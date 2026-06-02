package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName     string
	AppPort     string
	AppEnv      string
	DBDSN       string
	JWTSecret   string
	RabbitMQURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:     getenv("APP_NAME", "todo-app"),
		AppPort:     getenv("APP_USERS_PORT", "50052"),
		AppEnv:      getenv("APP_ENV", "development"),
		DBDSN:       getenv("DB_USERS_DSN", "postgres://postgres:postgres@localhost:5432/users_db?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "chien-apvn"),
		RabbitMQURL: getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
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
