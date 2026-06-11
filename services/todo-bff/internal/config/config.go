package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretLen = 32

type Config struct {
	AppName        string
	AppPort        string
	AppEnv         string
	TodosGRPCAddr  string
	UsersGRPCAddr  string
	JWTSecret      string
	RequestTimeout time.Duration
}

func Load() (*Config, error) {
	// Load .env if exists. Ignore error to allow pure environment-based config.
	_ = godotenv.Load()

	requestTimeout, err := time.ParseDuration(getenv("REQUEST_TIMEOUT", "5s"))
	if err != nil {
		return nil, fmt.Errorf("config: parse REQUEST_TIMEOUT: %w", err)
	}

	cfg := &Config{
		AppName:        getenv("APP_NAME", "todo-bff"),
		AppPort:        getenv("BFF_PORT", getenv("APP_PORT", "8080")),
		AppEnv:         getenv("APP_ENV", "development"),
		TodosGRPCAddr:  getenv("TODOS_GRPC_ADDR", "localhost:50051"),
		UsersGRPCAddr:  getenv("USERS_GRPC_ADDR", "localhost:50052"),
		JWTSecret:      getenv("JWT_SECRET", ""),
		RequestTimeout: requestTimeout,
	}

	if len(cfg.JWTSecret) < minJWTSecretLen {
		return nil, errors.New("config: JWT_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
