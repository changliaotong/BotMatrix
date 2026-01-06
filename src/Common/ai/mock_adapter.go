package ai

import (
	"BotMatrix/common/models"
	"context"
	"fmt"
	"strings"
)

type MockClient struct {
	BaseURL string
	APIKey  string
}

func NewMockClient(baseURL, apiKey string) *MockClient {
	return &MockClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *MockClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Simple rule-based response for testing the workflow
	lastMsg := ""
	if len(req.Messages) > 0 {
		content := req.Messages[len(req.Messages)-1].Content
		if str, ok := content.(string); ok {
			lastMsg = str
		}
	}
	fmt.Printf("[MockClient] Received lastMsg: %s\n", lastMsg)

	// Default response
	content := "I am a mock AI. I received your message."
	var toolCalls []ToolCall

	// Simulate ReAct logic based on the prompt content
	// Note: The ReAct agent sends "Observation: ..." as user message after tool execution.
	// The previous message from assistant would have been the ToolCall.

	// For the specific test case: "Read the file 'src/Common/cmd/dummy_test/main.go'..."
	if strings.Contains(lastMsg, "Read the file") && strings.Contains(lastMsg, "main.go") {
		// Step 1: Read file
		content = "Thought: I need to read the file to find the error.\n"
		toolCalls = []ToolCall{
			{
				ID:   "call_read_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "local_dev__dev_read_file",
					Arguments: `{"path": "src/Common/cmd/dummy_test/main.go"}`,
				},
			},
		}
	} else if strings.Contains(lastMsg, "package main") {
		// Likely the file content from Observation
		// Step 2: Fix file
		content = "Thought: I see the syntax error 'fmt.Printl'. I will fix it to 'fmt.Println'.\n"
		toolCalls = []ToolCall{
			{
				ID:   "call_write_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "local_dev__dev_write_file",
					Arguments: `{"path": "src/Common/cmd/dummy_test/main.go", "content": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n"}`,
				},
			},
		}
	} else if strings.Contains(lastMsg, "Successfully wrote to") {
		// Step 3: Run command
		content = "Thought: I have fixed the file. Now I will run it to verify.\n"
		toolCalls = []ToolCall{
			{
				ID:   "call_run_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "local_dev__dev_run_cmd",
					Arguments: `{"command": "go run src/Common/cmd/dummy_test/main.go"}`,
				},
			},
		}
	} else if strings.Contains(lastMsg, "Hello World") {
		// Step 4: Submit
		content = "Thought: Verification successful. Now I will submit the changes.\n"
		toolCalls = []ToolCall{
			{
				ID:   "call_commit_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "local_dev__dev_git_commit",
					Arguments: `{"message": "fix: syntax error in main.go (automated test)"}`,
				},
			},
		}
	} else if strings.Contains(lastMsg, "Git commit successful") || strings.Contains(lastMsg, "Nothing to commit") {
		// Step 5: Final Answer
		content = "Final Answer: I have fixed the syntax error, verified it, and submitted the code."
	}

	return &ChatResponse{
		ID: "mock-chat-id",
		// Object, Created, Model fields removed as they are not in types.ChatResponse
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:      "assistant",
					Content:   content,
					ToolCalls: toolCalls,
				},
				FinishReason: "stop",
			},
		},
		Usage: UsageInfo{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

func (c *MockClient) ChatStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamResponse, error) {
	ch := make(chan ChatStreamResponse)
	close(ch)
	return ch, nil
}

func (c *MockClient) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	return &EmbeddingResponse{
		// Object: "list", // Removed as not in types.EmbeddingResponse? Check definitions.
		Data: []EmbeddingData{
			{
				// Object:    "embedding", // Removed
				Embedding: make([]float32, 1536),
				Index:     0,
			},
		},
		Model: req.Model,
		Usage: UsageInfo{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}, nil
}

func (c *MockClient) GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error) {
	return nil, nil
}

func (c *MockClient) PlanTask(ctx context.Context, executionID string) error {
	return nil
}
