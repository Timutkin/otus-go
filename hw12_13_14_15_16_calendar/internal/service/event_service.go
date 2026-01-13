package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/grpc/pb"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService struct {
	eventStorage Storage
	lg           Logger
	eventMapper  EventMapper
	pb.UnimplementedEventServiceServer
}

type Storage interface {
	Create(ctx context.Context, event storage.Event) error
	GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]storage.Event, error)
	GetByID(ctx context.Context, eventID uuid.UUID) (storage.Event, error)
	Update(ctx context.Context, event storage.Event) error
	Delete(ctx context.Context, eventID uuid.UUID) error
}

type Logger interface {
	ErrorWithAny(msg string, name string, param any)
	ErrorWithParams(msg string, params map[string]string, err error)
	Error(msg string, err error)
	InfoWithParams(msg string, params map[string]string)
	Info(msg string)
}

type EventMapper interface {
	CreateEventRequestToEvent(rq *pb.CreateEventRequest) *storage.Event
	StorageEventToEvent(event storage.Event) *pb.Event
	UpdateEventRequestToEvent(rq *pb.UpdateEventRequest) storage.Event
}

func NewEventService(eventStorage Storage, lg Logger, eventMapper EventMapper) pb.EventServiceServer {
	return EventService{
		eventStorage: eventStorage,
		lg:           lg,
		eventMapper:  eventMapper,
	}
}

func (e EventService) CreateEvent(ctx context.Context, rq *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	requestEvent := rq.Event
	e.lg.InfoWithParams("create event request", map[string]string{
		"userId":  requestEvent.GetUserId(),
		"title":   requestEvent.GetTitle(),
		"eventId": requestEvent.GetId(),
		"method":  "CreateEvent",
	})
	err := validateEvent(requestEvent)
	if err != nil {
		e.lg.ErrorWithAny("validation failed", "event", requestEvent)
		return nil, err
	}
	event := e.eventMapper.CreateEventRequestToEvent(rq)
	for {
		err = e.eventStorage.Create(ctx, *event)
		if err != nil {
			if errors.Is(err, sqlstorage.ErrEventIDAlreadyExist) {
				e.lg.InfoWithParams("event id already exists, generating new id", map[string]string{
					"oldEventId": event.ID.String(),
				})
				event.ID = uuid.New()
				continue
			}
			e.lg.ErrorWithParams("failed to create event", map[string]string{
				"eventId": event.ID.String(),
				"userId":  requestEvent.GetUserId(),
			}, err)
			return nil, status.Error(codes.Internal, "failed to create event")
		}
		break
	}
	e.lg.InfoWithParams("event created successfully", map[string]string{
		"eventId": event.ID.String(),
		"userId":  requestEvent.GetUserId(),
		"title":   requestEvent.GetTitle(),
	})
	return &pb.CreateEventResponse{}, nil
}

func (e EventService) GetEventsByUserID(ctx context.Context, rq *pb.GetByUserIdRequest) (*pb.EventsResponse, error) {
	requestUserID := rq.GetUserId()
	e.lg.InfoWithParams("get events by user id request", map[string]string{
		"userId": requestUserID,
		"method": "GetEventsByUserID",
	})
	if requestUserID == "" {
		e.lg.Error("missing required field: userId", nil)
		return nil, status.Error(codes.InvalidArgument, "rq missing required field: userId")
	}
	id, err := uuid.Parse(requestUserID)
	if err != nil {
		e.lg.ErrorWithParams("invalid userId format", map[string]string{
			"userId": requestUserID,
		}, err)
		return nil, status.Error(codes.InvalidArgument, "invalid userId")
	}
	events, err := e.eventStorage.GetEventsByUserID(ctx, id)
	if err != nil {
		e.lg.ErrorWithParams("failed to get events by user id", map[string]string{
			"userId": requestUserID,
		}, err)
		return nil, status.Error(codes.Internal, "failed to get event by userId")
	}
	if len(events) == 0 {
		e.lg.InfoWithParams("no events found for user", map[string]string{
			"userId": requestUserID,
		})
		return &pb.EventsResponse{Events: make([]*pb.Event, 0)}, nil
	}
	res := make([]*pb.Event, 0, len(events))
	for _, v := range events {
		res = append(res, e.eventMapper.StorageEventToEvent(v))
	}
	e.lg.InfoWithParams("events retrieved successfully", map[string]string{
		"userId":      requestUserID,
		"eventsCount": strconv.Itoa(len(res)),
	})
	return &pb.EventsResponse{Events: res}, nil
}

func (e EventService) GetById(ctx context.Context, request *pb.ByIdRequest) (*pb.EventResponse, error) { //nolint
	requestEventID := request.GetEventId()
	e.lg.InfoWithParams("get event by id request", map[string]string{
		"eventId": requestEventID,
		"method":  "GetById",
	})
	if requestEventID == "" {
		e.lg.Error("missing required field: eventId", nil)
		return nil, status.Error(codes.InvalidArgument, "request missing required field: eventId")
	}
	id, err := uuid.Parse(requestEventID)
	if err != nil {
		e.lg.ErrorWithParams("invalid eventId format", map[string]string{
			"eventId": requestEventID,
		}, err)
		return nil, status.Error(codes.InvalidArgument, "invalid eventId")
	}
	event, err := e.eventStorage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFoundErr) {
			e.lg.InfoWithParams("event not found", map[string]string{
				"eventId": requestEventID,
			})
			return &pb.EventResponse{}, nil
		}
		e.lg.ErrorWithParams("failed to get event by id", map[string]string{
			"eventId": requestEventID,
		}, err)
		return nil, status.Error(codes.Internal, "failed to get event by userId")
	}
	response := e.eventMapper.StorageEventToEvent(event)
	e.lg.InfoWithParams("event retrieved successfully", map[string]string{
		"eventId": requestEventID,
		"userId":  response.GetUserId(),
	})
	return &pb.EventResponse{Event: response}, nil
}

func (e EventService) UpdateEvent(ctx context.Context, request *pb.UpdateEventRequest) (*pb.EventResponse, error) {
	requestID := request.GetId()
	e.lg.InfoWithParams("update event request", map[string]string{
		"eventId": requestID,
		"method":  "UpdateEvent",
	})
	if requestID == "" {
		e.lg.Error("missing required field: id", nil)
		return nil, status.Error(codes.InvalidArgument, "request missing required field: id")
	}
	id, err := uuid.Parse(requestID)
	if err != nil {
		e.lg.ErrorWithParams("invalid id format", map[string]string{
			"eventId": requestID,
		}, err)
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	event := e.eventMapper.UpdateEventRequestToEvent(request)
	err = e.eventStorage.Update(ctx, event)
	if err != nil {
		e.lg.ErrorWithParams("failed to update event", map[string]string{
			"eventId": requestID,
		}, err)
		return nil, status.Error(codes.Internal, "failed to update event")
	}
	updatedEvent, err := e.eventStorage.GetByID(ctx, id)
	if err != nil {
		e.lg.ErrorWithParams("failed to get updated event", map[string]string{
			"eventId": requestID,
		}, err)
		return nil, status.Error(codes.Internal, "failed to get updated event")
	}
	response := e.eventMapper.StorageEventToEvent(updatedEvent)
	e.lg.InfoWithParams("event updated successfully", map[string]string{
		"eventId": requestID,
		"userId":  response.GetUserId(),
	})
	return &pb.EventResponse{Event: response}, nil
}

func (e EventService) DeleteEvent(ctx context.Context, request *pb.ByIdRequest) (*pb.DeleteEventResponse, error) {
	requestEventID := request.GetEventId()
	e.lg.InfoWithParams("delete event request", map[string]string{
		"eventId": requestEventID,
		"method":  "DeleteEvent",
	})
	if requestEventID == "" {
		e.lg.Error("missing required field: eventId", nil)
		return nil, status.Error(codes.InvalidArgument, "request missing required field: eventId")
	}
	id, err := uuid.Parse(requestEventID)
	if err != nil {
		e.lg.ErrorWithParams("invalid eventId format", map[string]string{
			"eventId": requestEventID,
		}, err)
		return nil, status.Error(codes.InvalidArgument, "invalid eventId")
	}
	err = e.eventStorage.Delete(ctx, id)
	if err != nil {
		e.lg.ErrorWithParams("failed to delete event", map[string]string{
			"eventId": requestEventID,
		}, err)
		return nil, status.Error(codes.Internal, "failed to delete by eventId")
	}
	e.lg.InfoWithParams("event deleted successfully", map[string]string{
		"eventId": requestEventID,
	})
	return &pb.DeleteEventResponse{}, nil
}

func (e EventService) mustEmbedUnimplementedEventServiceServer() {} //nolint

func validateEvent(event *pb.Event) error {
	if event.GetTitle() == "" {
		return status.Errorf(codes.InvalidArgument, "request missing required field: title")
	}
	if event.GetDateTime() == nil {
		return status.Errorf(codes.InvalidArgument, "request missing required field: dateTime")
	}
	if event.GetEventDuration() == 0 {
		return status.Errorf(codes.InvalidArgument, "request missing required field: eventDuration")
	}
	if event.GetDescription() == "" {
		return status.Errorf(codes.InvalidArgument, "request missing required field: description")
	}
	if event.GetUserId() == "" {
		return status.Errorf(codes.InvalidArgument, "request missing required field: userId")
	}
	if _, err := uuid.Parse(event.GetUserId()); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid userId")
	}
	return nil
}
