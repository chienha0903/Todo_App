package consumer

import (
	"context"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/chienha0903/Todo_App/pkg/rabbitmq"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
)

const maxRetries = 3

type UserDeleteConsumer struct {
	todoCmd        gateway.TodoCommandGateway
	processedRepo  gateway.ProcessedEventGateway
	txGW           gateway.TransactionGateway
	amqpConn       *amqp.Connection
	workerPoolSize int
}

func NewUserDeleteConsumer(
	todoCmd gateway.TodoCommandGateway,
	processedRepo gateway.ProcessedEventGateway,
	txGW gateway.TransactionGateway,
	conn *amqp.Connection,
	workerPoolSize int,
) *UserDeleteConsumer {
	return &UserDeleteConsumer{
		todoCmd:        todoCmd,
		processedRepo:  processedRepo,
		txGW:           txGW,
		amqpConn:       conn,
		workerPoolSize: workerPoolSize,
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

	// QoS khớp với số worker: RabbitMQ không gửi quá workerPoolSize message chưa Ack
	if err := ch.Qos(c.workerPoolSize, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		rabbitmq.QueueUserDeleteRequested,
		"", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	// jobs là buffered channel dùng để phân phối message đến các worker.
	// Buffer = workerPoolSize để main loop không bị block khi tất cả worker đang bận.
	jobs := make(chan amqp.Delivery, c.workerPoolSize)

	var wg sync.WaitGroup
	for i := range c.workerPoolSize {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.runWorker(ctx, workerID, jobs)
		}(i)
	}

	slog.Info("todo consumer started", "worker_pool_size", c.workerPoolSize)

	for {
		select {
		case <-ctx.Done():
			close(jobs) // báo tất cả worker dừng sau khi drain
			wg.Wait()
			return nil
		case msg, ok := <-msgs:
			if !ok {
				close(jobs)
				wg.Wait()
				return nil
			}
			// Gửi vào jobs; nếu ctx bị cancel trước khi có slot thì Nack để requeue
			select {
			case jobs <- msg:
			case <-ctx.Done():
				_ = msg.Nack(false, true)
				close(jobs)
				wg.Wait()
				return nil
			}
		}
	}
}

// runWorker đọc message từ jobs channel và xử lý tuần tự.
// Khi jobs bị close, vòng range tự kết thúc.
func (c *UserDeleteConsumer) runWorker(ctx context.Context, workerID int, jobs <-chan amqp.Delivery) {
	slog.Info("worker started", "worker_id", workerID)
	for msg := range jobs {
		c.handleMessage(ctx, msg)
	}
	slog.Info("worker stopped", "worker_id", workerID)
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
