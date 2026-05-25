package user

import (
	"context"

	userpb "github.com/chienha0903/Todo_App/proto/user"
	"github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/mapper"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
)

type UserHandler struct {
	userpb.UnimplementedUserserviceServer
	creater usecase.UserCreater
	getter  usecase.UserGetter
	lister  usecase.UserLister
	updater usecase.UserUpdater
	deleter usecase.UserDeleter
	authenticator usecase.UserAuthenticator
}

func NewUserHandler(
	creater usecase.UserCreater,
	getter usecase.UserGetter,
	lister usecase.UserLister,
	updater usecase.UserUpdater,
	deleter usecase.UserDeleter,
	authenticator usecase.UserAuthenticator,
) *UserHandler {
	return &UserHandler{
		creater: creater,
		getter:  getter,
		lister:  lister,
		updater: updater,
		deleter: deleter,
		authenticator: authenticator,
	}
}

func (h *UserHandler) CreateUser(
	ctx context.Context,
	req *userpb.CreateUserRequest,
) (*userpb.CreateUserResponse, error) {
	out, err := h.creater.Create(ctx, mapper.ToCreateUserInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.CreateUserResponse{User: mapper.ToProtoUser(out)}, nil
}

func (h *UserHandler) GetUser(
	ctx context.Context,
	req *userpb.GetUserRequest,
) (*userpb.GetUserResponse, error) {
	out, err := h.getter.Get(ctx, mapper.ToGetUserInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.GetUserResponse{User: mapper.ToProtoUser(out)}, nil
}

func (h *UserHandler) ListUsers(
	ctx context.Context,
	req *userpb.ListUsersRequest,
) (*userpb.ListUsersResponse, error) {
	page, err := h.lister.List(ctx, mapper.ToListUsersInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.ListUsersResponse{
		Users:    mapper.ToProtoUsers(page.Items),
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}, nil
}

func (h *UserHandler) UpdateUser(
	ctx context.Context,
	req *userpb.UpdateUserRequest,
) (*userpb.UpdateUserResponse, error) {
	out, err := h.updater.Update(ctx, mapper.ToUpdateUserInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.UpdateUserResponse{User: mapper.ToProtoUser(out)}, nil
}

func (h *UserHandler) DeleteUser(
	ctx context.Context,
	req *userpb.DeleteUserRequest,
) (*userpb.DeleteUserResponse, error) {
	if _, err := h.deleter.Delete(ctx, mapper.ToDeleteUserInput(req)); err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.DeleteUserResponse{}, nil
}

func (h *UserHandler) Login(
	ctx context.Context,
	req *userpb.LoginRequest,
) (*userpb.LoginResponse, error){
	out, err := h.authenticator.Login(ctx, mapper.ToUserLoginInput(req))

	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.LoginResponse{
		AccessToken: out.AccessToken,
	}, nil
}
