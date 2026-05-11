package service

import (
	"context"
	"strings"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

var _ todousecase.TodoUpdater = (*TodoUpdater)(nil)

type TodoUpdater struct {
	gw gateway.TodoGateway
}

func NewTodoUpdater(gw gateway.TodoGateway) *TodoUpdater {
	return &TodoUpdater{gw: gw}
}

func (s *TodoUpdater) Update(ctx context.Context, in *input.UpdateTodo) (*output.Todo, error) {
	if err := validateUpdate(in); err != nil {
		return nil, err
	}
	return s.gw.UpdateTodo(ctx, *in)
}

func validateUpdate(in *input.UpdateTodo) error {
	if in.ID <= 0 {
		return apperror.InvalidArgument("id must be a positive integer")
	}
	if !hasUpdateField(in) {
		return apperror.InvalidArgument("at least one field is required")
	}
	return validateOptionalRFC3339(in.DueDate)
}

func hasUpdateField(in *input.UpdateTodo) bool {
	return strings.TrimSpace(in.Title) != "" ||
		strings.TrimSpace(in.Description) != "" ||
		strings.TrimSpace(in.Priority) != "" ||
		strings.TrimSpace(in.Status) != "" ||
		strings.TrimSpace(in.DueDate) != ""
}
