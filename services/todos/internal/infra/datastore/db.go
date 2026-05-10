package datastore

import (
	"fmt"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("datastore: connect db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("datastore: get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("datastore: ping db: %w", err)
	}

	if err := db.AutoMigrate(&model.Todo{}); err != nil {
		return nil, fmt.Errorf("datastore: auto migrate: %w", err)
	}

	return db, nil
}

func NewTodoCommandGateway(repo *todoCommandRepo) gateway.TodoCommandGateway {
	return repo
}

func NewTodoQueryGateway(repo *todoQueryRepo) gateway.TodoQueryGateway {
	return repo
}
