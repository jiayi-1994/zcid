package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides per-key sliding-window rate limiting backed by Redis,
// with an in-memory fallback when Redis is unavailable.
type RateLimiter struct {
	rdb    *redis.Client
	max    int
	window time.Duration

	mu       sync.Mutex
	counters map[string]*memEntry
}

type memEntry struct {
	tokens    int
	resetTime time.Time
}

func NewRateLimiter(rdb *redis.Client, max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:      rdb,
		max:      max,
		window:   window,
		counters: make(map[string]*memEntry),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	if rl.rdb != nil {
		return rl.allowRedis(key)
	}
	return rl.allowMemory(key)
}

func (rl *RateLimiter) allowRedis(key string) bool {
	ctx := context.Background()
	rKey := fmt.Sprintf("ratelimit:%s", key)
	pipe := rl.rdb.Pipeline()
	incr := pipe.Incr(ctx, rKey)
	pipe.Expire(ctx, rKey, rl.window)
	_, _ = pipe.Exec(ctx)
	return incr.Val() <= int64(rl.max)
}

func (rl *RateLimiter) allowMemory(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, ok := rl.counters[key]
	if !ok || now.After(entry.resetTime) {
		rl.counters[key] = &memEntry{tokens: 1, resetTime: now.Add(rl.window)}
		return true
	}
	entry.tokens++
	return entry.tokens <= rl.max
}

// RateLimit returns a Gin middleware that rate-limits requests by client IP.
func RateLimit(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}
