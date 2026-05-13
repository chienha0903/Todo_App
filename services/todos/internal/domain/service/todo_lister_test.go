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
		check      func(t *testing.T, got *output.TodoPage)
	}{
		{
			name:  "success - returns todos",
			input: &input.ListTodosInput{UserID: 7, Page: 1, PageSize: 20},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				todos := []*entity.Todo{newFixtureTodo()}
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7), int32(1), int32(20)).
					Return(todos, int64(1), nil)
			},
			check: func(t *testing.T, got *output.TodoPage) {
				if len(got.Items) != 1 {
					t.Fatalf("len = %d, want 1", len(got.Items))
				}
				if got.Items[0].ID != 5 || got.Items[0].UserID != 7 {
					t.Fatalf("id/user_id = %d/%d, want 5/7", got.Items[0].ID, got.Items[0].UserID)
				}
				if got.Items[0].Title != "Buy milk" {
					t.Fatalf("title = %q, want %q", got.Items[0].Title, "Buy milk")
				}
				if got.Total != 1 {
					t.Fatalf("total = %d, want 1", got.Total)
				}
				if got.Page != 1 {
					t.Fatalf("page = %d, want 1", got.Page)
				}
				if got.PageSize != 20 {
					t.Fatalf("pageSize = %d, want 20", got.PageSize)
				}
			},
		},
		{
			name:  "success - empty list",
			input: &input.ListTodosInput{UserID: 7, Page: 1, PageSize: 20},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7), int32(1), int32(20)).
					Return([]*entity.Todo{}, int64(0), nil)
			},
			check: func(t *testing.T, got *output.TodoPage) {
				if len(got.Items) != 0 {
					t.Fatalf("len = %d, want 0", len(got.Items))
				}
				if got.Total != 0 {
					t.Fatalf("total = %d, want 0", got.Total)
				}
			},
		},
		{
			name:  "fallback to defaults when page/pageSize <= 0",
			input: &input.ListTodosInput{UserID: 7, Page: 0, PageSize: 0},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7), int32(1), int32(20)).
					Return([]*entity.Todo{}, int64(0), nil)
			},
			check: func(t *testing.T, got *output.TodoPage) {
				if got.Page != 1 {
					t.Fatalf("page = %d, want 1 (default)", got.Page)
				}
				if got.PageSize != 20 {
					t.Fatalf("pageSize = %d, want 20 (default)", got.PageSize)
				}
			},
		},
		{
			name:  "gateway error",
			input: &input.ListTodosInput{UserID: 7, Page: 1, PageSize: 20},
			setupMock: func(repo *gatewaymock.MockTodoQueryGateway) {
				repo.EXPECT().
					GetTodos(gomock.Any(), entity.UserID(7), int32(1), int32(20)).
					Return(nil, int64(0), apperrors.NewAppError(apperrors.ReasonInternalServerError, "db error"))
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
