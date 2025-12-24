package main

import (
	"BotMatrix/common"
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"botworker/internal/server"
	"botworker/plugins"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// LogManager 处理日志滚动和获取
type LogManager struct {
	maxLines int
	logs     []string
	mutex    sync.RWMutex
}

func NewLogManager(maxLines int) *LogManager {
	return &LogManager{
		maxLines: maxLines,
		logs:     make([]string, 0, maxLines),
	}
}

func (m *LogManager) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	line := string(p)
	m.logs = append(m.logs, line)
	if len(m.logs) > m.maxLines {
		m.logs = m.logs[len(m.logs)-m.maxLines:]
	}
	return os.Stderr.Write(p)
}

func (m *LogManager) GetLogs(lines int) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if lines > len(m.logs) {
		lines = len(m.logs)
	}
	return m.logs[len(m.logs)-lines:]
}

var (
	logManager    *LogManager
	currentConfig *config.Config
	configPath    string
	configMutex   sync.RWMutex
	workerServer  *server.CombinedServer
	serverMutex   sync.RWMutex
	cancelFunc    context.CancelFunc
	ctxMutex      sync.Mutex
)

func main() {
	// 初始化日志管理器
	logManager = NewLogManager(1000)
	log.SetOutput(logManager)

	// 初始化翻译器
	common.InitTranslator("locales", "zh-CN")

	// 加载初始配置
	cfg, path, err := config.LoadFromCLI()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	currentConfig = cfg
	configPath = path

	// 启动 Web UI
	go startHTTPServer()

	// 启动机器人
	restartBot()

	// 阻塞主协程
	select {}
}

func restartBot() {
	ctxMutex.Lock()
	if cancelFunc != nil {
		cancelFunc()
	}
	var ctx context.Context
	ctx, cancelFunc = context.WithCancel(context.Background())
	ctxMutex.Unlock()

	go func() {
		err := startBot(ctx)
		if err != nil && err != context.Canceled {
			log.Printf("机器人运行错误: %v", err)
		}
	}()
}

func startBot(ctx context.Context) error {
	configMutex.RLock()
	cfg := currentConfig
	configMutex.RUnlock()

	log.Println(common.T("", "server_starting"), "BotWorker")

	// 测试数据库连接
	database, err := db.NewDBConnection(&cfg.Database)
	if err != nil {
		log.Printf("警告: 无法连接到数据库: %v", err)
	} else {
		log.Println("成功连接到数据库")
		plugins.SetGlobalDB(database)
		if err := db.InitDatabase(database); err != nil {
			log.Printf("警告: 初始化数据库表失败: %v", err)
		}
	}

	// 测试Redis连接
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Printf("警告: 无法连接到Redis服务器: %v", err)
	} else {
		log.Println("成功连接到Redis服务器")
		plugins.SetGlobalRedis(redisClient)
	}

	// 创建组合服务器
	serverMutex.Lock()
	workerServer = server.NewCombinedServer(cfg, redisClient)
	serverMutex.Unlock()

	// 获取插件管理器
	pluginManager := workerServer.GetPluginManager()

	// 加载所有插件
	loadAllPlugins(pluginManager, cfg, database, redisClient)

	// 打印已加载的插件
	log.Println("已加载的插件:")
	for _, p := range pluginManager.GetPlugins() {
		log.Printf("- %s v%s: %s", p.Name(), p.Version(), p.Description())
	}

	// 启动服务器
	log.Println("启动OneBot协议机器人服务器...")
	go func() {
		if err := workerServer.Run(); err != nil {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("停止 BotWorker...")
	serverMutex.Lock()
	if workerServer != nil {
		workerServer.Stop()
	}
	serverMutex.Unlock()
	return ctx.Err()
}

func loadAllPlugins(pluginManager *plugin.Manager, cfg *config.Config, database *db.DB, redisClient *redis.Client) {
	// 加载示例插件
	pluginManager.LoadPlugin(&plugins.EchoPlugin{})
	pluginManager.LoadPlugin(&plugins.WelcomePlugin{})
	pluginManager.LoadPlugin(plugins.NewGroupManagerPlugin(database, redisClient))
	pluginManager.LoadPlugin(plugins.NewWeatherPlugin(&cfg.Weather))

	pointsPlugin := plugins.NewPointsPlugin(database)
	pluginManager.LoadPlugin(pointsPlugin)

	pluginManager.LoadPlugin(plugins.NewSignInPlugin(pointsPlugin))
	pluginManager.LoadPlugin(plugins.NewAuctionPlugin(database, pointsPlugin))
	pluginManager.LoadPlugin(plugins.NewMedalPlugin())
	pluginManager.LoadPlugin(plugins.NewGamesPlugin())
	pluginManager.LoadPlugin(plugins.NewLotteryPlugin())
	pluginManager.LoadPlugin(plugins.NewMenuPlugin())
	pluginManager.LoadPlugin(plugins.NewTranslatePlugin(&cfg.Translate))
	pluginManager.LoadPlugin(plugins.NewMusicPlugin())
	pluginManager.LoadPlugin(plugins.NewPetPlugin(database, pointsPlugin))
	pluginManager.LoadPlugin(plugins.NewTimePlugin())
	pluginManager.LoadPlugin(plugins.NewQAPlugin())
	pluginManager.LoadPlugin(plugins.NewGiftPlugin(database))
	pluginManager.LoadPlugin(plugins.NewMarriagePlugin())
	pluginManager.LoadPlugin(plugins.NewBabyPlugin())
	pluginManager.LoadPlugin(plugins.NewBadgePlugin())
	pluginManager.LoadPlugin(plugins.NewSmallGamesPlugin())
	pluginManager.LoadPlugin(plugins.NewKnowledgeBasePlugin(database, cfg.AI.OfficialGroupID))
	pluginManager.LoadPlugin(plugins.NewDialogDemoPlugin())
	pluginManager.LoadPlugin(plugins.NewTestServerPlugin())
	pluginManager.LoadPlugin(plugins.NewRobberyPlugin(database))
	pluginManager.LoadPlugin(plugins.NewFishingPlugin(database))
	pluginManager.LoadPlugin(plugins.NewCultivationPlugin(database))
	pluginManager.LoadPlugin(plugins.NewFarmPlugin(database))
	pluginManager.LoadPlugin(plugins.NewTarotPlugin())
	pluginManager.LoadPlugin(plugins.NewWordGuessPlugin())
	pluginManager.LoadPlugin(plugins.NewIdiomGuessPlugin())
	pluginManager.LoadPlugin(plugins.NewPluginManagerPlugin(database))
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config-ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		lines := 100
		if l := r.URL.Query().Get("lines"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				lines = v
			}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		logs := logManager.GetLogs(lines)
		for _, line := range logs {
			fmt.Fprint(w, line)
		}
	})

	mux.HandleFunc("/config", handleConfig)
	mux.HandleFunc("/config-ui", handleConfigUI)

	configMutex.RLock()
	port := currentConfig.LogPort
	configMutex.RUnlock()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP Server at http://localhost%s (UI: /config-ui, Logs: /logs)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Failed to start HTTP Server: %v", err)
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		configMutex.RLock()
		json.NewEncoder(w).Encode(currentConfig)
		configMutex.RUnlock()
		return
	}

	if r.Method == http.MethodPost {
		var newCfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		configMutex.Lock()
		currentConfig = &newCfg
		configMutex.Unlock()

		// 保存到文件
		if configPath != "" {
			data, _ := json.MarshalIndent(newCfg, "", "  ")
			os.WriteFile(configPath, data, 0644)
		}

		restartBot()
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	cfg := currentConfig
	configMutex.RUnlock()

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>BotWorker Configuration</title>
    <style>
        body { font-family: sans-serif; margin: 20px; background: #f0f2f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #1a73e8; }
        .section { margin-bottom: 20px; padding: 15px; border: 1px solid #e0e0e0; border-radius: 4px; }
        .section-title { font-weight: bold; margin-bottom: 10px; color: #5f6368; border-bottom: 1px solid #eee; padding-bottom: 5px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; }
        input[type="text"], input[type="number"], input[type="password"], select {
            width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;
        }
        button {
            background: #1a73e8; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; font-size: 16px;
        }
        button:hover { background: #1557b0; }
        .logs { background: #000; color: #0f0; padding: 15px; border-radius: 4px; height: 300px; overflow-y: auto; font-family: monospace; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>BotWorker 控制面板</h1>
        
        <form id="configForm">
            <div class="section">
                <div class="section-title">基本设置</div>
                <div class="form-group">
                    <label>Worker ID</label>
                    <input type="text" name="worker_id" value="{{.WorkerID}}">
                </div>
                <div class="form-group">
                    <label>Web UI 端口 (需要手动重启生效)</label>
                    <input type="number" name="log_port" value="{{.LogPort}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">HTTP 服务器配置</div>
                <div class="form-group">
                    <label>监听地址</label>
                    <input type="text" name="http.addr" value="{{.HTTP.Addr}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">WebSocket 服务器配置</div>
                <div class="form-group">
                    <label>监听地址</label>
                    <input type="text" name="websocket.addr" value="{{.WebSocket.Addr}}">
                </div>
                <div class="form-group">
                    <label>检查来源</label>
                    <select name="websocket.check_origin">
                        <option value="true" {{if .WebSocket.CheckOrigin}}selected{{end}}>是</option>
                        <option value="false" {{if not .WebSocket.CheckOrigin}}selected{{end}}>否</option>
                    </select>
                </div>
            </div>

            <div class="section">
                <div class="section-title">数据库配置 (PostgreSQL)</div>
                <div class="form-group">
                    <label>Host</label>
                    <input type="text" name="database.host" value="{{.Database.Host}}">
                </div>
                <div class="form-group">
                    <label>Port</label>
                    <input type="number" name="database.port" value="{{.Database.Port}}">
                </div>
                <div class="form-group">
                    <label>User</label>
                    <input type="text" name="database.user" value="{{.Database.User}}">
                </div>
                <div class="form-group">
                    <label>Password</label>
                    <input type="password" name="database.password" value="{{.Database.Password}}">
                </div>
                <div class="form-group">
                    <label>Database Name</label>
                    <input type="text" name="database.dbname" value="{{.Database.DBName}}">
                </div>
                <div class="form-group">
                    <label>SSL Mode</label>
                    <input type="text" name="database.sslmode" value="{{.Database.SSLMode}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">Redis 配置</div>
                <div class="form-group">
                    <label>Host</label>
                    <input type="text" name="redis.host" value="{{.Redis.Host}}">
                </div>
                <div class="form-group">
                    <label>Port</label>
                    <input type="number" name="redis.port" value="{{.Redis.Port}}">
                </div>
                <div class="form-group">
                    <label>Password</label>
                    <input type="password" name="redis.password" value="{{.Redis.Password}}">
                </div>
                <div class="form-group">
                    <label>DB Index</label>
                    <input type="number" name="redis.db" value="{{.Redis.DB}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">天气 API 配置</div>
                <div class="form-group">
                    <label>API Key</label>
                    <input type="password" name="weather.api_key" value="{{.Weather.APIKey}}">
                </div>
                <div class="form-group">
                    <label>Endpoint</label>
                    <input type="text" name="weather.endpoint" value="{{.Weather.Endpoint}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">翻译 API 配置</div>
                <div class="form-group">
                    <label>API Key</label>
                    <input type="password" name="translate.api_key" value="{{.Translate.APIKey}}">
                </div>
                <div class="form-group">
                    <label>Endpoint</label>
                    <input type="text" name="translate.endpoint" value="{{.Translate.Endpoint}}">
                </div>
                <div class="form-group">
                    <label>Region</label>
                    <input type="text" name="translate.region" value="{{.Translate.Region}}">
                </div>
            </div>

            <div class="section">
                <div class="section-title">AI 配置</div>
                <div class="form-group">
                    <label>API Key</label>
                    <input type="password" name="ai.api_key" value="{{.AI.APIKey}}">
                </div>
                <div class="form-group">
                    <label>Endpoint</label>
                    <input type="text" name="ai.endpoint" value="{{.AI.Endpoint}}">
                </div>
                <div class="form-group">
                    <label>Model</label>
                    <input type="text" name="ai.model" value="{{.AI.Model}}">
                </div>
                <div class="form-group">
                    <label>Official Group ID</label>
                    <input type="text" name="ai.official_group_id" value="{{.AI.OfficialGroupID}}">
                </div>
            </div>

            <button type="button" onclick="saveConfig()">保存并重启机器人</button>
        </form>

        <h2>实时日志</h2>
        <div class="logs" id="logs"></div>
    </div>

    <script>
        function saveConfig() {
            const form = document.getElementById('configForm');
            const formData = new FormData(form);
            const config = {
                worker_id: formData.get('worker_id'),
                log_port: parseInt(formData.get('log_port')),
                http: {
                    addr: formData.get('http.addr'),
                    read_timeout: 30000000000,
                    write_timeout: 30000000000
                },
                websocket: {
                    addr: formData.get('websocket.addr'),
                    check_origin: formData.get('websocket.check_origin') === 'true',
                    read_timeout: 60000000000,
                    write_timeout: 10000000000,
                    pong_timeout: 60000000000
                },
                database: {
                    host: formData.get('database.host'),
                    port: parseInt(formData.get('database.port')),
                    user: formData.get('database.user'),
                    password: formData.get('database.password'),
                    dbname: formData.get('database.dbname'),
                    sslmode: formData.get('database.sslmode')
                },
                redis: {
                    host: formData.get('redis.host'),
                    port: parseInt(formData.get('redis.port')),
                    password: formData.get('redis.password'),
                    db: parseInt(formData.get('redis.db'))
                },
                weather: {
                    api_key: formData.get('weather.api_key'),
                    endpoint: formData.get('weather.endpoint'),
                    timeout: 10000000000
                },
                translate: {
                    api_key: formData.get('translate.api_key'),
                    endpoint: formData.get('translate.endpoint'),
                    region: formData.get('translate.region'),
                    timeout: 10000000000
                },
                ai: {
                    api_key: formData.get('ai.api_key'),
                    endpoint: formData.get('ai.endpoint'),
                    model: formData.get('ai.model'),
                    official_group_id: formData.get('ai.official_group_id'),
                    timeout: 15000000000
                },
                log: { level: "info", file: "" },
                plugin: { dir: "plugins", enabled: [] }
            };

            fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            }).then(response => {
                if (response.ok) alert('配置已保存，机器人正在重启...');
                else alert('保存失败');
            });
        }

        function updateLogs() {
            fetch('/logs?lines=50')
                .then(response => response.text())
                .then(text => {
                    const logsDiv = document.getElementById('logs');
                    logsDiv.textContent = text;
                    logsDiv.scrollTop = logsDiv.scrollHeight;
                });
        }

        setInterval(updateLogs, 2000);
        updateLogs();
    </script>
</body>
</html>
`
	t := template.Must(template.Must(template.New("config").Parse(tmpl)).Clone())
	t.Execute(w, cfg)
}
