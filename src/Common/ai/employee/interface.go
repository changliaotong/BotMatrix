package employee

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
)

// DigitalEmployeeService 数字员工核心服务接口
type DigitalEmployeeService interface {
	GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error)
	RecordKpi(employeeID uint, metric string, score float64) error
	UpdateOnlineStatus(botID string, status string) error
	ConsumeSalary(botID string, tokens int64) error
	CheckSalaryLimit(botID string) (bool, error)
	UpdateSalary(botID string, salaryToken *int64, salaryLimit *int64) error
	AutoEvolve(employeeID uint) error
	Recruit(orgID uint, jobID uint) (*models.DigitalEmployee, error)
	Fire(orgID uint, employeeID string) error
	Transfer(orgID uint, employeeID string, newJobID uint) error
	ChatWithEmployee(employee *models.DigitalEmployee, message types.InternalMessage, orgID uint) (string, error)
	GetCognitiveMemory(ctx context.Context, employeeID string, category string) ([]models.CognitiveMemory, error)
	AddCognitiveMemory(ctx context.Context, employeeID string, memory models.CognitiveMemory) error
	UpdateCognitiveMemory(ctx context.Context, employeeID string, memory models.CognitiveMemory) error
	DeleteCognitiveMemory(ctx context.Context, employeeID string, memoryID uint) error
	SearchCognitiveMemory(ctx context.Context, employeeID string, query string) ([]models.CognitiveMemory, error)
	AssignTask(ctx context.Context, employeeID string, task models.DigitalEmployeeTask) error
	GetTask(ctx context.Context, employeeID string, taskID uint) (*models.DigitalEmployeeTask, error)
	UpdateTaskStatus(ctx context.Context, employeeID string, taskID uint, status string) error
	GetTasksByStatus(ctx context.Context, employeeID string, status string) ([]models.DigitalEmployeeTask, error)
	GetAIService() types.AIService
}

// CognitiveMemoryService 认知记忆服务接口
type CognitiveMemoryService interface {
	GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemory, error)
	GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemory, error)
	SaveMemory(ctx context.Context, memory *models.CognitiveMemory) error
	ForgetMemory(ctx context.Context, memoryID uint) error
	SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemory, error)
	SetEmbeddingService(svc any) // Use any to avoid import cycle if needed, or rag.EmbeddingService
	ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc types.AIService) error
	LearnFromURL(ctx context.Context, botID string, url string, category string) error
	LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error
}

// DigitalEmployeeTaskService 数字员工任务服务接口
type DigitalEmployeeTaskService interface {
	CreateTask(ctx context.Context, task *models.DigitalEmployeeTask) error
	UpdateTaskStatus(ctx context.Context, executionID string, status string, progress int) error
	GetTaskByExecutionID(ctx context.Context, executionID string) (*models.DigitalEmployeeTask, error)
	AssignTask(ctx context.Context, executionID string, assigneeID uint) error
	PlanTask(ctx context.Context, executionID string) error
	ExecuteTask(ctx context.Context, executionID string) error
	ExecuteStep(ctx context.Context, executionID string, stepIndex int) error
	ApproveTask(ctx context.Context, executionID string) error
	CreateSubTask(ctx context.Context, parentExecutionID string, subTask *models.DigitalEmployeeTask) error
	RecordTaskResult(ctx context.Context, executionID string, result string, success bool) error
}

// DigitalEmployeeKPIService 数字员工KPI服务接口
type DigitalEmployeeKPIService interface {
	CalculateKPI(ctx context.Context, employeeID uint) (float64, error)
	OptimizeEmployee(ctx context.Context, employeeID uint) error
	GetPerformanceReport(ctx context.Context, employeeID uint, days int) (string, error)
}
