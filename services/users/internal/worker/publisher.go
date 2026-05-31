package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/chienha0903/Todo_App/pkg/rabbitmq"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
)

type OutboxPublisher struct {
	outboxQuery gateway.OutboxQueryGateway
	amqpConn    *amqp.Connection
}

func NewOutboxPublisher(q gateway.OutboxQueryGateway, conn *amqp.Connection) *OutboxPublisher {
	return &OutboxPublisher{outboxQuery: q, amqpConn: conn}
}

type publishedMessage struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"` 
}

func (p *OutboxPublisher) PublishBatch(ctx context.Context) error {
	ch, err := p.amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := rabbitmq.SetupTopology(ch); err != nil {
		return fmt.Errorf("setup topology: %w", err)
	}

	events, err := p.outboxQuery.GetUnpublishedEvents(ctx, 10)
	if err != nil {
		return fmt.Errorf("get unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	slog.Info("publishing outbox events", "count", len(events))

	for _, ev := range events {
		msg := publishedMessage{
			EventID:   ev.EventID,
			EventType: ev.EventType,
			Payload:   ev.Payload,
		}
		body, err := json.Marshal(msg)
		if err != nil {
			slog.Error("marshal event failed", "event_id", ev.EventID, "error", err)
			_ = p.outboxQuery.MarkAsFailed(ctx, ev.ID, err.Error())
			continue
		}

		err = ch.PublishWithContext(ctx,
			rabbitmq.ExchangeName,
			ev.EventType, 
			false,        
			false,        
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent, 
				MessageId:    ev.EventID,       
				Body:         body,
			},
		)
		if err != nil {
			slog.Error("publish event failed", "event_id", ev.EventID, "error", err)
			_ = p.outboxQuery.MarkAsFailed(ctx, ev.ID, err.Error())
			continue
		}

		if err := p.outboxQuery.MarkAsPublished(ctx, ev.ID); err != nil {
			slog.Error("mark published failed", "event_id", ev.EventID, "error", err)
		} else {
			slog.Info("event published", "event_id", ev.EventID, "type", ev.EventType)
		}
	}

	return nil
}
