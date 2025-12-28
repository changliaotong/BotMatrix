package main

import (
	log "BotMatrix/common/log"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"BotMatrix/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// GetAvatarURL 根据平台、ID 和是否为群组返回头像地址
func GetAvatarURL(platform string, id string, isGroup bool, providedAvatar string) string {
	platform = strings.ToUpper(platform)

	// 如果 ID 是纯数字且在特定范围内，处理 QQ 逻辑
	isQQ := platform == "QQ"
	if !isQQ && id != "" {
		isNumeric := true
		for _, c := range id {
			if c < '0' || c > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric && len(id) >= 5 && len(id) <= 11 {
			isQQ = true
		}
	}

	if isQQ {
		if isGroup {
			return "https://p.qlogo.cn/gh/" + id + "/" + id + "/640"
		}
		return "https://q.qlogo.cn/headimg_dl?dst_uin=" + id + "&spec=640"
	}

	// 非 QQ 协议，检查 ID 是否大于 980000000000
	if id != "" {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err == nil && idInt > 980000000000 {
			// 返回特殊前缀，由前端识别显示平台 logo
			return "platform://" + strings.ToLower(platform)
		}
	}

	// 否则使用提供的头像
	if providedAvatar != "" {
		return providedAvatar
	}

	return ""
}

// HandleLogin 处理登录请求
func HandleLogin(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			log.Printf(common.T("", "login_request_failed"), err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		log.Printf(common.T("", "login_attempt"), loginData.Username, r.RemoteAddr)

		m.UsersMutex.RLock()
		user, exists := m.Users[loginData.Username]
		m.UsersMutex.RUnlock()

		if !exists {
			row := m.DB.QueryRow(m.PrepareQuery("SELECT id, username, password_hash, is_admin, active, session_version, created_at, updated_at FROM users WHERE username = ?"), loginData.Username)
			var u common.User
			var createdAt, updatedAt interface{}
			err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.Active, &u.SessionVersion, &createdAt, &updatedAt)
			if err == nil {
				if createdAt != nil {
					switch v := createdAt.(type) {
					case time.Time:
						u.CreatedAt = v
					case string:
						u.CreatedAt, _ = time.Parse(time.RFC3339, v)
					}
				}
				if updatedAt != nil {
					switch v := updatedAt.(type) {
					case time.Time:
						u.UpdatedAt = v
					case string:
						u.UpdatedAt, _ = time.Parse(time.RFC3339, v)
					}
				}
				user = &u
				exists = true

				m.UsersMutex.Lock()
				m.Users[u.Username] = &u
				m.UsersMutex.Unlock()
			}
		}

		if !exists || !common.CheckPassword(loginData.Password, user.PasswordHash) {
			log.Printf(common.T("", "invalid_username_password") + ": " + loginData.Username)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_username_password"),
			})
			return
		}

		if !user.Active {
			log.Printf("用户未激活: %s", loginData.Username)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "user_not_active"),
			})
			return
		}

		token, err := m.GenerateToken(user)
		if err != nil {
			log.Printf(common.T("", "token_generation_failed")+": %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "token_generation_failed"),
			})
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		log.Printf(common.T("", "login_success"), user.Username, role)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"token":   token,
			"role":    role,
			"user": map[string]interface{}{
				"id":         user.ID,
				"username":   user.Username,
				"is_admin":   user.IsAdmin,
				"role":       role,
				"created_at": user.CreatedAt.Format(time.RFC3339),
			},
		})
	}
}

// HandleGetUserInfo 获取当前登录用户信息
func HandleGetUserInfo(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		claims, ok := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "not_logged_in"),
			})
			return
		}

		m.UsersMutex.RLock()
		user, exists := m.Users[claims.Username]
		m.UsersMutex.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "user_not_found"),
			})
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"user": map[string]interface{}{
				"id":         user.ID,
				"username":   user.Username,
				"is_admin":   user.IsAdmin,
				"role":       role,
				"created_at": user.CreatedAt.Format(time.RFC3339),
				"updated_at": user.UpdatedAt.Format(time.RFC3339),
			},
		})
	}
}

// HandleGetNexusStatus 获取 Nexus 运行状态
func HandleGetNexusStatus(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		// 获取连接统计
		m.ConnectionStats.Mutex.Lock()
		botCount := m.ConnectionStats.TotalBotConnections
		workerCount := m.ConnectionStats.TotalWorkerConnections
		m.ConnectionStats.Mutex.Unlock()

		// 获取当前在线数
		onlineBots := len(m.Bots)
		onlineWorkers := len(m.Workers)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"running":   true,
			"connected": onlineBots > 0 || onlineWorkers > 0,
			"stats": map[string]interface{}{
				"online_bots":    onlineBots,
				"online_workers": onlineWorkers,
				"total_bots":     botCount,
				"total_workers":  workerCount,
			},
			"version": VERSION,
		})
	}
}

// HandleGetStats 获取统计信息的请求
func HandleGetStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		m.StatsMutex.RLock()
		defer m.StatsMutex.RUnlock()

		onlineBots := len(m.Bots)
		onlineWorkers := len(m.Workers)
		totalBots := len(m.BotStats)

		if onlineBots > totalBots {
			totalBots = onlineBots
		}

		offlineBots := totalBots - onlineBots
		if offlineBots < 0 {
			offlineBots = 0
		}

		var mStats runtime.MemStats
		runtime.ReadMemStats(&mStats)

		cpuInfos, _ := cpu.Info()
		cpuModel := "Unknown"
		cpuCoresPhysical := 0
		cpuCoresLogical := 0
		cpuFreq := 0.0
		if len(cpuInfos) > 0 {
			cpuModel = cpuInfos[0].ModelName
			cpuCoresPhysical = int(cpuInfos[0].Cores)
			logical, _ := cpu.Counts(true)
			cpuCoresLogical = logical
			cpuFreq = cpuInfos[0].Mhz
		}

		cpuCount, _ := cpu.Counts(true)
		if cpuCount <= 0 {
			cpuCount = 1
		}

		cpuPercent, _ := cpu.Percent(0, false)
		var cpuUsage float64
		if len(cpuPercent) > 0 {
			cpuUsage = cpuPercent[0]
			if cpuUsage > 100 && cpuCount > 1 {
				cpuUsage = cpuUsage / float64(cpuCount)
			}
		}

		vm, _ := mem.VirtualMemory()
		hi, _ := host.Info()

		// 获取磁盘使用率 (主分区)
		var diskUsageStr string = "0%"
		usage, err := disk.Usage("/")
		if err != nil {
			// Windows 下尝试使用 C:
			usage, err = disk.Usage("C:")
		}
		if err == nil {
			diskUsageStr = fmt.Sprintf("%.1f%%", usage.UsedPercent)
		}

		m.HistoryMutex.RLock()
		cpuTrend := append([]float64{}, m.CPUTrend...)
		memTrend := append([]uint64{}, m.MemTrend...)
		msgTrend := append([]int64{}, m.MsgTrend...)
		sentTrend := append([]int64{}, m.SentTrend...)
		recvTrend := append([]int64{}, m.RecvTrend...)
		netSentTrend := append([]uint64{}, m.NetSentTrend...)
		netRecvTrend := append([]uint64{}, m.NetRecvTrend...)
		m.HistoryMutex.RUnlock()

		// 如果当前 CPU 使用率为 0，尝试从趋势中获取最新值
		if cpuUsage <= 0 && len(cpuTrend) > 0 {
			cpuUsage = cpuTrend[len(cpuTrend)-1]
		}

		uptimeSec := int64(time.Since(m.StartTime).Seconds())
		uptimeStr := fmt.Sprintf("%ds", uptimeSec)
		if uptimeSec >= 60 {
			uptimeStr = fmt.Sprintf("%dm %ds", uptimeSec/60, uptimeSec%60)
		}
		if uptimeSec >= 3600 {
			uptimeStr = fmt.Sprintf("%dh %dm", uptimeSec/3600, (uptimeSec%3600)/60)
		}
		if uptimeSec >= 86400 {
			uptimeStr = fmt.Sprintf("%dd %dh", uptimeSec/86400, (uptimeSec%86400)/3600)
		}

		stats := map[string]interface{}{
			"goroutines":          runtime.NumGoroutine(),
			"uptime":              uptimeStr,
			"memory_alloc":        mStats.Alloc,
			"memory_total":        vm.Total,
			"memory_used":         vm.Used,
			"memory_free":         vm.Free,
			"memory_used_percent": vm.UsedPercent,
			"disk_usage":          diskUsageStr,
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
			"bots_detail":         m.BotDetailedStats,
			"cpu_trend":           cpuTrend,
			"mem_trend":           memTrend,
			"msg_trend":           msgTrend,
			"sent_trend":          sentTrend,
			"recv_trend":          recvTrend,
			"net_sent_trend":      netSentTrend,
			"net_recv_trend":      netRecvTrend,
			"top_processes":       m.TopProcesses,
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"stats":   stats,
		})
	}
}

// HandleGetSystemStats 获取详细的系统运行统计
func HandleGetSystemStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cpuCount, _ := cpu.Counts(true)
		if cpuCount <= 0 {
			cpuCount = 1
		}

		cpuPercent, _ := cpu.Percent(time.Second, false)
		var cpuUsage float64
		if len(cpuPercent) > 0 {
			cpuUsage = cpuPercent[0]
			if cpuUsage > 100 && cpuCount > 1 {
				cpuUsage = cpuUsage / float64(cpuCount)
			}
		}

		vm, _ := mem.VirtualMemory()
		hi, _ := host.Info()

		partitions, _ := disk.Partitions(true)
		var diskUsage []map[string]interface{}
		seenMounts := make(map[string]bool)
		for _, p := range partitions {
			if seenMounts[p.Mountpoint] {
				continue
			}
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

		netIO, _ := net.IOCounters(false)
		var netUsage []map[string]interface{}
		for _, ioCounter := range netIO {
			netUsage = append(netUsage, map[string]interface{}{
				"name":      ioCounter.Name,
				"bytesSent": ioCounter.BytesSent,
				"bytesRecv": ioCounter.BytesRecv,
			})
		}

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

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"stats":   stats,
		})
	}
}

// HandleGetLogs 处理获取日志的请求
func HandleGetLogs(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 获取查询参数
		query := r.URL.Query()
		page, _ := strconv.Atoi(query.Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(query.Get("pageSize"))
		if pageSize < 1 {
			pageSize = 20
		}
		level := query.Get("level")
		botId := query.Get("botId")
		search := query.Get("search")
		sortBy := query.Get("sortBy")
		sortOrder := query.Get("sortOrder")

		// 获取所有日志
		allLogs := m.GetLogs(0) // 0 表示获取全部缓冲区日志

		// 1. 过滤
		var filteredLogs []common.LogEntry
		for _, entry := range allLogs {
			// 级别过滤
			if level != "" && level != "all" && !strings.EqualFold(entry.Level, level) {
				continue
			}
			// BotId/Source 过滤
			if botId != "" && botId != "all" && !strings.EqualFold(entry.Source, botId) {
				continue
			}

			// 搜索过滤
			if search != "" {
				searchLower := strings.ToLower(search)
				if !strings.Contains(strings.ToLower(entry.Message), searchLower) &&
					!strings.Contains(strings.ToLower(entry.Level), searchLower) &&
					!strings.Contains(strings.ToLower(entry.Source), searchLower) {
					continue
				}
			}
			filteredLogs = append(filteredLogs, entry)
		}

		// 2. 排序
		if sortBy != "" {
			sort.Slice(filteredLogs, func(i, j int) bool {
				var less bool
				switch sortBy {
				case "time", "timestamp":
					less = filteredLogs[i].Timestamp.Before(filteredLogs[j].Timestamp)
				case "level":
					less = filteredLogs[i].Level < filteredLogs[j].Level
				case "message":
					less = filteredLogs[i].Message < filteredLogs[j].Message
				default:
					less = filteredLogs[i].Timestamp.Before(filteredLogs[j].Timestamp)
				}
				if sortOrder == "desc" {
					return !less
				}
				return less
			})
		} else {
			// 默认按时间倒序
			sort.Slice(filteredLogs, func(i, j int) bool {
				return filteredLogs[i].Timestamp.After(filteredLogs[j].Timestamp)
			})
		}

		// 3. 分页
		total := len(filteredLogs)
		start := (page - 1) * pageSize
		end := start + pageSize

		var pagedLogs []common.LogEntry
		if start < total {
			if end > total {
				end = total
			}
			pagedLogs = filteredLogs[start:end]
		} else {
			pagedLogs = []common.LogEntry{}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"logs":    pagedLogs,
				"total":   total,
				"hasMore": end < total,
			},
		})
	}
}

// HandleClearLogs 处理清空日志的请求
func HandleClearLogs(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)
		m.ClearLogs()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "logs_cleared"),
		})
	}
}

// HandleGetBots 处理获取机器人列表的请求
func HandleGetBots(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		m.StatsMutex.RLock()
		defer m.StatsMutex.RUnlock()

		bots := make([]map[string]interface{}, 0, len(m.Bots))
		for id, bot := range m.Bots {
			remoteAddr := ""
			if bot.Conn != nil {
				remoteAddr = bot.Conn.RemoteAddr().String()
			}

			totalMsg := m.BotStats[id]
			todayMsg := m.BotStatsToday[id]

			bots = append(bots, map[string]interface{}{
				"id":              id,
				"self_id":         bot.SelfID,
				"nickname":        bot.Nickname,
				"avatar":          GetAvatarURL(bot.Platform, bot.SelfID, false, ""),
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
}

// HandleGetWorkers 处理获取Worker列表的请求
func HandleGetWorkers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		workers := make([]map[string]interface{}, 0, len(m.Workers))
		for _, worker := range m.Workers {
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
}

// HandleDockerList 获取 Docker 容器列表
func HandleDockerList(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			// 尝试延迟初始化
			if err := m.InitDockerClient(); err != nil {
				log.Printf("Docker client initialization failed: %v", err)
				// 不要返回 500，而是返回空列表，让前端知道 Docker 未就绪
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":     "warning",
					"message":    common.T(lang, "docker_not_init"),
					"containers": []interface{}{},
				})
				return
			}
		}

		containers, err := m.DockerClient.ContainerList(r.Context(), types.ContainerListOptions{All: true})
		if err != nil {
			log.Printf(common.T("", "docker_list_failed"), err)
			// 同理，返回空列表而不是 500
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "error",
				"message":    err.Error(),
				"containers": []interface{}{},
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "ok",
			"containers": containers,
		})
	}
}

// HandleDockerAction 处理 Docker 容器操作 (start/stop/restart)
func HandleDockerAction(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var req struct {
			ContainerID string `json:"container_id"`
			Action      string `json:"action"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		if m.DockerClient == nil {
			// 尝试延迟初始化
			if err := m.InitDockerClient(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": common.T(lang, "docker_not_init") + ": " + err.Error(),
				})
				return
			}
		}

		var err error
		switch req.Action {
		case "start":
			err = m.DockerClient.ContainerStart(r.Context(), req.ContainerID, types.ContainerStartOptions{})
		case "stop":
			timeout := 10
			err = m.DockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
		case "restart":
			timeout := 10
			err = m.DockerClient.ContainerRestart(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
		case "delete":
			timeout := 5
			m.DockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
			err = m.DockerClient.ContainerRemove(r.Context(), req.ContainerID, types.ContainerRemoveOptions{Force: true})
		default:
			err = fmt.Errorf(common.T(lang, "unsupported_action"), req.Action)
		}

		if err != nil {
			log.Printf(common.T("", "docker_action_failed"), req.Action, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		status := "running"
		if req.Action == "stop" {
			status = "exited"
		} else if req.Action == "delete" {
			status = "deleted"
		}
		m.BroadcastDockerEvent(req.Action, req.ContainerID, status)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"id":     req.ContainerID,
		})
	}
}

// HandleDockerAddBot 添加机器人容器
func HandleDockerAddBot(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "docker_not_init"),
			})
			return
		}

		var req struct {
			Platform string            `json:"platform"`
			Image    string            `json:"image"`
			Env      map[string]string `json:"env"`
			Cmd      []string          `json:"cmd"`
		}

		// 尝试解析请求体，如果解析失败则使用默认值 (向后兼容)
		imageName := "botmatrix-wxbot"
		platform := "WeChat"
		envVars := make(map[string]string)
		cmd := []string{"python", "onebot.py"}

		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				if req.Image != "" {
					imageName = req.Image
				}
				if req.Platform != "" {
					platform = req.Platform
				}
				if req.Env != nil {
					envVars = req.Env
				}
				if req.Cmd != nil {
					cmd = req.Cmd
				}
			}
		}

		ctx := context.Background()

		_, _, err := m.DockerClient.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			log.Printf(common.T("", "docker_pulling_image"), imageName)
			reader, err := m.DockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
			if err != nil {
				log.Printf(common.T("", "docker_pull_failed"), imageName, err)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": fmt.Sprintf(common.T(lang, "docker_image_not_exists"), imageName, err),
				})
				return
			}
			defer reader.Close()
			io.Copy(io.Discard, reader)
		}

		containerName := fmt.Sprintf("%s-%d", strings.ToLower(platform), time.Now().Unix())

		// 构建环境变量列表
		finalEnv := []string{
			"TZ=Asia/Shanghai",
		}

		// 如果没有指定 MANAGER_URL，使用默认值
		if _, ok := envVars["MANAGER_URL"]; !ok {
			finalEnv = append(finalEnv, "MANAGER_URL=ws://btmgr:3001")
		}
		if _, ok := envVars["NEXUS_ADDR"]; !ok {
			finalEnv = append(finalEnv, "NEXUS_ADDR=ws://btmgr:3001/ws/bots")
		}

		for k, v := range envVars {
			finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
		}

		// 如果没有指定 BOT_SELF_ID 且是微信平台，添加一个随机 ID (兼容旧版)
		if platform == "WeChat" {
			foundSelfID := false
			for k := range envVars {
				if k == "BOT_SELF_ID" {
					foundSelfID = true
					break
				}
			}
			if !foundSelfID {
				finalEnv = append(finalEnv, "BOT_SELF_ID="+fmt.Sprintf("%d", time.Now().Unix()%1000000))
			}
		}

		config := &container.Config{
			Image: imageName,
			Env:   finalEnv,
			Cmd:   cmd,
		}

		hostConfig := &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: "always"},
		}

		resp, err := m.DockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "docker_create_container_failed"), err),
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "docker_start_container_failed"), err),
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": common.T(lang, "bot_deploy_success"),
			"id":      resp.ID,
		})
	}
}

// HandleDockerAddWorker 添加 Worker 容器
func HandleDockerAddWorker(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "docker_not_init"),
			})
			return
		}

		var req struct {
			Image string            `json:"image"`
			Env   map[string]string `json:"env"`
			Cmd   []string          `json:"cmd"`
			Name  string            `json:"name"`
		}

		// 默认值
		imageName := "botmatrix-system-worker"
		workerName := fmt.Sprintf("sysworker-%d", time.Now().Unix())
		envVars := make(map[string]string)
		var cmd []string

		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				if req.Image != "" {
					imageName = req.Image
				}
				if req.Name != "" {
					workerName = req.Name
				}
				if req.Env != nil {
					envVars = req.Env
				}
				if req.Cmd != nil {
					cmd = req.Cmd
				}
			}
		}

		ctx := context.Background()

		_, _, err := m.DockerClient.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			log.Printf(common.T("", "docker_pulling_image"), imageName)
			reader, err := m.DockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
			if err != nil {
				log.Printf(common.T("", "docker_pull_failed"), imageName, err)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": fmt.Sprintf(common.T(lang, "docker_image_not_exists"), imageName, err),
				})
				return
			}
			defer reader.Close()
			io.Copy(io.Discard, reader)
		}

		// 构建环境变量列表
		finalEnv := []string{
			"TZ=Asia/Shanghai",
		}

		// 如果没有指定 BOT_MANAGER_URL，使用默认值
		if _, ok := envVars["BOT_MANAGER_URL"]; !ok {
			finalEnv = append(finalEnv, "BOT_MANAGER_URL=ws://btmgr:3001")
		}

		for k, v := range envVars {
			finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
		}

		config := &container.Config{
			Image: imageName,
			Env:   finalEnv,
			Cmd:   cmd,
		}

		hostConfig := &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: "always"},
		}

		resp, err := m.DockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, workerName)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "docker_create_container_failed"), err),
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "docker_start_container_failed"), err),
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": common.T(lang, "worker_deploy_success"),
			"id":      resp.ID,
		})
	}
}

// HandleChangePassword 修改用户密码
func HandleChangePassword(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		claims, ok := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "not_logged_in"),
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
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		m.UsersMutex.Lock()
		defer m.UsersMutex.Unlock()

		user, exists := m.Users[claims.Username]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "user_not_found"),
			})
			return
		}

		if !common.CheckPassword(data.OldPassword, user.PasswordHash) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "old_password_error"),
			})
			return
		}

		newHash, err := common.HashPassword(data.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "password_encrypt_failed"),
			})
			return
		}

		user.PasswordHash = newHash
		user.UpdatedAt = time.Now()

		if err := m.SaveUserToDB(user); err != nil {
			log.Printf(common.T("", "password_update_db_failed"), err)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "password_change_success"),
		})
	}
}

// HandleGetMessages 获取最新消息列表
func HandleGetMessages(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		if m.DB == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  false,
				"messages": []interface{}{},
				"error":    "Database not initialized",
			})
			return
		}

		query := `
		SELECT id, message_id, bot_id, user_id, group_id, type, content, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT ?
		`
		rows, err := m.DB.Query(m.PrepareQuery(query), limit)
		if err != nil {
			log.Printf("获取最新消息失败: %v", err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  false,
				"messages": []interface{}{},
				"error":    err.Error(),
			})
			return
		}
		defer rows.Close()

		messages := []map[string]interface{}{}
		for rows.Next() {
			var id int
			var messageID, botID, userID, groupID, msgType, content string
			var createdAt time.Time
			err := rows.Scan(&id, &messageID, &botID, &userID, &groupID, &msgType, &content, &createdAt)
			if err != nil {
				log.Printf("扫描消息失败: %v", err)
				continue
			}

			// 获取昵称和群名
			userName := userID
			groupName := groupID
			userAvatar := ""
			groupAvatar := ""

			m.Mutex.RLock()
			bot, hasBot := m.Bots[botID]
			platform := ""
			if hasBot {
				platform = bot.Platform
			}
			m.Mutex.RUnlock()

			m.CacheMutex.RLock()
			if friend, ok := m.FriendCache[userID]; ok {
				if n, ok := friend["nickname"].(string); ok {
					userName = n
				}
				if a, ok := friend["avatar"].(string); ok {
					userAvatar = a
				}
			}
			// 如果在好友缓存没找到头像，尝试在成员缓存找
			if userAvatar == "" && groupID != "" && groupID != "0" {
				memberKey := groupID + "_" + userID
				if member, ok := m.MemberCache[memberKey]; ok {
					if a, ok := member["avatar"].(string); ok {
						userAvatar = a
					}
					// 也可以尝试更新下昵称，如果之前没找到的话
					if userName == userID {
						if n, ok := member["nickname"].(string); ok {
							userName = n
						}
					}
				}
			}

			if group, ok := m.GroupCache[groupID]; ok {
				if n, ok := group["group_name"].(string); ok {
					groupName = n
				}
				if a, ok := group["avatar"].(string); ok {
					groupAvatar = a
				}
			}
			m.CacheMutex.RUnlock()

			messages = append(messages, map[string]interface{}{
				"id":           id,
				"message_id":   messageID,
				"bot_id":       botID,
				"user_id":      userID,
				"user_name":    userName,
				"user_avatar":  GetAvatarURL(platform, userID, false, userAvatar),
				"group_id":     groupID,
				"group_name":   groupName,
				"group_avatar": GetAvatarURL(platform, groupID, true, groupAvatar),
				"type":         msgType,
				"content":      content,
				"created_at":   createdAt.Format("2006-01-02 15:04:05"),
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"messages": messages,
		})
	}
}

// HandleGetContacts 获取联系人列表 (群组和好友)
func HandleGetContacts(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var botID string
		refresh := false

		if r.Method == http.MethodPost {
			var body struct {
				BotID   string `json:"bot_id"`
				Refresh bool   `json:"refresh"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				botID = body.BotID
				refresh = true // Sync endpoint defaults to refresh
			}
		} else {
			botID = r.URL.Query().Get("bot_id")
			refresh = r.URL.Query().Get("refresh") == "true"
		}

		lang := r.Header.Get("Accept-Language")
		if lang == "" {
			lang = "zh-CN"
		}

		// 如果不强制刷新，检查缓存是否为空
		if !refresh && botID != "" {
			m.CacheMutex.RLock()
			hasCache := false
			// 检查群组缓存
			for _, g := range m.GroupCache {
				if common.ToString(g["bot_id"]) == botID {
					hasCache = true
					break
				}
			}
			// 如果群组没有，检查好友缓存
			if !hasCache {
				for _, f := range m.FriendCache {
					if common.ToString(f["bot_id"]) == botID {
						hasCache = true
						break
					}
				}
			}
			m.CacheMutex.RUnlock()

			// 如果完全没有缓存，则强制触发一次刷新
			if !hasCache {
				refresh = true
			}
		}

		if refresh && botID != "" {
			m.Mutex.RLock()
			bot, ok := m.Bots[botID]
			m.Mutex.RUnlock()

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": common.T(lang, "bot_not_found|未找到机器人实例"),
				})
				return
			}

			// 检查机器人连接状态和心跳
			if bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": common.T(lang, "bot_disconnected|机器人已断开连接或心跳超时"),
				})
				return
			}

			echoGroups := "refresh_groups_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.PendingMutex.Lock()
			respChanGroups := make(chan map[string]interface{}, 1)
			m.PendingRequests[echoGroups] = respChanGroups
			m.PendingMutex.Unlock()

			bot.Mutex.Lock()
			err := bot.Conn.WriteJSON(map[string]interface{}{
				"action": "get_group_list",
				"params": map[string]interface{}{},
				"echo":   echoGroups,
			})
			bot.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send get_group_list to bot %s: %v", botID, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": common.T(lang, "bot_communication_error|与机器人通信失败"),
				})
				return
			}

			// Group fetch with its own timeout
			select {
			case resp := <-respChanGroups:
				if data, ok := resp["data"].([]interface{}); ok {
					m.CacheMutex.Lock()
					for _, g := range data {
						if group, ok := g.(map[string]interface{}); ok {
							gID := common.ToString(group["group_id"])
							gName := common.ToString(group["group_name"])

							m.GroupCache[gID] = map[string]interface{}{
								"group_id":   gID,
								"group_name": gName,
								"bot_id":     botID,
							}
							go m.SaveGroupToDB(gID, gName, botID)
						}
					}
					m.CacheMutex.Unlock()
				}
			case <-time.After(8 * time.Second):
				log.Printf(common.T("", "contacts_timeout_groups"), botID)
			}

			m.PendingMutex.Lock()
			delete(m.PendingRequests, echoGroups)
			m.PendingMutex.Unlock()

			echoFriends := "refresh_friends_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.PendingMutex.Lock()
			respChanFriends := make(chan map[string]interface{}, 1)
			m.PendingRequests[echoFriends] = respChanFriends
			m.PendingMutex.Unlock()

			bot.Mutex.Lock()
			err = bot.Conn.WriteJSON(map[string]interface{}{
				"action": "get_friend_list",
				"params": map[string]interface{}{},
				"echo":   echoFriends,
			})
			bot.Mutex.Unlock()

			if err == nil {
				// Friend fetch with its own timeout
				select {
				case resp := <-respChanFriends:
					if data, ok := resp["data"].([]interface{}); ok {
						m.CacheMutex.Lock()
						for _, f := range data {
							if friend, ok := f.(map[string]interface{}); ok {
								uID := common.ToString(friend["user_id"])
								nickname := common.ToString(friend["nickname"])

								m.FriendCache[uID] = map[string]interface{}{
									"user_id":  uID,
									"nickname": nickname,
									"bot_id":   botID,
								}
								go m.SaveFriendToDB(uID, nickname, botID)
							}
						}
						m.CacheMutex.Unlock()
					}
				case <-time.After(8 * time.Second):
					log.Printf(common.T("", "contacts_timeout_friends"), botID)
				}
			}

			m.PendingMutex.Lock()
			delete(m.PendingRequests, echoFriends)
			m.PendingMutex.Unlock()

			if bot.Platform == "qq_guild" || bot.Platform == "guild" {
				echoGuilds := "refresh_guilds_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
				m.PendingMutex.Lock()
				respChanGuilds := make(chan map[string]interface{}, 1)
				m.PendingRequests[echoGuilds] = respChanGuilds
				m.PendingMutex.Unlock()

				bot.Mutex.Lock()
				bot.Conn.WriteJSON(map[string]interface{}{
					"action": "get_guild_list",
					"params": map[string]interface{}{},
					"echo":   echoGuilds,
				})
				bot.Mutex.Unlock()

				// Guild fetch with its own timeout
				select {
				case resp := <-respChanGuilds:
					if data, ok := resp["data"].([]interface{}); ok {
						for _, g := range data {
							if guild, ok := g.(map[string]interface{}); ok {
								gID := common.ToString(guild["guild_id"])
								gName := common.ToString(guild["guild_name"])

								m.CacheMutex.Lock()
								m.GroupCache[gID] = map[string]interface{}{
									"group_id":   gID,
									"group_name": gName,
									"bot_id":     botID,
									"is_guild":   true,
								}
								m.CacheMutex.Unlock()
								go m.SaveGroupToDB(gID, gName, botID)
							}
						}
					}
				case <-time.After(8 * time.Second):
					log.Printf("[Contacts] 获取频道列表超时: %s", botID)
				}

				m.PendingMutex.Lock()
				delete(m.PendingRequests, echoGuilds)
				m.PendingMutex.Unlock()
			}
		}

		m.CacheMutex.RLock()
		defer m.CacheMutex.RUnlock()

		contacts := make([]map[string]interface{}, 0)
		for _, g := range m.GroupCache {
			gBotID, _ := g["bot_id"].(string)
			if botID == "" || gBotID == botID {
				platform := ""
				m.Mutex.RLock()
				if b, ok := m.Bots[gBotID]; ok {
					platform = b.Platform
				}
				m.Mutex.RUnlock()

				gID := common.ToString(g["group_id"])
				providedAvatar, _ := g["avatar"].(string)

				contacts = append(contacts, map[string]interface{}{
					"id":       gID,
					"name":     g["group_name"],
					"nickname": g["group_name"],
					"avatar":   GetAvatarURL(platform, gID, true, providedAvatar),
					"bot_id":   gBotID,
					"type":     "group",
				})
			}
		}

		for _, f := range m.FriendCache {
			fBotID, _ := f["bot_id"].(string)
			if botID == "" || fBotID == botID {
				platform := ""
				m.Mutex.RLock()
				if b, ok := m.Bots[fBotID]; ok {
					platform = b.Platform
				}
				m.Mutex.RUnlock()

				uID := common.ToString(f["user_id"])
				providedAvatar, _ := f["avatar"].(string)

				contacts = append(contacts, map[string]interface{}{
					"id":       uID,
					"name":     f["nickname"],
					"nickname": f["nickname"],
					"avatar":   GetAvatarURL(platform, uID, false, providedAvatar),
					"bot_id":   fBotID,
					"type":     "private",
				})
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"contacts": contacts,
		})
	}
}

// HandleGetGroupMembers 获取群成员列表
func HandleGetGroupMembers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		botID := r.URL.Query().Get("bot_id")
		groupID := r.URL.Query().Get("group_id")
		refresh := r.URL.Query().Get("refresh") == "true"

		if botID == "" || groupID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "missing_parameters"),
			})
			return
		}

		m.Mutex.RLock()
		bot, ok := m.Bots[botID]
		m.Mutex.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "bot_not_found"),
			})
			return
		}

		// 检查机器人连接状态和心跳
		if bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "bot_disconnected|机器人已断开连接或心跳超时"),
			})
			return
		}

		// 1. 优先尝试从缓存读取 (如果不需要强制刷新)
		if !refresh {
			m.CacheMutex.RLock()
			cachedMembers := make([]map[string]interface{}, 0)
			for _, member := range m.MemberCache {
				if common.ToString(member["group_id"]) == groupID {
					// 注入头像
					m.Mutex.RLock()
					platform := ""
					if botID, ok := member["bot_id"].(string); ok {
						if b, ok := m.Bots[botID]; ok {
							platform = b.Platform
						}
					}
					m.Mutex.RUnlock()

					uID := common.ToString(member["user_id"])
					providedAvatar, _ := member["avatar"].(string)
					member["avatar"] = GetAvatarURL(platform, uID, false, providedAvatar)

					cachedMembers = append(cachedMembers, member)
				}
			}
			m.CacheMutex.RUnlock()

			if len(cachedMembers) > 0 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    cachedMembers,
					"cached":  true,
				})
				return
			}
		}

		// 2. 如果需要刷新或缓存中没有，则从机器人获取
		echo := "get_members_" + groupID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
		m.PendingMutex.Lock()
		respChan := make(chan map[string]interface{}, 1)
		m.PendingRequests[echo] = respChan
		m.PendingMutex.Unlock()

		defer func() {
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echo)
			m.PendingMutex.Unlock()
		}()

		bot.Mutex.Lock()
		bot.Conn.WriteJSON(map[string]interface{}{
			"action": "get_group_member_list",
			"params": map[string]interface{}{
				"group_id": groupID,
			},
			"echo": echo,
		})
		bot.Mutex.Unlock()

		select {
		case resp := <-respChan:
			if data, ok := resp["data"].([]interface{}); ok {
				// 更新缓存
				m.CacheMutex.Lock()
				for _, it := range data {
					if member, ok := it.(map[string]interface{}); ok {
						uID := common.ToString(member["user_id"])
						nickname := common.ToString(member["nickname"])
						card := common.ToString(member["card"])
						key := fmt.Sprintf("%s:%s", groupID, uID)

						m.MemberCache[key] = map[string]interface{}{
							"group_id":  groupID,
							"user_id":   uID,
							"nickname":  nickname,
							"card":      card,
							"bot_id":    botID, // 确保缓存中有 bot_id
							"avatar":    GetAvatarURL(bot.Platform, uID, false, common.ToString(member["avatar"])),
							"is_cached": true,
						}
						// 同时更新返回给前端的数据
						member["avatar"] = GetAvatarURL(bot.Platform, uID, false, common.ToString(member["avatar"]))
						// 持久化到数据库
						go m.SaveMemberToDB(groupID, uID, nickname, card)
					}
				}
				m.CacheMutex.Unlock()

				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    data,
				})
				return
			}
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "bot_timeout"),
			})
			return
		}
	}
}

// HandleProxyAvatar 代理头像请求
func HandleProxyAvatar(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
			http.Error(w, "Invalid URL protocol", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest("GET", avatarURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
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

		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("X-Proxy-By", "BotAdmin")

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// HandleBatchSend 处理批量发送消息
func HandleBatchSend(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 拦截并设置 action 为 batch_send_msg
		// 这样可以复用 HandleSendAction 的逻辑
		r.Header.Set("X-Internal-Action", "batch_send_msg")
		HandleSendAction(m)(w, r)
	}
}

// HandleSendAction 处理发送 API 动作
func HandleSendAction(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

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
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		// 如果是内部路由调用，强制设置 action
		if req.Action == "" && r.Header.Get("X-Internal-Action") != "" {
			req.Action = r.Header.Get("X-Internal-Action")
		}

		if req.Action == "batch_send_msg" {
			targets, ok := req.Params["targets"].([]interface{})
			message, _ := req.Params["message"].(string)
			if !ok || message == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "failed",
					"message": common.T(lang, "batch_send_params_error"),
				})
				return
			}

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

					m.Mutex.RLock()
					bot, exists := m.Bots[targetBotID]
					m.Mutex.RUnlock()

					if !exists {
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
						if gid := toString(target["guild_id"]); gid != "" {
							params["guild_id"] = gid
						}
					}

					echo := fmt.Sprintf("batch|%d|%s", time.Now().UnixNano(), action)
					log.Printf("[API] [Batch] Sending action to bot %s: %s, params: %+v", bot.SelfID, action, params)
					msg := map[string]interface{}{
						"action": action,
						"params": params,
						"echo":   echo,
					}

					bot.Mutex.Lock()
					err := bot.Conn.WriteJSON(msg)
					bot.Mutex.Unlock()

					if err != nil {
						failed++
					} else {
						success++
					}
					time.Sleep(200 * time.Millisecond)
				}
				log.Printf("[BatchSend] Completed: %d success, %d failed", success, failed)
			}()

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"success": true,
				"message": common.T(lang, "batch_send_start", len(targets)),
			})
			return
		}

		m.Mutex.RLock()
		var bot *common.BotClient
		if req.BotID != "" {
			bot = m.Bots[req.BotID]
		} else {
			for _, b := range m.Bots {
				bot = b
				break
			}
		}
		m.Mutex.RUnlock()

		if bot == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "failed",
				"message": common.T(lang, "no_available_bot"),
			})
			return
		}

		echo := fmt.Sprintf("web|%d|%s", time.Now().UnixNano(), req.Action)

		respChan := make(chan map[string]interface{}, 1)
		m.PendingMutex.Lock()
		m.PendingRequests[echo] = respChan
		m.PendingTimestamps[echo] = time.Now()
		m.PendingMutex.Unlock()

		defer func() {
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echo)
			delete(m.PendingTimestamps, echo)
			m.PendingMutex.Unlock()
		}()

		msg := map[string]interface{}{
			"action": req.Action,
			"params": req.Params,
			"echo":   echo,
		}

		bot.Mutex.Lock()
		err := bot.Conn.WriteJSON(msg)
		bot.Mutex.Unlock()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "failed",
				"message": fmt.Sprintf(common.T(lang, "send_to_bot_failed"), err),
			})
			return
		}

		select {
		case resp := <-respChan:
			json.NewEncoder(w).Encode(resp)
		case <-time.After(30 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "failed",
				"message": common.T(lang, "bot_timeout"),
			})
		}
	}
}

// HandleGetChatStats 获取聊天统计信息
func HandleGetChatStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()
		m.CacheMutex.RLock()
		defer m.CacheMutex.RUnlock()

		groupNames := make(map[string]string)
		groupAvatars := make(map[string]string)
		for id, g := range m.GroupCache {
			if name, ok := g["group_name"].(string); ok {
				groupNames[id] = name
			}
			platform := ""
			if botID, ok := g["bot_id"].(string); ok {
				if b, ok := m.Bots[botID]; ok {
					platform = b.Platform
				}
			}
			avatar, _ := g["avatar"].(string)
			groupAvatars[id] = GetAvatarURL(platform, id, true, avatar)
		}

		userNames := make(map[string]string)
		userAvatars := make(map[string]string)
		for _, u := range m.MemberCache {
			id := common.ToString(u["user_id"])
			if name, ok := u["nickname"].(string); ok {
				userNames[id] = name
			}
			platform := ""
			if botID, ok := u["bot_id"].(string); ok {
				if b, ok := m.Bots[botID]; ok {
					platform = b.Platform
				}
			}
			avatar, _ := u["avatar"].(string)
			userAvatars[id] = GetAvatarURL(platform, id, false, avatar)
		}
		for id, f := range m.FriendCache {
			if name, ok := f["nickname"].(string); ok {
				userNames[id] = name
			}
			platform := ""
			if botID, ok := f["bot_id"].(string); ok {
				if b, ok := m.Bots[botID]; ok {
					platform = b.Platform
				}
			}
			avatar, _ := f["avatar"].(string)
			if _, exists := userAvatars[id]; !exists {
				userAvatars[id] = GetAvatarURL(platform, id, false, avatar)
			}
		}

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
			if _, exists := groupAvatars[k]; !exists {
				groupAvatars[k] = GetAvatarURL("", k, true, "")
			}
		}
		ust := make(map[string]int64)
		for k, v := range m.UserStatsToday {
			ust[k] = v
			if _, exists := userAvatars[k]; !exists {
				userAvatars[k] = GetAvatarURL("", k, false, "")
			}
		}

		resp := map[string]interface{}{
			"group_stats":       gs,
			"user_stats":        us,
			"group_stats_today": gst,
			"user_stats_today":  ust,
			"group_names":       groupNames,
			"user_names":        userNames,
			"group_avatars":     groupAvatars,
			"user_avatars":      userAvatars,
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    resp,
		})
	}
}

// HandleGetConfig 获取配置
func HandleGetConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log.Printf("[DEBUG] HandleGetConfig returning config: %+v", m.Config)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"status":  "ok",
			"config":  m.Config,
			"path":    common.GetResolvedConfigPath(),
		})
	}
}

// HandleUpdateConfig 更新配置
func HandleUpdateConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("[DEBUG] HandleUpdateConfig received body: %s", string(bodyBytes))

		// 使用当前配置作为基础，只更新提交的字段
		updatedConfig := *m.Config
		if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "config_format_error"),
			})
			return
		}

		log.Printf("[INFO] Updating config: LogLevel=%s, AutoReply=%v, EnableSkill=%v, PGPort=%d", updatedConfig.LogLevel, updatedConfig.AutoReply, updatedConfig.EnableSkill, updatedConfig.PGPort)

		// 更新管理器中的配置
		*m.Config = updatedConfig

		if err := m.SaveConfig(); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig save error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"status":  "error",
				"message": fmt.Sprintf(common.T(lang, "config_save_failed"), err),
			})
			return
		}

		log.Printf("[INFO] Config updated successfully, file path: %s", common.GetResolvedConfigPath())
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"status":  "ok",
			"message": common.T(lang, "config_updated"),
			"config":  m.Config,
		})
	}
}

// HandleGetRedisConfig 获取 Redis 动态配置
func HandleGetRedisConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		if m.Rdb == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "redis_not_connected"),
			})
			return
		}

		ctx := context.Background()

		// 获取限流配置
		rateLimit, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_CONFIG_RATELIMIT).Result()

		// 获取 TTL 配置
		ttl, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_CONFIG_TTL).Result()

		// 获取路由规则
		rules, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_DYNAMIC_RULES).Result()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"ratelimit": rateLimit,
			"ttl":       ttl,
			"rules":     rules,
		})
	}
}

// HandleUpdateRedisConfig 更新 Redis 动态配置
func HandleUpdateRedisConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		if m.Rdb == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "redis_not_connected"),
			})
			return
		}

		var data struct {
			Type  string            `json:"type"` // ratelimit, ttl, rules
			Data  map[string]string `json:"data"`
			Clear bool              `json:"clear"` // 是否先清空再设置
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		ctx := context.Background()
		var key string
		switch data.Type {
		case "ratelimit":
			key = common.REDIS_KEY_CONFIG_RATELIMIT
		case "ttl":
			key = common.REDIS_KEY_CONFIG_TTL
		case "rules":
			key = common.REDIS_KEY_DYNAMIC_RULES
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_config_type"),
			})
			return
		}

		if data.Clear {
			m.Rdb.Del(ctx, key)
		}

		if len(data.Data) > 0 {
			// 将 map[string]string 转换为 map[string]interface{} 以匹配 Redis HSet
			hsetData := make(map[string]interface{})
			for k, v := range data.Data {
				hsetData[k] = v
			}
			if err := m.Rdb.HSet(ctx, key, hsetData).Err(); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf(common.T(lang, "redis_update_failed"), err),
				})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "redis_config_updated"),
		})
	}
}

// HandleSubscriberWebSocket 处理订阅者 WebSocket 连接 (用于 UI 同步)
func HandleSubscriberWebSocket(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WS] Incoming subscriber connection from %s", r.RemoteAddr)
		claims, ok := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		if !ok {
			log.Printf("[WS] Unauthorized subscriber attempt from %s (No claims)", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Printf("[WS] Subscriber authorized: %s", claims.Username)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[WS] Subscriber upgrade failed for %s: %v", claims.Username, err)
			return
		}

		m.UsersMutex.RLock()
		user := m.Users[claims.Username]
		m.UsersMutex.RUnlock()

		subscriber := &common.Subscriber{
			Conn:  conn,
			Mutex: sync.Mutex{},
			User:  user,
		}

		m.Mutex.Lock()
		if m.Subscribers == nil {
			m.Subscribers = make(map[*websocket.Conn]*common.Subscriber)
		}
		m.Subscribers[conn] = subscriber
		m.Mutex.Unlock()

		log.Printf("Subscriber WebSocket connected: %s (User: %s)", conn.RemoteAddr(), claims.Username)

		m.Mutex.RLock()
		bots := make([]common.BotClient, 0, len(m.Bots))
		for _, bot := range m.Bots {
			bots = append(bots, *bot)
		}
		workers := make([]common.WorkerInfo, 0, len(m.Workers))
		for _, w := range m.Workers {
			workers = append(workers, common.WorkerInfo{
				ID:       w.ID,
				Type:     "worker",
				Status:   "online",
				LastSeen: time.Now().Format("15:04:05"),
			})
		}
		m.Mutex.RUnlock()

		m.CacheMutex.RLock()
		syncState := common.SyncState{
			Type:          "sync_state",
			Groups:        m.GroupCache,
			Friends:       m.FriendCache,
			Members:       m.MemberCache,
			Bots:          bots,
			Workers:       workers,
			TotalMessages: m.TotalMessages,
		}
		m.CacheMutex.RUnlock()

		subscriber.Mutex.Lock()
		if err := conn.WriteJSON(syncState); err != nil {
			log.Printf("[SUBSCRIBER] 发送初始同步状态失败: %v", err)
		}
		subscriber.Mutex.Unlock()

		defer func() {
			m.Mutex.Lock()
			delete(m.Subscribers, conn)
			m.Mutex.Unlock()
			conn.Close()
			log.Printf("Subscriber WebSocket disconnected: %s", conn.RemoteAddr())
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}

// HandleAdminListUsers 获取用户列表
func HandleAdminListUsers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		log.Printf("[DEBUG] HandleAdminListUsers called")

		rows, err := m.DB.Query(m.PrepareQuery("SELECT id, username, is_admin, active, created_at, updated_at FROM users"))
		if err != nil {
			log.Printf("[ERROR] HandleAdminListUsers DB query failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf(common.T(lang, "db_query_failed"), err),
			})
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var id int64
			var username string
			var createdAt, updatedAt interface{}
			var isAdmin, active bool
			if err := rows.Scan(&id, &username, &isAdmin, &active, &createdAt, &updatedAt); err != nil {
				log.Printf("[ERROR] HandleAdminListUsers scan failed: %v", err)
				continue
			}

			var createdAtStr, updatedAtStr string
			if createdAt != nil {
				if t, ok := createdAt.(time.Time); ok {
					createdAtStr = t.Format(time.RFC3339)
				} else if s, ok := createdAt.(string); ok {
					createdAtStr = s
				}
			}
			if updatedAt != nil {
				if t, ok := updatedAt.(time.Time); ok {
					updatedAtStr = t.Format(time.RFC3339)
				} else if s, ok := updatedAt.(string); ok {
					updatedAtStr = s
				}
			}

			users = append(users, map[string]interface{}{
				"id":         id,
				"username":   username,
				"is_admin":   isAdmin,
				"active":     active,
				"created_at": createdAtStr,
				"updated_at": updatedAtStr,
			})
		}

		log.Printf("[DEBUG] HandleAdminListUsers found %d users", len(users))

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"users":   users,
		})
	}
}

// HandleAdminManageUsers 用户管理操作 (create/delete/reset_pwd/toggle_status)
func HandleAdminManageUsers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var req struct {
			Action   string `json:"action"`
			Username string `json:"username"`
			Password string `json:"password"`
			IsAdmin  bool   `json:"is_admin"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		switch req.Action {
		case "create":
			handleAdminCreateUser(m, w, lang, req.Username, req.Password, req.IsAdmin)
		case "edit":
			handleAdminUpdateUser(m, w, lang, req.Username, req.IsAdmin)
		case "delete":
			handleAdminDeleteUser(m, w, lang, req.Username)
		case "reset_password":
			handleAdminResetPassword(m, w, lang, req.Username, req.Password)
		case "toggle_status", "toggle_active":
			handleAdminToggleUser(m, w, lang, req.Username)
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf(common.T(lang, "user_management_invalid_action"), req.Action),
			})
		}
	}
}

func handleAdminCreateUser(m *common.Manager, w http.ResponseWriter, lang, username, password string, isAdmin bool) {
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "user_pwd_empty"),
		})
		return
	}

	hash, err := common.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "password_encrypt_failed"),
		})
		return
	}

	user := &common.User{
		Username:     username,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := m.SaveUserToDB(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf(common.T(lang, "user_create_failed"), err),
		})
		return
	}

	m.UsersMutex.Lock()
	m.Users[username] = user
	m.UsersMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": common.T(lang, "user_created"),
	})
}

func handleAdminUpdateUser(m *common.Manager, w http.ResponseWriter, lang, username string, isAdmin bool) {
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "username_empty"),
		})
		return
	}

	if username == "admin" && !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "cannot_disable_default_admin"),
		})
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET is_admin = ?, updated_at = ? WHERE username = ?"), isAdmin, time.Now(), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf(common.T(lang, "user_update_failed"), err),
		})
		return
	}

	m.UsersMutex.Lock()
	if u, exists := m.Users[username]; exists {
		u.IsAdmin = isAdmin
		u.UpdatedAt = time.Now()
	}
	m.UsersMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": common.T(lang, "user_info_updated"),
	})
}

func handleAdminDeleteUser(m *common.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "cannot_delete_default_admin"),
		})
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("DELETE FROM users WHERE username = ?"), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf(common.T(lang, "user_delete_failed"), err),
		})
		return
	}

	m.UsersMutex.Lock()
	delete(m.Users, username)
	m.UsersMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": common.T(lang, "user_deleted"),
	})
}

func handleAdminResetPassword(m *common.Manager, w http.ResponseWriter, lang, username, password string) {
	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "new_password_empty"),
		})
		return
	}

	hash, err := common.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "password_encrypt_failed"),
		})
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET password_hash = ?, updated_at = ? WHERE username = ?"), hash, time.Now(), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf(common.T(lang, "user_update_failed"), err),
		})
		return
	}

	m.UsersMutex.Lock()
	if u, exists := m.Users[username]; exists {
		u.PasswordHash = hash
		u.UpdatedAt = time.Now()
	}
	m.UsersMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": common.T(lang, "password_reset_success"),
	})
}

func handleAdminToggleUser(m *common.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": common.T(lang, "cannot_disable_default_admin"),
		})
		return
	}

	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()

	user, exists := m.Users[username]
	var currentStatus bool

	if exists {
		currentStatus = user.Active
	} else {
		// If not in cache, load from DB
		err := m.DB.QueryRow(m.PrepareQuery("SELECT active FROM users WHERE username = ?"), username).Scan(&currentStatus)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "user_not_found"),
			})
			return
		}
	}

	newStatus := !currentStatus

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET active = ?, updated_at = ? WHERE username = ?"), newStatus, time.Now(), username); err != nil {
		log.Printf("更新用户状态失败: %v (username: %s, newStatus: %v)", err, username, newStatus)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf(common.T(lang, "user_update_failed"), err),
		})
		return
	}

	// Update cache if it exists
	if exists {
		user.Active = newStatus
	} else {
		// Optionally load full user into cache if needed,
		// but for now just letting it stay out of cache until next login/access
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": common.T(lang, "user_status_updated"),
		"active":  newStatus,
	})
}

// HandleGetRoutingRules 获取所有路由规则
func HandleGetRoutingRules(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"rules":   m.RoutingRules,
		})
	}
}

// HandleSetRoutingRule 设置路由规则
func HandleSetRoutingRule(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var rule struct {
			Key      string `json:"key"`
			WorkerID string `json:"worker_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		if rule.Key == "" || rule.WorkerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "routing_rule_invalid_params"),
			})
			return
		}

		m.Mutex.Lock()
		if m.RoutingRules == nil {
			m.RoutingRules = make(map[string]string)
		}
		m.RoutingRules[rule.Key] = rule.WorkerID
		m.Mutex.Unlock()

		if err := m.SaveRoutingRuleToDB(rule.Key, rule.WorkerID); err != nil {
			log.Printf(common.T(lang, "routing_rule_save_failed"), err)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "routing_rule_set_success"),
		})
	}
}

// HandleDeleteRoutingRule 删除路由规则
func HandleDeleteRoutingRule(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "routing_rule_key_empty"),
			})
			return
		}

		m.Mutex.Lock()
		if _, exists := m.RoutingRules[key]; exists {
			delete(m.RoutingRules, key)
			if err := m.DeleteRoutingRuleFromDB(key); err != nil {
				log.Printf(common.T(lang, "routing_rule_delete_failed"), err)
			}
		}
		m.Mutex.Unlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "routing_rule_delete_success"),
		})
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// HandleDockerLogs 获取 Docker 容器日志
func HandleDockerLogs(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		containerID := r.URL.Query().Get("container_id")
		if containerID == "" {
			containerID = r.URL.Query().Get("id")
		}

		if containerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "container_id_empty"),
			})
			return
		}

		if m.DockerClient == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": common.T(lang, "docker_not_init"),
			})
			return
		}

		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "200",
		}

		reader, err := m.DockerClient.ContainerLogs(r.Context(), containerID, options)
		if err != nil {
			log.Printf(common.T(lang, "get_docker_logs_failed"), err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		defer reader.Close()

		logs, _ := io.ReadAll(reader)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"logs":   string(logs),
		})
	}
}

// HandleGetManual 获取管理员手册
func HandleGetManual(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		manual := map[string]interface{}{
			"title": common.T(lang, "manual_title"),
			"sections": []map[string]interface{}{
				{
					"title":   common.T(lang, "manual_section_quickstart_title"),
					"content": common.T(lang, "manual_section_quickstart_content"),
				},
				{
					"title":   common.T(lang, "manual_section_docker_title"),
					"content": common.T(lang, "manual_section_docker_content"),
				},
				{
					"title":   common.T(lang, "manual_section_routing_title"),
					"content": common.T(lang, "manual_section_routing_content"),
				},
				{
					"title":   common.T(lang, "manual_section_users_title"),
					"content": common.T(lang, "manual_section_users_content"),
				},
			},
			"version": "1.0.0", // 使用硬编码版本号或从配置中获取
		}

		json.NewEncoder(w).Encode(manual)
	}
}
