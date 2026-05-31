package rabbitmq

import (
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "user.events"
	ExchangeDLX  = "user.events.dlx"

	QueueUserDeleteRequested    = "todo.user-delete-requested"
	QueueUserDeleteRequestedDLQ = "todo.user-delete-requested.dlq"

	RoutingKeyUserDeleteRequested = "UserDeleteRequested"
)

func Connect(url string) (*amqp.Connection, error) {
	var (
		conn *amqp.Connection
		err  error
	)
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			slog.Info("rabbitmq connected")
			return conn, nil
		}
		slog.Warn("rabbitmq not ready, retrying...", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("connect rabbitmq after 10 attempts: %w", err)
}

func SetupTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		ExchangeDLX, "direct",
		true, 
		false, 
		false,
		false, 
		nil,
	); err != nil {
		return fmt.Errorf("declare DLX: %w", err)
	}

	if err := ch.ExchangeDeclare(
		ExchangeName, "direct",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if _, err := ch.QueueDeclare(
		QueueUserDeleteRequestedDLQ,
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare DLQ: %w", err)
	}
	if err := ch.QueueBind(
		QueueUserDeleteRequestedDLQ,
		RoutingKeyUserDeleteRequested,
		ExchangeDLX,
		false, nil,
	); err != nil {
		return fmt.Errorf("bind DLQ: %w", err)
	}

	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    ExchangeDLX,
		"x-dead-letter-routing-key": RoutingKeyUserDeleteRequested,
	}
	if _, err := ch.QueueDeclare(
		QueueUserDeleteRequested,
		true, false, false, false, mainArgs,
	); err != nil {
		return fmt.Errorf("declare main queue: %w", err)
	}
	if err := ch.QueueBind(
		QueueUserDeleteRequested,
		RoutingKeyUserDeleteRequested,
		ExchangeName,
		false, nil,
	); err != nil {
		return fmt.Errorf("bind main queue: %w", err)
	}

	return nil
}
