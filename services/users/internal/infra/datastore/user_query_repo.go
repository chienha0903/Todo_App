package datastore

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/maper"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userQueryRepo struct {
	db *gorm.DB
}

func NewUserQueryRepo(db *gorm.DB) *userQueryRepo {
	return &userQueryRepo{db: db}
}

func (r *userQueryRepo) GetUser(ctx context.Context, id entity.UserID) (*entity.User, error) {
	var m model.User

	result := extractDB(ctx, r.db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, int64(id))
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("db get user: %w", result.Error)
	}

	u, err := maper.ToEntity(&m)
	if err != nil {
		return nil, fmt.Errorf("db get user mapper: %w", err)
	}

	return u, nil
}

func (r *userQueryRepo) GetUsers(ctx context.Context, page, pageSize int32) ([]*entity.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("db count users: %w", err)
	}

	offset := int((page - 1) * pageSize)
	var ms []model.User
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(offset).
		Find(&ms)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("db list users: %w", result.Error)
	}

	users := make([]*entity.User, 0, len(ms))
	for i := range ms {
		u, err := maper.ToEntity(&ms[i])
		if err != nil {
			return nil, 0, fmt.Errorf("db list users mapper: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *userQueryRepo) GetUsersByIDs(ctx context.Context, ids []entity.UserID) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rawIDs := make([]int64, len(ids))
	for i, id := range ids {
		rawIDs[i] = int64(id)
	}

	var ms []model.User
	if err := r.db.WithContext(ctx).Where("id IN ?", rawIDs).Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("db get users by ids: %w", err)
	}

	users := make([]*entity.User, 0, len(ms))
	for i := range ms {
		u, err := maper.ToEntity(&ms[i])
		if err != nil {
			return nil, fmt.Errorf("db get users by ids mapper: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *userQueryRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.User

	result := extractDB(ctx, r.db).WithContext(ctx).First(&m, "email = ?", email)
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("db get user by email: %w", result.Error)
	}

	return maper.ToEntity(&m)
}
