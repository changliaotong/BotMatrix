package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// NewManager 创建管理器实例
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

	// 启动各种监控任务
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

	// 启动WebSocket服务器
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

	// 静态文件服务
	webMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
		} else {
			http.ServeFile(w, r, r.URL.Path[1:])
		}
	})

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
