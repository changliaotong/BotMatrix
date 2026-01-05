package models

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
	ID            uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Name          string         `gorm:"size:255;not null;column:name" json:"name"`
	Type          string         `gorm:"size:50;not null;column:type" json:"type"`               // once, cron, delayed, condition
	ActionType    string         `gorm:"size:50;not null;column:action_type" json:"action_type"` // send_message, mute_group, unmute_group
	ActionParams  string         `gorm:"type:text;column:action_params" json:"action_params"`    // JSON 参数
	TriggerConfig string         `gorm:"type:text;column:trigger_config" json:"trigger_config"`  // JSON 触发配置 (cron, delay, conditions)
	Status        TaskStatus     `gorm:"size:50;default:'pending';column:status" json:"status"`
	CreatorID     uint           `gorm:"index;column:creator_id" json:"creator_id"`
	IsEnterprise  bool           `gorm:"default:false;column:is_enterprise" json:"is_enterprise"`
	Tags          []Tag          `gorm:"many2many:task_tags;foreignKey:ID;joinForeignKey:TaskID;References:ID;joinReferences:TagID" json:"tags"`
	Executions    []Execution    `gorm:"foreignKey:TaskID" json:"executions"`
	LastRunTime   *time.Time     `gorm:"column:last_run_time" json:"last_run_time"`
	NextRunTime   *time.Time     `gorm:"index;column:next_run_time" json:"next_run_time"`
}

func (Task) TableName() string {
	return "tasks"
}

// Execution 任务执行实例
type Execution struct {
	ID            uint            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt     time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index;column:deleted_at" json:"deleted_at"`
	TaskID        uint            `gorm:"index;not null;column:task_id" json:"task_id"`
	ExecutionID   string          `gorm:"size:100;uniqueIndex;not null;column:execution_id" json:"execution_id"`
	TriggerTime   time.Time       `gorm:"index;column:trigger_time" json:"trigger_time"`
	ActualTime    *time.Time      `gorm:"column:actual_time" json:"actual_time"`
	Status        ExecutionStatus `gorm:"size:50;default:'pending';column:status" json:"status"`
	Result        string          `gorm:"type:text;column:result" json:"result"` // JSON 结果或错误详情
	RetryCount    int             `gorm:"default:0;column:retry_count" json:"retry_count"`
	MaxRetries    int             `gorm:"default:3;column:max_retries" json:"max_retries"`
	NextRetryTime *time.Time      `gorm:"index;column:next_retry_time" json:"next_retry_time"`
	TraceID       string          `gorm:"size:100;column:trace_id" json:"trace_id"`
}

func (Execution) TableName() string {
	return "task_executions"
}

// Tag 标签定义
type Tag struct {
	ID         uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Name       string         `gorm:"size:100;not null;index;column:name" json:"name"`
	Type       string         `gorm:"size:50;not null;column:type" json:"type"` // group, friend
	TargetID   string         `gorm:"size:255;not null;index;column:target_id" json:"target_id"`
	IsInternal bool           `gorm:"default:false;column:is_internal" json:"is_internal"` // 系统内置标签
}

func (Tag) TableName() string {
	return "task_tags_def"
}

// Strategy 全局策略定义
type Strategy struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Name        string         `gorm:"size:100;not null;uniqueIndex;column:name" json:"name"`
	Type        string         `gorm:"size:50;column:type" json:"type"`       // rate_limit, maintenance, flow_control
	Config      string         `gorm:"type:text;column:config" json:"config"` // JSON 配置
	IsEnabled   bool           `gorm:"default:false;column:is_enabled" json:"is_enabled"`
	Description string         `gorm:"size:255;column:description" json:"description"`
}

func (Strategy) TableName() string {
	return "task_strategies"
}

// AIDraft AI 生成的任务草稿，等待确认
type AIDraft struct {
	ID         uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	DraftID    string         `gorm:"size:100;uniqueIndex;not null;column:draft_id" json:"draft_id"`
	UserID     uint           `gorm:"index;column:user_id" json:"user_id"`
	GroupID    string         `gorm:"size:100;column:group_id" json:"group_id"`              // 执行上下文：群号
	UserRole   string         `gorm:"size:50;column:user_role" json:"user_role"`             // 执行上下文：角色
	Intent     string         `gorm:"size:50;column:intent" json:"intent"`
	Data       string         `gorm:"type:text;column:data" json:"data"`                     // 序列化的结构化数据
	Status     string         `gorm:"size:20;default:'pending';column:status" json:"status"` // pending, confirmed, expired, rejected
	ExpireTime time.Time      `gorm:"index;column:expire_time" json:"expire_time"`
}

func (AIDraft) TableName() string {
	return "task_ai_drafts"
}

// UserIdentity 跨平台身份统一映射
type UserIdentity struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	NexusUID    string         `gorm:"size:100;uniqueIndex;not null;column:nexus_uid" json:"nexus_uid"` // 统一的 Nexus ID
	Platform    string         `gorm:"size:50;index;not null;column:platform" json:"platform"`          // 平台 (qq, wechat, tg)
	PlatformUID string         `gorm:"size:255;index;not null;column:platform_uid" json:"platform_uid"` // 平台内部 ID
	Nickname    string         `gorm:"size:255;column:nickname" json:"nickname"`
	Metadata    string         `gorm:"type:text;column:metadata" json:"metadata"` // 扩展属性 (积分、等级、偏好等)
}

func (UserIdentity) TableName() string {
	return "user_identities"
}

// ShadowRule 影子执行与 A/B 测试规则
type ShadowRule struct {
	ID             uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt      time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Name           string         `gorm:"size:100;not null;column:name" json:"name"`
	TargetWorkerID string         `gorm:"size:100;not null;column:target_worker_id" json:"target_worker_id"` // 影子 Worker ID
	MatchPattern   string         `gorm:"size:255;column:match_pattern" json:"match_pattern"`                // 匹配模式 (bot_*, group_*)
	IsEnabled      bool           `gorm:"default:false;column:is_enabled" json:"is_enabled"`
	TrafficPercent int            `gorm:"default:0;column:traffic_percent" json:"traffic_percent"` // 影子流量比例 (0-100)
}

func (ShadowRule) TableName() string {
	return "task_shadow_rules"
}

// TaskTag 任务与标签的中间表
type TaskTag struct {
	TaskID uint `gorm:"primaryKey;column:task_id" json:"task_id"`
	TagID  uint `gorm:"primaryKey;column:tag_id" json:"tag_id"`
}

func (TaskTag) TableName() string {
	return "task_tag_mapping"
}
