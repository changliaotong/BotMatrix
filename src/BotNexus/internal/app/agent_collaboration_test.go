package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// MockAIIntegrationService 模拟 AI 集成服务
type MockAIIntegrationService struct {
	LastMsg types.InternalMessage
}

func (m *MockAIIntegrationService) DispatchIntent(msg types.InternalMessage) (string, error) {
	return "", nil
}

func (m *MockAIIntegrationService) ChatWithEmployee(employee *models.DigitalEmployeeGORM, msg types.InternalMessage, targetOrgID uint) (string, error) {
	m.LastMsg = msg
	return fmt.Sprintf("Hello from %s", employee.Name), nil
}

func (m *MockAIIntegrationService) GetProvider(id uint) (*models.AIProviderGORM, error) {
	return nil, nil
}

func (m *MockAIIntegrationService) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	return nil, nil
}

func (m *MockAIIntegrationService) ChatAgent(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	return nil, nil
}

func (m *MockAIIntegrationService) ChatStream(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (<-chan ai.ChatStreamResponse, error) {
	return nil, nil
}

func (m *MockAIIntegrationService) CreateEmbedding(ctx context.Context, modelID uint, input any) (*ai.EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockAIIntegrationService) ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall ai.ToolCall) (any, error) {
	return nil, nil
}

func TestAgentCollaborationMCPHost(t *testing.T) {
	// 1. 设置内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.DigitalEmployeeGORM{})

	// 2. 插入测试数据
	emp := models.DigitalEmployeeGORM{
		EmployeeID:   "EMP001",
		Name:         "Collaborator",
		EnterpriseID: 1,
		BotID:        "bot123",
	}
	db.Create(&emp)

	// 3. 模拟 Manager 和 AI Service
	mockAI := &MockAIIntegrationService{}
	mgr := &Manager{
		AIIntegrationService: mockAI,
	}
	mgr.GORMDB = db

	host := NewAgentCollaborationMCPHost(mgr)

	t.Run("colleague_list", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "orgIDNum", uint(1))
		resp, err := host.CallTool(ctx, "collaboration", "colleague_list", nil)
		if err != nil {
			t.Fatalf("colleague_list failed: %v", err)
		}

		mcpResp := resp.(ai.MCPCallToolResponse)
		if len(mcpResp.Content) == 0 {
			t.Fatal("expected content in response")
		}
		text := mcpResp.Content[0].Text
		if !contains(text, "Collaborator") {
			t.Errorf("expected 'Collaborator' in response, got: %s", text)
		}
	})

	t.Run("colleague_consult", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "orgIDNum", uint(1))
		ctx = context.WithValue(ctx, "sessionID", "parent-session")
		args := map[string]any{
			"target_employee_id": "EMP001",
			"question":           "How are you?",
		}
		resp, err := host.CallTool(ctx, "collaboration", "colleague_consult", args)
		if err != nil {
			t.Fatalf("colleague_consult failed: %v", err)
		}

		mcpResp := resp.(ai.MCPCallToolResponse)
		text := mcpResp.Content[0].Text
		if !contains(text, "Collaborator") || !contains(text, "Hello from Collaborator") {
			t.Errorf("unexpected response: %s", text)
		}

		// 检查父 session ID 传递
		if mockAI.LastMsg.Extras["parentSessionID"] != "parent-session" {
			t.Errorf("expected parentSessionID to be 'parent-session', got: %v", mockAI.LastMsg.Extras["parentSessionID"])
		}
	})
}
