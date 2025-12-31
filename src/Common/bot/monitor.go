package bot

import (
	"time"

	"BotMatrix/common/types"
)

// StartBotTimeoutDetection starts bot timeout detection
func (m *Manager) StartBotTimeoutDetection() {
	m.LogInfo("[Bot Monitor] Starting timeout detection (60s interval)")
	for range NewTicker(60) {
		m.checkBotTimeouts()
	}
}

func (m *Manager) checkBotTimeouts() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.Bots = CheckBotTimeouts(
		m.Bots,
		m.LogWarn,
		m.LogInfo,
		m.TrackBotDisconnection,
	)
}

// CheckBotTimeouts checks for bot heartbeat timeouts
func CheckBotTimeouts(
	bots map[string]*types.BotClient,
	logWarn func(string, ...any),
	logInfo func(string, ...any),
	trackBotDisconnection func(string, string, time.Duration),
) map[string]*types.BotClient {
	now := time.Now()
	activeBots := make(map[string]*types.BotClient)

	for botID, bot := range bots {
		bot.Mutex.Lock()
		lastActive := bot.LastHeartbeat
		if lastActive.IsZero() {
			lastActive = bot.Connected
		}

		timeoutDuration := now.Sub(lastActive)

		if timeoutDuration < 2*time.Minute {
			activeBots[botID] = bot
		} else {
			logWarn("[Bot Monitor] Bot %s heartbeat timeout after %v, closing connection", botID, timeoutDuration)
			if bot.Conn != nil {
				bot.Conn.Close()
			}
			trackBotDisconnection(botID, "heartbeat_timeout", timeoutDuration)
		}
		bot.Mutex.Unlock()
	}

	removedCount := len(bots) - len(activeBots)
	if removedCount > 0 {
		logInfo("[Bot Monitor] Removed %d timeout bots, remaining: %d", removedCount, len(activeBots))
	}

	return activeBots
}
