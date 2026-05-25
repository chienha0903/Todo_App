package service

import (
	"context"
	"fmt"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserGetter = (*UserGetter)(nil)

type UserGetter struct {
	qryGW gateway.UserQueryGateway
}

func NewUserGetter(qryGW gateway.UserQueryGateway) *UserGetter {
	return &UserGetter{qryGW: qryGW}
}

func (s *UserGetter) Get(ctx context.Context, in *input.GetUserInput) (*output.UserGetterOutput, error) {
	user, err := s.qryGW.GetUser(ctx, entity.UserID(in.ID))
	
	if err != nil {
		return nil, fmt.Errorf("UserGetter.Get: %w", err)
	}

	if user == nil {
		return nil, pkgerrors.NewNotFound("user not found")
	}

	out := toOutput(user)
	return &out, nil
}
