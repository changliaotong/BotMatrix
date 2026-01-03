package tasks

import (
	"BotMatrix/common/ai"
	log "BotMatrix/common/log"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// AIService 定义任务系统需要的 AI 能力接口
type AIService interface {
	Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error)
	CreateEmbedding(ctx context.Context, modelID uint, input any) (*ai.EmbeddingResponse, error)
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
	AIActionSkillCall    AIActionType = "skill_call"  // 新增：技能调用
	AIActionCancelTask   AIActionType = "cancel_task" // 新增：取消任务
	AIActionBatch        AIActionType = "batch_task"  // 新增：批量/多重任务并行
)

// ParseRequest AI 解析请求
type ParseRequest struct {
	Input      string         `json:"input"`
	ActionType AIActionType   `json:"action_type"` // 可选，明确指定意图
	Context    map[string]any `json:"context"`     // 上下文信息
}

// MatchSkillByLLM 使用 LLM 进行语义匹配和参数提取
func (a *AIParser) MatchSkillByLLM(ctx context.Context, input string, modelID uint, parseCtx map[string]any) (*ParseResult, error) {
	if a.aiService == nil {
		return nil, fmt.Errorf("ai service not set")
	}

	// 1. 如果有知识库，先进行检索增强 (RAG)
	ragContext := ""
	if a.Manifest != nil && a.Manifest.KnowledgeBase != nil {
		// 构造搜索过滤器
		filter := &SearchFilter{Status: "active"}
		if parseCtx != nil {
			if bid, ok := parseCtx["bot_id"].(string); ok {
				filter.BotID = bid
			}
			// 优先匹配群组知识，其次是用户个人知识
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
			"以下是与用户请求相关的参考文档片段，请结合这些信息来理解系统功能、回答用户疑问或提取参数。\n" +
			"如果参考资料中包含操作指南，请优先遵循指南中的步骤。\n" +
			ragContext
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
	messages := []ai.Message{
		{
			Role:    ai.RoleSystem,
			Content: systemPrompt,
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
	analysis := ""
	if s, ok := choice.Message.Content.(string); ok {
		analysis = s
	} else if parts, ok := choice.Message.Content.([]ai.ContentPart); ok {
		for _, p := range parts {
			if p.Type == "text" {
				analysis += p.Text
			}
		}
	}

	result := &ParseResult{
		Analysis: analysis,
	}

	// 1. 优先尝试解析 ToolCalls (Function Calling 模式)
	if len(choice.Message.ToolCalls) > 0 {
		log.Printf("[AI-Task] LLM returned %d tool calls", len(choice.Message.ToolCalls))

		// 如果有多个工具调用，递归处理或循环处理
		for i, toolCall := range choice.Message.ToolCalls {
			functionName := toolCall.Function.Name
			subResult := &ParseResult{
				Summary: fmt.Sprintf("动作 %d: %s", i+1, functionName),
			}

			// 根据函数名映射意图
			switch functionName {
			case "create_task":
				subResult.Intent = AIActionCreateTask
			case "adjust_policy":
				subResult.Intent = AIActionAdjustPolicy
			case "manage_tags":
				subResult.Intent = AIActionManageTags
			case "cancel_task":
				subResult.Intent = AIActionCancelTask
			case "system_query":
				subResult.Intent = AIActionSystemQuery
			default:
				subResult.Intent = AIActionSkillCall
			}

			var params map[string]any
			json.Unmarshal([]byte(toolCall.Function.Arguments), &params)

			if subResult.Intent == AIActionSkillCall {
				subResult.Data = map[string]any{
					"skill":  functionName,
					"params": params,
				}
			} else {
				subResult.Data = params
			}
			subResult.IsSafe = true

			if i == 0 {
				// 第一个作为主结果
				result.Intent = subResult.Intent
				result.Summary = subResult.Summary
				result.Data = subResult.Data
				result.IsSafe = subResult.IsSafe
			} else {
				// 后续作为子动作
				result.SubActions = append(result.SubActions, subResult)
			}
		}

		if len(result.SubActions) > 0 {
			result.Summary = fmt.Sprintf("多重任务指令 (%d 个动作)", len(result.SubActions)+1)
		}

		return result, nil
	}

	// 2. 尝试从 Content 中解析 JSON (Structured Output 模式)
	content := analysis
	// 简单清洗 content，防止 AI 返回包含 Markdown 代码块
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var structured struct {
		Intent   AIActionType   `json:"intent"`
		Summary  string         `json:"summary"`
		Data     map[string]any `json:"data"`
		Analysis string         `json:"analysis"`
		IsSafe   bool           `json:"is_safe"`
	}

	if err := json.Unmarshal([]byte(content), &structured); err == nil && structured.Intent != "" {
		result.Intent = structured.Intent
		result.Summary = structured.Summary
		result.Data = structured.Data
		result.Analysis = structured.Analysis
		result.IsSafe = structured.IsSafe
		return result, nil
	}

	// 3. 如果都不是，则作为一般查询处理
	result.Intent = AIActionSystemQuery
	result.Summary = analysis
	result.IsSafe = true

	return result, nil
}

// ParseResult AI 解析结果
type ParseResult struct {
	DraftID    string         `json:"draft_id"` // 新增 DraftID
	Intent     AIActionType   `json:"intent"`
	Summary    string         `json:"summary"`
	Data       any            `json:"data"`        // 解析出的结构化数据
	IsSafe     bool           `json:"is_safe"`     // 是否安全（低风险）
	Analysis   string         `json:"analysis"`    // AI 的推理过程
	SubActions []*ParseResult `json:"sub_actions"` // 新增：支持多任务并行
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

	// 如果设置了 AI 服务，优先使用 LLM 进行深度解析
	if a.aiService != nil {
		modelID := uint(1) // 默认使用 ID 为 1 的模型，实际应从配置获取
		if val, ok := req.Context["model_id"].(float64); ok {
			modelID = uint(val)
		} else if val, ok := req.Context["model_id"].(int); ok {
			modelID = uint(val)
		}

		result, err := a.MatchSkillByLLM(context.Background(), input, modelID, req.Context)
		if err == nil && result.Intent != AIActionSystemQuery {
			return result, nil
		}
		// 如果 LLM 解析失败或识别为一般查询，则尝试正则和模拟解析作为兜底
	}

	// 1. 意图识别 (模拟/正则兜底)
	intent := req.ActionType
	if intent == "" {
		intent = a.recognizeIntent(input)
	}

	// 2. 根据意图进行结构化解析
	switch intent {
	case AIActionCreateTask:
		return a.parseTaskCreation(req)
	case AIActionAdjustPolicy:
		return a.parsePolicyAdjustment(req)
	case AIActionManageTags:
		return a.parseTagManagement(req)
	case AIActionSkillCall:
		return a.parseSkillCall(req)
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
	if containsOne(input, "提醒", "定时", "每天", "任务", "禁言", "踢人") {
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

func (a *AIParser) parseSkillCall(req ParseRequest) (*ParseResult, error) {
	input := req.Input
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

func (a *AIParser) parseTaskCreation(req ParseRequest) (*ParseResult, error) {
	input := req.Input
	botID, _ := req.Context["bot_id"].(string)
	groupID, _ := req.Context["group_id"].(string)

	// 基础回退逻辑：尝试通过正则提取关键信息
	res := &ParseResult{
		Intent:  AIActionCreateTask,
		IsSafe:  true,
		Summary: "创建自动化任务",
	}

	// 提取群号 (假设群号是 5-12 位数字)
	groupReg := regexp.MustCompile(`(?:群|group)\s*(\d{5,12})`)
	if match := groupReg.FindStringSubmatch(input); len(match) > 1 {
		groupID = match[1]
	}

	if containsOne(input, "禁言") {
		if containsOne(input, "随机") || containsOne(input, "套餐") {
			isSmart := containsOne(input, "智能", "最近", "活跃")
			res.Data = map[string]any{
				"name":           "AI 生成: 随机禁言套餐",
				"type":           "once",
				"action_type":    "mute_random",
				"action_params":  fmt.Sprintf(`{"bot_id": "%s", "group_id": "%s", "duration": 60, "count": 1, "smart": %v}`, botID, groupID, isSmart),
				"trigger_config": fmt.Sprintf(`{"time": "%s"}`, time.Now().Add(10*time.Second).Format(time.RFC3339)),
			}
			res.Analysis = "识别到您想要发起“随机禁言套餐”。我为您编排了一个 10 秒后执行的任务：从当前群成员中随机抽取 1 人，禁言 60 秒。此操作仅限群主或管理员确认执行。"
			if isSmart {
				res.Analysis += " 已启用智能模式，将优先从最近发言的活跃成员中抽取。"
			}
		} else {
			duration := 0
			if strings.Contains(input, "分钟") {
				duration = 60
			}
			res.Data = map[string]any{
				"name":           "AI 生成: 自动禁言",
				"type":           "cron",
				"action_type":    "mute_group",
				"action_params":  fmt.Sprintf(`{"bot_id": "%s", "group_id": "%s", "duration": %d}`, botID, groupID, duration),
				"trigger_config": `{"cron": "0 23 * * *"}`,
			}
			res.Analysis = "识别到'禁言'需求，已尝试提取群号，默认设置为每天 23:00 执行。"
		}
	} else if containsOne(input, "踢", "移除") {
		res.Data = map[string]any{
			"name":           "AI 生成: 移除成员",
			"type":           "once",
			"action_type":    "kick_member",
			"action_params":  fmt.Sprintf(`{"bot_id": "%s", "group_id": "%s", "user_id": ""}`, botID, groupID),
			"trigger_config": `{"time": "2026-01-01T00:00:00Z"}`,
		}
		res.Analysis = "识别到'移除成员'需求，请补充目标用户 ID。"
	} else {
		res.Data = map[string]any{
			"name":           "AI 生成: 消息提醒",
			"type":           "once",
			"action_type":    "send_message",
			"action_params":  fmt.Sprintf(`{"bot_id": "%s", "group_id": "%s", "message": "提醒内容"}`, botID, groupID),
			"trigger_config": `{"time": "2026-01-01T00:00:00Z"}`,
		}
		res.Analysis = "识别到'消息提醒'需求，默认设置了一次性任务，请确认内容和时间。"
	}
	return res, nil
}

func (a *AIParser) parsePolicyAdjustment(req ParseRequest) (*ParseResult, error) {
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

func (a *AIParser) parseTagManagement(req ParseRequest) (*ParseResult, error) {
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
