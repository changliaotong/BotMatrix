package app

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
)

// ReasoningMCPHost 提供推理相关的工具，如 Sequential Thinking
type ReasoningMCPHost struct {
	manager *Manager
}

func NewReasoningMCPHost(m *Manager) *ReasoningMCPHost {
	return &ReasoningMCPHost{manager: m}
}

func (h *ReasoningMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "sequential_thinking",
			Description: "A tool for dynamic and reflective problem-solving through a structured sequence of thoughts. Use this for complex tasks that require planning, step-by-step reasoning, and self-correction.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"thought": map[string]any{
						"type":        "string",
						"description": "The current thinking process or step.",
					},
					"thoughtNumber": map[string]any{
						"type":        "integer",
						"description": "The current step number in the reasoning chain.",
					},
					"totalThoughts": map[string]any{
						"type":        "integer",
						"description": "The estimated total number of steps (can be adjusted).",
					},
					"nextStep": map[string]any{
						"type":        "string",
						"description": "What to do next based on this thought.",
					},
					"isFinished": map[string]any{
						"type":        "boolean",
						"description": "Whether the reasoning process is complete.",
					},
				},
				"required": []string{"thought", "thoughtNumber", "nextStep"},
			},
		},
	}, nil
}

func (h *ReasoningMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *ReasoningMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *ReasoningMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "sequential_thinking" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	thought := arguments["thought"].(string)
	num := arguments["thoughtNumber"]
	isFinished, _ := arguments["isFinished"].(bool)

	// 在控制台输出推理过程，模拟“思考中”的状态
	fmt.Printf("\n[Reasoning] Step %v: %s\n", num, thought)

	msg := fmt.Sprintf("Thought accepted. Next: %v", arguments["nextStep"])
	if isFinished {
		msg = "Reasoning complete. Proceeding to final answer."
	}

	return ai.MCPCallToolResponse{
		Content: []ai.MCPContent{
			{
				Type: "text",
				Text: msg,
			},
		},
	}, nil
}

func (h *ReasoningMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *ReasoningMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}
