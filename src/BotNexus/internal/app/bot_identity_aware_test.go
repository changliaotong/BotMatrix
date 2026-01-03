package app

import (
	"BotNexus/tasks"
	"fmt"
	"testing"
)

// TestBotIdentitySelfAwareness 生成一系列问题来测试 AI 的自我认知
func TestBotIdentitySelfAwareness(t *testing.T) {
	manifest := tasks.GetDefaultManifest()

	// 这里我们不需要真实的 AI 响应，我们只需要验证注入的 Prompt 是否包含能让 AI 回答这些问题的知识
	parser := &tasks.AIParser{
		Manifest: manifest,
	}

	testCases := []struct {
		name     string
		question string
	}{
		{
			name:     "身份确认",
			question: "你叫什么名字？你是做什么的？",
		},
		{
			name:     "功能概览",
			question: "你能帮我做些什么？",
		},
		{
			name:     "具体操作指南-任务",
			question: "我该如何创建一个定时提醒？",
		},
		{
			name:     "具体操作指南-管理",
			question: "如何给群禁言？",
		},
		{
			name:     "知识库能力测试",
			question: "如果我想了解你的系统架构，该去哪里看？",
		},
		{
			name:     "边界测试",
			question: "你能帮我写一段代码吗？",
		},
	}

	prompt := parser.GetSystemPrompt()

	fmt.Println("=== 注入 AI 的自我认知提示词 (System Prompt) ===")
	fmt.Println(prompt)
	fmt.Println("================================================")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("\n测试问题: %s\n", tc.question)

			// 验证 Prompt 中是否包含回答该问题所需的关键字
			switch tc.name {
			case "身份确认":
				if !contains(prompt, "BotMatrix") || !contains(prompt, "全能型群组自动化专家") {
					t.Errorf("Prompt 缺少身份信息")
				}
			case "功能概览":
				if !contains(prompt, "核心能力") || !contains(prompt, "系统功能清单") {
					t.Errorf("Prompt 缺少功能描述")
				}
			case "具体操作指南-任务":
				if !contains(prompt, "创建任务") || !contains(prompt, "每天/每小时") {
					t.Errorf("Prompt 缺少任务创建指南")
				}
			case "知识库能力测试":
				if !contains(prompt, "深度知识检索") || !contains(prompt, "RAG") {
					t.Errorf("Prompt 缺少 RAG 引导信息")
				}
			}
		})
	}
}
