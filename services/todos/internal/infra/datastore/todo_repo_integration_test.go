//go:build integration

package datastore

import (
	"context"
	"errors"
	"testing"

	pkgerrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

func TestIntegration_CreateAndGetTodo(t *testing.T) {
	db := newTestDB(t)
	cmd := NewTodoCommandRepo(db)
	qry := NewTodoQueryRepo(db)
	ctx := context.Background()

	todo := newTestTodo(1)
	if err := cmd.CreateTodo(ctx, todo); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if todo.ID == 0 {
		t.Fatal("expected ID to be set after create")
	}

	got, err := qry.GetTodo(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetTodo: %v", err)
	}
	if got == nil {
		t.Fatal("expected todo, got nil")
	}
	if got.Title.Value() != todo.Title.Value() {
		t.Fatalf("title: got %q, want %q", got.Title.Value(), todo.Title.Value())
	}
	if got.UserID != todo.UserID {
		t.Fatalf("userID: got %d, want %d", got.UserID, todo.UserID)
	}
}

func TestIntegration_UpdateTodo(t *testing.T) {
	db := newTestDB(t)
	cmd := NewTodoCommandRepo(db)
	qry := NewTodoQueryRepo(db)
	ctx := context.Background()

	todo := newTestTodo(1)
	if err := cmd.CreateTodo(ctx, todo); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	newStatus, _ := vo.NewTodoStatus("IN_PROGRESS")
	todo.Status = newStatus
	if err := cmd.UpdateTodo(ctx, todo); err != nil {
		t.Fatalf("UpdateTodo: %v", err)
	}

	got, err := qry.GetTodo(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetTodo: %v", err)
	}
	if got.Status.String() != "IN_PROGRESS" {
		t.Fatalf("status: got %q, want IN_PROGRESS", got.Status.String())
	}
}

func TestIntegration_DeleteTodo(t *testing.T) {
	db := newTestDB(t)
	cmd := NewTodoCommandRepo(db)
	qry := NewTodoQueryRepo(db)
	ctx := context.Background()

	todo := newTestTodo(1)
	if err := cmd.CreateTodo(ctx, todo); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	if err := cmd.DeleteTodo(ctx, todo.ID); err != nil {
		t.Fatalf("DeleteTodo: %v", err)
	}

	got, err := qry.GetTodo(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetTodo after delete: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil after hard delete, got todo")
	}
}

func TestIntegration_DeleteTodo_NotFound(t *testing.T) {
	db := newTestDB(t)
	cmd := NewTodoCommandRepo(db)
	ctx := context.Background()

	err := cmd.DeleteTodo(ctx, entity.TodoID(99999))
	if !errors.Is(err, pkgerrors.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestIntegration_GetTodos_Pagination(t *testing.T) {
	db := newTestDB(t)
	cmd := NewTodoCommandRepo(db)
	qry := NewTodoQueryRepo(db)
	ctx := context.Background()

	const userID = int64(1)
	for range 5 {
		if err := cmd.CreateTodo(ctx, newTestTodo(userID)); err != nil {
			t.Fatalf("CreateTodo: %v", err)
		}
	}

	page1, total, err := qry.GetTodos(ctx, entity.UserID(userID), 1, 3)
	if err != nil {
		t.Fatalf("GetTodos page 1: %v", err)
	}
	if total != 5 {
		t.Fatalf("total: got %d, want 5", total)
	}
	if len(page1) != 3 {
		t.Fatalf("page 1 len: got %d, want 3", len(page1))
	}

	page2, _, err := qry.GetTodos(ctx, entity.UserID(userID), 2, 3)
	if err != nil {
		t.Fatalf("GetTodos page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page 2 len: got %d, want 2", len(page2))
	}
}
