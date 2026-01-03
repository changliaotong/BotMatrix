package app

import (
	"BotMatrix/common/bot"
	"BotMatrix/common/config"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	docker_types "github.com/docker/docker/api/types"
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
// @Summary 管理后台登录
// @Description 使用用户名和密码登录管理后台，获取访问 Token
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body object true "登录凭据"
// @Success 200 {object} utils.JSONResponse "登录成功，返回 Token"
// @Failure 401 {object} utils.JSONResponse "用户名或密码错误"
// @Failure 403 {object} utils.JSONResponse "用户未激活"
// @Router /api/login [post]
func HandleLogin(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			log.Printf("%s: %v", utils.T("", "login_request_failed"), err)
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		log.Printf("%s: %s, %s", utils.T("", "login_attempt"), loginData.Username, r.RemoteAddr)

		user, exists := m.GetOrLoadUser(loginData.Username)

		if !exists || !utils.CheckPassword(loginData.Password, user.PasswordHash) {
			log.Printf("%s: %s", utils.T("", "invalid_username_password|用户名或密码错误"), loginData.Username)
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_username_password|用户名或密码错误"), nil)
			return
		}

		if !user.Active {
			log.Printf("用户未激活: %s", loginData.Username)
			w.WriteHeader(http.StatusForbidden)
			utils.SendJSONResponse(w, false, utils.T(lang, "user_not_active|该账号已被禁用"), nil)
			return
		}

		token, err := m.GenerateToken(user)
		if err != nil {
			log.Printf(utils.T("", "token_generation_failed|Token生成失败")+": %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, utils.T(lang, "token_generation_failed|Token生成失败"), nil)
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		log.Printf(utils.T("", "login_success"), user.Username, role)

		utils.SendJSONResponse(w, true, "", struct {
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
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "用户信息"
// @Router /api/user/info [get]
// @Router /api/me [get]
func HandleGetUserInfo(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		user, exists := m.GetOrLoadUser(claims.Username)

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			utils.SendJSONResponse(w, false, utils.T(lang, "user_not_found"), nil)
			return
		}

		role := "user"
		if user.IsAdmin {
			role = "admin"
		}

		utils.SendJSONResponse(w, true, "", struct {
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
// @Summary 获取 Nexus 状态
// @Description 获取 BotNexus 服务的整体运行状态和版本信息
// @Tags System
// @Produce json
// @Success 200 {object} utils.JSONResponse "服务状态信息"
// @Router /api/admin/nexus/status [get]
func HandleGetNexusStatus(m *bot.Manager) http.HandlerFunc {
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

		utils.SendJSONResponse(w, true, "", struct {
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
// @Summary 获取系统统计
// @Description 获取 CPU、内存、磁盘、在线机器人及消息量趋势等统计数据
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "详细统计数据"
// @Router /api/admin/stats [get]
func HandleGetStats(m *bot.Manager) http.HandlerFunc {
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
		topProcesses := append([]types.ProcInfo{}, m.TopProcesses...)
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
			Goroutines        int       `json:"goroutines"`
			Uptime            string    `json:"uptime"`
			MemoryAlloc       uint64    `json:"memory_alloc"`
			MemoryTotal       uint64    `json:"memory_total"`
			MemoryUsed        uint64    `json:"memory_used"`
			MemoryFree        uint64    `json:"memory_free"`
			MemoryUsedPercent float64   `json:"memory_used_percent"`
			DiskUsage         string    `json:"disk_usage"`
			BotCount          int       `json:"bot_count"`
			WorkerCount       int       `json:"worker_count"`
			BotCountOffline   int       `json:"bot_count_offline"`
			BotCountTotal     int       `json:"bot_count_total"`
			ActiveGroupsToday int       `json:"active_groups_today"`
			ActiveGroups      int       `json:"active_groups"`
			ActiveUsersToday  int       `json:"active_users_today"`
			ActiveUsers       int       `json:"active_users"`
			MessageCount      int64     `json:"message_count"`
			SentMessageCount  int64     `json:"sent_message_count"`
			CPUUsage          float64   `json:"cpu_usage"`
			StartTime         int64     `json:"start_time"`
			CPUModel          string    `json:"cpu_model"`
			CPUCoresPhysical  int       `json:"cpu_cores_physical"`
			CPUCoresLogical   int       `json:"cpu_cores_logical"`
			CPUFreq           float64   `json:"cpu_freq"`
			OSPlatform        string    `json:"os_platform"`
			OSVersion         string    `json:"os_version"`
			OSArch            string    `json:"os_arch"`
			Timestamp         string    `json:"timestamp"`
			BotsDetail        any       `json:"bots_detail"`
			CPUTrend          []float64 `json:"cpu_trend"`
			MemTrend          []uint64  `json:"mem_trend"`
			MsgTrend          []int64   `json:"msg_trend"`
			SentTrend         []int64   `json:"sent_trend"`
			RecvTrend         []int64   `json:"recv_trend"`
			NetSentTrend      []uint64  `json:"net_sent_trend"`
			NetRecvTrend      []uint64  `json:"net_recv_trend"`
			TopProcesses      any       `json:"top_processes"`
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
			TopProcesses:      topProcesses,
		}

		utils.SendJSONResponse(w, true, "", struct {
			Stats statsResponse `json:"stats"`
		}{
			Stats: stats,
		})
	}
}

// HandleGetSystemStats 获取详细的系统运行统计
// @Summary 获取详细系统统计
// @Description 获取更详尽的系统硬件使用情况，包括 CPU、内存、各磁盘分区等
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "详细统计数据"
// @Router /api/system/stats [get]
func HandleGetSystemStats(m *bot.Manager) http.HandlerFunc {
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
			CPUUsage      float64          `json:"cpu_usage"`
			MemUsage      float64          `json:"mem_usage"`
			MemTotal      uint64           `json:"mem_total"`
			MemFree       uint64           `json:"mem_free"`
			DiskUsage     []DiskUsage      `json:"disk_usage"`
			NetIO         []NetUsage       `json:"net_io"`
			NetInterfaces []NetInterface   `json:"net_interfaces"`
			HostInfo      *host.InfoStat   `json:"host_info"`
			Processes     []types.ProcInfo `json:"processes"`
			Timestamp     int64            `json:"timestamp"`
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

		utils.SendJSONResponse(w, true, "", struct {
			Stats any `json:"stats"`
		}{
			Stats: stats,
		})
	}
}

// HandleGetLogs 处理获取日志的请求
// @Summary 获取系统日志
// @Description 分页获取系统运行日志，支持按级别、来源和关键词过滤，支持排序
// @Tags System
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param level query string false "日志级别"
// @Param botId query string false "机器人 ID 或来源"
// @Param search query string false "搜索关键词"
// @Param sortBy query string false "排序字段"
// @Param sortOrder query string false "排序顺序 (asc/desc)"
// @Success 200 {object} utils.JSONResponse "日志列表"
// @Router /api/admin/logs [get]
func HandleGetLogs(m *bot.Manager) http.HandlerFunc {
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
		var filteredLogs []types.LogEntry
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

		var pagedLogs []types.LogEntry
		if start < total {
			if end > total {
				end = total
			}
			pagedLogs = filteredLogs[start:end]
		} else {
			pagedLogs = []types.LogEntry{}
		}

		utils.SendJSONResponse(w, true, "", struct {
			Logs    []types.LogEntry `json:"logs"`
			Total   int              `json:"total"`
			HasMore bool             `json:"hasMore"`
		}{
			Logs:    pagedLogs,
			Total:   total,
			HasMore: end < total,
		})
	}
}

// HandleClearLogs 处理清空日志的请求
// @Summary 清空系统日志
// @Description 清空内存中的所有系统日志记录
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "清空结果"
// @Router /api/admin/logs/clear [post]
func HandleClearLogs(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		m.ClearLogs()
		utils.SendJSONResponse(w, true, utils.T(lang, "logs_cleared"), nil)
	}
}

// HandleGetBots 处理获取机器人列表的请求
// @Summary 获取机器人列表
// @Description 获取所有当前连接的机器人（OneBot 客户端）及其状态信息
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "机器人列表"
// @Router /api/admin/bots [get]
func HandleGetBots(m *bot.Manager) http.HandlerFunc {
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

		utils.SendJSONResponse(w, true, "", struct {
			Bots []BotInfo `json:"bots"`
		}{
			Bots: bots,
		})
	}
}

// HandleGetWorkers 处理获取Worker列表的请求
// @Summary 获取 Worker 列表
// @Description 获取所有当前在线和历史连接过的 Worker 列表及其详细状态
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "Worker 列表"
// @Router /api/admin/workers [get]
func HandleGetWorkers(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		onlineWorkers := make(map[string]bool)
		for _, w := range m.Workers {
			onlineWorkers[w.ID] = true
		}
		m.Mutex.RUnlock()

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
		workers := make([]WorkerInfo, 0)

		// 1. 获取在线 Worker
		m.Mutex.RLock()
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
		m.Mutex.RUnlock()

		// 2. 从 Redis 获取离线 Worker
		if m.Rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			iter := m.Rdb.Scan(ctx, 0, "botmatrix:worker:*:last_seen", 100).Iterator()
			for iter.Next(ctx) {
				key := iter.Val()
				parts := strings.Split(key, ":")
				if len(parts) < 4 {
					continue
				}
				workerID := parts[2]

				if onlineWorkers[workerID] {
					continue
				}

				lastSeen, _ := m.Rdb.Get(ctx, key).Int64()
				status := "Offline"
				isAlive := false
				// 如果 1 分钟内有更新，认为是存活的 (可能仅通过 Redis 通信)
				if time.Now().Unix()-lastSeen < 60 {
					status = "Active (Redis)"
					isAlive = true
				}

				workers = append(workers, WorkerInfo{
					ID:         workerID,
					RemoteAddr: workerID,
					Connected:  time.Unix(lastSeen, 0).Format("2006-01-02 15:04:05"),
					IsAlive:    isAlive,
					Status:     status,
				})
			}
		}

		utils.SendJSONResponse(w, true, "", struct {
			Workers []WorkerInfo `json:"workers"`
		}{
			Workers: workers,
		})
	}
}

// HandleListPlugins 处理获取插件列表的请求
// @Summary 获取插件列表
// @Description 获取系统中所有插件（包括 Nexus 中心插件和所有 Worker 节点的插件）及其运行状态
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "插件列表"
// @Router /api/admin/plugins [get]
func HandleListPlugins(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = utils.GetLangFromRequest(r) // 使用 _ 忽略未使用的变量

		// 1. 获取中心插件 (Nexus 自身的插件)
		centralPlugins := m.PluginManager.GetPlugins()

		type PluginInfo struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Author      string `json:"author"`
			State       string `json:"state"`
			Type        string `json:"type"`        // "central" 或 "worker"
			Source      string `json:"source"`      // "nexus" 或 workerID
			Online      bool   `json:"online"`      // 所在节点是否在线
			IsInternal  bool   `json:"is_internal"` // 是否为内部插件
		}

		var allPlugins []PluginInfo

		// 添加中心插件 (外部)
		for id, versions := range centralPlugins {
			for _, p := range versions {
				allPlugins = append(allPlugins, PluginInfo{
					ID:          id,
					Name:        p.Config.Name,
					Version:     p.Config.Version,
					Description: p.Config.Description,
					Author:      p.Config.Author,
					State:       p.State,
					Type:        "central",
					Source:      "nexus",
					Online:      true,
					IsInternal:  false,
				})
			}
		}

		// 添加中心插件 (内部)
		for name, p := range m.PluginManager.GetInternalPlugins() {
			allPlugins = append(allPlugins, PluginInfo{
				ID:          name,
				Name:        name,
				Version:     p.Version(),
				Description: p.Description(),
				Author:      "system",
				State:       "running",
				Type:        "central",
				Source:      "nexus",
				Online:      true,
				IsInternal:  true,
			})
		}

		// 2. 获取所有在线 Worker 的插件信息
		m.Mutex.RLock()
		onlineWorkers := make(map[string]bool)
		workers := make([]*types.WorkerClient, 0, len(m.Workers))
		for _, w := range m.Workers {
			workers = append(workers, w)
			onlineWorkers[w.ID] = true
		}
		m.Mutex.RUnlock()

		for _, worker := range workers {
			hasMetadata := false
			if worker.Metadata != nil {
				if pluginsRaw, ok := worker.Metadata["plugins"]; ok {
					if pluginsList, ok := pluginsRaw.([]any); ok {
						hasMetadata = true
						for _, pRaw := range pluginsList {
							if pMap, ok := pRaw.(map[string]any); ok {
								allPlugins = append(allPlugins, PluginInfo{
									ID:          utils.ToString(pMap["id"]),
									Name:        utils.ToString(pMap["name"]),
									Version:     utils.ToString(pMap["version"]),
									Description: utils.ToString(pMap["description"]),
									Author:      utils.ToString(pMap["author"]),
									State:       utils.ToString(pMap["state"]),
									Type:        "worker",
									Source:      worker.ID,
									Online:      true,
									IsInternal:  utils.ToBool(pMap["is_internal"]),
								})
							}
						}
					}
				}
			}

			// 如果内存中没有元数据且 Redis 可用，尝试从 Redis 读取
			if !hasMetadata && m.Rdb != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				key := fmt.Sprintf("botmatrix:worker:%s:plugins", worker.ID)
				val, err := m.Rdb.Get(ctx, key).Result()
				cancel()

				if err == nil && val != "" {
					var pluginsList []map[string]any
					if err := json.Unmarshal([]byte(val), &pluginsList); err == nil {
						for _, pMap := range pluginsList {
							allPlugins = append(allPlugins, PluginInfo{
								ID:          utils.ToString(pMap["id"]),
								Name:        utils.ToString(pMap["name"]),
								Version:     utils.ToString(pMap["version"]),
								Description: utils.ToString(pMap["description"]),
								Author:      utils.ToString(pMap["author"]),
								State:       utils.ToString(pMap["state"]),
								Type:        "worker",
								Source:      worker.ID,
								Online:      true,
								IsInternal:  utils.ToBool(pMap["is_internal"]),
							})
						}
					}
				}
			}
		}

		// 3. 从 Redis 中发现离线 Worker 的插件
		if m.Rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 使用 SCAN 查找所有 worker 插件键
			iter := m.Rdb.Scan(ctx, 0, "botmatrix:worker:*:plugins", 100).Iterator()
			for iter.Next(ctx) {
				key := iter.Val()
				// 提取 workerID: botmatrix:worker:WORKERID:plugins
				parts := strings.Split(key, ":")
				if len(parts) < 4 {
					continue
				}
				workerID := parts[2]

				// 如果该 Worker 在线，跳过（已经在上面处理过了）
				if onlineWorkers[workerID] {
					continue
				}

				// 检查 Worker 的活跃时间 (Redis last_seen)
				isOnline := false
				lastSeenVal, err := m.Rdb.Get(ctx, fmt.Sprintf("botmatrix:worker:%s:last_seen", workerID)).Int64()
				if err == nil {
					// 如果最后活跃时间在 60 秒内，视为在线
					if time.Now().Unix()-lastSeenVal < 60 {
						isOnline = true
					}
				}

				// 读取离线 Worker 的插件列表
				val, err := m.Rdb.Get(ctx, key).Result()
				if err == nil && val != "" {
					var pluginsList []map[string]any
					if err := json.Unmarshal([]byte(val), &pluginsList); err == nil {
						for _, pMap := range pluginsList {
							allPlugins = append(allPlugins, PluginInfo{
								ID:          utils.ToString(pMap["id"]),
								Name:        utils.ToString(pMap["name"]),
								Version:     utils.ToString(pMap["version"]),
								Description: utils.ToString(pMap["description"]),
								Author:      utils.ToString(pMap["author"]),
								State:       utils.ToString(pMap["state"]),
								Type:        "worker",
								Source:      workerID,
								Online:      isOnline,
								IsInternal:  utils.ToBool(pMap["is_internal"]),
							})
						}
					}
				}
			}
		}

		utils.SendJSONResponse(w, true, "", struct {
			Plugins []PluginInfo `json:"plugins"`
		}{
			Plugins: allPlugins,
		})
	}
}

// HandlePluginAction 处理插件操作 (启动、停止、重启)
// @Summary 操作插件
// @Description 启动、停止、重启或重载指定的插件
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "操作参数 (id, action, source)"
// @Success 200 {object} utils.JSONResponse "操作结果"
// @Router /api/admin/plugins/action [post]
func HandlePluginAction(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var req struct {
			ID     string `json:"id"`
			Action string `json:"action"` // "start", "stop", "restart", "reload"
			Source string `json:"source"` // "nexus" 或 workerID
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if req.Source == "nexus" {
			// 操作中心插件
			var err error
			switch req.Action {
			case "start":
				err = m.PluginManager.StartPlugin(req.ID, "")
			case "stop":
				err = m.PluginManager.StopPlugin(req.ID, "")
			case "restart":
				err = m.PluginManager.RestartPlugin(req.ID, "")
			case "reload":
				// 重新加载插件目录
				centralPluginsDir := filepath.Join("..", "..", "plugins", "central")
				if _, err := os.Stat(centralPluginsDir); os.IsNotExist(err) {
					os.MkdirAll(centralPluginsDir, 0755)
				}
				err = m.PluginManager.LoadPlugins(centralPluginsDir)
			default:
				err = fmt.Errorf("unknown action: %s", req.Action)
			}

			if err != nil {
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		} else {
			// 操作 Worker 插件，通过 Redis Pub/Sub 发送指令给 Worker
			// 发送指令给 Worker
			cmd := map[string]any{
				"action": "plugin_action",
				"params": map[string]any{
					"id":     req.ID,
					"action": req.Action,
				},
			}

			if m.Rdb != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				cmdJSON, _ := json.Marshal(cmd)
				channel := fmt.Sprintf("botmatrix:worker:%s:commands", req.Source)
				if err := m.Rdb.Publish(ctx, channel, cmdJSON).Err(); err != nil {
					utils.SendJSONResponse(w, false, "Failed to publish command to Redis: "+err.Error(), nil)
					return
				}
			} else {
				// Fallback to WebSocket if Redis is not available
				var worker *types.WorkerClient
				m.Mutex.RLock()
				for _, w := range m.Workers {
					if w.ID == req.Source {
						worker = w
						break
					}
				}
				m.Mutex.RUnlock()

				if worker == nil {
					utils.SendJSONResponse(w, false, "Worker not found and Redis is unavailable", nil)
					return
				}

				if err := worker.Conn.WriteJSON(cmd); err != nil {
					utils.SendJSONResponse(w, false, "Failed to send command via WebSocket: "+err.Error(), nil)
					return
				}
			}
		}

		utils.SendJSONResponse(w, true, "Success", nil)
	}
}

// HandleInstallPlugin 处理插件安装 (上传 .bmpk 文件)
// @Summary 安装插件
// @Description 上传 .bmpk 插件包并安装到指定节点（Nexus 或 Worker）
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param plugin formData file true "插件文件 (.bmpk)"
// @Param target formData string false "目标节点 (nexus 或 workerID)"
// @Success 200 {object} utils.JSONResponse "安装结果"
// @Router /api/admin/plugins/install [post]
func HandleInstallPlugin(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析上传的文件
		if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max
			utils.SendJSONResponse(w, false, "Failed to parse form: "+err.Error(), nil)
			return
		}

		file, header, err := r.FormFile("plugin")
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to get file: "+err.Error(), nil)
			return
		}
		defer file.Close()

		// 2. 检查文件名后缀
		if !strings.HasSuffix(strings.ToLower(header.Filename), ".bmpk") {
			utils.SendJSONResponse(w, false, "Invalid file type. Only .bmpk files are allowed.", nil)
			return
		}

		// 3. 保存到临时文件
		tmpFile, err := os.CreateTemp("", "plugin-*.bmpk")
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to create temp file: "+err.Error(), nil)
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, file); err != nil {
			utils.SendJSONResponse(w, false, "Failed to save file: "+err.Error(), nil)
			return
		}

		// 4. 确定安装目标
		target := r.FormValue("target") // "nexus" 或 workerID
		if target == "" {
			target = "nexus"
		}

		if target == "nexus" {
			// 安装到中心端
			centralPluginsDir := filepath.Join("..", "..", "plugins", "central")
			if err := m.PluginManager.InstallPlugin(tmpFile.Name(), centralPluginsDir); err != nil {
				utils.SendJSONResponse(w, false, "Installation failed: "+err.Error(), nil)
				return
			}
			utils.SendJSONResponse(w, true, "Plugin installed successfully on Nexus", nil)
		} else {
			// 读取文件内容并进行 Base64 编码
			content, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				utils.SendJSONResponse(w, false, "Failed to read plugin file: "+err.Error(), nil)
				return
			}
			base64Content := base64.StdEncoding.EncodeToString(content)

			// 插件安装指令
			cmd := map[string]any{
				"action": "plugin_install",
				"params": map[string]any{
					"filename": header.Filename,
					"content":  base64Content,
				},
			}

			if m.Rdb != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				cmdJSON, _ := json.Marshal(cmd)
				channel := fmt.Sprintf("botmatrix:worker:%s:commands", target)
				if err := m.Rdb.Publish(ctx, channel, cmdJSON).Err(); err != nil {
					utils.SendJSONResponse(w, false, "Failed to publish installation command to Redis: "+err.Error(), nil)
					return
				}
				utils.SendJSONResponse(w, true, "Plugin installation command published to Redis for worker", nil)
			} else {
				// Fallback to WebSocket
				var worker *types.WorkerClient
				m.Mutex.RLock()
				for _, w := range m.Workers {
					if w.ID == target {
						worker = w
						break
					}
				}
				m.Mutex.RUnlock()

				if worker == nil {
					utils.SendJSONResponse(w, false, "Worker not found and Redis is unavailable", nil)
					return
				}

				if err := worker.Conn.WriteJSON(cmd); err != nil {
					utils.SendJSONResponse(w, false, "Failed to send installation command via WebSocket: "+err.Error(), nil)
					return
				}
				utils.SendJSONResponse(w, true, "Plugin installation command sent via WebSocket to worker", nil)
			}
		}
	}
}

// HandleDeletePlugin 处理插件删除
// @Summary 删除插件
// @Description 从指定节点（Nexus 或 Worker）彻底删除指定的插件及其文件
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "删除参数 (id, version, source)"
// @Success 200 {object} utils.JSONResponse "删除结果"
// @Router /api/admin/plugins/delete [post]
func HandleDeletePlugin(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var req struct {
			ID      string `json:"id"`
			Version string `json:"version"`
			Source  string `json:"source"` // "nexus" 或 workerID
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if req.Source == "nexus" {
			// 删除中心插件
			plugin := m.PluginManager.GetPlugin(req.ID, req.Version)
			if plugin == nil {
				utils.SendJSONResponse(w, false, "Plugin not found", nil)
				return
			}

			// 停止运行中的插件
			if plugin.State == "running" {
				m.PluginManager.StopPlugin(req.ID, req.Version)
			}

			// 删除文件目录
			if err := os.RemoveAll(plugin.Dir); err != nil {
				utils.SendJSONResponse(w, false, "Failed to delete plugin files: "+err.Error(), nil)
				return
			}

			// 从内存中移除
			m.PluginManager.RemovePlugin(req.ID, req.Version)

			// 重新加载插件列表以更新内存状态
			centralPluginsDir := filepath.Join("..", "..", "plugins", "central")
			m.PluginManager.LoadPlugins(centralPluginsDir)

			utils.SendJSONResponse(w, true, "Plugin deleted successfully", nil)
		} else {
			// 发送删除指令给 Worker
			cmd := map[string]any{
				"action": "plugin_delete",
				"params": map[string]any{
					"id":      req.ID,
					"version": req.Version,
				},
			}

			if m.Rdb != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				cmdJSON, _ := json.Marshal(cmd)
				channel := fmt.Sprintf("botmatrix:worker:%s:commands", req.Source)
				if err := m.Rdb.Publish(ctx, channel, cmdJSON).Err(); err != nil {
					utils.SendJSONResponse(w, false, "Failed to publish deletion command to Redis: "+err.Error(), nil)
					return
				}
				utils.SendJSONResponse(w, true, "Plugin deletion command published to Redis for worker", nil)
			} else {
				// Fallback to WebSocket
				var worker *types.WorkerClient
				m.Mutex.RLock()
				for _, w := range m.Workers {
					if w.ID == req.Source {
						worker = w
						break
					}
				}
				m.Mutex.RUnlock()

				if worker == nil {
					utils.SendJSONResponse(w, false, "Worker not found and Redis is unavailable", nil)
					return
				}

				if err := worker.Conn.WriteJSON(cmd); err != nil {
					utils.SendJSONResponse(w, false, "Failed to send deletion command via WebSocket: "+err.Error(), nil)
					return
				}
				utils.SendJSONResponse(w, true, "Plugin deletion command sent via WebSocket to worker", nil)
			}
		}
	}
}

// HandleDockerList 获取 Docker 容器列表
// @Summary 获取 Docker 容器列表
// @Description 获取系统宿主机上的所有 Docker 容器列表
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "容器列表"
// @Router /api/admin/docker/list [get]
// @Router /api/docker/list [get]
// @Router /api/docker/containers [get]
func HandleDockerList(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		if m.DockerClient == nil {
			// 尝试延迟初始化
			if err := m.InitDockerClient(); err != nil {
				log.Printf("Docker client initialization failed: %v", err)
				// 不要返回 500，而是返回空列表，让前端知道 Docker 未就绪
				utils.SendJSONResponse(w, true, utils.T(lang, "docker_not_init"), struct {
					Status     string `json:"status"`
					Containers []any  `json:"containers"`
				}{
					Status:     "warning",
					Containers: []any{},
				})
				return
			}
		}

		containers, err := m.DockerClient.ContainerList(r.Context(), docker_types.ContainerListOptions{All: true})
		if err != nil {
			log.Printf(utils.T("", "docker_list_failed"), err)
			// 同理，返回空列表而不是 500
			utils.SendJSONResponse(w, false, err.Error(), struct {
				Status     string `json:"status"`
				Containers []any  `json:"containers"`
			}{
				Status:     "error",
				Containers: []any{},
			})
			return
		}

		utils.SendJSONResponse(w, true, "", struct {
			Status     string                   `json:"status"`
			Containers []docker_types.Container `json:"containers"`
		}{
			Status:     "ok",
			Containers: containers,
		})
	}
}

// HandleDockerAction 处理 Docker 容器操作 (start/stop/restart/delete)
// @Summary 操作 Docker 容器
// @Description 启动、停止、重启或删除指定的 Docker 容器
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "操作参数 (container_id, action)"
// @Success 200 {object} utils.JSONResponse "操作结果"
// @Router /api/admin/docker/action [post]
// @Router /api/docker/start [post]
// @Router /api/docker/stop [post]
// @Router /api/docker/restart [post]
// @Router /api/docker/remove [post]
func HandleDockerAction(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var req struct {
			ContainerID string `json:"container_id"`
			Action      string `json:"action"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.DockerClient == nil {
			// 检查是否是 Online 平台的机器人，如果是则不需要 Docker
			isOnlineBot := false
			m.Mutex.RLock()
			if bot, ok := m.Bots[req.ContainerID]; ok && bot.Platform == "Online" {
				isOnlineBot = true
			}
			m.Mutex.RUnlock()

			if !isOnlineBot {
				// 尝试延迟初始化
				if err := m.InitDockerClient(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					utils.SendJSONResponse(w, false, utils.T(lang, "docker_not_init")+": "+err.Error(), nil)
					return
				}
			}
		}

		// 处理 Online 平台机器人的动作
		m.Mutex.RLock()
		bot, exists := m.Bots[req.ContainerID]
		m.Mutex.RUnlock()

		if exists && bot.Platform == "Online" {
			if req.Action == "delete" {
				// 从内存移除
				m.Mutex.Lock()
				delete(m.Bots, req.ContainerID)
				m.Mutex.Unlock()

				// 从数据库移除
				if m.GORMDB != nil {
					if err := m.GORMDB.Where("self_id = ? AND platform = ?", req.ContainerID, "Online").Delete(&models.BotEntityGORM{}).Error; err != nil {
						log.Printf("[Admin] Failed to delete Online Bot from DB: %v", err)
					}
				}

				m.BroadcastDockerEvent("delete", req.ContainerID, "deleted")
				utils.SendJSONResponse(w, true, "Online bot deleted successfully", struct {
					ID string `json:"id"`
				}{
					ID: req.ContainerID,
				})
				return
			} else if req.Action == "restart" || req.Action == "start" {
				// 对于 Online 机器人，这些动作只是更新状态
				bot.Connected = time.Now()
				m.BroadcastDockerEvent(req.Action, req.ContainerID, "running")
				utils.SendJSONResponse(w, true, "Online bot status updated", struct {
					ID string `json:"id"`
				}{
					ID: req.ContainerID,
				})
				return
			} else if req.Action == "stop" {
				// 模拟停止
				bot.Connected = time.Time{}
				m.BroadcastDockerEvent("stop", req.ContainerID, "exited")
				utils.SendJSONResponse(w, true, "Online bot stopped", struct {
					ID string `json:"id"`
				}{
					ID: req.ContainerID,
				})
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
			err = m.DockerClient.ContainerRestart(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
		case "delete":
			timeout := 5
			m.DockerClient.ContainerStop(r.Context(), req.ContainerID, container.StopOptions{Timeout: &timeout})
			err = m.DockerClient.ContainerRemove(r.Context(), req.ContainerID, docker_types.ContainerRemoveOptions{Force: true})
		default:
			err = fmt.Errorf(utils.T(lang, "unsupported_action"), req.Action)
		}

		if err != nil {
			log.Printf(utils.T("", "docker_action_failed"), req.Action, err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		status := "running"
		if req.Action == "stop" {
			status = "exited"
		} else if req.Action == "delete" {
			status = "deleted"
		}
		m.BroadcastDockerEvent(req.Action, req.ContainerID, status)

		utils.SendJSONResponse(w, true, "", struct {
			ID string `json:"id"`
		}{
			ID: req.ContainerID,
		})
	}
}

// HandleDockerAddBot 添加机器人容器
// @Summary 添加机器人容器
// @Description 在 Docker 中创建并启动一个新的机器人容器，支持多种平台
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "机器人配置 (platform, image, env, cmd)"
// @Success 200 {object} utils.JSONResponse "添加结果"
// @Router /api/admin/docker/add-bot [post]
// @Router /api/docker/add-bot [post]
func HandleDockerAddBot(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var req struct {
			Platform string            `json:"platform"`
			Image    string            `json:"image"`
			Env      map[string]string `json:"env"`
			Cmd      []string          `json:"cmd"`
		}

		// 默认值
		imageName := "botmatrix-wxbot"
		platform := "WeChat"
		envVars := make(map[string]string)
		cmd := []string{"python", "onebot.py"}

		// 尝试解析请求体
		if r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 写回 Body 供后续使用（如果有）
			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &req); err == nil {
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
		}

		log.Printf("[Admin] Add Bot Request: Platform=%s, Image=%s", platform, imageName)

		// 处理在线机器人 (Web 模拟) - 不需要 Docker
		if platform == "Online" {
			botID := envVars["ONLINE_BOT_ID"]
			botName := envVars["ONLINE_BOT_NAME"]
			if botID == "" {
				botID = fmt.Sprintf("%d", time.Now().Unix())
			}
			if botName == "" {
				botName = "Online Bot " + botID
			}

			m.Mutex.Lock()
			if m.Bots == nil {
				m.Bots = make(map[string]*types.BotClient)
			}
			m.Bots[botID] = &types.BotClient{
				SelfID:    botID,
				Nickname:  botName,
				Platform:  "Online",
				Protocol:  "v11",
				Connected: time.Now(),
			}
			m.Mutex.Unlock()

			// 初始化模拟联系人 (群组和好友)
			m.CacheMutex.Lock()
			if m.GroupCache == nil {
				m.GroupCache = make(map[string]types.GroupInfo)
			}
			if m.FriendCache == nil {
				m.FriendCache = make(map[string]types.FriendInfo)
			}

			// 添加一个模拟群组
			mockGroupID := "10001"
			m.GroupCache[mockGroupID] = types.GroupInfo{
				BotID:     botID,
				GroupID:   mockGroupID,
				GroupName: "模拟群聊 (Online)",
			}

			// 添加一个模拟好友
			mockFriendID := "admin"
			m.FriendCache[mockFriendID] = types.FriendInfo{
				BotID:    botID,
				UserID:   mockFriendID,
				Nickname: "管理员 (Mock)",
			}
			m.CacheMutex.Unlock()

			// 持久化到数据库
			if err := m.SaveBotToDB(botID, botName, "Online", "v11"); err != nil {
				log.Printf("[Admin] Failed to save Online Bot to DB: %v", err)
			}

			log.Printf("[Admin] Added Online Bot: %s (%s)", botName, botID)
			utils.SendJSONResponse(w, true, utils.T(lang, "bot_deploy_success"), struct {
				Status string `json:"status"`
				ID     string `json:"id"`
			}{
				Status: "ok",
				ID:     botID,
			})
			return
		}

		// 其他平台需要 Docker
		if m.DockerClient == nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "docker_not_init"), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		ctx := context.Background()

		_, _, err := m.DockerClient.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			log.Printf(utils.T("", "docker_pulling_image"), imageName)
			reader, err := m.DockerClient.ImagePull(ctx, imageName, docker_types.ImagePullOptions{})
			if err != nil {
				log.Printf(utils.T("", "docker_pull_failed"), imageName, err)
				utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_image_not_exists"), imageName, err), struct {
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
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_create_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, docker_types.ContainerStartOptions{}); err != nil {
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_start_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		utils.SendJSONResponse(w, true, utils.T(lang, "bot_deploy_success"), struct {
			Status string `json:"status"`
			ID     string `json:"id"`
		}{
			Status: "ok",
			ID:     resp.ID,
		})
	}
}

// HandleDockerAddWorker 添加 Worker 容器
// @Summary 添加 Worker 容器
// @Description 在 Docker 中创建并启动一个新的 Worker 容器
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "Worker 配置 (image, env, cmd, name)"
// @Success 200 {object} utils.JSONResponse "添加结果"
// @Router /api/admin/docker/add-worker [post]
// @Router /api/docker/add-worker [post]
func HandleDockerAddWorker(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		if m.DockerClient == nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "docker_not_init"), struct {
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
			log.Printf(utils.T("", "docker_pulling_image"), imageName)
			reader, err := m.DockerClient.ImagePull(ctx, imageName, docker_types.ImagePullOptions{})
			if err != nil {
				log.Printf(utils.T("", "docker_pull_failed"), imageName, err)
				utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_image_not_exists"), imageName, err), struct {
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
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_create_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		if err := m.DockerClient.ContainerStart(ctx, resp.ID, docker_types.ContainerStartOptions{}); err != nil {
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "docker_start_container_failed"), err), struct {
				Status string `json:"status"`
			}{
				Status: "error",
			})
			return
		}

		m.BroadcastDockerEvent("create", resp.ID, "running")

		utils.SendJSONResponse(w, true, utils.T(lang, "worker_deploy_success"), struct {
			Status string `json:"status"`
			ID     string `json:"id"`
		}{
			Status: "ok",
			ID:     resp.ID,
		})
	}
}

// HandleChangePassword 修改用户密码
// @Summary 修改密码
// @Description 修改当前登录用户的登录密码
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "密码参数 (old_password, new_password)"
// @Success 200 {object} utils.JSONResponse "修改结果"
// @Router /api/user/password [post]
func HandleChangePassword(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		var data struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		user, exists := m.GetOrLoadUser(claims.Username)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			utils.SendJSONResponse(w, false, utils.T(lang, "user_not_found"), nil)
			return
		}

		m.UsersMutex.Lock()
		defer m.UsersMutex.Unlock()

		if !utils.CheckPassword(data.OldPassword, user.PasswordHash) {
			w.WriteHeader(http.StatusForbidden)
			utils.SendJSONResponse(w, false, utils.T(lang, "old_password_error"), nil)
			return
		}

		newHash, err := utils.HashPassword(data.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, utils.T(lang, "password_encrypt_failed"), nil)
			return
		}

		user.PasswordHash = newHash
		user.SessionVersion++ // 递增 SessionVersion 以使旧 Token 失效
		user.UpdatedAt = time.Now()

		if err := m.SaveUserToDB(user); err != nil {
			log.Printf(utils.T("", "password_update_db_failed"), err)
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "password_change_success"), nil)
	}
}

// HandleGetMessages 获取最新消息列表
// @Summary 获取消息列表
// @Description 从数据库获取最新的聊天消息记录，包括私聊和群聊
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "获取的消息数量 (默认 50)"
// @Success 200 {object} utils.JSONResponse "消息列表"
// @Router /api/admin/messages [get]
func HandleGetMessages(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		if m.DB == nil {
			utils.SendJSONResponse(w, false, "Database not initialized", struct {
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
			utils.SendJSONResponse(w, false, err.Error(), struct {
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

		utils.SendJSONResponse(w, true, "", struct {
			Messages []MessageInfo `json:"messages"`
		}{
			Messages: messages,
		})
	}
}

// HandleGetContacts 获取联系人列表 (群组和好友)
// @Summary 获取联系人列表
// @Description 获取指定机器人的群组和好友列表，支持强制刷新
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bot_id query string false "机器人 ID (GET 请求)"
// @Param refresh query bool false "是否强制刷新 (GET 请求)"
// @Param body body object false "请求体 (POST 请求，包含 bot_id 和 refresh)"
// @Success 200 {object} utils.JSONResponse "联系人列表"
// @Router /api/admin/contacts [get]
// @Router /api/admin/contacts [post]
func HandleGetContacts(m *bot.Manager) http.HandlerFunc {
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

		lang := utils.GetLangFromRequest(r)

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
				utils.SendJSONResponse(w, false, utils.T(lang, "bot_not_found"), nil)
				return
			}

			// 检查机器人连接状态和心跳
			if bot.Platform != "Online" && (bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute) {
				w.WriteHeader(http.StatusServiceUnavailable)
				utils.SendJSONResponse(w, false, utils.T(lang, "bot_disconnected"), nil)
				return
			}

			// 对于 Online 机器人，直接跳过真实通信，因为它们没有真实连接
			if bot.Platform == "Online" {
				log.Printf("[OnlineBot] Skipping real contact refresh for Online bot: %s", botID)
			} else {
				echoGroups := "refresh_groups_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
				m.PendingMutex.Lock()
				respChanGroups := make(chan types.InternalMessage, 1)
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
					utils.SendJSONResponse(w, false, utils.T(lang, "bot_communication_error"), nil)
					return
				}

				// Group fetch with its own timeout
				select {
				case resp := <-respChanGroups:
					if data, ok := resp.Extras["data"].([]any); ok {
						m.CacheMutex.Lock()
						for _, g := range data {
							if group, ok := g.(map[string]any); ok {
								gID := utils.ToString(group["group_id"])
								gName := utils.ToString(group["group_name"])

								m.GroupCache[gID] = types.GroupInfo{
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
					log.Printf(utils.T("", "contacts_timeout_groups"), botID)
				}

				m.PendingMutex.Lock()
				delete(m.PendingRequests, echoGroups)
				m.PendingMutex.Unlock()

				// Friend fetch
				echoFriends := "refresh_friends_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
				m.PendingMutex.Lock()
				respChanFriends := make(chan types.InternalMessage, 1)
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
					select {
					case resp := <-respChanFriends:
						if data, ok := resp.Extras["data"].([]any); ok {
							m.CacheMutex.Lock()
							for _, f := range data {
								if friend, ok := f.(map[string]any); ok {
									fID := utils.ToString(friend["user_id"])
									fNick := utils.ToString(friend["nickname"])

									m.FriendCache[fID] = types.FriendInfo{
										UserID:   fID,
										Nickname: fNick,
										BotID:    botID,
										LastSeen: time.Now(),
									}
									go m.SaveFriendToDB(fID, fNick, botID)
								}
							}
							m.CacheMutex.Unlock()
						}
					case <-time.After(8 * time.Second):
						log.Printf(utils.T("", "contacts_timeout_friends"), botID)
					}
				}

				m.PendingMutex.Lock()
				delete(m.PendingRequests, echoFriends)
				m.PendingMutex.Unlock()
			}

			if bot.Platform == "qq_guild" || bot.Platform == "guild" {
				echoGuilds := "refresh_guilds_" + botID + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
				m.PendingMutex.Lock()
				respChanGuilds := make(chan types.InternalMessage, 1)
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
								gID := utils.ToString(guild["guild_id"])
								gName := utils.ToString(guild["guild_name"])

								m.CacheMutex.Lock()
								m.GroupCache[gID] = types.GroupInfo{
									GroupID:   gID,
									GroupName: gName,
									BotID:     botID,
									IsCached:  true,
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

		utils.SendJSONResponse(w, true, "", struct {
			Contacts []ContactInfo `json:"contacts"`
		}{
			Contacts: contacts,
		})
	}
}

// HandleGetGroupMembers 获取群成员列表
// @Summary 获取群成员列表
// @Description 获取指定群组的成员列表，支持从机器人实时获取或从缓存读取
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param bot_id query string true "机器人 ID"
// @Param group_id query string true "群组 ID"
// @Param refresh query boolean false "是否强制刷新缓存"
// @Success 200 {object} utils.JSONResponse "成员列表"
// @Router /api/admin/group/members [get]
func HandleGetGroupMembers(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		botID := r.URL.Query().Get("bot_id")
		groupID := r.URL.Query().Get("group_id")
		refresh := r.URL.Query().Get("refresh") == "true"

		if botID == "" || groupID == "" {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "missing_parameters"), nil)
			return
		}

		m.Mutex.RLock()
		bot, ok := m.Bots[botID]
		m.Mutex.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			utils.SendJSONResponse(w, false, utils.T(lang, "bot_not_found"), nil)
			return
		}

		// 检查机器人连接状态和心跳
		if bot.Conn == nil || time.Since(bot.LastHeartbeat) > 5*time.Minute {
			w.WriteHeader(http.StatusServiceUnavailable)
			utils.SendJSONResponse(w, false, utils.T(lang, "bot_disconnected"), nil)
			return
		}

		// 1. 优先尝试从缓存读取 (如果不需要强制刷新)
		if !refresh {
			m.CacheMutex.RLock()
			cachedMembers := make([]types.MemberInfo, 0)
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
				utils.SendJSONResponse(w, true, "", struct {
					Data   []types.MemberInfo `json:"data"`
					Cached bool               `json:"cached"`
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
		respChan := make(chan types.InternalMessage, 1)
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
						uID := utils.ToString(member["user_id"])
						nickname := utils.ToString(member["nickname"])
						card := utils.ToString(member["card"])
						role := utils.ToString(member["role"])
						key := fmt.Sprintf("%s:%s", groupID, uID)

						m.MemberCache[key] = types.MemberInfo{
							GroupID:  groupID,
							UserID:   uID,
							Nickname: nickname,
							Card:     card,
							Role:     role,
							BotID:    botID, // 确保缓存中有 bot_id
							Avatar:   GetAvatarURL(bot.Platform, uID, false, utils.ToString(member["avatar"])),
							IsCached: true,
							LastSeen: time.Now(),
						}
						// 同时更新返回给前端的数据
						member["avatar"] = GetAvatarURL(bot.Platform, uID, false, utils.ToString(member["avatar"]))
						// 持久化到数据库
						go m.SaveMemberToDB(groupID, uID, nickname, card, role)
					}
				}
				m.CacheMutex.Unlock()

				utils.SendJSONResponse(w, true, "", struct {
					Data []any `json:"data"`
				}{
					Data: data,
				})
				return
			}
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			utils.SendJSONResponse(w, false, utils.T(lang, "bot_timeout"), nil)
			return
		}
	}
}

// HandleProxyAvatar 代理头像请求
// @Summary 代理头像
// @Description 代理并缓存外部头像图片，解决跨域或防盗链问题
// @Tags System
// @Param url query string true "原始头像 URL"
// @Success 200 {file} image "图片流"
// @Router /api/proxy/avatar [get]
func HandleProxyAvatar(m *bot.Manager) http.HandlerFunc {
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
// @Summary 批量发送消息
// @Description 向多个目标（群或私聊）批量发送相同内容的群发消息
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "群发参数 (targets, message)"
// @Success 200 {object} utils.JSONResponse "任务启动结果"
// @Router /api/admin/batch_send [post]
func HandleBatchSend(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 拦截并设置 action 为 batch_send_msg
		// 这样可以复用 HandleSendAction 的逻辑
		r.Header.Set("X-Internal-Action", "batch_send_msg")
		HandleSendAction(m)(w, r)
	}
}

// HandleSendAction 处理发送 API 动作
// @Summary 发送 API 动作
// @Description 向机器人发送 OneBot 标准 API 动作
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "动作参数 (bot_id, action, params)"
// @Success 200 {object} utils.JSONResponse "动作执行结果"
// @Router /api/action [post]
// @Router /api/smart_action [post]
func HandleSendAction(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

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
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
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
				utils.SendJSONResponse(w, false, utils.T(lang, "invalid_params"), nil)
				return
			}
			targets, ok := params["targets"].([]any)
			message, _ := params["message"].(string)
			if !ok || message == "" {
				w.WriteHeader(http.StatusBadRequest)
				utils.SendJSONResponse(w, false, utils.T(lang, "batch_send_params_error"), nil)
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

					targetID := utils.ToString(target["id"])
					targetBotID := utils.ToString(target["bot_id"])
					targetType := utils.ToString(target["type"])

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
						if gid := utils.ToString(target["guild_id"]); gid != "" {
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

			utils.SendJSONResponse(w, true, utils.T(lang, "batch_send_start", len(targets)), nil)
			return
		}

		m.Mutex.RLock()
		var bot *types.BotClient
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
			utils.SendJSONResponse(w, false, utils.T(lang, "no_available_bot"), nil)
			return
		}

		echo := fmt.Sprintf("web|%d|%s", time.Now().UnixNano(), req.Action)

		respChan := make(chan types.InternalMessage, 1)
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
		var err error

		// --- 优化：后台测试聊天绕过机器人通道，直接调用大模型 ---
		// 检查是否是数字员工，如果是，且是发送消息动作，直接由 AI 回复
		if m.DigitalEmployeeService != nil && m.AIIntegrationService != nil &&
			(req.Action == "send_private_msg" || req.Action == "send_group_msg" || req.Action == "send_msg") {
			employee, _ := m.DigitalEmployeeService.GetEmployeeByBotID(bot.SelfID)
			if employee != nil {
				params, ok := req.Params.(map[string]any)
				if ok {
					message := utils.ToString(params["message"])
					userID := utils.ToString(params["user_id"])
					if userID == "" {
						userID = "admin"
					}
					groupID := utils.ToString(params["group_id"])

					incomingMsg := types.InternalMessage{
						ID:         fmt.Sprintf("msg_%d", time.Now().UnixNano()),
						Time:       time.Now().Unix(),
						Platform:   bot.Platform,
						SelfID:     bot.SelfID,
						PostType:   "message",
						RawMessage: message,
						UserID:     userID,
						GroupID:    groupID,
					}
					if groupID != "" {
						incomingMsg.MessageType = "group"
					} else {
						incomingMsg.MessageType = "private"
					}

					log.Printf("[Admin] Digital Employee %s detected in inject, bypassing channel for direct AI response", bot.SelfID)

					response, err := m.AIIntegrationService.ChatWithEmployee(employee, incomingMsg)
					if err == nil && response != "" {
						bot.Mutex.Unlock() // 提前释放锁
						go func() {
							m.PendingMutex.Lock()
							if ch, ok := m.PendingRequests[echo]; ok {
								ch <- types.InternalMessage{
									SelfID: bot.SelfID,
									Time:   time.Now().Unix(),
									Extras: map[string]any{
										"status":  "ok",
										"retcode": 0,
										"data": map[string]any{
											"message_id": 12345,
											"reply":      response,
										},
									},
								}
							}
							m.PendingMutex.Unlock()
						}()
						return
					}
					log.Printf("[Admin] AI response failed or empty, falling back to normal flow: %v", err)
				}
			}
		}

		if bot.Conn != nil {
			err = bot.Conn.WriteJSON(msg)
		} else if bot.Platform == "Online" {

			// 模拟成功响应给 WebUI
			go func() {
				time.Sleep(50 * time.Millisecond)
				m.PendingMutex.Lock()
				if ch, ok := m.PendingRequests[echo]; ok {
					ch <- types.InternalMessage{
						SelfID: bot.SelfID,
						Time:   time.Now().Unix(),
						Extras: map[string]any{
							"status":  "ok",
							"retcode": 0,
							"data": map[string]any{
								"message_id": 12345,
							},
						},
					}
				}
				m.PendingMutex.Unlock()
			}()
		} else {
			err = fmt.Errorf("bot connection is closed")
		}
		bot.Mutex.Unlock()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "send_to_bot_failed"), err), nil)
			return
		}

		select {
		case resp := <-respChan:
			utils.SendJSONResponse(w, true, "", resp.Extras)
		case <-time.After(30 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			utils.SendJSONResponse(w, false, utils.T(lang, "bot_timeout"), nil)
		}
	}
}

// HandleGetChatStats 获取聊天统计信息
// @Summary 获取聊天统计
// @Description 获取所有群聊和私聊的各种统计数据，包括消息量、活跃度等
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "聊天统计数据"
// @Router /api/stats/chat [get]
func HandleGetChatStats(m *bot.Manager) http.HandlerFunc {
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

		utils.SendJSONResponse(w, true, "", struct {
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
// @Summary 获取配置
// @Description 获取当前应用的完整配置信息
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "应用配置"
// @Router /api/admin/config [get]
func HandleGetConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] HandleGetConfig returning config: %+v", m.Config)

		utils.SendJSONResponse(w, true, "", struct {
			Config *config.AppConfig `json:"config"`
			Path   string            `json:"path"`
		}{
			Config: m.Config,
			Path:   config.GetResolvedConfigPath(),
		})
	}
}

// HandleUpdateConfig 更新配置
// @Summary 更新配置
// @Description 更新当前应用的全局配置信息并持久化到文件
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body config.AppConfig true "新的配置对象"
// @Success 200 {object} utils.JSONResponse "更新后的配置"
// @Router /api/admin/config [post]
func HandleUpdateConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("[DEBUG] HandleUpdateConfig received body: %s", string(bodyBytes))

		// 使用当前配置作为基础，只更新提交的字段
		updatedConfig := *m.Config
		if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "config_format_error"), nil)
			return
		}

		log.Printf("[INFO] Updating config: LogLevel=%s, AutoReply=%v, EnableSkill=%v, PGPort=%d", updatedConfig.LogLevel, updatedConfig.AutoReply, updatedConfig.EnableSkill, updatedConfig.PGPort)

		// 更新管理器中的配置
		*m.Config = updatedConfig

		if err := m.SaveConfig(); err != nil {
			log.Printf("[ERROR] HandleUpdateConfig save error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "config_save_failed"), err), nil)
			return
		}

		log.Printf("[INFO] Config updated successfully, file path: %s", config.GetResolvedConfigPath())
		utils.SendJSONResponse(w, true, utils.T(lang, "config_updated"), struct {
			Config *config.AppConfig `json:"config"`
		}{
			Config: m.Config,
		})
	}
}

// HandleGetRedisConfig 获取 Redis 动态配置
// @Summary 获取 Redis 动态配置
// @Description 获取存储在 Redis 中的限流、TTL 和路由规则等动态配置
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "Redis 配置信息"
// @Router /api/admin/redis/config [get]
func HandleGetRedisConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		if m.Rdb == nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "redis_not_connected"), nil)
			return
		}

		ctx := context.Background()

		// 获取限流配置
		rateLimit, _ := m.Rdb.HGetAll(ctx, config.REDIS_KEY_CONFIG_RATELIMIT).Result()

		// 获取 TTL 配置
		ttl, _ := m.Rdb.HGetAll(ctx, config.REDIS_KEY_CONFIG_TTL).Result()

		// 获取路由规则
		rules, _ := m.Rdb.HGetAll(ctx, config.REDIS_KEY_DYNAMIC_RULES).Result()

		utils.SendJSONResponse(w, true, "", struct {
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
// @Summary 更新 Redis 配置
// @Description 更新存储在 Redis 中的动态配置项（限流、TTL 或路由）
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "配置更新参数 (type, data, clear)"
// @Success 200 {object} utils.JSONResponse "更新结果"
// @Router /api/admin/redis/config [post]
func HandleUpdateRedisConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		if m.Rdb == nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "redis_not_connected"), nil)
			return
		}

		var data struct {
			Type  string            `json:"type"` // ratelimit, ttl, rules
			Data  map[string]string `json:"data"`
			Clear bool              `json:"clear"` // 是否先清空再设置
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		ctx := context.Background()
		var key string
		switch data.Type {
		case "ratelimit":
			key = config.REDIS_KEY_CONFIG_RATELIMIT
		case "ttl":
			key = config.REDIS_KEY_CONFIG_TTL
		case "rules":
			key = config.REDIS_KEY_DYNAMIC_RULES
		default:
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_config_type"), nil)
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
				utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "redis_update_failed"), err), nil)
				return
			}
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "redis_config_updated"), nil)
	}
}

// HandleSubscriberWebSocket 处理订阅者 WebSocket 连接 (用于 UI 同步)
// @Summary 管理后台 WebSocket 连接
// @Description 用于 UI 实时同步状态的 WebSocket 连接
// @Tags System
// @Security BearerAuth
// @Router /ws/subscriber [get]
func HandleSubscriberWebSocket(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WS] Incoming subscriber connection from %s", r.RemoteAddr)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
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

		subscriber := &types.Subscriber{
			Conn:  conn,
			Mutex: sync.Mutex{},
			User:  user,
		}

		m.Mutex.Lock()
		if m.Subscribers == nil {
			m.Subscribers = make(map[*websocket.Conn]*types.Subscriber)
		}
		m.Subscribers[conn] = subscriber
		m.Mutex.Unlock()

		log.Printf("Subscriber WebSocket connected: %s (User: %s)", conn.RemoteAddr(), claims.Username)

		m.Mutex.RLock()
		bots := make([]types.BotClient, 0, len(m.Bots))
		for _, bot := range m.Bots {
			bots = append(bots, *bot)
		}
		workers := make([]types.WorkerInfo, 0, len(m.Workers))
		for _, w := range m.Workers {
			workers = append(workers, types.WorkerInfo{
				ID:       w.ID,
				Type:     "worker",
				Status:   "online",
				LastSeen: time.Now().Format("15:04:05"),
			})
		}
		m.Mutex.RUnlock()

		m.CacheMutex.RLock()
		syncState := types.SyncState{
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
// @Summary 获取用户列表
// @Description 从数据库获取所有系统用户的信息
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "用户列表"
// @Router /api/admin/users [get]
func HandleAdminListUsers(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		log.Printf("[DEBUG] HandleAdminListUsers called")

		var dbUsers []types.User
		if err := m.GORMDB.Find(&dbUsers).Error; err != nil {
			log.Printf("[ERROR] HandleAdminListUsers GORM query failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "db_query_failed"), err), nil)
			return
		}

		type UserInfo struct {
			ID            int64  `json:"id"`
			Username      string `json:"username"`
			IsAdmin       bool   `json:"is_admin"`
			IsSuperPoints bool   `json:"is_super_points"`
			Active        bool   `json:"active"`
			CreatedAt     string `json:"created_at"`
			UpdatedAt     string `json:"updated_at"`
		}

		users := make([]UserInfo, 0)
		for _, u := range dbUsers {
			users = append(users, UserInfo{
				ID:            u.ID,
				Username:      u.Username,
				IsAdmin:       u.IsAdmin,
				IsSuperPoints: u.IsSuperPoints,
				Active:        u.Active,
				CreatedAt:     u.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     u.UpdatedAt.Format(time.RFC3339),
			})
		}

		log.Printf("[DEBUG] HandleAdminListUsers found %d users", len(users))

		utils.SendJSONResponse(w, true, "", struct {
			Users []UserInfo `json:"users"`
		}{
			Users: users,
		})
	}
}

// HandleAdminManageUsers 用户管理操作 (create/delete/reset_pwd/toggle_status)
// @Summary 管理用户信息
// @Description 执行用户管理操作，包括创建、编辑、删除、重置密码和切换激活状态
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "管理操作参数 (action, username, password, is_admin, is_super_points)"
// @Success 200 {object} utils.JSONResponse "操作结果"
// @Router /api/admin/users [post]
func HandleAdminManageUsers(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var req struct {
			Action        string `json:"action"`
			Username      string `json:"username"`
			Password      string `json:"password"`
			IsAdmin       bool   `json:"is_admin"`
			IsSuperPoints bool   `json:"is_super_points"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		switch req.Action {
		case "create":
			handleAdminCreateUser(m, w, lang, req.Username, req.Password, req.IsAdmin, req.IsSuperPoints)
		case "edit":
			handleAdminUpdateUser(m, w, lang, req.Username, req.IsAdmin, req.IsSuperPoints)
		case "delete":
			handleAdminDeleteUser(m, w, lang, req.Username)
		case "reset_password":
			handleAdminResetPassword(m, w, lang, req.Username, req.Password)
		case "toggle_status", "toggle_active":
			handleAdminToggleUser(m, w, lang, req.Username)
		default:
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_management_invalid_action"), req.Action), nil)
		}
	}
}

func handleAdminCreateUser(m *bot.Manager, w http.ResponseWriter, lang, username, password string, isAdmin bool, isSuperPoints bool) {
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		utils.SendJSONResponse(w, false, utils.T(lang, "user_pwd_empty"), nil)
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, utils.T(lang, "password_encrypt_failed"), nil)
		return
	}

	user := &types.User{
		Username:       username,
		PasswordHash:   hash,
		IsAdmin:        isAdmin,
		IsSuperPoints:  isSuperPoints,
		Active:         true,
		SessionVersion: 1, // 初始化会话版本
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := m.SaveUserToDB(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_create_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	m.Users[username] = user
	m.UsersMutex.Unlock()

	utils.SendJSONResponse(w, true, utils.T(lang, "user_created"), nil)
}

func handleAdminUpdateUser(m *bot.Manager, w http.ResponseWriter, lang, username string, isAdmin bool, isSuperPoints bool) {
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		utils.SendJSONResponse(w, false, utils.T(lang, "username_empty"), nil)
		return
	}

	if username == "admin" && !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		utils.SendJSONResponse(w, false, utils.T(lang, "cannot_disable_default_admin"), nil)
		return
	}

	user, exists := m.GetOrLoadUser(username)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		utils.SendJSONResponse(w, false, utils.T(lang, "user_not_found"), nil)
		return
	}

	user.IsAdmin = isAdmin
	user.IsSuperPoints = isSuperPoints
	user.SessionVersion++
	user.UpdatedAt = time.Now()

	if err := m.SaveUserToDB(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_update_failed"), err), nil)
		return
	}

	utils.SendJSONResponse(w, true, utils.T(lang, "user_info_updated"), nil)
}

func handleAdminDeleteUser(m *bot.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		utils.SendJSONResponse(w, false, utils.T(lang, "cannot_delete_default_admin"), nil)
		return
	}

	if err := m.GORMDB.Where("username = ?", username).Delete(&models.UserGORM{}).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_delete_failed"), err), nil)
		return
	}

	m.UsersMutex.Lock()
	delete(m.Users, username)
	m.UsersMutex.Unlock()

	utils.SendJSONResponse(w, true, utils.T(lang, "user_deleted"), nil)
}

func handleAdminResetPassword(m *bot.Manager, w http.ResponseWriter, lang, username, password string) {
	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		utils.SendJSONResponse(w, false, utils.T(lang, "new_password_empty"), nil)
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, utils.T(lang, "password_encrypt_failed"), nil)
		return
	}

	user, exists := m.GetOrLoadUser(username)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		utils.SendJSONResponse(w, false, utils.T(lang, "user_not_found"), nil)
		return
	}

	user.PasswordHash = hash
	user.SessionVersion++
	user.UpdatedAt = time.Now()

	if err := m.SaveUserToDB(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_update_failed"), err), nil)
		return
	}

	utils.SendJSONResponse(w, true, utils.T(lang, "password_reset_success"), nil)
}

func handleAdminToggleUser(m *bot.Manager, w http.ResponseWriter, lang, username string) {
	if username == "admin" {
		w.WriteHeader(http.StatusForbidden)
		utils.SendJSONResponse(w, false, utils.T(lang, "cannot_disable_default_admin"), nil)
		return
	}

	user, exists := m.GetOrLoadUser(username)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		utils.SendJSONResponse(w, false, utils.T(lang, "user_not_found"), nil)
		return
	}

	user.Active = !user.Active
	user.SessionVersion++
	user.UpdatedAt = time.Now()

	if err := m.SaveUserToDB(user); err != nil {
		log.Printf("更新用户状态失败: %v (username: %s, active: %v)", err, username, user.Active)
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendJSONResponse(w, false, fmt.Sprintf(utils.T(lang, "user_update_failed"), err), nil)
		return
	}

	utils.SendJSONResponse(w, true, utils.T(lang, "user_status_updated"), struct {
		Active bool `json:"active"`
	}{
		Active: user.Active,
	})
}

// HandleGetRoutingRules 获取所有路由规则
// @Summary 获取路由规则
// @Description 获取所有的消息路由规则
// @Tags System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "路由规则列表"
// @Router /api/admin/routing [get]
func HandleGetRoutingRules(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()

		utils.SendJSONResponse(w, true, "", struct {
			Rules map[string]string `json:"rules"`
		}{
			Rules: m.RoutingRules,
		})
	}
}

// HandleSetRoutingRule 设置路由规则
// @Summary 设置路由规则
// @Description 创建或更新一条消息路由规则，将特定 Key 映射到 WorkerID
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "路由规则参数 (key, worker_id)"
// @Success 200 {object} utils.JSONResponse "设置结果"
// @Router /api/admin/routing [post]
func HandleSetRoutingRule(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var rule struct {
			Key      string `json:"key"`
			WorkerID string `json:"worker_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if rule.Key == "" || rule.WorkerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "routing_rule_invalid_params"), nil)
			return
		}

		m.Mutex.Lock()
		if m.RoutingRules == nil {
			m.RoutingRules = make(map[string]string)
		}
		m.RoutingRules[rule.Key] = rule.WorkerID
		m.Mutex.Unlock()

		if err := m.SaveRoutingRuleToDB(rule.Key, rule.WorkerID); err != nil {
			log.Printf(utils.T(lang, "routing_rule_save_failed"), err)
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "routing_rule_set_success"), nil)
	}
}

// HandleDeleteRoutingRule 删除路由规则
// @Summary 删除路由规则
// @Description 根据 Key 删除一条消息路由规则
// @Tags System
// @Produce json
// @Security BearerAuth
// @Param key query string true "规则 Key"
// @Success 200 {object} utils.JSONResponse "删除结果"
// @Router /api/admin/routing [delete]
func HandleDeleteRoutingRule(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "routing_rule_key_empty"), nil)
			return
		}

		m.Mutex.Lock()
		defer m.Mutex.Unlock()

		if _, exists := m.RoutingRules[key]; exists {
			delete(m.RoutingRules, key)
			if err := m.DeleteRoutingRuleFromDB(key); err != nil {
				log.Printf(utils.T(lang, "routing_rule_delete_failed"), err)
			}
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "routing_rule_delete_success"), nil)
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// HandleDockerLogs 获取 Docker 容器日志
// @Summary 获取 Docker 日志
// @Description 获取指定 Docker 容器的最近运行日志
// @Tags System
// @Produce json
// @Security BearerAuth
// @Param id query string true "容器 ID"
// @Success 200 {object} utils.JSONResponse "日志内容"
// @Router /api/docker/logs [get]
// @Router /api/admin/docker/logs [get]
func HandleDockerLogs(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		containerID := r.URL.Query().Get("id")

		if containerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.DockerClient == nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, utils.T(lang, "docker_not_init"), nil)
			return
		}

		options := docker_types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "100",
			Follow:     false,
		}

		reader, err := m.DockerClient.ContainerLogs(r.Context(), containerID, options)
		if err != nil {
			log.Printf(utils.T(lang, "get_docker_logs_failed"), err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		defer reader.Close()

		logs, _ := io.ReadAll(reader)

		utils.SendJSONResponse(w, true, "", struct {
			Logs string `json:"logs"`
		}{
			Logs: string(logs),
		})
	}
}

// HandleGetManual 获取管理员手册
// @Summary 获取管理员手册
// @Description 获取管理后台的帮助文档和操作说明
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "帮助手册内容"
// @Router /api/admin/manual [get]
func HandleGetManual(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

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
			Title: utils.T(lang, "manual_title"),
			Sections: []ManualSection{
				{
					Title:   utils.T(lang, "manual_section_quickstart_title"),
					Content: utils.T(lang, "manual_section_quickstart_content"),
				},
				{
					Title:   utils.T(lang, "manual_section_docker_title"),
					Content: utils.T(lang, "manual_section_docker_content"),
				},
				{
					Title:   utils.T(lang, "manual_section_routing_title"),
					Content: utils.T(lang, "manual_section_routing_content"),
				},
				{
					Title:   utils.T(lang, "manual_section_users_title"),
					Content: utils.T(lang, "manual_section_users_content"),
				},
			},
			Version: "1.0.0", // 使用硬编码版本号或从配置中获取
		}

		utils.SendJSONResponse(w, true, "", manual)
	}
}

// HandleListEmployees 获取数字员工列表
// @Summary 获取数字员工列表
// @Description 获取所有配置的数字员工信息
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "数字员工列表"
// @Router /api/admin/employees [get]
func HandleListEmployees(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var employees []models.DigitalEmployeeGORM
		if err := m.GORMDB.Find(&employees).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", employees)
	}
}

// HandleSaveEmployee 保存/更新数字员工信息
// @Summary 保存数字员工
// @Description 创建或更新数字员工的配置信息
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "数字员工信息"
// @Success 200 {object} utils.JSONResponse "保存结果"
// @Router /api/admin/employees [post]
func HandleSaveEmployee(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var employee models.DigitalEmployeeGORM
		if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		var err error
		if employee.ID > 0 {
			err = m.GORMDB.Save(&employee).Error
		} else {
			err = m.GORMDB.Create(&employee).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "Saved digital employee", employee)
	}
}

// HandleRecordEmployeeKpi 手动记录 KPI (如评价)
// @Summary 记录员工 KPI
// @Description 手动记录数字员工的关键绩效指标 (KPI)，如评分或特定指标
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "KPI 记录参数 (employee_id, metric, score)"
// @Success 200 {object} utils.JSONResponse "记录结果"
// @Router /api/admin/employees/kpi [post]
func HandleRecordEmployeeKpi(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EmployeeID uint    `json:"employee_id"`
			Metric     string  `json:"metric"`
			Score      float64 `json:"score"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		if m.DigitalEmployeeService == nil {
			utils.SendJSONResponse(w, false, "Employee service not initialized", nil)
			return
		}

		err := m.DigitalEmployeeService.RecordKpi(req.EmployeeID, req.Metric, req.Score)
		if err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "KPI recorded successfully", nil)
	}
}

// HandleListMemories 获取记忆列表
// @Summary 获取认知记忆列表
// @Description 获取所有数字员工的认知记忆记录，支持按 bot_id、user_id 或内容关键词过滤
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param bot_id query string false "机器人 ID"
// @Param user_id query string false "用户 ID"
// @Param q query string false "搜索关键词"
// @Success 200 {object} utils.JSONResponse "记忆列表"
// @Router /api/admin/memories [get]
func HandleListMemories(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		botID := r.URL.Query().Get("bot_id")
		userID := r.URL.Query().Get("user_id")
		query := r.URL.Query().Get("q")

		var memories []models.CognitiveMemoryGORM
		db := m.GORMDB.Model(&models.CognitiveMemoryGORM{})

		if botID != "" {
			db = db.Where("bot_id = ?", botID)
		}
		if userID != "" {
			db = db.Where("user_id = ?", userID)
		}
		if query != "" {
			db = db.Where("content LIKE ?", "%"+query+"%")
		}

		if err := db.Order("last_seen DESC").Find(&memories).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", memories)
	}
}

// HandleDeleteMemory 删除特定记忆
// @Summary 删除认知记忆
// @Description 根据 ID 删除一条特定的数字员工认知记忆
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "记忆 ID"
// @Success 200 {object} utils.JSONResponse "删除结果"
// @Router /api/admin/memories/{id} [delete]
func HandleDeleteMemory(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/memories/")
		if idStr == "" {
			utils.SendJSONResponse(w, false, "Missing memory ID", nil)
			return
		}

		if err := m.GORMDB.Delete(&models.CognitiveMemoryGORM{}, idStr).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "Memory deleted", nil)
	}
}

// HandleListB2BSkills 获取 B2B 技能共享列表
// @Summary 获取 B2B 技能列表
// @Description 获取所有已配置的 B2B 技能共享记录
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "技能列表"
// @Router /api/admin/b2b/skills [get]
func HandleListB2BSkills(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sharings []models.B2BSkillSharingGORM
		if err := m.GORMDB.Find(&sharings).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", sharings)
	}
}

// HandleSaveB2BSkill 保存/更新 B2B 技能共享
// @Summary 保存 B2B 技能
// @Description 创建或更新 B2B 技能共享配置
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "技能共享信息"
// @Success 200 {object} utils.JSONResponse "保存结果"
// @Router /api/admin/b2b/skills [post]
func HandleSaveB2BSkill(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sharing models.B2BSkillSharingGORM
		if err := json.NewDecoder(r.Body).Decode(&sharing); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		var err error
		if sharing.ID > 0 {
			err = m.GORMDB.Save(&sharing).Error
		} else {
			err = m.GORMDB.Create(&sharing).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "Saved B2B skill sharing", sharing)
	}
}

// HandleDeleteB2BSkill 删除 B2B 技能共享
// @Summary 删除 B2B 技能
// @Description 根据 ID 删除一条 B2B 技能共享记录
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "技能 ID"
// @Success 200 {object} utils.JSONResponse "删除结果"
// @Router /api/admin/b2b/skills/{id} [delete]
func HandleDeleteB2BSkill(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/b2b/skills/")
		if idStr == "" {
			utils.SendJSONResponse(w, false, "Missing sharing ID", nil)
			return
		}

		if err := m.GORMDB.Delete(&models.B2BSkillSharingGORM{}, idStr).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "B2B skill sharing deleted", nil)
	}
}

// HandleListB2BConnections 获取 B2B 连接列表
// @Summary 获取 B2B 连接列表
// @Description 获取所有已建立的企业间数字员工 (B2B) 连接记录
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "连接列表"
// @Router /api/admin/b2b/connections [get]
func HandleListB2BConnections(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var connections []models.B2BConnectionGORM
		if err := m.GORMDB.Find(&connections).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", connections)
	}
}

// HandleUpdateEmployeeStatus 手动更新数字员工状态/预算
// @Summary 更新数字员工状态
// @Description 手动更新数字员工的在线状态或薪资限制
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "状态参数 (bot_id, status, salary_limit)"
// @Success 200 {object} utils.JSONResponse "更新结果"
// @Router /api/admin/employees/status [post]
func HandleUpdateEmployeeStatus(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BotID       string `json:"bot_id"`
			Status      string `json:"status"`
			SalaryLimit *int64 `json:"salary_limit"`
			SalaryToken *int64 `json:"salary_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		if m.DigitalEmployeeService == nil {
			utils.SendJSONResponse(w, false, "Employee service not initialized", nil)
			return
		}

		if req.Status != "" {
			if err := m.DigitalEmployeeService.UpdateOnlineStatus(req.BotID, req.Status); err != nil {
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		}

		if req.SalaryLimit != nil || req.SalaryToken != nil {
			if err := m.DigitalEmployeeService.UpdateSalary(req.BotID, req.SalaryToken, req.SalaryLimit); err != nil {
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		}

		utils.SendJSONResponse(w, true, "Employee status/salary updated", nil)
	}
}
