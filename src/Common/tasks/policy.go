package tasks

import (
	"BotMatrix/common/types"
	"fmt"
	"strings"
)

// UserContext 执行 AI 任务的用户上下文
type UserContext struct {
	UserID  string
	GroupID string
	Role    string // owner, admin, member
}

// PolicyResult 策略检查结果
type PolicyResult struct {
	Allowed bool
	Reason  string
}

// CheckCapabilityPolicy 检查用户是否有权使用特定能力
func CheckCapabilityPolicy(manifest *types.SystemManifest, actionType string, ctx UserContext) PolicyResult {
	if manifest == nil {
		return PolicyResult{Allowed: true}
	}
	capability, ok := manifest.Actions[actionType]
	if !ok {
		return PolicyResult{Allowed: true} // 默认允许未知动作（或由后续逻辑处理）
	}

	// 1. 检查角色权限
	isAllowedRole := false
	for _, role := range capability.DefaultRoles {
		if ctx.Role == role {
			isAllowedRole = true
			break
		}
	}

	if !isAllowedRole {
		return PolicyResult{
			Allowed: false,
			Reason:  fmt.Sprintf("权限不足：该功能 [%s] 仅限 %s 使用，您的角色是 %s", capability.Name, strings.Join(capability.DefaultRoles, "/"), ctx.Role),
		}
	}

	// 2. 高风险动作额外提示
	if capability.RiskLevel == "high" && ctx.Role != "owner" {
		// 管理员虽然可以执行 high 风险操作，但可能需要更严格的审计或限制
		// 这里可以预留二次验证逻辑
	}

	return PolicyResult{Allowed: true}
}
