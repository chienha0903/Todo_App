package datastore

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type txTestTodo struct {
	ID    uint `gorm:"primaryKey"`
	Title string
}

func setupTransactionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&txTestTodo{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestGormTransactor_Rollback_WhenFunctionReturnsError(t *testing.T) {
	ctx := context.Background()
	db := setupTransactionTestDB(t)

	transactor := NewGormTransactor(db)

	err := transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		todo := txTestTodo{
			Title: "rollback test",
		}

		if err := extractDB(ctx, db).Create(&todo).Error; err != nil {
			return err
		}

		return errors.New("force rollback")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var count int64
	if err := db.Model(&txTestTodo{}).
		Where("title = ?", "rollback test").
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("expected record to be rolled back, but found %d records", count)
	}
}

func TestGormTransactor_Commit_WhenFunctionReturnsNil(t *testing.T) {
	ctx := context.Background()
	db := setupTransactionTestDB(t)

	transactor := NewGormTransactor(db)

	err := transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		todo := txTestTodo{
			Title: "commit test",
		}

		if err := extractDB(ctx, db).Create(&todo).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var count int64
	if err := db.Model(&txTestTodo{}).
		Where("title = ?", "commit test").
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("expected record to be committed, but found %d records", count)
	}
}
