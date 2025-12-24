package tasks

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Dispatcher 执行分发器
type Dispatcher struct {
	db      *gorm.DB
	rdb     *redis.Client
	manager interface{} // 引用 main.Manager 以便调用机器人操作
	actions map[string]ActionHandler
}

// ActionHandler 定义动作执行接口
type ActionHandler func(task Task, execution *Execution) error

func NewDispatcher(db *gorm.DB, rdb *redis.Client, manager interface{}) *Dispatcher {
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

func (d *Dispatcher) Dispatch(execution Execution) {
	// 1. 更新状态为 Dispatching
	if err := d.updateStatus(execution.ID, ExecDispatching, nil); err != nil {
		log.Printf("[Dispatcher] Failed to update status to dispatching: %v", err)
		return
	}

	// 2. 获取关联的 Task
	var task Task
	if err := d.db.Preload("Tags").First(&task, execution.TaskID).Error; err != nil {
		d.updateStatus(execution.ID, ExecFailed, fmt.Errorf("task not found: %v", err))
		return
	}

	// 3. 更新状态为 Running
	now := time.Now()
	execution.ActualTime = &now
	if err := d.updateStatus(execution.ID, ExecRunning, nil); err != nil {
		log.Printf("[Dispatcher] Failed to update status to running: %v", err)
		return
	}

	// 4. 查找处理器并执行
	handler, ok := d.actions[task.ActionType]
	if !ok {
		d.updateStatus(execution.ID, ExecFailed, fmt.Errorf("unknown action type: %s", task.ActionType))
		return
	}

	err := handler(task, &execution)
	if err != nil {
		// 5. 失败处理
		execution.RetryCount++
		updates := map[string]interface{}{
			"retry_count": execution.RetryCount,
		}

		if execution.RetryCount >= execution.MaxRetries {
			updates["status"] = ExecDead
			d.updateStatusDetailed(execution.ID, updates, err)
		} else {
			// 计算下次重试时间 (指数退避)
			nextRetry := time.Now().Add(time.Duration(execution.RetryCount*execution.RetryCount) * time.Minute)
			execution.NextRetryTime = &nextRetry
			updates["status"] = ExecFailed
			updates["next_retry_time"] = nextRetry
			d.updateStatusDetailed(execution.ID, updates, err)
		}
	} else {
		// 6. 成功处理
		d.updateStatus(execution.ID, ExecSuccess, nil)
	}
}

func (d *Dispatcher) updateStatusDetailed(id uint, updates map[string]interface{}, execErr error) error {
	if execErr != nil {
		result := map[string]string{
			"error": execErr.Error(),
			"time":  time.Now().Format(time.RFC3339),
		}
		resultJSON, _ := json.Marshal(result)
		updates["result"] = string(resultJSON)
	}

	return d.db.Model(&Execution{}).Where("id = ?", id).Updates(updates).Error
}

func (d *Dispatcher) updateStatus(id uint, status ExecutionStatus, execErr error) error {
	updates := map[string]interface{}{
		"status": status,
	}
	return d.updateStatusDetailed(id, updates, execErr)
}
