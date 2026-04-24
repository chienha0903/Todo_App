package service

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
)

type TodoDeleter interface {
	Delete(ctx context.Context, in *input.DeleteTodoInput) error
}

type todoDeleter struct {
	cmdGW gateway.TodoCommandGateway
}

func NewTodoDeleter(cmdGW gateway.TodoCommandGateway) TodoDeleter {
	return &todoDeleter{cmdGW: cmdGW}
}

func (s *todoDeleter) Delete(ctx context.Context, in *input.DeleteTodoInput) error {
	return s.cmdGW.DeleteTodo(ctx, entity.TodoID(in.ID))
}
