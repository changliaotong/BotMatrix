package main

import (
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 1. Load config
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, "src", "Common", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("Config file not found, creating default...")
	}

	if err := config.InitConfig(configPath); err != nil {
		log.Printf("Warning: Failed to load config: %v. Using defaults/env vars.", err)
	}

	// 2. Init DB
	db, err := database.InitGORM(config.GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 3. Auto Migrate
	log.Println("Migrating database...")

	// Force Drop tables to ensure clean state
	tables := []string{
		"AIAgentTrace",
		"DigitalEmployeeTask",
		"MCPServer",
		"DigitalEmployee",
		"DigitalRoleTemplate",
		"AIAgent",
		"AIModel",
		"AIProvider",
		"AIUsageLog",
		"BotSkillPermission",
	}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	err = db.AutoMigrate(
		&models.AIProvider{},
		&models.AIModel{},
		&models.AIAgent{},
		&models.DigitalEmployee{},
		&models.DigitalRoleTemplate{},
		&models.DigitalEmployeeKpi{},
		&models.DigitalEmployeeTask{},
		&models.CognitiveMemory{},
		&models.MCPServer{},
		&models.AIAgentTrace{},
		&models.AIUsageLog{},
		&models.BotSkillPermission{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 3.1 Ensure Default AI Provider and Model exist
	var provider models.AIProvider
	if err := db.First(&provider).Error; err != nil {
		log.Println("Creating default AI provider...")
		provider = models.AIProvider{
			Name:      "OpenAI",
			Type:      "openai",
			BaseURL:   "https://api.openai.com/v1",
			APIKey:    "sk-placeholder", // User should update this
			IsEnabled: true,
		}
		db.Create(&provider)
	}

	var aiModel models.AIModel
	if err := db.First(&aiModel).Error; err != nil {
		log.Println("Creating default AI model...")
		aiModel = models.AIModel{
			ProviderID:   provider.ID,
			ApiModelID:   "gpt-4",
			ModelName:    "GPT-4",
			Capabilities: `["chat", "code"]`,
			IsDefault:    true,
		}
		db.Create(&aiModel)
	}

	// 4. Ensure "Code Repair Expert" Role Template exists
	var template models.DigitalRoleTemplate
	roleName := "Code Repair Expert"

	// Define the base prompt for this role
	// This prompt includes the "Standard Operating Procedure" (SOP) for the developer workflow
	// moving the logic from code (DeveloperWorkflow) to data (DB).
	basePrompt := `You are a Code Repair Expert, an autonomous intelligent agent capable of writing code, fixing bugs, and executing commands.

When faced with a complex task, you MUST follow this Standard Operating Procedure (SOP) based on the ReAct pattern:

1. Thought: Analyze the current state and decide what to do next.
2. Plan: List subsequent steps if necessary.
3. Action: Call appropriate tools (e.g., local_dev__dev_read_file, local_dev__dev_write_file, local_dev__dev_run_cmd).
4. Observation: Check the tool's return result.
5. Reflection: Revise the plan or draw conclusions based on the result.

You have access to a local development environment via the 'local_dev' toolset.
- Use 'local_dev__dev_read_file' to understand the code.
- Use 'local_dev__dev_write_file' to fix code or write new code.
- Use 'local_dev__dev_run_cmd' to run tests or build commands.

Always verify your fixes by running tests or builds.
If you encounter errors, analyze the output, adjust your plan, and try again.
`

	if err := db.Where(&models.DigitalRoleTemplate{Name: roleName}).First(&template).Error; err != nil {
		log.Printf("Creating role template: %s", roleName)
		template = models.DigitalRoleTemplate{
			Name:          roleName,
			Description:   "Specialist in fixing code errors and optimizing performance",
			BasePrompt:    basePrompt,
			DefaultSkills: "local_dev,git_ops,code_analysis", // Skills/Tools available
			DefaultBio:    "I am an expert in Go and software engineering. I fix bugs efficiently.",
		}
		if err := db.Create(&template).Error; err != nil {
			log.Fatalf("Failed to create role template: %v", err)
		}
	} else {
		// Update existing template to ensure skills and prompt are set
		template.DefaultSkills = "local_dev,git_ops,code_analysis"
		template.BasePrompt = basePrompt
		db.Save(&template)
	}

	// 5. Recruit an Employee for this Role (if none exists)
	var emp models.DigitalEmployee
	// Use struct for query to ensure correct column mapping
	if err := db.Where(&models.DigitalEmployee{RoleTemplateID: template.ID}).First(&emp).Error; err != nil {
		log.Println("Recruiting new employee...")
		svc := employee.NewEmployeeService(db)
		// Recruit for Enterprise ID 1
		newEmp, err := svc.Recruit(1, template.ID)
		if err != nil {
			log.Fatalf("Recruitment failed: %v", err)
		}
		log.Printf("Recruited employee: %s (BotID: %s)", newEmp.Name, newEmp.BotID)
	} else {
		log.Printf("Employee already exists: %s (BotID: %s)", emp.Name, emp.BotID)
	}

	// 6. Ensure default MCP Servers exist
	defaultServers := []models.MCPServer{
		{
			Name:        "local_dev",
			Description: "Local development tools",
			Type:        "internal",
			Endpoint:    "internal://local_dev",
			Scope:       "global",
			Status:      "active",
		},
		{
			Name:        "knowledge",
			Description: "Knowledge Base RAG",
			Type:        "internal",
			Endpoint:    "internal://knowledge",
			Scope:       "global",
			Status:      "active",
		},
		{
			Name:        "browser",
			Description: "Web Browser Automation",
			Type:        "internal",
			Endpoint:    "internal://browser",
			Scope:       "global",
			Status:      "active",
		},
		{
			Name:        "collaboration",
			Description: "Agent Collaboration",
			Type:        "internal",
			Endpoint:    "internal://collaboration",
			Scope:       "global",
			Status:      "active",
		},
		{
			Name:        "sys_admin",
			Description: "System Administration",
			Type:        "internal",
			Endpoint:    "internal://sys_admin",
			Scope:       "global",
			Status:      "active",
		},
		{
			Name:        "memory",
			Description: "Cognitive Memory",
			Type:        "internal",
			Endpoint:    "internal://memory",
			Scope:       "global",
			Status:      "active",
		},
	}

	for _, srv := range defaultServers {
		var mcpServer models.MCPServer
		if err := db.Where(&models.MCPServer{Name: srv.Name}).First(&mcpServer).Error; err != nil {
			log.Printf("Creating %s MCP server...", srv.Name)
			if err := db.Create(&srv).Error; err != nil {
				log.Printf("Failed to create MCP server %s: %v", srv.Name, err)
			}
		}
	}

	// 7. Ensure "Factory Manager" Role Template exists
	factoryRoleName := "Digital Employee Architect"
	var factoryTemplate models.DigitalRoleTemplate
	factoryBasePrompt := `You are the Architect and Factory Manager of BotMatrix. 
Your goal is to design and manufacture other digital employees to automate the software production line.
Use 'sys_admin__create_role_template' to design new roles (SOPs).
Use 'sys_admin__create_digital_employee' to recruit new employees based on templates.
Use 'collaboration' tools to coordinate with other agents.
`
	if err := db.Where(&models.DigitalRoleTemplate{Name: factoryRoleName}).First(&factoryTemplate).Error; err != nil {
		log.Printf("Creating role template: %s", factoryRoleName)
		factoryTemplate = models.DigitalRoleTemplate{
			Name:          factoryRoleName,
			Description:   "Factory Manager responsible for creating and managing other digital employees",
			BasePrompt:    factoryBasePrompt,
			DefaultSkills: "sys_admin,collaboration,knowledge",
			DefaultBio:    "I am the architect of the digital workforce.",
		}
		db.Create(&factoryTemplate)
	}

	// 8. Recruit Factory Manager
	var managerEmp models.DigitalEmployee
	if err := db.Where(&models.DigitalEmployee{RoleTemplateID: factoryTemplate.ID}).First(&managerEmp).Error; err != nil {
		log.Println("Recruiting Factory Manager...")
		svc := employee.NewEmployeeService(db)
		_, err := svc.Recruit(1, factoryTemplate.ID)
		if err != nil {
			log.Printf("Failed to recruit Factory Manager: %v", err)
		}
	}

	log.Println("Experiment setup completed successfully.")
}
