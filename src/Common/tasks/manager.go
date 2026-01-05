package tasks

import (
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
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
	Syncer       *TaskSyncer  // 新增：任务同步器
}

// GetAI 获取 AI 解析器
func (tm *TaskManager) GetAI() types.AIParserInterface {
	return tm.AI
}

// GetDispatcher 获取任务分发器
func (tm *TaskManager) GetDispatcher() types.DispatcherInterface {
	return tm.Dispatcher
}

// GetInterceptors 获取拦截器管理器
func (tm *TaskManager) GetInterceptors() types.InterceptorManagerInterface {
	return tm.Interceptors
}

// GetTagging 获取标签管理器
func (tm *TaskManager) GetTagging() types.TaggingManagerInterface {
	return tm.Tagging
}

// TaskExecutor 定义了执行 AI 草稿的接口，用于解耦
type TaskExecutor interface {
	ExecuteAIDraft(draft *models.AIDraft) error
}

func NewTaskManager(db *gorm.DB, rdb *redis.Client, botManager BotManager, sourceID string) *TaskManager {
	// 自动迁移表结构
	err := db.AutoMigrate(
		&models.Task{},
		&models.Execution{},
		&models.Tag{},
		&models.TaskTag{},
		&models.Strategy{},
		&models.AIDraft{},
		&models.UserIdentity{},
		&models.ShadowRule{},
	)
	if err != nil {
		log.Printf("[TaskManager] AutoMigrate failed: %v", err)
	}

	dispatcher := NewDispatcher(db, rdb, botManager)
	scheduler := NewScheduler(db, dispatcher)
	tagging := NewTaggingManager(db)
	ai := NewAIParser()
	interceptors := NewInterceptorManager(db, ai)

	tm := &TaskManager{
		DB:           db,
		Rdb:          rdb,
		Scheduler:    scheduler,
		Dispatcher:   dispatcher,
		Tagging:      tagging,
		AI:           ai,
		Interceptors: interceptors,
		BotManager:   botManager,
	}

	tm.Syncer = NewTaskSyncer(db, rdb, tm, sourceID)
	return tm
}

// SetExecutor 设置任务执行器
func (tm *TaskManager) SetExecutor(executor TaskExecutor) {
	tm.Executor = executor
}

// Start 启动任务系统
func (tm *TaskManager) Start(startScheduler bool) {
	if startScheduler {
		tm.Scheduler.Start()
	}
	if tm.Syncer != nil {
		tm.Syncer.Start(context.Background())
	}
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
	var strategy models.Strategy
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

	var identity models.UserIdentity
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
	var identity models.UserIdentity
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

// CreateTask 创建任务
func (tm *TaskManager) CreateTask(task *models.Task, isEnterprise bool) error {
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

	if err := tm.DB.Create(task).Error; err != nil {
		return err
	}

	// 跨组件同步
	if tm.Syncer != nil {
		tm.Syncer.SyncTask(*task, SyncActionUpdate)
	}
	return nil
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID uint, userIDStr string) error {
	var task models.Task
	if err := tm.DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("未找到任务 #%d", taskID)
	}

	if task.Status == models.TaskDisabled || task.Status == models.TaskCompleted {
		return fmt.Errorf("任务当前状态为 %s，无需取消", task.Status)
	}

	// 权限校验
	creatorID, _ := strconv.ParseUint(userIDStr, 10, 64)
	if uint(creatorID) != task.CreatorID {
		// 如果不是创建者，检查是否是系统管理员（这里可以根据实际业务逻辑扩展）
		// 目前简单处理：非创建者不可取消
		return fmt.Errorf("权限不足：只有任务创建者可以取消该任务")
	}

	err := tm.DB.Model(&task).Updates(map[string]any{
		"status":        models.TaskDisabled,
		"next_run_time": nil,
	}).Error

	if err == nil && tm.Syncer != nil {
		tm.Syncer.SyncTask(task, SyncActionUpdate)
	}
	return err
}

// UpdateTask 更新任务
func (tm *TaskManager) UpdateTask(task *models.Task) error {
	err := tm.DB.Save(task).Error
	if err == nil && tm.Syncer != nil {
		tm.Syncer.SyncTask(*task, SyncActionUpdate)
	}
	return err
}

// DeleteTask 删除任务
func (tm *TaskManager) DeleteTask(taskID uint) error {
	var task models.Task
	if err := tm.DB.First(&task, taskID).Error; err == nil {
		err = tm.DB.Delete(&task).Error
		if err == nil && tm.Syncer != nil {
			tm.Syncer.SyncTask(task, SyncActionDelete)
		}
		return err
	}
	return tm.DB.Delete(&models.Task{}, taskID).Error
}

// CheckAndTriggerConditions 检查并触发条件任务
func (tm *TaskManager) CheckAndTriggerConditions(eventType string, context map[string]any) {
	var tasks []models.Task
	// 查找对应类型的条件任务
	err := tm.DB.Where("status = ? AND type = ?", models.TaskPending, "condition").Find(&tasks).Error
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

func (tm *TaskManager) matchCondition(task models.Task, eventType string, context map[string]any) bool {
	// 简化实现：检查 TriggerConfig 中的条件
	// 示例: {"event": "message", "keyword": "help"}
	return false // 实际应实现更复杂的逻辑
}

// GetExecutionHistory 获取执行历史
func (tm *TaskManager) GetExecutionHistory(taskID uint, limit int) ([]models.Execution, error) {
	var history []models.Execution
	err := tm.DB.Where("task_id = ?", taskID).Order("created_at desc").Limit(limit).Find(&history).Error
	return history, err
}

// ProcessChatMessage 处理群聊中的 AI 任务指令
func (tm *TaskManager) ProcessChatMessage(ctx context.Context, botID, groupID, userID, content string) error {
	isPrivate := groupID == "" || groupID == "0"

	// 记录/更新用户的默认群组（仅限群聊）
	if !isPrivate {
		tm.setUserDefaultGroup(userID, groupID)
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
			var draft models.AIDraft
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
				var lastTask models.Task
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

	// 2. 检查是否是“取消 [TaskID]”指令
	if strings.HasPrefix(content, "#取消 ") || strings.HasPrefix(content, "取消 ") {
		taskIDStr := strings.TrimPrefix(content, "#取消 ")
		taskIDStr = strings.TrimPrefix(taskIDStr, "取消 ")
		taskIDStr = strings.TrimSpace(taskIDStr)

		if taskID, err := strconv.ParseUint(taskIDStr, 10, 64); err == nil {
			if err := tm.CancelTask(uint(taskID), userID); err != nil {
				p := make(map[string]any)
				for k, v := range replyParams {
					p[k] = v
				}
				p["message"] = fmt.Sprintf("❌ 取消失败：%v", err)
				return tm.BotManager.SendBotAction(botID, replyAction, p)
			}

			p := make(map[string]any)
			for k, v := range replyParams {
				p[k] = v
			}
			p["message"] = fmt.Sprintf("✅ 任务 #%d 已成功取消", taskID)
			return tm.BotManager.SendBotAction(botID, replyAction, p)
		}
	}

	// 3. AI 意图解析与分发
	// TODO: 完善 AI 意图解析逻辑
	return nil
}
