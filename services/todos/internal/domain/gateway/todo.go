package gateway

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
)

//go:generate mockgen -source=todo.go -destination=mock/mock_todo.go -package=mock

type TodoCommandGateway interface {
	CreateTodo(ctx context.Context, todo *entity.Todo) error
	UpdateTodo(ctx context.Context, todo *entity.Todo) error
	DeleteTodo(ctx context.Context, id entity.TodoID) error
}

type TodoQueryGateway interface {
	GetTodo(ctx context.Context, id entity.TodoID) (*entity.Todo, error)
	GetTodos(ctx context.Context, userID entity.UserID) ([]*entity.Todo, error)
}
