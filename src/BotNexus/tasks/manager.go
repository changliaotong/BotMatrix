package tasks

import (
	"log"

	"gorm.io/gorm"
)

// TaskManager 任务系统总管理器
type TaskManager struct {
	DB             *gorm.DB
	Scheduler      *Scheduler
	Dispatcher     *Dispatcher
	Tagging        *TaggingManager
	AI             *AIParser
	Interceptors   *InterceptorManager
	BotManager     BotManager
}

func NewTaskManager(db *gorm.DB, botManager BotManager) *TaskManager {
	// 自动迁移表结构
	err := db.AutoMigrate(&Task{}, &Execution{}, &Tag{}, &TaskTag{}, &Strategy{}, &AIDraft{}, &UserIdentity{}, &ShadowRule{})
	if err != nil {
		log.Printf("[TaskManager] AutoMigrate failed: %v", err)
	}

	dispatcher := NewDispatcher(db, botManager)
	scheduler := NewScheduler(db, dispatcher)
	tagging := NewTaggingManager(db)
	ai := NewAIParser()
	interceptors := NewInterceptorManager(db)

	return &TaskManager{
		DB:           db,
		Scheduler:    scheduler,
		Dispatcher:   dispatcher,
		Tagging:      tagging,
		AI:           ai,
		Interceptors: interceptors,
		BotManager:   botManager,
	}
}

func (tm *TaskManager) Start() {
	tm.Scheduler.Start()
}

func (tm *TaskManager) Stop() {
	tm.Scheduler.Stop()
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

	return tm.DB.Create(task).Error
}

// CheckAndTriggerConditions 检查并触发条件任务
func (tm *TaskManager) CheckAndTriggerConditions(eventType string, context map[string]interface{}) {
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

func (tm *TaskManager) matchCondition(task Task, eventType string, context map[string]interface{}) bool {
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
