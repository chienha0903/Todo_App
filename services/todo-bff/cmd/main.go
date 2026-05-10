package main

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chienha0903/Todo_App/services/todo-bff/internal/config"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/di"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("app failed", "component", "bff", "event", "app_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	server, err := di.InitHTTPServer(cfg)
	if err != nil {
		return fmt.Errorf("init http server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := startHTTPServer(server, cfg)

	select {
	case <-ctx.Done():
		stop()
		return shutdownHTTPServer(server)
	case err := <-serveErr:
		if err != nil {
			return err
		}
		return nil
	}
}

func startHTTPServer(server *nethttp.Server, cfg *config.Config) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		logHTTPServerStarted(cfg)

		err := server.ListenAndServe()
		if err != nil && !stderrors.Is(err, nethttp.ErrServerClosed) {
			errCh <- fmt.Errorf("serve http server: %w", err)
			return
		}
		errCh <- nil
	}()

	return errCh
}

func shutdownHTTPServer(server *nethttp.Server) error {
	slog.Info("server stopping", "component", "http_server", "event", "server_stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	slog.Info("server stopped", "component", "http_server", "event", "server_stopped")
	return nil
}

func logHTTPServerStarted(cfg *config.Config) {
	slog.Info(
		"server started",
		"component", "http_server",
		"event", "server_started",
		"app", cfg.AppName,
		"port", cfg.AppPort,
		"env", cfg.AppEnv,
		"todos_grpc_addr", cfg.TodosGRPCAddr,
	)
}
