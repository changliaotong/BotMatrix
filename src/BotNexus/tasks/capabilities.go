package tasks

import (
	"BotMatrix/common/ai"
	"context"
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
	Regex        string            `json:"regex"`         // 新增：正则触发器
	RiskLevel    string            `json:"risk_level"`    // 新增：风险等级 (low, medium, high)
	DefaultRoles []string          `json:"default_roles"` // 新增：默认允许的角色 (owner, admin, member)
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

// GenerateSkillGuide 生成技能的操作指南 (用于 RAG 索引)
func (c *Capability) GenerateSkillGuide() string {
	guide := fmt.Sprintf("## 技能名称: %s\n", c.Name)
	guide += fmt.Sprintf("描述: %s\n", c.Description)
	if c.Category != "" {
		guide += fmt.Sprintf("分类: %s\n", c.Category)
	}
	if len(c.Params) > 0 {
		guide += "\n参数说明:\n"
		for name, desc := range c.Params {
			req := ""
			for _, r := range c.Required {
				if r == name {
					req = " (必填)"
					break
				}
			}
			guide += fmt.Sprintf("- %s: %s%s\n", name, desc, req)
		}
	}
	if c.Example != "" {
		guide += fmt.Sprintf("\n使用示例: %s\n", c.Example)
	}
	if c.RiskLevel != "" {
		guide += fmt.Sprintf("风险等级: %s\n", c.RiskLevel)
	}
	return guide
}

// BotIdentity 机器人身份定义 (自举核心)
type BotIdentity struct {
	Name        string            `json:"name"`        // 机器人名称
	Role        string            `json:"role"`        // 核心角色定位
	Personality string            `json:"personality"` // 性格特征描述
	Knowledge   []string          `json:"knowledge"`   // 核心背景知识 (静态注入)
	HowTo       map[string]string `json:"how_to"`      // 核心功能操作指南 (静态注入)
}

// DocChunk 文档片段
type DocChunk struct {
	ID      string  `json:"id"`
	Content string  `json:"content"`
	Source  string  `json:"source"`
	Score   float64 `json:"score"` // 新增：相似度评分
}

// SearchFilter 搜索过滤条件
type SearchFilter struct {
	OwnerType string // system, user, group, bot
	OwnerID   string // ID
	BotID     string // 机器人号码 (用于自动共享该机器人名下的知识)
	Status    string // active, paused
}

// KnowledgeBase 知识库接口
type KnowledgeBase interface {
	Search(ctx context.Context, query string, limit int, filter *SearchFilter) ([]DocChunk, error)
}

// SystemManifest 系统功能清单
type SystemManifest struct {
	Version       string                `json:"version"`
	Identity      BotIdentity           `json:"identity"`     // 新增：机器人身份
	Intents       map[string]string     `json:"intents"`      // 意图 -> 说明
	Actions       map[string]Capability `json:"actions"`      // 动作 -> 详情
	Triggers      map[string]Capability `json:"triggers"`     // 触发方式 -> 详情
	Skills        []Capability          `json:"skills"`       // 业务技能 (由 Worker 提供)
	GlobalRules   []string              `json:"global_rules"` // 全局约束
	KnowledgeBase KnowledgeBase         `json:"-"`            // 新增：外部知识库接口

	// RAGEnabled 是否启用知识库增强自我认知
	RAGEnabled bool `json:"rag_enabled"`
}

// GetDefaultManifest 获取系统默认功能清单
func GetDefaultManifest() *SystemManifest {
	return &SystemManifest{
		Version: "1.0",
		Identity: BotIdentity{
			Name:        "BotMatrix",
			Role:        "全能型群组自动化专家",
			Personality: "专业、高效、偶尔幽默，致力于通过自动化流程提升协作效率。",
			Knowledge: []string{
				"你是 BotMatrix 系统的核心 AI 调度器",
				"你可以创建定时任务（Cron）和一次性任务",
				"你可以管理群成员、设置管理员、执行禁言等操作",
				"你拥有扩展技能库，可以调用天气、翻译等外部服务",
				"你优先通过 Function Calling (工具调用) 来执行具体动作",
			},
			HowTo: map[string]string{
				"创建任务": "说出‘每天/每小时/在某时间点 [做某事]’即可。例如：‘每天早上8点提醒我开会’。",
				"取消任务": "说出‘取消任务 [ID]’或‘停止 [任务描述]’。例如：‘取消 ID 为 101 的任务’。",
				"群管理":  "可以直接要求禁言、踢人或设置管理员。例如：‘把群禁言 10 分钟’。",
				"技能调用": "直接描述需求，如‘查一下北京的天气’、‘把这句话翻译成英文’。",
				"帮助查询": "问‘你能做什么’或‘如何使用 [功能]’即可获得详细说明。",
			},
		},
		Intents: map[string]string{
			"create_task":  "创建定时或提醒任务",
			"adjust_rule":  "调整系统规则或配置",
			"manage_group": "执行群管理动作（禁言、踢人等）",
			"system_query": "查询系统状态、自己的能力或获取帮助",
			"skill_call":   "调用外部扩展技能",
		},
		RAGEnabled: true,
		Actions: map[string]Capability{
			"cancel_task": {
				Name:         "取消任务",
				Description:  "根据任务 ID 或描述取消一个正在运行或等待中的任务",
				RiskLevel:    "medium",
				DefaultRoles: []string{"owner", "admin"},
				Params: map[string]string{
					"task_id": "要取消的任务 ID",
				},
				Example: "取消 ID 为 123 的任务",
			},
			"system_query": {
				Name:         "系统查询与帮助",
				Description:  "向机器人咨询其身份、功能、状态或获取操作指南",
				RiskLevel:    "low",
				DefaultRoles: []string{"owner", "admin", "member"},
				Params: map[string]string{
					"query": "查询关键词 (如 identity, capabilities, help)",
				},
				Example: "你能做什么？",
			},
			"send_message": {
				Name:         "发送消息",
				Description:  "向指定群组或好友发送文本消息",
				RiskLevel:    "low",
				DefaultRoles: []string{"owner", "admin", "member"},
				Params: map[string]string{
					"message":  "消息内容",
					"group_id": "目标群号 (可选)",
					"user_id":  "目标用户 ID (可选)",
				},
				Example: "每天早上 8 点给群 123 发送 '早上好'",
			},
			"mute_group": {
				Name:         "群禁言",
				Description:  "开启全群禁言",
				RiskLevel:    "medium",
				DefaultRoles: []string{"owner", "admin"},
				Params: map[string]string{
					"group_id": "目标群号",
					"duration": "持续秒数 (0 为永久)",
				},
				Example: "晚上 11 点关闭群 123 的发言",
			},
			"unmute_group": {
				Name:         "取消群禁言",
				Description:  "解除全群禁言",
				RiskLevel:    "medium",
				DefaultRoles: []string{"owner", "admin"},
				Params: map[string]string{
					"group_id": "目标群号",
				},
				Example: "早上 9 点开启群 123 的发言",
			},
			"kick_member": {
				Name:         "踢出成员",
				Description:  "将指定用户移出群组",
				RiskLevel:    "high",
				DefaultRoles: []string{"owner", "admin"},
				Params: map[string]string{
					"group_id":           "目标群号",
					"user_id":            "目标用户 ID",
					"reject_add_request": "是否拒绝再次申请 (true/false)",
				},
				Example: "把群 123 里的坏人 456 踢了",
			},
			"set_group_admin": {
				Name:         "设置管理员",
				Description:  "设置或取消群管理员",
				RiskLevel:    "high",
				DefaultRoles: []string{"owner"},
				Params: map[string]string{
					"group_id": "目标群号",
					"user_id":  "目标用户 ID",
					"enable":   "是否设置为管理员 (true/false)",
				},
				Example: "把用户 789 设为群 123 的管理员",
			},
			"mute_random": {
				Name:         "随机禁言套餐",
				Description:  "从群成员中随机抽取幸运儿进行禁言",
				RiskLevel:    "medium",
				DefaultRoles: []string{"owner", "admin"},
				Params: map[string]string{
					"group_id": "目标群号",
					"duration": "禁言时长 (秒)",
					"count":    "抽取人数 (默认为 1)",
					"smart":    "是否启用智能模式 (true/false)，优先禁言最近发言的人",
				},
				Example: "给群 123 安排个智能随机禁言套餐，禁言 60 秒",
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

	ragNotice := ""
	if m.RAGEnabled {
		ragNotice = "\n5. **深度知识检索**: 你拥有一个强大的 RAG 知识库。如果用户的问题涉及到复杂的系统架构、详细的操作流程或你内置知识中没有涵盖的内容，请使用 'system_query' 意图，并在分析中说明你需要检索知识库。"
	}

	prompt := `你现在的身份是：%s (%s)。
你的性格特征是：%s。

你的核心背景知识：
%s

### 操作指南 (How-to Guide):
%s

你的职责是理解用户的自然语言需求，并将其转化为系统可识别的结构化 JSON 指令。

### 核心能力:
1. **意图识别**: 准确判断用户是想创建任务、调整策略、管理标签、查询系统还是调用某个特定技能。
2. **自我认知与帮助**: 当用户询问“你是谁”、“你能做什么”或请求“如何使用 [功能]”时，请优先从内置的“核心背景知识”和“操作指南”中回答。
3. **参数提取**: 从对话中提取技能执行所需的参数，对于缺失的必填参数，需在输出中标记。
4. **技能路由**: 当用户需求匹配到 'skills' 清单中的项时，使用 'skill_call' 意图。%s

### 系统功能清单 (Capability Manifest):
%s

### 输出规范:
1. 必须返回合法的 JSON 格式。
2. 包含 'intent' (意图), 'summary' (中文摘要), 'data' (结构化参数), 'analysis' (你的推理过程)。
3. 如果是系统查询 (system_query)，data 可以包含查询关键词或为空，主要通过 analysis 回复。
4. 如果是技能调用 (skill_call)，data 必须包含 'skill' 和 'params'。
5. 如果是创建任务 (create_task)，data 必须符合任务结构：{"name": "...", "type": "cron/once", "action_type": "...", "action_params": "JSON字符串", "trigger_config": "JSON字符串"}。

### 示例转换 (自我认知与帮助):
用户: "你是谁？"
输出: {
  "intent": "system_query",
  "summary": "自我介绍",
  "data": {"query": "identity"},
  "analysis": "你好！我是 BotMatrix，你的全能型群组自动化专家。我擅长管理群务、设置定时任务以及调用各种实用技能来提升你的群聊体验。你可以问我‘你能做什么’来了解详情。"
}

用户: "如何创建定时任务？"
输出: {
  "intent": "system_query",
  "summary": "功能使用指导",
  "data": {"query": "how_to_create_task"},
  "analysis": "创建定时任务非常简单！你只需要告诉我你想要在什么时间做某事。例如：‘每天早上8点提醒我开会’或‘每周五下午5点把群禁言’。我会自动为您生成任务草稿，您确认后即可生效。"
}
`

	knowledgeStr := ""
	for _, k := range m.Identity.Knowledge {
		knowledgeStr += fmt.Sprintf("- %s\n", k)
	}

	howtoStr := ""
	for k, v := range m.Identity.HowTo {
		howtoStr += fmt.Sprintf("- **%s**: %s\n", k, v)
	}

	return fmt.Sprintf(prompt,
		m.Identity.Name, m.Identity.Role, m.Identity.Personality,
		knowledgeStr, howtoStr, ragNotice, string(manifestJSON))
}
