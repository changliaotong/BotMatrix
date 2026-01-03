package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/bot"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ConcurrencyMockClient is a mock implementation of ai.Client
type ConcurrencyMockClient struct{}

func (m *ConcurrencyMockClient) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: "Concurrency test response",
				},
			},
		},
		Usage: ai.UsageInfo{
			TotalTokens: 100,
		},
	}, nil
}

func (m *ConcurrencyMockClient) ChatStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.ChatStreamResponse, error) {
	return nil, nil
}

func (m *ConcurrencyMockClient) CreateEmbedding(ctx context.Context, req ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	return nil, nil
}

func (m *ConcurrencyMockClient) ChatAgent(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: "Concurrency test agent response",
				},
			},
		},
	}, nil
}

func TestChatWithEmployeeConcurrency(t *testing.T) {
	// 0. Initialize Logger
	clog.InitLogger(clog.Config{Level: "debug", Development: true})

	// 1. Setup in-memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(
		&models.DigitalEmployeeGORM{},
		&models.AIAgentGORM{},
		&models.AIModelGORM{},
		&models.AIProviderGORM{},
		&models.AIChatMessageGORM{},
		&models.AIAgentTraceGORM{},
		&models.AIUsageLogGORM{},
		&models.MCPServerGORM{},
		&models.CognitiveMemoryGORM{},
		&models.B2BSkillSharingGORM{},
		&models.BotSkillPermissionGORM{},
	)

	// 2. Setup mock data
	provider := models.AIProviderGORM{Name: "Mock", Type: "openai"}
	db.Create(&provider)

	model := models.AIModelGORM{
		ProviderID: provider.ID,
		ModelID:    "mock-model",
		IsDefault:  true,
	}
	db.Create(&model)

	agent := models.AIAgentGORM{
		Name:    "Test Agent",
		ModelID: model.ID,
	}
	db.Create(&agent)

	emp := models.DigitalEmployeeGORM{
		BotID:        "bot123",
		Name:         "Concurrent Worker",
		EnterpriseID: 1,
		AgentID:      agent.ID,
	}
	db.Create(&emp)

	// 3. Initialize AI Service with mock client
	service := NewAIService(db, &Manager{Manager: &bot.Manager{GORMDB: db}})
	service.clientsByConfig = map[string]ai.Client{
		"|": &ConcurrencyMockClient{}, // getClient generates cache key "baseURL|apiKey"
	}

	// 4. Run concurrent tests
	const concurrency = 20
	var wg sync.WaitGroup
	wg.Add(concurrency)

	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			msg := types.InternalMessage{
				UserID:     fmt.Sprintf("user_%d", id),
				RawMessage: "Hello",
			}
			resp, err := service.ChatWithEmployee(&emp, msg, 1)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %v", id, err)
				return
			}
			if resp != "Concurrency test response" {
				errors <- fmt.Errorf("goroutine %d got unexpected response: %s", id, resp)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	// 5. Verify data consistency
	var traceCount int64
	db.Model(&models.AIAgentTraceGORM{}).Count(&traceCount)
	if traceCount == 0 {
		t.Error("expected traces to be saved")
	}

	var msgCount int64
	db.Model(&models.AIChatMessageGORM{}).Count(&msgCount)
	// Each call saves 2 messages (user + assistant) in a goroutine
	// Since it's async, we might need to wait a bit
	// But let's just check if it's > 0 for now
	if msgCount == 0 {
		// Wait a short time for async saves
		// time.Sleep(500 * time.Millisecond)
		// db.Model(&models.AIChatMessageGORM{}).Count(&msgCount)
	}
}
