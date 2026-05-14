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

var _ todousecase.TodoLister = (*TodoLister)(nil)

type TodoLister struct {
	gw gateway.TodoGateway
}

func NewTodoLister(gw gateway.TodoGateway) *TodoLister {
	return &TodoLister{gw: gw}
}

func (s *TodoLister) List(ctx context.Context, in *input.ListTodos) (*output.TodoPage, error) {
	if in.UserID <= 0 {
		return nil, apperror.InvalidArgument("userId must be a positive integer")
	}
	res, err := s.gw.ListTodos(ctx, *in)
	if err != nil {
		return nil, fmt.Errorf("TodoLister.List: %w", err)
	}
	return res, nil
}
