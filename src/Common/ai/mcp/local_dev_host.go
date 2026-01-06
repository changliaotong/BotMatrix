package mcp

import (
	"BotMatrix/common/types"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LocalDevMCPHost provides tools for the agent to develop on the host machine safely.
type LocalDevMCPHost struct {
	baseDir        string
	protectedPaths []string
	allowedCmds    []string
}

func NewLocalDevMCPHost(baseDir string) *LocalDevMCPHost {
	return &LocalDevMCPHost{
		baseDir: baseDir,
		protectedPaths: []string{
			".git",
			"config.json",
			".env",
			"id_rsa",
			"id_rsa.pub",
		},
		allowedCmds: []string{
			"go",
			"git",
			"ls",
			"dir",
			"grep",
			"cat",
			"find",
			"mkdir",
			// "rm", // Explicitly excluding rm for now, use a safer delete tool if needed
		},
	}
}

func (h *LocalDevMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "dev_read_file",
			Description: "Read a file from the project workspace.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative path to the file (e.g., 'src/main.go')",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "dev_write_file",
			Description: "Write content to a file safely (creates backup automatically).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative path to the file",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Content to write",
					},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "dev_run_cmd",
			Description: "Execute a whitelisted shell command in the workspace.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Command to run (e.g., 'go build ./...')",
					},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "dev_git_commit",
			Description: "Stage all changes and commit with a message.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{
						"type":        "string",
						"description": "Commit message (follow conventional commits)",
					},
				},
				"required": []string{"message"},
			},
		},
	}, nil
}

func (h *LocalDevMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *LocalDevMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *LocalDevMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	switch toolName {
	case "dev_read_file":
		path, _ := arguments["path"].(string)
		return h.readFile(path)
	case "dev_write_file":
		path, _ := arguments["path"].(string)
		content, _ := arguments["content"].(string)
		return h.writeFile(path, content)
	case "dev_run_cmd":
		cmd, _ := arguments["command"].(string)
		return h.runCmd(cmd)
	case "dev_git_commit":
		msg, _ := arguments["message"].(string)
		return h.gitCommit(msg)
	}
	return nil, fmt.Errorf("unknown tool: %s", toolName)
}

func (h *LocalDevMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, nil
}

func (h *LocalDevMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", nil
}

// Helper methods

func (h *LocalDevMCPHost) resolvePath(path string) (string, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid path: contains '..'")
	}

	fullPath := filepath.Join(h.baseDir, cleanPath)

	// Check protected paths
	for _, protected := range h.protectedPaths {
		if strings.HasPrefix(cleanPath, protected) {
			return "", fmt.Errorf("access denied: %s is a protected path", cleanPath)
		}
	}

	return fullPath, nil
}

func (h *LocalDevMCPHost) readFile(path string) (any, error) {
	fullPath, err := h.resolvePath(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: string(content)},
		},
	}, nil
}

func (h *LocalDevMCPHost) writeFile(path string, content string) (any, error) {
	fullPath, err := h.resolvePath(path)
	if err != nil {
		return nil, err
	}

	// Create backup if file exists
	if _, err := os.Stat(fullPath); err == nil {
		backupPath := filepath.Join(h.baseDir, ".backups", path+"."+time.Now().Format("20060102150405")+".bak")
		os.MkdirAll(filepath.Dir(backupPath), 0755)

		src, err := os.Open(fullPath)
		if err == nil {
			dst, err := os.Create(backupPath)
			if err == nil {
				io.Copy(dst, src)
				dst.Close()
			}
			src.Close()
		}
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return nil, err
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: fmt.Sprintf("Successfully wrote to %s (Backup created)", path)},
		},
	}, nil
}

func (h *LocalDevMCPHost) runCmd(command string) (any, error) {
	// Simple whitelist check
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	baseCmd := parts[0]
	allowed := false
	for _, cmd := range h.allowedCmds {
		if baseCmd == cmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, fmt.Errorf("command '%s' is not allowed", baseCmd)
	}

	// Execute
	cmd := exec.Command("powershell", "-Command", command)
	cmd.Dir = h.baseDir
	output, err := cmd.CombinedOutput()

	result := string(output)
	if err != nil {
		result += fmt.Sprintf("\nError: %v", err)
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: result},
		},
	}, nil
}

func (h *LocalDevMCPHost) gitCommit(message string) (any, error) {
	// git add .
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = h.baseDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git add failed: %s", string(out))
	}

	// git commit -m "message"
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = h.baseDir
	out, err := commitCmd.CombinedOutput()

	result := string(out)
	if err != nil {
		// Check if it's just "nothing to commit"
		if strings.Contains(result, "nothing to commit") {
			return types.MCPCallToolResponse{
				Content: []types.MCPContent{
					{Type: "text", Text: "Nothing to commit."},
				},
			}, nil
		}
		return nil, fmt.Errorf("git commit failed: %s", result)
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: fmt.Sprintf("Git commit successful:\n%s", result)},
		},
	}, nil
}
