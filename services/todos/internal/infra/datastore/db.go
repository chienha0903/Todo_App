package datastore

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
)

func NewDB(cfg *config.Config) (*gorm.DB, func(), error) {
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("datastore: connect db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("datastore: get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, nil, fmt.Errorf("datastore: ping db: %w", err)
	}

	return db, func() { _ = sqlDB.Close() }, nil
}

func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://services/todos/internal/infra/datastore/migrations",
		databaseURL,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func NewTodoCommandGateway(repo *todoCommandRepo) gateway.TodoCommandGateway {
	return repo
}

func NewTodoQueryGateway(repo *todoQueryRepo) gateway.TodoQueryGateway {
	return repo
}
