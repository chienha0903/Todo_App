package resolver

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/middleware"
	ucin "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
)

func (r *mutationResolver) CreateTodo(ctx context.Context, input model.CreateTodoInput) (*model.Todo, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	todo, err := r.creater.Create(ctx, &ucin.CreateTodo{
		UserID:      int64(input.UserID),
		Title:       input.Title,
		Description: input.Description,
		Priority:    string(input.Priority),
		DueDate:     derefStr(input.DueDate),
	})
	if err != nil {
		return nil, err
	}
	return toModel(todo), nil
}

func (r *mutationResolver) UpdateTodo(ctx context.Context, id string, input model.UpdateTodoInput) (*model.Todo, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	todoID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	todo, err := r.updater.Update(ctx, &ucin.UpdateTodo{
		ID:          todoID,
		Title:       derefStr(input.Title),
		Description: derefStr(input.Description),
		Priority:    derefPriority(input.Priority),
		Status:      derefStatus(input.Status),
		DueDate:     derefStr(input.DueDate),
	})
	if err != nil {
		return nil, err
	}
	return toModel(todo), nil
}

func (r *mutationResolver) DeleteTodo(ctx context.Context, id string) (bool, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return false, err
	}

	todoID, err := parseID(id)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if err := r.deleter.Delete(ctx, &ucin.DeleteTodo{ID: todoID}); err != nil {
		return false, err
	}
	return true, nil
}

func (r *queryResolver) Todo(ctx context.Context, id string) (*model.Todo, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	todoID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	todo, err := r.getter.Get(ctx, &ucin.GetTodo{ID: todoID})
	if err != nil {
		return nil, err
	}
	return toModel(todo), nil
}

func (r *queryResolver) Todos(ctx context.Context, userID int, page *int, pageSize *int) (*model.TodoPage, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	p, ps := 1, 20
	if page != nil && *page > 0 {
		p = *page
	}
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	result, err := r.lister.List(ctx, &ucin.ListTodos{
		UserID:   int64(userID),
		Page:     p,
		PageSize: ps,
	})
	if err != nil {
		return nil, err
	}
	return toPageModel(result), nil
}
