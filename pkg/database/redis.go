package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/config"
)

// NewRedis initializes a go-redis client.
func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// PingRedis checks Redis connectivity for health checks.
func PingRedis(client *redis.Client) error {
	return client.Ping(context.Background()).Err()
}
