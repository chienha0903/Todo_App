package mapper

import (
	"fmt"
	"time"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
)

// ToModel chuyển proto Todo → GraphQL model Todo.
// Đây là chiều duy nhất BFF cần: proto response → client response.
func ToModel(t *todopb.Todo) *model.Todo {
	if t == nil {
		return nil
	}

	m := &model.Todo{
		ID:          fmt.Sprintf("%d", t.Id),
		UserID:      int(t.UserId),
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	if t.DueDate != "" {
		dueDate := t.DueDate
		m.DueDate = &dueDate
	}

	return m
}

// ToModels chuyển slice proto Todo → slice GraphQL model Todo.
func ToModels(todos []*todopb.Todo) []*model.Todo {
	items := make([]*model.Todo, 0, len(todos))
	for _, t := range todos {
		items = append(items, ToModel(t))
	}
	return items
}

// ParseOptionalDueDate validate và parse RFC3339 string nếu có.
func ParseOptionalDueDate(value *string) (string, error) {
	if value == nil || *value == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, *value); err != nil {
		return "", fmt.Errorf("dueDate must be RFC3339 format")
	}
	return *value, nil
}
