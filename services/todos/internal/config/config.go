package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string
	DBDSN   string
}

func Load() (*Config, error) {
	// Load .env if exists. Ignore error to allow pure environment-based config.
	_ = godotenv.Load()

	cfg := &Config{
		AppName: getenv("APP_NAME", "todo-app"),
		AppPort: getenv("APP_PORT", "50051"),
		AppEnv:  getenv("APP_ENV", "development"),
		DBDSN:   getenv("DB_DSN", "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"),
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
