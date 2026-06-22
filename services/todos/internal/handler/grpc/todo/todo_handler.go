package todo

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/caller"
	"github.com/chienha0903/Todo_App/services/todos/internal/handler/grpc/mapper"
	"github.com/chienha0903/Todo_App/services/todos/internal/observability/tracing"
	todousecase "github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo"
)

// tracer is package-level and lazily delegates to the real TracerProvider once
// tracing.Init() sets it via otel.SetTracerProvider. Noop until then.
var tracer = otel.Tracer("todos/handler")

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
	ctx, hSpan := tracer.Start(ctx, "handler.CreateTodo",
		trace.WithAttributes(
			attribute.String("grpc.method", "CreateTodo"),
			attribute.Int64("user.id", req.UserId),
		))
	defer hSpan.End()

	c, ok := caller.FromContext(ctx)
	if !ok {
		err := status.Error(codes.Unauthenticated, "missing caller identity")
		tracing.RecordError(hSpan, err)
		return nil, err
	}

	if c.Role != "ADMIN" && req.UserId != c.UserID {
		err := status.Error(codes.PermissionDenied, "cannot create todo for another user")
		tracing.RecordError(hSpan, err)
		return nil, err
	}

	in, err := mapper.ToCreateTodoInput(req)
	if err != nil {
		tracing.RecordError(hSpan, err)
		return nil, toGRPCError(err)
	}

	ctx, ucSpan := tracer.Start(ctx, "usecase.CreateTodo",
		trace.WithAttributes(
			attribute.String("usecase.name", "CreateTodo"),
			attribute.Int64("user.id", req.UserId),
		))
	out, err := h.creater.Create(ctx, in)
	if err != nil {
		tracing.RecordError(ucSpan, err)
		ucSpan.End()
		tracing.RecordError(hSpan, err)
		return nil, toGRPCError(err)
	}
	ucSpan.SetStatus(otelcodes.Ok, "")
	ucSpan.End()

	hSpan.SetAttributes(attribute.Int64("todo.id", out.ID))
	hSpan.SetStatus(otelcodes.Ok, "")
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
