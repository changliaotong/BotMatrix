package rag

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotNexus/tasks"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRAGComplexQA(t *testing.T) {
	// 1. 初始化数据库
	dbPath := "rag_qa_test.db"
	defer os.Remove(dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.AIProviderGORM{}, &models.AIModelGORM{}, &models.AIUsageLogGORM{})

	// 2. 初始化 AI 服务模拟
	mockAI := &MockAIServiceForQA{}
	mockAI.db = db

	db.Create(&models.AIProviderGORM{ID: 1, Name: "Volcengine", Type: "openai", BaseURL: "https://ark.cn-beijing.volces.com/api/v3"})
	db.Create(&models.AIModelGORM{ID: 1, ProviderID: 1, ModelID: "doubao-embedding-vision-251215", ModelName: "Doubao Embedding Vision"})

	// 3. 初始化 RAG 组件
	es := rag.NewTaskAIEmbeddingService(mockAI, 1, "doubao-embedding-vision-251215")
	kb := rag.NewPostgresKnowledgeBase(db, es, mockAI, 1)
	kb.Setup()

	indexer := rag.NewIndexer(kb, mockAI, 1)
	ctx := context.Background()

	// 4. 注入核心知识
	docs := map[string]string{
		"architecture.md":  "BotNexus 采用插件化架构，核心组件包括 Dispatcher (负责分发消息), TaskManager (管理任务生命周期) 和 AIParser (解析用户意图)。",
		"weather_skill.md": "天气技能 (weather) 可以查询全球主要城市的天气。参数: city (必填)。风险等级: 低。",
	}

	for name, content := range docs {
		path := name
		os.WriteFile(path, []byte(content), 0644)
		defer os.Remove(path)
		indexer.IndexFile(ctx, path, "doc")
	}

	// 5. 执行 AI 解析请求
	parser := &tasks.AIParser{
		Manifest: tasks.GetDefaultManifest(),
	}
	parser.Manifest.KnowledgeBase = kb
	parser.SetAIService(mockAI)

	fmt.Println("=== 测试复杂问题 1: 系统架构 ===")
	_, err = parser.MatchSkillByLLM(ctx, "系统架构 分发消息", 1, nil)
	if err != nil {
		t.Fatalf("MatchSkillByLLM failed: %v", err)
	}

	if !strings.Contains(mockAI.LastSystemPrompt, "Dispatcher (负责分发消息)") {
		t.Error("RAG context missing expected architectural information")
	}

	fmt.Println("=== 测试复杂问题 2: 技能参数 ===")
	_, err = parser.MatchSkillByLLM(ctx, "天气技能 参数", 1, nil)
	if err != nil {
		t.Fatalf("MatchSkillByLLM failed: %v", err)
	}

	if !strings.Contains(mockAI.LastSystemPrompt, "参数: city (必填)") {
		t.Error("RAG context missing expected skill parameter information")
	}

	fmt.Println("QA Test passed successfully!")
}

// MockAIServiceForQA 记录最后一次 System Prompt 的模拟服务
type MockAIServiceForQA struct {
	AIServiceImpl
	LastSystemPrompt string
}

func (s *MockAIServiceForQA) Chat(ctx context.Context, modelID uint, messages []ai.Message, tools []ai.Tool) (*ai.ChatResponse, error) {
	for _, msg := range messages {
		if msg.Role == ai.RoleSystem {
			s.LastSystemPrompt, _ = msg.Content.(string)
		}
	}
	// 直接返回模拟结果，避免调用 AIServiceImpl.Chat 产生的数据库操作
	return &ai.ChatResponse{
		Choices: []ai.Choice{
			{
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: "这是模拟的 AI 回复内容。",
				},
			},
		},
		Usage: ai.UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}
