package tasks

import (
	"encoding/json"
	"fmt"
)

// Capability 定义系统的一个功能原子
type Capability struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Params       map[string]string `json:"params"`  // 参数名 -> 说明
	Example      string            `json:"example"` // 示例输入
	IsEnterprise bool              `json:"is_enterprise"`
}

// SystemManifest 系统功能清单
type SystemManifest struct {
	Version      string                `json:"version"`
	Intents      map[string]string     `json:"intents"`      // 意图 -> 说明
	Actions      map[string]Capability `json:"actions"`      // 动作 -> 详情
	Triggers     map[string]Capability `json:"triggers"`     // 触发方式 -> 详情
	Skills       []Capability          `json:"skills"`       // 业务技能 (由 Worker 提供)
	GlobalRules  []string              `json:"global_rules"` // 全局约束
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

// GenerateSystemPrompt 生成给 AI 的 System Prompt
func (m *SystemManifest) GenerateSystemPrompt() string {
	manifestJSON, _ := json.MarshalIndent(m, "", "  ")

	prompt := `你是一个 BotNexus 系统的智能调度助手。
你的任务是理解用户的自然语言需求，并将其转化为系统可识别的结构化 JSON 指令。

### 系统功能清单 (Capability Manifest):
%s

### 输出规范:
1. 必须返回合法的 JSON 格式。
2. 包含 'intent' (意图), 'summary' (中文摘要), 'data' (结构化参数), 'analysis' (你的推理过程)。
3. 如果用户需求不明确，请在 analysis 中指出并请求补充。
4. 始终考虑安全边界，对于高危操作（如全群禁言）需在 summary 中明确提醒。

### 示例转换:
用户: "帮我设置每天晚上11点禁言群 123"
输出: {
  "intent": "create_task",
  "summary": "创建每天 23:00 自动禁言群 123 的任务",
  "data": {
    "name": "夜间自动禁言",
    "type": "cron",
    "action_type": "mute_group",
    "action_params": "{\"group_id\": \"123\", \"duration\": 0}",
    "trigger_config": "{\"cron\": \"0 23 * * *\"}"
  },
  "analysis": "用户要求定时禁言，识别为 create_task 意图，使用 cron 触发器。"
}
`
	return fmt.Sprintf(prompt, string(manifestJSON))
}
