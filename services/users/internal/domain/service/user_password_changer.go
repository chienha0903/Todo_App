package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/event"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
)

var _ usecase.UserPasswordChanger = (*UserPasswordChanger)(nil)

type UserPasswordChanger struct {
	qryGW      gateway.UserQueryGateway
	cmdGW      gateway.UserCommandGateway
	outboxGW   gateway.OutboxCommandGateway
	transactor gateway.TransactionGateway
}

func NewUserPasswordChanger(
	qryGW gateway.UserQueryGateway,
	cmdGW gateway.UserCommandGateway,
	outboxGW gateway.OutboxCommandGateway,
	transactor gateway.TransactionGateway,
) *UserPasswordChanger {
	return &UserPasswordChanger{
		qryGW:      qryGW,
		cmdGW:      cmdGW,
		outboxGW:   outboxGW,
		transactor: transactor,
	}
}

func (s *UserPasswordChanger) ChangePassword(ctx context.Context, in *input.ChangePasswordInput) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.qryGW.GetUser(ctx, entity.UserID(in.UserID))
		if err != nil {
			return fmt.Errorf("UserPasswordChanger.ChangePassword: %w", err)
		}
		if user == nil {
			return pkgerrors.NewNotFound("user not found")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password.Value()), []byte(in.CurrentPassword)); err != nil {
			return pkgerrors.NewAuthN("current password is incorrect")
		}

		newPwd, err := vo.NewPassword(in.NewPassword)
		if err != nil {
			return err
		}
		hashedPwd, err := hashPassword(newPwd)
		if err != nil {
			return err
		}
		user.Password = hashedPwd

		if err := s.cmdGW.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("UserPasswordChanger.ChangePassword update: %w", err)
		}

		payload, err := (&event.DeleteUserTokensPayload{UserID: in.UserID}).ToJSON()
		if err != nil {
			return fmt.Errorf("UserPasswordChanger.ChangePassword marshal cache payload: %w", err)
		}
		if err := s.outboxGW.InsertOutboxEvent(ctx, &model.OutboxEvent{
			AggregateType: event.CacheAggregateType,
			AggregateID:   in.UserID,
			EventType:     event.CacheEventDeleteUserTokens,
			Payload:       payload,
		}); err != nil {
			return fmt.Errorf("UserPasswordChanger.ChangePassword insert cache event: %w", err)
		}

		return nil
	})
}
