package service

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

var _ todousecase.TodoGetter = (*TodoGetter)(nil)

type TodoGetter struct {
	gw gateway.TodoGateway
}

func NewTodoGetter(gw gateway.TodoGateway) *TodoGetter {
	return &TodoGetter{gw: gw}
}

func (s *TodoGetter) Get(ctx context.Context, in *input.GetTodo) (*output.Todo, error) {
	if in.ID <= 0 {
		return nil, apperror.InvalidArgument("id must be a positive integer")
	}
	res, err := s.gw.GetTodo(ctx, *in)
	if err != nil {
		return nil, fmt.Errorf("TodoGetter.Get: %w", err)
	}
	return res, nil
}
