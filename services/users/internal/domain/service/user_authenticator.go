package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	userjwt "github.com/chienha0903/Todo_App/services/users/internal/infra/jwt"
	"golang.org/x/crypto/bcrypt"
)

const accessTokenTTL = 24 * time.Hour

type UserAuthenticator struct {
	qryGW     gateway.UserQueryGateway
	jwtSecret string
}

func NewUserAuthenticator(qryGW gateway.UserQueryGateway, jwtSecret string) *UserAuthenticator {
	return &UserAuthenticator{
		qryGW:     qryGW,
		jwtSecret: jwtSecret,
	}
}

func (s *UserAuthenticator) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.qryGW.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("UserAuthenticator.Login: %w", err)
	}

	if user == nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.Value()), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := userjwt.Generate(int64(user.UserID), user.Role.String(), s.jwtSecret, accessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("UserAuthenticator.Login generate token: %w", err)
	}

	return token, nil
}
