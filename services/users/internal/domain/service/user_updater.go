package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserUpdater = (*UserUpdater)(nil)

type UserUpdater struct {
	cmdGW      gateway.UserCommandGateway
	qryGW      gateway.UserQueryGateway
	transactor gateway.TransactionGateway
}

func NewUserUpdater(
	cmdGW gateway.UserCommandGateway,
	qryGW gateway.UserQueryGateway,
	transactor gateway.TransactionGateway,
) *UserUpdater {
	return &UserUpdater{cmdGW: cmdGW, qryGW: qryGW, transactor: transactor}
}

func (s *UserUpdater) Update(ctx context.Context, in *input.UpdateUserInput) (*output.UserUpdaterOutput, error) {
	var result *output.UserUpdaterOutput

	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.qryGW.GetUser(ctx, entity.UserID(in.ID))
		if err != nil {
			return fmt.Errorf("UserUpdater.Update: %w", err)
		}
		if user == nil {
			return pkgerrors.NewNotFound("user not found")
		}

		if err := applyUserUpdates(user, in); err != nil {
			return err
		}

		user.UpdatedAt = time.Now()

		if err := s.cmdGW.UpdateUser(ctx, user); err != nil {
			if stderrors.Is(err, pkgerrors.ErrRecordNotFound) {
				return pkgerrors.NewNotFound("user not found")
			}
			return fmt.Errorf("UserUpdater.Update: %w", err)
		}

		out := toOutput(user)
		result = &out
		return nil
	})

	return result, err
}

func applyUserUpdates(user *entity.User, in *input.UpdateUserInput) error {
	var err error

	if in.Username != "" {
		user.Username, err = vo.NewUsername(in.Username)
		if err != nil {
			return err
		}
	}

	if in.Email != "" {
		user.Email, err = vo.NewEmail(in.Email)
		if err != nil {
			return err
		}
	}

	if in.Password != "" {
		user.Password, err = vo.NewPassword(in.Password)
		if err != nil {
			return err
		}
	}

	if in.Role != "" {
		user.Role, err = vo.NewUserRole(in.Role)
		if err != nil {
			return err
		}
	}

	return nil
}
