package mcp

import (
	"BotMatrix/common/browser"
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/sandbox"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

// InternalSkillProviderImpl 实现了 InternalSkillProvider 接口
type InternalSkillProviderImpl struct {
	m types.Manager
}

func NewInternalSkillProviderImpl(m types.Manager) *InternalSkillProviderImpl {
	return &InternalSkillProviderImpl{m: m}
}

func (p *InternalSkillProviderImpl) GetTools() []types.Tool {
	// 这里需要处理 TaskManager，暂时通过反射或类型断言，或者在 types 中定义更通用的接口
	// 为了解决编译，先简单处理
	return nil
}

func (p *InternalSkillProviderImpl) ExecuteAction(actionType string, params map[string]any) error {
	return fmt.Errorf("not implemented")
}

// IMServiceProviderImpl 实现了 IMServiceProvider 接口
type IMServiceProviderImpl struct {
	m types.Manager
}

func NewIMServiceProviderImpl(m types.Manager) *IMServiceProviderImpl {
	return &IMServiceProviderImpl{m: m}
}

func (p *IMServiceProviderImpl) SendMessage(platform, targetID, content string) error {
	// 实际调用 BotNexus 内部的 IM 适配器逻辑
	// 这里假设 Manager 有相关方法或可以直接调用
	fmt.Printf("[IM-Provider] Sending message via %s to %s: %s\n", platform, targetID, content)
	return nil
}

func (p *IMServiceProviderImpl) GetActivePlatforms() []string {
	return []string{"onebot_v11", "discord_gateway", "wechat_work"}
}

// CollaborationProviderImpl 实现了 CollaborationProvider 接口
type CollaborationProviderImpl struct {
	m types.Manager
}

func NewCollaborationProviderImpl(m types.Manager) *CollaborationProviderImpl {
	return &CollaborationProviderImpl{m: m}
}

func (p *CollaborationProviderImpl) GetEmployeesByOrg(ctx context.Context, orgID uint) ([]models.DigitalEmployee, error) {
	var employees []models.DigitalEmployee
	if err := p.m.GetGORMDB().WithContext(ctx).Where("enterprise_id = ?", orgID).Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (p *CollaborationProviderImpl) GetEmployeeByID(ctx context.Context, orgID uint, employeeID string) (*models.DigitalEmployee, error) {
	var targetEmp models.DigitalEmployee
	if err := p.m.GetGORMDB().WithContext(ctx).Where("enterprise_id = ? AND employee_id = ?", orgID, employeeID).First(&targetEmp).Error; err != nil {
		return nil, fmt.Errorf("未找到工号为 %s 的同事", employeeID)
	}
	return &targetEmp, nil
}

func (p *CollaborationProviderImpl) ChatWithEmployee(ctx context.Context, employee *models.DigitalEmployee, message *types.Message, orgID uint) (string, error) {
	return p.m.GetAIService().ChatWithEmployee(employee, types.InternalMessage{
		RawMessage: fmt.Sprintf("%v", message.Content),
	}, orgID)
}

func (p *CollaborationProviderImpl) CreateExecution(ctx context.Context, executionID, traceID string) error {
	now := time.Now()
	exec := models.Execution{
		ExecutionID: executionID,
		Status:      models.ExecRunning,
		TriggerTime: now,
		ActualTime:  &now,
		TraceID:     traceID,
	}
	return p.m.GetGORMDB().WithContext(ctx).Create(&exec).Error
}

func (p *CollaborationProviderImpl) UpdateExecution(ctx context.Context, executionID string, status string, result string) error {
	return p.m.GetGORMDB().WithContext(ctx).Model(&models.Execution{}).
		Where("execution_id = ?", executionID).
		Updates(map[string]any{
			"status": status,
			"result": result,
		}).Error
}

func (p *CollaborationProviderImpl) GetExecution(ctx context.Context, executionID string) (*CollaborationExecutionInfo, error) {
	var exec models.Execution
	if err := p.m.GetGORMDB().WithContext(ctx).Where("execution_id = ?", executionID).First(&exec).Error; err != nil {
		return nil, err
	}
	return &CollaborationExecutionInfo{
		ExecutionID: exec.ExecutionID,
		Status:      string(exec.Status),
		Result:      exec.Result,
	}, nil
}

func (p *CollaborationProviderImpl) NewInternalMessage(msgType, category, botID, role, content string) *types.Message {
	// 这里的 NewInternalMessage 可能不在 types.Manager 中，暂时返回一个基础 Message
	return &types.Message{
		Role:    types.Role(role),
		Content: content,
	}
}

// MCPManager 扩展了通用的 types.MCPManager，增加数据库加载和 Manager 绑定
type MCPManager struct {
	*types.MCPManager
	db      *gorm.DB
	manager types.Manager
}

func NewMCPManager(m types.Manager) *MCPManager {
	mgr := &MCPManager{
		MCPManager: types.NewMCPManager(),
		db:         m.GetGORMDB(),
		manager:    m,
	}

	// Initialize Docker client and Sandbox Manager
	dockerCli, err := utils.InitDockerClient()
	if err == nil {
		// Use a default image, e.g., python:3.10-slim
		sandboxMgr := sandbox.NewSandboxManager(dockerCli, "python:3.10-slim")
		mgr.RegisterServer(types.MCPServerInfo{
			ID:    "sandbox",
			Name:  "Code Sandbox",
			Scope: types.ScopeGlobal,
		}, NewSandboxMCPHost(sandboxMgr))
	} else {
		log.Warn(fmt.Sprintf("Docker client initialization failed, sandbox tool will be unavailable: %v", err))
	}

	// Register Local Developer Host (Safe File System & Git Access)
	// Assuming current working directory is the project root
	cwd, _ := os.Getwd()
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "local_dev",
		Name:  "Local Developer Environment",
		Scope: types.ScopeGlobal,
	}, NewLocalDevMCPHost(cwd))

	// Initialize Browser Manager
	// 默认 headless=true，下载路径为空（使用系统默认）
	browserMgr, err := browser.NewBrowserManager(true, "")
	if err == nil {
		// 尝试预启动，或者懒加载
		// 这里注册 MCP Server，实际调用时会自动启动
		mgr.RegisterServer(types.MCPServerInfo{
			ID:    "browser",
			Name:  "Web Browser",
			Scope: types.ScopeGlobal,
		}, NewBrowserMCPHost(browserMgr))
	} else {
		log.Warn(fmt.Sprintf("Browser manager initialization failed: %v", err))
	}

	// 添加演示用的自定义脱敏规则
	mgr.PrivacyFilter.AddCustomPattern("PROJECT", `(BotMatrix|ProjectX|InternalPlan)`)

	// 注册内置技能作为 MCP Server
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "internal_skills",
		Name:  "Internal Bot Skills",
		Scope: types.ScopeGlobal,
	}, NewInternalSkillMCPHost(NewInternalSkillProviderImpl(m)))

	// 注册推理增强工具
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "reasoning",
		Name:  "Reasoning Engine",
		Scope: types.ScopeGlobal,
	}, NewReasoningMCPHost(m.GetAIService()))

	// 注册安全搜索工具
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "search",
		Name:  "Secure Search",
		Scope: types.ScopeGlobal,
	}, NewSearchMCPHost(m.GetAIService()))

	// 注册本地知识库工具 (初始时不带 KB，后续通过 SetKnowledgeBase 注入)
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "knowledge",
		Name:  "Knowledge Base",
		Scope: types.ScopeGlobal,
	}, NewKnowledgeMCPHost(m.GetKnowledgeBase()))

	// 注册长期记忆工具
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "memory",
		Name:  "Agent Memory",
		Scope: types.ScopeGlobal,
	}, NewMemoryMCPHost(m.GetCognitiveMemoryService()))

	// 注册 IM 桥接工具 (适配器模式并行)
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "im_bridge",
		Name:  "IM Adapter Bridge",
		Scope: types.ScopeGlobal,
	}, NewIMBridgeMCPHost(NewIMServiceProviderImpl(m)))

	// 注册多智能体协作工具
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "collaboration",
		Name:  "Agent Collaboration",
		Scope: types.ScopeGlobal,
	}, NewAgentCollaborationMCPHost(NewCollaborationProviderImpl(m)))

	// Register System Administration Tools
	mgr.RegisterServer(types.MCPServerInfo{
		ID:    "sys_admin",
		Name:  "System Administration",
		Scope: types.ScopeGlobal,
	}, NewSysAdminMCPHost(m.GetGORMDB()))

	// 加载数据库配置
	mgr.LoadFromDB()
	return mgr
}

// SetKnowledgeBase 注入向量知识库实现
func (m *MCPManager) SetKnowledgeBase(kb types.KnowledgeBase) {
	if rs, ok := m.GetServer("knowledge"); ok {
		if host, ok := rs.Host.(*KnowledgeMCPHost); ok {
			host.kb = kb
		}
	}
}

// GetKnowledgeBase 获取注入的向量知识库实现
func (m *MCPManager) GetKnowledgeBase() types.KnowledgeBase {
	if rs, ok := m.GetServer("knowledge"); ok {
		if host, ok := rs.Host.(*KnowledgeMCPHost); ok {
			return host.kb
		}
	}
	return nil
}

// LoadFromDB 从数据库加载启用的 MCP 服务器
func (m *MCPManager) LoadFromDB() error {
	if m.db == nil {
		return nil
	}

	var configs []models.MCPServer
	// Use struct-based condition to handle column names automatically
	if err := m.db.Where(&models.MCPServer{Status: "active"}).Find(&configs).Error; err != nil {
		return err
	}

	for _, cfg := range configs {
		var host types.MCPHost
		switch cfg.Type {
		case "webhook":
			// Webhook 类型直接使用 Endpoint
			host = types.NewGenericWebhookMCPHost(cfg.Endpoint, cfg.APIKey, nil)
		case "internal":
			host = NewInternalSkillMCPHost(NewInternalSkillProviderImpl(m.manager))
		}

		if host != nil {
			m.RegisterServer(types.MCPServerInfo{
				ID:      fmt.Sprintf("db_%d", cfg.ID),
				Name:    cfg.Name,
				Scope:   types.MCPServerScope(cfg.Scope),
				OwnerID: cfg.OwnerID,
			}, host)
		}
	}
	return nil
}

// CallTool 包装了 types.MCPManager.CallTool，可以根据需要增加逻辑
func (m *MCPManager) CallTool(ctx context.Context, fullName string, args map[string]any) (any, error) {
	return m.MCPManager.CallTool(ctx, fullName, args)
}

// GetToolsForContext 包装了 types.MCPManager.GetToolsForContext
func (m *MCPManager) GetToolsForContext(ctx context.Context, userID uint, orgID uint) ([]types.Tool, error) {
	return m.MCPManager.GetToolsForContext(ctx, userID, orgID)
}
