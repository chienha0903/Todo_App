package todo

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

type TodoCreater interface {
	Create(ctx context.Context, in *input.CreateTodo) (*output.Todo, error)
}

type TodoGetter interface {
	Get(ctx context.Context, in *input.GetTodo) (*output.Todo, error)
}

type TodoLister interface {
	List(ctx context.Context, in *input.ListTodos) ([]*output.Todo, error)
}

type TodoUpdater interface {
	Update(ctx context.Context, in *input.UpdateTodo) (*output.Todo, error)
}

type TodoDeleter interface {
	Delete(ctx context.Context, in *input.DeleteTodo) error
}
