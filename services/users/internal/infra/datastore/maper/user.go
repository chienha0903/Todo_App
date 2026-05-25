package maper

import (
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

func ToModel(u *entity.User) *model.User {
	return &model.User{
		UserID:   int64(u.UserID),
		Username: u.Username.Value(),
		Email:    u.Email.Value(),
		Password: u.Password.Value(),
		Role:     u.Role.String(),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToEntity(m *model.User) (*entity.User, error) {
	username, err := vo.NewUsername(m.Username)
	if err != nil {
		return nil, err
	}

	email, err := vo.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	password, err := vo.NewPassword(m.Password)
	if err != nil {
		return nil, err
	}

	role, err := vo.NewUserRole(m.Role)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		UserID:    entity.UserID(m.UserID),
		Username:  username,
		Email:     email,
		Password:  password,
		Role:      role,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}
