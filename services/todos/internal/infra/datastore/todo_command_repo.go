package datastore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/mapper"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore/model"
	"github.com/chienha0903/Todo_App/services/todos/internal/observability/tracing"
)

var tracer = otel.Tracer("todos/datastore")

type todoCommandRepo struct {
	db *gorm.DB
}

func NewTodoCommandRepo(db *gorm.DB) *todoCommandRepo {
	return &todoCommandRepo{db: db}
}

func (r *todoCommandRepo) CreateTodo(ctx context.Context, t *entity.Todo) error {
	ctx, span := tracer.Start(ctx, "repository.todos.insert",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation.name", "INSERT"),
			attribute.String("db.collection.name", "todos"),
		))
	defer span.End()

	m := mapper.ToModel(t)

	result := extractDB(ctx, r.db).WithContext(ctx).Create(m)
	if result.Error != nil {
		tracing.RecordError(span, result.Error)
		return fmt.Errorf("db create todo: %w", result.Error)
	}

	t.ID = entity.TodoID(m.ID)
	span.SetAttributes(attribute.Int64("todo.id", int64(t.ID)))
	span.SetStatus(otelcodes.Ok, "")
	return nil
}

func (r *todoCommandRepo) UpdateTodo(ctx context.Context, t *entity.Todo) error {
	m := mapper.ToModel(t)

	result := extractDB(ctx, r.db).WithContext(ctx).Save(m)
	if result.Error != nil {
		return fmt.Errorf("db update todo: %w", result.Error)
	}

	return ensureTodoAffected(result.RowsAffected)
}

func (r *todoCommandRepo) DeleteTodo(ctx context.Context, id entity.TodoID) error {
	result := extractDB(ctx, r.db).WithContext(ctx).Delete(&model.Todo{}, int64(id))
	if result.Error != nil {
		return fmt.Errorf("db delete todo: %w", result.Error)
	}

	return ensureTodoAffected(result.RowsAffected)
}

func (r *todoCommandRepo) SoftDeleteByUserID(ctx context.Context, userID int64) error {
	now := time.Now()

	result := extractDB(ctx, r.db).WithContext(ctx).
		Model(&model.Todo{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("db soft delete todos by user: %w", result.Error)
	}

	slog.Info("soft deleted todos", "user_id", userID, "count", result.RowsAffected)

	return nil
}

func ensureTodoAffected(rowsAffected int64) error {
	if rowsAffected == 0 {
		return pkgerrors.ErrRecordNotFound
	}

	return nil
}
