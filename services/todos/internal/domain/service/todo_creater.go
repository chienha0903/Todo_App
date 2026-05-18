package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

var _ todousecase.TodoCreater = (*TodoCreater)(nil)

type TodoCreater struct {
	cmdGW gateway.TodoCommandGateway
}

func NewTodoCreater(cmdGW gateway.TodoCommandGateway) *TodoCreater {
	return &TodoCreater{cmdGW: cmdGW}
}

func (s *TodoCreater) Create(ctx context.Context, in *input.CreateTodoInput) (*output.TodoCreater, error) {
	todo, err := newTodoFromCreateInput(in, time.Now())
	if err != nil {
		return nil, err
	}

	if err := s.cmdGW.CreateTodo(ctx, todo); err != nil {
		return nil, fmt.Errorf("TodoCreater.Create: %w", err)
	}

	out := toOutput(todo)
	return &out, nil
}

func newTodoFromCreateInput(in *input.CreateTodoInput, now time.Time) (*entity.Todo, error) {
	title, err := vo.NewTodoTitle(in.Title)
	if err != nil {
		return nil, err
	}

	description, err := vo.NewTodoDescription(in.Description)
	if err != nil {
		return nil, err
	}

	priority, err := vo.NewTodoPriority(in.Priority)
	if err != nil {
		return nil, err
	}

	dueDate, err := newOptionalTodoDueDate(in.DueDate)
	if err != nil {
		return nil, err
	}

	return &entity.Todo{
		UserID:      entity.UserID(in.UserID),
		Title:       title,
		Description: description,
		Status:      vo.TodoStatusPending,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func newOptionalTodoDueDate(value *time.Time) (*vo.TodoDueDate, error) {
	if value == nil {
		return nil, nil
	}

	dueDate, err := vo.NewTodoDueDate(*value)
	if err != nil {
		return nil, err
	}
	return &dueDate, nil
}
