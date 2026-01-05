package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSessionStore implements SessionStore using Redis.
type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client interface{}) *RedisSessionStore {
	if c, ok := client.(*redis.Client); ok {
		return &RedisSessionStore{client: c}
	}
	panic(fmt.Sprintf("invalid redis client type: %T, expected *redis.Client (v9)", client))
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
