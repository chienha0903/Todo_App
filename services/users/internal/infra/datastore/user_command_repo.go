package datastore

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/maper"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

type userCommandRepo struct {
	db *gorm.DB
}

func NewUserCommandRepo(db *gorm.DB) *userCommandRepo {
	return &userCommandRepo{db: db}
}

func (r *userCommandRepo) CreateUser(ctx context.Context, u *entity.User) error {
	m := maper.ToModel(u)

	result := extractDB(ctx, r.db).WithContext(ctx).Create(m)
	if result.Error != nil {
		return fmt.Errorf("db create user: %w", result.Error)
	}

	u.UserID = entity.UserID(m.UserID)
	return nil
}

func (r *userCommandRepo) UpdateUser(ctx context.Context, u *entity.User) error {
	m := maper.ToModel(u)

	result := extractDB(ctx, r.db).WithContext(ctx).Save(m)
	if result.Error != nil {
		return fmt.Errorf("db update user: %w", result.Error)
	}

	return ensureUserAffected(result.RowsAffected)
}

func (r *userCommandRepo) DeleteUser(ctx context.Context, id entity.UserID) error {
	result := extractDB(ctx, r.db).WithContext(ctx).Delete(&model.User{}, int64(id))
	if result.Error != nil {
		return fmt.Errorf("db delete user: %w", result.Error)
	}

	return ensureUserAffected(result.RowsAffected)
}

func ensureUserAffected(rowsAffected int64) error {
	if rowsAffected == 0 {
		return pkgerrors.ErrRecordNotFound
	}
	return nil
}
