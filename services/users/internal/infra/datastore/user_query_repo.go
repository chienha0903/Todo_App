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
