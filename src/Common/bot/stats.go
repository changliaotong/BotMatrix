package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"BotMatrix/common/models"
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
	ctx := context.Background()
	if m.Rdb == nil || m.GORMDB == nil {
		return
	}

	// 每天凌晨执行：将昨天的 Redis 统计数据固化到数据库
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("2006-01-02")

	// 1. 扫描昨天的所有群组排名 Key
	iter := m.Rdb.Scan(ctx, 0, fmt.Sprintf("stats:rank:%s:group:*", dateStr), 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		groupID := strings.TrimPrefix(key, fmt.Sprintf("stats:rank:%s:group:", dateStr))

		// 获取该群组昨天的所有用户发言数据
		results, err := m.Rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}

		for _, z := range results {
			userID := z.Member.(string)
			count := int64(z.Score)

			// 写入数据库
			stat := &models.MessageStat{
				Date:    yesterday,
				GroupID: groupID,
				UserID:  userID,
				Count:   count,
			}

			// 使用 Upsert 逻辑，防止重复运行导致数据翻倍
			m.GORMDB.Where(models.MessageStat{
				Date:    yesterday,
				GroupID: groupID,
				UserID:  userID,
			}).Assign(models.MessageStat{Count: count}).FirstOrCreate(stat)
		}
	}

	m.LogInfo("[STATS] Yesterday's stats persisted to DB for date: %s", dateStr)
}

// StartPeriodicStatsSave starts a goroutine to periodically save stats
func (m *Manager) StartPeriodicStatsSave() {
	// 每天凌晨 1:00 执行一次固化逻辑
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-m.Ctx.Done():
				return
			case t := <-ticker.C:
				// 如果是凌晨 1 点左右，执行昨天的统计固化
				if t.Hour() == 1 {
					m.SaveAllStatsToDB()
				}
			}
		}
	}()
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
