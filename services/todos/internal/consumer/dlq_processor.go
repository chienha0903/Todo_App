package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/chienha0903/Todo_App/pkg/rabbitmq"
)

type DLQProcessor struct {
	amqpConn *amqp.Connection
}

func NewDLQProcessor(conn *amqp.Connection) *DLQProcessor {
	return &DLQProcessor{amqpConn: conn}
}

func (p *DLQProcessor) Start(ctx context.Context) error {
	ch, err := p.amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := rabbitmq.SetupTopology(ch); err != nil {
		return fmt.Errorf("setup topology: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	msgs, err := ch.Consume(
		rabbitmq.QueueUserDeleteRequestedDLQ,
		"dlq-processor",
		false,
		false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("consume DLQ: %w", err)
	}

	slog.Info("DLQ processor started", "queue", rabbitmq.QueueUserDeleteRequestedDLQ)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			p.handleDLQMessage(ctx, ch, msg)
		}
	}
}

type dlqMessageInfo struct {
	MessageID  string          `json:"message_id"`
	RoutingKey string          `json:"routing_key"`
	Body       json.RawMessage `json:"body"`
	DeathInfo  []deathEntry    `json:"death_info"`
}

type deathEntry struct {
	Queue    string `json:"queue"`
	Reason   string `json:"reason"`
	Count    int64  `json:"count"`
	Exchange string `json:"exchange"`
}

func (p *DLQProcessor) handleDLQMessage(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery) {
	info := dlqMessageInfo{
		MessageID:  msg.MessageId,
		RoutingKey: msg.RoutingKey,
		Body:       msg.Body,
		DeathInfo:  extractDeathInfo(msg),
	}

	infoJSON, _ := json.Marshal(info)
	slog.Error("DLQ message received — manual intervention required",
		"details", string(infoJSON),
	)

	if parsed, err := parseUserDeleteRequested(msg.Body); err == nil {
		slog.Error("DLQ: failed to delete todos for user",
			"user_id", parsed.Payload.UserID,
			"event_id", parsed.EventID,
		)
	}

	_ = msg.Ack(false)
}

func (p *DLQProcessor) Replay(ctx context.Context) error {
	ch, err := p.amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := rabbitmq.SetupTopology(ch); err != nil {
		return fmt.Errorf("setup topology: %w", err)
	}

	count := 0
	for {
		msg, ok, err := ch.Get(rabbitmq.QueueUserDeleteRequestedDLQ, false)
		if err != nil {
			return fmt.Errorf("get from DLQ: %w", err)
		}
		if !ok {
			break
		}

		err = ch.PublishWithContext(ctx,
			rabbitmq.ExchangeName,
			rabbitmq.RoutingKeyUserDeleteRequested,
			false, false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				MessageId:    msg.MessageId,
				Body:         msg.Body,
			},
		)
		if err != nil {
			_ = msg.Nack(false, true)
			return fmt.Errorf("replay publish: %w", err)
		}

		_ = msg.Ack(false)
		count++
		slog.Info("replayed DLQ message", "message_id", msg.MessageId)
	}

	slog.Info("DLQ replay complete", "replayed_count", count)
	return nil
}

func extractDeathInfo(msg amqp.Delivery) []deathEntry {
	deaths, ok := msg.Headers["x-death"]
	if !ok {
		return nil
	}
	deathList, ok := deaths.([]interface{})
	if !ok {
		return nil
	}

	entries := make([]deathEntry, 0, len(deathList))
	for _, d := range deathList {
		table, ok := d.(amqp.Table)
		if !ok {
			continue
		}
		entry := deathEntry{}
		if q, ok := table["queue"].(string); ok {
			entry.Queue = q
		}
		if r, ok := table["reason"].(string); ok {
			entry.Reason = r
		}
		if c, ok := table["count"].(int64); ok {
			entry.Count = c
		}
		if e, ok := table["exchange"].(string); ok {
			entry.Exchange = e
		}
		entries = append(entries, entry)
	}

	return entries
}
