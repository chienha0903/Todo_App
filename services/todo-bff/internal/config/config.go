package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName        string
	AppPort        string
	AppEnv         string
	TodosGRPCAddr  string
	RequestTimeout time.Duration
}

func Load() (*Config, error) {
	// Load .env if exists. Ignore error to allow pure environment-based config.
	_ = godotenv.Load()

	requestTimeout, err := time.ParseDuration(getenv("REQUEST_TIMEOUT", "5s"))
	if err != nil {
		return nil, fmt.Errorf("config: parse REQUEST_TIMEOUT: %w", err)
	}

	return &Config{
		AppName:        getenv("APP_NAME", "todo-bff"),
		AppPort:        getenv("BFF_PORT", getenv("APP_PORT", "8080")),
		AppEnv:         getenv("APP_ENV", "development"),
		TodosGRPCAddr:  getenv("TODOS_GRPC_ADDR", "localhost:50051"),
		RequestTimeout: requestTimeout,
	}, nil
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
