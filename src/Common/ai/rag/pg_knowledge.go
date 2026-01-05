package rag

import (
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// KnowledgeBase 知识库核心接口
type KnowledgeBase = types.KnowledgeBase

// AIService 定义 RAG 需要的 AI 能力接口
type AIService = types.AIService

// EmbeddingService 向量生成服务接口
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateQueryEmbedding(ctx context.Context, query string) ([]float32, error)
}

// PostgresKnowledgeBase 基于 PostgreSQL + pgvector 的知识库实现
type PostgresKnowledgeBase struct {
	db               *gorm.DB
	embeddingService EmbeddingService
	aiSvc            types.AIService // 用于 Query Refinement 等高级功能
	aiModelID        uint            // 用于 AI 服务的模型 ID
}

func NewPostgresKnowledgeBase(db *gorm.DB, es EmbeddingService, aiSvc types.AIService, aiModelID uint) *PostgresKnowledgeBase {
	return &PostgresKnowledgeBase{
		db:               db,
		embeddingService: es,
		aiSvc:            aiSvc,
		aiModelID:        aiModelID,
	}
}

// Search 执行语义搜索
func (p *PostgresKnowledgeBase) Search(ctx context.Context, query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error) {
	return p.SearchHybrid(ctx, query, limit, filter)
}

// SearchHybrid 执行混合搜索 (向量 + 关键词)
func (p *PostgresKnowledgeBase) SearchHybrid(ctx context.Context, query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error) {
	if p.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not set")
	}

	// --- RAG 2.0: Query Refinement (查询重写) ---
	refinedQuery := query
	if p.aiSvc != nil {
		// 尝试优化查询词，提取核心概念，去除语气词，补充背景
		// 这是一个简单的 Agentic 行为示例
		prompt := fmt.Sprintf("你是一个搜索专家。请将以下用户提问改写为一个或多个最适合在知识库中进行向量搜索和关键词搜索的关键词或短语。只需要返回改写后的内容，不要有任何解释。\n提问：%s\n改写结果：", query)
		resp, err := p.aiSvc.Chat(ctx, p.aiModelID, []types.Message{
			{Role: types.RoleUser, Content: prompt},
		}, nil)
		if err == nil && len(resp.Choices) > 0 {
			if s, ok := resp.Choices[0].Message.Content.(string); ok && s != "" {
				refinedQuery = s
				fmt.Printf("[RAG 2.0] Query Refined: '%s' -> '%s'\n", query, refinedQuery)
			}
		}
	}

	var err error

	// 1. 检查是否支持向量搜索 (pgvector 扩展是否存在)
	hasVector := false
	if p.db.Dialector.Name() == "postgres" {
		p.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVector)
	}

	// 2. 构造基础查询 (使用 DISTINCT 避免多所有者共享导致的切片重复)
	baseQuery := p.db.WithContext(ctx).
		Table("knowledge_chunks").
		Select("DISTINCT knowledge_chunks.*").
		Joins("JOIN knowledge_docs ON knowledge_chunks.doc_id = knowledge_docs.id").
		Joins("JOIN knowledge_doc_access ON knowledge_docs.id = knowledge_doc_access.doc_id").
		Where("knowledge_docs.deleted_at IS NULL")

	// 3. 应用过滤条件
	if filter != nil {
		if filter.Status != "" {
			baseQuery = baseQuery.Where("knowledge_docs.status = ?", filter.Status)
		} else {
			baseQuery = baseQuery.Where("knowledge_docs.status = ?", "active")
		}

		// 权限过滤逻辑：系统全局知识 + 目标对象知识 (用户或群组) + 机器人自有知识
		ownerCond := p.db.Where("knowledge_doc_access.owner_type = ?", "system")
		if filter.OwnerType == "user" && filter.OwnerID != "" {
			ownerCond = ownerCond.Or("knowledge_doc_access.owner_type = 'user' AND knowledge_doc_access.owner_id = ?", filter.OwnerID)
		} else if filter.OwnerType == "group" && filter.OwnerID != "" {
			ownerCond = ownerCond.Or("knowledge_doc_access.owner_type = 'group' AND knowledge_doc_access.owner_id = ?", filter.OwnerID)
		}

		// 增加机器人维度的自动共享：该机器人名下的所有知识
		if filter.BotID != "" {
			ownerCond = ownerCond.Or("knowledge_doc_access.owner_type = 'bot' AND knowledge_doc_access.owner_id = ?", filter.BotID)
		}

		// 增加 机器人+用户 联合维度的共享 (用于官方机器人场景下的个人跨群共享)
		if filter.BotID != "" && filter.OwnerType == "user" && filter.OwnerID != "" {
			botUserID := fmt.Sprintf("%s:%s", filter.BotID, filter.OwnerID)
			ownerCond = ownerCond.Or("knowledge_doc_access.owner_type = 'bot_user' AND knowledge_doc_access.owner_id = ?", botUserID)
		}

		baseQuery = baseQuery.Where(ownerCond)
	} else {
		// 默认只搜索 active 的系统知识
		baseQuery = baseQuery.Where("knowledge_docs.status = ? AND knowledge_doc_access.owner_type = ?", "active", "system")
	}

	// 4. 执行搜索
	// --- 向量搜索 (极致优化：优先搜索子切片但返回父上下文) ---
	var vectorChunks []KnowledgeChunk
	if hasVector {
		var vector []float32
		vector, err = p.embeddingService.GenerateQueryEmbedding(ctx, refinedQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %v", err)
		}
		vectorStr, _ := json.Marshal(vector)

		// 搜索时优先搜索 small 类型切片 (通常更精准)，如果是 parent 类型则作为背景
		err = baseQuery.Session(&gorm.Session{}).
			Preload("Doc").
			Order(fmt.Sprintf("embedding <=> '%s'", string(vectorStr))).
			Limit(limit).
			Find(&vectorChunks).Error
		if err != nil {
			fmt.Printf("Vector search failed, falling back to keyword: %v\n", err)
		}
	}

	// --- 关键词搜索 (极致优化：使用 PostgreSQL 全文检索提高精度) ---
	var keywordChunks []KnowledgeChunk
	cleanQuery := refinedQuery
	for _, punc := range []string{"？", "?", "！", "!", "。", ".", "，", ","} {
		cleanQuery = strings.ReplaceAll(cleanQuery, punc, " ")
	}
	words := strings.Fields(cleanQuery)

	if len(words) > 0 {
		kwQuery := baseQuery.Session(&gorm.Session{}).Preload("Doc")
		if p.db.Dialector.Name() == "postgres" {
			// 使用 to_tsquery 进行全文检索
			tsQuery := strings.Join(words, " & ")
			err = kwQuery.Where("to_tsvector('simple', knowledge_chunks.content) @@ to_tsquery('simple', ?)", tsQuery).
				Limit(limit).
				Find(&keywordChunks).Error
		}

		if err != nil || len(keywordChunks) == 0 {
			// 降级到 LIKE 搜索
			queryBuilder := ""
			args := make([]interface{}, 0)
			for i, word := range words {
				if i > 0 {
					queryBuilder += " OR "
				}
				queryBuilder += "knowledge_chunks.content LIKE ?"
				args = append(args, "%"+word+"%")
			}
			err = kwQuery.Where(queryBuilder, args...).Limit(limit).Find(&keywordChunks).Error
		}
	}

	fmt.Printf("Search Results - Vector: %d, Keyword: %d\n", len(vectorChunks), len(keywordChunks))

	// 2. 合并结果并去重，同时处理多级索引逻辑 (Parent-Child)
	seen := make(map[uint]bool)
	var intermediateChunks []KnowledgeChunk

	// 先添加关键词匹配的
	for _, c := range keywordChunks {
		if !seen[c.ID] {
			intermediateChunks = append(intermediateChunks, c)
			seen[c.ID] = true
		}
	}

	// 再添加向量匹配的
	for _, c := range vectorChunks {
		if !seen[c.ID] {
			intermediateChunks = append(intermediateChunks, c)
			seen[c.ID] = true
		}
	}

	// --- 极致优化：多级索引回溯 (如果搜到的是子切片，获取其父切片作为补充上下文) ---
	var finalResults []types.DocChunk
	for _, c := range intermediateChunks {
		content := c.Content
		if c.Type == "small" && c.ParentID != 0 {
			var parent KnowledgeChunk
			if err := p.db.First(&parent, c.ParentID).Error; err == nil {
				// 将父切片内容作为 Context 补充，或者直接替换为父切片内容
				// 这里采用“子切片内容在前，父切片背景在后”的策略
				content = fmt.Sprintf("%s\n\n[背景补充]\n%s", c.Content, parent.Content)
			}
		}

		finalResults = append(finalResults, types.DocChunk{
			ID:      fmt.Sprintf("chunk_%d", c.ID),
			Content: content,
			Source:  c.Doc.Source,
		})
	}

	// 截断到 limit
	if len(finalResults) > limit {
		finalResults = finalResults[:limit]
	}

	results := finalResults

	// --- RAG 2.0: GraphRAG Retrieval ---
	entities, relations, _ := p.SearchGraph(ctx, refinedQuery, 3)
	if len(entities) > 0 {
		var graphContext strings.Builder
		graphContext.WriteString("[知识图谱关联上下文]\n")
		graphContext.WriteString("相关实体：\n")
		for _, e := range entities {
			graphContext.WriteString(fmt.Sprintf("- %s (%s): %s\n", e.Name, e.Type, e.Description))
		}
		if len(relations) > 0 {
			graphContext.WriteString("实体关系：\n")
			for _, r := range relations {
				graphContext.WriteString(fmt.Sprintf("- [%s] --(%s)--> [%s]: %s\n", r.Subject.Name, r.Predicate, r.Object.Name, r.Description))
			}
		}
		results = append(results, types.DocChunk{
			ID:      "graph_context",
			Content: graphContext.String(),
			Source:  "Knowledge Graph",
		})
	}

	// --- RAG 2.0: Self-Reflection (检索相关性自检) ---
	// 使用并发处理提高自省效率，确保检索质量的同时不牺牲响应速度
	if p.aiSvc != nil && len(results) > 0 {
		fmt.Printf("[RAG 2.0] Starting Parallel Self-Reflection for %d chunks...\n", len(results))

		type reflectionResult struct {
			index int
			chunk types.DocChunk
			keep  bool
		}

		resultChan := make(chan reflectionResult, len(results))
		var wg sync.WaitGroup

		for i, res := range results {
			wg.Add(1)
			go func(idx int, chunk types.DocChunk) {
				defer wg.Done()

				reflectionPrompt := fmt.Sprintf("你是一个质量控制专家。请判断以下检索到的内容是否能够回答用户的提问，或者与提问高度相关。\n用户提问：%s\n检索内容：%s\n如果相关，请只输出 YES，如果不相关，请只输出 NO。如果内容包含具体的步骤或配置，且提问涉及“如何做”，请务必输出 YES。不要输出任何其他文字。", query, chunk.Content)

				resp, err := p.aiSvc.Chat(ctx, p.aiModelID, []types.Message{
					{Role: types.RoleUser, Content: reflectionPrompt},
				}, nil)

				keep := true // 默认保留（保守策略）
				if err == nil && len(resp.Choices) > 0 {
					if s, ok := resp.Choices[0].Message.Content.(string); ok {
						ans := strings.TrimSpace(strings.ToUpper(s))
						if !strings.Contains(ans, "YES") {
							fmt.Printf("[RAG 2.0] Self-Reflection: Filtered out chunk from %s (AI said %s)\n", chunk.Source, ans)
							keep = false
						}
					}
				}
				resultChan <- reflectionResult{index: idx, chunk: chunk, keep: keep}
			}(i, res)
		}

		// 异步关闭通道
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// 收集并恢复顺序
		reflectionMap := make(map[int]reflectionResult)
		for res := range resultChan {
			reflectionMap[res.index] = res
		}

		var filteredResults []types.DocChunk
		for i := 0; i < len(results); i++ {
			if res, ok := reflectionMap[i]; ok && res.keep {
				filteredResults = append(filteredResults, res.chunk)
			}
		}
		return filteredResults, nil
	}

	return results, nil
}

// SetDocStatus 设置文档状态 (启用/暂停)
func (p *PostgresKnowledgeBase) SetDocStatus(ctx context.Context, docID uint, status string) error {
	return p.db.WithContext(ctx).Model(&KnowledgeDoc{}).Where("id = ?", docID).Update("status", status).Error
}

// DeleteDoc 删除文档及其所有切片和权限记录
func (p *PostgresKnowledgeBase) DeleteDoc(ctx context.Context, docID uint) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除权限记录
		if err := tx.Where("doc_id = ?", docID).Delete(&KnowledgeDocAccess{}).Error; err != nil {
			return err
		}
		// 删除切片
		if err := tx.Where("doc_id = ?", docID).Delete(&KnowledgeChunk{}).Error; err != nil {
			return err
		}
		// 删除文档
		if err := tx.Delete(&KnowledgeDoc{}, docID).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetUserDocs 获取用户或群组可见的文档列表 (基于授权关系)
func (p *PostgresKnowledgeBase) GetUserDocs(ctx context.Context, ownerType, ownerID string) ([]KnowledgeDoc, error) {
	var docs []KnowledgeDoc
	err := p.db.WithContext(ctx).
		Table("knowledge_docs").
		Select("DISTINCT knowledge_docs.*").
		Joins("JOIN knowledge_doc_access ON knowledge_docs.id = knowledge_doc_access.doc_id").
		Where("knowledge_doc_access.owner_type = ? AND knowledge_doc_access.owner_id = ?", ownerType, ownerID).
		Find(&docs).Error
	return docs, err
}

// GetOwnedDocs 获取用户上传的原始文档列表 (拥有所有权)
func (p *PostgresKnowledgeBase) GetOwnedDocs(ctx context.Context, uploaderID string) ([]KnowledgeDoc, error) {
	var docs []KnowledgeDoc
	err := p.db.WithContext(ctx).Where("uploader_id = ?", uploaderID).Find(&docs).Error
	return docs, err
}

// IsDocOwner 检查用户是否为文档的所有者
func (p *PostgresKnowledgeBase) IsDocOwner(ctx context.Context, docID uint, userID string) bool {
	var count int64
	p.db.WithContext(ctx).Model(&KnowledgeDoc{}).
		Where("id = ? AND uploader_id = ?", docID, userID).
		Count(&count)
	return count > 0
}

// AddDocAccess 为文档添加访问权限 (实现共享)
func (p *PostgresKnowledgeBase) AddDocAccess(ctx context.Context, docID uint, ownerType, ownerID string) error {
	var count int64
	p.db.WithContext(ctx).Model(&KnowledgeDocAccess{}).
		Where("doc_id = ? AND owner_type = ? AND owner_id = ?", docID, ownerType, ownerID).
		Count(&count)

	if count > 0 {
		return nil // 已存在
	}

	access := KnowledgeDocAccess{
		DocID:     docID,
		OwnerType: ownerType,
		OwnerID:   ownerID,
	}
	return p.db.WithContext(ctx).Create(&access).Error
}

// RemoveDocAccess 移除文档的访问权限
func (p *PostgresKnowledgeBase) RemoveDocAccess(ctx context.Context, docID uint, ownerType, ownerID string) error {
	return p.db.WithContext(ctx).
		Where("doc_id = ? AND owner_type = ? AND owner_id = ?", docID, ownerType, ownerID).
		Delete(&KnowledgeDocAccess{}).Error
}

// SaveEntity 保存或更新实体，并生成向量
func (p *PostgresKnowledgeBase) SaveEntity(ctx context.Context, entity *KnowledgeEntity) error {
	// 如果已经存在同名同类型的实体，则尝试合并
	var existing KnowledgeEntity
	if err := p.db.WithContext(ctx).Model(&KnowledgeEntity{}).Where("name = ? AND type = ?", entity.Name, entity.Type).First(&existing).Error; err == nil {
		// 更新描述（简单追加或由 AI 以后期合并）
		if !strings.Contains(existing.Description, entity.Description) {
			existing.Description += "\n" + entity.Description
		}
		// 重新生成向量
		if p.embeddingService != nil {
			embedding, _ := p.embeddingService.GenerateEmbedding(ctx, existing.Name+": "+existing.Description)
			existing.Embedding = embedding
		}
		return p.db.WithContext(ctx).Model(&KnowledgeEntity{}).Save(&existing).Error
	}

	// 新建
	if p.embeddingService != nil {
		embedding, _ := p.embeddingService.GenerateEmbedding(ctx, entity.Name+": "+entity.Description)
		entity.Embedding = embedding
	}
	return p.db.WithContext(ctx).Model(&KnowledgeEntity{}).Create(entity).Error
}

// SaveRelation 保存关系
func (p *PostgresKnowledgeBase) SaveRelation(ctx context.Context, relation *KnowledgeRelation) error {
	return p.db.WithContext(ctx).Model(&KnowledgeRelation{}).Create(relation).Error
}

// SearchGraph 搜索关联的实体和关系
func (p *PostgresKnowledgeBase) SearchGraph(ctx context.Context, query string, limit int) ([]KnowledgeEntity, []KnowledgeRelation, error) {
	var entities []KnowledgeEntity
	var relations []KnowledgeRelation

	// 1. 向量搜索相关实体
	if p.embeddingService != nil && p.db.Dialector.Name() == "postgres" {
		// 检查是否有 vector 扩展
		hasVector := false
		p.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVector)
		if hasVector {
			vector, err := p.embeddingService.GenerateQueryEmbedding(ctx, query)
			if err == nil {
				vectorStr, _ := json.Marshal(vector)
				p.db.WithContext(ctx).
					Model(&KnowledgeEntity{}).
					Order(fmt.Sprintf("embedding <=> '%s'", string(vectorStr))).
					Limit(limit).
					Find(&entities)
			}
		}
	}

	// 2. 如果向量搜索没结果，尝试关键词匹配实体名
	if len(entities) == 0 {
		p.db.WithContext(ctx).
			Model(&KnowledgeEntity{}).
			Where("name LIKE ?", "%"+query+"%").
			Limit(limit).
			Find(&entities)
	}

	// 3. 获取这些实体的直接关系
	if len(entities) > 0 {
		var entityIDs []uint
		for _, e := range entities {
			entityIDs = append(entityIDs, e.ID)
		}
		p.db.WithContext(ctx).
			Model(&KnowledgeRelation{}).
			Preload("Subject").
			Preload("Object").
			Where("subject_id IN ? OR object_id IN ?", entityIDs, entityIDs).
			Limit(limit * 2).
			Find(&relations)
	}

	return entities, relations, nil
}

// Setup 初始化数据库表和扩展
func (p *PostgresKnowledgeBase) Setup() error {
	// 仅在 PostgreSQL 下尝试创建扩展
	if p.db.Dialector.Name() == "postgres" {
		if err := p.db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			fmt.Printf("Warning: failed to create vector extension: %v. Vector search will be disabled.\n", err)
			// 不返回错误，允许继续初始化，后续搜索会自动退化到关键词
		}
	}
	return p.db.AutoMigrate(&KnowledgeDoc{}, &KnowledgeDocAccess{}, &KnowledgeChunk{}, &KnowledgeEntity{}, &KnowledgeRelation{})
}
