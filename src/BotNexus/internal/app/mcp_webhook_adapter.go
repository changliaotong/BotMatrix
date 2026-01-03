package app

import (
	"BotMatrix/common/ai"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GenericWebhookMCPHost 实现了一个通用的 Webhook 适配器
// 它可以将任何符合简单 JSON 规范的国内业务 API 转换为 MCP Tool
type GenericWebhookMCPHost struct {
	Endpoint string
	APIKey   string
	Tools    []ai.MCPTool
}

func NewGenericWebhookMCPHost(endpoint, apiKey string, tools []ai.MCPTool) *GenericWebhookMCPHost {
	return &GenericWebhookMCPHost{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Tools:    tools,
	}
}

func (h *GenericWebhookMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return h.Tools, nil
}

func (h *GenericWebhookMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *GenericWebhookMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *GenericWebhookMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	reqBody, _ := json.Marshal(map[string]any{
		"tool":      toolName,
		"arguments": arguments,
		"server_id": serverID,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", h.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("webhook returned non-200 status: %d", resp.StatusCode)
	}

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *GenericWebhookMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource reading not supported via generic webhook yet")
}

func (h *GenericWebhookMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt templates not supported via generic webhook yet")
}
