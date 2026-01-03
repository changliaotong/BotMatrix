package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// MockAIServiceForPrivateChat 模拟私聊场景下的 AI 服务
type MockAIServiceForPrivateChat struct {
	LastSystemPrompt string
	ReturnCrossGroup bool
}

func (m *MockAIServiceForPrivateChat) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	m.LastSystemPrompt, _ = messages[0].Content.(string)

	// 如果用户输入包含“取消”，返回取消意图
	userInput, _ := messages[len(messages)-1].Content.(string)
	if strings.Contains(userInput, "取消") {
		return &ai.ChatResponse{
			Choices: []ai.Choice{
				{
					Message: ai.Message{
						Role: ai.RoleAssistant,
						ToolCalls: []ai.ToolCall{
							{
								ID:   "call_cancel",
								Type: "function",
								Function: ai.FunctionCall{
									Name:      "cancel_task",
									Arguments: "{\"task_id\":\"123\"}",
								},
							},
						},
					},
				},
			},
		}, nil
	}

	groupID := ""
	if m.ReturnCrossGroup {
		groupID = "999"
	}

	// 模拟 AI 解析：用户说“报时”，AI 自动补全了 group_id
	taskData := map[string]any{
		"name":           "报时任务",
		"type":           "cron",
		"action_type":    "send_message",
		"action_params":  fmt.Sprintf("{\"group_id\":\"%s\",\"message\":\"报时啦\"}", groupID),
		"trigger_config": "{\"cron\":\"0 * * * *\"}",
	}
	if groupID == "" {
		taskData["action_params"] = "{\"message\":\"报时啦\"}"
	}
	taskJSON, _ := json.Marshal(taskData)

	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role: ai.RoleAssistant,
					ToolCalls: []ai.ToolCall{
						{
							ID:   "call_private",
							Type: "function",
							Function: ai.FunctionCall{
								Name:      "create_task",
								Arguments: string(taskJSON),
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m *MockAIServiceForPrivateChat) CreateEmbedding(ctx context.Context, modelID uint, input any) (*ai.EmbeddingResponse, error) {
	return &ai.EmbeddingResponse{}, nil
}

func (m *MockAIServiceForPrivateChat) ChatStream(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (<-chan ai.ChatStreamResponse, error) {
	return nil, nil
}

func (m *MockAIServiceForPrivateChat) ChatAgent(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	return m.Chat(ctx, modelID, messages, tools)
}

// MockExecutor 模拟执行器
type MockExecutor struct {
	LastExecutedDraft *tasks.AIDraft
}

func (m *MockExecutor) ExecuteAIDraft(draft *tasks.AIDraft) error {
	m.LastExecutedDraft = draft
	return nil
}

func TestPrivateChatAndCrossGroup(t *testing.T) {
	// 1. 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	db.AutoMigrate(&tasks.Task{}, &tasks.AIDraft{}, &tasks.UserIdentity{}, &models.AIModelGORM{}, &models.AIProviderGORM{})

	// 预先创建用户身份，否则 setUserDefaultGroup 会失败
	db.Create(&tasks.UserIdentity{
		NexusUID:    "nexus_123",
		Platform:    "qq",
		PlatformUID: "user_123",
		Nickname:    "test_user",
	})

	// 2. 初始化 Mock 服务
	aiSvc := &MockAIServiceForPrivateChat{}
	botMgr := &MockBotManager{}
	executor := &MockExecutor{}
	tm := tasks.NewTaskManager(db, nil, botMgr)
	tm.SetExecutor(executor)
	tm.AI.SetAIService(aiSvc)
	tm.AI.Manifest = tasks.GetDefaultManifest() // 确保有 Manifest，否则 Parse 会失败

	// 3. 场景 A: 用户在群 101 中发过消息，记录默认群组
	fmt.Println("\n--- [场景 A: 用户在群 101 发起任务] ---")
	tm.ProcessChatMessage(context.Background(), "bot1", "101", "user_123", "AI 帮我报时")

	// 验证 UserIdentity 是否记录了默认群
	var identity tasks.UserIdentity
	db.Where("platform_uid = ?", "user_123").First(&identity)
	fmt.Printf("用户默认群组记录: %s\n", identity.Metadata)

	// 4. 场景 B: 用户在私聊中发起任务
	fmt.Println("\n--- [场景 B: 用户在私聊发起任务] ---")
	aiSvc.LastSystemPrompt = ""
	tm.ProcessChatMessage(context.Background(), "bot1", "", "user_123", "AI 报时")

	// 验证 System Prompt 是否包含了目标群组 ID
	if !strings.Contains(aiSvc.LastSystemPrompt, "目标群组 ID: 101") {
		t.Errorf("System Prompt should contain target group ID 101, got: %s", aiSvc.LastSystemPrompt)
	}
	fmt.Println("System Prompt 验证通过: 包含了默认群组 ID 101")

	// 验证草稿是否自动补全了 group_id
	var draft tasks.AIDraft
	db.Order("id desc").First(&draft)
	fmt.Printf("生成的草稿数据: %s\n", draft.Data)
	// 因为是在 JSON 字符串中，group_id 可能会带转义反斜杠，或者不带（取决于序列化深度）
	// 我们直接检查是否包含 group_id 和 101
	if !strings.Contains(draft.Data, "group_id") || !strings.Contains(draft.Data, "101") {
		t.Errorf("Draft should automatically include group_id: 101, got: %s", draft.Data)
	}
	fmt.Println("草稿验证通过: 自动补全了 group_id 101")

	// 5. 场景 C: 跨群操作拦截 (普通用户尝试操作群 999)
	fmt.Println("\n--- [场景 C: 跨群操作拦截] ---")
	aiSvc.ReturnCrossGroup = true
	botMgr.Actions = nil // 清空记录
	tm.ProcessChatMessage(context.Background(), "bot1", "101", "user_123", "给群 999 设置报时")

	// 验证是否收到了拦截消息
	intercepted := false
	for _, action := range botMgr.Actions {
		if strings.Contains(action, "权限拦截") || strings.Contains(action, "没有权限") {
			intercepted = true
			break
		}
	}
	if !intercepted {
		t.Errorf("Should have intercepted cross-group operation")
	}
	fmt.Println("跨群拦截验证通过: 成功拦截了非法跨群请求")

	// 6. 场景 D: 即时执行验证 (取消任务)
	fmt.Println("\n--- [场景 D: 即时执行验证 (取消任务)] ---")
	botMgr.Actions = nil
	executor.LastExecutedDraft = nil
	tm.ProcessChatMessage(context.Background(), "bot1", "101", "user_admin", "取消刚才的任务")

	if executor.LastExecutedDraft == nil {
		t.Errorf("Cancel task should be executed immediately, but no draft executed")
	} else if executor.LastExecutedDraft.Intent != string(tasks.AIActionCancelTask) {
		t.Errorf("Executed draft should be cancel_task, got: %s", executor.LastExecutedDraft.Intent)
	}
	fmt.Println("即时执行验证通过: 取消任务指令已直接执行")
}
