package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"botworker/internal/config"
)

// Client Redis客户端封装
type Client struct {
	*redis.Client
}

// NewClient 创建新的Redis客户端
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	// 构建Redis客户端配置
	redisCfg := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	// 创建Redis客户端
	client := redis.NewClient(redisCfg)

	// 测试连接
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("无法连接到Redis服务器: %w", err)
	}

	return &Client{Client: client}, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	return c.Client.Close()
}
