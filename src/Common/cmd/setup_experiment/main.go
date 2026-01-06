package main

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/ai/mcp"
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/service"
	"BotMatrix/common/types"
	"log"

	"gorm.io/gorm"
)

func main() {
	// 0. Init Logger
	clog.InitDefaultLogger()

	// 1. Load config (uses unified lookup logic in config package)
	if err := config.InitConfig(""); err != nil {
		log.Printf("Warning: Failed to load config: %v. Using defaults/env vars.", err)
	}

	// 2. Init DB
	db, err := database.InitGORM(config.GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 3. Auto Migrate
	log.Println("Migrating database (preserving existing data)...")

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
		&models.DigitalFactoryGoal{},
		&models.DigitalFactoryMilestone{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 3.1 Ensure Default AI Provider and Model exist
	var provider models.AIProvider

	// 1. Try to find an existing enabled provider first
	if err := db.Where("\"IsEnabled\" = ?", true).First(&provider).Error; err != nil {
		log.Println("No active AI provider found in DB, creating from config...")

		// Determine provider details from config or default to OpenAI (Placeholder)
		providerType := "openai"
		providerName := "OpenAI"
		baseURL := "https://api.openai.com/v1"
		apiKey := "sk-placeholder"

		if config.GlobalConfig.AIProviderType != "" && config.GlobalConfig.AIProviderType != "mock" {
			providerType = config.GlobalConfig.AIProviderType
			providerName = "PrimaryLLM"
			baseURL = config.GlobalConfig.AIBaseURL
			apiKey = config.GlobalConfig.AIApiKey
		}

		provider = models.AIProvider{
			Name:      providerName,
			Type:      providerType,
			BaseURL:   baseURL,
			APIKey:    apiKey,
			IsEnabled: true,
		}
		if err := db.Create(&provider).Error; err != nil {
			log.Fatalf("Failed to create AI provider: %v", err)
		}
	} else {
		log.Printf("Using existing AI provider from DB: %s (Type: %s)", provider.Name, provider.Type)
	}

	var aiModel models.AIModel
	// 2. Try to find a default model for this provider
	if err := db.Where(&models.AIModel{ProviderID: provider.ID, IsDefault: true}).First(&aiModel).Error; err != nil {
		log.Println("No default AI model found for provider, creating...")

		modelName := "GPT-4"
		modelID := "gpt-4"

		if config.GlobalConfig.AIModelName != "" {
			modelID = config.GlobalConfig.AIModelName
			modelName = config.GlobalConfig.AIModelName
		}

		aiModel = models.AIModel{
			ProviderID:   provider.ID,
			ApiModelID:   modelID,
			ModelName:    modelName,
			Capabilities: `["chat", "code"]`,
			IsDefault:    true,
		}
		if err := db.Create(&aiModel).Error; err != nil {
			log.Fatalf("Failed to create AI model: %v", err)
		}
	} else {
		log.Printf("Using existing default AI model from DB: %s (ID: %s)", aiModel.ModelName, aiModel.ApiModelID)
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

	// 8.1 Create Sample Factory Goals
	var goalCount int64
	db.Model(&models.DigitalFactoryGoal{}).Count(&goalCount)
	if goalCount == 0 {
		log.Println("Creating sample factory goals...")
		goals := []models.DigitalFactoryGoal{
			{
				Title:       "构建数字员工自动产线 (V1)",
				Description: "实现从代码提交到自动修复的完整闭环",
				Status:      "in_progress",
				Progress:    65,
				Priority:    1,
			},
			{
				Title:       "工厂扩容：接入多模型协作",
				Description: "支持 DeepSeek, OpenAI, Anthropic 多模型动态切换与协作",
				Status:      "pending",
				Progress:    20,
				Priority:    2,
			},
		}
		for _, g := range goals {
			db.Create(&g)
			// Create milestones for each goal
			if g.Title == "构建数字员工自动产线 (V1)" {
				milestones := []models.DigitalFactoryMilestone{
					{GoalID: g.ID, Title: "基础设施搭建", Status: "completed", Weight: 30, Order: 1},
					{GoalID: g.ID, Title: "GitLab Webhook 集成", Status: "completed", Weight: 20, Order: 2},
					{GoalID: g.ID, Title: "Web 监控看板 (基础版)", Status: "completed", Weight: 15, Order: 3},
					{GoalID: g.ID, Title: "真实 LLM 深度集成", Status: "in_progress", Weight: 35, Order: 4},
				}
				for _, m := range milestones {
					db.Create(&m)
				}
			}
		}
	}

	// 8.2 Create a Sample Task to show on dashboard
	var taskCount int64
	db.Model(&models.DigitalEmployeeTask{}).Count(&taskCount)
	if taskCount == 0 {
		log.Println("Creating sample task...")
		sampleTask := models.DigitalEmployeeTask{
			ExecutionID: "exec-init-001",
			Title:       "初始化产线监控",
			Description: "确保看板能够正确显示所有数字员工的状态",
			Status:      "done",
			AssigneeID:  emp.ID, // Assign to Code Repair Expert
		}
		db.Create(&sampleTask)
	}

	log.Println("Experiment setup completed successfully.")

	// 9. Start Dashboard Server
	log.Println("Starting Web Dashboard...")
	dashboardSvc := service.NewDashboardService(db)
	dashboardSvc.StartServer("8080")

	// 10. Start Webhook Service
	log.Println("Starting Webhook Service...")
	// We need Redis for TaskService
	redisMgr, err := database.NewRedisManager(config.GlobalConfig)
	if err != nil {
		log.Printf("Warning: Failed to init Redis for Webhook Service: %v. Using degraded TaskService.", err)
	}

	empSvc := employee.NewEmployeeService(db)

	// --- NEW: Initialize AI Service & MCP Manager with Circular Dependency Handling ---
	// 1. Create a simple manager to satisfy MCP requirements
	sm := &simpleManager{db: db}

	// 2. Create AI Service (initially without MCP)
	aiSvc := ai.NewAIService(db, nil, nil)
	sm.aiSvc = aiSvc

	// 3. Create real MCP Manager (registers local_dev, browser, etc.)
	mcpMgr := mcp.NewMCPManager(sm)

	// 4. Inject MCP Manager back into AI Service
	aiSvc.SetMCPManager(mcpMgr)

	empSvc.SetAIService(aiSvc)
	// -------------------------------------------------------

	taskSvc := employee.NewTaskService(db, empSvc, redisMgr)
	webhookSvc := service.NewWebhookService(db, taskSvc)
	webhookSvc.StartServer("8081")

	// Keep main process alive
	log.Println("Factory is running. Dashboard: http://localhost:8080, Webhooks: http://localhost:8081")
	log.Println("Press Ctrl+C to stop.")
	select {}
}

// simpleManager implements types.Manager to satisfy MCP requirements
type simpleManager struct {
	db    *gorm.DB
	aiSvc types.AIService
}

func (m *simpleManager) GetGORMDB() *gorm.DB                                     { return m.db }
func (m *simpleManager) GetKnowledgeBase() types.KnowledgeBase                   { return nil }
func (m *simpleManager) GetAIService() types.AIService                           { return m.aiSvc }
func (m *simpleManager) GetB2BService() types.B2BService                         { return nil }
func (m *simpleManager) GetCognitiveMemoryService() types.CognitiveMemoryService { return nil }
func (m *simpleManager) GetDigitalEmployeeService() types.DigitalEmployeeService { return nil }
func (m *simpleManager) GetDigitalEmployeeTaskService() types.DigitalEmployeeTaskService {
	return nil
}
func (m *simpleManager) GetTaskManager() types.TaskManagerInterface            { return nil }
func (m *simpleManager) GetMCPManager() types.MCPManagerInterface              { return nil }
func (m *simpleManager) ValidateToken(token string) (*types.UserClaims, error) { return nil, nil }
