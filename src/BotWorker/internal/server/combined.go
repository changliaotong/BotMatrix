package server

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/bot"
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	commononebot "BotMatrix/common/onebot"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/session"
	"BotMatrix/common/tasks"
	"BotMatrix/common/types"
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/redis"
	"botworker/plugins"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type CombinedServer struct {
	botService             *bot.BaseBot
	wsServer               *WebSocketServer
	httpServer             *HTTPServer
	pluginManager          *core.PluginManager
	redisClient            *redis.Client
	config                 *config.Config
	actionRouter           func(string, string, map[string]any) (any, error)
	lastSelfID             int64
	lastPlatform           string
	skills                 map[string]core.Skill
	skillCapabilities      []core.SkillCapability
	skillsMu               sync.RWMutex
	aiService              ai.AIService
	employeeService        employee.DigitalEmployeeService
	cognitiveMemoryService employee.CognitiveMemoryService
	taskManager            *tasks.TaskManager
}

func (s *CombinedServer) SetAIService(aiSvc ai.AIService) {
	s.aiService = aiSvc
	if plugins.GlobalGORMDB != nil {
		s.employeeService = employee.NewEmployeeService(plugins.GlobalGORMDB)
		s.cognitiveMemoryService = employee.NewCognitiveMemoryService(plugins.GlobalGORMDB)

		// 初始化任务管理器
		if s.taskManager == nil {
			s.taskManager = tasks.NewTaskManager(plugins.GlobalGORMDB, s.redisClient.Client, s, s.config.WorkerID)
			s.taskManager.AI.SetAIService(aiSvc)
			// Worker 仅作为任务生成端和同步端，不执行调度触发
			s.taskManager.Start(false)
			log.Info("[Worker] TaskManager started (Scheduler Disabled)")

			// 注册任务系统消息处理器
			s.OnMessage(func(e *onebot.Event) error {
				if s.taskManager == nil {
					return nil
				}
				// 转换 OneBot 事件为任务系统需要的格式
				ctx := context.Background()
				botID := fmt.Sprintf("%v", e.SelfID)
				groupID := e.GroupID.String()
				userID := e.UserID.String()
				content := e.RawMessage

				// 异步处理消息，避免阻塞消息流水线
				go func() {
					if err := s.taskManager.ProcessChatMessage(ctx, botID, groupID, userID, content); err != nil {
						log.Errorf("[Worker] TaskManager.ProcessChatMessage error: %v", err)
					}
				}()
				return nil
			})
		}
	}
}

func NewCombinedServer(botService *bot.BaseBot, cfg *config.Config, rdb *redis.Client) *CombinedServer {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	server := &CombinedServer{
		botService:    botService,
		wsServer:      NewWebSocketServer(&cfg.WebSocket),
		httpServer:    NewHTTPServer(&cfg.HTTP),
		redisClient:   rdb,
		config:        cfg,
		skills:        make(map[string]core.Skill),
		pluginManager: core.NewPluginManager(),
	}
	server.registerStorageHandlers()
	server.registerCoreHandlers()
	return server
}

func (s *CombinedServer) registerCoreHandlers() {
	s.OnMessage(func(e *onebot.Event) error {
		if plugins.GlobalStore == nil {
			return nil
		}

		// 1. 记录消息日志
		logEntry := &models.MessageLog{
			BotID:     fmt.Sprintf("%v", e.SelfID),
			UserID:    e.UserID.String(),
			GroupID:   e.GroupID.String(),
			Content:   e.RawMessage,
			Platform:  e.Platform,
			Direction: "incoming",
			CreatedAt: time.Now(),
		}
		if raw, err := json.Marshal(e); err == nil {
			logEntry.RawData = string(raw)
		}
		_ = plugins.GlobalStore.Messages.LogMessage(logEntry)

		// 2. 更新消息统计
		if e.GroupID.String() != "" && e.UserID.String() != "" {
			_ = plugins.GlobalStore.Messages.UpdateStat(e.GroupID.String(), e.UserID.String(), time.Now(), 1)
		}

		// 3. 异步更新缓存 (不阻塞主流程)
		go func() {
			// 更新群组缓存
			if e.GroupID.String() != "" {
				groupCache := &models.GroupCache{
					GroupID:  e.GroupID.String(),
					BotID:    fmt.Sprintf("%v", e.SelfID),
					LastSeen: time.Now(),
				}
				// 尝试获取群名（如果 event 中有）
				// 注意：OneBot 事件中不一定有群名，通常需要调用 API 获取
				_ = plugins.GlobalStore.Caches.UpdateGroupCache(groupCache)
			}

			// 更新成员缓存
			if e.GroupID.String() != "" && e.UserID.String() != "" {
				memberCache := &models.MemberCache{
					GroupID:  e.GroupID.String(),
					UserID:   e.UserID.String(),
					LastSeen: time.Now(),
				}
				// 尝试获取昵称（如果 event 中有）
				if e.Sender.Nickname != "" {
					memberCache.Nickname = e.Sender.Nickname
				}
				if e.Sender.Card != "" {
					memberCache.Card = e.Sender.Card
				}
				if e.Sender.Role != "" {
					memberCache.Role = e.Sender.Role
				}
				_ = plugins.GlobalStore.Caches.UpdateMemberCache(memberCache)
			}
		}()
		return nil
	})
}

func (s *CombinedServer) registerStorageHandlers() {
	s.HandleAPI("storage.get", func(req *onebot.Request) (*onebot.Response, error) {
		params, ok := req.Params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid parameters")
		}
		key, _ := params["key"].(string)
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}

		var value any
		ctx := context.Background()
		if s.redisClient != nil {
			store := session.NewRedisSessionStore(s.redisClient.Client)
			err := store.Get(ctx, key, &value)
			if err == nil {
				return &onebot.Response{Data: map[string]any{"value": value}}, nil
			}
		}

		// Fallback to Database for specific keys
		// Example: table:users:id:{userId}:is_super_points
		var userId int64
		if n, err := fmt.Sscanf(key, "table:users:id:%d:is_super_points", &userId); err == nil && n == 1 {
			if database := plugins.GlobalDB; database != nil {
				user, err := db.GetUserByUserID(database, userId)
				if err == nil && user != nil {
					return &onebot.Response{Data: map[string]any{"value": user.IsSuperPoints}}, nil
				}
			}
		}

		// Example: table:users:id:{userId}:global_points
		if n, err := fmt.Sscanf(key, "table:users:id:%d:global_points", &userId); err == nil && n == 1 {
			if database := plugins.GlobalDB; database != nil {
				user, err := db.GetUserByUserID(database, userId)
				if err == nil && user != nil {
					return &onebot.Response{Data: map[string]any{"value": user.Points}}, nil
				}
			}
		}

		// Example: table:bot_friends:bot:{botId}:user:{userId}:local_points
		var botId int64
		if n, err := fmt.Sscanf(key, "table:bot_friends:bot:%d:user:%d:local_points", &botId, &userId); err == nil && n == 2 {
			if database := plugins.GlobalDB; database != nil {
				points, err := db.GetLocalPoints(database, botId, userId)
				if err == nil {
					return &onebot.Response{Data: map[string]any{"value": points}}, nil
				}
			}
		}

		// Example: table:group_members:group:{groupId}:user:{userId}:points
		var groupId int64
		if n, err := fmt.Sscanf(key, "table:group_members:group:%d:user:%d:points", &groupId, &userId); err == nil && n == 2 {
			if database := plugins.GlobalDB; database != nil {
				points, err := db.GetGroupPoints(database, userId, groupId)
				if err == nil {
					return &onebot.Response{Data: map[string]any{"value": points}}, nil
				}
			}
		}

		return &onebot.Response{Data: map[string]any{"value": nil}}, nil
	})

	s.HandleAPI("storage.set", func(req *onebot.Request) (*onebot.Response, error) {
		params, ok := req.Params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid parameters")
		}
		key, _ := params["key"].(string)
		value := params["value"]
		expirationMs, _ := params["expiration"].(float64)

		if key == "" {
			return nil, fmt.Errorf("key is required")
		}

		ctx := context.Background()
		if s.redisClient != nil {
			store := session.NewRedisSessionStore(s.redisClient.Client)
			expiration := time.Duration(expirationMs) * time.Millisecond
			err := store.Set(ctx, key, value, expiration)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("storage not available (redis not initialized)")
		}

		// Sync to Database for specific keys
		// Example: table:users:id:{userId}:is_super_points
		var userId int64
		if n, err := fmt.Sscanf(key, "table:users:id:%d:is_super_points", &userId); err == nil && n == 1 {
			if isSuper, ok := value.(bool); ok {
				if database := plugins.GlobalDB; database != nil {
					_ = db.UpdateUserSuperPoints(database, userId, isSuper)
				}
			}
		}

		// Example: table:users:id:{userId}:global_points
		if n, err := fmt.Sscanf(key, "table:users:id:%d:global_points", &userId); err == nil && n == 1 {
			var points int
			switch v := value.(type) {
			case int:
				points = v
			case int64:
				points = int(v)
			case float64:
				points = int(v)
			}
			if database := plugins.GlobalDB; database != nil {
				_ = db.UpdateUserPoints(database, userId, points)
			}
		}

		// Example: table:bot_friends:bot:{botId}:user:{userId}:local_points
		var botId int64
		if n, err := fmt.Sscanf(key, "table:bot_friends:bot:%d:user:%d:local_points", &botId, &userId); err == nil && n == 2 {
			var points int64
			switch v := value.(type) {
			case int:
				points = int64(v)
			case int64:
				points = v
			case float64:
				points = int64(v)
			}
			if database := plugins.GlobalDB; database != nil {
				_ = db.UpdateLocalPoints(database, botId, userId, points)
			}
		}

		// Example: table:group_members:group:{groupId}:user:{userId}:points
		var groupId int64
		if n, err := fmt.Sscanf(key, "table:group_members:group:%d:user:%d:points", &groupId, &userId); err == nil && n == 2 {
			var points int64
			switch v := value.(type) {
			case int:
				points = int64(v)
			case int64:
				points = v
			case float64:
				points = int64(v)
			}
			if database := plugins.GlobalDB; database != nil {
				_ = db.UpdateGroupPoints(database, userId, groupId, points)
			}
		}

		return &onebot.Response{Data: map[string]any{"status": "ok"}}, nil
	})

	s.HandleAPI("storage.delete", func(req *onebot.Request) (*onebot.Response, error) {
		params, ok := req.Params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid parameters")
		}
		key, _ := params["key"].(string)
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}

		ctx := context.Background()
		if s.redisClient != nil {
			store := session.NewRedisSessionStore(s.redisClient.Client)
			err := store.Delete(ctx, key)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("storage not available (redis not initialized)")
		}

		return &onebot.Response{Data: map[string]any{"status": "ok"}}, nil
	})

	s.HandleAPI("storage.exists", func(req *onebot.Request) (*onebot.Response, error) {
		params, ok := req.Params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid parameters")
		}
		key, _ := params["key"].(string)
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}

		ctx := context.Background()
		if s.redisClient != nil {
			store := session.NewRedisSessionStore(s.redisClient.Client)
			exists, err := store.Exists(ctx, key)
			if err != nil {
				return nil, err
			}
			return &onebot.Response{Data: map[string]any{"exists": exists}}, nil
		} else {
			return nil, fmt.Errorf("storage not available (redis not initialized)")
		}
	})
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

func (s *CombinedServer) HandleAPI(action string, fn any) {
	if handler, ok := fn.(onebot.RequestHandler); ok {
		s.wsServer.HandleAPI(action, handler)
		s.httpServer.HandleAPI(action, handler)
	} else if handler, ok := fn.(onebot.EventHandler); ok {
		// 如果是事件处理器，根据 action 名称注册
		switch action {
		case "on_message":
			s.OnMessage(handler)
		case "on_notice":
			s.OnNotice(handler)
		case "on_request":
			s.OnRequest(handler)
		default:
			s.OnEvent(action, handler)
		}
	} else if handler, ok := fn.(func(map[string]any)); ok {
		// 兼容 map[string]any 类型的处理器
		wrappedHandler := func(e *onebot.Event) error {
			payload := map[string]any{
				"from":     e.UserID.String(),
				"group_id": e.GroupID.String(),
				"user_id":  e.UserID.String(),
				"text":     e.RawMessage,
				"platform": e.Platform,
				"self_id":  fmt.Sprintf("%v", e.SelfID),
			}
			handler(payload)
			return nil
		}
		s.HandleAPI(action, wrappedHandler)
	}
}

func (s *CombinedServer) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	log.Printf("[Worker] Sending message: %v", params.Message)
	// 优先使用WebSocket发送消息
	resp, err := s.wsServer.SendMessage(params)
	if err == nil {
		return resp, nil
	}

	// 转发动作给 BotNexus
	log.Printf("[Worker] Forwarding action '%s' to BotNexus via Redis", "send_msg")
	if pubErr := s.publishActionToNexus("send_msg", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}

	return resp, err
}

func (s *CombinedServer) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	resp, err := s.wsServer.DeleteMessage(params)
	if err == nil {
		return resp, nil
	}
	if pubErr := s.publishActionToNexus("delete_msg", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}
	return resp, err
}

func (s *CombinedServer) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	resp, err := s.wsServer.SendLike(params)
	if err == nil {
		return resp, nil
	}
	if pubErr := s.publishActionToNexus("send_like", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}
	return resp, err
}

func (s *CombinedServer) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	resp, err := s.wsServer.SetGroupKick(params)
	if err == nil {
		return resp, nil
	}
	if pubErr := s.publishActionToNexus("set_group_kick", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}
	return resp, err
}

func (s *CombinedServer) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	resp, err := s.wsServer.SetGroupBan(params)
	if err == nil {
		return resp, nil
	}
	if pubErr := s.publishActionToNexus("set_group_ban", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}
	return resp, err
}

func (s *CombinedServer) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	return s.wsServer.GetGroupMemberList(params)
}

func (s *CombinedServer) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	return s.wsServer.GetGroupMemberInfo(params)
}

func (s *CombinedServer) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	resp, err := s.wsServer.SetGroupSpecialTitle(params)
	if err == nil {
		return resp, nil
	}
	if pubErr := s.publishActionToNexus("set_group_special_title", params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}
	return resp, err
}

func (s *CombinedServer) CallBotAction(action string, params any) (any, error) {
	// 优先使用WebSocket发送消息
	resp, err := s.wsServer.CallAction(action, params)
	if err == nil {
		return resp, nil
	}

	// 转发动作给 BotNexus
	log.Printf("[Combined] Forwarding action '%s' to BotNexus via Redis", action)
	if pubErr := s.publishActionToNexus(action, params); pubErr == nil {
		return &onebot.Response{Status: "ok"}, nil
	}

	return resp, err
}

func (s *CombinedServer) GetSelfID() int64 {
	if s.lastSelfID != 0 {
		return s.lastSelfID
	}
	return s.wsServer.GetSelfID()
}

// --- BotManager Interface Implementation ---

func (s *CombinedServer) SendBotAction(botID string, action string, params any) error {
	_, err := s.CallBotAction(action, params)
	return err
}

func (s *CombinedServer) SendToWorker(workerID string, msg types.WorkerCommand) error {
	// Worker 无法直接发送给另一个 Worker，转发给 Nexus
	return s.botService.SendNexusCommand("worker_command", map[string]any{
		"target_worker": workerID,
		"command":       msg,
	})
}

func (s *CombinedServer) FindWorkerBySkill(skillName string) string {
	// 目前简单返回自己，后续可以通过 Nexus 查找
	return s.config.WorkerID
}

func (s *CombinedServer) GetTags(targetType string, targetID string) []string {
	if plugins.GlobalGORMDB == nil {
		return nil
	}
	var tags []models.Tag
	if err := plugins.GlobalGORMDB.Where("type = ? AND target_id = ?", targetType, targetID).Find(&tags).Error; err != nil {
		log.Printf("[BotWorker] GetTags error: %v", err)
		return nil
	}
	var names []string
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}

func (s *CombinedServer) GetTargetsByTags(targetType string, tags []string, logic string) []string {
	if plugins.GlobalGORMDB == nil || len(tags) == 0 {
		return nil
	}

	var results []string
	if logic == "AND" {
		// 所有的标签都必须存在
		err := plugins.GlobalGORMDB.Model(&models.Tag{}).
			Where("type = ? AND name IN ?", targetType, tags).
			Group("target_id").
			Having("COUNT(DISTINCT name) = ?", len(tags)).
			Pluck("target_id", &results).Error
		if err != nil {
			log.Printf("[BotWorker] GetTargetsByTags error: %v", err)
			return nil
		}
	} else {
		// 只要有一个标签存在即可 (OR)
		err := plugins.GlobalGORMDB.Model(&models.Tag{}).
			Where("type = ? AND name IN ?", targetType, tags).
			Distinct("target_id").
			Pluck("target_id", &results).Error
		if err != nil {
			log.Printf("[BotWorker] GetTargetsByTags error: %v", err)
			return nil
		}
	}

	return results
}

func (s *CombinedServer) GetGroupMembers(botID string, groupID string) ([]types.MemberInfo, error) {
	gid, _ := strconv.ParseInt(groupID, 10, 64)
	resp, err := s.wsServer.GetGroupMemberList(&onebot.GetGroupMemberListParams{
		GroupID: onebot.FlexibleInt64(gid),
	})
	if err != nil {
		return nil, err
	}

	members, ok := resp.Data.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response data type: %T", resp.Data)
	}

	var result []types.MemberInfo
	for _, mAny := range members {
		mMap, ok := mAny.(map[string]any)
		if !ok {
			continue
		}

		m := types.MemberInfo{
			GroupID:  groupID,
			UserID:   fmt.Sprintf("%v", mMap["user_id"]),
			Nickname: fmt.Sprintf("%v", mMap["nickname"]),
			Card:     fmt.Sprintf("%v", mMap["card"]),
			Role:     fmt.Sprintf("%v", mMap["role"]),
			BotID:    botID,
		}
		result = append(result, m)
	}
	return result, nil
}

func (s *CombinedServer) publishActionToNexus(action string, params any) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	msg := map[string]any{
		"type":      "action",
		"action":    action,
		"params":    params,
		"worker_id": s.config.WorkerID,
		"timestamp": time.Now().Unix(),
	}

	// 提取 platform 和 self_id
	var platform, selfID, groupID, userID, message string

	// 尝试从 params 中提取 (如果 params 是 struct 指针)
	if params != nil {
		v := reflect.ValueOf(params)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			// 尝试获取 Platform 字段
			pf := v.FieldByName("Platform")
			if pf.IsValid() && pf.Kind() == reflect.String {
				platform = pf.String()
			}
			// 尝试获取 SelfID 字段
			sf := v.FieldByName("SelfID")
			if sf.IsValid() {
				selfID = fmt.Sprintf("%v", sf.Interface())
			}
			// 尝试获取 GroupID 字段
			gf := v.FieldByName("GroupID")
			if gf.IsValid() {
				groupID = fmt.Sprintf("%v", gf.Interface())
			}
			// 尝试获取 UserID 字段
			uf := v.FieldByName("UserID")
			if uf.IsValid() {
				userID = fmt.Sprintf("%v", uf.Interface())
			}
			// 尝试获取 Message 字段
			mf := v.FieldByName("Message")
			if mf.IsValid() {
				message = fmt.Sprintf("%v", mf.Interface())
			}
		} else if v.Kind() == reflect.Map {
			// 如果是 map，尝试直接提取
			if val, ok := params.(map[string]any); ok {
				if v, ok := val["platform"]; ok {
					platform = fmt.Sprintf("%v", v)
				}
				if v, ok := val["self_id"]; ok {
					selfID = fmt.Sprintf("%v", v)
				}
				if v, ok := val["group_id"]; ok {
					groupID = fmt.Sprintf("%v", v)
				}
				if v, ok := val["user_id"]; ok {
					userID = fmt.Sprintf("%v", v)
				}
				if v, ok := val["message"]; ok {
					message = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	// 如果 params 中没有，使用最后的记录
	if selfID == "" && s.lastSelfID != 0 {
		selfID = fmt.Sprintf("%d", s.lastSelfID)
	}
	if platform == "" && s.lastPlatform != "" {
		platform = s.lastPlatform
	}

	if selfID != "" {
		msg["self_id"] = selfID
	}
	if platform != "" {
		msg["platform"] = platform
	}
	if groupID != "" {
		msg["group_id"] = groupID
	}
	if userID != "" {
		msg["user_id"] = userID
	}
	if message != "" {
		msg["reply"] = message // BotNexus 期望的字段是 reply
	}

	// 打印详细的 Payload 内容
	if payloadJson, err := json.Marshal(msg); err == nil {
		log.Printf("[Nexus] Action Payload: %s", string(payloadJson))
	}

	log.Printf("[Nexus] Sending action to Nexus: %s", action)
	s.botService.SendToNexus(msg)
	return nil
}

// Session & State Management 实现
func (s *CombinedServer) GetSessionContext(platform, userID string) (*types.SessionContext, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.GetSessionContext(platform, userID)
}

func (s *CombinedServer) SetSessionState(platform, userID string, state types.SessionState, ttl time.Duration) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return s.redisClient.SetSessionState(platform, userID, state, ttl)
}

func (s *CombinedServer) GetSessionState(platform, userID string) (*types.SessionState, error) {
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
func (s *CombinedServer) HandleSkill(skillName string, fn func(ctx core.BaseContext, params map[string]string) (string, error)) {
	s.skillsMu.Lock()
	defer s.skillsMu.Unlock()
	s.skills[skillName] = fn
}

// RegisterSkill 注册带有元数据的技能
func (s *CombinedServer) RegisterSkill(capability core.SkillCapability, fn func(ctx core.BaseContext, params map[string]string) (string, error)) {
	s.skillsMu.Lock()
	defer s.skillsMu.Unlock()
	s.skills[capability.Name] = fn

	// 检查是否已经存在同名能力，如果存在则更新，否则添加
	found := false
	for i, c := range s.skillCapabilities {
		if c.Name == capability.Name {
			s.skillCapabilities[i] = capability
			found = true
			break
		}
	}
	if !found {
		s.skillCapabilities = append(s.skillCapabilities, capability)
	}
	log.Printf("[Worker] Registered skill: %s (Regex: %s)", capability.Name, capability.Regex)
}

func (s *CombinedServer) CreateBaseContext(botUin int64, groupID int64, userID int64) core.BaseContext {
	botInfo := &core.BotInfo{
		Uin:      botUin,
		Platform: s.lastPlatform,
	}

	var group *models.Sz84Group
	var member *models.Sz84GroupMember
	var user *models.Sz84User

	if plugins.GlobalStore != nil {
		if groupID != 0 {
			group, _ = plugins.GlobalStore.Groups.GetByID(groupID)
			member, _ = plugins.GlobalStore.Members.Get(groupID, userID)
		}
		user, _ = plugins.GlobalStore.Users.GetByID(userID)
	}

	var sz84Store *models.Sz84Store
	if plugins.GlobalGORMDB != nil {
		sz84Store = models.NewSz84Store(plugins.GlobalGORMDB, nil) // Redis can be added if needed
	}

	return NewBaseContext(botInfo, group, member, user, sz84Store)
}

// routeMessageToSkill 尝试将消息路由到特定插件技能
func (s *CombinedServer) routeMessageToSkill(e *onebot.Event) (bool, error) {
	s.skillsMu.RLock()
	defer s.skillsMu.RUnlock()

	// 1. 优先从 Redis 获取动态插件路由规则 (例如：group_123 -> weather)
	if s.redisClient != nil {
		ctx := context.Background()
		matchKeys := []string{
			fmt.Sprintf("user_%v", e.UserID),
			fmt.Sprintf("group_%v", e.GroupID),
			fmt.Sprintf("bot_%v", e.SelfID),
		}

		// 规则存储在 worker:{workerID}:plugin_rules 哈希表中
		rulesKey := fmt.Sprintf("worker:%s:plugin_rules", s.config.WorkerID)
		for _, key := range matchKeys {
			if skillName, err := s.redisClient.HGet(ctx, rulesKey, key).Result(); err == nil && skillName != "" {
				if _, ok := s.skills[skillName]; ok {
					log.Printf("[Routing] Dynamic plugin route matched (Redis): %s -> %s", key, skillName)
					params := map[string]string{
						"user_id":  e.UserID.String(),
						"group_id": e.GroupID.String(),
						"message":  e.RawMessage,
						"platform": e.Platform,
						"self_id":  fmt.Sprintf("%v", e.SelfID),
					}
					baseCtx := s.CreateBaseContext(int64(e.SelfID), int64(e.GroupID), int64(e.UserID))
					_, err := s.InvokeSkill(baseCtx, skillName, params)
					if err != nil {
						return true, fmt.Errorf("failed to invoke dynamic skill %s: %v", skillName, err)
					}
					return true, nil
				}
			}
		}
	}

	// 2. 遍历已注册能力的 Regex 进行匹配
	for _, cap := range s.skillCapabilities {
		if cap.Regex == "" {
			continue
		}

		matched, err := regexp.MatchString(cap.Regex, e.RawMessage)
		if err == nil && matched {
			log.Printf("[Routing] Message matched skill: %s (Regex: %s)", cap.Name, cap.Regex)
			params := map[string]string{
				"user_id":  e.UserID.String(),
				"group_id": e.GroupID.String(),
				"message":  e.RawMessage,
				"platform": e.Platform,
				"self_id":  fmt.Sprintf("%v", e.SelfID),
			}
			baseCtx := s.CreateBaseContext(int64(e.SelfID), int64(e.GroupID), int64(e.UserID))
			_, err := s.InvokeSkill(baseCtx, cap.Name, params)
			if err != nil {
				return true, fmt.Errorf("failed to invoke skill %s: %v", cap.Name, err)
			}
			return true, nil
		}
	}

	// 2. 如果没有任何匹配，路由到 sz84 兜底
	if handler, ok := s.skills["sz84"]; ok {
		log.Printf("[Routing] No specific skill matched, falling back to sz84")
		params := map[string]string{
			"user_id":  e.UserID.String(),
			"group_id": e.GroupID.String(),
			"message":  e.RawMessage,
			"platform": e.Platform,
			"self_id":  fmt.Sprintf("%v", e.SelfID),
		}
		baseCtx := s.CreateBaseContext(int64(e.SelfID), int64(e.GroupID), int64(e.UserID))
		_, err := handler(baseCtx, params)
		if err != nil {
			return true, fmt.Errorf("failed to invoke fallback sz84 skill: %v", err)
		}
		return true, nil
	}

	return false, nil
}

// InvokeSkill 调用已注册的技能
func (s *CombinedServer) InvokeSkill(ctx core.BaseContext, skillName string, params map[string]string) (string, error) {
	s.skillsMu.RLock()
	fn, ok := s.skills[skillName]
	s.skillsMu.RUnlock()

	if !ok {
		return "", fmt.Errorf("skill %s not found", skillName)
	}

	return fn(ctx, params)
}

func (s *CombinedServer) CallPluginAction(pluginID string, action string, payload map[string]any) (any, error) {
	if s.actionRouter != nil {
		return s.actionRouter(pluginID, action, payload)
	}
	return nil, fmt.Errorf("action router not initialized")
}

func (s *CombinedServer) SetActionRouter(router func(string, string, map[string]any) (any, error)) {
	s.actionRouter = router
}

// 插件管理
func (s *CombinedServer) GetPluginManager() *core.PluginManager {
	return s.pluginManager
}

func (s *CombinedServer) GetConfig() *config.Config {
	return s.config
}

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

	// 启动Redis队列监听 (如果配置了Redis)
	if s.redisClient != nil {
		go s.startRedisQueueListener()
	}

	// 保持主运行状态
	select {}
}

func (s *CombinedServer) startRedisQueueListener() {
	if s.redisClient == nil {
		return
	}

	workerID := s.config.WorkerID
	groupName := "botmatrix:group:workers"
	consumerName := fmt.Sprintf("worker:%s", workerID)

	// 准备队列名（Streams）
	streams := []string{"botmatrix:queue:default"}
	if workerID != "" {
		streams = append(streams, fmt.Sprintf("botmatrix:queue:worker:%s", workerID))
	}

	ctx := context.Background()

	// 初始化消费组
	for _, stream := range streams {
		// 先尝试删除旧的 List Key (如果是从旧版本升级)
		typeInfo, _ := s.redisClient.Type(ctx, stream).Result()
		if typeInfo == "list" {
			log.Printf("[RedisStreams] Found old list key %s, deleting it to convert to stream", stream)
			s.redisClient.Del(ctx, stream)
		}

		// 检查 Stream 是否存在，如果不存在 XGroupCreate 会报错
		err := s.redisClient.XGroupCreateMkStream(ctx, stream, groupName, "0").Err()
		if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
			log.Printf("[RedisStreams] Error creating group for %s: %v", stream, err)
		}
	}

	log.Printf("[RedisStreams] Starting listener for streams: %v, Group: %s, Consumer: %s", streams, groupName, consumerName)

	// 构建 XReadGroup 参数：[stream1, stream2, ..., id1, id2, ...]
	// 对于新消息，ID 应该使用 ">"
	readArgs := &goredis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  make([]string, len(streams)*2),
		Count:    1,
		Block:    30 * time.Second,
	}
	for i, stream := range streams {
		readArgs.Streams[i] = stream
		readArgs.Streams[i+len(streams)] = ">"
	}

	for {
		entries, err := s.redisClient.XReadGroup(ctx, readArgs).Result()
		if err != nil {
			if err != goredis.Nil {
				log.Printf("[RedisStreams] Error reading from streams: %v", err)
				time.Sleep(5 * time.Second)
			}
			continue
		}

		for _, streamResult := range entries {
			streamName := streamResult.Stream
			for _, xmsg := range streamResult.Messages {
				payload, ok := xmsg.Values["payload"].(string)
				if !ok {
					log.Printf("[RedisStreams] Invalid message format in %s: %v", streamName, xmsg.Values)
					s.redisClient.XAck(ctx, streamName, groupName, xmsg.ID)
					continue
				}

				log.Printf("[RedisStreams] Received message from %s (ID: %s)", streamName, xmsg.ID)

				var msg map[string]any
				if err := json.Unmarshal([]byte(payload), &msg); err != nil {
					log.Printf("[RedisStreams] Failed to unmarshal message: %v", err)
					s.redisClient.XAck(ctx, streamName, groupName, xmsg.ID)
					continue
				}

				// 处理消息
				go func(stream, group, msgID string, m map[string]any) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[RedisStreams] Panic in message processor: %v", r)
						}
					}()
					s.processQueueMessage(m)
					// 处理成功后发送 ACK
					s.redisClient.XAck(ctx, stream, group, msgID)
				}(streamName, groupName, xmsg.ID, msg)
			}
		}
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

	s.skillsMu.RLock()
	handler, ok := s.skills[skillName]
	s.skillsMu.RUnlock()
	if !ok {
		log.Printf("[SkillCall] No handler for skill: %s", skillName)
		s.reportSkillResult(taskID, executionID, skillName, "", fmt.Errorf("no handler for skill: %s", skillName))
		return
	}

	// 从参数中提取上下文信息 (如果存在)
	var botUin, groupID, userID int64
	if v, ok := params["bot_uin"]; ok {
		botUin, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := params["group_id"]; ok {
		groupID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := params["user_id"]; ok {
		userID, _ = strconv.ParseInt(v, 10, 64)
	}

	baseCtx := s.CreateBaseContext(botUin, groupID, userID)
	result, err := handler(baseCtx, params)
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

	log.Printf("[SkillResult] Reporting result for %s: %s (error: %v)", skillName, result, err)

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
	// 只有非元事件才打印详细日志
	if postType != "meta_event" {
		log.Printf("[Worker] Processing queue event: post_type=%s, message_type=%s, msg=%v", postType, messageType, msg)
	}

	// 增加表情占位符转换逻辑 (处理旧版数据库中的占位符)
	if postType == "message" {
		if rawMsg, ok := msg["raw_message"].(string); ok && rawMsg != "" {
			newMsg := commononebot.ConvertLegacyPlaceholders(rawMsg)
			if newMsg != rawMsg {
				log.Printf("[Worker] Converted legacy placeholders: %s -> %s", rawMsg, newMsg)
				msg["raw_message"] = newMsg
				// 同时更新 message 字段，确保后续处理使用转换后的内容
				msg["message"] = newMsg
			}
		}
	}

	// 将 map 转换为 onebot.Event
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Worker] Failed to marshal queue message: %v", err)
		return
	}

	var event onebot.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[Worker] Failed to unmarshal queue event: %v", err)
		return
	}

	if postType != "meta_event" {
		log.Printf("[Worker] Unmarshaled Event: message=%v, raw=%s", event.Message, event.RawMessage)
	}

	// 更新最后处理的机器人 ID 和平台
	s.lastSelfID = int64(event.SelfID)
	if p, ok := msg["platform"].(string); ok {
		s.lastPlatform = p
	} else {
		// Default to qq if not provided, or keep last if same bot
		if s.lastPlatform == "" {
			s.lastPlatform = "qq"
		}
	}

	// 处理 QQGuild ID 生成（确保 ID 映射正确）
	processEventIDs(&event)

	// --- 技能路由逻辑 (New) ---
	if event.PostType == "message" && event.RawMessage != "" {
		handled, err := s.routeMessageToSkill(&event)
		if err != nil {
			log.Errorf("[Worker] routeMessageToSkill error: %v", err)
		}
		if handled {
			return
		}
	}

	// --- 智能体 (Digital Employee) 处理逻辑 ---
	// 如果该 Bot 被定义为“数字员工”，则在 Worker 端直接进行 AI 响应
	if s.employeeService != nil && s.aiService != nil && event.PostType == "message" && event.UserID.String() != fmt.Sprintf("%v", event.SelfID) {
		employee, err := s.employeeService.GetEmployeeByBotID(fmt.Sprintf("%v", event.SelfID))
		if err == nil && employee != nil {
			log.Printf("[Agent] Bot %v is a Digital Employee: %s (%s)", event.SelfID, employee.Name, employee.Title)

			// 只有文本消息才触发 AI
			if event.RawMessage != "" {
				// 调用 AI 进行数字员工响应 (带上下文历史)
				// 注意：这里需要将 onebot.Event 转换为 types.InternalMessage
				internalMsg := types.InternalMessage{
					ID:          event.MessageID.String(),
					Time:        event.Time,
					Platform:    event.Platform,
					SelfID:      fmt.Sprintf("%v", event.SelfID),
					UserID:      event.UserID.String(),
					GroupID:     event.GroupID.String(),
					MessageType: event.MessageType,
					RawMessage:  event.RawMessage,
				}

				response, err := s.aiService.ChatWithEmployee(employee, internalMsg, employee.EnterpriseID)
				if err == nil && response != "" {
					log.Printf("[Agent] AI Response for Bot %v: %s", event.SelfID, response)
					// 发送回复
					s.SendMessage(&onebot.SendMessageParams{
						MessageType: event.MessageType,
						UserID:      event.UserID,
						GroupID:     event.GroupID,
						Message:     response,
					})

					// 数字员工回复后，通常不需要再分发给插件处理通用逻辑
					return
				} else if err != nil {
					log.Printf("[Agent] AI Chat failed: %v", err)
				}
			}
		}
	}
	// ------------------------------------------

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
	case "meta_event":
		// 默认不再向插件转发元事件（如心跳），以减少噪音和系统负载
		// 如果以后有插件需要心跳，可以在这里增加白名单
		// s.wsServer.DispatchEvent(event)
	}
}

func (s *CombinedServer) Stop() {
	s.wsServer.Stop()
	s.httpServer.Stop()
}
