package user

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/output"
)

type UserGetter interface {
	Get(ctx context.Context, in *input.GetUser) (*output.User, error)
}

type UserLister interface {
	List(ctx context.Context, in *input.ListUsers) (*output.UserPage, error)
}

type UserUpdater interface {
	Update(ctx context.Context, in *input.UpdateUser) (*output.User, error)
}

type UserDeleter interface {
	Delete(ctx context.Context, in *input.DeleteUser) error
}
