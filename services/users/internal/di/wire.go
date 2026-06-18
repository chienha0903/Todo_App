//go:build wireinject

package di

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/service"
	grpchandler "github.com/chienha0903/Todo_App/services/users/internal/handler/grpc"
	userhandler "github.com/chienha0903/Todo_App/services/users/internal/handler/grpc/user"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore"
	redisstore "github.com/chienha0903/Todo_App/services/users/internal/infra/redis"
	"github.com/chienha0903/Todo_App/services/users/internal/usecase"
)

func InitializeApp(cfg *config.Config) (*grpc.Server, func(), error) {
	wire.Build(
		// infra
		datastore.NewDB,
		datastore.NewUserCommandRepo,
		datastore.NewUserCommandGateway,
		datastore.NewUserQueryRepo,
		datastore.NewUserQueryGateway,
		datastore.NewGormTransactor,
		wire.Bind(new(gateway.TransactionGateway), new(*datastore.GormTransactor)),
		datastore.NewOutboxRepo,
		datastore.NewOutboxCommandGateway,

		// Redis infra
		redisstore.NewClient,
		redisstore.NewRefreshTokenRepo,
		redisstore.NewRefreshTokenCommandGateway,
		redisstore.NewRefreshTokenQueryGateway,

		// domain service
		service.NewUserCreater,
		wire.Bind(new(usecase.UserCreater), new(*service.UserCreater)),
		service.NewUserGetter,
		wire.Bind(new(usecase.UserGetter), new(*service.UserGetter)),
		service.NewUserBatchGetter,
		wire.Bind(new(usecase.UserBatchGetter), new(*service.UserBatchGetter)),
		service.NewUserUpdater,
		wire.Bind(new(usecase.UserUpdater), new(*service.UserUpdater)),
		service.NewUserDeleter,
		wire.Bind(new(usecase.UserDeleter), new(*service.UserDeleter)),
		service.NewUserLister,
		wire.Bind(new(usecase.UserLister), new(*service.UserLister)),
		service.NewUserAuthenticator,
		wire.Bind(new(usecase.UserAuthenticator), new(*service.UserAuthenticator)),
		service.NewUserRefresher,
		wire.Bind(new(usecase.UserRefresher), new(*service.UserRefresher)),
		service.NewUserPasswordChanger,
		wire.Bind(new(usecase.UserPasswordChanger), new(*service.UserPasswordChanger)),
		// handler
		userhandler.NewUserHandler,
		grpchandler.NewGRPCServer,
	)
	return nil, nil, nil
}
