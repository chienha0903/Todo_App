//go:build wireinject

package di

import (
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/service"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/graph/resolver"
	infratodo "github.com/chienha0903/Todo_App/services/todo-bff/internal/infra/todo"
	todousecase "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo"
	"github.com/google/wire"
)

func InitializeApp(cfg *config.Config) (*resolver.Resolver, func(), error) {
	wire.Build(
		// infra
		infratodo.NewGRPCConn,
		infratodo.NewTodoServiceClient,
		infratodo.NewGRPCGateway,

		// domain service
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

		// handler
		resolver.NewResolver,
	)
	return nil, nil, nil
}
