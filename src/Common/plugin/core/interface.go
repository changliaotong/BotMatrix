package core

import (
	"BotMatrix/common/types"
	"time"
)

// Plugin 内部插件接口
type PluginModule interface {
	Name() string
	Description() string
	Version() string
	Init(robot Robot)
}

// Skill 插件提供的技能函数类型
type Skill func(ctx BaseContext, params map[string]string) (string, error)

// SkillCapability 描述插件提供的技能
type SkillCapability struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Usage       string            `json:"usage"`
	Params      map[string]string `json:"params"`
	Regex       string            `json:"regex"` // 新增：指令正则触发器
}

// SkillCapable 插件可选实现的接口，用于报备技能
type SkillCapable interface {
	GetSkills() []SkillCapability
}

// Robot 插件可调用的机器人接口
type Robot interface {
	// 基础事件监听 (由具体的适配器实现)
	// HandleAPI 处理 API 请求
	HandleAPI(action string, fn any)
	// CallBotAction 调用机器人动作
	CallBotAction(action string, params any) (any, error)
	// CallPluginAction 调用其他插件动作
	CallPluginAction(pluginID string, action string, payload map[string]any) (any, error)

	// 会话状态管理
	GetSessionContext(platform, userID string) (*types.SessionContext, error)
	SetSessionState(platform, userID string, state types.SessionState, ttl time.Duration) error
	GetSessionState(platform, userID string) (*types.SessionState, error)
	ClearSessionState(platform, userID string) error

	// 任务与技能管理
	HandleSkill(skillName string, fn func(ctx BaseContext, params map[string]string) (string, error))
	RegisterSkill(capability SkillCapability, fn func(ctx BaseContext, params map[string]string) (string, error))
}
