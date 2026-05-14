package service

import (
	"context"
	"fmt"

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

func (s *TodoLister) List(ctx context.Context, in *input.ListTodosInput) (*output.TodoPage, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	todos, total, err := s.qryGW.GetTodos(ctx, entity.UserID(in.UserID), page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("TodoLister.List: %w", err)
	}
	return &output.TodoPage{
		Items:    toOutputSlice(todos),
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
