package collaboration

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Workflow 通用团队协作流程
// 实现数字员工之间的协作流程管理

type Workflow struct {
	mu          sync.RWMutex
	ID          string
	Name        string
	Description string
	Status      WorkflowStatus
	Steps       []*WorkflowStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusCreated  WorkflowStatus = "created"
	WorkflowStatusRunning  WorkflowStatus = "running"
	WorkflowStatusPaused   WorkflowStatus = "paused"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed   WorkflowStatus = "failed"
)

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID          string
	Name        string
	Description string
	Type        WorkflowStepType
	RoleType    string
	Status      WorkflowStepStatus
	Input       map[string]interface{}
	Output      map[string]interface{}
	Dependencies []string // 依赖的步骤ID
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

// WorkflowStepType 工作流步骤类型
type WorkflowStepType string

const (
	WorkflowStepTypeTask       WorkflowStepType = "task"
	WorkflowStepTypeDecision   WorkflowStepType = "decision"
	WorkflowStepTypeParallel   WorkflowStepType = "parallel"
	WorkflowStepTypeSequential WorkflowStepType = "sequential"
)

// WorkflowStepStatus 工作流步骤状态
type WorkflowStepStatus string

const (
	WorkflowStepStatusCreated  WorkflowStepStatus = "created"
	WorkflowStepStatusRunning  WorkflowStepStatus = "running"
	WorkflowStepStatusCompleted WorkflowStepStatus = "completed"
	WorkflowStepStatusFailed   WorkflowStepStatus = "failed"
	WorkflowStepStatusSkipped  WorkflowStepStatus = "skipped"
)

// WorkflowManager 工作流管理器
type WorkflowManager struct {
	mu          sync.RWMutex
	messageBus  MessageBus
	workflows   map[string]*Workflow
}

// NewWorkflowManager 创建新的工作流管理器
func NewWorkflowManager(messageBus MessageBus) *WorkflowManager {
	return &WorkflowManager{
		messageBus: messageBus,
		workflows:  make(map[string]*Workflow),
	}
}

// CreateWorkflow 创建新的工作流
func (wm *WorkflowManager) CreateWorkflow(name, description string, steps []*WorkflowStep) (*Workflow, error) {
	workflow := &Workflow{
		ID:          generateWorkflowID(),
		Name:        name,
		Description: description,
		Status:      WorkflowStatusCreated,
		Steps:       steps,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	wm.mu.Lock()
	wm.workflows[workflow.ID] = workflow
	wm.mu.Unlock()

	log.Printf("Created workflow: %s (ID: %s)", name, workflow.ID)
	return workflow, nil
}

// StartWorkflow 启动工作流
func (wm *WorkflowManager) StartWorkflow(workflowID string) error {
	wm.mu.RLock()
	workflow, exists := wm.workflows[workflowID]
	wm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("workflow with ID %s not found", workflowID)
	}

	workflow.Status = WorkflowStatusRunning
	workflow.UpdatedAt = time.Now()

	log.Printf("Started workflow: %s (ID: %s)", workflow.Name, workflow.ID)

	// 异步执行工作流
	go wm.executeWorkflow(workflow)

	return nil
}

// executeWorkflow 执行工作流
func (wm *WorkflowManager) executeWorkflow(workflow *Workflow) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Workflow execution failed: %v", r)
			workflow.Status = WorkflowStatusFailed
			workflow.UpdatedAt = time.Now()
		}
	}()

	// 拓扑排序步骤
	orderedSteps, err := wm.topologicalSortSteps(workflow.Steps)
	if err != nil {
		log.Printf("Failed to sort workflow steps: %v", err)
		workflow.Status = WorkflowStatusFailed
		workflow.UpdatedAt = time.Now()
		return
	}

	// 执行步骤
	for _, step := range orderedSteps {
		if workflow.Status != WorkflowStatusRunning {
			break
		}

		if err := wm.executeStep(workflow, step); err != nil {
			log.Printf("Step %s failed: %v", step.Name, err)
			step.Status = WorkflowStepStatusFailed
			workflow.Status = WorkflowStatusFailed
			workflow.UpdatedAt = time.Now()
			return
		}

		step.Status = WorkflowStepStatusCompleted
		step.CompletedAt = time.Now()
		workflow.UpdatedAt = time.Now()

		log.Printf("Completed step: %s (ID: %s)", step.Name, step.ID)
	}

	workflow.Status = WorkflowStatusCompleted
	workflow.UpdatedAt = time.Now()
	log.Printf("Workflow completed: %s (ID: %s)", workflow.Name, workflow.ID)
}

// executeStep 执行工作流步骤
func (wm *WorkflowManager) executeStep(workflow *Workflow, step *WorkflowStep) error {
	step.Status = WorkflowStepStatusRunning
	step.StartedAt = time.Now()

	log.Printf("Executing step: %s (ID: %s)", step.Name, step.ID)

	// 根据步骤类型执行
	switch step.Type {
	case WorkflowStepTypeTask:
		return wm.executeTaskStep(workflow, step)
	case WorkflowStepTypeDecision:
		return wm.executeDecisionStep(workflow, step)
	case WorkflowStepTypeParallel:
		return wm.executeParallelStep(workflow, step)
	case WorkflowStepTypeSequential:
		return wm.executeSequentialStep(workflow, step)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeTaskStep 执行任务步骤
func (wm *WorkflowManager) executeTaskStep(workflow *Workflow, step *WorkflowStep) error {
	// 创建任务
	task := Task{
		ID:          generateTaskID(),
		Type:        step.RoleType,
		Description: step.Description,
		Priority:    PriorityHigh,
		Input:       step.Input,
	}

	// 分配任务
	// 这里需要集成任务分配器
	// 暂时简化处理
	log.Printf("Task step executed: %s", step.Name)
	return nil
}

// executeDecisionStep 执行决策步骤
func (wm *WorkflowManager) executeDecisionStep(workflow *Workflow, step *WorkflowStep) error {
	// 简单的决策逻辑
	// 实际应用中可以根据输入进行复杂决策
	log.Printf("Decision step executed: %s", step.Name)
	return nil
}

// executeParallelStep 执行并行步骤
func (wm *WorkflowManager) executeParallelStep(workflow *Workflow, step *WorkflowStep) error {
	// 并行执行多个子步骤
	log.Printf("Parallel step executed: %s", step.Name)
	return nil
}

// executeSequentialStep 执行顺序步骤
func (wm *WorkflowManager) executeSequentialStep(workflow *Workflow, step *WorkflowStep) error {
	// 顺序执行多个子步骤
	log.Printf("Sequential step executed: %s", step.Name)
	return nil
}

// topologicalSortSteps 拓扑排序步骤
func (wm *WorkflowManager) topologicalSortSteps(steps []*WorkflowStep) ([]*WorkflowStep, error) {
	// 简单的拓扑排序实现
	// 实际应用中需要处理循环依赖
	return steps, nil
}

// generateWorkflowID 生成工作流ID
func generateWorkflowID() string {
	return fmt.Sprintf("workflow_%d", time.Now().UnixNano())
}