package common

import (
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

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
			m.SaveAllStatsToDB()
		case <-midnightTicker.C:
			// 检查是否跨天 (更稳健的检查方法)
			now := time.Now()
			currentDate := now.Format("2006-01-02")

			m.StatsMutex.Lock()
			if m.LastResetDate != currentDate {
				log.Printf("[STATS] 检测到跨天，执行每日统计重置: %s -> %s", m.LastResetDate, currentDate)
				m.UserStatsToday = make(map[string]int64)
				m.GroupStatsToday = make(map[string]int64)
				m.BotStatsToday = make(map[string]int64)
				m.LastResetDate = currentDate

				// 重置后立即保存一次数据库，清空数据库中的今日统计
				m.StatsMutex.Unlock()
				m.SaveAllStatsToDB()
			} else {
				m.StatsMutex.Unlock()
			}
		}
	}
}

// UpdateBotStats 更新Bot统计信息
func (m *Manager) UpdateBotStats(botID string, userID, groupID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.StatsMutex.Lock()
	defer m.StatsMutex.Unlock()

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
	go m.SaveStatToDB("total_messages", m.TotalMessages)
}

// UpdateBotSentStats 更新发送消息统计
func (m *Manager) UpdateBotSentStats(botID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.StatsMutex.Lock()
	defer m.StatsMutex.Unlock()

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
	go m.SaveStatToDB("total_messages", m.TotalMessages)
	go m.SaveStatToDB("sent_messages", m.SentMessages)
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
		m.StatsMutex.RLock()
		total := m.TotalMessages
		sent := m.SentMessages
		m.StatsMutex.RUnlock()

		// 计算增量 (用于趋势图)
		var deltaTotal, deltaSent int64
		if m.LastTrendTotal > 0 {
			deltaTotal = total - m.LastTrendTotal
			if deltaTotal < 0 {
				deltaTotal = 0
			}
		}
		if m.LastTrendSent > 0 {
			deltaSent = sent - m.LastTrendSent
			if deltaSent < 0 {
				deltaSent = 0
			}
		}

		// 如果是第一次收集，或者 total 为 0，则 delta 为 0
		if m.LastTrendTotal == 0 {
			deltaTotal = 0
		}
		if m.LastTrendSent == 0 {
			deltaSent = 0
		}

		m.LastTrendTotal = total
		m.LastTrendSent = sent

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
			if cached, ok := m.ProcMap[p.Pid]; ok {
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
		m.ProcMap = newProcMap

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

// TrackWorkerConnection 记录Worker连接
func (m *Manager) TrackWorkerConnection(workerID string) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.TotalWorkerConnections++
	m.ConnectionStats.LastWorkerActivity[workerID] = time.Now()

	m.LogInfo("[Stats] Worker %s connected (total: %d)", workerID, m.ConnectionStats.TotalWorkerConnections)
}

// TrackWorkerDisconnection 记录Worker断开
func (m *Manager) TrackWorkerDisconnection(workerID string, reason string, duration time.Duration) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.WorkerConnectionDurations[workerID] = duration
	m.ConnectionStats.WorkerDisconnectReasons[reason]++

	m.LogInfo("[Stats] Worker %s disconnected: reason=%s, duration=%v", workerID, reason, duration)
}

// TrackBotConnection 记录Bot连接
func (m *Manager) TrackBotConnection(botID string) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.TotalBotConnections++
	m.ConnectionStats.LastBotActivity[botID] = time.Now()

	m.LogInfo("[Stats] Bot %s connected (total: %d)", botID, m.ConnectionStats.TotalBotConnections)
}

// TrackBotDisconnection 记录Bot断开
func (m *Manager) TrackBotDisconnection(botID string, reason string, duration time.Duration) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.BotConnectionDurations[botID] = duration
	m.ConnectionStats.BotDisconnectReasons[reason]++

	m.LogInfo("[Stats] Bot %s disconnected: reason=%s, duration=%v", botID, reason, duration)
}

// GetConnectionStats 获取连接统计（线程安全）
func (m *Manager) GetConnectionStats() map[string]any {
	m.ConnectionStats.Mutex.RLock()
	defer m.ConnectionStats.Mutex.RUnlock()

	// 复制数据避免锁竞争
	botDurations := make(map[string]string)
	for k, v := range m.ConnectionStats.BotConnectionDurations {
		botDurations[k] = v.String()
	}

	workerDurations := make(map[string]string)
	for k, v := range m.ConnectionStats.WorkerConnectionDurations {
		workerDurations[k] = v.String()
	}

	stats := map[string]any{
		"total_bot_connections":       m.ConnectionStats.TotalBotConnections,
		"total_worker_connections":    m.ConnectionStats.TotalWorkerConnections,
		"bot_connection_durations":    botDurations,
		"worker_connection_durations": workerDurations,
		"bot_disconnect_reasons":      m.ConnectionStats.BotDisconnectReasons,
		"worker_disconnect_reasons":   m.ConnectionStats.WorkerDisconnectReasons,
		"last_bot_activity":           m.ConnectionStats.LastBotActivity,
		"last_worker_activity":        m.ConnectionStats.LastWorkerActivity,
	}

	return stats
}

// GetStatsSummary 获取统计摘要
func (m *Manager) GetStatsSummary() map[string]any {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	m.ConnectionStats.Mutex.RLock()
	defer m.ConnectionStats.Mutex.RUnlock()

	summary := map[string]any{
		"active_bots":              len(m.Bots),
		"active_workers":           len(m.Workers),
		"total_bot_connections":    m.ConnectionStats.TotalBotConnections,
		"total_worker_connections": m.ConnectionStats.TotalWorkerConnections,
		"timestamp":                time.Now().Format("2006-01-02 15:04:05"),
	}

	return summary
}
