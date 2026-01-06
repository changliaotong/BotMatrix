package main

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/employee" // Now we use employee package which has TaskService
	"BotMatrix/common/ai/mcp"
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func main() {
	// 1. Load Config & DB
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, "src", "Common", "config.json")
	if err := config.InitConfig(configPath); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	// Init Logger
	logConfig := clog.Config{
		Level:       "info",
		Format:      "console",
		Development: true,
	}
	if err := clog.InitLogger(logConfig); err != nil {
		log.Printf("Warning: Failed to init logger: %v", err)
	}

	db, err := database.InitGORM(config.GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 2. Initialize Services
	// Mock AI Service for testing flow logic (or use real one if configured)
	// For this test, we want to see if the workflow *starts* and generates the prompt.
	// Since we might not have a real OpenAI key, we might get an error or need a mock.
	// But let's try to set up the real structure.

	// MCP Manager
	// We need to register LocalDevHost
	mockMgr := &MockManager{db: db}
	mcpManager := mcp.NewMCPManager(mockMgr)
	// Register local_dev tool
	localDev := mcp.NewLocalDevMCPHost(cwd)

	mcpManager.RegisterServer(types.MCPServerInfo{ // Note: types.MCPServerInfo or models?
		ID:    "local_dev",
		Name:  "Local Development Environment",
		Scope: types.ScopeGlobal, // types.ScopeGlobal?
	}, localDev)

	// 2. Setup AI Provider (Mock)
	// Upsert the mock provider into DB
	mockProviderModel := models.AIProvider{
		Name:      "MockProvider",
		Type:      "mock",
		BaseURL:   "http://mock",
		APIKey:    "mock-key",
		IsEnabled: true,
	}
	if err := db.Where(models.AIProvider{Name: "MockProvider"}).FirstOrCreate(&mockProviderModel).Error; err != nil {
		log.Fatalf("Failed to create mock provider: %v", err)
	}

	// Update the default model (ID 1) to use the mock provider
	var model models.AIModel
	if err := db.First(&model, 1).Error; err == nil {
		model.ProviderID = mockProviderModel.ID
		db.Save(&model)
	} else {
		log.Printf("Warning: Model ID 1 not found. Creating it...")
		model = models.AIModel{
			ID:           1,
			ProviderID:   mockProviderModel.ID,
			ApiModelID:   "gpt-4-mock",
			ModelName:    "GPT-4 Mock",
			Capabilities: `["chat"]`,
			IsDefault:    true,
		}
		db.Create(&model)
	}

	provider := &MockAIServiceProvider{db: db}
	aiSvc := ai.NewAIService(db, provider, mcpManager)
	mockMgr.aiSvc = aiSvc

	// Employee Service
	empSvc := employee.NewEmployeeService(db)
	empSvc.SetAIService(aiSvc)
	aiSvc.SetEmployeeService(empSvc) // Add this to close the loop if needed

	// 3. Get the Employee
	var emp models.DigitalEmployee
	// Use struct-based condition to handle column names automatically
	if err := db.Where(&models.DigitalEmployee{Name: "Code Repair Expert"}).First(&emp).Error; err != nil {
		log.Fatalf("Employee not found: %v. Run setup_experiment first.", err)
	}
	// Need to load Agent for System Prompt
	if emp.AgentID != 0 {
		db.First(&emp.Agent, emp.AgentID)
	}

	// 4. Start Workflow via TaskService (Generic, DB-driven)
	log.Println("Creating generic task for employee...")

	// Init Redis
	redisMgr, err := database.NewRedisManager(config.GlobalConfig)
	if err != nil {
		log.Printf("Warning: Redis init failed: %v. TaskService persistence might be limited.", err)
	}

	// Create a new task service instance
	taskSvc := employee.NewTaskService(db, empSvc, redisMgr)

	ctx := context.Background()

	task := models.DigitalEmployeeTask{
		Description: "Read the file 'src/Common/cmd/dummy_test/main.go', find the syntax error (fmt.Printl), and fix it using 'dev_write_file'. Then run 'go run src/Common/cmd/dummy_test/main.go' to verify. Finally, submit the changes using 'dev_git_commit'.",
		AssigneeID:  emp.ID,
		Status:      "pending",
		TaskType:    "ai",
		// Priority:    1, // Assuming Priority field exists or will be added. Commented out if not exists.
		// Title:       "Fix dummy test compilation error", // Assuming Title field exists or will be added.
	}

	// Create task in DB
	if err := taskSvc.CreateTask(ctx, &task); err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}

	log.Printf("Task created with ID: %d", task.ID)

	// Execute task (Generic Execution)
	log.Println("Executing task...")
	if err := taskSvc.ExecuteTask(ctx, fmt.Sprintf("%d", task.ID)); err != nil {
		log.Fatalf("Task execution failed: %v", err)
	}

	log.Println("Task execution completed successfully!")

	// 5. Verify Result
	// Read the file again to check if it's fixed
	content, _ := os.ReadFile("src/Common/cmd/dummy_test/main.go")
	log.Printf("Final file content:\n%s", string(content))

	if strings.Contains(string(content), "fmt.Println") {
		log.Println("Verification Success: 'fmt.Printl' was fixed to 'fmt.Println'")
	} else {
		log.Println("Verification Failed: File content does not seem fixed.")
	}
}

// MockManager implements types.Manager for testing
type MockManager struct {
	db    *gorm.DB
	aiSvc types.AIService
}

func (m *MockManager) GetGORMDB() *gorm.DB {
	return m.db
}
func (m *MockManager) GetKnowledgeBase() types.KnowledgeBase {
	return nil
}
func (m *MockManager) GetAIService() types.AIService {
	return m.aiSvc
}
func (m *MockManager) GetB2BService() types.B2BService {
	return nil
}
func (m *MockManager) GetCognitiveMemoryService() types.CognitiveMemoryService {
	return nil
}
func (m *MockManager) GetDigitalEmployeeService() types.DigitalEmployeeService {
	return nil
}
func (m *MockManager) GetDigitalEmployeeTaskService() types.DigitalEmployeeTaskService {
	return nil
}
func (m *MockManager) GetTaskManager() types.TaskManagerInterface {
	return nil
}
func (m *MockManager) GetMCPManager() types.MCPManagerInterface {
	return nil
}
func (m *MockManager) ValidateToken(token string) (*types.UserClaims, error) {
	return nil, nil
}

// MockAIServiceProvider implements ai.AIServiceProvider for testing
type MockAIServiceProvider struct {
	db *gorm.DB
}

func (m *MockAIServiceProvider) SyncSkillCall(ctx context.Context, skillName string, params map[string]any) (any, error) {
	return nil, nil
}
func (m *MockAIServiceProvider) GetWorkers() []types.WorkerInfo {
	return nil
}
func (m *MockAIServiceProvider) CheckPermission(ctx context.Context, botID string, userID uint, orgID uint, skillName string) (bool, error) {
	return true, nil
}
func (m *MockAIServiceProvider) GetGORMDB() *gorm.DB {
	return m.db
}
func (m *MockAIServiceProvider) GetKnowledgeBase() types.KnowledgeBase {
	return nil
}
func (m *MockAIServiceProvider) GetManifest() *types.SystemManifest {
	return nil
}
func (m *MockAIServiceProvider) IsDigitalEmployeeEnabled() bool {
	return true
}
