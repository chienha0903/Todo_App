//go:build wireinject

package di

import (
	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/service"
	grpchandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc"
	todohandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// InitGRPCServer wires all dependencies and returns a ready *grpc.Server.
func InitGRPCServer(cfg *config.Config) (*grpc.Server, error) {
	wire.Build(
		datastore.NewDB,
		datastore.NewTodoRepo,
		datastore.NewTodoCommandGateway,
		datastore.NewTodoQueryGateway,
		service.NewTodoCreator,
		service.NewTodoGetter,
		service.NewTodoLister,
		service.NewTodoUpdater,
		service.NewTodoDeleter,
		todohandler.NewTodoHandler,
		grpchandler.NewGRPCServer,
	)
	return nil, nil
}
