package ai

import (
	"context"
)

// MCP (Model Context Protocol) 核心结构定义
// 参考: https://modelcontextprotocol.io/

// MCPTool 定义了 MCP 协议中的工具格式
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPResource 定义了 MCP 协议中的资源
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCPPrompt 定义了 MCP 协议中的提示词模板
type MCPPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Arguments   []MCPPromptArgument `json:"arguments,omitempty"`
}

type MCPPromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// MCPServerScope 定义 MCP 服务器的可用范围
type MCPServerScope string

const (
	ScopeGlobal MCPServerScope = "global" // 全局可用
	ScopeOrg    MCPServerScope = "org"    // 组织/群组可用
	ScopeUser   MCPServerScope = "user"   // 仅个人可用
)

// MCPServerInfo 包含 MCP 服务器的元数据
type MCPServerInfo struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Scope       MCPServerScope `json:"scope"`
	OwnerID     uint           `json:"owner_id,omitempty"` // 所属用户或组织 ID
}

// MCPListToolsResponse tools/list 响应
type MCPListToolsResponse struct {
	Tools []MCPTool `json:"tools"`
}

// MCPCallToolRequest tools/call 请求
type MCPCallToolRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// MCPCallToolResponse tools/call 响应
type MCPCallToolResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
	Type string `json:"type"` // text, image, resource
	Text string `json:"text,omitempty"`
}

// MCPHost 定义了作为 MCP 客户端（Host）的行为
type MCPHost interface {
	// 列表发现
	ListTools(ctx context.Context, serverID string) ([]MCPTool, error)
	ListResources(ctx context.Context, serverID string) ([]MCPResource, error)
	ListPrompts(ctx context.Context, serverID string) ([]MCPPrompt, error)

	// 调用执行
	CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error)
	ReadResource(ctx context.Context, serverID string, uri string) (any, error)
	GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error)
}

// MCPServer 定义了作为 MCP 服务端的行为
type MCPServer interface {
	// 注册服务
	RegisterTool(tool MCPTool, handler func(ctx context.Context, args map[string]any) (any, error))
	RegisterResource(resource MCPResource, provider func(ctx context.Context, uri string) (any, error))
	RegisterPrompt(prompt MCPPrompt, generator func(ctx context.Context, args map[string]any) (string, error))
}

// 将 MCP 工具转换为 OpenAI 格式的工具
func (t *MCPTool) ToOpenAITool() Tool {
	return Tool{
		Type: "function",
		Function: FunctionDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.InputSchema,
		},
	}
}
