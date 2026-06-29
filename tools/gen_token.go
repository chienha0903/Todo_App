//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func sign(secret string, userID int64, role string) string {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign error: %v\n", err)
		os.Exit(1)
	}
	return t
}

func main() {
	secret := "chien-apvn-super-secret-key-2024"

	fmt.Println("=== TOKEN ADMIN (user_id=1, Role=ADMIN) ===")
	fmt.Println(sign(secret, 1, "ADMIN"))
	fmt.Println()
	fmt.Println("=== TOKEN USER (user_id=2, Role=USER) ===")
	fmt.Println(sign(secret, 2, "USER"))
}
