package consumer

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/chienha0903/Todo_App/pkg/rabbitmq"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
)

const maxRetries = 3

type UserDeleteConsumer struct {
	todoCmd       gateway.TodoCommandGateway
	processedRepo gateway.ProcessedEventGateway
	txGW          gateway.TransactionGateway
	amqpConn      *amqp.Connection
}

func NewUserDeleteConsumer(
	todoCmd gateway.TodoCommandGateway,
	processedRepo gateway.ProcessedEventGateway,
	txGW gateway.TransactionGateway,
	conn *amqp.Connection,
) *UserDeleteConsumer {
	return &UserDeleteConsumer{
		todoCmd:       todoCmd,
		processedRepo: processedRepo,
		txGW:          txGW,
		amqpConn:      conn,
	}
}

func (c *UserDeleteConsumer) Start(ctx context.Context) error {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := rabbitmq.SetupTopology(ch); err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		rabbitmq.QueueUserDeleteRequested,
		"", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	slog.Info("todo consumer started")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *UserDeleteConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	parsed, err := parseUserDeleteRequested(msg.Body)

	if err != nil {
		slog.Error("parse message failed", "error", err, "message_id", msg.MessageId)
		_ = msg.Nack(false, false)
		return
	}

	slog.Info("processing UserDeleteRequested",
		"user_id", parsed.Payload.UserID,
		"event_id", parsed.EventID,
	)

	if err := c.process(ctx, parsed); err != nil {
		retryCount := getDeathCount(msg)

		slog.Error("process failed",
			"error", err,
			"user_id", parsed.Payload.UserID,
			"retry_count", retryCount,
		)
		if retryCount >= maxRetries {
			slog.Warn("max retries exceeded, moving to DLQ",
				"event_id", parsed.EventID,
				"user_id", parsed.Payload.UserID,
			)
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Nack(false, true)
		}
		return
	}

	_ = msg.Ack(false)
	slog.Info("UserDeleteRequested processed",
		"user_id", parsed.Payload.UserID,
		"event_id", parsed.EventID,
	)
}

func (c *UserDeleteConsumer) process(ctx context.Context, msg *UserDeleteRequestedMessage) error {
	exists, err := c.processedRepo.Exists(ctx, msg.EventID)
	if err != nil {
		return err
	}
	if exists {
		slog.Info("duplicate event, skipping", "event_id", msg.EventID)
		return nil
	}

	return c.txGW.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := c.todoCmd.SoftDeleteByUserID(txCtx, msg.Payload.UserID); err != nil {
			return err
		}
		return c.processedRepo.Insert(txCtx, msg.EventID)
	})
}

func getDeathCount(msg amqp.Delivery) int64 {
	deaths, ok := msg.Headers["x-death"]
	if !ok {
		return 0
	}

	deathList, ok := deaths.([]interface{})
	if !ok || len(deathList) == 0 {
		return 0
	}

	entry, ok := deathList[0].(amqp.Table)
	if !ok {
		return 0
	}

	count, _ := entry["count"].(int64)

	return count
}
