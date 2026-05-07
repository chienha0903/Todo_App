package datastore

import (
	"context"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/mapper"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
	"gorm.io/gorm"
)

type todoCommandRepo struct {
	db *gorm.DB
}

func NewTodoCommandRepo(db *gorm.DB) *todoCommandRepo {
	return &todoCommandRepo{db: db}
}

func (r *todoCommandRepo) CreateTodo(ctx context.Context, t *entity.Todo) error {
	m := mapper.ToModel(t)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		return result.Error
	}
	t.ID = entity.TodoID(m.ID)
	return nil
}

func (r *todoCommandRepo) UpdateTodo(ctx context.Context, t *entity.Todo) error {
	m := mapper.ToModel(t)
	result := r.db.WithContext(ctx).Save(m)
	if result.Error != nil {
		return result.Error
	}
	return ensureTodoAffected(result.RowsAffected)
}

func (r *todoCommandRepo) DeleteTodo(ctx context.Context, id entity.TodoID) error {
	result := r.db.WithContext(ctx).Delete(&model.Todo{}, int64(id))
	if result.Error != nil {
		return result.Error
	}
	return ensureTodoAffected(result.RowsAffected)
}

func ensureTodoAffected(rowsAffected int64) error {
	if rowsAffected == 0 {
		return apperrors.NewAppError(apperrors.ReasonNotFound, "Todo not found")
	}
	return nil
}
