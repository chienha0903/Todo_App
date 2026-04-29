package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/di"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "component", "app", "event", "load_config_failed", "error", err)
		os.Exit(1)
	}

	srv, err := di.InitGRPCServer(cfg)
	if err != nil {
		slog.Error(
			"init server failed",
			"component", "grpc_server",
			"event", "init_server_failed",
			"app", cfg.AppName,
			"env", cfg.AppEnv,
			"error", err,
		)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		slog.Error(
			"listen failed",
			"component", "grpc_server",
			"event", "listen_failed",
			"app", cfg.AppName,
			"port", cfg.AppPort,
			"env", cfg.AppEnv,
			"error", err,
		)
		os.Exit(1)
	}

	slog.Info(
		"server started",
		"component", "grpc_server",
		"event", "server_started",
		"app", cfg.AppName,
		"port", cfg.AppPort,
		"env", cfg.AppEnv,
	)
	if err := srv.Serve(lis); err != nil {
		slog.Error(
			"serve failed",
			"component", "grpc_server",
			"event", "serve_failed",
			"app", cfg.AppName,
			"port", cfg.AppPort,
			"env", cfg.AppEnv,
			"error", err,
		)
		os.Exit(1)
	}
}
