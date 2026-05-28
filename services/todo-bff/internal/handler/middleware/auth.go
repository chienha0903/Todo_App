package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	userjwt "github.com/chienha0903/Todo_App/services/todo-bff/internal/jwt"
)

type contextKey string

const (
	ctxUserID contextKey = "user_id"
	ctxRole   contextKey = "role"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)

			if token != "" {
				claim, err := userjwt.ParseToken(token, jwtSecret)
				if err == nil {
					ctx := context.WithValue(r.Context(), ctxUserID, claim.UserID)
					ctx = context.WithValue(ctx, ctxRole, claim.Role)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(authHeader, "Bearer ")
}

func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(ctxUserID).(int64)
	return userID, ok
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ctxRole).(string)
	return role, ok
}

func RequireAuth(ctx context.Context) error {
	userID, ok := GetUserID(ctx)
	if !ok || userID == 0 {
		return apperror.Unauthorized()
	}
	
	return nil
}
