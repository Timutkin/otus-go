//go:build migrations
// +build migrations

package integration_test

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	. "github.com/onsi/ginkgo/v2" //nolint
	g "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/logger"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/timutkin/otus-go/hw12_13_14_15_calendar/migrations"
)

var _ = Describe("NotificationScheduler", Ordered, func() {
	ctx := context.Background()
	cfg := config.NewCalendarConfig("../../configs/calendar_config.yaml")

	postgresContainer, connStr := upPostgres(ctx, cfg)
	var notificationScheduler scheduler.NotificationScheduler
	var storage scheduler.Storage
	var lg scheduler.NotificationSchedulerLogger
	var mockSender MockSenderService

	BeforeEach(func() {
		lg = logger.New()
		sql := sqlstorage.New(connStr, cfg.DB)
		err := sql.Connect(context.Background())
		if err != nil {
			log.Printf("Connection string: %s", connStr)
			log.Fatal().Err(err).Msg("failed connect to db")
		}
		storage = sql
		mockSender = MockSenderService{messages: [][]byte{}}
		notificationScheduler = scheduler.NewNotificationScheduler(storage, &mockSender, lg, "test-queue")
	})

	When("create notification scheduler", func() {
		It("should initialize scheduler with correct parameters", func(ctx SpecContext) {
			g.Expect(notificationScheduler).ShouldNot(g.BeNil())
		}, SpecTimeout(time.Second*1))
	})

	When("get scheduler jobs", func() {
		It("should return jobs list with two jobs", func(ctx SpecContext) {
			jobs := notificationScheduler.GetJobs()
			g.Expect(jobs).Should(g.HaveLen(2))
		}, SpecTimeout(time.Second*1))

		It("should have correct cron expressions", func(ctx SpecContext) {
			jobs := notificationScheduler.GetJobs()
			g.Expect(jobs[0].Cron).Should(g.Equal("* * * * *"))
			g.Expect(jobs[1].Cron).Should(g.Equal("0 0 * * *"))
		}, SpecTimeout(time.Second*1))

		It("should have callable functions", func(ctx SpecContext) {
			jobs := notificationScheduler.GetJobs()
			g.Expect(jobs[0].Function).ShouldNot(g.BeNil())
			g.Expect(jobs[1].Function).ShouldNot(g.BeNil())
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

type MockSenderService struct {
	messages [][]byte
}

func (m *MockSenderService) Send(queueName string, message []byte) error {
	m.messages = append(m.messages, message)
	return nil
}

func upPostgres(
	ctx context.Context, cfg config.CalendarConfig,
) (testcontainers.Container, string) {
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

	return postgresContainer, connStr
}
