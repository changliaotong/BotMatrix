package ai

import (
	"BotMatrix/common/ai/b2b"
	"BotMatrix/common/ai/employee"
	// "BotMatrix/common/ai/employee" // Moved to SetEmployeeService to break cycle

	// "BotMatrix/common/ai/employee" // Moved to SetEmployeeService to break cycle
	"BotMatrix/common/ai/rag"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIServiceImpl struct {
	db              *gorm.DB
	provider        AIServiceProvider
	clientsByConfig map[string]Client
	mu              sync.RWMutex
	privacyFilter   *PrivacyFilter
	contextManager  *ContextManager
	mcpManager      MCPManagerInterface
	skillManager    *SkillManager
	memoryService   types.CognitiveMemoryService // Use interface
	employeeService types.DigitalEmployeeService // Use interface
	b2bService      b2b.B2BService
}

func NewAIService(db *gorm.DB, provider AIServiceProvider, mcp MCPManagerInterface) *AIServiceImpl {
	return &AIServiceImpl{
		db:              db,
		provider:        provider,
		clientsByConfig: make(map[string]Client),
		privacyFilter:   types.NewPrivacyFilter(),
		contextManager:  NewContextManager(8192), // 默认支持 8k token 上下文
		mcpManager:      mcp,
		skillManager:    NewSkillManager(db, provider, mcp),
		// memoryService:   employee.NewCognitiveMemoryService(db), // Initialize later or use interface wrapper if needed
		// employeeService: employee.NewEmployeeService(db), // Initialize later
	}
}

func (s *AIServiceImpl) SetEmployeeService(svc types.DigitalEmployeeService) {
	s.employeeService = svc
}

func (s *AIServiceImpl) SetCognitiveMemoryService(svc types.CognitiveMemoryService) {
	s.memoryService = svc
}

func (s *AIServiceImpl) SetB2BService(b2bSvc b2b.B2BService) {
	s.b2bService = b2bSvc
	if s.skillManager != nil {
		s.skillManager.SetB2BService(b2bSvc)
	}
}

func (s *AIServiceImpl) SetKnowledgeBase(kb rag.KnowledgeBase) {
	if s.mcpManager != nil {
		s.mcpManager.SetKnowledgeBase(kb)
	}
}

func (s *AIServiceImpl) GetMCPManager() MCPManagerInterface {
	return s.mcpManager
}

func (s *AIServiceImpl) GetProvider(id uint) (*models.AIProvider, error) {
	var provider models.AIProvider
	if err := s.db.First(&provider, id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *AIServiceImpl) getClient(provider *models.AIProvider, model *models.AIModel) (Client, error) {
	baseURL := provider.BaseURL
	apiKey := provider.APIKey

	if provider.Type == "worker" {
		return NewWorkerAIClient(s.provider), nil
	}

	if provider.Type == "mock" {
		return NewMockClient(baseURL, apiKey), nil
	}

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
	newClient := NewOpenAIAdapter(baseURL, apiKey)

	s.mu.Lock()
	if s.clientsByConfig == nil {
		s.clientsByConfig = make(map[string]Client)
	}
	s.clientsByConfig[cacheKey] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *AIServiceImpl) prepareChat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []Tool, *types.MaskContext) {
	sessionID, _ := ctx.Value("sessionID").(string)
	botID, _ := ctx.Value("botID").(string)
	step, _ := ctx.Value("step").(int)

	// 2. 认知记忆注入 (Memory Injection)
	if s.memoryService != nil {
		userID, _ := ctx.Value("userID").(string)

		if userID != "" && botID != "" {
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
				messages = append([]Message{{
					Role:    "system",
					Content: memoryContext,
				}}, messages...)

				// 记录追踪
				if sessionID != "" {
					s.SaveTrace(sessionID, botID, step, "memory_retrieval", fmt.Sprintf("Retrieved %d memories", len(memories)), "")
				}
			}
		}
	}

	// 3. 知识库检索 (RAG)
	if s.provider != nil && s.provider.GetKnowledgeBase() != nil {
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
			kb := s.provider.GetKnowledgeBase()
			chunks, err := kb.Search(ctx, query, 3, nil)
			if err == nil && len(chunks) > 0 {
				kbContext := "参考知识库内容：\n"
				for _, chunk := range chunks {
					kbContext += fmt.Sprintf("- %s\n", chunk.Content)
				}
				messages = append([]Message{{
					Role:    "system",
					Content: kbContext,
				}}, messages...)

				// 记录追踪
				if sessionID != "" {
					s.SaveTrace(sessionID, botID, step, "knowledge_retrieval", fmt.Sprintf("Retrieved %d chunks from KB", len(chunks)), "")
				}
			}
		}
	}

	// 4. 获取技能与工具并合并
	if s.skillManager != nil {
		userIDNum, _ := ctx.Value("userIDNum").(uint)
		orgIDNum, _ := ctx.Value("orgIDNum").(uint)

		// 智能机器人核心优化：自动注入推理引擎和基础技能，无需数字员工权限
		if botID != "" {
			// 获取机器人可用技能
			extraTools, _ := s.skillManager.GetAvailableSkillsForBot(ctx, botID, userIDNum, orgIDNum)

			// 始终注入推理引擎 (如果未被禁用)
			if s.mcpManager != nil {
				reasoningTools, _ := s.mcpManager.GetToolsForContext(ctx, userIDNum, orgIDNum)
				for _, rt := range reasoningTools {
					if strings.Contains(rt.Function.Name, "sequential_thinking") {
						extraTools = append(extraTools, rt)
					}
				}
			}

			if len(extraTools) > 0 {
				tools = append(tools, extraTools...)
				// 记录追踪
				if sessionID != "" {
					skillNames := make([]string, len(extraTools))
					for i, t := range extraTools {
						skillNames[i] = t.Function.Name
					}
					s.SaveTrace(sessionID, botID, step, "skill_injection", strings.Join(skillNames, ", "), "")
				}
			}
		}
	}

	// 5. 上下文管理 (Context Pruning)
	if s.contextManager != nil {
		messages = s.contextManager.PruneMessages(messages)
	}

	// 隐私脱敏处理 (PII Masking)
	maskCtx := types.NewMaskContext()
	maskedMessages := s.maskMessages(messages, maskCtx)

	return maskedMessages, tools, maskCtx
}

func (s *AIServiceImpl) GetMemoryService() employee.CognitiveMemoryService {
	return s.memoryService
}

func (s *AIServiceImpl) GetKnowledgeBase() rag.KnowledgeBase {
	if s.mcpManager != nil {
		return s.mcpManager.GetKnowledgeBase()
	}
	return nil
}

func (s *AIServiceImpl) Chat(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	maskedMessages, finalTools, maskCtx := s.prepareChat(ctx, messages, tools)

	var model models.AIModel
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

	req := ChatRequest{
		Model:    model.ApiModelID,
		Messages: maskedMessages,
		Tools:    finalTools,
	}

	// 设置超时 (默认 60s)
	chatCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	startTime := time.Now()
	resp, err := client.Chat(chatCtx, req)
	duration := time.Since(startTime)

	if err != nil {
		clog.Error("[AI] Chat failed", zap.Error(err), zap.Uint("model_id", modelID))
		return nil, err
	}

	if resp != nil && len(resp.Choices) > 0 {
		for i := range resp.Choices {
			resp.Choices[i].Message.Content = s.unmaskContent(resp.Choices[i].Message.Content, maskCtx)
		}

		// 异步记录使用日志
		go func() {
			usage := models.AIUsageLog{
				UserID:       0, // TODO: 从 context 获取
				ModelName:    model.ModelName,
				ProviderType: provider.Type,
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				DurationMS:   int(duration.Milliseconds()),
				Status:       "success",
				CreatedAt:    time.Now(),
			}
			s.db.Create(&usage)
		}()
	}

	return resp, nil
}

// ChatAgent 提供自主 Agent 循环能力，自动处理工具调用并返回最终结果
func (s *AIServiceImpl) ChatAgent(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error) {
	// 提取元数据
	botID, _ := ctx.Value("botID").(string)
	userID, _ := ctx.Value("userID").(string)
	sessionID, _ := ctx.Value("sessionID").(string)

	if botID == "" {
		botID = "default_bot"
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("agent_%d", time.Now().UnixNano())
	}

	// 使用新的 AgentExecutor
	executor := NewAgentExecutor(s, modelID, botID, userID, sessionID)

	// 执行 ReAct 循环
	// 注意：tools 已经在 prepareChat 中准备好了，但在这里我们需要确保它们包含 Sandbox 等工具
	// 这里的 tools 参数通常是外部传入的，而 prepareChat 会注入更多系统级工具
	// 为了复用 prepareChat 的注入逻辑 (Memory, RAG, MCP Tools)，我们需要先调用一次 prepareChat
	// 但 prepareChat 只是准备数据，不执行。

	// 正确的做法是：
	// 1. prepareChat 获取完整的上下文和工具列表
	maskedMessages, finalTools, maskCtx := s.prepareChat(ctx, messages, tools)

	// 2. 将这些传递给 Executor
	resp, err := executor.Execute(ctx, maskedMessages, finalTools)
	if err != nil {
		return nil, err
	}

	// 3. 处理响应中的隐私脱敏
	if resp != nil && len(resp.Choices) > 0 {
		for i := range resp.Choices {
			resp.Choices[i].Message.Content = s.unmaskContent(resp.Choices[i].Message.Content, maskCtx)
		}

		// 异步记录日志与洞察 (复用原有逻辑)
		// ...
		go func() {
			// 这里简单记录最后一次调用的 token 使用情况
			// 实际应该累加整个 Session 的消耗
			var model models.AIModel
			s.db.First(&model, modelID)
			usage := models.AIUsageLog{
				UserID:       0, // TODO
				ModelName:    model.ModelName,
				ProviderType: "agent",
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				Status:       "success",
				CreatedAt:    time.Now(),
			}
			s.db.Create(&usage)
		}()

		if s.memoryService != nil && botID != "" && userID != "" {
			// 注意：这里需要传入完整的 agent 历史记录才能提取洞察
			// Executor 内部维护了历史，但没有返回给我们。
			// 暂时只用最后几轮
			go s.extractInsightsAndSave(context.Background(), modelID, userID, botID, maskedMessages)
		}
	}

	return resp, nil
}

// extractInsightsAndSave 异步提取对话中的重要信息并保存到认知记忆
func (s *AIServiceImpl) extractInsightsAndSave(ctx context.Context, modelID uint, userID, botID string, messages []Message) {
	// 只有当对话轮数超过一定数量，或者包含重要信息时才执行
	if len(messages) < 4 {
		return
	}

	// 构造提取 Prompt
	prompt := "你是一个洞察力极强的观察者。请从以下用户与 AI 的对话记录中，提取出关于用户的偏好、习惯、需求或重要背景信息。\n"
	prompt += "要求：\n1. 只提取对未来交互有价值的长期记忆。\n2. 格式：[类别] 洞察内容 (重要性: 1-5)。\n3. 如果没有发现有价值的信息，请输出 NONE。\n\n对话记录：\n"

	// 只取最近的几轮对话进行分析
	startIdx := 0
	if len(messages) > 10 {
		startIdx = len(messages) - 10
	}
	for _, m := range messages[startIdx:] {
		role := string(m.Role)
		content, _ := m.Content.(string)
		if content != "" {
			prompt += fmt.Sprintf("%s: %s\n", role, content)
		}
	}

	resp, err := s.Chat(ctx, modelID, []Message{{Role: RoleUser, Content: prompt}}, nil)
	if err != nil || len(resp.Choices) == 0 {
		return
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || strings.Contains(content, "NONE") {
		return
	}

	// 解析并保存
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 匹配格式: [类别] 内容 (重要性: N)
		re := regexp.MustCompile(`\[(.*?)\] (.*?) \(重要性: (\d)\)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 {
			importance := 3
			fmt.Sscanf(matches[3], "%d", &importance)

			s.memoryService.SaveMemory(ctx, &models.CognitiveMemory{
				UserID:     userID,
				BotID:      botID,
				Category:   matches[1],
				Content:    matches[2],
				Importance: importance,
				LastSeen:   time.Now(),
			})
		}
	}
}

// ExecuteTool 执行工具调用，支持普通 Skill 和 MCP Tool
// ExecuteTool 执行工具调用，支持普通 Skill 和 MCP Tool
func (s *AIServiceImpl) ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall ToolCall) (any, error) {
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
		var mcpArgs map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &mcpArgs); err != nil {
			return nil, fmt.Errorf("invalid MCP arguments: %v", err)
		}
		return s.mcpManager.CallTool(ctx, name, mcpArgs)
	}

	if s.provider != nil {
		return s.provider.SyncSkillCall(ctx, name, args)
	}

	return nil, fmt.Errorf("standard skill execution failed: provider not initialized")
}

// maskMessages 对消息列表进行脱敏
func (s *AIServiceImpl) maskMessages(messages []Message, ctx *types.MaskContext) []Message {
	if s.privacyFilter == nil {
		return messages
	}

	masked := make([]Message, len(messages))
	for i, msg := range messages {
		masked[i] = msg
		switch v := msg.Content.(type) {
		case string:
			masked[i].Content = s.privacyFilter.Mask(v, ctx)
		case []ContentPart:
			newParts := make([]ContentPart, len(v))
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
func (s *AIServiceImpl) unmaskContent(content any, ctx *types.MaskContext) any {
	if s.privacyFilter == nil {
		return content
	}

	switch v := content.(type) {
	case string:
		return s.privacyFilter.Unmask(v, ctx)
	case []ContentPart:
		newParts := make([]ContentPart, len(v))
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

func (s *AIServiceImpl) ChatStream(ctx context.Context, modelID uint, messages []Message, tools []Tool) (<-chan ChatStreamResponse, error) {
	// 获取技能与工具并合并
	if s.skillManager != nil {
		// TODO: 从上下文获取真实 UserID 和 OrgID
		botID := "default_bot"
		extraTools, _ := s.skillManager.GetAvailableSkillsForBot(ctx, botID, 0, 0)
		if len(extraTools) > 0 {
			tools = append(tools, extraTools...)
		}
	}

	var model models.AIModel
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
	maskCtx := types.NewMaskContext()
	maskedMessages := s.maskMessages(messages, maskCtx)

	req := ChatRequest{
		Model:    model.ApiModelID,
		Messages: maskedMessages,
		Tools:    tools,
		Stream:   true,
	}

	// 打印 LLM 流式调用详情
	fmt.Printf("\n--- [LLM STREAM START] ---\n")
	fmt.Printf("Model: %s (%s)\n", model.ModelName, model.ApiModelID)
	for i, msg := range maskedMessages {
		fmt.Printf("Message #%d [%s]: %v\n", i, msg.Role, msg.Content)
	}
	fmt.Printf("--- [STREAMING...] ---\n\n")

	return client.ChatStream(ctx, req)
}

func (s *AIServiceImpl) CreateEmbedding(ctx context.Context, modelID uint, input any) (*EmbeddingResponse, error) {
	var model models.AIModel
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

	req := EmbeddingRequest{
		Model: model.ApiModelID,
		Input: input,
	}

	resp, err := client.CreateEmbedding(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *AIServiceImpl) CreateEmbeddingSimple(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	// 找到默认的嵌入模型
	var model models.AIModel
	if err := s.db.Where("is_default = ? AND type = ?", true, "embedding").First(&model).Error; err != nil {
		if err := s.db.Where("type = ?", "embedding").First(&model).Error; err != nil {
			// 如果没有专门的 embedding 模型，尝试默认模型
			if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
				return nil, fmt.Errorf("no embedding model available: %v", err)
			}
		}
	}
	return s.CreateEmbedding(ctx, model.ID, req.Input)
}

func (s *AIServiceImpl) ChatSimple(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var modelID uint
	modelIDStr := req.Model

	// 1. 尝试根据 Model 名称查找模型
	if modelIDStr != "" {
		var model models.AIModel
		if err := s.db.Where("api_model_id = ?", modelIDStr).First(&model).Error; err == nil {
			modelID = model.ID
		}
	}

	// 2. 如果未指定或未找到，使用默认模型
	if modelID == 0 {
		var model models.AIModel
		if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
			if err := s.db.First(&model).Error; err != nil {
				return nil, fmt.Errorf("no chat model available: %v", err)
			}
		}
		modelID = model.ID
	}
	return s.Chat(ctx, modelID, req.Messages, req.Tools)
}

// SaveTrace 记录 AI Agent 的执行追踪日志
func (s *AIServiceImpl) SaveTrace(sessionID, botID string, step int, traceType, content, metadata string) {
	trace := models.AIAgentTrace{
		SessionID: sessionID,
		BotID:     botID,
		Step:      step,
		Type:      traceType,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	// 异步保存，不阻塞主流程
	go func() {
		if err := s.db.Create(&trace).Error; err != nil {
			clog.Error("failed to save agent trace", zap.Error(err), zap.String("session", sessionID))
		}
	}()
}

func (s *AIServiceImpl) DispatchIntent(msg types.InternalMessage) (string, error) {
	// 1. 获取所有可用工具定义 (如果任务管理器可用)
	var tools []Tool
	var systemPrompt string

	if s.provider != nil && s.provider.GetManifest() != nil {
		tools = s.provider.GetManifest().GenerateTools()
		systemPrompt = s.provider.GetManifest().GenerateSystemPrompt()
	}

	// 2. 找到默认的对话模型
	var model models.AIModel
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

	messages := []Message{
		{
			Role:    RoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    RoleUser,
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

func (s *AIServiceImpl) ChatWithEmployee(employee *models.DigitalEmployee, msg types.InternalMessage, targetOrgID uint) (string, error) {
	// 0. 检查功能开关
	if s.provider != nil && !s.provider.IsDigitalEmployeeEnabled() {
		return "对不起，数字员工服务当前已禁用。请联系管理员开启 EnableDigitalEmployee。", nil
	}

	// 1. 验证访问权限 (本地或外派)
	if employee.EnterpriseID != targetOrgID {
		if s.b2bService != nil {
			allowed, err := s.b2bService.CheckDispatchPermission(employee.ID, targetOrgID, "chat")
			if err != nil {
				return "", fmt.Errorf("failed to check dispatch permission: %w", err)
			}
			if !allowed {
				return "对不起，我没有权限为您提供服务 (未被外派或权限不足)。", nil
			}
		} else {
			return "对不起，跨企业协作服务未开启。", nil
		}
	}

	// 2. 获取会话 ID
	sessionID := fmt.Sprintf("bot:%s:user:%s", employee.BotID, msg.UserID)
	if msg.GroupID != "" {
		sessionID = fmt.Sprintf("bot:%s:group:%s", employee.BotID, msg.GroupID)
	}

	// 2. 获取最近历史记录 (最近 10 条)
	var historyMessages []models.AIChatMessage
	s.db.Where("session_id = ?", sessionID).Order("id desc").Limit(10).Find(&historyMessages)

	// 3. 构造 AI 消息列表
	var messages []Message

	// 添加人设 Prompt
	systemPrompt := employee.Bio
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("你现在扮演 %s，职位是 %s。", employee.Name, employee.Title)
	}

	messages = append(messages, Message{
		Role:    RoleSystem,
		Content: systemPrompt,
	})

	// 添加历史记录 (需要反转顺序)
	for i := len(historyMessages) - 1; i >= 0; i-- {
		m := historyMessages[i]
		messages = append(messages, Message{
			Role:    Role(m.Role),
			Content: m.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, Message{
		Role:    RoleUser,
		Content: msg.RawMessage,
	})

	// 4. 获取模型 ID (如果数字员工绑定了智能体，使用智能体的模型)
	modelID := uint(0)
	if employee.AgentID > 0 {
		var agent models.AIAgent
		if err := s.db.First(&agent, employee.AgentID).Error; err == nil {
			modelID = agent.ModelID
		}
	}

	// 如果还是没模型，用默认的
	if modelID == 0 {
		var model models.AIModel
		if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
			s.db.First(&model)
			modelID = model.ID
		} else {
			modelID = model.ID
		}
	}

	// 5. 调用 AI 前检查薪资预算
	if s.employeeService != nil {
		ok, err := s.employeeService.CheckSalaryLimit(employee.BotID)
		if err != nil {
			clog.Error("[AI] Failed to check salary limit", zap.Error(err))
		} else if !ok {
			return "对不起，我的今日预算已耗尽，请联系管理员增加额度。", nil
		}
	}

	// 6. 调用 AI
	newSessionID := fmt.Sprintf("chat:%s:user:%s:%d", employee.BotID, msg.UserID, time.Now().Unix())
	chatCtx := context.WithValue(context.Background(), "botID", employee.BotID)
	chatCtx = context.WithValue(chatCtx, "userID", fmt.Sprintf("%v", msg.UserID))
	chatCtx = context.WithValue(chatCtx, "orgIDNum", targetOrgID) // 传入目标企业 ID
	chatCtx = context.WithValue(chatCtx, "sessionID", newSessionID)
	chatCtx = context.WithValue(chatCtx, "step", 0)

	// 链路追踪：关联父 SessionID
	parentSessionID, _ := msg.Extras["parentSessionID"].(string)
	if parentSessionID != "" {
		chatCtx = context.WithValue(chatCtx, "parentSessionID", parentSessionID)
		s.SaveTrace(newSessionID, employee.BotID, 0, "collaboration_link", fmt.Sprintf("Parent Session: %s", parentSessionID), "")
	}

	// 任务追踪：关联 ExecutionID
	executionID, _ := msg.Extras["executionID"].(string)
	if executionID != "" {
		chatCtx = context.WithValue(chatCtx, "executionID", executionID)
		s.SaveTrace(newSessionID, employee.BotID, 0, "task_link", fmt.Sprintf("Execution ID: %s", executionID), "")

		// 在系统提示词中注入任务信息，让 AI 意识到自己正在执行一个委派任务
		taskInfo := fmt.Sprintf("\n\n注意：你当前正在执行一个被委派的任务（ID: %s）。完成任务后，请务必使用 `task_report` 工具汇报进度或结果。", executionID)
		messages[0].Content = messages[0].Content.(string) + taskInfo
	}

	if employee.EnterpriseID != targetOrgID {
		chatCtx = context.WithValue(chatCtx, "isDispatched", true)
		chatCtx = context.WithValue(chatCtx, "sourceOrgID", employee.EnterpriseID)
	}

	// 使用 ChatAgent 以便支持工具调用
	resp, err := s.ChatAgent(chatCtx, modelID, messages, nil)
	if err != nil {
		s.SaveTrace(sessionID, employee.BotID, 0, "error", err.Error(), "")
		return "", err
	}

	if len(resp.Choices) > 0 {
		content, _ := resp.Choices[0].Message.Content.(string)
		// ChatAgent 内部已经记录了最终响应的 trace，这里不需要重复记录

		// 6. 异步保存消息到历史并更新薪资消耗
		go func() {
			// 保存用户消息
			s.db.Create(&models.AIChatMessage{
				SessionID: sessionID,
				UserID:    uint(0), // 这里的 UserID 是指 Nexus 用户，机器人对话中的用户 ID 存在 SessionID 中
				Role:      string(RoleUser),
				Content:   msg.RawMessage,
			})
			// 保存 AI 回复
			s.db.Create(&models.AIChatMessage{
				SessionID:  sessionID,
				UserID:     uint(0),
				Role:       string(RoleAssistant),
				Content:    content,
				UsageToken: resp.Usage.TotalTokens,
			})

			// 更新数字员工的薪资消耗 (Token 统计)
			if s.employeeService != nil {
				s.employeeService.ConsumeSalary(employee.BotID, int64(resp.Usage.TotalTokens))
			}

			// 7. 认知记忆自动提取 (异步执行，不影响主流程)
			userIDStr := fmt.Sprintf("%v", msg.UserID)
			sideCtx := context.WithValue(context.Background(), "sessionID", newSessionID)
			sideCtx = context.WithValue(sideCtx, "botID", employee.BotID)
			sideCtx = context.WithValue(sideCtx, "orgIDNum", targetOrgID)
			// 为后台任务设置独立超时
			sideCtx, _ = context.WithTimeout(sideCtx, 30*time.Second)

			go s.ExtractAndSaveMemories(sideCtx, userIDStr, employee.BotID, messages[len(messages)-2:])

			// 8. AI 自动 KPI 评分
			go s.EvaluateAndRecordKpi(sideCtx, employee, msg.RawMessage, content)

			// 9. 数字员工自动学习 (异步执行)
			go s.AutoLearnFromConversation(sideCtx, employee, messages[len(messages)-2:])

			// 10. 定期固化记忆 (每 20 条消息触发一次，或者根据记忆数量触发)
			// 这里简单演示：如果记忆数量超过一定阈值就触发
			go s.MaybeConsolidateMemories(context.Background(), userIDStr, employee.BotID)
		}()

		return content, nil
	}

	return "", fmt.Errorf("no response from ai")
}

// ExtractAndSaveMemories 从对话中提取并保存新的记忆
func (s *AIServiceImpl) ExtractAndSaveMemories(ctx context.Context, userID string, botID string, messages []Message) {
	if s.memoryService == nil || len(messages) < 2 {
		return
	}

	sessionID, _ := ctx.Value("sessionID").(string)

	// 构造提取 Prompt
	prompt := "你是一个记忆提取专家。请从以下对话中提取出关于用户的、具有长期价值的事实或偏好（如：姓名、生日、职业、喜好、厌恶、重要经历等）。\n"
	prompt += "规则：\n1. 只提取新出现的、有价值的信息。\n2. 格式为：[类别] 事实内容。\n3. 如果没有发现新信息，请回复 'NONE'。\n4. 不要提取对话过程，只提取事实。\n\n对话内容：\n"

	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %v\n", m.Role, m.Content)
	}

	// 获取默认模型
	var model models.AIModel
	if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		if err := s.db.First(&model).Error; err != nil {
			clog.Error("[Memory] No default model found for extraction")
			return
		}
	}

	// 调用 AI 进行提取
	provider, _ := s.GetProvider(model.ProviderID)
	client, _ := s.getClient(provider, &model)
	if client == nil {
		return
	}

	req := ChatRequest{
		Model: model.ApiModelID,
		Messages: []Message{
			{Role: RoleSystem, Content: prompt},
		},
	}

	resp, err := client.Chat(ctx, req)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		if err != nil {
			clog.Error("[Memory] AI extraction failed", zap.Error(err))
		}
		return
	}

	content, _ := resp.Choices[0].Message.Content.(string)
	if strings.TrimSpace(content) == "" || strings.ToUpper(strings.TrimSpace(content)) == "NONE" {
		return
	}

	// 记录追踪
	if sessionID != "" {
		s.SaveTrace(sessionID, botID, 99, "memory_extracted", content, "")
	}

	// 解析并保存
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.ToUpper(line) == "NONE" {
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

		// --- 去重与冲突解决逻辑 ---
		// 检索该用户下该机器人是否已有相似记忆
		existing, _ := s.memoryService.SearchMemories(ctx, botID, fact, category)
		var memory *models.CognitiveMemory

		if len(existing) > 0 {
			// 找到相似记忆
			topMatch := existing[0]
			// 假设 SearchMemories 返回的第一个是最相似的，我们可以根据业务逻辑判断是否合并
			// 这里简单判断：如果内容非常相似，则合并
			if strings.Contains(topMatch.Content, fact) || strings.Contains(fact, topMatch.Content) {
				memory = &topMatch
				memory.LastSeen = time.Now()
				memory.Importance++ // 再次提到，增加重要性
				if len(fact) > len(memory.Content) {
					memory.Content = fact
				}
			} else {
				// 冲突解决：让 AI 决定是否合并或作为新记忆
				mergePrompt := fmt.Sprintf("请判断以下两个关于用户的记忆是否描述同一件事。如果是，请合并它们；如果不是，请回复 'NEW'。\n现有记忆：%s\n新提炼记忆：%s\n请直接输出合并后的内容或 'NEW'。", topMatch.Content, fact)
				mergeResp, err := client.Chat(ctx, ChatRequest{
					Model: model.ApiModelID,
					Messages: []Message{
						{Role: RoleSystem, Content: mergePrompt},
					},
				})
				if err == nil && len(mergeResp.Choices) > 0 {
					mergeResult, _ := mergeResp.Choices[0].Message.Content.(string)
					mergeResult = strings.TrimSpace(mergeResult)
					if mergeResult != "NEW" && mergeResult != "" {
						memory = &topMatch
						memory.Content = mergeResult
						memory.LastSeen = time.Now()
						memory.Importance++
					}
				}
			}
		}

		if memory == nil {
			// 创建新记忆
			memory = &models.CognitiveMemory{
				UserID:     userID,
				BotID:      botID,
				Category:   category,
				Content:    fact,
				Importance: 1,
				LastSeen:   time.Now(),
			}
		}

		if err := s.memoryService.SaveMemory(ctx, memory); err != nil {
			clog.Error("[Memory] Failed to save memory", zap.Error(err))
		}
	}
}

// AutoLearnFromConversation 从对话中自动学习新知识或技能
func (s *AIServiceImpl) AutoLearnFromConversation(ctx context.Context, employee *models.DigitalEmployee, messages []Message) {
	if s.provider == nil || s.provider.GetManifest() == nil || s.provider.GetManifest().KnowledgeBase == nil {
		return
	}

	sessionID, _ := ctx.Value("sessionID").(string)

	kb, ok := s.provider.GetManifest().KnowledgeBase.(*rag.PostgresKnowledgeBase)
	if !ok {
		return
	}

	// 1. 构造提取 Prompt
	prompt := `你是一个知识提取与管理专家。请分析以下对话，提取出其中包含的“通用知识”、“业务规则”、“技术步骤”或“操作技巧”。

规则：
1. 只提取具有普遍价值的新信息，不要提取关于特定用户的私有记忆（如姓名、喜好）。
2. 请直接输出 JSON 数组格式，每个元素包含：
   - category: 类别 (业务知识/操作规程/技术细节/常见问题)
   - content: 知识内容的详细描述
   - summary: 一句话摘要
3. 如果没有发现有价值的新知识，请回复 '[]'。

对话内容：
`
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %v\n", m.Role, m.Content)
	}

	// 2. 获取默认模型
	var model models.AIModel
	if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		if err := s.db.First(&model).Error; err != nil {
			return
		}
	}

	// 3. 调用 AI 提取
	provider, _ := s.GetProvider(model.ProviderID)
	client, _ := s.getClient(provider, &model)
	if client == nil {
		return
	}

	req := ChatRequest{
		Model: model.ApiModelID,
		Messages: []Message{
			{Role: RoleSystem, Content: prompt},
		},
	}

	resp, err := client.Chat(ctx, req)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		return
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || strings.TrimSpace(content) == "" || strings.TrimSpace(content) == "[]" {
		return
	}

	// 记录追踪
	if sessionID != "" {
		s.SaveTrace(sessionID, employee.BotID, 99, "auto_learned_raw", content, "")
	}

	// 解析 JSON (增加鲁棒性)
	jsonStr := content
	if idx := strings.Index(jsonStr, "["); idx != -1 {
		if lastIdx := strings.LastIndex(jsonStr, "]"); lastIdx != -1 && lastIdx > idx {
			jsonStr = jsonStr[idx : lastIdx+1]
		}
	}

	var learnedItems []struct {
		Category string `json:"category"`
		Content  string `json:"content"`
		Summary  string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &learnedItems); err != nil {
		clog.Warn("[AutoLearn] Failed to unmarshal JSON, fallback to manual parsing", zap.Error(err))
		// 简单的手动解析逻辑 (如果 JSON 失败)
		if strings.Contains(content, "content:") {
			// 这里可以添加更复杂的手动解析逻辑，暂时先跳过
		}
	}

	// 4. 使用 Indexer 入库
	indexer := rag.NewIndexer(kb, s, model.ID)
	for _, item := range learnedItems {
		category := item.Category
		fact := item.Content

		// --- 冲突检测与合并逻辑 ---
		// 搜索是否已有相似知识
		existingChunks, err := kb.Search(ctx, fact, 1, &SearchFilter{
			BotID: employee.BotID,
		})

		if err == nil && len(existingChunks) > 0 {
			topMatch := existingChunks[0]
			if topMatch.Score > 0.95 {
				// 极高相似度：跳过，避免重复
				continue
			} else if topMatch.Score > 0.6 {
				// 中等相似度：尝试合并知识
				mergePrompt := fmt.Sprintf("请合并以下两条相关的知识点，生成一个更全面、准确的版本。\n知识点1：%s\n知识点2：%s\n请直接输出合并后的内容，不要有其他解释。", topMatch.Content, fact)
				mergeResp, err := client.Chat(ctx, ChatRequest{
					Model: model.ApiModelID,
					Messages: []Message{
						{Role: RoleSystem, Content: mergePrompt},
					},
				})
				if err == nil && len(mergeResp.Choices) > 0 {
					mergedContent, _ := mergeResp.Choices[0].Message.Content.(string)
					fact = mergedContent
					clog.Info("[AutoLearn] Merged knowledge", zap.String("botID", employee.BotID))
				}
			}
		}

		// 索引内容
		title := item.Summary
		if title == "" {
			title = fmt.Sprintf("自动学习: %s", category)
		}
		source := fmt.Sprintf("auto_learn://%s/%d", employee.BotID, time.Now().UnixNano())

		// 授权给该数字员工所在的组织
		err = indexer.IndexContent(ctx, title, source, []byte(fact), "learned", "system", "bot", employee.BotID)
		if err != nil {
			clog.Error("[AutoLearn] Failed to index learned content", zap.Error(err))
		}
	}
}

// EvaluateAndRecordKpi AI 自动对回复质量进行打分
func (s *AIServiceImpl) EvaluateAndRecordKpi(ctx context.Context, employee *models.DigitalEmployee, userMsg, aiResp string) {
	if s.employeeService == nil {
		return
	}

	sessionID, _ := ctx.Value("sessionID").(string)

	prompt := fmt.Sprintf(`你是一个专业的质量检查员。请根据以下对话，对数字员工的回复进行评分。
数字员工角色：%s (%s)
用户输入：%s
员工回复：%s

评分标准 (0-100分)：
1. 专业度：回复是否符合角色设定和专业背景。
2. 准确性：信息是否准确。
3. 礼貌度：语气是否得体。
4. 效率：是否直接解决了用户的问题。

请仅回复一个 0-100 之间的数字，不要有任何其他文字。`, employee.Name, employee.Title, userMsg, aiResp)

	// 获取默认模型
	var model models.AIModel
	if err := s.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		if err := s.db.First(&model).Error; err != nil {
			return
		}
	}

	provider, _ := s.GetProvider(model.ProviderID)
	client, _ := s.getClient(provider, &model)
	if client == nil {
		return
	}

	req := ChatRequest{
		Model: model.ApiModelID,
		Messages: []Message{
			{Role: RoleSystem, Content: prompt},
		},
	}

	resp, err := client.Chat(ctx, req)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		return
	}

	content, _ := resp.Choices[0].Message.Content.(string)
	content = strings.TrimSpace(content)

	// 解析分数
	var score float64
	fmt.Sscanf(content, "%f", &score)

	if score > 0 {
		// 限制分数范围
		if score > 100 {
			score = 100
		}
		s.employeeService.RecordKpi(employee.ID, "ai_auto_eval", score)
		clog.Info("[KPI] AI Auto Scored", zap.Uint("employeeID", employee.ID), zap.Float64("score", score))

		// 记录追踪
		if sessionID != "" {
			s.SaveTrace(sessionID, employee.BotID, 99, "kpi_score", fmt.Sprintf("%.1f", score), "")
		}
	}
}

// MaybeConsolidateMemories 检查并决定是否需要固化记忆
func (s *AIServiceImpl) MaybeConsolidateMemories(ctx context.Context, userID string, botID string) {
	if s.memoryService == nil {
		return
	}

	// 简单的触发逻辑：获取该用户记忆总数
	var count int64
	s.db.Model(&models.CognitiveMemory{}).Where("user_id = ? AND bot_id = ?", userID, botID).Count(&count)

	// 如果记忆超过 20 条，尝试固化
	if count >= 20 {
		go s.memoryService.ConsolidateMemories(ctx, userID, botID, s)
	}
}

// ConsolidateUserMemories 固化并优化用户记忆 (已迁移到 CognitiveMemoryService)
func (s *AIServiceImpl) ConsolidateUserMemories(ctx context.Context, userID string, botID string) {
	s.memoryService.ConsolidateMemories(ctx, userID, botID, s)
}
