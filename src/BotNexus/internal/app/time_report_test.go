package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/bot"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotNexus/tasks"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// MockBotManager 模拟机器人管理器
type MockBotManager struct {
	Actions []string
}

func (m *MockBotManager) SendBotAction(botID string, action string, params any) error {
	p, _ := json.Marshal(params)
	msg := fmt.Sprintf("[BOT %s] Action: %s, Params: %s", botID, action, string(p))
	fmt.Printf(">>> %s\n", msg)
	m.Actions = append(m.Actions, msg)
	return nil
}
func (m *MockBotManager) SendToWorker(workerID string, msg types.WorkerCommand) error { return nil }
func (m *MockBotManager) FindWorkerBySkill(skillName string) string                   { return "" }
func (m *MockBotManager) GetTags(targetType string, targetID string) []string         { return nil }
func (m *MockBotManager) GetTargetsByTags(targetType string, tags []string, logic string) []string {
	return nil
}
func (m *MockBotManager) GetGroupMembers(botID string, groupID string) ([]types.MemberInfo, error) {
	return []types.MemberInfo{
		{UserID: "admin_1", Role: "admin"},
		{UserID: "user_admin", Role: "admin"},
		{UserID: "user_123", Role: "member"},
	}, nil
}

// MockClientForTimeReport 模拟专门用于整点报时的 AI 客户端
type MockClientForTimeReport struct{}

func (m *MockClientForTimeReport) ChatStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.ChatStreamResponse, error) {
	return nil, nil
}

func (m *MockClientForTimeReport) CreateEmbedding(ctx context.Context, req ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockClientForTimeReport) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	// 模拟 AI 识别意图并调用 create_task
	// 假设用户说：帮我设置一个整点报时，每天 8 点到 22 点在群 100 报时

	taskData := map[string]any{
		"name":           "整点报时任务",
		"type":           "cron",
		"action_type":    "send_message",
		"action_params":  "{\"bot_id\":\"bot123\",\"group_id\":\"100\",\"message\":\"🕙 现在是整点时间，休息一下吧！\"}",
		"trigger_config": "{\"cron\":\"0 8-22 * * *\"}",
	}
	taskJSON, _ := json.Marshal(taskData)

	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role: ai.RoleAssistant,
					ToolCalls: []ai.ToolCall{
						{
							ID:   "call_123",
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

func TestTimeReportTask(t *testing.T) {
	fmt.Println("\n===== [整点报时任务演示测试] =====")

	// 1. 初始化环境
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	err = db.AutoMigrate(&tasks.Task{}, &tasks.AIDraft{}, &models.AIProviderGORM{}, &models.AIModelGORM{}, &models.AIUsageLogGORM{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	bm := &MockBotManager{}
	tm := tasks.NewTaskManager(db, nil, bm)

	// 设置 AI 解析器
	tm.AI.Manifest = tasks.GetDefaultManifest()
	aiSvc := NewAIService(db, nil)
	// 注入模拟的客户端逻辑 (这里需要注意，AIServiceImpl 内部现在使用 getClient 来获取客户端)
	// 为了让 Mock 工作，我们需要模拟一个 Client 接口
	mockClient := &MockClientForTimeReport{}
	// 由于 AIServiceImpl 现在使用 BaseURL+APIKey 的哈希作为缓存，
	// 我们需要确保 Provider 存在并匹配
	provider := models.AIProviderGORM{ID: 1, Name: "Test", BaseURL: "https://api.test.com", APIKey: "test-key"}
	db.Create(&provider)
	cacheKey := "https://api.test.com|test-key"
	aiSvc.clientsByConfig[cacheKey] = mockClient
	tm.AI.SetAIService(aiSvc)
	db.Create(&models.AIModelGORM{
		ID:         1,
		ProviderID: 1,
		ModelID:    "test-model",
		ModelName:  "Test Model",
		IsDefault:  true,
	})

	// 2. 模拟用户在群里说话
	fmt.Println("\n[用户输入]: 帮我设置一个整点报时，每天 8 点到 22 点在群 100 报时")

	// 由于 ProcessChatMessage 内部会调用 tm.AI.Parse，而 tm.AI.Parse 内部又会调用 aiSvc.Chat
	// 我们已经在 ai_service.go 中加了打印，这里会触发打印
	err = tm.ProcessChatMessage(context.Background(), "bot123", "100", "admin_1", "帮我设置一个整点报时，每天 8 点到 22 点在群 100 报时")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// 3. 检查是否生成了草稿
	var draft tasks.AIDraft
	db.First(&draft)
	fmt.Printf("\n[系统生成草稿]: ID=%s, Intent=%s\n", draft.DraftID, draft.Intent)

	// 4. 模拟用户确认
	confirmCmd := "#确认 " + draft.DraftID
	fmt.Printf("\n[用户确认]: %s\n", confirmCmd)

	// 设置执行器 (Manager 实现了 TaskExecutor)
	mgr := &Manager{
		Manager: &bot.Manager{
			GORMDB: db,
		},
		TaskManager: tm,
	}
	tm.Executor = mgr

	err = tm.ProcessChatMessage(context.Background(), "bot123", "100", "admin_1", confirmCmd)
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}

	// 5. 验证任务是否创建成功
	var task tasks.Task
	err = db.Order("id desc").First(&task).Error
	if err != nil {
		fmt.Printf("\n[验证失败]: 未找到创建的任务: %v\n", err)
	} else {
		fmt.Printf("\n[任务创建成功]:\n")
		fmt.Printf("  ID: %d\n", task.ID)
		fmt.Printf("  名称: %s\n", task.Name)
		fmt.Printf("  类型: %s\n", task.Type)
		fmt.Printf("  Cron: %s\n", task.TriggerConfig)
		fmt.Printf("  动作: %s\n", task.ActionType)
		fmt.Printf("  参数: %s\n", task.ActionParams)
	}

	fmt.Println("\n===== [演示结束] =====")
}
