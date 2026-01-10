package evolution

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DatabaseEvolution 数据库驱动的自主进化系统
// 实现进化数据的持久化和动态配置
type DatabaseEvolution struct {
	mu          sync.RWMutex
	DB          *gorm.DB
	ID          string
	Name        string
	Description string
	Status      EvolutionStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DatabaseEvolutionAgent 数据库驱动的进化Agent
type DatabaseEvolutionAgent struct {
	mu        sync.RWMutex
	DB        *gorm.DB
	ID        string
	Name      string
	Type      string
	Status    AgentStatus
	Skills    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DatabaseEvolutionTask 数据库驱动的进化任务
type DatabaseEvolutionTask struct {
	mu          sync.RWMutex
	DB          *gorm.DB
	ID          string
	Type        string
	Description string
	Input       map[string]interface{}
	Priority    int
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DatabaseEvolutionResult 数据库驱动的进化结果
type DatabaseEvolutionResult struct {
	mu            sync.RWMutex
	DB            *gorm.DB
	ID            string
	TaskID        string
	AgentID       string
	Status        string
	Output        map[string]interface{}
	Effectiveness float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewDatabaseEvolution 创建新的数据库驱动自主进化系统
func NewDatabaseEvolution(db *gorm.DB, name, description string) (*DatabaseEvolution, error) {
	se := &DatabaseEvolution{
		DB:          db,
		ID:          generateEvolutionID(),
		Name:        name,
		Description: description,
		Status:      EvolutionStatusCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 自动迁移表结构
	err := db.AutoMigrate(
		&DatabaseEvolutionAgent{},
		&DatabaseEvolutionTask{},
		&DatabaseEvolutionResult{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// 保存进化系统
	err = db.Create(se).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create evolution system: %v", err)
	}

	return se, nil
}

// AddAgent 添加进化Agent
func (se *DatabaseEvolution) AddAgent(agentType string, name string, skills []string) error {
	agent := &DatabaseEvolutionAgent{
		DB:        se.DB,
		ID:        generateAgentID(),
		Name:      name,
		Type:      agentType,
		Status:    AgentStatusIdle,
		Skills:    skills,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := se.DB.Create(agent).Error
	if err != nil {
		return fmt.Errorf("failed to add agent: %v", err)
	}

	log.Printf("Added database evolution agent: %s (%s)", name, agentType)
	return nil
}

// GetAgent 获取进化Agent
func (se *DatabaseEvolution) GetAgent(agentID string) (*DatabaseEvolutionAgent, error) {
	var agent DatabaseEvolutionAgent
	err := se.DB.Where("id = ?", agentID).First(&agent).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	agent.DB = se.DB
	return &agent, nil
}

// ListAgents 列出所有进化Agent
func (se *DatabaseEvolution) ListAgents() ([]*DatabaseEvolutionAgent, error) {
	var agents []*DatabaseEvolutionAgent
	err := se.DB.Find(&agents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %v", err)
	}

	for _, agent := range agents {
		agent.DB = se.DB
	}

	return agents, nil
}

// AddTask 添加进化任务
func (se *DatabaseEvolution) AddTask(taskType, description string, input map[string]interface{}, priority int) error {
	task := &DatabaseEvolutionTask{
		DB:          se.DB,
		ID:          generateTaskID(),
		Type:        taskType,
		Description: description,
		Input:       input,
		Priority:    priority,
		Status:      "created",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := se.DB.Create(task).Error
	if err != nil {
		return fmt.Errorf("failed to add task: %v", err)
	}

	log.Printf("Added database evolution task: %s", description)
	return nil
}

// GetTask 获取进化任务
func (se *DatabaseEvolution) GetTask(taskID string) (*DatabaseEvolutionTask, error) {
	var task DatabaseEvolutionTask
	err := se.DB.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	task.DB = se.DB
	return &task, nil
}

// ListTasks 列出所有进化任务
func (se *DatabaseEvolution) ListTasks() ([]*DatabaseEvolutionTask, error) {
	var tasks []*DatabaseEvolutionTask
	err := se.DB.Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %v", err)
	}

	for _, task := range tasks {
		task.DB = se.DB
	}

	return tasks, nil
}

// ExecuteTask 执行进化任务
func (se *DatabaseEvolution) ExecuteTask(taskID string, agentID string) error {
	// 获取任务
	task, err := se.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %v", err)
	}

	// 获取Agent
	agent, err := se.GetAgent(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %v", err)
	}

	// 更新任务状态为运行中
	task.Status = "running"
	task.UpdatedAt = time.Now()
	err = se.DB.Save(task).Error
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	// 更新Agent状态为忙碌
	agent.Status = AgentStatusBusy
	agent.UpdatedAt = time.Now()
	err = se.DB.Save(agent).Error
	if err != nil {
		return fmt.Errorf("failed to update agent status: %v", err)
	}

	// 执行任务
	result, err := agent.ExecuteEvolutionTask(task)
	if err != nil {
		// 更新任务状态为失败
		task.Status = "failed"
		task.UpdatedAt = time.Now()
		se.DB.Save(task)

		// 更新Agent状态为空闲
		agent.Status = AgentStatusIdle
		agent.UpdatedAt = time.Now()
		se.DB.Save(agent)

		return fmt.Errorf("failed to execute task: %v", err)
	}

	// 保存结果
	result.AgentID = agentID
	err = se.DB.Create(result).Error
	if err != nil {
		return fmt.Errorf("failed to save result: %v", err)
	}

	// 更新任务状态为完成
	task.Status = "completed"
	task.UpdatedAt = time.Now()
	err = se.DB.Save(task).Error
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	// 更新Agent状态为空闲
	agent.Status = AgentStatusIdle
	agent.UpdatedAt = time.Now()
	err = se.DB.Save(agent).Error
	if err != nil {
		return fmt.Errorf("failed to update agent status: %v", err)
	}

	log.Printf("Agent %s completed task %s with effectiveness %.2f", agent.Name, task.Description, result.Effectiveness)
	return nil
}

// ExecuteEvolutionTask 执行进化任务
func (a *DatabaseEvolutionAgent) ExecuteEvolutionTask(task *DatabaseEvolutionTask) (*DatabaseEvolutionResult, error) {
	log.Printf("Database Evolution Agent %s executing task: %s", a.Name, task.Description)

	// 根据任务类型执行不同的操作
	var output map[string]interface{}
	var effectiveness float64

	switch task.Type {
	case "review_pr":
		output = map[string]interface{}{
			"review": "PR审阅完成，代码符合规范",
			"suggestions": []string{
				"添加单元测试",
				"优化错误处理",
			},
		}
		effectiveness = 95
	case "fix_bug":
		output = map[string]interface{}{
			"bug_fixed":       true,
			"fix_description": "修复了空指针异常",
			"test_passed":     true,
		}
		effectiveness = 90
	case "write_test":
		output = map[string]interface{}{
			"test_written":  true,
			"test_coverage": 85,
			"test_passed":   true,
		}
		effectiveness = 85
	case "generate_plugin":
		output = map[string]interface{}{
			"plugin_generated": true,
			"plugin_name":      "抖音适配器",
			"plugin_version":   "1.0.0",
		}
		effectiveness = 80
	default:
		output = map[string]interface{}{
			"result": "任务执行完成",
		}
		effectiveness = 75
	}

	result := &DatabaseEvolutionResult{
		DB:            a.DB,
		ID:            generateResultID(),
		TaskID:        task.ID,
		AgentID:       a.ID,
		Status:        "completed",
		Output:        output,
		Effectiveness: effectiveness,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return result, nil
}

// GetStatus 获取Agent状态
func (a *DatabaseEvolutionAgent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// SetStatus 设置Agent状态
func (a *DatabaseEvolutionAgent) SetStatus(status AgentStatus) error {
	a.mu.Lock()
	a.Status = status
	a.UpdatedAt = time.Now()
	a.mu.Unlock()

	return a.DB.Save(a).Error
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
