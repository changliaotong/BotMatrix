package ai

import (
	"fmt"
)

// ContextManager 处理对话上下文的修剪和摘要
type ContextManager struct {
	MaxTokens int
}

func NewContextManager(maxTokens int) *ContextManager {
	if maxTokens <= 0 {
		maxTokens = 4096 // 默认值
	}
	return &ContextManager{MaxTokens: maxTokens}
}

// EstimateTokens 估算消息列表的 Token 数量 (简单字符估算：1 token ≈ 4 字符)
func (m *ContextManager) EstimateTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += 10 // 基础开销
		if content, ok := msg.Content.(string); ok {
			total += len(content) / 4
		} else if parts, ok := msg.Content.([]ContentPart); ok {
			for _, part := range parts {
				if part.Type == "text" {
					total += len(part.Text) / 4
				}
			}
		}
	}
	return total
}

// PruneMessages 修剪消息列表以适应 Token 限制
// 它会保留 System 消息，并从旧到新修剪 User/Assistant 消息
func (m *ContextManager) PruneMessages(messages []Message) []Message {
	if m.EstimateTokens(messages) <= m.MaxTokens {
		return messages
	}

	// 始终保留第一条 System 消息 (通常是全局指令)
	var systemMsg *Message
	if len(messages) > 0 && messages[0].Role == "system" {
		systemMsg = &messages[0]
	}

	// 尝试保留最近的消息
	pruned := make([]Message, 0)
	if systemMsg != nil {
		pruned = append(pruned, *systemMsg)
	}

	// 从后往前添加消息，直到达到限制
	temp := make([]Message, 0)
	currentTokens := 0
	if systemMsg != nil {
		currentTokens = m.EstimateTokens([]Message{*systemMsg})
	}

	for i := len(messages) - 1; i >= 0; i-- {
		if systemMsg != nil && i == 0 {
			continue // 已经处理过了
		}
		msgTokens := m.EstimateTokens([]Message{messages[i]})
		if currentTokens+msgTokens > m.MaxTokens {
			break
		}
		temp = append([]Message{messages[i]}, temp...)
		currentTokens += msgTokens
	}

	return append(pruned, temp...)
}

// SummarizeOldMessages (占位符) 后续可以调用 LLM 生成摘要
func (m *ContextManager) SummarizeOldMessages(oldMessages []Message) (string, error) {
	if len(oldMessages) == 0 {
		return "", nil
	}
	// 实际实现中，这里会调用一次简短的 LLM 请求来总结这些消息
	return fmt.Sprintf("[此处是较早对话的自动摘要，共 %d 条消息]", len(oldMessages)), nil
}
