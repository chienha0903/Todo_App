package resolver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

type Resolver struct {
	creater todousecase.TodoCreater
	getter  todousecase.TodoGetter
	lister  todousecase.TodoLister
	updater todousecase.TodoUpdater
	deleter todousecase.TodoDeleter
	timeout time.Duration
}

func NewResolver(
	cfg *config.Config,
	creater todousecase.TodoCreater,
	getter todousecase.TodoGetter,
	lister todousecase.TodoLister,
	updater todousecase.TodoUpdater,
	deleter todousecase.TodoDeleter,
) *Resolver {
	return &Resolver{
		creater: creater,
		getter:  getter,
		lister:  lister,
		updater: updater,
		deleter: deleter,
		timeout: cfg.RequestTimeout,
	}
}

func parseID(id string) (int64, error) {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, apperror.InvalidArgument("id must be a positive integer")
	}
	return parsed, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefPriority(p *model.TodoPriority) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func derefStatus(s *model.TodoStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func toModel(t *output.Todo) *model.Todo {
	if t == nil {
		return nil
	}
	m := &model.Todo{
		ID:          fmt.Sprintf("%d", t.ID),
		UserID:      int(t.UserID),
		Title:       t.Title,
		Description: t.Description,
		Status:      model.TodoStatus(t.Status),
		Priority:    model.TodoPriority(t.Priority),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != "" {
		dd := t.DueDate
		m.DueDate = &dd
	}
	return m
}

func toModels(todos []*output.Todo) []*model.Todo {
	items := make([]*model.Todo, 0, len(todos))
	for _, t := range todos {
		items = append(items, toModel(t))
	}
	return items
}

func toPageModel(p *output.TodoPage) *model.TodoPage {
	return &model.TodoPage{
		Items:    toModels(p.Items),
		Total:    p.Total,
		Page:     p.Page,
		PageSize: p.PageSize,
		HasNext:  p.HasNext,
	}
}
