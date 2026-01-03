package app

import (
	"BotMatrix/common/config"
	"BotMatrix/common/log"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"BotNexus/tasks"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HandleListTasks 获取任务列表
func HandleListTasks(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskList []tasks.Task
		m.GORMDB.Preload("Tags").Find(&taskList)
		utils.SendJSONResponse(w, true, "", taskList)
	}
}

// HandleListSystemCapabilities 获取系统任务处理能力 (Actions & Interceptors)
func HandleListSystemCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actions := m.TaskManager.Dispatcher.GetActions()
		interceptors := m.TaskManager.Interceptors.GetInterceptors()

		utils.SendJSONResponse(w, true, "", struct {
			Actions      []string `json:"actions"`
			Interceptors []string `json:"interceptors"`
		}{
			Actions:      actions,
			Interceptors: interceptors,
		})
	}
}

// HandleCreateTask 创建任务
func HandleCreateTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var task tasks.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		// 简单检查是否为企业版 (这里可以根据用户权限或配置决定)
		isEnterprise := claims.IsAdmin

		if err := m.TaskManager.CreateTask(&task, isEnterprise); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, utils.T(lang, "create_task_failed"), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", task)
	}
}

// HandleGetExecutions 获取执行记录
func HandleGetExecutions(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskIDStr := r.URL.Query().Get("task_id")
		taskID, _ := strconv.Atoi(taskIDStr)

		history, err := m.TaskManager.GetExecutionHistory(uint(taskID), 20)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		utils.SendJSONResponse(w, true, "", history)
	}
}

// HandleAIParse AI 解析
func HandleAIParse(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tasks.ParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 提取执行上下文信息
		groupID, _ := req.Context["group_id"].(string)
		platformUID, _ := req.Context["user_id"].(string)
		botID, _ := req.Context["bot_id"].(string)

		// 获取用户角色 (从缓存或 BotManager)
		userRole := "member"
		if groupID != "" && platformUID != "" {
			members, err := m.GetGroupMembers(botID, groupID)
			if err == nil {
				for _, mem := range members {
					if mem.UserID == platformUID {
						userRole = mem.Role
						break
					}
				}
			}
		}

		// 频率限制检查 (防止 AI 刷屏)
		// 优先从策略配置中读取限制，否则使用默认值 (20次/小时)
		limit := 20
		window := time.Hour
		var rateLimitCfg struct {
			Limit  int `json:"limit"`
			Window int `json:"window_seconds"`
		}
		if m.TaskManager.GetStrategyConfig("ai_rate_limit", &rateLimitCfg) {
			limit = rateLimitCfg.Limit
			if rateLimitCfg.Window > 0 {
				window = time.Duration(rateLimitCfg.Window) * time.Second
			}
		}

		if groupID != "" {
			allowed, err := m.TaskManager.CheckRateLimit(r.Context(), "group:"+groupID, limit, window)
			if err != nil {
				log.Printf("[AIParse] Rate limit check failed: %v", err)
			} else if !allowed {
				utils.SendJSONResponse(w, false, fmt.Sprintf("频率限制：该群在指定时间内 AI 任务生成已达上限 (%d次)", limit), nil)
				return
			}
		}

		result, err := m.TaskManager.AI.Parse(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		// 权限检查
		if result.Intent == tasks.AIActionCreateTask {
			// 从 Data 中提取 action_type
			var taskData struct {
				ActionType string `json:"action_type"`
			}
			json.Unmarshal([]byte(fmt.Sprintf("%v", result.Data)), &taskData) // 这里 Data 可能是 map，需要处理

			// 兼容处理：如果 result.Data 是 map[string]any
			if dataMap, ok := result.Data.(map[string]any); ok {
				if at, ok := dataMap["action_type"].(string); ok {
					taskData.ActionType = at
				}
			}

			policyCtx := tasks.UserContext{
				UserID:  platformUID,
				GroupID: groupID,
				Role:    userRole,
			}
			policy := tasks.CheckCapabilityPolicy(m.TaskManager.AI.Manifest, taskData.ActionType, policyCtx)
			if !policy.Allowed {
				utils.SendJSONResponse(w, false, policy.Reason, nil)
				return
			}
		}

		// 将解析结果存入草稿箱
		draftID := utils.GenerateRandomToken(8)
		dataJSON, _ := json.Marshal(result.Data)
		claims := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)

		draft := tasks.AIDraft{
			DraftID:    draftID,
			UserID:     uint(claims.UserID),
			GroupID:    groupID,
			UserRole:   userRole,
			Intent:     string(result.Intent),
			Data:       string(dataJSON),
			ExpireTime: time.Now().Add(15 * time.Minute), // 15分钟有效
		}
		m.GORMDB.Create(&draft)

		result.DraftID = draftID

		utils.SendJSONResponse(w, true, "", result)
	}
}

// HandleAIConfirm AI 确认执行
func HandleAIConfirm(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			utils.SendJSONResponse(w, false, "草稿不存在或已失效", nil)
			return
		}

		if time.Now().After(draft.ExpireTime) {
			m.GORMDB.Model(&draft).Update("status", "expired")
			w.WriteHeader(http.StatusGone)
			utils.SendJSONResponse(w, false, "草稿已过期", nil)
			return
		}

		// 再次进行权限检查 (防止角色变更)
		if draft.Intent == string(tasks.AIActionCreateTask) {
			var dataMap map[string]any
			json.Unmarshal([]byte(draft.Data), &dataMap)
			actionType, _ := dataMap["action_type"].(string)

			policyCtx := tasks.UserContext{
				UserID:  strconv.Itoa(int(draft.UserID)), // 这里可能需要更准确的映射，先用 UserID 占位
				GroupID: draft.GroupID,
				Role:    draft.UserRole,
			}
			policy := tasks.CheckCapabilityPolicy(m.TaskManager.AI.Manifest, actionType, policyCtx)
			if !policy.Allowed {
				utils.SendJSONResponse(w, false, "执行前权限校验失败: "+policy.Reason, nil)
				return
			}
		}

		// 使用统一的执行逻辑
		err := m.executeAIDraft(&draft)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		m.GORMDB.Model(&draft).Update("status", "confirmed")
		utils.SendJSONResponse(w, true, "执行成功", nil)
	}
}

// HandleTranslate 翻译接口 (目前使用 Azure 服务)
func HandleTranslate(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Text       string `json:"text"`
			TargetLang string `json:"target_lang"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if config.GlobalConfig.AzureTranslateKey == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			utils.SendJSONResponse(w, false, "Azure Translate API key not configured", nil)
			return
		}

		endpoint := config.GlobalConfig.AzureTranslateEndpoint
		if endpoint == "" {
			endpoint = "https://api.cognitive.microsofttranslator.com"
		}

		url := fmt.Sprintf("%s/translate?api-version=3.0&to=%s", endpoint, req.TargetLang)

		body := []struct {
			Text string `json:"text"`
		}{
			{Text: req.Text},
		}
		jsonBody, _ := json.Marshal(body)

		azureReq, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		azureReq.Header.Set("Content-Type", "application/json")
		azureReq.Header.Set("Ocp-Apim-Subscription-Key", config.GlobalConfig.AzureTranslateKey)
		if config.GlobalConfig.AzureTranslateRegion != "" {
			azureReq.Header.Set("Ocp-Apim-Subscription-Region", config.GlobalConfig.AzureTranslateRegion)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(azureReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, "Azure API error", nil)
			return
		}

		var azureResp []struct {
			Translations []struct {
				Text string `json:"text"`
				To   string `json:"to"`
			} `json:"translations"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&azureResp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, "Failed to parse response", nil)
			return
		}

		if len(azureResp) == 0 || len(azureResp[0].Translations) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, "No translation returned", nil)
			return
		}

		utils.SendJSONResponse(w, true, "", struct {
			TranslatedText string `json:"translated_text"`
		}{
			TranslatedText: azureResp[0].Translations[0].Text,
		})
	}
}

// HandleManageTags 标签管理
func HandleManageTags(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		utils.SendJSONResponse(w, true, "", nil)
	}
}

// HandleGetCapabilities 获取系统能力清单 (用于 AI 提示或功能展示)
func HandleGetCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		manifest := m.TaskManager.AI.Manifest
		prompt := m.TaskManager.AI.GetSystemPrompt()

		utils.SendJSONResponse(w, true, "", struct {
			Manifest any    `json:"manifest"`
			Prompt   string `json:"prompt"`
		}{
			Manifest: manifest,
			Prompt:   prompt,
		})
	}
}

// HandleListStrategies 获取策略列表
func HandleListStrategies(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		strategies := m.TaskManager.Interceptors.GetStrategies()
		utils.SendJSONResponse(w, true, "", strategies)
	}
}

// HandleGetStrategy 获取策略详情
func HandleGetStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		strategy, err := m.TaskManager.Interceptors.GetStrategy(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		utils.SendJSONResponse(w, true, "", strategy)
	}
}

// HandleSaveStrategy 保存策略
func HandleSaveStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		utils.SendJSONResponse(w, true, "", strategy)
	}
}

// HandleDeleteStrategy 删除策略
func HandleDeleteStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		m.TaskManager.Interceptors.DeleteStrategy(name)
		utils.SendJSONResponse(w, true, "", nil)
	}
}

// HandleListIdentities 获取身份列表
func HandleListIdentities(m *Manager) http.HandlerFunc {
	type identityResponse struct {
		tasks.UserIdentity
		Avatar string `json:"avatar"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var identities []tasks.UserIdentity
		m.GORMDB.Find(&identities)

		// 增加头像逻辑
		respIdentities := make([]identityResponse, 0, len(identities))
		for _, id := range identities {
			respIdentities = append(respIdentities, identityResponse{
				UserIdentity: id,
				Avatar:       GetAvatarURL(id.Platform, id.PlatformUID, false, ""),
			})
		}

		utils.SendJSONResponse(w, true, "", respIdentities)
	}
}

// HandleSaveIdentity 保存身份
func HandleSaveIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		utils.SendJSONResponse(w, true, "", identity)
	}
}

// HandleDeleteIdentity 删除身份
func HandleDeleteIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		m.GORMDB.Delete(&tasks.UserIdentity{}, id)
		utils.SendJSONResponse(w, true, "", nil)
	}
}

// HandleListShadowRules 获取影子规则列表
func HandleListShadowRules(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rules []tasks.ShadowRule
		m.GORMDB.Find(&rules)
		utils.SendJSONResponse(w, true, "", rules)
	}
}

// HandleSaveShadowRule 保存影子规则
func HandleSaveShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		utils.SendJSONResponse(w, true, "", rule)
	}
}

// HandleDeleteShadowRule 删除影子规则
func HandleDeleteShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		m.GORMDB.Delete(&tasks.ShadowRule{}, id)
		utils.SendJSONResponse(w, true, "", nil)
	}
}
