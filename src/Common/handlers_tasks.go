package common

import (
	"BotMatrix/common/config"
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/tasks"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// --- 数字员工任务引擎 (New Task Engine) 接口 ---

// HandleListEmployeeTasks 获取数字员工任务列表
// @Summary 获取数字员工任务列表
// @Description 获取指定数字员工的所有任务，支持状态过滤
// @Tags DigitalEmployee
// @Produce json
// @Param employee_id query string false "员工 ID"
// @Param status query string false "任务状态 (pending, executing, completed, failed, pending_approval)"
// @Success 200 {object} utils.JSONResponse "任务列表"
// @Router /api/admin/employees/tasks [get]
func HandleListEmployeeTasks(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskList []models.DigitalEmployeeTaskGORM
		query := m.GORMDB.Model(&models.DigitalEmployeeTaskGORM{})

		// 过滤条件
		if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
			query = query.Where("assignee_id = ?", employeeID)
		} else if id := r.URL.Query().Get("id"); id != "" { // 兼容之前的 id 参数
			query = query.Where("assignee_id = ?", id)
		}

		if status := r.URL.Query().Get("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		if err := query.Order("created_at DESC").Find(&taskList).Error; err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", taskList)
	}
}

// HandleCreateEmployeeTask 创建数字员工任务
func HandleCreateEmployeeTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task models.DigitalEmployeeTaskGORM
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if err := m.DigitalEmployeeTaskService.CreateTask(r.Context(), &task); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		// 异步开始规划
		go m.DigitalEmployeeTaskService.PlanTask(context.Background(), task.ExecutionID)

		utils.SendJSONResponse(w, true, "Task created and planning started", task)
	}
}

// HandleApproveEmployeeTask 审批数字员工任务
func HandleApproveEmployeeTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		executionID := r.URL.Query().Get("execution_id")
		if executionID == "" {
			utils.SendJSONResponse(w, false, "Missing execution_id", nil)
			return
		}

		if err := m.DigitalEmployeeTaskService.ApproveTask(r.Context(), executionID); err != nil {
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		// 审批后自动恢复执行
		go m.DigitalEmployeeTaskService.ExecuteTask(context.Background(), executionID)

		utils.SendJSONResponse(w, true, "Task approved and execution resumed", nil)
	}
}

// HandleExecuteEmployeeTask 手动触发/恢复执行任务
func HandleExecuteEmployeeTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		executionID := r.URL.Query().Get("execution_id")
		if executionID == "" {
			utils.SendJSONResponse(w, false, "Missing execution_id", nil)
			return
		}

		go m.DigitalEmployeeTaskService.ExecuteTask(context.Background(), executionID)

		utils.SendJSONResponse(w, true, "Task execution triggered", nil)
	}
}

// --- 原有任务系统接口 (Legacy Tasks) ---

// HandleListTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取所有已定义的自动化任务列表
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Task "任务列表"
// @Router /api/admin/tasks [get]
func HandleListTasks(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskList []models.Task
		m.GORMDB.Preload("Tags").Find(&taskList)
		utils.SendJSONResponse(w, true, "", taskList)
	}
}

// HandleListSystemCapabilities 获取系统任务处理能力 (Actions & Interceptors)
// @Summary 获取系统能力
// @Description 获取系统支持的所有动作 (Actions) 和拦截器 (Interceptors)
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "系统能力列表"
// @Router /api/admin/tasks/capabilities [get]
func HandleListSystemCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tm := m.GetTaskManager()
		actions := tm.GetDispatcher().GetActions()
		interceptors := tm.GetInterceptors().GetInterceptors()

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
// @Summary 创建自动化任务
// @Description 创建一个新的自动化任务，支持定时、周期或触发式执行
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.Task true "任务定义"
// @Success 200 {object} models.Task "创建成功的任务"
// @Router /api/admin/tasks [post]
func HandleCreateTask(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		// 简单检查是否为企业版 (这里可以根据用户权限或配置决定)
		isEnterprise := claims.IsAdmin

		if err := m.GetTaskManager().CreateTask(&task, isEnterprise); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, utils.T(lang, "create_task_failed"), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", task)
	}
}

// HandleGetExecutions 获取执行记录
// @Summary 获取执行记录
// @Description 获取指定任务的最近执行历史记录
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param task_id query int true "任务 ID"
// @Success 200 {array} models.Execution "执行记录列表"
// @Router /api/admin/tasks/executions [get]
func HandleGetExecutions(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")
		var executions []models.Execution
		m.GORMDB.Where("task_id = ?", taskID).Order("created_at DESC").Find(&executions)
		utils.SendJSONResponse(w, true, "", executions)
	}
}

// HandleAIParse AI 解析
// @Summary AI 意图解析
// @Description 使用 AI 解析用户自然语言指令，识别意图并生成任务草稿
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body tasks.ParseRequest true "解析请求"
// @Success 200 {object} utils.JSONResponse "解析结果"
// @Router /api/admin/tasks/ai-parse [post]
func HandleAIParse(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ParseRequest
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
		tm := m.GetTaskManager()
		if tm.GetStrategyConfig("ai_rate_limit", &rateLimitCfg) {
			limit = rateLimitCfg.Limit
			if rateLimitCfg.Window > 0 {
				window = time.Duration(rateLimitCfg.Window) * time.Second
			}
		}

		if groupID != "" {
			allowed, err := tm.CheckRateLimit(r.Context(), "group:"+groupID, limit, window)
			if err != nil {
				log.Printf("[AIParse] Rate limit check failed: %v", err)
			} else if !allowed {
				utils.SendJSONResponse(w, false, fmt.Sprintf("频率限制：该群在指定时间内 AI 任务生成已达上限 (%d次)", limit), nil)
				return
			}
		}

		// 执行解析
		result, err := tm.GetAI().MatchSkillByLLM(r.Context(), req.Input, 0, req.Context)
		if err != nil {
			utils.SendJSONResponse(w, false, "AI解析失败: "+err.Error(), nil)
			return
		}

		// 检查策略限制 (针对解析出的 Action)
		if result.Intent != "" {
			claims := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
			policyCtx := tasks.UserContext{
				UserID:  strconv.Itoa(int(claims.UserID)),
				GroupID: groupID,
				Role:    userRole,
			}
			policy := tasks.CheckCapabilityPolicy(tm.GetAI().GetManifest(), string(result.Intent), policyCtx)
			if !policy.Allowed {
				utils.SendJSONResponse(w, false, policy.Reason, nil)
				return
			}
		}

		// 将解析结果存入草稿箱
		draftID := utils.GenerateRandomToken(8)
		dataJSON, _ := json.Marshal(result.Data)
		claims := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)

		draft := models.AIDraft{
			DraftID:    draftID,
			UserID:     uint(claims.UserID),
			GroupID:    groupID,
			UserRole:   userRole,
			Intent:     string(result.Intent),
			Data:       string(dataJSON),
			Status:     "pending",
			ExpireTime: time.Now().Add(15 * time.Minute), // 15分钟有效
		}
		m.GORMDB.Create(&draft)

		result.DraftID = draftID

		utils.SendJSONResponse(w, true, "", result)
	}
}

// HandleAIConfirm AI 确认执行
// @Summary AI 执行确认
// @Description 确认并执行之前由 AI 生成的任务草稿
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "确认请求"
// @Success 200 {object} utils.JSONResponse "执行成功"
// @Router /api/admin/tasks/ai-confirm [post]
func HandleAIConfirm(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var draft models.AIDraft
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
		if draft.Intent == string(types.AIActionCreateTask) {
			var dataMap map[string]any
			json.Unmarshal([]byte(draft.Data), &dataMap)
			actionType, _ := dataMap["action_type"].(string)

			policyCtx := tasks.UserContext{
				UserID:  strconv.Itoa(int(draft.UserID)), // 这里可能需要更准确的映射，先用 UserID 占位
				GroupID: draft.GroupID,
				Role:    draft.UserRole,
			}
			policy := tasks.CheckCapabilityPolicy(m.GetTaskManager().GetAI().GetManifest(), actionType, policyCtx)
			if !policy.Allowed {
				utils.SendJSONResponse(w, false, "执行前权限校验失败: "+policy.Reason, nil)
				return
			}
		}

		// 使用统一的执行逻辑
		err := executeAIDraft(r.Context(), m, &draft)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendJSONResponse(w, false, err.Error(), nil)
			return
		}

		m.GORMDB.Model(&draft).Update("status", "confirmed")
		utils.SendJSONResponse(w, true, "执行成功", nil)
	}
}

func executeAIDraft(ctx context.Context, m *Manager, draft *models.AIDraft) error {
	tm := m.GetTaskManager()
	if tm == nil {
		return fmt.Errorf("task manager not available")
	}

	switch types.AIActionType(draft.Intent) {
	case types.AIActionCreateTask:
		var task models.Task
		if err := json.Unmarshal([]byte(draft.Data), &task); err != nil {
			return fmt.Errorf("failed to parse task data: %v", err)
		}
		// 强制设置创建者和企业属性 (如果适用)
		task.CreatorID = uint(draft.UserID)
		// task.IsEnterprise = ... // 根据上下文决定

		if err := tm.CreateTask(&task, false); err != nil {
			return fmt.Errorf("failed to create task: %v", err)
		}
		return nil

	case types.AIActionSkillCall:
		var req struct {
			Action string         `json:"action"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &req); err != nil {
			return fmt.Errorf("failed to parse skill data: %v", err)
		}
		_, err := tm.GetDispatcher().ExecuteAction(ctx, req.Action, req.Params)
		return err

	default:
		return fmt.Errorf("unsupported AI intent: %s", draft.Intent)
	}
}

// HandleTranslate 翻译接口 (目前使用 Azure 服务)
// @Summary 文本翻译
// @Description 使用 Azure Translate 服务进行多语言文本翻译
// @Tags Tools
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "翻译请求"
// @Success 200 {object} object "翻译结果"
// @Router /api/admin/tools/translate [post]
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
// @Summary 管理标签
// @Description 为指定资源（任务、用户等）添加或移除标签
// @Tags Tools
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "标签操作请求"
// @Success 200 {object} utils.JSONResponse "操作成功"
// @Router /api/admin/tools/tags [post]
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
		tm := m.GetTaskManager()
		if req.Action == "add" {
			err = tm.GetTagging().AddTag(req.TargetType, req.TargetID, req.TagName)
		} else {
			err = tm.GetTagging().RemoveTag(req.TargetType, req.TargetID, req.TagName)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		utils.SendJSONResponse(w, true, "", nil)
	}
}

// HandleGetCapabilities 获取系统能力清单 (用于 AI 提示或功能展示)
// @Summary 获取系统能力清单
// @Description 获取系统支持的 Manifest 配置和系统 Prompt 模板
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "系统能力数据"
// @Router /api/admin/tasks/manifest [get]
func HandleGetCapabilities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tm := m.GetTaskManager()
		manifest := tm.GetAI().GetManifest()
		prompt := tm.GetAI().GetSystemPrompt()

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
// @Summary 获取策略列表
// @Description 获取系统定义的所有任务执行拦截策略列表
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Strategy "策略列表"
// @Router /api/admin/tasks/strategies [get]
func HandleListStrategies(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		strategies := m.GetTaskManager().GetInterceptors().GetStrategies()
		utils.SendJSONResponse(w, true, "", strategies)
	}
}

// HandleGetStrategy 获取特定策略详情
// @Summary 获取特定策略详情
// @Description 获取指定名称的拦截器策略详情
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param name query string true "策略名称"
// @Success 200 {object} models.Strategy "策略详情"
// @Router /api/admin/tasks/strategies/detail [get]
func HandleGetStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		strategy, err := m.GetTaskManager().GetInterceptors().GetStrategy(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		utils.SendJSONResponse(w, true, "", strategy)
	}
}

// HandleUpdateStrategy 更新策略配置
// @Summary 更新策略配置
// @Description 更新拦截器策略配置，如维护模式开关、全局限流等
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.Strategy true "策略配置"
// @Success 200 {object} utils.JSONResponse "保存成功"
// @Router /api/admin/tasks/strategies [post]
func HandleUpdateStrategy(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var strategy models.Strategy
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
		m.GetTaskManager().GetInterceptors().DeleteStrategy(name)
		utils.SendJSONResponse(w, true, "策略已删除", nil)
	}
}

func HandleListIdentities(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var identities []models.UserIdentity
		m.GORMDB.Find(&identities)
		utils.SendJSONResponse(w, true, "", identities)
	}
}

func HandleSaveIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var identity models.UserIdentity
		if err := json.NewDecoder(r.Body).Decode(&identity); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.GORMDB.Save(&identity)
		utils.SendJSONResponse(w, true, "Saved", identity)
	}
}

func HandleDeleteIdentity(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		m.GORMDB.Delete(&models.UserIdentity{}, id)
		utils.SendJSONResponse(w, true, "Deleted", nil)
	}
}

func HandleListShadowRules(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rules []models.ShadowRule
		m.GORMDB.Find(&rules)
		utils.SendJSONResponse(w, true, "", rules)
	}
}

func HandleSaveShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rule models.ShadowRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.GORMDB.Save(&rule)
		utils.SendJSONResponse(w, true, "Saved", rule)
	}
}

func HandleDeleteShadowRule(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		m.GORMDB.Delete(&models.ShadowRule{}, id)
		utils.SendJSONResponse(w, true, "Deleted", nil)
	}
}
