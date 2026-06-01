package datastore

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

type outboxRepo struct {
	db *gorm.DB
}

func NewOutboxRepo(db *gorm.DB) *outboxRepo {
	return &outboxRepo{db: db}
}

func NewOutboxCommandGateway(repo *outboxRepo) gateway.OutboxCommandGateway {
	return repo
}

func NewOutboxQueryGateway(repo *outboxRepo) gateway.OutboxQueryGateway {
	return repo
}

func (r *outboxRepo) InsertOutboxEvent(ctx context.Context, event *model.OutboxEvent) error {
	result := extractDB(ctx, r.db).WithContext(ctx).Create(event)

	if result.Error != nil {
		return fmt.Errorf("db insert outbox event: %w", result.Error)
	}

	return nil
}

func (r *outboxRepo) GetUnpublishedEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error) {
	var events []*model.OutboxEvent

	err := r.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("db get unpublished events: %w", err)
	}

	return events, nil
}

func (r *outboxRepo) MarkAsPublished(ctx context.Context, id int64) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Update("published_at", now)
	if result.Error != nil {
		return fmt.Errorf("db mark outbox published: %w", result.Error)
	}

	return nil
}

func (r *outboxRepo) MarkAsFailed(ctx context.Context, id int64, errMsg string) error {
	result := r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMsg,
		})

	if result.Error != nil {
		return fmt.Errorf("db mark outbox failed: %w", result.Error)
	}

	return nil
}
