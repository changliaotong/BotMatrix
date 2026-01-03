package app

import (
	"BotMatrix/common/ai"
	"BotNexus/internal/rag"
	"BotNexus/tasks"
	"context"
	"fmt"
	"strings"
)

// KnowledgeMCPHost 提供本地知识库检索能力
type KnowledgeMCPHost struct {
	manager *Manager
	kb      *rag.PostgresKnowledgeBase
}

func NewKnowledgeMCPHost(m *Manager, kb *rag.PostgresKnowledgeBase) *KnowledgeMCPHost {
	return &KnowledgeMCPHost{
		manager: m,
		kb:      kb,
	}
}

func (h *KnowledgeMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "search_knowledge",
			Description: "Search the knowledge base for technical documentation, architecture details, and project information using semantic search.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query or keywords in natural language.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 3).",
					},
				},
				"required": []string{"query"},
			},
		},
	}, nil
}

func (h *KnowledgeMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *KnowledgeMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *KnowledgeMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "search_knowledge" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	if h.kb == nil {
		return nil, fmt.Errorf("knowledge base not initialized")
	}

	query := arguments["query"].(string)
	limit := 3
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	// 执行语义搜索
	filter := &tasks.SearchFilter{
		Status: "active",
	}
	chunks, err := h.kb.Search(ctx, query, limit, filter)
	if err != nil {
		return nil, fmt.Errorf("knowledge search failed: %v", err)
	}

	if len(chunks) == 0 {
		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{{Type: "text", Text: "No relevant information found in the knowledge base."}},
		}, nil
	}

	var results []string
	for _, c := range chunks {
		results = append(results, fmt.Sprintf("Source: %s\nContent: %s", c.Source, c.Content))
	}

	return ai.MCPCallToolResponse{
		Content: []ai.MCPContent{{Type: "text", Text: "Found relevant information:\n\n" + strings.Join(results, "\n\n---\n\n")}},
	}, nil
}

func (h *KnowledgeMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *KnowledgeMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}
