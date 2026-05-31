package grpc

import (
	userpb "github.com/chienha0903/Todo_App/proto/user"
	userhandler "github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(h *userhandler.UserHandler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryRecoveryInterceptor,
			UnaryLoggingInterceptor,
		),
	)

	userpb.RegisterUserserviceServer(srv, h)

	reflection.Register(srv)

	return srv
}
