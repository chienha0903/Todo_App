package resolver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/middleware"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	todooutput "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
	useroutput "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/output"
)

type Resolver struct {
	creater     todousecase.TodoCreater
	getter      todousecase.TodoGetter
	lister      todousecase.TodoLister
	updater     todousecase.TodoUpdater
	deleter     todousecase.TodoDeleter
	userGateway gateway.UserGateway
	timeout     time.Duration
}

func NewResolver(
	cfg *config.Config,
	creater todousecase.TodoCreater,
	getter todousecase.TodoGetter,
	lister todousecase.TodoLister,
	updater todousecase.TodoUpdater,
	deleter todousecase.TodoDeleter,
	userGateway gateway.UserGateway,
) *Resolver {
	return &Resolver{
		creater:     creater,
		getter:      getter,
		lister:      lister,
		updater:     updater,
		deleter:     deleter,
		userGateway: userGateway,
		timeout:     cfg.RequestTimeout,
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

func toModel(t *todooutput.Todo) *model.Todo {
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

func toModels(todos []*todooutput.Todo) []*model.Todo {
	items := make([]*model.Todo, 0, len(todos))

	for _, t := range todos {
		items = append(items, toModel(t))
	}

	return items
}

func toPageModel(p *todooutput.TodoPage) *model.TodoPage {
	return &model.TodoPage{
		Items:    toModels(p.Items),
		Total:    p.Total,
		Page:     p.Page,
		PageSize: p.PageSize,
		HasNext:  p.HasNext,
	}
}

func toUserModel(u *useroutput.User) *model.User {
	if u == nil {
		return nil
	}

	return &model.User{
		ID:        fmt.Sprintf("%d", u.ID),
		Email:     u.Email,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func requireAdmin(ctx context.Context) error {
	if err := middleware.RequireAuth(ctx); err != nil {
		return apperror.Unauthorized()
	}

	role, _ := middleware.GetRole(ctx)
	if role != "ADMIN" {
		return apperror.PermissionDenied()
	}

	return nil
}

// populateUsers fills the User field on each todo using the per-request DataLoader.
//
// Two-phase pattern (idiomatic graph-gophers/dataloader/v7):
//  1. Call Load() for every todo — non-blocking, registers all IDs into the same
//     batch window without triggering any RPC yet.
//  2. Execute each thunk — the first call dispatches the batch; the rest read from
//     the already-completed result. Duplicate IDs share the same thunk automatically.
func populateUsers(ctx context.Context, todos []*model.Todo) {
	loaders := middleware.GetLoaders(ctx)
	if loaders == nil || len(todos) == 0 {
		return
	}

	// Phase 1: register all IDs — all end up in the same 2ms batch window.
	thunks := make([]func() (*useroutput.User, error), len(todos))
	for i, t := range todos {
		thunks[i] = loaders.UserByID.Load(ctx, int64(t.UserID))
	}

	// Phase 2: resolve — batch RPC fires on the first thunk() call.
	for i, thunk := range thunks {
		u, err := thunk()
		if err == nil && u != nil {
			todos[i].User = toUserModel(u)
		}
	}
}

func toUserPageModel(p *useroutput.UserPage) *model.UserPage {
	items := make([]*model.User, 0, len(p.Items))

	for _, u := range p.Items {
		items = append(items, toUserModel(u))
	}

	return &model.UserPage{
		Items:    items,
		Total:    int(p.Total),
		Page:     int(p.Page),
		PageSize: int(p.PageSize),
	}
}
