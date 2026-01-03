package app

import (
	"BotMatrix/common/config"
	"BotMatrix/common/log"
	"BotMatrix/common/onebot"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"BotNexus/tasks"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin
	},
}

// handleBotWebSocket handles Bot WebSocket connections
func (m *Manager) handleBotWebSocket(w http.ResponseWriter, r *http.Request) {
	// Record detailed info of Bot connection attempt
	log.Info("Bot WebSocket connection attempt",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("self_id", r.Header.Get("X-Self-ID")),
		zap.String("platform", r.Header.Get("X-Platform")))

	// Determine Protocol Version
	protocol := "v11" // Default
	wsProtocol := r.Header.Get("Sec-WebSocket-Protocol")
	responseHeader := http.Header{}
	if strings.Contains(wsProtocol, "12.onebot.v12") || strings.Contains(wsProtocol, "onebot.v12") {
		protocol = "v12"
		responseHeader.Set("Sec-WebSocket-Protocol", "12.onebot.v12")
	}

	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		log.Error("Bot WebSocket upgrade failed", zap.Error(err))
		return
	}

	// Generate Bot ID
	selfID := r.Header.Get("X-Self-ID")
	platform := r.Header.Get("X-Platform")
	if platform == "" {
		platform = "qq" // Default platform
	}

	// If no ID is provided, use remote address as temporary ID
	if selfID == "" {
		selfID = conn.RemoteAddr().String()
	}

	// Create Bot client
	bot := &types.BotClient{
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Platform:      platform,
		SelfID:        selfID,
		Protocol:      protocol,
	}

	// Register Bot
	m.Mutex.Lock()
	if m.Bots == nil {
		m.Bots = make(map[string]*types.BotClient)
	}
	botKey := bot.SelfID
	m.Bots[botKey] = bot
	m.Mutex.Unlock()

	// Update online status to online
	if m.DigitalEmployeeService != nil {
		go m.DigitalEmployeeService.UpdateOnlineStatus(bot.SelfID, "online")
	}

	// Update connection stats
	m.ConnectionStats.Mutex.Lock()
	m.ConnectionStats.TotalBotConnections++
	m.ConnectionStats.LastBotActivity[botKey] = time.Now()
	m.ConnectionStats.Mutex.Unlock()

	log.Printf("Bot WebSocket connected: %s (Platform: %s, ID: %s)", conn.RemoteAddr(), bot.Platform, bot.SelfID)

	// Fetch Bot info asynchronously
	go m.fetchBotInfo(bot)

	// Start connection handling loop
	go m.handleBotConnection(bot)
}

// fetchBotInfo actively fetches detailed info of the Bot
func (m *Manager) fetchBotInfo(bot *types.BotClient) {
	// Wait for 1 second to ensure connection is fully established and handshake is complete
	time.Sleep(1 * time.Second)

	log.Printf("[Bot] Fetching info for bot: %s", bot.SelfID)

	// 1. Get login info (nickname, etc.)
	echoInfo := "fetch_info_" + bot.SelfID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	m.PendingMutex.Lock()
	m.PendingRequests[echoInfo] = make(chan types.InternalMessage, 1)
	m.PendingMutex.Unlock()

	reqInfo := types.InternalAction{
		Action: "get_login_info",
		Echo:   echoInfo,
	}

	bot.Mutex.Lock()
	bot.Conn.WriteJSON(reqInfo)
	bot.Mutex.Unlock()

	// 2. Get group list (get group count)
	echoGroups := "fetch_groups_" + bot.SelfID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	m.PendingMutex.Lock()
	m.PendingRequests[echoGroups] = make(chan types.InternalMessage, 1)
	m.PendingMutex.Unlock()

	reqGroups := types.InternalAction{
		Action: "get_group_list",
		Echo:   echoGroups,
	}

	bot.Mutex.Lock()
	bot.Conn.WriteJSON(reqGroups)
	bot.Mutex.Unlock()

	// Wait for response (with timeout)
	timeout := time.After(10 * time.Second)
	infoDone := false
	groupsDone := false

	for !infoDone || !groupsDone {
		select {
		case resp := <-m.PendingRequests[echoInfo]:
			var info types.LoginInfo
			if err := utils.DecodeMapToStruct(resp.Extras["data"], &info); err == nil {
				bot.Mutex.Lock()
				newNickname := info.Nickname
				newSelfID := info.UserID

				if newSelfID != "" && newSelfID != bot.SelfID {
					oldID := bot.SelfID
					oldKey := fmt.Sprintf("%s:%s", bot.Platform, oldID)
					bot.Mutex.Unlock() // Unlock before map operations

					m.Mutex.Lock()
					delete(m.Bots, oldKey)
					bot.SelfID = newSelfID
					bot.Nickname = newNickname
					newKey := fmt.Sprintf("%s:%s", bot.Platform, newSelfID)
					m.Bots[newKey] = bot
					m.Mutex.Unlock()

					bot.Mutex.Lock() // Re-lock
					log.Printf("[Bot] Updated Bot ID from %s to %s via get_login_info", oldID, newSelfID)
				} else {
					bot.Nickname = newNickname
				}
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated info for %s: Nickname=%s", bot.SelfID, bot.Nickname)

				// Persist bot info to database
				go m.SaveBotToDB(bot.SelfID, bot.Nickname, bot.Platform, bot.Protocol)
			}
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echoInfo)
			m.PendingMutex.Unlock()
			infoDone = true

		case resp := <-m.PendingRequests[echoGroups]:
			var groups []types.GroupListItem
			if err := utils.DecodeMapToStruct(resp.Extras["data"], &groups); err == nil {
				bot.Mutex.Lock()
				bot.GroupCount = len(groups)
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated group count for %s: %d", bot.SelfID, bot.GroupCount)

				// Update group cache for subsequent routing API requests
				m.CacheMutex.Lock()
				for _, group := range groups {
					gID := group.GroupID

					if gID != "" {
						name := group.GroupName
						if name == "" {
							name = fmt.Sprintf("Group %s (Auto)", gID)
						}
						// Create GroupInfo struct
						cachedGroup := types.GroupInfo{
							GroupID:   gID,
							GroupName: name,
							BotID:     bot.SelfID,
							IsCached:  true,
							LastSeen:  time.Now(),
						}
						m.GroupCache[gID] = cachedGroup
						// Persist to database
						go m.SaveGroupToDB(gID, name, bot.SelfID)
					}
				}
				m.CacheMutex.Unlock()
				log.Printf("[Bot] Cached %d groups for Bot %s", len(groups), bot.SelfID)
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

// handleBotConnection handles message loop for a single Bot connection
func (m *Manager) handleBotConnection(bot *types.BotClient) {
	// Start heartbeat sending goroutine
	stopHeartbeat := make(chan struct{})
	go m.sendBotHeartbeat(bot, stopHeartbeat)

	defer func() {
		close(stopHeartbeat) // Stop heartbeat
		// Cleanup work when connection is closed
		m.removeBot(bot.SelfID)
		bot.Conn.Close()

		// Update online status to offline
		if m.DigitalEmployeeService != nil {
			go m.DigitalEmployeeService.UpdateOnlineStatus(bot.SelfID, "offline")
		}

		// Record disconnection
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

	// Set read deadline (extended to 120 seconds)
	bot.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	bot.Conn.SetPongHandler(func(string) error {
		bot.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		bot.LastHeartbeat = time.Now()
		// log.Printf("Bot %s received pong", bot.SelfID)
		return nil
	})

	for {
		_, rawMsg, err := bot.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Bot %s read error: %v", bot.SelfID, err)
			}
			break
		}

		// Handle message
		// 0. Convert to InternalMessage for Neural Nexus
		var internalMsg types.InternalMessage
		if bot.Protocol == "v12" {
			var v12Msg onebot.V12RawMessage
			decoder := json.NewDecoder(bytes.NewReader(rawMsg))
			decoder.UseNumber()
			if err := decoder.Decode(&v12Msg); err != nil {
				log.Printf("Bot %s v12 unmarshal error: %v", bot.SelfID, err)
				continue
			}
			internalMsg = m.v12ToInternal(v12Msg)
			log.Printf("[Nexus][%s:%s] Converted v12 to internal: message=%v, raw=%s", bot.Platform, bot.SelfID, internalMsg.Message, internalMsg.RawMessage)
		} else {
			var v11Msg onebot.V11RawMessage
			decoder := json.NewDecoder(bytes.NewReader(rawMsg))
			decoder.UseNumber()
			if err := decoder.Decode(&v11Msg); err != nil {
				log.Printf("Bot %s v11 unmarshal error: %v", bot.SelfID, err)
				continue
			}
			// Ensure self_id is set for v11
			if v11Msg.SelfID == nil || v11Msg.SelfID == "" {
				v11Msg.SelfID = bot.SelfID
			}
			internalMsg = m.v11ToInternal(v11Msg, bot.Platform)
			log.Printf("[Nexus][%s:%s] Converted v11 to internal: message=%v, raw=%s", bot.Platform, bot.SelfID, internalMsg.Message, internalMsg.RawMessage)
		}

		m.handleBotMessage(bot, internalMsg)

		// Update activity time
		bot.LastHeartbeat = time.Now()
		botKey := fmt.Sprintf("%s:%s", bot.Platform, bot.SelfID)
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.LastBotActivity[botKey] = time.Now()
		m.ConnectionStats.Mutex.Unlock()
	}
}

// sendBotHeartbeat sends heartbeat packets to Bot periodically
func (m *Manager) sendBotHeartbeat(bot *types.BotClient, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second) // Send heartbeat every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send ping frame
			bot.Mutex.Lock()
			err := bot.Conn.WriteMessage(websocket.PingMessage, []byte{})
			bot.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send ping to Bot %s: %v", bot.SelfID, err)
				return
			}
			// log.Printf("Sent ping to Bot %s", bot.SelfID)

		case <-stop:
			return
		}
	}
}

// sendWorkerHeartbeat sends heartbeat packets to Worker periodically
func (m *Manager) sendWorkerHeartbeat(worker *types.WorkerClient, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second) // Send heartbeat every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send ping frame
			worker.Mutex.Lock()
			err := worker.Conn.WriteMessage(websocket.PingMessage, []byte{})
			worker.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send ping to Worker %s: %v", worker.ID, err)
				return
			}
			// log.Printf("Sent ping to Worker %s", worker.ID)

		case <-stop:
			return
		}
	}
}

// handleBotMessage handles Bot messages
func (m *Manager) handleBotMessage(bot *types.BotClient, msg types.InternalMessage) {
	// 1. Core plugin intercept
	if allowed, reason, err := m.Core.ProcessMessage(msg); !allowed {
		log.Printf("[Core] Message blocked: %s (reason: %s)", bot.SelfID, reason)
		if err != nil {
			log.Printf("[Core] Error processing message: %v", err)
		}

		// If admin command is intercepted by system (e.g. system closed), try to handle admin command
		if reason == "system_closed" && m.Core.identifyMessageType(msg) == "admin_command" {
			// Continue to let handleBotMessageEvent handle or intercept directly here
		} else {
			return
		}
	}

	// Core plugin handles admin commands
	if m.Core.identifyMessageType(msg) == "admin_command" {
		resp, err := m.Core.HandleAdminCommand(msg)
		if err == nil && resp != "" {
			// Send response back to bot
			m.sendBotMessage(bot, msg, resp)
			return
		}
	}

	// 2. Check if self_id is included and update (if current is temporary ID)
	msgSelfID := msg.SelfID
	if msgSelfID != "" {
		if bot.SelfID != msgSelfID && strings.Contains(bot.SelfID, ":") {
			// Current is temporary IP ID, received formal ID, updating
			oldID := bot.SelfID
			m.Mutex.Lock()
			delete(m.Bots, oldID)
			bot.SelfID = msgSelfID
			m.Bots[bot.SelfID] = bot
			m.Mutex.Unlock()
			log.Printf("[Bot] Updated Bot ID from %s to %s", oldID, msgSelfID)
		}
	}

	// Check if it's an API response (has echo field)
	echo := msg.Echo
	if echo != "" {
		// Broadcast routing event: Bot -> Nexus (Response)
		m.BroadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "response", nil)

		// This is an API response, need to send back to corresponding Worker
		m.PendingMutex.Lock()
		respChan, exists := m.PendingRequests[echo]
		sendTime, timeExists := m.PendingTimestamps[echo]
		delete(m.PendingTimestamps, echo)
		m.PendingMutex.Unlock()

		if exists {
			// Record RTT
			if timeExists {
				rtt := time.Since(sendTime)
				// Find corresponding Worker and update RTT
				// Use | as separator: workerID|originalEcho
				if parts := strings.Split(echo, "|"); len(parts) >= 2 {
					workerID := parts[0]
					m.Mutex.RLock()
					for _, w := range m.Workers {
						if w.ID == workerID {
							w.Mutex.Lock()
							w.LastRTT = rtt
							w.RTTSamples = append(w.RTTSamples, rtt)
							if len(w.RTTSamples) > 20 { // Keep at most 20 samples
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

					// Broadcast routing event: Nexus -> Worker (Response Forward)
					m.BroadcastRoutingEvent("Nexus", workerID, "nexus_to_worker", "response", nil)
				}
			}

			// Send response to waiting Worker request
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

	// Original message processing logic
	// Get message type
	postType := msg.PostType
	if postType == "" {
		log.Printf("[Bot] Warning: Received message from Bot %s without echo or post_type: %v", bot.SelfID, msg)
		return
	}

	// Update statistics
	m.Mutex.Lock()
	bot.RecvCount++
	m.Mutex.Unlock()

	// Only log non-spammy message types to console
	postTypeLower := strings.ToLower(postType)
	// Filter out common high-frequency events and anything containing "log"
	isLogType := strings.Contains(postTypeLower, "log")
	if !isLogType && postTypeLower != "meta_event" && postTypeLower != "message" && postTypeLower != "heartbeat" && !strings.Contains(postTypeLower, "terminal") {
		log.Printf("Received %s event from Bot %s", postType, bot.SelfID)
	}

	// Handle according to message type
	userID := msg.UserID
	groupID := msg.GroupID

	switch postType {
	case "log":
		// 机器人上报的运行日志，直接丢弃，不进行转发
		return
	case "meta_event":
		// Meta event (heartbeat etc.)
		if msg.MetaType == "heartbeat" {
			// Heartbeat event, update status
			// For now we don't need detailed status check from Original
			return // 心跳消息处理完内部状态后直接返回，不转发给 Worker 和 Web UI
		}
	case "message", "message_sent", "notice", "request":
		// Debug log for self-messages
		if userID == bot.SelfID || postType == "message_sent" {
			log.Printf("[Bot] Received self-message from Bot %s: %v", bot.SelfID, msg)
		}

		// 3. Check idempotency (prevent duplicate replies)
		msgID := msg.ID
		if msgID == "" {
			// notice/request may not have message_id, try using post_type + time + user_id as unique identifier
			msgID = fmt.Sprintf("%s:%v:%v", postType, msg.Time, msg.UserID)
		}

		// Self-messages skip idempotency and rate limit checks to ensure they are tracked
		if userID != bot.SelfID && postType != "message_sent" {
			if msgID != "" && !m.CheckIdempotency(msgID) {
				log.Printf("[REDIS] Duplicate %s detected: %s, skipping", postType, msgID)
				return
			}

			// 4. Check rate limit (prevent spam and abuse)
			if !m.CheckRateLimit(userID, groupID) {
				log.Printf("[REDIS] Rate limit exceeded for user %s / group %s", userID, groupID)
				return
			}
		} else {
			if config.GlobalConfig.LogLevel == "DEBUG" {
				log.Printf("[Bot] Self-message %s skipping idempotency and rate limit checks", msgID)
			}
		}

		// 5. Update session context (supports TTL)
		if userID != "" {
			m.UpdateContext(bot.Platform, userID, msg)
		}

		// --- 任务系统 AI 意图识别 (Chat-to-Task) ---
		if m.TaskManager != nil && postType == "message" && userID != bot.SelfID {
			// 只有满足触发条件（唤醒词或关键词）才进入 AI 流程
			isAITrigger := strings.HasPrefix(strings.ToUpper(msg.RawMessage), "AI") ||
				strings.HasPrefix(msg.RawMessage, "#确认") ||
				strings.HasPrefix(msg.RawMessage, "确认") ||
				utils.ContainsOne(msg.RawMessage, "任务", "定时", "禁言套餐")

			log.Printf("[AI-Task] Checking trigger for: '%s', isTrigger: %v", msg.RawMessage, isAITrigger)

			if isAITrigger {
				ctx := context.Background()
				err := m.TaskManager.ProcessChatMessage(ctx, bot.SelfID, groupID, userID, msg.RawMessage)
				if err == nil {
					return // 拦截成功，不再分发
				}
				log.Printf("[AI-Task] ProcessChatMessage error: %v", err)
			}
		}
		// ------------------------------------------

		// 拦截器检查：在分发给 Worker 之前进行全局控制
		if m.TaskManager != nil {
			interceptorCtx := &tasks.InterceptorContext{
				Platform: bot.Platform,
				SelfID:   bot.SelfID,
				UserID:   userID,
				GroupID:  groupID,
				Message:  &msg, // Pass the whole InternalMessage instead of map
			}
			if !m.TaskManager.Interceptors.ProcessBeforeDispatch(interceptorCtx) {
				// 如果被拦截器拦截，则不继续分发
				return
			}
		}
	}

	// Cache group/member/friend info (based on message)
	m.cacheBotDataFromMessage(bot, msg)

	// Enrich message with cached data (names, cards) before forwarding to workers
	m.enrichMessageWithCache(&msg)

	if postType == "message" || postType == "message_sent" {
		m.handleBotMessageEvent(bot, msg)
	}

	// 5.5 Broadcast to subscribers (Web UI Monitor)
	m.BroadcastEvent(msg.ToV11Map())

	// --- 智能体 (Digital Employee) 处理逻辑 ---
	// 如果该 Bot 被定义为“数字员工”，则尝试进行 AI 响应
	if m.DigitalEmployeeService != nil && m.AIIntegrationService != nil && postType == "message" && userID != bot.SelfID {
		employee, err := m.DigitalEmployeeService.GetEmployeeByBotID(bot.SelfID)
		if err == nil && employee != nil {
			log.Printf("[Agent] Bot %s is a Digital Employee: %s (%s)", bot.SelfID, employee.Name, employee.Title)

			// 只有文本消息才触发 AI
			if msg.RawMessage != "" {
				// 调用 AI 进行数字员工响应 (带上下文历史)
				response, err := m.AIIntegrationService.ChatWithEmployee(employee, msg, employee.EnterpriseID)
				if err == nil && response != "" {
					log.Printf("[Agent] AI Response for Bot %s: %s", bot.SelfID, response)
					// 发送回复
					m.sendBotMessage(bot, msg, response)

					// 数字员工回复后，通常不需要再分发给 Worker 处理通用逻辑
					return
				} else if err != nil {
					log.Printf("[Agent] AI Chat failed: %v", err)
				}
			}
		}
	}
	// ------------------------------------------

	// Forward to Worker for processing
	// 打印转发详情，方便排查频繁转发的消息
	log.Printf("[Forwarding] Bot: %s, PostType: %v, MessageType: %v, GroupID: %v, UserID: %v",
		bot.SelfID, msg.PostType, msg.MessageType, msg.GroupID, msg.UserID)

	// 6. Use Redis queue for asynchronous decoupling
	// 如果启用了技能系统，则优先进入 Redis 队列进行异步分发
	if config.ENABLE_SKILL && m.Rdb != nil {
		targetWorkerID := m.getTargetWorkerID(msg)
		if userID == bot.SelfID {
			log.Printf("[Routing] Self-message from Bot %s, TargetWorkerID: %s", bot.SelfID, targetWorkerID)
		}
		err := m.PushToRedisQueue(targetWorkerID, msg)
		if err == nil {
			m.BroadcastRoutingEvent("Nexus", "RedisQueue", "nexus_to_redis", "push", nil)
			if userID == bot.SelfID {
				log.Printf("[Routing] Self-message from Bot %s pushed to Redis queue", bot.SelfID)
			}

			// 影子执行 (Shadow Mode): 如果有影子 Worker，额外推送一份
			var shadowWorkerID string
			if sID, ok := msg.Extras["shadow_worker_id"].(string); ok {
				shadowWorkerID = sID
			}

			if shadowWorkerID != "" {
				log.Printf("[Shadow] Pushing shadow copy to worker: %s", shadowWorkerID)
				// 克隆消息并打上影子标记
				shadowMsg := msg // Struct copy
				// How to mark shadow? InternalMessage doesn't have IsShadow field.
				// We can add it to Original if it's a map, but we'll use ToV11Map later anyway.
				m.PushToRedisQueue(shadowWorkerID, shadowMsg)
			}
			return
		}
		log.Printf("[REDIS] Failed to push to queue: %v. Falling back to WebSocket direct forwarding.", err)
	}

	// Fallback: If Redis is unavailable or push fails, use original WebSocket direct forwarding
	m.forwardMessageToWorker(msg)

	// 影子执行 (Shadow Mode) - Fallback 路径
	var shadowWorkerID string
	if sID, ok := msg.Extras["shadow_worker_id"].(string); ok {
		shadowWorkerID = sID
	}
	if shadowWorkerID != "" {
		m.forwardMessageToWorkerWithTarget(msg, shadowWorkerID)
	}
}

// enrichMessageWithCache supplements the message with cached group and user information
func (m *Manager) enrichMessageWithCache(msg *types.InternalMessage) {
	if msg.MessageType != "private" && msg.MessageType != "group" {
		return
	}

	m.CacheMutex.RLock()
	defer m.CacheMutex.RUnlock()

	groupID := msg.GroupID
	userID := msg.UserID

	// 1. Enrich Group Name
	if groupID != "" {
		if group, ok := m.GroupCache[groupID]; ok {
			if group.GroupName != "" {
				if msg.GroupName == "" || strings.Contains(msg.GroupName, groupID) {
					msg.GroupName = group.GroupName
				}
			}
		}
	}

	// 2. Enrich User Nickname and Group Card
	if userID != "" {
		nickname := ""
		card := ""

		// Try member cache first (for group messages)
		if groupID != "" {
			memberKey := fmt.Sprintf("%s:%s", groupID, userID)
			if member, ok := m.MemberCache[memberKey]; ok {
				nickname = member.Nickname
				card = member.Card
			}
		}

		// Try friend cache (or fallback for nickname)
		if nickname == "" {
			if friend, ok := m.FriendCache[userID]; ok {
				nickname = friend.Nickname
			}
		}

		if nickname != "" || card != "" {
			if msg.SenderName == "" && nickname != "" {
				msg.SenderName = nickname
			}
			if msg.SenderCard == "" && card != "" {
				msg.SenderCard = card
			}
		}
	}
}

// getTargetWorkerID helper method: Get target Worker ID based on routing rules
func (m *Manager) getTargetWorkerID(msg types.InternalMessage) string {
	// 0. 优先处理正则触发器 (Fast-Track Regex Matching)
	// 插件上报的正则指令具有最高优先级，命中后直接路由
	if m.TaskManager != nil && m.TaskManager.AI != nil {
		if skill, matched := m.TaskManager.AI.MatchSkillByRegex(msg.RawMessage); matched {
			log.Printf("[Routing] Regex matched skill: %s. Routing to capable worker.", skill.Name)
			workerID := m.FindWorkerBySkill(skill.Name)
			if workerID != "" {
				log.Printf("[Routing] Regex Fast-Track: skill=%s -> worker=%s", skill.Name, workerID)
				return workerID
			}
		}
	}

	// 1. 处理拦截器注入的语义路由提示 (Intelligent Semantic Routing)
	if hint, ok := msg.Extras["intent_hint"].(string); ok && hint != "" {
		log.Printf("[Routing] Using semantic intent hint: %s", hint)
		// 如果提示是技能名称，则寻找具备该能力的 Worker
		if strings.HasPrefix(hint, "skill:") {
			skillName := strings.TrimPrefix(hint, "skill:")
			workerID := m.FindWorkerBySkill(skillName)
			if workerID != "" {
				log.Printf("[Routing] Resolved skill hint %s to worker %s", skillName, workerID)
				return workerID
			}
		}
		return hint
	}

	var matchKeys []string

	// Extract match keys
	if msg.UserID != "" && msg.UserID != "0" {
		matchKeys = append(matchKeys, fmt.Sprintf("user_%s", msg.UserID), msg.UserID)
	}
	if msg.GroupID != "" && msg.GroupID != "0" {
		matchKeys = append(matchKeys, fmt.Sprintf("group_%s", msg.GroupID), msg.GroupID)
	}
	if msg.SelfID != "" {
		matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", msg.SelfID), msg.SelfID)
	}

	// 1. Prioritize getting dynamic routing rules from Redis
	if m.Rdb != nil {
		ctx := context.Background()
		for _, key := range matchKeys {
			if wID, err := m.Rdb.HGet(ctx, config.REDIS_KEY_DYNAMIC_RULES, key).Result(); err == nil && wID != "" {
				log.Printf("[REDIS] Dynamic route matched (Redis): %s -> %s", key, wID)
				return wID
			}
		}

		rules, err := m.Rdb.HGetAll(ctx, config.REDIS_KEY_DYNAMIC_RULES).Result()
		if err == nil && len(rules) > 0 {
			for p, w := range rules {
				if strings.Contains(p, "*") {
					for _, key := range matchKeys {
						if utils.MatchRoutePattern(p, key) {
							log.Printf("[REDIS] Dynamic wildcard route matched (Redis): %s (%s) -> %s", p, key, w)
							return w
						}
					}
				}
			}
		}
	}

	// 2. Fallback to static routing rules in memory
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	// A. Exact match
	for _, key := range matchKeys {
		if wID, exists := m.RoutingRules[key]; exists && wID != "" {
			return wID
		}
	}

	// B. Wildcard match
	for p, w := range m.RoutingRules {
		if strings.Contains(p, "*") {
			for _, key := range matchKeys {
				if utils.MatchRoutePattern(p, key) {
					return w
				}
			}
		}
	}

	return ""
}

// cacheBotDataFromMessage extracts and caches data from messages
func (m *Manager) cacheBotDataFromMessage(bot *types.BotClient, msg types.InternalMessage) {
	// Only cache from messages (not notice/request/etc for now, or adapt as needed)
	// In InternalMessage, MessageType usually indicates private/group
	if msg.MessageType == "" {
		return
	}

	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// Cache group info
	if msg.GroupID != "" {
		gID := msg.GroupID

		existingGroup, exists := m.GroupCache[gID]
		groupName := ""
		if exists {
			groupName = existingGroup.GroupName
		}

		if groupName == "" || strings.Contains(groupName, "(Cached)") {
			if groupName == "" {
				groupName = "Group " + gID
			}
		}

		m.GroupCache[gID] = types.GroupInfo{
			GroupID:   gID,
			GroupName: groupName,
			BotID:     bot.SelfID,
			IsCached:  true,
			LastSeen:  time.Now(),
		}
		go m.SaveGroupToDB(gID, groupName, bot.SelfID)

		// Cache member info
		if msg.UserID != "" {
			uID := msg.UserID
			key := fmt.Sprintf("%s:%s", gID, uID)

			nickname := msg.SenderName
			card := msg.SenderCard
			role := ""
			if msg.Extras != nil {
				if r, ok := msg.Extras["role"].(string); ok {
					role = r
				}
			}

			m.MemberCache[key] = types.MemberInfo{
				GroupID:  gID,
				UserID:   uID,
				Nickname: nickname,
				Card:     card,
				Role:     role,
				IsCached: true,
				LastSeen: time.Now(),
			}
			go m.SaveMemberToDB(gID, uID, nickname, card, role)
		}
	} else if msg.UserID != "" {
		// Cache friend info (private chat)
		uID := msg.UserID
		if _, exists := m.FriendCache[uID]; !exists {
			nickname := msg.SenderName

			m.FriendCache[uID] = types.FriendInfo{
				UserID:   uID,
				Nickname: nickname,
				BotID:    bot.SelfID,
				IsCached: true,
				LastSeen: time.Now(),
			}
			go m.SaveFriendToDB(uID, nickname, bot.SelfID)
		}
	}
}

// isSystemMessage checks if it's a system message or meaningless statistical data
func isSystemMessage(msg types.InternalMessage) bool {
	// 1. Extract basic fields
	userID := msg.UserID
	msgType := msg.MessageType
	message := msg.RawMessage
	subType := msg.SubType

	return checkSystemMessage(userID, msgType, subType, message)
}

// isSystemAction checks if it's a system action (send message to system account)
func isSystemAction(action types.InternalAction) bool {
	userID := action.UserID
	if userID == "" && action.Params != nil {
		userID = utils.ToString(action.Params["user_id"])
	}

	msgType := action.MessageType
	if msgType == "" && action.Params != nil {
		msgType = utils.ToString(action.Params["message_type"])
	}

	message := utils.ToString(action.Message)
	if message == "" && action.Params != nil {
		message = utils.ToString(action.Params["message"])
	}

	subType := "" // Actions usually don't have sub_type at top level

	return checkSystemMessage(userID, msgType, subType, message)
}

// checkSystemMessage is a helper for both isSystemMessage and isSystemAction
func checkSystemMessage(userID, msgType, subType, message string) bool {
	// 1. Check common system user IDs
	systemIDs := map[string]bool{
		"10000":      true, // System message
		"1000000":    true, // Group system message
		"1000001":    true, // Friend system message
		"1000002":    true, // Operation message
		"80000000":   true, // Anonymous message
		"2852199017": true, // QQ Official
	}
	if systemIDs[userID] {
		return true
	}

	// 2. Check message type and sub-type
	if msgType == "system" || subType == "system" {
		return true
	}

	// 3. Check if content is empty
	if message == "" {
		return true
	}

	return false
}

// sendBotMessage sends text message back to bot (helper method)
func (m *Manager) sendBotMessage(bot *types.BotClient, originalMsg types.InternalMessage, text string) {
	action := types.InternalAction{}

	if bot.Protocol == "v12" {
		// OneBot v12 format
		action.Action = "send_message" // v12 standard action name
		action.Message = []types.MessageSegment{
			{
				Type: "text",
				Data: types.TextSegmentData{
					Text: text,
				},
			},
		}
		if originalMsg.GroupID != "" {
			action.GroupID = originalMsg.GroupID
			action.DetailType = "group"
		} else if originalMsg.UserID != "" {
			action.UserID = originalMsg.UserID
			action.DetailType = "private"
		}
	} else {
		// OneBot v11 format
		action.Action = "send_msg"
		action.Message = text
		if originalMsg.GroupID != "" {
			action.GroupID = originalMsg.GroupID
		} else if originalMsg.UserID != "" {
			action.UserID = originalMsg.UserID
		}
	}

	bot.Mutex.Lock()
	defer bot.Mutex.Unlock()

	// 如果是 Online 平台或者连接断开，我们通过广播事件来让 WebUI 收到消息
	if bot.Conn == nil || bot.Platform == "Online" {
		// 构造一个响应消息
		resp := types.InternalMessage{
			ID:          fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Time:        time.Now().Unix(),
			Platform:    bot.Platform,
			SelfID:      bot.SelfID,
			PostType:    "message_sent",
			MessageType: originalMsg.MessageType,
			RawMessage:  text,
			UserID:      originalMsg.UserID,
			GroupID:     originalMsg.GroupID,
			SenderName:  "AI Assistant",
		}

		// 广播给 Web UI Monitor
		m.BroadcastEvent(resp.ToV11Map())

		// 同时通过 RoutingEvent 记录
		params := &types.RoutingParams{
			UserID:   originalMsg.UserID,
			Content:  text,
			Platform: bot.Platform,
			UserName: "AI Assistant",
		}
		if originalMsg.GroupID != "" {
			params.GroupID = originalMsg.GroupID
			m.BroadcastRoutingEvent(bot.SelfID, originalMsg.GroupID, "bot_to_group", "message", params)
		} else {
			m.BroadcastRoutingEvent(bot.SelfID, originalMsg.UserID, "bot_to_user", "message", params)
		}

		log.Printf("[Bot] Message sent via Broadcast for Bot %s (Online/NoConn)", bot.SelfID)
		return
	}

	err := bot.Conn.WriteJSON(action)
	if err != nil {
		log.Printf("[Bot] Failed to send message to Bot %s: %v", bot.SelfID, err)
	}
}

// handleBotMessageEvent handles Bot message events
func (m *Manager) handleBotMessageEvent(bot *types.BotClient, msg types.InternalMessage) {
	// Extract message info
	userID := msg.UserID
	groupID := msg.GroupID
	message := msg.RawMessage

	// Prepare broadcast routing params
	params := &types.RoutingParams{
		UserID:   userID,
		Content:  message,
		Platform: bot.Platform,
	}

	// Extract sender info (nickname and avatar) from message if available
	if msg.SenderName != "" {
		params.UserName = msg.SenderName
	}

	// Platform specific avatar logic
	if msg.UserAvatar != "" {
		params.UserAvatar = msg.UserAvatar
	} else {
		switch strings.ToUpper(bot.Platform) {
		case "QQ":
			if userID != "" {
				params.UserAvatar = fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", userID)
			}
		case "WECHAT", "WX", "TENCENT":
			if strings.ToUpper(bot.Platform) == "WECHAT" || strings.ToUpper(bot.Platform) == "WX" {
				params.UserAvatar = "/static/avatars/wechat_default.png"
			}
		}
	}

	if groupID != "" {
		params.GroupID = groupID
		// Try to get group name from cache
		m.CacheMutex.RLock()
		if group, ok := m.GroupCache[groupID]; ok {
			if group.GroupName != "" {
				params.GroupName = group.GroupName
			}
		}
		m.CacheMutex.RUnlock()
	}

	// 1. Broadcast routing event: Path based on group or private
	if groupID != "" {
		// Group path: User -> Group -> Nexus
		if userID != "" {
			params.SourceType = "user"
			params.TargetType = "group"
			params.SourceLabel = params.UserName
			params.TargetLabel = params.GroupName
			m.BroadcastRoutingEvent(userID, groupID, "user_to_group", "message", params)
		}

		// Group -> Nexus
		params.SourceType = "group"
		params.TargetType = "nexus"
		params.SourceLabel = params.GroupName
		m.BroadcastRoutingEvent(groupID, "Nexus", "group_to_nexus", "message", params)
	} else {
		// Private path: User -> Bot -> Nexus
		if userID != "" {
			params.SourceType = "user"
			params.TargetType = "bot"
			params.SourceLabel = params.UserName
			m.BroadcastRoutingEvent(userID, bot.SelfID, "user_to_bot", "message", params)
		}

		// Bot -> Nexus
		params.SourceType = "bot"
		params.TargetType = "nexus"
		params.SourceLabel = bot.Nickname
		m.BroadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "message", params)
	}

	// Update detailed stats (exclude system messages)
	if !isSystemMessage(msg) {
		m.UpdateBotStats(bot.SelfID, userID, groupID)

		// 保存到数据库
		messageID := msg.ID
		msgType := msg.MessageType
		rawMsg, _ := json.Marshal(msg)
		go m.SaveMessageToDB(messageID, bot.SelfID, userID, groupID, msgType, message, string(rawMsg))
	}
}

// removeBot removes Bot connection
func (m *Manager) removeBot(botID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if _, exists := m.Bots[botID]; exists {
		delete(m.Bots, botID)
		log.Printf("Removed Bot %s from active connections", botID)
	}
}

// cacheMessage caches messages that cannot be processed immediately
func (m *Manager) cacheMessage(msg types.InternalMessage) {
	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// Limit cache size to prevent memory overflow
	if len(m.MessageCache) > 1000 {
		m.MessageCache = m.MessageCache[1:] // Discard oldest message
	}
	m.MessageCache = append(m.MessageCache, msg)
	log.Printf("[CACHE] No workers available, message cached (Total: %d)", len(m.MessageCache))
}

// flushMessageCache sends cached messages when a new Worker connects
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
		// Resend via routing
		go m.forwardMessageToWorker(msg)
	}
}

// matchRoutePattern checks if a string matches a pattern (supports * wildcard)
func matchRoutePattern(pattern, value string) bool {
	return utils.MatchRoutePattern(pattern, value)
}

// findWorkerByID finds a Worker by ID
func (m *Manager) findWorkerByID(workerID string) *types.WorkerClient {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	for _, w := range m.Workers {
		if w.ID == workerID {
			return w
		}
	}
	return nil
}

// forwardMessageToWorker forwards the message to a Worker for processing
func (m *Manager) forwardMessageToWorker(msg types.InternalMessage) {
	m.forwardMessageToWorkerWithRetry(msg, 0)
}

// forwardMessageToWorkerWithTarget 直接将消息发送给指定的 Worker
func (m *Manager) forwardMessageToWorkerWithTarget(msg types.InternalMessage, targetWorkerID string) {
	if targetWorkerID == "" {
		return
	}

	// 确保消息有 echo
	echo := msg.Echo
	if echo == "" {
		echo = fmt.Sprintf("shadow_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	}

	if w := m.findWorkerByID(targetWorkerID); w != nil {
		log.Printf("[ROUTING] Direct forwarding to Target Worker: %s", targetWorkerID)

		// 根据 Worker 协议转换消息格式
		var finalMsg any
		if w.Protocol == "v12" {
			finalMsg = msg.ToV12Map()
		} else {
			v11Map := msg.ToV11Map()
			v11Map["echo"] = echo
			finalMsg = v11Map
		}

		w.Mutex.Lock()
		err := w.Conn.WriteJSON(finalMsg)
		w.Mutex.Unlock()

		if err == nil {
			m.Mutex.Lock()
			w.HandledCount++
			m.Mutex.Unlock()
			m.BroadcastRoutingEvent("Nexus", targetWorkerID, "nexus_to_worker", "message_forward", nil)
		} else {
			log.Printf("[ROUTING] Failed to forward to worker %s: %v", targetWorkerID, err)
		}
	} else {
		log.Printf("[ROUTING] Target Worker %s not found for direct forwarding", targetWorkerID)
	}
}

// forwardMessageToWorkerWithRetry message forwarding with retry limit
func (m *Manager) forwardMessageToWorkerWithRetry(msg types.InternalMessage, retryCount int) {
	if retryCount > 3 {
		log.Printf("[ROUTING] [ERROR] Maximum retry count exceeded for message. Caching message.")
		m.cacheMessage(msg)
		return
	}

	// 1. Try to use routing rules
	var targetWorkerID string
	var matchKeys []string

	// Extract matching keys
	if msg.UserID != "" && msg.UserID != "0" {
		matchKeys = append(matchKeys, fmt.Sprintf("user_%s", msg.UserID))
		matchKeys = append(matchKeys, msg.UserID)
	}

	if msg.GroupID != "" && msg.GroupID != "0" {
		matchKeys = append(matchKeys, fmt.Sprintf("group_%s", msg.GroupID))
		matchKeys = append(matchKeys, msg.GroupID)
	}

	if msg.SelfID != "" {
		matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", msg.SelfID))
		matchKeys = append(matchKeys, msg.SelfID)
	}

	// Find matching rules
	m.Mutex.RLock()
	var matchedKey string

	// A. Exact match priority
	for _, key := range matchKeys {
		if wID, exists := m.RoutingRules[key]; exists && wID != "" {
			targetWorkerID = wID
			matchedKey = key
			break
		}
	}

	// B. If no exact match, try wildcard match
	if targetWorkerID == "" {
		for p, w := range m.RoutingRules {
			if strings.Contains(p, "*") {
				for _, key := range matchKeys {
					if utils.MatchRoutePattern(p, key) {
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

	// Ensure the message has a unique echo
	echo := msg.Echo
	if echo == "" {
		echo = fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	}

	// If a target Worker is found, try to get it
	if targetWorkerID != "" {
		if w := m.findWorkerByID(targetWorkerID); w != nil {
			log.Printf("[ROUTING] Rule Matched: %s -> Target Worker: %s", matchedKey, targetWorkerID)

			// 根据 Worker 协议转换
			var finalMsg any
			if w.Protocol == "v12" {
				finalMsg = msg.ToV12Map()
			} else {
				v11Map := msg.ToV11Map()
				v11Map["echo"] = echo
				finalMsg = v11Map
			}

			w.Mutex.Lock()
			err := w.Conn.WriteJSON(finalMsg)
			w.Mutex.Unlock()

			if err == nil {
				m.WorkerRequestMutex.Lock()
				m.WorkerRequestTimes[echo] = time.Now()
				m.WorkerRequestMutex.Unlock()

				m.Mutex.Lock()
				w.HandledCount++
				m.Mutex.Unlock()

				m.BroadcastRoutingEvent("Nexus", targetWorkerID, "nexus_to_worker", "message", nil)
				return
			}
			log.Printf("[ROUTING] [ERROR] Failed to send to target worker %s: %v. Falling back to load balancer.", targetWorkerID, err)
		} else {
			log.Printf("[ROUTING] [WARNING] Target worker %s defined in rule (%s) is OFFLINE or NOT FOUND. Falling back to load balancer.", targetWorkerID, matchedKey)
		}
	}

	// 2. Load balancing forwarding
	m.Mutex.RLock()
	var healthyWorkers []*types.WorkerClient
	for _, w := range m.Workers {
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

	var selectedWorker *types.WorkerClient
	m.Mutex.Lock()

	if len(healthyWorkers) == 1 {
		selectedWorker = healthyWorkers[0]
	} else {
		var unhandled []*types.WorkerClient
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
			var minProcessTime time.Duration = -1
			for _, w := range healthyWorkers {
				if w.AvgProcessTime > 0 {
					if minProcessTime == -1 || w.AvgProcessTime < minProcessTime {
						minProcessTime = w.AvgProcessTime
						selectedWorker = w
					}
				}
			}

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

			if selectedWorker == nil {
				idx := m.WorkerIndex % len(healthyWorkers)
				selectedWorker = healthyWorkers[idx]
				m.WorkerIndex++
			}
		}
	}
	m.Mutex.Unlock()

	if selectedWorker != nil {
		// 根据协议转换
		var finalMsg any
		if selectedWorker.Protocol == "v12" {
			finalMsg = msg.ToV12Map()
		} else {
			v11Map := msg.ToV11Map()
			v11Map["echo"] = echo
			finalMsg = v11Map
		}

		selectedWorker.Mutex.Lock()
		err := selectedWorker.Conn.WriteJSON(finalMsg)
		selectedWorker.Mutex.Unlock()

		if err == nil {
			m.WorkerRequestMutex.Lock()
			m.WorkerRequestTimes[echo] = time.Now()
			m.WorkerRequestMutex.Unlock()

			m.Mutex.Lock()
			selectedWorker.HandledCount++
			m.Mutex.Unlock()

			m.BroadcastRoutingEvent("Nexus", selectedWorker.ID, "nexus_to_worker", "message", nil)
			log.Printf("[ROUTING] Forwarded to worker %s (AvgRTT: %v, Handled: %d)", selectedWorker.ID, selectedWorker.AvgRTT, selectedWorker.HandledCount)
		} else {
			log.Printf("[ROUTING] [ERROR] Failed to forward to selected worker %s: %v. Removing and retrying...", selectedWorker.ID, err)
			m.removeWorker(selectedWorker.ID)
			m.forwardMessageToWorkerWithRetry(msg, retryCount+1)
		}
	}
}

// handleWorkerWebSocket handles Worker WebSocket connections
func (m *Manager) handleWorkerWebSocket(w http.ResponseWriter, r *http.Request) {
	// Record detailed information of Worker connection attempt
	log.Printf("Worker WebSocket connection attempt from %s - Headers: X-Self-ID=%s, X-Platform=%s",
		r.RemoteAddr, r.Header.Get("X-Self-ID"), r.Header.Get("X-Platform"))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Worker WebSocket upgrade failed: %v", err)
		return
	}

	// Generate Worker ID (prefer X-Self-ID from header if available)
	workerID := r.Header.Get("X-Self-ID")
	if workerID == "" {
		workerID = conn.RemoteAddr().String()
	}

	// Determine Protocol Version for Worker
	protocol := "v11" // Default for backward compatibility
	wsProtocol := r.Header.Get("Sec-WebSocket-Protocol")
	if strings.Contains(wsProtocol, "v12") || r.Header.Get("X-Protocol") == "v12" {
		protocol = "v12"
	}

	// Create Worker client - explicitly identified as a Worker connection
	worker := &types.WorkerClient{
		ID:            workerID,
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Protocol:      protocol,
	}

	log.Printf("Worker client created successfully: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// Register Worker
	m.Mutex.Lock()
	m.Workers = append(m.Workers, worker)
	m.Mutex.Unlock()

	// 广播 Worker 状态更新事件
	go m.BroadcastEvent(types.WorkerUpdateEvent{
		Type: "worker_update",
		Data: types.WorkerInfo{
			ID:           worker.ID,
			RemoteAddr:   worker.ID,
			Connected:    worker.Connected.Format("2006-01-02 15:04:05"),
			HandledCount: worker.HandledCount,
			AvgRTT:       worker.AvgRTT.String(),
			LastRTT:      worker.LastRTT.String(),
			IsAlive:      true,
			Status:       "Online",
		},
	})

	// Update connection statistics
	m.ConnectionStats.Mutex.Lock()
	m.ConnectionStats.TotalWorkerConnections++
	if m.ConnectionStats.LastWorkerActivity == nil {
		m.ConnectionStats.LastWorkerActivity = make(map[string]time.Time)
	}
	m.ConnectionStats.LastWorkerActivity[workerID] = time.Now()
	m.ConnectionStats.Mutex.Unlock()

	log.Printf("Worker WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// 尝试从 Redis 恢复能力缓存 (如果存在)
	m.loadWorkerCapabilitiesFromRedis(worker)

	// Try to send cached messages
	go m.flushMessageCache()

	// Start heartbeat loop
	stopChan := make(chan struct{})
	go m.sendWorkerHeartbeat(worker, stopChan)

	// Start connection processing loop
	go func() {
		m.handleWorkerConnection(worker)
		close(stopChan) // Stop heartbeat when connection processing ends
	}()
}

// loadWorkerCapabilitiesFromRedis 尝试从 Redis 恢复 Worker 的能力列表
func (m *Manager) loadWorkerCapabilitiesFromRedis(worker *types.WorkerClient) {
	if m.Rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. 恢复能力列表 (Capabilities)
	capKey := fmt.Sprintf("botmatrix:worker:%s:capabilities", worker.ID)
	capsData, err := m.Rdb.Get(ctx, capKey).Result()
	if err == nil && capsData != "" {
		var capabilities []types.WorkerCapability
		if err := json.Unmarshal([]byte(capsData), &capabilities); err == nil {
			worker.Capabilities = capabilities
			log.Printf("[Worker] Restored %d capabilities for %s from Redis", len(capabilities), worker.ID)
			m.SyncWorkerSkills()
		}
	}

	// 2. 恢复元数据 (Metadata/Plugins)
	metaKey := fmt.Sprintf("botmatrix:worker:%s:plugins", worker.ID)
	metaData, err := m.Rdb.Get(ctx, metaKey).Result()
	if err == nil && metaData != "" {
		var pluginsInfo []map[string]any
		if err := json.Unmarshal([]byte(metaData), &pluginsInfo); err == nil {
			worker.Metadata = map[string]any{
				"plugins": pluginsInfo,
			}
			log.Printf("[Worker] Restored metadata for %s from Redis", worker.ID)
		}
	}
}

// FindWorkerBySkill 寻找具备指定能力的 Worker，返回其 ID
func (m *Manager) FindWorkerBySkill(skillName string) string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	// 收集所有具备该能力的候选者
	var candidates []string
	for _, w := range m.Workers {
		// 检查心跳，确保 Worker 在线 (1分钟内有活动)
		if time.Since(w.LastHeartbeat) > 1*time.Minute {
			continue
		}

		for _, cap := range w.Capabilities {
			if cap.Name == skillName {
				candidates = append(candidates, w.ID)
				break
			}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// 简单的负载均衡：随机选择一个
	return candidates[rand.Intn(len(candidates))]
}

// SyncWorkerSkills 汇总所有 Worker 的能力并同步给 AI 解析器
func (m *Manager) SyncWorkerSkills() {
	m.Mutex.RLock()
	var allSkills []tasks.Capability
	seen := make(map[string]bool)

	for _, w := range m.Workers {
		for _, cap := range w.Capabilities {
			if !seen[cap.Name] {
				allSkills = append(allSkills, tasks.Capability{
					Name:        cap.Name,
					Description: cap.Description,
					Example:     cap.Usage,
					Params:      cap.Params,
					Regex:       cap.Regex,
				})
				seen[cap.Name] = true
			}
		}
	}
	m.Mutex.RUnlock()

	if m.TaskManager != nil && m.TaskManager.AI != nil {
		m.TaskManager.AI.UpdateSkills(allSkills)
		log.Printf("[AI] Synced %d unique skills from %d workers", len(allSkills), len(m.Workers))
	} else {
		log.Printf("[AI] TaskManager or AI not initialized, skipping skill sync")
	}
}

// handleWorkerConnection handles the message loop for a single Worker connection
func (m *Manager) handleWorkerConnection(worker *types.WorkerClient) {
	// Start heartbeat goroutine
	stopHeartbeat := make(chan struct{})
	go m.sendWorkerHeartbeat(worker, stopHeartbeat)

	defer func() {
		close(stopHeartbeat)
		// Cleanup work when connection is closed
		m.removeWorker(worker.ID)
		worker.Conn.Close()

		// 广播 Worker 离线事件
		go m.BroadcastEvent(types.WorkerUpdateEvent{
			Type: "worker_update",
			Data: types.WorkerInfo{
				ID:      worker.ID,
				IsAlive: false,
				Status:  "Offline",
			},
		})

		// Record disconnection
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

	// Set read deadline
	worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	worker.Conn.SetPongHandler(func(string) error {
		worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		worker.LastHeartbeat = time.Now()
		return nil
	})

	for {
		var msg types.WorkerMessage
		err := utils.ReadJSONWithNumber(worker.Conn, &msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Worker %s read error: %v", worker.ID, err)
			}
			break
		}

		// Handle Worker response
		m.handleWorkerMessage(worker, msg)

		// Update activity time
		worker.LastHeartbeat = time.Now()
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.LastWorkerActivity[worker.ID] = time.Now()
		m.ConnectionStats.Mutex.Unlock()
	}
}

// handleWorkerMessage handles Worker messages
func (m *Manager) handleWorkerMessage(worker *types.WorkerClient, msg types.WorkerMessage) {
	// 打印从 Worker 接收到的详细 Payload
	if payloadJson, err := json.Marshal(msg); err == nil {
		log.Printf("[WorkerMsg] Received from Worker %s: %s", worker.ID, string(payloadJson))
	}

	// Only record key information, do not print full message
	msgType := msg.Type

	// 处理 Worker 报备的能力列表
	if msgType == "register_capabilities" || msgType == "update_metadata" {
		updated := false
		if len(msg.Capabilities) > 0 {
			worker.Capabilities = msg.Capabilities
			m.SyncWorkerSkills()
			log.Printf("[Worker] Worker %s registered %d capabilities", worker.ID, len(msg.Capabilities))
			updated = true

			// 持久化到 Redis
			if m.Rdb != nil {
				go func(wID string, caps []types.WorkerCapability) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					key := fmt.Sprintf("botmatrix:worker:%s:capabilities", wID)
					data, _ := json.Marshal(caps)
					m.Rdb.Set(ctx, key, string(data), 24*time.Hour)
				}(worker.ID, msg.Capabilities)
			}
		}
		if msg.Metadata != nil {
			worker.Metadata = msg.Metadata
			log.Printf("[Worker] Worker %s updated metadata", worker.ID)
			updated = true

			// 持久化到 Redis
			if m.Rdb != nil {
				go func(wID string, meta map[string]any) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					key := fmt.Sprintf("botmatrix:worker:%s:plugins", wID)
					data, _ := json.Marshal(meta)
					m.Rdb.Set(ctx, key, string(data), 24*time.Hour)
				}(worker.ID, msg.Metadata)
			}
		}
		if updated {
			return
		}
	}

	action := msg.Action
	echo := msg.Echo

	if action != "" || (echo != "" && msgType == "") {
		// This is an API request initiated by a Worker, needs to be forwarded to the Bot
		log.Printf("Worker %s API request: action=%s, echo=%s", worker.ID, action, echo)

		// Broadcast routing event: Worker -> Nexus (Request)
		m.BroadcastRoutingEvent(worker.ID, "Nexus", "worker_to_nexus", "request", nil)

		// Convert to InternalAction
		internalAction := types.InternalAction{
			Action:      action,
			Echo:        echo,
			SelfID:      msg.SelfID,
			Platform:    msg.Platform,
			GroupID:     msg.GroupID,
			UserID:      msg.UserID,
			MessageType: msg.MessageType,
			Message:     msg.Reply,
		}

		if msg.Params != nil {
			internalAction.Params = make(map[string]any)
			for k, v := range msg.Params {
				// 允许这些字段保留在 Params 中，因为 OneBot v11 需要它们在 params 内部
				if k == "self_id" || k == "platform" || k == "message_type" {
					continue
				}
				internalAction.Params[k] = v
			}
		}

		// Forward to Bot
		m.forwardWorkerRequestToBot(worker, internalAction, echo)
		return
	} else if msgType == "skill_result" {
		if config.ENABLE_SKILL {
			// 处理通过 WebSocket 上报的技能执行结果
			skillResult := types.SkillResult{
				TaskID:      msg.TaskID,
				ExecutionID: msg.ExecutionID,
				Status:      msg.Status,
				Result:      msg.Result,
				Error:       msg.Error,
				WorkerID:    worker.ID,
			}
			m.HandleSkillResult(skillResult)
		}
	} else {
		log.Printf("Worker %s event/response: type=%s", worker.ID, msgType)

		// Update statistics
		m.Mutex.Lock()
		worker.HandledCount++
		m.Mutex.Unlock()

		// 1. Statistics processing time
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

				// 广播 Worker 状态更新（包含处理时间）
				go m.BroadcastEvent(types.WorkerUpdateEvent{
					Type: "worker_update",
					Data: types.WorkerInfo{
						ID:              worker.ID,
						HandledCount:    worker.HandledCount,
						LastProcessTime: worker.LastProcessTime.String(),
						AvgProcessTime:  worker.AvgProcessTime.String(),
						Status:          "Online",
						IsAlive:         true,
					},
				})
			} else {
				m.WorkerRequestMutex.Unlock()
			}
		}

		// 2. Check if it contains a reply (some frameworks allow direct reply in event response)
		if msg.Reply != "" {
			log.Printf("Worker %s sent passive reply: %s", worker.ID, msg.Reply)
			// Construct a send_msg request and forward it to the Bot
			m.handleWorkerPassiveReply(worker, msg)
		}
	}
}

// handleWorkerPassiveReply handles passive replies from Workers
func (m *Manager) handleWorkerPassiveReply(worker *types.WorkerClient, msg types.WorkerMessage) {
	// Extract echo (if the Worker included echo in the passive reply)
	echo := msg.Echo

	// Construct InternalAction
	action := types.InternalAction{
		Action:      "send_msg",
		Echo:        echo,
		SelfID:      msg.SelfID,
		Platform:    msg.Platform,
		Message:     msg.Reply,
		GroupID:     msg.GroupID,
		UserID:      msg.UserID,
		MessageType: msg.MessageType,
	}

	// If there are other params, merge them
	if msg.Params != nil {
		action.Params = make(map[string]any)
		for k, v := range msg.Params {
			// 允许这些字段保留在 Params 中，因为 OneBot v11 需要它们在 params 内部
			if k == "message_type" {
				continue
			}
			action.Params[k] = v
		}
	}

	// Forward to Bot
	m.forwardWorkerRequestToBot(worker, action, echo)
}

// removeWorker removes a Worker connection
func (m *Manager) removeWorker(workerID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	// Remove from Workers array
	newWorkers := make([]*types.WorkerClient, 0, len(m.Workers))
	for _, w := range m.Workers {
		if w.ID != workerID {
			newWorkers = append(newWorkers, w)
		}
	}
	m.Workers = newWorkers

	log.Printf("Removed Worker %s from active connections", workerID)
}

// forwardWorkerRequestToBot forwards Worker request to Bot
func (m *Manager) forwardWorkerRequestToBot(worker *types.WorkerClient, action types.InternalAction, originalEcho string) {
	// Construct internal echo, including worker ID for tracking and recording RTT
	// Add timestamp to ensure internalEcho is unique even if originalEcho is empty or duplicated
	internalEcho := fmt.Sprintf("%s|%s|%d", worker.ID, originalEcho, time.Now().UnixNano())

	// Save request mapping
	respChan := make(chan types.InternalMessage, 1)
	m.PendingMutex.Lock()
	m.PendingRequests[internalEcho] = respChan
	m.PendingTimestamps[internalEcho] = time.Now() // Record send time
	m.PendingMutex.Unlock()

	// Update Worker RTT tracking (start)
	m.WorkerRequestMutex.Lock()
	m.WorkerRequestTimes[internalEcho] = time.Now()
	m.WorkerRequestMutex.Unlock()

	// Intelligent routing logic: select the correct Bot based on self_id or group_id
	var targetBot *types.BotClient
	routeSource := "specified" // For logging

	// 1. Try to extract self_id from action
	selfID := action.SelfID
	if selfID == "" {
		if sid, ok := action.Params["self_id"]; ok {
			selfID = utils.ToString(sid)
		}
	}

	m.Mutex.RLock()
	if selfID != "" {
		if bot, exists := m.Bots[selfID]; exists {
			targetBot = bot
		}
	}

	// 2. If no self_id, try to find the corresponding Bot from cache based on group_id
	if targetBot == nil {
		groupID := action.GroupID
		if groupID == "" {
			if gid, ok := action.Params["group_id"]; ok {
				groupID = utils.ToString(gid)
			}
		}

		if groupID != "" {
			m.CacheMutex.RLock()
			if groupData, exists := m.GroupCache[groupID]; exists {
				botID := groupData.BotID
				if botID != "" {
					if bot, exists := m.Bots[botID]; exists {
						targetBot = bot
						routeSource = "group_cache"
					}
				}
			}
			m.CacheMutex.RUnlock()
		}
	}

	// 3. Fallback scheme: if still not found, select the first available Bot
	if targetBot == nil {
		for _, bot := range m.Bots {
			targetBot = bot
			routeSource = "fallback (first available)"
			break
		}
	}
	m.Mutex.RUnlock()

	if targetBot != nil && routeSource != "specified" {
		log.Printf("[ROUTING] Routed Worker %s request to Bot %s via %s", worker.ID, targetBot.SelfID, routeSource)
	}

	if targetBot == nil {
		log.Printf("[ROUTING] [ERROR] No available Bot to handle Worker %s request (echo: %s, action: %v)", worker.ID, originalEcho, action.Action)

		// Return error response to Worker
		response := types.InternalMessage{
			Status:  "failed",
			Retcode: 1404,
			Msg:     "No Bot available",
			Echo:    originalEcho,
		}

		var finalResponse any
		if worker.Protocol == "v12" {
			finalResponse = response.ToV12Map()
		} else {
			finalResponse = response.ToV11Map()
		}

		worker.Mutex.Lock()
		worker.Conn.WriteJSON(finalResponse)
		worker.Mutex.Unlock()

		// Clean up mapping
		m.PendingMutex.Lock()
		delete(m.PendingRequests, internalEcho)
		delete(m.PendingTimestamps, internalEcho)
		m.PendingMutex.Unlock()
		return
	}

	// Forward request to Bot
	action.Echo = internalEcho

	// 统一构造消息广播，无论机器人是在线还是离线/真实机器人
	// 这样 WebUI 就能实时看到机器人发出的消息
	actionName := action.Action
	if actionName == "send_msg" || actionName == "send_group_msg" || actionName == "send_private_msg" {
		go func() {
			msgContent := utils.ToString(action.Message)
			if msgContent == "" {
				msgContent = utils.ToString(action.Params["message"])
			}

			userID := action.UserID
			if userID == "" {
				userID = utils.ToString(action.Params["user_id"])
			}

			groupID := action.GroupID
			if groupID == "" {
				groupID = utils.ToString(action.Params["group_id"])
			}

			// 模拟 OneBot v11 消息事件，用于 WebUI 显示
			event := map[string]any{
				"post_type":    "message",
				"message_type": action.MessageType,
				"sub_type":     "normal",
				"user_id":      targetBot.SelfID, // 发送者是机器人
				"target_id":    userID,           // 接收者是用户
				"group_id":     groupID,
				"message":      msgContent,
				"raw_message":  msgContent,
				"message_id":   fmt.Sprintf("msg_%d", time.Now().UnixNano()),
				"self_id":      targetBot.SelfID,
				"time":         time.Now().Unix(),
				"sender": map[string]any{
					"user_id":  targetBot.SelfID,
					"nickname": targetBot.Nickname,
					"role":     "bot",
				},
			}
			if action.MessageType == "" {
				if groupID != "" {
					event["message_type"] = "group"
				} else {
					event["message_type"] = "private"
				}
			}
			m.BroadcastEvent(event)
		}()
	}

	targetBot.Mutex.Lock()
	var err error
	if targetBot.Conn != nil {
		err = targetBot.Conn.WriteJSON(action)
	} else if targetBot.Platform == "Online" {
		// 对于在线机器人，模拟发送成功
		log.Printf("[OnlineBot] Worker %s reply: %s", worker.ID, action.Message)

		// 模拟响应给 Worker，表示消息已成功发送
		go func() {
			time.Sleep(10 * time.Millisecond)
			respChan <- types.InternalMessage{
				SelfID: targetBot.SelfID,
				Status: "ok",
				Extras: map[string]any{
					"status":  "ok",
					"retcode": 0,
					"data": map[string]any{
						"message_id": fmt.Sprintf("msg_%d", time.Now().UnixNano()),
					},
				},
				Echo: internalEcho,
			}
		}()
	} else {
		err = fmt.Errorf("bot connection is closed")
	}
	targetBot.Mutex.Unlock()

	// Broadcast outgoing routing events: Nexus -> Group -> User
	if err == nil {
		go func() {
			actionName := action.Action
			if actionName == "send_msg" || actionName == "send_group_msg" || actionName == "send_private_msg" {
				groupID := action.GroupID
				if groupID == "" {
					groupID = utils.ToString(action.Params["group_id"])
				}

				userID := action.UserID
				if userID == "" {
					userID = utils.ToString(action.Params["user_id"])
				}

				content := utils.ToString(action.Message)
				if content == "" {
					content = utils.ToString(action.Params["message"])
				}

				params := &types.RoutingParams{
					Content:  content,
					Platform: targetBot.Platform,
				}

				if groupID != "" {
					// Nexus -> Group
					m.CacheMutex.RLock()
					if group, ok := m.GroupCache[groupID]; ok {
						if name := group.GroupName; name != "" {
							params.GroupName = name
						}
					}
					m.CacheMutex.RUnlock()

					params.SourceType = "nexus"
					params.TargetType = "group"
					params.TargetLabel = params.GroupName
					m.BroadcastRoutingEvent("Nexus", groupID, "nexus_to_group", "response", params)

					// Group -> User
					if userID != "" {
						params.SourceType = "group"
						params.TargetType = "user"
						params.SourceLabel = params.GroupName
						m.BroadcastRoutingEvent(groupID, userID, "group_to_user", "response", params)
					}
				} else if userID != "" {
					// Nexus -> Bot -> User (Private)
					params.SourceType = "nexus"
					params.TargetType = "bot"
					params.TargetLabel = targetBot.Nickname
					m.BroadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "response", params)

					params.SourceType = "bot"
					params.TargetType = "user"
					params.SourceLabel = targetBot.Nickname
					m.BroadcastRoutingEvent(targetBot.SelfID, userID, "bot_to_user", "response", params)
				}
			}
		}()
	}

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward Worker %s request to Bot %s: %v. Attempting fallback...", worker.ID, targetBot.SelfID, err)

		// Remove failed Bot and try another one
		m.removeBot(targetBot.SelfID)

		// Simple retry logic: find another available Bot
		var fallbackBot *types.BotClient
		m.Mutex.RLock()
		for _, bot := range m.Bots {
			fallbackBot = bot
			break
		}
		m.Mutex.RUnlock()

		if fallbackBot != nil {
			log.Printf("[ROUTING] Falling back to Bot %s", fallbackBot.SelfID)
			fallbackBot.Mutex.Lock()
			err = fallbackBot.Conn.WriteJSON(action)
			fallbackBot.Mutex.Unlock()
			if err == nil {
				targetBot = fallbackBot
			}
		}

		if targetBot == nil || err != nil {
			// Return error response to Worker
			response := types.InternalMessage{
				Status:  "failed",
				Retcode: 1400,
				Msg:     "Failed to forward to any Bot",
				Echo:    originalEcho,
			}

			var finalResponse any
			if worker.Protocol == "v12" {
				finalResponse = response.ToV12Map()
			} else {
				finalResponse = response.ToV11Map()
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(finalResponse)
			worker.Mutex.Unlock()

			// 清理映射
			m.PendingMutex.Lock()
			delete(m.PendingRequests, internalEcho)
			delete(m.PendingTimestamps, internalEcho)
			m.PendingMutex.Unlock()
			return
		}
	}

	// Successfully sent to Bot, start processing response
	// Broadcast routing event: Nexus -> Bot (Request Forward)
	params := &types.RoutingParams{
		Platform: targetBot.Platform,
	}
	// Try to extract content from message for display
	if action.Params != nil {
		if content, ok := action.Params["message"]; ok {
			params.Content = utils.ToString(content)
		}
		if userID, ok := action.Params["user_id"]; ok {
			params.UserID = utils.ToString(userID)
		}
	}

	m.BroadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "request", params)

	// Update sending statistics (if it is a send message type operation)
	actionName = action.Action
	if actionName == "send_msg" || actionName == "send_private_msg" || actionName == "send_group_msg" || actionName == "send_guild_channel_msg" {
		if !isSystemAction(action) {
			m.UpdateBotSentStats(targetBot.SelfID)
		}
	}

	// If it is a send message, additionally broadcast an event from Bot to User to complete the visual effect loop
	if params.UserID != "" {
		m.BroadcastRoutingEvent(targetBot.SelfID, params.UserID, "bot_to_user", "message", params)
	}

	log.Printf("[ROUTING] Forwarded Worker %s request to Bot %s (Source: %s)", worker.ID, targetBot.SelfID, "dynamic")

	// Start timeout processing (must receive response within 30 seconds)
	go func() {
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		select {
		case response := <-respChan:
			// Received response, restore original echo and forward to Worker
			// Check response status, if Bot is found not in group, clear cache
			if response.Status == "failed" {
				// Try to extract group_id
				var groupID string
				if gid, ok := action.Params["group_id"]; ok {
					groupID = utils.ToString(gid)
				}

				if groupID != "" {
					// Check if error message contains "not in group" or similar hints
					respMsg := response.Msg
					retcode := response.Retcode

					if strings.Contains(respMsg, "not in group") || strings.Contains(respMsg, "removed") || retcode == 1200 {
						log.Printf("[ROUTING] Bot %s reported group error for %s: %s (retcode: %d). Clearing cache.",
							targetBot.SelfID, groupID, respMsg, retcode)

						m.CacheMutex.Lock()
						if data, exists := m.GroupCache[groupID]; exists {
							if data.BotID == targetBot.SelfID {
								delete(m.GroupCache, groupID)
								log.Printf("[ROUTING] Removed stale group cache for %s (Bot: %s)", groupID, targetBot.SelfID)
							}
						}
						m.CacheMutex.Unlock()

						// Asynchronously trigger refresh of Bot info and group list
						go m.fetchBotInfo(targetBot)
					}
				}
			}

			// Prepare final response for Worker
			var finalResponse any
			if worker.Protocol == "v12" {
				// For v12, we can send InternalMessage converted to map
				resMap := response.ToV12Map()
				resMap["echo"] = originalEcho
				finalResponse = resMap
			} else {
				// For v11, use ToV11Map
				resMap := response.ToV11Map()
				resMap["echo"] = originalEcho
				finalResponse = resMap
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(finalResponse)
			worker.Mutex.Unlock()
			log.Printf("Forwarded Bot response (echo: %s) to Worker %s", originalEcho, worker.ID)

		case <-timeout.C:
			// Timeout, return error response
			log.Printf("Worker request (echo: %s) timed out", originalEcho)

			response := types.InternalMessage{
				Status:  "failed",
				Retcode: 1401,
				Msg:     "Request timeout",
				Echo:    originalEcho,
			}

			var finalResponse any
			if worker.Protocol == "v12" {
				finalResponse = response.ToV12Map()
			} else {
				finalResponse = response.ToV11Map()
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(finalResponse)
			worker.Mutex.Unlock()
		}

		// Clean up mapping
		m.PendingMutex.Lock()
		delete(m.PendingRequests, internalEcho)
		delete(m.PendingTimestamps, internalEcho)
		m.PendingMutex.Unlock()
	}()
}

// cleanupPendingRequests cleans up expired request mappings
func (m *Manager) cleanupPendingRequests() {
	m.PendingMutex.Lock()
	defer m.PendingMutex.Unlock()

	// Clean up all pending requests (usually called when system shuts down)
	for echo, ch := range m.PendingRequests {
		close(ch)
		delete(m.PendingRequests, echo)
		log.Printf("Cleaned up pending request: %s", echo)
	}
}

// BroadcastEvent broadcasts any event to all subscribers
// BroadcastRoutingEvent broadcasts routing events to all subscribers
func (m *Manager) BroadcastRoutingEvent(source, target, direction, msgType string, params *types.RoutingParams) {
	event := types.RoutingEvent{
		Type:      "routing_event",
		Source:    source,
		Target:    target,
		Direction: direction,
		MsgType:   msgType,
		Timestamp: time.Now(),
	}

	if params != nil {
		event.SourceType = params.SourceType
		event.TargetType = params.TargetType
		event.SourceLabel = params.SourceLabel
		event.TargetLabel = params.TargetLabel
		event.UserID = params.UserID
		event.UserName = params.UserName
		event.UserAvatar = params.UserAvatar
		event.Content = params.Content
		event.Platform = params.Platform
		event.GroupID = params.GroupID
		event.GroupName = params.GroupName
	}

	m.BroadcastEvent(event)
}

// StartIdempotencyCleanup regularly cleans up local idempotency cache
func (m *Manager) StartIdempotencyCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := 0
			now := time.Now()
			m.LocalIdempotency.Range(func(key, value any) bool {
				if lastSeen, ok := value.(time.Time); ok {
					if now.Sub(lastSeen) > 1*time.Hour {
						m.LocalIdempotency.Delete(key)
						count++
					}
				}
				return true
			})
			if count > 0 {
				log.Printf("[REDIS] Cleaned up %d expired items from local idempotency cache", count)
			}
		}
	}
}

// CheckIdempotency checks message idempotency
func (m *Manager) CheckIdempotency(msgID string) bool {
	if msgID == "" {
		return true
	}

	// 1. Check local hot cache (fast interception)
	if lastSeen, ok := m.LocalIdempotency.Load(msgID); ok {
		if time.Since(lastSeen.(time.Time)) < 10*time.Minute {
			return false // Duplicate message
		}
	}

	if m.Rdb == nil {
		m.LocalIdempotency.Store(msgID, time.Now())
		return true
	}
	ctx := context.Background()
	key := fmt.Sprintf(config.REDIS_KEY_IDEMPOTENCY, msgID)

	// 2. Get dynamic TTL configuration from local cache
	ttl := 1 * time.Hour // Default 1 hour
	m.ConfigCacheMu.RLock()
	if val, ok := m.ConfigCache["ttl:idempotency_ttl_sec"]; ok {
		if ttlSec, err := strconv.ParseInt(val, 10, 64); err == nil && ttlSec > 0 {
			ttl = time.Duration(ttlSec) * time.Second
		}
	}
	m.ConfigCacheMu.RUnlock()

	// 3. Use SETNX to set expiration time
	ok, err := m.Rdb.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		log.Printf("[REDIS] Idempotency check error: %v", err)
		return true // Allow through when Redis fails to ensure availability
	}

	if ok {
		// Store in local hot cache
		m.LocalIdempotency.Store(msgID, time.Now())

		// Regularly clean up local cache (simple logic: clean up expired data every 1000 stores)
		// Actual production environment suggests using more rigorous cleanup goroutines
	}

	return ok
}

// CheckRateLimit checks rate limit (supports dynamic configuration)
func (m *Manager) CheckRateLimit(userID, groupID string) bool {
	if m.Rdb == nil {
		return true
	}
	ctx := context.Background()

	// 1. Get dynamic rate limit configuration from local cache
	// Default values: 20 per minute for users, 100 per minute for groups
	userLimit := int64(20)
	groupLimit := int64(100)

	m.ConfigCacheMu.RLock()
	if val, ok := m.ConfigCache["ratelimit:user_limit_per_min"]; ok {
		if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
			userLimit = limit
		}
	}
	if val, ok := m.ConfigCache["ratelimit:group_limit_per_min"]; ok {
		if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
			groupLimit = limit
		}
	}

	// Check individual/specific group override configuration
	if userID != "" {
		if val, ok := m.ConfigCache[fmt.Sprintf("ratelimit:user:%s:limit", userID)]; ok {
			if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
				userLimit = limit
			}
		}
	}
	if groupID != "" {
		if val, ok := m.ConfigCache[fmt.Sprintf("ratelimit:group:%s:limit", groupID)]; ok {
			if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
				groupLimit = limit
			}
		}
	}
	m.ConfigCacheMu.RUnlock()

	// 2. User level rate limit
	if userID != "" {
		userKey := fmt.Sprintf(config.REDIS_KEY_RATELIMIT_USER, userID)
		count, _ := m.Rdb.Incr(ctx, userKey).Result()
		if count == 1 {
			m.Rdb.Expire(ctx, userKey, 60*time.Second)
		}
		if count > userLimit {
			log.Printf("[RATELIMIT] User %s exceeded limit (%d/%d min)", userID, count, userLimit)
			return false
		}
	}

	// 3. Group level rate limit
	if groupID != "" {
		groupKey := fmt.Sprintf(config.REDIS_KEY_RATELIMIT_GROUP, groupID)
		count, _ := m.Rdb.Incr(ctx, groupKey).Result()
		if count == 1 {
			m.Rdb.Expire(ctx, groupKey, 60*time.Second)
		}
		if count > groupLimit {
			log.Printf("[RATELIMIT] Group %s exceeded limit (%d/%d min)", groupID, count, groupLimit)
			return false
		}
	}

	return true
}

// PushToRedisQueue 将消息推送到 Redis 队列 (支持重试)
func (m *Manager) PushToRedisQueue(targetWorkerID string, msg types.InternalMessage) error {
	if m.Rdb == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 转换为 map 格式以便 Worker 解析
	// 注意：这里默认转换为 v11 格式，因为目前大多数 Worker 还是基于 v11 协议
	// 如果未来有纯 v12 的 Worker，可以根据 targetWorkerID 的协议动态选择
	msgMap := msg.ToV11Map()

	data, err := json.Marshal(msgMap)
	if err != nil {
		return err
	}

	ctx := context.Background()
	queueKey := config.REDIS_KEY_QUEUE_DEFAULT
	if targetWorkerID != "" {
		queueKey = fmt.Sprintf(config.REDIS_KEY_QUEUE_WORKER, targetWorkerID)
	}

	// 使用指数退避重试策略 (最多 3 次)
	var lastErr error
	for i := 0; i < 3; i++ {
		// 使用 Redis Streams (XAdd) 代替 RPush，以匹配 Worker 的实现
		err = m.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: queueKey,
			Values: map[string]interface{}{
				"payload": string(data),
			},
		}).Err()

		if err == nil {
			log.Printf("[REDIS] Message pushed to stream %s (attempt %d)", queueKey, i+1)
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
	}

	return fmt.Errorf("failed to push to redis after 3 attempts: %v", lastErr)
}

// UpdateContext 更新会话上下文 (支持更丰富的状态存储)
func (m *Manager) UpdateContext(platform, userID string, msg types.InternalMessage) {
	if m.Rdb == nil || userID == "" || platform == "" {
		return
	}
	ctx := context.Background()
	key := fmt.Sprintf(config.REDIS_KEY_SESSION_CONTEXT, platform, userID)

	// 1. 获取动态 TTL 配置 (从本地缓存获取)
	ttl := 24 * time.Hour // 默认 24 小时
	m.ConfigCacheMu.RLock()
	if val, ok := m.ConfigCache["ttl:session_ttl_sec"]; ok {
		if ttlSec, err := strconv.ParseInt(val, 10, 64); err == nil && ttlSec > 0 {
			ttl = time.Duration(ttlSec) * time.Second
		}
	}
	m.ConfigCacheMu.RUnlock()

	// 2. 获取现有上下文 (优先从本地缓存获取，减少 Redis 读取)
	var sessionContext types.SessionContext
	cacheKey := fmt.Sprintf("%s:%s", platform, userID)
	if val, ok := m.SessionCache.Load(cacheKey); ok {
		sessionContext = val.(types.SessionContext)
	} else if val, err := m.Rdb.Get(ctx, key).Result(); err == nil {
		json.Unmarshal([]byte(val), &sessionContext)
	}

	// 3. 更新信息
	if sessionContext.CreatedAt.IsZero() {
		sessionContext.CreatedAt = time.Now()
	}
	sessionContext.UpdatedAt = time.Now()
	sessionContext.Platform = platform
	sessionContext.UserID = userID
	sessionContext.LastMsg = msg

	// 记录消息历史 (最近 5 条)
	sessionContext.History = append(sessionContext.History, msg)
	if len(sessionContext.History) > 5 {
		sessionContext.History = sessionContext.History[len(sessionContext.History)-5:]
	}

	// 4. 更新本地缓存
	m.SessionCache.Store(cacheKey, sessionContext)

	// 5. 异步写入 Redis (减少主流程延迟)
	go func() {
		data, _ := json.Marshal(sessionContext)
		m.Rdb.Set(context.Background(), key, data, ttl)
	}()
}

// GetSessionContext 获取会话上下文
func (m *Manager) GetSessionContext(platform, userID string) *types.SessionContext {
	if m.Rdb == nil || userID == "" || platform == "" {
		return nil
	}

	// 1. 优先从本地缓存获取
	cacheKey := fmt.Sprintf("%s:%s", platform, userID)
	if val, ok := m.SessionCache.Load(cacheKey); ok {
		sessionContext := val.(types.SessionContext)
		return &sessionContext
	}

	// 2. 缓存未命中，从 Redis 获取
	ctx := context.Background()
	key := fmt.Sprintf(config.REDIS_KEY_SESSION_CONTEXT, platform, userID)

	val, err := m.Rdb.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var sessionContext types.SessionContext
	if err := json.Unmarshal([]byte(val), &sessionContext); err != nil {
		return nil
	}

	// 存入本地缓存，方便下次使用
	m.SessionCache.Store(cacheKey, sessionContext)

	return &sessionContext
}

// SetSessionState 设置特定的会话状态 (例如：正在输入、等待确认等)
func (m *Manager) SetSessionState(platform, userID string, state types.SessionState, ttl time.Duration) error {
	if m.Rdb == nil || userID == "" || platform == "" {
		return fmt.Errorf("redis not initialized or invalid params")
	}
	ctx := context.Background()
	key := fmt.Sprintf("botmatrix:session:state:%s:%s", platform, userID)

	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now()
	}

	data, _ := json.Marshal(state)
	return m.Rdb.Set(ctx, key, data, ttl).Err()
}

// GetSessionState 获取特定的会话状态
func (m *Manager) GetSessionState(platform, userID string) *types.SessionState {
	if m.Rdb == nil || userID == "" || platform == "" {
		return nil
	}
	ctx := context.Background()
	key := fmt.Sprintf("botmatrix:session:state:%s:%s", platform, userID)

	val, err := m.Rdb.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var state types.SessionState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil
	}
	return &state
}
