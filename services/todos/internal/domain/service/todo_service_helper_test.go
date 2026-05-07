package service

import (
	stderrors "errors"
	"testing"
	"time"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

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

func newFixtureTodo() *entity.Todo {
	title, _ := vo.NewTodoTitle("Buy milk")
	desc, _ := vo.NewTodoDescription("Go to market")
	priority, _ := vo.NewTodoPriority("low")
	now := time.Now()
	return &entity.Todo{
		ID:          entity.TodoID(5),
		UserID:      entity.UserID(7),
		Title:       title,
		Description: desc,
		Status:      vo.TODO_STATUS_PENDING,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
