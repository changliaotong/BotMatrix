package app

import (
	"BotMatrix/common/models"
	"BotNexus/internal/rag"
	"BotNexus/tasks"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRAGIngestion(t *testing.T) {
	// 1. 初始化数据库 (使用本地文件进行持久化测试，或使用 :memory:)
	dbPath := "rag_test.db"
	defer os.Remove(dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.AIProviderGORM{}, &models.AIModelGORM{})

	// 2. 初始化 AI 服务模拟
	mockAI := NewAIService(db, nil)

	// 创建测试模型
	db.Create(&models.AIProviderGORM{ID: 1, Name: "Volcengine", Type: "openai", BaseURL: "https://ark.cn-beijing.volces.com/api/v3"})
	db.Create(&models.AIModelGORM{ID: 1, ProviderID: 1, ModelID: "doubao-embedding-vision-251215", ModelName: "Doubao Embedding Vision"})

	// 3. 初始化 RAG 组件
	es := rag.NewTaskAIEmbeddingService(mockAI, 1, "doubao-embedding-vision-251215")
	kb := rag.NewPostgresKnowledgeBase(db, es, mockAI, 1)

	// 注意：SQLite 不支持 pgvector 的 <=> 操作符，所以这里的 Search 会报错
	// 但我们可以测试 Setup 和 Indexing 过程
	if err := kb.Setup(); err != nil {
		t.Fatalf("RAG setup failed: %v", err)
	}

	indexer := rag.NewIndexer(kb, mockAI, 1)
	ctx := context.Background()

	// 4. 递归索引整个项目文档
	baseDir := "d:/projects/BotMatrix"
	fmt.Println("=== 开始全量灌入知识库 ===")

	// 索引文档 (.md)
	docsDirs := []string{
		"docs",
		"src/BotNexus/tasks",
	}
	for _, dir := range docsDirs {
		err = indexer.IndexDirectory(ctx, filepath.Join(baseDir, dir), "doc", []string{".md"})
		if err != nil {
			t.Errorf("索引文档目录 %s 失败: %v", dir, err)
		}
	}

	// 索引根目录文档
	err = indexer.IndexFile(ctx, filepath.Join(baseDir, "README.md"), "doc")
	if err != nil {
		fmt.Printf("跳过 README.md (可能不存在): %v\n", err)
	}

	// 索引核心代码 (.go) - 只选关键目录
	codeDirs := []string{
		"src/BotNexus/tasks",
		"src/BotNexus/internal/rag",
		"src/BotNexus/internal/app",
		"src/Common/ai",
		"src/Common/models",
	}

	for _, dir := range codeDirs {
		err = indexer.IndexDirectory(ctx, filepath.Join(baseDir, dir), "code", []string{".go"})
		if err != nil {
			t.Errorf("索引代码目录 %s 失败: %v", dir, err)
		}
	}

	// 索引内置技能
	testSkills := []tasks.Capability{
		{
			Name:        "weather",
			Description: "获取指定城市的天气预报",
			Params:      map[string]string{"city": "城市名称，如北京、上海"},
			Required:    []string{"city"},
			Example:     "北京天气怎么样？",
			Category:    "tools",
		},
		{
			Name:        "translate",
			Description: "将文本翻译为指定语言",
			Params:      map[string]string{"text": "待翻译文本", "to": "目标语言"},
			Required:    []string{"text"},
			Example:     "帮我把‘Hello’翻译成中文",
			Category:    "tools",
		},
	}
	err = indexer.IndexSkills(ctx, testSkills)
	if err != nil {
		t.Errorf("索引技能失败: %v", err)
	}

	// 5. 验证数据是否入库
	var docCount int64
	db.Model(&rag.KnowledgeDoc{}).Count(&docCount)
	fmt.Printf("成功灌入文档数量: %d\n", docCount)

	var chunkCount int64
	db.Model(&rag.KnowledgeChunk{}).Count(&chunkCount)
	fmt.Printf("成功生成知识切片数量: %d\n", chunkCount)

	if docCount == 0 || chunkCount == 0 {
		t.Errorf("灌入结果异常: docs=%d, chunks=%d", docCount, chunkCount)
	}

	// 打印一些切片内容
	var sampleChunks []rag.KnowledgeChunk
	db.Limit(3).Find(&sampleChunks)
	for i, c := range sampleChunks {
		fmt.Printf("切片 %d (DocID: %d): %s...\n", i+1, c.DocID, c.Content[:100])
	}

	// 6. 验证是否包含核心文档
	var dCount int64
	db.Model(&rag.KnowledgeDoc{}).Count(&dCount)
	fmt.Printf("当前数据库中已索引文档数: %d\n", dCount)

	var cCount int64
	db.Model(&rag.KnowledgeChunk{}).Count(&cCount)
	fmt.Printf("当前数据库中已生成向量分片数: %d\n", cCount)

	if dCount == 0 {
		t.Log("警告：当前数据库中没有已向量化的文档。")
	}
}
