package employee

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
)

// DigitalEmployeeService 数字员工核心服务接口
type DigitalEmployeeService interface {
	GetEmployeeByBotID(botID string) (*models.DigitalEmployeeGORM, error)
	RecordKpi(employeeID uint, metric string, score float64) error
	UpdateOnlineStatus(botID string, status string) error
	ConsumeSalary(botID string, tokens int64) error
	CheckSalaryLimit(botID string) (bool, error)
	UpdateSalary(botID string, salaryToken *int64, salaryLimit *int64) error
	AutoEvolve(employeeID uint) error
}

// CognitiveMemoryService 认知记忆服务接口
type CognitiveMemoryService interface {
	GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemoryGORM, error)
	GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemoryGORM, error)
	SaveMemory(ctx context.Context, memory *models.CognitiveMemoryGORM) error
	ForgetMemory(ctx context.Context, memoryID uint) error
	SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemoryGORM, error)
	SetEmbeddingService(svc any) // Use any to avoid import cycle if needed, or rag.EmbeddingService
	ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc types.AIService) error
	LearnFromURL(ctx context.Context, botID string, url string, category string) error
	LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error
}

// DigitalEmployeeTaskService 数字员工任务服务接口
type DigitalEmployeeTaskService interface {
	CreateTask(ctx context.Context, task *models.DigitalEmployeeTaskGORM) error
	UpdateTaskStatus(ctx context.Context, executionID string, status string, progress int) error
	GetTaskByExecutionID(ctx context.Context, executionID string) (*models.DigitalEmployeeTaskGORM, error)
	AssignTask(ctx context.Context, executionID string, assigneeID uint) error
	PlanTask(ctx context.Context, executionID string) error
	ExecuteTask(ctx context.Context, executionID string) error
	ExecuteStep(ctx context.Context, executionID string, stepIndex int) error
	ApproveTask(ctx context.Context, executionID string) error
	CreateSubTask(ctx context.Context, parentExecutionID string, subTask *models.DigitalEmployeeTaskGORM) error
	RecordTaskResult(ctx context.Context, executionID string, result string, success bool) error
}

// DigitalEmployeeKPIService 数字员工KPI服务接口
type DigitalEmployeeKPIService interface {
	CalculateKPI(ctx context.Context, employeeID uint) (float64, error)
	OptimizeEmployee(ctx context.Context, employeeID uint) error
	GetPerformanceReport(ctx context.Context, employeeID uint, days int) (string, error)
}
