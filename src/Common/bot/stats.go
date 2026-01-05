package bot

import (
	"time"

	"BotMatrix/common/types"
)

// ResetTodayStats resets today's statistics
func ResetTodayStats(userStatsToday, groupStatsToday, botStatsToday map[string]int64) (map[string]int64, map[string]int64, map[string]int64) {
	return make(map[string]int64), make(map[string]int64), make(map[string]int64)
}

// UpdateBotStats updates bot statistics in the provided maps
func UpdateBotStats(
	botID, userID, groupID string,
	botDetailedStats map[string]*types.BotStatDetail,
	userStats, groupStats, botStats map[string]int64,
	userStatsToday, groupStatsToday, botStatsToday map[string]int64,
) {
	if _, exists := botDetailedStats[botID]; !exists {
		botDetailedStats[botID] = &types.BotStatDetail{
			Users:  make(map[string]int64),
			Groups: make(map[string]int64),
		}
	}

	stats := botDetailedStats[botID]
	stats.Received++
	stats.LastMsg = time.Now()

	if userID != "" && userID != "0" {
		stats.Users[userID]++
		userStats[userID]++
		userStatsToday[userID]++
	}
	if groupID != "" && groupID != "0" {
		stats.Groups[groupID]++
		groupStats[groupID]++
		groupStatsToday[groupID]++
	}
	botStats[botID]++
	botStatsToday[botID]++
}

// SaveAllStatsToDB saves all statistics to the database
func (m *Manager) SaveAllStatsToDB() {
	// For now, this is a stub as there is no clear database schema for generic stats
	// In a real scenario, this would iterate through m.UserStats, m.GroupStats, etc.
	// and save them to the database.
	m.LogDebug("[STATS] Saving all stats to DB (stub)")
}

// StartPeriodicStatsSave starts periodic stats saving
func (m *Manager) StartPeriodicStatsSave() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	midnightTicker := time.NewTicker(1 * time.Hour)
	defer midnightTicker.Stop()

	for {
		select {
		case <-ticker.C:
			m.SaveAllStatsToDB()
		case <-midnightTicker.C:
			now := time.Now()
			currentDate := now.Format("2006-01-02")

			m.StatsMutex.Lock()
			if m.LastResetDate != currentDate {
				m.LogInfo("[STATS] Date change detected, performing daily reset: %s -> %s", m.LastResetDate, currentDate)
				m.UserStatsToday, m.GroupStatsToday, m.BotStatsToday = ResetTodayStats(m.UserStatsToday, m.GroupStatsToday, m.BotStatsToday)
				m.LastResetDate = currentDate
				m.StatsMutex.Unlock()
				m.SaveAllStatsToDB()
			} else {
				m.StatsMutex.Unlock()
			}
		}
	}
}

// UpdateBotStats updates bot statistics
func (m *Manager) UpdateBotStats(botID string, userID, groupID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.StatsMutex.Lock()
	defer m.StatsMutex.Unlock()

	// Initialize statistics data structure (if needed)
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*types.BotStatDetail)
	}

	UpdateBotStats(
		botID, userID, groupID,
		m.BotDetailedStats,
		m.UserStats, m.GroupStats, m.BotStats,
		m.UserStatsToday, m.GroupStatsToday, m.BotStatsToday,
	)
}
