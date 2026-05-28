package gateway

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/output"
)

type UserGateway interface {
	Login(ctx context.Context, email, password string) (string, error)
	GetUser(ctx context.Context, in *input.GetUser) (*output.User, error)
	ListUsers(ctx context.Context, in *input.ListUsers) (*output.UserPage, error)
	UpdateUser(ctx context.Context, in *input.UpdateUser) (*output.User, error)
	DeleteUser(ctx context.Context, in *input.DeleteUser) error
}
