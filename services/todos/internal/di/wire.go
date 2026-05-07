//go:build wireinject

package di

import (
	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/service"
	grpchandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc"
	todohandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// InitGRPCServer wires all dependencies and returns a ready *grpc.Server.
func InitGRPCServer(cfg *config.Config) (*grpc.Server, error) {
	wire.Build(
		datastore.NewDB,
		datastore.NewTodoCommandRepo,
		datastore.NewTodoCommandGateway,
		datastore.NewTodoQueryRepo,
		datastore.NewTodoQueryGateway,
		service.NewTodoCreater,
		wire.Bind(new(todousecase.TodoCreater), new(*service.TodoCreater)),
		service.NewTodoGetter,
		wire.Bind(new(todousecase.TodoGetter), new(*service.TodoGetter)),
		service.NewTodoLister,
		wire.Bind(new(todousecase.TodoLister), new(*service.TodoLister)),
		service.NewTodoUpdater,
		wire.Bind(new(todousecase.TodoUpdater), new(*service.TodoUpdater)),
		service.NewTodoDeleter,
		wire.Bind(new(todousecase.TodoDeleter), new(*service.TodoDeleter)),
		todohandler.NewTodoHandler,
		grpchandler.NewGRPCServer,
	)
	return nil, nil
}
