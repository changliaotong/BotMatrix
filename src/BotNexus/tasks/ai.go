package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AIParser AI 解析器
type AIParser struct {
	Manifest *SystemManifest
}

// AIActionType AI 解析的目标动作类型
type AIActionType string

const (
	AIActionCreateTask   AIActionType = "create_task"
	AIActionAdjustPolicy AIActionType = "adjust_policy"
	AIActionManageTags   AIActionType = "manage_tags"
	AIActionSystemQuery  AIActionType = "system_query"
)

// ParseRequest AI 解析请求
type ParseRequest struct {
	Input      string         `json:"input"`
	ActionType AIActionType   `json:"action_type"` // 可选，明确指定意图
	Context    map[string]any `json:"context"`     // 上下文信息
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
		Manifest: GetDefaultManifest(),
	}
}

// GetSystemPrompt 获取系统提示词，用于喂给大模型
func (a *AIParser) GetSystemPrompt() string {
	return a.Manifest.GenerateSystemPrompt()
}

// UpdateSkills 更新 Worker 报备的业务技能
func (a *AIParser) UpdateSkills(skills []Capability) {
	a.Manifest.Skills = skills
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
	return AIActionSystemQuery
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

func (a *AIParser) parseSkillCall(input string) (*ParseResult, error) {
	// 模拟解析逻辑：从 Skills 清单中匹配
	// 实际应由 LLM 根据 SystemPrompt 生成
	return &ParseResult{
		Intent:   "skill_call",
		Summary:  "调用业务技能",
		Data:     map[string]any{"skill": "weather", "params": map[string]string{"city": "北京"}},
		Analysis: "用户请求业务功能，匹配到 Worker 报备的技能。",
		IsSafe:   true,
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
