package app

import (
	"context"
	"time"

	clog "BotMatrix/common/log"
	"BotMatrix/common/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CognitiveMemoryService 认知记忆服务接口
type CognitiveMemoryService interface {
	// GetRelevantMemories 获取与当前上下文相关的记忆
	GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemoryGORM, error)
	// SaveMemory 保存或更新记忆
	SaveMemory(ctx context.Context, memory *models.CognitiveMemoryGORM) error
	// ForgetMemory 删除记忆
	ForgetMemory(ctx context.Context, memoryID uint) error
}

// CognitiveMemoryServiceImpl 认知记忆服务实现
type CognitiveMemoryServiceImpl struct {
	db *gorm.DB
}

// NewCognitiveMemoryService 创建新的认知记忆服务
func NewCognitiveMemoryService(db *gorm.DB) CognitiveMemoryService {
	return &CognitiveMemoryServiceImpl{
		db: db,
	}
}

func (s *CognitiveMemoryServiceImpl) GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemoryGORM, error) {
	var memories []models.CognitiveMemoryGORM

	// 1. 基础过滤：UserID + BotID
	queryBuilder := s.db.WithContext(ctx).Where("user_id = ? AND bot_id = ?", userID, botID)

	// 2. 语义搜索 (Vector Search Placeholder)
	// 如果 query 不为空，且系统配置了向量数据库，这里应该执行 Embedding + Vector Search
	if query != "" {
		// 模拟语义检索逻辑
		clog.Debug("[Memory] Performing semantic search for query", zap.String("query", query))
		// queryBuilder = queryBuilder.Where("content LIKE ?", "%"+query+"%") // 临时退化为关键词搜索
	}

	// 3. 按照重要性、最后发现时间、衰减因子排序
	// 简单的排序算法：importance * 10 + recency_score
	err := queryBuilder.
		Order("importance DESC, last_seen DESC").
		Limit(10).
		Find(&memories).Error

	if err == nil {
		clog.Info("[Memory] Retrieved memories",
			zap.String("user_id", userID),
			zap.Int("count", len(memories)))
	}

	return memories, err
}

func (s *CognitiveMemoryServiceImpl) SaveMemory(ctx context.Context, memory *models.CognitiveMemoryGORM) error {
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.LastSeen = time.Now()

	// 如果已有相同 Category 的记忆，可以考虑合并或覆盖
	// 这里简单处理：如果 ID 为 0 则创建，否则更新
	return s.db.WithContext(ctx).Save(memory).Error
}

func (s *CognitiveMemoryServiceImpl) ForgetMemory(ctx context.Context, memoryID uint) error {
	return s.db.WithContext(ctx).Delete(&models.CognitiveMemoryGORM{}, memoryID).Error
}
