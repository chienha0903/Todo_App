package mapper

import (
	"time"

	"github.com/chienha0903/Todo_App/pkg/errors"
	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

func ToCreateTodoInput(req *todopb.CreateTodoRequest) (*input.CreateTodoInput, error) {
	dueDate, err := parseOptionalDueDate(req.DueDate)
	if err != nil {
		return nil, err
	}

	in := &input.CreateTodoInput{
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     dueDate,
	}
	return in, nil
}

func ToGetTodoInput(req *todopb.GetTodoRequest) *input.GetTodoInput {
	return &input.GetTodoInput{ID: req.Id}
}

func ToListTodosInput(req *todopb.ListTodosRequest) *input.ListTodosInput {
	return &input.ListTodosInput{UserID: req.UserId}
}

func ToUpdateTodoInput(req *todopb.UpdateTodoRequest) (*input.UpdateTodoInput, error) {
	dueDate, err := parseOptionalDueDate(req.DueDate)
	if err != nil {
		return nil, err
	}

	in := &input.UpdateTodoInput{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      req.Status,
		DueDate:     dueDate,
	}
	return in, nil
}

func ToDeleteTodoInput(req *todopb.DeleteTodoRequest) *input.DeleteTodoInput {
	return &input.DeleteTodoInput{ID: req.Id}
}

func ToProtoTodo(t *output.Todo) *todopb.Todo {
	proto := &todopb.Todo{
		Id:          int64(t.ID),
		UserId:      int64(t.UserID),
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
	if t.DueDate != nil {
		proto.DueDate = t.DueDate.Format(time.RFC3339)
	}
	return proto
}

func ToProtoTodos(todos output.TodoLister) []*todopb.Todo {
	items := make([]*todopb.Todo, 0, len(todos))
	for i := range todos {
		t := todos[i]
		items = append(items, ToProtoTodo(&t))
	}
	return items
}

func parseOptionalDueDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	dueDate, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, errors.NewAppError(
			errors.ReasonInvalidParameter,
			"Due date must be RFC3339 format",
		)
	}

	return &dueDate, nil
}
