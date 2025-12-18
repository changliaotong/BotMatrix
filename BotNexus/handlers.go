package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// hashPassword 对密码进行哈希处理
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword 验证密码是否匹配
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// saveUsersToRedis 将用户数据保存到Redis
func (m *Manager) saveUsersToRedis() error {
	if m.rdb == nil {
		return nil // Redis未连接，跳过保存
	}

	// 锁定用户存储
	m.usersMutex.RLock()
	defer m.usersMutex.RUnlock()

	// 将用户数据转换为JSON
	usersJSON, err := json.Marshal(m.users)
	if err != nil {
		return err
	}

	// 保存到Redis，过期时间设为7天
	err = m.rdb.Set(context.Background(), "botnexus:users", usersJSON, 7*24*time.Hour).Err()
	if err != nil {
		return err
	}

	log.Printf("用户数据已保存到Redis，共 %d 个用户", len(m.users))
	return nil
}

// loadUsersFromRedis 从Redis加载用户数据
func (m *Manager) loadUsersFromRedis() error {
	if m.rdb == nil {
		return nil // Redis未连接，跳过加载
	}

	// 从Redis获取用户数据
	usersJSON, err := m.rdb.Get(context.Background(), "botnexus:users").Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Redis中无用户数据，将使用默认配置")
			return nil // Redis中没有用户数据，正常返回
		}
		return err
	}

	// 解析JSON到用户映射
	users := make(map[string]*User)
	err = json.Unmarshal([]byte(usersJSON), &users)
	if err != nil {
		return err
	}

	// 验证用户数据完整性
	validUsers := make(map[string]*User)
	for username, user := range users {
		// 检查必要字段
		if user.Username == "" || user.PasswordHash == "" {
			log.Printf("从Redis加载的用户数据不完整，忽略用户: %s", username)
			continue
		}
		validUsers[username] = user
	}

	// 锁定并更新用户存储
	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	// 更新用户数据
	m.users = validUsers
	log.Printf("从Redis加载用户数据成功，共 %d 个有效用户", len(m.users))

	// 如果没有有效用户，初始化默认管理员
	if len(m.users) == 0 {
		log.Printf("Redis中没有有效用户，将重新初始化默认管理员")
		m.initDefaultAdmin()
	}

	return nil
}

// initDefaultAdmin 初始化默认管理员用户
// 注意：调用此函数时需要确保已经持有usersMutex锁
func (m *Manager) initDefaultAdmin() {
	// 检查是否已存在管理员用户
	if _, exists := m.users["admin"]; exists {
		return // 已存在，无需初始化
	}

	// 对默认密码进行哈希处理
	hashedPassword, err := hashPassword(DEFAULT_ADMIN_PASSWORD)
	if err != nil {
		log.Printf("初始化默认管理员密码哈希失败: %v", err)
		return
	}

	// 创建默认管理员用户
	adminUser := &User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hashedPassword,
		IsAdmin:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 存储到用户映射
	m.users["admin"] = adminUser
	log.Printf("默认管理员用户已初始化，用户名: admin, 密码: %s", DEFAULT_ADMIN_PASSWORD)

	// 保存到Redis
	go func() {
		if err := m.saveUsersToRedis(); err != nil {
			log.Printf("保存默认管理员用户到Redis失败: %v", err)
		}
	}()
}

// generateToken 生成JWT token
func (m *Manager) generateToken(user *User) (string, error) {
	// 设置token的声明
	claims := UserClaims{
		UserID:         user.ID,
		Username:       user.Username,
		IsAdmin:        user.IsAdmin,
		SessionVersion: user.SessionVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用JWT_SECRET创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWT_SECRET))

	return tokenString, err
}

// handleLogin 处理登录请求
func (m *Manager) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("收到登录请求，客户端地址: %s", r.RemoteAddr)

	// 解析请求体
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		log.Printf("登录请求解析失败: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	log.Printf("登录尝试 - 用户名: %s", loginData.Username)

	// 验证用户名和密码
	m.usersMutex.RLock()
	user, exists := m.users[loginData.Username]
	m.usersMutex.RUnlock()

	if !exists {
		log.Printf("登录失败 - 用户不存在: %s", loginData.Username)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名或密码错误",
		})
		return
	}

	if !checkPassword(loginData.Password, user.PasswordHash) {
		log.Printf("登录失败 - 密码错误: %s", loginData.Username)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成JWT token
	token, err := m.generateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Token生成失败",
		})
		return
	}

	// 返回成功响应
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"role":    user.IsAdmin,
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// handleGetStats 处理获取统计信息的请求
func (m *Manager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取连接统计
	m.connectionStats.Mutex.Lock()
	botDurations := make(map[string]string)
	workerDurations := make(map[string]string)

	for k, v := range m.connectionStats.BotConnectionDurations {
		botDurations[k] = v.String()
	}
	for k, v := range m.connectionStats.WorkerConnectionDurations {
		workerDurations[k] = v.String()
	}

	stats := map[string]interface{}{
		"bots": map[string]interface{}{
			"count":       len(m.bots),
			"durations":   botDurations,
			"disconnects": m.connectionStats.BotDisconnectReasons,
		},
		"workers": map[string]interface{}{
			"count":       len(m.workers),
			"durations":   workerDurations,
			"disconnects": m.connectionStats.WorkerDisconnectReasons,
		},
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	m.connectionStats.Mutex.Unlock()

	json.NewEncoder(w).Encode(stats)
}

// handleGetLogs 处理获取日志的请求
func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": []string{"日志功能开发中..."},
	})
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

	// 创建Bot客户端
	bot := &BotClient{
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Platform:      "qq", // 默认平台为QQ
	}

	// 生成Bot ID（使用连接地址作为临时ID）
	bot.SelfID = conn.RemoteAddr().String()

	// 注册Bot
	m.mutex.Lock()
	if m.bots == nil {
		m.bots = make(map[string]*BotClient)
	}
	m.bots[bot.SelfID] = bot
	m.mutex.Unlock()

	// 更新连接统计
	m.connectionStats.Mutex.Lock()
	m.connectionStats.TotalBotConnections++
	m.connectionStats.LastBotActivity[bot.SelfID] = time.Now()
	m.connectionStats.Mutex.Unlock()

	log.Printf("Bot WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), bot.SelfID)

	// 启动连接处理循环
	go m.handleBotConnection(bot)
}

// handleBotConnection 处理单个Bot连接的消息循环
func (m *Manager) handleBotConnection(bot *BotClient) {
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
		m.connectionStats.Mutex.Lock()
		m.connectionStats.BotConnectionDurations[bot.SelfID] = duration
		if m.connectionStats.BotDisconnectReasons == nil {
			m.connectionStats.BotDisconnectReasons = make(map[string]int64)
		}
		m.connectionStats.BotDisconnectReasons["connection_closed"]++
		m.connectionStats.Mutex.Unlock()

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
		err := bot.Conn.ReadJSON(&msg)
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
		m.connectionStats.Mutex.Lock()
		m.connectionStats.LastBotActivity[bot.SelfID] = time.Now()
		m.connectionStats.Mutex.Unlock()
	}
}

// sendBotHeartbeat 定期发送心跳包给Bot
func (m *Manager) sendBotHeartbeat(bot *BotClient, stop chan struct{}) {
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

// handleBotMessage 处理Bot消息
func (m *Manager) handleBotMessage(bot *BotClient, msg map[string]interface{}) {
	// 检查是否是API响应（有echo字段）
	if echo, ok := msg["echo"].(string); ok && echo != "" {
		// 这是API响应，需要回传给对应的Worker
		m.pendingMutex.Lock()
		respChan, exists := m.pendingRequests[echo]
		m.pendingMutex.Unlock()

		if exists {
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
	msgType, ok := msg["post_type"].(string)
	if !ok {
		return
	}

	log.Printf("Received %s message from Bot %s", msgType, bot.SelfID)

	// 更新统计信息
	m.mutex.Lock()
	bot.RecvCount++
	m.mutex.Unlock()

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
	m.forwardMessageToWorker(msg)
}

// handleBotMessageEvent 处理Bot消息事件
func (m *Manager) handleBotMessageEvent(bot *BotClient, msg map[string]interface{}) {
	// 提取消息信息
	userID, _ := msg["user_id"].(float64)
	groupID, _ := msg["group_id"].(float64)
	message, _ := msg["message"].(string)

	log.Printf("Bot %s received message from user %d: %s", bot.SelfID, int64(userID), message)

	// 更新详细统计
	m.updateBotStats(bot.SelfID, int64(userID), int64(groupID))
}

// removeBot 移除Bot连接
func (m *Manager) removeBot(botID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.bots[botID]; exists {
		delete(m.bots, botID)
		log.Printf("Removed Bot %s from active connections", botID)
	}
}

// forwardMessageToWorker 将消息转发给Worker处理
func (m *Manager) forwardMessageToWorker(msg map[string]interface{}) {
	m.mutex.RLock()
	workers := make([]*WorkerClient, len(m.workers))
	copy(workers, m.workers)
	m.mutex.RUnlock()

	if len(workers) == 0 {
		log.Printf("No workers available to handle message")
		return
	}

	// 简单的轮询选择Worker
	m.mutex.Lock()
	if m.workerIndex >= len(workers) {
		m.workerIndex = 0
	}
	worker := workers[m.workerIndex]
	m.workerIndex++
	m.mutex.Unlock()

	// 发送消息给选中的Worker
	worker.Mutex.Lock()
	err := worker.Conn.WriteJSON(msg)
	worker.Mutex.Unlock()

	if err != nil {
		log.Printf("Failed to forward message to worker %s: %v", worker.ID, err)
	} else {
		worker.HandledCount++
		log.Printf("Forwarded message to worker %s", worker.ID)
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
	worker := &WorkerClient{
		ID:            workerID,
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
	}

	log.Printf("Worker client created successfully: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// 注册Worker
	m.mutex.Lock()
	m.workers = append(m.workers, worker)
	m.mutex.Unlock()

	// 更新连接统计
	m.connectionStats.Mutex.Lock()
	m.connectionStats.TotalWorkerConnections++
	if m.connectionStats.LastWorkerActivity == nil {
		m.connectionStats.LastWorkerActivity = make(map[string]time.Time)
	}
	m.connectionStats.LastWorkerActivity[workerID] = time.Now()
	m.connectionStats.Mutex.Unlock()

	log.Printf("Worker WebSocket connected: %s (ID: %s)", conn.RemoteAddr(), workerID)

	// 启动连接处理循环
	go m.handleWorkerConnection(worker)
}

// handleWorkerConnection 处理单个Worker连接的消息循环
func (m *Manager) handleWorkerConnection(worker *WorkerClient) {
	defer func() {
		// 连接关闭时的清理工作
		m.removeWorker(worker.ID)
		worker.Conn.Close()

		// 记录断开连接
		duration := time.Since(worker.Connected)
		m.connectionStats.Mutex.Lock()
		if m.connectionStats.WorkerConnectionDurations == nil {
			m.connectionStats.WorkerConnectionDurations = make(map[string]time.Duration)
		}
		m.connectionStats.WorkerConnectionDurations[worker.ID] = duration
		if m.connectionStats.WorkerDisconnectReasons == nil {
			m.connectionStats.WorkerDisconnectReasons = make(map[string]int64)
		}
		m.connectionStats.WorkerDisconnectReasons["connection_closed"]++
		m.connectionStats.Mutex.Unlock()

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
		err := worker.Conn.ReadJSON(&msg)
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
		m.connectionStats.Mutex.Lock()
		m.connectionStats.LastWorkerActivity[worker.ID] = time.Now()
		m.connectionStats.Mutex.Unlock()
	}
}

// handleWorkerMessage 处理Worker消息
func (m *Manager) handleWorkerMessage(worker *WorkerClient, msg map[string]interface{}) {
	// 只记录关键信息，不打印完整消息
	msgType, _ := msg["type"].(string)
	echo, hasEcho := msg["echo"].(string)

	if hasEcho {
		log.Printf("Worker %s request: type=%s, echo=%s", worker.ID, msgType, echo)

		// 这是一个Worker发起的API请求，需要转发给Bot
		m.forwardWorkerRequestToBot(worker, msg, echo)
	} else {
		log.Printf("Worker %s response: type=%s", worker.ID, msgType)

		// 更新统计信息
		m.mutex.Lock()
		worker.HandledCount++
		m.mutex.Unlock()

		// 这里可以处理Worker的响应消息
		// 例如：将响应转发回对应的Bot，或者处理业务逻辑
	}
}

// removeWorker 移除Worker连接
func (m *Manager) removeWorker(workerID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 从workers数组中移除
	newWorkers := make([]*WorkerClient, 0, len(m.workers))
	for _, w := range m.workers {
		if w.ID != workerID {
			newWorkers = append(newWorkers, w)
		}
	}
	m.workers = newWorkers

	log.Printf("Removed Worker %s from active connections", workerID)
}

// handleSubscriberWebSocket 处理Subscriber WebSocket连接
func (m *Manager) handleSubscriberWebSocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Subscriber WebSocket upgrade failed: %v", err)
		return
	}

	// 注册Subscriber
	m.usersMutex.RLock()
	user := m.users[claims.Username]
	m.usersMutex.RUnlock()

	subscriber := &Subscriber{
		Conn:  conn,
		Mutex: sync.Mutex{},
		User:  user,
	}

	m.mutex.Lock()
	if m.subscribers == nil {
		m.subscribers = make(map[*websocket.Conn]*Subscriber)
	}
	m.subscribers[conn] = subscriber
	m.mutex.Unlock()

	log.Printf("Subscriber WebSocket connected: %s (User: %s)", conn.RemoteAddr(), claims.Username)

	// 启动读取循环以检测连接断开
	defer func() {
		m.mutex.Lock()
		delete(m.subscribers, conn)
		m.mutex.Unlock()
		conn.Close()
		log.Printf("Subscriber WebSocket disconnected: %s", conn.RemoteAddr())
	}()

	for {
		// Subscriber 通常只接收消息，不发送。
		// 但我们需要读取以检测连接关闭。
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// forwardWorkerRequestToBot 将Worker请求转发给Bot
func (m *Manager) forwardWorkerRequestToBot(worker *WorkerClient, msg map[string]interface{}, echo string) {
	// 保存请求映射：echo -> worker
	m.pendingMutex.Lock()
	m.pendingRequests[echo] = make(chan map[string]interface{}, 1)
	m.pendingMutex.Unlock()

	// 选择目标Bot（这里可以添加更复杂的路由逻辑）
	m.mutex.RLock()
	var targetBot *BotClient
	for _, bot := range m.bots {
		targetBot = bot
		break
	}
	m.mutex.RUnlock()

	if targetBot == nil {
		log.Printf("No available Bot to handle Worker request (echo: %s)", echo)

		// 返回错误响应给Worker
		response := map[string]interface{}{
			"status":  "failed",
			"retcode": 1404,
			"msg":     "No Bot available",
			"echo":    echo,
			"data":    nil,
		}

		worker.Mutex.Lock()
		worker.Conn.WriteJSON(response)
		worker.Mutex.Unlock()

		// 清理映射
		m.pendingMutex.Lock()
		delete(m.pendingRequests, echo)
		m.pendingMutex.Unlock()
		return
	}

	// 转发请求给Bot
	targetBot.Mutex.Lock()
	err := targetBot.Conn.WriteJSON(msg)
	targetBot.Mutex.Unlock()

	if err != nil {
		log.Printf("Failed to forward Worker request to Bot %s: %v", targetBot.SelfID, err)

		// 返回错误响应给Worker
		response := map[string]interface{}{
			"status":  "failed",
			"retcode": 1400,
			"msg":     "Failed to forward to Bot",
			"echo":    echo,
			"data":    nil,
		}

		worker.Mutex.Lock()
		worker.Conn.WriteJSON(response)
		worker.Mutex.Unlock()

		// 清理映射
		m.pendingMutex.Lock()
		delete(m.pendingRequests, echo)
		m.pendingMutex.Unlock()
		return
	}

	log.Printf("Forwarded Worker request (echo: %s) to Bot %s", echo, targetBot.SelfID)

	// 启动超时处理（30秒内必须收到响应）
	go func() {
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		select {
		case response := <-m.pendingRequests[echo]:
			// 收到响应，转发给Worker
			worker.Mutex.Lock()
			worker.Conn.WriteJSON(response)
			worker.Mutex.Unlock()
			log.Printf("Forwarded Bot response (echo: %s) to Worker %s", echo, worker.ID)

		case <-timeout.C:
			// 超时，返回错误响应
			log.Printf("Worker request (echo: %s) timed out", echo)

			response := map[string]interface{}{
				"status":  "failed",
				"retcode": 1401,
				"msg":     "Request timeout",
				"echo":    echo,
				"data":    nil,
			}

			worker.Mutex.Lock()
			worker.Conn.WriteJSON(response)
			worker.Mutex.Unlock()
		}

		// 清理映射
		m.pendingMutex.Lock()
		delete(m.pendingRequests, echo)
		m.pendingMutex.Unlock()
	}()
}

// cleanupPendingRequests 清理过期的请求映射
func (m *Manager) cleanupPendingRequests() {
	m.pendingMutex.Lock()
	defer m.pendingMutex.Unlock()

	// 清理所有pending请求（通常在系统关闭时调用）
	for echo, ch := range m.pendingRequests {
		close(ch)
		delete(m.pendingRequests, echo)
		log.Printf("Cleaned up pending request: %s", echo)
	}
}

// StartPeriodicStatsSave 启动定期保存统计信息
func (m *Manager) StartPeriodicStatsSave() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.LogInfo("Saving stats...")
		// 这里可以添加保存逻辑
	}
}

// updateBotStats 更新Bot统计信息
func (m *Manager) updateBotStats(botID string, userID, groupID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 初始化Bot详细统计
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}

	if _, exists := m.BotDetailedStats[botID]; !exists {
		m.BotDetailedStats[botID] = &BotStatDetail{
			Users:  make(map[int64]int64),
			Groups: make(map[int64]int64),
		}
	}

	stats := m.BotDetailedStats[botID]
	stats.Received++
	stats.LastMsg = time.Now()

	if userID > 0 {
		stats.Users[userID]++
	}
	if groupID > 0 {
		stats.Groups[groupID]++
	}

	// 更新全局统计
	if m.UserStats == nil {
		m.UserStats = make(map[int64]int64)
	}
	if m.GroupStats == nil {
		m.GroupStats = make(map[int64]int64)
	}

	m.UserStats[userID]++
	m.GroupStats[groupID]++
	m.TotalMessages++
}

// StartTrendCollection 启动趋势收集
func (m *Manager) StartTrendCollection() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 趋势收集逻辑（静默执行，不打印日志）
		// m.LogDebug("Collecting trends...")
	}
}

// ==================== 用户管理相关接口 ====================

// handleGetUserInfo 获取当前登录用户信息
func (m *Manager) handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
		})
		return
	}

	m.usersMutex.RLock()
	user, exists := m.users[claims.Username]
	m.usersMutex.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt.Format(time.RFC3339),
			"updated_at": user.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// handleChangePassword 修改用户密码
func (m *Manager) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
		})
		return
	}

	var data struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	user, exists := m.users[claims.Username]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	// 验证旧密码
	if !checkPassword(data.OldPassword, user.PasswordHash) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "原密码错误",
		})
		return
	}

	// 哈希新密码
	newHash, err := hashPassword(data.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "密码加密失败",
		})
		return
	}

	// 更新密码
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	// 保存到Redis
	go m.saveUsersToRedis()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "密码修改成功",
	})
}

// ==================== 路由规则管理接口 ====================

// handleGetRoutingRules 获取所有路由规则
func (m *Manager) handleGetRoutingRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rules":   m.routingRules,
	})
}

// handleSetRoutingRule 设置路由规则
func (m *Manager) handleSetRoutingRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rule struct {
		Key      string `json:"key"`
		WorkerID string `json:"worker_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	if rule.Key == "" || rule.WorkerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Key和WorkerID不能为空",
		})
		return
	}

	m.mutex.Lock()
	if m.routingRules == nil {
		m.routingRules = make(map[string]string)
	}
	m.routingRules[rule.Key] = rule.WorkerID
	m.mutex.Unlock()

	log.Printf("[ADMIN] Set routing rule: %s -> %s", rule.Key, rule.WorkerID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "路由规则设置成功",
	})
}

// handleDeleteRoutingRule 删除路由规则
func (m *Manager) handleDeleteRoutingRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Key不能为空",
		})
		return
	}

	m.mutex.Lock()
	if _, exists := m.routingRules[key]; exists {
		delete(m.routingRules, key)
		log.Printf("[ADMIN] Deleted routing rule for key: %s", key)
	}
	m.mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "路由规则删除成功",
	})
}
