package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

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

	// Connection Stats (New)
	connectionStats ConnectionStats
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// BotStatDetail represents detailed stats for a bot
type BotStatDetail struct {
	Sent     int64           `json:"sent"`
	Received int64           `json:"received"`
	Users    map[int64]int64 `json:"users"`  // UserID -> Count
	Groups   map[int64]int64 `json:"groups"` // GroupID -> Count
	LastMsg  time.Time       `json:"last_msg"`
}

func NewManager() *Manager {
	m := &Manager{
		bots:            make(map[string]*BotClient),
		subscribers:     make(map[*websocket.Conn]*Subscriber),
		workers:         make([]*WorkerClient, 0),
		pendingRequests: make(map[string]chan map[string]interface{}),
		routingRules:    make(map[string]string),
		// Stats
		UserStats:        make(map[int64]int64),
		GroupStats:       make(map[int64]int64),
		BotStats:         make(map[string]int64),
		BotStatsSent:     make(map[string]int64),
		UserStatsToday:   make(map[int64]int64),
		GroupStatsToday:  make(map[int64]int64),
		BotStatsToday:    make(map[string]int64),
		BotDetailedStats: make(map[string]*BotStatDetail),
		LastResetDate:    time.Now().Format("2006-01-02"),
		// Connection Stats
		connectionStats: ConnectionStats{
			BotConnectionDurations:    make(map[string]time.Duration),
			WorkerConnectionDurations: make(map[string]time.Duration),
			BotDisconnectReasons:      make(map[string]int64),
			WorkerDisconnectReasons:   make(map[string]int64),
			LastBotActivity:           make(map[string]time.Time),
			LastWorkerActivity:        make(map[string]time.Time),
		},
	}

	// 初始化Redis连接
	m.rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PWD,
		DB:       0,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
	} else {
		log.Println("Redis connected successfully")
	}

	return m
}

func main() {
	manager := NewManager()

	// 启动各种goroutine
	go manager.StartWorkerTimeoutDetection()
	go manager.StartBotTimeoutDetection()
	go manager.StartPeriodicStatsSave()
	go manager.StartTrendCollection()

	// 设置HTTP路由
	mux := http.NewServeMux()

	// WebSocket endpoints
	mux.HandleFunc("/ws/bots", manager.handleBotWebSocket)
	mux.HandleFunc("/ws/workers", manager.handleWorkerWebSocket)
	mux.HandleFunc("/ws/subscribers", manager.handleSubscriberWebSocket)

	// API endpoints
	mux.HandleFunc("/api/login", manager.handleLogin)
	mux.HandleFunc("/api/stats", manager.handleGetStats)
	mux.HandleFunc("/api/logs", manager.handleGetLogs)

	// 启动服务器
	go func() {
		log.Printf("Starting WebSocket server on %s", WS_PORT)
		if err := http.ListenAndServe(WS_PORT, mux); err != nil {
			log.Fatal("WebSocket server failed: ", err)
		}
	}()

	// 启动WebUI服务器
	webMux := http.NewServeMux()
	webMux.HandleFunc("/api/login", manager.handleLogin)
	webMux.HandleFunc("/api/stats", manager.handleGetStats)
	webMux.HandleFunc("/api/logs", manager.handleGetLogs)

	go func() {
		log.Printf("Starting WebUI server on %s", WEBUI_PORT)
		if err := http.ListenAndServe(WEBUI_PORT, webMux); err != nil {
			log.Fatal("WebUI server failed: ", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
