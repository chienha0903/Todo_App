package service

import (
	"context"
	stderrors "errors"
	"fmt"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserDeleter = (*UserDeleter)(nil)

type UserDeleter struct {
	cmdGW gateway.UserCommandGateway
}

func NewUserDeleter(cmdGW gateway.UserCommandGateway) *UserDeleter {
	return &UserDeleter{cmdGW: cmdGW}
}

func (s *UserDeleter) Delete(ctx context.Context, in *input.DeleteUserInput) (*output.UserDeleterOutput, error) {
	err := s.cmdGW.DeleteUser(ctx, entity.UserID(in.ID))
	if err != nil {
		if stderrors.Is(err, pkgerrors.ErrRecordNotFound) {
			return nil, pkgerrors.NewNotFound("user not found")
		}
		return nil, fmt.Errorf("UserDeleter.Delete: %w", err)
	}
	return &output.UserDeleterOutput{ID: in.ID}, nil
}
