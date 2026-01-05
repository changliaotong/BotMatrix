package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/mcp"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"botworker/internal/config"
	"botworker/plugins"
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// WorkerAIService 为 Worker 提供的简单 AI 服务实现
type WorkerAIService struct {
	config          *config.Config
	clientsByConfig map[string]ai.Client
	kb              *rag.PostgresKnowledgeBase
	mcp             *mcp.MCPManager
	mu              sync.RWMutex
}

func NewWorkerAIService(cfg *config.Config) *WorkerAIService {
	s := &WorkerAIService{
		config:          cfg,
		clientsByConfig: make(map[string]ai.Client),
	}
	s.mcp = mcp.NewMCPManager(s)

	// 如果配置了数据库，初始化 KB
	if cfg.Database.Host != "" && plugins.GlobalGORMDB != nil {
		// Worker 端通常使用简单的 OpenAIExtractor 作为 EmbeddingService
		// 或者通过配置指定
		es := ai.NewOpenAIAdapter(cfg.AI.Endpoint, cfg.AI.APIKey)
		kb := rag.NewPostgresKnowledgeBase(plugins.GlobalGORMDB, es, s, 0)
		s.kb = kb
	}

	return s
}

func (s *WorkerAIService) GetGORMDB() *gorm.DB                                     { return plugins.GlobalGORMDB }
func (s *WorkerAIService) GetAIService() types.AIService                           { return s }
func (s *WorkerAIService) GetB2BService() types.B2BService                         { return nil }
func (s *WorkerAIService) GetCognitiveMemoryService() types.CognitiveMemoryService { return nil }
func (s *WorkerAIService) GetDigitalEmployeeService() types.DigitalEmployeeService { return nil }
func (s *WorkerAIService) GetDigitalEmployeeTaskService() types.DigitalEmployeeTaskService {
	return nil
}
func (s *WorkerAIService) GetTaskManager() types.TaskManagerInterface            { return nil }
func (s *WorkerAIService) ValidateToken(token string) (*types.UserClaims, error) { return nil, nil }

func (s *WorkerAIService) GetMCPManager() types.MCPManagerInterface {
	return s.mcp
}

func (s *WorkerAIService) GetKnowledgeBase() types.KnowledgeBase {
	return s.kb
}

func (s *WorkerAIService) getClient() (ai.Client, error) {
	baseURL := s.config.AI.Endpoint
	apiKey := s.config.AI.APIKey

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("AI config not set (endpoint or api_key is empty)")
	}

	cacheKey := fmt.Sprintf("%s|%s", baseURL, apiKey)

	s.mu.RLock()
	client, ok := s.clientsByConfig[cacheKey]
	s.mu.RUnlock()
	if ok {
		return client, nil
	}

	newClient := ai.NewOpenAIAdapter(baseURL, apiKey)
	s.mu.Lock()
	s.clientsByConfig[cacheKey] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *WorkerAIService) Chat(ctx context.Context, modelID uint, messages []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	model := s.config.AI.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	req := types.ChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}

	resp, err := client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *WorkerAIService) ChatAgent(ctx context.Context, modelID uint, messages []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	// Worker 端暂不实现复杂的 Agent 逻辑，直接调用 Chat
	return s.Chat(ctx, modelID, messages, tools)
}

func (s *WorkerAIService) CreateEmbedding(ctx context.Context, modelID uint, input any) (*types.EmbeddingResponse, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	req := types.EmbeddingRequest{
		Input: input,
	}

	resp, err := client.CreateEmbedding(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *WorkerAIService) ChatSimple(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	return s.Chat(ctx, 0, req.Messages, req.Tools)
}

func (s *WorkerAIService) ChatStream(ctx context.Context, modelID uint, messages []types.Message, tools []types.Tool) (<-chan types.ChatStreamResponse, error) {
	return nil, fmt.Errorf("ChatStream not implemented on WorkerAIService")
}

func (s *WorkerAIService) CreateEmbeddingSimple(ctx context.Context, req types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	return s.CreateEmbedding(ctx, 0, req.Input)
}

func (s *WorkerAIService) ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall types.ToolCall) (any, error) {
	return nil, fmt.Errorf("ExecuteTool not implemented on WorkerAIService")
}

func (s *WorkerAIService) GetProvider(id uint) (*models.AIProviderGORM, error) {
	return nil, fmt.Errorf("GetProvider not implemented on WorkerAIService")
}

func (s *WorkerAIService) DispatchIntent(msg types.InternalMessage) (string, error) {
	return "", fmt.Errorf("DispatchIntent not implemented on WorkerAIService")
}

func (s *WorkerAIService) ChatWithEmployee(employee *models.DigitalEmployeeGORM, msg types.InternalMessage, targetOrgID uint) (string, error) {
	return "", fmt.Errorf("ChatWithEmployee not implemented on WorkerAIService")
}
