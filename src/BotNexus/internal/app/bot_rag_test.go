package app

import (
	"BotNexus/tasks"
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestBotRAGInjection(t *testing.T) {
	manifest := tasks.GetDefaultManifest()
	manifest.KnowledgeBase = &MockKnowledgeBase{} // 注入模拟知识库

	mockAI := &MockAIServiceForIdentity{}
	parser := &tasks.AIParser{
		Manifest: manifest,
	}
	parser.SetAIService(mockAI)

	ctx := context.Background()
	// 询问涉及知识库的问题
	_, err := parser.MatchSkillByLLM(ctx, "系统架构是怎么样的？", 1, nil)
	if err != nil {
		t.Fatalf("MatchSkillByLLM failed: %v", err)
	}

	// 验证 System Prompt 是否包含 RAG 内容
	if !strings.Contains(mockAI.LastSystemPrompt, "### 参考文档 (RAG):") {
		t.Error("System Prompt missing RAG header")
	}

	if !strings.Contains(mockAI.LastSystemPrompt, "任务系统由 Task, Execution, Scheduler, Dispatcher 四大核心组件组成") {
		t.Error("System Prompt missing retrieved knowledge chunk")
	}

	if !strings.Contains(mockAI.LastSystemPrompt, "来源: DOCS.md") {
		t.Error("System Prompt missing source info")
	}

	fmt.Println("RAG Injection successful!")
}
