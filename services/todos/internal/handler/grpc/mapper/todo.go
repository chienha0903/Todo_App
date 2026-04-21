package mapper

import (
	"time"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

// ToCreateTodoInput converts a gRPC request to usecase input.
func ToCreateTodoInput(req *todopb.CreateTodoRequest) input.CreateTodoInput {
	in := input.CreateTodoInput{
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

// ToProtoTodo converts usecase output to a gRPC Todo message.
func ToProtoTodo(out *output.TodoOutput) *todopb.Todo {
	proto := &todopb.Todo{
		Id:          out.ID,
		UserId:      out.UserID,
		Title:       out.Title,
		Description: out.Description,
		Status:      out.Status,
		Priority:    out.Priority,
		CreatedAt:   out.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   out.UpdatedAt.Format(time.RFC3339),
	}

	if out.DueDate != nil {
		proto.DueDate = out.DueDate.Format(time.RFC3339)
	}

	return proto
}
