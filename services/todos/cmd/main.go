package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/di"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
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

	if err := datastore.RunMigrations(cfg.DBDSN); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	srv, cleanup, err := di.InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("init grpc server: %w", err)
	}
	defer cleanup()

	lis, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		return fmt.Errorf("listen grpc server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logGRPCServerStarted(cfg)
		if err := srv.Serve(lis); err != nil {
			errCh <- fmt.Errorf("serve grpc server: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		stop()
		slog.Info("server stopping", "component", "grpc_server", "event", "server_stopping")
		srv.GracefulStop()
		slog.Info("server stopped", "component", "grpc_server", "event", "server_stopped")
		return nil
	case err := <-errCh:
		return err
	}
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
