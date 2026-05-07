package service

import (
	"context"
	"testing"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	gatewaymock "github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway/mock"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
	"go.uber.org/mock/gomock"
)

func TestTodoUpdaterUpdate(t *testing.T) {
	tests := []struct {
		name       string
		input      *input.UpdateTodoInput
		setupMock  func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway)
		wantErr    bool
		wantReason apperrors.Reason
		check      func(t *testing.T, got *output.TodoUpdater)
	}{
		{
			name: "success - update title and priority",
			input: &input.UpdateTodoInput{
				ID:       5,
				Title:    "Read book",
				Priority: "high",
			},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().GetTodo(gomock.Any(), entity.TodoID(5)).Return(newFixtureTodo(), nil)
				cmdRepo.EXPECT().UpdateTodo(gomock.Any(), gomock.Any()).Return(nil)
			},
			check: func(t *testing.T, got *output.TodoUpdater) {
				if got.Title != "Read book" {
					t.Fatalf("title = %q, want %q", got.Title, "Read book")
				}
				if got.Priority != string(vo.TODO_PRIORITY_HIGH) {
					t.Fatalf("priority = %q, want %q", got.Priority, vo.TODO_PRIORITY_HIGH)
				}
			},
		},
		{
			name: "success - update status",
			input: &input.UpdateTodoInput{
				ID:     5,
				Status: "completed",
			},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().GetTodo(gomock.Any(), entity.TodoID(5)).Return(newFixtureTodo(), nil)
				cmdRepo.EXPECT().UpdateTodo(gomock.Any(), gomock.Any()).Return(nil)
			},
			check: func(t *testing.T, got *output.TodoUpdater) {
				if got.Status != string(vo.TODO_STATUS_COMPLETED) {
					t.Fatalf("status = %q, want %q", got.Status, vo.TODO_STATUS_COMPLETED)
				}
			},
		},
		{
			name:  "todo not found",
			input: &input.UpdateTodoInput{ID: 99, Title: "Read book"},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().
					GetTodo(gomock.Any(), entity.TodoID(99)).
					Return(nil, apperrors.NewAppError(apperrors.ReasonNotFound, "todo not found"))
				cmdRepo.EXPECT().UpdateTodo(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr:    true,
			wantReason: apperrors.ReasonNotFound,
		},
		{
			name:  "invalid input - invalid status",
			input: &input.UpdateTodoInput{ID: 5, Status: "INVALID_STATUS"},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().GetTodo(gomock.Any(), entity.TodoID(5)).Return(newFixtureTodo(), nil)
				cmdRepo.EXPECT().UpdateTodo(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInvalidParameter,
		},
		{
			name:  "invalid input - invalid priority",
			input: &input.UpdateTodoInput{ID: 5, Priority: "urgent"},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().GetTodo(gomock.Any(), entity.TodoID(5)).Return(newFixtureTodo(), nil)
				cmdRepo.EXPECT().UpdateTodo(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInvalidParameter,
		},
		{
			name:  "update gateway error",
			input: &input.UpdateTodoInput{ID: 5, Title: "Read book"},
			setupMock: func(cmdRepo *gatewaymock.MockTodoCommandGateway, qryRepo *gatewaymock.MockTodoQueryGateway) {
				qryRepo.EXPECT().GetTodo(gomock.Any(), entity.TodoID(5)).Return(newFixtureTodo(), nil)
				cmdRepo.EXPECT().
					UpdateTodo(gomock.Any(), gomock.Any()).
					Return(apperrors.NewAppError(apperrors.ReasonInternalServerError, "update failed"))
			},
			wantErr:    true,
			wantReason: apperrors.ReasonInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdRepo := gatewaymock.NewMockTodoCommandGateway(ctrl)
			qryRepo := gatewaymock.NewMockTodoQueryGateway(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(cmdRepo, qryRepo)
			}

			svc := NewTodoUpdater(cmdRepo, qryRepo)
			got, err := svc.Update(context.Background(), tt.input)

			if tt.wantErr {
				if got != nil {
					t.Fatalf("Update() output = %#v, want nil", got)
				}
				assertAppErrorReason(t, err, tt.wantReason)
				return
			}

			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
