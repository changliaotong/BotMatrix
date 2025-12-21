package main

import (
	"time"
)

// 启动Bot超时检测
func (m *Manager) StartBotTimeoutDetection() {
	m.LogInfo("[Bot Monitor] Starting timeout detection (60s interval)")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.checkBotTimeouts()
	}
}

// 检查Bot心跳超时
func (m *Manager) checkBotTimeouts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	activeBots := make(map[string]*BotClient)

	if len(m.bots) > 0 {
		// m.LogDebug("[Bot Monitor] Checking timeouts for %d bots", len(m.bots))
	}

	for botID, bot := range m.bots {
		bot.Mutex.Lock()
		lastActive := bot.LastHeartbeat
		if lastActive.IsZero() {
			lastActive = bot.Connected
		}

		timeoutDuration := now.Sub(lastActive)
		/*
			m.LogDebug("[Bot Monitor] Bot %s - Last active: %v, Timeout: %v",
				botID, lastActive.Format("15:04:05"), timeoutDuration)
		*/

		if timeoutDuration < 2*time.Minute {
			activeBots[botID] = bot
			// m.LogDebug("[Bot Monitor] Bot %s is active", botID)
		} else {
			// 超时Bot，关闭连接
			m.LogWarn("[Bot Monitor] Bot %s heartbeat timeout after %v, closing connection",
				botID, timeoutDuration)
			bot.Conn.Close()

			// 记录断开连接
			m.TrackBotDisconnection(botID, "heartbeat_timeout", timeoutDuration)
		}
		bot.Mutex.Unlock()
	}

	removedCount := len(m.bots) - len(activeBots)
	if removedCount > 0 {
		m.LogInfo("[Bot Monitor] Removed %d timeout bots, remaining: %d", removedCount, len(activeBots))
	}

	m.bots = activeBots
}

// 检查特定Bot状态（用于调试）
func (m *Manager) LogBotStatus(botID string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if bot, exists := m.bots[botID]; exists {
		bot.Mutex.Lock()
		now := time.Now()
		lastActive := bot.LastHeartbeat
		if lastActive.IsZero() {
			lastActive = bot.Connected
		}

		m.LogInfo("[Bot Status] ID: %s, Connected: %v, LastActive: %v, Timeout: %v",
			botID,
			bot.Connected.Format("15:04:05"),
			lastActive.Format("15:04:05"),
			now.Sub(lastActive))
		bot.Mutex.Unlock()
	} else {
		m.LogWarn("[Bot Status] Bot %s not found", botID)
	}
}

// 检查所有Bot状态（用于调试）
func (m *Manager) LogAllBotStatus() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	m.LogInfo("[Bot Status] Total bots: %d", len(m.bots))
	for botID, bot := range m.bots {
		bot.Mutex.Lock()
		now := time.Now()
		lastActive := bot.LastHeartbeat
		if lastActive.IsZero() {
			lastActive = bot.Connected
		}

		m.LogInfo("[Bot Status] ID: %s, Timeout: %v",
			botID, now.Sub(lastActive))
		bot.Mutex.Unlock()
	}
}
