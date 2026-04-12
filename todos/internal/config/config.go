package config

import "os"

type Config struct {
	AppName string
	AppPort string
	AppEnv  string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppName: getenv("APP_NAME", "todo-app"),
		AppPort: getenv("APP_PORT", "8080"),
		AppEnv:  getenv("APP_ENV", "development"),
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
