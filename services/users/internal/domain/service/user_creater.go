package service

import (
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
	"golang.org/x/crypto/bcrypt"
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

func hashPassword(p vo.Password) (vo.Password, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Value()), bcrypt.DefaultCost)
	if err != nil {
		return vo.Password{}, pkgerrors.NewInternal("failed to hash password")
	}
	return vo.NewPassword(string(hash))
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

	rawPassword, err := vo.NewPassword(in.Password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	role, err := vo.NewUserRole(in.Role)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
