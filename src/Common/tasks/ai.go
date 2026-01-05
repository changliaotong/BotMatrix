package tasks

import (
	log "BotMatrix/common/log"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
)

// AIParser AI 解析器
type AIParser struct {
	Manifest     *types.SystemManifest
	regexCache   map[string]*regexp.Regexp
	regexCacheMu sync.RWMutex
	aiService    types.AIService
}

func (a *AIParser) SetAIService(svc types.AIService) {
	a.aiService = svc
}

func (a *AIParser) GetAIService() types.AIService {
	return a.aiService
}

func (a *AIParser) GetManifest() *types.SystemManifest {
	return a.Manifest
}

func (a *AIParser) GetSystemPrompt() string {
	if a.Manifest == nil {
		return ""
	}
	return a.Manifest.GenerateSystemPrompt()
}

// MatchSkillByRegex 检查输入是否匹配任何已报备技能的正则表达式
func (a *AIParser) MatchSkillByRegex(input string) (*types.Capability, bool) {
	if a.Manifest == nil || len(a.Manifest.Skills) == 0 {
		return nil, false
	}

	a.regexCacheMu.RLock()
	defer a.regexCacheMu.RUnlock()

	for _, skill := range a.Manifest.Skills {
		if skill.Regex == "" {
			continue
		}

		re, ok := a.regexCache[skill.Name]
		if !ok {
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

// AIActionType, ParseResult were moved to types/ai.go

// MatchSkillByLLM 使用 LLM 进行语义匹配和参数提取
func (a *AIParser) MatchSkillByLLM(ctx context.Context, input string, modelID uint, parseCtx map[string]any) (*types.ParseResult, error) {
	if a.aiService == nil {
		return nil, fmt.Errorf("ai service not set")
	}

	ragContext := ""
	if a.Manifest != nil && a.Manifest.KnowledgeBase != nil {
		filter := &types.SearchFilter{Status: "active"}
		if parseCtx != nil {
			if bid, ok := parseCtx["bot_id"].(string); ok {
				filter.BotID = bid
			}
			if gid, ok := parseCtx["effective_group_id"].(string); ok && gid != "" {
				filter.OwnerType = "group"
				filter.OwnerID = gid
			} else if uid, ok := parseCtx["user_id"].(string); ok && uid != "" {
				filter.OwnerType = "user"
				filter.OwnerID = uid
			}
		}

		chunks, err := a.Manifest.KnowledgeBase.Search(ctx, input, 3, filter)
		if err == nil && len(chunks) > 0 {
			ragContext = "\n\n### 参考文档 (RAG):\n"
			for i, chunk := range chunks {
				ragContext += fmt.Sprintf("[%d] 来源: %s\n%s\n", i+1, chunk.Source, chunk.Content)
			}
			log.Printf("[AI-Task] RAG retrieved %d chunks for query: %s", len(chunks), input)
		}
	}

	systemPrompt := a.Manifest.GenerateSystemPrompt()
	if ragContext != "" {
		systemPrompt += "\n\n" +
			"### 知识库参考资料 (RAG)\n" +
			"以下是与用户请求相关的参考文档片段，请结合 these 信息来理解系统功能、回答用户疑问或提取参数。\n" +
			"如果参考资料中包含操作指南，请优先遵循指南中的步骤。\n" +
			" + ragContext"
	}
	if parseCtx != nil {
		contextInfo := "\n\n当前运行环境信息："
		if gid, ok := parseCtx["effective_group_id"].(string); ok && gid != "" {
			contextInfo += fmt.Sprintf("\n- 目标群组 ID: %s (如果用户未指定群组，请默认使用此 ID)", gid)
		}
		if isPrivate, ok := parseCtx["is_private"].(bool); ok {
			chatType := "群聊"
			if isPrivate {
				chatType = "私聊"
			}
			contextInfo += fmt.Sprintf("\n- 当前对话类型: %s", chatType)
		}
		if role, ok := parseCtx["user_role"].(string); ok && role != "" {
			contextInfo += fmt.Sprintf("\n- 当前用户角色: %s", role)
		}
		if botID, ok := parseCtx["bot_id"].(string); ok && botID != "" {
			contextInfo += fmt.Sprintf("\n- 当前机器人 ID: %s", botID)
		}
		systemPrompt += contextInfo
	}

	tools := a.Manifest.GenerateTools()
	messages := []types.Message{
		{
			Role:    types.RoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    types.RoleUser,
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
	analysis := ""
	if s, ok := choice.Message.Content.(string); ok {
		analysis = s
	} else if parts, ok := choice.Message.Content.([]types.ContentPart); ok {
		for _, p := range parts {
			if p.Type == "text" {
				analysis += p.Text
			}
		}
	}

	result := &types.ParseResult{
		Analysis: analysis,
	}

	if len(choice.Message.ToolCalls) > 0 {
		for i, toolCall := range choice.Message.ToolCalls {
			functionName := toolCall.Function.Name
			subResult := &types.ParseResult{
				Summary: fmt.Sprintf("动作 %d: %s", i+1, functionName),
			}

			switch functionName {
			case "create_task":
				subResult.Intent = types.AIActionCreateTask
			case "adjust_policy":
				subResult.Intent = types.AIActionAdjustPolicy
			case "manage_tags":
				subResult.Intent = types.AIActionManageTags
			case "cancel_task":
				subResult.Intent = types.AIActionCancelTask
			case "system_query":
				subResult.Intent = types.AIActionSystemQuery
			default:
				subResult.Intent = types.AIActionSkillCall
				subResult.Summary = fmt.Sprintf("调用技能: %s", functionName)
			}

			var args map[string]any
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			subResult.Data = args
			subResult.IsSafe = true

			result.SubActions = append(result.SubActions, subResult)
		}

		if len(result.SubActions) == 1 {
			result.Intent = result.SubActions[0].Intent
			result.Summary = result.SubActions[0].Summary
			result.Data = result.SubActions[0].Data
			result.IsSafe = true
		} else {
			result.Intent = types.AIActionBatch
			result.Summary = fmt.Sprintf("识别到 %d 个组合动作", len(result.SubActions))
			result.IsSafe = true
		}

		return result, nil
	}

	// 尝试解析结构化回复 (如果模型没有调用工具而是直接返回了 JSON)
	var structured struct {
		Intent   types.AIActionType `json:"intent"`
		Summary  string             `json:"summary"`
		Data     any                `json:"data"`
		Analysis string             `json:"analysis"`
		IsSafe   bool               `json:"is_safe"`
	}

	if err := json.Unmarshal([]byte(analysis), &structured); err == nil && structured.Intent != "" {
		result.Intent = structured.Intent
		result.Summary = structured.Summary
		result.Data = structured.Data
		result.Analysis = structured.Analysis
		result.IsSafe = structured.IsSafe
		return result, nil
	}

	result.Intent = types.AIActionSystemQuery
	result.Summary = analysis
	result.IsSafe = true

	return result, nil
}

func NewAIParser() *AIParser {
	return &AIParser{
		Manifest:   GetDefaultManifest(),
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// GetDefaultManifest 获取默认的系统功能清单
func GetDefaultManifest() *types.SystemManifest {
	return &types.SystemManifest{
		Version: "2.0.0",
		Identity: types.BotIdentity{
			Name:        "BotMatrix Core",
			Role:        "分布式机器人矩阵枢纽",
			Personality: "专业、高效、可靠、具备极强的多智能体协作能力",
			Knowledge: []string{
				"BotMatrix 是一个 AI 原生分布式机器人技能平台",
				"支持 MCP (Model Context Protocol) 协议集成工具",
				"具备 RAG 2.0 增强检索能力",
			},
		},
		Intents: map[string]string{
			"create_task":   "创建新的异步执行任务",
			"adjust_policy": "调整系统运行策略或安全规则",
			"system_query":  "查询系统运行状态或知识库",
			"skill_call":    "调用特定的业务技能或插件",
		},
		Actions:     make(map[string]types.Capability),
		Triggers:    make(map[string]types.Capability),
		GlobalRules: []string{"严禁泄露用户 PII 信息", "所有操作必须符合安全合规要求"},
		RAGEnabled:  true,
	}
}

// UpdateSkills 更新 Worker 报备的业务技能
func (a *AIParser) UpdateSkills(skills []types.Capability) {
	if a.Manifest == nil {
		a.Manifest = &types.SystemManifest{}
	}
	a.Manifest.Skills = skills

	a.regexCacheMu.Lock()
	defer a.regexCacheMu.Unlock()
	a.regexCache = make(map[string]*regexp.Regexp)
	for _, skill := range skills {
		if skill.Regex != "" {
			if re, err := regexp.Compile(skill.Regex); err == nil {
				a.regexCache[skill.Name] = re
			}
		}
	}
}
