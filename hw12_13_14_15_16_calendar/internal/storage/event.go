package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrEventNotFoundErr = errors.New("event not found")

type Event struct {
	ID               uuid.UUID      `db:"id"`
	Title            *string        `db:"title"`
	DateTime         *time.Time     `db:"date_time"`
	EventDuration    *time.Duration `db:"event_duration"`
	Description      *string        `db:"description"`
	UserID           *uuid.UUID     `db:"user_id"`
	NotificationTime *time.Time     `db:"notification_time"`
}
