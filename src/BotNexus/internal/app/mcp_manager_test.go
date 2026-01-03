package app

import (
	"BotMatrix/common/ai"
	"context"
	"testing"
)

// MockMCPHost 模拟一个 MCP 服务器
type MockMCPHost struct{}

func (m *MockMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "get_weather",
			Description: "Get weather for a city",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{"type": "string"},
				},
			},
		},
	}, nil
}

func (m *MockMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (m *MockMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (m *MockMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName == "get_weather" {
		return "Sunny, 25°C", nil
	}
	return nil, nil
}

func (m *MockMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, nil
}

func (m *MockMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", nil
}

func TestMCPManager(t *testing.T) {
	m := NewMCPManager(nil, &Manager{})
	m.RegisterServer(ai.MCPServerInfo{
		ID:    "weather_service",
		Name:  "Weather Service",
		Scope: ai.ScopeGlobal,
	}, &MockMCPHost{})

	ctx := context.Background()
	tools, err := m.GetToolsForContext(ctx, 0, 0)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	toolName := tools[0].Function.Name
	if toolName != "weather_service__get_weather" {
		t.Errorf("Unexpected tool name: %s", toolName)
	}

	res, err := m.CallTool(ctx, toolName, map[string]any{"city": "Beijing"})
	if err != nil {
		t.Fatalf("Failed to call tool: %v", err)
	}

	if res != "Sunny, 25°C" {
		t.Errorf("Unexpected result: %v", res)
	}

	// 测试权限过滤
	m.RegisterServer(ai.MCPServerInfo{
		ID:      "private_service",
		Name:    "Private Service",
		Scope:   ai.ScopeUser,
		OwnerID: 123,
	}, &MockMCPHost{})

	// 匿名上下文应找不到私有工具
	tools, _ = m.GetToolsForContext(ctx, 0, 0)
	if len(tools) != 1 {
		t.Errorf("Anonymous user should only see 1 tool, got %d", len(tools))
	}

	// 指定用户 ID 应看到私有工具
	tools, _ = m.GetToolsForContext(ctx, 123, 0)
	if len(tools) != 2 {
		t.Errorf("User 123 should see 2 tools, got %d", len(tools))
	}
}
