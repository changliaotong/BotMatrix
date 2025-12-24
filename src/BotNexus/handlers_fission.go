package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"BotMatrix/common"
	"strings"
)

// HandleGetFissionConfig 获取裂变配置
func HandleGetFissionConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var config common.FissionConfigGORM
		if m.GORMDB != nil {
			m.GORMDB.First(&config)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"config":  config,
		})
	}
}

// HandleUpdateFissionConfig 更新裂变配置
func HandleUpdateFissionConfig(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var config common.FissionConfigGORM
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		if m.GORMDB != nil {
			config.ID = 1 // 强制使用 ID 1
			config.UpdatedAt = time.Now()
			if err := m.GORMDB.Save(&config).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "action_success"),
		})
	}
}

// HandleGetFissionTasks 获取裂变任务列表
func HandleGetFissionTasks(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var tasks []common.FissionTaskGORM
		if m.GORMDB != nil {
			m.GORMDB.Find(&tasks)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"tasks":   tasks,
		})
	}
}

// HandleSaveFissionTask 保存或更新裂变任务
func HandleSaveFissionTask(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		var task common.FissionTaskGORM
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "invalid_request_format"),
			})
			return
		}

		if m.GORMDB != nil {
			task.UpdatedAt = time.Now()
			if task.ID == 0 {
				task.CreatedAt = time.Now()
			}
			if err := m.GORMDB.Save(&task).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "action_success"),
		})
	}
}

// HandleDeleteFissionTask 删除裂变任务
func HandleDeleteFissionTask(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)

		// 从 URL 路径中提取 ID
		pathParts := strings.Split(r.URL.Path, "/")
		idStr := pathParts[len(pathParts)-1]
		id, _ := strconv.Atoi(idStr)

		if m.GORMDB != nil {
			if err := m.GORMDB.Delete(&common.FissionTaskGORM{}, id).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": common.T(lang, "action_success"),
		})
	}
}

// HandleGetFissionStats 获取裂变统计数据
func HandleGetFissionStats(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stats := map[string]interface{}{
			"total_invitations": 0,
			"total_points":      0,
			"active_inviters":   0,
			"today_invitations": 0,
			"daily_trends":      []map[string]interface{}{},
		}

		if m.GORMDB != nil {
			var totalInvitations int64
			m.GORMDB.Model(&common.InvitationGORM{}).Count(&totalInvitations)
			stats["total_invitations"] = totalInvitations

			var totalPoints float64
			m.GORMDB.Model(&common.UserFissionRecordGORM{}).Select("sum(points)").Row().Scan(&totalPoints)
			stats["total_points"] = totalPoints

			var activeInviters int64
			m.GORMDB.Model(&common.UserFissionRecordGORM{}).Where("invite_count > 0").Count(&activeInviters)
			stats["active_inviters"] = activeInviters

			var todayInvitations int64
			today := time.Now().Truncate(24 * time.Hour)
			m.GORMDB.Model(&common.InvitationGORM{}).Where("created_at >= ?", today).Count(&todayInvitations)
			stats["today_invitations"] = todayInvitations

			// 获取最近 7 天的趋势
			var trends []map[string]interface{}
			for i := 6; i >= 0; i-- {
				date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
				startTime := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -i)
				endTime := startTime.AddDate(0, 0, 1)

				var count int64
				m.GORMDB.Model(&common.InvitationGORM{}).Where("created_at >= ? AND created_at < ?", startTime, endTime).Count(&count)

				trends = append(trends, map[string]interface{}{
					"date":  date,
					"count": count,
				})
			}
			stats["daily_trends"] = trends
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"stats":   stats,
		})
	}
}

// HandleGetInvitations 获取邀请记录列表
func HandleGetInvitations(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var invitations []common.InvitationGORM
		if m.GORMDB != nil {
			m.GORMDB.Order("created_at desc").Limit(100).Find(&invitations)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"invitations": invitations,
		})
	}
}

// HandleGetFissionLeaderboard 获取裂变排行榜 (后台用)
func HandleGetFissionLeaderboard(m *common.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var rank []common.UserFissionRecordGORM
		if m.GORMDB != nil {
			m.GORMDB.Order("invite_count desc, points desc").Limit(20).Find(&rank)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"rank":    rank,
		})
	}
}
