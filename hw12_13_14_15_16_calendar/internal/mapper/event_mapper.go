package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/grpc/pb"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventMapper struct{}

func (e EventMapper) CreateEventRequestToEvent(rq *pb.CreateEventRequest) *storage.Event {
	id, _ := uuid.NewRandom()
	event := rq.GetEvent()
	dateTime := event.DateTime.AsTime()
	eventDuration := time.Duration(event.EventDuration)
	userID, _ := uuid.Parse(event.UserId)
	var notificationTime *time.Time
	if event.GetNotificationTime() != nil {
		asTime := event.NotificationTime.AsTime()
		notificationTime = &asTime
	}
	return &storage.Event{
		ID:               id,
		Title:            &event.Title,
		DateTime:         &dateTime,
		EventDuration:    &eventDuration,
		Description:      &event.Description,
		UserID:           &userID,
		NotificationTime: notificationTime,
	}
}

func (e EventMapper) StorageEventToEvent(event storage.Event) *pb.Event {
	pbEvent := &pb.Event{Id: event.ID.String()}
	if event.Title != nil {
		pbEvent.Title = *event.Title
	}
	if event.EventDuration != nil {
		pbEvent.EventDuration = int64(*event.EventDuration)
	}
	if event.Description != nil {
		pbEvent.Description = *event.Description
	}
	if event.UserID != nil {
		pbEvent.UserId = event.UserID.String()
	}
	if event.DateTime != nil {
		pbEvent.DateTime = timestamppb.New(*event.DateTime)
	}
	if event.NotificationTime != nil {
		pbEvent.NotificationTime = timestamppb.New(*event.NotificationTime)
	}
	return pbEvent
}

func (e EventMapper) UpdateEventRequestToEvent(rq *pb.UpdateEventRequest) storage.Event {
	id, _ := uuid.Parse(rq.Id)
	storageEvent := &storage.Event{
		ID: id,
	}
	if rq.Title != nil {
		storageEvent.Title = rq.Title
	}
	if rq.Description != nil {
		storageEvent.Description = rq.Description
	}
	if rq.DateTime != nil {
		asTime := rq.DateTime.AsTime()
		storageEvent.DateTime = &asTime
	}
	if rq.EventDuration != nil {
		duration := time.Duration(*rq.EventDuration)
		storageEvent.EventDuration = &duration
	}
	if rq.NotificationTime != nil {
		asTime := rq.NotificationTime.AsTime()
		storageEvent.NotificationTime = &asTime
	}

	return *storageEvent
}
