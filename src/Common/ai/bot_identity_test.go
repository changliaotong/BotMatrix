package ai

import (
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// MockKnowledgeBase 模拟知识库
type MockKnowledgeBase struct{}

func (m *MockKnowledgeBase) Search(ctx context.Context, query string, limit int, filter *SearchFilter) ([]DocChunk, error) {
	if strings.Contains(query, "架构") {
		return []DocChunk{
			{
				ID:      "chunk_1",
				Source:  "DOCS.md",
				Content: "任务系统由 Task, Execution, Scheduler, Dispatcher 四大核心组件组成。",
			},
		}, nil
	}
	return nil, nil
}

// MockAIServiceForIdentity 模拟 AI 服务用于身份识别测试
type MockAIServiceForIdentity struct {
	LastSystemPrompt string
}

func (m *MockAIServiceForIdentity) ChatStream(ctx context.Context, modelID uint, messages []Message, tools []Tool) (<-chan ChatStreamResponse, error) {
	return nil, nil
}

func (m *MockAIServiceForIdentity) Chat(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	for _, msg := range messages {
		if msg.Role == RoleSystem {
			m.LastSystemPrompt, _ = msg.Content.(string)
		}
	}

	// 模拟 AI 回复：如果是询问身份，返回 system_query
	var lastUserMsg string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == RoleUser {
			lastUserMsg, _ = messages[i].Content.(string)
			break
		}
	}

	intent := "create_task"
	if strings.Contains(lastUserMsg, "你是谁") || strings.Contains(lastUserMsg, "介绍") {
		intent = "system_query"
	}

	return &ChatResponse{
		Choices: []Choice{
			{
				Message: Message{
					Role:    RoleAssistant,
					Content: "{\"intent\":\"" + intent + "\", \"summary\":\"自我介绍\", \"analysis\":\"你好！我是 BotMatrix，你的全能型群组自动化专家。\", \"data\":{}}",
					ToolCalls: []ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: FunctionCall{
								Name:      intent,
								Arguments: "{}",
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m *MockAIServiceForIdentity) CreateEmbedding(ctx context.Context, modelID uint, input any) (*EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockAIServiceForIdentity) ChatAgent(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	return m.Chat(ctx, modelID, messages, tools)
}

func TestBotIdentityInjection(t *testing.T) {
	manifest := tasks.GetDefaultManifest()
	mockAI := &MockAIServiceForIdentity{}
	parser := &tasks.AIParser{
		Manifest: manifest,
	}
	parser.SetAIService(mockAI)

	ctx := context.Background()
	parseCtx := map[string]any{
		"bot_id":             "bot_123",
		"effective_group_id": "group_456",
		"user_role":          "admin",
		"is_private":         true,
	}

	result, err := parser.MatchSkillByLLM(ctx, "你是谁？", 1, parseCtx)
	if err != nil {
		t.Fatalf("MatchSkillByLLM failed: %v", err)
	}

	// 1. 验证 System Prompt 是否包含身份信息
	if mockAI.LastSystemPrompt == "" {
		t.Error("System Prompt was not captured")
	}

	expectedParts := []string{
		"你现在的身份是：BotMatrix (全能型群组自动化专家)",
		"你的性格特征是：专业、高效、偶尔幽默",
		"### 操作指南 (How-to Guide):",
		"- **创建任务**:",
		"- **群管理**:",
		"当前运行环境信息：",
		"- 目标群组 ID: group_456",
		"- 当前用户角色: admin",
		"- 当前对话类型: 私聊",
	}

	for _, part := range expectedParts {
		if !contains(mockAI.LastSystemPrompt, part) {
			t.Errorf("System Prompt missing expected part: %s", part)
		}
	}

	// 2. 验证解析结果
	if result.Intent != tasks.AIActionSystemQuery {
		t.Errorf("Expected intent %s, got %s", tasks.AIActionSystemQuery, result.Intent)
	}

	if !contains(result.Analysis, "BotMatrix") {
		t.Errorf("Analysis should mention BotMatrix, got: %s", result.Analysis)
	}

	fmt.Printf("Parsed Result: %+v\n", result)
}

func contains(s, substr string) bool {
	return (s != "" && substr != "" && (len(s) >= len(substr)) && (s == substr || (s[:len(substr)] == substr) || (s[len(s)-len(substr):] == substr) || (len(s) > len(substr) && (s[1:len(substr)+1] == substr || contains(s[1:], substr)))))
}

func TestBotIdentityJson(t *testing.T) {
	manifest := tasks.GetDefaultManifest()
	data, _ := json.MarshalIndent(manifest, "", "  ")

	var m map[string]any
	json.Unmarshal(data, &m)

	identity, ok := m["identity"].(map[string]any)
	if !ok {
		t.Fatal("Identity field missing in JSON manifest")
	}

	if identity["name"] != "BotMatrix" {
		t.Errorf("Expected name BotMatrix, got %v", identity["name"])
	}
}
