package models

import "time"

// RoutingRule represents a routing rule for messages
type RoutingRule struct {
	Pattern        string    `json:"pattern"`
	TargetWorkerID string    `json:"target_worker_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GroupCache represents cached group information
type GroupCache struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	BotID     string    `json:"bot_id"`
	LastSeen  time.Time `json:"last_seen"`
}

// FriendCache represents cached friend information
type FriendCache struct {
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Remark   string    `json:"remark"`
	BotID    string    `json:"bot_id"`
	LastSeen time.Time `json:"last_seen"`
}

// MemberCache represents cached group member information
type MemberCache struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Card     string    `json:"card"`
	Role     string    `json:"role"`
	LastSeen time.Time `json:"last_seen"`
}
