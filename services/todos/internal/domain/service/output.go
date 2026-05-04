package service

import (
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

func toOutput(t *entity.Todo) output.Todo {
	out := output.Todo{
		ID:          int64(t.ID),
		UserID:      int64(t.UserID),
		Title:       t.Title.Value(),
		Description: t.Description.Value(),
		Status:      t.Status.String(),
		Priority:    t.Priority.String(),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != nil {
		v := t.DueDate.Value()
		out.DueDate = &v
	}
	return out
}

func toOutputs(todos []*entity.Todo) output.TodoLister {
	out := make(output.TodoLister, 0, len(todos))
	for _, todo := range todos {
		out = append(out, toOutput(todo))
	}
	return out
}
