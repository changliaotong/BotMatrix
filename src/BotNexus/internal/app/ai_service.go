package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type AIServiceImpl struct {
	db      *gorm.DB
	manager *Manager
	clients map[uint]ai.Client
	mu      sync.RWMutex
}

func NewAIService(db *gorm.DB, m *Manager) *AIServiceImpl {
	return &AIServiceImpl{
		db:      db,
		manager: m,
		clients: make(map[uint]ai.Client),
	}
}

func (s *AIServiceImpl) GetProvider(id uint) (*models.AIProviderGORM, error) {
	var provider models.AIProviderGORM
	if err := s.db.First(&provider, id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *AIServiceImpl) getClient(providerID uint) (ai.Client, error) {
	s.mu.RLock()
	client, ok := s.clients[providerID]
	s.mu.RUnlock()
	if ok {
		return client, nil
	}

	provider, err := s.GetProvider(providerID)
	if err != nil {
		return nil, err
	}

	// 目前只实现了 OpenAI 兼容适配器，大部分主流模型（DeepSeek, Ollama, Azure）都支持
	newClient := ai.NewOpenAIAdapter(provider.BaseURL, provider.APIKey)

	s.mu.Lock()
	s.clients[providerID] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *AIServiceImpl) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	client, err := s.getClient(model.ProviderID)
	if err != nil {
		return nil, err
	}

	req := ai.ChatRequest{
		Model:    model.ModelID,
		Messages: messages,
		Tools:    tools,
	}

	// 打印 LLM 调用详情
	fmt.Printf("\n--- [LLM CALL START] ---\n")
	fmt.Printf("Model: %s (%s)\n", model.ModelName, model.ModelID)
	for i, msg := range messages {
		fmt.Printf("Message #%d [%s]: %s\n", i, msg.Role, msg.Content)
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
		choice := resp.Choices[0]
		if choice.Message.Content != "" {
			fmt.Printf("Response: %s\n", choice.Message.Content)
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

func (s *AIServiceImpl) ChatStream(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (<-chan ai.ChatStreamResponse, error) {
	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	client, err := s.getClient(model.ProviderID)
	if err != nil {
		return nil, err
	}

	req := ai.ChatRequest{
		Model:    model.ModelID,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	}

	// 打印 LLM 流式调用详情
	fmt.Printf("\n--- [LLM STREAM START] ---\n")
	fmt.Printf("Model: %s (%s)\n", model.ModelName, model.ModelID)
	for i, msg := range messages {
		fmt.Printf("Message #%d [%s]: %s\n", i, msg.Role, msg.Content)
	}
	fmt.Printf("--- [STREAMING...] ---\n\n")

	return client.ChatStream(ctx, req)
}

func (s *AIServiceImpl) CreateEmbedding(ctx context.Context, modelID uint, input []string) (*ai.EmbeddingResponse, error) {
	var model models.AIModelGORM
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}

	client, err := s.getClient(model.ProviderID)
	if err != nil {
		return nil, err
	}

	req := ai.EmbeddingRequest{
		Model: model.ModelID,
		Input: input,
	}

	return client.CreateEmbedding(ctx, req)
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
		return choice.Message.Content, nil
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
	resp, err := s.Chat(context.Background(), modelID, messages, nil)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content

		// 6. 异步保存消息到历史
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
		}()

		return content, nil
	}

	return "", fmt.Errorf("no response from ai")
}
