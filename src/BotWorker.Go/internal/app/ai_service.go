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
	"strconv"
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

func (s *WorkerAIService) getClient(provider *models.AIProvider, model *models.AIModel) (ai.Client, error) {
	baseURL := provider.BaseURL
	apiKey := provider.APIKey

	if provider.Type == "mock" {
		return ai.NewMockClient(baseURL, apiKey), nil
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

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("AI config not set for provider %s", provider.Name)
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
	if s.clientsByConfig == nil {
		s.clientsByConfig = make(map[string]ai.Client)
	}
	s.clientsByConfig[cacheKey] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *WorkerAIService) Chat(ctx context.Context, modelID uint, messages []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	var model models.AIModel
	var err error

	if plugins.GlobalGORMDB != nil {
		if modelID > 0 {
			err = plugins.GlobalGORMDB.First(&model, modelID).Error
		} else {
			// 使用默认模型
			err = plugins.GlobalGORMDB.Where("\"IsDefault\" = ?", true).First(&model).Error
			if err != nil {
				err = plugins.GlobalGORMDB.First(&model).Error
			}
		}
	}

	var provider models.AIProvider
	if err == nil && plugins.GlobalGORMDB != nil {
		err = plugins.GlobalGORMDB.First(&provider, model.ProviderID).Error
	}

	var client ai.Client
	if err == nil && plugins.GlobalGORMDB != nil {
		client, err = s.getClient(&provider, &model)
	} else {
		// 回退到配置文件
		client, err = s.getLegacyClient()
	}

	if err != nil {
		return nil, err
	}

	apiModelID := model.ApiModelID
	if apiModelID == "" {
		apiModelID = s.config.AI.Model
	}
	if apiModelID == "" {
		apiModelID = "gpt-3.5-turbo"
	}

	req := types.ChatRequest{
		Model:    apiModelID,
		Messages: messages,
		Tools:    tools,
	}

	resp, err := client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *WorkerAIService) getLegacyClient() (ai.Client, error) {
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
	if s.clientsByConfig == nil {
		s.clientsByConfig = make(map[string]ai.Client)
	}
	s.clientsByConfig[cacheKey] = newClient
	s.mu.Unlock()

	return newClient, nil
}

func (s *WorkerAIService) ChatAgent(ctx context.Context, modelID uint, messages []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	// 获取上下文中的 SessionID, BotID 等
	sessionID, _ := ctx.Value("session_id").(string)
	botID, _ := ctx.Value("bot_id").(string)
	userIDStr, _ := ctx.Value("user_id").(string)
	orgIDStr, _ := ctx.Value("org_id").(string)

	if sessionID == "" {
		sessionID = "worker_session"
	}
	if botID == "" {
		botID = "worker_bot"
	}

	// 自动注入 MCP 工具
	if s.mcp != nil {
		uid, _ := strconv.ParseUint(userIDStr, 10, 64)
		oid, _ := strconv.ParseUint(orgIDStr, 10, 64)
		mcpTools, err := s.mcp.GetToolsForContext(ctx, uint(uid), uint(oid))
		if err == nil && len(mcpTools) > 0 {
			if tools == nil {
				tools = make([]types.Tool, 0)
			}
			tools = append(tools, mcpTools...)
		}
	}

	// 使用 Common/ai 中的 AgentExecutor
	executor := ai.NewAgentExecutor(s, modelID, botID, userIDStr, sessionID)
	return executor.Execute(ctx, messages, tools)
}

func (s *WorkerAIService) CreateEmbedding(ctx context.Context, modelID uint, input any) (*types.EmbeddingResponse, error) {
	var model models.AIModel
	var err error

	if plugins.GlobalGORMDB != nil {
		if modelID > 0 {
			err = plugins.GlobalGORMDB.First(&model, modelID).Error
		} else {
			// 使用默认模型
			err = plugins.GlobalGORMDB.Where("\"IsDefault\" = ?", true).First(&model).Error
			if err != nil {
				err = plugins.GlobalGORMDB.First(&model).Error
			}
		}
	}

	var provider models.AIProvider
	if err == nil && plugins.GlobalGORMDB != nil {
		err = plugins.GlobalGORMDB.First(&provider, model.ProviderID).Error
	}

	var client ai.Client
	if err == nil && plugins.GlobalGORMDB != nil {
		client, err = s.getClient(&provider, &model)
	} else {
		// 回退到配置文件
		client, err = s.getLegacyClient()
	}

	if err != nil {
		return nil, err
	}

	apiModelID := model.ApiModelID
	if apiModelID == "" {
		apiModelID = s.config.AI.Model
	}
	if apiModelID == "" {
		apiModelID = "text-embedding-3-small"
	}

	req := types.EmbeddingRequest{
		Model: apiModelID,
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

func (s *WorkerAIService) GetProvider(id uint) (*models.AIProvider, error) {
	return nil, fmt.Errorf("GetProvider not implemented on WorkerAIService")
}

func (s *WorkerAIService) DispatchIntent(msg types.InternalMessage) (string, error) {
	return "", fmt.Errorf("DispatchIntent not implemented on WorkerAIService")
}

func (s *WorkerAIService) ChatWithEmployee(employee *models.DigitalEmployee, msg types.InternalMessage, targetOrgID uint) (string, error) {
	return "", fmt.Errorf("ChatWithEmployee not implemented on WorkerAIService")
}
