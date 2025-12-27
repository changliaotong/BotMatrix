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
		m.GORMDB.Preload("Tags").Find(&taskList)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    taskList,
		})
	}
}

// HandleListSystemCapabilities 获取系统任务处理能力 (Actions & Interceptors)
func HandleListSystemCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		actions := m.TaskManager.Dispatcher.GetActions()
		interceptors := m.TaskManager.Interceptors.GetInterceptors()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"actions":      actions,
				"interceptors": interceptors,
			},
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
		draftID := common.GenerateRandomToken(8)
		dataJSON, _ := json.Marshal(result.Data)
		claims := r.Context().Value(common.UserClaimsKey).(*common.UserClaims)
		
		draft := tasks.AIDraft{
			DraftID:    draftID,
			UserID:     uint(claims.UserID),
			Intent:     string(result.Intent),
			Data:       string(dataJSON),
			ExpireTime: time.Now().Add(15 * time.Minute), // 15分钟有效
		}
		m.GORMDB.Create(&draft)

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
		if err := m.GORMDB.Where("draft_id = ? AND status = 'pending'", req.DraftID).First(&draft).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "草稿不存在或已失效",
			})
			return
		}

		if time.Now().After(draft.ExpireTime) {
			m.GORMDB.Model(&draft).Update("status", "expired")
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
			workerID := m.FindWorkerBySkill(skillReq.Skill)
			if workerID == "" {
				err = fmt.Errorf("未找到具备该能力的 Worker: %s", skillReq.Skill)
			} else {
				// 构造指令发送给 Worker
				cmd := map[string]interface{}{
					"type":   "skill_call",
					"skill":  skillReq.Skill,
					"params": skillReq.Params,
					"user_id": draft.UserID,
				}
				err = m.SendToWorker(workerID, cmd)
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

		m.GORMDB.Model(&draft).Update("status", "confirmed")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "执行成功",
		})
	}
}

// HandleTranslate 翻译接口 (目前使用 Azure 服务)
func HandleTranslate(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Text       string `json:"text"`
			TargetLang string `json:"target_lang"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if common.AZURE_TRANSLATE_KEY == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Azure Translate API key not configured",
			})
			return
		}

		// 构建 Azure 翻译请求
		endpoint := common.AZURE_TRANSLATE_END
		if endpoint == "" {
			endpoint = "https://api.cognitive.microsofttranslator.com/translate"
		}

		apiURL := fmt.Sprintf("%s?api-version=3.0&to=%s", endpoint, req.TargetLang)
		
		// Azure 目标语言映射 (简单的映射，实际可能更复杂)
		targetLang := req.TargetLang
		switch targetLang {
		case "zh-CN":
			targetLang = "zh-Hans"
		case "zh-TW":
			targetLang = "zh-Hant"
		case "en-US":
			targetLang = "en"
		case "ja-JP":
			targetLang = "ja"
		}
		apiURL = fmt.Sprintf("%s?api-version=3.0&to=%s", endpoint, targetLang)

		requestBody := []map[string]string{
			{"Text": req.Text},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		client := &http.Client{Timeout: 10 * time.Second}
		azureReq, err := http.NewRequest("POST", apiURL, strings.NewReader(string(bodyBytes)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
			return
		}

		azureReq.Header.Set("Content-Type", "application/json")
		azureReq.Header.Set("Ocp-Apim-Subscription-Key", common.AZURE_TRANSLATE_KEY)
		if common.AZURE_TRANSLATE_REG != "" {
			azureReq.Header.Set("Ocp-Apim-Subscription-Region", common.AZURE_TRANSLATE_REG)
		}

		resp, err := client.Do(azureReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Azure API error"})
			return
		}

		var result []struct {
			Translations []struct {
				Text string `json:"text"`
				To   string `json:"to"`
			} `json:"translations"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to parse response"})
			return
		}

		translatedText := ""
		if len(result) > 0 && len(result[0].Translations) > 0 {
			translatedText = result[0].Translations[0].Text
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"translated_text": translatedText,
			},
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

// HandleListStrategies 获取策略列表
func HandleListStrategies(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var strategies []tasks.Strategy
		m.GORMDB.Find(&strategies)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    strategies,
		})
	}
}

// HandleSaveStrategy 保存策略 (新建或更新)
func HandleSaveStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var strategy tasks.Strategy
		if err := json.NewDecoder(r.Body).Decode(&strategy); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strategy.ID > 0 {
			m.GORMDB.Save(&strategy)
		} else {
			m.GORMDB.Create(&strategy)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": strategy})
	}
}

// HandleDeleteStrategy 删除策略
func HandleDeleteStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		m.GORMDB.Delete(&tasks.Strategy{}, id)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// HandleListIdentities 获取身份列表
func HandleListIdentities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var identities []tasks.UserIdentity
		m.GORMDB.Find(&identities)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    identities,
		})
	}
}

// HandleSaveIdentity 保存身份
func HandleSaveIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var identity tasks.UserIdentity
		if err := json.NewDecoder(r.Body).Decode(&identity); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if identity.ID > 0 {
			m.GORMDB.Save(&identity)
		} else {
			m.GORMDB.Create(&identity)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": identity})
	}
}

// HandleDeleteIdentity 删除身份
func HandleDeleteIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		m.GORMDB.Delete(&tasks.UserIdentity{}, id)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// HandleListShadowRules 获取影子规则列表
func HandleListShadowRules(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var rules []tasks.ShadowRule
		m.GORMDB.Find(&rules)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    rules,
		})
	}
}

// HandleSaveShadowRule 保存影子规则
func HandleSaveShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var rule tasks.ShadowRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if rule.ID > 0 {
			m.GORMDB.Save(&rule)
		} else {
			m.GORMDB.Create(&rule)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": rule})
	}
}

// HandleDeleteShadowRule 删除影子规则
func HandleDeleteShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		m.GORMDB.Delete(&tasks.ShadowRule{}, id)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}
