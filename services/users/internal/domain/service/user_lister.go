package service

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

var _ usecase.UserLister = (*UserLister)(nil)

type UserLister struct {
	qryGW gateway.UserQueryGateway
}

func NewUserLister(qryGW gateway.UserQueryGateway) *UserLister {
	return &UserLister{qryGW: qryGW}
}

func (s *UserLister) List(ctx context.Context, in *input.ListUsersInput) (*output.UserPage, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	users, total, err := s.qryGW.GetUsers(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("UserLister.List: %w", err)
	}

	items := make([]output.User, 0, len(users))
	for _, u := range users {
		items = append(items, toOutput(u))
	}

	return &output.UserPage{
		Items:    items,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
