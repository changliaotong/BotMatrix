package rag

import (
	"BotNexus/tasks"
	"context"
	"fmt"
)

// TaskAIEmbeddingService 使用 tasks.AIService 生成向量
type TaskAIEmbeddingService struct {
	svc     tasks.AIService
	modelID uint
}

func NewTaskAIEmbeddingService(svc tasks.AIService, modelID uint) *TaskAIEmbeddingService {
	return &TaskAIEmbeddingService{
		svc:     svc,
		modelID: modelID,
	}
}

func (s *TaskAIEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := s.svc.CreateEmbedding(ctx, s.modelID, []string{text})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}
