package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	jwtpkg "github.com/chienha0903/Todo_App/pkg/jwt"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/caller"
)

func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		token, err := extractTokenFromMD(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := jwtpkg.ParseToken(token, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = caller.WithContext(ctx, claims.UserID, claims.Role)
		return handler(ctx, req)
	}
}

func extractTokenFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 || values[0] == "" {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == values[0] {
		return "", status.Error(codes.Unauthenticated, "authorization must be Bearer token")
	}

	return token, nil
}
