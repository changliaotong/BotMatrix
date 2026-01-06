package types

import "time"

// TaskType 定义任务类型
type TaskType string

const (
	TaskTypeRule   TaskType = "rule"   // 规则任务：完全确定、无推理空间
	TaskTypeAI     TaskType = "ai"     // 智能任务：由大模型生成执行步骤
	TaskTypeHybrid TaskType = "hybrid" // 混合任务：AI 规划，规则执行
)

// TaskStatus 定义任务状态
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusPaused  TaskStatus = "paused"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusFailed  TaskStatus = "failed"
)

// RedisTaskSchema 定义 Redis 中任务对象的唯一合法结构
// 对应 Redis Schema 规范
type RedisTaskSchema struct {
	TaskID      string          `json:"task_id"`
	TaskType    TaskType        `json:"task_type"`
	Status      TaskStatus      `json:"status"`
	Owner       TaskOwner       `json:"owner"`
	Context     TaskContext     `json:"context"`
	Steps       []TaskStep      `json:"steps"`
	CurrentStep int             `json:"current_step"`
	Timestamps  TaskTimestamps  `json:"timestamps"`
}

type TaskOwner struct {
	WorkerID   string `json:"worker_id"`
	EmployeeID string `json:"employee_id"`
}

type TaskContext struct {
	Source   string                 `json:"source"`  // timer | user | system | ai
	Trigger  string                 `json:"trigger"` // cron | webhook | message
	Metadata map[string]interface{} `json:"metadata"`
}

type TaskStep struct {
	StepID int                    `json:"step_id"`
	Tool   string                 `json:"tool"` // MCP Tool Name
	Params map[string]interface{} `json:"params"`
	Status TaskStatus             `json:"status"`
	Result interface{}            `json:"result"` // 允许 null
	Error  interface{}            `json:"error"`  // 允许 null
}

type TaskTimestamps struct {
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
