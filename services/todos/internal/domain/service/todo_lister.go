package service

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

var _ todousecase.TodoLister = (*TodoLister)(nil)

type TodoLister struct {
	qryGW gateway.TodoQueryGateway
}

func NewTodoLister(qryGW gateway.TodoQueryGateway) *TodoLister {
	return &TodoLister{qryGW: qryGW}
}

func (s *TodoLister) List(ctx context.Context, in *input.ListTodosInput) (output.TodoLister, error) {
	todos, err := s.qryGW.GetTodos(ctx, entity.UserID(in.UserID))
	if err != nil {
		return nil, err
	}
	return toOutputs(todos), nil
}
