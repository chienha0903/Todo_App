package usecase

import (
	"context"

	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

type UserCreater interface {
	Create(ctx context.Context, in *input.CreateUserInput) (*output.UserCreaterOutput, error)
}

type UserGetter interface {
	Get(ctx context.Context, in *input.GetUserInput) (*output.UserGetterOutput, error)
}

type UserUpdater interface {
	Update(ctx context.Context, in *input.UpdateUserInput) (*output.UserUpdaterOutput, error)
}

type UserDeleter interface {
	Delete(ctx context.Context, in *input.DeleteUserInput) (*output.UserDeleterOutput, error)
}

type UserLister interface {
	List(ctx context.Context, in *input.ListUsersInput) (*output.UserPage, error)
}

type UserAuthenticator interface {
	Login(ctx context.Context, in *input.UserLoginInput) (*output.UserLoginOutput, error)
}

type UserRefresher interface {
	Refresh(ctx context.Context, in *input.UserRefreshTokenInput) (*output.UserRefreshTokenOutput, error)
}
