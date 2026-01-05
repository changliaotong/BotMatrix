package tasks

import (
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	TaskEventStream   = "botmatrix:tasks:events"
	TaskCommandStream = "botmatrix:tasks:commands"

	SyncActionUpdate = "updated"
	SyncActionDelete = "deleted"
)

type TaskEvent struct {
	Type      string `json:"type"` // "created", "updated", "deleted", "triggered"
	TaskID    uint   `json:"task_id"`
	Payload   any    `json:"payload,omitempty"`
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"` // worker_id or "nexus"
}

type TaskSyncer struct {
	db       *gorm.DB
	rdb      *redis.Client
	manager  *TaskManager
	sourceID string
}

func NewTaskSyncer(db *gorm.DB, rdb *redis.Client, manager *TaskManager, sourceID string) *TaskSyncer {
	return &TaskSyncer{
		db:       db,
		rdb:      rdb,
		manager:  manager,
		sourceID: sourceID,
	}
}

func (s *TaskSyncer) Start(ctx context.Context) {
	if s.rdb == nil {
		return
	}
	go s.listenEvents(ctx)
}

func (s *TaskSyncer) PublishEvent(eventType string, taskID uint, payload any) error {
	if s.rdb == nil {
		return nil
	}

	event := TaskEvent{
		Type:      eventType,
		TaskID:    taskID,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		Source:    s.sourceID,
	}

	data, _ := json.Marshal(event)
	return s.rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: TaskEventStream,
		Values: map[string]interface{}{"data": data},
	}).Err()
}

func (s *TaskSyncer) listenEvents(ctx context.Context) {
	lastID := "$"
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := s.rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{TaskEventStream, lastID},
				Count:   10,
				Block:   0,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					log.Printf("[TaskSyncer] XRead error: %v", err)
					time.Sleep(time.Second)
				}
				continue
			}

			for _, stream := range res {
				for _, msg := range stream.Messages {
					lastID = msg.ID
					var event TaskEvent
					if data, ok := msg.Values["data"].(string); ok {
						if err := json.Unmarshal([]byte(data), &event); err != nil {
							continue
						}
						// 忽略自己发出的事件
						if event.Source == s.sourceID {
							continue
						}
						s.handleEvent(event)
					}
				}
			}
		}
	}
}

// SyncTask 同步任务状态变更
func (s *TaskSyncer) SyncTask(task models.Task, action string) {
	s.PublishEvent(action, task.ID, task)
}

func (s *TaskSyncer) handleEvent(event TaskEvent) {
	log.Printf("[TaskSyncer] Received event: %s for Task #%d from %s", event.Type, event.TaskID, event.Source)

	switch event.Type {
	case "created", "updated":
		// 任务更新，如果是 Scheduler 在运行，可能需要重新计算时间
		// 这里的实现依赖于 Scheduler 是轮询 DB 的，所以这里其实不需要做什么，
		// 除非 Scheduler 是内存型的。
	case "triggered":
		// 任务已被其他节点触发
	}
}

// RequestTaskExecution 请求执行一个任务 (通过 Redis 分发)
func (s *TaskSyncer) RequestTaskExecution(taskID uint) error {
	if s.rdb == nil {
		return fmt.Errorf("redis not available")
	}

	cmd := map[string]any{
		"action":  "execute_task",
		"task_id": taskID,
		"source":  s.sourceID,
		"time":    time.Now().Unix(),
	}
	data, _ := json.Marshal(cmd)

	return s.rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: TaskCommandStream,
		Values: map[string]interface{}{"data": data},
	}).Err()
}
