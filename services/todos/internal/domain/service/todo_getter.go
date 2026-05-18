package service

import (
	"context"
	"fmt"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

var _ todousecase.TodoGetter = (*TodoGetter)(nil)

type TodoGetter struct {
	qryGW gateway.TodoQueryGateway
}

func NewTodoGetter(qryGW gateway.TodoQueryGateway) *TodoGetter {
	return &TodoGetter{qryGW: qryGW}
}

func (s *TodoGetter) Get(ctx context.Context, in *input.GetTodoInput) (*output.TodoGetter, error) {
	todo, err := s.qryGW.GetTodo(ctx, entity.TodoID(in.ID))
	if err != nil {
		return nil, fmt.Errorf("TodoGetter.Get: %w", err)
	}

	if todo == nil {
		return nil, pkgerrors.NewNotFound("todo not found")
	}

	out := toOutput(todo)
	return &out, nil
}
