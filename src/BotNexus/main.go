// BotNexus - 统一构建入口文件
package main

import (
	"BotMatrix/common"
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// 版本号定义
const VERSION = "86"

// LogManager 用于捕获日志并显示在 Web UI
type LogManager struct {
	logs []string
	max  int
	mu   sync.Mutex
}

func (lm *LogManager) Write(p []byte) (n int, err error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.logs = append(lm.logs, string(p))
	if len(lm.logs) > lm.max {
		lm.logs = lm.logs[len(lm.logs)-lm.max:]
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
	log.Println("重启 BotNexus...")
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
	// 设置日志输出
	log.SetOutput(logMgr)

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

	// 设置页面 (简单 Web UI)
	mux.HandleFunc("/config-ui", handleConfigUI)
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetConfig(manager.Manager)(w, r)
		case http.MethodPost:
			handleConfigUpdate(manager.Manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/logs", handleLogs)

	// --- WebSocket 接口 (仅供管理后台 UI 使用) ---
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

	// 初始化GORM (任务系统需要)
	m.GORMManager = common.NewGORMManager()
	if err := m.GORMManager.InitGORM(); err != nil {
		log.Printf("[GORM] 初始化失败: %v", err)
	} else {
		m.GORMDB = m.GORMManager.DB
		log.Printf("[GORM] 任务系统已准备就绪")
	}

	// 初始化任务管理器
	m.TaskManager = tasks.NewTaskManager(m.GORMDB, m.Rdb, m)
	m.TaskManager.Start()

	// 启动 Redis 订阅监听 (用于接收 Worker 报备的能力)
	if m.Rdb != nil {
		go m.startRedisWorkerSubscription()
	}

	return m
}

func (m *Manager) startRedisWorkerSubscription() {
	ctx := context.Background()
	pubsub := m.Rdb.Subscribe(ctx, "botmatrix:worker:register")
	defer pubsub.Close()

	log.Printf("[Redis] Subscribed to botmatrix:worker:register")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("[Redis] Subscription error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var regMsg struct {
			Type         string                    `json:"type"`
			WorkerID     string                    `json:"worker_id"`
			Capabilities []common.WorkerCapability `json:"capabilities"`
		}

		if err := json.Unmarshal([]byte(msg.Payload), &regMsg); err != nil {
			log.Printf("[Redis] Failed to unmarshal worker registration: %v", err)
			continue
		}

		if regMsg.Type == "worker_register" {
			log.Printf("[Redis] Received registration from worker: %s (%d capabilities)", regMsg.WorkerID, len(regMsg.Capabilities))

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
		}
	}
}

// Web UI 处理器
func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BotNexus 配置中心</title>
    <style>
        :root { --primary-color: #007bff; --success-color: #28a745; --danger-color: #dc3545; --bg-color: #f4f7f6; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: var(--bg-color); margin: 0; display: flex; height: 100vh; }
        .sidebar { width: 280px; background: #2c3e50; color: white; display: flex; flex-direction: column; }
        .sidebar-header { padding: 20px; font-size: 20px; font-weight: bold; border-bottom: 1px solid #34495e; }
        .nav-item { padding: 15px 20px; cursor: pointer; transition: background 0.2s; display: flex; align-items: center; gap: 10px; }
        .nav-item:hover { background: #34495e; }
        .nav-item.active { background: var(--primary-color); }
        .main-content { flex: 1; overflow-y: auto; padding: 30px; }
        .card { background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); padding: 25px; margin-bottom: 25px; }
        .section-title { font-size: 18px; font-weight: 600; margin-bottom: 20px; color: #2c3e50; display: flex; justify-content: space-between; align-items: center; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; color: #666; }
        input[type="text"], input[type="number"], input[type="password"], select { 
            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; 
        }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: 500; transition: opacity 0.2s; }
        .btn-primary { background: var(--primary-color); color: white; }
        .btn-success { background: var(--success-color); color: white; }
        .btn-danger { background: var(--danger-color); color: white; }
        .btn:hover { opacity: 0.9; }
        .logs-container { background: #1e1e1e; color: #d4d4d4; padding: 15px; border-radius: 6px; font-family: 'Consolas', monospace; height: 500px; overflow-y: auto; font-size: 13px; line-height: 1.5; }
        .checkbox-group { display: flex; align-items: center; gap: 8px; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">BotNexus Hub</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('db')">数据库与Redis</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">核心网关配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>WebSocket 端口 (Core)</label>
                        <input type="text" id="ws_port">
                    </div>
                    <div class="form-group">
                        <label>Web UI 端口</label>
                        <input type="text" id="webui_port">
                    </div>
                    <div class="form-group">
                        <label>JWT 密钥</label>
                        <input type="password" id="jwt_secret">
                    </div>
                    <div class="form-group">
                        <label>默认管理员密码</label>
                        <input type="password" id="default_admin_password">
                    </div>
                    <div class="form-group">
                        <label>统计文件路径</label>
                        <input type="text" id="stats_file">
                    </div>
                </div>
            </div>

            <div style="text-align: center; margin-top: 30px;">
                <button class="btn btn-primary" style="padding: 15px 40px; font-size: 16px;" onclick="saveConfig()">保存配置并重启</button>
            </div>
        </div>

        <div id="db-tab" style="display: none;">
            <div class="card">
                <div class="section-title">持久化数据库 (SQLite/PostgreSQL)</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>数据库类型</label>
                        <select id="db_type">
                            <option value="sqlite">SQLite</option>
                            <option value="postgres">PostgreSQL</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>PG 主机</label>
                        <input type="text" id="pg_host">
                    </div>
                    <div class="form-group">
                        <label>PG 端口</label>
                        <input type="number" id="pg_port">
                    </div>
                    <div class="form-group">
                        <label>PG 用户</label>
                        <input type="text" id="pg_user">
                    </div>
                    <div class="form-group">
                        <label>PG 密码</label>
                        <input type="password" id="pg_password">
                    </div>
                    <div class="form-group">
                        <label>PG 数据库名</label>
                        <input type="text" id="pg_dbname">
                    </div>
                    <div class="form-group">
                        <label>PG SSL 模式</label>
                        <input type="text" id="pg_sslmode">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">Redis 配置 (缓存与队列)</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>Redis 地址</label>
                        <input type="text" id="redis_addr">
                    </div>
                    <div class="form-group">
                        <label>Redis 密码</label>
                        <input type="password" id="redis_pwd">
                    </div>
                </div>
            </div>

            <div style="text-align: center; margin-top: 30px;">
                <button class="btn btn-primary" style="padding: 15px 40px; font-size: 16px;" onclick="saveConfig()">保存配置并重启</button>
            </div>
        </div>

        <div id="logs-tab" style="display: none;">
            <div class="card">
                <div class="section-title">
                    系统日志 (最近 100 行)
                    <button class="btn btn-danger" onclick="clearLogs()">清空显示</button>
                </div>
                <div id="logs" class="logs-container"></div>
            </div>
        </div>
    </div>

    <script>
        let currentTab = 'config';
        function switchTab(tab) {
            document.getElementById(currentTab + '-tab').style.display = 'none';
            document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
            
            document.getElementById(tab + '-tab').style.display = 'block';
            event.currentTarget.classList.add('active');
            currentTab = tab;
        }

        async function loadConfig() {
            const resp = await fetch('/config');
            const config = await resp.json();
            
            document.getElementById('ws_port').value = config.ws_port || ':3001';
            document.getElementById('webui_port').value = config.webui_port || ':5000';
            document.getElementById('jwt_secret').value = config.jwt_secret || '';
            document.getElementById('default_admin_password').value = config.default_admin_password || '';
            document.getElementById('stats_file').value = config.stats_file || 'stats.json';
            
            document.getElementById('db_type').value = config.db_type || 'sqlite';
            document.getElementById('pg_host').value = config.pg_host || 'localhost';
            document.getElementById('pg_port').value = config.pg_port || 5432;
            document.getElementById('pg_user').value = config.pg_user || 'postgres';
            document.getElementById('pg_password').value = config.pg_password || '';
            document.getElementById('pg_dbname').value = config.pg_dbname || 'botnexus';
            document.getElementById('pg_sslmode').value = config.pg_sslmode || 'disable';
            
            document.getElementById('redis_addr').value = config.redis_addr || '';
            document.getElementById('redis_pwd').value = config.redis_pwd || '';
        }

        async function saveConfig() {
            const config = {
                ws_port: document.getElementById('ws_port').value,
                webui_port: document.getElementById('webui_port').value,
                jwt_secret: document.getElementById('jwt_secret').value,
                default_admin_password: document.getElementById('default_admin_password').value,
                stats_file: document.getElementById('stats_file').value,
                
                db_type: document.getElementById('db_type').value,
                pg_host: document.getElementById('pg_host').value,
                pg_port: parseInt(document.getElementById('pg_port').value),
                pg_user: document.getElementById('pg_user').value,
                pg_password: document.getElementById('pg_password').value,
                pg_dbname: document.getElementById('pg_dbname').value,
                pg_sslmode: document.getElementById('pg_sslmode').value,
                
                redis_addr: document.getElementById('redis_addr').value,
                redis_pwd: document.getElementById('redis_pwd').value
            };

            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });

            if (resp.ok) {
                alert('配置已保存，BotNexus 正在重启...');
                setTimeout(() => window.location.reload(), 3000);
            } else {
                const err = await resp.text();
                alert('保存失败: ' + err);
            }
        }

        async function updateLogs() {
            if (currentTab !== 'logs') return;
            try {
                const resp = await fetch('/logs?lines=100');
                const text = await resp.text();
                const logsDiv = document.getElementById('logs');
                logsDiv.innerText = text;
                logsDiv.scrollTop = logsDiv.scrollHeight;
            } catch (e) {}
        }

        function clearLogs() {
            document.getElementById('logs').innerText = '';
        }

        setInterval(updateLogs, 2000);
        loadConfig();
    </script>
</body>
</html>
	`)
}

func handleConfigUpdate(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var newConfig common.AppConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "config_format_error"),
			})
			return
		}

		// 更新配置
		m.Config.WSPort = newConfig.WSPort
		m.Config.WebUIPort = newConfig.WebUIPort
		m.Config.RedisAddr = newConfig.RedisAddr
		m.Config.RedisPwd = newConfig.RedisPwd
		m.Config.JWTSecret = newConfig.JWTSecret
		m.Config.DefaultAdminPassword = newConfig.DefaultAdminPassword
		m.Config.StatsFile = newConfig.StatsFile

		m.Config.DBType = newConfig.DBType
		m.Config.PGHost = newConfig.PGHost
		m.Config.PGPort = newConfig.PGPort
		m.Config.PGUser = newConfig.PGUser
		m.Config.PGPassword = newConfig.PGPassword
		m.Config.PGDBName = newConfig.PGDBName
		m.Config.PGSSLMode = newConfig.PGSSLMode

		if err := m.SaveConfig(); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "config_save_failed"), err),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": common.T(lang, "config_updated"),
		})

		// 异步重启
		go func() {
			time.Sleep(1 * time.Second)
			restartBot()
		}()
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	lines := 100
	fmt.Sscanf(r.URL.Query().Get("lines"), "%d", &lines)
	fmt.Fprint(w, logMgr.GetLogs(lines))
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
