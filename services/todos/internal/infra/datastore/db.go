package datastore

import (
	"context"
	"fmt"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet registers all datastore providers for Wire.
var ProviderSet = wire.NewSet(
	NewDB,
	NewTodoRepo,
	wire.Bind(new(gateway.TodoCommandGateway), new(*todoRepo)),
	wire.Bind(new(gateway.TodoQueryGateway), new(*todoRepo)),
)

func NewDB(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("datastore: connect db: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("datastore: ping db: %w", err)
	}
	return pool, nil
}
