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
	const q = `
		INSERT INTO todos (user_id, title, description, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	var dueDate *time.Time
	if t.DueDate != nil {
		v := t.DueDate.Value()
		dueDate = &v
	}

	return r.db.QueryRow(ctx, q,
		t.UserID,
		t.Title.Value(),
		t.Description.Value(),
		t.Status.String(),
		t.Priority.String(),
		dueDate,
		t.CreatedAt,
		t.UpdatedAt,
	).Scan(&t.ID)
}

func (r *todoRepo) UpdateTodo(ctx context.Context, t *entity.Todo) error {
	const q = `
		UPDATE todos
		SET title=$1, description=$2, status=$3, priority=$4, due_date=$5, updated_at=$6
		WHERE id=$7`

	var dueDate *time.Time
	if t.DueDate != nil {
		v := t.DueDate.Value()
		dueDate = &v
	}

	result, err := r.db.Exec(ctx, q,
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
	if result.RowsAffected() == 0 {
		return apperrors.New(apperrors.REASON_NOT_FOUND, "Todo not found")
	}
	return nil
}

func (r *todoRepo) DeleteTodo(ctx context.Context, id entity.TodoID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM todos WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return apperrors.New(apperrors.REASON_NOT_FOUND, "Todo not found")
	}
	return nil
}

// --- TodoQueryGateway ---

func (r *todoRepo) GetTodo(ctx context.Context, id entity.TodoID) (*entity.Todo, error) {
	const q = `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos WHERE id=$1`

	row := r.db.QueryRow(ctx, q, id)
	todo, err := scanTodo(row)
	if err != nil {
		return nil, mapTodoRepoError(err)
	}
	return todo, nil
}

func (r *todoRepo) GetTodos(ctx context.Context, userID entity.UserID) ([]*entity.Todo, error) {
	const q = `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos WHERE user_id=$1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	var (
		id          int64
		userID      int64
		title       string
		description string
		status      string
		priority    string
		dueDate     *time.Time
		createdAt   time.Time
		updatedAt   time.Time
	)

	if err := s.Scan(&id, &userID, &title, &description, &status, &priority, &dueDate, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scanTodo: %w", err)
	}

	titleVO, err := vo.NewTodoTitle(title)
	if err != nil {
		return nil, err
	}
	descVO, err := vo.NewTodoDescription(description)
	if err != nil {
		return nil, err
	}
	statusVO, err := vo.NewTodoStatus(status)
	if err != nil {
		return nil, err
	}
	priorityVO, err := vo.NewTodoPriority(priority)
	if err != nil {
		return nil, err
	}

	t := &entity.Todo{
		ID:          entity.TodoID(id),
		UserID:      entity.UserID(userID),
		Title:       titleVO,
		Description: descVO,
		Status:      statusVO,
		Priority:    priorityVO,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if dueDate != nil {
		dd, err := vo.NewTodoDueDate(*dueDate)
		if err != nil {
			return nil, err
		}
		t.DueDate = &dd
	}

	return t, nil
}

func mapTodoRepoError(err error) error {
	if stderrors.Is(err, pgx.ErrNoRows) {
		return apperrors.New(apperrors.REASON_NOT_FOUND, "Todo not found")
	}
	return err
}
