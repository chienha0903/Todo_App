package service

import (
	"context"
	"testing"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	gatewaymock "github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway/mock"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"go.uber.org/mock/gomock"
)

func TestTodoDeleterDelete(t *testing.T) {
	tests := []struct {
		name       string
		input      *input.DeleteTodoInput
		setupMock  func(repo *gatewaymock.MockTodoCommandGateway)
		wantErr    bool
		wantReason apperrors.Reason
	}{
		{
			name:  "success",
			input: &input.DeleteTodoInput{ID: 5},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().
					DeleteTodo(gomock.Any(), entity.TodoID(5)).
					Return(nil)
			},
		},
		{
			name:  "not found",
			input: &input.DeleteTodoInput{ID: 99},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().
					DeleteTodo(gomock.Any(), entity.TodoID(99)).
					Return(apperrors.NewAppError(apperrors.ReasonNotFound, "todo not found"))
			},
			wantErr:    true,
			wantReason: apperrors.ReasonNotFound,
		},
		{
			name:  "gateway internal error",
			input: &input.DeleteTodoInput{ID: 5},
			setupMock: func(repo *gatewaymock.MockTodoCommandGateway) {
				repo.EXPECT().
					DeleteTodo(gomock.Any(), entity.TodoID(5)).
					Return(apperrors.NewAppError(apperrors.ReasonInternalServerError, "delete failed"))
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

			svc := NewTodoDeleter(repo)
			err := svc.Delete(context.Background(), tt.input)

			if tt.wantErr {
				assertAppErrorReason(t, err, tt.wantReason)
				return
			}

			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
		})
	}
}
