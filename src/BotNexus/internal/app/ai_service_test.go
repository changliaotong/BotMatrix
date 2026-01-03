package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// MockClient 模拟 AI 客户端
type MockClient struct{}

func (m *MockClient) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return &ai.ChatResponse{
		ID: "test-id",
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: "这是模拟的 AI 回复内容。",
				},
				FinishReason: "stop",
			},
		},
		Usage: ai.UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

func (m *MockClient) ChatStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.ChatStreamResponse, error) {
	ch := make(chan ai.ChatStreamResponse)
	go func() {
		defer close(ch)
		ch <- ai.ChatStreamResponse{
			ID: "test-id",
			Choices: []ai.StreamChoice{
				{
					Delta: ai.MessageDelta{Content: "这是模拟的"},
				},
			},
		}
		ch <- ai.ChatStreamResponse{
			ID: "test-id",
			Choices: []ai.StreamChoice{
				{
					Delta: ai.MessageDelta{Content: "流式回复"},
				},
			},
		}
	}()
	return ch, nil
}

func (m *MockClient) CreateEmbedding(ctx context.Context, req ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	var inputLen int
	if ss, ok := req.Input.([]string); ok {
		inputLen = len(ss)
	} else if ms, ok := req.Input.([]map[string]any); ok {
		inputLen = len(ms)
	} else {
		inputLen = 1
	}

	embeddings := make([]ai.EmbeddingData, inputLen)
	for i := 0; i < inputLen; i++ {
		// 生成一个伪随机向量 (2048 维)
		vec := make([]float32, 2048)
		for j := 0; j < 2048; j++ {
			vec[j] = float32(i+j) / 1000.0
		}
		embeddings[i] = ai.EmbeddingData{
			Embedding: vec,
			Index:     i,
		}
	}
	return &ai.EmbeddingResponse{
		Data:  embeddings,
		Model: req.Model,
		Usage: ai.UsageInfo{TotalTokens: 100},
	}, nil
}

func TestAIServiceLogging(t *testing.T) {
	// 1. 设置内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.AIProviderGORM{}, &models.AIModelGORM{}, &models.AIUsageLogGORM{})

	// 2. 插入测试数据
	provider := models.AIProviderGORM{
		Name:    "TestProvider",
		BaseURL: "https://api.test.com",
		APIKey:  "test-key",
	}
	db.Create(&provider)

	model := models.AIModelGORM{
		ProviderID: provider.ID,
		ModelName:  "TestModel",
		ModelID:    "gpt-3.5-turbo",
	}
	db.Create(&model)

	// 3. 创建 AI 服务并手动注入模拟客户端
	s := NewAIService(db, nil)
	// 使用 clientsByConfig 注入，Key 需要匹配 s.getClient 生成的规则
	cacheKey := "https://api.test.com|test-key"
	s.clientsByConfig[cacheKey] = &MockClient{}

	// 4. 执行 Chat 调用，触发打印
	fmt.Println("\n=== 开始测试 Chat 日志打印 ===")
	messages := []ai.Message{
		{Role: ai.RoleSystem, Content: "你是一个助手"},
		{Role: ai.RoleUser, Content: "你好，请自我介绍"},
	}
	tools := []ai.Tool{
		{
			Type: "function",
			Function: ai.FunctionDefinition{
				Name:        "test_tool",
				Description: "测试工具",
			},
		},
	}

	_, err = s.Chat(context.Background(), model.ID, messages, tools)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// 5. 执行 ChatStream 调用，触发打印
	fmt.Println("\n=== 开始测试 ChatStream 日志打印 ===")
	_, err = s.ChatStream(context.Background(), model.ID, messages, nil)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	// 给异步记录日志一点时间
	time.Sleep(100 * time.Millisecond)
}
