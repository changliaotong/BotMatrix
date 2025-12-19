// BotNexus - 统一构建入口文件
// 这个文件用于Docker构建，整合所有模块但避免重复定义

package main

import (
	"context"
	"encoding/json"
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

// 版本号定义
const VERSION = "80"

// 注意：常量和upgrader定义从其他文件导入，此处不再重复定义

// 主函数 - 整合所有功能
func main() {
	log.Printf("启动 BotNexus 服务... 版本号: %s", VERSION)

	// 创建管理器 (内部会初始化数据库和管理员)
	manager := NewManager()

	// 启动超时检测
	go manager.StartWorkerTimeoutDetection()
	go manager.StartBotTimeoutDetection()

	// 启动统计信息收集
	go manager.StartTrendCollection()

	// 启动统计信息重置定时器
	go manager.StartStatsResetTimer()

	// 定期保存统计数据 (如果需要)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			// 这里可以保存其他非持久化数据
		}
	}()

	// 设置HTTP路由
	http.HandleFunc("/login", manager.handleLogin)
	http.HandleFunc("/api/login", manager.handleLogin)
	http.HandleFunc("/api/stats", manager.JWTMiddleware(manager.handleGetStats))
	http.HandleFunc("/api/system/stats", manager.JWTMiddleware(manager.handleGetSystemStats))
	http.HandleFunc("/api/me", manager.JWTMiddleware(manager.handleGetUserInfo))
	http.HandleFunc("/api/user/info", manager.JWTMiddleware(manager.handleGetUserInfo))
	http.HandleFunc("/api/user/password", manager.JWTMiddleware(manager.handleChangePassword))
	http.HandleFunc("/api/bots", manager.JWTMiddleware(manager.handleGetBots))
	http.HandleFunc("/api/workers", manager.JWTMiddleware(manager.handleGetWorkers))
	http.HandleFunc("/api/logs", manager.JWTMiddleware(manager.handleGetLogs))

	// Docker 接口
	http.HandleFunc("/api/docker/list", manager.JWTMiddleware(manager.handleDockerList))
	http.HandleFunc("/api/docker/action", manager.JWTMiddleware(manager.handleDockerAction))
	http.HandleFunc("/api/docker/add-bot", manager.JWTMiddleware(manager.handleDockerAddBot))
	http.HandleFunc("/api/docker/add-worker", manager.JWTMiddleware(manager.handleDockerAddWorker))

	// 管理员接口
	http.HandleFunc("/api/admin/routing", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			manager.handleGetRoutingRules(w, r)
		case http.MethodPost:
			manager.handleSetRoutingRule(w, r)
		case http.MethodDelete:
			manager.handleDeleteRoutingRule(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// 用户管理接口 (仅限管理员)
	http.HandleFunc("/api/admin/users", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			manager.handleAdminListUsers(w, r)
		case http.MethodPost:
			manager.handleAdminCreateUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/admin/user/reset-password", manager.AdminMiddleware(manager.handleAdminResetPassword))

	http.HandleFunc("/ws/bots", manager.handleBotWebSocket)
	http.HandleFunc("/ws/workers", manager.handleWorkerWebSocket)
	http.HandleFunc("/ws/subscriber", manager.JWTMiddleware(manager.handleSubscriberWebSocket))
	// 静态文件服务 - 同时支持本地开发和Docker环境
	webDir := "./web"
	if _, err := os.Stat("/app/web"); err == nil {
		webDir = "/app/web"
	}
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// 启动HTTP服务器
	go func() {
		log.Printf("WebSocket 服务启动在端口 %s", WS_PORT)
		if err := http.ListenAndServe(WS_PORT, nil); err != nil {
			log.Fatal("WebSocket 服务启动失败:", err)
		}
	}()

	// 启动WebUI服务器 (暂时注释掉以避免端口冲突)
	go func() {
		log.Printf("WebUI 服务启动在端口 %s", WEBUI_PORT)
		if err := http.ListenAndServe(WEBUI_PORT, manager.createWebUIHandler()); err != nil {
			log.Fatal("WebUI 服务启动失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭数据库连接
	if manager.db != nil {
		manager.db.Close()
		log.Printf("[INFO] 数据库已关闭")
	}

	// 关闭Redis连接
	if manager.rdb != nil {
		if err := manager.rdb.Close(); err != nil {
			log.Printf("关闭Redis连接失败: %v", err)
		}
	}

	log.Println("服务已关闭")
}

func (m *Manager) createWebUIHandler() http.Handler {
	mux := http.NewServeMux()

	// --- 公开接口 ---
	mux.HandleFunc("/login", m.handleLogin)
	mux.HandleFunc("/api/login", m.handleLogin)

	// --- 需要登录的接口 ---
	mux.HandleFunc("/api/me", m.JWTMiddleware(m.handleGetUserInfo))
	mux.HandleFunc("/api/user/password", m.JWTMiddleware(m.handleChangePassword))

	mux.HandleFunc("/api/bots", m.JWTMiddleware(m.handleGetBots))
	mux.HandleFunc("/api/workers", m.JWTMiddleware(m.handleGetWorkers))
	mux.HandleFunc("/api/stats", m.JWTMiddleware(m.handleGetStats))
	mux.HandleFunc("/api/system/stats", m.JWTMiddleware(m.handleGetSystemStats))
	mux.HandleFunc("/api/logs", m.JWTMiddleware(m.handleGetLogs))

	// --- WebSocket 接口 (供 WebUI 使用) ---
	mux.HandleFunc("/ws/subscriber", m.JWTMiddleware(m.handleSubscriberWebSocket))

	// --- 需要管理员权限的接口 ---
	mux.HandleFunc("/api/admin/routing", m.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			m.handleGetRoutingRules(w, r)
		case http.MethodPost:
			m.handleSetRoutingRule(w, r)
		case http.MethodDelete:
			m.handleDeleteRoutingRule(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// 用户管理接口 (仅限管理员)
	mux.HandleFunc("/api/admin/users", m.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			m.handleAdminListUsers(w, r)
		case http.MethodPost:
			m.handleAdminCreateUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/user/reset-password", m.AdminMiddleware(m.handleAdminResetPassword))

	// 静态文件服务 - 同时支持本地开发和Docker环境
	webDir := "./web"
	if _, err := os.Stat("/app/web"); err == nil {
		webDir = "/app/web"
	}
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	// Overmind (Flutter Web) 服务
	overmindDir := "../Overmind/build/web"
	if _, err := os.Stat("/app/overmind"); err == nil {
		overmindDir = "/app/overmind"
	}
	mux.Handle("/overmind/", http.StripPrefix("/overmind/", http.FileServer(http.Dir(overmindDir))))

	return mux
}

// handleGetBots 处理获取Bot列表的请求
func (m *Manager) handleGetBots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	bots := make([]map[string]interface{}, 0, len(m.bots))
	for id, bot := range m.bots {
		bots = append(bots, map[string]interface{}{
			"id":           id,
			"self_id":      bot.SelfID,
			"nickname":     bot.Nickname,
			"platform":     bot.Platform,
			"connected":    bot.Connected.Format("2006-01-02 15:04:05"),
			"group_count":  bot.GroupCount,
			"friend_count": bot.FriendCount,
			"is_alive":     true, // 当前连接的都是在线的
		})
	}
	m.mutex.RUnlock()

	json.NewEncoder(w).Encode(bots)
}

// handleGetWorkers 处理获取Worker列表的请求
func (m *Manager) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	workers := make([]map[string]interface{}, 0, len(m.workers))
	for _, worker := range m.workers {
		workers = append(workers, map[string]interface{}{
			"id":            worker.ID,
			"remote_addr":   worker.ID,
			"connected":     worker.Connected.Format("2006-01-02 15:04:05"),
			"handled_count": worker.HandledCount,
			"is_alive":      true,
			"status":        "Online",
		})
	}
	m.mutex.RUnlock()

	json.NewEncoder(w).Encode(workers)
}

// 简化的管理器创建函数
func NewManager() *Manager {
	m := &Manager{
		bots:              make(map[string]*BotClient),
		subscribers:       make(map[*websocket.Conn]*Subscriber),
		workers:           make([]*WorkerClient, 0),
		pendingRequests:   make(map[string]chan map[string]interface{}),
		pendingTimestamps: make(map[string]time.Time),
		routingRules:      make(map[string]string),
		UserStats:         make(map[int64]int64),
		GroupStats:        make(map[int64]int64),
		BotStats:          make(map[string]int64),
		BotStatsSent:      make(map[string]int64),
		UserStatsToday:    make(map[int64]int64),
		GroupStatsToday:   make(map[int64]int64),
		BotStatsToday:     make(map[string]int64),
		LastResetDate:     time.Now().Format("2006-01-02"),
		StartTime:         time.Now(),
		connectionStats: ConnectionStats{
			BotConnectionDurations:    make(map[string]time.Duration),
			WorkerConnectionDurations: make(map[string]time.Duration),
			BotDisconnectReasons:      make(map[string]int64),
			WorkerDisconnectReasons:   make(map[string]int64),
			LastBotActivity:           make(map[string]time.Time),
			LastWorkerActivity:        make(map[string]time.Time),
		},
		statsMutex: sync.RWMutex{},
		mutex:      sync.RWMutex{},
		// Bot Data Cache
		groupCache:  make(map[string]map[string]interface{}),
		memberCache: make(map[string]map[string]interface{}),
		friendCache: make(map[string]map[string]interface{}),
		cacheMutex:  sync.RWMutex{},

		// User Management
		users:      make(map[string]*User),
		usersMutex: sync.RWMutex{},
	}
	// 初始化数据库
	if err := m.initDB(); err != nil {
		log.Printf("[ERROR] 数据库初始化失败: %v", err)
	} else {
		// 从数据库加载用户
		if err := m.loadUsersFromDB(); err != nil {
			log.Printf("[WARN] 从数据库加载用户失败: %v", err)
		}
	}

	// 初始化默认管理员用户 (如果不存在)
	m.usersMutex.Lock()
	m.initDefaultAdmin()
	m.usersMutex.Unlock()

	// 初始化Redis (用于统计信息等非持久化数据)
	m.rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PWD,
		DB:       0,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] 无法连接到Redis: %v", err)
		m.rdb = nil
	} else {
		log.Printf("[INFO] 已连接到Redis")
	}

	// 初始化Docker客户端
	cli, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("[WARN] 无法初始化Docker客户端: %v", err)
	} else {
		m.dockerClient = cli
		log.Printf("[INFO] Docker客户端已初始化")
	}

	return m
}

// 注意：超时检测和统计重置方法已从其他文件导入，此处不再重复定义
