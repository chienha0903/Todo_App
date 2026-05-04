package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/di"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("app failed", "component", "grpc_server", "event", "app_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	srv, err := di.InitGRPCServer(cfg)
	if err != nil {
		return fmt.Errorf("init grpc server: %w", err)
	}

	lis, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		return fmt.Errorf("listen grpc server: %w", err)
	}

	logGRPCServerStarted(cfg)

	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc server: %w", err)
	}

	return nil
}

func logGRPCServerStarted(cfg *config.Config) {
	slog.Info(
		"server started",
		"component", "grpc_server",
		"event", "server_started",
		"app", cfg.AppName,
		"port", cfg.AppPort,
		"env", cfg.AppEnv,
	)
}
