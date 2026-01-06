package database

import (
	"context"
	"fmt"
	"time"

	"BotMatrix/common/config"

	"github.com/redis/go-redis/v9"
)

// RedisManager handles Redis connections
type RedisManager struct {
	Client *redis.Client
}

// NewRedisManager creates a new Redis manager
func NewRedisManager(cfg *config.AppConfig) (*RedisManager, error) {
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("redis address not configured")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd, // no password set
		DB:       0,            // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	return &RedisManager{
		Client: rdb,
	}, nil
}

// Set stores a value in Redis
func (r *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisManager) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del deletes a value from Redis
func (r *RedisManager) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
