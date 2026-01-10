package collaboration

import (
	"log"
	"time"
)

// Example 通用协作示例
func Example() {
	// 创建消息总线
	messageBus := NewMessageBus()
	defer messageBus.Stop()

	// 创建工作流管理器
	workflowManager := NewWorkflowManager(messageBus)

	// 创建任务分配器
	taskAssigner := NewTaskAssigner(messageBus)

	// 创建动态角色加载器
	dynamicRoleLoader := NewDynamicRoleLoader(messageBus)

	// 创建可视化
	visualization := NewVisualization(messageBus, workflowManager)

	// 注册角色工厂
	// 这里需要注册具体的角色工厂
	// 例如：开发团队角色工厂

	// 加载角色配置
	// err := dynamicRoleLoader.LoadRolesFromDirectory("config/roles")
	// if err != nil {
	//     log.Printf("Failed to load roles: %v", err)
	//     return
	// }

	// 创建工作流
	steps := []*WorkflowStep{
		{
			ID:          "step_1",
			Name:        "需求分析",
			Description: "分析项目需求",
			Type:        WorkflowStepTypeTask,
			RoleType:    "architect",
			MaxRetries:  3,
		},
		{
			ID:          "step_2",
			Name:        "代码实现",
			Description: "实现核心功能",
			Type:        WorkflowStepTypeTask,
			RoleType:    "programmer",
			Dependencies: []string{"step_1"},
			MaxRetries:  3,
		},
		{
			ID:          "step_3",
			Name:        "测试",
			Description: "测试功能正确性",
			Type:        WorkflowStepTypeTask,
			RoleType:    "tester",
			Dependencies: []string{"step_2"},
			MaxRetries:  3,
		},
		{
			ID:          "step_4",
			Name:        "代码审查",
			Description: "审查代码质量",
			Type:        WorkflowStepTypeTask,
			RoleType:    "reviewer",
			Dependencies: []string{"step_3"},
			MaxRetries:  3,
		},
	}

	workflow, err := workflowManager.CreateWorkflow(
		"软件开发流程",
		"完整的软件开发工作流",
		steps,
	)
	if err != nil {
		log.Printf("Failed to create workflow: %v", err)
		return
	}

	// 启动工作流
	if err := workflowManager.StartWorkflow(workflow.ID); err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	// 等待工作流执行
	time.Sleep(10 * time.Second)

	// 导出可视化数据
	jsonData, err := visualization.ExportToJSON()
	if err != nil {
		log.Printf("Failed to export visualization data: %v", err)
		return
	}
	log.Printf("Visualization data: %s", jsonData)

	// 导出Graphviz格式
	graphvizData, err := visualization.ExportToGraphviz()
	if err != nil {
		log.Printf("Failed to export graphviz data: %v", err)
		return
	}
	log.Printf("Graphviz data: %s", graphvizData)

	// 导出Mermaid格式
	mermaidData, err := visualization.ExportToMermaid()
	if err != nil {
		log.Printf("Failed to export mermaid data: %v", err)
		return
	}
	log.Printf("Mermaid data: %s", mermaidData)
}