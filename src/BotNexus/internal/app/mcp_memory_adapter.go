package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"context"
	"fmt"
	"strings"
	"time"
)

// MemoryMCPHost 提供基于数据库和向量检索的长期记忆能力
type MemoryMCPHost struct {
	manager *Manager
}

func NewMemoryMCPHost(m *Manager) *MemoryMCPHost {
	return &MemoryMCPHost{
		manager: m,
	}
}

func (h *MemoryMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "store_memory",
			Description: "Store important information in long-term memory. Supports automatic semantic indexing.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{
						"type":        "string",
						"description": "The information to remember.",
					},
					"category": map[string]any{
						"type":        "string",
						"description": "Optional category (e.g., 'user_preference', 'project_info').",
					},
					"importance": map[string]any{
						"type":        "integer",
						"description": "Importance level (1-10), default is 5.",
						"minimum":     1,
						"maximum":     10,
					},
				},
				"required": []string{"content"},
			},
		},
		{
			Name:        "search_memory",
			Description: "Search through long-term memory using semantic similarity or keywords. Returns memory IDs for potential management.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query to find relevant memories.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "forget_memory",
			Description: "Remove a specific piece of information from long-term memory using its memory ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"memory_id": map[string]any{
						"type":        "integer",
						"description": "The unique ID of the memory to forget.",
					},
				},
				"required": []string{"memory_id"},
			},
		},
	}, nil
}

func (h *MemoryMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *MemoryMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *MemoryMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if h.manager.CognitiveMemoryService == nil {
		return nil, fmt.Errorf("CognitiveMemoryService not initialized")
	}

	// 尝试从上下文中获取当前 BotID 和 UserID
	// 注意：在实际调用中，这些应该从 MCP 请求的元数据或上下文中透传过来
	// 这里先使用演示用的默认值，或从 arguments 中尝试获取（如果 prompt 设计了透传）
	botID := "system"
	userID := "system"

	switch toolName {
	case "store_memory":
		content := arguments["content"].(string)
		category, _ := arguments["category"].(string)
		importance, _ := arguments["importance"].(float64)
		if importance == 0 {
			importance = 5
		}

		memory := &models.CognitiveMemoryGORM{
			BotID:      botID,
			UserID:     userID,
			Content:    content,
			Category:   category,
			Importance: int(importance),
			LastSeen:   time.Now(),
		}

		err := h.manager.CognitiveMemoryService.SaveMemory(ctx, memory)
		if err != nil {
			return nil, err
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{{Type: "text", Text: "Memory saved and indexed successfully."}},
		}, nil

	case "search_memory":
		query := arguments["query"].(string)
		memories, err := h.manager.CognitiveMemoryService.GetRelevantMemories(ctx, userID, botID, query)
		if err != nil {
			return nil, err
		}

		if len(memories) == 0 {
			return ai.MCPCallToolResponse{
				Content: []ai.MCPContent{{Type: "text", Text: "No relevant memories found."}},
			}, nil
		}

		var results []string
		for _, m := range memories {
			results = append(results, fmt.Sprintf("[ID: %d] [%s] %s (Importance: %d)", m.ID, m.Category, m.Content, m.Importance))
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{{Type: "text", Text: "Found relevant memories:\n" + fmt.Sprintf("- %s", strings.Join(results, "\n- "))}},
		}, nil

	case "forget_memory":
		memoryIDFloat, ok := arguments["memory_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid memory_id: must be an integer")
		}
		memoryID := uint(memoryIDFloat)

		err := h.manager.CognitiveMemoryService.ForgetMemory(ctx, memoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to forget memory: %v", err)
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{{Type: "text", Text: fmt.Sprintf("Memory ID %d has been removed from long-term memory.", memoryID)}},
		}, nil

	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func (h *MemoryMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *MemoryMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}
