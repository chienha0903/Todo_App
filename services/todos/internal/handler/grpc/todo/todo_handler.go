package todo

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/caller"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/mapper"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
)

type TodoHandler struct {
	todopb.UnimplementedTodoServiceServer
	creater todousecase.TodoCreater
	getter  todousecase.TodoGetter
	lister  todousecase.TodoLister
	updater todousecase.TodoUpdater
	deleter todousecase.TodoDeleter
}

func NewTodoHandler(
	creater todousecase.TodoCreater,
	getter todousecase.TodoGetter,
	lister todousecase.TodoLister,
	updater todousecase.TodoUpdater,
	deleter todousecase.TodoDeleter,
) *TodoHandler {
	return &TodoHandler{
		creater: creater,
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
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" && req.UserId != c.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot create todo for another user")
	}

	in, err := mapper.ToCreateTodoInput(req)
	if err != nil {
		return nil, toGRPCError(err)
	}

	out, err := h.creater.Create(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &todopb.CreateTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) GetTodo(
	ctx context.Context,
	req *todopb.GetTodoRequest,
) (*todopb.GetTodoResponse, error) {
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	out, err := h.getter.Get(ctx, mapper.ToGetTodoInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	if c.Role != "ADMIN" && out.UserID != c.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot access another user's todo")
	}

	return &todopb.GetTodoResponse{Todo: mapper.ToProtoTodo(out)}, nil
}

func (h *TodoHandler) ListTodos(
	ctx context.Context,
	req *todopb.ListTodosRequest,
) (*todopb.ListTodosResponse, error) {
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" {
		req.UserId = c.UserID
	}

	page, err := h.lister.List(ctx, mapper.ToListTodosInput(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &todopb.ListTodosResponse{
		Todos:    mapper.ToProtoTodos(page.Items),
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}, nil
}

func (h *TodoHandler) UpdateTodo(
	ctx context.Context,
	req *todopb.UpdateTodoRequest,
) (*todopb.UpdateTodoResponse, error) {
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" {
		existing, err := h.getter.Get(ctx, mapper.ToGetTodoInput(&todopb.GetTodoRequest{Id: req.Id}))
		if err != nil {
			return nil, toGRPCError(err)
		}
		if existing.UserID != c.UserID {
			return nil, status.Error(codes.PermissionDenied, "cannot update another user's todo")
		}
	}

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
	c, ok := caller.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing caller identity")
	}

	if c.Role != "ADMIN" {
		existing, err := h.getter.Get(ctx, mapper.ToGetTodoInput(&todopb.GetTodoRequest{Id: req.Id}))
		if err != nil {
			return nil, toGRPCError(err)
		}
		if existing.UserID != c.UserID {
			return nil, status.Error(codes.PermissionDenied, "cannot delete another user's todo")
		}
	}

	if err := h.deleter.Delete(ctx, mapper.ToDeleteTodoInput(req)); err != nil {
		return nil, toGRPCError(err)
	}

	return &todopb.DeleteTodoResponse{}, nil
}
