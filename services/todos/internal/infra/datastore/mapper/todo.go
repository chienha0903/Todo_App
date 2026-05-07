package mapper

import (
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
)

func ToModel(t *entity.Todo) *model.Todo {
	var dueDate *time.Time
	if t.DueDate != nil {
		v := t.DueDate.Value()
		dueDate = &v
	}
	return &model.Todo{
		ID:          int64(t.ID),
		UserID:      int64(t.UserID),
		Title:       t.Title.Value(),
		Description: t.Description.Value(),
		Status:      t.Status.String(),
		Priority:    t.Priority.String(),
		DueDate:     dueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToEntity(m *model.Todo) (*entity.Todo, error) {
	title, err := vo.NewTodoTitle(m.Title)
	if err != nil {
		return nil, err
	}

	description, err := vo.NewTodoDescription(m.Description)
	if err != nil {
		return nil, err
	}

	status, err := vo.NewTodoStatus(m.Status)
	if err != nil {
		return nil, err
	}

	priority, err := vo.NewTodoPriority(m.Priority)
	if err != nil {
		return nil, err
	}

	t := &entity.Todo{
		ID:          entity.TodoID(m.ID),
		UserID:      entity.UserID(m.UserID),
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if m.DueDate != nil {
		dueDate, err := vo.NewTodoDueDate(*m.DueDate)
		if err != nil {
			return nil, err
		}
		t.DueDate = &dueDate
	}

	return t, nil
}
