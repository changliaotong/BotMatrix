package types

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SkillResult represents the result of a skill execution from a Worker
type SkillResult struct {
	TaskID        any    `json:"task_id"`
	ExecutionID   any    `json:"execution_id"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Status        string `json:"status"`
	Result        string `json:"result"`
	Error         string `json:"error"`
	WorkerID      string `json:"worker_id"`
}

// WorkerCommand represents a command sent from Nexus to a Worker
type WorkerCommand struct {
	Type          string         `json:"type"`
	Skill         string         `json:"skill,omitempty"`
	Params        map[string]any `json:"params,omitempty"`
	UserID        string         `json:"user_id,omitempty"`
	TaskID        any            `json:"task_id,omitempty"`
	ExecutionID   any            `json:"execution_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Timestamp     int64          `json:"timestamp,omitempty"`
}

// WorkerMessage represents any message coming from a Worker via WebSocket
type WorkerMessage struct {
	Type         string             `json:"type"`
	Action       string             `json:"action"`
	Echo         string             `json:"echo"`
	Capabilities []WorkerCapability `json:"capabilities"`
	Params       map[string]any     `json:"params"`
	SelfID       string             `json:"self_id"`
	Platform     string             `json:"platform"`
	Reply        string             `json:"reply"`
	Status       string             `json:"status"`
	Result       string             `json:"result"`
	Error        string             `json:"error"`
	TaskID       any                `json:"task_id"`
	ExecutionID  any                `json:"execution_id"`
	Metadata     map[string]any     `json:"metadata"` // 额外元数据

	// Common OneBot fields that might appear at top level in passive replies
	GroupID     string `json:"group_id"`
	UserID      string `json:"user_id"`
	MessageType string `json:"message_type"`
}

// WorkerCapability 定义 Worker 具备的能力（如：签到、天气）
type WorkerCapability struct {
	Name        string            `json:"name"`        // 能力名称 (例如: "checkin")
	Description string            `json:"description"` // 描述 (例如: "每日签到获取积分")
	Usage       string            `json:"usage"`       // 使用示例
	Params      map[string]string `json:"params"`      // 参数说明
	Regex       string            `json:"regex"`       // 新增：指令正则触发器 (例如: "^签到$")
}

// WorkerClient represents a business logic worker
type WorkerClient struct {
	ID            string // Worker标识
	Conn          *websocket.Conn
	Mutex         sync.Mutex
	Connected     time.Time
	HandledCount  int64
	LastHeartbeat time.Time
	Capabilities  []WorkerCapability `json:"capabilities"` // Worker 报备的能力列表
	Protocol      string             `json:"protocol"`     // "v11" or "v12"

	// RTT Tracking
	AvgRTT     time.Duration `json:"avg_rtt"`
	LastRTT    time.Duration `json:"last_rtt"`
	RTTSamples []time.Duration
	MaxSamples int

	// Process Time Tracking (Worker processing duration)
	AvgProcessTime     time.Duration `json:"avg_process_time"`
	LastProcessTime    time.Duration `json:"last_process_time"`
	ProcessTimeSamples []time.Duration
	Metadata           map[string]any `json:"metadata"` // 额外元数据 (如：插件列表、系统信息等)
}

type WorkerInfo struct {
	ID              string             `json:"id"`
	RemoteAddr      string             `json:"remote_addr"`
	Type            string             `json:"type"`
	Status          string             `json:"status"`
	Connected       string             `json:"connected"`
	LastSeen        string             `json:"last_seen"`
	HandledCount    int64              `json:"handled_count"`
	AvgRTT          string             `json:"avg_rtt"`
	LastRTT         string             `json:"last_rtt"`
	AvgProcessTime  string             `json:"avg_process_time"`
	LastProcessTime string             `json:"last_process_time"`
	Capabilities    []WorkerCapability `json:"capabilities"`
	IsAlive         bool               `json:"is_alive"`
}

// WorkerUpdateEvent represents a worker status update event
type WorkerUpdateEvent struct {
	Type string     `json:"type"` // Always "worker_update"
	Data WorkerInfo `json:"data"`
}
