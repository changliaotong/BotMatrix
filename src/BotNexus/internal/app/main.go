// BotNexus - 统一构建入口文件
// @title BotNexus API
// @version 1.0
// @description BotNexus 管理后台 API 接口文档
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package app

import (
	"BotMatrix/common"
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/b2b"
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/ai/mcp"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/bot"
	"BotMatrix/common/config"
	clog "BotMatrix/common/log"
	"BotMatrix/common/middleware"
	"BotMatrix/common/models"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/tasks"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"

	// "BotNexus/internal/config"
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

	"github.com/botuniverse/go-libonebot"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 版本号定义
const VERSION = "87"

// 插件系统集成
var pluginManager *core.PluginManager

// LogManager 用于捕获日志并显示在 Web UI
type LogManager struct {
	logs    []string
	max     int
	mu      sync.Mutex
	manager *bot.Manager
}

func (lm *LogManager) SetManager(m *bot.Manager) {
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
	clog.Info("重启 BotNexus...")
	os.Exit(0)
}

// PlatformAdapter 定义第三方平台适配器接口 (公众号, 抖音等)
type PlatformAdapter interface {
	GetPlatform() string
	HandleWebhook(w http.ResponseWriter, r *http.Request)
	SendMessage(targetID string, content string) error
}

// AIIntegrationService 定义 AI 调度与管理接口
type AIIntegrationService interface {
	types.AIService
}

// Manager 核心管理器
type Manager struct {
	*bot.Manager
	Core                       *common.CorePlugin
	TaskManager                *tasks.TaskManager
	OneBot                     *libonebot.OneBot
	PluginManager              *core.PluginManager
	PlatformAdapters           map[string]PlatformAdapter
	B2BService                 b2b.B2BService
	AIIntegrationService       AIIntegrationService
	DigitalEmployeeService     employee.DigitalEmployeeService
	DigitalEmployeeTaskService employee.DigitalEmployeeTaskService
	DigitalEmployeeKPIService  employee.DigitalEmployeeKPIService
	CognitiveMemoryService     employee.CognitiveMemoryService
	KnowledgeBase              types.KnowledgeBase
	pendingSkillRes            sync.Map // map[string]chan any
	MCPManager                 *mcp.MCPManager

	// 测试钩子
	OnCommandSent func(workerID string, msg types.WorkerCommand)
}

// NewInternalMessage 构造一个统一的内部消息
func (m *Manager) NewInternalMessage(platform, protocol, botID, userID, content string) types.InternalMessage {
	return types.InternalMessage{
		ID:         fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Time:       time.Now().Unix(),
		Platform:   platform,
		Protocol:   protocol,
		SelfID:     botID,
		UserID:     userID,
		RawMessage: content,
		Extras:     make(map[string]any),
	}
}

// Run 启动 BotNexus
func Run() {
	// 初始化配置 (优先从当前目录 config.json 加载)
	config.InitConfig("config.json")

	// 初始化日志系统
	clog.InitDefaultLogger()

	// 初始化翻译器
	utils.InitTranslator("locales", "zh-CN")

	clog.Info(utils.T("", "server_starting", VERSION))

	// 创建管理器 (内部会初始化数据库 and 管理员)
	manager := NewManager()
	logMgr.SetManager(manager.Manager)

	// 初始化插件系统 (中心插件：如数据统计、消息拦截等)
	manager.PluginManager = core.NewPluginManager()
	centralPluginsDir := filepath.Join("..", "..", "plugins", "central")
	// 确保目录存在
	if _, err := os.Stat(centralPluginsDir); os.IsNotExist(err) {
		os.MkdirAll(centralPluginsDir, 0755)
	}
	if err := manager.PluginManager.LoadPlugins(centralPluginsDir); err != nil {
		clog.Error("加载中心插件失败", zap.Error(err))
	}

	// 启动超时检测 (Nexus 维护连接状态)
	go manager.StartWorkerTimeoutDetection()
	go manager.StartBotTimeoutDetection()

	// 启动定期保存统计信息 (每天凌晨固化昨天的 Redis 数据到 DB)
	go manager.StartPeriodicStatsSave()

	// 启动幂等性缓存清理
	go manager.StartIdempotencyCleanup()

	// 启动配置缓存刷新 (从 Redis 同步 RateLimit 和 TTL 配置)
	go manager.StartConfigCacheRefresh()

	// 统计信息重置和保存已迁移至 BotWorker
	// go manager.StartPeriodicStatsSave()

	// 启动 Core Gateway (WebSocket 转发引擎 - 仅处理机器人和工作节点连接)
	coreMux := manager.createCoreHandler()
	go func() {
		clog.Info(utils.T("", "core_engine_starting", config.WS_PORT))
		if err := http.ListenAndServe(config.WS_PORT, coreMux); err != nil {
			clog.Error(utils.T("", "core_engine_failed", err))
		}
	}()

	// 启动管理后台 HTTP 服务
	webuiPort := manager.Config.WebUIPort
	if webuiPort == "" {
		webuiPort = ":5000"
	}

	mux := http.NewServeMux()

	// --- 1. 公开接口 (无需认证) ---
	mux.HandleFunc("/api/login", HandleLogin(manager.Manager))
	mux.HandleFunc("/api/register", HandleRegister(manager.Manager))
	mux.HandleFunc("/api/auth/token-login", HandleTokenLogin(manager.Manager))
	mux.HandleFunc("/api/proxy/avatar", HandleProxyAvatar(manager.Manager))

	// AI 相关的请求转发给 Worker 处理 (Nexus 仅做代理)
	mux.HandleFunc("/api/ai/chat/stream", manager.SkillMiddleware(ai.HandleAIChatStream(manager)))
	mux.HandleFunc("/api/ai/", manager.handleWorkerProxy)
	mux.HandleFunc("/api/knowledge/", manager.handleWorkerProxy)
	mux.HandleFunc("/api/admin/ai/", manager.handleWorkerProxy)
	mux.HandleFunc("/api/admin/employees", manager.handleWorkerProxy)
	mux.HandleFunc("/api/admin/departments", manager.handleWorkerProxy)
	mux.HandleFunc("/api/admin/role-templates", manager.handleWorkerProxy)

	// --- 2. 普通用户接口 (仅需 JWT 认证) ---
	// 用户个人信息
	mux.HandleFunc("/api/me", manager.JWTMiddleware(HandleGetUserInfo(manager.Manager)))
	mux.HandleFunc("/api/user/info", manager.JWTMiddleware(HandleGetUserInfo(manager.Manager)))
	mux.HandleFunc("/api/user/profile", manager.JWTMiddleware(HandleUpdateUserProfile(manager.Manager)))
	mux.HandleFunc("/api/user/password", manager.JWTMiddleware(HandleChangePassword(manager.Manager)))

	// 基础统计与状态 (用户可见版本)
	mux.HandleFunc("/api/stats", manager.JWTMiddleware(HandleGetStats(manager.Manager)))
	mux.HandleFunc("/api/stats/chat", manager.JWTMiddleware(HandleGetChatStats(manager.Manager)))
	mux.HandleFunc("/api/nexus/status", manager.JWTMiddleware(HandleGetNexusStatus(manager.Manager)))

	// 机器人与联系人 (用户可见版本)
	mux.HandleFunc("/api/bots", manager.JWTMiddleware(HandleGetBots(manager.Manager)))
	mux.HandleFunc("/api/workers", manager.JWTMiddleware(HandleGetWorkers(manager.Manager)))
	mux.HandleFunc("/api/contacts", manager.JWTMiddleware(HandleGetContacts(manager.Manager)))
	mux.HandleFunc("/api/action", manager.JWTMiddleware(HandleSendAction(manager)))
	mux.HandleFunc("/api/smart_action", manager.JWTMiddleware(HandleSendAction(manager)))

	// 任务与能力
	mux.HandleFunc("/api/tasks", manager.JWTMiddleware(manager.SkillMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			common.HandleListTasks(manager.Manager)(w, r)
		case http.MethodPost:
			common.HandleCreateTask(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/api/tasks/executions", manager.JWTMiddleware(manager.SkillMiddleware(common.HandleGetExecutions(manager.Manager))))
	mux.HandleFunc("/api/system/capabilities", manager.JWTMiddleware(manager.SkillMiddleware(common.HandleGetCapabilities(manager.Manager))))
	mux.HandleFunc("/api/tags", manager.JWTMiddleware(manager.SkillMiddleware(common.HandleManageTags(manager.Manager))))

	// --- 3. 管理员接口 (需 Admin 权限) ---
	// 系统监控与统计 (高权限版本)
	mux.HandleFunc("/api/admin/stats", manager.AdminMiddleware(HandleGetStats(manager.Manager)))
	mux.HandleFunc("/api/admin/system/stats", manager.AdminMiddleware(HandleGetSystemStats(manager.Manager)))
	mux.HandleFunc("/api/admin/nexus/status", manager.AdminMiddleware(HandleGetNexusStatus(manager.Manager)))

	// 资源管理
	mux.HandleFunc("/api/admin/bots", manager.AdminMiddleware(HandleGetBots(manager.Manager)))
	mux.HandleFunc("/api/admin/setup/members", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetMemberSetup(manager.Manager)(w, r)
		case http.MethodPut:
			HandleUpdateMemberSetup(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/setup/groups", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetGroupSetup(manager.Manager)(w, r)
		case http.MethodPut:
			HandleUpdateGroupSetup(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/workers", manager.AdminMiddleware(HandleGetWorkers(manager.Manager)))
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

	// 消息与联系人管理
	mux.HandleFunc("/api/admin/contacts", manager.AdminMiddleware(HandleGetContacts(manager.Manager)))
	mux.HandleFunc("/api/admin/contacts/sync", manager.AdminMiddleware(HandleGetContacts(manager.Manager)))
	mux.HandleFunc("/api/admin/group/members", manager.AdminMiddleware(HandleGetGroupMembers(manager.Manager)))
	mux.HandleFunc("/api/admin/messages", manager.AdminMiddleware(HandleGetMessages(manager.Manager)))
	mux.HandleFunc("/api/admin/batch_send", manager.AdminMiddleware(HandleBatchSend(manager)))

	// 系统日志
	mux.HandleFunc("/api/admin/logs", manager.AdminMiddleware(HandleGetLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/logs/clear", manager.AdminMiddleware(HandleClearLogs(manager.Manager)))

	// 数字员工审计与审批
	mux.HandleFunc("/api/admin/audit/tools", manager.AdminMiddleware(HandleToolAuditActions(manager.Manager)))
	mux.HandleFunc("/api/admin/audit/tools/approve", manager.AdminMiddleware(HandleToolAuditActions(manager.Manager)))
	mux.HandleFunc("/api/admin/audit/tools/reject", manager.AdminMiddleware(HandleToolAuditActions(manager.Manager)))

	// 系统配置
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
	mux.HandleFunc("/api/admin/manual", manager.AdminMiddleware(HandleGetManual(manager.Manager)))

	// 基础设施与插件
	mux.HandleFunc("/api/admin/docker/list", manager.AdminMiddleware(HandleDockerList(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/action", manager.AdminMiddleware(HandleDockerAction(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/logs", manager.AdminMiddleware(HandleDockerLogs(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/add-bot", manager.AdminMiddleware(HandleDockerAddBot(manager.Manager)))
	mux.HandleFunc("/api/admin/docker/add-worker", manager.AdminMiddleware(HandleDockerAddWorker(manager.Manager)))

	mux.HandleFunc("/api/admin/plugins/list", manager.AdminMiddleware(HandleListPlugins(manager)))
	mux.HandleFunc("/api/admin/plugins/action", manager.AdminMiddleware(HandlePluginAction(manager)))
	mux.HandleFunc("/api/admin/plugins/install", manager.AdminMiddleware(HandleInstallPlugin(manager)))
	mux.HandleFunc("/api/admin/plugins/delete", manager.AdminMiddleware(HandleDeletePlugin(manager)))

	// 裂变营销系统
	mux.HandleFunc("/api/admin/fission/config", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			common.HandleGetFissionConfig(manager.Manager)(w, r)
		case http.MethodPost:
			common.HandleUpdateFissionConfig(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/tasks", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			common.HandleGetFissionTasks(manager.Manager)(w, r)
		case http.MethodPost:
			common.HandleSaveFissionTask(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/tasks/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			common.HandleDeleteFissionTask(manager.Manager)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/fission/stats", manager.AdminMiddleware(common.HandleGetFissionStats(manager.Manager)))
	mux.HandleFunc("/api/admin/fission/invitations", manager.AdminMiddleware(common.HandleGetInvitations(manager.Manager)))
	mux.HandleFunc("/api/admin/fission/leaderboard", manager.AdminMiddleware(common.HandleGetFissionLeaderboard(manager.Manager)))

	// --- MCP Server 接口 (Global Agent Mesh 核心) ---
	mux.HandleFunc("/api/mcp/v1/sse", mcp.HandleMCPSSE(manager))
	mux.HandleFunc("/api/mcp/v1/tools", manager.B2BMiddleware(mcp.HandleMCPListTools(manager)))
	mux.HandleFunc("/api/mcp/v1/tools/call", manager.B2BMiddleware(mcp.HandleMCPCallTool(manager)))

	// --- Global Agent Mesh 发现与连接接口 ---
	mux.HandleFunc("/api/mesh/discover", manager.JWTMiddleware(employee.HandleMeshDiscover(manager.B2BService)))
	mux.HandleFunc("/api/mesh/register", manager.AdminMiddleware(employee.HandleMeshRegister(manager.B2BService)))
	mux.HandleFunc("/api/mesh/connect", manager.AdminMiddleware(employee.HandleMeshConnect(manager.B2BService)))
	mux.HandleFunc("/api/mesh/call", manager.AdminMiddleware(employee.HandleMeshCall(manager.B2BService)))
	mux.HandleFunc("/api/b2b/handshake", manager.DigitalEmployeeMiddleware(employee.HandleB2BHandshake(manager.B2BService)))
	mux.HandleFunc("/api/b2b/skills/request", manager.JWTMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BRequestSkill(manager.B2BService))))
	mux.HandleFunc("/api/b2b/skills/approve", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BApproveSkill(manager.B2BService))))
	mux.HandleFunc("/api/b2b/skills/list", manager.JWTMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BListSkills(manager.B2BService))))
	mux.HandleFunc("/api/b2b/dispatch", manager.JWTMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BDispatchEmployee(manager.B2BService))))
	mux.HandleFunc("/api/b2b/dispatch/approve", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BApproveDispatch(manager.B2BService))))
	mux.HandleFunc("/api/b2b/dispatch/list", manager.JWTMiddleware(manager.DigitalEmployeeMiddleware(employee.HandleB2BListDispatches(manager.B2BService))))

	mux.HandleFunc("/api/admin/debug/fix-data", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 将所有 agent 的 model_id 设置为 1 (假设 ID 1 的模型存在)
		if err := manager.GORMDB.Model(&models.AIAgentGORM{}).Where("model_id = ?", 0).Update("model_id", 1).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "Fixed model_id for agents", nil)
	}))

	mux.HandleFunc("/api/admin/ai/providers", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// HandleListAIProviders(manager)(w, r)
		case http.MethodPost:
			// HandleSaveAIProvider(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/ai/providers/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			// HandleDeleteAIProvider(manager)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/ai/models", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// HandleListAIModels(manager)(w, r)
		case http.MethodPost:
			// HandleSaveAIModel(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/ai/models/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			// HandleDeleteAIModel(manager)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/ai/agents", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// HandleListAIAgents(manager)(w, r)
		case http.MethodPost:
			// HandleSaveAIAgent(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/ai/agents/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// HandleGetAIAgent(manager)(w, r)
		case http.MethodDelete:
			// HandleDeleteAIAgent(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// --- 数字员工管理接口 ---
	// (已迁移到 Worker，Nexus 仅做代理)
	/*
		mux.HandleFunc("/api/admin/employees", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				HandleListEmployees(manager)(w, r)
			case http.MethodPost:
				HandleSaveEmployee(manager)(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})))
		mux.HandleFunc("/api/admin/employees/kpi", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				HandleGetEmployeeKPI(manager)(w, r)
			case http.MethodPost:
				HandleRecordEmployeeKpi(manager)(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})))

		mux.HandleFunc("/api/admin/department/kpi", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleGetDepartmentKPISummary(manager))))
		mux.HandleFunc("/api/admin/employees/optimize", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleOptimizeEmployee(manager))))
		mux.HandleFunc("/api/admin/employees/tasks", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleListEmployeeTasks(manager))))
		mux.HandleFunc("/api/admin/employees/status", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleUpdateEmployeeStatus(manager))))
		mux.HandleFunc("/api/admin/departments", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleListDepartments(manager))))
		mux.HandleFunc("/api/admin/role-templates", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleListRoleTemplates(manager))))
	*/

	// --- B2B 技能与连接管理 ---
	// (已迁移到 Worker)
	/*
		mux.HandleFunc("/api/admin/b2b/skills", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				HandleListB2BSkills(manager)(w, r)
			case http.MethodPost:
				HandleSaveB2BSkill(manager)(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})))
		mux.HandleFunc("/api/admin/b2b/skills/", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				HandleDeleteB2BSkill(manager)(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})))
		mux.HandleFunc("/api/admin/b2b/connections", manager.AdminMiddleware(manager.DigitalEmployeeMiddleware(HandleListB2BConnections(manager))))
	*/

	// --- 认知记忆管理接口 ---
	mux.HandleFunc("/api/admin/memories", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleListMemories(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/admin/memories/", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			HandleDeleteMemory(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// --- 认知记忆自主学习与管理 ---
	// (已迁移到 Worker)

	// AI 试用接口 (流式)
	// (已迁移到 Worker)

	// 静态文件服务已移除，WebUI 彻底分离独立运行
	// 如需 Nexus 托管静态文件，请恢复以下代码
	/*
		webDir := "../WebUI/web"
		if _, err := os.Stat("./web"); err == nil {
			webDir = "./web"
		} else if _, err := os.Stat("../WebUI/dist"); err == nil {
			webDir = "../WebUI/dist"
		} else if _, err := os.Stat("/app/web"); err == nil {
			webDir = "/app/web"
		}

		// 统一处理 WebUI 路由
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// 排除 API 和 WS 路径（理论上 ServeMux 会优先匹配更长的路径，但为了保险）
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
				http.NotFound(w, r)
				return
			}

			path := filepath.Join(webDir, r.URL.Path)
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				// 物理文件存在，正常服务
				http.ServeFile(w, r, path)
				return
			}
			// 物理文件不存在或为目录，返回 index.html 支持 SPA 路由
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
		})
	*/

	clog.Info(utils.T("", "admin_starting", webuiPort))
	// 使用 CORS 中间件包装 mux
	if err := http.ListenAndServe(webuiPort, middleware.CORSMiddleware(mux)); err != nil {
		clog.Fatal(utils.T("", "admin_failed", err))
	}
}

func (m *Manager) createCoreHandler() http.Handler {
	mux := http.NewServeMux()

	// 仅处理转发核心的 WebSocket 连接
	mux.HandleFunc("/ws/bots", m.handleBotWebSocket)
	mux.HandleFunc("/ws/workers", m.handleWorkerWebSocket)

	return mux
}

// SkillMiddleware 检查技能系统是否启用的中间件
func (m *Manager) SkillMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !config.ENABLE_SKILL || m.GORMDB == nil || m.TaskManager == nil {
			lang := r.Header.Get("Accept-Language")
			if lang == "" {
				lang = "zh-CN"
			}
			msg := utils.T(lang, "skill_system_disabled|技能系统已禁用。请在配置中设置 ENABLE_SKILL=true 以启用任务和策略功能。")

			// 使用 200 状态码以允许前端优雅地处理“服务已关闭”的提示，而不是触发网络错误
			w.WriteHeader(http.StatusOK)
			utils.SendJSONResponseWithCode(w, false, msg, "SKILL_DISABLED", nil)
			return
		}
		next(w, r)
	}
}

// DigitalEmployeeMiddleware 检查数字员工功能是否启用
func (m *Manager) DigitalEmployeeMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.Config.EnableDigitalEmployee {
			lang := r.Header.Get("Accept-Language")
			if lang == "" {
				lang = "zh-CN"
			}
			msg := utils.T(lang, "digital_employee_disabled|数字员工功能已禁用。请在配置中设置 EnableDigitalEmployee=true 以启用该功能。")

			w.WriteHeader(http.StatusOK)
			utils.SendJSONResponseWithCode(w, false, msg, "DIGITAL_EMPLOYEE_DISABLED", nil)
			return
		}
		next(w, r)
	}
}

// B2BMiddleware 企业间通信中间件
func (m *Manager) B2BMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return ai.B2BMiddleware(m, next)
}

// GetGORMDB 获取 GORM 数据库连接
func (m *Manager) GetGORMDB() *gorm.DB {
	return m.GORMDB
}

// GetKnowledgeBase 获取知识库服务
func (m *Manager) GetKnowledgeBase() types.KnowledgeBase {
	return m.KnowledgeBase
}

// GetAIService 获取 AI 服务
func (m *Manager) GetAIService() types.AIService {
	// AIIntegrationService 实现了 AIService
	if svc, ok := m.AIIntegrationService.(types.AIService); ok {
		return svc
	}
	return nil
}

// GetB2BService 获取 B2B 服务
func (m *Manager) GetB2BService() types.B2BService {
	return m.B2BService
}

// GetCognitiveMemoryService 获取认知记忆服务
func (m *Manager) GetCognitiveMemoryService() types.CognitiveMemoryService {
	return m.CognitiveMemoryService
}

// GetDigitalEmployeeService 获取数字员工服务
func (m *Manager) GetDigitalEmployeeService() types.DigitalEmployeeService {
	return m.DigitalEmployeeService
}

// GetDigitalEmployeeTaskService 获取数字员工任务服务
func (m *Manager) GetDigitalEmployeeTaskService() types.DigitalEmployeeTaskService {
	return m.DigitalEmployeeTaskService
}

// GetTaskManager 获取任务管理器
func (m *Manager) GetTaskManager() types.TaskManagerInterface {
	return m.TaskManager
}

// GetMCPManager 获取 MCP 管理器
func (m *Manager) GetMCPManager() types.MCPManagerInterface {
	return m.MCPManager
}

// 简化的管理器创建函数
func NewManager() *Manager {
	m := &Manager{
		Manager: bot.NewManager(),
	}

	// 初始化配置指针，指向全局配置
	if m.Config == nil {
		m.Config = config.GlobalConfig
	}

	// 初始化数据库
	if err := m.InitDB(); err != nil {
		clog.Error(utils.T("", "db_init_failed", err))
	} else {
		// 从数据库加载路由规则
		if err := m.LoadRoutingRulesFromDB(); err != nil {
			clog.Error(utils.T("", "load_route_rules_failed"), zap.Error(err))
		}
		// 从数据库加载联系人缓存
		if err := m.LoadCachesFromDB(); err != nil {
			clog.Error(utils.T("", "load_contacts_failed"), zap.Error(err))
		}
		// 从数据库加载用户
		if err := m.LoadUsersFromDB(); err != nil {
			clog.Error("加载用户失败", zap.Error(err))
		}
		// 从数据库加载系统统计
		if err := m.LoadStatsFromDB(); err != nil {
			clog.Error(utils.T("", "load_stats_failed"), zap.Error(err))
		}
		// 从数据库加载 Online 机器人
		if err := m.LoadBotsFromDB(); err != nil {
			clog.Error("加载在线机器人失败", zap.Error(err))
		}
		// 初始化默认岗位模板
		employee.InitDefaultRoleTemplates(m.GORMDB)
	}

	// 初始化Redis (用于统计信息等非持久化数据)
	m.Rdb = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR,
		Password: config.REDIS_PWD,
		DB:       0,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Rdb.Ping(ctx).Err(); err != nil {
		clog.Error(utils.T("", "redis_conn_failed"), zap.Error(err), zap.String("addr", config.REDIS_ADDR))
		m.Rdb = nil
	} else {
		clog.Info(utils.T("", "redis_connected"))
	}

	// 初始化 Docker 客户端
	if err := m.InitDockerClient(); err != nil {
		clog.Error("Docker 初始化失败", zap.Error(err))
	}

	// 初始化核心插件
	m.Core = common.NewCorePlugin(m.Manager)

	// 初始化平台适配器与服务实现
	m.PlatformAdapters = make(map[string]PlatformAdapter)

	// AI 业务逻辑已迁移到 Worker，Nexus 仅作为消息中转和基础管理
	/*
		if m.GORMDB != nil {
	*/

	// 初始化 OneBot v12 实现
	m.OneBot = libonebot.NewOneBot("botmatrix", &libonebot.Self{Platform: "nexus", UserID: "nexus"}, &libonebot.Config{})
	m.initOneBotActions()

	// 初始化测试钩子
	m.OnCommandSent = nil

	// 技能系统 (任务系统)
	if config.GlobalConfig.EnableSkill {
		clog.Info("[SKILL] 技能系统正在启动...")
		// GORM 已经在 InitDB 中初始化过了
		if m.GORMManager != nil && m.GORMManager.DB != nil {
			m.GORMDB = m.GORMManager.DB
			clog.Info("[GORM] 任务系统已准备就绪")

			// 初始化任务管理器 (仅在 GORMDB 成功初始化后)
			m.TaskManager = tasks.NewTaskManager(m.GORMDB, m.Rdb, m, "nexus")
			m.TaskManager.Executor = m // 设置执行器，用于处理群聊 AI 草稿确认
			if m.AIIntegrationService != nil {
				m.TaskManager.AI.SetAIService(m.AIIntegrationService)

				// 初始化 RAG 知识库 (PostgreSQL + pgvector)
				// 优先从配置中获取模型 ID，如果没有则尝试查找包含 embedding 关键字的模型
				var embedModel models.AIModelGORM
				var findErr error
				if config.GlobalConfig.AIEmbeddingModel != "" {
					findErr = m.GORMDB.Where("api_model_id = ?", config.GlobalConfig.AIEmbeddingModel).First(&embedModel).Error
				} else {
					findErr = m.GORMDB.Where("api_model_id LIKE ?", "%embedding%").First(&embedModel).Error
				}

				if findErr == nil {
					// 获取默认对话模型用于 RAG 2.0 (Query Refinement / Self-Reflection)
					var chatModel models.AIModelGORM
					m.GORMDB.Where("is_default = ?", true).First(&chatModel)
					if chatModel.ID == 0 {
						chatModel = embedModel // 兜底
					}

					es := rag.NewTaskAIEmbeddingService(m.AIIntegrationService, embedModel.ID, embedModel.ModelName)
					kb := rag.NewPostgresKnowledgeBase(m.GORMDB, es, m.AIIntegrationService, chatModel.ID)

					// 将向量服务注入认知记忆系统
					if aiSvc, ok := m.AIIntegrationService.(*ai.AIServiceImpl); ok {
						if ms := aiSvc.GetMemoryService(); ms != nil {
							ms.SetEmbeddingService(es)
						}
					}

					if err := kb.Setup(); err == nil {
						m.TaskManager.AI.Manifest.KnowledgeBase = kb

						// 注入到 MCP 管理器，供知识库工具使用
						if aiSvc, ok := m.AIIntegrationService.(*ai.AIServiceImpl); ok {
							aiSvc.SetKnowledgeBase(kb)
						}

						clog.Info("[RAG] 知识库已就绪", zap.String("model", embedModel.ModelName))

						// 自动同步系统文档
						go m.SyncSystemKnowledge()
					} else {
						clog.Warn("[RAG] 知识库初始化失败", zap.Error(err))
					}
				} else {
					clog.Warn("[RAG] 未找到可用的向量模型，RAG 功能将受限")
				}
			}
			// Nexus 作为任务系统的管理中心和调度触发端
			m.TaskManager.Start(true)
			clog.Info("[Nexus] TaskManager started (Scheduler Enabled)")

			// 启动 Redis 订阅监听 (用于接收 Worker 报备的能力)
			if m.Rdb != nil {
				go m.startRedisWorkerSubscription()
			}
		} else {
			clog.Error("[GORM] 任务系统启动失败：GORM 未初始化")
		}
	} else {
		clog.Info("[SKILL] 技能系统已禁用 (EnableSkill=false)")
	}

	return m
}

// SendToWorker 实现 tasks.BotManager 接口，支持 Redis 和 WebSocket 双通道发送
func (m *Manager) SendToWorker(workerID string, msg types.WorkerCommand) error {
	// 0. 尝试通过测试钩子发送
	if m.OnCommandSent != nil {
		m.OnCommandSent(workerID, msg)
		return nil
	}

	payload, _ := json.Marshal(msg)

	// 1. 尝试通过 Redis 发送 (仅在启用技能系统时)
	if config.ENABLE_SKILL && m.Rdb != nil {
		queue := "botmatrix:queue:default"
		if workerID != "" {
			queue = fmt.Sprintf("botmatrix:queue:worker:%s", workerID)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := m.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: queue,
			Values: map[string]interface{}{"payload": payload},
			MaxLen: 1000,
			Approx: true,
		}).Err()
		if err == nil {
			clog.Info("[Dispatcher] Sent message to worker via Redis Stream", zap.String("worker_id", workerID), zap.String("stream", queue))
			return nil
		}
		clog.Warn("[Dispatcher] Failed to send via Redis Stream. Falling back to WebSocket.", zap.Error(err))
	}

	// 2. 尝试通过 WebSocket 发送
	if workerID != "" {
		if w := m.findWorkerByID(workerID); w != nil {
			if w.Conn == nil {
				return fmt.Errorf("worker %s has no active websocket connection", workerID)
			}
			w.Mutex.Lock()
			err := w.Conn.WriteJSON(msg)
			w.Mutex.Unlock()
			if err == nil {
				clog.Info("[Dispatcher] Sent message to worker via WebSocket", zap.String("worker_id", workerID))
				return nil
			}
			return fmt.Errorf("websocket send failed: %v", err)
		}
		return fmt.Errorf("worker %s not found (offline)", workerID)
	}

	return fmt.Errorf("no target worker specified and Redis is unavailable")
}

// SyncSkillCall 同步调用一个技能并等待结果
func (m *Manager) SyncSkillCall(ctx context.Context, skillName string, params map[string]any) (any, error) {
	// 1. 寻找具备该能力的 Worker
	var workerID string
	m.Mutex.RLock()
	for _, w := range m.Workers {
		for _, cap := range w.Capabilities {
			if cap.Name == skillName {
				workerID = w.ID
				break
			}
		}
		if workerID != "" {
			break
		}
	}
	m.Mutex.RUnlock()

	if workerID == "" {
		return nil, fmt.Errorf("no worker available for skill: %s", skillName)
	}

	// 2. 生成唯一的 Correlation ID
	correlationID := fmt.Sprintf("sync_%d_%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)

	// 3. 准备接收结果的 Channel
	resChan := make(chan any, 1)
	m.pendingSkillRes.Store(correlationID, resChan)
	defer m.pendingSkillRes.Delete(correlationID)

	// 4. 发送指令
	cmd := types.WorkerCommand{
		Type:          "skill_call",
		Skill:         skillName,
		Params:        params,
		CorrelationID: correlationID,
		Timestamp:     time.Now().Unix(),
	}

	if err := m.SendToWorker(workerID, cmd); err != nil {
		return nil, fmt.Errorf("failed to send skill call: %w", err)
	}

	// 5. 等待结果或超时
	select {
	case res := <-resChan:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("skill call timeout")
	}
}

// HandleSkillResult 统一处理技能执行结果 (由 Redis 订阅或 WebSocket 触发)
func (m *Manager) HandleSkillResult(skillResult types.SkillResult) {
	taskIDStr := fmt.Sprint(skillResult.TaskID)
	executionID := fmt.Sprint(skillResult.ExecutionID)
	correlationID := skillResult.CorrelationID
	statusStr := skillResult.Status
	result := skillResult.Result
	errStr := skillResult.Error
	workerID := skillResult.WorkerID

	clog.Info("[Task] Received skill result",
		zap.String("worker_id", workerID),
		zap.String("task_id", taskIDStr),
		zap.String("execution_id", executionID),
		zap.String("correlation_id", correlationID),
		zap.String("status", statusStr))

	// 1. 检查是否有正在等待同步结果的请求
	if correlationID != "" {
		if resChanVal, ok := m.pendingSkillRes.Load(correlationID); ok {
			if resChan, ok := resChanVal.(chan any); ok {
				if errStr != "" {
					resChan <- fmt.Errorf("%s", errStr)
				} else {
					resChan <- result
				}
				// 既然已经发送给同步等待者，是否还需要继续执行异步更新数据库逻辑？
				// 通常还是需要的，因为同步调用可能只是 AI 链路的一部分，任务系统仍需记录。
			}
		}
	}

	// 转换状态
	status := models.ExecSuccess
	if statusStr == "failed" {
		status = models.ExecFailed
	}

	// 更新执行状态
	updates := map[string]any{
		"status": status,
		"result": result,
	}
	if errStr != "" {
		updates["result"] = fmt.Sprintf("Error: %s\nResult: %s", errStr, result)
	}

	// 如果是成功或失败，设置实际完成时间
	now := time.Now()
	updates["actual_time"] = &now

	// 如果没有数据库连接，直接返回
	if m.GORMDB == nil {
		return
	}

	// 如果有 executionID，优先根据 executionID 更新
	if executionID != "" && executionID != "<nil>" && executionID != "0" {
		if err := m.GORMDB.Model(&models.Execution{}).Where("execution_id = ?", executionID).Updates(updates).Error; err != nil {
			clog.Error("[Task] Failed to update execution", zap.String("execution_id", executionID), zap.Error(err))
		}
	} else {
		// 否则根据 taskID 更新最新的执行记录 (兼容旧版 Worker)
		if err := m.GORMDB.Model(&models.Execution{}).Where("task_id = ?", taskIDStr).Order("created_at desc").Limit(1).Updates(updates).Error; err != nil {
			clog.Error("[Task] Failed to update execution for task", zap.String("task_id", taskIDStr), zap.Error(err))
		}
	}

	// 如果任务成功，可能需要更新 Task 表的最后运行时间
	if status == models.ExecSuccess {
		taskID, _ := strconv.ParseUint(taskIDStr, 10, 32)
		if taskID > 0 {
			m.GORMDB.Model(&models.Task{}).Where("id = ?", taskID).Update("last_run_time", &now)
		}
	}

	// 广播到 WebUI，确保前端能实时更新状态和显示结果
	m.BroadcastEvent(map[string]any{
		"type": "skill_result",
		"data": skillResult,
	})
}

// SyncSystemKnowledge (已迁移到 Worker)
func (m *Manager) SyncSystemKnowledge() {
}

func (m *Manager) startRedisWorkerSubscription() {
	ctx := context.Background()
	pubsub := m.Rdb.Subscribe(ctx, "botmatrix:worker:register", "botmatrix:worker:skill_result", config.REDIS_KEY_ACTION_QUEUE, "botmatrix:actions")
	defer pubsub.Close()

	clog.Info("[Redis] Subscribed to worker channels: register, skill_result, actions, botmatrix:actions")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			clog.Error("[Redis] Subscription error", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		clog.Info("[Redis] Received raw message", zap.String("channel", msg.Channel), zap.String("payload", msg.Payload))

		var rawMsg map[string]any
		if err := json.Unmarshal([]byte(msg.Payload), &rawMsg); err != nil {
			clog.Error("[Redis] Failed to unmarshal message", zap.Error(err))
			continue
		}

		msgType, _ := rawMsg["type"].(string)

		// 如果没有 type，但有 action 和 self_id，则视为 action (兼容模式)
		if msgType == "" {
			if _, ok := rawMsg["action"]; ok {
				if _, ok := rawMsg["self_id"]; ok {
					msgType = "action"
				}
			}
		}

		// 如果还是没有 type，但有 skill_id，则视为 skill_result
		if msgType == "" {
			if _, ok := rawMsg["skill_id"]; ok {
				msgType = "skill_result"
			}
		}

		switch msgType {
		case "action":
			// 处理来自 Worker 的 Action
			m.handleWorkerAction(rawMsg)

		case "api_response":
			// 处理来自机器人的 API 响应 (用于回复 Echo)
			// m.handleApiResponse(rawMsg)

		case "skill_result":
			var skillResult types.SkillResult
			payloadBytes, _ := json.Marshal(rawMsg)
			json.Unmarshal(payloadBytes, &skillResult)
			m.HandleSkillResult(skillResult)
		}
	}
}

func (m *Manager) handleWorkerAction(msg map[string]any) {
	actionType, _ := msg["action"].(string)
	selfID, _ := msg["self_id"].(string)
	if selfID == "" {
		selfID, _ = msg["bot_id"].(string) // 向后兼容
	}
	platform, _ := msg["platform"].(string)
	if platform == "" {
		platform = "qq" // Default platform
	}
	params := msg["params"]

	clog.Info("[WorkerAction] Received action from worker",
		zap.String("action", actionType),
		zap.String("platform", platform),
		zap.String("self_id", selfID),
		zap.Any("params", params))

	// 查找对应的机器人
	bot, exists := m.GetBot(platform, selfID)

	if !exists {
		// 尝试从持久化缓存中查找以获取更多信息
		m.CacheMutex.RLock()
		cachedBot, cachedExists := m.BotCache[selfID]
		m.CacheMutex.RUnlock()

		if cachedExists {
			clog.Warn("[WorkerAction] Bot is offline",
				zap.String("self_id", selfID),
				zap.String("platform", cachedBot.Platform),
				zap.String("nickname", cachedBot.Nickname))
		} else {
			clog.Warn("[WorkerAction] Bot not found in memory or cache",
				zap.String("self_id", selfID),
				zap.String("platform", platform))
		}
		return
	}

	// 构造 OneBot 请求
	req := map[string]any{
		"action":  actionType,
		"self_id": selfID,
	}

	if paramsMap, ok := params.(map[string]any); ok {
		req["params"] = paramsMap
		// 确保 params 中也包含 self_id，有些 OneBot 实现可能需要
		if _, exists := paramsMap["self_id"]; !exists {
			paramsMap["self_id"] = selfID
		}
	}

	if echo, ok := msg["echo"]; ok {
		req["echo"] = echo
	}

	// 打印最终发送给机器人的完整数据，用于调试
	clog.Info("[WorkerAction] Final request to bot",
		zap.String("self_id", selfID),
		zap.String("action", actionType),
		zap.Any("req", req))

	// 发送给机器人
	bot.Mutex.Lock()
	err := bot.Conn.WriteJSON(req)
	bot.Mutex.Unlock()

	if err != nil {
		clog.Error("[WorkerAction] Failed to send action to bot", zap.Error(err))
	} else {
		clog.Info("[WorkerAction] Successfully sent action to bot", zap.String("self_id", selfID))

		// 广播到 WebUI，确保前端能实时看到机器人发出的消息
		if actionType == "send_msg" || actionType == "send_group_msg" || actionType == "send_private_msg" {
			go func() {
				paramsMap, _ := params.(map[string]any)
				msgContent := utils.ToString(paramsMap["message"])
				userID := utils.ToString(paramsMap["user_id"])
				groupID := utils.ToString(paramsMap["group_id"])
				msgType := utils.ToString(paramsMap["message_type"])

				// 模拟 OneBot v11 消息事件，用于 WebUI 显示
				event := map[string]any{
					"post_type":    "message",
					"message_type": msgType,
					"sub_type":     "normal",
					"message_id":   fmt.Sprintf("reply_%d", time.Now().UnixNano()),
					"user_id":      bot.SelfID, // 发送者是机器人
					"target_id":    userID,     // 接收者是用户
					"group_id":     groupID,
					"message":      msgContent,
					"raw_message":  msgContent,
					"self_id":      bot.SelfID,
					"time":         time.Now().Unix(),
					"sender": map[string]any{
						"user_id":  bot.SelfID,
						"nickname": bot.Nickname,
						"role":     "bot",
					},
				}
				if msgType == "" {
					if groupID != "" {
						event["message_type"] = "group"
					} else {
						event["message_type"] = "private"
					}
				}
				m.BroadcastEvent(event)
			}()
		}
	}
}

// 实现 tasks.BotManager 接口
func (m *Manager) SendBotAction(botID string, action string, params any) error {
	m.Mutex.RLock()
	// 尝试直接查找 (可能已经是 platform:id 格式)
	bot, exists := m.Bots[botID]
	if !exists && !strings.Contains(botID, ":") {
		// 如果不是 platform:id 格式，尝试加上默认的 qq:
		botKey := fmt.Sprintf("qq:%s", botID)
		bot, exists = m.Bots[botKey]
	}
	m.Mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot %s not found (tried both direct and qq: prefix)", botID)
	}

	echo := fmt.Sprintf("task|%d|%s", time.Now().UnixNano(), action)
	msg := struct {
		Action string `json:"action"`
		Params any    `json:"params"`
		Echo   string `json:"echo"`
	}{
		Action: action,
		Params: params,
		Echo:   echo,
	}

	bot.Mutex.Lock()
	defer bot.Mutex.Unlock()

	clog.Info("[BotAction] Sending action to bot",
		zap.String("bot_id", botID),
		zap.String("action", action),
		zap.String("echo", echo))

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
