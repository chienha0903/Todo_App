package service

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

type TodoGetter interface {
	Get(ctx context.Context, in *input.GetTodoInput) (*output.TodoGetter, error)
}

type todoGetter struct {
	qryGW gateway.TodoQueryGateway
}

func NewTodoGetter(qryGW gateway.TodoQueryGateway) TodoGetter {
	return &todoGetter{qryGW: qryGW}
}

func (s *todoGetter) Get(ctx context.Context, in *input.GetTodoInput) (*output.TodoGetter, error) {
	todo, err := s.qryGW.GetTodo(ctx, entity.TodoID(in.ID))
	if err != nil {
		return nil, err
	}
	out := toOutput(todo)
	return &out, nil
}
