package cache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisCache_NilClientReturnsError(t *testing.T) {
	c := NewRedisCache(nil, "test:", time.Minute)

	if _, err := c.Get(context.Background(), "k"); err == nil {
		t.Fatal("expected get error when client is nil")
	}
	if err := c.Set(context.Background(), "k", "v", time.Second); err == nil {
		t.Fatal("expected set error when client is nil")
	}
	if err := c.Delete(context.Background(), "k"); err == nil {
		t.Fatal("expected delete error when client is nil")
	}
}

func TestRedisCache_GetReturnsCacheMiss(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() { _ = client.Close() })

	c := NewRedisCache(client, "test:", time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "missing")
	if err == nil {
		t.Fatal("expected get error")
	}
	if errors.Is(err, ErrCacheMiss) {
		return
	}
	if !strings.Contains(err.Error(), "redis get") {
		t.Fatalf("expected wrapped redis get error, got %v", err)
	}
}

func TestRedisCache_SetAndDeleteReturnErrorWhenRedisUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() { _ = client.Close() })

	c := NewRedisCache(client, "test:", time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := c.Set(ctx, "k", "v", 0); err == nil {
		t.Fatal("expected set error when redis is unavailable")
	}
	if err := c.Delete(ctx, "k"); err == nil {
		t.Fatal("expected delete error when redis is unavailable")
	}
}
