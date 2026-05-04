package grpc

import (
	todopb "github.com/chienha0903/Todo_App/proto/todo"
	todohandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/todo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(h *todohandler.TodoHandler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryRecoveryInterceptor,
			UnaryLoggingInterceptor,
		),
	)
	todopb.RegisterTodoServiceServer(srv, h)
	reflection.Register(srv)
	return srv
}
