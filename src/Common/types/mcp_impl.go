package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RegisteredServer 已注册的服务器
type RegisteredServer struct {
	Info MCPServerInfo
	Host MCPHost
}

// MCPManager 管理 MCP 服务器的连接与工具发现
type MCPManager struct {
	servers       map[string]*RegisteredServer // serverID -> RegisteredServer
	PrivacyFilter *PrivacyFilter
	mu            sync.RWMutex
}

func NewMCPManager() *MCPManager {
	return &MCPManager{
		servers:       make(map[string]*RegisteredServer),
		PrivacyFilter: NewPrivacyFilter(),
	}
}

// RegisterServer 注册一个远程或本地的 MCP 服务器
func (m *MCPManager) RegisterServer(info MCPServerInfo, host MCPHost) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[info.ID] = &RegisteredServer{
		Info: info,
		Host: host,
	}
}

// UnregisterServer 注销服务器
func (m *MCPManager) UnregisterServer(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.servers, serverID)
}

// GetServer 获取服务器
func (m *MCPManager) GetServer(serverID string) (*RegisteredServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rs, ok := m.servers[serverID]
	return rs, ok
}

// GetKnowledgeBase 获取注入的向量知识库实现
func (m *MCPManager) GetKnowledgeBase() KnowledgeBase {
	return nil
}

// SetKnowledgeBase 注入向量知识库实现
func (m *MCPManager) SetKnowledgeBase(kb KnowledgeBase) {
	// 基础实现不处理 KB，由扩展实现处理
}

// GetToolsForContext 根据上下文（用户、组织等）获取可用的工具
func (m *MCPManager) GetToolsForContext(ctx context.Context, userID uint, orgID uint) ([]Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allTools := make([]Tool, 0)
	for _, rs := range m.servers {
		allowed := false
		switch rs.Info.Scope {
		case ScopeGlobal:
			allowed = true
		case ScopeOrg:
			if rs.Info.OwnerID == orgID {
				allowed = true
			}
		case ScopeUser:
			if rs.Info.OwnerID == userID {
				allowed = true
			}
		}

		if !allowed {
			continue
		}

		mcpTools, err := rs.Host.ListTools(ctx, rs.Info.ID)
		if err != nil {
			continue
		}

		for _, mt := range mcpTools {
			tool := mt.ToOpenAITool()
			tool.Function.Name = fmt.Sprintf("%s__%s", rs.Info.ID, tool.Function.Name)
			allTools = append(allTools, tool)
		}
	}
	return allTools, nil
}

// CallTool 调用指定的 MCP 工具
func (m *MCPManager) CallTool(ctx context.Context, fullName string, args map[string]any) (any, error) {
	parts := strings.SplitN(fullName, "__", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid MCP tool name format: %s", fullName)
	}
	serverID := parts[0]
	toolName := parts[1]

	m.mu.RLock()
	rs, ok := m.servers[serverID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("MCP server not found: %s", serverID)
	}

	var pCtx *MaskContext
	if !strings.HasPrefix(rs.Info.ID, "internal") && m.PrivacyFilter != nil {
		maskedArgs := make(map[string]any)
		pCtx = NewMaskContext()
		for k, v := range args {
			if strV, ok := v.(string); ok {
				maskedArgs[k] = m.PrivacyFilter.Mask(strV, pCtx)
			} else {
				maskedArgs[k] = v
			}
		}
		args = maskedArgs
	}

	result, err := rs.Host.CallTool(ctx, serverID, toolName, args)
	if err != nil {
		return nil, err
	}

	if pCtx != nil && pCtx.Counter > 0 && m.PrivacyFilter != nil {
		if mcpResp, ok := result.(MCPCallToolResponse); ok {
			for i := range mcpResp.Content {
				if mcpResp.Content[i].Type == "text" {
					mcpResp.Content[i].Text = m.PrivacyFilter.Unmask(mcpResp.Content[i].Text, pCtx)
				}
			}
			result = mcpResp
		} else if strRes, ok := result.(string); ok {
			result = m.PrivacyFilter.Unmask(strRes, pCtx)
		}
	}

	return result, nil
}

// GenericWebhookMCPHost 实现了一个通用的 Webhook 适配器
// 它可以将任何符合简单 JSON 规范的国内业务 API 转换为 MCP Tool
type GenericWebhookMCPHost struct {
	Endpoint string
	APIKey   string
	Tools    []MCPTool
}

func NewGenericWebhookMCPHost(endpoint, apiKey string, tools []MCPTool) *GenericWebhookMCPHost {
	return &GenericWebhookMCPHost{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Tools:    tools,
	}
}

func (h *GenericWebhookMCPHost) ListTools(ctx context.Context, serverID string) ([]MCPTool, error) {
	return h.Tools, nil
}

func (h *GenericWebhookMCPHost) ListResources(ctx context.Context, serverID string) ([]MCPResource, error) {
	return nil, nil
}

func (h *GenericWebhookMCPHost) ListPrompts(ctx context.Context, serverID string) ([]MCPPrompt, error) {
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
