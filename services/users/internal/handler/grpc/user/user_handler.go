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
	updater usecase.UserUpdater
	deleter usecase.UserDeleter
}

func NewUserHandler(
	creater usecase.UserCreater,
	getter usecase.UserGetter,
	updater usecase.UserUpdater,
	deleter usecase.UserDeleter,
) *UserHandler {
	return &UserHandler{
		creater: creater,
		getter:  getter,
		updater: updater,
		deleter: deleter,
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
