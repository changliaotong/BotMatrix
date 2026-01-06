package mcp

import (
	"BotMatrix/common/browser"
	"BotMatrix/common/types"
	"context"
	"encoding/base64"
	"fmt"
)

type BrowserMCPHost struct {
	manager *browser.BrowserManager
}

func NewBrowserMCPHost(manager *browser.BrowserManager) *BrowserMCPHost {
	return &BrowserMCPHost{manager: manager}
}

func (h *BrowserMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "browser_navigate",
			Description: "Navigate to a URL and return the text content of the page.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to visit (e.g., https://example.com)",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "browser_screenshot",
			Description: "Take a full-page screenshot of a URL and return base64 encoded image.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to visit",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "browser_extract",
			Description: "Extract text content from a specific CSS selector on a page.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to visit",
					},
					"selector": map[string]any{
						"type":        "string",
						"description": "CSS selector to extract (e.g., '#main-content', '.article-body')",
					},
				},
				"required": []string{"url", "selector"},
			},
		},
	}, nil
}

func (h *BrowserMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *BrowserMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *BrowserMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	switch toolName {
	case "browser_navigate":
		url, _ := arguments["url"].(string)
		if url == "" {
			return nil, fmt.Errorf("missing url")
		}

		content, err := h.manager.Navigate(ctx, url)
		if err != nil {
			return nil, err
		}

		// 截断过长的内容，避免 Context 溢出
		// 简单的截断策略，后续可以优化为 Summarize
		maxLen := 10000
		if len(content) > maxLen {
			content = content[:maxLen] + "\n... (content truncated)"
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: content,
				},
			},
		}, nil

	case "browser_screenshot":
		url, _ := arguments["url"].(string)
		if url == "" {
			return nil, fmt.Errorf("missing url")
		}

		imgData, err := h.manager.Screenshot(ctx, url)
		if err != nil {
			return nil, err
		}

		base64Str := base64.StdEncoding.EncodeToString(imgData)
		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "image", // 注意：标准 MCP 可能是 image/png，这里使用 text 传输 base64 或者自定义类型
					// 这里假设下游能处理 image 类型，或者作为 text 返回 data uri
					// 为了兼容性，先返回 Text 描述，实际图像数据可能需要通过 Resource 访问
					// 或者直接返回 base64 文本
					Text: fmt.Sprintf("Screenshot taken. Base64 data length: %d", len(base64Str)),
				},
				{
					Type: "text",
					Text: "data:image/png;base64," + base64Str,
				},
			},
		}, nil

	case "browser_extract":
		url, _ := arguments["url"].(string)
		selector, _ := arguments["selector"].(string)
		if url == "" || selector == "" {
			return nil, fmt.Errorf("missing url or selector")
		}

		content, err := h.manager.ExtractContent(ctx, url, selector)
		if err != nil {
			return nil, err
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: content,
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (h *BrowserMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource not found")
}

func (h *BrowserMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt not found")
}
