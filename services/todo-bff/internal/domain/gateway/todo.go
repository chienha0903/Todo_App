package gateway

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
)

//go:generate mockgen -source=todo.go -destination=mock/mock_todo.go -package=mock

type TodoGateway interface {
	CreateTodo(ctx context.Context, in input.CreateTodo) (*output.Todo, error)
	GetTodo(ctx context.Context, in input.GetTodo) (*output.Todo, error)
	ListTodos(ctx context.Context, in input.ListTodos) (*output.TodoPage, error)
	UpdateTodo(ctx context.Context, in input.UpdateTodo) (*output.Todo, error)
	DeleteTodo(ctx context.Context, in input.DeleteTodo) error
}
