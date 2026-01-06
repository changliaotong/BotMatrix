package ai

import (
	"BotMatrix/common/models"
	"context"
	"encoding/json"
	"fmt"
)

// WorkerAIClient 实现 ai.Client 接口，通过 SyncSkillCall 调用远程 Worker
type WorkerAIClient struct {
	provider AIServiceProvider
}

var _ Client = (*WorkerAIClient)(nil)

func NewWorkerAIClient(provider AIServiceProvider) *WorkerAIClient {
	return &WorkerAIClient{
		provider: provider,
	}
}

func (c *WorkerAIClient) GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error) {
	res, err := c.provider.SyncSkillCall(context.Background(), "employee_get_info", map[string]any{
		"bot_id": botID,
	})
	if err != nil {
		return nil, err
	}
	resStr, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}
	var emp models.DigitalEmployee
	json.Unmarshal([]byte(resStr), &emp)
	return &emp, nil
}

func (c *WorkerAIClient) PlanTask(ctx context.Context, executionID string) error {
	_, err := c.provider.SyncSkillCall(ctx, "employee_plan_task", map[string]any{
		"execution_id": executionID,
	})
	return err
}

func (c *WorkerAIClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	reqBytes, _ := json.Marshal(req)

	// 通过 SyncSkillCall 调用 Worker 的 ai_chat 技能
	res, err := c.provider.SyncSkillCall(ctx, "ai_chat", map[string]any{
		"request": string(reqBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("worker chat failed: %w", err)
	}

	resStr, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected worker response type: %T", res)
	}

	var resp ChatResponse
	if err := json.Unmarshal([]byte(resStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker response: %w", err)
	}

	return &resp, nil
}

func (c *WorkerAIClient) ChatStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamResponse, error) {
	// 目前 SyncSkillCall 不支持流式返回，暂不实现
	return nil, fmt.Errorf("distributed stream chat not supported yet")
}

func (c *WorkerAIClient) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	reqBytes, _ := json.Marshal(req)

	// 通过 SyncSkillCall 调用 Worker 的 ai_embedding 技能
	res, err := c.provider.SyncSkillCall(ctx, "ai_embedding", map[string]any{
		"request": string(reqBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("worker embedding failed: %w", err)
	}

	resStr, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected worker response type: %T", res)
	}

	var resp EmbeddingResponse
	if err := json.Unmarshal([]byte(resStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker response: %w", err)
	}

	return &resp, nil
}
