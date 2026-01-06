package mcp

import (
	"BotMatrix/common/sandbox"
	"BotMatrix/common/types"
	"context"
	"fmt"
)

type SandboxMCPHost struct {
	manager *sandbox.SandboxManager
}

func NewSandboxMCPHost(manager *sandbox.SandboxManager) *SandboxMCPHost {
	return &SandboxMCPHost{manager: manager}
}

func (h *SandboxMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "sandbox_create",
			Description: "Create a new isolated sandbox environment (Docker container). Returns the sandbox_id.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image": map[string]any{
						"type":        "string",
						"description": "Docker image to use (default: python:3.10-slim)",
					},
				},
			},
		},
		{
			Name:        "sandbox_exec",
			Description: "Execute a shell command inside the sandbox.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sandbox_id": map[string]any{
						"type":        "string",
						"description": "The ID of the sandbox returned by sandbox_create",
					},
					"command": map[string]any{
						"type":        "string",
						"description": "Shell command to execute (e.g., 'ls -la', 'python script.py')",
					},
				},
				"required": []string{"sandbox_id", "command"},
			},
		},
		{
			Name:        "sandbox_write_file",
			Description: "Write content to a file inside the sandbox.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sandbox_id": map[string]any{
						"type":        "string",
						"description": "The ID of the sandbox",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Absolute path to the file (e.g., '/workspace/script.py')",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Text content to write",
					},
				},
				"required": []string{"sandbox_id", "path", "content"},
			},
		},
		{
			Name:        "sandbox_read_file",
			Description: "Read content from a file inside the sandbox.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sandbox_id": map[string]any{
						"type":        "string",
						"description": "The ID of the sandbox",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Absolute path to the file",
					},
				},
				"required": []string{"sandbox_id", "path"},
			},
		},
		{
			Name:        "sandbox_destroy",
			Description: "Destroy the sandbox and release resources.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sandbox_id": map[string]any{
						"type":        "string",
						"description": "The ID of the sandbox",
					},
				},
				"required": []string{"sandbox_id"},
			},
		},
	}, nil
}

func (h *SandboxMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *SandboxMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *SandboxMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	switch toolName {
	case "sandbox_create":
		image, _ := arguments["image"].(string)
		sb, err := h.manager.CreateSandbox(ctx, image)
		if err != nil {
			return nil, err
		}
		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Sandbox created successfully. ID: %s", sb.ID),
				},
			},
		}, nil

	case "sandbox_exec":
		id, _ := arguments["sandbox_id"].(string)
		cmd, _ := arguments["command"].(string)
		if id == "" || cmd == "" {
			return nil, fmt.Errorf("missing sandbox_id or command")
		}

		// Reconstruct sandbox object (assuming ID is valid for now)
		sb := &sandbox.Sandbox{ID: id, Manager: h.manager}
		stdout, stderr, err := sb.Exec(ctx, cmd)
		if err != nil {
			return nil, err
		}

		output := fmt.Sprintf("STDOUT:\n%s\n", stdout)
		if stderr != "" {
			output += fmt.Sprintf("\nSTDERR:\n%s", stderr)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: output,
				},
			},
		}, nil

	case "sandbox_write_file":
		id, _ := arguments["sandbox_id"].(string)
		path, _ := arguments["path"].(string)
		content, _ := arguments["content"].(string)
		if id == "" || path == "" {
			return nil, fmt.Errorf("missing sandbox_id or path")
		}

		sb := &sandbox.Sandbox{ID: id, Manager: h.manager}
		err := sb.WriteFile(ctx, path, []byte(content))
		if err != nil {
			return nil, err
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Successfully wrote to %s", path),
				},
			},
		}, nil

	case "sandbox_read_file":
		id, _ := arguments["sandbox_id"].(string)
		path, _ := arguments["path"].(string)
		if id == "" || path == "" {
			return nil, fmt.Errorf("missing sandbox_id or path")
		}

		sb := &sandbox.Sandbox{ID: id, Manager: h.manager}
		content, err := sb.ReadFile(ctx, path)
		if err != nil {
			return nil, err
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: string(content),
				},
			},
		}, nil

	case "sandbox_destroy":
		id, _ := arguments["sandbox_id"].(string)
		if id == "" {
			return nil, fmt.Errorf("missing sandbox_id")
		}

		sb := &sandbox.Sandbox{ID: id, Manager: h.manager}
		err := sb.Destroy(ctx)
		if err != nil {
			return nil, err
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: "Sandbox destroyed successfully",
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown tool: %s", toolName)
}

func (h *SandboxMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource not found")
}

func (h *SandboxMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt not found")
}
