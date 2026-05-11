package resolver

import (
	stderrors "errors"
	"fmt"
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

func toGraphQLError(err error) error {
	var appErr *apperror.Error
	if stderrors.As(err, &appErr) {
		switch appErr.Code {
		case apperror.CodeNotFound:
			return fmt.Errorf("not found: %s", appErr.Message)
		case apperror.CodeInvalidArgument:
			return fmt.Errorf("invalid argument: %s", appErr.Message)
		case apperror.CodeUnauthorized:
			return stderrors.New("unauthorized")
		case apperror.CodePermissionDenied:
			return stderrors.New("permission denied")
		case apperror.CodeTimeout:
			return stderrors.New("request timed out")
		case apperror.CodeUnavailable:
			return stderrors.New("service unavailable")
		}
	}
	return stderrors.New("internal server error")
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
		Status:      t.Status,
		Priority:    t.Priority,
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
