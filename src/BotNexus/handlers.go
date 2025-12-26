package main

import (
	"BotMatrix/common"
	"BotMatrix/common/log"
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

	conn, err := upgrader.Upgrade(w, r, nil)
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
	bot := &common.BotClient{
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Platform:      platform,
		SelfID:        selfID,
	}

	// Register Bot
	m.Mutex.Lock()
	if m.Bots == nil {
		m.Bots = make(map[string]*common.BotClient)
	}
	m.Bots[bot.SelfID] = bot
	m.Mutex.Unlock()

	// Update connection stats
	m.ConnectionStats.Mutex.Lock()
	m.ConnectionStats.TotalBotConnections++
	m.ConnectionStats.LastBotActivity[bot.SelfID] = time.Now()
	m.ConnectionStats.Mutex.Unlock()

	log.Printf("Bot WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), bot.SelfID)

	// Fetch Bot info asynchronously
	go m.fetchBotInfo(bot)

	// Start connection handling loop
	go m.handleBotConnection(bot)
}

// fetchBotInfo actively fetches detailed info of the Bot
func (m *Manager) fetchBotInfo(bot *common.BotClient) {
	// Wait for 1 second to ensure connection is fully established and handshake is complete
	time.Sleep(1 * time.Second)

	log.Printf("[Bot] Fetching info for bot: %s", bot.SelfID)

	// 1. Get login info (nickname, etc.)
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

	// 2. Get group list (get group count)
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

	// Wait for response (with timeout)
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

				// Update group cache for subsequent routing API requests
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
							// Deep copy all fields
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
							// Persist to database
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

// handleBotConnection handles message loop for a single Bot connection
func (m *Manager) handleBotConnection(bot *common.BotClient) {
	// Start heartbeat sending goroutine
	stopHeartbeat := make(chan struct{})
	go m.sendBotHeartbeat(bot, stopHeartbeat)

	defer func() {
		close(stopHeartbeat) // Stop heartbeat
		// Cleanup work when connection is closed
		m.removeBot(bot.SelfID)
		bot.Conn.Close()

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

		// Handle message
		m.handleBotMessage(bot, msg)

		// Update activity time
		bot.LastHeartbeat = time.Now()
		m.ConnectionStats.Mutex.Lock()
		m.ConnectionStats.LastBotActivity[bot.SelfID] = time.Now()
		m.ConnectionStats.Mutex.Unlock()
	}
}

// sendBotHeartbeat sends heartbeat packets to Bot periodically
func (m *Manager) sendBotHeartbeat(bot *common.BotClient, stop chan struct{}) {
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
			log.Printf("Sent ping to Bot %s", bot.SelfID)

		case <-stop:
			return
		}
	}
}

// sendWorkerHeartbeat sends heartbeat packets to Worker periodically
func (m *Manager) sendWorkerHeartbeat(worker *common.WorkerClient, stop chan struct{}) {
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
			log.Printf("Sent ping to Worker %s", worker.ID)

		case <-stop:
			return
		}
	}
}

// handleBotMessage handles Bot messages
func (m *Manager) handleBotMessage(bot *common.BotClient, msg map[string]interface{}) {
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
	msgSelfID := common.ToString(msg["self_id"])
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
	echo := common.ToString(msg["echo"])
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
	msgType := common.ToString(msg["post_type"])
	if msgType == "" {
		log.Printf("[Bot] Warning: Received message from Bot %s without echo or post_type: %v", bot.SelfID, msg)
		return
	}

	// Update statistics
	m.Mutex.Lock()
	bot.RecvCount++
	m.Mutex.Unlock()

	// Only log non-spammy message types to console
	msgTypeLower := strings.ToLower(msgType)
	// Filter out common high-frequency events and anything containing "log"
	isLogType := strings.Contains(msgTypeLower, "log")
	if !isLogType && msgTypeLower != "meta_event" && msgTypeLower != "message" && msgTypeLower != "heartbeat" && !strings.Contains(msgTypeLower, "terminal") {
		log.Printf("Received %s event from Bot %s", msgType, bot.SelfID)
	}

	// Handle according to message type
	switch msgType {
	case "meta_event":
		// Meta event (heartbeat etc.)
		if heartbeat, ok := msg["meta_event_type"].(string); ok && heartbeat == "heartbeat" {
			// Heartbeat event, update status
			if status, ok := msg["status"].(map[string]interface{}); ok {
				if online, ok := status["online"].(bool); ok {
					// Only log heartbeat if debug is enabled or status changed
					if common.GlobalConfig.LogLevel == "DEBUG" {
						log.Printf("Bot %s heartbeat: online=%v", bot.SelfID, online)
					}
				}
			}
		}
	case "message", "notice", "request":
		// 3. Check idempotency (prevent duplicate replies)
		msgID := common.ToString(msg["message_id"])
		var userID, groupID string
		if msgID == "" {
			// notice/request may not have message_id, try using post_type + time + user_id as unique identifier
			msgID = fmt.Sprintf("%s:%v:%v", msgType, msg["time"], msg["user_id"])
		}

		if msgID != "" && !m.CheckIdempotency(msgID) {
			log.Printf("[REDIS] Duplicate %s detected: %s, skipping", msgType, msgID)
			return
		}

		// 4. Check rate limit (prevent spam and abuse)
		userID = common.ToString(msg["user_id"])
		groupID = common.ToString(msg["group_id"])
		if !m.CheckRateLimit(userID, groupID) {
			log.Printf("[REDIS] Rate limit exceeded for user %s / group %s", userID, groupID)
			return
		}

		// 5. Update session context (supports TTL)
		if userID != "" {
			m.UpdateContext(bot.Platform, userID, msg)
		}

		// 拦截器检查：在分发给 Worker 之前进行全局控制
		if m.TaskManager != nil {
			interceptorCtx := &tasks.InterceptorContext{
				Platform: bot.Platform,
				SelfID:   bot.SelfID,
				UserID:   userID,
				GroupID:  groupID,
				Event:    msg,
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
	m.enrichMessageWithCache(msg)

	if msgType == "message" {
		m.handleBotMessageEvent(bot, msg)
	}

	// Forward to Worker for processing
	// Ensure message format conforms to OneBot 11 standard (especially ID type)
	if userID, ok := msg["user_id"]; ok {
		msg["user_id"] = common.ToString(userID)
	}
	if groupID, ok := msg["group_id"]; ok {
		msg["group_id"] = common.ToString(groupID)
	}

	// Complete and standardize self_id (OneBot 11 standard should be int64, but we are compatible with string)
	if selfID, ok := msg["self_id"]; ok {
		msg["self_id"] = common.ToString(selfID)
	} else {
		// Complete self_id
		msg["self_id"] = bot.SelfID
	}

	// Complete time (int64 unix timestamp)
	if _, ok := msg["time"]; !ok {
		msg["time"] = time.Now().Unix()
	}

	// Complete post_type if missing
	if _, ok := msg["post_type"]; !ok {
		msg["post_type"] = "message"
	}

	// Complete platform
	if _, ok := msg["platform"]; !ok {
		msg["platform"] = bot.Platform
	}

	// 5.5 Broadcast to subscribers (Web UI Monitor)
	m.BroadcastEvent(msg)

	// 6. Use Redis queue for asynchronous decoupling
	// 如果启用了技能系统，则优先进入 Redis 队列进行异步分发
	if common.ENABLE_SKILL && m.Rdb != nil {
		targetWorkerID := m.getTargetWorkerID(msg)
		err := m.PushToRedisQueue(targetWorkerID, msg)
		if err == nil {
			m.BroadcastRoutingEvent("Nexus", "RedisQueue", "nexus_to_redis", "push", nil)

			// 影子执行 (Shadow Mode): 如果有影子 Worker，额外推送一份
			if shadowID, ok := msg["shadow_worker_id"].(string); ok && shadowID != "" {
				log.Printf("[Shadow] Pushing shadow copy to worker: %s", shadowID)
				// 克隆消息并打上影子标记
				shadowMsg := make(map[string]interface{})
				for k, v := range msg {
					shadowMsg[k] = v
				}
				shadowMsg["is_shadow"] = true
				m.PushToRedisQueue(shadowID, shadowMsg)
			}
			return
		}
		log.Printf("[REDIS] Failed to push to queue: %v. Falling back to WebSocket direct forwarding.", err)
	}

	// Fallback: If Redis is unavailable or push fails, use original WebSocket direct forwarding
	m.forwardMessageToWorker(msg)

	// 影子执行 (Shadow Mode) - Fallback 路径
	if shadowID, ok := msg["shadow_worker_id"].(string); ok && shadowID != "" {
		shadowMsg := make(map[string]interface{})
		for k, v := range msg {
			shadowMsg[k] = v
		}
		shadowMsg["is_shadow"] = true
		m.forwardMessageToWorkerWithTarget(shadowMsg, shadowID)
	}
}

// enrichMessageWithCache supplements the message with cached group and user information
func (m *Manager) enrichMessageWithCache(msg map[string]interface{}) {
	postType := common.ToString(msg["post_type"])
	if postType != "message" && postType != "notice" && postType != "request" {
		return
	}

	m.CacheMutex.RLock()
	defer m.CacheMutex.RUnlock()

	groupID := common.ToString(msg["group_id"])
	userID := common.ToString(msg["user_id"])

	// Ensure sender object exists for message type
	var sender map[string]interface{}
	if postType == "message" {
		if s, ok := msg["sender"].(map[string]interface{}); ok {
			sender = s
		} else {
			sender = make(map[string]interface{})
			msg["sender"] = sender
		}
	}

	// 1. Enrich Group Name
	if groupID != "" {
		if group, ok := m.GroupCache[groupID]; ok {
			if gName, ok := group["group_name"].(string); ok && gName != "" {
				// Only set if not already present or more descriptive than "Group ID"
				existingName := common.ToString(msg["group_name"])
				if existingName == "" || strings.Contains(existingName, groupID) {
					msg["group_name"] = gName
				}
			}
		}
	}

	// 2. Enrich User Nickname and Group Card
	if userID != "" {
		// Try member cache first (for group messages)
		if groupID != "" {
			memberKey := fmt.Sprintf("%s:%s", groupID, userID)
			if member, ok := m.MemberCache[memberKey]; ok {
				nickname := common.ToString(member["nickname"])
				card := common.ToString(member["card"])

				if sender != nil {
					if common.ToString(sender["nickname"]) == "" && nickname != "" {
						sender["nickname"] = nickname
					}
					if common.ToString(sender["card"]) == "" && card != "" {
						sender["card"] = card
					}
				}
				// Also add to top level if missing
				if common.ToString(msg["nickname"]) == "" && nickname != "" {
					msg["nickname"] = nickname
				}
			}
		}

		// Try friend cache (or fallback for nickname)
		if friend, ok := m.FriendCache[userID]; ok {
			nickname := common.ToString(friend["nickname"])
			if nickname != "" {
				if sender != nil && common.ToString(sender["nickname"]) == "" {
					sender["nickname"] = nickname
				}
				if common.ToString(msg["nickname"]) == "" {
					msg["nickname"] = nickname
				}
			}
		}
	}
}

// getTargetWorkerID helper method: Get target Worker ID based on routing rules
func (m *Manager) getTargetWorkerID(msg map[string]interface{}) string {
	// 0. 优先处理拦截器注入的语义路由提示 (Intelligent Semantic Routing)
	if hint, ok := msg["intent_hint"].(string); ok && hint != "" {
		log.Printf("[Routing] Using semantic intent hint: %s", hint)
		// 查找对应意图的 Worker (这里假设意图直接映射到 Worker ID 或标签)
		// 实际可建立 intent -> worker_id 的映射表
		return hint
	}

	var matchKeys []string

	// Extract match keys (logic same as forwardMessageToWorkerWithRetry)
	if userID, ok := msg["user_id"]; ok {
		sID := common.ToString(userID)
		if sID != "0" && sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("user_%s", sID), sID)
		}
	}
	if groupID, ok := msg["group_id"]; ok {
		sID := common.ToString(groupID)
		if sID != "0" && sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("group_%s", sID), sID)
		}
	}
	if selfID, ok := msg["self_id"]; ok {
		sID := common.ToString(selfID)
		if sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", sID), sID)
		}
	}

	// 1. Prioritize getting dynamic routing rules from Redis
	if m.Rdb != nil {
		ctx := context.Background()
		for _, key := range matchKeys {
			if wID, err := m.Rdb.HGet(ctx, common.REDIS_KEY_DYNAMIC_RULES, key).Result(); err == nil && wID != "" {
				log.Printf("[REDIS] Dynamic route matched (Redis): %s -> %s", key, wID)
				return wID
			}
		}

		// Check wildcard rules (Redis)
		// Note: HGetAll has performance impact when there are many rules, but in routing scenarios the number of rules is usually limited
		rules, err := m.Rdb.HGetAll(ctx, common.REDIS_KEY_DYNAMIC_RULES).Result()
		if err == nil && len(rules) > 0 {
			for p, w := range rules {
				if strings.Contains(p, "*") {
					for _, key := range matchKeys {
						if common.MatchRoutePattern(p, key) {
							log.Printf("[REDIS] Dynamic wildcard route matched (Redis): %s (%s) -> %s", p, key, w)
							return w
						}
					}
				}
			}
		}
	}

	// 2. Fallback to static routing rules in memory (Fail-open/Local cache)
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
				if common.MatchRoutePattern(p, key) {
					return w
				}
			}
		}
	}

	return ""
}

// cacheBotDataFromMessage extracts and caches data from messages (especially for Tencent channel bots)
func (m *Manager) cacheBotDataFromMessage(bot *common.BotClient, msg map[string]interface{}) {
	postType, _ := msg["post_type"].(string)
	if postType != "message" {
		return
	}

	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// Cache group info
	if groupIDVal, ok := msg["group_id"]; ok && groupIDVal != nil {
		gID := common.ToString(groupIDVal)

		// If group already exists in cache, keep original name, unless original name contains "(Cached)" and current message provides more accurate info
		existingGroup, exists := m.GroupCache[gID]
		groupName := ""
		if exists {
			if name, ok := existingGroup["group_name"].(string); ok {
				groupName = name
			}
		}

		// If no name or name contains "(Cached)", try setting a default value
		if groupName == "" || strings.Contains(groupName, "(Cached)") {
			// Only set (Cached) format when there is absolutely no name
			if groupName == "" {
				groupName = "Group " + gID
			}
		}

		// Update or add group cache
		m.GroupCache[gID] = map[string]interface{}{
			"group_id":   gID,
			"group_name": groupName,
			"bot_id":     bot.SelfID,
			"is_cached":  true,
			"reason":     "Automatically updated from message",
			"last_seen":  time.Now(),
		}
		// Persist to database
		go m.SaveGroupToDB(gID, groupName, bot.SelfID)

		// Cache member info
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
			// Persist to database
			go m.SaveMemberToDB(gID, uID, nickname, card)
		}
	} else if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
		// Cache friend info (private chat)
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
			// Persist to database
			go m.SaveFriendToDB(uID, nickname)
		}
	}
}

// isSystemMessage checks if it's a system message or meaningless statistical data
func isSystemMessage(msg map[string]interface{}) bool {
	// 1. Extract basic fields, supports direct structure and params nested structure
	userID := common.ToString(msg["user_id"])
	msgType := common.ToString(msg["message_type"])
	subType := common.ToString(msg["sub_type"])
	message := ""

	// Try to extract from params (if it's a send request)
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

	// Extract content
	if message == "" {
		if rm, ok := msg["raw_message"].(string); ok && rm != "" {
			message = rm
		} else if m, ok := msg["message"].(string); ok && m != "" {
			message = m
		}
	}

	// 2. Check common system user IDs
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

	// 3. Check message type and sub-type
	if msgType == "system" || subType == "system" {
		return true
	}

	// 4. Check if content is empty
	if message == "" {
		return true
	}

	return false
}

// sendBotMessage sends text message back to bot (helper method)
func (m *Manager) sendBotMessage(bot *common.BotClient, originalMsg map[string]interface{}, text string) {
	resp := map[string]interface{}{
		"action": "send_msg",
		"params": map[string]interface{}{
			"message": text,
		},
	}

	// Copy key fields from original message to determine destination
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

// handleBotMessageEvent handles Bot message events
func (m *Manager) handleBotMessageEvent(bot *common.BotClient, msg map[string]interface{}) {
	// Extract message info
	userID := common.ToString(msg["user_id"])
	groupID := common.ToString(msg["group_id"])

	// Prefer raw_message as it usually contains more complete text content
	message := ""
	if rm, ok := msg["raw_message"].(string); ok && rm != "" {
		message = rm
	} else if m, ok := msg["message"].(string); ok && m != "" {
		message = m
	}

	// Prepare broadcast extra info
	extras := map[string]interface{}{
		"user_id":  userID,
		"content":  message,
		"platform": bot.Platform,
	}

	// Extract sender info (nickname and avatar)
	if sender, ok := msg["sender"].(map[string]interface{}); ok {
		if nickname, ok := sender["nickname"].(string); ok && nickname != "" {
			extras["user_name"] = nickname
		}

		// Platform specific avatar logic
		switch strings.ToUpper(bot.Platform) {
		case "QQ":
			if userID != "" {
				extras["user_avatar"] = fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", userID)
			}
		case "WECHAT", "WX":
			// WeChat avatars usually cannot be spliced directly via ID, use default or get from sender
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			} else {
				extras["user_avatar"] = "/static/avatars/wechat_default.png"
			}
		case "TENCENT":
			// Tencent channel avatars are usually provided directly in sender
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			}
		default:
			// Other platforms try to get avatar from sender
			if avatar, ok := sender["avatar"].(string); ok && avatar != "" {
				extras["user_avatar"] = avatar
			}
		}
	}

	if groupID != "" {
		extras["group_id"] = groupID
		// Try to get group name from cache
		m.CacheMutex.RLock()
		if group, ok := m.GroupCache[groupID]; ok {
			if name, ok := group["group_name"].(string); ok {
				extras["group_name"] = name
			}
		}
		m.CacheMutex.RUnlock()
	}

	// 1. Broadcast routing event: Path based on group or private
	if groupID != "" {
		// Group path: User -> Group -> Nexus
		if userID != "" {
			extras["source_type"] = "user"
			extras["target_type"] = "group"
			extras["source_label"] = extras["user_name"]
			extras["target_label"] = extras["group_name"]
			m.BroadcastRoutingEvent(userID, groupID, "user_to_group", "message", extras)
		}

		// Group -> Nexus
		extras["source_id"] = groupID
		extras["source_type"] = "group"
		extras["target_type"] = "nexus"
		extras["source_label"] = extras["group_name"]
		m.BroadcastRoutingEvent(groupID, "Nexus", "group_to_nexus", "message", extras)
	} else {
		// Private path: User -> Bot -> Nexus
		if userID != "" {
			extras["source_type"] = "user"
			extras["target_type"] = "bot"
			extras["source_label"] = extras["user_name"]
			m.BroadcastRoutingEvent(userID, bot.SelfID, "user_to_bot", "message", extras)
		}

		// Bot -> Nexus
		extras["source_type"] = "bot"
		extras["target_type"] = "nexus"
		extras["source_label"] = bot.Nickname
		m.BroadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "message", extras)
	}

	// Update detailed stats (exclude system messages)
	if !isSystemMessage(msg) {
		m.UpdateBotStats(bot.SelfID, userID, groupID)
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
func (m *Manager) cacheMessage(msg map[string]interface{}) {
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
	return common.MatchRoutePattern(pattern, value)
}

// findWorkerByID finds a Worker by ID
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

// forwardMessageToWorker forwards the message to a Worker for processing
func (m *Manager) forwardMessageToWorker(msg map[string]interface{}) {
	m.forwardMessageToWorkerWithRetry(msg, 0)
}

// forwardMessageToWorkerWithTarget 直接将消息发送给指定的 Worker
func (m *Manager) forwardMessageToWorkerWithTarget(msg map[string]interface{}, targetWorkerID string) {
	if targetWorkerID == "" {
		return
	}

	// 确保消息有 echo
	if _, ok := msg["echo"]; !ok {
		msg["echo"] = fmt.Sprintf("shadow_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	}

	if w := m.findWorkerByID(targetWorkerID); w != nil {
		log.Printf("[ROUTING] Direct forwarding to Target Worker: %s", targetWorkerID)
		w.Mutex.Lock()
		err := w.Conn.WriteJSON(msg)
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
func (m *Manager) forwardMessageToWorkerWithRetry(msg map[string]interface{}, retryCount int) {
	if retryCount > 3 {
		log.Printf("[ROUTING] [ERROR] Maximum retry count exceeded for message. Caching message.")
		m.cacheMessage(msg)
		return
	}

	// 1. Try to use routing rules
	var targetWorkerID string
	var matchKeys []string

	// Extract matching keys
	// User ID match: user_123456
	if userID, ok := msg["user_id"]; ok {
		sID := common.ToString(userID)
		if sID != "0" && sID != "" {
			uID := fmt.Sprintf("user_%s", sID)
			matchKeys = append(matchKeys, uID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// Group ID match: group_789012
	if groupID, ok := msg["group_id"]; ok {
		sID := common.ToString(groupID)
		if sID != "0" && sID != "" {
			gID := fmt.Sprintf("group_%s", sID)
			matchKeys = append(matchKeys, gID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// Bot ID match: bot_123 or self_id
	if selfID, ok := msg["self_id"]; ok {
		sID := common.ToString(selfID)
		if sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", sID))
			matchKeys = append(matchKeys, sID)
		}
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
		// Since RoutingRules is a map, we cannot sort directly,
		// but common.MatchRoutePattern already handles wildcard logic.
		// For performance and simplicity, we iterate directly
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

	// Ensure the message has a unique echo for tracking processing time and matching replies
	if _, ok := msg["echo"]; !ok {
		msg["echo"] = fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	}

	// If a target Worker is found, try to get it
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

				// Broadcast routing event: Nexus -> Worker (Message Forward - Rule Match)
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
	// Filter out healthy workers
	var healthyWorkers []*common.WorkerClient
	for _, w := range m.Workers {
		// Simple check: if no heartbeat for 60 seconds and not just connected, it is considered unhealthy
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

	// Load balancing algorithm
	if len(healthyWorkers) == 1 {
		selectedWorker = healthyWorkers[0]
	} else {
		// 1. Prioritize choosing a Worker that has never processed a message (to get RTT)
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
			// 2. Prioritize choosing the Worker with the minimum AvgProcessTime (lowest load)
			var minProcessTime time.Duration = -1
			for _, w := range healthyWorkers {
				if w.AvgProcessTime > 0 {
					if minProcessTime == -1 || w.AvgProcessTime < minProcessTime {
						minProcessTime = w.AvgProcessTime
						selectedWorker = w
					}
				}
			}

			// 3. If AvgProcessTime has no data or is not selected, choose the Worker with the minimum AvgRTT
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

			// 4. Fall back to global round-robin
			if selectedWorker == nil {
				idx := m.WorkerIndex % len(healthyWorkers)
				selectedWorker = healthyWorkers[idx]
				m.WorkerIndex++
			}
		}
	}
	m.Mutex.Unlock()

	// Send the message to the selected Worker
	selectedWorker.Mutex.Lock()
	err := selectedWorker.Conn.WriteJSON(msg)
	selectedWorker.Mutex.Unlock()

	if err == nil {
		// Record the time the message was sent to the Worker, used to calculate processing time
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

		// Broadcast routing event: Nexus -> Worker (Message Forward - LB)
		m.BroadcastRoutingEvent("Nexus", selectedWorker.ID, "nexus_to_worker", "message", nil)

		log.Printf("[ROUTING] Forwarded to worker %s (AvgRTT: %v, Handled: %d)", selectedWorker.ID, selectedWorker.AvgRTT, selectedWorker.HandledCount)
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

	// Generate Worker ID
	workerID := conn.RemoteAddr().String()

	// Create Worker client - explicitly identified as a Worker connection
	worker := &common.WorkerClient{
		ID:            workerID,
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
	}

	log.Printf("Worker client created successfully: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// Register Worker
	m.Mutex.Lock()
	m.Workers = append(m.Workers, worker)
	m.Mutex.Unlock()

	// 广播 Worker 状态更新事件
	go m.BroadcastEvent(map[string]interface{}{
		"type": "worker_update",
		"data": map[string]interface{}{
			"id":            worker.ID,
			"remote_addr":   worker.ID,
			"connected":     worker.Connected.Format("2006-01-02 15:04:05"),
			"handled_count": worker.HandledCount,
			"avg_rtt":       worker.AvgRTT.String(),
			"last_rtt":      worker.LastRTT.String(),
			"is_alive":      true,
			"status":        "Online",
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
				})
				seen[cap.Name] = true
			}
		}
	}
	m.Mutex.RUnlock()

	m.TaskManager.AI.UpdateSkills(allSkills)
	log.Printf("[AI] Synced %d unique skills from %d workers", len(allSkills), len(m.Workers))
}

// handleWorkerConnection handles the message loop for a single Worker connection
func (m *Manager) handleWorkerConnection(worker *common.WorkerClient) {
	// Start heartbeat goroutine
	stopHeartbeat := make(chan struct{})
	go m.sendWorkerHeartbeat(worker, stopHeartbeat)

	defer func() {
		close(stopHeartbeat)
		// Cleanup work when connection is closed
		m.removeWorker(worker.ID)
		worker.Conn.Close()

		// 广播 Worker 离线事件
		go m.BroadcastEvent(map[string]interface{}{
			"type": "worker_update",
			"data": map[string]interface{}{
				"id":       worker.ID,
				"is_alive": false,
				"status":   "Offline",
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
		var msg map[string]interface{}
		err := common.ReadJSONWithNumber(worker.Conn, &msg)
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
func (m *Manager) handleWorkerMessage(worker *common.WorkerClient, msg map[string]interface{}) {
	// Only record key information, do not print full message
	msgType := common.ToString(msg["type"])

	// 处理 Worker 报备的能力列表
	if msgType == "register_capabilities" {
		if caps, ok := msg["capabilities"].([]interface{}); ok {
			var workerCaps []common.WorkerCapability
			capsData, _ := json.Marshal(caps)
			json.Unmarshal(capsData, &workerCaps)
			worker.Capabilities = workerCaps
			m.SyncWorkerSkills()
			log.Printf("[Worker] Worker %s registered %d capabilities", worker.ID, len(workerCaps))
			return
		}
	}

	action := common.ToString(msg["action"])
	echo := common.ToString(msg["echo"])

	if action != "" || (echo != "" && msgType == "") {
		// This is an API request initiated by a Worker, needs to be forwarded to the Bot
		log.Printf("Worker %s API request: action=%s, echo=%s", worker.ID, action, echo)

		// Broadcast routing event: Worker -> Nexus (Request)
		m.BroadcastRoutingEvent(worker.ID, "Nexus", "worker_to_nexus", "request", nil)

		m.forwardWorkerRequestToBot(worker, msg, echo)
	} else if msgType == "skill_result" {
		if common.ENABLE_SKILL {
			// 处理通过 WebSocket 上报的技能执行结果
			m.HandleSkillResult(msg)
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
				go m.BroadcastEvent(map[string]interface{}{
					"type": "worker_update",
					"data": map[string]interface{}{
						"id":                worker.ID,
						"handled_count":     worker.HandledCount,
						"last_process_time": worker.LastProcessTime.String(),
						"avg_process_time":  worker.AvgProcessTime.String(),
						"status":            "Online",
					},
				})
			} else {
				m.WorkerRequestMutex.Unlock()
			}
		}

		// 2. Check if it contains a reply (some frameworks allow direct reply in event response)
		reply := common.ToString(msg["reply"])
		if reply != "" {
			log.Printf("Worker %s sent passive reply: %s", worker.ID, reply)
			// Construct a send_msg request and forward it to the Bot
			m.handleWorkerPassiveReply(worker, msg)
		}
	}
}

// handleWorkerPassiveReply handles passive replies from Workers
func (m *Manager) handleWorkerPassiveReply(worker *common.WorkerClient, msg map[string]interface{}) {
	// Extract echo (if the Worker included echo in the passive reply)
	echo := common.ToString(msg["echo"])

	// Construct OneBot 11 send_msg request
	params := make(map[string]interface{})
	action := map[string]interface{}{
		"action": "send_msg",
		"params": params,
	}

	// Iterate and forward all fields returned by the Worker
	for k, v := range msg {
		switch k {
		case "reply":
			params["message"] = v
		case "action", "echo":
			// Ignore, action is fixed to send_msg, echo is extracted separately
			continue
		case "self_id", "platform":
			// Routing key fields, put in top level and params
			action[k] = v
			params[k] = v
		default:
			// Other fields pass through to params (group_id, user_id, message_type, auto_escape, etc.)
			params[k] = v
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
	newWorkers := make([]*common.WorkerClient, 0, len(m.Workers))
	for _, w := range m.Workers {
		if w.ID != workerID {
			newWorkers = append(newWorkers, w)
		}
	}
	m.Workers = newWorkers

	log.Printf("Removed Worker %s from active connections", workerID)
}

// forwardWorkerRequestToBot forwards Worker request to Bot
func (m *Manager) forwardWorkerRequestToBot(worker *common.WorkerClient, msg map[string]interface{}, originalEcho string) {
	// Construct internal echo, including worker ID for tracking and recording RTT
	// Add timestamp to ensure internalEcho is unique even if originalEcho is empty or duplicated
	internalEcho := fmt.Sprintf("%s|%s|%d", worker.ID, originalEcho, time.Now().UnixNano())

	// Save request mapping
	respChan := make(chan map[string]interface{}, 1)
	m.PendingMutex.Lock()
	m.PendingRequests[internalEcho] = respChan
	m.PendingTimestamps[internalEcho] = time.Now() // Record send time
	m.PendingMutex.Unlock()

	// Modify echo in message to internal echo
	msg["echo"] = internalEcho

	// Intelligent routing logic: select the correct Bot based on self_id or group_id
	var targetBot *common.BotClient
	var routeSource string

	// 1. Try to extract self_id from request parameters
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

	// 2. If no self_id, try to find the corresponding Bot from cache based on group_id
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

	// 3. Fallback scheme: if still not found, select the first available Bot
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

		// Return error response to Worker
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

		// Clean up mapping
		m.PendingMutex.Lock()
		delete(m.PendingRequests, internalEcho)
		delete(m.PendingTimestamps, internalEcho)
		m.PendingMutex.Unlock()
		return
	}

	// Forward request to Bot
	targetBot.Mutex.Lock()
	err := targetBot.Conn.WriteJSON(msg)
	targetBot.Mutex.Unlock()

	// Broadcast outgoing routing events: Nexus -> Group -> User
	if err == nil {
		go func() {
			action := common.ToString(msg["action"])
			if action == "send_msg" || action == "send_group_msg" || action == "send_private_msg" {
				params, _ := msg["params"].(map[string]interface{})
				if params == nil {
					params = msg
				}

				groupID := common.ToString(params["group_id"])
				userID := common.ToString(params["user_id"])
				content := common.ToString(params["message"])

				extras := map[string]interface{}{
					"content":  content,
					"platform": targetBot.Platform,
				}

				if groupID != "" {
					// Nexus -> Group
					m.CacheMutex.RLock()
					if group, ok := m.GroupCache[groupID]; ok {
						if name, ok := group["group_name"].(string); ok {
							extras["group_name"] = name
						}
					}
					m.CacheMutex.RUnlock()

					extras["source_type"] = "nexus"
					extras["target_type"] = "group"
					extras["target_label"] = extras["group_name"]
					m.BroadcastRoutingEvent("Nexus", groupID, "nexus_to_group", "response", extras)

					// Group -> User
					if userID != "" {
						extras["source_id"] = groupID
						extras["source_type"] = "group"
						extras["target_type"] = "user"
						extras["source_label"] = extras["group_name"]
						m.BroadcastRoutingEvent(groupID, userID, "group_to_user", "response", extras)
					}
				} else if userID != "" {
					// Nexus -> Bot -> User (Private)
					extras["source_type"] = "nexus"
					extras["target_type"] = "bot"
					extras["target_label"] = targetBot.Nickname
					m.BroadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "response", extras)

					extras["source_type"] = "bot"
					extras["target_type"] = "user"
					extras["source_label"] = targetBot.Nickname
					m.BroadcastRoutingEvent(targetBot.SelfID, userID, "bot_to_user", "response", extras)
				}
			}
		}()
	}

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward Worker %s request to Bot %s: %v. Attempting fallback...", worker.ID, targetBot.SelfID, err)

		// Remove failed Bot and try another one
		m.removeBot(targetBot.SelfID)

		// Simple retry logic: find another available Bot
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
			// Return error response to Worker
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

	// Successfully sent to Bot, start processing response
	// Broadcast routing event: Nexus -> Bot (Request Forward)
	extras := map[string]interface{}{
		"platform": targetBot.Platform,
	}
	// Try to extract content from message for display
	if params, ok := msg["params"].(map[string]interface{}); ok {
		if content, ok := params["message"]; ok {
			extras["content"] = common.ToString(content)
		}
		if userID, ok := params["user_id"]; ok {
			extras["user_id"] = common.ToString(userID)
		}
	}

	m.BroadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "request", extras)

	// Update sending statistics (if it is a send message type operation)
	action := common.ToString(msg["action"])
	if action == "send_msg" || action == "send_private_msg" || action == "send_group_msg" || action == "send_guild_channel_msg" {
		// Add check to exclude system messages
		if !isSystemMessage(msg) {
			m.UpdateBotSentStats(targetBot.SelfID)
		}
	}

	// If it is a send message, additionally broadcast an event from Bot to User to complete the visual effect loop
	if userID, ok := extras["user_id"].(string); ok && userID != "" {
		m.BroadcastRoutingEvent(targetBot.SelfID, userID, "bot_to_user", "message", extras)
	}

	log.Printf("[ROUTING] Forwarded Worker %s request to Bot %s (Source: %s)", worker.ID, targetBot.SelfID, routeSource)

	// Start timeout processing (must receive response within 30 seconds)
	go func() {
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		select {
		case response := <-respChan:
			// Received response, restore original echo and forward to Worker
			if response != nil {
				// Check response status, if Bot is found not in group, clear cache
				retcode := common.ToInt64(response["retcode"])
				status, _ := response["status"].(string)

				if retcode == 1200 || status == "failed" {
					// Try to extract group_id
					var groupID string
					if gid, ok := msg["group_id"]; ok {
						groupID = common.ToString(gid)
					} else if params, ok := msg["params"].(map[string]interface{}); ok {
						if gid, ok := params["group_id"]; ok {
							groupID = common.ToString(gid)
						}
					}

					if groupID != "" {
						// Check if error message contains "not in group" or similar hints
						respMsg, _ := response["msg"].(string)
						if strings.Contains(respMsg, "not in group") || strings.Contains(respMsg, "removed") || retcode == 1200 {
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

							// Asynchronously trigger refresh of Bot info and group list
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
			// Timeout, return error response
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
			m.LocalIdempotency.Range(func(key, value interface{}) bool {
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
	key := fmt.Sprintf(common.REDIS_KEY_IDEMPOTENCY, msgID)

	// 2. Get dynamic TTL configuration
	ttl := 1 * time.Hour // Default 1 hour
	if val, err := m.Rdb.HGet(ctx, common.REDIS_KEY_CONFIG_TTL, "idempotency_ttl_sec").Result(); err == nil {
		if ttlSec, err := strconv.ParseInt(val, 10, 64); err == nil && ttlSec > 0 {
			ttl = time.Duration(ttlSec) * time.Second
		}
	}

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

	// 1. Get dynamic rate limit configuration (get from Redis Hash, if not exists use default values)
	// Default values: 20 per minute for users, 100 per minute for groups
	userLimit := int64(20)
	groupLimit := int64(100)

	// Try to get configuration from Redis (for performance, actual production environment suggests local cache + expiration sync)
	config, err := m.Rdb.HGetAll(ctx, common.REDIS_KEY_CONFIG_RATELIMIT).Result()
	if err == nil && len(config) > 0 {
		if val, ok := config["user_limit_per_min"]; ok {
			if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
				userLimit = limit
			}
		}
		if val, ok := config["group_limit_per_min"]; ok {
			if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
				groupLimit = limit
			}
		}

		// Check individual/specific group override configuration
		if userID != "" {
			if val, ok := config[fmt.Sprintf("user:%s:limit", userID)]; ok {
				if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
					userLimit = limit
				}
			}
		}
		if groupID != "" {
			if val, ok := config[fmt.Sprintf("group:%s:limit", groupID)]; ok {
				if limit, err := strconv.ParseInt(val, 10, 64); err == nil {
					groupLimit = limit
				}
			}
		}
	}

	// 2. User level rate limit
	if userID != "" {
		userKey := fmt.Sprintf(common.REDIS_KEY_RATELIMIT_USER, userID)
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
		groupKey := fmt.Sprintf(common.REDIS_KEY_RATELIMIT_GROUP, groupID)
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
func (m *Manager) PushToRedisQueue(targetWorkerID string, msg map[string]interface{}) error {
	if m.Rdb == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	queueKey := common.REDIS_KEY_QUEUE_DEFAULT
	if targetWorkerID != "" {
		queueKey = fmt.Sprintf(common.REDIS_KEY_QUEUE_WORKER, targetWorkerID)
	}

	// 使用指数退避重试策略 (最多 3 次)
	var lastErr error
	for i := 0; i < 3; i++ {
		err = m.Rdb.RPush(ctx, queueKey, data).Err()
		if err == nil {
			log.Printf("[REDIS] Message pushed to queue %s (attempt %d)", queueKey, i+1)
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
	}

	return fmt.Errorf("failed to push to redis after 3 attempts: %v", lastErr)
}

// UpdateContext 更新会话上下文 (支持更丰富的状态存储)
func (m *Manager) UpdateContext(platform, userID string, msg map[string]interface{}) {
	if m.Rdb == nil || userID == "" || platform == "" {
		return
	}
	ctx := context.Background()
	key := fmt.Sprintf(common.REDIS_KEY_SESSION_CONTEXT, platform, userID)

	// 1. 获取动态 TTL 配置
	ttl := 24 * time.Hour // 默认 24 小时
	if val, err := m.Rdb.HGet(ctx, common.REDIS_KEY_CONFIG_TTL, "session_ttl_sec").Result(); err == nil {
		if ttlSec, err := strconv.ParseInt(val, 10, 64); err == nil && ttlSec > 0 {
			ttl = time.Duration(ttlSec) * time.Second
		}
	}

	// 2. 获取现有上下文 (如果存在)
	var contextData map[string]interface{}
	if val, err := m.Rdb.Get(ctx, key).Result(); err == nil {
		json.Unmarshal([]byte(val), &contextData)
	}
	if contextData == nil {
		contextData = make(map[string]interface{})
	}

	// 3. 更新基本信息
	contextData["last_msg"] = msg["message"]
	contextData["last_time"] = time.Now().Unix()
	contextData["platform"] = platform

	// 记录消息历史 (最近 5 条)
	var history []interface{}
	if h, ok := contextData["history"].([]interface{}); ok {
		history = h
	}
	history = append(history, msg["message"])
	if len(history) > 5 {
		history = history[len(history)-5:]
	}
	contextData["history"] = history

	data, _ := json.Marshal(contextData)

	// 4. 设置过期时间
	m.Rdb.Set(ctx, key, data, ttl)
}

// GetSessionContext 获取会话上下文
func (m *Manager) GetSessionContext(platform, userID string) map[string]interface{} {
	if m.Rdb == nil || userID == "" || platform == "" {
		return nil
	}
	ctx := context.Background()
	key := fmt.Sprintf(common.REDIS_KEY_SESSION_CONTEXT, platform, userID)

	val, err := m.Rdb.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &contextData); err != nil {
		return nil
	}
	return contextData
}

// SetSessionState 设置特定的会话状态 (例如：正在输入、等待确认等)
func (m *Manager) SetSessionState(platform, userID string, state map[string]interface{}, ttl time.Duration) error {
	if m.Rdb == nil || userID == "" || platform == "" {
		return fmt.Errorf("redis not initialized or invalid params")
	}
	ctx := context.Background()
	key := fmt.Sprintf("botmatrix:session:state:%s:%s", platform, userID)

	data, _ := json.Marshal(state)
	return m.Rdb.Set(ctx, key, data, ttl).Err()
}

// GetSessionState 获取特定的会话状态
func (m *Manager) GetSessionState(platform, userID string) map[string]interface{} {
	if m.Rdb == nil || userID == "" || platform == "" {
		return nil
	}
	ctx := context.Background()
	key := fmt.Sprintf("botmatrix:session:state:%s:%s", platform, userID)

	val, err := m.Rdb.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var state map[string]interface{}
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil
	}
	return state
}
