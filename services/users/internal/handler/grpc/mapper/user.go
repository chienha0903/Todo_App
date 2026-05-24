package mapper

import (
	"time"

	userpb "github.com/chienha0903/Todo_App/proto/user"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

func ToCreateUserInput(req *userpb.CreateUserRequest) *input.CreateUserInput {
	return &input.CreateUserInput{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: req.PasswordHash,
		Role:         req.Role,
	}
}

func ToGetUserInput(req *userpb.GetUserRequest) *input.GetUserInput {
	return &input.GetUserInput{ID: req.Id}
}

func ToUpdateUserInput(req *userpb.UpdateUserRequest) *input.UpdateUserInput {
	return &input.UpdateUserInput{
		ID:           req.Id,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: req.PasswordHash,
		Role:         req.Role,
	}
}

func ToDeleteUserInput(req *userpb.DeleteUserRequest) *input.DeleteUserInput {
	return &input.DeleteUserInput{ID: req.Id}
}

func ToProtoUser(u *output.User) *userpb.User {
	return &userpb.User{
		Id:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    time.Unix(u.CreatedAt, 0).UTC().Format(time.RFC3339),
		UpdatedAt:    time.Unix(u.UpdatedAt, 0).UTC().Format(time.RFC3339),
	}
}
