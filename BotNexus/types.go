package main

import (
	"database/sql"
	"sync"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/process"
)

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
	Mutex         sync.Mutex      `json:"-"`
	SentCount     int64           `json:"sent_count"`     // Track sent messages per bot session
	RecvCount     int64           `json:"recv_count"`     // Track received messages per bot session
	LastHeartbeat time.Time       `json:"last_heartbeat"` // Track last heartbeat for timeout detection
}

// WorkerClient represents a business logic worker
type WorkerClient struct {
	ID            string // Worker标识
	Conn          *websocket.Conn
	Mutex         sync.Mutex
	Connected     time.Time
	HandledCount  int64
	LastHeartbeat time.Time

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
}

// SyncState represents the initial state for subscribers
type SyncState struct {
	Type          string                            `json:"type"` // Always "sync_state"
	Groups        map[string]map[string]interface{} `json:"groups"`
	Friends       map[string]map[string]interface{} `json:"friends"`
	Members       map[string]map[string]interface{} `json:"members"`
	Bots          []BotClient                       `json:"bots"`
	Workers       []WorkerInfo                      `json:"workers"`
	TotalMessages int64                             `json:"total_messages"`
}

type WorkerInfo struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen"`
}

// RoutingEvent represents a message routing event for visualization
type RoutingEvent struct {
	Type          string    `json:"type"`      // Always "routing_event"
	Source        string    `json:"source"`    // BotID or WorkerID or UserID
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
}

// Manager holds the state
type Manager struct {
	config      *AppConfig
	bots        map[string]*BotClient
	subscribers map[*websocket.Conn]*Subscriber // UI or other consumers (Broadcast)
	workers     []*WorkerClient                 // Business logic workers (Round-Robin)
	workerIndex int                             // For Round-Robin
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
	logBuffer   []LogEntry
	logMutex    sync.RWMutex

	// Pending Requests (Echo -> Channel)
	pendingRequests   map[string]chan map[string]interface{}
	pendingTimestamps map[string]time.Time // Echo -> Send Time for RTT tracking
	pendingMutex      sync.Mutex

	// Worker Processing Tracking (Echo -> Send Time to Worker)
	workerRequestTimes map[string]time.Time // Echo -> Time when message sent to worker
	workerRequestMutex sync.Mutex

	// Redis
	rdb *redis.Client

	// Docker
	dockerClient *dclient.Client

	// 临时固定路由规则 (测试用)
	routingRules map[string]string // group_id/bot_id -> worker_id

	// Chat Stats
	statsMutex      sync.RWMutex
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
	procMap      map[int32]*process.Process

	// For delta calculation
	lastTrendTotal int64
	lastTrendSent  int64

	// Connection Stats (New)
	connectionStats ConnectionStats

	// User Management
	users      map[string]*User // 用户名 -> 用户信息
	usersMutex sync.RWMutex     // 用户存储的并发保护
	db         *sql.DB          // SQLite 数据库连接

	// Message Cache (For when no workers are available)
	messageCache      []map[string]interface{}
	messageCacheMutex sync.Mutex

	// Bot Data Cache
	groupCache  map[string]map[string]interface{} // group_id -> data
	memberCache map[string]map[string]interface{} // group_id:user_id -> data
	friendCache map[string]map[string]interface{} // user_id -> data
	cacheMutex  sync.RWMutex
}
