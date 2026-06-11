package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const minJWTSecretLen = 32

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
		AppName:     getenv("APP_NAME", "users"),
		AppPort:     getenv("APP_USERS_PORT", "50052"),
		AppEnv:      getenv("APP_ENV", "development"),
		DBDSN:       getenv("DB_USERS_DSN", "postgres://postgres:postgres@localhost:5433/user_db?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", ""),
		RabbitMQURL: getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
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
