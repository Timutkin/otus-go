package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	t.Run("create event", func(t *testing.T) {
		event := createEvent()
		ms := New()
		err := ms.Create(context.Background(), event)
		assert.NoError(t, err)
		val, ok := ms.evenIDByEvent[event.ID]
		assert.True(t, ok)
		assert.Equal(t, event, val)
		events, ok := ms.userIDByEvent[*event.UserID]
		assert.True(t, ok)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, event, events[0])
	})

	t.Run("delete event", func(t *testing.T) {
		event := createEvent()
		ms := New()
		_ = ms.Create(context.Background(), event)
		err := ms.Delete(context.Background(), event.ID)
		assert.NoError(t, err)
		_, ok := ms.userIDByEvent[event.ID]
		assert.False(t, ok)
		_, ok = ms.userIDByEvent[*event.UserID]
		assert.False(t, ok)
	})

	t.Run("get events by user id", func(t *testing.T) {
		ms := New()
		userID := uuid.New()
		e1 := createEvent()
		e2 := createEvent()
		e1.UserID = &userID
		e2.UserID = &userID
		_ = ms.Create(context.Background(), e1)
		_ = ms.Create(context.Background(), e2)
		id, err := ms.GetEventsByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.Equal(t, id[0], e1)
		assert.Equal(t, id[1], e2)
	})

	t.Run("get event by id", func(t *testing.T) {
		event := createEvent()
		ms := New()
		_ = ms.Create(context.Background(), event)
		e, _ := ms.GetByID(context.Background(), event.ID)
		assert.Equal(t, event, e)
	})

	t.Run("update event", func(t *testing.T) {
		t.Run("all fields", func(t *testing.T) {
			event := createEvent()
			ms := New()
			_ = ms.Create(context.Background(), event)

			newTitle := "new title"
			newDescription := "new description"
			eventDateTime := *event.DateTime
			newDateTime := eventDateTime.Add(2 * time.Hour)
			eventNotificationTime := *event.NotificationTime
			newNotificationTime := eventNotificationTime.Add(time.Minute)
			eventDuration := *event.EventDuration
			newDuration := eventDuration + time.Minute

			err := ms.Update(context.Background(), storage.Event{
				ID:               event.ID,
				Title:            &newTitle,
				Description:      &newDescription,
				DateTime:         &newDateTime,
				NotificationTime: &newNotificationTime,
				EventDuration:    &newDuration,
			})
			assert.NoError(t, err)

			updatedEvent, err := ms.GetByID(context.Background(), event.ID)
			assert.NoError(t, err)
			assert.Equal(t, newTitle, *updatedEvent.Title)
			assert.Equal(t, newDescription, *updatedEvent.Description)
			assert.Equal(t, newDateTime, *updatedEvent.DateTime)
			assert.Equal(t, newNotificationTime, *updatedEvent.NotificationTime)
			assert.Equal(t, newDuration, *updatedEvent.EventDuration)

			events, err := ms.GetEventsByUserID(context.Background(), *event.UserID)
			updatedEvent = events[0]
			assert.NoError(t, err)
			assert.Equal(t, newTitle, *updatedEvent.Title)
			assert.Equal(t, newDescription, *updatedEvent.Description)
			assert.Equal(t, newDateTime, *updatedEvent.DateTime)
			assert.Equal(t, newNotificationTime, *updatedEvent.NotificationTime)
			assert.Equal(t, newDuration, *updatedEvent.EventDuration)
		})

		t.Run("single field", func(t *testing.T) {
			event := createEvent()
			ms := New()
			_ = ms.Create(context.Background(), event)

			newTitle := "only title changed"
			err := ms.Update(context.Background(), storage.Event{
				ID:    event.ID,
				Title: &newTitle,
			})
			assert.NoError(t, err)
			updatedEvent, err := ms.GetByID(context.Background(), event.ID)
			assert.NoError(t, err)
			assert.Equal(t, newTitle, *updatedEvent.Title)
			assert.Equal(t, *event.Description, *updatedEvent.Description)
			assert.Equal(t, *event.DateTime, *updatedEvent.DateTime)
			assert.Equal(t, *event.EventDuration, *updatedEvent.EventDuration)
			assert.Equal(t, *event.NotificationTime, *updatedEvent.NotificationTime)
		})

		t.Run("user id update returns error", func(t *testing.T) {
			event := createEvent()
			ms := New()
			_ = ms.Create(context.Background(), event)

			newUserID := uuid.New()
			err := ms.Update(context.Background(), storage.Event{
				ID:     event.ID,
				UserID: &newUserID,
			})
			assert.ErrorIs(t, err, storage.ErrUserIDShouldBeNil)
		})
	})
}

func createEvent() storage.Event {
	id := uuid.New()
	userID := uuid.New()
	title := "title"
	description := "description"
	now := time.Now()
	nTime := time.Now()
	duration, _ := time.ParseDuration("1h")
	event := storage.Event{
		ID:               id,
		Title:            &title,
		DateTime:         &now,
		EventDuration:    &duration,
		Description:      &description,
		UserID:           &userID,
		NotificationTime: &nTime,
	}
	return event
}
