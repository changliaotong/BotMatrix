package development_team

import (
	"BotMatrix/common/models"
	"BotMatrix/common/database"
	"log"
	"time"
)

// ExampleDBUsage demonstrates how to use the database-driven digital team
// 数据库驱动的数字团队使用示例
func ExampleDBUsage() {
	// Initialize database manager
	dbManager, err := NewDatabaseManager()
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}
	
	// Seed initial roles if not exist
	err = dbManager.SeedInitialRoles()
	if err != nil {
		log.Printf("Failed to seed initial roles: %v", err)
	}
	
	// Load a role from database
	loader := NewDynamicRoleLoader(database.GetDB())
	programmer, err := loader.LoadRole("程序员")
	if err != nil {
		log.Fatalf("Failed to load programmer role: %v", err)
	}
	
	log.Printf("Loaded programmer role with %d experience", programmer.GetExperience())
	
	// Create a task
	task := models.DigitalTask{
		TaskName:    "Generate API endpoint",
		TaskType:    "generate_code",
		Description: "Generate a REST API endpoint for user management",
		InputData: map[string]interface{}{
			"language": "Go",
			"framework": "Gin",
			"requirements": "Create CRUD operations for users",
		},
		Priority: 2,
	}
	
	err = dbManager.CreateTask(&task)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
	}
	
	// Execute the task
	result, err := programmer.ExecuteTask(Task{
		ID:          task.ID,
		Type:        task.TaskType,
		Description: task.Description,
		Input:       task.InputData,
		Priority:    task.Priority,
	})
	if err != nil {
		log.Printf("Failed to execute task: %v", err)
	}
	
	// Update task with result
	task.Status = "completed"
	task.OutputData = result.Output
	task.ExecutionTime = result.ExecutionTime
	if result.Error != nil {
		task.Status = "failed"
		task.ErrorMsg = result.Error.Error()
	}
	
	err = dbManager.UpdateTask(&task)
	if err != nil {
		log.Printf("Failed to update task: %v", err)
	}
	
	// Create a project
	project := models.DigitalProject{
		ProjectName: "User Management System",
		Description: "A complete user management system with API and database",
		Requirements: "Create a user management system with authentication, authorization, and user profiles",
		TechStack:   []string{"Go", "Gin", "PostgreSQL", "JWT"},
		Status:      "in_progress",
	}
	
	err = dbManager.CreateProject(&project)
	if err != nil {
		log.Printf("Failed to create project: %v", err)
	}
	
	// Evolve the programmer role based on performance
	evolution := NewDatabaseSelfEvolution(database.GetDB())
	performanceData := map[string]interface{}{
		"performance_score": 95.0,
		"top_skill":         "API design",
	}
	
	// Get the role model
	roleModel, err := dbManager.GetRole("程序员")
	if err != nil {
		log.Printf("Failed to get role model: %v", err)
		return
	}
	
	err = evolution.EvolveRole(roleModel, performanceData)
	if err != nil {
		log.Printf("Failed to evolve role: %v", err)
	}
	
	// Update the role instance
	err = loader.UpdateRole(programmer)
	if err != nil {
		log.Printf("Failed to update role instance: %v", err)
	}
	
	// Get evolution history
	evolutions, err := evolution.GetEvolutionHistory(roleModel.ID)
	if err != nil {
		log.Printf("Failed to get evolution history: %v", err)
	} else {
		log.Printf("Evolution history for %s: %d entries", roleModel.RoleName, len(evolutions))
	}
	
	// Generate performance report
	startDate := time.Now().AddDate(0, 0, -7) // Last 7 days
	endDate := time.Now()
	
	report, err := evolution.GetPerformanceReport(roleModel.ID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to generate performance report: %v", err)
	} else {
		log.Printf("Performance report: %+v", report)
	}
	
	log.Println("Database-driven digital team example completed successfully")
}

// ExampleTeamManager demonstrates how to use the team manager with database
// 数据库驱动的团队管理器使用示例
func ExampleTeamManager() {
	// Initialize database manager
	dbManager, err := NewDatabaseManager()
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}
	
	// Seed initial roles
	err = dbManager.SeedInitialRoles()
	if err != nil {
		log.Printf("Failed to seed initial roles: %v", err)
	}
	
	// Load all roles
	loader := NewDynamicRoleLoader(database.GetDB())
	roles, err := loader.LoadAllRoles()
	if err != nil {
		log.Fatalf("Failed to load roles: %v", err)
	}
	
	// Create team manager
	team := NewDevelopmentTeam()
	for _, role := range roles {
		team.AddRole(role)
	}
	
	log.Printf("Created digital team with %d roles", len(roles))
	
	// Create a project
	project := models.DigitalProject{
		ProjectName: "E-commerce Platform",
		Description: "A complete e-commerce platform with product catalog, shopping cart, and checkout",
		Requirements: "Create a scalable e-commerce platform with microservices architecture",
		TechStack:   []string{"Go", "gRPC", "PostgreSQL", "Redis", "Kubernetes"},
		Status:      "in_progress",
	}
	
	err = dbManager.CreateProject(&project)
	if err != nil {
		log.Printf("Failed to create project: %v", err)
	}
	
	// Execute project tasks
	tasks := []Task{
		{
			ID:          "1",
			Type:        "design_architecture",
			Description: "Design microservices architecture for e-commerce platform",
			Input: map[string]interface{}{
				"requirements": "Create scalable microservices architecture",
			},
			Priority: 1,
		},
		{
			ID:          "2",
			Type:        "generate_code",
			Description: "Generate product service API",
			Input: map[string]interface{}{
				"language": "Go",
				"framework": "gRPC",
			},
			Priority: 2,
		},
		{
			ID:          "3",
			Type:        "design_database",
			Description: "Design database schema for product catalog",
			Input: map[string]interface{}{
				"database": "PostgreSQL",
			},
			Priority: 2,
		},
	}
	
	for _, task := range tasks {
		result, err := team.ExecuteTask(task)
		if err != nil {
			log.Printf("Task %s failed: %v", task.ID, err)
			project.Status = "failed"
			break
		}
		
		// Record task result
		project.Results[task.ID] = result
		log.Printf("Task %s completed successfully", task.ID)
	}
	
	if project.Status != "failed" {
		project.Status = "completed"
		project.Progress = 100.0
	}
	
	err = dbManager.UpdateProject(&project)
	if err != nil {
		log.Printf("Failed to update project: %v", err)
	}
	
	log.Println("Team manager example completed successfully")
}