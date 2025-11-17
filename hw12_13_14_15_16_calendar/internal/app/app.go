package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	storage Storage
	logger  Logger
}

type Logger interface {
	DebugWithParams(msg string, params map[string]string)
}

type Storage interface {
	Create(ctx context.Context, event storage.Event) error
	GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]storage.Event, error)
	GetByID(ctx context.Context, eventID uuid.UUID) (storage.Event, error)
	Update(ctx context.Context, event storage.Event) error
	Delete(ctx context.Context, eventID uuid.UUID) error
}

func New(logger Logger, storage Storage) *App {
	return &App{
		storage: storage,
		logger:  logger,
	}
}

//nolint
func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}
