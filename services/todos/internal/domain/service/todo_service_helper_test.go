package service

import (
	stderrors "errors"
	"testing"
	"time"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

func assertAppErrorCode(t *testing.T, err error, code apperrors.ErrorCode) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperrors.AppError
	if !stderrors.As(err, &appErr) {
		if code == apperrors.ErrInternal {
			return // plain technical errors are treated as internal
		}
		t.Fatalf("expected *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %q, want %q", appErr.Code, code)
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
		Status:      vo.TodoStatusPending,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
