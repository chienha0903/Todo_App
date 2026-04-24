package service

import (
	"context"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

type TodoUpdater interface {
	Update(ctx context.Context, in *input.UpdateTodoInput) (*output.TodoUpdater, error)
}

type todoUpdater struct {
	cmdGW gateway.TodoCommandGateway
	qryGW gateway.TodoQueryGateway
}

func NewTodoUpdater(cmdGW gateway.TodoCommandGateway, qryGW gateway.TodoQueryGateway) TodoUpdater {
	return &todoUpdater{cmdGW: cmdGW, qryGW: qryGW}
}

func (s *todoUpdater) Update(ctx context.Context, in *input.UpdateTodoInput) (*output.TodoUpdater, error) {
	todo, err := s.qryGW.GetTodo(ctx, entity.TodoID(in.ID))
	if err != nil {
		return nil, err
	}

	if in.Title != "" {
		todo.Title, err = vo.NewTodoTitle(in.Title)
		if err != nil {
			return nil, err
		}
	}
	if in.Description != "" {
		todo.Description, err = vo.NewTodoDescription(in.Description)
		if err != nil {
			return nil, err
		}
	}
	if in.Priority != "" {
		todo.Priority, err = vo.NewTodoPriority(in.Priority)
		if err != nil {
			return nil, err
		}
	}
	if in.Status != "" {
		todo.Status, err = vo.NewTodoStatus(in.Status)
		if err != nil {
			return nil, err
		}
	}
	if in.DueDate != nil {
		dd, err := vo.NewTodoDueDate(*in.DueDate)
		if err != nil {
			return nil, err
		}
		todo.DueDate = &dd
	}

	todo.UpdatedAt = time.Now()

	if err := s.cmdGW.UpdateTodo(ctx, todo); err != nil {
		return nil, err
	}

	out := toOutput(todo)
	return &out, nil
}
