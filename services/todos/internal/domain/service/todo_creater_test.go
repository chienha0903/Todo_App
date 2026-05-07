package service

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	gatewaymock "github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway/mock"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
	"go.uber.org/mock/gomock"
)

func TestTodoCreaterCreate(t *testing.T) {
	dueDate := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		input      *input.CreateTodoInput
		setupMock  func(repo *gatewaymock.MockTodoCommandGateway)
		wantErr    bool
		wantReason apperrors.Reason
		check      func(t *testing.T, got *output.TodoCreater)
	}{
		{
			name: "success",
			input: &input.CreateTodoInput{
				UserID:      7,
				Title:       "  Buy milk  ",
				Description: "  Go to market  ",
				Priority:    "high",
				DueDate:     &dueDate,
			},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().
					CreateTodo(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, todo *entity.Todo) error {
						todo.ID = entity.TodoID(10)
						return nil
					})
			},
			check: func(t *testing.T, got *output.TodoCreater) {
				if got.ID != 10 || got.UserID != 7 {
					t.Fatalf("id/user_id = %d/%d, want 10/7", got.ID, got.UserID)
				}
				if got.Title != "Buy milk" {
					t.Fatalf("title = %q, want %q", got.Title, "Buy milk")
				}
				if got.Description != "Go to market" {
					t.Fatalf("description = %q, want %q", got.Description, "Go to market")
				}
				if got.Priority != string(vo.TODO_PRIORITY_HIGH) {
					t.Fatalf("priority = %q, want %q", got.Priority, vo.TODO_PRIORITY_HIGH)
				}
				if got.Status != string(vo.TODO_STATUS_PENDING) {
					t.Fatalf("status = %q, want %q", got.Status, vo.TODO_STATUS_PENDING)
				}
				if got.DueDate == nil || !got.DueDate.Equal(dueDate) {
					t.Fatalf("due_date = %v, want %v", got.DueDate, dueDate)
				}
			},
		},
		{
			name: "invalid input - empty title",
			input: &input.CreateTodoInput{
				UserID:      7,
				Title:       "",
				Description: "Go to market",
				Priority:    "high",
			},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().CreateTodo(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInvalidParameter,
		},
		{
			name: "invalid input - invalid priority",
			input: &input.CreateTodoInput{
				UserID:      7,
				Title:       "Buy milk",
				Description: "Go to market",
				Priority:    "urgent",
			},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().CreateTodo(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInvalidParameter,
		},
		{
			name: "gateway error",
			input: &input.CreateTodoInput{
				UserID:      7,
				Title:       "Buy milk",
				Description: "Go to market",
				Priority:    "high",
			},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().
					CreateTodo(gomock.Any(), gomock.Any()).
					Return(apperrors.NewAppError(apperrors.ReasonInternalServerError, "create failed"))
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := gatewaymock.NewMockTodoCommandGateway(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			svc := NewTodoCreater(repo)
			got, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				if got != nil {
					t.Fatalf("Create() output = %#v, want nil", got)
				}
				assertAppErrorReason(t, err, tt.wantReason)
				return
			}

			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
