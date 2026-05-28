package jwt

import (
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

func Generate(userID int64, role, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

