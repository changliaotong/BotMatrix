// BotNexus - 统一构建入口文件
package main

import (
	"BotMatrix/common"
	"BotNexus/tasks"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// 版本号定义
const VERSION = "86"

// Manager 是 BotNexus 本地的包装结构，允许在其上定义方法
type Manager struct {
	*common.Manager
	Core        *CorePlugin
	TaskManager *tasks.TaskManager
}

// 主函数 - 整合所有功能
func main() {
	// 初始化翻译器
	common.InitTranslator("locales", "zh-CN")

	log.Printf(common.T("", "server_starting"), VERSION)

	// 创建管理器 (内部会初始化数据库和管理员)
	manager := NewManager()

	// 启动超时检测
	go manager.StartWorkerTimeoutDetection()
	go manager.StartBotTimeoutDetection()

	// 启动统计信息收集
	go manager.StartTrendCollection()

	// 启动幂等性缓存清理
	go manager.StartIdempotencyCleanup()

	// 启动统计信息重置和定期保存
	go manager.StartPeriodicStatsSave()

	// 启动 Core Gateway (WebSocket 转发引擎 - 仅处理机器人和工作节点连接)
	coreMux := manager.createCoreHandler()
	go func() {
		log.Printf(common.T("", "core_engine_starting"), common.WS_PORT)
		if err := http.ListenAndServe(common.WS_PORT, coreMux); err != nil {
			log.Fatalf(common.T("", "core_engine_failed"), err)
		}
	}()

	// 启动管理后台 HTTP 服务
	webuiPort := manager.Config.WebUIPort
	if webuiPort == "" {
		webuiPort = ":5000"
	}

	mux := http.NewServeMux()

	// 公开接口
	mux.HandleFunc("/login", HandleLogin(manager.Manager))
	mux.HandleFunc("/api/login", HandleLogin(manager.Manager))

	// 需要登录的接口
	mux.HandleFunc("/api/me", manager.JWTMiddleware(HandleGetUserInfo(manager.Manager)))
	mux.HandleFunc("/api/user/info", manager.JWTMiddleware(HandleGetUserInfo(manager.Manager)))
	mux.HandleFunc("/api/user/password", manager.JWTMiddleware(HandleChangePassword(manager.Manager)))

	mux.HandleFunc("/api/bots", manager.JWTMiddleware(HandleGetBots(manager.Manager)))
	mux.HandleFunc("/api/workers", manager.JWTMiddleware(HandleGetWorkers(manager.Manager)))
	mux.HandleFunc("/api/proxy/avatar", HandleProxyAvatar(manager.Manager))
	mux.HandleFunc("/api/stats", manager.JWTMiddleware(HandleGetStats(manager.Manager)))
	mux.HandleFunc("/api/stats/chat", manager.JWTMiddleware(HandleGetChatStats(manager.Manager)))
	mux.HandleFunc("/api/system/stats", manager.JWTMiddleware(HandleGetSystemStats(manager.Manager)))
	mux.HandleFunc("/api/logs", manager.JWTMiddleware(HandleGetLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/logs", manager.JWTMiddleware(HandleGetLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/logs/clear", manager.AdminMiddleware(HandleClearLogs(manager.Manager)))
	mux.HandleFunc("/api/contacts", manager.JWTMiddleware(HandleGetContacts(manager.Manager)))
	mux.HandleFunc("/api/action", manager.JWTMiddleware(HandleSendAction(manager.Manager)))
	mux.HandleFunc("/api/smart_action", manager.JWTMiddleware(HandleSendAction(manager.Manager)))

	// 任务系统接口
	mux.HandleFunc("/api/tasks", manager.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleListTasks(manager)(w, r)
		case http.MethodPost:
			HandleCreateTask(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/tasks/executions", manager.JWTMiddleware(HandleGetExecutions(manager)))
	mux.HandleFunc("/api/ai/parse", manager.JWTMiddleware(HandleAIParse(manager)))
	mux.HandleFunc("/api/ai/confirm", manager.JWTMiddleware(HandleAIConfirm(manager)))
	mux.HandleFunc("/api/system/capabilities", manager.JWTMiddleware(HandleGetCapabilities(manager)))
	mux.HandleFunc("/api/tags", manager.JWTMiddleware(HandleManageTags(manager)))

	// 管理员接口
	mux.HandleFunc("/api/admin/config", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetConfig(manager.Manager)(w, r)
		case http.MethodPost:
			HandleUpdateConfig(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/redis/config", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetRedisConfig(manager.Manager)(w, r)
		case http.MethodPost:
			HandleUpdateRedisConfig(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/users", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleAdminListUsers(manager.Manager)(w, r)
		case http.MethodPost:
			HandleAdminManageUsers(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/routing", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetRoutingRules(manager.Manager)(w, r)
		case http.MethodPost:
			HandleSetRoutingRule(manager.Manager)(w, r)
		case http.MethodDelete:
			HandleDeleteRoutingRule(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// Docker 接口
	mux.HandleFunc("/api/docker/list", manager.AdminMiddleware(HandleDockerList(manager.Manager)))
	mux.HandleFunc("/api/docker/containers", manager.AdminMiddleware(HandleDockerList(manager.Manager)))
	mux.HandleFunc("/api/docker/start", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/docker/stop", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/docker/restart", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/docker/remove", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/docker/add-bot", manager.AdminMiddleware(HandleDockerAddBot(manager.Manager)))
	mux.HandleFunc("/api/docker/add-worker", manager.AdminMiddleware(HandleDockerAddWorker(manager.Manager)))
	mux.HandleFunc("/api/docker/logs", manager.AdminMiddleware(HandleDockerLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/list", manager.AdminMiddleware(HandleDockerList(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/action", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/logs", manager.AdminMiddleware(HandleDockerLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/add-bot", manager.AdminMiddleware(HandleDockerAddBot(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/add-worker", manager.AdminMiddleware(HandleDockerAddWorker(manager.Manager)))

	// 管理员接口 - 帮助手册
	mux.HandleFunc("/api/admin/manual", manager.AdminMiddleware(HandleGetManual(manager.Manager)))

	// WebSocket 接口 (管理后台 UI 使用)
	mux.HandleFunc("/ws/subscriber", manager.JWTMiddleware(HandleSubscriberWebSocket(manager.Manager)))

	// 静态文件服务
	webDir := "../WebUI/web"
	if _, err := os.Stat("./web"); err == nil {
		webDir = "./web"
	} else if _, err := os.Stat("/app/web"); err == nil {
		webDir = "/app/web"
	}
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	// Overmind (Flutter Web) 静态服务
	overmindDir := "../WebUI/overmind"
	if _, err := os.Stat("./overmind"); err == nil {
		overmindDir = "./overmind"
	} else if _, err := os.Stat("/app/overmind"); err == nil {
		overmindDir = "/app/overmind"
	}
	mux.Handle("/overmind/", http.StripPrefix("/overmind/", http.FileServer(http.Dir(overmindDir))))

	log.Printf(common.T("", "admin_starting"), webuiPort)
	if err := http.ListenAndServe(webuiPort, mux); err != nil {
		log.Fatalf(common.T("", "admin_failed"), err)
	}
}

func (m *Manager) createCoreHandler() http.Handler {
	mux := http.NewServeMux()

	// 仅处理转发核心的 WebSocket 连接
	mux.HandleFunc("/ws/bots", m.handleBotWebSocket)
	mux.HandleFunc("/ws/workers", m.handleWorkerWebSocket)

	return mux
}

// 简化的管理器创建函数
func NewManager() *Manager {
	m := &Manager{
		Manager: common.NewManager(),
	}

	// 初始化配置指针，指向全局配置
	if m.Config == nil {
		m.Config = common.GlobalConfig
	}

	// 初始化数据库
	if err := m.InitDB(); err != nil {
		log.Printf(common.T("", "db_init_failed"), err)
	} else {
		// 从数据库加载路由规则
		if err := m.LoadRoutingRulesFromDB(); err != nil {
			log.Printf(common.T("", "load_route_rules_failed"), err)
		}
		// 从数据库加载联系人缓存
		if err := m.LoadCachesFromDB(); err != nil {
			log.Printf(common.T("", "load_contacts_failed"), err)
		}
		// 从数据库加载系统统计
		if err := m.LoadStatsFromDB(); err != nil {
			log.Printf(common.T("", "load_stats_failed"), err)
		}
	}

	// 初始化Redis (用于统计信息等非持久化数据)
	m.Rdb = redis.NewClient(&redis.Options{
		Addr:     common.REDIS_ADDR,
		Password: common.REDIS_PWD,
		DB:       0,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.Rdb.Ping(ctx).Err(); err != nil {
		log.Printf(common.T("", "redis_conn_failed"), err)
		m.Rdb = nil
	} else {
		log.Printf(common.T("", "redis_connected"))
	}

	// 初始化核心插件
	m.Core = NewCorePlugin(m)

	// 初始化任务管理器
	m.TaskManager = tasks.NewTaskManager(m.DB, m)
	m.TaskManager.Start()

	return m
}

// 实现 tasks.BotManager 接口
func (m *Manager) SendBotAction(botID string, action string, params map[string]interface{}) error {
	m.Mutex.RLock()
	bot, exists := m.Bots[botID]
	m.Mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot %s not found", botID)
	}

	echo := fmt.Sprintf("task|%d|%s", time.Now().UnixNano(), action)
	msg := map[string]interface{}{
		"action": action,
		"params": params,
		"echo":   echo,
	}

	bot.Mutex.Lock()
	defer bot.Mutex.Unlock()
	return bot.Conn.WriteJSON(msg)
}

func (m *Manager) GetTags(targetType string, targetID string) []string {
	tags, _ := m.TaskManager.Tagging.GetTagsByTarget(targetType, targetID)
	return tags
}

func (m *Manager) GetTargetsByTags(targetType string, tags []string, logic string) []string {
	targets, _ := m.TaskManager.Tagging.GetTargetsByTags(targetType, tags, logic)
	return targets
}
