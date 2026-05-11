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
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/handler/server"
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

	gqlResolver, cleanup, err := di.InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("init app: %w", err)
	}
	defer cleanup()

	httpServer := server.NewHTTPServer(cfg, gqlResolver)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := startHTTPServer(httpServer, cfg)

	select {
	case <-ctx.Done():
		stop()
		return shutdownHTTPServer(httpServer)
	case err := <-serveErr:
		if err != nil {
			return err
		}
		return nil
	}
}

func startHTTPServer(httpServer *nethttp.Server, cfg *config.Config) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		logHTTPServerStarted(cfg)

		err := httpServer.ListenAndServe()
		if err != nil && !stderrors.Is(err, nethttp.ErrServerClosed) {
			errCh <- fmt.Errorf("serve http server: %w", err)
			return
		}
		errCh <- nil
	}()

	return errCh
}

func shutdownHTTPServer(httpServer *nethttp.Server) error {
	slog.Info("server stopping", "component", "http_server", "event", "server_stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
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
