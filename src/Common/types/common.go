package types

import "time"

// ApiResponse represents a standard API response
type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Time      string    `json:"time"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source,omitempty"`
}

// SessionContext 存储会话上下文信息
type SessionContext struct {
	Platform  string            `json:"platform"`
	UserID    string            `json:"user_id"`
	LastMsg   InternalMessage   `json:"last_msg"`
	History   []InternalMessage `json:"history"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SessionState represents a specific state of a session (e.g., waiting for input)
type SessionState struct {
	Step       string         `json:"step"`
	Action     string         `json:"action"`
	Data       map[string]any `json:"data"` // Internal data for the state
	WaitingFor string         `json:"waiting_for"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// RoutingParams 用于 BroadcastRoutingEvent 的可选参数
type RoutingParams struct {
	SourceType  string `json:"source_type,omitempty"`
	SourceLabel string `json:"source_label,omitempty"`
	TargetType  string `json:"target_type,omitempty"`
	TargetLabel string `json:"target_label,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	UserName    string `json:"user_name,omitempty"`
	UserAvatar  string `json:"user_avatar,omitempty"`
	Content     string `json:"content,omitempty"`
	Platform    string `json:"platform,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
}

// SyncState represents the initial state for subscribers
type SyncState struct {
	Type          string                `json:"type"` // Always "sync_state"
	Groups        map[string]GroupInfo  `json:"groups"`
	Friends       map[string]FriendInfo `json:"friends"`
	Members       map[string]MemberInfo `json:"members"`
	Bots          []BotClient           `json:"bots"`
	Workers       []WorkerInfo          `json:"workers"`
	TotalMessages int64                 `json:"total_messages"`
	Uptime        string                `json:"uptime"`
	Version       string                `json:"version"`
}

// RoutingEvent represents a message routing event for visualization
type RoutingEvent struct {
	Type          string    `json:"type"`   // Always "routing_event"
	Source        string    `json:"source"` // BotID or WorkerID or UserID
	SourceType    string    `json:"source_type"`
	SourceLabel   string    `json:"source_label"`
	Target        string    `json:"target"` // WorkerID or BotID or "Nexus"
	TargetType    string    `json:"target_type"`
	TargetLabel   string    `json:"target_label"`
	Direction     string    `json:"direction"` // "user_to_bot", "bot_to_nexus", "nexus_to_worker", etc.
	MsgType       string    `json:"msg_type"`  // "message", "request", "response"
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`        // Optional: User ID
	UserName      string    `json:"user_name"`      // Optional: User Nickname
	UserAvatar    string    `json:"user_avatar"`    // Optional: User Avatar URL
	Content       string    `json:"content"`        // Optional: Message Content
	Platform      string    `json:"platform"`       // Optional: Platform (QQ, WeChat, etc.)
	GroupID       string    `json:"group_id"`       // Optional: Group ID
	GroupName     string    `json:"group_name"`     // Optional: Group Name
	Color         string    `json:"color"`          // Optional: Custom node color
	TotalMessages int64     `json:"total_messages"` // Current total message count
}

// DockerEvent represents a docker container state change
type DockerEvent struct {
	Type        string    `json:"type"` // Always "docker_event"
	Action      string    `json:"action"`
	ContainerID string    `json:"container_id"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}
