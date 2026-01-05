package rag

import (
	"BotMatrix/common/ai"
	"context"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type MockEmbeddingService struct{}

func (m *MockEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return make([]float32, 1536), nil
}
func (m *MockEmbeddingService) GenerateQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	return make([]float32, 1536), nil
}

type MockAIService struct{}

func (m *MockAIService) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	content := "YES" // 默认返回 YES 以通过 Self-Reflection
	if len(messages) > 0 {
		if msg, ok := messages[len(messages)-1].Content.(string); ok {
			if strings.Contains(msg, "改写") {
				// 提取提问内容，模拟提取关键词
				if strings.Contains(msg, "提问：") {
					parts := strings.Split(msg, "提问：")
					if len(parts) > 1 {
						queryPart := parts[1]
						if idx := strings.Index(queryPart, "\n"); idx != -1 {
							content = strings.TrimSpace(queryPart[:idx])
						} else {
							content = strings.TrimSpace(queryPart)
						}
					}
				}
			} else if strings.Contains(msg, "输出格式必须为 JSON") {
				content = `{
					"entities": [{"name": "Postgres", "type": "Database", "description": "A powerful relational database."}],
					"relations": [{"subject": "Postgres", "predicate": "supports", "object": "pgvector", "description": "Postgres supports vector search via pgvector."}]
				}`
			} else if strings.Contains(msg, "判断以下检索到的内容") {
				content = "YES"
			}
		}
	}
	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: content,
				},
			},
		},
	}, nil
}

func (m *MockAIService) ChatAgent(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	return m.Chat(ctx, modelID, messages, tools)
}

func (m *MockAIService) CreateEmbedding(ctx context.Context, modelID uint, input any) (*ai.EmbeddingResponse, error) {
	return &ai.EmbeddingResponse{
		Data: []ai.EmbeddingData{
			{Embedding: make([]float32, 1536)},
		},
	}, nil
}

func TestRAGIntegration(t *testing.T) {
	// 1. 设置内存数据库 (使用共享缓存以确保多连接下表不丢失)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	kb := &PostgresKnowledgeBase{
		db:               db,
		embeddingService: &MockEmbeddingService{},
		aiSvc:            &MockAIService{},
		aiModelID:        1,
	}
	if err := kb.Setup(); err != nil {
		t.Fatalf("failed to setup kb: %v", err)
	}

	indexer := NewIndexer(kb, &MockAIService{}, 0)
	ctx := context.Background()

	// 2. 测试 BotID 过滤和共享
	t.Run("BotID Filtering", func(t *testing.T) {
		// 上传一个绑定到机器人 A 的文档
		err := indexer.IndexContent(ctx, "Bot A Knowledge", "bot_a_doc.txt", []byte("This is knowledge for Bot A"), "txt", "user1", "bot", "BOT_A")
		if err != nil {
			t.Fatalf("failed to index bot knowledge: %v", err)
		}

		// 上传一个绑定到用户 1 的个人知识 (跨群共享)
		err = indexer.IndexContent(ctx, "User 1 Personal", "user1_doc.txt", []byte("This is personal knowledge for User 1"), "txt", "user1", "user", "")
		if err != nil {
			t.Fatalf("failed to index personal knowledge: %v", err)
		}

		// 场景 A: 机器人 A 搜索 (应该能看到机器人 A 的文档 + 用户 1 的个人文档)
		resultsA, err := kb.Search(ctx, "knowledge", 10, &ai.SearchFilter{
			BotID:  "BOT_A",
			UserID: "user1",
		})
		if err != nil {
			t.Fatalf("search A failed: %v", err)
		}
		if len(resultsA) < 2 {
			t.Errorf("expected at least 2 results for Bot A, got %d", len(resultsA))
		}

		// 场景 B: 机器人 B 搜索 (应该只能看到用户 1 的个人文档，不能看到机器人 A 的文档)
		resultsB, err := kb.Search(ctx, "knowledge", 10, &ai.SearchFilter{
			BotID:  "BOT_B",
			UserID: "user1",
		})
		if err != nil {
			t.Fatalf("search B failed: %v", err)
		}
		// 这里在 mock 环境下由于没有真正的向量搜索，可能返回不准确，但逻辑上应该过滤
		// 在 PostgresKnowledgeBase.SearchHybrid 中有真正的 SQL 过滤逻辑
		for _, res := range resultsB {
			if res.BotID == "BOT_A" {
				t.Errorf("Bot B should not see Bot A's knowledge")
			}
		}
	})
}
