package ai

import (
	"BotMatrix/common/log"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestDistributedAIChat(t *testing.T) {
	// 0. Init Logger
	_ = log.InitLogger(log.Config{
		Level:       "info",
		Format:      "console",
		Development: true,
	})

	// 1. Setup Nexus Manager
	mgr := NewManager()

	// Set up OnCommandSent hook to capture the command and prevent actual network calls
	mgr.OnCommandSent = func(workerID string, msg types.WorkerCommand) {
		fmt.Printf("Mocked: Command %s sent to worker %s\n", msg.Type, workerID)
	}

	// 2. Mock a Worker with ai_chat capability
	workerID := "test_ai_worker"

	// Register the worker in manager
	mgr.Mutex.Lock()
	mgr.Workers = append(mgr.Workers, &types.WorkerClient{
		ID: workerID,
		Capabilities: []types.WorkerCapability{
			{
				Name:        "ai_chat",
				Description: "Distributed Chat",
			},
		},
	})
	mgr.Mutex.Unlock()

	// 3. Create WorkerAIClient
	client := NewWorkerAIClient(mgr)

	// 4. Run Chat in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		// Simulate Nexus receiving the result from Redis/WS
		// We need to wait for the command to be "sent" first
		// In this mock, we'll just wait a bit and then push to pendingSkillRes
		time.Sleep(100 * time.Millisecond)

		// Find the correlation ID from pendingSkillRes
		var correlationID string
		mgr.pendingSkillRes.Range(func(key, value any) bool {
			correlationID = key.(string)
			return false
		})

		if correlationID != "" {
			resp := ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    RoleAssistant,
							Content: "Hello from distributed worker!",
						},
					},
				},
			}
			respJSON, _ := json.Marshal(resp)

			// Trigger HandleSkillResult
			mgr.HandleSkillResult(types.SkillResult{
				CorrelationID: correlationID,
				Result:        string(respJSON),
				WorkerID:      workerID,
				Status:        "success",
			})
		}
	}()

	// 5. Execute Chat
	resp, err := client.Chat(ctx, ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Hi"}},
	})

	if err != nil {
		t.Fatalf("Distributed Chat failed: %v", err)
	}

	if resp.Choices[0].Message.Content != "Hello from distributed worker!" {
		t.Errorf("Unexpected response: %v", resp.Choices[0].Message.Content)
	}

	fmt.Println("Distributed AI Chat test passed!")
}

func TestDistributedAIEmbedding(t *testing.T) {
	// 0. Init Logger
	_ = log.InitLogger(log.Config{
		Level:       "info",
		Format:      "console",
		Development: true,
	})

	// 1. Setup Nexus Manager
	mgr := NewManager()

	// Set up OnCommandSent hook
	mgr.OnCommandSent = func(workerID string, msg types.WorkerCommand) {
		fmt.Printf("Mocked: Command %s sent to worker %s\n", msg.Type, workerID)
	}

	// 2. Mock a Worker with ai_embedding capability
	workerID := "test_ai_worker"

	mgr.Mutex.Lock()
	mgr.Workers = append(mgr.Workers, &types.WorkerClient{
		ID: workerID,
		Capabilities: []types.WorkerCapability{
			{
				Name:        "ai_embedding",
				Description: "Distributed Embedding",
			},
		},
	})
	mgr.Mutex.Unlock()

	// 3. Create WorkerAIClient
	client := NewWorkerAIClient(mgr)

	// 4. Run CreateEmbedding in a goroutine to simulate result
	go func() {
		time.Sleep(100 * time.Millisecond)

		var correlationID string
		mgr.pendingSkillRes.Range(func(key, value any) bool {
			correlationID = key.(string)
			return false
		})

		if correlationID != "" {
			resp := EmbeddingResponse{
				Data: []EmbeddingData{
					{
						Embedding: []float32{0.1, 0.2, 0.3},
						Index:     0,
					},
				},
				Usage: UsageInfo{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			}
			respJSON, _ := json.Marshal(resp)

			mgr.HandleSkillResult(types.SkillResult{
				CorrelationID: correlationID,
				Result:        string(respJSON),
				WorkerID:      workerID,
				Status:        "success",
			})
		}
	}()

	// 5. Execute Embedding
	resp, err := client.CreateEmbedding(context.Background(), EmbeddingRequest{
		Input: "Hello world",
	})

	if err != nil {
		t.Fatalf("Distributed Embedding failed: %v", err)
	}

	if len(resp.Data) == 0 || resp.Data[0].Embedding[0] != 0.1 {
		t.Errorf("Unexpected response: %v", resp)
	}

	fmt.Println("Distributed AI Embedding test passed!")
}
