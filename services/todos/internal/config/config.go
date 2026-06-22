package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const minJWTSecretLen = 32

type Config struct {
	AppName         string
	AppPort         string
	AppEnv          string
	DBDSN           string
	JWTSecret       string
	RabbitMQURL     string
	EnableDebugRace bool
	DebugPort       string
}

func Load() (*Config, error) {
	// Load .env if exists. Ignore error to allow pure environment-based config.
	_ = godotenv.Load()

	cfg := &Config{
		AppName:         getenv("APP_NAME", "todo-app"),
		AppPort:         getenv("APP_TODO_PORT", "50051"),
		AppEnv:          getenv("APP_ENV", "development"),
		DBDSN:           getenv("DB_TODO_DSN", "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"),
		JWTSecret:       getenv("JWT_SECRET", ""),
		RabbitMQURL:     getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		EnableDebugRace: getenv("ENABLE_DEBUG_RACE", "") == "true",
		DebugPort:       getenv("DEBUG_PORT", "8081"),
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
