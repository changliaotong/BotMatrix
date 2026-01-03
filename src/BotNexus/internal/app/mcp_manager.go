package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotNexus/internal/rag"
	"context"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// MCPManager 管理 MCP 服务器的连接与工具发现
type MCPManager struct {
	db            *gorm.DB
	manager       *Manager
	servers       map[string]*registeredServer // serverID -> registeredServer
	PrivacyFilter *ai.PrivacyFilter
	mu            sync.RWMutex
}

type registeredServer struct {
	info ai.MCPServerInfo
	host ai.MCPHost
}

func NewMCPManager(db *gorm.DB, m *Manager) *MCPManager {
	mgr := &MCPManager{
		db:            db,
		manager:       m,
		servers:       make(map[string]*registeredServer),
		PrivacyFilter: ai.NewPrivacyFilter(),
	}

	// 添加演示用的自定义脱敏规则
	mgr.PrivacyFilter.AddCustomPattern("PROJECT", `(BotMatrix|ProjectX|InternalPlan)`)

	// 注册内置技能作为 MCP Server
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "internal_skills",
		Name:  "Internal Bot Skills",
		Scope: ai.ScopeGlobal,
	}, NewInternalSkillMCPHost(m))

	// 注册推理增强工具
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "reasoning",
		Name:  "Reasoning Engine",
		Scope: ai.ScopeGlobal,
	}, NewReasoningMCPHost(m))

	// 注册安全搜索工具
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "search",
		Name:  "Secure Search",
		Scope: ai.ScopeGlobal,
	}, NewSearchMCPHost(m))

	// 注册本地知识库工具 (初始时不带 KB，后续通过 SetKnowledgeBase 注入)
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "knowledge",
		Name:  "Knowledge Base",
		Scope: ai.ScopeGlobal,
	}, NewKnowledgeMCPHost(m, nil))

	// 注册长期记忆工具
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "memory",
		Name:  "Agent Memory",
		Scope: ai.ScopeGlobal,
	}, NewMemoryMCPHost(m))

	// 注册 IM 桥接工具 (适配器模式并行)
	mgr.RegisterServer(ai.MCPServerInfo{
		ID:    "im_bridge",
		Name:  "IM Adapter Bridge",
		Scope: ai.ScopeGlobal,
	}, NewIMBridgeMCPHost(m))

	// 加载数据库配置
	mgr.LoadFromDB()
	return mgr
}

// SetKnowledgeBase 注入向量知识库实现
func (m *MCPManager) SetKnowledgeBase(kb *rag.PostgresKnowledgeBase) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.servers["knowledge"]; ok {
		if host, ok := s.host.(*KnowledgeMCPHost); ok {
			host.kb = kb
		}
	}
}

// LoadFromDB 从数据库加载启用的 MCP 服务器
func (m *MCPManager) LoadFromDB() error {
	if m.db == nil {
		return nil
	}

	var configs []models.MCPServerGORM
	if err := m.db.Where("status = ?", "active").Find(&configs).Error; err != nil {
		return err
	}

	for _, cfg := range configs {
		var host ai.MCPHost
		switch cfg.Type {
		case "webhook":
			// Webhook 类型直接使用 Endpoint
			host = NewGenericWebhookMCPHost(cfg.Endpoint, cfg.APIKey, nil)
		case "internal":
			host = NewInternalSkillMCPHost(m.manager)
		}

		if host != nil {
			m.RegisterServer(ai.MCPServerInfo{
				ID:      fmt.Sprintf("db_%d", cfg.ID),
				Name:    cfg.Name,
				Scope:   ai.MCPServerScope(cfg.Scope),
				OwnerID: cfg.OwnerID,
			}, host)
		}
	}
	return nil
}

// RegisterServer 注册一个远程或本地的 MCP 服务器
func (m *MCPManager) RegisterServer(info ai.MCPServerInfo, host ai.MCPHost) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[info.ID] = &registeredServer{
		info: info,
		host: host,
	}
	fmt.Printf("[MCP] Registered server: %s (Scope: %s)\n", info.Name, info.Scope)
}

// GetToolsForContext 根据上下文（用户、组织等）获取可用的工具
func (m *MCPManager) GetToolsForContext(ctx context.Context, userID uint, orgID uint) ([]ai.Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allTools := make([]ai.Tool, 0)
	for _, rs := range m.servers {
		// 权限检查
		allowed := false
		switch rs.info.Scope {
		case ai.ScopeGlobal:
			allowed = true
		case ai.ScopeOrg:
			if rs.info.OwnerID == orgID {
				allowed = true
			}
		case ai.ScopeUser:
			if rs.info.OwnerID == userID {
				allowed = true
			}
		}

		if !allowed {
			continue
		}

		mcpTools, err := rs.host.ListTools(ctx, rs.info.ID)
		if err != nil {
			fmt.Printf("[MCP] Failed to list tools for server %s: %v\n", rs.info.ID, err)
			continue
		}

		for _, mt := range mcpTools {
			tool := mt.ToOpenAITool()
			tool.Function.Name = fmt.Sprintf("%s__%s", rs.info.ID, tool.Function.Name)
			allTools = append(allTools, tool)
		}
	}
	return allTools, nil
}

// CallTool 调用指定的 MCP 工具
func (m *MCPManager) CallTool(ctx context.Context, fullName string, args map[string]any) (any, error) {
	// 使用 strings.SplitN 分割 serverID 和 toolName
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

	// 隐私脱敏 (Privacy Bastion)
	// 仅对非 internal 服务器进行脱敏，或者根据配置决定
	var pCtx *ai.MaskContext
	if rs.info.ID != "internal_skills" && m.PrivacyFilter != nil {
		maskedArgs := make(map[string]any)
		pCtx = ai.NewMaskContext()
		for k, v := range args {
			if strV, ok := v.(string); ok {
				maskedArgs[k] = m.PrivacyFilter.Mask(strV, pCtx)
			} else {
				maskedArgs[k] = v
			}
		}
		args = maskedArgs
		fmt.Printf("[MCP] Privacy Bastion: Masked %d sensitive items for tool %s\n", pCtx.Counter, fullName)
	}

	result, err := rs.host.CallTool(ctx, serverID, toolName, args)
	if err != nil {
		return nil, err
	}

	// 还原结果 (Restore sensitive data)
	if pCtx != nil && pCtx.Counter > 0 && m.PrivacyFilter != nil {
		if mcpResp, ok := result.(ai.MCPCallToolResponse); ok {
			for i := range mcpResp.Content {
				if mcpResp.Content[i].Type == "text" {
					mcpResp.Content[i].Text = m.PrivacyFilter.Restore(mcpResp.Content[i].Text, pCtx)
				}
			}
			result = mcpResp
		} else if strRes, ok := result.(string); ok {
			result = m.PrivacyFilter.Restore(strRes, pCtx)
		}
		// 可以根据需要添加更多类型的处理 (如 map[string]any)
	}

	return result, nil
}
