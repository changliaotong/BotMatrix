package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisSessionStore implements SessionStore using Redis.
type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func (s *RedisSessionStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, expiration).Err()
}

func (s *RedisSessionStore) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (s *RedisSessionStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *RedisSessionStore) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}
