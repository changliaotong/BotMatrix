package main

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/config"
	"BotMatrix/common/models"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	if err := config.InitConfig("config.json"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := config.GlobalConfig

	// 2. 连接数据库
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDBName, cfg.PGSSLMode)
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 3. 获取向量模型配置
	var embedModel models.AIModelGORM
	if err := db.Where("model_name LIKE ?", "%embedding%").First(&embedModel).Error; err != nil {
		log.Fatalf("未找到向量模型: %v", err)
	}
	fmt.Printf("使用向量模型: %s (ID: %d, ProviderID: %d, BaseURL: %s)\n",
		embedModel.ModelID, embedModel.ID, embedModel.ProviderID, embedModel.BaseURL)

	// 4. 初始化 AI 服务和 RAG
	aiService := ai.NewAIService(db, nil, nil)
	es := rag.NewTaskAIEmbeddingService(aiService, embedModel.ID, embedModel.ModelID)
	kb := rag.NewPostgresKnowledgeBase(db, es, aiService, embedModel.ID)

	// 5. 准备测试数据
	// 首先尝试索引真实的项目文档
	indexProjectDocs(db, kb, aiService, embedModel.ID)
	// 如果还是没数据，则注入模拟数据
	setupTestData(db, kb, es)

	// 6. 执行搜索测试
	queries := []string{
		"什么是机器人自举 (Bootstrap) 机制？",
		"为什么 RAG 架构选型选择了 pgvector？",
		"系统推荐使用什么向量化模型，如何部署？",
		"如何更新机器人的自我认知？",
		"数据库连不上怎么办？",
	}

	for _, query := range queries {
		fmt.Printf("\n--- 问题: %s ---\n", query)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		results, err := kb.Search(ctx, query, 3, nil)
		cancel()

		if err != nil {
			fmt.Printf("❌ 搜索失败: %v\n", err)
			continue
		}

		if len(results) == 0 {
			fmt.Println("⚠️ 未找到相关结果")
			continue
		}

		for i, res := range results {
			fmt.Printf("[%d] ID: %s\n内容: %s\n来源: %s\n", i+1, res.ID, res.Content, res.Source)
		}
	}
}

func indexProjectDocs(db *gorm.DB, kb *rag.PostgresKnowledgeBase, svc *ai.AIServiceImpl, modelID uint) {
	indexer := rag.NewIndexer(kb, svc, modelID)
	ctx := context.Background()

	// 1. 索引核心文档目录
	docsDir := "../../docs/zh-CN/core"
	fmt.Printf("正在索引目录: %s ...\n", docsDir)
	err := indexer.IndexDirectory(ctx, docsDir, "doc", []string{".md"})
	if err != nil {
		fmt.Printf("⚠️ 索引目录 %s 失败: %v\n", docsDir, err)
	}

	// 2. 索引组件文档
	compDir := "../../docs/zh-CN/components"
	fmt.Printf("正在索引目录: %s ...\n", compDir)
	err = indexer.IndexDirectory(ctx, compDir, "doc", []string{".md"})
	if err != nil {
		fmt.Printf("⚠️ 索引目录 %s 失败: %v\n", compDir, err)
	}

	fmt.Println("✅ 真实文档索引完成")
}

func setupTestData(db *gorm.DB, kb *rag.PostgresKnowledgeBase, es rag.EmbeddingService) {
	// 1. 初始化表结构和扩展
	if err := kb.Setup(); err != nil {
		fmt.Printf("⚠️ 初始化知识库失败: %v\n", err)
	}

	// 2. 检查是否已有数据
	var count int64
	db.Model(&rag.KnowledgeChunk{}).Count(&count)
	if count > 0 {
		fmt.Printf("当前知识库切片数量: %d\n", count)
		return
	}

	fmt.Println("正在初始化测试数据...")

	// 3. 创建测试文档
	doc := rag.KnowledgeDoc{
		Title:  "BotNexus 深度技术文档",
		Source: "technical_manual_v1.md",
		Type:   "doc",
		Content: `BotNexus 系统配置与故障排查：
- 数据库连接：如果遇到数据库无法连接，请检查 config.json 中的 pg_host (默认 localhost), pg_port (5432), pg_user 和 pg_password。确保 PostgreSQL 服务已启动并允许远程连接。
- 模型端点：确保 base_url 包含 /api/v3 后缀（针对火山引擎）。

AI 任务与技能开发进阶：
- 天气技能配置：WeatherSkill 需要在配置文件中提供 weather_api_key。它接受 city 和 date 参数，支持查询未来 3 天的天气预报。
- 错误处理机制：在 Execute 方法中返回 error 时，系统会捕获该异常并将其包装成友好的提示返回给用户，同时在后台记录详细的错误日志以便审计。
- 多模态支持：BotNexus 深度集成多模态能力。通过配置 doubao-embedding-vision 模型，系统可以同时处理文本、图像和视频的向量化检索。
- 安全性：所有自定义技能都在受限的沙箱环境中执行，无法直接访问宿主机文件系统或敏感环境变量，除非在权限清单中明确声明。`,
	}

	if err := db.Create(&doc).Error; err != nil {
		log.Printf("❌ 创建文档失败: %v\n", err)
		return
	}

	// 4. 创建测试切片
	chunks := []rag.KnowledgeChunk{
		{
			DocID:    doc.ID,
			Content:  "数据库连接故障排查：检查 config.json 中的 pg_host, pg_port, user, password。确保 PG 服务运行正常。",
			Metadata: `{"topic": "troubleshooting", "category": "database"}`,
		},
		{
			DocID:    doc.ID,
			Content:  "天气技能 (WeatherSkill) 配置：需要 weather_api_key，支持通过城市名和日期查询天气。",
			Metadata: `{"topic": "skills", "category": "weather"}`,
		},
		{
			DocID:    doc.ID,
			Content:  "技能执行错误处理：Execute 方法返回的 error 会被系统捕获并记录日志，同时向用户返回友好提示。",
			Metadata: `{"topic": "development", "category": "error_handling"}`,
		},
		{
			DocID:    doc.ID,
			Content:  "多模态能力：系统支持处理图片和视频，使用 doubao-embedding-vision 模型进行跨模态检索。",
			Metadata: `{"topic": "features", "category": "multimodal"}`,
		},
		{
			DocID:    doc.ID,
			Content:  "安全性与沙箱：技能在受限环境下运行，权限受清单控制，保护宿主机安全。",
			Metadata: `{"topic": "security", "category": "sandbox"}`,
		},
	}

	for i := range chunks {
		// 生成向量 (如果向量服务可用)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		embedding, err := es.GenerateEmbedding(ctx, chunks[i].Content)
		cancel()

		if err == nil {
			chunks[i].Embedding = embedding
		} else {
			fmt.Printf("⚠️ 无法为切片 %d 生成向量: %v\n", i, err)
		}

		if err := db.Create(&chunks[i]).Error; err != nil {
			log.Printf("❌ 创建切片 %d 失败: %v\n", i, err)
		}
	}
	fmt.Printf("✅ 成功创建 %d 条测试数据\n", len(chunks))
}
