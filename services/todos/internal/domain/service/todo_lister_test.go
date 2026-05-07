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

func TestTodoListerList(t *testing.T) {
	tests := []struct {
		name       string
		input      *input.ListTodosInput
		setupMock  func(repo *gatewaymock.MockTodoQueryGateway)
		wantErr    bool
		wantReason apperrors.Reason
		check      func(t *testing.T, got output.TodoLister)
	}{
		{
			name:  "success - returns todos",
			input: &input.ListTodosInput{UserID: 7},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				todos := []*entity.Todo{newFixtureTodo()}
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7)).
					Return(todos, nil)
			},
			check: func(t *testing.T, got output.TodoLister) {
				if len(got) != 1 {
					t.Fatalf("len = %d, want 1", len(got))
				}
				if got[0].ID != 5 || got[0].UserID != 7 {
					t.Fatalf("id/user_id = %d/%d, want 5/7", got[0].ID, got[0].UserID)
				}
				if got[0].Title != "Buy milk" {
					t.Fatalf("title = %q, want %q", got[0].Title, "Buy milk")
				}
			},
		},
		{
			name:  "success - empty list",
			input: &input.ListTodosInput{UserID: 7},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7)).
					Return([]*entity.Todo{}, nil)
			},
			check: func(t *testing.T, got output.TodoLister) {
				if len(got) != 0 {
					t.Fatalf("len = %d, want 0", len(got))
				}
			},
		},
		{
			name:  "gateway error",
			input: &input.ListTodosInput{UserID: 7},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7)).
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

			svc := NewTodoLister(repo)
			got, err := svc.List(context.Background(), tt.input)

			if tt.wantErr {
				if got != nil {
					t.Fatalf("List() output = %#v, want nil", got)
				}
				assertAppErrorReason(t, err, tt.wantReason)
				return
			}

			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
