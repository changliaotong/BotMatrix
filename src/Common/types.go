package common

import (
	"database/sql"
	"sync"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/process"
	"gorm.io/gorm"
)

// ==================== 统一消息标准 (Neural Nexus) ====================

// MessageSegment represents a structured message segment (OneBot v12 compatible)
type MessageSegment struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// TextSegmentData represents the data for a text message segment
type TextSegmentData struct {
	Text string `json:"text"`
}

// ImageSegmentData represents the data for an image message segment
type ImageSegmentData struct {
	File string `json:"file"`
	URL  string `json:"url,omitempty"`
}

// InternalMessage is the unified message format used within BotMatrix
type InternalMessage struct {
	ID          string           `json:"id"`           // Message ID
	Time        int64            `json:"time"`         // Timestamp
	Platform    string           `json:"platform"`     // qq, wechat, etc.
	SelfID      string           `json:"self_id"`      // Bot ID
	Protocol    string           `json:"protocol"`     // v11, v12, etc.
	PostType    string           `json:"post_type"`    // message, notice, request, meta_event
	MessageType string           `json:"message_type"` // private, group
	SubType     string           `json:"sub_type"`     // friend, normal, etc.
	UserID      string           `json:"user_id"`      // Sender ID
	GroupID     string           `json:"group_id"`     // Group ID (if applicable)
	GroupName   string           `json:"group_name"`   // Group Name (if applicable)
	Message     []MessageSegment `json:"message"`      // Structured message
	RawMessage  string           `json:"raw_message"`  // Original raw message string
	SenderName  string           `json:"sender_name"`  // Sender nickname
	SenderCard  string           `json:"sender_card"`  // Sender card/alias in group
	UserAvatar  string           `json:"user_avatar"`  // User avatar URL
	Echo        string           `json:"echo"`         // Echo for tracking
	Status      string           `json:"status"`       // ok, failed
	Retcode     int              `json:"retcode"`      // OneBot return code
	Msg         string           `json:"msg"`          // Error message or info
	MetaType    string           `json:"meta_type"`    // heartbeat, lifecycle
	Extras      map[string]any   `json:"extras"`       // Additional platform-specific fields
}

// InternalAction is the unified action format used within BotMatrix
type InternalAction struct {
	Action   string         `json:"action"`
	Params   map[string]any `json:"params"`
	Echo     string         `json:"echo"`
	SelfID   string         `json:"self_id,omitempty"`
	Platform string         `json:"platform,omitempty"`

	// Common fields to avoid map usage
	UserID      string `json:"user_id,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	DetailType  string `json:"detail_type,omitempty"`
	Message     any    `json:"message,omitempty"`
}

// SkillResult represents the result of a skill execution from a Worker
type SkillResult struct {
	TaskID      any    `json:"task_id"`
	ExecutionID any    `json:"execution_id"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	Error       string `json:"error"`
	WorkerID    string `json:"worker_id"`
}

// WorkerCommand represents a command sent from Nexus to a Worker
type WorkerCommand struct {
	Type        string         `json:"type"`
	Skill       string         `json:"skill,omitempty"`
	Params      map[string]any `json:"params,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	TaskID      any            `json:"task_id,omitempty"`
	ExecutionID any            `json:"execution_id,omitempty"`
	Timestamp   int64          `json:"timestamp,omitempty"`
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

	// Common OneBot fields that might appear at top level in passive replies
	GroupID     string `json:"group_id"`
	UserID      string `json:"user_id"`
	MessageType string `json:"message_type"`
}

// GroupInfo represents cached group information
type GroupInfo struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	BotID     string    `json:"bot_id"`
	IsCached  bool      `json:"is_cached"`
	LastSeen  time.Time `json:"last_seen"`
	Avatar    string    `json:"avatar"`
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

// LoginInfo represents the data returned by get_login_info
type LoginInfo struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

// ApiResponse represents a standard API response
type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// GroupListItem represents an item in the list returned by get_group_list
type GroupListItem struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
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

// FriendInfo represents cached friend information
type FriendInfo struct {
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	BotID    string    `json:"bot_id"`
	IsCached bool      `json:"is_cached"`
	LastSeen time.Time `json:"last_seen"`
	Avatar   string    `json:"avatar"`
}

// MemberInfo represents cached group member information
type MemberInfo struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Card     string    `json:"card"`
	IsCached bool      `json:"is_cached"`
	LastSeen time.Time `json:"last_seen"`
}

// ==================== 基础结构体 ====================

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn          *websocket.Conn `json:"-"`
	SelfID        string          `json:"self_id"`
	Nickname      string          `json:"nickname"`
	GroupCount    int             `json:"group_count"`
	FriendCount   int             `json:"friend_count"`
	Connected     time.Time       `json:"connected"`
	Platform      string          `json:"platform"`
	Protocol      string          `json:"protocol"` // "v11" or "v12"
	Mutex         sync.Mutex      `json:"-"`
	SentCount     int64           `json:"sent_count"`     // Track sent messages per bot session
	RecvCount     int64           `json:"recv_count"`     // Track received messages per bot session
	LastHeartbeat time.Time       `json:"last_heartbeat"` // Track last heartbeat for timeout detection
}

// WorkerCapability 定义 Worker 具备的能力（如：签到、天气）
type WorkerCapability struct {
	Name        string            `json:"name"`        // 能力名称 (例如: "checkin")
	Description string            `json:"description"` // 描述 (例如: "每日签到获取积分")
	Usage       string            `json:"usage"`       // 使用示例
	Params      map[string]string `json:"params"`      // 参数说明
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
}

type ProcInfo struct {
	Pid    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"`
}

// Subscriber represents a UI or other consumer
type Subscriber struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	User  *User
}

// User represents a user with password hash
type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"` // 密码哈希，不序列化到JSON
	IsAdmin        bool      `json:"is_admin"`
	SessionVersion int       `json:"session_version"` // 用于同步登录状态和强制登出
	Active         bool      `json:"active"`          // 用户状态：启用/禁用
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserClaims for JWT
type UserClaims struct {
	UserID         int64  `json:"user_id"`
	Username       string `json:"username"`
	IsAdmin        bool   `json:"is_admin"`
	SessionVersion int    `json:"session_version"`
	jwt.RegisteredClaims
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Time      string    `json:"time"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source,omitempty"`
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

func NewManager() *Manager {
	m := &Manager{
		Bots:              make(map[string]*BotClient),
		Subscribers:       make(map[*websocket.Conn]*Subscriber),
		Workers:           make([]*WorkerClient, 0),
		PendingRequests:   make(map[string]chan InternalMessage),
		PendingTimestamps: make(map[string]time.Time),
		RoutingRules:      make(map[string]string),
		UserStats:         make(map[string]int64),
		GroupStats:        make(map[string]int64),
		BotStats:          make(map[string]int64),
		BotStatsSent:      make(map[string]int64),
		UserStatsToday:    make(map[string]int64),
		GroupStatsToday:   make(map[string]int64),
		BotStatsToday:     make(map[string]int64),
		LastResetDate:     time.Now().Format("2006-01-02"),
		StartTime:         time.Now(),
		ConnectionStats: ConnectionStats{
			BotConnectionDurations:    make(map[string]time.Duration),
			WorkerConnectionDurations: make(map[string]time.Duration),
			BotDisconnectReasons:      make(map[string]int64),
			WorkerDisconnectReasons:   make(map[string]int64),
			LastBotActivity:           make(map[string]time.Time),
			LastWorkerActivity:        make(map[string]time.Time),
		},
		StatsMutex: sync.RWMutex{},
		Mutex:      sync.RWMutex{},
		// Bot Data Cache
		GroupCache:  make(map[string]GroupInfo),
		MemberCache: make(map[string]MemberInfo),
		FriendCache: make(map[string]FriendInfo),
		CacheMutex:  sync.RWMutex{},

		// User Management
		Users:              make(map[string]*User),
		UsersMutex:         sync.RWMutex{},
		ProcMap:            make(map[int32]*process.Process),
		WorkerRequestTimes: make(map[string]time.Time),
		LogBuffer:          make([]LogEntry, 0),
	}
	return m
}

type WorkerInfo struct {
	ID              string `json:"id"`
	RemoteAddr      string `json:"remote_addr"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	Connected       string `json:"connected"`
	LastSeen        string `json:"last_seen"`
	HandledCount    int64  `json:"handled_count"`
	AvgRTT          string `json:"avg_rtt"`
	LastRTT         string `json:"last_rtt"`
	AvgProcessTime  string `json:"avg_process_time"`
	LastProcessTime string `json:"last_process_time"`
	IsAlive         bool   `json:"is_alive"`
}

// WorkerUpdateEvent represents a worker status update event
type WorkerUpdateEvent struct {
	Type string     `json:"type"` // Always "worker_update"
	Data WorkerInfo `json:"data"`
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

// ==================== 统计结构体 ====================

// ConnectionStats tracks connection lifecycle statistics
type ConnectionStats struct {
	TotalBotConnections       int64                    `json:"total_bot_connections"`
	TotalWorkerConnections    int64                    `json:"total_worker_connections"`
	BotConnectionDurations    map[string]time.Duration `json:"bot_connection_durations"`    // bot_id -> duration
	WorkerConnectionDurations map[string]time.Duration `json:"worker_connection_durations"` // worker_id -> duration
	BotDisconnectReasons      map[string]int64         `json:"bot_disconnect_reasons"`      // reason -> count
	WorkerDisconnectReasons   map[string]int64         `json:"worker_disconnect_reasons"`   // reason -> count
	LastBotActivity           map[string]time.Time     `json:"last_bot_activity"`           // bot_id -> last activity
	LastWorkerActivity        map[string]time.Time     `json:"last_worker_activity"`        // worker_id -> last activity
	Mutex                     sync.RWMutex
}

// BotStatDetail represents detailed stats for a bot
type BotStatDetail struct {
	Sent     int64            `json:"sent"`
	Received int64            `json:"received"`
	Users    map[string]int64 `json:"users"`  // UserID -> Count
	Groups   map[string]int64 `json:"groups"` // GroupID -> Count
	LastMsg  time.Time        `json:"last_msg"`
}

// ==================== 管理器结构体 ====================

// AppConfig represents the backend configuration
type AppConfig struct {
	WSPort               string `json:"ws_port"`
	WebUIPort            string `json:"webui_port"`
	RedisAddr            string `json:"redis_addr"`
	RedisPwd             string `json:"redis_pwd"`
	JWTSecret            string `json:"jwt_secret"`
	DefaultAdminPassword string `json:"default_admin_password"`
	StatsFile            string `json:"stats_file"`

	// Database Configuration
	PGHost     string `json:"pg_host"`
	PGPort     int    `json:"pg_port"`
	PGUser     string `json:"pg_user"`
	PGPassword string `json:"pg_password"`
	PGDBName   string `json:"pg_dbname"`
	PGSSLMode  string `json:"pg_sslmode"`

	// Feature Flags
	EnableSkill bool   `json:"enable_skill"`
	LogLevel    string `json:"log_level"`
	AutoReply   bool   `json:"auto_reply"`

	// Azure Translator Config
	AzureTranslateKey      string `json:"azure_translate_key"`
	AzureTranslateEndpoint string `json:"azure_translate_endpoint"`
	AzureTranslateRegion   string `json:"azure_translate_region"`
}

// Manager holds the state
type Manager struct {
	Config      *AppConfig
	Bots        map[string]*BotClient
	Subscribers map[*websocket.Conn]*Subscriber // UI or other consumers (Broadcast)
	Workers     []*WorkerClient                 // Business logic workers (Round-Robin)
	WorkerIndex int                             // For Round-Robin
	Mutex       sync.RWMutex
	Upgrader    websocket.Upgrader
	LogBuffer   []LogEntry
	LogMutex    sync.RWMutex

	// Pending Requests (Echo -> Channel)
	PendingRequests   map[string]chan InternalMessage
	PendingTimestamps map[string]time.Time // Echo -> Send Time for RTT tracking
	PendingMutex      sync.Mutex

	// Worker Processing Tracking (Echo -> Send Time to Worker)
	WorkerRequestTimes map[string]time.Time // Echo -> Time when message sent to worker
	WorkerRequestMutex sync.Mutex

	// Redis
	Rdb *redis.Client

	// Docker
	DockerClient *dclient.Client

	// 临时固定路由规则 (测试用)
	RoutingRules map[string]string // group_id/bot_id -> worker_id

	// Chat Stats
	StatsMutex      sync.RWMutex
	StartTime       time.Time        `json:"start_time"`        // Server start time
	TotalMessages   int64            `json:"total_messages"`    // Global counter
	SentMessages    int64            `json:"sent_messages"`     // Global sent counter
	UserStats       map[string]int64 `json:"user_stats"`        // UserID -> Count (Total)
	GroupStats      map[string]int64 `json:"group_stats"`       // GroupID -> Count (Total)
	BotStats        map[string]int64 `json:"bot_stats"`         // BotID -> Count (Total Recv)
	BotStatsSent    map[string]int64 `json:"bot_stats_sent"`    // BotID -> Count (Total Sent)
	UserStatsToday  map[string]int64 `json:"user_stats_today"`  // UserID -> Count (Today)
	GroupStatsToday map[string]int64 `json:"group_stats_today"` // GroupID -> Count (Today)
	BotStatsToday   map[string]int64 `json:"bot_stats_today"`   // BotID -> Count (Today)
	LastResetDate   string           `json:"last_reset_date"`   // YYYY-MM-DD

	// Granular Stats (Per Bot)
	BotDetailedStats map[string]*BotStatDetail `json:"bot_detailed_stats"` // BotID -> Detail

	// System Resource Stats
	HistoryMutex sync.RWMutex
	CPUTrend     []float64  `json:"cpu_trend"`
	MemTrend     []uint64   `json:"mem_trend"`
	MsgTrend     []int64    `json:"msg_trend"`
	SentTrend    []int64    `json:"sent_trend"`
	RecvTrend    []int64    `json:"recv_trend"`
	NetSentTrend []uint64   `json:"net_sent_trend"`
	NetRecvTrend []uint64   `json:"net_recv_trend"`
	TrendLabels  []string   `json:"trend_labels"`
	TopProcesses []ProcInfo `json:"top_processes"`
	ProcMap      map[int32]*process.Process

	// For delta calculation
	LastTrendTotal int64
	LastTrendSent  int64

	// Connection Stats (New)
	ConnectionStats ConnectionStats

	// User Management
	Users      map[string]*User // 用户名 -> 用户信息
	UsersMutex sync.RWMutex     // 用户存储的并发保护
	DB         *sql.DB          // 数据库连接 (PostgreSQL)

	// GORM Support
	GORMDB      *gorm.DB     // GORM数据库连接
	GORMManager *GORMManager // GORM管理器

	// Message Cache (For when no workers are available)
	MessageCache []InternalMessage
	CacheMutex   sync.RWMutex
	GroupCache   map[string]GroupInfo
	MemberCache  map[string]MemberInfo
	FriendCache  map[string]FriendInfo

	// Local Idempotency Cache (reduce Redis pressure)
	LocalIdempotency sync.Map // msgID -> time.Time
}
