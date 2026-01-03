package app

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
)

// IMBridgeMCPHost 将现有的 IM 适配器能力桥接为 MCP 工具
type IMBridgeMCPHost struct {
	manager *Manager
}

func NewIMBridgeMCPHost(m *Manager) *IMBridgeMCPHost {
	return &IMBridgeMCPHost{manager: m}
}

func (h *IMBridgeMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "im_send_message",
			Description: "通过现有的 IM 适配器（微信、Discord 等）发送消息。这是适配器模式与 MCP 模式并行的典型应用。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"platform": map[string]any{
						"type":        "string",
						"description": "目标平台，如 'wechat', 'discord', 'onebot'",
					},
					"target_id": map[string]any{
						"type":        "string",
						"description": "接收者 ID、群号或频道 ID",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "消息文本内容",
					},
				},
				"required": []string{"platform", "target_id", "content"},
			},
		},
		{
			Name:        "im_list_platforms",
			Description: "列出当前系统已连接并激活的 IM 平台适配器",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
	}, nil
}

func (h *IMBridgeMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	switch toolName {
	case "im_send_message":
		platform, _ := arguments["platform"].(string)
		targetID, _ := arguments["target_id"].(string)
		content, _ := arguments["content"].(string)

		// 模拟调用现有适配器逻辑
		// 在实际代码中，这里会调用 manager.SendMessage(platform, targetID, content)
		fmt.Printf("[IM-Bridge] Calling legacy adapter: platform=%s, target=%s, content=%s\n", platform, targetID, content)

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("已通过适配器模式向 %s(%s) 发送消息: %s", platform, targetID, content),
				},
			},
		}, nil

	case "im_list_platforms":
		// 模拟获取已加载的适配器列表
		platforms := []string{"onebot_v11", "discord_gateway", "wechat_work"}
		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("当前激活的适配器: %v", platforms),
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func (h *IMBridgeMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *IMBridgeMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource not supported")
}

func (h *IMBridgeMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *IMBridgeMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt not supported")
}
