package main

import (
	"BotMatrix/common"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// handleBotWebSocket 处理Bot WebSocket连接
func (m *Manager) handleBotWebSocket(w http.ResponseWriter, r *http.Request) {
	// 记录Bot连接尝试的详细信息
	log.Printf("Bot WebSocket connection attempt from %s - Headers: X-Self-ID=%s, X-Platform=%s",
		r.RemoteAddr, r.Header.Get("X-Self-ID"), r.Header.Get("X-Platform"))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Bot WebSocket upgrade failed: %v", err)
		return
	}

	// 生成Bot ID
	selfID := r.Header.Get("X-Self-ID")
	platform := r.Header.Get("X-Platform")
	if platform == "" {
		platform = "qq" // 默认平台
	}

	// 如果没有提供 ID，使用连接地址作为临时ID
	if selfID == "" {
		selfID = conn.RemoteAddr().String()
	}

	// 创建Bot客户端
	bot := &common.BotClient{
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Platform:      platform,
		SelfID:        selfID,
	}

	// 注册Bot
	m.Mutex.Lock()
	if m.Bots == nil {
		m.Bots = make(map[string]*common.BotClient)
	}
	m.Bots[bot.SelfID] = bot
	m.Mutex.Unlock()

	// 更新连接统计
	m.ConnectionStats.Mutex.Lock()
	m.ConnectionStats.TotalBotConnections++
	m.ConnectionStats.LastBotActivity[bot.SelfID] = time.Now()
	m.ConnectionStats.Mutex.Unlock()

	log.Printf("Bot WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), bot.SelfID)

	// 异步获取 Bot 信息
	go m.fetchBotInfo(bot)

	// 启动连接处理循环
	go m.handleBotConnection(bot)
}

// fetchBotInfo 主动获取 Bot 的详细信息
func (m *Manager) fetchBotInfo(bot *common.BotClient) {
	// 等待一秒，确保连接完全建立且握手完成
	time.Sleep(1 * time.Second)

	log.Printf("[Bot] Fetching info for bot: %s", bot.SelfID)

	// 1. 获取登录信息 (昵称等)
	echoInfo := "fetch_info_" + bot.SelfID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	m.PendingMutex.Lock()
	m.PendingRequests[echoInfo] = make(chan map[string]interface{}, 1)
	m.PendingMutex.Unlock()

	reqInfo := map[string]interface{}{
		"action": "get_login_info",
		"params": map[string]interface{}{},
		"echo":   echoInfo,
	}

	bot.Mutex.Lock()
	bot.Conn.WriteJSON(reqInfo)
	bot.Mutex.Unlock()

	// 2. 获取群列表 (获取群数量)
	echoGroups := "fetch_groups_" + bot.SelfID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	m.PendingMutex.Lock()
	m.PendingRequests[echoGroups] = make(chan map[string]interface{}, 1)
	m.PendingMutex.Unlock()

	reqGroups := map[string]interface{}{
		"action": "get_group_list",
		"params": map[string]interface{}{},
		"echo":   echoGroups,
	}

	bot.Mutex.Lock()
	bot.Conn.WriteJSON(reqGroups)
	bot.Mutex.Unlock()

	// 等待响应 (带超时)
	timeout := time.After(10 * time.Second)
	infoDone := false
	groupsDone := false

	for !infoDone || !groupsDone {
		select {
		case resp := <-m.PendingRequests[echoInfo]:
			if data, ok := resp["data"].(map[string]interface{}); ok {
				bot.Mutex.Lock()
				var newNickname string
				var newSelfID string

				if nickname, ok := data["nickname"].(string); ok {
					newNickname = nickname
				}
				if userID, ok := data["user_id"]; ok {
					newSelfID = common.ToString(userID)
				}

				if newSelfID != "" && newSelfID != bot.SelfID {
					oldID := bot.SelfID
					bot.Mutex.Unlock() // Unlock before map operations

					m.Mutex.Lock()
					delete(m.Bots, oldID)
					bot.SelfID = newSelfID
					bot.Nickname = newNickname
					m.Bots[bot.SelfID] = bot
					m.Mutex.Unlock()

					bot.Mutex.Lock() // Re-lock
					log.Printf("[Bot] Updated Bot ID from %s to %s via get_login_info", oldID, newSelfID)
				} else {
					bot.Nickname = newNickname
				}
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated info for %s: Nickname=%s", bot.SelfID, bot.Nickname)
			}
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echoInfo)
			m.PendingMutex.Unlock()
			infoDone = true

		case resp := <-m.PendingRequests[echoGroups]:
			if data, ok := resp["data"].([]interface{}); ok {
				bot.Mutex.Lock()
				bot.GroupCount = len(data)
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated group count for %s: %d", bot.SelfID, bot.GroupCount)

				// 更新群组缓存，以便后续路由 API 请求
				m.CacheMutex.Lock()
				for _, item := range data {
					if group, ok := item.(map[string]interface{}); ok {
						var gID string
						if idVal, ok := group["group_id"]; ok {
							gID = common.ToString(idVal)
						}

						if gID != "" {
							name, _ := group["group_name"].(string)
							if name == "" {
								name = fmt.Sprintf("Group %s (Auto)", gID)
							}
							// 深度复制所有字段
							cachedGroup := make(map[string]interface{})
							for k, v := range group {
								cachedGroup[k] = v
							}
							cachedGroup["group_id"] = gID
							cachedGroup["group_name"] = name
							cachedGroup["bot_id"] = bot.SelfID
							cachedGroup["is_cached"] = true
							cachedGroup["source"] = "get_group_list"
							m.GroupCache[gID] = cachedGroup
							// 持久化到数据库
							go m.SaveGroupToDB(gID, name, bot.SelfID)
						}
					}
				}
				m.CacheMutex.Unlock()
				log.Printf("[Bot] Cached %d groups for Bot %s", len(data), bot.SelfID)
			}
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echoGroups)
			m.PendingMutex.Unlock()
			groupsDone = true

		case <-timeout:
			log.Printf("[Bot] Timeout fetching info for bot %s", bot.SelfID)
			infoDone = true
			groupsDone = true
		}
	}
}

// handleBotConnection 处理单个Bot连接的消息循环
func (m *Manager) handleBotConnection(bot *common.BotClient) {
	// 启动心跳发送协程
	stopHeartbeat := make(chan struct{})
	go m.sendBotHeartbeat(bot, stopHeartbeat)

	defer func() {
		close(stopHeartbeat) // 停止心跳
		// 连接关闭时的清理工作
		m.removeBot(bot.SelfID)
		bot.Conn.Close()

		// 记录断开连接
		duration := time.Since(bot.Connected)
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.BotConnectionDurations[bot.SelfID] = duration
		if m.ConnectionStats.BotDisconnectReasons == nil {
			m.ConnectionStats.BotDisconnectReasons = make(map[string]int64)
		}
		m.ConnectionStats.BotDisconnectReasons["connection_closed"]++
		m.ConnectionStats.Mutex.Unlock()

		log.Printf("Bot WebSocket disconnected: %s (duration: %v)", bot.SelfID, duration)
	}()

	// 设置读取超时（延长到120秒）
	bot.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	bot.Conn.SetPongHandler(func(string) error {
		bot.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		bot.LastHeartbeat = time.Now()
		log.Printf("Bot %s received pong", bot.SelfID)
		return nil
	})

	for {
		var msg map[string]interface{}
		err := common.ReadJSONWithNumber(bot.Conn, &msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Bot %s read error: %v", bot.SelfID, err)
			}
			break
		}

		// 处理消息
		m.handleBotMessage(bot, msg)

		// 更新活动时间
		bot.LastHeartbeat = time.Now()
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.LastBotActivity[bot.SelfID] = time.Now()
		m.ConnectionStats.Mutex.Unlock()
	}
}

// sendBotHeartbeat 定期发送心跳包给Bot
func (m *Manager) sendBotHeartbeat(bot *common.BotClient, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送ping帧
			bot.Mutex.Lock()
			err := bot.Conn.WriteMessage(websocket.PingMessage, []byte{})
			bot.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send ping to Bot %s: %v", bot.SelfID, err)
				return
			}
			log.Printf("Sent ping to Bot %s", bot.SelfID)

		case <-stop:
			return
		}
	}
}

// sendWorkerHeartbeat 定期发送心跳包给Worker
func (m *Manager) sendWorkerHeartbeat(worker *common.WorkerClient, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送ping帧
			worker.Mutex.Lock()
			err := worker.Conn.WriteMessage(websocket.PingMessage, []byte{})
			worker.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send ping to Worker %s: %v", worker.ID, err)
				return
			}
			log.Printf("Sent ping to Worker %s", worker.ID)

		case <-stop:
			return
		}
	}
}

// handleBotMessage 处理Bot消息
func (m *Manager) handleBotMessage(bot *common.BotClient, msg map[string]interface{}) {
	// 1. 核心插件拦截
	if allowed, reason, err := m.Core.ProcessMessage(msg); !allowed {
		log.Printf("[Core] Message blocked: %s (reason: %s)", bot.SelfID, reason)
		if err != nil {
			log.Printf("[Core] Error processing message: %v", err)
		}

		// 如果是管理员指令被系统拦截（例如系统关闭时），尝试处理管理员指令
		if reason == "system_closed" && m.Core.identifyMessageType(msg) == "admin_command" {
			// 继续往下走，让 handleBotMessageEvent 处理或直接在这里拦截
		} else {
			return
		}
	}

	// 核心插件处理管理员指令
	if m.Core.identifyMessageType(msg) == "admin_command" {
		resp, err := m.Core.HandleAdminCommand(msg)
		if err == nil && resp != "" {
			// 发送响应回机器人
			m.sendBotMessage(bot, msg, resp)
			return
		}
	}

	// 2. 检查是否包含 self_id 并更新（如果当前是临时 ID）
	msgSelfID := common.ToString(msg["self_id"])
	if msgSelfID != "" {
		if bot.SelfID != msgSelfID && strings.Contains(bot.SelfID, ":") {
			// 当前是临时 IP ID，收到正式 ID，进行更新
			oldID := bot.SelfID
			m.Mutex.Lock()
			delete(m.Bots, oldID)
			bot.SelfID = msgSelfID
			m.Bots[bot.SelfID] = bot
			m.Mutex.Unlock()
			log.Printf("[Bot] Updated Bot ID from %s to %s", oldID, msgSelfID)
		}
	}

	// 检查是否是API响应（有echo字段）
	echo := common.ToString(msg["echo"])
	if echo != "" {
		// 广播路由事件: Bot -> Nexus (Response)
		m.BroadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "response", nil)

		// 这是API响应，需要回传给对应的Worker
		m.PendingMutex.Lock()
		respChan, exists := m.PendingRequests[echo]
		sendTime, timeExists := m.PendingTimestamps[echo]
		delete(m.PendingTimestamps, echo)
		m.PendingMutex.Unlock()

		if exists {
			// 记录 RTT
			if timeExists {
				rtt := time.Since(sendTime)
				// 找到对应的 Worker 并更新 RTT
				// 使用 | 作为分隔符: workerID|originalEcho
				if parts := strings.Split(echo, "|"); len(parts) >= 2 {
					workerID := parts[0]
					m.Mutex.RLock()
					for _, w := range m.Workers {
						if w.ID == workerID {
							w.Mutex.Lock()
							w.LastRTT = rtt
							w.RTTSamples = append(w.RTTSamples, rtt)
							if len(w.RTTSamples) > 20 { // 最多保留20个样本
								w.RTTSamples = w.RTTSamples[1:]
							}
							var total time.Duration
							for _, s := range w.RTTSamples {
								total += s
							}
							w.AvgRTT = total / time.Duration(len(w.RTTSamples))
							w.Mutex.Unlock()
							log.Printf("[RTT] Worker %s AvgRTT: %v, LastRTT: %v", workerID, w.AvgRTT, w.LastRTT)
							break
						}
					}
					m.Mutex.RUnlock()

					// 广播路由事件: Nexus -> Worker (Response Forward)
					m.BroadcastRoutingEvent("Nexus", workerID, "nexus_to_worker", "response", nil)
				}
			}

			// 将响应发送给等待的Worker请求
			select {
			case respChan <- msg:
				log.Printf("Forwarded Bot %s API response (echo: %s) to Worker", bot.SelfID, echo)
			default:
				log.Printf("Failed to forward Bot %s API response (echo: %s): channel full", bot.SelfID, echo)
			}
		} else {
			log.Printf("Received Bot %s API response (echo: %s) but no pending request found", bot.SelfID, echo)
		}
		return
	}

	// 原有的消息处理逻辑
	// 获取消息类型
	msgType := common.ToString(msg["post_type"])
	if msgType == "" {
		log.Printf("[Bot] Warning: Received message from Bot %s without echo or post_type: %v", bot.SelfID, msg)
		return
	}

	log.Printf("Received %s message from Bot %s", msgType, bot.SelfID)

	// 更新统计信息
	m.Mutex.Lock()
	bot.RecvCount++
	m.Mutex.Unlock()

	// 根据消息类型处理
	switch msgType {
	case "meta_event":
		// 元事件（心跳等）
		if heartbeat, ok := msg["meta_event_type"].(string); ok && heartbeat == "heartbeat" {
			// 心跳事件，更新状态
			if status, ok := msg["status"].(map[string]interface{}); ok {
				if online, ok := status["online"].(bool); ok {
					log.Printf("Bot %s heartbeat: online=%v", bot.SelfID, online)
				}
			}
		}
	case "message":
		// 消息事件
		messageType, _ := msg["message_type"].(string)
		log.Printf("Message type: %s", messageType)
		m.handleBotMessageEvent(bot, msg)
	}

	// 转发给Worker处理
	// 确保消息格式符合 OneBot 11 标准 (特别是 ID 类型)
	if userID, ok := msg["user_id"]; ok {
		msg["user_id"] = common.ToString(userID)
	}
	if groupID, ok := msg["group_id"]; ok {
		msg["group_id"] = common.ToString(groupID)
	}

	// 补全并标准化 self_id (OneBot 11 标准应为 int64，但我们兼容字符串)
	if selfID, ok := msg["self_id"]; ok {
		msg["self_id"] = common.ToString(selfID)
	} else {
		// 补全 self_id
		msg["self_id"] = bot.SelfID
	}

	// 补全 time (int64 unix timestamp)
	if _, ok := msg["time"]; !ok {
		msg["time"] = time.Now().Unix()
	}

	// 补全 post_type 如果缺失
	if _, ok := msg["post_type"]; !ok {
		msg["post_type"] = "message"
	}

	// 补全 platform
	if _, ok := msg["platform"]; !ok {
		msg["platform"] = bot.Platform
	}

	m.forwardMessageToWorker(msg)

	// 缓存群/成员/好友信息 (基于消息)
	m.cacheBotDataFromMessage(bot, msg)
}

// cacheBotDataFromMessage 从消息中提取并缓存数据 (特别针对腾讯频道机器人)
func (m *Manager) cacheBotDataFromMessage(bot *common.BotClient, msg map[string]interface{}) {
	postType, _ := msg["post_type"].(string)
	if postType != "message" {
		return
	}

	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// 缓存群信息
	if groupIDVal, ok := msg["group_id"]; ok && groupIDVal != nil {
		gID := common.ToString(groupIDVal)

		// 如果缓存中已存在该群组，则保留原有名称，除非原有名称包含 "(Cached)" 而当前消息提供了更准确的信息
		existingGroup, exists := m.GroupCache[gID]
		groupName := ""
		if exists {
			if name, ok := existingGroup["group_name"].(string); ok {
				groupName = name
			}
		}

		// 如果没有名称或者名称包含 "(Cached)"，尝试设置一个默认值
		if groupName == "" || strings.Contains(groupName, "(Cached)") {
			// 只有在完全没有名称时才设置 (Cached) 格式
			if groupName == "" {
				groupName = "Group " + gID
			}
		}

		// 更新或添加群组缓存
		m.GroupCache[gID] = map[string]interface{}{
			"group_id":   gID,
			"group_name": groupName,
			"bot_id":     bot.SelfID,
			"is_cached":  true,
			"reason":     "Automatically updated from message",
			"last_seen":  time.Now(),
		}
		// 持久化到数据库
		go m.SaveGroupToDB(gID, groupName, bot.SelfID)

		// 缓存成员信息
		if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
			uID := common.ToString(userIDVal)
			key := fmt.Sprintf("%s:%s", gID, uID)
			sender, _ := msg["sender"].(map[string]interface{})
			nickname := ""
			card := ""
			if sender != nil {
				nickname, _ = sender["nickname"].(string)
				card, _ = sender["card"].(string)
			}
			m.MemberCache[key] = map[string]interface{}{
				"group_id":  gID,
				"user_id":   uID,
				"nickname":  nickname,
				"card":      card,
				"is_cached": true,
			}
			// 持久化到数据库
			go m.SaveMemberToDB(gID, uID, nickname, card)
		}
	} else if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
		// 缓存好友信息 (私聊)
		uID := common.ToString(userIDVal)
		if _, exists := m.FriendCache[uID]; !exists {
			sender, _ := msg["sender"].(map[string]interface{})
			nickname := ""
			if sender != nil {
				nickname, _ = sender["nickname"].(string)
			}
			m.FriendCache[uID] = map[string]interface{}{
				"user_id":   uID,
				"nickname":  nickname,
				"is_cached": true,
			}
			// 持久化到数据库
			go m.SaveFriendToDB(uID, nickname)
		}
	}
}

// isSystemMessage 检查是否为系统消息或无意义的统计数据
func isSystemMessage(msg map[string]interface{}) bool {
	// 1. 提取基本字段，支持直接结构和 params 嵌套结构
	userID := common.ToString(msg["user_id"])
	msgType := common.ToString(msg["message_type"])
	subType := common.ToString(msg["sub_type"])
	message := ""

	// 尝试从 params 中提取 (如果是发送请求)
	if params, ok := msg["params"].(map[string]interface{}); ok {
		if userID == "" {
			userID = common.ToString(params["user_id"])
		}
		if msgType == "" {
			msgType = common.ToString(params["message_type"])
		}
		if m, ok := params["message"].(string); ok {
			message = m
		}
	}

	// 提取内容
	if message == "" {
		if rm, ok := msg["raw_message"].(string); ok && rm != "" {
			message = rm
		} else if m, ok := msg["message"].(string); ok && m != "" {
			message = m
		}
	}

	// 2. 检查常见的系统用户 ID
	systemIDs := map[string]bool{
		"10000":      true, // 系统消息
		"1000000":    true, // 群系统消息
		"1000001":    true, // 好友系统消息
		"1000002":    true, // 运营消息
		"80000000":   true, // 匿名消息
		"2852199017": true, // QQ官方
	}
	if systemIDs[userID] {
		return true
	}

	// 3. 检查消息类型和子类型
	if msgType == "system" || subType == "system" {
		return true
	}

	// 4. 检查内容是否为空
	if message == "" {
		return true
	}

	return false
}

// sendBotMessage 发送文本消息回机器人 (辅助方法)
func (m *Manager) sendBotMessage(bot *common.BotClient, originalMsg map[string]interface{}, text string) {
	resp := map[string]interface{}{
		"action": "send_msg",
		"params": map[string]interface{}{
			"message": text,
		},
	}

	// 复制原始消息的关键字段以确定发送目标
	if groupID, ok := originalMsg["group_id"]; ok {
		resp["params"].(map[string]interface{})["group_id"] = groupID
	} else if userID, ok := originalMsg["user_id"]; ok {
		resp["params"].(map[string]interface{})["user_id"] = userID
	}

	bot.Mutex.Lock()
	err := bot.Conn.WriteJSON(resp)
	bot.Mutex.Unlock()

	if err != nil {
		log.Printf("[Bot] Failed to send message to Bot %s: %v", bot.SelfID, err)
	}
}

// handleBotMessageEvent 处理Bot消息事件
func (m *Manager) handleBotMessageEvent(bot *common.BotClient, msg map[string]interface{}) {
	// 提取消息信息
	userID := common.ToString(msg["user_id"])
	groupID := common.ToString(msg["group_id"])

	// 优先使用 raw_message，因为它通常包含更完整的文本内容
	message := ""
	if rm, ok := msg["raw_message"].(string); ok && rm != "" {
		message = rm
	} else if m, ok := msg["message"].(string); ok && m != "" {
		message = m
	}

	// 准备广播扩展信息
	extras := map[string]interface{}{
		"user_id":  userID,
		"content":  message,
		"platform": bot.Platform,
	}

	// 提取发送者信息 (昵称和头像)
	if sender, ok := msg["sender"].(map[string]interface{}); ok {
		if nickname, ok := sender["nickname"].(string); ok && nickname != "" {
			extras["user_name"] = nickname
		}

		// 平台特定的头像逻辑
		switch strings.ToUpper(bot.Platform) {
		case "QQ":
			if userID != "" {
				extras["user_avatar"] = fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", userID)
			}
		case "WECHAT", "WX":
			// 微信头像通常无法直接通过 ID 拼接，这里先使用默认或从 sender 中获取
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			} else {
				extras["user_avatar"] = "/static/avatars/wechat_default.png"
			}
		case "TENCENT":
			// 腾讯频道头像通常在 sender 中直接提供
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			}
		default:
			// 其他平台尝试从 sender 中获取头像
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			}
		}
	}

	if groupID != "" {
		extras["group_id"] = groupID
		// 尝试从缓存中获取群名
		m.CacheMutex.RLock()
		if group, ok := m.GroupCache[groupID]; ok {
			if name, ok := group["group_name"].(string); ok {
				extras["group_name"] = name
			}
		}
		m.CacheMutex.RUnlock()
	}

	// 1. 广播路由事件: User -> Bot (如果用户 ID 存在)
	if userID != "" {
		m.BroadcastRoutingEvent(userID, bot.SelfID, "user_to_bot", "message", extras)
	}

	// 2. 广播路由事件: Bot -> Nexus
	m.BroadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "message", extras)

	// 更新详细统计 (排除系统消息)
	if !isSystemMessage(msg) {
		m.UpdateBotStats(bot.SelfID, userID, groupID)
	}
}

// removeBot 移除Bot连接
func (m *Manager) removeBot(botID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if _, exists := m.Bots[botID]; exists {
		delete(m.Bots, botID)
		log.Printf("Removed Bot %s from active connections", botID)
	}
}

// cacheMessage 缓存无法立即处理的消息
func (m *Manager) cacheMessage(msg map[string]interface{}) {
	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// 限制缓存大小，防止内存溢出
	if len(m.MessageCache) > 1000 {
		m.MessageCache = m.MessageCache[1:] // 丢弃最旧的消息
	}
	m.MessageCache = append(m.MessageCache, msg)
	log.Printf("[CACHE] No workers available, message cached (Total: %d)", len(m.MessageCache))
}

// flushMessageCache 当有新 Worker 连接时，发送缓存的消息
func (m *Manager) flushMessageCache() {
	m.Mutex.RLock()
	workerCount := len(m.Workers)
	m.Mutex.RUnlock()

	if workerCount == 0 {
		return
	}

	m.CacheMutex.Lock()
	if len(m.MessageCache) == 0 {
		m.CacheMutex.Unlock()
		return
	}

	cache := m.MessageCache
	m.MessageCache = nil
	m.CacheMutex.Unlock()

	log.Printf("[CACHE] Flushing %d cached messages to workers", len(cache))
	for _, msg := range cache {
		// 重新通过路由发送
		go m.forwardMessageToWorker(msg)
	}
}

// matchRoutePattern 检查字符串是否匹配模式 (支持 * 通配符)
func matchRoutePattern(pattern, value string) bool {
	return common.MatchRoutePattern(pattern, value)
}

// findWorkerByID 根据ID查找Worker
func (m *Manager) findWorkerByID(workerID string) *common.WorkerClient {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	for _, w := range m.Workers {
		if w.ID == workerID {
			return w
		}
	}
	return nil
}

// forwardMessageToWorker 将消息转发给Worker处理
func (m *Manager) forwardMessageToWorker(msg map[string]interface{}) {
	m.forwardMessageToWorkerWithRetry(msg, 0)
}

// forwardMessageToWorkerWithRetry 带重试限制的消息转发
func (m *Manager) forwardMessageToWorkerWithRetry(msg map[string]interface{}, retryCount int) {
	if retryCount > 3 {
		log.Printf("[ROUTING] [ERROR] Maximum retry count exceeded for message. Caching message.")
		m.cacheMessage(msg)
		return
	}

	// 1. 尝试使用路由规则
	var targetWorkerID string
	var matchKeys []string

	// 提取匹配键
	// 用户 ID 匹配: user_123456
	if userID, ok := msg["user_id"]; ok {
		sID := common.ToString(userID)
		if sID != "0" && sID != "" {
			uID := fmt.Sprintf("user_%s", sID)
			matchKeys = append(matchKeys, uID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// 群组 ID 匹配: group_789012
	if groupID, ok := msg["group_id"]; ok {
		sID := common.ToString(groupID)
		if sID != "0" && sID != "" {
			gID := fmt.Sprintf("group_%s", sID)
			matchKeys = append(matchKeys, gID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// 机器人 ID 匹配: bot_123 or self_id
	if selfID, ok := msg["self_id"]; ok {
		sID := common.ToString(selfID)
		if sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", sID))
			matchKeys = append(matchKeys, sID)
		}
	}

	// 查找匹配规则
	m.Mutex.RLock()
	var matchedKey string

	// A. 精确匹配优先
	for _, key := range matchKeys {
		if wID, exists := m.RoutingRules[key]; exists && wID != "" {
			targetWorkerID = wID
			matchedKey = key
			break
		}
	}

	// B. 如果没有精确匹配，尝试通配符匹配
	if targetWorkerID == "" {
		// 由于 RoutingRules 是 map，我们无法直接排序，
		// 但 common.MatchRoutePattern 已经处理了通配符逻辑。
		// 为了性能和简化，我们直接遍历
		for p, w := range m.RoutingRules {
			if strings.Contains(p, "*") {
				for _, key := range matchKeys {
					if common.MatchRoutePattern(p, key) {
						targetWorkerID = w
						matchedKey = fmt.Sprintf("%s (via pattern %s)", key, p)
						break
					}
				}
			}
			if targetWorkerID != "" {
				break
			}
		}
	}
	m.Mutex.RUnlock()

	// 确保消息有一个唯一的 echo，用于追踪处理耗时和回复匹配
	if _, ok := msg["echo"]; !ok {
		msg["echo"] = fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	}

	// 如果找到了目标 Worker，尝试获取它
	if targetWorkerID != "" {
		if w := m.findWorkerByID(targetWorkerID); w != nil {
			log.Printf("[ROUTING] Rule Matched: %s -> Target Worker: %s", matchedKey, targetWorkerID)
			w.Mutex.Lock()
			err := w.Conn.WriteJSON(msg)
			w.Mutex.Unlock()

			if err == nil {
				m.Mutex.Lock()
				w.HandledCount++
				m.Mutex.Unlock()

				// 广播路由事件: Nexus -> Worker (Message Forward - Rule Match)
				m.BroadcastRoutingEvent("Nexus", targetWorkerID, "nexus_to_worker", "message", nil)
				return
			}
			log.Printf("[ROUTING] [ERROR] Failed to send to target worker %s: %v. Falling back to load balancer.", targetWorkerID, err)
		} else {
			log.Printf("[ROUTING] [WARNING] Target worker %s defined in rule (%s) is OFFLINE or NOT FOUND. Falling back to load balancer.", targetWorkerID, matchedKey)
		}
	}

	// 2. 负载均衡转发
	m.Mutex.RLock()
	// 过滤出健康的 worker
	var healthyWorkers []*common.WorkerClient
	for _, w := range m.Workers {
		// 简单检查：如果 60 秒没有心跳且不是刚连接，视为不健康
		if time.Since(w.LastHeartbeat) < 60*time.Second || time.Since(w.Connected) < 10*time.Second {
			healthyWorkers = append(healthyWorkers, w)
		}
	}
	m.Mutex.RUnlock()

	if len(healthyWorkers) == 0 {
		log.Printf("[ROUTING] [WARNING] No healthy workers available. Caching message.")
		m.cacheMessage(msg)
		return
	}

	var selectedWorker *common.WorkerClient
	m.Mutex.Lock()

	// 负载均衡算法
	if len(healthyWorkers) == 1 {
		selectedWorker = healthyWorkers[0]
	} else {
		// 1. 优先选择从未处理过消息的 Worker (以便获取 RTT)
		var unhandled []*common.WorkerClient
		for _, w := range healthyWorkers {
			if w.HandledCount == 0 {
				unhandled = append(unhandled, w)
			}
		}

		if len(unhandled) > 0 {
			idx := m.WorkerIndex % len(unhandled)
			selectedWorker = unhandled[idx]
			m.WorkerIndex++
		} else {
			// 2. 优先选择 AvgProcessTime 最小的 Worker (负载最低)
			var minProcessTime time.Duration = -1
			for _, w := range healthyWorkers {
				if w.AvgProcessTime > 0 {
					if minProcessTime == -1 || w.AvgProcessTime < minProcessTime {
						minProcessTime = w.AvgProcessTime
						selectedWorker = w
					}
				}
			}

			// 3. 如果 AvgProcessTime 都没数据，或者没选出，选择 AvgRTT 最小的 Worker
			if selectedWorker == nil {
				var minRTT time.Duration = -1
				for _, w := range healthyWorkers {
					if w.AvgRTT > 0 {
						if minRTT == -1 || w.AvgRTT < minRTT {
							minRTT = w.AvgRTT
							selectedWorker = w
						}
					}
				}
			}

			// 4. 退回到全局轮询
			if selectedWorker == nil {
				idx := m.WorkerIndex % len(healthyWorkers)
				selectedWorker = healthyWorkers[idx]
				m.WorkerIndex++
			}
		}
	}
	m.Mutex.Unlock()

	// 发送消息给选中的Worker
	selectedWorker.Mutex.Lock()
	err := selectedWorker.Conn.WriteJSON(msg)
	selectedWorker.Mutex.Unlock()

	if err == nil {
		// 记录发送给 Worker 的时间，用于计算处理耗时
		echo := common.ToString(msg["echo"])
		if echo != "" {
			m.WorkerRequestMutex.Lock()
			m.WorkerRequestTimes[echo] = time.Now()
			m.WorkerRequestMutex.Unlock()
		}
	}

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward to worker %s: %v. Removing and retrying...", selectedWorker.ID, err)
		m.removeWorker(selectedWorker.ID)
		m.forwardMessageToWorkerWithRetry(msg, retryCount+1)
	} else {
		m.Mutex.Lock()
		selectedWorker.HandledCount++
		m.Mutex.Unlock()

		// 广播路由事件: Nexus -> Worker (Message Forward - LB)
		m.BroadcastRoutingEvent("Nexus", selectedWorker.ID, "nexus_to_worker", "message", nil)

		log.Printf("[ROUTING] Forwarded to worker %s (AvgRTT: %v, Handled: %d)", selectedWorker.ID, selectedWorker.AvgRTT, selectedWorker.HandledCount)
	}
}

// handleWorkerWebSocket 处理Worker WebSocket连接
func (m *Manager) handleWorkerWebSocket(w http.ResponseWriter, r *http.Request) {
	// 记录Worker连接尝试的详细信息
	log.Printf("Worker WebSocket connection attempt from %s - Headers: X-Self-ID=%s, X-Platform=%s",
		r.RemoteAddr, r.Header.Get("X-Self-ID"), r.Header.Get("X-Platform"))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Worker WebSocket upgrade failed: %v", err)
		return
	}

	// 生成Worker ID
	workerID := conn.RemoteAddr().String()

	// 创建Worker客户端 - 明确标识为Worker连接
	worker := &common.WorkerClient{
		ID:            workerID,
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
	}

	log.Printf("Worker client created successfully: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// 注册Worker
	m.Mutex.Lock()
	m.Workers = append(m.Workers, worker)
	m.Mutex.Unlock()

	// 更新连接统计
	m.ConnectionStats.Mutex.Lock()
	m.ConnectionStats.TotalWorkerConnections++
	if m.ConnectionStats.LastWorkerActivity == nil {
		m.ConnectionStats.LastWorkerActivity = make(map[string]time.Time)
	}
	m.ConnectionStats.LastWorkerActivity[workerID] = time.Now()
	m.ConnectionStats.Mutex.Unlock()

	log.Printf("Worker WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// 尝试发送缓存的消息
	go m.flushMessageCache()

	// 启动心跳包循环
	stopChan := make(chan struct{})
	go m.sendWorkerHeartbeat(worker, stopChan)

	// 启动连接处理循环
	go func() {
		m.handleWorkerConnection(worker)
		close(stopChan) // 当连接处理结束时停止心跳
	}()
}

// handleWorkerConnection 处理单个Worker连接的消息循环
func (m *Manager) handleWorkerConnection(worker *common.WorkerClient) {
	// 启动心跳协程
	stopHeartbeat := make(chan struct{})
	go m.sendWorkerHeartbeat(worker, stopHeartbeat)

	defer func() {
		close(stopHeartbeat)
		// 连接关闭时的清理工作
		m.removeWorker(worker.ID)
		worker.Conn.Close()

		// 记录断开连接
		duration := time.Since(worker.Connected)
		m.ConnectionStats.Mutex.Lock()
		if m.ConnectionStats.WorkerConnectionDurations == nil {
			m.ConnectionStats.WorkerConnectionDurations = make(map[string]time.Duration)
		}
		m.ConnectionStats.WorkerConnectionDurations[worker.ID] = duration
		if m.ConnectionStats.WorkerDisconnectReasons == nil {
			m.ConnectionStats.WorkerDisconnectReasons = make(map[string]int64)
		}
		m.ConnectionStats.WorkerDisconnectReasons["connection_closed"]++
		m.ConnectionStats.Mutex.Unlock()

		log.Printf("Worker WebSocket disconnected: %s (duration: %v)", worker.ID, duration)
	}()

	// 设置读取超时
	worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	worker.Conn.SetPongHandler(func(string) error {
		worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		worker.LastHeartbeat = time.Now()
		return nil
	})

	for {
		var msg map[string]interface{}
		err := common.ReadJSONWithNumber(worker.Conn, &msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Worker %s read error: %v", worker.ID, err)
			}
			break
		}

		// 处理Worker响应
		m.handleWorkerMessage(worker, msg)

		// 更新活动时间
		worker.LastHeartbeat = time.Now()
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.LastWorkerActivity[worker.ID] = time.Now()
		m.ConnectionStats.Mutex.Unlock()
	}
}

// handleWorkerMessage 处理Worker消息
func (m *Manager) handleWorkerMessage(worker *common.WorkerClient, msg map[string]interface{}) {
	// 只记录关键信息，不打印完整消息
	msgType := common.ToString(msg["type"])
	action := common.ToString(msg["action"])
	echo := common.ToString(msg["echo"])

	if action != "" || (echo != "" && msgType == "") {
		// 这是一个Worker发起的API请求，需要转发给Bot
		log.Printf("Worker %s API request: action=%s, echo=%s", worker.ID, action, echo)

		// 广播路由事件: Worker -> Nexus (Request)
		m.BroadcastRoutingEvent(worker.ID, "Nexus", "worker_to_nexus", "request", nil)

		m.forwardWorkerRequestToBot(worker, msg, echo)
	} else {
		log.Printf("Worker %s event/response: type=%s", worker.ID, msgType)

		// 更新统计信息
		m.Mutex.Lock()
		worker.HandledCount++
		m.Mutex.Unlock()

		// 1. 统计处理耗时
		if echo != "" {
			m.WorkerRequestMutex.Lock()
			if startTime, exists := m.WorkerRequestTimes[echo]; exists {
				duration := time.Since(startTime)
				delete(m.WorkerRequestTimes, echo)
				m.WorkerRequestMutex.Unlock()

				worker.Mutex.Lock()
				worker.LastProcessTime = duration
				worker.ProcessTimeSamples = append(worker.ProcessTimeSamples, duration)
				if len(worker.ProcessTimeSamples) > 20 {
					worker.ProcessTimeSamples = worker.ProcessTimeSamples[1:]
				}
				var total time.Duration
				for _, s := range worker.ProcessTimeSamples {
					total += s
				}
				worker.AvgProcessTime = total / time.Duration(len(worker.ProcessTimeSamples))
				worker.Mutex.Unlock()
				log.Printf("[METRIC] Worker %s processed message in %v (Avg: %v)", worker.ID, duration, worker.AvgProcessTime)
			} else {
				m.WorkerRequestMutex.Unlock()
			}
		}

		// 2. 检查是否包含回复内容 (一些框架允许在事件响应中直接返回回复)
		reply := common.ToString(msg["reply"])
		if reply != "" {
			log.Printf("Worker %s sent passive reply: %s", worker.ID, reply)
			// 构造一个 send_msg 请求转发给 Bot
			m.handleWorkerPassiveReply(worker, msg)
		}
	}
}

// handleWorkerPassiveReply 处理Worker的被动回复
func (m *Manager) handleWorkerPassiveReply(worker *common.WorkerClient, msg map[string]interface{}) {
	// 提取 echo (如果 Worker 在被动回复中带了 echo)
	echo := common.ToString(msg["echo"])

	// 构造 OneBot 11 的 send_msg 请求
	params := make(map[string]interface{})
	action := map[string]interface{}{
		"action": "send_msg",
		"params": params,
	}

	// 遍历并转发所有 Worker 返回的字段
	for k, v := range msg {
		switch k {
		case "reply":
			params["message"] = v
		case "action", "echo":
			// 忽略，action 已固定为 send_msg，echo 另外提取
			continue
		case "self_id", "platform":
			// 路由关键字段，放在顶层也放在 params
			action[k] = v
			params[k] = v
		default:
			// 其他字段透传到 params (如 group_id, user_id, message_type, auto_escape 等)
			params[k] = v
		}
	}

	// 转发给 Bot
	m.forwardWorkerRequestToBot(worker, action, echo)
}

// removeWorker 移除Worker连接
func (m *Manager) removeWorker(workerID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	// 从Workers数组中移除
	newWorkers := make([]*common.WorkerClient, 0, len(m.Workers))
	for _, w := range m.Workers {
		if w.ID != workerID {
			newWorkers = append(newWorkers, w)
		}
	}
	m.Workers = newWorkers

	log.Printf("Removed Worker %s from active connections", workerID)
}

// forwardWorkerRequestToBot 将Worker请求转发给Bot
func (m *Manager) forwardWorkerRequestToBot(worker *common.WorkerClient, msg map[string]interface{}, originalEcho string) {
	// 构造内部 echo，包含 worker ID 以便追踪和记录 RTT
	// 加上时间戳确保即使 originalEcho 为空或重复，internalEcho 也是唯一的
	internalEcho := fmt.Sprintf("%s|%s|%d", worker.ID, originalEcho, time.Now().UnixNano())

	// 保存请求映射
	respChan := make(chan map[string]interface{}, 1)
	m.PendingMutex.Lock()
	m.PendingRequests[internalEcho] = respChan
	m.PendingTimestamps[internalEcho] = time.Now() // 记录发送时间
	m.PendingMutex.Unlock()

	// 修改消息中的 echo 为内部 echo
	msg["echo"] = internalEcho

	// 智能路由逻辑：根据 self_id 或 group_id 选择正确的 Bot
	var targetBot *common.BotClient
	var routeSource string

	// 1. 尝试从请求参数中提取 self_id
	var selfID string
	if sid, ok := msg["self_id"]; ok {
		selfID = common.ToString(sid)
	} else if params, ok := msg["params"].(map[string]interface{}); ok {
		if sid, ok := params["self_id"]; ok {
			selfID = common.ToString(sid)
		}
	}

	m.Mutex.RLock()
	if selfID != "" {
		if bot, exists := m.Bots[selfID]; exists {
			targetBot = bot
			routeSource = "self_id"
		}
	}

	// 2. 如果没有 self_id，尝试根据 group_id 从缓存中查找对应的 Bot
	if targetBot == nil {
		var groupID string
		if gid, ok := msg["group_id"]; ok {
			groupID = common.ToString(gid)
		} else if params, ok := msg["params"].(map[string]interface{}); ok {
			if gid, ok := params["group_id"]; ok {
				groupID = common.ToString(gid)
			}
		}

		if groupID != "" {
			m.CacheMutex.RLock()
			if groupData, exists := m.GroupCache[groupID]; exists {
				if botID, ok := groupData["bot_id"].(string); ok {
					if bot, exists := m.Bots[botID]; exists {
						targetBot = bot
						routeSource = fmt.Sprintf("group_id cache (%s)", groupID)
					}
				}
			}
			m.CacheMutex.RUnlock()
		}
	}

	// 3. 兜底方案：如果还是找不到，选第一个可用的 Bot
	if targetBot == nil {
		for _, bot := range m.Bots {
			targetBot = bot
			routeSource = "fallback (first available)"
			break
		}
	}
	m.Mutex.RUnlock()

	if targetBot == nil {
		log.Printf("[ROUTING] [ERROR] No available Bot to handle Worker %s request (echo: %s, type: %v)", worker.ID, originalEcho, msg["action"])

		// 返回错误响应给Worker
		response := map[string]interface{}{
			"status":  "failed",
			"retcode": 1404,
			"msg":     "No Bot available",
			"echo":    originalEcho,
			"data":    nil,
		}

		worker.Mutex.Lock()
		worker.Conn.WriteJSON(response)
		worker.Mutex.Unlock()

		// 清理映射
		m.PendingMutex.Lock()
		delete(m.PendingRequests, internalEcho)
		delete(m.PendingTimestamps, internalEcho)
		m.PendingMutex.Unlock()
		return
	}

	// 转发请求给Bot
	targetBot.Mutex.Lock()
	err := targetBot.Conn.WriteJSON(msg)
	targetBot.Mutex.Unlock()

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward Worker %s request to Bot %s: %v. Attempting fallback...", worker.ID, targetBot.SelfID, err)

		// 移除失效的 Bot 并尝试另一个
		m.removeBot(targetBot.SelfID)

		// 简单的重试逻辑：找另一个可用的 Bot
		var fallbackBot *common.BotClient
		m.Mutex.RLock()
		for _, bot := range m.Bots {
			fallbackBot = bot
			break
		}
		m.Mutex.RUnlock()

		if fallbackBot != nil {
			log.Printf("[ROUTING] Falling back to Bot %s", fallbackBot.SelfID)
			fallbackBot.Mutex.Lock()
			err = fallbackBot.Conn.WriteJSON(msg)
			fallbackBot.Mutex.Unlock()
			if err == nil {
				targetBot = fallbackBot
			}
		}

		if targetBot == nil || err != nil {
			// 返回错误响应给Worker
			response := map[string]interface{}{
				"status":  "failed",
				"retcode": 1400,
				"msg":     "Failed to forward to any Bot",
				"echo":    originalEcho,
				"data":    nil,
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(response)
			worker.Mutex.Unlock()

			// 清理映射
			m.PendingMutex.Lock()
			delete(m.PendingRequests, internalEcho)
			delete(m.PendingTimestamps, internalEcho)
			m.PendingMutex.Unlock()
			return
		}
	}

	// 成功发送到 Bot，开始处理响应
	// 广播路由事件: Nexus -> Bot (Request Forward)
	extras := map[string]interface{}{
		"platform": targetBot.Platform,
	}
	// 尝试从消息中提取内容用于展示
	if params, ok := msg["params"].(map[string]interface{}); ok {
		if content, ok := params["message"]; ok {
			extras["content"] = common.ToString(content)
		}
		if userID, ok := params["user_id"]; ok {
			extras["user_id"] = common.ToString(userID)
		}
	}

	m.BroadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "request", extras)

	// 更新发送统计 (如果是发送消息类操作)
	action := common.ToString(msg["action"])
	if action == "send_msg" || action == "send_private_msg" || action == "send_group_msg" || action == "send_guild_channel_msg" {
		// 增加排除系统消息的判断
		if !isSystemMessage(msg) {
			m.UpdateBotSentStats(targetBot.SelfID)
		}
	}

	// 如果是发送消息，额外广播一个从 Bot 到 User 的事件，让特效闭环
	if userID, ok := extras["user_id"].(string); ok && userID != "" {
		m.BroadcastRoutingEvent(targetBot.SelfID, userID, "bot_to_user", "message", extras)
	}

	log.Printf("[ROUTING] Forwarded Worker %s request to Bot %s (Source: %s)", worker.ID, targetBot.SelfID, routeSource)

	// 启动超时处理（30秒内必须收到响应）
	go func() {
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		select {
		case response := <-respChan:
			// 收到响应，还原原始 echo 并转发给 Worker
			if response != nil {
				// 检查响应状态，如果发现 Bot 不在群组中，清除缓存
				retcode := common.ToInt64(response["retcode"])
				status, _ := response["status"].(string)

				if retcode == 1200 || status == "failed" {
					// 尝试提取 group_id
					var groupID string
					if gid, ok := msg["group_id"]; ok {
						groupID = common.ToString(gid)
					} else if params, ok := msg["params"].(map[string]interface{}); ok {
						if gid, ok := params["group_id"]; ok {
							groupID = common.ToString(gid)
						}
					}

					if groupID != "" {
						// 检查错误信息是否包含“不在群”或类似提示
						respMsg, _ := response["msg"].(string)
						if strings.Contains(respMsg, "不在群") || strings.Contains(respMsg, "移出") || retcode == 1200 {
							log.Printf("[ROUTING] Bot %s reported group error for %s: %s (retcode: %d). Clearing cache.",
								targetBot.SelfID, groupID, respMsg, retcode)

							m.CacheMutex.Lock()
							if data, exists := m.GroupCache[groupID]; exists {
								if cachedBotID, _ := data["bot_id"].(string); cachedBotID == targetBot.SelfID {
									delete(m.GroupCache, groupID)
									log.Printf("[ROUTING] Removed stale group cache for %s (Bot: %s)", groupID, targetBot.SelfID)
								}
							}
							m.CacheMutex.Unlock()

							// 异步触发刷新该 Bot 的信息和群列表
							go m.fetchBotInfo(targetBot)
						}
					}
				}

				response["echo"] = originalEcho
				worker.Mutex.Lock()
				worker.Conn.WriteJSON(response)
				worker.Mutex.Unlock()
				log.Printf("Forwarded Bot response (echo: %s) to Worker %s", originalEcho, worker.ID)
			}

		case <-timeout.C:
			// 超时，返回错误响应
			log.Printf("Worker request (echo: %s) timed out", originalEcho)

			response := map[string]interface{}{
				"status":  "failed",
				"retcode": 1401,
				"msg":     "Request timeout",
				"echo":    originalEcho,
				"data":    nil,
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(response)
			worker.Mutex.Unlock()
		}

		// 清理映射
		m.PendingMutex.Lock()
		delete(m.PendingRequests, internalEcho)
		delete(m.PendingTimestamps, internalEcho)
		m.PendingMutex.Unlock()
	}()
}

// cleanupPendingRequests 清理过期的请求映射
func (m *Manager) cleanupPendingRequests() {
	m.PendingMutex.Lock()
	defer m.PendingMutex.Unlock()

	// 清理所有pending请求（通常在系统关闭时调用）
	for echo, ch := range m.PendingRequests {
		close(ch)
		delete(m.PendingRequests, echo)
		log.Printf("Cleaned up pending request: %s", echo)
	}
}

// BroadcastRoutingEvent 广播路由事件到所有订阅者
func (m *Manager) BroadcastRoutingEvent(source, target, direction, msgType string, extras map[string]interface{}) {
	event := map[string]interface{}{
		"type":      "routing_event",
		"source":    source,
		"target":    target,
		"direction": direction,
		"msg_type":  msgType,
		"timestamp": time.Now().Unix(),
	}
	if extras != nil {
		for k, v := range extras {
			event[k] = v
		}
	}

	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for conn := range m.Subscribers {
		err := conn.WriteJSON(event)
		if err != nil {
			log.Printf("Failed to send routing event to subscriber: %v", err)
		}
	}
}
