package client

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RabbitLogger interface {
	ErrorWithParams(msg string, params map[string]string, err error)
}

type RabbitClient struct {
	connectionString string
	connection       *amqp.Connection
	logger           RabbitLogger
}

func NewRabbitClient(connectionString string, logger RabbitLogger) *RabbitClient {
	client := &RabbitClient{
		connectionString: connectionString,
		logger:           logger,
	}
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("create connect to RabbitMQ")
	}
	client.connection = conn
	return client
}

func (c *RabbitClient) Send(queueName string, message []byte) error {
	ch, err := c.connection.Channel()
	if err != nil {
		c.logger.ErrorWithParams(
			"get channel", map[string]string{
				"queueName": queueName,
			},
			err,
		)
	}
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		c.logger.ErrorWithParams(
			"declare queue", map[string]string{
				"queueName": queueName,
			},
			err,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
	return err
}

func (c *RabbitClient) Consume(queueName string) (<-chan amqp.Delivery, error) {
	ch, err := c.connection.Channel()
	if err != nil {
		c.logger.ErrorWithParams(
			"get channel", map[string]string{
				"queueName": queueName,
			},
			err,
		)
	}
	messages, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}
	return messages, err
}
