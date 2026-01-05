package models

import (
	"time"

	"gorm.io/gorm"
)

// DigitalJob 定义具体的职位/工种 (Role/Position)
// 这是数字员工的社会身份，决定了它的职责范围和KPI标准
type DigitalJob struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"size:100;uniqueIndex;not null;column:name" json:"name"` // e.g. "Senior Go Developer", "Marketing Specialist"
	Description string         `gorm:"type:text;column:description" json:"description"`
	Level       int            `gorm:"default:1;column:level" json:"level"` // 职级 (1-10)
	Department  string         `gorm:"size:100;column:department" json:"department"`
	BaseSalary  int64          `gorm:"default:0;column:base_salary" json:"base_salary"` // 基础薪资 (Token/Day)
	IsActive    bool           `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (DigitalJob) TableName() string {
	return "DigitalJob"
}

// DigitalCapability 定义底层能力 (Capabilities)
// 能力是通用的、工具属性的，通常对应一个 MCP Server 或一组 API 调用权限
// 例如：FileSystemAccess, BrowserAccess, DockerControl
type DigitalCapability struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null;column:name" json:"name"` // e.g. "fs_read_write"
	Description string    `gorm:"type:text;column:description" json:"description"`
	Type        string    `gorm:"size:50;column:type" json:"type"`       // "mcp_server", "native_plugin", "api_scope"
	Target      string    `gorm:"size:255;column:target" json:"target"`  // MCP Server Name (e.g. "filesystem") or Plugin ID
	Config      string    `gorm:"type:text;column:config" json:"config"` // 默认配置 JSON (e.g. {"allowed_paths": ["/tmp"]})
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DigitalCapability) TableName() string {
	return "DigitalCapability"
}

// JobCapabilityRelation 职位与能力的关联 (职位要求)
// 定义了一个 Job 需要具备哪些底层能力
type JobCapabilityRelation struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	JobID          uint      `gorm:"index;column:job_id" json:"job_id"`
	CapabilityID   uint      `gorm:"index;column:capability_id" json:"capability_id"`
	IsRequired     bool      `gorm:"default:true;column:is_required" json:"is_required"`
	ConfigOverride string    `gorm:"type:text;column:config_override" json:"config_override"` // 针对该职位的特定配置覆盖
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (JobCapabilityRelation) TableName() string {
	return "JobCapabilityRelation"
}

// EmployeeSkillRelation 员工与技能的关联 (技能树)
// 记录员工对某个具体 Skill (AISkill表) 的熟练程度
// 这是“自进化”的核心数据：随着任务完成，Experience 和 Proficiency 会增加
type EmployeeSkillRelation struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EmployeeID  uint      `gorm:"index;column:employee_id" json:"employee_id"`      // 关联 DigitalEmployee
	SkillID     uint      `gorm:"index;column:skill_id" json:"skill_id"`            // 关联 AISkill
	Proficiency float64   `gorm:"default:0;column:proficiency" json:"proficiency"` // 0-100% 熟练度
	Experience  int64     `gorm:"default:0;column:experience" json:"experience"`   // 经验值 (XP)
	Level       int       `gorm:"default:1;column:level" json:"level"`             // 当前技能等级
	LastUsedAt  time.Time `gorm:"column:last_used_at" json:"last_used_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (EmployeeSkillRelation) TableName() string {
	return "EmployeeSkillRelation"
}

// EmployeeJobRelation 员工与职位的关联 (任职记录)
// 支持一人多职，或者职位变动历史
type EmployeeJobRelation struct {
	ID         uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EmployeeID uint       `gorm:"index;column:employee_id" json:"employee_id"`
	JobID      uint       `gorm:"index;column:job_id" json:"job_id"`
	IsPrimary  bool       `gorm:"default:false;column:is_primary" json:"is_primary"`
	AssignedAt time.Time  `gorm:"column:assigned_at" json:"assigned_at"`
	EndAt      *time.Time `gorm:"column:end_at" json:"end_at"`                           // 如果离职/转岗，则不为空
	Status     string     `gorm:"size:20;default:'active';column:status" json:"status"` // active, suspended, terminated
}

func (EmployeeJobRelation) TableName() string {
	return "EmployeeJobRelation"
}
