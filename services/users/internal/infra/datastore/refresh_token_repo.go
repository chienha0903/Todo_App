package datastore

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore/model"
)

type RefreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func NewRefreshTokenCommandGateway(repo *RefreshTokenRepo) gateway.RefreshTokenCommandGateway {
	return repo
}

func NewRefreshTokenQueryGateway(repo *RefreshTokenRepo) gateway.RefreshTokenQueryGateway {
	return repo
}

func (r *RefreshTokenRepo) StoreRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	m := &model.RefreshToken{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}

	result := extractDB(ctx, r.db).WithContext(ctx).Create(m)
	if result.Error != nil {
		return fmt.Errorf("db store refresh token: %w", result.Error)
	}

	token.ID = m.ID
	return nil
}

func (r *RefreshTokenRepo) MarkUsed(ctx context.Context, tokenHash string, usedAt time.Time) error {
	result := extractDB(ctx, r.db).WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("used_at", usedAt)

	if result.Error != nil {
		return fmt.Errorf("db mark refresh token used: %w", result.Error)
	}

	return nil
}

func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	result := extractDB(ctx, r.db).WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("db delete refresh tokens by user: %w", result.Error)
	}

	return nil
}

func (r *RefreshTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var m model.RefreshToken

	result := extractDB(ctx, r.db).WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&m)

	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("db find refresh token: %w", result.Error)
	}

	return &entity.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}, nil
}
