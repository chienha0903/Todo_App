package todo

import (
	"context"
	"errors"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func NewGRPCConn(cfg *config.Config) (*grpc.ClientConn, func(), error) {
	conn, err := grpc.NewClient(
		cfg.TodosGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}
	return conn, func() { _ = conn.Close() }, nil
}

func NewTodoServiceClient(conn *grpc.ClientConn) todopb.TodoServiceClient {
	return todopb.NewTodoServiceClient(conn)
}

type grpcGateway struct {
	client todopb.TodoServiceClient
}

func NewGRPCGateway(client todopb.TodoServiceClient) gateway.TodoGateway {
	return &grpcGateway{client: client}
}

func (g *grpcGateway) CreateTodo(ctx context.Context, in input.CreateTodo) (*output.Todo, error) {
	resp, err := g.client.CreateTodo(ctx, &todopb.CreateTodoRequest{
		UserId:      in.UserID,
		Title:       in.Title,
		Description: in.Description,
		Priority:    in.Priority,
		DueDate:     in.DueDate,
	})
	if err != nil {
		return nil, toAppError(err)
	}
	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) GetTodo(ctx context.Context, in input.GetTodo) (*output.Todo, error) {
	resp, err := g.client.GetTodo(ctx, &todopb.GetTodoRequest{Id: in.ID})
	if err != nil {
		return nil, toAppError(err)
	}
	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) ListTodos(ctx context.Context, in input.ListTodos) ([]*output.Todo, error) {
	resp, err := g.client.ListTodos(ctx, &todopb.ListTodosRequest{UserId: in.UserID})
	if err != nil {
		return nil, toAppError(err)
	}
	return toOutputs(resp.GetTodos()), nil
}

func (g *grpcGateway) UpdateTodo(ctx context.Context, in input.UpdateTodo) (*output.Todo, error) {
	resp, err := g.client.UpdateTodo(ctx, &todopb.UpdateTodoRequest{
		Id:          in.ID,
		Title:       in.Title,
		Description: in.Description,
		Priority:    in.Priority,
		Status:      in.Status,
		DueDate:     in.DueDate,
	})
	if err != nil {
		return nil, toAppError(err)
	}
	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) DeleteTodo(ctx context.Context, in input.DeleteTodo) error {
	_, err := g.client.DeleteTodo(ctx, &todopb.DeleteTodoRequest{Id: in.ID})
	if err != nil {
		return toAppError(err)
	}
	return nil
}

func toAppError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.Timeout()
	}
	st, ok := status.FromError(err)
	if !ok {
		return apperror.Internal()
	}
	switch st.Code() {
	case codes.NotFound:
		return apperror.NotFound(st.Message())
	case codes.InvalidArgument:
		return apperror.InvalidArgument(st.Message())
	case codes.Unauthenticated:
		return apperror.Unauthorized()
	case codes.PermissionDenied:
		return apperror.PermissionDenied()
	case codes.DeadlineExceeded:
		return apperror.Timeout()
	case codes.Unavailable:
		return apperror.Unavailable()
	default:
		return apperror.Internal()
	}
}

func toOutput(t *todopb.Todo) *output.Todo {
	if t == nil {
		return nil
	}
	return &output.Todo{
		ID:          t.Id,
		UserID:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toOutputs(todos []*todopb.Todo) []*output.Todo {
	items := make([]*output.Todo, 0, len(todos))
	for _, t := range todos {
		items = append(items, toOutput(t))
	}
	return items
}
