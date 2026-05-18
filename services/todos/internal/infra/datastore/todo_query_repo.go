package datastore

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/mapper"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
	"gorm.io/gorm"
)

type todoQueryRepo struct {
	db *gorm.DB
}

func NewTodoQueryRepo(db *gorm.DB) *todoQueryRepo {
	return &todoQueryRepo{db: db}
}

func (r *todoQueryRepo) GetTodo(ctx context.Context, id entity.TodoID) (*entity.Todo, error) {
	var m model.Todo
	
	result := r.db.WithContext(ctx).First(&m, int64(id))
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // service layer sẽ tạo NewNotFound
		}
		return nil, fmt.Errorf("db get todo: %w", result.Error)
	}

	t, err := mapper.ToEntity(&m)
	if err != nil {
		return nil, fmt.Errorf("db get todo mapper: %w", err)
	}

	return t, nil
}

func (r *todoQueryRepo) GetTodos(ctx context.Context, userID entity.UserID, page, pageSize int32) ([]*entity.Todo, int64, error) {
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Todo{}).
		Where("user_id = ?", int64(userID)).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("db count todos: %w", err)
	}

	offset := int((page - 1) * pageSize)
	var ms []model.Todo
	result := r.db.WithContext(ctx).
		Where("user_id = ?", int64(userID)).
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(offset).
		Find(&ms)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("db list todos: %w", result.Error)
	}

	todos := make([]*entity.Todo, 0, len(ms))
	for i := range ms {
		t, err := mapper.ToEntity(&ms[i])
		if err != nil {
			return nil, 0, fmt.Errorf("db list todos mapper: %w", err)
		}
		todos = append(todos, t)
	}
	
	return todos, total, nil
}
