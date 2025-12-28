package tasks

import (
	"encoding/json"
	log "BotMatrix/common/log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler 调度器
type Scheduler struct {
	db         *gorm.DB
	dispatcher *Dispatcher
	stopChan   chan struct{}
	wg         sync.WaitGroup
	interval   time.Duration
}

func NewScheduler(db *gorm.DB, dispatcher *Dispatcher) *Scheduler {
	return &Scheduler{
		db:         db,
		dispatcher: dispatcher,
		stopChan:   make(chan struct{}),
		interval:   10 * time.Second, // 默认每10秒扫描一次
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Println("[Scheduler] Started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("[Scheduler] Stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.scanAndTrigger()
		}
	}
}

// scanAndTrigger 扫描并触发任务
func (s *Scheduler) scanAndTrigger() {
	var tasks []Task
	now := time.Now()

	// 查找需要执行的任务
	// 1. Status 为 pending
	// 2. NextRunTime <= now 且不为 NULL
	err := s.db.Where("status = ? AND next_run_time IS NOT NULL AND next_run_time <= ?", TaskPending, now).Find(&tasks).Error
	if err != nil {
		log.Printf("[Scheduler] Failed to scan tasks: %v", err)
		return
	}

	for _, task := range tasks {
		s.triggerTask(task)
	}

	// 扫描并重试失败的 Execution
	s.scanAndRetryExecutions()

	// 清理过期的 AI 草稿
	s.cleanExpiredDrafts()
}

func (s *Scheduler) cleanExpiredDrafts() {
	now := time.Now()
	err := s.db.Model(&AIDraft{}).Where("status = ? AND expire_time <= ?", "pending", now).Update("status", "expired").Error
	if err != nil {
		log.Printf("[Scheduler] Failed to clean expired drafts: %v", err)
	}
}

func (s *Scheduler) triggerTask(task Task) {
	// 原子操作：创建 Execution 并更新 Task 的 NextRunTime
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 再次检查状态，防止并发问题 (虽然目前是单中心，但保留扩展性)
		var currentTask Task
		if err := tx.First(&currentTask, task.ID).Error; err != nil {
			return err
		}
		if currentTask.Status != TaskPending || (currentTask.NextRunTime != nil && currentTask.NextRunTime.After(time.Now())) {
			return nil // 已经被处理或状态变更
		}

		executionID := uuid.New().String()
		now := time.Now()
		triggerTime := now
		if task.NextRunTime != nil {
			triggerTime = *task.NextRunTime
		}
		execution := Execution{
			TaskID:      task.ID,
			ExecutionID: executionID,
			TriggerTime: triggerTime,
			Status:      ExecPending,
		}

		if err := tx.Create(&execution).Error; err != nil {
			return err
		}

		// 计算下一次执行时间
		nextRun := s.calculateNextRun(task)
		updates := map[string]interface{}{
			"last_run_time": time.Now(),
			"next_run_time": nextRun,
		}
		if task.Type == "once" {
			updates["status"] = TaskCompleted
		}

		if err := tx.Model(&task).Updates(updates).Error; err != nil {
			return err
		}

		// 异步分发执行
		go s.dispatcher.Dispatch(execution)

		return nil
	})

	if err != nil {
		log.Printf("[Scheduler] Failed to trigger task %d: %v", task.ID, err)
	}
}

func (s *Scheduler) calculateNextRun(task Task) *time.Time {
	switch task.Type {
	case "once":
		return nil
	case "cron":
		var config struct {
			Cron string `json:"cron"`
		}
		if err := json.Unmarshal([]byte(task.TriggerConfig), &config); err != nil {
			return nil
		}
		schedule, err := cron.ParseStandard(config.Cron)
		if err != nil {
			return nil
		}
		next := schedule.Next(time.Now())
		return &next
	case "delayed":
		var config struct {
			Delay int `json:"delay"` // 秒
		}
		if err := json.Unmarshal([]byte(task.TriggerConfig), &config); err != nil {
			return nil
		}
		next := time.Now().Add(time.Duration(config.Delay) * time.Second)
		return &next
	}
	return nil
}

func (s *Scheduler) scanAndRetryExecutions() {
	var executions []Execution
	now := time.Now()

	// 查找需要重试的 Execution
	err := s.db.Where("status = ? AND next_retry_time <= ? AND retry_count < max_retries", ExecFailed, now).Find(&executions).Error
	if err != nil {
		log.Printf("[Scheduler] Failed to scan executions for retry: %v", err)
		return
	}

	for _, exec := range executions {
		go s.dispatcher.Dispatch(exec)
	}
}
