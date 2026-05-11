package service

import (
	"context"
	"strings"
	"time"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

var _ todousecase.TodoCreater = (*TodoCreater)(nil)

type TodoCreater struct {
	gw gateway.TodoGateway
}

func NewTodoCreater(gw gateway.TodoGateway) *TodoCreater {
	return &TodoCreater{gw: gw}
}

func (s *TodoCreater) Create(ctx context.Context, in *input.CreateTodo) (*output.Todo, error) {
	if err := validateCreate(in); err != nil {
		return nil, err
	}
	return s.gw.CreateTodo(ctx, *in)
}

func validateCreate(in *input.CreateTodo) error {
	if in.UserID <= 0 {
		return apperror.InvalidArgument("user_id must be a positive integer")
	}
	if strings.TrimSpace(in.Title) == "" {
		return apperror.InvalidArgument("title is required")
	}
	if strings.TrimSpace(in.Description) == "" {
		return apperror.InvalidArgument("description is required")
	}
	if strings.TrimSpace(in.Priority) == "" {
		return apperror.InvalidArgument("priority is required")
	}
	return validateOptionalRFC3339(in.DueDate)
}

func validateOptionalRFC3339(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return apperror.InvalidArgument("due_date must be RFC3339 format")
	}
	return nil
}
