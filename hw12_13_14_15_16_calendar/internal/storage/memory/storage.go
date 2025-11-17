package memorystorage

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	userIDByEvent map[uuid.UUID][]storage.Event
	evenIDByEvent map[uuid.UUID]storage.Event
	mu            sync.RWMutex
}

func (s *Storage) Update(ctx context.Context, newEvent storage.Event) error {
	e, err := s.GetByID(ctx, newEvent.ID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		return err
	}
	if newEvent.UserID != nil {
		return storage.ErrUserIDShouldBeNil
	}
	if newEvent.Title != nil {
		e.Title = newEvent.Title
	}
	if newEvent.DateTime != nil {
		e.DateTime = newEvent.DateTime
	}
	if newEvent.EventDuration != nil {
		e.EventDuration = newEvent.EventDuration
	}
	if newEvent.Description != nil {
		e.Description = newEvent.Description
	}
	if newEvent.NotificationTime != nil {
		e.NotificationTime = newEvent.NotificationTime
	}
	s.evenIDByEvent[e.ID] = e
	events := s.userIDByEvent[*e.UserID]
	for i, val := range events {
		if val.ID == e.ID {
			events[i] = e
			break
		}
	}
	return nil
}

func (s *Storage) Create(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userIDByEvent[*event.UserID] = append(s.userIDByEvent[*event.UserID], event)
	s.evenIDByEvent[event.ID] = event
	return nil
}

func (s *Storage) Delete(ctx context.Context, eventID uuid.UUID) error {
	event, err := s.GetByID(ctx, eventID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		return err
	}
	delete(s.evenIDByEvent, event.ID)
	delete(s.userIDByEvent, *event.UserID)
	return nil
}

func (s *Storage) GetEventsByUserID(_ context.Context, userID uuid.UUID) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userIDByEvent[userID], nil
}

func (s *Storage) GetByID(_ context.Context, eventID uuid.UUID) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	event, ok := s.evenIDByEvent[eventID]
	if !ok {
		return storage.Event{}, storage.ErrEventNotFoundErr
	}
	return event, nil
}

func New() *Storage {
	return &Storage{
		userIDByEvent: make(map[uuid.UUID][]storage.Event),
		evenIDByEvent: make(map[uuid.UUID]storage.Event),
	}
}
