package datastore

import (
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
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

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	return db, func() { _ = sqlDB.Close() }, nil
}

func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://services/users/internal/infra/datastore/migrations",
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

func NewUserCommandGateway(repo *userCommandRepo) gateway.UserCommandGateway {
	return repo
}

func NewUserQueryGateway(repo *userQueryRepo) gateway.UserQueryGateway {
	return repo
}
