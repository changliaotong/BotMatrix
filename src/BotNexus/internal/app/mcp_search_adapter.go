package app

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
)

// SearchMCPHost 提供安全联网搜索工具
type SearchMCPHost struct {
	manager *Manager
}

func NewSearchMCPHost(m *Manager) *SearchMCPHost {
	return &SearchMCPHost{manager: m}
}

func (h *SearchMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "web_search",
			Description: "Search the web for real-time information. Privacy Bastion will automatically protect your sensitive data in the query.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query.",
					},
				},
				"required": []string{"query"},
			},
		},
	}, nil
}

func (h *SearchMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *SearchMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *SearchMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "web_search" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	query := arguments["query"].(string)

	// 这里模拟搜索结果。
	// 注意：在真实的 MCPManager.CallTool 中，query 已经被脱敏过了。
	// 这里返回的结果如果包含敏感词，也会在 MCPManager.CallTool 中被还原。
	
	fmt.Printf("[Search] Executing web search for: %s\n", query)

	// 模拟结果，故意包含一些可能被脱敏的词，看看还原效果
	// 假设 "BotMatrix" 或 "ProjectX" 是敏感词
	mockResults := fmt.Sprintf("Search results for '%s':\n1. Latest updates on BotMatrix project status.\n2. ProjectX internal documentation and roadmap.", query)

	return ai.MCPCallToolResponse{
		Content: []ai.MCPContent{
			{
				Type: "text",
				Text: mockResults,
			},
		},
	}, nil
}

func (h *SearchMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *SearchMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}
