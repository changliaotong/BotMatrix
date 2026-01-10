package models

import (
	"time"

	"gorm.io/gorm"
)

// DigitalRole represents a digital development team role
// 数字开发团队角色模型
type DigitalRole struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName    string         `gorm:"unique;not null;size:100" json:"role_name"` // 角色名称（如：架构师、程序员）
	Description string         `gorm:"type:text" json:"description"` // 角色描述
	Skills      []string       `gorm:"type:jsonb" json:"skills"` // 角色技能列表
	Experience  int            `gorm:"default:0" json:"experience"` // 经验值
	Level       int            `gorm:"default:1" json:"level"` // 角色等级
	Prompt      string         `gorm:"type:text" json:"prompt"` // AI提示词模板
	Config      map[string]interface{} `gorm:"type:jsonb" json:"config"` // 角色配置
	IsActive    bool           `gorm:"default:true" json:"is_active"` // 是否启用
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DigitalTask represents a task for the digital team
// 数字团队任务模型
type DigitalTask struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskName    string         `gorm:"not null;size:200" json:"task_name"` // 任务名称
	TaskType    string         `gorm:"not null;size:50" json:"task_type"` // 任务类型（如：设计、编码、测试）
	Description string         `gorm:"type:text" json:"description"` // 任务描述
	InputData   map[string]interface{} `gorm:"type:jsonb" json:"input_data"` // 任务输入数据
	OutputData  map[string]interface{} `gorm:"type:jsonb" json:"output_data"` // 任务输出数据
	Status      string         `gorm:"default:pending;size:20" json:"status"` // 任务状态（pending, in_progress, completed, failed）
	AssignedTo  string         `gorm:"size:100" json:"assigned_to"` // 分配的角色
	Priority    int            `gorm:"default:1" json:"priority"` // 任务优先级（1-5）
	ExecutionTime float64      `gorm:"default:0" json:"execution_time"` // 执行时间（秒）
	ErrorMsg    string         `gorm:"type:text" json:"error_msg"` // 错误信息
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
}

// DigitalProject represents a project managed by the digital team
// 数字团队项目模型
type DigitalProject struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectName string         `gorm:"not null;size:200" json:"project_name"` // 项目名称
	Description string         `gorm:"type:text" json:"description"` // 项目描述
	Requirements string        `gorm:"type:text" json:"requirements"` // 项目需求
	TechStack   []string       `gorm:"type:jsonb" json:"tech_stack"` // 技术栈
	Status      string         `gorm:"default:pending;size:20" json:"status"` // 项目状态
	Progress    float64        `gorm:"default:0" json:"progress"` // 项目进度（0-100）
	Results     map[string]interface{} `gorm:"type:jsonb" json:"results"` // 项目结果
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
}

// DigitalEvolution represents the self-evolution record of the digital team
// 数字团队自我进化记录模型
type DigitalEvolution struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID      uint           `gorm:"index" json:"role_id"` // 角色ID
	RoleName    string         `gorm:"size:100" json:"role_name"` // 角色名称
	EvolutionType string       `gorm:"size:50" json:"evolution_type"` // 进化类型（如：学习、优化、修复）
	Description string         `gorm:"type:text" json:"description"` // 进化描述
	BeforeData  map[string]interface{} `gorm:"type:jsonb" json:"before_data"` // 进化前数据
	AfterData   map[string]interface{} `gorm:"type:jsonb" json:"after_data"` // 进化后数据
	Effectiveness float64      `gorm:"default:0" json:"effectiveness"` // 进化效果（0-100）
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
}

// TableName methods

func (DigitalRole) TableName() string {
	return "digital_roles"
}

func (DigitalTask) TableName() string {
	return "digital_tasks"
}

func (DigitalProject) TableName() string {
	return "digital_projects"
}

func (DigitalEvolution) TableName() string {
	return "digital_evolutions"
}