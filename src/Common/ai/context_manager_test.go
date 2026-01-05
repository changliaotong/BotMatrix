package ai

import (
	"testing"
)

func TestContextManager_PruneMessages(t *testing.T) {
	cm := NewContextManager(100) // 极小的限制用于测试

	messages := []Message{
		{Role: "system", Content: "System Instruction"},
		{Role: "user", Content: "Very long message 1 that should be pruned eventually because it's quite long and takes up tokens."},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Short message 2"},
		{Role: "assistant", Content: "Response 2"},
	}

	pruned := cm.PruneMessages(messages)

	// 验证 System 消息是否保留
	if len(pruned) == 0 || pruned[0].Role != "system" {
		t.Errorf("System message should be preserved, got %v", pruned)
	}

	// 验证 Token 总数是否在限制内
	tokens := cm.EstimateTokens(pruned)
	if tokens > cm.MaxTokens {
		t.Errorf("Pruned messages tokens (%d) exceed max tokens (%d)", tokens, cm.MaxTokens)
	}

	// 验证是否保留了最近的消息
	lastMsg := pruned[len(pruned)-1]
	if lastMsg.Content != "Response 2" {
		t.Errorf("Last message should be 'Response 2', got %v", lastMsg.Content)
	}
}

func TestContextManager_EstimateTokens(t *testing.T) {
	cm := NewContextManager(1000)
	
	msg := []Message{{Role: "user", Content: "hello world"}}
	tokens := cm.EstimateTokens(msg)
	
	// "hello world" is 11 chars. 11/4 = 2. 10 (base) + 2 = 12 tokens.
	if tokens != 12 {
		t.Errorf("Expected 12 tokens, got %d", tokens)
	}
}
