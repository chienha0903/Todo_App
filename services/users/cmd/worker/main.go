package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chienha0903/Todo_App/pkg/rabbitmq"
	"github.com/chienha0903/Todo_App/services/users/internal/config"
	"github.com/chienha0903/Todo_App/services/users/internal/infra/datastore"
	"github.com/chienha0903/Todo_App/services/users/internal/worker"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("worker failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, cleanup, err := datastore.NewDB(cfg)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer cleanup()

	conn, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer conn.Close()

	outboxRepo := datastore.NewOutboxRepo(db)
	outboxQueryGW := datastore.NewOutboxQueryGateway(outboxRepo)
	outboxCmdGW := datastore.NewOutboxCommandGateway(outboxRepo)
	publisher := worker.NewOutboxPublisher(outboxQueryGW, outboxCmdGW, conn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	slog.Info("outbox worker started", "poll_interval", "5s")

	if err := publisher.PublishBatch(ctx); err != nil {
		slog.Error("initial publish batch failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox worker shutting down")
			return nil
		case <-ticker.C:
			if err := publisher.PublishBatch(ctx); err != nil {
				slog.Error("publish batch failed", "error", err)
			}
		}
	}
}
