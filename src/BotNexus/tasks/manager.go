package tasks

import (
	log "BotMatrix/common/log"
	"BotMatrix/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TaskManager 任务系统总管理器
type TaskManager struct {
	DB           *gorm.DB
	Rdb          *redis.Client
	Scheduler    *Scheduler
	Dispatcher   *Dispatcher
	Tagging      *TaggingManager
	AI           *AIParser
	Interceptors *InterceptorManager
	BotManager   BotManager
	Executor     TaskExecutor // 新增：任务执行器接口
}

// TaskExecutor 定义了执行 AI 草稿的接口，用于解耦
type TaskExecutor interface {
	ExecuteAIDraft(draft *AIDraft) error
}

func NewTaskManager(db *gorm.DB, rdb *redis.Client, botManager BotManager) *TaskManager {
	// 自动迁移表结构
	err := db.AutoMigrate(&Task{}, &Execution{}, &Tag{}, &TaskTag{}, &Strategy{}, &AIDraft{}, &UserIdentity{}, &ShadowRule{})
	if err != nil {
		log.Printf("[TaskManager] AutoMigrate failed: %v", err)
	}

	dispatcher := NewDispatcher(db, rdb, botManager)
	scheduler := NewScheduler(db, dispatcher)
	tagging := NewTaggingManager(db)
	ai := NewAIParser()
	interceptors := NewInterceptorManager(db, ai)

	return &TaskManager{
		DB:           db,
		Rdb:          rdb,
		Scheduler:    scheduler,
		Dispatcher:   dispatcher,
		Tagging:      tagging,
		AI:           ai,
		Interceptors: interceptors,
		BotManager:   botManager,
	}
}

// SetExecutor 设置任务执行器
func (tm *TaskManager) SetExecutor(executor TaskExecutor) {
	tm.Executor = executor
}

func (tm *TaskManager) Start() {
	tm.Scheduler.Start()
}

func (tm *TaskManager) Stop() {
	tm.Scheduler.Stop()
}

// CheckRateLimit 检查频率限制 (Redis 实现)
func (tm *TaskManager) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if tm.Rdb == nil {
		return true, nil // 如果没有 Redis，默认跳过检查 (或实现内存版)
	}

	redisKey := fmt.Sprintf("ratelimit:ai_task:%s", key)

	// 使用 Redis 事务 (PipeLine)
	pipe := tm.Rdb.TxPipeline()
	count := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return int(count.Val()) <= limit, nil
}

// GetStrategyConfig 获取策略配置
func (tm *TaskManager) GetStrategyConfig(name string, out any) bool {
	var strategy Strategy
	if err := tm.DB.Where("name = ? AND is_enabled = true", name).First(&strategy).Error; err != nil {
		return false
	}
	err := json.Unmarshal([]byte(strategy.Config), out)
	return err == nil
}

// setUserDefaultGroup 记录或设置用户的默认操作群组
func (tm *TaskManager) setUserDefaultGroup(userID string, groupID string) {
	if groupID == "" || groupID == "0" {
		return
	}

	var identity UserIdentity
	if err := tm.DB.Where("platform_uid = ?", userID).First(&identity).Error; err == nil {
		var metadata map[string]any
		json.Unmarshal([]byte(identity.Metadata), &metadata)
		if metadata == nil {
			metadata = make(map[string]any)
		}
		metadata["default_group"] = groupID
		metaJSON, _ := json.Marshal(metadata)
		tm.DB.Model(&identity).Update("metadata", string(metaJSON))
	}
}

// getUserDefaultGroup 获取用户的默认操作群组
func (tm *TaskManager) getUserDefaultGroup(userID string) string {
	var identity UserIdentity
	if err := tm.DB.Where("platform_uid = ?", userID).First(&identity).Error; err == nil {
		var metadata map[string]any
		json.Unmarshal([]byte(identity.Metadata), &metadata)
		if metadata != nil {
			if groupID, ok := metadata["default_group"].(string); ok {
				return groupID
			}
		}
	}
	return ""
}

// ProcessChatMessage 处理群聊中的 AI 任务指令
func (tm *TaskManager) ProcessChatMessage(ctx context.Context, botID, groupID, userID, content string) error {
	isPrivate := groupID == "" || groupID == "0"
	effectiveGroupID := groupID

	// 记录/更新用户的默认群组（仅限群聊）
	if !isPrivate {
		tm.setUserDefaultGroup(userID, groupID)
	} else {
		// 私聊模式，尝试获取默认群组
		effectiveGroupID = tm.getUserDefaultGroup(userID)
	}

	// 确定回复方式
	replyAction := "send_group_msg"
	replyParams := map[string]any{"group_id": groupID}
	if isPrivate {
		replyAction = "send_private_msg"
		replyParams = map[string]any{"user_id": userID}
	}

	// 1. 检查是否是“确认 [DraftID]”指令
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "#确认 ") || strings.HasPrefix(content, "确认 ") {
		draftID := strings.TrimPrefix(content, "#确认 ")
		draftID = strings.TrimPrefix(draftID, "确认 ")
		// 只取第一行并去除空格
		lines := strings.Split(draftID, "\n")
		draftID = strings.TrimSpace(lines[0])

		if len(draftID) == 16 {
			var draft AIDraft
			if err := tm.DB.Where("draft_id = ? AND status = 'pending'", draftID).First(&draft).Error; err == nil {
				// 权限校验 (简化：确认者必须是该群成员)
				// TODO: 进一步细化权限，例如只有发起者或管理员可以确认

				// 发送执行中提示
				p := make(map[string]any)
				for k, v := range replyParams {
					p[k] = v
				}
				p["message"] = fmt.Sprintf("🚀 收到确认！正在执行任务 [%s]...", draftID)
				tm.BotManager.SendBotAction(botID, replyAction, p)

				// 调用执行器执行任务
				if tm.Executor != nil {
					err := tm.Executor.ExecuteAIDraft(&draft)
					if err != nil {
						p["message"] = fmt.Sprintf("❌ 执行失败：%v", err)
						return tm.BotManager.SendBotAction(botID, replyAction, p)
					}
				}

				tm.DB.Model(&draft).Update("status", "confirmed")

				// 获取最后插入的任务 ID (如果有的话)
				var lastTask Task
				var successMsg string
				if err := tm.DB.Where("creator_id = ?", draft.UserID).Order("id DESC").First(&lastTask).Error; err == nil {
					successMsg = fmt.Sprintf("✅ 任务 [%s] 已成功执行！\n\n📌 任务 ID: #%d\nℹ️ 如需取消此任务，请回复：\n#取消 %d\n或对我说：\n\"取消刚才的任务\"", draftID, lastTask.ID, lastTask.ID)
				} else {
					successMsg = fmt.Sprintf("✅ 任务 [%s] 已成功执行！", draftID)
				}

				p["message"] = successMsg
				return tm.BotManager.SendBotAction(botID, replyAction, p)
			}
		}
	}

	// 1.5 检查是否是“取消 [TaskID]”指令
	if strings.HasPrefix(content, "#取消 ") || strings.HasPrefix(content, "取消 ") {
		taskIDStr := strings.TrimPrefix(content, "#取消 ")
		taskIDStr = strings.TrimPrefix(taskIDStr, "取消 ")
		taskIDStr = strings.TrimSpace(taskIDStr)

		if taskIDStr != "" {
			taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
			if err == nil {
				err = tm.CancelTask(uint(taskID), userID)
				p := make(map[string]any)
				for k, v := range replyParams {
					p[k] = v
				}
				if err != nil {
					p["message"] = fmt.Sprintf("❌ 取消失败：%v", err)
					return tm.BotManager.SendBotAction(botID, replyAction, p)
				}
				p["message"] = fmt.Sprintf("✅ 任务 #%d 已成功取消，调度器已停止该任务的后续执行。", taskID)
				return tm.BotManager.SendBotAction(botID, replyAction, p)
			}
		}
	}

	// 1.6 检查是否是帮助指令
	if content == "#帮助" || content == "帮助" || content == "#help" || content == "help" {
		helpMsg := "🤖 **BotNexus 任务指令帮助**\n\n" +
			"1️⃣ **自然语言任务**\n" +
			"直接对我说你想做的事，例如：\n" +
			"• \"每天上午10点提醒写周报\"\n" +
			"• \"每隔1小时报一次时\"\n" +
			"• \"取消刚才的任务\"\n\n" +
			"2️⃣ **快捷指令**\n" +
			"• `#确认 [草稿ID]` - 确认执行 AI 生成的任务\n" +
			"• `#取消 [任务ID]` - 取消指定的自动化任务\n" +
			"• `#帮助` - 显示本帮助信息\n\n" +
			"💡 *提示：所有 AI 生成的任务都需要回复确认指令后才会生效。*"

		p := make(map[string]any)
		for k, v := range replyParams {
			p[k] = v
		}
		p["message"] = helpMsg
		return tm.BotManager.SendBotAction(botID, replyAction, p)
	}

	// 2. 频率限制检查
	limitKey := "group:" + groupID
	if isPrivate {
		limitKey = "user:" + userID
	}
	allowed, _ := tm.CheckRateLimit(ctx, limitKey, 20, time.Hour)
	if !allowed {
		p := make(map[string]any)
		for k, v := range replyParams {
			p[k] = v
		}
		p["message"] = "⚠️ 哎呀，操作太频繁了！每小时只能发起 20 次 AI 任务，请稍后再试哦。"
		return tm.BotManager.SendBotAction(botID, replyAction, p)
	}

	// 3. 获取用户角色 (用于给 AI 提供上下文)
	userRole := "member"
	checkGroupID := groupID
	if isPrivate {
		checkGroupID = effectiveGroupID
	}
	if checkGroupID != "" {
		members, err := tm.BotManager.GetGroupMembers(botID, checkGroupID)
		if err == nil && len(members) > 0 {
			for _, m := range members {
				if m.UserID == userID {
					userRole = m.Role
					break
				}
			}
		}
	}
	// 兜底逻辑：管理员账号赋予 owner 权限
	if userRole == "" || userRole == "member" {
		if userID == "1653346663" || userID == "admin" || userID == "888888" {
			userRole = "owner"
		}
	}

	// 4. 调用 AI 解析
	sessionID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	req := ParseRequest{
		Input: content,
		Context: map[string]any{
			"bot_id":             botID,
			"group_id":           groupID,
			"effective_group_id": effectiveGroupID,
			"user_id":            userID,
			"user_role":          userRole,
			"is_private":         isPrivate,
			"session_id":         sessionID,
			"step":               0,
		},
	}

	log.Printf("[AI-Task] Calling AI.Parse for content: %s (Session: %s)", content, sessionID)
	result, err := tm.AI.Parse(req)
	if err != nil {
		log.Printf("[AI-Task] AI.Parse error: %v", err)
		// 如果有 AIService 接口支持 saveTrace，可以在这里记录
		return err
	}
	log.Printf("[AI-Task] AI.Parse result: intent=%s, summary=%s", result.Intent, result.Summary)

	// 记录解析结果追踪
	if svc, ok := tm.AI.GetAIService().(interface {
		SaveTrace(sessionID, botID string, step int, traceType, content, metadata string)
	}); ok {
		svc.SaveTrace(sessionID, botID, 0, "intent_parse", string(result.Intent), result.Summary)
	}

	// 5. 如果是系统查询，直接回复 Analysis
	if result.Intent == AIActionSystemQuery {
		p := make(map[string]any)
		for k, v := range replyParams {
			p[k] = v
		}
		p["message"] = fmt.Sprintf("🤖 AI 助手回复：\n\n%s", result.Analysis)
		return tm.BotManager.SendBotAction(botID, replyAction, p)
	}

	// 6. 收集并校验所有动作 (支持多任务并行)
	allActions := append([]*ParseResult{result}, result.SubActions...)
	log.Printf("[AI-Task] User role: %s", userRole)

	for _, action := range allActions {
		if action.Intent == AIActionCreateTask || action.Intent == AIActionCancelTask {
			var actionType string
			if action.Intent == AIActionCreateTask {
				if dataMap, ok := action.Data.(map[string]any); ok {
					actionType, _ = dataMap["action_type"].(string)

					// --- 跨群操作校验与自动补全 ---
					if paramsStr, ok := dataMap["action_params"].(string); ok {
						var p map[string]any
						if err := json.Unmarshal([]byte(paramsStr), &p); err == nil {
							targetGroupID, _ := p["group_id"].(string)
							if targetGroupID != "" && targetGroupID != effectiveGroupID {
								if userRole != "owner" && userRole != "admin" {
									p := make(map[string]any)
									for k, v := range replyParams {
										p[k] = v
									}
									p["message"] = fmt.Sprintf("🚫 权限拦截：您没有权限为群组 %s 创建任务。", targetGroupID)
									return tm.BotManager.SendBotAction(botID, replyAction, p)
								}
							}
							if targetGroupID == "" && effectiveGroupID != "" {
								p["group_id"] = effectiveGroupID
								newParams, _ := json.Marshal(p)
								dataMap["action_params"] = string(newParams)
							}
						}
					}
				}
			} else {
				actionType = "cancel_task"
			}

			policyCtx := UserContext{
				UserID:  userID,
				GroupID: checkGroupID,
				Role:    userRole,
			}
			policy := CheckCapabilityPolicy(tm.AI.Manifest, actionType, policyCtx)
			if !policy.Allowed {
				p := make(map[string]any)
				for k, v := range replyParams {
					p[k] = v
				}
				p["message"] = fmt.Sprintf("🚫 权限拦截 (%s)：\n%s", action.Summary, policy.Reason)
				return tm.BotManager.SendBotAction(botID, replyAction, p)
			}
		}
	}

	// 5. 判断是否需要即时执行 (如果所有动作都是低风险或即时类的)
	allImmediate := true
	for _, action := range allActions {
		immediate := false
		if action.Intent == AIActionCancelTask || action.Intent == AIActionSkillCall {
			immediate = true
		} else if action.Intent == AIActionCreateTask {
			if dataMap, ok := action.Data.(map[string]any); ok {
				actionType, _ := dataMap["action_type"].(string)
				triggerType, _ := dataMap["type"].(string)
				capability, ok := tm.AI.Manifest.Actions[actionType]
				if ok && capability.RiskLevel == "low" && triggerType == "once" {
					immediate = true
				}
			}
		}

		if !immediate {
			allImmediate = false
			break
		}
	}

	if allImmediate && tm.Executor != nil {
		log.Printf("[AI-Task] All %d actions are immediate-eligible", len(allActions))
		var uID uint
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uID = uint(id)
		}

		var executeErrors []error
		for _, action := range allActions {
			dataJSON, _ := json.Marshal(action.Data)
			tempDraft := &AIDraft{
				UserID:   uID,
				GroupID:  effectiveGroupID,
				UserRole: userRole,
				Intent:   string(action.Intent),
				Data:     string(dataJSON),
				Status:   "confirmed",
			}
			if err := tm.Executor.ExecuteAIDraft(tempDraft); err != nil {
				executeErrors = append(executeErrors, err)
			}
		}

		p := make(map[string]any)
		for k, v := range replyParams {
			p[k] = v
		}

		if len(executeErrors) > 0 {
			p["message"] = fmt.Sprintf("⚠️ 部分操作执行失败 (%d/%d)：\n%v", len(executeErrors), len(allActions), executeErrors[0])
		} else {
			p["message"] = fmt.Sprintf("✅ 已为您完成操作：\n%s", result.Summary)
		}
		return tm.BotManager.SendBotAction(botID, replyAction, p)
	}

	// 6. 生成草稿并存储 (针对需要确认的高风险/持久化操作)
	draftID := utils.GenerateRandomToken(8)
	log.Printf("[AI-Task] Generating draft: %s", draftID)

	var uID uint
	if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
		uID = uint(id)
	}

	draft := AIDraft{
		DraftID:    draftID,
		UserID:     uID,
		GroupID:    effectiveGroupID,
		UserRole:   userRole,
		Status:     "pending",
		ExpireTime: time.Now().Add(15 * time.Minute),
	}

	if len(allActions) > 1 {
		// 批量任务草稿
		draft.Intent = string(AIActionBatch)
		var batchData []map[string]any
		for _, a := range allActions {
			batchData = append(batchData, map[string]any{
				"intent": string(a.Intent),
				"data":   a.Data,
			})
		}
		dataJSON, _ := json.Marshal(batchData)
		draft.Data = string(dataJSON)
	} else {
		// 单个任务草稿
		draft.Intent = string(result.Intent)
		dataJSON, _ := json.Marshal(result.Data)
		draft.Data = string(dataJSON)
	}

	if err := tm.DB.Create(&draft).Error; err != nil {
		log.Printf("[AI-Task] Failed to save draft: %v", err)
		p := make(map[string]any)
		for k, v := range replyParams {
			p[k] = v
		}
		p["message"] = fmt.Sprintf("❌ 系统错误：保存任务草稿失败 (%v)", err)
		return tm.BotManager.SendBotAction(botID, replyAction, p)
	}

	// 发送带引导的回复
	summary := result.Summary
	if isPrivate && effectiveGroupID != "" {
		summary = fmt.Sprintf("%s (目标群组: %s)", summary, effectiveGroupID)
	}

	guideMsg := fmt.Sprintf("🤖 AI 已为您编排好任务草稿：\n\n📝 摘要：%s\n💡 推理：%s\n\n✅ 确认执行请在 15 分钟内回复：\n#确认 %s\n\n❌ 如需取消请忽略此消息。",
		summary, result.Analysis, draftID)

	p := make(map[string]any)
	for k, v := range replyParams {
		p[k] = v
	}
	p["message"] = guideMsg
	return tm.BotManager.SendBotAction(botID, replyAction, p)
}

// CreateTask 创建任务，包含版本限制逻辑
func (tm *TaskManager) CreateTask(task *Task, isEnterprise bool) error {
	// 试用版与企业版功能差异校验
	if !isEnterprise {
		// 试用版限制：
		// 1. 基础任务类型 (once, cron)
		if task.Type != "once" && task.Type != "cron" {
			return gorm.ErrInvalidData // 简化错误处理
		}
		// 2. 单群限制 (在 ActionParams 中校验)
		// 3. 标签数量限制
		if len(task.Tags) > 1 {
			return gorm.ErrInvalidData
		}
	}

	// 初始化下一次执行时间
	if task.NextRunTime == nil {
		task.NextRunTime = tm.Scheduler.CalculateNextRun(*task)
	}

	return tm.DB.Create(task).Error
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID uint, userIDStr string) error {
	var task Task
	if err := tm.DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("未找到任务 #%d", taskID)
	}

	if task.Status == TaskDisabled || task.Status == TaskCompleted {
		return fmt.Errorf("任务当前状态为 %s，无需取消", task.Status)
	}

	// 权限校验
	creatorID, _ := strconv.ParseUint(userIDStr, 10, 64)
	if uint(creatorID) != task.CreatorID {
		// 如果不是创建者，检查是否是系统管理员（这里可以根据实际业务逻辑扩展）
		// 目前简单处理：非创建者不可取消
		return fmt.Errorf("权限不足：只有任务创建者可以取消该任务")
	}

	return tm.DB.Model(&task).Updates(map[string]any{
		"status":        TaskDisabled,
		"next_run_time": nil,
	}).Error
}

// CheckAndTriggerConditions 检查并触发条件任务
func (tm *TaskManager) CheckAndTriggerConditions(eventType string, context map[string]any) {
	var tasks []Task
	// 查找对应类型的条件任务
	err := tm.DB.Where("status = ? AND type = ?", TaskPending, "condition").Find(&tasks).Error
	if err != nil {
		return
	}

	for _, task := range tasks {
		if tm.matchCondition(task, eventType, context) {
			// 条件满足，立即触发执行
			tm.Scheduler.triggerTask(task)
		}
	}
}

func (tm *TaskManager) matchCondition(task Task, eventType string, context map[string]any) bool {
	// 简化实现：检查 TriggerConfig 中的条件
	// 示例: {"event": "message", "keyword": "help"}
	return false // 实际应实现更复杂的逻辑
}

// GetExecutionHistory 获取执行历史
func (tm *TaskManager) GetExecutionHistory(taskID uint, limit int) ([]Execution, error) {
	var history []Execution
	err := tm.DB.Where("task_id = ?", taskID).Order("created_at desc").Limit(limit).Find(&history).Error
	return history, err
}
