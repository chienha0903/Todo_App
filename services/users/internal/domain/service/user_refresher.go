package service

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
	userjwt "github.com/chienha0903/Todo_App/services/users/internal/infra/jwt"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
)

type UserRefresher struct {
	jwtSecret string
}

func NewUserRefresher(cfg *config.Config) *UserRefresher {
	return &UserRefresher{jwtSecret: cfg.JWTSecret}
}

func (s *UserRefresher) Refresh(_ context.Context, in *input.UserRefreshTokenInput) (*output.UserRefreshTokenOutput, error) {
	claims, err := userjwt.Parse(in.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	accessToken, err := userjwt.Generate(claims.UserID, claims.Role, s.jwtSecret, accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh generate access token: %w", err)
	}

	refreshToken, err := userjwt.Generate(claims.UserID, claims.Role, s.jwtSecret, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserRefresher.Refresh generate refresh token: %w", err)
	}

	return &output.UserRefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
