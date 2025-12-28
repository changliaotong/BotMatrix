package common

import "time"

// RoutingRule 路由规则结构体
type RoutingRule struct {
	Pattern        string    `json:"pattern"`
	TargetWorkerID string    `json:"target_worker_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GroupCache 群组缓存结构体
type GroupCache struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	BotID     string    `json:"bot_id"`
	LastSeen  time.Time `json:"last_seen"`
}

// FriendCache 好友缓存结构体
type FriendCache struct {
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	BotID    string    `json:"bot_id"`
	LastSeen time.Time `json:"last_seen"`
}

// MemberCache 群成员缓存结构体
type MemberCache struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Card     string    `json:"card"`
	LastSeen time.Time `json:"last_seen"`
}
