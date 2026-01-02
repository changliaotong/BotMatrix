package rag

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Indexer 知识索引器
type Indexer struct {
	kb *PostgresKnowledgeBase
}

func NewIndexer(kb *PostgresKnowledgeBase) *Indexer {
	return &Indexer{kb: kb}
}

// IndexFile 索引单个文件
func (idx *Indexer) IndexFile(ctx context.Context, path string, docType string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", path, err)
	}

	ext := filepath.Ext(path)
	var chunks []Chunk

	if ext == ".md" {
		chunks = SimpleMarkdownChunker(string(content), 50)
	} else if ext == ".go" {
		chunks = SimpleCodeChunker(string(content))
	} else {
		chunks = []Chunk{{Content: string(content), Title: filepath.Base(path)}}
	}

	// 1. 创建文档记录
	doc := KnowledgeDoc{
		Title:   filepath.Base(path),
		Source:  path,
		Type:    docType,
		Content: string(content),
	}

	// 检查文档是否已存在，如果存在则更新，不存在则创建
	var existingDoc KnowledgeDoc
	if err := idx.kb.db.Where("source = ?", path).First(&existingDoc).Error; err == nil {
		doc.ID = existingDoc.ID
		idx.kb.db.Save(&doc)
		// 删除旧的切片
		idx.kb.db.Where("doc_id = ?", doc.ID).Delete(&KnowledgeChunk{})
	} else {
		idx.kb.db.Create(&doc)
	}

	// 2. 处理切片并生成向量
	for _, c := range chunks {
		embedding, err := idx.kb.embeddingService.GenerateEmbedding(ctx, c.Content)
		if err != nil {
			log.Printf("[Indexer] Failed to generate embedding for chunk in %s: %v", path, err)
			continue
		}

		chunk := KnowledgeChunk{
			DocID:     doc.ID,
			Content:   c.Content,
			Embedding: embedding,
			Metadata:  fmt.Sprintf(`{"title": "%s"}`, c.Title),
		}
		idx.kb.db.Create(&chunk)
	}

	log.Printf("[Indexer] Indexed %s (%d chunks)", path, len(chunks))
	return nil
}

// IndexDirectory 递归索引目录
func (idx *Indexer) IndexDirectory(ctx context.Context, dir string, docType string, extensions []string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 检查扩展名
		match := false
		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range extensions {
			if ext == strings.ToLower(e) {
				match = true
				break
			}
		}

		if match {
			return idx.IndexFile(ctx, path, docType)
		}
		return nil
	})
}
