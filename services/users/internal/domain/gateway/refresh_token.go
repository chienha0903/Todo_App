package gateway

import (
	"context"
	"time"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
)

type RefreshTokenCommandGateway interface {
	StoreRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	MarkUsed(ctx context.Context, tokenHash string, usedAt time.Time) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

type RefreshTokenQueryGateway interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
}
