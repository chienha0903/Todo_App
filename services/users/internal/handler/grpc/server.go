package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	userpb "github.com/chienha0903/Todo_App/proto/user"
	userhandler "github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/user"
)

func NewGRPCServer(h *userhandler.UserHandler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryRecoveryInterceptor,
			UnaryLoggingInterceptor,
		),
	)

	userpb.RegisterUserserviceServer(srv, h)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())

	reflection.Register(srv)

	return srv
}
