package sender

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/scheduler"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	InfoWithParams(msg string, params map[string]string)
	Info(msg string)
	ErrorWithParams(msg string, params map[string]string, err error)
}

type NotificationConsumer interface {
	Consume(queueName string) (<-chan amqp.Delivery, error)
}

type Storage interface {
	Update(ctx context.Context, newEvent storage.Event) error
}

type NotificationSender struct {
	consumer  NotificationConsumer
	queueName string
	logger    Logger
	storage   Storage
}

func NewNotificationSender(
	consumer NotificationConsumer, queueName string, logger Logger, storage Storage,
) NotificationSender {
	return NotificationSender{
		consumer:  consumer,
		queueName: queueName,
		logger:    logger,
		storage:   storage,
	}
}

func (n NotificationSender) StartListening(ctx context.Context) {
	messages, err := n.consumer.Consume(n.queueName)
	for err != nil {
		n.logger.ErrorWithParams(
			"consume messages", map[string]string{
				"queueName": n.queueName,
			},
			err,
		)
		time.Sleep(time.Second * 60)
		messages, err = n.consumer.Consume(n.queueName)
	}
	n.logger.Info("start listening queue with name " + n.queueName)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			body := string(msg.Body)
			n.logger.InfoWithParams("got message", map[string]string{"queueName": n.queueName, "message": body})
			var notification scheduler.Notification
			err := json.Unmarshal(msg.Body, &notification)
			if err != nil {
				n.logger.ErrorWithParams(
					"unmarshal notification", map[string]string{"queueName": n.queueName, "message": body}, err,
				)
				return
			}
			status := "SENT"
			err = n.storage.Update(context.Background(), storage.Event{
				ID:                 uuid.MustParse(notification.ID),
				NotificationStatus: &status,
			})
			if err != nil {
				n.logger.ErrorWithParams(
					"update status to SENT", map[string]string{"queueName": n.queueName, "message": body}, err,
				)
				return
			}
		}
	}
}
