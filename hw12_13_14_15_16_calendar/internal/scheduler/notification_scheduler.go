package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

type NotificationSchedulerLogger interface {
	Error(msg string, err error)
	Debug(msg string)
	ErrorWithParams(msg string, params map[string]string, err error)
}

type Scheduler interface {
	CreateJob(cron string, function any, functionParam ...any) error
}

type SenderService interface {
	Send(queueName string, message []byte) error
}

type Storage interface {
	FindByCurrentTimeByMinutesAndPendingStatus() ([]storage.Event, error)
	Update(ctx context.Context, newEvent storage.Event) error
	FindByDateTimeMoreOrEqual(dateTime time.Time) ([]storage.Event, error)
	Delete(ctx context.Context, eventID uuid.UUID) error
}

type NotificationScheduler struct {
	storage   Storage
	sender    SenderService
	logger    NotificationSchedulerLogger
	queueName string
}

func NewNotificationScheduler(
	storage Storage, sender SenderService, logger NotificationSchedulerLogger, queueName string,
) NotificationScheduler {
	notificationScheduler := NotificationScheduler{
		storage:   storage,
		sender:    sender,
		logger:    logger,
		queueName: queueName,
	}
	return notificationScheduler
}

type Notification struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	DateTime time.Time `json:"dateTime"`
	UserID   string    `json:"userId"`
}

func (n NotificationScheduler) GetJobs() []Job {
	return []Job{
		{
			Function:       n.sendEvents(),
			FunctionParams: nil,
			Cron:           "* * * * *",
		},
		{
			Function:       n.deleteOldEvents(),
			FunctionParams: nil,
			Cron:           "* * * * *",
		},
	}
}

func (n NotificationScheduler) sendEvents() func() {
	return func() {
		events, err := n.storage.FindByCurrentTimeByMinutesAndPendingStatus()
		if err != nil {
			n.logger.Error("get events for notification", err)
		}
		if len(events) == 0 {
			n.logger.Debug("found 0 events to sending")
			return
		}
		wg := sync.WaitGroup{}
		for _, e := range events {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n.handleEventForNotification(e)
				status := "SENT"
				err := n.storage.Update(context.Background(), storage.Event{
					ID:                 e.ID,
					NotificationStatus: &status,
				})
				if err != nil {
					n.logger.ErrorWithParams(
						"update event status", map[string]string{"eventId": e.ID.String()}, err,
					)
				}
			}()
		}
		wg.Wait()
		n.logger.Debug(fmt.Sprintf("finish processed %d events", len(events)))
	}
}

func (n NotificationScheduler) deleteOldEvents() func() {
	return func() {
		dateTime := time.Now().Add(-time.Hour * 24 * 365)
		events, err := n.storage.FindByDateTimeMoreOrEqual(dateTime)
		if err != nil {
			n.logger.Error("get old events", err)
			return
		}
		wg := sync.WaitGroup{}
		for _, e := range events {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := n.storage.Delete(context.Background(), e.ID)
				if err != nil {
					n.logger.ErrorWithParams(
						"delete old event",
						map[string]string{
							"id":       e.ID.String(),
							"dateTime": e.NotificationTime.String(),
						},
						err,
					)
				}
			}()
		}
		wg.Wait()
		n.logger.Debug(fmt.Sprintf("deleted %d events", len(events)))
	}
}

func (n NotificationScheduler) handleEventForNotification(e storage.Event) {
	notification, err := json.Marshal(Notification{
		ID:       e.ID.String(),
		Title:    *e.Title,
		DateTime: *e.NotificationTime,
		UserID:   e.UserID.String(),
	})
	if err != nil {
		n.logger.ErrorWithParams(
			"marshal event to notification",
			map[string]string{
				"id":       e.ID.String(),
				"title":    *e.Title,
				"dateTime": e.NotificationTime.String(),
				"userId":   e.UserID.String(),
			},
			err,
		)
		return
	}
	err = n.sender.Send(n.queueName, notification)
	if err != nil {
		n.logger.ErrorWithParams(
			"send event to queue",
			map[string]string{
				"queueName":    n.queueName,
				"notification": string(notification),
			},
			err,
		)
	}
}
