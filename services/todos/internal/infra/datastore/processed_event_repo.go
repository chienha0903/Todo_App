package datastore

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
)

type processedEventRepo struct {
	db *gorm.DB
}

func NewProcessedEventRepo(db *gorm.DB) *processedEventRepo {
	return &processedEventRepo{db: db}
}

func NewProcessedEventGateway(repo *processedEventRepo) gateway.ProcessedEventGateway {
	return repo
}

func (r *processedEventRepo) Exists(ctx context.Context, eventID string) (bool, error) {
	var m model.ProcessedEvent
	
	result := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&m)

	if result.Error == nil {
		return true, nil
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, fmt.Errorf("db check processed event: %w", result.Error)
}

func (r *processedEventRepo) Insert(ctx context.Context, eventID string) error {
	m := &model.ProcessedEvent{EventID: eventID}

	result := extractDB(ctx, r.db).WithContext(ctx).Create(m)
	if result.Error != nil {
		return fmt.Errorf("db insert processed event: %w", result.Error)
	}

	return nil
}
