package main

import (
	"context"
	"log"
)

// MCPServer 定义了MCP服务器接口
type MCPServer interface {
	RegisterTool(tool string, handler func(ctx context.Context, args map[string]any) (any, error))
	RegisterResource(resource string, provider func(ctx context.Context, uri string) (any, error))
	RegisterPrompt(prompt string, generator func(ctx context.Context, args map[string]any) (string, error))
}

// MockMCPServer 模拟MCP服务器
// 用于示例和测试
type MockMCPServer struct {
	tools     map[string]func(ctx context.Context, args map[string]any) (any, error)
	resources map[string]func(ctx context.Context, uri string) (any, error)
	prompts   map[string]func(ctx context.Context, args map[string]any) (string, error)
}

// NewMockMCPServer 创建新的模拟MCP服务器
func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		tools:     make(map[string]func(ctx context.Context, args map[string]any) (any, error)),
		resources: make(map[string]func(ctx context.Context, uri string) (any, error)),
		prompts:   make(map[string]func(ctx context.Context, args map[string]any) (string, error)),
	}
}

// RegisterTool 注册工具
func (m *MockMCPServer) RegisterTool(tool string, handler func(ctx context.Context, args map[string]any) (any, error)) {
	m.tools[tool] = handler
	log.Printf("Registered MCP tool: %s", tool)
}

// RegisterResource 注册资源
func (m *MockMCPServer) RegisterResource(resource string, provider func(ctx context.Context, uri string) (any, error)) {
	m.resources[resource] = provider
	log.Printf("Registered MCP resource: %s", resource)
}

// RegisterPrompt 注册提示
func (m *MockMCPServer) RegisterPrompt(prompt string, generator func(ctx context.Context, args map[string]any) (string, error)) {
	m.prompts[prompt] = generator
	log.Printf("Registered MCP prompt: %s", prompt)
}

// main 主函数
func main() {
	// 创建模拟MCP服务器
	mockMCP := NewMockMCPServer()

	// 注册测试工具
	mockMCP.RegisterTool("test_tool", func(ctx context.Context, args map[string]any) (any, error) {
		return map[string]interface{}{
			"result": "success",
			"args":   args,
		}, nil
	})

	log.Printf("Mock MCP server created with %d tools, %d resources, %d prompts",
		len(mockMCP.tools), len(mockMCP.resources), len(mockMCP.prompts))

	log.Println("Test completed successfully!")
}