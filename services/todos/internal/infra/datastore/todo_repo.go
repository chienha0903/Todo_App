package datastore

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type todoRepo struct {
	db *pgxpool.Pool
}

// NewTodoRepo returns an implementation of both TodoCommandGateway and TodoQueryGateway.
func NewTodoRepo(db *pgxpool.Pool) *todoRepo {
	return &todoRepo{db: db}
}

// --- TodoCommandGateway ---

func (r *todoRepo) CreateTodo(ctx context.Context, t *entity.Todo) error {
	const query = `
		INSERT INTO todos (user_id, title, description, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	dueDate := todoDueDateValue(t.DueDate)

	row := r.db.QueryRow(
		ctx,
		query,
		t.UserID,
		t.Title.Value(),
		t.Description.Value(),
		t.Status.String(),
		t.Priority.String(),
		dueDate,
		t.CreatedAt,
		t.UpdatedAt,
	)

	return row.Scan(&t.ID)
}

func (r *todoRepo) UpdateTodo(ctx context.Context, t *entity.Todo) error {
	const query = `
		UPDATE todos
		SET title=$1, description=$2, status=$3, priority=$4, due_date=$5, updated_at=$6
		WHERE id=$7`

	dueDate := todoDueDateValue(t.DueDate)

	result, err := r.db.Exec(
		ctx,
		query,
		t.Title.Value(),
		t.Description.Value(),
		t.Status.String(),
		t.Priority.String(),
		dueDate,
		t.UpdatedAt,
		t.ID,
	)
	if err != nil {
		return err
	}
	return ensureTodoAffected(result.RowsAffected())
}

func (r *todoRepo) DeleteTodo(ctx context.Context, id entity.TodoID) error {
	const query = `DELETE FROM todos WHERE id=$1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return ensureTodoAffected(result.RowsAffected())
}

// --- TodoQueryGateway ---

func (r *todoRepo) GetTodo(ctx context.Context, id entity.TodoID) (*entity.Todo, error) {
	const query = `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos WHERE id=$1`

	row := r.db.QueryRow(ctx, query, id)
	todo, err := scanTodo(row)
	if err != nil {
		return nil, mapTodoRepoError(err)
	}
	return todo, nil
}

func (r *todoRepo) GetTodos(ctx context.Context, userID entity.UserID) ([]*entity.Todo, error) {
	const query = `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos WHERE user_id=$1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTodos(rows)
}

func scanTodos(rows pgx.Rows) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

// scanner is implemented by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanTodo(s scanner) (*entity.Todo, error) {
	row, err := scanTodoRow(s)
	if err != nil {
		return nil, err
	}
	return todoFromRow(row)
}

type todoRow struct {
	id          int64
	userID      int64
	title       string
	description string
	status      string
	priority    string
	dueDate     *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func scanTodoRow(s scanner) (todoRow, error) {
	var row todoRow

	err := s.Scan(
		&row.id,
		&row.userID,
		&row.title,
		&row.description,
		&row.status,
		&row.priority,
		&row.dueDate,
		&row.createdAt,
		&row.updatedAt,
	)
	if err != nil {
		return todoRow{}, fmt.Errorf("scan todo row: %w", err)
	}

	return row, nil
}

func todoFromRow(row todoRow) (*entity.Todo, error) {
	title, err := vo.NewTodoTitle(row.title)
	if err != nil {
		return nil, err
	}

	description, err := vo.NewTodoDescription(row.description)
	if err != nil {
		return nil, err
	}

	status, err := vo.NewTodoStatus(row.status)
	if err != nil {
		return nil, err
	}

	priority, err := vo.NewTodoPriority(row.priority)
	if err != nil {
		return nil, err
	}

	t := &entity.Todo{
		ID:          entity.TodoID(row.id),
		UserID:      entity.UserID(row.userID),
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		CreatedAt:   row.createdAt,
		UpdatedAt:   row.updatedAt,
	}

	if row.dueDate != nil {
		dueDate, err := vo.NewTodoDueDate(*row.dueDate)
		if err != nil {
			return nil, err
		}
		t.DueDate = &dueDate
	}

	return t, nil
}

func todoDueDateValue(dueDate *vo.TodoDueDate) *time.Time {
	if dueDate == nil {
		return nil
	}

	value := dueDate.Value()
	return &value
}

func ensureTodoAffected(rowsAffected int64) error {
	if rowsAffected == 0 {
		return apperrors.NewAppError(apperrors.ReasonNotFound, "Todo not found")
	}
	return nil
}

func mapTodoRepoError(err error) error {
	if stderrors.Is(err, pgx.ErrNoRows) {
		return apperrors.NewAppError(apperrors.ReasonNotFound, "Todo not found")
	}
	return err
}
