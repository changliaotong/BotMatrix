package tasks

import (
	"BotMatrix/common/ai"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// AIService 定义任务系统需要的 AI 能力接口
type AIService interface {
	Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error)
}

// AIParser AI 解析器
type AIParser struct {
	Manifest     *SystemManifest
	regexCache   map[string]*regexp.Regexp
	regexCacheMu sync.RWMutex
	aiService    AIService
}

func (a *AIParser) SetAIService(svc AIService) {
	a.aiService = svc
}

// MatchSkillByRegex 检查输入是否匹配任何已报备技能的正则表达式
func (a *AIParser) MatchSkillByRegex(input string) (*Capability, bool) {
	if a.Manifest == nil || len(a.Manifest.Skills) == 0 {
		return nil, false
	}

	a.regexCacheMu.RLock()
	defer a.regexCacheMu.RUnlock()

	for _, skill := range a.Manifest.Skills {
		if skill.Regex == "" {
			continue
		}

		// 优先从缓存获取已编译的正则
		re, ok := a.regexCache[skill.Name]
		if !ok {
			// 如果没在缓存中，则现场编译（理论上 UpdateSkills 会预编译，这里做个兜底）
			var err error
			re, err = regexp.Compile(skill.Regex)
			if err != nil {
				continue
			}
		}

		if re.MatchString(input) {
			return &skill, true
		}
	}

	return nil, false
}

// AIActionType AI 解析的目标动作类型
type AIActionType string

const (
	AIActionCreateTask   AIActionType = "create_task"
	AIActionAdjustPolicy AIActionType = "adjust_policy"
	AIActionManageTags   AIActionType = "manage_tags"
	AIActionSystemQuery  AIActionType = "system_query"
	AIActionSkillCall    AIActionType = "skill_call" // 新增：技能调用
)

// ParseRequest AI 解析请求
type ParseRequest struct {
	Input      string         `json:"input"`
	ActionType AIActionType   `json:"action_type"` // 可选，明确指定意图
	Context    map[string]any `json:"context"`     // 上下文信息
}

// MatchSkillByLLM 使用 LLM 进行语义匹配和参数提取
func (a *AIParser) MatchSkillByLLM(ctx context.Context, input string, modelID uint) (*ParseResult, error) {
	if a.aiService == nil {
		return nil, fmt.Errorf("ai service not set")
	}

	tools := a.Manifest.GenerateTools()
	messages := []ai.Message{
		{
			Role:    ai.RoleSystem,
			Content: a.Manifest.GenerateSystemPrompt(),
		},
		{
			Role:    ai.RoleUser,
			Content: input,
		},
	}

	resp, err := a.aiService.Chat(ctx, modelID, messages, tools)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from ai")
	}

	choice := resp.Choices[0]
	result := &ParseResult{
		Analysis: choice.Message.Content,
	}

	if len(choice.Message.ToolCalls) > 0 {
		toolCall := choice.Message.ToolCalls[0]
		result.Intent = AIActionSkillCall
		result.Summary = fmt.Sprintf("调用技能: %s", toolCall.Function.Name)

		var args map[string]any
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		result.Data = map[string]any{
			"skill": toolCall.Function.Name,
			"args":  args,
		}
	} else {
		result.Intent = AIActionSystemQuery
		result.Summary = choice.Message.Content
	}

	return result, nil
}

// ParseResult AI 解析结果
type ParseResult struct {
	DraftID  string       `json:"draft_id"` // 新增 DraftID
	Intent   AIActionType `json:"intent"`
	Summary  string       `json:"summary"`
	Data     any          `json:"data"` // 解析出的结构化数据
	IsSafe   bool         `json:"is_safe"`
	Analysis string       `json:"analysis"` // AI 的推理过程
}

func NewAIParser() *AIParser {
	return &AIParser{
		Manifest:   GetDefaultManifest(),
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// GetSystemPrompt 获取系统提示词，用于喂给大模型
func (a *AIParser) GetSystemPrompt() string {
	return a.Manifest.GenerateSystemPrompt()
}

// UpdateSkills 更新 Worker 报备的业务技能
func (a *AIParser) UpdateSkills(skills []Capability) {
	a.Manifest.Skills = skills

	// 预编译正则并存入缓存
	newCache := make(map[string]*regexp.Regexp)
	for _, skill := range skills {
		if skill.Regex != "" {
			if re, err := regexp.Compile(skill.Regex); err == nil {
				newCache[skill.Name] = re
			}
		}
	}

	a.regexCacheMu.Lock()
	a.regexCache = newCache
	a.regexCacheMu.Unlock()
}

// Parse 统一 AI 解析接口
func (a *AIParser) Parse(req ParseRequest) (*ParseResult, error) {
	input := req.Input

	// 1. 意图识别 (模拟)
	intent := req.ActionType
	if intent == "" {
		intent = a.recognizeIntent(input)
	}

	// 2. 根据意图进行结构化解析 (模拟)
	switch intent {
	case AIActionCreateTask:
		return a.parseTaskCreation(input)
	case AIActionAdjustPolicy:
		return a.parsePolicyAdjustment(input)
	case AIActionManageTags:
		return a.parseTagManagement(input)
	case "skill_call":
		return a.parseSkillCall(input)
	default:
		return &ParseResult{
			Intent:   AIActionSystemQuery,
			Summary:  "系统咨询",
			Analysis: "识别为一般性咨询，建议查看文档。",
			IsSafe:   true,
		}, nil
	}
}

func (a *AIParser) recognizeIntent(input string) AIActionType {
	if containsOne(input, "提醒", "定时", "每天", "任务") {
		return AIActionCreateTask
	}
	if containsOne(input, "维护", "模式", "限制", "策略", "开关") {
		return AIActionAdjustPolicy
	}
	if containsOne(input, "打标签", "分类", "标记", "标签") {
		return AIActionManageTags
	}
	// 检查是否匹配已知技能
	for _, skill := range a.Manifest.Skills {
		if strings.Contains(input, skill.Name) || strings.Contains(input, skill.Description) {
			return AIActionSkillCall
		}
	}
	return AIActionSystemQuery
}

func (a *AIParser) parseSkillCall(input string) (*ParseResult, error) {
	// 模拟 AI 提取技能参数
	var matchedSkill *Capability
	for _, skill := range a.Manifest.Skills {
		if strings.Contains(input, skill.Name) {
			matchedSkill = &skill
			break
		}
	}

	if matchedSkill == nil {
		return &ParseResult{
			Intent:   AIActionSystemQuery,
			Summary:  "未能匹配到具体技能",
			Analysis: "用户意图似乎是调用技能，但未能在当前在线 Worker 中找到匹配项。",
			IsSafe:   true,
		}, nil
	}

	return &ParseResult{
		Intent:  AIActionSkillCall,
		Summary: fmt.Sprintf("调用技能: %s", matchedSkill.Name),
		Data: map[string]any{
			"skill_name": matchedSkill.Name,
			"params":     map[string]any{}, // 实际应由 LLM 提取
		},
		IsSafe:   true,
		Analysis: fmt.Sprintf("识别到用户想要执行 '%s' 技能。", matchedSkill.Name),
	}, nil
}

func (a *AIParser) parseTaskCreation(input string) (*ParseResult, error) {
	// 模拟解析逻辑
	res := &ParseResult{
		Intent:  AIActionCreateTask,
		IsSafe:  true,
		Summary: "创建自动化任务",
	}

	if containsOne(input, "禁言") {
		res.Data = map[string]any{
			"name":           "AI 生成: 自动禁言",
			"type":           "cron",
			"action_type":    "mute_group",
			"action_params":  `{"duration": 0}`,
			"trigger_config": `{"cron": "0 23 * * *"}`,
		}
		res.Analysis = "识别到'禁言'需求，建议设置为每天 23:00 执行。"
	} else {
		res.Data = map[string]any{
			"name":           "AI 生成: 消息提醒",
			"type":           "once",
			"action_type":    "send_message",
			"action_params":  `{"message": "AI 提醒内容"}`,
			"trigger_config": `{"time": "2025-12-24T00:00:00Z"}`,
		}
		res.Analysis = "识别到'消息提醒'需求，默认设置了一次性任务。"
	}
	return res, nil
}

func (a *AIParser) parsePolicyAdjustment(input string) (*ParseResult, error) {
	return &ParseResult{
		Intent:  AIActionAdjustPolicy,
		IsSafe:  true,
		Summary: "调整全局策略",
		Data: map[string]any{
			"strategy_name": "maintenance_mode",
			"action":        "enable",
			"duration":      "3600",
		},
		Analysis: "识别到需要进入'维护模式'，建议开启 1 小时。",
	}, nil
}

func (a *AIParser) parseTagManagement(input string) (*ParseResult, error) {
	return &ParseResult{
		Intent:  AIActionManageTags,
		IsSafe:  true,
		Summary: "批量管理标签",
		Data: map[string]any{
			"target_type": "group",
			"tag_name":    "VIP",
			"condition":   "all_active",
		},
		Analysis: "识别到标签管理需求，建议对活跃群组批量标记 'VIP'。",
	}, nil
}

func contains(s string, keywords ...string) bool {
	for _, k := range keywords {
		if !strings.Contains(s, k) {
			return false
		}
	}
	return true
}

func containsOne(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// SimulateExecution 模拟执行
func (a *AIParser) SimulateExecution(taskJSON string) (string, error) {
	var task Task
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		return "", err
	}

	// 生成模拟报告
	report := fmt.Sprintf("模拟执行报告:\n- 任务名称: %s\n- 任务类型: %s\n- 预计下次触发: %s\n- 动作: %s\n- 影响范围: 预计影响 1 个群组",
		task.Name, task.Type, "2025-12-23 23:00:00", task.ActionType)

	return report, nil
}
