package rag

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SyncSystemKnowledge 将系统文档和设计文档同步到 RAG 知识库
// baseDir 是 docs 目录所在的根路径，例如 ".." 或 "../.."
func SyncSystemKnowledge(ctx context.Context, indexer *Indexer, baseDir string) {
	if indexer == nil {
		return
	}

	log.Println("[RAG] 开始同步系统设计文档...")

	// 1. 同步核心设计文档 (刚生成的 RAG 方案)
	designDocPath := filepath.Join(baseDir, "docs", "zh-CN", "core", "RAG_USER_DOCS_PLAN.md")
	if _, err := os.Stat(designDocPath); err == nil {
		if err := indexer.IndexFile(ctx, designDocPath, "design"); err != nil {
			log.Printf("[RAG] 索引设计文档失败: %s, err: %v", designDocPath, err)
		} else {
			log.Printf("[RAG] 已成功索引设计文档: %s", designDocPath)
		}
	}

	// 2. 同步项目通用文档目录
	docsDir := filepath.Join(baseDir, "docs", "zh-CN")
	if _, err := os.Stat(docsDir); err == nil {
		extensions := []string{".md"}
		if err := indexer.IndexDirectory(ctx, docsDir, "system", extensions); err != nil {
			log.Printf("[RAG] 索引系统文档目录失败: %s, err: %v", docsDir, err)
		} else {
			log.Printf("[RAG] 系统文档目录同步完成: %s", docsDir)
		}
	}

	log.Println("[RAG] 系统知识同步任务已完成")
}

// StartAutoSync 启动定时同步任务
func StartAutoSync(ctx context.Context, indexer *Indexer, baseDir string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即执行一次
	SyncSystemKnowledge(ctx, indexer, baseDir)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			SyncSystemKnowledge(ctx, indexer, baseDir)
		}
	}
}
