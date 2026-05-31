package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/event"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserDeleter = (*UserDeleter)(nil)

type UserDeleter struct {
	cmdGW    gateway.UserCommandGateway
	qryGW    gateway.UserQueryGateway
	outboxGW gateway.OutboxCommandGateway
	txGW     gateway.TransactionGateway
}

func NewUserDeleter(
	cmdGW gateway.UserCommandGateway,
	qryGW gateway.UserQueryGateway,
	outboxGW gateway.OutboxCommandGateway,
	txGW gateway.TransactionGateway,
) *UserDeleter {
	return &UserDeleter{
		cmdGW:    cmdGW,
		qryGW:    qryGW,
		outboxGW: outboxGW,
		txGW:     txGW,
	}
}

func (s *UserDeleter) Delete(ctx context.Context, in *input.DeleteUserInput) (*output.UserDeleterOutput, error) {
	userID := entity.UserID(in.ID)

	user, err := s.qryGW.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserDeleter.Delete get user: %w", err)
	}
	if user == nil {
		return nil, pkgerrors.NewNotFound("user not found")
	}

	err = s.txGW.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := time.Now()
		user.Status = vo.UserStatusDeleting
		user.DeletedAt = &now
		user.UpdatedAt = now

		if err := s.cmdGW.UpdateUser(txCtx, user); err != nil {
			if stderrors.Is(err, pkgerrors.ErrRecordNotFound) {
				return pkgerrors.NewNotFound("user not found")
			}
			return fmt.Errorf("update user: %w", err)
		}

		payload := &event.UserDeleteRequestedPayload{
			UserID:    int64(userID),
			Email:     user.Email.Value(),
			Username:  user.Username.Value(),
			DeletedAt: now,
		}
		payloadBytes, err := payload.ToJSON()
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		outboxEvent := &model.OutboxEvent{
			AggregateType: "user",
			AggregateID:   int64(userID),
			EventType:     event.EventTypeUserDeleteRequested,
			Payload:       payloadBytes,
		}
		if err := s.outboxGW.InsertOutboxEvent(txCtx, outboxEvent); err != nil {
			return fmt.Errorf("insert outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("UserDeleter.Delete: %w", err)
	}

	return &output.UserDeleterOutput{ID: in.ID}, nil
}
