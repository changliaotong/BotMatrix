package ai

import (
	"context"
)

// MCP (Model Context Protocol) 核心结构定义
// 参考: https://modelcontextprotocol.io/

// MCPServer 定义了作为 MCP 服务端的行为
type MCPServer interface {
	// 注册服务
	RegisterTool(tool MCPTool, handler func(ctx context.Context, args map[string]any) (any, error))
	RegisterResource(resource MCPResource, provider func(ctx context.Context, uri string) (any, error))
	RegisterPrompt(prompt MCPPrompt, generator func(ctx context.Context, args map[string]any) (string, error))
}
