package main

import (
	"time"
)

// TrackWorkerConnection 记录Worker连接
func (m *Manager) TrackWorkerConnection(workerID string) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.TotalWorkerConnections++
	m.connectionStats.LastWorkerActivity[workerID] = time.Now()

	m.LogInfo("[Stats] Worker %s connected (total: %d)", workerID, m.connectionStats.TotalWorkerConnections)
}

// TrackWorkerDisconnection 记录Worker断开
func (m *Manager) TrackWorkerDisconnection(workerID string, reason string, duration time.Duration) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.WorkerConnectionDurations[workerID] = duration
	m.connectionStats.WorkerDisconnectReasons[reason]++

	m.LogInfo("[Stats] Worker %s disconnected: reason=%s, duration=%v", workerID, reason, duration)
}

// TrackBotConnection 记录Bot连接
func (m *Manager) TrackBotConnection(botID string) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.TotalBotConnections++
	m.connectionStats.LastBotActivity[botID] = time.Now()

	m.LogInfo("[Stats] Bot %s connected (total: %d)", botID, m.connectionStats.TotalBotConnections)
}

// TrackBotDisconnection 记录Bot断开
func (m *Manager) TrackBotDisconnection(botID string, reason string, duration time.Duration) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.BotConnectionDurations[botID] = duration
	m.connectionStats.BotDisconnectReasons[reason]++

	m.LogInfo("[Stats] Bot %s disconnected: reason=%s, duration=%v", botID, reason, duration)
}

// GetConnectionStats 获取连接统计（线程安全）
func (m *Manager) GetConnectionStats() map[string]interface{} {
	m.connectionStats.Mutex.RLock()
	defer m.connectionStats.Mutex.RUnlock()

	// 复制数据避免锁竞争
	botDurations := make(map[string]string)
	for k, v := range m.connectionStats.BotConnectionDurations {
		botDurations[k] = v.String()
	}

	workerDurations := make(map[string]string)
	for k, v := range m.connectionStats.WorkerConnectionDurations {
		workerDurations[k] = v.String()
	}

	stats := map[string]interface{}{
		"total_bot_connections":       m.connectionStats.TotalBotConnections,
		"total_worker_connections":    m.connectionStats.TotalWorkerConnections,
		"bot_connection_durations":    botDurations,
		"worker_connection_durations": workerDurations,
		"bot_disconnect_reasons":      m.connectionStats.BotDisconnectReasons,
		"worker_disconnect_reasons":   m.connectionStats.WorkerDisconnectReasons,
		"last_bot_activity":           m.connectionStats.LastBotActivity,
		"last_worker_activity":        m.connectionStats.LastWorkerActivity,
	}

	m.LogDebug("[Stats] Retrieved connection stats: %d bots, %d workers",
		m.connectionStats.TotalBotConnections, m.connectionStats.TotalWorkerConnections)

	return stats
}

// GetStatsSummary 获取统计摘要
func (m *Manager) GetStatsSummary() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	m.connectionStats.Mutex.RLock()
	defer m.connectionStats.Mutex.RUnlock()

	summary := map[string]interface{}{
		"active_bots":              len(m.bots),
		"active_workers":           len(m.workers),
		"total_bot_connections":    m.connectionStats.TotalBotConnections,
		"total_worker_connections": m.connectionStats.TotalWorkerConnections,
		"timestamp":                time.Now().Format("2006-01-02 15:04:05"),
	}

	m.LogDebug("[Stats] Retrieved stats summary: %v", summary)

	return summary
}

// StartStatsResetTimer 启动统计信息重置定时器
func (m *Manager) StartStatsResetTimer() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.resetDailyStats()
	}
}

// resetDailyStats 重置每日统计
func (m *Manager) resetDailyStats() {
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if m.LastResetDate != currentDate {
		m.LogInfo("[Stats] 重置每日统计信息: %s -> %s", m.LastResetDate, currentDate)
		m.UserStatsToday = make(map[string]int64)
		m.GroupStatsToday = make(map[string]int64)
		m.BotStatsToday = make(map[string]int64)
		m.LastResetDate = currentDate
	}
}
