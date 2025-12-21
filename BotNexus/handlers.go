package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"strconv"
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

	// totalBots 应该是在数据库中或统计列表中的所有机器人
	// 这里使用 m.BotStats 作为基准，它记录了所有见过的机器人
	totalBots := len(m.BotStats)

	// 修正逻辑：如果当前在线的大于总数（可能由于内存状态未同步），则更新总数
	if onlineBots > totalBots {
		totalBots = onlineBots
	}

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
	cpuCount, _ := cpu.Counts(true)
	if cpuCount <= 0 {
		cpuCount = 1
	}

	cpuPercent, _ := cpu.Percent(0, false)
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
		// 如果 CPU 使用率超过 100 且有多个核心，说明是总和，需要归一化
		if cpuUsage > 100 && cpuCount > 1 {
			cpuUsage = cpuUsage / float64(cpuCount)
		}
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
	cpuCount, _ := cpu.Counts(true)
	if cpuCount <= 0 {
		cpuCount = 1
	}

	cpuPercent, _ := cpu.Percent(time.Second, false)
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
		// 如果 CPU 使用率超过 100 且有多个核心，说明是总和，需要归一化
		if cpuUsage > 100 && cpuCount > 1 {
			cpuUsage = cpuUsage / float64(cpuCount)
		}
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

	// 获取缓存的 Top 进程
	m.HistoryMutex.RLock()
	processList := m.TopProcesses
	m.HistoryMutex.RUnlock()

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

// handleGetBots 处理获取机器人列表的请求
func (m *Manager) handleGetBots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	bots := make([]map[string]interface{}, 0, len(m.bots))
	for id, bot := range m.bots {
		// 获取远程地址
		remoteAddr := ""
		if bot.Conn != nil {
			remoteAddr = bot.Conn.RemoteAddr().String()
		}

		// 获取统计信息
		totalMsg := m.BotStats[id]
		todayMsg := m.BotStatsToday[id]

		bots = append(bots, map[string]interface{}{
			"id":              id,
			"self_id":         bot.SelfID,
			"nickname":        bot.Nickname,
			"group_count":     bot.GroupCount,
			"friend_count":    bot.FriendCount,
			"connected":       bot.Connected.Format("2006-01-02 15:04:05"),
			"platform":        bot.Platform,
			"sent_count":      bot.SentCount,
			"recv_count":      bot.RecvCount,
			"msg_count":       totalMsg,
			"msg_count_today": todayMsg,
			"remote_addr":     remoteAddr,
			"last_heartbeat":  bot.LastHeartbeat.Format("2006-01-02 15:04:05"),
			"is_alive":        true,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"bots":    bots,
	})
}

// handleGetWorkers 处理获取Worker列表的请求
func (m *Manager) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	workers := make([]map[string]interface{}, 0, len(m.workers))
	for _, worker := range m.workers {
		workers = append(workers, map[string]interface{}{
			"id":            worker.ID,
			"remote_addr":   worker.ID,
			"connected":     worker.Connected.Format("2006-01-02 15:04:05"),
			"handled_count": worker.HandledCount,
			"avg_rtt":       worker.AvgRTT.String(),
			"last_rtt":      worker.LastRTT.String(),
			"is_alive":      true,
			"status":        "Online",
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"workers": workers,
	})
}

// handleGetContacts 处理获取联系人/群组信息的请求
func (m *Manager) handleGetContacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	botID := r.URL.Query().Get("bot_id")
	refresh := r.URL.Query().Get("refresh") == "true"

	if refresh && botID != "" {
		m.mutex.RLock()
		bot, ok := m.bots[botID]
		m.mutex.RUnlock()

		if ok {
			// 如果需要刷新，先尝试从机器人获取最新数据
			// 注意：这里是同步等待，但带超时
			log.Printf("[Contacts] Refreshing contacts for bot: %s", botID)

			// 1. 获取群列表
			echoGroups := "refresh_groups_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.pendingMutex.Lock()
			respChanGroups := make(chan map[string]interface{}, 1)
			m.pendingRequests[echoGroups] = respChanGroups
			m.pendingMutex.Unlock()

			bot.Mutex.Lock()
			bot.Conn.WriteJSON(map[string]interface{}{
				"action": "get_group_list",
				"params": map[string]interface{}{},
				"echo":   echoGroups,
			})
			bot.Mutex.Unlock()

			// 2. 获取好友列表
			echoFriends := "refresh_friends_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.pendingMutex.Lock()
			respChanFriends := make(chan map[string]interface{}, 1)
			m.pendingRequests[echoFriends] = respChanFriends
			m.pendingMutex.Unlock()

			bot.Mutex.Lock()
			bot.Conn.WriteJSON(map[string]interface{}{
				"action": "get_friend_list",
				"params": map[string]interface{}{},
				"echo":   echoFriends,
			})
			bot.Mutex.Unlock()

			// 3. 获取频道列表 (Guild)
			echoGuilds := "refresh_guilds_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.pendingMutex.Lock()
			respChanGuilds := make(chan map[string]interface{}, 1)
			m.pendingRequests[echoGuilds] = respChanGuilds
			m.pendingMutex.Unlock()

			bot.Mutex.Lock()
			bot.Conn.WriteJSON(map[string]interface{}{
				"action": "get_guild_list",
				"params": map[string]interface{}{},
				"echo":   echoGuilds,
			})
			bot.Mutex.Unlock()

			// 等待响应，最长 5 秒
			timeout := time.After(5 * time.Second)
			groupsDone := false
			friendsDone := false
			guildsDone := false

			for !groupsDone || !friendsDone || !guildsDone {
				select {
				case resp := <-respChanGroups:
					if data, ok := resp["data"].([]interface{}); ok {
						m.cacheMutex.Lock()
						for _, item := range data {
							if group, ok := item.(map[string]interface{}); ok {
								gID := toString(group["group_id"])
								// 深度复制所有字段，确保 member_count 等信息不丢失
								cachedGroup := make(map[string]interface{})
								for k, v := range group {
									cachedGroup[k] = v
								}
								cachedGroup["id"] = gID
								cachedGroup["name"] = toString(group["group_name"])
								cachedGroup["bot_id"] = botID
								if cachedGroup["type"] == nil {
									cachedGroup["type"] = "group"
								}
								cachedGroup["last_seen"] = time.Now()
								m.groupCache[gID] = cachedGroup
								go m.saveGroupToDB(gID, toString(group["group_name"]), botID)
							}
						}
						m.cacheMutex.Unlock()
					} else if dataMap, ok := resp["data"].(map[string]interface{}); ok {
						// 兼容某些非标准实现返回 {"groups": [...]} 的情况
						if groups, ok := dataMap["groups"].([]interface{}); ok {
							m.cacheMutex.Lock()
							for _, item := range groups {
								if group, ok := item.(map[string]interface{}); ok {
									gID := toString(group["group_id"])
									cachedGroup := make(map[string]interface{})
									for k, v := range group {
										cachedGroup[k] = v
									}
									cachedGroup["id"] = gID
									cachedGroup["name"] = toString(group["group_name"])
									cachedGroup["bot_id"] = botID
									if cachedGroup["type"] == nil {
										cachedGroup["type"] = "group"
									}
									cachedGroup["last_seen"] = time.Now()
									m.groupCache[gID] = cachedGroup
									go m.saveGroupToDB(gID, toString(group["group_name"]), botID)
								}
							}
							m.cacheMutex.Unlock()
						}
					}
					m.pendingMutex.Lock()
					delete(m.pendingRequests, echoGroups)
					m.pendingMutex.Unlock()
					groupsDone = true
				case resp := <-respChanFriends:
					if data, ok := resp["data"].([]interface{}); ok {
						m.cacheMutex.Lock()
						for _, item := range data {
							if friend, ok := item.(map[string]interface{}); ok {
								uID := toString(friend["user_id"])
								cachedFriend := make(map[string]interface{})
								for k, v := range friend {
									cachedFriend[k] = v
								}
								cachedFriend["id"] = uID
								cachedFriend["name"] = toString(friend["nickname"])
								cachedFriend["bot_id"] = botID
								if cachedFriend["type"] == nil {
									cachedFriend["type"] = "private"
								}
								cachedFriend["last_seen"] = time.Now()
								m.friendCache[uID] = cachedFriend
								go m.saveFriendToDB(uID, toString(friend["nickname"]))
							}
						}
						m.cacheMutex.Unlock()
					} else if dataMap, ok := resp["data"].(map[string]interface{}); ok {
						// 兼容 {"friends": [...]}
						if friends, ok := dataMap["friends"].([]interface{}); ok {
							m.cacheMutex.Lock()
							for _, item := range friends {
								if friend, ok := item.(map[string]interface{}); ok {
									uID := toString(friend["user_id"])
									cachedFriend := make(map[string]interface{})
									for k, v := range friend {
										cachedFriend[k] = v
									}
									cachedFriend["id"] = uID
									cachedFriend["name"] = toString(friend["nickname"])
									cachedFriend["bot_id"] = botID
									if cachedFriend["type"] == nil {
										cachedFriend["type"] = "private"
									}
									cachedFriend["last_seen"] = time.Now()
									m.friendCache[uID] = cachedFriend
									go m.saveFriendToDB(uID, toString(friend["nickname"]))
								}
							}
							m.cacheMutex.Unlock()
						}
					}
					m.pendingMutex.Lock()
					delete(m.pendingRequests, echoFriends)
					m.pendingMutex.Unlock()
					friendsDone = true
				case resp := <-respChanGuilds:
					if data, ok := resp["data"].([]interface{}); ok {
						m.cacheMutex.Lock()
						for _, item := range data {
							if guild, ok := item.(map[string]interface{}); ok {
								gID := toString(guild["guild_id"])
								cachedGuild := make(map[string]interface{})
								for k, v := range guild {
									cachedGuild[k] = v
								}
								cachedGuild["id"] = gID
								cachedGuild["name"] = toString(guild["guild_name"])
								cachedGuild["bot_id"] = botID
								cachedGuild["type"] = "guild"
								cachedGuild["last_seen"] = time.Now()
								m.groupCache[gID] = cachedGuild
							}
						}
						m.cacheMutex.Unlock()
					}
					m.pendingMutex.Lock()
					delete(m.pendingRequests, echoGuilds)
					m.pendingMutex.Unlock()
					guildsDone = true
				case <-timeout:
					log.Printf("[Contacts] Refresh timeout for bot: %s", botID)
					groupsDone = true
					friendsDone = true
					guildsDone = true
				}
			}
		}
	}

	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	// 整合群组和好友信息
	contacts := make([]map[string]interface{}, 0)

	for _, group := range m.groupCache {
		if botID == "" || group["bot_id"] == botID {
			item := make(map[string]interface{})
			for k, v := range group {
				item[k] = v
			}
			// 确保有通用的 name 和 id 字段供前端使用
			if name, ok := item["group_name"]; ok {
				item["name"] = name
			} else if name, ok := item["guild_name"]; ok {
				item["name"] = name
			}

			if id, ok := item["group_id"]; ok {
				item["id"] = id
			} else if id, ok := item["guild_id"]; ok {
				item["id"] = id
			}
			contacts = append(contacts, item)
		}
	}

	for _, friend := range m.friendCache {
		if botID == "" || friend["bot_id"] == botID {
			item := make(map[string]interface{})
			for k, v := range friend {
				item[k] = v
			}
			// 确保有通用的 name 和 id 字段供前端使用
			if nickname, ok := item["nickname"]; ok {
				item["name"] = nickname
			}
			if userID, ok := item["user_id"]; ok {
				item["id"] = userID
			}
			contacts = append(contacts, item)
		}
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
							m.groupCache[gID] = cachedGroup
							// 持久化到数据库
							go m.saveGroupToDB(gID, name, bot.SelfID)
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

// sendWorkerHeartbeat 定期发送心跳包给Worker
func (m *Manager) sendWorkerHeartbeat(worker *WorkerClient, stop chan struct{}) {
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
func (m *Manager) handleBotMessage(bot *BotClient, msg map[string]interface{}) {
	// 检查是否包含 self_id 并更新（如果当前是临时 ID）
	msgSelfID := toString(msg["self_id"])
	if msgSelfID != "" {
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
	echo := toString(msg["echo"])
	if echo != "" {
		// 广播路由事件: Bot -> Nexus (Response)
		m.broadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "response")

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

					// 广播路由事件: Nexus -> Worker (Response Forward)
					m.broadcastRoutingEvent("Nexus", workerID, "nexus_to_worker", "response")
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
	msgType := toString(msg["post_type"])
	if msgType == "" {
		log.Printf("[Bot] Warning: Received message from Bot %s without echo or post_type: %v", bot.SelfID, msg)
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
	// 确保消息格式符合 OneBot 11 标准 (特别是 ID 类型)
	if userID, ok := msg["user_id"]; ok {
		msg["user_id"] = toString(userID)
	}
	if groupID, ok := msg["group_id"]; ok {
		msg["group_id"] = toString(groupID)
	}

	// 补全并标准化 self_id (OneBot 11 标准应为 int64，但我们兼容字符串)
	if selfID, ok := msg["self_id"]; ok {
		msg["self_id"] = toString(selfID)
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
func (m *Manager) cacheBotDataFromMessage(bot *BotClient, msg map[string]interface{}) {
	postType, _ := msg["post_type"].(string)
	if postType != "message" {
		return
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// 缓存群信息
	if groupIDVal, ok := msg["group_id"]; ok && groupIDVal != nil {
		gID := toString(groupIDVal)

		// 如果缓存中已存在该群组，则保留原有名称，除非原有名称包含 "(Cached)" 而当前消息提供了更准确的信息
		existingGroup, exists := m.groupCache[gID]
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
		m.groupCache[gID] = map[string]interface{}{
			"group_id":   gID,
			"group_name": groupName,
			"bot_id":     bot.SelfID,
			"is_cached":  true,
			"reason":     "Automatically updated from message",
			"last_seen":  time.Now(),
		}
		// 持久化到数据库
		go m.saveGroupToDB(gID, groupName, bot.SelfID)

		// 缓存成员信息
		if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
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
				"group_id":  gID,
				"user_id":   uID,
				"nickname":  nickname,
				"card":      card,
				"is_cached": true,
			}
			// 持久化到数据库
			go m.saveMemberToDB(gID, uID, nickname, card)
		}
	} else if userIDVal, ok := msg["user_id"]; ok && userIDVal != nil {
		// 缓存好友信息 (私聊)
		uID := toString(userIDVal)
		if _, exists := m.friendCache[uID]; !exists {
			sender, _ := msg["sender"].(map[string]interface{})
			nickname := ""
			if sender != nil {
				nickname, _ = sender["nickname"].(string)
			}
			m.friendCache[uID] = map[string]interface{}{
				"user_id":   uID,
				"nickname":  nickname,
				"is_cached": true,
			}
			// 持久化到数据库
			go m.saveFriendToDB(uID, nickname)
		}
	}
}

// isSystemMessage 检查是否为系统消息或无意义的统计数据
func isSystemMessage(msg map[string]interface{}) bool {
	// 1. 提取基本字段，支持直接结构和 params 嵌套结构
	userID := toString(msg["user_id"])
	msgType := toString(msg["message_type"])
	subType := toString(msg["sub_type"])
	message := ""

	// 尝试从 params 中提取 (如果是发送请求)
	if params, ok := msg["params"].(map[string]interface{}); ok {
		if userID == "" {
			userID = toString(params["user_id"])
		}
		if msgType == "" {
			msgType = toString(params["message_type"])
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

// handleBotMessageEvent 处理Bot消息事件
func (m *Manager) handleBotMessageEvent(bot *BotClient, msg map[string]interface{}) {
	// 提取消息信息
	userID := toString(msg["user_id"])
	groupID := toString(msg["group_id"])

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
		m.cacheMutex.RLock()
		if group, ok := m.groupCache[groupID]; ok {
			if name, ok := group["group_name"].(string); ok {
				extras["group_name"] = name
			}
		}
		m.cacheMutex.RUnlock()
	}

	// 1. 广播路由事件: User -> Bot (如果用户 ID 存在)
	if userID != "" {
		m.broadcastRoutingEvent(userID, bot.SelfID, "user_to_bot", "message", extras)
	}

	// 2. 广播路由事件: Bot -> Nexus
	m.broadcastRoutingEvent(bot.SelfID, "Nexus", "bot_to_nexus", "message", extras)

	// 更新详细统计 (排除系统消息)
	if !isSystemMessage(msg) {
		m.updateBotStats(bot.SelfID, userID, groupID)
	}
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
				m.mutex.Lock()
				w.HandledCount++
				m.mutex.Unlock()

				// 广播路由事件: Nexus -> Worker (Message Forward - Rule Match)
				m.broadcastRoutingEvent("Nexus", targetWorkerID, "nexus_to_worker", "message")
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

	if err == nil {
		// 记录发送给 Worker 的时间，用于计算处理耗时
		echo := toString(msg["echo"])
		if echo != "" {
			m.workerRequestMutex.Lock()
			m.workerRequestTimes[echo] = time.Now()
			m.workerRequestMutex.Unlock()
		}
	}

	if err != nil {
		log.Printf("[ROUTING] [ERROR] Failed to forward to worker %s: %v. Removing and retrying...", selectedWorker.ID, err)
		m.removeWorker(selectedWorker.ID)
		m.forwardMessageToWorkerWithRetry(msg, retryCount+1)
	} else {
		m.mutex.Lock()
		selectedWorker.HandledCount++
		m.mutex.Unlock()

		// 广播路由事件: Nexus -> Worker (Message Forward - LB)
		m.broadcastRoutingEvent("Nexus", selectedWorker.ID, "nexus_to_worker", "message")

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
func (m *Manager) handleWorkerConnection(worker *WorkerClient) {
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
	msgType := toString(msg["type"])
	action := toString(msg["action"])
	echo := toString(msg["echo"])

	if action != "" || (echo != "" && msgType == "") {
		// 这是一个Worker发起的API请求，需要转发给Bot
		log.Printf("Worker %s API request: action=%s, echo=%s", worker.ID, action, echo)

		// 广播路由事件: Worker -> Nexus (Request)
		m.broadcastRoutingEvent(worker.ID, "Nexus", "worker_to_nexus", "request")

		m.forwardWorkerRequestToBot(worker, msg, echo)
	} else {
		log.Printf("Worker %s event/response: type=%s", worker.ID, msgType)

		// 更新统计信息
		m.mutex.Lock()
		worker.HandledCount++
		m.mutex.Unlock()

		// 1. 统计处理耗时
		if echo != "" {
			m.workerRequestMutex.Lock()
			if startTime, exists := m.workerRequestTimes[echo]; exists {
				duration := time.Since(startTime)
				delete(m.workerRequestTimes, echo)
				m.workerRequestMutex.Unlock()

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
				m.workerRequestMutex.Unlock()
			}
		}

		// 2. 检查是否包含回复内容 (一些框架允许在事件响应中直接返回回复)
		reply := toString(msg["reply"])
		if reply != "" {
			log.Printf("Worker %s sent passive reply: %s", worker.ID, reply)
			// 构造一个 send_msg 请求转发给 Bot
			m.handleWorkerPassiveReply(worker, msg)
		}
	}
}

// handleWorkerPassiveReply 处理Worker的被动回复
func (m *Manager) handleWorkerPassiveReply(worker *WorkerClient, msg map[string]interface{}) {
	// 提取 echo (如果 Worker 在被动回复中带了 echo)
	echo := toString(msg["echo"])

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

	// 发送初始同步状态
	m.mutex.RLock()
	bots := make([]BotClient, 0, len(m.bots))
	for _, bot := range m.bots {
		bots = append(bots, *bot)
	}
	workers := make([]WorkerInfo, 0, len(m.workers))
	for _, w := range m.workers {
		workers = append(workers, WorkerInfo{
			ID:       w.ID,
			Type:     "worker",
			Status:   "online",
			LastSeen: time.Now().Format("15:04:05"),
		})
	}
	m.mutex.RUnlock()

	m.cacheMutex.RLock()
	syncState := SyncState{
		Type:          "sync_state",
		Groups:        m.groupCache,
		Friends:       m.friendCache,
		Members:       m.memberCache,
		Bots:          bots,
		Workers:       workers,
		TotalMessages: m.TotalMessages,
	}
	m.cacheMutex.RUnlock()

	subscriber.Mutex.Lock()
	if err := conn.WriteJSON(syncState); err != nil {
		log.Printf("[SUBSCRIBER] 发送初始同步状态失败: %v", err)
	}
	subscriber.Mutex.Unlock()

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

// broadcastRoutingEvent 向所有订阅者广播路由事件
func (m *Manager) broadcastRoutingEvent(source, target, direction, msgType string, extras ...interface{}) {
	m.mutex.RLock()
	event := RoutingEvent{
		Type:          "routing_event",
		Source:        source,
		Target:        target,
		Direction:     direction,
		MsgType:       msgType,
		Timestamp:     time.Now(),
		TotalMessages: m.TotalMessages,
	}

	// Determine Source Type and Label
	if source == "Nexus" {
		event.SourceType = "nexus"
		event.SourceLabel = "NEXUS"
	} else if bot, ok := m.bots[source]; ok {
		event.SourceType = "bot"
		event.SourceLabel = bot.Nickname
		if event.SourceLabel == "" {
			event.SourceLabel = bot.SelfID
		}
	} else {
		isWorker := false
		for _, w := range m.workers {
			if w.ID == source {
				event.SourceType = "worker"
				event.SourceLabel = "WORKER"
				isWorker = true
				break
			}
		}
		if !isWorker {
			event.SourceType = "user"
			event.SourceLabel = source
		}
	}

	// Determine Target Type and Label
	if target == "Nexus" {
		event.TargetType = "nexus"
		event.TargetLabel = "NEXUS"
	} else if bot, ok := m.bots[target]; ok {
		event.TargetType = "bot"
		event.TargetLabel = bot.Nickname
		if event.TargetLabel == "" {
			event.TargetLabel = bot.SelfID
		}
	} else {
		isWorker := false
		for _, w := range m.workers {
			if w.ID == target {
				event.TargetType = "worker"
				event.TargetLabel = "WORKER"
				isWorker = true
				break
			}
		}
		if !isWorker {
			event.TargetType = "user"
			event.TargetLabel = target
		}
	}
	m.mutex.RUnlock()

	// Handle extra data
	if len(extras) > 0 {
		if data, ok := extras[0].(map[string]interface{}); ok {
			if uid, ok := data["user_id"]; ok {
				event.UserID = toString(uid)
			}
			if uname, ok := data["user_name"]; ok {
				event.UserName = toString(uname)
			}
			if uavatar, ok := data["user_avatar"]; ok {
				event.UserAvatar = toString(uavatar)
			}
			if content, ok := data["content"]; ok {
				event.Content = toString(content)
			}
			if platform, ok := data["platform"]; ok {
				event.Platform = toString(platform)
			}
			if gid, ok := data["group_id"]; ok {
				event.GroupID = toString(gid)
			}
			if gname, ok := data["group_name"]; ok {
				event.GroupName = toString(gname)
			}
			if color, ok := data["color"]; ok {
				// Allow passing custom color via extras
				// Frontend will use it if present
			}
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, sub := range m.subscribers {
		go func(s *Subscriber, e RoutingEvent) {
			s.Mutex.Lock()
			defer s.Mutex.Unlock()
			err := s.Conn.WriteJSON(e)
			if err != nil {
				log.Printf("[SUBSCRIBER] Failed to send routing event to subscriber: %v", err)
			}
		}(sub, event)
	}
}

// broadcastDockerEvent 向所有订阅者广播 Docker 事件
func (m *Manager) broadcastDockerEvent(action, containerID, status string) {
	event := DockerEvent{
		Type:        "docker_event",
		Action:      action,
		ContainerID: containerID,
		Status:      status,
		Timestamp:   time.Now(),
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, sub := range m.subscribers {
		go func(s *Subscriber, e DockerEvent) {
			s.Mutex.Lock()
			defer s.Mutex.Unlock()
			err := s.Conn.WriteJSON(e)
			if err != nil {
				log.Printf("[SUBSCRIBER] Failed to send docker event to subscriber: %v", err)
			}
		}(sub, event)
	}
}

// forwardWorkerRequestToBot 将Worker请求转发给Bot
func (m *Manager) forwardWorkerRequestToBot(worker *WorkerClient, msg map[string]interface{}, originalEcho string) {
	// 构造内部 echo，包含 worker ID 以便追踪和记录 RTT
	// 加上时间戳确保即使 originalEcho 为空或重复，internalEcho 也是唯一的
	internalEcho := fmt.Sprintf("%s|%s|%d", worker.ID, originalEcho, time.Now().UnixNano())

	// 保存请求映射
	respChan := make(chan map[string]interface{}, 1)
	m.pendingMutex.Lock()
	m.pendingRequests[internalEcho] = respChan
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
		selfID = toString(sid)
	} else if params, ok := msg["params"].(map[string]interface{}); ok {
		if sid, ok := params["self_id"]; ok {
			selfID = toString(sid)
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
			groupID = toString(gid)
		} else if params, ok := msg["params"].(map[string]interface{}); ok {
			if gid, ok := params["group_id"]; ok {
				groupID = toString(gid)
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
			m.pendingMutex.Lock()
			delete(m.pendingRequests, internalEcho)
			delete(m.pendingTimestamps, internalEcho)
			m.pendingMutex.Unlock()
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
			extras["content"] = toString(content)
		}
		if userID, ok := params["user_id"]; ok {
			extras["user_id"] = toString(userID)
		}
	}

	m.broadcastRoutingEvent("Nexus", targetBot.SelfID, "nexus_to_bot", "request", extras)

	// 更新发送统计 (如果是发送消息类操作)
	action := toString(msg["action"])
	if action == "send_msg" || action == "send_private_msg" || action == "send_group_msg" || action == "send_guild_channel_msg" {
		// 增加排除系统消息的判断
		if !isSystemMessage(msg) {
			m.updateBotSentStats(targetBot.SelfID)
		}
	}

	// 如果是发送消息，额外广播一个从 Bot 到 User 的事件，让特效闭环
	if userID, ok := extras["user_id"].(string); ok && userID != "" {
		m.broadcastRoutingEvent(targetBot.SelfID, userID, "bot_to_user", "message", extras)
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
				retcode := toInt64(response["retcode"])
				status, _ := response["status"].(string)

				if retcode == 1200 || status == "failed" {
					// 尝试提取 group_id
					var groupID string
					if gid, ok := msg["group_id"]; ok {
						groupID = toString(gid)
					} else if params, ok := msg["params"].(map[string]interface{}); ok {
						if gid, ok := params["group_id"]; ok {
							groupID = toString(gid)
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
	// 每 5 分钟保存一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 每天凌晨清空今日统计
	midnightTicker := time.NewTicker(1 * time.Hour)
	defer midnightTicker.Stop()

	for {
		select {
		case <-ticker.C:
			m.saveAllStatsToDB()
		case <-midnightTicker.C:
			// 检查是否跨天 (更稳健的检查方法)
			now := time.Now()
			currentDate := now.Format("2006-01-02")
			
			m.statsMutex.Lock()
			if m.LastResetDate != currentDate {
				log.Printf("[STATS] 检测到跨天，执行每日统计重置: %s -> %s", m.LastResetDate, currentDate)
				m.UserStatsToday = make(map[string]int64)
				m.GroupStatsToday = make(map[string]int64)
				m.BotStatsToday = make(map[string]int64)
				m.LastResetDate = currentDate
				
				// 重置后立即保存一次数据库，清空数据库中的今日统计
				m.statsMutex.Unlock()
				m.saveAllStatsToDB()
			} else {
				m.statsMutex.Unlock()
			}
		}
	}
}

// updateBotStats 更新Bot统计信息
func (m *Manager) updateBotStats(botID string, userID, groupID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// 初始化统计数据结构 (如果需要)
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}
	if m.UserStats == nil {
		m.UserStats = make(map[string]int64)
	}
	if m.GroupStats == nil {
		m.GroupStats = make(map[string]int64)
	}
	if m.BotStats == nil {
		m.BotStats = make(map[string]int64)
	}
	if m.UserStatsToday == nil {
		m.UserStatsToday = make(map[string]int64)
	}
	if m.GroupStatsToday == nil {
		m.GroupStatsToday = make(map[string]int64)
	}
	if m.BotStatsToday == nil {
		m.BotStatsToday = make(map[string]int64)
	}

	// 更新Bot详细统计
	if _, exists := m.BotDetailedStats[botID]; !exists {
		m.BotDetailedStats[botID] = &BotStatDetail{
			Users:  make(map[string]int64),
			Groups: make(map[string]int64),
		}
	}

	stats := m.BotDetailedStats[botID]
	stats.Received++
	stats.LastMsg = time.Now()

	if userID != "" && userID != "0" {
		stats.Users[userID]++
		m.UserStats[userID]++
		m.UserStatsToday[userID]++
	}
	if groupID != "" && groupID != "0" {
		stats.Groups[groupID]++
		m.GroupStats[groupID]++
		m.GroupStatsToday[groupID]++
	}

	// 更新全局和今日统计
	m.BotStats[botID]++
	m.BotStatsToday[botID]++
	m.TotalMessages++

	// 持久化到数据库
	go m.saveStatToDB("total_messages", m.TotalMessages)
}

// updateBotSentStats 更新发送消息统计
func (m *Manager) updateBotSentStats(botID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// 初始化统计数据结构 (如果需要)
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}
	if m.BotStatsSent == nil {
		m.BotStatsSent = make(map[string]int64)
	}

	// 更新Bot详细统计
	if _, exists := m.BotDetailedStats[botID]; !exists {
		m.BotDetailedStats[botID] = &BotStatDetail{
			Users:  make(map[string]int64),
			Groups: make(map[string]int64),
		}
	}

	stats := m.BotDetailedStats[botID]
	stats.Sent++

	// 更新全局和今日统计
	m.BotStatsSent[botID]++
	m.SentMessages++
	m.TotalMessages++

	// 持久化到数据库
	go m.saveStatToDB("total_messages", m.TotalMessages)
	go m.saveStatToDB("sent_messages", m.SentMessages)
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

		// 获取 CPU 核心数用于归一化处理，防止超过 100%
		cpuCount, err := cpu.Counts(true)
		if err != nil || cpuCount <= 0 {
			cpuCount = runtime.NumCPU()
			if cpuCount <= 0 {
				cpuCount = 1
			}
		}

		cpuPercent, _ := cpu.Percent(0, false)
		var currentCPU float64
		if len(cpuPercent) > 0 {
			currentCPU = cpuPercent[0]
			// 如果 CPU 使用率超过 100 且有多个核心，说明是总和，需要归一化
			if currentCPU > 100 && cpuCount > 1 {
				currentCPU = currentCPU / float64(cpuCount)
			}
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

		// 计算增量 (用于趋势图)
		var deltaTotal, deltaSent int64
		if m.lastTrendTotal > 0 {
			deltaTotal = total - m.lastTrendTotal
			if deltaTotal < 0 {
				deltaTotal = 0
			}
		}
		if m.lastTrendSent > 0 {
			deltaSent = sent - m.lastTrendSent
			if deltaSent < 0 {
				deltaSent = 0
			}
		}

		// 如果是第一次收集，或者 total 为 0，则 delta 为 0
		if m.lastTrendTotal == 0 {
			deltaTotal = 0
		}
		if m.lastTrendSent == 0 {
			deltaSent = 0
		}

		m.lastTrendTotal = total
		m.lastTrendSent = sent

		// 存入趋势数组
		m.MsgTrend = append(m.MsgTrend, deltaTotal)
		m.SentTrend = append(m.SentTrend, deltaSent)
		m.RecvTrend = append(m.RecvTrend, deltaTotal-deltaSent)

		// 获取 Top 进程
		procs, _ := process.Processes()
		var allProcs []ProcInfo

		newProcMap := make(map[int32]*process.Process)
		for _, p := range procs {
			// 只追踪 CPU 较高的前 200 个进程
			name, _ := p.Name()

			// 尝试从缓存获取，以获得准确的 CPU 百分比
			var procObj *process.Process
			if cached, ok := m.procMap[p.Pid]; ok {
				procObj = cached
			} else {
				procObj = p
			}
			newProcMap[p.Pid] = procObj

			cp, _ := procObj.CPUPercent()
			// 归一化处理：除以核心数
			cp = cp / float64(cpuCount)

			mp, _ := procObj.MemoryInfo()
			if mp != nil {
				allProcs = append(allProcs, ProcInfo{
					Pid:    p.Pid,
					Name:   name,
					CPU:    cp,
					Memory: mp.RSS,
				})
			}
		}
		m.procMap = newProcMap

		// 按 CPU 排序
		sort.Slice(allProcs, func(i, j int) bool {
			return allProcs[i].CPU > allProcs[j].CPU
		})
		if len(allProcs) > 10 {
			m.TopProcesses = allProcs[:10]
		} else {
			m.TopProcesses = allProcs
		}

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
	case "delete":
		// 先停止容器再删除
		timeout := 5
		m.dockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
		err = m.dockerClient.ContainerRemove(r.Context(), req.ContainerID, types.ContainerRemoveOptions{Force: true})
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

	// 广播 Docker 事件
	status := "running"
	if req.Action == "stop" {
		status = "exited"
	} else if req.Action == "delete" {
		status = "deleted"
	}
	m.broadcastDockerEvent(req.Action, req.ContainerID, status)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"id":     req.ContainerID,
	})
}

// handleDockerAddBot 添加机器人容器
func (m *Manager) handleDockerAddBot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.dockerClient == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Docker 客户端未初始化",
		})
		return
	}

	ctx := context.Background()
	imageName := "botmatrix-wxbot" // 假设已经构建好的镜像名

	// 1. 检查镜像是否存在，不存在则尝试拉取
	_, _, err := m.dockerClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		log.Printf("[Docker] 镜像 %s 在本地未找到，正在尝试从仓库拉取...", imageName)
		reader, err := m.dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			log.Printf("[Docker] 无法拉取镜像 %s: %v", imageName, err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("镜像 %s 不存在且无法拉取。请确保镜像已正确构建或存在于仓库中。错误: %v", imageName, err),
			})
			return
		}
		defer reader.Close()
		io.Copy(io.Discard, reader)
		log.Printf("[Docker] 镜像 %s 拉取成功", imageName)
	}

	// 2. 生成唯一的容器名
	containerName := fmt.Sprintf("wxbot-%d", time.Now().Unix())

	// 3. 配置容器
	config := &container.Config{
		Image: imageName,
		Env: []string{
			"TZ=Asia/Shanghai",
			"MANAGER_URL=ws://btmgr:3001", // 假设在同一个网络中
			"BOT_SELF_ID=" + strconv.FormatInt(time.Now().Unix()%1000000, 10),
		},
		Cmd: []string{"python", "onebot.py"},
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}

	// 4. 创建容器
	resp, err := m.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("创建容器失败: %v", err),
		})
		return
	}

	// 5. 启动容器
	if err := m.dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("启动容器失败: %v", err),
		})
		return
	}

	// 广播 Docker 事件
	m.broadcastDockerEvent("create", resp.ID, "running")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "机器人容器部署成功",
		"id":      resp.ID,
	})
}

// handleDockerAddWorker 添加 Worker 容器
func (m *Manager) handleDockerAddWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.dockerClient == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Docker 客户端未初始化",
		})
		return
	}

	ctx := context.Background()
	imageName := "botmatrix-system-worker"

	// 检查镜像
	_, _, err := m.dockerClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		log.Printf("[Docker] 镜像 %s 在本地未找到，正在尝试从仓库拉取...", imageName)
		reader, err := m.dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			log.Printf("[Docker] 无法拉取镜像 %s: %v", imageName, err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("镜像 %s 不存在且无法拉取。请确保镜像已正确构建或存在于仓库中。错误: %v", imageName, err),
			})
			return
		}
		defer reader.Close()
		io.Copy(io.Discard, reader)
		log.Printf("[Docker] 镜像 %s 拉取成功", imageName)
	}

	containerName := fmt.Sprintf("sysworker-%d", time.Now().Unix())

	config := &container.Config{
		Image: imageName,
		Env: []string{
			"TZ=Asia/Shanghai",
			"BOT_MANAGER_URL=ws://btmgr:3001",
		},
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}

	resp, err := m.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("创建容器失败: %v", err),
		})
		return
	}

	if err := m.dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("启动容器失败: %v", err),
		})
		return
	}

	// 广播 Docker 事件
	m.broadcastDockerEvent("create", resp.ID, "running")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Worker容器部署成功",
		"id":      resp.ID,
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

// handleAdminManageUsers 统一处理用户管理操作 (创建/删除/重置密码)
func (m *Manager) handleAdminManageUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "仅支持 POST 请求",
		})
		return
	}

	var req struct {
		Action   string `json:"action"`
		Username string `json:"username"`
		Password string `json:"password"` // 对应 create 和 reset_password
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求格式错误",
		})
		return
	}

	switch req.Action {
	case "create":
		m.processAdminCreateUser(w, req.Username, req.Password, req.IsAdmin)
	case "delete":
		m.processAdminDeleteUser(w, r, req.Username)
	case "reset_password":
		m.processAdminResetPassword(w, req.Username, req.Password)
	case "toggle_active":
		m.processAdminToggleUser(w, req.Username)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "不支持的操作: " + req.Action,
		})
	}
}

// processAdminCreateUser 内部处理创建用户
func (m *Manager) processAdminCreateUser(w http.ResponseWriter, username, password string, isAdmin bool) {
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名和密码不能为空",
		})
		return
	}

	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	if _, exists := m.users[username]; exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户已存在",
		})
		return
	}

	// 检查数据库
	var existingID int64
	err := m.db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&existingID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户已存在",
		})
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "密码哈希失败",
		})
		return
	}

	newUser := &User{
		Username:       username,
		PasswordHash:   hashedPassword,
		IsAdmin:        isAdmin,
		SessionVersion: 1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := m.saveUserToDB(newUser); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存用户失败: " + err.Error(),
		})
		return
	}

	// 重新读取 ID
	m.db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&newUser.ID)
	m.users[username] = newUser

	log.Printf("[ADMIN] Created user: %s (Admin: %v)", username, isAdmin)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户创建成功",
	})
}

// processAdminDeleteUser 内部处理删除用户
func (m *Manager) processAdminDeleteUser(w http.ResponseWriter, r *http.Request, username string) {
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

	// 禁止删除自己
	claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
	if ok && claims != nil && claims.Username == username {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "不能删除当前登录用户",
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

	if _, err := m.db.Exec("DELETE FROM users WHERE username = ?", username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	delete(m.users, username)
	log.Printf("[ADMIN] Deleted user: %s", username)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户删除成功",
	})
}

// processAdminResetPassword 内部处理重置密码
func (m *Manager) processAdminResetPassword(w http.ResponseWriter, username, newPassword string) {
	if username == "" || newPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户名和新密码不能为空",
		})
		return
	}

	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()

	user, exists := m.users[username]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "密码哈希失败",
		})
		return
	}

	user.PasswordHash = hashedPassword
	user.SessionVersion++
	user.UpdatedAt = time.Now()

	if err := m.saveUserToDB(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存失败",
		})
		return
	}

	log.Printf("[ADMIN] Reset password for user: %s", username)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "密码重置成功",
	})
}

// processAdminToggleUser 内部处理切换用户状态
func (m *Manager) processAdminToggleUser(w http.ResponseWriter, username string) {
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
			"message": "不能禁用默认管理员",
		})
		return
	}

	m.usersMutex.Lock()
	user, exists := m.users[username]
	if !exists {
		m.usersMutex.Unlock()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	newStatus := !user.Active
	user.Active = newStatus
	m.usersMutex.Unlock()

	// 更新数据库
	activeInt := 0
	if newStatus {
		activeInt = 1
	}
	if _, err := m.db.Exec("UPDATE users SET active = ? WHERE username = ?", activeInt, username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	log.Printf("[ADMIN] Toggled user status: %s (Active: %v)", username, newStatus)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户状态已更新",
		"active":  newStatus,
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

// handleProxyAvatar 代理头像请求，解决跨域问题
func (m *Manager) handleProxyAvatar(w http.ResponseWriter, r *http.Request) {
	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	avatarURL := r.URL.Query().Get("url")
	if avatarURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// 验证 URL 协议
	if !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
		http.Error(w, "Invalid URL protocol", http.StatusBadRequest)
		return
	}

	// 创建请求
	req, err := http.NewRequest("GET", avatarURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 伪造 User-Agent 以防被拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	// 一些头像服务器需要特定的 Accept 头
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	// 移除 Referer 以防被防盗链拦截
	req.Header.Del("Referer")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[PROXY] Failed to fetch avatar: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制关键响应头
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	}

	// 强制允许跨域
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("X-Proxy-By", "BotNexus")

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// handleSendAction 处理发送 API 动作的请求
func (m *Manager) handleSendAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BotID  string                 `json:"bot_id"`
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "请求格式错误",
		})
		return
	}

	// 支持批量发送功能
	if req.Action == "batch_send_msg" {
		targets, ok := req.Params["targets"].([]interface{})
		message, _ := req.Params["message"].(string)
		if !ok || message == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "failed",
				"message": "批量发送参数错误",
			})
			return
		}

		// 异步处理批量发送，防止请求阻塞
		go func() {
			log.Printf("[BatchSend] Starting batch send for %d targets", len(targets))
			success := 0
			failed := 0
			for _, t := range targets {
				target, ok := t.(map[string]interface{})
				if !ok {
					continue
				}

				targetID := toString(target["id"])
				targetBotID := toString(target["bot_id"])
				targetType := toString(target["type"])

				m.mutex.RLock()
				bot, exists := m.bots[targetBotID]
				m.mutex.RUnlock()

				if !exists {
					log.Printf("[BatchSend] Bot %s not found for target %s", targetBotID, targetID)
					failed++
					continue
				}

				action := "send_group_msg"
				params := map[string]interface{}{
					"group_id": targetID,
					"message":  message,
				}
				if targetType == "private" {
					action = "send_private_msg"
					params = map[string]interface{}{
						"user_id": targetID,
						"message": message,
					}
				} else if targetType == "guild" {
					action = "send_msg"
					params = map[string]interface{}{
						"message_type": "guild",
						"channel_id":   targetID,
						"message":      message,
					}
					// 如果有 guild_id，也带上
					if gid := toString(target["guild_id"]); gid != "" {
						params["guild_id"] = gid
					}
				}

				// 发送请求，不等待响应（或者设置短超时）
				echo := fmt.Sprintf("batch|%d|%s", time.Now().UnixNano(), action)
				msg := map[string]interface{}{
					"action": action,
					"params": params,
					"echo":   echo,
				}

				bot.Mutex.Lock()
				err := bot.Conn.WriteJSON(msg)
				bot.Mutex.Unlock()

				if err != nil {
					log.Printf("[BatchSend] Failed to send to %s: %v", targetID, err)
					failed++
				} else {
					success++
				}
				// 稍微停顿一下，防止发送过快被风控
				time.Sleep(200 * time.Millisecond)
			}
			log.Printf("[BatchSend] Completed: %d success, %d failed", success, failed)
		}()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"success": true,
			"message": fmt.Sprintf("已启动批量发送任务，共 %d 个目标", len(targets)),
		})
		return
	}

	// 如果没有指定 bot_id，尝试寻找一个可用的
	m.mutex.RLock()
	var bot *BotClient
	if req.BotID != "" {
		bot = m.bots[req.BotID]
	} else {
		// 选第一个
		for _, b := range m.bots {
			bot = b
			break
		}
	}
	m.mutex.RUnlock()

	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "未找到可用的 Bot",
		})
		return
	}

	// 构造 echo
	echo := fmt.Sprintf("web|%d|%s", time.Now().UnixNano(), req.Action)

	// 注册等待响应
	respChan := make(chan map[string]interface{}, 1)
	m.pendingMutex.Lock()
	m.pendingRequests[echo] = respChan
	m.pendingTimestamps[echo] = time.Now()
	m.pendingMutex.Unlock()

	defer func() {
		m.pendingMutex.Lock()
		delete(m.pendingRequests, echo)
		delete(m.pendingTimestamps, echo)
		m.pendingMutex.Unlock()
	}()

	// 构造消息
	msg := map[string]interface{}{
		"action": req.Action,
		"params": req.Params,
		"echo":   echo,
	}

	// 发送给 Bot
	bot.Mutex.Lock()
	err := bot.Conn.WriteJSON(msg)
	bot.Mutex.Unlock()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "发送请求到 Bot 失败: " + err.Error(),
		})
		return
	}

	// 等待响应
	select {
	case resp := <-respChan:
		json.NewEncoder(w).Encode(resp)
	case <-time.After(30 * time.Second):
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "等待 Bot 响应超时",
		})
	}
}

// handleGetChatStats 获取聊天统计信息
func (m *Manager) handleGetChatStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	// 提取群组名称
	groupNames := make(map[string]string)
	for id, g := range m.groupCache {
		if name, ok := g["group_name"].(string); ok {
			groupNames[id] = name
		}
	}

	// 提取用户名称
	userNames := make(map[string]string)
	for id, u := range m.memberCache {
		if name, ok := u["nickname"].(string); ok {
			userNames[id] = name
		}
	}
	for id, f := range m.friendCache {
		if name, ok := f["nickname"].(string); ok {
			userNames[id] = name
		}
	}

	// 转换为 map[string]int64 以便前端使用
	gs := make(map[string]int64)
	for k, v := range m.GroupStats {
		gs[k] = v
	}
	us := make(map[string]int64)
	for k, v := range m.UserStats {
		us[k] = v
	}
	gst := make(map[string]int64)
	for k, v := range m.GroupStatsToday {
		gst[k] = v
	}
	ust := make(map[string]int64)
	for k, v := range m.UserStatsToday {
		ust[k] = v
	}

	resp := map[string]interface{}{
		"group_stats":       gs,
		"user_stats":        us,
		"group_stats_today": gst,
		"user_stats_today":  ust,
		"group_names":       groupNames,
		"user_names":        userNames,
	}

	json.NewEncoder(w).Encode(resp)
}
