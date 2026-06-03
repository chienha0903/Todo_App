package user

import (
	"context"
	"fmt"

	userpb "github.com/chienha0903/Todo_App/proto/user"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/domain/gateway"
	userinput "github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/input"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/usecase/user/output"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConn grpc.ClientConn

func NewGRPCConn(cfg *config.Config) (*ClientConn, func(), error) {
	conn, err := grpc.NewClient(
		cfg.UsersGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("grpc dial %s: %w", cfg.UsersGRPCAddr, err)
	}

	return (*ClientConn)(conn), func() { _ = conn.Close() }, nil
}

func NewUserServiceClient(conn *ClientConn) userpb.UserserviceClient {
	return userpb.NewUserserviceClient((*grpc.ClientConn)(conn))
}

type grpcGateway struct {
	client userpb.UserserviceClient
}

func NewGRPCGateway(client userpb.UserserviceClient) gateway.UserGateway {
	return &grpcGateway{client: client}
}

func (g *grpcGateway) Login(ctx context.Context, email, password string) (*gateway.AuthTokens, error) {
	resp, err := g.client.Login(ctx, &userpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return &gateway.AuthTokens{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (g *grpcGateway) RefreshToken(ctx context.Context, in *userinput.RefreshToken) (*gateway.AuthTokens, error) {
	resp, err := g.client.RefreshToken(ctx, &userpb.RefreshTokenRequest{
		RefreshToken: in.RefreshToken,
	})

	if err != nil {
		return nil, err
	}

	return &gateway.AuthTokens{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (g *grpcGateway) ChangePassword(ctx context.Context, in *userinput.ChangePassword) error {
	_, err := g.client.ChangePassword(ctx, &userpb.ChangePasswordRequest{
		UserId:          in.UserID,
		CurrentPassword: in.CurrentPassword,
		NewPassword:     in.NewPassword,
	})
	return err
}

func (g *grpcGateway) GetUser(ctx context.Context, in *userinput.GetUser) (*output.User, error) {
	resp, err := g.client.GetUser(ctx, &userpb.GetUserRequest{Id: in.ID})
	if err != nil {
		return nil, err
	}
	return toOutput(resp.GetUser()), nil
}

func (g *grpcGateway) GetUsersByIDs(ctx context.Context, in *userinput.GetUsersByIDs) ([]*output.User, error) {
	resp, err := g.client.GetUsersByIDs(ctx, &userpb.GetUsersByIDsRequest{Ids: in.IDs})
	if err != nil {
		return nil, err
	}
	return toOutputs(resp.GetUsers()), nil
}

func (g *grpcGateway) ListUsers(ctx context.Context, in *userinput.ListUsers) (*output.UserPage, error) {
	resp, err := g.client.ListUsers(ctx, &userpb.ListUsersRequest{
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &output.UserPage{
		Items:    toOutputs(resp.GetUsers()),
		Total:    resp.Total,
		Page:     resp.Page,
		PageSize: resp.PageSize,
	}, nil
}

func (g *grpcGateway) UpdateUser(ctx context.Context, in *userinput.UpdateUser) (*output.User, error) {
	resp, err := g.client.UpdateUser(ctx, &userpb.UpdateUserRequest{
		Id:       in.ID,
		Email:    in.Email,
		Username: in.Username,
		Password: in.Password,
		Role:     in.Role,
	})
	if err != nil {
		return nil, err
	}
	return toOutput(resp.GetUser()), nil
}

func (g *grpcGateway) DeleteUser(ctx context.Context, in *userinput.DeleteUser) error {
	_, err := g.client.DeleteUser(ctx, &userpb.DeleteUserRequest{Id: in.ID})
	return err
}

func toOutput(u *userpb.User) *output.User {
	if u == nil {
		return nil
	}
	return &output.User{
		ID:        u.Id,
		Email:     u.Email,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toOutputs(users []*userpb.User) []*output.User {
	items := make([]*output.User, 0, len(users))
	for _, u := range users {
		items = append(items, toOutput(u))
	}
	return items
}
