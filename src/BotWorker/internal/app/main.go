package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/mcp"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/bot"
	common_config "BotMatrix/common/config"
	"BotMatrix/common/log"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/redis"
	"botworker/internal/server"
	"botworker/plugins"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	botService    *bot.BaseBot
	currentConfig *config.Config
	configPath    string
	configMutex   sync.RWMutex
	workerServer  *server.CombinedServer
	serverMutex   sync.RWMutex
	cancelFunc    context.CancelFunc
	ctxMutex      sync.Mutex
	redisClient   *redis.Client
	pluginBridge  *PluginBridge
	bridgeMutex   sync.RWMutex
)

const VERSION = "1.0.0"

// Run 启动 BotWorker
func Run() {
	// 初始化基础机器人，默认端口 8082
	botService = bot.NewBaseBot(8082)

	// 设置日志输出到 BaseBot 的 LogManager
	log.SetOutput(botService.LogManager)

	// 初始化翻译器
	utils.InitTranslator("locales", "zh-CN")

	// 加载初始配置
	cfg, path, err := config.LoadFromCLI()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	currentConfig = cfg
	configPath = path

	// 同步配置到 common 包
	common_config.InitConfig(path)

	// 设置标准处理器
	setupHandlers()

	// 注册详细统计信息 API
	botService.Mux.HandleFunc("/api/stats/group", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		groupID := r.URL.Query().Get("group_id")
		period := r.URL.Query().Get("period") // today, yesterday, week, month, last_week, last_month

		if groupID == "" {
			http.Error(w, "group_id is required", http.StatusBadRequest)
			return
		}

		type StatResult struct {
			UserID   string `json:"user_id"`
			Nickname string `json:"nickname"`
			Count    int64  `json:"count"`
		}
		var results []StatResult

		now := time.Now()
		todayStr := now.Format("2006-01-02")

		switch period {
		case "today":
			// 从 Redis 读取今日实时数据
			redisKey := fmt.Sprintf("stats:rank:%s:group:%s", todayStr, groupID)
			zResults, _ := redisClient.ZRevRangeWithScores(ctx, redisKey, 0, 19).Result()
			for _, z := range zResults {
				uid := z.Member.(string)
				results = append(results, StatResult{
					UserID: uid,
					Count:  int64(z.Score),
				})
			}

		case "yesterday":
			// 从 DB 读取昨日固化数据
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			if plugins.GlobalStore != nil {
				stats, err := plugins.GlobalStore.Messages.GetTopStats(groupID, yesterday, 20)
				if err == nil {
					for _, s := range stats {
						results = append(results, StatResult{
							UserID: s.UserID,
							Count:  s.Count,
						})
					}
				}
			}

		case "week", "month", "last_week", "last_month":
			// 聚合查询
			var startDate, endDate time.Time
			includeToday := false

			switch period {
			case "week":
				// 本周（周一到现在）
				offset := int(now.Weekday()) - 1
				if offset < 0 {
					offset = 6
				}
				startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offset)
				endDate = now
				includeToday = true
			case "last_week":
				// 上周
				offset := int(now.Weekday()) - 1
				if offset < 0 {
					offset = 6
				}
				endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offset-1)
				startDate = endDate.AddDate(0, 0, -6)
			case "month":
				// 本月
				startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				endDate = now
				includeToday = true
			case "last_month":
				// 上月
				startDate = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
				endDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
			}

			// 1. 从 DB 查询历史数据
			if plugins.GlobalStore != nil {
				stats, err := plugins.GlobalStore.Messages.GetStatsRange(groupID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), 20)
				if err == nil {
					for _, s := range stats {
						results = append(results, StatResult{
							UserID: s.UserID,
							Count:  s.Count,
						})
					}
				}
			}

			// 2. 如果包含今天，则合并 Redis 数据
			if includeToday {
				redisKey := fmt.Sprintf("stats:rank:%s:group:%s", todayStr, groupID)
				todayData, _ := redisClient.ZRangeWithScores(ctx, redisKey, 0, -1).Result()

				resMap := make(map[string]int64)
				for _, r := range results {
					resMap[r.UserID] = r.Count
				}
				for _, z := range todayData {
					uid := z.Member.(string)
					resMap[uid] += int64(z.Score)
				}

				// 重新排序
				results = nil
				for uid, count := range resMap {
					results = append(results, StatResult{UserID: uid, Count: count})
				}
				sort.Slice(results, func(i, j int) bool {
					return results[i].Count > results[j].Count
				})
				if len(results) > 20 {
					results = results[:20]
				}
			}
		}

		// 补全昵称
		for i := range results {
			if plugins.GlobalStore != nil {
				if member, err := plugins.GlobalStore.Caches.GetMember(groupID, results[i].UserID); err == nil {
					if member.Card != "" {
						results[i].Nickname = member.Card
					} else {
						results[i].Nickname = member.Nickname
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// 注册统计信息 API
	botService.Mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		res := make(map[string]any)

		// 1. 获取总消息数
		total, _ := redisClient.Get(ctx, "stats:total_messages").Int64()
		res["total_messages"] = total

		// 2. 获取用户排名 (Top 10)
		userRank, _ := redisClient.HGetAll(ctx, "stats:users:rank").Result()
		res["user_rank"] = userRank

		// 3. 获取群组排名 (Top 10)
		groupRank, _ := redisClient.HGetAll(ctx, "stats:groups:rank").Result()
		res["group_rank"] = groupRank

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	// 启动 HTTP 服务器
	go botService.StartHTTPServer()

	// 启动机器人
	restartBot()

	// 等待退出信号并处理清理
	botService.WaitExitSignal()
	stopWorker()
}

func setupHandlers() {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 使用 BaseBot 的标准处理器
	botService.SetupStandardHandlers("BotWorker", currentConfig, restartBot, []bot.ConfigSection{
		{
			Title: "基础设置",
			Fields: []bot.ConfigField{
				{Label: "Worker ID", ID: "worker_id", Type: "text", Value: currentConfig.WorkerID},
				{Label: "启用技能系统", ID: "enable_skill", Type: "checkbox", Value: currentConfig.EnableSkill},
			},
		},
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Token", ID: "bot_token", Type: "password", Value: currentConfig.BotToken},
				{Label: "Bot Nexus 地址", ID: "nexus_addr", Type: "text", Value: currentConfig.NexusAddr},
				{Label: "Web UI 端口", ID: "log_port", Type: "number", Value: currentConfig.LogPort},
			},
		},
		{
			Title: "HTTP 服务配置",
			Fields: []bot.ConfigField{
				{Label: "监听地址", ID: "http.addr", Type: "text", Value: currentConfig.HTTP.Addr},
			},
		},
		{
			Title: "WebSocket 服务配置",
			Fields: []bot.ConfigField{
				{Label: "监听地址", ID: "websocket.addr", Type: "text", Value: currentConfig.WebSocket.Addr},
			},
		},
		{
			Title: "数据库配置",
			Fields: []bot.ConfigField{
				{Label: "Host", ID: "database.host", Type: "text", Value: currentConfig.Database.Host},
				{Label: "Port", ID: "database.port", Type: "number", Value: currentConfig.Database.Port},
				{Label: "User", ID: "database.user", Type: "text", Value: currentConfig.Database.User},
				{Label: "Password", ID: "database.password", Type: "password", Value: currentConfig.Database.Password},
				{Label: "DB Name", ID: "database.dbname", Type: "text", Value: currentConfig.Database.DBName},
			},
		},
		{
			Title: "Redis 配置",
			Fields: []bot.ConfigField{
				{Label: "Host", ID: "redis.host", Type: "text", Value: currentConfig.Redis.Host},
				{Label: "Port", ID: "redis.port", Type: "number", Value: currentConfig.Redis.Port},
				{Label: "Password", ID: "redis.password", Type: "password", Value: currentConfig.Redis.Password},
				{Label: "DB", ID: "redis.db", Type: "number", Value: currentConfig.Redis.DB},
			},
		},
		{
			Title: "AI 配置",
			Fields: []bot.ConfigField{
				{Label: "API Key", ID: "ai.api_key", Type: "password", Value: currentConfig.AI.APIKey},
				{Label: "Endpoint", ID: "ai.endpoint", Type: "text", Value: currentConfig.AI.Endpoint},
				{Label: "Model", ID: "ai.model", Type: "text", Value: currentConfig.AI.Model},
			},
		},
	})
}

func stopWorker() {
	log.Info("正在停止 BotWorker...")
	ctxMutex.Lock()
	if cancelFunc != nil {
		cancelFunc()
	}
	ctxMutex.Unlock()

	serverMutex.Lock()
	if workerServer != nil {
		workerServer.Stop()
	}
	serverMutex.Unlock()
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
			log.Error("机器人运行错误", zap.Error(err))
		}
	}()

	// 连接到 Nexus
	configMutex.RLock()
	nexusAddr := currentConfig.NexusAddr
	workerID := currentConfig.WorkerID
	configMutex.RUnlock()

	if nexusAddr != "" {
		botService.StartNexusConnection(ctx, nexusAddr, "Worker", workerID, handleNexusCommand)
	}
}

func startBot(ctx context.Context) error {
	configMutex.RLock()
	cfg := currentConfig
	configMutex.RUnlock()

	log.Info(utils.T("", "server_starting", VERSION), zap.String("component", "BotWorker"))

	// 测试数据库连接
	database, err := db.NewDBConnection(&cfg.Database)
	if err != nil {
		log.Warn("无法连接到数据库", zap.Error(err))
	} else {
		log.Info("成功连接到数据库")
		plugins.SetGlobalDB(database)
		/* 按照要求，不再通过代码修改数据库结构
		if err := db.InitDatabase(database); err != nil {
			log.Warn("初始化数据库表失败", zap.Error(err))
		}
		*/
	}

	// 初始化 GORM 连接
	gormDB, err := db.NewGORMConnection(&cfg.Database)
	if err != nil {
		log.Warn("无法连接到 GORM 数据库", zap.Error(err))
	} else {
		log.Info("成功连接到 GORM 数据库")
		plugins.SetGlobalGORMDB(gormDB)
	}

	// 测试Redis连接
	var err2 error
	redisClient, err2 = redis.NewClient(&cfg.Redis)
	if err2 != nil {
		log.Warn("无法连接到Redis服务器", zap.Error(err2))
	} else {
		log.Info("成功连接到Redis服务器")
		plugins.SetGlobalRedis(redisClient)
	}

	// 创建组合服务器
	serverMutex.Lock()
	workerServer = server.NewCombinedServer(botService, cfg, redisClient)
	serverMutex.Unlock()

	// 初始化 AI 服务
	var aiService types.AIService
	if plugins.GlobalGORMDB != nil {
		// 1. 初始化 RAG 知识库
		var kb types.KnowledgeBase
		if cfg.Database.Host != "" {
			// Worker 端通常使用简单的 OpenAIExtractor 作为 EmbeddingService
			es := ai.NewOpenAIAdapter(cfg.AI.Endpoint, cfg.AI.APIKey)
			kb = rag.NewPostgresKnowledgeBase(plugins.GlobalGORMDB, es, nil, 0)
		}

		// 2. 初始化 AI Provider
		provider := NewWorkerAIProvider(cfg, kb)

		// 3. 初始化基础 WorkerAIService 作为 Manager 的一部分
		baseAISvc := NewWorkerAIService(cfg)

		// 4. 初始化 MCP Manager (WorkerAIService 实现了 Manager 接口)
		mcpManager := mcp.NewMCPManager(baseAISvc)
		if kb != nil {
			mcpManager.SetKnowledgeBase(kb)
		}

		// 5. 初始化核心 AI 服务
		aiService = ai.NewAIService(plugins.GlobalGORMDB, provider, mcpManager)
		log.Info("核心 AI 服务初始化成功")
	} else {
		// 回退到基础 AI 服务
		aiService = NewWorkerAIService(cfg)
		log.Info("已回退到基础 WorkerAIService")
	}

	// 获取插件管理器
	pluginManager := workerServer.GetPluginManager()

	// 将 AI 服务设置到 CombinedServer
	workerServer.SetAIService(aiService)

	// 初始化插件桥接器 (负责扫描和加载外部进程插件)
	bridge := NewPluginBridge(workerServer, aiService)
	bridgeMutex.Lock()
	pluginBridge = bridge
	bridgeMutex.Unlock()

	if err := bridge.LoadInternalPlugins(); err != nil {
		log.Error("加载内部插件失败", zap.Error(err))
	}
	if err := bridge.LoadExternalPlugins(); err != nil {
		log.Error("加载外部插件失败", zap.Error(err))
	}

	// 打印已加载的外部插件
	log.Info("已加载的外部插件:")
	externalNames := make(map[string]bool)
	for id, versions := range pluginManager.GetPlugins() {
		for _, p := range versions {
			log.Info("外部插件",
				zap.String("id", id),
				zap.String("version", p.Config.Version),
				zap.String("state", p.State))
			externalNames[p.Config.Name] = true
		}
	}

	// 打印已加载的内部插件
	log.Info("已加载的内部插件:")
	for _, p := range pluginManager.GetInternalPlugins() {
		if externalNames[p.Name()] {
			continue
		}
		log.Info("内部插件",
			zap.String("name", p.Name()),
			zap.String("version", p.Version()),
			zap.String("description", p.Description()))
	}
	log.Info("管理后台已启动", zap.String("url", fmt.Sprintf("http://localhost:%d/config-ui", cfg.LogPort)))

	// 启动服务器
	log.Info("启动OneBot协议机器人服务器...")
	go func() {
		if err := workerServer.Run(); err != nil {
			log.Error("服务器启动失败", zap.Error(err))
		}
	}()

	// 延迟上报状态，确保连接已建立
	go func() {
		time.Sleep(2 * time.Second)
		reportWorkerStatus()
	}()

	// 定期上报状态
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reportWorkerStatus()
			}
		}
	}()

	// 启动 RAG 自动同步任务 (系统文档同步)
	if kb, ok := aiService.GetMCPManager().GetKnowledgeBase().(*rag.PostgresKnowledgeBase); ok {
		// 自动查找 docs 目录，尝试多种可能的路径
		baseDirs := []string{".", "..", "../.."}
		var docsDir string
		for _, d := range baseDirs {
			testDir := filepath.Join(d, "docs", "zh-CN")
			if _, err := os.Stat(testDir); err == nil {
				docsDir = d
				break
			}
		}

		if docsDir != "" {
			indexer := rag.NewIndexer(kb, aiService, 0)
			go rag.StartAutoSync(ctx, indexer, docsDir, 1*time.Hour)
			log.Info("RAG 系统文档自动同步任务已启动", zap.String("base_dir", docsDir))
		} else {
			log.Warn("未找到 docs 目录，跳过 RAG 系统文档自动同步")
		}
	}

	// 监听 Redis 命令
	if redisClient != nil {
		go func() {
			ctx := context.Background()
			channel := fmt.Sprintf("botmatrix:worker:%s:commands", currentConfig.WorkerID)
			pubsub := redisClient.Subscribe(ctx, channel)
			defer pubsub.Close()

			log.Info("开始监听 Redis 命令通道", zap.String("channel", channel))

			ch := pubsub.Channel()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-ch:
					if msg == nil {
						return
					}
					log.Info("从 Redis 收到命令", zap.String("payload", msg.Payload))
					handleNexusCommand([]byte(msg.Payload))
				}
			}
		}()
	}

	<-ctx.Done()
	log.Info("停止 BotWorker...")
	serverMutex.Lock()
	if workerServer != nil {
		workerServer.Stop()
	}
	serverMutex.Unlock()
	return ctx.Err()
}

func loadAllPlugins(pluginManager *core.PluginManager, cfg *config.Config, database *sql.DB, redisClient *redis.Client) {
	// 加载插件

}

func handleNexusCommand(data []byte) {
	var req struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		log.Error("解析 Nexus 命令失败", zap.Error(err))
		return
	}

	log.Info("收到 Nexus 命令", zap.String("action", req.Action))

	// 将命令转发给 CombinedServer 处理
	serverMutex.RLock()
	s := workerServer
	serverMutex.RUnlock()
	if s == nil {
		return
	}

	if req.Action == "plugin_install" {
		filename := utils.ToString(req.Params["filename"])
		contentBase64 := utils.ToString(req.Params["content"])

		// 解码 Base64
		content, err := base64.StdEncoding.DecodeString(contentBase64)
		if err != nil {
			log.Error("解码插件内容失败", zap.Error(err))
			return
		}

		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "worker-plugin-*.bmpk")
		if err != nil {
			log.Error("创建临时插件文件失败", zap.Error(err))
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := tmpFile.Write(content); err != nil {
			log.Error("写入临时插件文件失败", zap.Error(err))
			return
		}

		pm := s.GetPluginManager()
		if pm != nil {
			configMutex.RLock()
			workerPluginsDir := currentConfig.Plugin.Dir
			configMutex.RUnlock()
			if workerPluginsDir == "" {
				workerPluginsDir = "plugins/worker"
			}

			// 确保目录存在
			if _, err := os.Stat(workerPluginsDir); os.IsNotExist(err) {
				os.MkdirAll(workerPluginsDir, 0755)
			}

			if err := pm.InstallPlugin(tmpFile.Name(), workerPluginsDir); err != nil {
				log.Error("安装插件失败", zap.String("filename", filename), zap.Error(err))
			} else {
				log.Info("成功安装插件", zap.String("filename", filename))
				// 重新加载插件列表
				bridgeMutex.RLock()
				if pluginBridge != nil {
					pluginBridge.Reload()
				}
				bridgeMutex.RUnlock()
			}
		}
	} else if req.Action == "invoke_skill" {
		skillName := utils.ToString(req.Params["skill_name"])
		payload, _ := req.Params["payload"].(map[string]any)

		// 转换 payload 为 map[string]string 以适配 InvokeSkill
		params := make(map[string]string)
		for k, v := range payload {
			params[k] = fmt.Sprintf("%v", v)
		}

		serverMutex.RLock()
		s := workerServer
		serverMutex.RUnlock()

		if s != nil {
			go func() {
				result, err := s.InvokeSkill(skillName, params)
				// 如果有 echo，则将结果发送回 Nexus (通过 Redis 或 WebSocket)
				if req.Echo != "" {
					resp := map[string]any{
						"type":   "skill_response",
						"echo":   req.Echo,
						"result": result,
					}
					if err != nil {
						resp["error"] = err.Error()
					}
					// 这里简单通过 botService 发送
					botService.SendToNexus(resp)
				}
			}()
		}
		return
	}

	if req.Action == "plugin_action" {
		pluginID := utils.ToString(req.Params["id"])
		action := utils.ToString(req.Params["action"])
		pm := s.GetPluginManager()
		if pm != nil {
			var err error
			switch action {
			case "start":
				err = pm.StartPlugin(pluginID, "")
			case "stop":
				err = pm.StopPlugin(pluginID, "")
			case "restart":
				err = pm.RestartPlugin(pluginID, "")
			case "reload":
				bridgeMutex.RLock()
				if pluginBridge != nil {
					err = pluginBridge.Reload()
				} else {
					configMutex.RLock()
					workerPluginsDir := currentConfig.Plugin.Dir
					configMutex.RUnlock()
					if workerPluginsDir == "" {
						workerPluginsDir = "plugins/worker"
					}
					err = pm.LoadPlugins(workerPluginsDir)
				}
				bridgeMutex.RUnlock()
			}
			if err != nil {
				log.Error("执行插件操作失败", zap.String("id", pluginID), zap.String("action", action), zap.Error(err))
			} else {
				log.Info("成功执行插件操作", zap.String("id", pluginID), zap.String("action", action))
				reportWorkerStatus()
			}
		}
		return
	}

	if req.Action == "plugin_delete" {
		pluginID := utils.ToString(req.Params["id"])
		version := utils.ToString(req.Params["version"])
		pm := s.GetPluginManager()
		if pm != nil {
			plugin := pm.GetPlugin(pluginID, version)
			if plugin != nil {
				// 停止插件
				if plugin.State == "running" {
					pm.StopPlugin(pluginID, version)
				}
				// 删除目录
				if err := os.RemoveAll(plugin.Dir); err != nil {
					log.Error("删除插件文件失败", zap.String("id", pluginID), zap.Error(err))
				} else {
					log.Info("成功删除插件", zap.String("id", pluginID))
					// 从内存中移除
					pm.RemovePlugin(pluginID, version)
					// 重新加载插件列表
					bridgeMutex.RLock()
					if pluginBridge != nil {
						pluginBridge.Reload()
					} else {
						configMutex.RLock()
						workerPluginsDir := currentConfig.Plugin.Dir
						configMutex.RUnlock()
						if workerPluginsDir == "" {
							workerPluginsDir = "plugins/worker"
						}
						pm.LoadPlugins(workerPluginsDir)
					}
					bridgeMutex.RUnlock()
					reportWorkerStatus()
				}
			}
		}
		return
	}
}

// reportWorkerStatus 收集并上报 Worker 状态（如插件列表、能力列表）到 Nexus
func reportWorkerStatus() {
	serverMutex.RLock()
	s := workerServer
	serverMutex.RUnlock()
	if s == nil {
		return
	}

	pm := s.GetPluginManager()
	if pm == nil {
		return
	}

	var pluginsInfo []map[string]any

	// 1. 外部插件 (优先上报，因为信息更全)
	externalIDs := make(map[string]bool)
	for id, versions := range pm.GetPlugins() {
		for _, p := range versions {
			pluginEntry := map[string]any{
				"id":          id,
				"name":        p.Config.Name,
				"version":     p.Config.Version,
				"description": p.Config.Description,
				"author":      p.Config.Author,
				"state":       p.State,
				"is_internal": false,
			}

			// 收集插件能力
			if sc, ok := p.(core.SkillCapable); ok {
				var skills []types.WorkerCapability
				for _, skill := range sc.GetSkills() {
					skills = append(skills, types.WorkerCapability{
						Name:        skill.Name,
						Description: skill.Description,
						Usage:       skill.Usage,
						Params:      skill.Params,
						Regex:       skill.Regex,
					})
				}
				pluginEntry["skills"] = skills
			}

			pluginsInfo = append(pluginsInfo, pluginEntry)
			externalIDs[p.Config.Name] = true
		}
	}

	// 2. 内部插件
	for _, p := range pm.GetInternalPlugins() {
		// 如果该插件名已作为外部插件上报过，则在列表中跳过它以避免重复显示
		if externalIDs[p.Name()] {
			continue
		}

		pluginEntry := map[string]any{
			"id":          p.Name(),
			"name":        p.Name(),
			"version":     p.Version(),
			"description": p.Description(),
			"author":      "system",
			"state":       "running",
			"is_internal": true,
		}

		// 收集插件能力
		if sc, ok := p.(core.SkillCapable); ok {
			var skills []types.WorkerCapability
			for _, skill := range sc.GetSkills() {
				skills = append(skills, types.WorkerCapability{
					Name:        skill.Name,
					Description: skill.Description,
					Usage:       skill.Usage,
					Params:      skill.Params,
					Regex:       skill.Regex,
				})
			}
			pluginEntry["skills"] = skills
		}

		pluginsInfo = append(pluginsInfo, pluginEntry)
	}

	msg := map[string]any{
		"type": "update_metadata",
		"metadata": map[string]any{
			"plugins":   pluginsInfo,
			"http_addr": currentConfig.HTTP.Addr,
		},
	}

	log.Debug("上报 Worker 状态到 Nexus",
		zap.Int("plugin_count", len(pluginsInfo)))

	botService.SendToNexus(msg)

	// 更新活跃时间 (备份)
	if redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		redisClient.Set(ctx, fmt.Sprintf("botmatrix:worker:%s:last_seen", currentConfig.WorkerID), time.Now().Unix(), 24*time.Hour)
	}
}
