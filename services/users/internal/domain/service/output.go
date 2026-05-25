package service

import (
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

func toOutput(u *entity.User) output.User {
	return output.User{
		ID:        int64(u.UserID),
		Email:     u.Email.Value(),
		Username:  u.Username.Value(),
		Password:  u.Password.Value(),
		Role:      u.Role.String(),
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}
