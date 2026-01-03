package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotNexus/internal/rag"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CognitiveMemoryService 认知记忆服务接口
type CognitiveMemoryService interface {
	// GetRelevantMemories 获取与当前上下文相关的记忆（包括用户记忆和数字员工角色记忆）
	GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemoryGORM, error)
	// GetRoleMemories 获取数字员工的角色定义/通用知识记忆
	GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemoryGORM, error)
	// SaveMemory 保存或更新记忆
	SaveMemory(ctx context.Context, memory *models.CognitiveMemoryGORM) error
	// ForgetMemory 删除记忆
	ForgetMemory(ctx context.Context, memoryID uint) error
	// SearchMemories 搜索特定记忆（支持关键词）
	SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemoryGORM, error)
	// SetEmbeddingService 设置向量服务
	SetEmbeddingService(svc rag.EmbeddingService)
}

// CognitiveMemoryServiceImpl 认知记忆服务实现
type CognitiveMemoryServiceImpl struct {
	db           *gorm.DB
	embeddingSvc rag.EmbeddingService
}

// NewCognitiveMemoryService 创建新的认知记忆服务
func NewCognitiveMemoryService(db *gorm.DB) CognitiveMemoryService {
	return &CognitiveMemoryServiceImpl{
		db: db,
	}
}

func (s *CognitiveMemoryServiceImpl) SetEmbeddingService(svc rag.EmbeddingService) {
	s.embeddingSvc = svc
}

func (s *CognitiveMemoryServiceImpl) GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemoryGORM, error) {
	var userMemories []models.CognitiveMemoryGORM
	var roleMemories []models.CognitiveMemoryGORM

	// 1. 获取用户特定记忆 (UserID + BotID)
	// 优先尝试向量检索 (如果 query 不为空且 embeddingSvc 可用)
	hasVector := false
	if query != "" && s.embeddingSvc != nil && s.db.Dialector.Name() == "postgres" {
		s.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVector)
	}

	if hasVector {
		vector, err := s.embeddingSvc.GenerateQueryEmbedding(ctx, query)
		if err == nil {
			vectorStr, _ := json.Marshal(vector)
			// 执行向量搜索
			err = s.db.WithContext(ctx).
				Where("user_id = ? AND bot_id = ?", userID, botID).
				Order(fmt.Sprintf("embedding <=> '%s'", string(vectorStr))).
				Limit(10).
				Find(&userMemories).Error
			if err != nil {
				clog.Warn("[Memory] Vector search failed, falling back to keyword", zap.Error(err))
			}
		} else {
			clog.Warn("[Memory] Failed to generate embedding for query", zap.Error(err))
		}
	}

	// 如果向量检索未命中或失败，使用关键词检索兜底
	if len(userMemories) == 0 {
		userQuery := s.db.WithContext(ctx).Where("user_id = ? AND bot_id = ?", userID, botID)
		if query != "" {
			userQuery = userQuery.Where("content LIKE ?", "%"+query+"%")
		}
		err := userQuery.Order("importance DESC, last_seen DESC").Limit(10).Find(&userMemories).Error
		if err != nil {
			clog.Error("[Memory] Failed to get user memories", zap.Error(err))
		}
	}

	// 2. 获取数字员工角色记忆
	roleQuery := s.db.WithContext(ctx).Where("user_id = '' AND bot_id = ?", botID)
	err := roleQuery.Order("importance DESC").Limit(5).Find(&roleMemories).Error
	if err != nil {
		clog.Error("[Memory] Failed to get role memories", zap.Error(err))
	}

	// 3. 合并记忆
	allMemories := append(roleMemories, userMemories...)
	return allMemories, nil
}

func (s *CognitiveMemoryServiceImpl) GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemoryGORM, error) {
	var memories []models.CognitiveMemoryGORM
	err := s.db.WithContext(ctx).
		Where("user_id = '' AND bot_id = ?", botID).
		Order("importance DESC").
		Find(&memories).Error
	return memories, err
}

func (s *CognitiveMemoryServiceImpl) SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemoryGORM, error) {
	var memories []models.CognitiveMemoryGORM
	db := s.db.WithContext(ctx).Where("bot_id = ?", botID)

	if query != "" {
		db = db.Where("content LIKE ?", "%"+query+"%")
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}

	err := db.Order("last_seen DESC").Limit(20).Find(&memories).Error
	return memories, err
}

func (s *CognitiveMemoryServiceImpl) SaveMemory(ctx context.Context, memory *models.CognitiveMemoryGORM) error {
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.LastSeen = time.Now()

	// 如果向量服务可用，生成向量
	if s.embeddingSvc != nil && memory.Content != "" {
		vec, err := s.embeddingSvc.GenerateEmbedding(ctx, memory.Content)
		if err == nil {
			vecJSON, _ := json.Marshal(vec)
			memory.Embedding = string(vecJSON)
		} else {
			clog.Warn("[Memory] Failed to generate embedding for memory", zap.Error(err))
		}
	}

	// 如果已有相同 Category 的记忆，可以考虑合并或覆盖
	// 这里简单处理：如果 ID 为 0 则创建，否则更新
	return s.db.WithContext(ctx).Save(memory).Error
}

func (s *CognitiveMemoryServiceImpl) ForgetMemory(ctx context.Context, memoryID uint) error {
	return s.db.WithContext(ctx).Delete(&models.CognitiveMemoryGORM{}, memoryID).Error
}
