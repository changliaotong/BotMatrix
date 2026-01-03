package rag

import (
	"BotMatrix/common/ai"
	"BotNexus/tasks"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Indexer 负责将文档转换为切片并入库
type Indexer struct {
	kb      *PostgresKnowledgeBase
	parsers map[string]ContentParser
	svc     tasks.AIService
	modelID uint
}

func NewIndexer(kb *PostgresKnowledgeBase, svc tasks.AIService, modelID uint) *Indexer {
	idx := &Indexer{
		kb:      kb,
		parsers: make(map[string]ContentParser),
		svc:     svc,
		modelID: modelID,
	}

	aiParser := AIParser{svc: svc, modelID: modelID}

	// 注册默认解析器
	idx.RegisterParser(".md", &MarkdownParser{MinSize: 50})
	idx.RegisterParser(".go", &CodeParser{})
	idx.RegisterParser(".docx", &DocxParser{})
	idx.RegisterParser(".doc", &DocParser{})
	idx.RegisterParser(".pdf", &PDFParser{AIParser: aiParser})
	idx.RegisterParser(".txt", &TxtParser{MinSize: 20})
	idx.RegisterParser(".xlsx", &ExcelParser{})
	idx.RegisterParser(".xls", &ExcelParser{})

	// 注册图像解析器
	imgParser := &ImageParser{AIParser: aiParser}
	idx.RegisterParser(".png", imgParser)
	idx.RegisterParser(".jpg", imgParser)
	idx.RegisterParser(".jpeg", imgParser)
	idx.RegisterParser(".gif", imgParser)
	idx.RegisterParser(".webp", imgParser)

	idx.RegisterParser("default", &DefaultParser{})
	return idx
}

func (idx *Indexer) RegisterParser(ext string, parser ContentParser) {
	idx.parsers[strings.ToLower(ext)] = parser
}

// IndexFile 索引单个文件 (默认为系统级别)
func (idx *Indexer) IndexFile(ctx context.Context, path string, docType string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", path, err)
	}

	title := filepath.Base(path)
	return idx.IndexContent(ctx, title, path, content, docType, "system", "system", "system")
}

// IndexContent 索引通用内容
func (idx *Indexer) IndexContent(ctx context.Context, title, source string, content []byte, docType, uploaderID, targetType, targetID string) error {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// 1. 检查文档是否已存在且哈希一致
	var existingDoc KnowledgeDoc
	if err := idx.kb.db.Where("source = ?", source).First(&existingDoc).Error; err == nil {
		if existingDoc.Hash == hash {
			// 内容未变，但可能需要确保新的授权关系存在
			if targetType != "" && targetID != "" {
				idx.kb.AddDocAccess(ctx, existingDoc.ID, targetType, targetID)
			}
			log.Printf("[Indexer] Skip unchanged content: %s", source)
			return nil
		}
		// 如果哈希不一致，删除旧的切片，准备重新索引
		idx.kb.db.Where("doc_id = ?", existingDoc.ID).Delete(&KnowledgeChunk{})
		existingDoc.Hash = hash
		existingDoc.UploaderID = uploaderID
		existingDoc.UpdatedAt = time.Now()
		// 注意：Content 字段会在解析后更新
	} else {
		// 不存在则创建
		existingDoc = KnowledgeDoc{
			Title:      title,
			Source:     source,
			Type:       docType,
			Hash:       hash,
			UploaderID: uploaderID,
			Status:     "active",
		}
		if err := idx.kb.db.Create(&existingDoc).Error; err != nil {
			return err
		}
		// 创建初始授权关系
		if targetType != "" && targetID != "" {
			idx.kb.AddDocAccess(ctx, existingDoc.ID, targetType, targetID)
		}
		// 默认也授权给上传者本人，方便在“我的文档”中查看
		if uploaderID != "" {
			idx.kb.AddDocAccess(ctx, existingDoc.ID, "user", uploaderID)
		}
	}

	// 2. 根据类型选择解析器并切分
	var chunks []Chunk
	ext := strings.ToLower(filepath.Ext(source))

	if parser, ok := idx.parsers[ext]; ok {
		chunks = parser.Parse(ctx, content)
	} else if strings.Contains(strings.ToLower(title), "readme") {
		chunks = idx.parsers[".md"].Parse(ctx, content)
	} else {
		chunks = idx.parsers["default"].Parse(ctx, content)
	}

	// 更新文档的纯文本内容 (由所有切片合并而成)
	var fullText strings.Builder
	for _, c := range chunks {
		fullText.WriteString(c.Content)
		fullText.WriteString("\n")
	}
	existingDoc.Content = fullText.String()
	idx.kb.db.Save(&existingDoc)

	// 3. 处理切片并生成向量
	for _, c := range chunks {
		embedding, err := idx.kb.embeddingService.GenerateEmbedding(ctx, c.Content)
		if err != nil {
			log.Printf("[Indexer] Failed to generate embedding for chunk in %s: %v", source, err)
			continue
		}

		chunk := KnowledgeChunk{
			DocID:     existingDoc.ID,
			Content:   c.Content,
			Embedding: embedding,
			Metadata:  fmt.Sprintf(`{"title": "%s", "source": "%s"}`, c.Title, source),
		}
		idx.kb.db.Create(&chunk)
		// --- RAG 2.0: GraphRAG Entity Extraction ---
		go idx.ExtractAndIndexGraph(ctx, c.Content, existingDoc.ID)
	}

	log.Printf("[Indexer] Indexed %s (%d chunks)", source, len(chunks))
	return nil
}

// ExtractAndIndexGraph 从文本中提取实体和关系并入库
func (idx *Indexer) ExtractAndIndexGraph(ctx context.Context, content string, docID uint) {
	if idx.svc == nil {
		return
	}

	prompt := `你是一个知识图谱专家。请从以下文本中提取关键实体及其关系。
输出格式必须为 JSON，结构如下：
{
  "entities": [{"name": "实体名", "type": "类型", "description": "简要描述"}],
  "relations": [{"subject": "主体实体名", "predicate": "关系谓语", "object": "客体实体名", "description": "关系描述"}]
}
文本内容：
` + content

	resp, err := idx.svc.Chat(ctx, idx.modelID, []ai.Message{
		{Role: ai.RoleUser, Content: prompt},
	}, nil)

	if err != nil || len(resp.Choices) == 0 {
		return
	}

	resStr, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return
	}

	// 清理可能的 Markdown 格式
	resStr = strings.TrimPrefix(resStr, "```json")
	resStr = strings.TrimSuffix(resStr, "```")
	resStr = strings.TrimSpace(resStr)

	var graph struct {
		Entities []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"entities"`
		Relations []struct {
			Subject     string `json:"subject"`
			Predicate   string `json:"predicate"`
			Object      string `json:"object"`
			Description string `json:"description"`
		} `json:"relations"`
	}

	if err := json.Unmarshal([]byte(resStr), &graph); err != nil {
		return
	}

	// 保存实体
	entityMap := make(map[string]uint)
	for _, e := range graph.Entities {
		entity := &KnowledgeEntity{
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
		}
		if err := idx.kb.SaveEntity(ctx, entity); err == nil {
			entityMap[e.Name] = entity.ID
		}
	}

	// 保存关系
	for _, r := range graph.Relations {
		subID, ok1 := entityMap[r.Subject]
		objID, ok2 := entityMap[r.Object]
		if ok1 && ok2 {
			relation := &KnowledgeRelation{
				SubjectID:   subID,
				Predicate:   r.Predicate,
				ObjectID:    objID,
				Description: r.Description,
				DocID:       docID,
			}
			idx.kb.SaveRelation(ctx, relation)
		}
	}
}

// IndexSkills 索引技能列表
func (idx *Indexer) IndexSkills(ctx context.Context, skills []tasks.Capability) error {
	for _, skill := range skills {
		content := skill.GenerateSkillGuide()
		source := fmt.Sprintf("skill://%s", skill.Name)
		title := fmt.Sprintf("技能: %s", skill.Name)
		if err := idx.IndexContent(ctx, title, source, []byte(content), "skill", "system", "system", "system"); err != nil {
			log.Printf("[Indexer] Failed to index skill %s: %v", skill.Name, err)
		}
	}
	log.Printf("[Indexer] Indexed %d skills", len(skills))
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
