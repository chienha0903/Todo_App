//go:build wireinject

package di

import (
	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	grpchandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/service"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// InitGRPCServer wires all dependencies and returns a ready *grpc.Server.
func InitGRPCServer(cfg *config.Config) (*grpc.Server, error) {
	wire.Build(
		datastore.ProviderSet,
		service.ProviderSet,
		grpchandler.ProviderSet,
	)
	return nil, nil
}
