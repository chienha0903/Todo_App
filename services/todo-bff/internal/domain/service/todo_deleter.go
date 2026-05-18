package service

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
)

var _ todousecase.TodoDeleter = (*TodoDeleter)(nil)

type TodoDeleter struct {
	gw gateway.TodoGateway
}

func NewTodoDeleter(gw gateway.TodoGateway) *TodoDeleter {
	return &TodoDeleter{gw: gw}
}

func (s *TodoDeleter) Delete(ctx context.Context, in *input.DeleteTodo) error {
	if in.ID <= 0 {
		return apperror.InvalidArgument("id must be a positive integer")
	}

	if err := s.gw.DeleteTodo(ctx, *in); err != nil {
		return fmt.Errorf("TodoDeleter.Delete: %w", err)
	}

	return nil
}
