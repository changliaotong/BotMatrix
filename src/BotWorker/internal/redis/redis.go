package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"botworker/internal/config"
)

// Redis Key 设计 (需与 Nexus 保持一致)
const (
	REDIS_KEY_SESSION_CONTEXT = "botmatrix:session:%s:%s"       // platform:user_id
	REDIS_KEY_SESSION_STATE   = "botmatrix:session:state:%s:%s" // platform:user_id
	REDIS_KEY_CONFIG_TTL      = "botmatrix:config:ttl"
)

// Client Redis客户端封装
type Client struct {
	*redis.Client
}

// Nil Redis空结果错误
var Nil = redis.Nil

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

// GetSessionContext 获取会话上下文
func (c *Client) GetSessionContext(platform, userID string) (map[string]interface{}, error) {
	ctx := context.Background()
	key := fmt.Sprintf(REDIS_KEY_SESSION_CONTEXT, platform, userID)

	val, err := c.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &contextData); err != nil {
		return nil, err
	}
	return contextData, nil
}

// SetSessionState 设置会话状态
func (c *Client) SetSessionState(platform, userID string, state map[string]interface{}, ttl time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf(REDIS_KEY_SESSION_STATE, platform, userID)

	// 如果没有提供 TTL，尝试从配置获取
	if ttl == 0 {
		ttl = 24 * time.Hour // 默认
		if val, err := c.HGet(ctx, REDIS_KEY_CONFIG_TTL, "session_ttl_sec").Result(); err == nil {
			if ttlSec, err := strconv.ParseInt(val, 10, 64); err == nil && ttlSec > 0 {
				ttl = time.Duration(ttlSec) * time.Second
			}
		}
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, ttl).Err()
}

// GetSessionState 获取会话状态
func (c *Client) GetSessionState(platform, userID string) (map[string]interface{}, error) {
	ctx := context.Background()
	key := fmt.Sprintf(REDIS_KEY_SESSION_STATE, platform, userID)

	val, err := c.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var state map[string]interface{}
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, err
	}
	return state, nil
}
