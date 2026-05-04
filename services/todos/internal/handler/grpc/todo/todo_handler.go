package todo

import (
	"context"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/mapper"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
)

type TodoHandler struct {
	todopb.UnimplementedTodoServiceServer
	creator todousecase.TodoCreator
	getter  todousecase.TodoGetter
	lister  todousecase.TodoLister
	updater todousecase.TodoUpdater
	deleter todousecase.TodoDeleter
}

func NewTodoHandler(
	creator todousecase.TodoCreator,
	getter todousecase.TodoGetter,
	lister todousecase.TodoLister,
	updater todousecase.TodoUpdater,
	deleter todousecase.TodoDeleter,
) *TodoHandler {
	return &TodoHandler{
		creator: creator,
		getter:  getter,
		lister:  lister,
		updater: updater,
		deleter: deleter,
	}
}

func (h *TodoHandler) CreateTodo(
	ctx context.Context,
	req *todopb.CreateTodoRequest,
) (*todopb.CreateTodoResponse, error) {
	in, err := mapper.ToCreateTodoInput(req)
	if err != nil {
		return nil, toGRPCError(err)
	}
	out, err := h.creator.Create(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &todopb.CreateTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) GetTodo(
	ctx context.Context,
	req *todopb.GetTodoRequest,
) (*todopb.GetTodoResponse, error) {
	out, err := h.getter.Get(ctx, mapper.ToGetTodoInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &todopb.GetTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) ListTodos(
	ctx context.Context,
	req *todopb.ListTodosRequest,
) (*todopb.ListTodosResponse, error) {
	todos, err := h.lister.List(ctx, mapper.ToListTodosInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &todopb.ListTodosResponse{Todos: mapper.ToProtoTodos(todos)}, nil
}

func (h *TodoHandler) UpdateTodo(
	ctx context.Context,
	req *todopb.UpdateTodoRequest,
) (*todopb.UpdateTodoResponse, error) {
	in, err := mapper.ToUpdateTodoInput(req)
	if err != nil {
		return nil, toGRPCError(err)
	}
	out, err := h.updater.Update(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &todopb.UpdateTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) DeleteTodo(
	ctx context.Context,
	req *todopb.DeleteTodoRequest,
) (*todopb.DeleteTodoResponse, error) {
	if err := h.deleter.Delete(ctx, mapper.ToDeleteTodoInput(req)); err != nil {
		return nil, toGRPCError(err)
	}
	return &todopb.DeleteTodoResponse{}, nil
}
