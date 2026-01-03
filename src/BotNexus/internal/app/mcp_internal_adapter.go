package app

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
)

// InternalSkillMCPHost 将 BotNexus 现有的 Skill Center 桥接为 MCP 协议
type InternalSkillMCPHost struct {
	manager *Manager
}

func NewInternalSkillMCPHost(m *Manager) *InternalSkillMCPHost {
	return &InternalSkillMCPHost{manager: m}
}

func (h *InternalSkillMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	if h.manager.TaskManager == nil || h.manager.TaskManager.AI == nil {
		return nil, nil
	}

	// 从 TaskManager.AI.Manifest 获取工具定义
	aiTools := h.manager.TaskManager.AI.Manifest.GenerateTools()
	mcpTools := make([]ai.MCPTool, len(aiTools))
	for i, t := range aiTools {
		mcpTools[i] = ai.MCPTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
	}
	return mcpTools, nil
}

func (h *InternalSkillMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	// 暂时不实现资源列表
	return nil, nil
}

func (h *InternalSkillMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	// 暂时不实现提示词模板列表
	return nil, nil
}

func (h *InternalSkillMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if h.manager.TaskManager == nil || h.manager.TaskManager.Dispatcher == nil {
		return nil, fmt.Errorf("task manager or dispatcher not initialized")
	}

	// 桥接到 Dispatcher 执行
	err := h.manager.TaskManager.Dispatcher.ExecuteAction(toolName, arguments)
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "success", "message": "Action executed"}, nil
}

func (h *InternalSkillMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource reading not supported for internal skills")
}

func (h *InternalSkillMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt templates not supported for internal skills")
}
