package datastore

import (
	"context"
	stderrors "errors"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
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
		return nil, mapTodoRepoError(result.Error)
	}
	return mapper.ToEntity(&m)
}

func (r *todoQueryRepo) GetTodos(ctx context.Context, userID entity.UserID) ([]*entity.Todo, error) {
	var ms []model.Todo
	result := r.db.WithContext(ctx).
		Where("user_id = ?", int64(userID)).
		Order("created_at DESC").
		Find(&ms)
	if result.Error != nil {
		return nil, result.Error
	}

	todos := make([]*entity.Todo, 0, len(ms))
	for i := range ms {
		t, err := mapper.ToEntity(&ms[i])
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func mapTodoRepoError(err error) error {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.NewAppError(apperrors.ReasonNotFound, "Todo not found")
	}
	return err
}
