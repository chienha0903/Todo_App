package redisstore

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/chienha0903/Todo_App/services/users/internal/config"
)

func NewClient(cfg *config.Config) (*redis.Client, func(), error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, nil, fmt.Errorf("redis: parse URL: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	cleanup := func() { _ = client.Close() }
	return client, cleanup, nil
}
