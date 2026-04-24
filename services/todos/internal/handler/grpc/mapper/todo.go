package mapper

import (
	"time"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

func ToCreateTodoInput(req *todopb.CreateTodoRequest) *input.CreateTodoInput {
	in := &input.CreateTodoInput{
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}
	if req.DueDate != "" {
		if t, err := time.Parse(time.RFC3339, req.DueDate); err == nil {
			in.DueDate = &t
		}
	}
	return in
}

func ToGetTodoInput(req *todopb.GetTodoRequest) *input.GetTodoInput {
	return &input.GetTodoInput{ID: req.Id}
}

func ToListTodosInput(req *todopb.ListTodosRequest) *input.ListTodosInput {
	return &input.ListTodosInput{UserID: req.UserId}
}

func ToUpdateTodoInput(req *todopb.UpdateTodoRequest) *input.UpdateTodoInput {
	in := &input.UpdateTodoInput{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      req.Status,
	}
	if req.DueDate != "" {
		if t, err := time.Parse(time.RFC3339, req.DueDate); err == nil {
			in.DueDate = &t
		}
	}
	return in
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
