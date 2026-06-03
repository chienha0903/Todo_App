package gateway

import (
	"context"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
)

//go:generate mockgen -source=user.go -destination=mock/mock_user.go -package=mock

type UserCommandGateway interface {
	CreateUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id entity.UserID) error
}

type UserQueryGateway interface {
	GetUser(ctx context.Context, id entity.UserID) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUsers(ctx context.Context, page, pageSize int32) ([]*entity.User, int64, error)
	GetUsersByIDs(ctx context.Context, ids []entity.UserID) ([]*entity.User, error)
}

type TransactionGateway interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
