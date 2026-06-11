package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/event"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

type CacheInvalidator struct {
	outboxQuery gateway.OutboxQueryGateway
	outboxCmd   gateway.OutboxCommandGateway
	tokenCmdGW  gateway.RefreshTokenCommandGateway
}

func NewCacheInvalidator(
	outboxQuery gateway.OutboxQueryGateway,
	outboxCmd gateway.OutboxCommandGateway,
	tokenCmdGW gateway.RefreshTokenCommandGateway,
) *CacheInvalidator {
	return &CacheInvalidator{
		outboxQuery: outboxQuery,
		outboxCmd:   outboxCmd,
		tokenCmdGW:  tokenCmdGW,
	}
}

func (c *CacheInvalidator) ProcessBatch(ctx context.Context) error {
	events, err := c.outboxQuery.GetUnpublishedEventsByAggregateType(ctx, event.CacheAggregateType, 10)
	if err != nil {
		return fmt.Errorf("get cache events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	slog.Info("processing cache invalidation events", "count", len(events))

	for _, ev := range events {
		if err := c.process(ctx, ev); err != nil {
			slog.Error("cache invalidation failed", "event_id", ev.EventID, "type", ev.EventType, "error", err)
			_ = c.outboxCmd.MarkAsFailed(ctx, ev.ID, err.Error())
			continue
		}

		if err := c.outboxCmd.MarkAsPublished(ctx, ev.ID); err != nil {
			slog.Error("mark cache event processed failed", "event_id", ev.EventID, "error", err)
		} else {
			slog.Info("cache event processed", "event_id", ev.EventID, "type", ev.EventType)
		}
	}

	return nil
}

func (c *CacheInvalidator) process(ctx context.Context, ev *model.OutboxEvent) error {
	switch ev.EventType {
	case event.CacheEventDeleteUserTokens:
		var p event.DeleteUserTokensPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
		return c.tokenCmdGW.DeleteByUserID(ctx, p.UserID)
	default:
		return fmt.Errorf("unknown cache event type: %s", ev.EventType)
	}
}
