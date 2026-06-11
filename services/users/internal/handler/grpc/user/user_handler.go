package user

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/chienha0903/Todo_App/proto/user"
	"github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/caller"
	"github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/mapper"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
)

type UserHandler struct {
	userpb.UnimplementedUserserviceServer
	creater         usecase.UserCreater
	getter          usecase.UserGetter
	batchGetter     usecase.UserBatchGetter
	lister          usecase.UserLister
	updater         usecase.UserUpdater
	deleter         usecase.UserDeleter
	authenticator   usecase.UserAuthenticator
	refresher       usecase.UserRefresher
	passwordChanger usecase.UserPasswordChanger
}

func NewUserHandler(
	creater usecase.UserCreater,
	getter usecase.UserGetter,
	batchGetter usecase.UserBatchGetter,
	lister usecase.UserLister,
	updater usecase.UserUpdater,
	deleter usecase.UserDeleter,
	authenticator usecase.UserAuthenticator,
	refresher usecase.UserRefresher,
	passwordChanger usecase.UserPasswordChanger,
) *UserHandler {
	return &UserHandler{
		creater:         creater,
		getter:          getter,
		batchGetter:     batchGetter,
		lister:          lister,
		updater:         updater,
		deleter:         deleter,
		authenticator:   authenticator,
		refresher:       refresher,
		passwordChanger: passwordChanger,
	}
}

// CreateUser — public, không cần auth check (interceptor đã skip)
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
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" && req.Id != c.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot access another user's profile")
	}

	out, err := h.getter.Get(ctx, mapper.ToGetUserInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.GetUserResponse{User: mapper.ToProtoUser(out)}, nil
}

// GetUsersByIDs — chỉ dùng nội bộ bởi BFF (dataloader), không expose ra ngoài qua BFF auth layer
func (h *UserHandler) GetUsersByIDs(
	ctx context.Context,
	req *userpb.GetUsersByIDsRequest,
) (*userpb.GetUsersByIDsResponse, error) {
	users, err := h.batchGetter.GetByIDs(ctx, mapper.ToGetUsersByIDsInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.GetUsersByIDsResponse{Users: mapper.ToProtoUsers(users)}, nil
}

func (h *UserHandler) ListUsers(
	ctx context.Context,
	req *userpb.ListUsersRequest,
) (*userpb.ListUsersResponse, error) {
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" {
		return nil, status.Error(codes.PermissionDenied, "only admin can list all users")
	}

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
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" && req.Id != c.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot update another user's profile")
	}

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
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" {
		return nil, status.Error(codes.PermissionDenied, "only admin can delete users")
	}

	if _, err := h.deleter.Delete(ctx, mapper.ToDeleteUserInput(req)); err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.DeleteUserResponse{}, nil
}

// Login — public endpoint, interceptor skip qua publicMethods
func (h *UserHandler) Login(
	ctx context.Context,
	req *userpb.LoginRequest,
) (*userpb.LoginResponse, error) {
	out, err := h.authenticator.Login(ctx, mapper.ToUserLoginInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.LoginResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}

func (h *UserHandler) ChangePassword(
	ctx context.Context,
	req *userpb.ChangePasswordRequest,
) (*userpb.ChangePasswordResponse, error) {
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if req.UserId != c.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot change another user's password")
	}

	if err := h.passwordChanger.ChangePassword(ctx, mapper.ToChangePasswordInput(req)); err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.ChangePasswordResponse{}, nil
}

// RefreshToken — public endpoint, interceptor skip qua publicMethods
func (h *UserHandler) RefreshToken(
	ctx context.Context,
	req *userpb.RefreshTokenRequest,
) (*userpb.RefreshTokenResponse, error) {
	out, err := h.refresher.Refresh(ctx, mapper.ToRefreshTokenInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userpb.RefreshTokenResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}
