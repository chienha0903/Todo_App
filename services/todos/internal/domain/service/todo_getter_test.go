package service

import (
	"context"
	"testing"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	gatewaymock "github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway/mock"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
	"go.uber.org/mock/gomock"
)

func TestTodoGetterGet(t *testing.T) {
	tests := []struct {
		name       string
		input      *input.GetTodoInput
		setupMock  func(repo *gatewaymock.MockTodoQueryGateway)
		wantErr    bool
		wantReason apperrors.Reason
		check      func(t *testing.T, got *output.TodoGetter)
	}{
		{
			name:  "success",
			input: &input.GetTodoInput{ID: 5},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodo(gomock.Any(), entity.TodoID(5)).
					Return(newFixtureTodo(), nil)
			},
			check: func(t *testing.T, got *output.TodoGetter) {
				if got.ID != 5 || got.UserID != 7 {
					t.Fatalf("id/user_id = %d/%d, want 5/7", got.ID, got.UserID)
				}
				if got.Title != "Buy milk" {
					t.Fatalf("title = %q, want %q", got.Title, "Buy milk")
				}
				if got.Description != "Go to market" {
					t.Fatalf("description = %q, want %q", got.Description, "Go to market")
				}
			},
		},
		{
			name:  "not found",
			input: &input.GetTodoInput{ID: 99},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodo(gomock.Any(), entity.TodoID(99)).
					Return(nil, apperrors.NewAppError(apperrors.ReasonNotFound, "todo not found"))
			},
			wantErr:    true,
			wantReason: apperrors.ReasonNotFound,
		},
		{
			name:  "gateway internal error",
			input: &input.GetTodoInput{ID: 5},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodo(gomock.Any(), entity.TodoID(5)).
					Return(nil, apperrors.NewAppError(apperrors.ReasonInternalServerError, "db error"))
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := gatewaymock.NewMockTodoQueryGateway(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			svc := NewTodoGetter(repo)
			got, err := svc.Get(context.Background(), tt.input)

			if tt.wantErr {
				if got != nil {
					t.Fatalf("Get() output = %#v, want nil", got)
				}
				assertAppErrorReason(t, err, tt.wantReason)
				return
			}

			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
