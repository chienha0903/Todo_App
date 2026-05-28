package resolver

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/model"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/middleware"
	ucin "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/input"
)

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	token, err := r.userGateway.Login(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{AccessToken: token}, nil
}

func (r *queryResolver) GetUser(ctx context.Context, id string) (*model.User, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	u, err := r.userGateway.GetUser(ctx, &ucin.GetUser{ID: userID})
	if err != nil {
		return nil, err
	}
	return toUserModel(u), nil
}

func (r *queryResolver) ListUsers(ctx context.Context, page *int, pageSize *int) (*model.UserPage, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	p, ps := int32(1), int32(20)
	if page != nil && *page > 0 {
		p = int32(*page)
	}
	if pageSize != nil && *pageSize > 0 {
		ps = int32(*pageSize)
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result, err := r.userGateway.ListUsers(ctx, &ucin.ListUsers{Page: p, PageSize: ps})
	if err != nil {
		return nil, err
	}
	return toUserPageModel(result), nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input model.UpdateUserInput) (*model.User, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	u, err := r.userGateway.UpdateUser(ctx, &ucin.UpdateUser{
		ID:       userID,
		Email:    derefStr(input.Email),
		Username: derefStr(input.Username),
		Password: derefStr(input.Password),
		Role:     derefStr(input.Role),
	})
	if err != nil {
		return nil, err
	}
	return toUserModel(u), nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
	if err := requireAdmin(ctx); err != nil {
		return false, err
	}

	userID, err := parseID(id)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if err := r.userGateway.DeleteUser(ctx, &ucin.DeleteUser{ID: userID}); err != nil {
		return false, err
	}
	return true, nil
}

func requireAdmin(ctx context.Context) error {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return apperror.Unauthorized()
	}
	role, _ := middleware.GetRole(ctx)
	if role != "admin" {
		return apperror.PermissionDenied()
	}
	return nil
}
