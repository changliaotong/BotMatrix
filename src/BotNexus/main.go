// BotNexus - 统一构建入口文件
package main

import (
	"BotMatrix/common"
	"BotMatrix/common/log"
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 版本号定义
const VERSION = "87"

// 插件系统集成
var pluginManager *PluginManager

// LogManager 用于捕获日志并显示在 Web UI
type LogManager struct {
	logs    []string
	max     int
	mu      sync.Mutex
	manager *common.Manager
}

func (lm *LogManager) SetManager(m *common.Manager) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.manager = m
}

func (lm *LogManager) Write(p []byte) (n int, err error) {
	lm.mu.Lock()
	msg := string(p)
	lm.logs = append(lm.logs, msg)
	if len(lm.logs) > lm.max {
		lm.logs = lm.logs[len(lm.logs)-lm.max:]
	}
	m := lm.manager
	lm.mu.Unlock()

	if m != nil {
		// 避免 AddLog 产生的 fmt.Printf 再次进入这里 (虽然 AddLog 现在用 fmt 了，但还是加个保险)
		// 提取日志级别，默认为 INFO
		level := "INFO"
		if strings.Contains(msg, "[DEBUG]") {
			level = "DEBUG"
		} else if strings.Contains(msg, "[WARN]") {
			level = "WARN"
		} else if strings.Contains(msg, "[ERROR]") {
			level = "ERROR"
		}

		// 移除可能的换行符，因为 AddLog 会处理显示
		cleanMsg := strings.TrimSpace(msg)
		if cleanMsg != "" {
			// 使用 go func 异步调用，防止死锁或阻塞主线程日志输出
			go m.AddLog(level, cleanMsg)
		}
	}

	return os.Stdout.Write(p)
}

func (lm *LogManager) GetLogs(lines int) string {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	start := 0
	if len(lm.logs) > lines {
		start = len(lm.logs) - lines
	}
	var res string
	for i := start; i < len(lm.logs); i++ {
		res += lm.logs[i]
	}
	return res
}

var logMgr = &LogManager{max: 1000}

func restartBot() {
	log.Info("重启 BotNexus...")
	os.Exit(0)
}

// Manager 是 BotNexus 本地的包装结构，允许在其上定义方法
type Manager struct {
	*common.Manager
	Core        *CorePlugin
	TaskManager *tasks.TaskManager
}

// 主函数 - 整合所有功能
func main() {
	// 初始化日志系统
	log.InitDefaultLogger()

	// 初始化翻译器
	common.InitTranslator("locales", "zh-CN")

	log.Info(common.T("", "server_starting"), zap.String("version", VERSION))

	// 创建管理器 (内部会初始化数据库和管理员)
	manager := NewManager()
	logMgr.SetManager(manager.Manager)

	// 初始化插件系统
	pluginManager := NewPluginManager()
	pluginsDir := filepath.Join("..", "..", "plugins")
	if err := pluginManager.LoadPlugins(pluginsDir); err != nil {
		log.Error("加载插件失败", zap.Error(err))
	}

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
		log.Info(common.T("", "core_engine_starting"), zap.String("port", common.WS_PORT))
		if err := http.ListenAndServe(common.WS_PORT, coreMux); err != nil {
			log.Error(common.T("", "core_engine_failed"), zap.Error(err))
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
	mux.HandleFunc("/api/admin/bots", manager.JWTMiddleware(HandleGetBots(manager.Manager)))
	mux.HandleFunc("/api/workers", manager.JWTMiddleware(HandleGetWorkers(manager.Manager)))
	mux.HandleFunc("/api/proxy/avatar", HandleProxyAvatar(manager.Manager))
	mux.HandleFunc("/api/stats", manager.JWTMiddleware(HandleGetStats(manager.Manager)))
	mux.HandleFunc("/api/admin/stats", manager.JWTMiddleware(HandleGetStats(manager.Manager)))
	mux.HandleFunc("/api/stats/chat", manager.JWTMiddleware(HandleGetChatStats(manager.Manager)))
	mux.HandleFunc("/api/system/stats", manager.JWTMiddleware(HandleGetSystemStats(manager.Manager)))
	mux.HandleFunc("/api/logs", manager.JWTMiddleware(HandleGetLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/logs", manager.JWTMiddleware(HandleGetLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/logs/clear", manager.AdminMiddleware(HandleClearLogs(manager.Manager)))
	mux.HandleFunc("/api/contacts", manager.JWTMiddleware(HandleGetContacts(manager.Manager)))
	mux.HandleFunc("/api/action", manager.JWTMiddleware(HandleSendAction(manager.Manager)))
	mux.HandleFunc("/api/smart_action", manager.JWTMiddleware(HandleSendAction(manager.Manager)))

	// 任务系统接口
	mux.HandleFunc("/api/admin/tasks/capabilities", manager.AdminMiddleware(HandleListSystemCapabilities(manager)))
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

	// 高级任务与策略接口
	mux.HandleFunc("/api/admin/strategies", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleListStrategies(manager)(w, r)
		case http.MethodPost:
			HandleSaveStrategy(manager)(w, r)
		case http.MethodDelete:
			HandleDeleteStrategy(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/identities", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleListIdentities(manager)(w, r)
		case http.MethodPost:
			HandleSaveIdentity(manager)(w, r)
		case http.MethodDelete:
			HandleDeleteIdentity(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/shadow-rules", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleListShadowRules(manager)(w, r)
		case http.MethodPost:
			HandleSaveShadowRule(manager)(w, r)
		case http.MethodDelete:
			HandleDeleteShadowRule(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

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

	// 裂变系统接口
	mux.HandleFunc("/api/admin/fission/config", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetFissionConfig(manager.Manager)(w, r)
		case http.MethodPost:
			HandleUpdateFissionConfig(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/tasks", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetFissionTasks(manager.Manager)(w, r)
		case http.MethodPost:
			HandleSaveFissionTask(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/tasks/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			HandleDeleteFissionTask(manager.Manager)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/stats", manager.AdminMiddleware(HandleGetFissionStats(manager.Manager)))
	mux.HandleFunc("/api/admin/fission/invitations", manager.AdminMiddleware(HandleGetInvitations(manager.Manager)))
	mux.HandleFunc("/api/admin/fission/leaderboard", manager.AdminMiddleware(HandleGetFissionLeaderboard(manager.Manager)))

	// --- WebSocket 接口 (仅供管理后台 UI 使用) ---
	mux.HandleFunc("/ws/subscriber", manager.JWTMiddleware(HandleSubscriberWebSocket(manager.Manager)))

	mux.HandleFunc("/api/admin/nexus/status", manager.JWTMiddleware(HandleGetNexusStatus(manager.Manager)))

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

	log.Info(common.T("", "admin_starting"), zap.String("port", webuiPort))
	if err := http.ListenAndServe(webuiPort, mux); err != nil {
		log.Fatal(common.T("", "admin_failed"), zap.Error(err))
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
		log.Error(common.T("", "db_init_failed"), zap.Error(err))
	} else {
		// 从数据库加载路由规则
		if err := m.LoadRoutingRulesFromDB(); err != nil {
			log.Error(common.T("", "load_route_rules_failed"), zap.Error(err))
		}
		// 从数据库加载联系人缓存
		if err := m.LoadCachesFromDB(); err != nil {
			log.Error(common.T("", "load_contacts_failed"), zap.Error(err))
		}
		// 从数据库加载系统统计
		if err := m.LoadStatsFromDB(); err != nil {
			log.Error(common.T("", "load_stats_failed"), zap.Error(err))
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
		log.Error(common.T("", "redis_conn_failed"), zap.Error(err))
		m.Rdb = nil
	} else {
		log.Info(common.T("", "redis_connected"))
	}

	// 初始化 Docker 客户端
	if err := m.InitDockerClient(); err != nil {
		log.Error("Docker 初始化失败", zap.Error(err))
	}

	// 初始化核心插件
	m.Core = NewCorePlugin(m)

	// 技能系统 (任务系统) 默认关闭，仅在 ENABLE_SKILL 为 true 时开启
	if common.ENABLE_SKILL {
		log.Info("[SKILL] 技能系统正在启动...")
		// 初始化GORM (任务系统需要)
		m.GORMManager = common.NewGORMManager()
		if err := m.GORMManager.InitGORM(); err != nil {
			log.Error("[GORM] 初始化失败", zap.Error(err))
		} else {
			m.GORMDB = m.GORMManager.DB
			log.Info("[GORM] 任务系统已准备就绪")
		}

		// 初始化任务管理器
		m.TaskManager = tasks.NewTaskManager(m.GORMDB, m.Rdb, m)
		m.TaskManager.Start()

		// 启动 Redis 订阅监听 (用于接收 Worker 报备的能力)
		if m.Rdb != nil {
			go m.startRedisWorkerSubscription()
		}
	} else {
		log.Info("[SKILL] 技能系统已禁用 (ENABLE_SKILL=false)")
	}

	return m
}

// SendToWorker 实现 tasks.BotManager 接口，支持 Redis 和 WebSocket 双通道发送
func (m *Manager) SendToWorker(workerID string, msg map[string]interface{}) error {
	payload, _ := json.Marshal(msg)

	// 1. 尝试通过 Redis 发送 (仅在启用技能系统时)
	if common.ENABLE_SKILL && m.Rdb != nil {
		queue := "botmatrix:queue:default"
		if workerID != "" {
			queue = fmt.Sprintf("botmatrix:queue:worker:%s", workerID)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := m.Rdb.RPush(ctx, queue, payload).Err()
		if err == nil {
			log.Info("[Dispatcher] Sent message to worker via Redis queue", zap.String("worker_id", workerID), zap.String("queue", queue))
			return nil
		}
		log.Warn("[Dispatcher] Failed to send via Redis. Falling back to WebSocket.", zap.Error(err))
	}

	// 2. 尝试通过 WebSocket 发送
	if workerID != "" {
		if w := m.findWorkerByID(workerID); w != nil {
			w.Mutex.Lock()
			err := w.Conn.WriteJSON(msg)
			w.Mutex.Unlock()
			if err == nil {
				log.Info("[Dispatcher] Sent message to worker via WebSocket", zap.String("worker_id", workerID))
				return nil
			}
			return fmt.Errorf("websocket send failed: %v", err)
		}
		return fmt.Errorf("worker %s not found (offline)", workerID)
	}

	return fmt.Errorf("no target worker specified and Redis is unavailable")
}

// HandleSkillResult 统一处理技能执行结果 (由 Redis 订阅或 WebSocket 触发)
func (m *Manager) HandleSkillResult(rawMsg map[string]interface{}) {
	taskIDStr := fmt.Sprint(rawMsg["task_id"])
	executionID := fmt.Sprint(rawMsg["execution_id"])
	statusStr := fmt.Sprint(rawMsg["status"])
	result := fmt.Sprint(rawMsg["result"])
	errStr := fmt.Sprint(rawMsg["error"])
	workerID := fmt.Sprint(rawMsg["worker_id"])

	log.Info("[Task] Received skill result", 
		zap.String("worker_id", workerID), 
		zap.String("task_id", taskIDStr), 
		zap.String("execution_id", executionID), 
		zap.String("status", statusStr))

	// 转换状态
	status := tasks.ExecSuccess
	if statusStr == "failed" {
		status = tasks.ExecFailed
	}

	// 更新执行状态
	updates := map[string]interface{}{
		"status": status,
		"result": result,
	}
	if errStr != "" {
		updates["result"] = fmt.Sprintf("Error: %s\nResult: %s", errStr, result)
	}

	// 如果是成功或失败，设置实际完成时间
	now := time.Now()
	updates["actual_time"] = &now

	// 如果有 executionID，优先根据 executionID 更新
	if executionID != "" && executionID != "<nil>" && executionID != "0" {
		if err := m.GORMDB.Model(&tasks.Execution{}).Where("execution_id = ?", executionID).Updates(updates).Error; err != nil {
			log.Error("[Task] Failed to update execution", zap.String("execution_id", executionID), zap.Error(err))
		}
	} else {
		// 否则根据 taskID 更新最新的执行记录 (兼容旧版 Worker)
		if err := m.GORMDB.Model(&tasks.Execution{}).Where("task_id = ?", taskIDStr).Order("created_at desc").Limit(1).Updates(updates).Error; err != nil {
			log.Error("[Task] Failed to update execution for task", zap.String("task_id", taskIDStr), zap.Error(err))
		}
	}

	// 如果任务成功，可能需要更新 Task 表的最后运行时间
	if status == tasks.ExecSuccess {
		taskID, _ := strconv.ParseUint(taskIDStr, 10, 32)
		if taskID > 0 {
			m.GORMDB.Model(&tasks.Task{}).Where("id = ?", taskID).Update("last_run_time", &now)
		}
	}
}

func (m *Manager) startRedisWorkerSubscription() {
	ctx := context.Background()
	pubsub := m.Rdb.Subscribe(ctx, "botmatrix:worker:register", "botmatrix:worker:skill_result")
	defer pubsub.Close()

	log.Info("[Redis] Subscribed to worker channels: register, skill_result")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Error("[Redis] Subscription error", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		var rawMsg map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &rawMsg); err != nil {
			log.Error("[Redis] Failed to unmarshal message", zap.Error(err))
			continue
		}

		msgType, _ := rawMsg["type"].(string)

		switch msgType {
		case "worker_register":
			var regMsg struct {
				WorkerID     string                    `json:"worker_id"`
				Capabilities []common.WorkerCapability `json:"capabilities"`
			}
			payloadBytes, _ := json.Marshal(rawMsg)
			if err := json.Unmarshal(payloadBytes, &regMsg); err != nil {
				log.Error("[Redis] Failed to unmarshal worker registration", zap.Error(err))
				continue
			}

			log.Info("[Redis] Received registration from worker", 
				zap.String("worker_id", regMsg.WorkerID), 
				zap.Int("capabilities_count", len(regMsg.Capabilities)))

			// 更新或添加 Worker 信息，确保调度器能找到它
			m.Mutex.Lock()
			found := false
			for i, w := range m.Workers {
				if w.ID == regMsg.WorkerID {
					m.Workers[i].Capabilities = regMsg.Capabilities
					m.Workers[i].LastHeartbeat = time.Now()
					found = true
					break
				}
			}
			if !found {
				m.Workers = append(m.Workers, &common.WorkerClient{
					ID:            regMsg.WorkerID,
					Capabilities:  regMsg.Capabilities,
					Connected:     time.Now(),
					LastHeartbeat: time.Now(),
				})
			}
			m.Mutex.Unlock()

			// 更新 AI 解析器的技能列表
			var skills []tasks.Capability
			for _, cap := range regMsg.Capabilities {
				skills = append(skills, tasks.Capability{
					Name:        cap.Name,
					Description: cap.Description,
					Example:     cap.Usage,
					Params:      cap.Params,
				})
			}
			m.TaskManager.AI.UpdateSkills(skills)

		case "skill_result":
			m.HandleSkillResult(rawMsg)
		}
	}
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
