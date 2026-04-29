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

type TodoCreator interface {
	Create(ctx context.Context, in *input.CreateTodoInput) (*output.TodoCreator, error)
}

type todoCreator struct {
	cmdGW gateway.TodoCommandGateway
}

func NewTodoCreator(cmdGW gateway.TodoCommandGateway) TodoCreator {
	return &todoCreator{cmdGW: cmdGW}
}

func (s *todoCreator) Create(ctx context.Context, in *input.CreateTodoInput) (*output.TodoCreator, error) {
	titleVO, err := vo.NewTodoTitle(in.Title)
	if err != nil {
		return nil, err
	}

	descVO, err := vo.NewTodoDescription(in.Description)
	if err != nil {
		return nil, err
	}

	priorityVO, err := vo.NewTodoPriority(in.Priority)
	if err != nil {
		return nil, err
	}

	var dueDateVO *vo.TodoDueDate
	if in.DueDate != nil {
		dd, err := vo.NewTodoDueDate(*in.DueDate)
		if err != nil {
			return nil, err
		}
		dueDateVO = &dd
	}

	now := time.Now()
	todo := &entity.Todo{
		UserID:      entity.UserID(in.UserID),
		Title:       titleVO,
		Description: descVO,
		Status:      vo.TODO_STATUS_PENDING,
		Priority:    priorityVO,
		DueDate:     dueDateVO,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.cmdGW.CreateTodo(ctx, todo); err != nil {
		return nil, err
	}

	out := toOutput(todo)
	return &out, nil
}
