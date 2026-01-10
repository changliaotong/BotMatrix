package evolution

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// SelfEvolution 自主进化系统
// 实现开发团队的自主进化能力
type SelfEvolution struct {
	mu          sync.RWMutex
	ID          string
	Name        string
	Description string
	Status      EvolutionStatus
	Agents      []EvolutionAgent
	FeedbackLoop *FeedbackLoop
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EvolutionStatus 进化状态
type EvolutionStatus string

const (
	EvolutionStatusCreated  EvolutionStatus = "created"
	EvolutionStatusRunning  EvolutionStatus = "running"
	EvolutionStatusPaused   EvolutionStatus = "paused"
	EvolutionStatusCompleted EvolutionStatus = "completed"
)

// EvolutionAgent 进化Agent
type EvolutionAgent interface {
	// 获取Agent ID
	GetID() string
	// 获取Agent名称
	GetName() string
	// 获取Agent类型
	GetType() string
	// 执行进化任务
	ExecuteEvolutionTask(task EvolutionTask) (EvolutionResult, error)
	// 获取Agent状态
	GetStatus() AgentStatus
	// 设置Agent状态
	SetStatus(status AgentStatus) error
}

// AgentStatus Agent状态
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusError   AgentStatus = "error"
)

// EvolutionTask 进化任务
type EvolutionTask struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Priority    int               `json:"priority"`
	CreatedAt   time.Time         `json:"created_at"`
}

// EvolutionResult 进化结果
type EvolutionResult struct {
	ID          string            `json:"id"`
	TaskID      string            `json:"task_id"`
	Status      string            `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Effectiveness float64         `json:"effectiveness"`
	CompletedAt time.Time         `json:"completed_at"`
}

// FeedbackLoop 反馈循环
type FeedbackLoop struct {
	mu          sync.RWMutex
	ID          string
	Name        string
	Description string
	Status      FeedbackLoopStatus
	Tasks       []*EvolutionTask
	Results     []*EvolutionResult
	CreatedAt   time.Time
}

// FeedbackLoopStatus 反馈循环状态
type FeedbackLoopStatus string

const (
	FeedbackLoopStatusActive  FeedbackLoopStatus = "active"
	FeedbackLoopStatusInactive FeedbackLoopStatus = "inactive"
)

// NewSelfEvolution 创建新的自主进化系统
func NewSelfEvolution(name, description string) *SelfEvolution {
	return &SelfEvolution{
		ID:          generateEvolutionID(),
		Name:        name,
		Description: description,
		Status:      EvolutionStatusCreated,
		Agents:      []EvolutionAgent{},
		FeedbackLoop: &FeedbackLoop{
			ID:          generateFeedbackLoopID(),
			Name:        "Main Feedback Loop",
			Description: "Main feedback loop for self-evolution",
			Status:      FeedbackLoopStatusActive,
			CreatedAt:   time.Now(),
		},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddAgent 添加进化Agent
func (se *SelfEvolution) AddAgent(agent EvolutionAgent) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	se.Agents = append(se.Agents, agent)
	log.Printf("Added evolution agent: %s (%s)", agent.GetName(), agent.GetType())
	return nil
}

// Start 启动自主进化系统
func (se *SelfEvolution) Start() error {
	se.mu.Lock()
	se.Status = EvolutionStatusRunning
	se.UpdatedAt = time.Now()
	se.mu.Unlock()

	log.Printf("Started self-evolution system: %s", se.Name)

	// 启动反馈循环
	go se.runFeedbackLoop()

	return nil
}

// runFeedbackLoop 运行反馈循环
func (se *SelfEvolution) runFeedbackLoop() {
	for {
		se.mu.RLock()
		status := se.Status
		feedbackStatus := se.FeedbackLoop.Status
		se.mu.RUnlock()

		if status != EvolutionStatusRunning || feedbackStatus != FeedbackLoopStatusActive {
			break
		}

		// 处理任务
		se.processTasks()

		// 等待一段时间
		time.Sleep(10 * time.Second)
	}
}

// processTasks 处理进化任务
func (se *SelfEvolution) processTasks() {
	se.mu.Lock()
	defer se.mu.Unlock()

	for _, task := range se.FeedbackLoop.Tasks {
		// 找到合适的Agent执行任务
		agent := se.findSuitableAgent(task)
		if agent == nil {
			log.Printf("No suitable agent for task: %s", task.Description)
			continue
		}

		// 执行任务
		result, err := agent.ExecuteEvolutionTask(task)
		if err != nil {
			log.Printf("Agent %s failed to execute task %s: %v", agent.GetName(), task.ID, err)
			continue
		}

		// 保存结果
		se.FeedbackLoop.Results = append(se.FeedbackLoop.Results, &result)

		log.Printf("Agent %s completed task %s with effectiveness %.2f", agent.GetName(), task.ID, result.Effectiveness)
	}

	// 清空已处理的任务
	se.FeedbackLoop.Tasks = []*EvolutionTask{}
}

// findSuitableAgent 找到合适的Agent
func (se *SelfEvolution) findSuitableAgent(task EvolutionTask) EvolutionAgent {
	// 简单的匹配逻辑
	// 实际应用中可以根据Agent技能和任务类型进行匹配
	for _, agent := range se.Agents {
		if agent.GetStatus() == AgentStatusIdle {
			return agent
		}
	}
	return nil
}

// AddTask 添加进化任务
func (se *SelfEvolution) AddTask(task EvolutionTask) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	se.FeedbackLoop.Tasks = append(se.FeedbackLoop.Tasks, &task)
	log.Printf("Added evolution task: %s", task.Description)
	return nil
}

// Stop 停止自主进化系统
func (se *SelfEvolution) Stop() error {
	se.mu.Lock()
	se.Status = EvolutionStatusPaused
	se.FeedbackLoop.Status = FeedbackLoopStatusInactive
	se.UpdatedAt = time.Now()
	se.mu.Unlock()

	log.Printf("Stopped self-evolution system: %s", se.Name)
	return nil
}

// generateEvolutionID 生成进化ID
func generateEvolutionID() string {
	return fmt.Sprintf("evolution_%d", time.Now().UnixNano())
}

// generateFeedbackLoopID 生成反馈循环ID
func generateFeedbackLoopID() string {
	return fmt.Sprintf("feedback_loop_%d", time.Now().UnixNano())
}