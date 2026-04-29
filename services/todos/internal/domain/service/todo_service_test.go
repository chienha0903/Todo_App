package service

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
)

func TestTodoCreatorCreate(t *testing.T) {
	dueDate := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)
	repo := &mockTodoGateway{
		createFn: func(ctx context.Context, todo *entity.Todo) error {
			todo.ID = entity.TodoID(10)
			return nil
		},
	}
	svc := NewTodoCreator(repo)

	got, err := svc.Create(context.Background(), &input.CreateTodoInput{
		UserID:      7,
		Title:       "  Buy milk  ",
		Description: "  Go to market  ",
		Priority:    "high",
		DueDate:     &dueDate,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.createdTodo == nil {
		t.Fatal("expected gateway CreateTodo to be called")
	}
	if got.ID != 10 || got.UserID != 7 {
		t.Fatalf("output id/user_id = %d/%d, want 10/7", got.ID, got.UserID)
	}
	if got.Title != "Buy milk" {
		t.Fatalf("output title = %q, want %q", got.Title, "Buy milk")
	}
	if got.Description != "Go to market" {
		t.Fatalf("output description = %q, want %q", got.Description, "Go to market")
	}
	if got.Priority != string(vo.TODO_PRIORITY_HIGH) {
		t.Fatalf("output priority = %q, want %q", got.Priority, vo.TODO_PRIORITY_HIGH)
	}
	if got.Status != string(vo.TODO_STATUS_PENDING) {
		t.Fatalf("output status = %q, want %q", got.Status, vo.TODO_STATUS_PENDING)
	}
	if got.DueDate == nil || !got.DueDate.Equal(dueDate) {
		t.Fatalf("output due date = %v, want %v", got.DueDate, dueDate)
	}
}

func TestTodoCreatorCreateInvalidInput(t *testing.T) {
	repo := &mockTodoGateway{}
	svc := NewTodoCreator(repo)

	got, err := svc.Create(context.Background(), &input.CreateTodoInput{
		UserID:      7,
		Title:       "",
		Description: "Go to market",
		Priority:    "HIGH",
	})
	if got != nil {
		t.Fatalf("Create() output = %#v, want nil", got)
	}
	assertAppErrorReason(t, err, apperrors.REASON_INVALID_PARAMETER)
	if repo.createCalls != 0 {
		t.Fatalf("CreateTodo calls = %d, want 0", repo.createCalls)
	}
}

func TestTodoCreatorCreateGatewayError(t *testing.T) {
	wantErr := apperrors.New(apperrors.REASON_INTERNAL_SERVER_ERROR, "create failed")
	repo := &mockTodoGateway{
		createFn: func(ctx context.Context, todo *entity.Todo) error {
			return wantErr
		},
	}
	svc := NewTodoCreator(repo)

	got, err := svc.Create(context.Background(), &input.CreateTodoInput{
		UserID:      7,
		Title:       "Buy milk",
		Description: "Go to market",
		Priority:    "HIGH",
	})
	if got != nil {
		t.Fatalf("Create() output = %#v, want nil", got)
	}
	if !stderrors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
}

type mockTodoGateway struct {
	createFn func(context.Context, *entity.Todo) error

	createCalls int
	createdTodo *entity.Todo
}

func (m *mockTodoGateway) CreateTodo(ctx context.Context, todo *entity.Todo) error {
	m.createCalls++
	m.createdTodo = todo
	if m.createFn != nil {
		return m.createFn(ctx, todo)
	}
	return nil
}

func (m *mockTodoGateway) UpdateTodo(ctx context.Context, todo *entity.Todo) error {
	return nil
}

func (m *mockTodoGateway) DeleteTodo(ctx context.Context, id entity.TodoID) error {
	return nil
}

func (m *mockTodoGateway) GetTodo(ctx context.Context, id entity.TodoID) (*entity.Todo, error) {
	return nil, nil
}

func (m *mockTodoGateway) GetTodos(ctx context.Context, userID entity.UserID) ([]*entity.Todo, error) {
	return nil, nil
}

func assertAppErrorReason(t *testing.T, err error, reason apperrors.Reason) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperrors.Error
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Reason != reason {
		t.Fatalf("error reason = %q, want %q", appErr.Reason, reason)
	}
}
