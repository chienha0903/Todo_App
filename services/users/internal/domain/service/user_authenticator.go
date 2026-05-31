package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	userjwt "github.com/chienha0903/Todo_App/services/users/internal/infra/jwt"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/input"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase/output"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type UserAuthenticator struct {
	qryGW      gateway.UserQueryGateway
	tokenCmdGW gateway.RefreshTokenCommandGateway
	jwtSecret  string
}

func NewUserAuthenticator(qryGW gateway.UserQueryGateway, tokenCmdGW gateway.RefreshTokenCommandGateway, cfg *config.Config) *UserAuthenticator {
	return &UserAuthenticator{
		qryGW:      qryGW,
		tokenCmdGW: tokenCmdGW,
		jwtSecret:  cfg.JWTSecret,
	}
}

func (s *UserAuthenticator) Login(ctx context.Context, in *input.UserLoginInput) (*output.UserLoginOutput, error) {
	user, err := s.qryGW.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, fmt.Errorf("UserAuthenticator.Login: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.Value()), []byte(in.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	if user.Status.IsDeleting() || user.Status.IsDeleted() {
		return nil, errors.New("account is being deleted")
	}

	accessToken, err := userjwt.Generate(int64(user.UserID), user.Role.String(), s.jwtSecret, accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserAuthenticator.Login generate access token: %w", err)
	}

	refreshToken, err := userjwt.Generate(int64(user.UserID), user.Role.String(), s.jwtSecret, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("UserAuthenticator.Login generate refresh token: %w", err)
	}

	if err := s.tokenCmdGW.StoreRefreshToken(ctx, &entity.RefreshToken{
		UserID:    int64(user.UserID),
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}); err != nil {
		return nil, fmt.Errorf("UserAuthenticator.Login store refresh token: %w", err)
	}

	return &output.UserLoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
