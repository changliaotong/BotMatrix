package tasks

import (
	"BotMatrix/common/ai"
	"encoding/json"
	"fmt"
)

// Capability 定义系统的一个功能原子
type Capability struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Category     string            `json:"category"` // 新增：分类 (e.g., tools, entertainment, ai)
	Params       map[string]string `json:"params"`   // 参数名 -> 说明
	Required     []string          `json:"required"` // 新增：必填参数列表
	Example      string            `json:"example"`  // 示例输入
	IsEnterprise bool              `json:"is_enterprise"`
	Regex        string            `json:"regex"` // 新增：正则触发器
}

// ToTool 将 Capability 转换为 LLM 可识别的工具定义
func (c *Capability) ToTool() ai.Tool {
	properties := make(map[string]any)
	for name, desc := range c.Params {
		properties[name] = map[string]string{
			"type":        "string",
			"description": desc,
		}
	}

	return ai.Tool{
		Type: "function",
		Function: ai.FunctionDefinition{
			Name:        c.Name,
			Description: c.Description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   c.Required,
			},
		},
	}
}

// SystemManifest 系统功能清单
type SystemManifest struct {
	Version     string                `json:"version"`
	Intents     map[string]string     `json:"intents"`      // 意图 -> 说明
	Actions     map[string]Capability `json:"actions"`      // 动作 -> 详情
	Triggers    map[string]Capability `json:"triggers"`     // 触发方式 -> 详情
	Skills      []Capability          `json:"skills"`       // 业务技能 (由 Worker 提供)
	GlobalRules []string              `json:"global_rules"` // 全局约束
}

// GetDefaultManifest 获取系统默认功能清单
func GetDefaultManifest() *SystemManifest {
	return &SystemManifest{
		Version: "1.0",
		Intents: map[string]string{
			"create_task":   "创建一个新的自动化任务（如定时消息、自动禁言）",
			"adjust_policy": "调整系统全局策略（如维护模式、限流开关）",
			"manage_tags":   "管理群组或好友的标签",
			"system_query":  "咨询系统功能、状态或获取帮助",
		},
		Actions: map[string]Capability{
			"send_message": {
				Name:        "发送消息",
				Description: "向指定群组或好友发送文本消息",
				Params: map[string]string{
					"message":  "消息内容",
					"group_id": "目标群号 (可选)",
					"user_id":  "目标用户 ID (可选)",
				},
				Example: "每天早上 8 点给群 123 发送 '早上好'",
			},
			"mute_group": {
				Name:        "群禁言",
				Description: "开启全群禁言",
				Params: map[string]string{
					"group_id": "目标群号",
					"duration": "持续秒数 (0 为永久)",
				},
				Example: "晚上 11 点关闭群 123 的发言",
			},
			"unmute_group": {
				Name:        "取消群禁言",
				Description: "解除全群禁言",
				Params: map[string]string{
					"group_id": "目标群号",
				},
				Example: "早上 9 点开启群 123 的发言",
			},
		},
		Triggers: map[string]Capability{
			"cron": {
				Name:        "周期触发 (Cron)",
				Description: "使用标准 Cron 表达式定时触发",
				Params: map[string]string{
					"cron": "Cron 表达式 (例如: '0 23 * * *')",
				},
			},
			"once": {
				Name:        "一次性触发",
				Description: "在指定时间点触发一次",
				Params: map[string]string{
					"time": "ISO8601 时间字符串",
				},
			},
			"condition": {
				Name:         "条件触发",
				Description:  "当满足特定事件或关键词时触发",
				IsEnterprise: true,
				Params: map[string]string{
					"event":   "事件类型 (message, join, etc.)",
					"keyword": "触发关键词",
				},
			},
		},
		GlobalRules: []string{
			"所有任务执行必须生成 Execution 记录",
			"试用版仅支持单群、单标签任务",
			"企业版支持多标签组合和延迟任务",
			"AI 生成的指令必须经过人工确认后方可执行",
		},
	}
}

// GenerateTools 生成所有可用工具列表，用于 LLM Function Calling
func (m *SystemManifest) GenerateTools() []ai.Tool {
	var tools []ai.Tool

	// 1. 添加业务技能 (Skills)
	for _, skill := range m.Skills {
		tools = append(tools, skill.ToTool())
	}

	// 2. 添加核心动作 (Actions)
	for _, action := range m.Actions {
		tools = append(tools, action.ToTool())
	}

	return tools
}

// GenerateSystemPrompt 生成给 AI 的 System Prompt
func (m *SystemManifest) GenerateSystemPrompt() string {
	manifestJSON, _ := json.MarshalIndent(m, "", "  ")

	prompt := `你是一个 BotNexus 系统的智能技能机器人 (Skill Bot) 调度专家。
你的任务是理解用户的自然语言需求，并将其转化为系统可识别的结构化 JSON 指令。

### 核心能力:
1. **意图识别**: 准确判断用户是想创建任务、调整策略还是调用某个特定技能。
2. **参数提取**: 从对话中提取技能执行所需的参数，对于缺失的必填参数，需在输出中标记。
3. **技能路由**: 当用户需求匹配到 'skills' 清单中的项时，使用 'skill_call' 意图。

### 系统功能清单 (Capability Manifest):
%s

### 输出规范:
1. 必须返回合法的 JSON 格式。
2. 包含 'intent' (意图), 'summary' (中文摘要), 'data' (结构化参数), 'analysis' (你的推理过程)。
3. 如果是技能调用 (skill_call)，data 必须包含 'skill_name' 和 'params'。
4. 如果用户需求不明确或缺少必填参数，请在 analysis 中指出并请求补充。

### 示例转换 (技能调用):
用户: "帮我查询上海的天气"
输出: {
  "intent": "skill_call",
  "summary": "调用天气查询技能",
  "data": {
    "skill_name": "weather_query",
    "params": {"city": "上海"}
  },
  "analysis": "用户想要查询天气，匹配到 'weather_query' 技能，提取城市为上海。"
}
`
	return fmt.Sprintf(prompt, string(manifestJSON))
}
