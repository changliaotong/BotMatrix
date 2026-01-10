package evolution

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SystemEvolution 系统进化模块
// 实现系统级别的自主进化能力
type SystemEvolution struct {
	mu          sync.RWMutex
	DB          *gorm.DB
	ID          string
	Name        string
	Description string
	Version     string
	Status      EvolutionStatus
	Iterations  []*SystemIteration
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SystemIteration 系统迭代
type SystemIteration struct {
	mu          sync.RWMutex
	ID          string
	Version     string
	Description string
	Status      IterationStatus
	Tasks       []*IterationTask
	Results     []*IterationResult
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

// IterationStatus 迭代状态
type IterationStatus string

const (
	IterationStatusCreated   IterationStatus = "created"
	IterationStatusRunning   IterationStatus = "running"
	IterationStatusCompleted IterationStatus = "completed"
	IterationStatusFailed    IterationStatus = "failed"
)

// IterationTask 迭代任务
type IterationTask struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Input       map[string]interface{} `json:"input"`
	Status      TaskStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusCreated   TaskStatus = "created"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// IterationResult 迭代结果
type IterationResult struct {
	ID            string                 `json:"id"`
	TaskID        string                 `json:"task_id"`
	Status        string                 `json:"status"`
	Output        map[string]interface{} `json:"output"`
	Effectiveness float64                `json:"effectiveness"`
	CreatedAt     time.Time              `json:"created_at"`
}

// DigitalEmployeeEvolution 数字员工进化
type DigitalEmployeeEvolution struct {
	mu        sync.RWMutex
	DB        *gorm.DB
	ID        string
	Name      string
	Type      string
	Prompt    string
	Skills    []string
	Version   string
	Status    EmployeeStatus
	UpdatedAt time.Time
}

// EmployeeStatus 员工状态
type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
)

// NewSystemEvolution 创建新的系统进化模块
func NewSystemEvolution(db *gorm.DB, name, description, version string) (*SystemEvolution, error) {
	se := &SystemEvolution{
		DB:          db,
		ID:          generateEvolutionID(),
		Name:        name,
		Description: description,
		Version:     version,
		Status:      EvolutionStatusCreated,
		Iterations:  []*SystemIteration{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 自动迁移表结构
	err := db.AutoMigrate(
		&SystemIteration{},
		&IterationTask{},
		&IterationResult{},
		&DigitalEmployeeEvolution{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// 保存系统进化模块
	err = db.Create(se).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create system evolution: %v", err)
	}

	return se, nil
}

// CreateIteration 创建系统迭代
func (se *SystemEvolution) CreateIteration(version, description string) (*SystemIteration, error) {
	iteration := &SystemIteration{
		ID:          generateIterationID(),
		Version:     version,
		Description: description,
		Status:      IterationStatusCreated,
		Tasks:       []*IterationTask{},
		Results:     []*IterationResult{},
		CreatedAt:   time.Now(),
	}

	err := se.DB.Create(iteration).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create iteration: %v", err)
	}

	se.mu.Lock()
	se.Iterations = append(se.Iterations, iteration)
	se.mu.Unlock()

	log.Printf("Created system iteration: %s (%s)", version, description)
	return iteration, nil
}

// StartIteration 启动系统迭代
func (se *SystemEvolution) StartIteration(iterationID string) error {
	var iteration SystemIteration
	err := se.DB.Where("id = ?", iterationID).First(&iteration).Error
	if err != nil {
		return fmt.Errorf("failed to find iteration: %v", err)
	}

	iteration.Status = IterationStatusRunning
	iteration.StartedAt = time.Now()
	err = se.DB.Save(&iteration).Error
	if err != nil {
		return fmt.Errorf("failed to start iteration: %v", err)
	}

	log.Printf("Started system iteration: %s", iteration.Version)
	return nil
}

// CompleteIteration 完成系统迭代
func (se *SystemEvolution) CompleteIteration(iterationID string) error {
	var iteration SystemIteration
	err := se.DB.Where("id = ?", iterationID).First(&iteration).Error
	if err != nil {
		return fmt.Errorf("failed to find iteration: %v", err)
	}

	iteration.Status = IterationStatusCompleted
	iteration.CompletedAt = time.Now()
	err = se.DB.Save(&iteration).Error
	if err != nil {
		return fmt.Errorf("failed to complete iteration: %v", err)
	}

	// 更新系统版本
	se.mu.Lock()
	se.Version = iteration.Version
	se.UpdatedAt = time.Now()
	se.mu.Unlock()

	err = se.DB.Save(se).Error
	if err != nil {
		return fmt.Errorf("failed to update system version: %v", err)
	}

	log.Printf("Completed system iteration: %s, new version: %s", iteration.Version, se.Version)
	return nil
}

// AddIterationTask 添加迭代任务
func (se *SystemEvolution) AddIterationTask(iterationID string, task *IterationTask) error {
	var iteration SystemIteration
	err := se.DB.Where("id = ?", iterationID).First(&iteration).Error
	if err != nil {
		return fmt.Errorf("failed to find iteration: %v", err)
	}

	if task.ID == "" {
		task.ID = generateTaskID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	err = se.DB.Create(task).Error
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	log.Printf("Added iteration task: %s", task.Name)
	return nil
}

// UpdateDigitalEmployee 更新数字员工
func (se *SystemEvolution) UpdateDigitalEmployee(employeeID string, prompt string, skills []string) error {
	var employee DigitalEmployeeEvolution
	err := se.DB.Where("id = ?", employeeID).First(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to find employee: %v", err)
	}

	employee.Prompt = prompt
	employee.Skills = skills
	employee.Version = fmt.Sprintf("v%d", time.Now().Unix())
	employee.UpdatedAt = time.Now()

	err = se.DB.Save(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to update employee: %v", err)
	}

	log.Printf("Updated digital employee: %s, new version: %s", employee.Name, employee.Version)
	return nil
}

// AddDigitalEmployeeSkill 为数字员工添加技能
func (se *SystemEvolution) AddDigitalEmployeeSkill(employeeID string, skill string) error {
	var employee DigitalEmployeeEvolution
	err := se.DB.Where("id = ?", employeeID).First(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to find employee: %v", err)
	}

	// 检查技能是否已存在
	for _, s := range employee.Skills {
		if s == skill {
			return fmt.Errorf("skill %s already exists", skill)
		}
	}

	employee.Skills = append(employee.Skills, skill)
	employee.Version = fmt.Sprintf("v%d", time.Now().Unix())
	employee.UpdatedAt = time.Now()

	err = se.DB.Save(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to add skill: %v", err)
	}

	log.Printf("Added skill %s to digital employee: %s", skill, employee.Name)
	return nil
}

// RemoveDigitalEmployeeSkill 移除数字员工的技能
func (se *SystemEvolution) RemoveDigitalEmployeeSkill(employeeID string, skill string) error {
	var employee DigitalEmployeeEvolution
	err := se.DB.Where("id = ?", employeeID).First(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to find employee: %v", err)
	}

	// 查找并移除技能
	for i, s := range employee.Skills {
		if s == skill {
			employee.Skills = append(employee.Skills[:i], employee.Skills[i+1:]...)
			break
		}
	}

	employee.Version = fmt.Sprintf("v%d", time.Now().Unix())
	employee.UpdatedAt = time.Now()

	err = se.DB.Save(&employee).Error
	if err != nil {
		return fmt.Errorf("failed to remove skill: %v", err)
	}

	log.Printf("Removed skill %s from digital employee: %s", skill, employee.Name)
	return nil
}

// CreateDigitalEmployee 创建数字员工
func (se *SystemEvolution) CreateDigitalEmployee(name, employeeType, prompt string, skills []string) (*DigitalEmployeeEvolution, error) {
	employee := &DigitalEmployeeEvolution{
		ID:        generateEmployeeID(),
		Name:      name,
		Type:      employeeType,
		Prompt:    prompt,
		Skills:    skills,
		Version:   fmt.Sprintf("v%d", time.Now().Unix()),
		Status:    EmployeeStatusActive,
		UpdatedAt: time.Now(),
	}

	err := se.DB.Create(employee).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %v", err)
	}

	log.Printf("Created digital employee: %s (%s)", name, employeeType)
	return employee, nil
}

// generateIterationID 生成迭代ID
func generateIterationID() string {
	return fmt.Sprintf("iteration_%d", time.Now().UnixNano())
}

// generateEmployeeID 生成员工ID
func generateEmployeeID() string {
	return fmt.Sprintf("employee_%d", time.Now().UnixNano())
}
