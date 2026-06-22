package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/di"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/service"
	debughandler "github.com/chienha0903/Todo_App/services/todos/internal/handler/debug"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
	"github.com/chienha0903/Todo_App/services/todos/internal/observability/tracing"
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

	shutdownTracer, err := tracing.Init(context.Background(), cfg.AppName, "", cfg.AppEnv)
	if err != nil {
		slog.Warn("tracing init failed, continuing without tracing", "error", err)
		shutdownTracer = func(context.Context) error { return nil }
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := shutdownTracer(ctx); shutdownErr != nil {
			slog.Error("tracing shutdown error", "error", shutdownErr)
		}
	}()

	if err := datastore.RunMigrations(cfg.DBDSN); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	srv, cleanup, err := di.InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("init grpc server: %w", err)
	}
	defer cleanup()

	if cfg.EnableDebugRace {
		if err := startDebugServer(cfg); err != nil {
			return fmt.Errorf("start debug server: %w", err)
		}
	}

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

// startDebugServer khởi tạo DB connection riêng + HTTP server riêng cho race demo.
// Hoàn toàn tách biệt khỏi gRPC server và connection pool production.
func startDebugServer(cfg *config.Config) error {
	db, dbCleanup, err := datastore.NewDB(cfg)
	if err != nil {
		return fmt.Errorf("debug db: %w", err)
	}

	transactor := datastore.NewGormTransactor(db)
	repo := datastore.NewRaceDemoRepo(db)
	gw := datastore.NewRaceDemoGateway(repo)
	debugger := service.NewRaceDebugger(gw, transactor)

	mux := http.NewServeMux()
	debughandler.NewRaceHandler(debugger).RegisterRoutes(mux)

	go func() {
		addr := ":" + cfg.DebugPort
		slog.Info("debug server started", "component", "debug_http", "port", cfg.DebugPort)
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			slog.Error("debug server error", "error", err)
		}
		dbCleanup()
	}()

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
