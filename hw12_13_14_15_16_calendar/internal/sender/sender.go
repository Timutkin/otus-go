package sender

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SenderLogger interface {
	InfoWithParams(msg string, params map[string]string)
	Info(msg string)
}

type NotificationConsumer interface {
	Consume(queueName string) <-chan amqp.Delivery
}

type NotificationSender struct {
	consumer  NotificationConsumer
	queueName string
	logger    SenderLogger
}

func NewNotificationSender(consumer NotificationConsumer, queueName string, logger SenderLogger) NotificationSender {
	return NotificationSender{
		consumer:  consumer,
		queueName: queueName,
		logger:    logger,
	}
}

func (n NotificationSender) StartListening(ctx context.Context) {
	messages := n.consumer.Consume(n.queueName)
	n.logger.Info("start listening queue with name " + n.queueName)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			body := string(msg.Body)
			n.logger.InfoWithParams("got message", map[string]string{"queueName": n.queueName, "message": body})
		}
	}
}
