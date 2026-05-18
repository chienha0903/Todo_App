package todo

import (
	"context"
	"fmt"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/todo/output"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCConn(cfg *config.Config) (*grpc.ClientConn, func(), error) {
	conn, err := grpc.NewClient(
		cfg.TodosGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc dial %s: %w", cfg.TodosGRPCAddr, err)
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
		return nil, err
	}

	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) GetTodo(ctx context.Context, in input.GetTodo) (*output.Todo, error) {
	resp, err := g.client.GetTodo(ctx, &todopb.GetTodoRequest{Id: in.ID})
	if err != nil {
		return nil, err
	}

	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) ListTodos(ctx context.Context, in input.ListTodos) (*output.TodoPage, error) {
	resp, err := g.client.ListTodos(ctx, &todopb.ListTodosRequest{
		UserId:   in.UserID,
		Page:     int32(in.Page),
		PageSize: int32(in.PageSize),
	})
	if err != nil {
		return nil, err
	}

	total := int(resp.Total)
	page := int(resp.Page)
	pSize := int(resp.PageSize)

	return &output.TodoPage{
		Items:    toOutputs(resp.GetTodos()),
		Total:    total,
		Page:     page,
		PageSize: pSize,
		HasNext:  page*pSize < total,
	}, nil
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
		return nil, err
	}

	return toOutput(resp.GetTodo()), nil
}

func (g *grpcGateway) DeleteTodo(ctx context.Context, in input.DeleteTodo) error {
	_, err := g.client.DeleteTodo(ctx, &todopb.DeleteTodoRequest{Id: in.ID})
	return err
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
