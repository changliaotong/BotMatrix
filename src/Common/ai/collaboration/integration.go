package collaboration

import (
	"fmt"
	"log"
)

// Integration 集成模块
// 实现与现有开发团队架构的集成
type Integration struct {
	messageBus         MessageBus
	taskAssigner       *TaskAssigner
	dynamicRoleLoader  *DynamicRoleLoader
	workflowManager    *WorkflowManager
	visualization      *Visualization
}

// NewIntegration 创建新的集成实例
func NewIntegration() (*Integration, error) {
	// 创建消息总线
	messageBus := NewMessageBus()

	// 创建工作流管理器
	workflowManager := NewWorkflowManager(messageBus)

	// 创建任务分配器
	taskAssigner := NewTaskAssigner(messageBus)

	// 创建动态角色加载器
	dynamicRoleLoader := NewDynamicRoleLoader(messageBus)

	// 创建可视化
	visualization := NewVisualization(messageBus, workflowManager)

	return &Integration{
		messageBus:         messageBus,
		taskAssigner:       taskAssigner,
		dynamicRoleLoader:  dynamicRoleLoader,
		workflowManager:    workflowManager,
		visualization:      visualization,
	}, nil
}

// IntegrateDevelopmentTeam 集成开发团队角色
func (i *Integration) IntegrateDevelopmentTeam(roles []interface{}) error {
	for _, role := range roles {
		// 使用角色适配器将开发团队角色适配为通用协作角色
		adaptedRole := NewRoleAdapter(role)

		// 注册角色到任务分配器
		err := i.taskAssigner.AddRole(adaptedRole)
		if err != nil {
			return fmt.Errorf("failed to add role: %v", err)
		}

		// 注册角色到动态角色加载器
		// 这里需要创建角色工厂
		// factory := NewDevelopmentTeamRoleFactory()
		// err = i.dynamicRoleLoader.RegisterRoleFactory(factory)
		// if err != nil {
		//     return fmt.Errorf("failed to register role factory: %v", err)
		// }

		log.Printf("Integrated role: %s (%s)", adaptedRole.GetName(), adaptedRole.GetType())
	}

	return nil
}

// CreateDevelopmentWorkflow 创建开发工作流
func (i *Integration) CreateDevelopmentWorkflow() (*Workflow, error) {
	steps := []*WorkflowStep{
		{
			ID:          "step_requirement_analysis",
			Name:        "需求分析",
			Description: "分析项目需求和用户故事",
			Type:        WorkflowStepTypeTask,
			RoleType:    "architect",
			MaxRetries:  3,
		},
		{
			ID:          "step_system_design",
			Name:        "系统设计",
			Description: "设计系统架构和模块划分",
			Type:        WorkflowStepTypeTask,
			RoleType:    "architect",
			Dependencies: []string{"step_requirement_analysis"},
			MaxRetries:  3,
		},
		{
			ID:          "step_database_design",
			Name:        "数据库设计",
			Description: "设计数据库表结构和索引",
			Type:        WorkflowStepTypeTask,
			RoleType:    "database_specialist",
			Dependencies: []string{"step_system_design"},
			MaxRetries:  3,
		},
		{
			ID:          "step_code_implementation",
			Name:        "代码实现",
			Description: "实现核心功能模块",
			Type:        WorkflowStepTypeTask,
			RoleType:    "programmer",
			Dependencies: []string{"step_database_design"},
			MaxRetries:  3,
		},
		{
			ID:          "step_testing",
			Name:        "测试",
			Description: "进行单元测试和集成测试",
			Type:        WorkflowStepTypeTask,
			RoleType:    "tester",
			Dependencies: []string{"step_code_implementation"},
			MaxRetries:  3,
		},
		{
			ID:          "step_code_review",
			Name:        "代码审查",
			Description: "审查代码质量和规范性",
			Type:        WorkflowStepTypeTask,
			RoleType:    "reviewer",
			Dependencies: []string{"step_testing"},
			MaxRetries:  3,
		},
		{
			ID:          "step_deployment",
			Name:        "部署",
			Description: "部署到生产环境",
			Type:        WorkflowStepTypeTask,
			RoleType:    "architect",
			Dependencies: []string{"step_code_review"},
			MaxRetries:  3,
		},
	}

	return i.workflowManager.CreateWorkflow(
		"软件开发流程",
		"完整的软件开发工作流，从需求分析到部署",
		steps,
	)
}

// RunDevelopmentWorkflow 运行开发工作流
func (i *Integration) RunDevelopmentWorkflow(workflowID string) error {
	return i.workflowManager.StartWorkflow(workflowID)
}

// ExportVisualization 导出可视化数据
func (i *Integration) ExportVisualization() (string, error) {
	return i.visualization.ExportToJSON()
}

// ExportGraphviz 导出Graphviz格式
func (i *Integration) ExportGraphviz() (string, error) {
	return i.visualization.ExportToGraphviz()
}

// ExportMermaid 导出Mermaid格式
func (i *Integration) ExportMermaid() (string, error) {
	return i.visualization.ExportToMermaid()
}