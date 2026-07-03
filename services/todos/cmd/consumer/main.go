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
	"github.com/chienha0903/Todo_App/services/todos/internal/config"
	"github.com/chienha0903/Todo_App/services/todos/internal/consumer"
	"github.com/chienha0903/Todo_App/services/todos/internal/infra/datastore"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("consumer failed", "error", err)
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

	todoRepo := datastore.NewTodoCommandRepo(db)
	todoCmd := datastore.NewTodoCommandGateway(todoRepo)
	processedGW := datastore.NewProcessedEventGateway(datastore.NewProcessedEventRepo(db))
	txGW := datastore.NewGormTransactor(db)

	mainConsumer := consumer.NewUserDeleteConsumer(todoCmd, processedGW, txGW, conn, cfg.ConsumerWorkerPoolSize)
	dlqProcessor := consumer.NewDLQProcessor(conn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting consumers")

	go runDLQProcessor(ctx, dlqProcessor)

	for {
		if err := mainConsumer.Start(ctx); err != nil {
			slog.Error("main consumer error", "error", err)
		}

		select {
		case <-ctx.Done():
			slog.Info("consumer shutdown complete")
			return nil
		case <-time.After(5 * time.Second):
			slog.Info("restarting main consumer...")
		}
	}
}

func runDLQProcessor(ctx context.Context, p *consumer.DLQProcessor) {
	for {
		if err := p.Start(ctx); err != nil {
			slog.Error("DLQ processor error", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			slog.Info("restarting DLQ processor...")
		}
	}
}
