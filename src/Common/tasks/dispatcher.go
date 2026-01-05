package tasks

import (
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Dispatcher 执行分发器
type Dispatcher struct {
	db      *gorm.DB
	rdb     *redis.Client
	manager BotManager
	actions map[string]ActionHandler
}

// GetActions 获取所有注册的动作
func (d *Dispatcher) GetActions() []string {
	var names []string
	for name := range d.actions {
		names = append(names, name)
	}
	return names
}

// ActionHandler 定义动作执行接口
type ActionHandler func(task models.Task, execution *models.Execution) error

// BotManager 定义调度中心需要的机器人管理能力
type BotManager interface {
	SendBotAction(botID string, action string, params any) error
	SendToWorker(workerID string, msg types.WorkerCommand) error
	FindWorkerBySkill(skillName string) string // 返回 WorkerID
	GetTags(targetType string, targetID string) []string
	GetTargetsByTags(targetType string, tags []string, logic string) []string
	GetGroupMembers(botID string, groupID string) ([]types.MemberInfo, error)
}

func NewDispatcher(db *gorm.DB, rdb *redis.Client, manager BotManager) *Dispatcher {
	d := &Dispatcher{
		db:      db,
		rdb:     rdb,
		manager: manager,
		actions: make(map[string]ActionHandler),
	}
	d.registerDefaultActions()
	return d
}

func (d *Dispatcher) RegisterAction(name string, handler ActionHandler) {
	d.actions[name] = handler
}

// ExecuteAction 直接执行一个动作 (用于同步调用)
func (d *Dispatcher) ExecuteAction(ctx context.Context, action string, params any) (any, error) {
	handler, ok := d.actions[action]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %s", action)
	}

	// 构造一个临时的 Task 和 Execution
	paramsJSON, _ := json.Marshal(params)
	task := models.Task{
		ActionType:   action,
		ActionParams: string(paramsJSON),
	}
	execution := models.Execution{}

	err := handler(task, &execution)
	if err != nil {
		return nil, err
	}

	return execution.Result, nil
}

func (d *Dispatcher) Dispatch(execution models.Execution) {
	// 1. 更新状态为 Dispatching
	if err := d.updateStatus(execution.ID, models.ExecDispatching, nil); err != nil {
		log.Printf("[Dispatcher] Failed to update status to dispatching: %v", err)
		return
	}

	// 2. 获取关联的 Task
	var task models.Task
	if err := d.db.Preload("Tags").First(&task, execution.TaskID).Error; err != nil {
		d.updateStatus(execution.ID, models.ExecFailed, fmt.Errorf("task not found: %v", err))
		return
	}

	// 3. 更新状态为 Running
	now := time.Now()
	execution.ActualTime = &now
	if err := d.updateStatus(execution.ID, models.ExecRunning, nil); err != nil {
		log.Printf("[Dispatcher] Failed to update status to running: %v", err)
		return
	}

	// 4. 查找处理器并执行
	handler, ok := d.actions[task.ActionType]
	if !ok {
		d.updateStatus(execution.ID, models.ExecFailed, fmt.Errorf("unknown action type: %s", task.ActionType))
		return
	}

	err := handler(task, &execution)
	if err != nil {
		// 5. 失败处理
		execution.RetryCount++
		updates := map[string]any{
			"retry_count": execution.RetryCount,
		}

		if execution.RetryCount >= execution.MaxRetries {
			updates["status"] = models.ExecDead
			d.updateStatusDetailed(execution.ID, updates, err)
		} else {
			// 计算下次重试时间 (指数退避)
			nextRetry := time.Now().Add(time.Duration(execution.RetryCount*execution.RetryCount) * time.Minute)
			execution.NextRetryTime = &nextRetry
			updates["status"] = models.ExecFailed
			updates["next_retry_time"] = nextRetry
			d.updateStatusDetailed(execution.ID, updates, err)
		}
	} else {
		// 6. 成功处理
		d.updateStatus(execution.ID, models.ExecSuccess, nil)
	}
}

func (d *Dispatcher) updateStatusDetailed(id uint, updates map[string]any, execErr error) error {
	if execErr != nil {
		result := map[string]string{
			"error": execErr.Error(),
			"time":  time.Now().Format(time.RFC3339),
		}
		resultJSON, _ := json.Marshal(result)
		updates["result"] = string(resultJSON)
	}

	return d.db.Model(&models.Execution{}).Where("id = ?", id).Updates(updates).Error
}

func (d *Dispatcher) updateStatus(id uint, status models.ExecutionStatus, execErr error) error {
	updates := map[string]any{
		"status": status,
	}
	return d.updateStatusDetailed(id, updates, execErr)
}
