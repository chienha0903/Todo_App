package grpc

import (
	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ProviderSet registers handler providers for Wire.
var ProviderSet = wire.NewSet(NewTodoHandler, NewGRPCServer)

func NewGRPCServer(todoHandler *todoHandler) *grpc.Server {
	srv := grpc.NewServer()
	todopb.RegisterTodoServiceServer(srv, todoHandler)
	// reflection cho phép dùng grpcurl để test
	reflection.Register(srv)
	return srv
}
