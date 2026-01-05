package rag

import (
	"BotMatrix/common/types"
	"context"
	"fmt"
	"strings"
)

// TaskAIEmbeddingService 使用 types.AIService 生成向量
type TaskAIEmbeddingService struct {
	svc       types.AIService
	modelID   uint
	modelName string // 缓存模型名称，用于判断是否需要添加特定前缀
}

func NewTaskAIEmbeddingService(svc types.AIService, modelID uint, modelName string) *TaskAIEmbeddingService {
	return &TaskAIEmbeddingService{
		svc:       svc,
		modelID:   modelID,
		modelName: modelName,
	}
}

func (s *TaskAIEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var input any = []string{text}
	if strings.Contains(strings.ToLower(s.modelName), "vision") {
		// 适配多模态格式
		input = []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		}
	}

	resp, err := s.svc.CreateEmbedding(ctx, s.modelID, input)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

// GenerateQueryEmbedding 为检索生成向量，会自动添加豆包建议的前缀
func (s *TaskAIEmbeddingService) GenerateQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	// 针对豆包模型 (含 doubao 关键字)，添加官方建议的检索前缀
	processedQuery := query
	// 不论是 model_id 还是 model_name 包含 doubao，都应用前缀
	if strings.Contains(strings.ToLower(s.modelName), "doubao") {
		prefix := "为这个句子生成表示以用于检索相关文章："
		processedQuery = prefix + query
	}

	var input any = []string{processedQuery}
	if strings.Contains(strings.ToLower(s.modelName), "vision") {
		// 适配多模态格式
		input = []map[string]any{
			{
				"type": "text",
				"text": processedQuery,
			},
		}
	}

	resp, err := s.svc.CreateEmbedding(ctx, s.modelID, input)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
