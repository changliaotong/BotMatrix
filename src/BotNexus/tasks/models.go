package tasks

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"   // 启用中
	TaskDisabled  TaskStatus = "disabled"  // 暂停
	TaskCompleted TaskStatus = "completed" // 已完成（一次性任务）
)

// ExecutionStatus 执行实例状态
type ExecutionStatus string

const (
	ExecPending     ExecutionStatus = "pending"     // 等待中
	ExecDispatching ExecutionStatus = "dispatching" // 调度中
	ExecRunning     ExecutionStatus = "running"     // 执行中
	ExecSuccess     ExecutionStatus = "success"     // 成功
	ExecFailed      ExecutionStatus = "failed"      // 失败
	ExecDead        ExecutionStatus = "dead"        // 已放弃
)

// Task 任务定义
type Task struct {
	gorm.Model
	Name          string      `gorm:"size:255;not null"`
	Type          string      `gorm:"size:50;not null"` // once, cron, delayed, condition
	ActionType    string      `gorm:"size:50;not null"` // send_message, mute_group, unmute_group
	ActionParams  string      `gorm:"type:text"`        // JSON 参数
	TriggerConfig string      `gorm:"type:text"`        // JSON 触发配置 (cron, delay, conditions)
	Status        TaskStatus  `gorm:"size:50;default:'pending'"`
	CreatorID     uint        `gorm:"index"`
	IsEnterprise  bool        `gorm:"default:false"`
	Tags          []Tag       `gorm:"many2many:task_tags;"`
	Executions    []Execution `gorm:"foreignKey:TaskID"`
	LastRunTime   *time.Time
	NextRunTime   *time.Time `gorm:"index"`
}

// Execution 任务执行实例
type Execution struct {
	gorm.Model
	TaskID        uint      `gorm:"index;not null"`
	ExecutionID   string    `gorm:"size:100;uniqueIndex;not null"`
	TriggerTime   time.Time `gorm:"index"`
	ActualTime    *time.Time
	Status        ExecutionStatus `gorm:"size:50;default:'pending'"`
	Result        string          `gorm:"type:text"` // JSON 结果或错误详情
	RetryCount    int             `gorm:"default:0"`
	MaxRetries    int             `gorm:"default:3"`
	NextRetryTime *time.Time      `gorm:"index"`
	TraceID       string          `gorm:"size:100"`
}

// Tag 标签定义
type Tag struct {
	gorm.Model
	Name       string `gorm:"size:100;not null;index"`
	Type       string `gorm:"size:50;not null"` // group, friend
	TargetID   string `gorm:"size:255;not null;index"`
	IsInternal bool   `gorm:"default:false"` // 系统内置标签
}

// Strategy 全局策略定义
type Strategy struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex"`
	Type        string `gorm:"size:50"`   // rate_limit, maintenance, flow_control
	Config      string `gorm:"type:text"` // JSON 配置
	IsEnabled   bool   `gorm:"default:false"`
	Description string `gorm:"size:255"`
}

// AIDraft AI 生成的任务草稿，等待确认
type AIDraft struct {
	gorm.Model
	DraftID    string    `gorm:"size:100;uniqueIndex;not null"`
	UserID     uint      `gorm:"index"`
	Intent     string    `gorm:"size:50"`
	Data       string    `gorm:"type:text"`                 // 序列化的结构化数据
	Status     string    `gorm:"size:20;default:'pending'"` // pending, confirmed, expired
	ExpireTime time.Time `gorm:"index"`
}

// UserIdentity 跨平台身份统一映射
type UserIdentity struct {
	gorm.Model
	NexusUID    string `gorm:"size:100;uniqueIndex;not null"` // 统一的 Nexus ID
	Platform    string `gorm:"size:50;index;not null"`        // 平台 (qq, wechat, tg)
	PlatformUID string `gorm:"size:255;index;not null"`       // 平台内部 ID
	Nickname    string `gorm:"size:255"`
	Metadata    string `gorm:"type:text"` // 扩展属性 (积分、等级、偏好等)
}

// ShadowRule 影子执行与 A/B 测试规则
type ShadowRule struct {
	gorm.Model
	Name           string `gorm:"size:100;not null"`
	TargetWorkerID string `gorm:"size:100;not null"` // 影子 Worker ID
	MatchPattern   string `gorm:"size:255"`          // 匹配模式 (bot_*, group_*)
	IsEnabled      bool   `gorm:"default:false"`
	TrafficPercent int    `gorm:"default:0"` // 影子流量比例 (0-100)
}

// TaskTag 任务与标签的中间表
type TaskTag struct {
	TaskID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}
