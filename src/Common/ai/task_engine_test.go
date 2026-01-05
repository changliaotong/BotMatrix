package ai

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockAIService 模拟 AI 服务
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) DispatchIntent(msg types.InternalMessage) (string, error) {
	return "", nil
}

func (m *MockAIService) ChatWithEmployee(employee *models.DigitalEmployeeGORM, msg types.InternalMessage, targetOrgID uint) (string, error) {
	return "", nil
}

func (m *MockAIService) Chat(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	args := m.Called(ctx, modelID, messages, tools)
	return args.Get(0).(*ChatResponse), args.Error(1)
}

func (m *MockAIService) ChatAgent(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	args := m.Called(ctx, modelID, messages, tools)
	return args.Get(0).(*ChatResponse), args.Error(1)
}

func (m *MockAIService) ChatStream(ctx context.Context, modelID uint, messages []Message, tools []Tool) (<-chan ChatStreamResponse, error) {
	return nil, nil
}

func (m *MockAIService) CreateEmbedding(ctx context.Context, modelID uint, input any) (*EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockAIService) ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall ToolCall) (any, error) {
	return nil, nil
}

func (m *MockAIService) GetProvider(id uint) (*models.AIProviderGORM, error) {
	return nil, nil
}

func TestTaskEngineFlow(t *testing.T) {
	// 1. 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.DigitalEmployeeTaskGORM{})

	// 2. 初始化 Mock AI 服务
	mockAI := new(MockAIService)

	// 3. 初始化任务服务
	taskSvc := NewDigitalEmployeeTaskService(db, nil)
	taskSvc.SetAIService(mockAI)

	ctx := context.Background()
	executionID := "test-exec-001"

	// --- 阶段 1: 创建任务 ---
	task := &models.DigitalEmployeeTaskGORM{
		ExecutionID: executionID,
		Title:       "分析财务报表",
		Description: "分析 2025 年 Q4 财报，找出支出异常项。",
	}
	err = taskSvc.CreateTask(ctx, task)
	assert.NoError(t, err)

	// --- 阶段 2: 任务规划 ---
	// 模拟 AI 返回的计划
	planJSON := `{"steps": [
		{"index": 1, "title": "读取数据", "description": "从数据库读取 Q4 支出数据"},
		{"index": 2, "title": "对比分析", "description": "对比 Q3 数据，标记增长超过 20% 的项", "requires_approval": true},
		{"index": 3, "title": "生成报告", "description": "汇总异常项并给出建议"}
	]}`
	mockAI.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&ChatResponse{
		Choices: []Choice{
			{Message: Message{Content: planJSON}},
		},
	}, nil).Once()

	err = taskSvc.PlanTask(ctx, executionID)
	assert.NoError(t, err)

	var updatedTask models.DigitalEmployeeTaskGORM
	db.Where("execution_id = ?", executionID).First(&updatedTask)
	assert.Equal(t, "executing", updatedTask.Status)
	assert.Contains(t, updatedTask.PlanRaw, "读取数据")

	// --- 阶段 3: 执行任务 (步骤 1) ---
	// 模拟步骤 1 的 AI 执行结果
	mockAI.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&ChatResponse{
		Choices: []Choice{
			{Message: Message{Content: "已读取数据，共 100 条支出记录。"}},
		},
	}, nil).Once()

	err = taskSvc.ExecuteTask(ctx, executionID)
	assert.NoError(t, err)

	db.Where("execution_id = ?", executionID).First(&updatedTask)
	// 因为步骤 2 需要审批，所以状态应该是 pending_approval
	assert.Equal(t, "pending_approval", updatedTask.Status)
	assert.Equal(t, 1, updatedTask.CurrentStepIndex)
	assert.Contains(t, updatedTask.ResultRaw, "已读取数据")

	// --- 阶段 4: 审批并继续 ---
	err = taskSvc.ApproveTask(ctx, executionID)
	assert.NoError(t, err)

	// 模拟步骤 2 和 3 的 AI 执行结果
	mockAI.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&ChatResponse{
		Choices: []Choice{
			{Message: Message{Content: "发现差旅费增长 35%，属于异常。"}},
		},
	}, nil).Once()
	mockAI.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&ChatResponse{
		Choices: []Choice{
			{Message: Message{Content: "报告已生成：建议加强差旅合规检查。"}},
		},
	}, nil).Once()

	err = taskSvc.ExecuteTask(ctx, executionID)
	assert.NoError(t, err)

	db.Where("execution_id = ?", executionID).First(&updatedTask)
	assert.Equal(t, "completed", updatedTask.Status)
	assert.Equal(t, 100, updatedTask.Progress)
	assert.Contains(t, updatedTask.ResultRaw, "建议加强差旅合规检查")
}
