package app

import (
	"BotMatrix/common/ai"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SkillManager 负责数字员工技能的发现、授权与执行调度
type SkillManager struct {
	db         *gorm.DB
	manager    *Manager
	mcpManager *MCPManager
	b2bService B2BService
}

func NewSkillManager(db *gorm.DB, manager *Manager, mcpManager *MCPManager) *SkillManager {
	return &SkillManager{
		db:         db,
		manager:    manager,
		mcpManager: mcpManager,
	}
}

func (sm *SkillManager) SetB2BService(b2b B2BService) {
	sm.b2bService = b2b
}

// GetAvailableSkillsForBot 获取特定机器人可用的所有技能描述（用于 AI Tool 定义）
func (sm *SkillManager) GetAvailableSkillsForBot(ctx context.Context, botID string, userID uint, orgID uint) ([]ai.Tool, error) {
	var tools []ai.Tool

	// 1. 获取基础插件技能 (从 Worker 上报的 Capabilities 中获取)
	sm.manager.Mutex.RLock()
	seenSkills := make(map[string]bool)
	for _, w := range sm.manager.Workers {
		for _, cap := range w.Capabilities {
			if !seenSkills[cap.Name] {
				// 权限检查
				allowed, _ := sm.CheckPermission(ctx, botID, userID, orgID, cap.Name)
				if !allowed {
					continue
				}

				tools = append(tools, ai.Tool{
					Type: "function",
					Function: ai.FunctionDefinition{
						Name:        cap.Name,
						Description: cap.Description,
						Parameters:  cap.Parameters,
					},
				})
				seenSkills[cap.Name] = true
			}
		}
	}
	sm.manager.Mutex.RUnlock()

	// 2. 获取 MCP 挂载的技能
	if sm.mcpManager != nil {
		mcpTools, err := sm.mcpManager.GetToolsForContext(ctx, userID, orgID)
		if err == nil {
			for _, mt := range mcpTools {
				// 权限检查 (包含 Bot 级精细化授权)
				allowed, _ := sm.CheckPermission(ctx, botID, userID, orgID, mt.Function.Name)
				if allowed {
					tools = append(tools, mt)
				}
			}
		}
	}

	// 3. 获取 B2B 共享技能
	if orgID > 0 {
		var sharedSkills []models.B2BSkillSharingGORM
		err := sm.db.WithContext(ctx).Where("target_ent_id = ? AND is_active = ?", orgID, true).Find(&sharedSkills).Error
		if err == nil {
			for _, ss := range sharedSkills {
				// 构造 B2B 技能标识: b2b__entID__skillName
				b2bSkillName := fmt.Sprintf("b2b__%d__%s", ss.SourceEntID, ss.SkillName)
				tools = append(tools, ai.Tool{
					Type: "function",
					Function: ai.FunctionDefinition{
						Name:        b2bSkillName,
						Description: fmt.Sprintf("[B2B 共享技能] %s", ss.SkillName),
						Parameters:  map[string]any{"type": "object"}, // 远程技能参数动态，暂定 object
					},
				})
			}
		}
	}

	return tools, nil
}

// GetToolsForContext 获取当前上下文下可用的所有工具 (整合了基础技能与 MCP 技能)
// 这是对 GetAvailableSkillsForBot 的别名，提供更符合语义的上下文接口
func (sm *SkillManager) GetToolsForContext(ctx context.Context, botID string, userID uint, orgID uint) ([]ai.Tool, error) {
	return sm.GetAvailableSkillsForBot(ctx, botID, userID, orgID)
}

// CheckPermission 检查机器人是否有权执行该技能
func (sm *SkillManager) CheckPermission(ctx context.Context, botID string, userID uint, orgID uint, skillName string) (bool, error) {
	// 1. 系统级核心技能 (如 send_message) 默认允许
	coreSkills := map[string]bool{
		"send_message": true,
		"system_query": true,
	}
	if coreSkills[skillName] {
		return true, nil
	}

	// 2. 如果是 MCP 工具 (包含 __ 分隔符)，需要验证用户是否有权访问该 MCP Server
	if strings.Contains(skillName, "__") {
		serverName := strings.Split(skillName, "__")[0]
		// 检查该用户/组织是否挂载了该 MCP Server
		var server models.MCPServerGORM
		err := sm.db.WithContext(ctx).
			Where("name = ? AND (owner_id = ? OR scope = 'global')", serverName, orgID).
			First(&server).Error

		if err != nil {
			clog.Warn("[Skill] MCP Server not found or access denied",
				zap.String("server", serverName),
				zap.Uint("org_id", orgID))
			return false, nil
		}
	}

	// 3. 从数据库查询授权记录 (Bot 级别的精细化授权)
	var permission models.BotSkillPermissionGORM
	err := sm.db.WithContext(ctx).
		Where("bot_id = ? AND skill_name = ?", botID, skillName).
		First(&permission).Error

	if err == nil {
		return permission.IsAllowed, nil
	}

	if err == gorm.ErrRecordNotFound {
		// 默认策略：如果未显式配置，核心插件技能默认允许，MCP 工具默认需要显式授权
		if strings.Contains(skillName, "__") {
			return false, nil // MCP 工具默认禁用
		}
		return true, nil // 核心插件技能默认启用
	}

	return false, err
}

// ExecuteSkill 执行技能调用
func (sm *SkillManager) ExecuteSkill(ctx context.Context, botID string, userID uint, orgID uint, toolCall ai.ToolCall) (any, error) {
	name := toolCall.Function.Name
	clog.Info("[Skill] Executing skill for bot",
		zap.String("bot_id", botID),
		zap.String("skill", name))

	// 1. 权限检查
	allowed, err := sm.CheckPermission(ctx, botID, userID, orgID, name)
	if err != nil || !allowed {
		return nil, fmt.Errorf("permission denied for skill: %s", name)
	}

	// 2. 路由执行
	var args map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 如果是 MCP 工具 (包含 __ 分隔符)
	if strings.HasPrefix(name, "mcp__") || strings.Contains(name, "__") && !strings.HasPrefix(name, "b2b__") {
		return sm.mcpManager.CallTool(ctx, name, args)
	}

	// 如果是 B2B 共享技能 (标识为 b2b__entID__skillName)
	if strings.HasPrefix(name, "b2b__") {
		parts := strings.Split(name, "__")
		if len(parts) >= 3 && sm.b2bService != nil {
			var sourceEntID uint
			fmt.Sscanf(parts[1], "%d", &sourceEntID)
			realSkillName := strings.Join(parts[2:], "__")
			return sm.b2bService.CallRemoteTool(orgID, sourceEntID, realSkillName, args)
		}
		return nil, fmt.Errorf("invalid b2b skill format or b2b service not available")
	}

	// 如果是传统 Worker 插件技能
	return sm.manager.SyncSkillCall(ctx, name, args)
}
