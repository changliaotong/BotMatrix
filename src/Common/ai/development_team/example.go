package development_team

import (
	"BotMatrix/common/ai"
	"fmt"
)

func ExampleDevelopmentTeam(aiSvc ai.AIService) {
	// 创建开发团队
	team := NewDevelopmentTeam(aiSvc)
	
	// 打印团队成员
	fmt.Println("Development Team Members:")
	for _, member := range team.GetTeamMembers() {
		fmt.Printf("- %s (Experience: %d)", member.GetRole(), member.GetExperience())
		fmt.Printf("  Skills: %v\n", member.GetSkills())
	}
	
	// 创建一个项目
	project := &Project{
		ID:          "proj_001",
		Name:        "数字开发团队自我进化系统",
		Description: "构建一个能够自我开发、自我进化的机器人系统",
		Status:      "not_started",
		Tasks: []Task{
			{
				ID:          "task_001",
				Type:        "design_architecture",
				Description: "设计数字开发团队自我进化系统的架构",
				Input: map[string]interface{}{
					"requirements": "构建一个能够自我开发、自我进化的机器人系统，包括架构师、程序员、数据库专员、测试人员、审查员等角色",
				},
				Priority: 1,
			},
			{
				ID:          "task_002",
				Type:        "generate_code",
				Description: "生成系统核心模块的代码",
				Input: map[string]interface{}{
					"prompt":    "生成数字开发团队自我进化系统的核心模块代码",
					"language": "Go",
				},
				Priority: 2,
			},
			{
				ID:          "task_003",
				Type:        "generate_schema",
				Description: "生成数据库Schema",
				Input: map[string]interface{}{
					"requirements": "为数字开发团队自我进化系统设计数据库Schema",
				},
				Priority: 3,
			},
			{
				ID:          "task_004",
				Type:        "generate_test_cases",
				Description: "生成系统测试用例",
				Input: map[string]interface{}{
					"code":      "// 这里应该是生成的代码",
					"test_type": "unit",
				},
				Priority: 4,
			},
			{
				ID:          "task_005",
				Type:        "review_code",
				Description: "审查系统代码",
				Input: map[string]interface{}{
					"code":      "// 这里应该是生成的代码",
					"standards": []string{"代码规范", "性能优化", "安全检查"},
				},
				Priority: 5,
			},
		},
	}
	
	// 启动项目
	fmt.Println("\nStarting Project:", project.Name)
	err := team.StartProject(project)
	if err != nil {
		fmt.Printf("Project failed: %v\n", err)
	} else {
		fmt.Printf("Project completed successfully!\n")
		fmt.Printf("Results: %v\n", project.Results)
	}
	
	// 打印团队统计信息
	fmt.Println("\nTeam Statistics:")
	stats := team.GetTeamStats()
	for role, stat := range stats {
		fmt.Printf("%s: %v\n", role, stat)
	}
	
	// 训练团队
	fmt.Println("\nTraining team on AI development...")
	team.TrainTeam("ai_development", 20)
	
	// 打印更新后的团队统计信息
	fmt.Println("\nUpdated Team Statistics:")
	stats = team.GetTeamStats()
	for role, stat := range stats {
		fmt.Printf("%s: %v\n", role, stat)
	}
}

func ExampleArchitect() {
	// 这里需要实际的AI服务实例
	// aiSvc := ai.NewAIService(...) 
	// architect := NewArchitect(aiSvc)
	// result := architect.DesignArchitecture("构建一个分布式系统")
	// fmt.Println(result)
}