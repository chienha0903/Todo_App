package service

import (
	"context"
	stderrors "errors"
	"fmt"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
)

var _ todousecase.TodoDeleter = (*TodoDeleter)(nil)

type TodoDeleter struct {
	cmdGW gateway.TodoCommandGateway
}

func NewTodoDeleter(cmdGW gateway.TodoCommandGateway) *TodoDeleter {
	return &TodoDeleter{cmdGW: cmdGW}
}

func (s *TodoDeleter) Delete(ctx context.Context, in *input.DeleteTodoInput) error {
	err := s.cmdGW.DeleteTodo(ctx, entity.TodoID(in.ID))
	if err != nil {
		if stderrors.Is(err, pkgerrors.ErrRecordNotFound) {
			return pkgerrors.NewNotFound("todo not found")
		}
		return fmt.Errorf("TodoDeleter.Delete: %w", err)
	}
	return nil
}
