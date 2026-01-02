package rag

import (
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// EmbeddingService 向量生成服务接口
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// PostgresKnowledgeBase 基于 PostgreSQL + pgvector 的知识库实现
type PostgresKnowledgeBase struct {
	db               *gorm.DB
	embeddingService EmbeddingService
}

func NewPostgresKnowledgeBase(db *gorm.DB, es EmbeddingService) *PostgresKnowledgeBase {
	return &PostgresKnowledgeBase{
		db:               db,
		embeddingService: es,
	}
}

// Search 执行语义搜索
func (p *PostgresKnowledgeBase) Search(ctx context.Context, query string, limit int) ([]tasks.DocChunk, error) {
	if p.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not set")
	}

	// 1. 生成查询文本的向量
	vector, err := p.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	// 2. 将 float32 数组转换为 PostgreSQL 向量格式 [v1,v2,...]
	vectorStr, _ := json.Marshal(vector)

	// 3. 使用 pgvector 执行余弦相似度搜索 (<=> 是余弦距离)
	var chunks []KnowledgeChunk
	err = p.db.WithContext(ctx).
		Preload("Doc").
		Order(fmt.Sprintf("embedding <=> '%s'", string(vectorStr))).
		Limit(limit).
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %v", err)
	}

	// 4. 转换为任务系统通用的 DocChunk 格式
	var results []tasks.DocChunk
	for _, c := range chunks {
		results = append(results, tasks.DocChunk{
			ID:      fmt.Sprintf("chunk_%d", c.ID),
			Content: c.Content,
			Source:  c.Doc.Source,
		})
	}

	return results, nil
}

// Setup 初始化数据库表和扩展
func (p *PostgresKnowledgeBase) Setup() error {
	// 仅在 PostgreSQL 下尝试创建扩展
	if p.db.Dialector.Name() == "postgres" {
		if err := p.db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			return fmt.Errorf("failed to create vector extension: %v", err)
		}
	}
	return p.db.AutoMigrate(&KnowledgeDoc{}, &KnowledgeChunk{})
}
