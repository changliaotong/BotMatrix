package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
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
	onlineWorkers := len(m.workers)
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
	netSentTrend := append([]uint64{}, m.NetSentTrend...)
	netRecvTrend := append([]uint64{}, m.NetRecvTrend...)
	m.HistoryMutex.RUnlock()

	// 获取内存使用情况
	vm, _ := mem.VirtualMemory()

	stats := map[string]interface{}{
		"goroutines":          runtime.NumGoroutine(),
		"memory_alloc":        mStats.Alloc,
		"memory_total":        vm.Total,
		"memory_used":         vm.Used,
		"memory_free":         vm.Free,
		"memory_used_percent": vm.UsedPercent,
		"bot_count":           onlineBots,
		"worker_count":        onlineWorkers,
		"bot_count_offline":   offlineBots,
		"bot_count_total":     totalBots,
		"active_groups_today": len(m.GroupStatsToday),
		"active_groups":       len(m.GroupStats),
		"active_users_today":  len(m.UserStatsToday),
		"active_users":        len(m.UserStats),
		"message_count":       m.TotalMessages,
		"sent_message_count":  m.SentMessages,
		"cpu_usage":           cpuUsage,
		"start_time":          m.StartTime.Unix(),
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
		"cpu_trend":      cpuTrend,
		"mem_trend":      memTrend,
		"msg_trend":      msgTrend,
		"sent_trend":     sentTrend,
		"recv_trend":     recvTrend,
		"net_sent_trend": netSentTrend,
		"net_recv_trend": netRecvTrend,
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

	// 获取磁盘使用情况
	partitions, _ := disk.Partitions(true) // 获取所有分区，包括物理分区
	var diskUsage []map[string]interface{}
	seenMounts := make(map[string]bool)
	for _, p := range partitions {
		// 过滤掉常见的非物理文件系统和重复挂载点
		if seenMounts[p.Mountpoint] {
			continue
		}
		// 排除一些虚拟文件系统 (Linux 常用)
		if strings.HasPrefix(p.Device, "/dev/loop") ||
			p.Fstype == "tmpfs" ||
			p.Fstype == "devtmpfs" ||
			p.Fstype == "overlay" {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err == nil && usage.Total > 0 {
			diskUsage = append(diskUsage, map[string]interface{}{
				"path":        p.Mountpoint,
				"total":       usage.Total,
				"free":        usage.Free,
				"used":        usage.Used,
				"usedPercent": usage.UsedPercent,
			})
			seenMounts[p.Mountpoint] = true
		}
	}

	// 获取网络 IO
	netIO, _ := net.IOCounters(false)
	var netUsage []map[string]interface{}
	for _, io := range netIO {
		netUsage = append(netUsage, map[string]interface{}{
			"name":      io.Name,
			"bytesSent": io.BytesSent,
			"bytesRecv": io.BytesRecv,
		})
	}

	// 获取网络接口
	interfaces, _ := net.Interfaces()
	var netInterfaces []map[string]interface{}
	for _, i := range interfaces {
		var addrs []map[string]interface{}
		for _, addr := range i.Addrs {
			addrs = append(addrs, map[string]interface{}{
				"addr": addr.Addr,
			})
		}
		netInterfaces = append(netInterfaces, map[string]interface{}{
			"name":  i.Name,
			"addrs": addrs,
		})
	}

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
		"cpu_usage":      cpuUsage,
		"mem_usage":      vm.UsedPercent,
		"mem_total":      vm.Total,
		"mem_free":       vm.Free,
		"disk_usage":     diskUsage,
		"net_io":         netUsage,
		"net_interfaces": netInterfaces,
		"host_info":      hi,
		"processes":      processList,
		"timestamp":      time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(stats)
}

// handleGetLogs 处理获取日志的请求
func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs := m.GetLogs(100) // 获取最近100条日志
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"logs":    logs,
	})
}

// handleGetContacts 处理获取联系人/群组信息的请求
func (m *Manager) handleGetContacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	// 整合群组和好友信息
	contacts := make([]map[string]interface{}, 0)

	for _, group := range m.groupCache {
		contacts = append(contacts, group)
	}

	for _, friend := range m.friendCache {
		contacts = append(contacts, friend)
	}

	json.NewEncoder(w).Encode(contacts)
}

// handleGetConfig 获取当前配置
func (m *Manager) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	syncToGlobalConfig()
	json.NewEncoder(w).Encode(GlobalConfig)
}

// handleUpdateConfig 更新配置
func (m *Manager) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newConfig AppConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的配置格式",
		})
		return
	}

	// 更新全局变量 (即时生效的部分)
	redisChanged := false
	if newConfig.WSPort != "" {
		WS_PORT = newConfig.WSPort
	}
	if newConfig.WebUIPort != "" {
		WEBUI_PORT = newConfig.WebUIPort
	}
	if newConfig.StatsFile != "" {
		STATS_FILE = newConfig.StatsFile
	}
	if newConfig.RedisAddr != "" {
		if REDIS_ADDR != newConfig.RedisAddr {
			REDIS_ADDR = newConfig.RedisAddr
			redisChanged = true
		}
	}
	if newConfig.RedisPwd != "" {
		if REDIS_PWD != newConfig.RedisPwd {
			REDIS_PWD = newConfig.RedisPwd
			redisChanged = true
		}
	}
	if newConfig.JWTSecret != "" {
		JWT_SECRET = newConfig.JWTSecret
	}
	if newConfig.DefaultAdminPassword != "" {
		DEFAULT_ADMIN_PASSWORD = newConfig.DefaultAdminPassword
	}

	// 如果 Redis 配置发生变化，重新初始化 Redis 客户端
	if redisChanged {
		log.Printf("[INFO] Redis 配置发生变化，正在重新初始化客户端...")
		if m.rdb != nil {
			m.rdb.Close()
		}
		m.rdb = redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Password: REDIS_PWD,
			DB:       0,
		})
		// 测试新连接
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.rdb.Ping(ctx).Err(); err != nil {
			log.Printf("[WARN] 无法连接到新 Redis: %v", err)
			// 注意：这里不设为 nil，让它保留客户端对象以便之后重试或报错
		} else {
			log.Printf("[INFO] 已成功连接到新 Redis")
		}
	}

	// 同步并保存到文件
	if err := saveConfigToFile(); err != nil {
		log.Printf("[ERROR] 保存配置失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存配置失败",
		})
		return
	}

	log.Printf("[INFO] 配置已更新并保存")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置已更新，部分配置(如端口)可能需要重启服务生效",
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
	bot := &BotClient{
		Conn:          conn,
		Connected:     time.Now(),
		LastHeartbeat: time.Now(),
		Platform:      platform,
		SelfID:        selfID,
	}

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

	// 异步获取 Bot 信息
	go m.fetchBotInfo(bot)

	// 启动连接处理循环
	go m.handleBotConnection(bot)
}

// fetchBotInfo 主动获取 Bot 的详细信息
func (m *Manager) fetchBotInfo(bot *BotClient) {
	// 等待一秒，确保连接完全建立且握手完成
	time.Sleep(1 * time.Second)

	log.Printf("[Bot] Fetching info for bot: %s", bot.SelfID)

	// 1. 获取登录信息 (昵称等)
	echoInfo := "fetch_info_" + bot.SelfID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	m.pendingMutex.Lock()
	m.pendingRequests[echoInfo] = make(chan map[string]interface{}, 1)
	m.pendingMutex.Unlock()

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
	m.pendingMutex.Lock()
	m.pendingRequests[echoGroups] = make(chan map[string]interface{}, 1)
	m.pendingMutex.Unlock()

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
		case resp := <-m.pendingRequests[echoInfo]:
			if data, ok := resp["data"].(map[string]interface{}); ok {
				bot.Mutex.Lock()
				var newNickname string
				var newSelfID string

				if nickname, ok := data["nickname"].(string); ok {
					newNickname = nickname
				}
				if userID, ok := data["user_id"]; ok {
					newSelfID = toString(userID)
				}

				if newSelfID != "" && newSelfID != bot.SelfID {
					oldID := bot.SelfID
					bot.Mutex.Unlock() // Unlock before map operations

					m.mutex.Lock()
					delete(m.bots, oldID)
					bot.SelfID = newSelfID
					bot.Nickname = newNickname
					m.bots[bot.SelfID] = bot
					m.mutex.Unlock()

					bot.Mutex.Lock() // Re-lock
					log.Printf("[Bot] Updated Bot ID from %s to %s via get_login_info", oldID, newSelfID)
				} else {
					bot.Nickname = newNickname
				}
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated info for %s: Nickname=%s", bot.SelfID, bot.Nickname)
			}
			m.pendingMutex.Lock()
			delete(m.pendingRequests, echoInfo)
			m.pendingMutex.Unlock()
			infoDone = true

		case resp := <-m.pendingRequests[echoGroups]:
			if data, ok := resp["data"].([]interface{}); ok {
				bot.Mutex.Lock()
				bot.GroupCount = len(data)
				bot.Mutex.Unlock()
				log.Printf("[Bot] Updated group count for %s: %d", bot.SelfID, bot.GroupCount)

				// 更新群组缓存，以便后续路由 API 请求
				m.cacheMutex.Lock()
				for _, item := range data {
					if group, ok := item.(map[string]interface{}); ok {
						var gID string
						if idVal, ok := group["group_id"]; ok {
							gID = toString(idVal)
						}

						if gID != "" {
							name, _ := group["group_name"].(string)
							if name == "" {
								name = fmt.Sprintf("Group %s (Auto)", gID)
							}
							m.groupCache[gID] = map[string]interface{}{
								"group_id":   gID,
								"group_name": name,
								"bot_id":     bot.SelfID,
								"is_cached":  true,
								"source":     "get_group_list",
							}
						}
					}
				}
				m.cacheMutex.Unlock()
				log.Printf("[Bot] Cached %d groups for Bot %s", len(data), bot.SelfID)
			}
			m.pendingMutex.Lock()
			delete(m.pendingRequests, echoGroups)
			m.pendingMutex.Unlock()
			groupsDone = true

		case <-timeout:
			log.Printf("[Bot] Timeout fetching info for bot %s", bot.SelfID)
			infoDone = true
			groupsDone = true
		}
	}
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
		err := ReadJSONWithNumber(bot.Conn, &msg)
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
	// 检查是否包含 self_id 并更新（如果当前是临时 ID）
	if msgSelfID, ok := msg["self_id"].(string); ok && msgSelfID != "" {
		if bot.SelfID != msgSelfID && strings.Contains(bot.SelfID, ":") {
			// 当前是临时 IP ID，收到正式 ID，进行更新
			oldID := bot.SelfID
			m.mutex.Lock()
			delete(m.bots, oldID)
			bot.SelfID = msgSelfID
			m.bots[bot.SelfID] = bot
			m.mutex.Unlock()
			log.Printf("[Bot] Updated Bot ID from %s to %s", oldID, msgSelfID)
		}
	}

	// 检查是否是API响应（有echo字段）
	if echo, ok := msg["echo"].(string); ok && echo != "" {
		// 这是API响应，需要回传给对应的Worker
		m.pendingMutex.Lock()
		respChan, exists := m.pendingRequests[echo]
		sendTime, timeExists := m.pendingTimestamps[echo]
		delete(m.pendingTimestamps, echo)
		m.pendingMutex.Unlock()

		if exists {
			// 记录 RTT
			if timeExists {
				rtt := time.Since(sendTime)
				// 找到对应的 Worker 并更新 RTT
				// 使用 | 作为分隔符: workerID|originalEcho
				if parts := strings.Split(echo, "|"); len(parts) >= 2 {
					workerID := parts[0]
					m.mutex.RLock()
					for _, w := range m.workers {
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
					m.mutex.RUnlock()
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

	// 缓存群/成员/好友信息 (基于消息)
	m.cacheBotDataFromMessage(bot, msg)
}

// cacheBotDataFromMessage 从消息中提取并缓存数据 (特别针对腾讯频道机器人)
func (m *Manager) cacheBotDataFromMessage(bot *BotClient, msg map[string]interface{}) {
	postType, _ := msg["post_type"].(string)
	if postType != "message" {
		return
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// 缓存群信息
	if groupIDVal, ok := msg["group_id"]; ok && groupIDVal != nil {
		groupID := toInt64(groupIDVal)
		gID := toString(groupIDVal)
		// 更新或添加群组缓存
		m.groupCache[gID] = map[string]interface{}{
			"group_id":   groupID,
			"group_name": fmt.Sprintf("Group %s (Cached)", gID),
			"bot_id":     bot.SelfID,
			"is_cached":  true,
			"reason":     "Automatically updated from message",
			"last_seen":  time.Now(),
		}

		// 缓存成员信息
		if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
			userID := toInt64(userIDVal)
			uID := toString(userIDVal)
			key := fmt.Sprintf("%s:%s", gID, uID)
			sender, _ := msg["sender"].(map[string]interface{})
			nickname := ""
			card := ""
			if sender != nil {
				nickname, _ = sender["nickname"].(string)
				card, _ = sender["card"].(string)
			}
			m.memberCache[key] = map[string]interface{}{
				"group_id":  groupID,
				"user_id":   userID,
				"nickname":  nickname,
				"card":      card,
				"is_cached": true,
			}
		}
	} else if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
		// 缓存好友信息 (私聊)
		userID := toInt64(userIDVal)
		uID := toString(userIDVal)
		if _, exists := m.friendCache[uID]; !exists {
			sender, _ := msg["sender"].(map[string]interface{})
			nickname := ""
			if sender != nil {
				nickname, _ = sender["nickname"].(string)
			}
			m.friendCache[uID] = map[string]interface{}{
				"user_id":   userID,
				"nickname":  nickname,
				"is_cached": true,
			}
		}
	}
}

// handleBotMessageEvent 处理Bot消息事件
func (m *Manager) handleBotMessageEvent(bot *BotClient, msg map[string]interface{}) {
	// 提取消息信息
	userID := toInt64(msg["user_id"])
	groupID := toInt64(msg["group_id"])
	message, _ := msg["message"].(string)

	log.Printf("Bot %s received message from user %d: %s", bot.SelfID, userID, message)

	// 更新详细统计
	m.updateBotStats(bot.SelfID, userID, groupID)
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

// cacheMessage 缓存无法立即处理的消息
func (m *Manager) cacheMessage(msg map[string]interface{}) {
	m.messageCacheMutex.Lock()
	defer m.messageCacheMutex.Unlock()

	// 限制缓存大小，防止内存溢出
	if len(m.messageCache) > 1000 {
		m.messageCache = m.messageCache[1:] // 丢弃最旧的消息
	}
	m.messageCache = append(m.messageCache, msg)
	log.Printf("[CACHE] No workers available, message cached (Total: %d)", len(m.messageCache))
}

// flushMessageCache 当有新 Worker 连接时，发送缓存的消息
func (m *Manager) flushMessageCache() {
	m.mutex.RLock()
	workerCount := len(m.workers)
	m.mutex.RUnlock()

	if workerCount == 0 {
		return
	}

	m.messageCacheMutex.Lock()
	if len(m.messageCache) == 0 {
		m.messageCacheMutex.Unlock()
		return
	}

	cache := m.messageCache
	m.messageCache = nil
	m.messageCacheMutex.Unlock()

	log.Printf("[CACHE] Flushing %d cached messages to workers", len(cache))
	for _, msg := range cache {
		// 重新通过路由发送
		go m.forwardMessageToWorker(msg)
	}
}

// matchRoutePattern 检查字符串是否匹配模式 (支持 * 通配符)
func matchRoutePattern(pattern, value string) bool {
	if pattern == value {
		return true
	}
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(value, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(value, suffix)
	}
	return false
}

// findWorkerByID 根据ID查找Worker
func (m *Manager) findWorkerByID(workerID string) *WorkerClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, w := range m.workers {
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
		sID := toString(userID)
		if sID != "0" && sID != "" {
			uID := fmt.Sprintf("user_%s", sID)
			matchKeys = append(matchKeys, uID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// 群组 ID 匹配: group_789012
	if groupID, ok := msg["group_id"]; ok {
		sID := toString(groupID)
		if sID != "0" && sID != "" {
			gID := fmt.Sprintf("group_%s", sID)
			matchKeys = append(matchKeys, gID)
			matchKeys = append(matchKeys, sID)
		}
	}

	// 机器人 ID 匹配: bot_123 or self_id
	if selfID, ok := msg["self_id"]; ok {
		sID := toString(selfID)
		if sID != "" {
			matchKeys = append(matchKeys, fmt.Sprintf("bot_%s", sID))
			matchKeys = append(matchKeys, sID)
		}
	}

	// 查找匹配规则
	m.mutex.RLock()
	var matchedKey string

	// A. 精确匹配优先
	for _, key := range matchKeys {
		if wID, exists := m.routingRules[key]; exists && wID != "" {
			targetWorkerID = wID
			matchedKey = key
			break
		}
	}

	// B. 如果没有精确匹配，尝试通配符匹配 (按长度倒序排序，确保最具体匹配优先且结果确定)
	if targetWorkerID == "" {
		// 收集并排序通配符规则
		type rule struct {
			pattern string
			wID     string
		}
		var wildcards []rule
		for p, w := range m.routingRules {
			if strings.Contains(p, "*") {
				wildcards = append(wildcards, rule{p, w})
			}
		}

		// 按模式长度降序排列，越长的模式通常越具体
		sort.Slice(wildcards, func(i, j int) bool {
			return len(wildcards[i].pattern) > len(wildcards[j].pattern)
		})

		for _, r := range wildcards {
			for _, key := range matchKeys {
				if matchRoutePattern(r.pattern, key) {
					targetWorkerID = r.wID
					matchedKey = fmt.Sprintf("%s (via pattern %s)", key, r.pattern)
					break
				}
			}
			if targetWorkerID != "" {
				break
			}
		}
	}
	m.mutex.RUnlock()

	// 如果找到了目标 Worker，尝试获取它
	if targetWorkerID != "" {
		if w := m.findWorkerByID(targetWorkerID); w != nil {
			log.Printf("[ROUTING] Rule Matched: %s -> Target Worker: %s", matchedKey, targetWorkerID)
			w.Mutex.Lock()
			err := w.Conn.WriteJSON(msg)
			w.Mutex.Unlock()

			if err == nil {
				m.mutex.Lock()
				w.HandledCount++
				m.mutex.Unlock()
				return
			}
			log.Printf("[ROUTING] [ERROR] Failed to send to target worker %s: %v. Falling back to load balancer.", targetWorkerID, err)
		} else {
			log.Printf("[ROUTING] [WARNING] Target worker %s defined in rule (%s) is OFFLINE or NOT FOUND. Falling back to load balancer.", targetWorkerID, matchedKey)
		}
	}

	// 2. 负载均衡转发
	m.mutex.RLock()
	// 过滤出健康的 worker
	var healthyWorkers []*WorkerClient
	for _, w := range m.workers {
		// 简单检查：如果 60 秒没有心跳且不是刚连接，视为不健康
		if time.Since(w.LastHeartbeat) < 60*time.Second || time.Since(w.Connected) < 10*time.Second {
			healthyWorkers = append(healthyWorkers, w)
		}
	}
	m.mutex.RUnlock()

	if len(healthyWorkers) == 0 {
		log.Printf("[ROUTING] [WARNING] No healthy workers available. Caching message.")
		m.cacheMessage(msg)
		return
	}

	var selectedWorker *WorkerClient
	m.mutex.Lock()

	// 负载均衡算法
	if len(healthyWorkers) == 1 {
		selectedWorker = healthyWorkers[0]
	} else {
		// 1. 优先选择从未处理过消息的 Worker (以便获取 RTT)
		var unhandled []*WorkerClient
		for _, w := range healthyWorkers {
			if w.HandledCount == 0 {
				unhandled = append(unhandled, w)
			}
		}

		if len(unhandled) > 0 {
			idx := m.workerIndex % len(unhandled)
			selectedWorker = unhandled[idx]
			m.workerIndex++
		} else {
			// 2. 选择 AvgRTT 最小的 Worker
			var minRTT time.Duration = -1
			for _, w := range healthyWorkers {
				if w.AvgRTT > 0 {
					if minRTT == -1 || w.AvgRTT < minRTT {
						minRTT = w.AvgRTT
						selectedWorker = w
					}
				}
			}

			// 3. 退回到全局轮询
			if selectedWorker == nil {
				idx := m.workerIndex % len(healthyWorkers)
				selectedWorker = healthyWorkers[idx]
				m.workerIndex++
			}
		}
	}
	m.mutex.Unlock()

	// 发送消息给选中的Worker
	selectedWorker.Mutex.Lock()
	err := selectedWorker.Conn.WriteJSON(msg)
	selectedWorker.Mutex.Unlock()

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward to worker %s: %v. Removing and retrying...", selectedWorker.ID, err)
		m.removeWorker(selectedWorker.ID)
		m.forwardMessageToWorkerWithRetry(msg, retryCount+1)
	} else {
		m.mutex.Lock()
		selectedWorker.HandledCount++
		m.mutex.Unlock()
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

	// 尝试发送缓存的消息
	go m.flushMessageCache()

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
		err := ReadJSONWithNumber(worker.Conn, &msg)
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
func (m *Manager) forwardWorkerRequestToBot(worker *WorkerClient, msg map[string]interface{}, originalEcho string) {
	// 构造内部 echo，包含 worker ID 以便追踪和记录 RTT
	// 使用 | 作为分隔符，因为 worker ID 可能包含 : (如 IP:Port)
	internalEcho := fmt.Sprintf("%s|%s", worker.ID, originalEcho)

	// 保存请求映射
	m.pendingMutex.Lock()
	m.pendingRequests[internalEcho] = make(chan map[string]interface{}, 1)
	m.pendingTimestamps[internalEcho] = time.Now() // 记录发送时间
	m.pendingMutex.Unlock()

	// 修改消息中的 echo 为内部 echo
	msg["echo"] = internalEcho

	// 智能路由逻辑：根据 self_id 或 group_id 选择正确的 Bot
	var targetBot *BotClient
	var routeSource string

	// 1. 尝试从请求参数中提取 self_id
	var selfID string
	if sid, ok := msg["self_id"]; ok {
		selfID = fmt.Sprintf("%v", sid)
	} else if params, ok := msg["params"].(map[string]interface{}); ok {
		if sid, ok := params["self_id"]; ok {
			selfID = fmt.Sprintf("%v", sid)
		}
	}

	m.mutex.RLock()
	if selfID != "" {
		if bot, exists := m.bots[selfID]; exists {
			targetBot = bot
			routeSource = "self_id"
		}
	}

	// 2. 如果没有 self_id，尝试根据 group_id 从缓存中查找对应的 Bot
	if targetBot == nil {
		var groupID string
		if gid, ok := msg["group_id"]; ok {
			groupID = fmt.Sprintf("%v", gid)
		} else if params, ok := msg["params"].(map[string]interface{}); ok {
			if gid, ok := params["group_id"]; ok {
				groupID = fmt.Sprintf("%v", gid)
			}
		}

		if groupID != "" {
			m.cacheMutex.RLock()
			if groupData, exists := m.groupCache[groupID]; exists {
				if botID, ok := groupData["bot_id"].(string); ok {
					if bot, exists := m.bots[botID]; exists {
						targetBot = bot
						routeSource = fmt.Sprintf("group_id cache (%s)", groupID)
					}
				}
			}
			m.cacheMutex.RUnlock()
		}
	}

	// 3. 兜底方案：如果还是找不到，选第一个可用的 Bot
	if targetBot == nil {
		for _, bot := range m.bots {
			targetBot = bot
			routeSource = "fallback (first available)"
			break
		}
	}
	m.mutex.RUnlock()

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
		m.pendingMutex.Lock()
		delete(m.pendingRequests, internalEcho)
		delete(m.pendingTimestamps, internalEcho)
		m.pendingMutex.Unlock()
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
		var fallbackBot *BotClient
		m.mutex.RLock()
		for _, bot := range m.bots {
			fallbackBot = bot
			break
		}
		m.mutex.RUnlock()

		if fallbackBot != nil {
			log.Printf("[ROUTING] Falling back to Bot %s", fallbackBot.SelfID)
			fallbackBot.Mutex.Lock()
			err = fallbackBot.Conn.WriteJSON(msg)
			fallbackBot.Mutex.Unlock()
		}

		if fallbackBot == nil || err != nil {
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
			m.pendingMutex.Lock()
			delete(m.pendingRequests, internalEcho)
			delete(m.pendingTimestamps, internalEcho)
			m.pendingMutex.Unlock()
			return
		}
	}

	log.Printf("[ROUTING] Forwarded Worker %s request (action: %v, echo: %s) to Bot %s via %s",
		worker.ID, msg["action"], originalEcho, targetBot.SelfID, routeSource)

	// 启动超时处理（30秒内必须收到响应）
	go func() {
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		select {
		case response := <-m.pendingRequests[internalEcho]:
			// 收到响应，还原原始 echo 并转发给 Worker
			if response != nil {
				// 检查响应状态，如果发现 Bot 不在群组中，清除缓存
				retcode := toInt64(response["retcode"])
				status, _ := response["status"].(string)

				if retcode == 1200 || status == "failed" {
					// 尝试提取 group_id
					var groupID string
					if gid, ok := msg["group_id"]; ok {
						groupID = fmt.Sprintf("%v", gid)
					} else if params, ok := msg["params"].(map[string]interface{}); ok {
						if gid, ok := params["group_id"]; ok {
							groupID = fmt.Sprintf("%v", gid)
						}
					}

					if groupID != "" {
						// 检查错误信息是否包含“不在群”或类似提示
						respMsg, _ := response["msg"].(string)
						if strings.Contains(respMsg, "不在群") || strings.Contains(respMsg, "移出") || retcode == 1200 {
							log.Printf("[ROUTING] Bot %s reported group error for %s: %s (retcode: %d). Clearing cache.",
								targetBot.SelfID, groupID, respMsg, retcode)

							m.cacheMutex.Lock()
							if data, exists := m.groupCache[groupID]; exists {
								if cachedBotID, _ := data["bot_id"].(string); cachedBotID == targetBot.SelfID {
									delete(m.groupCache, groupID)
									log.Printf("[ROUTING] Removed stale group cache for %s (Bot: %s)", groupID, targetBot.SelfID)
								}
							}
							m.cacheMutex.Unlock()

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
		m.pendingMutex.Lock()
		delete(m.pendingRequests, internalEcho)
		delete(m.pendingTimestamps, internalEcho)
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

		// 网络 IO 增量
		netIO, _ := net.IOCounters(false)
		if len(netIO) > 0 {
			m.NetSentTrend = append(m.NetSentTrend, netIO[0].BytesSent)
			m.NetRecvTrend = append(m.NetRecvTrend, netIO[0].BytesRecv)
		}

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
			if len(m.NetSentTrend) > 0 {
				m.NetSentTrend = m.NetSentTrend[1:]
			}
			if len(m.NetRecvTrend) > 0 {
				m.NetRecvTrend = m.NetRecvTrend[1:]
			}
		}

		m.HistoryMutex.Unlock()
	}
}

// ==================== Docker 管理接口 ====================

// handleDockerList 获取 Docker 容器列表
func (m *Manager) handleDockerList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.dockerClient == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Docker 客户端未初始化",
		})
		return
	}

	containers, err := m.dockerClient.ContainerList(r.Context(), types.ContainerListOptions{All: true})
	if err != nil {
		log.Printf("[ERROR] 获取 Docker 容器列表失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(containers)
}

// handleDockerAction 处理 Docker 容器操作 (start/stop/restart)
func (m *Manager) handleDockerAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ContainerID string `json:"container_id"`
		Action      string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "请求格式错误",
		})
		return
	}

	if m.dockerClient == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Docker 客户端未初始化",
		})
		return
	}

	var err error
	switch req.Action {
	case "start":
		err = m.dockerClient.ContainerStart(r.Context(), req.ContainerID, types.ContainerStartOptions{})
	case "stop":
		timeout := 10
		err = m.dockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
	case "restart":
		timeout := 10
		err = m.dockerClient.ContainerRestart(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
	default:
		err = fmt.Errorf("不支持的操作: %s", req.Action)
	}

	if err != nil {
		log.Printf("[ERROR] Docker 操作 %s 失败: %v", req.Action, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"id":     req.ContainerID,
	})
}

// handleDockerAddBot 添加机器人容器 (演示用，实际需要根据配置创建)
func (m *Manager) handleDockerAddBot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// 这里可以实现根据参数创建 WxBot 或其他类型的 Bot 容器
	// 暂时返回一个模拟成功，实际逻辑需要根据 docker-compose.yml 里的配置来创建
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": "自动部署机器人功能正在开发中，请手动使用 docker-compose 部署",
	})
}

// handleDockerAddWorker 添加 Worker 容器
func (m *Manager) handleDockerAddWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": "自动部署 Worker 功能正在开发中，请手动使用 docker-compose 部署",
	})
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

// handleAdminListUsers 获取所有用户列表 (仅限管理员)
func (m *Manager) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.usersMutex.RLock()
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	m.usersMutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

// handleAdminCreateUser 创建新用户 (仅限管理员)
func (m *Manager) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	if data.Username == "" || data.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名和密码不能为空",
		})
		return
	}

	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	if _, exists := m.users[data.Username]; exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户已存在",
		})
		return
	}

	hashedPassword, err := hashPassword(data.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "密码哈希失败",
		})
		return
	}

	newUser := &User{
		Username:       data.Username,
		PasswordHash:   hashedPassword,
		IsAdmin:        data.IsAdmin,
		SessionVersion: 1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := m.saveUserToDB(newUser); err != nil {
		log.Printf("保存新用户到数据库失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存用户失败",
		})
		return
	}

	m.users[newUser.Username] = newUser

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户创建成功",
		"user":    newUser,
	})
}

// handleAdminDeleteUser 删除用户 (仅限管理员)
func (m *Manager) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未提供用户名",
		})
		return
	}

	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "不能删除默认管理员",
		})
		return
	}

	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	if _, exists := m.users[username]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	// 从数据库删除
	if _, err := m.db.Exec("DELETE FROM users WHERE username = ?", username); err != nil {
		log.Printf("从数据库删除用户失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "删除用户失败",
		})
		return
	}

	delete(m.users, username)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户删除成功",
	})
}

// handleAdminResetPassword 重置用户密码 (仅限管理员)
func (m *Manager) handleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data struct {
		Username    string `json:"username"`
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

	user, exists := m.users[data.Username]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	hashedPassword, err := hashPassword(data.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "密码哈希失败",
		})
		return
	}

	user.PasswordHash = hashedPassword
	user.SessionVersion++ // 强制该用户重新登录
	user.UpdatedAt = time.Now()

	if err := m.saveUserToDB(user); err != nil {
		log.Printf("重置密码保存到数据库失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存密码失败",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "密码重置成功",
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

	// 保存到数据库
	if err := m.saveRoutingRuleToDB(rule.Key, rule.WorkerID); err != nil {
		log.Printf("[ERROR] 保存路由规则到数据库失败: %v", err)
	}

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
		// 从数据库删除
		if err := m.deleteRoutingRuleFromDB(key); err != nil {
			log.Printf("[ERROR] 从数据库删除路由规则失败: %v", err)
		}
		log.Printf("[ADMIN] Deleted routing rule for key: %s", key)
	}
	m.mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "路由规则删除成功",
	})
}
