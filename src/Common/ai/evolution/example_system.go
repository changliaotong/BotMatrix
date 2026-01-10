package main

import (
	"BotMatrix/common/ai/evolution"
	"context"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockMCPServer 模拟MCP服务器
// 用于示例和测试
type MockMCPServer struct {
	tools     map[string]func(ctx context.Context, args map[string]any) (any, error)
	resources map[string]func(ctx context.Context, uri string) (any, error)
	prompts   map[string]func(ctx context.Context, args map[string]any) (string, error)
}

// NewMockMCPServer 创建新的模拟MCP服务器
func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		tools:     make(map[string]func(ctx context.Context, args map[string]any) (any, error)),
		resources: make(map[string]func(ctx context.Context, uri string) (any, error)),
		prompts:   make(map[string]func(ctx context.Context, args map[string]any) (string, error)),
	}
}

// RegisterTool 注册工具
func (m *MockMCPServer) RegisterTool(tool string, handler func(ctx context.Context, args map[string]any) (any, error)) {
	m.tools[tool] = handler
	log.Printf("Registered MCP tool: %s", tool)
}

// RegisterResource 注册资源
func (m *MockMCPServer) RegisterResource(resource string, provider func(ctx context.Context, uri string) (any, error)) {
	m.resources[resource] = provider
	log.Printf("Registered MCP resource: %s", resource)
}

// RegisterPrompt 注册提示
func (m *MockMCPServer) RegisterPrompt(prompt string, generator func(ctx context.Context, args map[string]any) (string, error)) {
	m.prompts[prompt] = generator
	log.Printf("Registered MCP prompt: %s", prompt)
}

// main 主函数
func main() {
	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open("system_evolution.db"), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	// 创建系统进化模块
	systemEvolution, err := evolution.NewSystemEvolution(db, "数字开发团队自主进化系统", "基于迭代的系统自主进化系统", "v1.0.0")
	if err != nil {
		log.Printf("Failed to create system evolution: %v", err)
		return
	}

	// 创建实用进化系统
	mockMCP := NewMockMCPServer()
	practicalEvolution, err := evolution.NewPracticalEvolution(db, mockMCP, "实用自主进化系统", "真正可运行的自主进化系统", "v1.0.0")
	if err != nil {
		log.Printf("Failed to create practical evolution: %v", err)
		return
	}

	// 创建数字员工
	architect, err := systemEvolution.CreateDigitalEmployee(
		"架构师",
		"architect",
		"你是一名资深系统架构师，负责系统设计和技术选型",
		[]string{"系统设计", "架构优化", "技术选型"},
	)
	if err != nil {
		log.Printf("Failed to create architect employee: %v", err)
		return
	}

	programmer, err := systemEvolution.CreateDigitalEmployee(
		"程序员",
		"programmer",
		"你是一名资深Go程序员，负责代码编写和Bug修复",
		[]string{"代码编写", "Bug修复", "单元测试"},
	)
	if err != nil {
		log.Printf("Failed to create programmer employee: %v", err)
		return
	}

	// 为数字员工添加技能
	err = systemEvolution.AddDigitalEmployeeSkill(programmer.ID, "性能优化")
	if err != nil {
		log.Printf("Failed to add skill to programmer: %v", err)
		return
	}

	// 创建系统迭代
	iteration, err := systemEvolution.CreateIteration(
		"v1.1.0",
		"实现抖音适配器功能",
	)
	if err != nil {
		log.Printf("Failed to create iteration: %v", err)
		return
	}

	// 添加迭代任务
	tasks := []*evolution.IterationTask{
		{
			Name:        "需求分析",
			Description: "分析抖音适配器需求",
			Type:        "requirement_analysis",
			Input: map[string]interface{}{
				"platform": "抖音",
				"features": []string{"视频上传", "评论管理", "数据分析"},
			},
			Status: evolution.TaskStatusCreated,
		},
		{
			Name:        "系统设计",
			Description: "设计抖音适配器架构",
			Type:        "system_design",
			Input: map[string]interface{}{
				"platform":   "抖音",
				"tech_stack": []string{"Go", "REST API", "WebSocket"},
			},
			Status: evolution.TaskStatusCreated,
		},
		{
			Name:        "代码实现",
			Description: "实现抖音适配器核心功能",
			Type:        "code_implementation",
			Input: map[string]interface{}{
				"platform": "抖音",
				"features": []string{"视频上传", "评论管理", "数据分析"},
			},
			Status: evolution.TaskStatusCreated,
		},
		{
			Name:        "单元测试",
			Description: "编写抖音适配器单元测试",
			Type:        "unit_test",
			Input: map[string]interface{}{
				"platform": "抖音",
				"coverage": 85,
			},
			Status: TaskStatusCreated,
		},
		{
			Name:        "集成测试",
			Description: "进行抖音适配器集成测试",
			Type:        "integration_test",
			Input: map[string]interface{}{
				"platform":   "抖音",
				"test_cases": 50,
			},
			Status: TaskStatusCreated,
		},
		{
			Name:        "部署上线",
			Description: "部署抖音适配器到生产环境",
			Type:        "deployment",
			Input: map[string]interface{}{
				"platform":    "抖音",
				"environment": "production",
			},
			Status: TaskStatusCreated,
		},
	}

	for _, task := range tasks {
		err = systemEvolution.AddIterationTask(iteration.ID, task)
		if err != nil {
			log.Printf("Failed to add iteration task: %v", err)
			return
		}
	}

	// 启动迭代
	err = systemEvolution.StartIteration(iteration.ID)
	if err != nil {
		log.Printf("Failed to start iteration: %v", err)
		return
	}

	// 模拟迭代执行
	log.Printf("Running iteration %s...", iteration.Version)
	time.Sleep(5 * time.Second)

	// 完成迭代
	err = systemEvolution.CompleteIteration(iteration.ID)
	if err != nil {
		log.Printf("Failed to complete iteration: %v", err)
		return
	}

	// 更新数字员工提示词
	newPrompt := "你是一名资深Go程序员，负责代码编写、Bug修复和性能优化，精通抖音API集成"
	err = systemEvolution.UpdateDigitalEmployee(programmer.ID, newPrompt, programmer.Skills)
	if err != nil {
		log.Printf("Failed to update programmer prompt: %v", err)
		return
	}

	log.Println("System evolution example completed successfully")
	log.Printf("Current system version: %s", systemEvolution.Version)
}
