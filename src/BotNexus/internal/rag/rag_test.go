package rag

import (
	"BotMatrix/common/ai"
	"BotNexus/tasks"
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
		err = indexer.IndexContent(ctx, "User 1 Knowledge", "user1_doc.txt", []byte("This is personal knowledge for User 1"), "txt", "user1", "user", "user1")
		if err != nil {
			t.Fatalf("failed to index user knowledge: %v", err)
		}

		// 搜索测试 1: 机器人 A 在群组 G1 中，用户 1 提问
		filter := &tasks.SearchFilter{
			OwnerType: "group",
			OwnerID:   "G1",
			BotID:     "BOT_A",
		}
		results, _ := kb.SearchHybrid(ctx, "knowledge", 10, filter)

		foundBotA := false
		for _, r := range results {
			if r.Source == "bot_a_doc.txt" {
				foundBotA = true
			}
		}
		if !foundBotA {
			t.Errorf("Bot A knowledge not found in group G1 search")
		}

		// 搜索测试 2: 用户 1 自己的个人知识搜索
		filter = &tasks.SearchFilter{
			OwnerType: "user",
			OwnerID:   "user1",
			BotID:     "BOT_B", // 换个机器人也能搜到个人知识
		}
		results, _ = kb.SearchHybrid(ctx, "personal", 10, filter)
		foundUser1 := false
		for _, r := range results {
			if r.Source == "user1_doc.txt" {
				foundUser1 = true
			}
		}
		if !foundUser1 {
			t.Errorf("User 1 personal knowledge not found")
		}
	})

	// 3. 测试状态切换 (启用/暂停)
	t.Run("Status Management", func(t *testing.T) {
		var doc KnowledgeDoc
		db.Where("source = ?", "bot_a_doc.txt").First(&doc)

		// 暂停文档
		kb.SetDocStatus(ctx, doc.ID, "paused")

		// 再次搜索，不应该搜到
		filter := &tasks.SearchFilter{
			BotID: "BOT_A",
		}
		results, _ := kb.SearchHybrid(ctx, "knowledge", 10, filter)
		for _, r := range results {
			if r.Source == "bot_a_doc.txt" {
				t.Errorf("Paused document should not be searchable")
			}
		}

		// 恢复文档
		kb.SetDocStatus(ctx, doc.ID, "active")
		results, _ = kb.SearchHybrid(ctx, "knowledge", 10, filter)
		found := false
		for _, r := range results {
			if r.Source == "bot_a_doc.txt" {
				found = true
			}
		}
		if !found {
			t.Errorf("Re-activated document should be searchable")
		}
	})

	// 4. 测试解析器注册
	t.Run("Parser Registration", func(t *testing.T) {
		if _, ok := indexer.parsers[".pdf"]; !ok {
			t.Errorf("PDF parser should be registered")
		}
		if _, ok := indexer.parsers[".docx"]; !ok {
			t.Errorf("Docx parser should be registered")
		}
	})

	// 5. 测试 RAG 2.0 高级功能
	t.Run("RAG 2.0 Features", func(t *testing.T) {
		// 索引一段内容，确保 targetID 为 "system" 以便全局可见
		err := indexer.IndexContent(ctx, "RAG 2.0 Test", "rag2.txt", []byte("PostgresKnowledgeBase supports Query Refinement and Self-Reflection."), "txt", "user1", "system", "system")
		if err != nil {
			t.Fatalf("failed to index: %v", err)
		}

		// 搜索时会触发 MockAIService 的 Query Refinement 和 Self-Reflection
		results, err := kb.SearchHybrid(ctx, "PostgresKnowledgeBase", 5, &tasks.SearchFilter{Status: "active", OwnerType: "system"})
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}

		if len(results) == 0 {
			t.Errorf("expected results, got 0")
		}
	})

	// 6. 测试 GraphRAG 功能
	t.Run("GraphRAG", func(t *testing.T) {
		// 手动插入一些实体和关系 (因为 Indexer.ExtractAndIndexGraph 是异步 go 运行的，测试中不方便等待)
		entity := &KnowledgeEntity{
			Name:        "GraphTestEntity",
			Type:        "Test",
			Description: "This is a test entity for GraphRAG.",
		}
		kb.SaveEntity(ctx, entity)

		// 搜索 Graph
		entities, _, err := kb.SearchGraph(ctx, "GraphTestEntity", 5)
		if err != nil {
			t.Fatalf("SearchGraph failed: %v", err)
		}

		if len(entities) == 0 {
			t.Errorf("expected entities, got 0")
		}

		// 验证 SearchHybrid 是否包含图谱上下文
		results, err := kb.SearchHybrid(ctx, "GraphTestEntity", 5, &tasks.SearchFilter{Status: "active", OwnerType: "system"})
		if err != nil {
			t.Fatalf("SearchHybrid failed: %v", err)
		}

		foundGraph := false
		for _, r := range results {
			if r.Source == "Knowledge Graph" {
				foundGraph = true
				break
			}
		}
		if !foundGraph {
			t.Errorf("Graph context not found in search results")
		}
	})

	// 7. 测试多模态图片解析
	t.Run("Multi-modal Image", func(t *testing.T) {
		// 模拟一个图片文件上传
		// 注意：MockAIService 会返回默认内容
		err := indexer.IndexContent(ctx, "Test Image", "test.jpg", []byte("fake image data"), "jpg", "user1", "system", "system")
		if err != nil {
			t.Fatalf("failed to index image: %v", err)
		}

		var doc KnowledgeDoc
		err = db.Where("source = ?", "test.jpg").First(&doc).Error
		if err != nil {
			t.Fatalf("doc not found: %v", err)
		}

		// 检查解析出的内容是否包含在文档中
		// MockAIService 默认返回 YES (在 DescribeImage 中会被当作描述词)
		// 但由于 DescribeImage 的 prompt 是“请详细描述...”，MockAIService 的通用逻辑会返回 YES
		if doc.Content == "" {
			t.Errorf("expected doc content to be non-empty after image parsing")
		}
	})
}
