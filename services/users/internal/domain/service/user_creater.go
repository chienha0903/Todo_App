package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserCreater = (*UserCreater)(nil)

type UserCreater struct {
	cmdGW gateway.UserCommandGateway
}

func NewUserCreater(cmdGW gateway.UserCommandGateway) *UserCreater {
	return &UserCreater{cmdGW: cmdGW}
}

func (s *UserCreater) Create(ctx context.Context, in *input.CreateUserInput) (*output.UserCreaterOutput, error) {
	user, err := newUserFromCreateInput(in, time.Now())
	if err != nil {
		return nil, err
	}

	if err := s.cmdGW.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("UserCreater.Create: %w", err)
	}

	out := toOutput(user)
	return &out, nil
}

func newUserFromCreateInput(in *input.CreateUserInput, now time.Time) (*entity.User, error) {
	username, err := vo.NewUsername(in.Username)
	if err != nil {
		return nil, err
	}

	email, err := vo.NewEmail(in.Email)
	if err != nil {
		return nil, err
	}

	passwordHash, err := vo.NewPasswordHash(in.PasswordHash)
	if err != nil {
		return nil, err
	}

	role, err := vo.NewUserRole(in.Role)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
