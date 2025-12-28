package server

import (
	"BotMatrix/common"
	"BotMatrix/common/log"
	"botworker/internal/config"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type CombinedServer struct {
	wsServer      *WebSocketServer
	httpServer    *HTTPServer
	pluginManager *plugin.Manager
	redisClient   *redis.Client
	config        *config.Config
	skillHandlers map[string]func(map[string]string) (string, error)
}

func NewCombinedServer(cfg *config.Config, rdb *redis.Client) *CombinedServer {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	server := &CombinedServer{
		wsServer:      NewWebSocketServer(&cfg.WebSocket),
		httpServer:    NewHTTPServer(&cfg.HTTP),
		redisClient:   rdb,
		config:        cfg,
		skillHandlers: make(map[string]func(map[string]string) (string, error)),
	}
	server.pluginManager = plugin.NewManager(server)
	return server
}

// 实现plugin.Robot接口
func (s *CombinedServer) OnMessage(fn onebot.EventHandler) {
	s.wsServer.OnMessage(fn)
	s.httpServer.OnMessage(fn)
}

func (s *CombinedServer) OnNotice(fn onebot.EventHandler) {
	s.wsServer.OnNotice(fn)
	s.httpServer.OnNotice(fn)
}

func (s *CombinedServer) OnRequest(fn onebot.EventHandler) {
	s.wsServer.OnRequest(fn)
	s.httpServer.OnRequest(fn)
}

func (s *CombinedServer) OnEvent(eventName string, fn onebot.EventHandler) {
	s.wsServer.OnEvent(eventName, fn)
	s.httpServer.OnEvent(eventName, fn)
}

func (s *CombinedServer) HandleAPI(action string, fn onebot.RequestHandler) {
	s.wsServer.HandleAPI(action, fn)
	s.httpServer.HandleAPI(action, fn)
}

func (s *CombinedServer) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	// 优先使用WebSocket发送消息
	return s.wsServer.SendMessage(params)
}

func (s *CombinedServer) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	return s.wsServer.DeleteMessage(params)
}

func (s *CombinedServer) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	return s.wsServer.SendLike(params)
}

func (s *CombinedServer) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	return s.wsServer.SetGroupKick(params)
}

func (s *CombinedServer) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	return s.wsServer.SetGroupBan(params)
}

func (s *CombinedServer) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	return s.wsServer.GetGroupMemberList(params)
}

func (s *CombinedServer) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	return s.wsServer.GetGroupMemberInfo(params)
}

func (s *CombinedServer) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	return s.wsServer.SetGroupSpecialTitle(params)
}

func (s *CombinedServer) GetSelfID() int64 {
	return s.wsServer.GetSelfID()
}

// Session & State Management 实现
func (s *CombinedServer) GetSessionContext(platform, userID string) (*common.SessionContext, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.GetSessionContext(platform, userID)
}

func (s *CombinedServer) SetSessionState(platform, userID string, state common.SessionState, ttl time.Duration) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.SetSessionState(platform, userID, state, ttl)
}

func (s *CombinedServer) GetSessionState(platform, userID string) (*common.SessionState, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.GetSessionState(platform, userID)
}

func (s *CombinedServer) ClearSessionState(platform, userID string) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.ClearSessionState(platform, userID)
}

// HandleSkill 注册技能处理器
func (s *CombinedServer) HandleSkill(skillName string, fn func(params map[string]string) (string, error)) {
	s.skillHandlers[skillName] = fn
}

// 插件管理
func (s *CombinedServer) GetPluginManager() *plugin.Manager {
	return s.pluginManager
}

// 启动服务器
func (s *CombinedServer) Run() error {
	// 启动HTTP服务器
	go func() {
		if err := s.httpServer.Run(); err != nil {
			panic(err)
		}
	}()

	// 启动WebSocket服务器
	go func() {
		if err := s.wsServer.Run(); err != nil {
			log.Printf("[Combined] WebSocket server error: %v", err)
		}
	}()

	// 报备能力给 BotNexus
	go s.reportCapabilities()

	// 启动Redis队列监听 (如果配置了Redis)
	if s.redisClient != nil {
		go s.startRedisQueueListener()
	}

	// 保持主运行状态
	select {}
}

func (s *CombinedServer) reportCapabilities() {
	if s.redisClient == nil {
		return
	}

	// 如果禁用了技能系统，则不报备能力
	if !s.config.EnableSkill {
		log.Printf("[SKILL] Skill system is disabled, skipping capability reporting")
		return
	}

	// 等待插件加载完成
	time.Sleep(2 * time.Second)

	capabilities := []map[string]any{}
	for _, p := range s.pluginManager.GetPlugins() {
		if cp, ok := p.(plugin.SkillCapable); ok {
			for _, skill := range cp.GetSkills() {
				capabilities = append(capabilities, map[string]any{
					"name":        skill.Name,
					"description": skill.Description,
					"usage":       skill.Usage,
					"params":      skill.Params,
				})
			}
		}
	}

	if len(capabilities) == 0 {
		return
	}

	regMsg := map[string]any{
		"type":         "worker_register",
		"worker_id":    s.config.WorkerID,
		"capabilities": capabilities,
		"timestamp":    time.Now().Unix(),
	}

	payload, _ := json.Marshal(regMsg)
	ctx := context.Background()
	err := s.redisClient.Publish(ctx, "botmatrix:worker:register", payload).Err()
	if err != nil {
		log.Printf("[Combined] Failed to report capabilities: %v", err)
	} else {
		log.Printf("[Combined] Successfully reported %d capabilities to BotNexus", len(capabilities))
	}
}

func (s *CombinedServer) startRedisQueueListener() {
	if s.redisClient == nil {
		return
	}

	workerID := s.config.WorkerID
	// 监听两个队列：自己的专用队列和公共队列
	queues := []string{
		"botmatrix:queue:default",
	}
	if workerID != "" {
		// 优先处理专用队列
		queues = append([]string{fmt.Sprintf("botmatrix:queue:worker:%s", workerID)}, queues...)
	}

	log.Printf("[RedisQueue] Starting listener for queues: %v", queues)

	ctx := context.Background()
	for {
		// 使用 BLPOP 阻塞式获取消息，超时时间设为 30 秒
		result, err := s.redisClient.BLPop(ctx, 30*time.Second, queues...).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("[RedisQueue] Error popping from queue: %v", err)
				time.Sleep(5 * time.Second) // 出错后等待重试
			}
			continue
		}

		if len(result) < 2 {
			continue
		}

		// result[0] 是队列名，result[1] 是消息内容
		queueName := result[0]
		payload := result[1]

		log.Printf("[RedisQueue] Received message from %s", queueName)

		// 解析消息并分发
		var msg map[string]any
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			log.Printf("[RedisQueue] Failed to unmarshal message: %v", err)
			continue
		}

// 异步处理消息，避免阻塞监听器
	go func(queue string, m map[string]any) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[RedisQueue] Panic in message processor: %v", r)
			}
		}()
		s.processQueueMessage(m)
	}(queueName, msg)
}
}

func (s *CombinedServer) processQueueMessage(msg map[string]any) {
	// 1. 检查是否为指令 (skill_call)
	if msgType, ok := msg["type"].(string); ok && msgType == "skill_call" {
		if s.config.EnableSkill {
			s.handleSkillCall(msg)
		} else {
			log.Printf("[SKILL] Skill system is disabled, ignoring skill_call message")
		}
		return
	}

	// 2. 检查是否为工作节点注册响应或其他控制消息 (可选)
	if msgType, ok := msg["type"].(string); ok && msgType == "control" {
		log.Printf("[RedisQueue] Received control message: %v", msg)
		return
	}

	// 3. 直接分发给 OneBot 事件处理流程
	// 不再通过 pluginManager.HandleEvent 中转，以减少调用栈深度和潜在的循环依赖
	s.HandleQueueEvent(msg)
}

func (s *CombinedServer) handleSkillCall(msg map[string]any) {
	skillName, _ := msg["skill"].(string)
	paramsMap, _ := msg["params"].(map[string]any)
	taskID := fmt.Sprint(msg["task_id"])
	executionID := fmt.Sprint(msg["execution_id"])

	// 转换参数为 map[string]string
	params := make(map[string]string)
	for k, v := range paramsMap {
		params[k] = fmt.Sprint(v)
	}

	log.Printf("[SkillCall] Handling skill: %s (TaskID: %s, ExecID: %s) with params: %v", skillName, taskID, executionID, params)

	handler, ok := s.skillHandlers[skillName]
	if !ok {
		log.Printf("[SkillCall] No handler for skill: %s", skillName)
		s.reportSkillResult(taskID, executionID, skillName, "", fmt.Errorf("no handler for skill: %s", skillName))
		return
	}

	result, err := handler(params)
	if err != nil {
		log.Printf("[SkillCall] Error executing skill %s: %v", skillName, err)
		s.reportSkillResult(taskID, executionID, skillName, "", err)
		return
	}

	log.Printf("[SkillCall] Skill %s executed successfully: %s", skillName, result)

	// 将结果报备回 BotNexus (如果配置了 Redis)
	s.reportSkillResult(taskID, executionID, skillName, result, nil)
}

func (s *CombinedServer) reportSkillResult(taskID, executionID, skillName, result string, err error) {
	if taskID == "" || taskID == "<nil>" {
		return
	}

	status := "success"
	errorMessage := ""
	if err != nil {
		status = "failed"
		errorMessage = err.Error()
	}

	report := map[string]any{
		"type":         "skill_result",
		"worker_id":    s.config.WorkerID,
		"task_id":      taskID,
		"execution_id": executionID,
		"skill":        skillName,
		"status":       status,
		"result":       result,
		"error":        errorMessage,
		"timestamp":    time.Now().Unix(),
	}

	payload, _ := json.Marshal(report)

	// 1. 尝试通过 Redis 报备 (Publish)
	if s.redisClient != nil {
		ctx := context.Background()
		pubErr := s.redisClient.Publish(ctx, "botmatrix:worker:skill_result", payload).Err()
		if pubErr == nil {
			log.Printf("[SkillCall] Reported result for task %s via Redis", taskID)
			return
		}
		log.Printf("[SkillCall] Failed to report result via Redis: %v. Trying WebSocket.", pubErr)
	}

	// 2. 尝试通过 WebSocket 报备 (如果连接可用)
	if s.wsServer != nil {
		err := s.wsServer.BroadcastJSON(report)
		if err == nil {
			log.Printf("[SkillCall] Reported result for task %s via WebSocket", taskID)
			return
		}
		log.Printf("[SkillCall] Failed to report result via WebSocket: %v", err)
	}
}

func (s *CombinedServer) HandleQueueEvent(msg map[string]any) {
	// 记录原始消息的一些关键信息，方便调试
	postType, _ := msg["post_type"].(string)
	messageType, _ := msg["message_type"].(string)
	log.Printf("[Combined] Processing queue event: post_type=%s, message_type=%s", postType, messageType)

	// 将 map 转换为 onebot.Event
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Combined] Failed to marshal queue message: %v", err)
		return
	}

	var event onebot.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[Combined] Failed to unmarshal queue event: %v", err)
		return
	}

	// 处理 QQGuild ID 生成（确保 ID 映射正确）
	processEventIDs(&event)

	// 分发到内部处理器
	s.dispatchInternalEvent(&event)
}

func (s *CombinedServer) dispatchInternalEvent(event *onebot.Event) {
	// 这里的逻辑应该与 WebSocketServer.handleEvent 保持同步
	// 或者直接让 CombinedServer 拥有自己的 handler 列表

	// 目前简单做法是调用 wsServer 的处理逻辑（如果它暴露了）
	// 或者在 CombinedServer 中维护一套 handler
	
	// 实际上，CombinedServer 的 OnMessage 等方法是将 handler 注册到了 wsServer 和 httpServer
	// 所以我们应该从 wsServer 中获取 handler 并执行，或者在 CombinedServer 中也存一份

	// 既然 CombinedServer 的 Run 方法中启动了 wsServer，
	// 我们可以考虑让 CombinedServer 统一管理 handler
	
	// 为了不破坏现有结构，我们暂时通过反射或者修改 WebSocketServer 来支持
	// 最好的办法是在 CombinedServer 中实现一套通用的分发逻辑
	
	// 分发到对应的事件处理器 (从 wsServer 借用逻辑)
	switch event.PostType {
	case "message":
		// 这里需要访问 wsServer 的私有字段，或者让 wsServer 暴露一个 Dispatch 方法
		s.wsServer.DispatchEvent(event)
	case "notice":
		s.wsServer.DispatchEvent(event)
	case "request":
		s.wsServer.DispatchEvent(event)
	}
}

func (s *CombinedServer) Stop() {
	s.wsServer.Stop()
	s.httpServer.Stop()
}
