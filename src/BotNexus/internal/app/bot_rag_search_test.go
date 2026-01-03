package app

import (
	"BotMatrix/common/models"
	"BotNexus/internal/rag"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRAGHybridSearch(t *testing.T) {
	// 1. 初始化数据库 (使用内存数据库)
	dbPath := "rag_search_test.db"
	defer os.Remove(dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.AIProviderGORM{}, &models.AIModelGORM{})

	// 2. 初始化 AI 服务模拟
	mockAI := NewAIService(db, nil)

	db.Create(&models.AIProviderGORM{ID: 1, Name: "Volcengine", Type: "openai", BaseURL: "https://ark.cn-beijing.volces.com/api/v3"})
	db.Create(&models.AIModelGORM{ID: 1, ProviderID: 1, ModelID: "doubao-embedding-vision-251215", ModelName: "Doubao Embedding Vision"})

	// 3. 初始化 RAG 组件
	es := rag.NewTaskAIEmbeddingService(mockAI, 1, "doubao-embedding-vision-251215")
	kb := rag.NewPostgresKnowledgeBase(db, es, mockAI, 1)
	kb.Setup()

	indexer := rag.NewIndexer(kb, mockAI, 1)
	ctx := context.Background()

	// 4. 准备测试数据
	testDoc := `package test
// GetUserByID 根据 ID 获取用户信息
func GetUserByID(id string) (*User, error) {
    return &User{ID: id, Name: "Test"}, nil
}
`
	testPath := "test_code.go"
	os.WriteFile(testPath, []byte(testDoc), 0644)
	defer os.Remove(testPath)

	err = indexer.IndexFile(ctx, testPath, "code")
	if err != nil {
		t.Fatalf("IndexFile failed: %v", err)
	}

	// 5. 测试关键词搜索 (Hybrid Search 应该能搜到 "GetUserByID")
	fmt.Println("=== 开始测试关键词搜索 ===")
	results, err := kb.Search(ctx, "GetUserByID", 3, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	found := false
	for _, res := range results {
		fmt.Printf("找到结果: %s\n", res.Content)
		if stringsContains(res.Content, "GetUserByID") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Hybrid search failed to find content by keyword 'GetUserByID'")
	} else {
		fmt.Println("Hybrid search successfully found keyword!")
	}
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
