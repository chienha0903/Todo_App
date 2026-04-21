package grpc

import (
	"context"

	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/mapper"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type todoHandler struct {
	todopb.UnimplementedTodoServiceServer
	uc todousecase.TodoUsecase
}

func NewTodoHandler(uc todousecase.TodoUsecase) *todoHandler {
	return &todoHandler{uc: uc}
}

func (h *todoHandler) CreateTodo(ctx context.Context, req *todopb.CreateTodoRequest) (*todopb.CreateTodoResponse, error) {
	in := mapper.ToCreateTodoInput(req)

	out, err := h.uc.CreateTodo(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &todopb.CreateTodoResponse{
		Todo: mapper.ToProtoTodo(out),
	}, nil
}

func toGRPCError(err error) error {
	// Keep error mapping simple: return as InvalidArgument for now.
	// Extend with errors.Error type check when needed.
	return status.Error(codes.InvalidArgument, err.Error())
}
