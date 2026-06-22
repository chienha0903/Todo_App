package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	todohandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/todo"
)

func NewGRPCServer(cfg *config.Config, h *todohandler.TodoHandler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			UnaryRecoveryInterceptor,
			UnaryLoggingInterceptor,
			UnaryAuthInterceptor(cfg.JWTSecret),
		),
	)
	todopb.RegisterTodoServiceServer(srv, h)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	reflection.Register(srv)
	return srv
}
