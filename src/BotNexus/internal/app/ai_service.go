package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"sync"

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

	return client.Chat(ctx, req)
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

	return client.ChatStream(ctx, req)
}

func (s *AIServiceImpl) DispatchIntent(msg types.InternalMessage) (string, error) {
	if s.manager.TaskManager == nil || s.manager.TaskManager.AI == nil {
		return "", fmt.Errorf("task manager or ai parser not initialized")
	}

	// 1. 获取所有可用工具定义
	tools := s.manager.TaskManager.AI.Manifest.GenerateTools()
	if len(tools) == 0 {
		return "", fmt.Errorf("no tools available")
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
	messages := []ai.Message{
		{
			Role:    ai.RoleSystem,
			Content: s.manager.TaskManager.AI.Manifest.GenerateSystemPrompt(),
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
