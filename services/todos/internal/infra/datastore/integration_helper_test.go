//go:build integration

package datastore

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

// migrationsDir trả về absolute path đến thư mục migrations.
// go test đổi working directory về package dir nên không thể dùng relative path từ project root.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "migrations")
}

func runTestMigrations(t *testing.T, databaseURL string) {
	t.Helper()
	src := "file://" + filepath.ToSlash(migrationsDir())
	m, err := migrate.New(src, databaseURL)
	if err != nil {
		t.Fatalf("init migrate: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
}

// newTestDB tạo một PostgreSQL database tạm thời, random tên, chạy migration,
// và tự động DROP database đó sau khi test kết thúc.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	baseDSN := os.Getenv("TODO_DB_DSN")
	if baseDSN == "" {
		baseDSN = "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"
	}

	dbName := fmt.Sprintf("todo_test_%d_%d", time.Now().UnixNano(), rand.Intn(10000))

	u, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}

	// kết nối admin DB để tạo test DB
	adminURL := *u
	adminURL.Path = "/postgres"
	adminDB, err := gorm.Open(postgres.Open(adminURL.String()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect admin db: %v", err)
	}
	if err := adminDB.Exec("CREATE DATABASE " + dbName).Error; err != nil {
		t.Fatalf("create test db %q: %v", dbName, err)
	}

	// chạy migration trên test DB dùng absolute path
	testURL := *u
	testURL.Path = "/" + dbName
	testDSN := testURL.String()
	runTestMigrations(t, testDSN)

	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_ = adminDB.Exec("DROP DATABASE " + dbName).Error
		adminSQLDB, _ := adminDB.DB()
		_ = adminSQLDB.Close()
	})

	return db
}

func newTestTodo(userID int64) *entity.Todo {
	title, _ := vo.NewTodoTitle("integration test todo")
	desc, _ := vo.NewTodoDescription("integration test description")
	status, _ := vo.NewTodoStatus("PENDING")
	priority, _ := vo.NewTodoPriority("MEDIUM")
	return &entity.Todo{
		UserID:      entity.UserID(userID),
		Title:       title,
		Description: desc,
		Status:      status,
		Priority:    priority,
	}
}
