package models

import (
	"time"

	"gorm.io/gorm"
)

// ABExperiment 定义一个 A/B 测试实验
// 用于对比不同设定（Prompt, Model, Parameters, Memory）对数字员工绩效的影响
type ABExperiment struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"size:100;not null;column:name" json:"name"`
	Description string         `gorm:"type:text;column:description" json:"description"`
	TargetJobID uint           `gorm:"index;column:target_job_id" json:"target_job_id"`       // 测试针对的目标岗位
	Status      string         `gorm:"size:20;default:'planned';column:status" json:"status"` // planned, running, completed, halted
	Metric      string         `gorm:"size:50;column:metric" json:"metric"`                   // 核心考核指标: kpi_score, revenue, task_completion_rate
	StartDate   *time.Time     `gorm:"column:start_date" json:"start_date"`
	EndDate     *time.Time     `gorm:"column:end_date" json:"end_date"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (ABExperiment) TableName() string {
	return "ABExperiment"
}

// ABVariant 定义实验的一个变体 (Group A/B/C)
// 变体存储了相对于基准 Job 的配置差异 (Delta)
type ABVariant struct {
	ID           uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ExperimentID uint   `gorm:"index;column:experiment_id" json:"experiment_id"`
	Name         string `gorm:"size:100;not null;column:name" json:"name"` // e.g. "Control Group", "High Temp Group", "Aggressive Persona"
	Description  string `gorm:"size:255;column:description" json:"description"`

	// ConfigOverride 存储 JSON 格式的配置覆盖
	// 支持覆盖: Prompt, Temperature, ModelID, Tools, InitMemorySet
	// e.g. {"temperature": 0.9, "prompt_suffix": "Be very concise."}
	ConfigOverride string `gorm:"type:text;column:config_override" json:"config_override"`

	Allocation  int       `gorm:"default:0;column:allocation" json:"allocation"`     // 流量/实例分配权重 (0-100)
	SampleCount int       `gorm:"default:0;column:sample_count" json:"sample_count"` // 当前已生成的实例数量
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ABVariant) TableName() string {
	return "ABVariant"
}

// EmployeeExperimentRelation 记录员工参与的实验
type EmployeeExperimentRelation struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EmployeeID   uint      `gorm:"index;column:EmployeeId" json:"employee_id"`
	ExperimentID uint      `gorm:"index;column:ExperimentId" json:"experiment_id"`
	VariantID    uint      `gorm:"index;column:VariantId" json:"variant_id"`
	JoinedAt     time.Time `gorm:"column:JoinedAt" json:"joined_at"`
	FinalScore   float64   `gorm:"column:FinalScore" json:"final_score"` // 实验结束时的评分
}

func (EmployeeExperimentRelation) TableName() string {
	return "EmployeeExperimentRelation"
}

// MemorySnapshot 记忆快照 (用于克隆高绩效员工的记忆)
type MemorySnapshot struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SourceBotID string    `gorm:"index;size:64;column:SourceBotId" json:"source_bot_id"`
	Name        string    `gorm:"size:100;column:Name" json:"name"`
	Description string    `gorm:"size:255;column:Description" json:"description"`
	SnapshotAt  time.Time `gorm:"column:SnapshotAt" json:"snapshot_at"`
	MemoryCount int       `gorm:"column:MemoryCount" json:"memory_count"`
	// 实际上我们可能不需要物理复制 CognitiveMemory 表的所有行，
	// 而是通过 Tag 或 Reference 来让新员工“挂载”这份记忆。
	// 但为了独立性，这里假设是物理快照的元数据。
}

func (MemorySnapshot) TableName() string {
	return "MemorySnapshot"
}
