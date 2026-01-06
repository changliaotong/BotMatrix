package employee

import (
	// "BotMatrix/common/ai" // Removed to break import cycle

	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskServiceImpl struct {
	db          *gorm.DB
	employeeSvc DigitalEmployeeService
	redisMgr    *database.RedisManager
}

var _ DigitalEmployeeTaskService = (*TaskServiceImpl)(nil)

func NewTaskService(db *gorm.DB, employeeSvc DigitalEmployeeService, redisMgr *database.RedisManager) *TaskServiceImpl {
	return &TaskServiceImpl{
		db:          db,
		employeeSvc: employeeSvc,
		redisMgr:    redisMgr,
	}
}

func (s *TaskServiceImpl) CreateTask(ctx context.Context, task *models.DigitalEmployeeTask) error {
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.ExecutionID == "" {
		task.ExecutionID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(task).Error
}

func (s *TaskServiceImpl) UpdateTaskStatus(ctx context.Context, executionID string, status string, progress int) error {
	updates := map[string]interface{}{
		"Status":   status,
		"Progress": progress,
	}
	if status == "done" || status == "failed" {
		now := time.Now()
		updates["EndTime"] = &now
	}
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("\"ExecutionId\" = ?", executionID).Updates(updates).Error
}

func (s *TaskServiceImpl) GetTaskByExecutionID(ctx context.Context, executionID string) (*models.DigitalEmployeeTask, error) {
	var task models.DigitalEmployeeTask
	if err := s.db.WithContext(ctx).Where("\"ExecutionId\" = ?", executionID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskServiceImpl) AssignTask(ctx context.Context, executionID string, assigneeID uint) error {
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("\"ExecutionId\" = ?", executionID).Update("\"AssigneeId\"", assigneeID).Error
}

func (s *TaskServiceImpl) PlanTask(ctx context.Context, executionID string) error {
	// Simple plan: Just mark as in_progress
	return s.UpdateTaskStatus(ctx, executionID, "running", 0)
}

// ExecuteTask implements the generic execution logic driven by DB configuration
func (s *TaskServiceImpl) ExecuteTask(ctx context.Context, executionID string) error {
	// 1. Load Task
	task, err := s.GetTaskByExecutionID(ctx, executionID)
	if err != nil {
		return err
	}

	// 2. Load Assignee (Digital Employee)
	var emp models.DigitalEmployee
	if err := s.db.WithContext(ctx).Preload("Agent").First(&emp, task.AssigneeID).Error; err != nil {
		return fmt.Errorf("assignee not found: %v", err)
	}

	// 3. Get AI Service
	aiSvc := s.employeeSvc.GetAIService()
	if aiSvc == nil {
		return fmt.Errorf("AI service not initialized")
	}

	// 4. Get Tools
	mcpMgr := aiSvc.GetMCPManager()
	if mcpMgr == nil {
		return fmt.Errorf("MCP manager not available")
	}

	// Get tools available for this context (Global + Org + User)
	// We use 0 for UserID if not available
	tools, err := mcpMgr.GetToolsForContext(ctx, 0, emp.EnterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get tools: %v", err)
	}

	// 6. Dispatch based on TaskType
	switch task.TaskType {
	case "rule":
		return s.processRuleTask(ctx, task)
	case "hybrid":
		return s.processHybridTask(ctx, task, &emp, aiSvc, tools)
	case "ai":
		fallthrough
	default:
		return s.processAITask(ctx, task, &emp, aiSvc, tools)
	}
}

func (s *TaskServiceImpl) processRuleTask(ctx context.Context, task *models.DigitalEmployeeTask) error {
	// RuleTask: Strictly defined steps, no AI.
	// For now, assume Context contains a script name or pre-defined logic.
	// In a real implementation, this would look up a Go function or script from a registry.
	// TODO: Implement Rule Registry
	return fmt.Errorf("RuleTask execution not yet implemented - requires Rule Registry")
}

func (s *TaskServiceImpl) processHybridTask(ctx context.Context, task *models.DigitalEmployeeTask, emp *models.DigitalEmployee, aiSvc types.AIService, tools []types.Tool) error {
	// HybridTask:
	// 1. AI Generates Plan (List of Steps) -> Save to Redis
	// 2. Execute Steps Deterministically

	// Phase 1: Planning
	// Call AI to generate steps (not execute them)
	// We need a specific prompt to force JSON output of steps.
	// For now, we reuse the generic AI process but would need to parse the output into Redis steps.

	// Implementation placeholder
	return fmt.Errorf("HybridTask execution not yet implemented")
}

func (s *TaskServiceImpl) processAITask(ctx context.Context, task *models.DigitalEmployeeTask, emp *models.DigitalEmployee, aiSvc types.AIService, tools []types.Tool) error {
	// AITask: ReAct Pattern (Reason -> Act -> Observation)
	// Key requirement: Every step must be persisted to Redis before execution.

	// 5. Construct Prompt
	// Combine System Prompt from Agent (Role Template) + Task Description
	systemPrompt := emp.Agent.SystemPrompt
	if emp.Agent.SystemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	msgs := []types.Message{
		{Role: types.RoleSystem, Content: systemPrompt},
		{Role: types.RoleUser, Content: fmt.Sprintf("Task: %s\n\nPlease execute this task.", task.Description)},
	}

	// 6. Execute Agent via AI Service
	sessionID := fmt.Sprintf("task_%d_%d", task.ID, time.Now().Unix())

	// Pass metadata via context
	ctx = context.WithValue(ctx, "botID", emp.BotID)
	ctx = context.WithValue(ctx, "sessionID", sessionID)
	// We can also pass userID if needed, e.g. emp.OwnerID or similar

	// Use ChatAgent from AIService interface
	// NOTE: ChatAgent currently handles the loop internally.
	// To strictly comply with "Write Step to Redis -> Execute", we ideally need to
	// hook into ChatAgent or reimplement the loop here.
	// For now, we assume ChatAgent is compliant or we will refactor it later.
	// The current implementation is "AITask" style.

	resp, err := aiSvc.ChatAgent(ctx, emp.Agent.ModelID, msgs, tools)
	if err != nil {
		s.RecordTaskResult(ctx, task.ExecutionID, fmt.Sprintf("Execution failed: %v", err), false)
		return err
	}

	// 7. Record Result
	result := ""
	if len(resp.Choices) > 0 {
		if content, ok := resp.Choices[0].Message.Content.(string); ok {
			result = content
		}
	}

	return s.RecordTaskResult(ctx, task.ExecutionID, result, true)
}

func (s *TaskServiceImpl) ExecuteStep(ctx context.Context, executionID string, stepIndex int) error {
	return fmt.Errorf("not implemented")
}

func (s *TaskServiceImpl) ApproveTask(ctx context.Context, executionID string) error {
	return s.UpdateTaskStatus(ctx, executionID, "approved", 100)
}

func (s *TaskServiceImpl) CreateSubTask(ctx context.Context, parentExecutionID string, subTask *models.DigitalEmployeeTask) error {
	// Assuming ParentTaskID exists in model, if not, need migration
	// subTask.ParentTaskID = parentExecutionID
	return s.CreateTask(ctx, subTask)
}

func (s *TaskServiceImpl) RecordTaskResult(ctx context.Context, executionID string, result string, success bool) error {
	status := "done"
	if !success {
		status = "failed"
	}

	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("\"ExecutionId\" = ?", executionID).
		Updates(map[string]interface{}{
			"Status":  status,
			"Result":  result,
			"EndTime": time.Now(),
		}).Error
}
