package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
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

// initDefaultAdmin 初始化默认管理员用户
// 注意：调用此函数时需要确保已经持有usersMutex锁
func (m *Manager) initDefaultAdmin() {
	log.Printf("[INFO] 正在初始化默认管理员检测...")

	// 检查是否已存在管理员用户 (先查缓存)
	if _, exists := m.users["admin"]; exists {
		log.Printf("[INFO] 管理员用户 'admin' 已存在于内存缓存中，跳过初始化")
		return
	}

	// 再次检查数据库 (双重检查)
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err == nil && count > 0 {
		log.Printf("[INFO] 管理员用户 'admin' 已存在于数据库中，正在重新加载...")
		m.loadUsersFromDBNoLock()
		return
	}

	// 对默认密码进行哈希处理
	hashedPassword, err := hashPassword(DEFAULT_ADMIN_PASSWORD)
	if err != nil {
		log.Printf("[ERROR] 初始化默认管理员密码哈希失败: %v", err)
		return
	}

	// 创建默认管理员用户
	adminUser := &User{
		Username:       "admin",
		PasswordHash:   hashedPassword,
		IsAdmin:        true,
		SessionVersion: 1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 存储到用户映射
	m.users["admin"] = adminUser
	log.Printf("[INFO] 默认管理员用户已创建，用户名: admin, 初始密码: %s", DEFAULT_ADMIN_PASSWORD)

	// 保存到数据库
	if err := m.saveUserToDB(adminUser); err != nil {
		log.Printf("[ERROR] 保存默认管理员用户到数据库失败: %v", err)
	}
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

	// 解析请求体
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		log.Printf("[WARN] 登录请求解析失败: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	log.Printf("[INFO] 登录尝试 - 用户名: %s, 客户端IP: %s", loginData.Username, r.RemoteAddr)

	// 验证用户名和密码
	m.usersMutex.RLock()
	user, exists := m.users[loginData.Username]
	m.usersMutex.RUnlock()

	// 如果内存中不存在，尝试从数据库加载
	if !exists {
		log.Printf("[INFO] 用户 %s 不在内存缓存中，尝试从数据库查询...", loginData.Username)
		// 简单的单用户查询
		row := m.db.QueryRow("SELECT id, username, password_hash, is_admin, session_version, created_at, updated_at FROM users WHERE username = ?", loginData.Username)
		var u User
		var createdAt, updatedAt string
		err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.SessionVersion, &createdAt, &updatedAt)
		if err == nil {
			if createdAt != "" {
				u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
			}
			if updatedAt != "" {
				u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
			}
			user = &u
			exists = true

			// 更新内存缓存
			m.usersMutex.Lock()
			m.users[u.Username] = &u
			m.usersMutex.Unlock()
			log.Printf("[INFO] 从数据库成功加载用户: %s", u.Username)
		}
	}

	if !exists {
		log.Printf("[WARN] 登录失败 - 用户不存在: %s", loginData.Username)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if !checkPassword(loginData.Password, user.PasswordHash) {
		log.Printf("[WARN] 登录失败 - 密码不匹配: %s", loginData.Username)
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
		log.Printf("[ERROR] Token生成失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Token生成失败",
		})
		return
	}

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	log.Printf("[INFO] 登录成功 - 用户: %s, 角色: %s", user.Username, role)

	// 返回成功响应
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"role":    role,
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

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	m.connectionStats.Mutex.RLock()
	defer m.connectionStats.Mutex.RUnlock()

	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	// 计算在线/离线机器人
	onlineBots := len(m.bots)
	totalBots := len(m.BotStats)
	offlineBots := totalBots - onlineBots
	if offlineBots < 0 {
		offlineBots = 0
	}

	// 系统运行时信息
	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)

	// CPU 信息
	cpuInfos, _ := cpu.Info()
	cpuModel := "Unknown"
	cpuCoresPhysical := 0
	cpuCoresLogical := 0
	cpuFreq := 0.0
	if len(cpuInfos) > 0 {
		cpuModel = cpuInfos[0].ModelName
		cpuCoresPhysical = int(cpuInfos[0].Cores)
		// 逻辑核心通常通过 cpu.Counts(true) 获取
		logical, _ := cpu.Counts(true)
		cpuCoresLogical = logical
		cpuFreq = cpuInfos[0].Mhz
	}

	// CPU 使用率
	cpuPercent, _ := cpu.Percent(0, false)
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// OS 信息
	hi, _ := host.Info()

	m.HistoryMutex.RLock()
	cpuTrend := append([]float64{}, m.CPUTrend...)
	memTrend := append([]uint64{}, m.MemTrend...)
	msgTrend := append([]int64{}, m.MsgTrend...)
	sentTrend := append([]int64{}, m.SentTrend...)
	recvTrend := append([]int64{}, m.RecvTrend...)
	m.HistoryMutex.RUnlock()

	stats := map[string]interface{}{
		"goroutines":          runtime.NumGoroutine(),
		"memory_alloc":        mStats.Alloc,
		"memory_total":        mStats.Sys,
		"bot_count":           onlineBots,
		"bot_count_offline":   offlineBots,
		"bot_count_total":     totalBots,
		"active_groups_today": len(m.GroupStatsToday),
		"active_groups":       len(m.GroupStats),
		"active_users_today":  len(m.UserStatsToday),
		"active_users":        len(m.UserStats),
		"message_count":       m.TotalMessages,
		"sent_message_count":  m.SentMessages,
		"cpu_usage":           cpuUsage,
		"cpu_model":           cpuModel,
		"cpu_cores_physical":  cpuCoresPhysical,
		"cpu_cores_logical":   cpuCoresLogical,
		"cpu_freq":            cpuFreq,
		"os_platform":         hi.Platform,
		"os_version":          hi.PlatformVersion,
		"os_arch":             hi.KernelArch,
		"timestamp":           time.Now().Format("2006-01-02 15:04:05"),
		// 详情数据
		"bots_detail": m.BotDetailedStats,
		// 趋势数据
		"cpu_trend":  cpuTrend,
		"mem_trend":  memTrend,
		"msg_trend":  msgTrend,
		"sent_trend": sentTrend,
		"recv_trend": recvTrend,
	}

	json.NewEncoder(w).Encode(stats)
}

// handleGetSystemStats 获取详细的系统运行统计
func (m *Manager) handleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取CPU使用率
	cpuPercent, _ := cpu.Percent(time.Second, false)
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// 获取内存使用情况
	vm, _ := mem.VirtualMemory()

	// 获取主机信息
	hi, _ := host.Info()

	// 获取前5个CPU消耗最高的进程
	procs, _ := process.Processes()
	type procInfo struct {
		Pid    int32   `json:"pid"`
		Name   string  `json:"name"`
		CPU    float64 `json:"cpu"`
		Memory uint64  `json:"memory"`
	}
	var processList []procInfo
	for i, p := range procs {
		if i > 50 {
			break
		} // 限制扫描数量
		name, _ := p.Name()
		cp, _ := p.CPUPercent()
		mp, _ := p.MemoryInfo()
		if mp != nil {
			processList = append(processList, procInfo{
				Pid:    p.Pid,
				Name:   name,
				CPU:    cp,
				Memory: mp.RSS,
			})
		}
	}

	stats := map[string]interface{}{
		"cpu_usage": cpuUsage,
		"mem_usage": vm.UsedPercent,
		"mem_total": vm.Total,
		"mem_free":  vm.Free,
		"host_info": hi,
		"processes": processList,
		"timestamp": time.Now().Unix(),
	}

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

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// 初始化统计数据结构 (如果需要)
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}
	if m.UserStats == nil {
		m.UserStats = make(map[int64]int64)
	}
	if m.GroupStats == nil {
		m.GroupStats = make(map[int64]int64)
	}
	if m.BotStats == nil {
		m.BotStats = make(map[string]int64)
	}
	if m.UserStatsToday == nil {
		m.UserStatsToday = make(map[int64]int64)
	}
	if m.GroupStatsToday == nil {
		m.GroupStatsToday = make(map[int64]int64)
	}
	if m.BotStatsToday == nil {
		m.BotStatsToday = make(map[string]int64)
	}

	// 更新Bot详细统计
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
		m.UserStats[userID]++
		m.UserStatsToday[userID]++
	}
	if groupID > 0 {
		stats.Groups[groupID]++
		m.GroupStats[groupID]++
		m.GroupStatsToday[groupID]++
	}

	// 更新全局和今日统计
	m.BotStats[botID]++
	m.BotStatsToday[botID]++
	m.TotalMessages++
}

// StartTrendCollection 启动趋势收集
func (m *Manager) StartTrendCollection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.HistoryMutex.Lock()

		// 获取系统指标
		var mStats runtime.MemStats
		runtime.ReadMemStats(&mStats)

		cpuPercent, _ := cpu.Percent(0, false)
		var currentCPU float64
		if len(cpuPercent) > 0 {
			currentCPU = cpuPercent[0]
		}

		// 更新趋势数组
		m.CPUTrend = append(m.CPUTrend, currentCPU)
		m.MemTrend = append(m.MemTrend, mStats.Alloc)

		// 消息增量计算
		m.statsMutex.RLock()
		total := m.TotalMessages
		sent := m.SentMessages
		m.statsMutex.RUnlock()
		
		if len(m.MsgTrend) > 0 {
			// 这里我们存的是增量，但前端代码逻辑可能需要处理
			// 实际上前端 index.html 2402行 在做 Moving Sum，所以我们这里存增量是对的
		}

		// 简单起见，我们先存当前的 Total，然后在 GetStats 里计算增量或者让前端处理
		// 修正：前端期望的是每个时间点的消息数，它会自己计算增量
		m.MsgTrend = append(m.MsgTrend, total)
		m.SentTrend = append(m.SentTrend, sent)
		m.RecvTrend = append(m.RecvTrend, total-sent)

		// 保持长度，限制为 60 个点 (5秒一个点，共5分钟)
		maxPoints := 60
		if len(m.CPUTrend) > maxPoints {
			m.CPUTrend = m.CPUTrend[1:]
			m.MemTrend = m.MemTrend[1:]
			m.MsgTrend = m.MsgTrend[1:]
			m.SentTrend = m.SentTrend[1:]
			m.RecvTrend = m.RecvTrend[1:]
		}

		m.HistoryMutex.Unlock()
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

	// 如果内存中不存在，尝试从数据库加载
	if !exists {
		row := m.db.QueryRow("SELECT id, username, password_hash, is_admin, session_version, created_at, updated_at FROM users WHERE username = ?", claims.Username)
		var u User
		var createdAt, updatedAt string
		err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.SessionVersion, &createdAt, &updatedAt)
		if err == nil {
			if createdAt != "" {
				u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
			}
			if updatedAt != "" {
				u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
			}
			user = &u
			exists = true

			// 更新内存缓存
			m.usersMutex.Lock()
			m.users[u.Username] = &u
			m.usersMutex.Unlock()
		}
	}

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

	// 保存到数据库
	if err := m.saveUserToDB(user); err != nil {
		log.Printf("更新密码到数据库失败: %v", err)
	}

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
