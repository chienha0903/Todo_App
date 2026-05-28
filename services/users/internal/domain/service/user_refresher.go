package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	userjwt "github.com/chienha0903/Todo_App/services/users/internal/infra/jwt"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

type UserRefresher struct {
	tokenCmdGW gateway.RefreshTokenCommandGateway
	tokenQryGW gateway.RefreshTokenQueryGateway
	jwtSecret  string
}

func NewUserRefresher(tokenCmdGW gateway.RefreshTokenCommandGateway, tokenQryGW gateway.RefreshTokenQueryGateway, cfg *config.Config) *UserRefresher {
	return &UserRefresher{
		tokenCmdGW: tokenCmdGW,
		tokenQryGW: tokenQryGW,
		jwtSecret:  cfg.JWTSecret,
	}
}

func (s *UserRefresher) Refresh(ctx context.Context, in *input.UserRefreshTokenInput) (*output.UserRefreshTokenOutput, error) {
	claims, err := userjwt.Parse(in.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, pkgerrors.NewAuthN("invalid refresh token")
	}

	tokenHash := hashToken(in.RefreshToken)
	stored, err := s.tokenQryGW.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh: %w", err)
	}
	if stored == nil || stored.IsExpired() {
		return nil, pkgerrors.NewAuthN("refresh token not found or expired")
	}

	if stored.IsUsed() {
		return nil, pkgerrors.NewAuthN("refresh token already used")
	}

	now := time.Now()
	if err := s.tokenCmdGW.MarkUsed(ctx, tokenHash, now); err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh mark used: %w", err)
	}

	accessToken, err := userjwt.Generate(claims.UserID, claims.Role, s.jwtSecret, accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh generate access token: %w", err)
	}

	refreshToken, err := userjwt.Generate(claims.UserID, claims.Role, s.jwtSecret, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh generate refresh token: %w", err)
	}

	if err := s.tokenCmdGW.StoreRefreshToken(ctx, &entity.RefreshToken{
		UserID:    claims.UserID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: now.Add(refreshTokenTTL),
	}); err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh store new token: %w", err)
	}

	return &output.UserRefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
