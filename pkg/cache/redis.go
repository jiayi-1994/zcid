package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type RedisCache struct {
	client     *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

func NewRedisCache(client *redis.Client, keyPrefix string, defaultTTL time.Duration) *RedisCache {
	return &RedisCache{
		client:     client,
		keyPrefix:  strings.TrimSpace(keyPrefix),
		defaultTTL: defaultTTL,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", errors.New("redis client is nil")
	}

	fullKey := c.buildKey(key)
	value, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("redis get %s: %w", fullKey, err)
	}

	return value, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c.client == nil {
		return errors.New("redis client is nil")
	}

	fullKey := c.buildKey(key)
	effectiveTTL := ttl
	if effectiveTTL <= 0 {
		effectiveTTL = c.defaultTTL
	}

	if err := c.client.Set(ctx, fullKey, value, effectiveTTL).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", fullKey, err)
	}

	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return errors.New("redis client is nil")
	}

	fullKey := c.buildKey(key)
	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("redis delete %s: %w", fullKey, err)
	}

	return nil
}

func (c *RedisCache) buildKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return c.keyPrefix + ":" + key
}
