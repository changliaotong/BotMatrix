package evolution

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseExample 数据库驱动自主进化系统示例
func DatabaseExample() {
	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open("evolution.db"), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	// 创建数据库驱动自主进化系统
	selfEvolution, err := NewDatabaseEvolution(db, "数字开发团队自主进化系统", "基于数据库的数字开发团队自主进化系统")
	if err != nil {
		log.Printf("Failed to create database evolution system: %v", err)
		return
	}

	// 添加DevOps Agent
	err = selfEvolution.AddAgent("architect", "架构师", []string{"系统设计", "架构优化", "技术选型"})
	if err != nil {
		log.Printf("Failed to add architect agent: %v", err)
		return
	}

	err = selfEvolution.AddAgent("programmer", "程序员", []string{"代码编写", "Bug修复", "单元测试"})
	if err != nil {
		log.Printf("Failed to add programmer agent: %v", err)
		return
	}

	err = selfEvolution.AddAgent("code_reviewer", "代码审计员", []string{"代码审查", "性能优化", "安全检查"})
	if err != nil {
		log.Printf("Failed to add code reviewer agent: %v", err)
		return
	}

	// 添加进化任务
	err = selfEvolution.AddTask("review_pr", "审阅PR #123", map[string]interface{}{"pr_id": "123", "branch": "feature/douyin-adapter"}, 1)
	if err != nil {
		log.Printf("Failed to add review PR task: %v", err)
		return
	}

	err = selfEvolution.AddTask("fix_bug", "修复空指针异常", map[string]interface{}{"bug_id": "BUG-456", "file": "src/douyin/adapter.go"}, 2)
	if err != nil {
		log.Printf("Failed to add fix bug task: %v", err)
		return
	}

	err = selfEvolution.AddTask("write_test", "编写单元测试", map[string]interface{}{"test_file": "src/douyin/adapter_test.go", "coverage": 85}, 3)
	if err != nil {
		log.Printf("Failed to add write test task: %v", err)
		return
	}

	err = selfEvolution.AddTask("generate_plugin", "生成抖音适配器", map[string]interface{}{"plugin_name": "抖音适配器", "version": "1.0.0"}, 1)
	if err != nil {
		log.Printf("Failed to add generate plugin task: %v", err)
		return
	}

	// 列出所有Agent
	agents, err := selfEvolution.ListAgents()
	if err != nil {
		log.Printf("Failed to list agents: %v", err)
		return
	}

	log.Printf("Available agents: %d", len(agents))
	for _, agent := range agents {
		log.Printf("  - %s (%s) - Status: %s", agent.Name, agent.Type, agent.Status)
	}

	// 列出所有任务
	tasks, err := selfEvolution.ListTasks()
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		return
	}

	log.Printf("Available tasks: %d", len(tasks))
	for _, task := range tasks {
		log.Printf("  - %s (%s) - Priority: %d - Status: %s", task.Description, task.Type, task.Priority, task.Status)
	}

	// 执行任务
	if len(agents) > 0 && len(tasks) > 0 {
		log.Printf("Executing task %s with agent %s", tasks[0].Description, agents[0].Name)
		err = selfEvolution.ExecuteTask(tasks[0].ID, agents[0].ID)
		if err != nil {
			log.Printf("Failed to execute task: %v", err)
			return
		}
		log.Printf("Task executed successfully")
	}

	// 列出所有任务（更新后）
	tasks, err = selfEvolution.ListTasks()
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		return
	}

	log.Printf("Updated tasks: %d", len(tasks))
	for _, task := range tasks {
		log.Printf("  - %s (%s) - Priority: %d - Status: %s", task.Description, task.Type, task.Priority, task.Status)
	}

	log.Println("Database evolution system example completed successfully")
}