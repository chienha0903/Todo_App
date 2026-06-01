package gateway

import (
	"context"

	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

type OutboxCommandGateway interface {
	InsertOutboxEvent(ctx context.Context, event *model.OutboxEvent) error
	MarkAsPublished(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errMsg string) error
}

type OutboxQueryGateway interface {
	GetUnpublishedEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error)
}
