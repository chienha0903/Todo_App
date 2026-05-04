package todo

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
)

type TodoCreator interface {
	Create(ctx context.Context, in *input.CreateTodoInput) (*output.TodoCreater, error)
}

type TodoGetter interface {
	Get(ctx context.Context, in *input.GetTodoInput) (*output.TodoGetter, error)
}

type TodoLister interface {
	List(ctx context.Context, in *input.ListTodosInput) (output.TodoLister, error)
}

type TodoUpdater interface {
	Update(ctx context.Context, in *input.UpdateTodoInput) (*output.TodoUpdater, error)
}

type TodoDeleter interface {
	Delete(ctx context.Context, in *input.DeleteTodoInput) error
}
