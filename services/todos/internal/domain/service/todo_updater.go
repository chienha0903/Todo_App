package service

import (
	"context"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

type todoUpdater struct {
	cmdGW gateway.TodoCommandGateway
	qryGW gateway.TodoQueryGateway
}

func NewTodoUpdater(
	cmdGW gateway.TodoCommandGateway,
	qryGW gateway.TodoQueryGateway,
) todousecase.TodoUpdater {
	return &todoUpdater{cmdGW: cmdGW, qryGW: qryGW}
}

func (s *todoUpdater) Update(
	ctx context.Context,
	in *input.UpdateTodoInput,
) (*output.TodoUpdater, error) {
	todo, err := s.qryGW.GetTodo(ctx, entity.TodoID(in.ID))
	if err != nil {
		return nil, err
	}

	if err := applyTodoUpdates(todo, in); err != nil {
		return nil, err
	}

	todo.UpdatedAt = time.Now()

	if err := s.cmdGW.UpdateTodo(ctx, todo); err != nil {
		return nil, err
	}

	out := toOutput(todo)
	return &out, nil
}

func applyTodoUpdates(todo *entity.Todo, in *input.UpdateTodoInput) error {
	var err error

	if in.Title != "" {
		todo.Title, err = vo.NewTodoTitle(in.Title)
		if err != nil {
			return err
		}
	}

	if in.Description != "" {
		todo.Description, err = vo.NewTodoDescription(in.Description)
		if err != nil {
			return err
		}
	}

	if in.Priority != "" {
		todo.Priority, err = vo.NewTodoPriority(in.Priority)
		if err != nil {
			return err
		}
	}

	if in.Status != "" {
		todo.Status, err = vo.NewTodoStatus(in.Status)
		if err != nil {
			return err
		}
	}

	if in.DueDate != nil {
		dueDate, err := newOptionalTodoDueDate(in.DueDate)
		if err != nil {
			return err
		}
		todo.DueDate = dueDate
	}

	return nil
}
