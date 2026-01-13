package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/log"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/session"
	"botworker/internal/onebot"
	"botworker/internal/server"
	"botworker/plugins"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// PluginBridge 桥接我们的插件系统和BotWorker的插件系统
type PluginBridge struct {
	pluginManager    *core.PluginManager
	server           *server.CombinedServer
	aiService        ai.AIService
	pendingResponses sync.Map // map[string]chan string
	watcher          *fsnotify.Watcher
}

func NewPluginBridge(server *server.CombinedServer, aiService ai.AIService) *PluginBridge {
	pm := server.GetPluginManager()
	// 配置插件路径：优先使用配置中的路径，默认为 plugins/worker
	workerPluginsDir := "plugins/worker"
	if server.GetConfig() != nil && server.GetConfig().Plugin.Dir != "" {
		workerPluginsDir = server.GetConfig().Plugin.Dir
	}
	pm.SetPluginPath(workerPluginsDir)

	bridge := &PluginBridge{
		pluginManager: pm,
		server:        server,
		aiService:     aiService,
	}

	// 初始化文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("[PluginBridge] 无法创建文件监听器: %v", err)
	} else {
		bridge.watcher = watcher
		go bridge.watchPlugins()
	}

	return bridge
}

func (pb *PluginBridge) watchPlugins() {
	if pb.watcher == nil {
		return
	}
	defer pb.watcher.Close()

	// 监听所有配置的插件目录
	dirsToWatch := []string{pb.pluginManager.GetPluginPath()}
	if pb.server.GetConfig() != nil {
		dirsToWatch = append(dirsToWatch, pb.server.GetConfig().Plugin.DevDirs...)
	}

	for _, dir := range dirsToWatch {
		if dir == "" {
			continue
		}
		// 确保目录存在
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}

		// 递归监听目录
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() {
				// 跳过 .runtime 目录，不将其加入监听列表
				if info.Name() == ".runtime" {
					return filepath.SkipDir
				}
				pb.watcher.Add(path)
			}
			return nil
		})
		log.Printf("[PluginBridge] 正在监听插件目录: %s", dir)
	}

	// 防抖处理，避免频繁触发重载
	var timer *time.Timer
	for {
		select {
		case event, ok := <-pb.watcher.Events:
			if !ok {
				return
			}
			// 监听写入、重命名或删除事件
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// 忽略 .runtime 目录下的变化，避免无限循环重载
				if strings.Contains(event.Name, ".runtime") {
					continue
				}

				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(3*time.Second, func() {
					log.Printf("[PluginBridge] 检测到插件目录变化 (%s), 正在尝试热重载...", event.Name)
					pb.Reload()
				})
			}
		case err, ok := <-pb.watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("[PluginBridge] 监听器错误: %v", err)
		}
	}
}

func (pb *PluginBridge) Reload() error {
	// 记录当前已加载的插件 ID 和版本，用于对比
	existingPlugins := make(map[string]*core.Plugin)
	for id, versions := range pb.pluginManager.GetPlugins() {
		for _, v := range versions {
			existingPlugins[id+":"+v.Config.Version] = v
		}
	}

	// 扫描所有配置的插件目录
	dirsToScan := []string{pb.pluginManager.GetPluginPath()}
	if pb.server.GetConfig() != nil {
		dirsToScan = append(dirsToScan, pb.server.GetConfig().Plugin.DevDirs...)
	}

	for _, dir := range dirsToScan {
		if dir == "" {
			continue
		}
		if err := pb.pluginManager.ScanPlugins(dir); err != nil {
			log.Errorf("[PluginBridge] 扫描插件目录失败 %s: %v", dir, err)
		}
	}

	// 处理插件变化
	for id, versions := range pb.pluginManager.GetPlugins() {
		if id == "" {
			continue
		}
		for _, v := range versions {
			key := id + ":" + v.Config.Version
			_, exists := existingPlugins[key]

			if !exists {
				log.Printf("[PluginBridge] 发现新插件: %s (v%s), 正在启动...", id, v.Config.Version)
				pb.startAndRegisterPlugin(v)
			} else {
				// 如果插件已存在，但我们收到了热重载信号（通过触发 Reload）
				// 对于开发模式下的插件，我们强制重启以应用可能的二进制更新
				// 注意：由于我们使用了 Shadow Copy (影子拷贝)，新旧进程运行目录不同
				// 我们在这里快速重启，以将停机时间降至最低
				log.Printf("[PluginBridge] 正在快速重载插件: %s (v%s)...", id, v.Config.Version)

				// 停止旧进程
				pb.pluginManager.StopPlugin(id, v.Config.Version)

				// 立即启动新进程 (Shadow Copy 会确保没有文件冲突)
				pb.startAndRegisterPlugin(v)
			}
		}
	}

	return nil
}

func (pb *PluginBridge) startAndRegisterPlugin(v *core.Plugin) {
	// 启动进程
	if err := pb.pluginManager.StartPlugin(v.ID, v.Config.Version); err != nil {
		log.Errorf("启动插件 %s (v%s) 失败: %v", v.ID, v.Config.Version, err)
		return
	}

	// 包装并注册到 BotWorker 的插件管理器
	wrapper := &ExternalPluginWrapper{
		plugin: v,
		pm:     pb.pluginManager,
		bridge: pb,
	}
	pb.pluginManager.LoadPluginModule(wrapper, pb.server)
}

func (pb *PluginBridge) LoadInternalPlugins() error {
	// 加载 AIPlugin
	if pb.aiService != nil {
		aiPlugin := plugins.NewAIPlugin(pb.aiService)
		// 初始化数字员工服务
		empSvc := employee.NewEmployeeService(plugins.GlobalGORMDB)
		empSvc.SetAIService(pb.aiService)
		memSvc := employee.NewCognitiveMemoryService(plugins.GlobalGORMDB)
		taskSvc := employee.NewDigitalEmployeeTaskService(plugins.GlobalGORMDB, pb.aiService.GetMCPManager())
		taskSvc.SetAIService(pb.aiService)

		aiPlugin.SetEmployeeServices(empSvc, memSvc, taskSvc)
		if err := pb.pluginManager.LoadPluginModule(aiPlugin, pb.server); err != nil {
			log.Errorf("加载 AIPlugin 失败: %v", err)
		}
	}

	// 加载 PointsProxy
	pointsProxy := &plugins.PointsProxy{}
	if err := pb.pluginManager.LoadPluginModule(pointsProxy, pb.server); err != nil {
		return fmt.Errorf("加载 PointsProxy 失败: %v", err)
	}

	// 加载 PluginBuilder
	builder := plugins.NewBuilderPlugin(pb.pluginManager.GetPluginPath())
	if err := pb.pluginManager.LoadPluginModule(builder, pb.server); err != nil {
		return fmt.Errorf("加载 PluginBuilder 失败: %v", err)
	}
	return nil
}

func (pb *PluginBridge) LoadExternalPlugins() error {
	// 扫描所有配置的插件目录
	dirsToScan := []string{pb.pluginManager.GetPluginPath()}
	if pb.server.GetConfig() != nil {
		dirsToScan = append(dirsToScan, pb.server.GetConfig().Plugin.DevDirs...)
	}

	for _, dir := range dirsToScan {
		if dir == "" {
			continue
		}
		if err := pb.pluginManager.ScanPlugins(dir); err != nil {
			log.Errorf("[PluginBridge] 初始扫描插件目录失败 %s: %v", dir, err)
		}
	}

	// 启动所有扫描到的插件
	for id, versions := range pb.pluginManager.GetPlugins() {
		if id == "" {
			continue
		}
		for _, v := range versions {
			pb.startAndRegisterPlugin(v)
		}
	}

	// 注册全局动作路由，允许 Go 插件调用外部插件
	pb.server.SetActionRouter(func(pluginID string, action string, payload map[string]any) (any, error) {
		// 1. 首先检查是否是内部插件且目标是一个已注册的技能
		if pluginID != "" {
			internalPlugins := pb.pluginManager.GetInternalPlugins()
			if _, ok := internalPlugins[pluginID]; ok {
				// 如果是内部插件，尝试作为技能调用
				params := make(map[string]string)
				for k, v := range payload {
					params[k] = fmt.Sprintf("%v", v)
				}
				result, err := pb.server.InvokeSkill(action, params)
				if err == nil {
					return result, nil
				}
				// 如果技能未找到，继续尝试作为外部插件动作分发（可能是包装器）
			}
		}

		// 2. 分发给外部进程插件
		coreEvent := &core.EventMessage{
			ID:      fmt.Sprintf("action_%d", core.NextID()),
			Type:    "request",
			Name:    action,
			Payload: payload,
		}
		if pluginID != "" {
			// 目前简单处理，发送给所有版本的该 ID 插件
			pb.pluginManager.DispatchEventToPlugin(pluginID, "", coreEvent)
		} else {
			pb.pluginManager.DispatchEvent(coreEvent)
		}
		return "sent", nil
	})

	// 注册全局动作处理 (处理插件发出的动作，如发送消息)
	pb.pluginManager.RegisterActionHandler(func(p *core.Plugin, a *core.Action) {
		log.Printf("[PluginAction] Received action from plugin %s: %+v", p.ID, a)

		// 处理跨插件技能调用 (当 routeSkillCall 返回 false 时落到这里)
		if a.Type == "call_skill" {
			go func() {
				skillName, _ := a.Payload["skill_name"].(string)
				payload, _ := a.Payload["payload"].(map[string]any)

				// 转换 payload 为 map[string]string 以适配 InvokeSkill
				params := make(map[string]string)
				for k, v := range payload {
					params[k] = fmt.Sprintf("%v", v)
				}

				result, err := pb.server.InvokeSkill(skillName, params)

				// 发送响应回插件
				correlationID := a.CorrelationID
				if correlationID == "" {
					correlationID, _ = a.Payload["correlation_id"].(string)
				}

				if correlationID != "" {
					resp := &core.EventMessage{
						ID:            fmt.Sprintf("resp_%d", core.NextID()),
						Type:          "event",
						Name:          "skill_response",
						CorrelationId: correlationID,
						Payload: map[string]any{
							"result": result,
							"error":  "",
						},
					}
					if err != nil {
						resp.Payload.(map[string]any)["error"] = err.Error()
					}
					pb.pluginManager.DispatchEventToPlugin(p.ID, p.Config.Version, resp)
				}
			}()
			return
		}

		// 处理存储操作 (内部逻辑)
		if a.Type == "storage.get" || a.Type == "storage.set" || a.Type == "storage.delete" || a.Type == "storage.exists" {
			go pb.handleStorageAction(p, a)
			return
		}

		platform := ""
		selfID := ""
		if a.Payload != nil {
			platform, _ = a.Payload["platform"].(string)
			selfID, _ = a.Payload["self_id"].(string)
		}

		// 处理同步响应
		if a.Type == "reply" || a.Type == "skill_response" {
			if targetID, ok := a.Payload["target_event_id"].(string); ok {
				if ch, ok := pb.pendingResponses.Load(targetID); ok {
					log.Printf("[PluginAction] Fulfilling pending response for %s", targetID)

					// 提取消息内容
					message := a.Text
					if message == "" {
						message = fmt.Sprintf("%v", a.Payload["message"])
					}
					if message == "<nil>" || message == "" {
						message = fmt.Sprintf("%v", a.Payload["text"])
					}
					if message == "<nil>" || message == "" {
						message = fmt.Sprintf("%v", a.Payload["result"])
					}

					log.Printf("[PluginAction] Extracted result for %s: %v", targetID, message)

					resultStr := fmt.Sprintf("%v", message)
					select {
					case ch.(chan string) <- resultStr:
						log.Printf("[PluginAction] Successfully sent result to channel for %s", targetID)
					default:
						log.Printf("[PluginAction] Warning: Could not send to channel for %s (maybe timed out or full)", targetID)
					}

					// 如果是同步响应，我们不需要继续执行发送消息的逻辑
					if a.Type == "skill_response" {
						return
					}
				}
			}
		}

		// 构造通用参数
		params := make(map[string]any)
		if a.Payload != nil {
			for k, v := range a.Payload {
				params[k] = v
			}
		}

		// 兼容性处理
		if v, ok := params["user_id"]; !ok || fmt.Sprintf("%v", v) == "0" || fmt.Sprintf("%v", v) == "" {
			if a.Target != "" {
				if id, err := strconv.ParseInt(a.Target, 10, 64); err == nil {
					params["user_id"] = id
				} else {
					params["user_id"] = a.Target
				}
			}
		}
		if v, ok := params["group_id"]; !ok || fmt.Sprintf("%v", v) == "0" || fmt.Sprintf("%v", v) == "" {
			if a.TargetID != "" {
				if id, err := strconv.ParseInt(a.TargetID, 10, 64); err == nil {
					params["group_id"] = id
				} else {
					params["group_id"] = a.TargetID
				}
			}
		}

		if v, ok := params["message"]; !ok || v == nil || v == "" {
			if a.Text != "" {
				params["message"] = a.Text
			} else if text, ok := params["text"]; ok && text != nil && text != "" {
				params["message"] = text
			}
		}

		if a.CorrelationID != "" && params["correlation_id"] == nil {
			params["correlation_id"] = a.CorrelationID
		}

		// 补充平台和机器人ID
		if _, ok := params["platform"]; !ok && platform != "" {
			params["platform"] = platform
		}
		if _, ok := params["self_id"]; !ok && selfID != "" {
			params["self_id"] = selfID
		}
		if _, ok := params["bot_id"]; !ok && selfID != "" {
			params["bot_id"] = selfID
		}

		// 调用通用动作转发
		actionType := a.Type
		if actionType == "send_message" || actionType == "reply" {
			actionType = "send_msg"
		}

		log.Printf("[PluginAction] Forwarding action %s to server with params: %+v", actionType, params)
		resp, err := pb.server.CallBotAction(actionType, params)
		if err != nil {
			log.Printf("[PluginAction] [ERROR] CallBotAction failed for action %s: %v", actionType, err)
		} else {
			log.Printf("[PluginAction] CallBotAction success for action %s, response: %+v", actionType, resp)
		}
	})

	// 桥接 OneBot 事件到所有外部插件
	handler := func(name string) func(e *onebot.Event) error {
		return func(e *onebot.Event) error {
			userID := e.UserID.String()
			if e.TargetUserID != "" {
				userID = e.TargetUserID
			}
			groupID := e.GroupID.String()
			if e.TargetGroupID != "" {
				groupID = e.TargetGroupID
			}

			payload := map[string]any{
				"from":     userID,
				"group_id": groupID,
				"user_id":  userID,
				"text":     e.RawMessage,
				"platform": e.Platform,
				"self_id":  fmt.Sprintf("%v", e.SelfID),
				"raw":      e,
			}

			// 如果是 meta_event，添加额外字段
			if e.PostType == "meta_event" {
				payload["meta_event_type"] = e.MetaEventType
			}

			coreEvent := &core.EventMessage{
				ID:      fmt.Sprintf("ob_%d", e.Time),
				Type:    "event",
				Name:    name,
				Payload: payload,
			}
			pb.pluginManager.DispatchEvent(coreEvent)
			return nil
		}
	}

	pb.server.OnMessage(handler("on_message"))
	pb.server.OnNotice(handler("on_notice"))
	pb.server.OnRequest(handler("on_request"))
	pb.server.OnEvent("meta_event", handler("on_meta_event"))

	return nil
}

func (pb *PluginBridge) handleStorageAction(p *core.Plugin, a *core.Action) {
	if plugins.GlobalRedis == nil {
		fmt.Printf("Redis 未初始化，无法执行存储操作 %s\n", a.Type)
		return
	}

	ctx := context.Background()
	store := session.NewRedisSessionStore(plugins.GlobalRedis.Client)
	correlationID, _ := a.Payload["correlation_id"].(string)
	key, _ := a.Payload["key"].(string)

	var result any
	var err error

	switch a.Type {
	case "storage.get":
		var val any
		err = store.Get(ctx, key, &val)
		if err == nil {
			result = val
		}
	case "storage.set":
		value := a.Payload["value"]
		expire, _ := a.Payload["expire"].(float64) // JSON numbers are float64
		err = store.Set(ctx, key, value, time.Duration(expire)*time.Second)
		result = "ok"
	case "storage.delete":
		err = store.Delete(ctx, key)
		result = "ok"
	case "storage.exists":
		var val any
		err = store.Get(ctx, key, &val)
		result = (err == nil)
	}

	// 发送响应回插件
	if correlationID != "" {
		resp := &core.EventMessage{
			ID:            fmt.Sprintf("resp_%d", core.NextID()),
			Type:          "response",
			Name:          a.Type + "_response",
			CorrelationId: correlationID,
			Payload: map[string]any{
				"value": result,
				"error": func() string {
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
			},
		}
		pb.pluginManager.DispatchEventToPlugin(p.ID, p.Config.Version, resp)
	}
}

// 实现plugin.Plugin接口的包装器
type ExternalPluginWrapper struct {
	plugin *core.Plugin
	pm     *core.PluginManager
	bridge *PluginBridge
}

func (w *ExternalPluginWrapper) Name() string {
	return w.plugin.Config.Name
}

func (w *ExternalPluginWrapper) Description() string {
	return w.plugin.Config.Description
}

func (w *ExternalPluginWrapper) Version() string {
	return w.plugin.Config.Version
}

func (w *ExternalPluginWrapper) Init(robot core.Robot) {
	// 1. 注册 Intents 为技能
	for _, intent := range w.plugin.Config.Intents {
		intentName := intent.Name
		capability := core.SkillCapability{
			Name:        intentName,
			Description: w.plugin.Config.Description,
			Regex:       intent.Regex,
			Usage:       fmt.Sprintf("Keywords: %v", intent.Keywords),
		}
		w.registerSkill(robot, capability)
	}

	// 2. 注册 Capabilities 为技能 (支持 sz84 等没有明确 intent 的传统插件)
	for _, capName := range w.plugin.Config.Capabilities {
		// 如果已经作为 intent 注册过了，跳过
		alreadyRegistered := false
		for _, intent := range w.plugin.Config.Intents {
			if intent.Name == capName {
				alreadyRegistered = true
				break
			}
		}
		if alreadyRegistered {
			continue
		}

		capability := core.SkillCapability{
			Name:        capName,
			Description: fmt.Sprintf("Capability: %s", capName),
			Usage:       fmt.Sprintf("Directly call capability %s", capName),
		}
		w.registerSkill(robot, capability)
	}
}

func (w *ExternalPluginWrapper) registerSkill(robot core.Robot, capability core.SkillCapability) {
	skillName := capability.Name
	robot.RegisterSkill(capability, func(params map[string]string) (string, error) {
		log.Printf("[Skill] Triggered skill: %s, params: %+v", skillName, params)

		eventID := fmt.Sprintf("skill_%d", core.NextID())
		// 将技能调用转换为事件发送给插件
		coreEvent := &core.EventMessage{
			ID:   eventID,
			Type: "request",
			Name: "call_skill",
			Payload: map[string]any{
				"skill":    skillName,
				"params":   params,
				"from":     params["user_id"],
				"group_id": params["group_id"],
			},
		}

		// 创建等待响应的通道
		respCh := make(chan string, 1)
		w.bridge.pendingResponses.Store(eventID, respCh)
		defer w.bridge.pendingResponses.Delete(eventID)

		// 分发事件给插件
		w.pm.DispatchEventToPlugin(w.plugin.ID, w.plugin.Version, coreEvent)

		// 等待响应或超时
		select {
		case result := <-respCh:
			log.Printf("[Skill] Received synchronous result for %s: %s", skillName, result)
			return result, nil
		case <-time.After(10 * time.Second):
			log.Printf("[Skill] Timeout waiting for result of %s", skillName)
			return fmt.Sprintf("Skill %s timed out", skillName), nil
		}
	})
}

// 实现 SkillCapable 接口
func (w *ExternalPluginWrapper) GetSkills() []core.SkillCapability {
	skills := []core.SkillCapability{}
	// 1. Intents
	for _, intent := range w.plugin.Config.Intents {
		skills = append(skills, core.SkillCapability{
			Name:        intent.Name,
			Description: w.plugin.Config.Description,
			Usage:       fmt.Sprintf("Keywords: %v", intent.Keywords),
			Regex:       intent.Regex,
		})
	}
	// 2. Capabilities
	for _, capName := range w.plugin.Config.Capabilities {
		alreadyAdded := false
		for _, s := range skills {
			if s.Name == capName {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}
		skills = append(skills, core.SkillCapability{
			Name:        capName,
			Description: fmt.Sprintf("Capability: %s", capName),
			Usage:       fmt.Sprintf("Directly call capability %s", capName),
		})
	}
	return skills
}
