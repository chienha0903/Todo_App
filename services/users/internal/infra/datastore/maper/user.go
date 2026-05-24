package maper

import (
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/users/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

func ToModel(u *entity.User) *model.User {
	return &model.User{
		UserID:       int64(u.UserID),
		Username:     u.Username.Value(),
		Email:        u.Email.Value(),
		PasswordHash: u.PasswordHash.Value(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
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

	passwordHash, err := vo.NewPasswordHash(m.PasswordHash)
	if err != nil {
		return nil, err
	}

	u := &entity.User{
		UserID:       entity.UserID(m.UserID),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}

	return u, nil
}
