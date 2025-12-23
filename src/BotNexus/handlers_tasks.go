package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"BotNexus/tasks"
	"BotMatrix/common"
)

// HandleListTasks 获取任务列表
func HandleListTasks(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var taskList []tasks.Task
		m.DB.Preload("Tags").Find(&taskList)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    taskList,
		})
	}
}

// HandleCreateTask 创建任务
func HandleCreateTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lang := common.GetLangFromRequest(r)
		
		var task tasks.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		// 简单检查是否为企业版 (这里可以根据用户权限或配置决定)
		isEnterprise := claims.IsAdmin 

		if err := m.TaskManager.CreateTask(&task, isEnterprise); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": common.T(lang, "create_task_failed"),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    task,
		})
	}
}

// HandleGetExecutions 获取执行记录
func HandleGetExecutions(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		taskIDStr := r.URL.Query().Get("task_id")
		taskID, _ := strconv.Atoi(taskIDStr)

		history, err := m.TaskManager.GetExecutionHistory(uint(taskID), 20)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    history,
		})
	}
}

// HandleAIParse AI 解析
func HandleAIParse(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req tasks.ParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := m.TaskManager.AI.Parse(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// 将解析结果存入草稿箱
		draftID := common.GenerateUUID()
		dataJSON, _ := json.Marshal(result.Data)
		claims := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		
		draft := tasks.AIDraft{
			DraftID:    draftID,
			UserID:     uint(claims.ID),
			Intent:     string(result.Intent),
			Data:       string(dataJSON),
			ExpireTime: time.Now().Add(15 * time.Minute), // 15分钟有效
		}
		m.DB.Create(&draft)

		result.DraftID = draftID

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    result,
		})
	}
}

// HandleAIConfirm AI 确认执行
func HandleAIConfirm(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var draft tasks.AIDraft
		if err := m.DB.Where("draft_id = ? AND status = 'pending'", req.DraftID).First(&draft).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "草稿不存在或已失效",
			})
			return
		}

		if time.Now().After(draft.ExpireTime) {
			m.DB.Model(&draft).Update("status", "expired")
			w.WriteHeader(http.StatusGone)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "草稿已过期",
			})
			return
		}

		// 根据意图执行具体动作
		var err error
		switch draft.Intent {
		case string(tasks.AIActionCreateTask):
			var task tasks.Task
			json.Unmarshal([]byte(draft.Data), &task)
			task.CreatorID = draft.UserID
			err = m.TaskManager.CreateTask(&task, true) // 默认 AI 创建的按企业版权限处理或根据用户权限
		case string(tasks.AIActionAdjustPolicy):
			// 策略调整逻辑...
		case string(tasks.AIActionManageTags):
			// 标签管理逻辑...
		case "skill_call":
			var skillReq struct {
				Skill  string            `json:"skill"`
				Params map[string]string `json:"params"`
			}
			json.Unmarshal([]byte(draft.Data), &skillReq)
			worker := m.FindWorkerBySkill(skillReq.Skill)
			if worker == nil {
				err = fmt.Errorf("未找到具备该能力的 Worker: %s", skillReq.Skill)
			} else {
				// 构造指令发送给 Worker
				cmd := map[string]interface{}{
					"type":   "skill_call",
					"skill":  skillReq.Skill,
					"params": skillReq.Params,
					"user_id": draft.UserID,
				}
				worker.Mutex.Lock()
				err = worker.Conn.WriteJSON(cmd)
				worker.Mutex.Unlock()
			}
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		m.DB.Model(&draft).Update("status", "confirmed")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "执行成功",
		})
	}
}

// HandleManageTags 标签管理
func HandleManageTags(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Action     string `json:"action"` // add, remove
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
			TagName    string `json:"tag_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var err error
		if req.Action == "add" {
			err = m.TaskManager.Tagging.AddTag(req.TargetType, req.TargetID, req.TagName)
		} else {
			err = m.TaskManager.Tagging.RemoveTag(req.TargetType, req.TargetID, req.TagName)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// HandleGetCapabilities 获取系统能力清单 (用于 AI 提示或功能展示)
func HandleGetCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		manifest := m.TaskManager.AI.Manifest
		prompt := m.TaskManager.AI.GetSystemPrompt()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"manifest": manifest,
				"prompt":   prompt,
			},
		})
	}
}
