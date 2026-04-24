package service

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

type TodoLister interface {
	List(ctx context.Context, in *input.ListTodosInput) (output.TodoLister, error)
}

type todoLister struct {
	qryGW gateway.TodoQueryGateway
}

func NewTodoLister(qryGW gateway.TodoQueryGateway) TodoLister {
	return &todoLister{qryGW: qryGW}
}

func (s *todoLister) List(ctx context.Context, in *input.ListTodosInput) (output.TodoLister, error) {
	todos, err := s.qryGW.GetTodos(ctx, entity.UserID(in.UserID))
	if err != nil {
		return nil, err
	}
	out := make(output.TodoLister, 0, len(todos))
	for _, t := range todos {
		out = append(out, toOutput(t))
	}
	return out, nil
}
