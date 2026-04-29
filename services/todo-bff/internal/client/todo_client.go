package client

import (
	"fmt"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TodoClient struct {
	conn *grpc.ClientConn
	api  todopb.TodoServiceClient
}

func NewTodoClient(cfg *config.Config) (*TodoClient, error) {
	conn, err := grpc.NewClient(
		cfg.TodosGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("todo client: create grpc client: %w", err)
	}

	return &TodoClient{
		conn: conn,
		api:  todopb.NewTodoServiceClient(conn),
	}, nil
}

func (c *TodoClient) Service() todopb.TodoServiceClient {
	return c.api
}

func (c *TodoClient) Close() error {
	return c.conn.Close()
}
