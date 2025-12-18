package main

import (
	"database/sql"
	"sync"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// ==================== 基础结构体 ====================

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn          *websocket.Conn
	SelfID        string
	Nickname      string
	GroupCount    int
	FriendCount   int
	Connected     time.Time
	Platform      string
	Mutex         sync.Mutex
	SentCount     int64     // Track sent messages per bot session
	RecvCount     int64     // Track received messages per bot session
	LastHeartbeat time.Time // Track last heartbeat for timeout detection
}

// WorkerClient represents a business logic worker
type WorkerClient struct {
	ID            string // Worker标识
	Conn          *websocket.Conn
	Mutex         sync.Mutex
	Connected     time.Time
	HandledCount  int64
	LastHeartbeat time.Time
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
	Timestamp time.Time `json:"timestamp"`
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
	Sent     int64           `json:"sent"`
	Received int64           `json:"received"`
	Users    map[int64]int64 `json:"users"`  // UserID -> Count
	Groups   map[int64]int64 `json:"groups"` // GroupID -> Count
	LastMsg  time.Time       `json:"last_msg"`
}

// ==================== 管理器结构体 ====================

// Manager holds the state
type Manager struct {
	bots        map[string]*BotClient
	subscribers map[*websocket.Conn]*Subscriber // UI or other consumers (Broadcast)
	workers     []*WorkerClient                 // Business logic workers (Round-Robin)
	workerIndex int                             // For Round-Robin
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
	logBuffer   []LogEntry
	logMutex    sync.RWMutex

	// Pending Requests (Echo -> Channel)
	pendingRequests map[string]chan map[string]interface{}
	pendingMutex    sync.Mutex

	// Redis
	rdb *redis.Client

	// Docker
	dockerClient *dclient.Client

	// 临时固定路由规则 (测试用)
	routingRules map[string]string // group_id/bot_id -> worker_id

	// Chat Stats
	statsMutex      sync.RWMutex
	TotalMessages   int64            `json:"total_messages"`    // Global counter
	SentMessages    int64            `json:"sent_messages"`     // Global sent counter
	UserStats       map[int64]int64  `json:"user_stats"`        // UserID -> Count (Total)
	GroupStats      map[int64]int64  `json:"group_stats"`       // GroupID -> Count (Total)
	BotStats        map[string]int64 `json:"bot_stats"`         // BotID -> Count (Total Recv)
	BotStatsSent    map[string]int64 `json:"bot_stats_sent"`    // BotID -> Count (Total Sent)
	UserStatsToday  map[int64]int64  `json:"user_stats_today"`  // UserID -> Count (Today)
	GroupStatsToday map[int64]int64  `json:"group_stats_today"` // GroupID -> Count (Today)
	BotStatsToday   map[string]int64 `json:"bot_stats_today"`   // BotID -> Count (Today)
	LastResetDate   string           `json:"last_reset_date"`   // YYYY-MM-DD

	// Granular Stats (Per Bot)
	BotDetailedStats map[string]*BotStatDetail `json:"bot_detailed_stats"` // BotID -> Detail

	// Time Series Stats (New)
	HistoryMutex sync.RWMutex
	CPUTrend     []float64 `json:"cpu_trend"`
	MemTrend     []uint64  `json:"mem_trend"`
	MsgTrend     []int64   `json:"msg_trend"`
	SentTrend    []int64   `json:"sent_trend"`
	RecvTrend    []int64   `json:"recv_trend"`
	TrendLabels  []string  `json:"trend_labels"`

	// Connection Stats (New)
	connectionStats ConnectionStats

	// User Management
	users      map[string]*User // 用户名 -> 用户信息
	usersMutex sync.RWMutex     // 用户存储的并发保护
	db         *sql.DB          // SQLite 数据库连接
}
