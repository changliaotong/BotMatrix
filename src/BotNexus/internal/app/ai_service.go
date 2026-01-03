package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type AIServiceImpl struct {
	db              *gorm.DB
	manager         *Manager
	clientsByConfig map[string]ai.Client
	mu              sync.RWMutex
	privacyFilter   *ai.PrivacyFilter
	mcpManager      *MCPManager
	skillManager    *SkillManager
	memoryService   CognitiveMemoryService
	employeeService DigitalEmployeeService
}

func NewAIService(db *gorm.DB, m *Manager) *AIServiceImpl {
	mcp := NewMCPManager(db, m)
	return &AIServiceImpl{
		db:              db,
		manager:         m,
		clientsByConfig: make(map[string]ai.Client),
		privacyFilter:   ai.NewPrivacyFilter(),
		mcpManager:      mcp,
		skillManager:    NewSkillManager(db, m, mcp),
		memoryService:   NewCognitiveMemoryService(db),
		employeeService: NewEmployeeService(db),
	}
}

func (s *AIServiceImpl) GetProvider(id uint) (*models.AIProviderGORM, error) {
	var provider models.AIProviderGORM
	if err := s.db.First(&provider, id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *AIServiceImpl) getClient(provider *models.AIProviderGORM, model *models.AIModelGORM) (ai.Client, error) {
	baseURL := provider.BaseURL
	apiKey := provider.APIKey

	// 模型级别覆盖
	if model != nil {
		if model.BaseURL != "" {
			baseURL = model.BaseURL
		}
		if model.APIKey != "" {
			apiKey = model.APIKey
		}
	}

	// 使用 URL + Key 的哈希作为缓存键
	cacheKey := fmt.Sprintf("%s|%s", baseURL, apiKey)

	s.mu.RLock()
	client, ok := s.clientsByConfig[cacheKey]
	s.mu.RUnlock()
	if ok {
		return client, nil
	}

	// 目前只实现了 OpenAI 兼容适配器
	newClient := ai.NewOpenAIAdapter(baseURL, apiKey)

	s.mu.Lock()
	if s.clientsByConfig == nil {
		s.clientsByConfig = make(map[string]ai.Client)
	}
	s.clientsByConfig[cacheKey] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *AIServiceImpl) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	// 2. 认知记忆注入 (Memory Injection)
	if s.memoryService != nil {
		// 尝试从上下文提取真实的 userID 和 botID
		userID, _ := ctx.Value("userID").(string)
		botID, _ := ctx.Value("botID").(string)

		if userID == "" {
			userID = "default_user"
		}
		if botID == "" {
			botID = "default_bot"
		}

		// 提取最后一条用户消息作为检索 query
		query := ""
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				if q, ok := messages[i].Content.(string); ok {
					query = q
					break
				}
			}
		}

		memories, _ := s.memoryService.GetRelevantMemories(ctx, userID, botID, query)
		if len(memories) > 0 {
			memoryContext := "你拥有以下关于用户的认知记忆：\n"
			for _, m := range memories {
				memoryContext += fmt.Sprintf("- [%s] %s\n", m.Category, m.Content)
			}
			// 插入到第一条消息，引导 AI 行为
			messages = append([]ai.Message{{
				Role:    "system",
				Content: memoryContext,
			}}, messages...)
		}
	}

	// 3. 知识库检索 (RAG)
	if s.manager != nil && s.manager.TaskManager != nil && s.manager.TaskManager.AI.Manifest.KnowledgeBase != nil {
		// 从上下文或 Agent 配置中获取 KnowledgeBase 过滤条件
		// 暂时使用全局搜索，后续可以根据 botID 绑定特定知识库
		query := ""
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				if q, ok := messages[i].Content.(string); ok {
					query = q
					break
				}
			}
		}

		if query != "" {
			kb := s.manager.TaskManager.AI.Manifest.KnowledgeBase
			// 这里的 filter 可以根据业务需求定制
			chunks, err := kb.Search(ctx, query, 3, nil)
			if err == nil && len(chunks) > 0 {
				kbContext := "参考知识库内容：\n"
				for _, chunk := range chunks {
					kbContext += fmt.Sprintf("- %s\n", chunk.Content)
				}
				// 插入到消息列表中
				messages = append([]ai.Message{{
					Role:    "system",
					Content: kbContext,
				}}, messages...)
			}
		}
	}

	// 4. 获取技能与工具并合并
	if s.skillManager != nil {
		botID, _ := ctx.Value("botID").(string)
		userIDNum, _ := ctx.Value("userIDNum").(uint)
		orgIDNum, _ := ctx.Value("orgIDNum").(uint)

		if botID == "" {
			botID = "default_bot"
		}
		extraTools, _ := s.skillManager.GetAvailableSkillsForBot(ctx, botID, userIDNum, orgIDNum)
		if len(extraTools) > 0 {
			tools = append(tools, extraTools...)
		}
	}

	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	provider, err := s.GetProvider(model.ProviderID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(provider, &model)
	if err != nil {
		return nil, err
	}

	// 隐私脱敏处理 (PII Masking)
	maskCtx := ai.NewMaskContext()
	maskedMessages := s.maskMessages(messages, maskCtx)

	req := ai.ChatRequest{
		Model:    model.ModelID,
		Messages: maskedMessages,
		Tools:    tools,
	}

	// 打印 LLM 调用详情
	fmt.Printf("\n--- [LLM CALL START] ---\n")
	fmt.Printf("Model: %s (%s)\n", model.ModelName, model.ModelID)
	for i, msg := range maskedMessages {
		contentStr := ""
		if s, ok := msg.Content.(string); ok {
			contentStr = s
		} else {
			contentStr = "[Multi-modal Content]"
		}
		fmt.Printf("[%d] %s: %s\n", i, msg.Role, contentStr)
	}
	if len(tools) > 0 {
		fmt.Printf("Tools Available: %d\n", len(tools))
	}

	startTime := time.Now()
	resp, err := client.Chat(ctx, req)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("LLM Error: %v\n", err)
	} else if resp != nil && len(resp.Choices) > 0 {
		// 还原响应中的敏感信息 (Unmasking)
		for i := range resp.Choices {
			resp.Choices[i].Message.Content = s.unmaskContent(resp.Choices[i].Message.Content, maskCtx)
		}

		choice := resp.Choices[0]
		contentStr, _ := choice.Message.Content.(string)
		if contentStr != "" {
			fmt.Printf("Response: %s\n", contentStr)
		}
		for _, tc := range choice.Message.ToolCalls {
			fmt.Printf("Tool Call: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
		}
		fmt.Printf("Tokens: Input=%d, Output=%d, Total=%d\n",
			resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	}
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("--- [LLM CALL END] ---\n\n")

	if err == nil && resp != nil {
		// 异步记录使用日志
		go func() {
			usage := models.AIUsageLogGORM{
				UserID:       0, // TODO: 从 context 或参数中传递真正的 UserID
				ModelName:    model.ModelName,
				ProviderType: "openai", // 暂时硬编码，可从 provider 字段获取
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				DurationMS:   int(duration.Milliseconds()),
				Status:       "success",
				CreatedAt:    time.Now(),
			}
			s.db.Create(&usage)
		}()
	}

	return resp, err
}

// ExecuteTool 执行工具调用，支持普通 Skill 和 MCP Tool
func (s *AIServiceImpl) ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall ai.ToolCall) (any, error) {
	if s.skillManager != nil {
		return s.skillManager.ExecuteSkill(ctx, botID, userID, orgID, toolCall)
	}

	// 回退逻辑 (如果 skillManager 未初始化)
	name := toolCall.Function.Name
	var args map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	if strings.Contains(name, "__") {
		// MCP Tool 同样需要基础的权限验证 (简化版)
		return s.mcpManager.CallTool(ctx, name, args)
	}

	if s.manager != nil {
		return s.manager.SyncSkillCall(ctx, name, args)
	}

	return nil, fmt.Errorf("standard skill execution failed: manager not initialized")
}

// maskMessages 对消息列表进行脱敏
func (s *AIServiceImpl) maskMessages(messages []ai.Message, ctx *ai.MaskContext) []ai.Message {
	if s.privacyFilter == nil {
		return messages
	}

	masked := make([]ai.Message, len(messages))
	for i, msg := range messages {
		masked[i] = msg
		switch v := msg.Content.(type) {
		case string:
			masked[i].Content = s.privacyFilter.Mask(v, ctx)
		case []ai.ContentPart:
			newParts := make([]ai.ContentPart, len(v))
			for j, part := range v {
				newParts[j] = part
				if part.Type == "text" {
					newParts[j].Text = s.privacyFilter.Mask(part.Text, ctx)
				}
			}
			masked[i].Content = newParts
		}
	}
	return masked
}

// unmaskContent 还原内容中的敏感信息
func (s *AIServiceImpl) unmaskContent(content any, ctx *ai.MaskContext) any {
	if s.privacyFilter == nil {
		return content
	}

	switch v := content.(type) {
	case string:
		return s.privacyFilter.Unmask(v, ctx)
	case []ai.ContentPart:
		newParts := make([]ai.ContentPart, len(v))
		for i, part := range v {
			newParts[i] = part
			if part.Type == "text" {
				newParts[i].Text = s.privacyFilter.Unmask(part.Text, ctx)
			}
		}
		return newParts
	default:
		return content
	}
}
func (s *AIServiceImpl) ChatStream(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (<-chan ai.ChatStreamResponse, error) {
	// 获取技能与工具并合并
	if s.skillManager != nil {
		// TODO: 从上下文获取真实 UserID 和 OrgID
		botID := "default_bot"
		extraTools, _ := s.skillManager.GetAvailableSkillsForBot(ctx, botID, 0, 0)
		if len(extraTools) > 0 {
			tools = append(tools, extraTools...)
		}
	}

	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	provider, err := s.GetProvider(model.ProviderID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(provider, &model)
	if err != nil {
		return nil, err
	}

	// 隐私脱敏处理 (PII Masking)
	maskCtx := ai.NewMaskContext()
	maskedMessages := s.maskMessages(messages, maskCtx)

	req := ai.ChatRequest{
		Model:    model.ModelID,
		Messages: maskedMessages,
		Tools:    tools,
		Stream:   true,
	}

	// 打印 LLM 流式调用详情
	fmt.Printf("\n--- [LLM STREAM START] ---\n")
	fmt.Printf("Model: %s (%s)\n", model.ModelName, model.ModelID)
	for i, msg := range maskedMessages {
		fmt.Printf("Message #%d [%s]: %v\n", i, msg.Role, msg.Content)
	}
	fmt.Printf("--- [STREAMING...] ---\n\n")

	return client.ChatStream(ctx, req)
}

func (s *AIServiceImpl) CreateEmbedding(ctx context.Context, modelID uint, input any) (*ai.EmbeddingResponse, error) {
	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	provider, err := s.GetProvider(model.ProviderID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(provider, &model)
	if err != nil {
		return nil, err
	}

	req := ai.EmbeddingRequest{
		Model: model.ModelID,
		Input: input,
	}

	resp, err := client.CreateEmbedding(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *AIServiceImpl) DispatchIntent(msg types.InternalMessage) (string, error) {
	// 1. 获取所有可用工具定义 (如果任务管理器可用)
	var tools []ai.Tool
	var systemPrompt string

	if s.manager.TaskManager != nil && s.manager.TaskManager.AI != nil {
		tools = s.manager.TaskManager.AI.Manifest.GenerateTools()
		systemPrompt = s.manager.TaskManager.AI.Manifest.GenerateSystemPrompt()
	}

	// 2. 找到默认的对话模型
	var model models.AIModelGORM
	if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		// 如果没有默认模型，尝试获取第一个模型
		if err := s.db.First(&model).Error; err != nil {
			return "", fmt.Errorf("no ai model available: %v", err)
		}
	}

	// 3. 构造消息
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	messages := []ai.Message{
		{
			Role:    ai.RoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    ai.RoleUser,
			Content: msg.RawMessage,
		},
	}

	// 4. 调用 AI 进行意图识别 (Function Calling)
	resp, err := s.Chat(context.Background(), model.ID, messages, tools)
	if err != nil {
		return "", err
	}

	// 5. 解析结果
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if len(choice.Message.ToolCalls) > 0 {
			// 命中工具调用
			toolCall := choice.Message.ToolCalls[0]
			return fmt.Sprintf("Skill Hit: %s with args %s", toolCall.Function.Name, toolCall.Function.Arguments), nil
		}
		content, _ := choice.Message.Content.(string)
		return content, nil
	}

	return "No response from AI", nil
}

func (s *AIServiceImpl) ChatWithEmployee(employee *models.DigitalEmployeeGORM, msg types.InternalMessage) (string, error) {
	// 1. 获取会话 ID
	sessionID := fmt.Sprintf("bot:%s:user:%s", employee.BotID, msg.UserID)
	if msg.GroupID != "" {
		sessionID = fmt.Sprintf("bot:%s:group:%s", employee.BotID, msg.GroupID)
	}

	// 2. 获取最近历史记录 (最近 10 条)
	var historyMessages []models.AIChatMessageGORM
	s.db.Where("session_id = ?", sessionID).Order("id desc").Limit(10).Find(&historyMessages)

	// 3. 构造 AI 消息列表
	var messages []ai.Message

	// 添加人设 Prompt
	systemPrompt := employee.Bio
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("你现在扮演 %s，职位是 %s。", employee.Name, employee.Title)
	}

	messages = append(messages, ai.Message{
		Role:    ai.RoleSystem,
		Content: systemPrompt,
	})

	// 添加历史记录 (需要反转顺序)
	for i := len(historyMessages) - 1; i >= 0; i-- {
		m := historyMessages[i]
		messages = append(messages, ai.Message{
			Role:    ai.Role(m.Role),
			Content: m.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, ai.Message{
		Role:    ai.RoleUser,
		Content: msg.RawMessage,
	})

	// 4. 获取模型 ID (如果数字员工绑定了智能体，使用智能体的模型)
	modelID := uint(0)
	if employee.AgentID > 0 {
		var agent models.AIAgentGORM
		if err := s.db.First(&agent, employee.AgentID).Error; err == nil {
			modelID = agent.ModelID
		}
	}

	// 如果还是没模型，用默认的
	if modelID == 0 {
		var model models.AIModelGORM
		if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
			s.db.First(&model)
			modelID = model.ID
		} else {
			modelID = model.ID
		}
	}

	// 5. 调用 AI
	chatCtx := context.WithValue(context.Background(), "botID", employee.BotID)
	chatCtx = context.WithValue(chatCtx, "userID", fmt.Sprintf("%v", msg.UserID))
	// 如果能从 employee 或 msg 获取 OrgID, 也可以传进去
	// chatCtx = context.WithValue(chatCtx, "orgIDNum", employee.OrgID)

	resp, err := s.Chat(chatCtx, modelID, messages, nil)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		content, _ := resp.Choices[0].Message.Content.(string)

		// 6. 异步保存消息到历史并更新薪资消耗
		go func() {
			// 保存用户消息
			s.db.Create(&models.AIChatMessageGORM{
				SessionID: sessionID,
				UserID:    uint(0), // 这里的 UserID 是指 Nexus 用户，机器人对话中的用户 ID 存在 SessionID 中
				Role:      string(ai.RoleUser),
				Content:   msg.RawMessage,
			})
			// 保存 AI 回复
			s.db.Create(&models.AIChatMessageGORM{
				SessionID:  sessionID,
				UserID:     uint(0),
				Role:       string(ai.RoleAssistant),
				Content:    content,
				UsageToken: resp.Usage.TotalTokens,
			})

			// 更新数字员工的薪资消耗 (Token 统计)
			if s.employeeService != nil {
				s.employeeService.ConsumeSalary(employee.BotID, int64(resp.Usage.TotalTokens))
			}

			// 7. 认知记忆自动提取
			userIDStr := fmt.Sprintf("%v", msg.UserID)
			s.ExtractAndSaveMemories(context.Background(), userIDStr, employee.BotID, messages[len(messages)-2:])
		}()

		return content, nil
	}

	return "", fmt.Errorf("no response from ai")
}

// ExtractAndSaveMemories 从对话中提取并保存新的记忆
func (s *AIServiceImpl) ExtractAndSaveMemories(ctx context.Context, userID string, botID string, messages []ai.Message) {
	if s.memoryService == nil || len(messages) < 2 {
		return
	}

	// 构造提取 Prompt
	prompt := "你是一个记忆提取专家。请从以下对话中提取出关于用户的、具有长期价值的事实或偏好（如：姓名、生日、职业、喜好、厌恶、重要经历等）。\n"
	prompt += "规则：\n1. 只提取新出现的、有价值的信息。\n2. 格式为：[类别] 事实内容。\n3. 如果没有发现新信息，请回复 'NONE'。\n4. 不要提取对话过程，只提取事实。\n\n对话内容：\n"

	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %v\n", m.Role, m.Content)
	}

	// 获取默认模型
	var model models.AIModelGORM
	if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		if err := s.db.First(&model).Error; err != nil {
			return
		}
	}

	// 调用 AI 进行提取 (直接调用 Chat 内部逻辑，避免死循环)
	// 这里使用一个简化的 Chat 调用
	provider, _ := s.GetProvider(model.ProviderID)
	client, _ := s.getClient(provider, &model)
	if client == nil {
		return
	}

	req := ai.ChatRequest{
		Model: model.ModelID,
		Messages: []ai.Message{
			{Role: ai.RoleSystem, Content: prompt},
		},
	}

	resp, err := client.Chat(ctx, req)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		return
	}

	content, _ := resp.Choices[0].Message.Content.(string)
	if strings.TrimSpace(content) == "" || strings.ToUpper(content) == "NONE" {
		return
	}

	// 解析并保存
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "NONE" {
			continue
		}

		// 尝试解析 [类别] 内容
		category := "general"
		fact := line
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			idx := strings.Index(line, "]")
			category = line[1:idx]
			fact = strings.TrimSpace(line[idx+1:])
		}

		memory := &models.CognitiveMemoryGORM{
			UserID:     userID,
			BotID:      botID,
			Category:   category,
			Content:    fact,
			Importance: 1,
		}
		s.memoryService.SaveMemory(ctx, memory)
	}
}
