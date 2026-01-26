//go:build migrations
// +build migrations

package integration_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib"
	. "github.com/onsi/ginkgo/v2" //nolint
	g "github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/grpc/pb"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/logger"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/mapper"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/service"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/timutkin/otus-go/hw12_13_14_15_calendar/migrations"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("EventService", Ordered, func() {
	ctx := context.Background()
	cfg := config.NewCalendarConfig("../../configs/calendar_config.yaml")

	postgresContainer, connStr := upPostgresAndGoMigrate(ctx, cfg)
	var eventService pb.EventServiceServer
	var storage service.Storage
	var lg service.Logger
	var eventRq pb.CreateEventRequest
	userID := uuid.NewString()
	eventDuration := GinkgoRandomSeed()
	dateTime := timestamppb.Now()

	BeforeEach(func() {
		eventRq = pb.CreateEventRequest{Event: &pb.Event{
			Title:         "Test Event",
			Description:   "This is a test event",
			DateTime:      dateTime,
			UserId:        userID,
			EventDuration: eventDuration,
		}}

		lg = logger.New()
		sql := sqlstorage.New(connStr, cfg.DB)
		err := sql.Connect(context.Background())
		if err != nil {
			log.Printf("Connection string: %s", connStr)
			log.Fatal().Err(err).Msg("failed connect to db")
		}
		storage = sql
		eventService = service.NewEventService(storage, lg, mapper.EventMapper{})
	})

	When("create event without notification", func() {
		It("should save event without notification_status", func(ctx SpecContext) {
			_, err := eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())
		}, SpecTimeout(time.Second*1))

		It("should retrieve event by userID with correct fields", func(ctx SpecContext) {
			_, err := eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())

			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			g.Expect(events).Should(g.HaveLen(1))

			event := events[0]
			g.Expect(eventRq.Event.Title).Should(g.Equal(*event.Title))
			g.Expect(eventRq.Event.Description).Should(g.Equal(*event.Description))
			g.Expect(eventRq.Event.DateTime.AsTime()).Should(g.Equal(event.DateTime.UTC()))
			g.Expect(eventRq.Event.UserId).Should(g.Equal(event.UserID.String()))
			g.Expect(time.Duration(eventRq.Event.EventDuration)).Should(g.Equal(*event.EventDuration))
			g.Expect(eventRq.Event.NotificationTime).Should(g.BeNil())
		}, SpecTimeout(time.Second*1))

		AfterEach(func() {
			events, _ := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			for _, event := range events {
				_ = storage.Delete(context.Background(), event.ID)
			}
		})
	})

	When("create event with notification", func() {
		BeforeEach(func() {
			eventRq.Event.NotificationTime = timestamppb.Now()
		})

		It("should save event with notification_status", func(ctx SpecContext) {
			_, err := eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())
		}, SpecTimeout(time.Second*1))

		It("should retrieve event with notification details", func(ctx SpecContext) {
			_, err := eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())

			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			event := events[0]

			g.Expect("PENDING").Should(g.Equal(*event.NotificationStatus))
			g.Expect(eventRq.Event.NotificationTime.AsTime()).Should(g.Equal(event.NotificationTime.UTC()))
		}, SpecTimeout(time.Second*1))

		AfterEach(func() {
			events, _ := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			for _, event := range events {
				_ = storage.Delete(context.Background(), event.ID)
			}
			eventRq.Event.NotificationTime = nil
		})
	})

	When("get event by id", func() {
		var createdEventID string

		BeforeEach(func() {
			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			if len(events) > 0 {
				_ = storage.Delete(context.Background(), events[0].ID)
			}

			_, err = eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())
			events, _ = storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			createdEventID = events[0].ID.String()
		})

		It("should retrieve event by id", func(ctx SpecContext) {
			getReq := &pb.ByIdRequest{EventId: createdEventID}
			resp, err := eventService.GetById(context.Background(), getReq)
			g.Expect(err).Should(g.BeNil())
			g.Expect(resp.Event.Id).Should(g.Equal(createdEventID))
			g.Expect(resp.Event.Title).Should(g.Equal(eventRq.Event.Title))
			g.Expect(resp.Event.Description).Should(g.Equal(eventRq.Event.Description))
		}, SpecTimeout(time.Second*1))

		AfterEach(func() {
			events, _ := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			for _, event := range events {
				_ = storage.Delete(context.Background(), event.ID)
			}
		})
	})

	When("update event", func() {
		var createdEventID string

		BeforeEach(func() {
			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			if len(events) > 0 {
				_ = storage.Delete(context.Background(), events[0].ID)
			}

			_, err = eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())
			events, _ = storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			createdEventID = events[0].ID.String()
		})

		It("should update event fields", func(ctx SpecContext) {
			updatedTitle := "Updated Title"
			updatedDescription := "Updated Description"
			newDateTime := timestamppb.Now()

			updateReq := &pb.UpdateEventRequest{
				Id:          createdEventID,
				Title:       &updatedTitle,
				Description: &updatedDescription,
				DateTime:    newDateTime,
			}

			resp, err := eventService.UpdateEvent(context.Background(), updateReq)
			g.Expect(err).Should(g.BeNil())
			g.Expect(resp.Event.Title).Should(g.Equal(updatedTitle))
			g.Expect(resp.Event.Description).Should(g.Equal(updatedDescription))
		}, SpecTimeout(time.Second*1))

		AfterEach(func() {
			events, _ := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			for _, event := range events {
				_ = storage.Delete(context.Background(), event.ID)
			}
		})
	})

	When("delete event", func() {
		var createdEventID string

		BeforeEach(func() {
			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			if len(events) > 0 {
				_ = storage.Delete(context.Background(), events[0].ID)
			}

			_, err = eventService.CreateEvent(context.Background(), &eventRq)
			g.Expect(err).Should(g.BeNil())
			events, _ = storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			createdEventID = events[0].ID.String()
		})

		It("should delete event successfully", func(ctx SpecContext) {
			deleteReq := &pb.ByIdRequest{EventId: createdEventID}
			_, err := eventService.DeleteEvent(context.Background(), deleteReq)
			g.Expect(err).Should(g.BeNil())

			events, err := storage.GetEventsByUserID(context.Background(), uuid.MustParse(eventRq.Event.UserId))
			g.Expect(err).Should(g.BeNil())
			g.Expect(events).Should(g.HaveLen(0))
		}, SpecTimeout(time.Second*1))
	})

	AfterAll(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		} else {
			log.Printf("postgresql container was terminated")
		}
	})
})

func upPostgresAndGoMigrate(ctx context.Context, cfg config.CalendarConfig) (testcontainers.Container, string) {
	dbCfg := cfg.DB
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbCfg.Dbname),
		postgres.WithUsername(dbCfg.User),
		postgres.WithPassword(dbCfg.Password),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, ""
	}

	log.Print("postgresql container started ...")

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get connection string from container")
	}

	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("error while connect to db")
	}
	defer db.Close()

	if err = goose.RunContext(ctx, "up", db, "../../migrations"); err != nil {
		log.Fatal().Err(err).Msg("goose run failed")
	}

	return postgresContainer, connStr
}
