package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"BotMatrix/common/bot"
	"BotMatrix/common/models"
	"BotMatrix/common/utils"
	"strings"
)

// HandleGetFissionConfig 获取裂变配置
func HandleGetFissionConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config models.FissionConfigGORM
		if m.GORMDB != nil {
			m.GORMDB.First(&config)
		}

		utils.SendJSONResponse(w, true, "", struct {
			Config models.FissionConfigGORM `json:"config"`
		}{
			Config: config,
		})
	}
}

// HandleUpdateFissionConfig 更新裂变配置
func HandleUpdateFissionConfig(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var config models.FissionConfigGORM
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.GORMDB != nil {
			config.ID = 1 // 强制使用 ID 1
			config.UpdatedAt = time.Now()
			if err := m.GORMDB.Save(&config).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "action_success"), nil)
	}
}

// HandleGetFissionTasks 获取裂变任务列表
func HandleGetFissionTasks(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tasks []models.FissionTaskGORM
		if m.GORMDB != nil {
			m.GORMDB.Find(&tasks)
		}

		utils.SendJSONResponse(w, true, "", struct {
			Tasks []models.FissionTaskGORM `json:"tasks"`
		}{
			Tasks: tasks,
		})
	}
}

// HandleSaveFissionTask 保存或更新裂变任务
func HandleSaveFissionTask(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var task models.FissionTaskGORM
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request_format"), nil)
			return
		}

		if m.GORMDB != nil {
			task.UpdatedAt = time.Now()
			if task.ID == 0 {
				task.CreatedAt = time.Now()
			}
			if err := m.GORMDB.Save(&task).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "action_success"), nil)
	}
}

// HandleDeleteFissionTask 删除裂变任务
func HandleDeleteFissionTask(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		// 从 URL 路径中提取 ID
		pathParts := strings.Split(r.URL.Path, "/")
		idStr := pathParts[len(pathParts)-1]
		id, _ := strconv.Atoi(idStr)

		if m.GORMDB != nil {
			if err := m.GORMDB.Delete(&models.FissionTaskGORM{}, id).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				utils.SendJSONResponse(w, false, err.Error(), nil)
				return
			}
		}

		utils.SendJSONResponse(w, true, utils.T(lang, "action_success"), nil)
	}
}

// HandleGetFissionStats 获取裂变统计数据
func HandleGetFissionStats(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type DailyTrend struct {
			Date  string `json:"date"`
			Count int64  `json:"count"`
		}

		type FissionStats struct {
			TotalInvitations int64        `json:"total_invitations"`
			TotalPoints      float64      `json:"total_points"`
			ActiveInviters   int64        `json:"active_inviters"`
			TodayInvitations int64        `json:"today_invitations"`
			DailyTrends      []DailyTrend `json:"daily_trends"`
		}

		stats := FissionStats{
			DailyTrends: []DailyTrend{},
		}

		if m.GORMDB != nil {
			var totalInvitations int64
			m.GORMDB.Model(&models.InvitationGORM{}).Count(&totalInvitations)
			stats.TotalInvitations = totalInvitations

			var totalPoints float64
			m.GORMDB.Model(&models.UserFissionRecordGORM{}).Select("sum(points)").Row().Scan(&totalPoints)
			stats.TotalPoints = totalPoints

			var activeInviters int64
			m.GORMDB.Model(&models.UserFissionRecordGORM{}).Where("invite_count > 0").Count(&activeInviters)
			stats.ActiveInviters = activeInviters

			var todayInvitations int64
			today := time.Now().Truncate(24 * time.Hour)
			m.GORMDB.Model(&models.InvitationGORM{}).Where("created_at >= ?", today).Count(&todayInvitations)
			stats.TodayInvitations = todayInvitations

			// 获取最近 7 天的趋势
			for i := 6; i >= 0; i-- {
				date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
				startTime := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -i)
				endTime := startTime.AddDate(0, 0, 1)

				var count int64
				m.GORMDB.Model(&models.InvitationGORM{}).Where("created_at >= ? AND created_at < ?", startTime, endTime).Count(&count)

				stats.DailyTrends = append(stats.DailyTrends, DailyTrend{
					Date:  date,
					Count: count,
				})
			}
		}

		utils.SendJSONResponse(w, true, "", struct {
			Stats FissionStats `json:"stats"`
		}{
			Stats: stats,
		})
	}
}

// HandleGetInvitations 获取邀请记录列表
func HandleGetInvitations(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var invitations []models.InvitationGORM
		if m.GORMDB != nil {
			m.GORMDB.Order("created_at desc").Limit(100).Find(&invitations)
		}

		utils.SendJSONResponse(w, true, "", struct {
			Invitations []models.InvitationGORM `json:"invitations"`
		}{
			Invitations: invitations,
		})
	}
}

// HandleGetFissionLeaderboard 获取裂变排行榜 (后台用)
func HandleGetFissionLeaderboard(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rank []models.UserFissionRecordGORM
		if m.GORMDB != nil {
			m.GORMDB.Order("invite_count desc, points desc").Limit(20).Find(&rank)
		}

		utils.SendJSONResponse(w, true, "", struct {
			Rank []models.UserFissionRecordGORM `json:"rank"`
		}{
			Rank: rank,
		})
	}
}
