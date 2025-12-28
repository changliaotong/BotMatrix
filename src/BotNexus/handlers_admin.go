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
		lang := common.GetLangFromRequest(r)

		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			log.Printf(common.T("", "login_request_failed"), err)
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		log.Printf(common.T("", "login_attempt"), loginData.Username, r.RemoteAddr)

		m.UsersMutex.RLock()
		user, exists := m.Users[loginData.Username]
		m.UsersMutex.RUnlock()

		if !exists {
			var u common.User
			result := m.GORMDB.Where("username = ?", loginData.Username).First(&u)
			if result.Error == nil {
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
			common.SendJSONResponse(w, false, common.T(lang, "invalid_username_password"), nil)
			return
		}

		if !user.Active {
			log.Printf("用户未激活: %s", loginData.Username)
			w.WriteHeader(http.StatusForbidden)
			common.SendJSONResponse(w, false, common.T(lang, "user_not_active"), nil)
			return
		}

		token, err := m.GenerateToken(user)
		if err != nil {
			log.Printf(common.T("", "token_generation_failed")+": %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, common.T(lang, "token_generation_failed"), nil)
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		log.Printf(common.T("", "login_success"), user.Username, role)

		common.SendJSONResponse(w, true, "", struct {
			Token string `json:"token"`
			Role  string `json:"role"`
			User  struct {
				ID        int64     `json:"id"`
				Username  string    `json:"username"`
				IsAdmin   bool      `json:"is_admin"`
				Role      string    `json:"role"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"user"`
		}{
			Token: token,
			Role:  role,
			User: struct {
				ID        int64     `json:"id"`
				Username  string    `json:"username"`
				IsAdmin   bool      `json:"is_admin"`
				Role      string    `json:"role"`
				CreatedAt time.Time `json:"created_at"`
			}{
				ID:        user.ID,
				Username:  user.Username,
				IsAdmin:   user.IsAdmin,
				Role:      role,
				CreatedAt: user.CreatedAt,
			},
		})
	}
}

// HandleGetUserInfo 获取当前登录用户信息
func HandleGetUserInfo(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		claims, ok := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			common.SendJSONResponse(w, false, common.T(lang, "not_logged_in"), nil)
			return
		}

		m.UsersMutex.RLock()
		user, exists := m.Users[claims.Username]
		m.UsersMutex.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			common.SendJSONResponse(w, false, common.T(lang, "user_not_found"), nil)
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		common.SendJSONResponse(w, true, "", struct {
			User struct {
				ID        int64     `json:"id"`
				Username  string    `json:"username"`
				IsAdmin   bool      `json:"is_admin"`
				Role      string    `json:"role"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			} `json:"user"`
		}{
			User: struct {
				ID        int64     `json:"id"`
				Username  string    `json:"username"`
				IsAdmin   bool      `json:"is_admin"`
				Role      string    `json:"role"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}{
				ID:        user.ID,
				Username:  user.Username,
				IsAdmin:   user.IsAdmin,
				Role:      role,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

// HandleGetNexusStatus 获取 Nexus 运行状态
func HandleGetNexusStatus(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		common.SendJSONResponse(w, true, "", struct {
			Running   bool `json:"running"`
			Connected bool `json:"connected"`
			Stats     struct {
				OnlineBots    int   `json:"online_bots"`
				OnlineWorkers int   `json:"online_workers"`
				TotalBots     int64 `json:"total_bots"`
				TotalWorkers  int64 `json:"total_workers"`
			} `json:"stats"`
			Version string `json:"version"`
		}{
			Running:   true,
			Connected: onlineBots > 0 || onlineWorkers > 0,
			Stats: struct {
				OnlineBots    int   `json:"online_bots"`
				OnlineWorkers int   `json:"online_workers"`
				TotalBots     int64 `json:"total_bots"`
				TotalWorkers  int64 `json:"total_workers"`
			}{
				OnlineBots:    onlineBots,
				OnlineWorkers: onlineWorkers,
				TotalBots:     botCount,
				TotalWorkers:  workerCount,
			},
			Version: VERSION,
		})
	}
}

// HandleGetStats 获取统计信息的请求
func HandleGetStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		type statsResponse struct {
			Goroutines        int         `json:"goroutines"`
			Uptime            string      `json:"uptime"`
			MemoryAlloc       uint64      `json:"memory_alloc"`
			MemoryTotal       uint64      `json:"memory_total"`
			MemoryUsed        uint64      `json:"memory_used"`
			MemoryFree        uint64      `json:"memory_free"`
			MemoryUsedPercent float64     `json:"memory_used_percent"`
			DiskUsage         string      `json:"disk_usage"`
			BotCount          int         `json:"bot_count"`
			WorkerCount       int         `json:"worker_count"`
			BotCountOffline   int         `json:"bot_count_offline"`
			BotCountTotal     int         `json:"bot_count_total"`
			ActiveGroupsToday int         `json:"active_groups_today"`
			ActiveGroups      int         `json:"active_groups"`
			ActiveUsersToday  int         `json:"active_users_today"`
			ActiveUsers       int         `json:"active_users"`
			MessageCount      int64       `json:"message_count"`
			SentMessageCount  int64       `json:"sent_message_count"`
			CPUUsage          float64     `json:"cpu_usage"`
			StartTime         int64       `json:"start_time"`
			CPUModel          string      `json:"cpu_model"`
			CPUCoresPhysical  int         `json:"cpu_cores_physical"`
			CPUCoresLogical   int         `json:"cpu_cores_logical"`
			CPUFreq           float64     `json:"cpu_freq"`
			OSPlatform        string      `json:"os_platform"`
			OSVersion         string      `json:"os_version"`
			OSArch            string      `json:"os_arch"`
			Timestamp         string      `json:"timestamp"`
			BotsDetail        any         `json:"bots_detail"`
			CPUTrend          []float64   `json:"cpu_trend"`
			MemTrend          []uint64    `json:"mem_trend"`
			MsgTrend          []int64     `json:"msg_trend"`
			SentTrend         []int64     `json:"sent_trend"`
			RecvTrend         []int64     `json:"recv_trend"`
			NetSentTrend      []uint64    `json:"net_sent_trend"`
			NetRecvTrend      []uint64    `json:"net_recv_trend"`
			TopProcesses      any         `json:"top_processes"`
		}

		stats := statsResponse{
			Goroutines:        runtime.NumGoroutine(),
			Uptime:            uptimeStr,
			MemoryAlloc:       mStats.Alloc,
			MemoryTotal:       vm.Total,
			MemoryUsed:        vm.Used,
			MemoryFree:        vm.Free,
			MemoryUsedPercent: vm.UsedPercent,
			DiskUsage:         diskUsageStr,
			BotCount:          onlineBots,
			WorkerCount:       onlineWorkers,
			BotCountOffline:   offlineBots,
			BotCountTotal:     totalBots,
			ActiveGroupsToday: len(m.GroupStatsToday),
			ActiveGroups:      len(m.GroupStats),
			ActiveUsersToday:  len(m.UserStatsToday),
			ActiveUsers:       len(m.UserStats),
			MessageCount:      m.TotalMessages,
			SentMessageCount:  m.SentMessages,
			CPUUsage:          cpuUsage,
			StartTime:         m.StartTime.Unix(),
			CPUModel:          cpuModel,
			CPUCoresPhysical:  cpuCoresPhysical,
			CPUCoresLogical:   cpuCoresLogical,
			CPUFreq:           cpuFreq,
			OSPlatform:        hi.Platform,
			OSVersion:         hi.PlatformVersion,
			OSArch:            hi.KernelArch,
			Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
			BotsDetail:        m.BotDetailedStats,
			CPUTrend:          cpuTrend,
			MemTrend:          memTrend,
			MsgTrend:          msgTrend,
			SentTrend:         sentTrend,
			RecvTrend:         recvTrend,
			NetSentTrend:      netSentTrend,
			NetRecvTrend:      netRecvTrend,
			TopProcesses:      m.TopProcesses,
		}

		common.SendJSONResponse(w, true, "", struct {
			Stats statsResponse `json:"stats"`
		}{
			Stats: stats,
		})
	}
}

// HandleGetSystemStats 获取详细的系统运行统计
func HandleGetSystemStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		type DiskUsage struct {
			Path        string  `json:"path"`
			Total       uint64  `json:"total"`
			Free        uint64  `json:"free"`
			Used        uint64  `json:"used"`
			UsedPercent float64 `json:"usedPercent"`
		}
		var diskUsage []DiskUsage
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
				diskUsage = append(diskUsage, DiskUsage{
					Path:        p.Mountpoint,
					Total:       usage.Total,
					Free:        usage.Free,
					Used:        usage.Used,
					UsedPercent: usage.UsedPercent,
				})
				seenMounts[p.Mountpoint] = true
			}
		}

		netIO, _ := net.IOCounters(false)
		type NetUsage struct {
			Name      string `json:"name"`
			BytesSent uint64 `json:"bytesSent"`
			BytesRecv uint64 `json:"bytesRecv"`
		}
		var netUsage []NetUsage
		for _, ioCounter := range netIO {
			netUsage = append(netUsage, NetUsage{
				Name:      ioCounter.Name,
				BytesSent: ioCounter.BytesSent,
				BytesRecv: ioCounter.BytesRecv,
			})
		}

		interfaces, _ := net.Interfaces()
		type NetInterface struct {
			Name  string `json:"name"`
			Addrs []struct {
				Addr string `json:"addr"`
			} `json:"addrs"`
		}
		var netInterfaces []NetInterface
		for _, i := range interfaces {
			var addrs []struct {
				Addr string `json:"addr"`
			}
			for _, addr := range i.Addrs {
				addrs = append(addrs, struct {
					Addr string `json:"addr"`
				}{
					Addr: addr.Addr,
				})
			}
			netInterfaces = append(netInterfaces, NetInterface{
				Name:  i.Name,
				Addrs: addrs,
			})
		}

		m.HistoryMutex.RLock()
		processList := m.TopProcesses
		m.HistoryMutex.RUnlock()

		stats := struct {
			CPUUsage      float64              `json:"cpu_usage"`
			MemUsage      float64              `json:"mem_usage"`
			MemTotal      uint64               `json:"mem_total"`
			MemFree       uint64               `json:"mem_free"`
			DiskUsage     []DiskUsage          `json:"disk_usage"`
			NetIO         []NetUsage           `json:"net_io"`
			NetInterfaces []NetInterface       `json:"net_interfaces"`
			HostInfo      *host.InfoStat       `json:"host_info"`
			Processes     []common.ProcessInfo `json:"processes"`
			Timestamp     int64                `json:"timestamp"`
		}{
			CPUUsage:      cpuUsage,
			MemUsage:      vm.UsedPercent,
			MemTotal:      vm.Total,
			MemFree:       vm.Free,
			DiskUsage:     diskUsage,
			NetIO:         netUsage,
			NetInterfaces: netInterfaces,
			HostInfo:      hi,
			Processes:     processList,
			Timestamp:     time.Now().Unix(),
		}

		common.SendJSONResponse(w, true, "", struct {
			Stats any `json:"stats"`
		}{
			Stats: stats,
		})
	}
}

// HandleGetLogs 处理获取日志的请求
func HandleGetLogs(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		common.SendJSONResponse(w, true, "", struct {
			Logs    []common.LogEntry `json:"logs"`
			Total   int               `json:"total"`
			HasMore bool              `json:"hasMore"`
		}{
			Logs:    pagedLogs,
			Total:   total,
			HasMore: end < total,
		})
	}
}

// HandleClearLogs 处理清空日志的请求
func HandleClearLogs(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)
		m.ClearLogs()
		common.SendJSONResponse(w, true, common.T(lang, "logs_cleared"), nil)
	}
}

// HandleGetBots 处理获取机器人列表的请求
func HandleGetBots(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		m.StatsMutex.RLock()
		defer m.StatsMutex.RUnlock()

		type BotInfo struct {
			ID            string `json:"id"`
			SelfID        string `json:"self_id"`
			Nickname      string `json:"nickname"`
			Avatar        string `json:"avatar"`
			GroupCount    int    `json:"group_count"`
			FriendCount   int    `json:"friend_count"`
			Connected     string `json:"connected"`
			Platform      string `json:"platform"`
			SentCount     int64  `json:"sent_count"`
			RecvCount     int64  `json:"recv_count"`
			MsgCount      int64  `json:"msg_count"`
			MsgCountToday int64  `json:"msg_count_today"`
			RemoteAddr    string `json:"remote_addr"`
			LastHeartbeat string `json:"last_heartbeat"`
			IsAlive       bool   `json:"is_alive"`
		}
		bots := make([]BotInfo, 0, len(m.Bots))
		for id, bot := range m.Bots {
			remoteAddr := ""
			if bot.Conn != nil {
				remoteAddr = bot.Conn.RemoteAddr().String()
			}

			totalMsg := m.BotStats[id]
			todayMsg := m.BotStatsToday[id]

			bots = append(bots, BotInfo{
				ID:            id,
				SelfID:        bot.SelfID,
				Nickname:      bot.Nickname,
				Avatar:        GetAvatarURL(bot.Platform, bot.SelfID, false, ""),
				GroupCount:    bot.GroupCount,
				FriendCount:   bot.FriendCount,
				Connected:     bot.Connected.Format("2006-01-02 15:04:05"),
				Platform:      bot.Platform,
				SentCount:     bot.SentCount,
				RecvCount:     bot.RecvCount,
				MsgCount:      totalMsg,
				MsgCountToday: todayMsg,
				RemoteAddr:    remoteAddr,
				LastHeartbeat: bot.LastHeartbeat.Format("2006-01-02 15:04:05"),
				IsAlive:       true,
			})
		}

		common.SendJSONResponse(w, true, "", struct {
			Bots []BotInfo `json:"bots"`
		}{
			Bots: bots,
		})
	}
}

// HandleGetWorkers 处理获取Worker列表的请求
func HandleGetWorkers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		type WorkerInfo struct {
			ID           string `json:"id"`
			RemoteAddr   string `json:"remote_addr"`
			Connected    string `json:"connected"`
			HandledCount int64  `json:"handled_count"`
			AvgRTT       string `json:"avg_rtt"`
			LastRTT      string `json:"last_rtt"`
			IsAlive      bool   `json:"is_alive"`
			Status       string `json:"status"`
		}
		workers := make([]WorkerInfo, 0, len(m.Workers))
		for _, worker := range m.Workers {
			workers = append(workers, WorkerInfo{
				ID:           worker.ID,
				RemoteAddr:   worker.ID,
				Connected:    worker.Connected.Format("2006-01-02 15:04:05"),
				HandledCount: worker.HandledCount,
				AvgRTT:       worker.AvgRTT.String(),
				LastRTT:      worker.LastRTT.String(),
				IsAlive:      true,
				Status:       "Online",
			})
		}

		common.SendJSONResponse(w, true, "", struct {
			Workers []WorkerInfo `json:"workers"`
		}{
			Workers: workers,
		})
	}
}

// HandleDockerList 获取 Docker 容器列表
func HandleDockerList(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			// 尝试延迟初始化
			if err := m.InitDockerClient(); err != nil {
				log.Printf("Docker client initialization failed: %v", err)
				// 不要返回 500，而是返回空列表，让前端知道 Docker 未就绪
				common.SendJSONResponse(w, true, common.T(lang, "docker_not_init"), struct {
					Status     string `json:"status"`
					Containers []any  `json:"containers"`
				}{
					Status:     "warning",
					Containers: []any{},
				})
				return
			}
		}

		containers, err := m.DockerClient.ContainerList(r.Context(), types.ContainerListOptions{All: true})
		if err != nil {
			log.Printf(common.T("", "docker_list_failed"), err)
			// 同理，返回空列表而不是 500
			common.SendJSONResponse(w, false, err.Error(), struct {
				Status     string `json:"status"`
				Containers []any  `json:"containers"`
			}{
				Status:     "error",
				Containers: []any{},
			})
			return
		}

		common.SendJSONResponse(w, true, "", struct {
			Status     string            `json:"status"`
			Containers []types.Container `json:"containers"`
		}{
			Status:     "ok",
			Containers: containers,
		})
	}
}

// HandleDockerAction 处理 Docker 容器操作 (start/stop/restart/delete)
func HandleDockerAction(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		var req struct {
			ContainerID string `json:"container_id"`
			Action      string `json:"action"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.DockerClient == nil {
			// 尝试延迟初始化
			if err := m.InitDockerClient(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				common.SendJSONResponse(w, false, common.T(lang, "docker_not_init")+": "+err.Error(), nil)
				return
			}
		}

		var err error
		switch req.Action {
		case "start":
			err = m.DockerClient.ContainerStart(r.Context(), req.ContainerID, container.StartOptions{})
		case "stop":
			timeout := 10
			err = m.DockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
		case "restart":
			timeout := 10
			err = m.DockerClient.ContainerRestart(r.Context(), req.ContainerID, container.RestartOptions{Timeout: &timeout})
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
			common.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		status := "running"
		if req.Action == "stop" {
			status = "exited"
		} else if req.Action == "delete" {
			status = "deleted"
		}
		m.BroadcastDockerEvent(req.Action, req.ContainerID, status)

		common.SendJSONResponse(w, true, "", struct {
			ID string `json:"id"`
		}{
			ID: req.ContainerID,
		})
	}
}

// HandleDockerAddBot 添加机器人容器
func HandleDockerAddBot(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			common.SendJSONResponse(w, false, common.T(lang, "docker_not_init"), struct {
				Status string `json:"status"`
			}{
				Status: "error",
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
				common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_image_not_exists"), imageName, err), struct {
					Status string `json:"status"`
				}{
					Status: "error",
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
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_create_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_start_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		common.SendJSONResponse(w, true, common.T(lang, "bot_deploy_success"), struct {
			Status string `json:"status"`
			ID     string `json:"id"`
		}{
			Status: "ok",
			ID:     resp.ID,
		})
	}
}

// HandleDockerAddWorker 添加 Worker 容器
func HandleDockerAddWorker(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		if m.DockerClient == nil {
			common.SendJSONResponse(w, false, common.T(lang, "docker_not_init"), struct {
				Status string `json:"status"`
			}{
				Status: "error",
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
				common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_image_not_exists"), imageName, err), struct {
					Status string `json:"status"`
				}{
					Status: "error",
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
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_create_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "docker_start_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		common.SendJSONResponse(w, true, common.T(lang, "worker_deploy_success"), struct {
			Status string `json:"status"`
			ID     string `json:"id"`
		}{
			Status: "ok",
			ID:     resp.ID,
		})
	}
}

// HandleChangePassword 修改用户密码
func HandleChangePassword(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		claims, ok := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			common.SendJSONResponse(w, false, common.T(lang, "not_logged_in"), nil)
			return
		}

		var data struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		m.UsersMutex.Lock()
		defer m.UsersMutex.Unlock()

		user, exists := m.Users[claims.Username]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			common.SendJSONResponse(w, false, common.T(lang, "user_not_found"), nil)
			return
		}

		if !common.CheckPassword(data.OldPassword, user.PasswordHash) {
			w.WriteHeader(http.StatusForbidden)
			common.SendJSONResponse(w, false, common.T(lang, "old_password_error"), nil)
			return
		}

		newHash, err := common.HashPassword(data.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, common.T(lang, "password_encrypt_failed"), nil)
			return
		}

		user.PasswordHash = newHash
		user.UpdatedAt = time.Now()

		if err := m.SaveUserToDB(user); err != nil {
			log.Printf(common.T("", "password_update_db_failed"), err)
		}

		common.SendJSONResponse(w, true, common.T(lang, "password_change_success"), nil)
	}
}

// HandleGetMessages 获取最新消息列表
func HandleGetMessages(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		if m.DB == nil {
			common.SendJSONResponse(w, false, "Database not initialized", struct {
				Messages []interface{} `json:"messages"`
			}{
				Messages: []interface{}{},
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
			common.SendJSONResponse(w, false, err.Error(), struct {
				Messages []interface{} `json:"messages"`
			}{
				Messages: []interface{}{},
			})
			return
		}
		defer rows.Close()

		type MessageInfo struct {
			ID          int    `json:"id"`
			MessageID   string `json:"message_id"`
			BotID       string `json:"bot_id"`
			UserID      string `json:"user_id"`
			UserName    string `json:"user_name"`
			UserAvatar  string `json:"user_avatar"`
			GroupID     string `json:"group_id"`
			GroupName   string `json:"group_name"`
			GroupAvatar string `json:"group_avatar"`
			Type        string `json:"type"`
			Content     string `json:"content"`
			CreatedAt   string `json:"created_at"`
		}

		messages := []MessageInfo{}
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
				if n := friend.Nickname; n != "" {
					userName = n
				}
				if a := friend.Avatar; a != "" {
					userAvatar = a
				}
			}
			// 如果在好友缓存没找到头像，尝试在成员缓存找
			if userAvatar == "" && groupID != "" && groupID != "0" {
				memberKey := groupID + "_" + userID
				if member, ok := m.MemberCache[memberKey]; ok {
					if a := member.Avatar; a != "" {
						userAvatar = a
					}
					// 也可以尝试更新下昵称，如果之前没找到的话
					if userName == userID {
						if n := member.Nickname; n != "" {
							userName = n
						}
					}
				}
			}

			if group, ok := m.GroupCache[groupID]; ok {
				if n := group.GroupName; n != "" {
					groupName = n
				}
				if a := group.Avatar; a != "" {
					groupAvatar = a
				}
			}
			m.CacheMutex.RUnlock()

			messages = append(messages, MessageInfo{
				ID:          id,
				MessageID:   messageID,
				BotID:       botID,
				UserID:      userID,
				UserName:    userName,
				UserAvatar:  GetAvatarURL(platform, userID, false, userAvatar),
				GroupID:     groupID,
				GroupName:   groupName,
				GroupAvatar: GetAvatarURL(platform, groupID, true, groupAvatar),
				Type:        msgType,
				Content:     content,
				CreatedAt:   createdAt.Format("2006-01-02 15:04:05"),
			})
		}

		common.SendJSONResponse(w, true, "", struct {
			Messages []MessageInfo `json:"messages"`
		}{
			Messages: messages,
		})
	}
}

// HandleGetContacts 获取联系人列表 (群组和好友)
func HandleGetContacts(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		lang := common.GetLangFromRequest(r)

		// 如果不强制刷新，检查缓存是否为空
		if !refresh && botID != "" {
			m.CacheMutex.RLock()
			hasCache := false
			// 检查群组缓存
			for _, g := range m.GroupCache {
				if g.BotID == botID {
					hasCache = true
					break
				}
			}
			// 如果群组没有，检查好友缓存
			if !hasCache {
				for _, f := range m.FriendCache {
					if f.BotID == botID {
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
				common.SendJSONResponse(w, false, common.T(lang, "bot_not_found"), nil)
				return
			}

			// 检查机器人连接状态和心跳
			if bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute {
				w.WriteHeader(http.StatusServiceUnavailable)
				common.SendJSONResponse(w, false, common.T(lang, "bot_disconnected"), nil)
				return
			}

			echoGroups := "refresh_groups_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
			m.PendingMutex.Lock()
			respChanGroups := make(chan common.InternalMessage, 1)
			m.PendingRequests[echoGroups] = respChanGroups
			m.PendingMutex.Unlock()

			bot.Mutex.Lock()
			err := bot.Conn.WriteJSON(struct {
				Action string `json:"action"`
				Params any    `json:"params"`
				Echo   string `json:"echo"`
			}{
				Action: "get_group_list",
				Params: struct{}{},
				Echo:   echoGroups,
			})
			bot.Mutex.Unlock()

			if err != nil {
				log.Printf("Failed to send get_group_list to bot %s: %v", botID, err)
				w.WriteHeader(http.StatusInternalServerError)
				common.SendJSONResponse(w, false, common.T(lang, "bot_communication_error"), nil)
				return
			}

			// Group fetch with its own timeout
			select {
			case resp := <-respChanGroups:
				if data, ok := resp.Extras["data"].([]any); ok {
					m.CacheMutex.Lock()
					for _, g := range data {
						if group, ok := g.(map[string]any); ok {
							gID := common.ToString(group["group_id"])
							gName := common.ToString(group["group_name"])

							m.GroupCache[gID] = common.GroupInfo{
								GroupID:   gID,
								GroupName: gName,
								BotID:     botID,
								LastSeen:  time.Now(),
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
			respChanFriends := make(chan common.InternalMessage, 1)
			m.PendingRequests[echoFriends] = respChanFriends
			m.PendingMutex.Unlock()

			bot.Mutex.Lock()
			err = bot.Conn.WriteJSON(struct {
				Action string `json:"action"`
				Params any    `json:"params"`
				Echo   string `json:"echo"`
			}{
				Action: "get_friend_list",
				Params: struct{}{},
				Echo:   echoFriends,
			})
			bot.Mutex.Unlock()

			if err == nil {
				// Friend fetch with its own timeout
				select {
				case resp := <-respChanFriends:
					if data, ok := resp.Extras["data"].([]any); ok {
						m.CacheMutex.Lock()
						for _, f := range data {
							if friend, ok := f.(map[string]any); ok {
								uID := common.ToString(friend["user_id"])
								nickname := common.ToString(friend["nickname"])

								m.FriendCache[uID] = common.FriendInfo{
									UserID:   uID,
									Nickname: nickname,
									BotID:    botID,
									LastSeen: time.Now(),
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
				respChanGuilds := make(chan common.InternalMessage, 1)
				m.PendingRequests[echoGuilds] = respChanGuilds
				m.PendingMutex.Unlock()

				bot.Mutex.Lock()
				bot.Conn.WriteJSON(struct {
					Action string `json:"action"`
					Params any    `json:"params"`
					Echo   string `json:"echo"`
				}{
					Action: "get_guild_list",
					Params: struct{}{},
					Echo:   echoGuilds,
				})
				bot.Mutex.Unlock()

				// Guild fetch with its own timeout
				select {
				case resp := <-respChanGuilds:
					if data, ok := resp.Extras["data"].([]any); ok {
						for _, g := range data {
							if guild, ok := g.(map[string]any); ok {
								gID := common.ToString(guild["guild_id"])
								gName := common.ToString(guild["guild_name"])

								m.CacheMutex.Lock()
								m.GroupCache[gID] = common.GroupInfo{
									GroupID:   gID,
									GroupName: gName,
									BotID:     botID,
									IsCached:  true,
									Source:    "get_guild_list",
									LastSeen:  time.Now(),
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

		type ContactInfo struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
			BotID    string `json:"bot_id"`
			Type     string `json:"type"`
		}

		contacts := make([]ContactInfo, 0)
		for _, g := range m.GroupCache {
			gBotID := g.BotID
			if botID == "" || gBotID == botID {
				platform := ""
				m.Mutex.RLock()
				if b, ok := m.Bots[gBotID]; ok {
					platform = b.Platform
				}
				m.Mutex.RUnlock()

				gID := g.GroupID
				providedAvatar := g.Avatar

				contacts = append(contacts, ContactInfo{
					ID:       gID,
					Name:     g.GroupName,
					Nickname: g.GroupName,
					Avatar:   GetAvatarURL(platform, gID, true, providedAvatar),
					BotID:    gBotID,
					Type:     "group",
				})
			}
		}

		for _, f := range m.FriendCache {
			fBotID := f.BotID
			if botID == "" || fBotID == botID {
				platform := ""
				m.Mutex.RLock()
				if b, ok := m.Bots[fBotID]; ok {
					platform = b.Platform
				}
				m.Mutex.RUnlock()

				uID := f.UserID
				providedAvatar := f.Avatar

				contacts = append(contacts, ContactInfo{
					ID:       uID,
					Name:     f.Nickname,
					Nickname: f.Nickname,
					Avatar:   GetAvatarURL(platform, uID, false, providedAvatar),
					BotID:    fBotID,
					Type:     "private",
				})
			}
		}

		common.SendJSONResponse(w, true, "", struct {
			Contacts []ContactInfo `json:"contacts"`
		}{
			Contacts: contacts,
		})
	}
}

// HandleGetGroupMembers 获取群成员列表
func HandleGetGroupMembers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		botID := r.URL.Query().Get("bot_id")
		groupID := r.URL.Query().Get("group_id")
		refresh := r.URL.Query().Get("refresh") == "true"

		if botID == "" || groupID == "" {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "missing_parameters"), nil)
			return
		}

		m.Mutex.RLock()
		bot, ok := m.Bots[botID]
		m.Mutex.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			common.SendJSONResponse(w, false, common.T(lang, "bot_not_found"), nil)
			return
		}

		// 检查机器人连接状态和心跳
		if bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute {
			w.WriteHeader(http.StatusServiceUnavailable)
			common.SendJSONResponse(w, false, common.T(lang, "bot_disconnected"), nil)
			return
		}

		// 1. 优先尝试从缓存读取 (如果不需要强制刷新)
		if !refresh {
			m.CacheMutex.RLock()
			cachedMembers := make([]common.MemberInfo, 0)
			for _, member := range m.MemberCache {
				if member.GroupID == groupID {
					// 注入头像
					m.Mutex.RLock()
					platform := ""
					if b, ok := m.Bots[member.BotID]; ok {
						platform = b.Platform
					}
					m.Mutex.RUnlock()

					uID := member.UserID
					providedAvatar := member.Avatar
					member.Avatar = GetAvatarURL(platform, uID, false, providedAvatar)

					cachedMembers = append(cachedMembers, member)
				}
			}
			m.CacheMutex.RUnlock()

			if len(cachedMembers) > 0 {
				common.SendJSONResponse(w, true, "", struct {
					Data   []common.MemberInfo `json:"data"`
					Cached bool                `json:"cached"`
				}{
					Data:   cachedMembers,
					Cached: true,
				})
				return
			}
		}

		// 2. 如果需要刷新或缓存中没有，则从机器人获取
		echo := "get_members_" + groupID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
		m.PendingMutex.Lock()
		respChan := make(chan common.InternalMessage, 1)
		m.PendingRequests[echo] = respChan
		m.PendingMutex.Unlock()

		defer func() {
			m.PendingMutex.Lock()
			delete(m.PendingRequests, echo)
			m.PendingMutex.Unlock()
		}()

		bot.Mutex.Lock()
		bot.Conn.WriteJSON(struct {
			Action string `json:"action"`
			Params any    `json:"params"`
			Echo   string `json:"echo"`
		}{
			Action: "get_group_member_list",
			Params: struct {
				GroupID string `json:"group_id"`
			}{
				GroupID: groupID,
			},
			Echo: echo,
		})
		bot.Mutex.Unlock()

		select {
		case resp := <-respChan:
			if data, ok := resp.Extras["data"].([]any); ok {
				// 更新缓存
				m.CacheMutex.Lock()
				for _, it := range data {
					if member, ok := it.(map[string]any); ok {
						uID := common.ToString(member["user_id"])
						nickname := common.ToString(member["nickname"])
						card := common.ToString(member["card"])
						key := fmt.Sprintf("%s:%s", groupID, uID)

						m.MemberCache[key] = common.MemberInfo{
							GroupID:  groupID,
							UserID:   uID,
							Nickname: nickname,
							Card:     card,
							BotID:    botID, // 确保缓存中有 bot_id
							Avatar:   GetAvatarURL(bot.Platform, uID, false, common.ToString(member["avatar"])),
							IsCached: true,
							LastSeen: time.Now(),
						}
						// 同时更新返回给前端的数据
						member["avatar"] = GetAvatarURL(bot.Platform, uID, false, common.ToString(member["avatar"]))
						// 持久化到数据库
						go m.SaveMemberToDB(groupID, uID, nickname, card)
					}
				}
				m.CacheMutex.Unlock()

				common.SendJSONResponse(w, true, "", struct {
					Data []any `json:"data"`
				}{
					Data: data,
				})
				return
			}
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			common.SendJSONResponse(w, false, common.T(lang, "bot_timeout"), nil)
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
		lang := common.GetLangFromRequest(r)

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			BotID  string `json:"bot_id"`
			Action string `json:"action"`
			Params any    `json:"params"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		// 如果是内部路由调用，强制设置 action
		if req.Action == "" && r.Header.Get("X-Internal-Action") != "" {
			req.Action = r.Header.Get("X-Internal-Action")
		}

		if req.Action == "batch_send_msg" {
			params, ok := req.Params.(map[string]any)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				common.SendJSONResponse(w, false, common.T(lang, "invalid_params"), nil)
				return
			}
			targets, ok := params["targets"].([]any)
			message, _ := params["message"].(string)
			if !ok || message == "" {
				w.WriteHeader(http.StatusBadRequest)
				common.SendJSONResponse(w, false, common.T(lang, "batch_send_params_error"), nil)
				return
			}

			go func() {
				log.Printf("[BatchSend] Starting batch send for %d targets", len(targets))
				success := 0
				failed := 0
				for _, t := range targets {
					target, ok := t.(map[string]any)
					if !ok {
						continue
					}

					targetID := common.ToString(target["id"])
					targetBotID := common.ToString(target["bot_id"])
					targetType := common.ToString(target["type"])

					m.Mutex.RLock()
					bot, exists := m.Bots[targetBotID]
					m.Mutex.RUnlock()

					if !exists {
						failed++
						continue
					}

					action := "send_group_msg"
					var actionParams any
					if targetType == "private" {
						action = "send_private_msg"
						actionParams = struct {
							UserID  string `json:"user_id"`
							Message string `json:"message"`
						}{
							UserID:  targetID,
							Message: message,
						}
					} else if targetType == "guild" {
						action = "send_msg"
						type GuildParams struct {
							MessageType string `json:"message_type"`
							ChannelID   string `json:"channel_id"`
							Message     string `json:"message"`
							GuildID     string `json:"guild_id,omitempty"`
						}
						gp := GuildParams{
							MessageType: "guild",
							ChannelID:   targetID,
							Message:     message,
						}
						if gid := common.ToString(target["guild_id"]); gid != "" {
							gp.GuildID = gid
						}
						actionParams = gp
					} else {
						actionParams = struct {
							GroupID string `json:"group_id"`
							Message string `json:"message"`
						}{
							GroupID: targetID,
							Message: message,
						}
					}

					echo := fmt.Sprintf("batch|%d|%s", time.Now().UnixNano(), action)
					log.Printf("[API] [Batch] Sending action to bot %s: %s, params: %+v", bot.SelfID, action, actionParams)
					msg := struct {
						Action string `json:"action"`
						Params any    `json:"params"`
						Echo   string `json:"echo"`
					}{
						Action: action,
						Params: actionParams,
						Echo:   echo,
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

			common.SendJSONResponse(w, true, common.T(lang, "batch_send_start", len(targets)), nil)
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
			common.SendJSONResponse(w, false, common.T(lang, "no_available_bot"), nil)
			return
		}

		echo := fmt.Sprintf("web|%d|%s", time.Now().UnixNano(), req.Action)

		respChan := make(chan common.InternalMessage, 1)
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

		msg := struct {
			Action string `json:"action"`
			Params any    `json:"params"`
			Echo   string `json:"echo"`
		}{
			Action: req.Action,
			Params: req.Params,
			Echo:   echo,
		}

		bot.Mutex.Lock()
		err := bot.Conn.WriteJSON(msg)
		bot.Mutex.Unlock()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "send_to_bot_failed"), err), nil)
			return
		}

		select {
		case resp := <-respChan:
			common.SendJSONResponse(w, true, "", resp.Extras)
		case <-time.After(30 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			common.SendJSONResponse(w, false, common.T(lang, "bot_timeout"), nil)
		}
	}
}

// HandleGetChatStats 获取聊天统计信息
func HandleGetChatStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()
		m.CacheMutex.RLock()
		defer m.CacheMutex.RUnlock()

		groupNames := make(map[string]string)
		groupAvatars := make(map[string]string)
		for id, g := range m.GroupCache {
			if g.GroupName != "" {
				groupNames[id] = g.GroupName
			}
			platform := ""
			if b, ok := m.Bots[g.BotID]; ok {
				platform = b.Platform
			}
			avatar := g.Avatar
			groupAvatars[id] = GetAvatarURL(platform, id, true, avatar)
		}

		userNames := make(map[string]string)
		userAvatars := make(map[string]string)
		for _, u := range m.MemberCache {
			id := u.UserID
			if u.Nickname != "" {
				userNames[id] = u.Nickname
			}
			platform := ""
			if b, ok := m.Bots[u.BotID]; ok {
				platform = b.Platform
			}
			avatar := u.Avatar
			userAvatars[id] = GetAvatarURL(platform, id, false, avatar)
		}
		for id, f := range m.FriendCache {
			if f.Nickname != "" {
				userNames[id] = f.Nickname
			}
			platform := ""
			if b, ok := m.Bots[f.BotID]; ok {
				platform = b.Platform
			}
			avatar := f.Avatar
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

		common.SendJSONResponse(w, true, "", struct {
			GroupStats      map[string]int64  `json:"group_stats"`
			UserStats       map[string]int64  `json:"user_stats"`
			GroupStatsToday map[string]int64  `json:"group_stats_today"`
			UserStatsToday  map[string]int64  `json:"user_stats_today"`
			GroupNames      map[string]string `json:"group_names"`
			UserNames       map[string]string `json:"user_names"`
			GroupAvatars    map[string]string `json:"group_avatars"`
			UserAvatars     map[string]string `json:"user_avatars"`
		}{
			GroupStats:      gs,
			UserStats:       us,
			GroupStatsToday: gst,
			UserStatsToday:  ust,
			GroupNames:      groupNames,
			UserNames:       userNames,
			GroupAvatars:    groupAvatars,
			UserAvatars:     userAvatars,
		})
	}
}

// HandleGetConfig 获取配置
func HandleGetConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] HandleGetConfig returning config: %+v", m.Config)

		common.SendJSONResponse(w, true, "", struct {
			Config *common.AppConfig `json:"config"`
			Path   string            `json:"path"`
		}{
			Config: m.Config,
			Path:   common.GetResolvedConfigPath(),
		})
	}
}

// HandleUpdateConfig 更新配置
func HandleUpdateConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("[DEBUG] HandleUpdateConfig received body: %s", string(bodyBytes))

		// 使用当前配置作为基础，只更新提交的字段
		updatedConfig := *m.Config
		if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "config_format_error"), nil)
			return
		}

		log.Printf("[INFO] Updating config: LogLevel=%s, AutoReply=%v, EnableSkill=%v, PGPort=%d", updatedConfig.LogLevel, updatedConfig.AutoReply, updatedConfig.EnableSkill, updatedConfig.PGPort)

		// 更新管理器中的配置
		*m.Config = updatedConfig

		if err := m.SaveConfig(); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig save error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "config_save_failed"), err), nil)
			return
		}

		log.Printf("[INFO] Config updated successfully, file path: %s", common.GetResolvedConfigPath())
		common.SendJSONResponse(w, true, common.T(lang, "config_updated"), struct {
			Config *common.AppConfig `json:"config"`
		}{
			Config: m.Config,
		})
	}
}

// HandleGetRedisConfig 获取 Redis 动态配置
func HandleGetRedisConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		if m.Rdb == nil {
			common.SendJSONResponse(w, false, common.T(lang, "redis_not_connected"), nil)
			return
		}

		ctx := context.Background()

		// 获取限流配置
		rateLimit, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_CONFIG_RATELIMIT).Result()

		// 获取 TTL 配置
		ttl, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_CONFIG_TTL).Result()

		// 获取路由规则
		rules, _ := m.Rdb.HGetAll(ctx, common.REDIS_KEY_DYNAMIC_RULES).Result()

		common.SendJSONResponse(w, true, "", struct {
			Ratelimit map[string]string `json:"ratelimit"`
			TTL       map[string]string `json:"ttl"`
			Rules     map[string]string `json:"rules"`
		}{
			Ratelimit: rateLimit,
			TTL:       ttl,
			Rules:     rules,
		})
	}
}

// HandleUpdateRedisConfig 更新 Redis 动态配置
func HandleUpdateRedisConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		if m.Rdb == nil {
			common.SendJSONResponse(w, false, common.T(lang, "redis_not_connected"), nil)
			return
		}

		var data struct {
			Type  string            `json:"type"` // ratelimit, ttl, rules
			Data  map[string]string `json:"data"`
			Clear bool              `json:"clear"` // 是否先清空再设置
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
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
			common.SendJSONResponse(w, false, common.T(lang, "invalid_config_type"), nil)
			return
		}

		if data.Clear {
			m.Rdb.Del(ctx, key)
		}

		if len(data.Data) > 0 {
			// 将 map[string]string 转换为 map[string]any 以匹配 Redis HSet
			hsetData := make(map[string]any)
			for k, v := range data.Data {
				hsetData[k] = v
			}
			if err := m.Rdb.HSet(ctx, key, hsetData).Err(); err != nil {
				common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "redis_update_failed"), err), nil)
				return
			}
		}

		common.SendJSONResponse(w, true, common.T(lang, "redis_config_updated"), nil)
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
		lang := common.GetLangFromRequest(r)

		log.Printf("[DEBUG] HandleAdminListUsers called")

		var dbUsers []common.User
		if err := m.GORMDB.Find(&dbUsers).Error; err != nil {
			log.Printf("[ERROR] HandleAdminListUsers GORM query failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "db_query_failed"), err), nil)
			return
		}

		type UserInfo struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			IsAdmin   bool   `json:"is_admin"`
			Active    bool   `json:"active"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}

		users := make([]UserInfo, 0)
		for _, u := range dbUsers {
			users = append(users, UserInfo{
				ID:        u.ID,
				Username:  u.Username,
				IsAdmin:   u.IsAdmin,
				Active:    u.Active,
				CreatedAt: u.CreatedAt.Format(time.RFC3339),
				UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
			})
		}

		log.Printf("[DEBUG] HandleAdminListUsers found %d users", len(users))

		common.SendJSONResponse(w, true, "", struct {
			Users []UserInfo `json:"users"`
		}{
			Users: users,
		})
	}
}

// HandleAdminManageUsers 用户管理操作 (create/delete/reset_pwd/toggle_status)
func HandleAdminManageUsers(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		var req struct {
			Action   string `json:"action"`
			Username string `json:"username"`
			Password string `json:"password"`
			IsAdmin  bool   `json:"is_admin"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
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
			common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_management_invalid_action"), req.Action), nil)
		}
	}
}

func handleAdminCreateUser(m *common.Manager, w http.ResponseWriter, lang, username, password string, isAdmin bool) {
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		common.SendJSONResponse(w, false, common.T(lang, "user_pwd_empty"), nil)
		return
	}

	hash, err := common.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, common.T(lang, "password_encrypt_failed"), nil)
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
		common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_create_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	m.Users[username] = user
	m.UsersMutex.Unlock()

	common.SendJSONResponse(w, true, common.T(lang, "user_created"), nil)
}

func handleAdminUpdateUser(m *common.Manager, w http.ResponseWriter, lang, username string, isAdmin bool) {
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		common.SendJSONResponse(w, false, common.T(lang, "username_empty"), nil)
		return
	}

	if username == "admin" && !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		common.SendJSONResponse(w, false, common.T(lang, "cannot_disable_default_admin"), nil)
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET is_admin = ?, updated_at = ? WHERE username = ?"), isAdmin, time.Now(), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_update_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	if u, exists := m.Users[username]; exists {
		u.IsAdmin = isAdmin
		u.UpdatedAt = time.Now()
	}
	m.UsersMutex.Unlock()

	common.SendJSONResponse(w, true, common.T(lang, "user_info_updated"), nil)
}

func handleAdminDeleteUser(m *common.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		common.SendJSONResponse(w, false, common.T(lang, "cannot_delete_default_admin"), nil)
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("DELETE FROM users WHERE username = ?"), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_delete_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	delete(m.Users, username)
	m.UsersMutex.Unlock()

	common.SendJSONResponse(w, true, common.T(lang, "user_deleted"), nil)
}

func handleAdminResetPassword(m *common.Manager, w http.ResponseWriter, lang, username, password string) {
	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		common.SendJSONResponse(w, false, common.T(lang, "new_password_empty"), nil)
		return
	}

	hash, err := common.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, common.T(lang, "password_encrypt_failed"), nil)
		return
	}

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET password_hash = ?, updated_at = ? WHERE username = ?"), hash, time.Now(), username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_update_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	if u, exists := m.Users[username]; exists {
		u.PasswordHash = hash
		u.UpdatedAt = time.Now()
	}
	m.UsersMutex.Unlock()

	common.SendJSONResponse(w, true, common.T(lang, "password_reset_success"), nil)
}

func handleAdminToggleUser(m *common.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		common.SendJSONResponse(w, false, common.T(lang, "cannot_disable_default_admin"), nil)
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
			common.SendJSONResponse(w, false, common.T(lang, "user_not_found"), nil)
			return
		}
	}

	newStatus := !currentStatus

	if _, err := m.DB.Exec(m.PrepareQuery("UPDATE users SET active = ?, updated_at = ? WHERE username = ?"), newStatus, time.Now(), username); err != nil {
		log.Printf("更新用户状态失败: %v (username: %s, newStatus: %v)", err, username, newStatus)
		w.WriteHeader(http.StatusInternalServerError)
		common.SendJSONResponse(w, false, fmt.Sprintf(common.T(lang, "user_update_failed"), err), nil)
		return
	}

	// Update cache if it exists
	if exists {
		user.Active = newStatus
	} else {
		// Optionally load full user into cache if needed,
		// but for now just letting it stay out of cache until next login/access
	}

	common.SendJSONResponse(w, true, common.T(lang, "user_status_updated"), struct {
		Active bool `json:"active"`
	}{
		Active: newStatus,
	})
}

// HandleGetRoutingRules 获取所有路由规则
func HandleGetRoutingRules(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		common.SendJSONResponse(w, true, "", struct {
			Rules map[string]string `json:"rules"`
		}{
			Rules: m.RoutingRules,
		})
	}
}

// HandleSetRoutingRule 设置路由规则
func HandleSetRoutingRule(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		var rule struct {
			Key      string `json:"key"`
			WorkerID string `json:"worker_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		if rule.Key == "" || rule.WorkerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "routing_rule_invalid_params"), nil)
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

		common.SendJSONResponse(w, true, common.T(lang, "routing_rule_set_success"), nil)
	}
}

// HandleDeleteRoutingRule 删除路由规则
func HandleDeleteRoutingRule(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "routing_rule_key_empty"), nil)
			return
		}

		m.Mutex.Lock()
		defer m.Mutex.Unlock()

		if _, exists := m.RoutingRules[key]; exists {
			delete(m.RoutingRules, key)
			if err := m.DeleteRoutingRuleFromDB(key); err != nil {
				log.Printf(common.T(lang, "routing_rule_delete_failed"), err)
			}
		}

		common.SendJSONResponse(w, true, common.T(lang, "routing_rule_delete_success"), nil)
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
		lang := common.GetLangFromRequest(r)

		containerID := r.URL.Query().Get("id")

		if containerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			common.SendJSONResponse(w, false, common.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.DockerClient == nil {
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, common.T(lang, "docker_not_init"), nil)
			return
		}

		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "100",
			Follow:     false,
		}

		reader, err := m.DockerClient.ContainerLogs(r.Context(), containerID, options)
		if err != nil {
			log.Printf(common.T(lang, "get_docker_logs_failed"), err)
			w.WriteHeader(http.StatusInternalServerError)
			common.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		defer reader.Close()

		logs, _ := io.ReadAll(reader)

		common.SendJSONResponse(w, true, "", struct {
			Logs string `json:"logs"`
		}{
			Logs: string(logs),
		})
	}
}

// HandleGetManual 获取管理员手册
func HandleGetManual(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := common.GetLangFromRequest(r)

		type ManualSection struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		type ManualInfo struct {
			Title    string          `json:"title"`
			Sections []ManualSection `json:"sections"`
			Version  string          `json:"version"`
		}

		manual := ManualInfo{
			Title: common.T(lang, "manual_title"),
			Sections: []ManualSection{
				{
					Title:   common.T(lang, "manual_section_quickstart_title"),
					Content: common.T(lang, "manual_section_quickstart_content"),
				},
				{
					Title:   common.T(lang, "manual_section_docker_title"),
					Content: common.T(lang, "manual_section_docker_content"),
				},
				{
					Title:   common.T(lang, "manual_section_routing_title"),
					Content: common.T(lang, "manual_section_routing_content"),
				},
				{
					Title:   common.T(lang, "manual_section_users_title"),
					Content: common.T(lang, "manual_section_users_content"),
				},
			},
			Version: "1.0.0", // 使用硬编码版本号或从配置中获取
		}

		common.SendJSONResponse(w, true, "", manual)
	}
}
