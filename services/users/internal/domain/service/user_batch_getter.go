package service

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserBatchGetter = (*UserBatchGetter)(nil)

type UserBatchGetter struct {
	qryGW gateway.UserQueryGateway
}

func NewUserBatchGetter(qryGW gateway.UserQueryGateway) *UserBatchGetter {
	return &UserBatchGetter{qryGW: qryGW}
}

func (s *UserBatchGetter) GetByIDs(ctx context.Context, in *input.GetUsersByIDsInput) ([]output.User, error) {
	ids := make([]entity.UserID, len(in.IDs))
	for i, id := range in.IDs {
		ids[i] = entity.UserID(id)
	}

	users, err := s.qryGW.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("UserBatchGetter.GetByIDs: %w", err)
	}

	result := make([]output.User, 0, len(users))
	for _, u := range users {
		result = append(result, toOutput(u))
	}

	return result, nil
}
