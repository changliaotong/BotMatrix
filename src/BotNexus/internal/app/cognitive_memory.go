package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"BotMatrix/common/ai"
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
	// ConsolidateMemories 合并冗余或相似的记忆 (需要传入 AI 服务进行总结)
	ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc AIIntegrationService) error
	// LearnFromURL 从 URL 自动学习并存入记忆
	LearnFromURL(ctx context.Context, botID string, url string, category string) error
	// LearnFromContent 从原始内容自动学习并存入记忆
	LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error
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

func (s *CognitiveMemoryServiceImpl) ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc AIIntegrationService) error {
	if aiSvc == nil {
		return fmt.Errorf("AI service is required for consolidation")
	}

	// 获取该用户和机器人的所有记忆
	var memories []models.CognitiveMemoryGORM
	query := s.db.WithContext(ctx).Where("bot_id = ?", botID)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id = '' OR user_id IS NULL")
	}

	err := query.Order("category, created_at ASC").Find(&memories).Error
	if err != nil {
		return err
	}

	if len(memories) < 10 {
		// 记忆太少，不需要合并
		clog.Info("[Memory] Too few memories to consolidate", zap.String("bot_id", botID), zap.Int("count", len(memories)))
		return nil
	}

	// 构造固化 Prompt
	prompt := "你是一个记忆管理专家。以下是关于某个数字员工或其与用户交互的碎片化记忆片段。请将这些记忆进行逻辑合并、去重并提炼。\n"
	prompt += "规则：\n1. 合并相似或相关的片段（例如：‘喜欢苹果’和‘喜欢红富士’可以合并为‘喜欢各种苹果’）。\n2. 保持分类清晰。\n3. 提炼出更有深度的洞察，而不仅仅是堆砌事实。\n4. 格式：[类别] 提炼后的内容。\n\n记忆片段：\n"

	for _, m := range memories {
		cat := m.Category
		if cat == "" {
			cat = "general"
		}
		prompt += fmt.Sprintf("- [%s] %s\n", cat, m.Content)
	}

	// 调用 AI 进行固化 (使用默认模型)
	msgs := []ai.Message{
		{Role: ai.RoleSystem, Content: prompt},
	}

	// 尝试获取默认模型 ID (这里假设 AIIntegrationService 能处理默认模型，或者我们传 0)
	resp, err := aiSvc.Chat(ctx, 0, msgs, nil)
	if err != nil {
		return fmt.Errorf("AI chat failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("AI returned no choices")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || strings.TrimSpace(content) == "" {
		return fmt.Errorf("AI returned empty content")
	}

	// 解析新记忆并替换旧记忆
	lines := strings.Split(content, "\n")
	var newMemories []models.CognitiveMemoryGORM
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		category := "general"
		fact := line
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			idx := strings.Index(line, "]")
			category = line[1:idx]
			fact = strings.TrimSpace(line[idx+1:])
		}

		newMemories = append(newMemories, models.CognitiveMemoryGORM{
			UserID:     userID,
			BotID:      botID,
			Category:   category,
			Content:    fact,
			Importance: 3, // 固化后的记忆重要性更高
			LastSeen:   time.Now(),
		})
	}

	if len(newMemories) > 0 {
		return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 删除旧记忆
			delQuery := tx.Where("bot_id = ?", botID)
			if userID != "" {
				delQuery = delQuery.Where("user_id = ?", userID)
			} else {
				delQuery = delQuery.Where("user_id = '' OR user_id IS NULL")
			}

			if err := delQuery.Delete(&models.CognitiveMemoryGORM{}).Error; err != nil {
				return err
			}

			// 保存新记忆 (SaveMemory 会处理 Embedding)
			for _, m := range newMemories {
				// 注意：这里我们直接用 tx.Create，因为我们想在事务中执行。
				// 但 SaveMemory 内部使用了 s.db。
				// 为了保持事务和 Embedding，我们手动处理 Embedding 或重构 SaveMemory。
				// 这里简单处理：先 Create，后续由后台任务补充 Embedding，或者直接在这里调 Embedding。
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}

	return nil
}

func (s *CognitiveMemoryServiceImpl) LearnFromURL(ctx context.Context, botID string, url string, category string) error {
	clog.Info("[Memory] Learning from URL", zap.String("bot_id", botID), zap.String("url", url))

	// 1. 获取内容 (简单实现，后续可扩展为更强大的爬虫)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("URL returned status: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	// 2. 提取文件名或判断类型
	filename := filepath.Base(url)
	if !strings.Contains(filename, ".") {
		// 如果 URL 没后缀，根据 Content-Type 猜测
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			filename = "index.html"
		} else if strings.Contains(contentType, "application/pdf") {
			filename = "doc.pdf"
		} else {
			filename = "content.txt"
		}
	}

	return s.LearnFromContent(ctx, botID, content, filename, category)
}

func (s *CognitiveMemoryServiceImpl) LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error {
	clog.Info("[Memory] Learning from content", zap.String("bot_id", botID), zap.String("filename", filename))

	// 1. 选择解析器
	var parser rag.ContentParser
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		parser = &rag.PDFParser{}
	case ".xlsx", ".xls":
		parser = &rag.ExcelParser{}
	case ".docx":
		parser = &rag.DocxParser{}
	case ".md", ".markdown":
		parser = &rag.MarkdownParser{MinSize: 50}
	case ".go", ".py", ".js", ".ts", ".java", ".c", ".cpp":
		parser = &rag.CodeParser{}
	case ".html", ".htm":
		// 简单的 HTML 处理：去除标签
		re := regexp.MustCompile("<[^>]*>")
		stripped := re.ReplaceAllString(string(content), "")
		content = []byte(stripped)
		parser = &rag.TxtParser{MinSize: 50}
	default:
		parser = &rag.TxtParser{MinSize: 50}
	}

	// 2. 解析
	chunks := parser.Parse(ctx, content)
	if len(chunks) == 0 {
		return fmt.Errorf("no content extracted from %s", filename)
	}

	// 3. 存储
	for i, chunk := range chunks {
		mem := &models.CognitiveMemoryGORM{
			BotID:      botID,
			UserID:     "", // 角色记忆
			Content:    chunk.Content,
			Category:   category,
			Importance: 3, // 默认中等重要性
			Metadata:   fmt.Sprintf("Source: %s, Part: %d", filename, i+1),
		}
		if chunk.Title != "" {
			mem.Metadata += ", Title: " + chunk.Title
		}

		if err := s.SaveMemory(ctx, mem); err != nil {
			clog.Error("[Memory] Failed to save learned chunk", zap.Error(err), zap.Int("index", i))
		}
	}

	clog.Info("[Memory] Learning complete", zap.String("bot_id", botID), zap.Int("chunks", len(chunks)))
	return nil
}
