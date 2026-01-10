package development_team

import (
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"errors"
	"log"

	"gorm.io/gorm"
)

// DatabaseManager handles database operations for the digital team
// 数据库驱动的数字团队管理器
type DatabaseManager struct {
	DB *gorm.DB
}

// NewDatabaseManager creates a new DatabaseManager
func NewDatabaseManager() (*DatabaseManager, error) {
	// Get the main database connection
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	// Auto-migrate the models
	err := db.AutoMigrate(
		&models.DigitalRole{},
		&models.DigitalTask{},
		&models.DigitalProject{},
		&models.DigitalEvolution{},
	)
	if err != nil {
		log.Printf("Failed to auto-migrate digital team models: %v", err)
		return nil, err
	}

	return &DatabaseManager{DB: db}, nil
}

// CreateRole creates a new digital role in the database
func (dm *DatabaseManager) CreateRole(role *models.DigitalRole) error {
	return dm.DB.Create(role).Error
}

// GetRole retrieves a digital role by name
func (dm *DatabaseManager) GetRole(roleName string) (*models.DigitalRole, error) {
	var role models.DigitalRole
	err := dm.DB.Where("role_name = ? AND is_active = ?", roleName, true).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetAllRoles retrieves all active digital roles
func (dm *DatabaseManager) GetAllRoles() ([]models.DigitalRole, error) {
	var roles []models.DigitalRole
	err := dm.DB.Where("is_active = ?", true).Find(&roles).Error
	return roles, err
}

// UpdateRole updates a digital role in the database
func (dm *DatabaseManager) UpdateRole(role *models.DigitalRole) error {
	return dm.DB.Save(role).Error
}

// DeleteRole deletes a digital role from the database
func (dm *DatabaseManager) DeleteRole(roleName string) error {
	return dm.DB.Where("role_name = ?", roleName).Delete(&models.DigitalRole{}).Error
}

// CreateTask creates a new task in the database
func (dm *DatabaseManager) CreateTask(task *models.DigitalTask) error {
	return dm.DB.Create(task).Error
}

// GetTask retrieves a task by ID
func (dm *DatabaseManager) GetTask(taskID uint) (*models.DigitalTask, error) {
	var task models.DigitalTask
	err := dm.DB.First(&task, taskID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask updates a task in the database
func (dm *DatabaseManager) UpdateTask(task *models.DigitalTask) error {
	return dm.DB.Save(task).Error
}

// CreateProject creates a new project in the database
func (dm *DatabaseManager) CreateProject(project *models.DigitalProject) error {
	return dm.DB.Create(project).Error
}

// GetProject retrieves a project by ID
func (dm *DatabaseManager) GetProject(projectID uint) (*models.DigitalProject, error) {
	var project models.DigitalProject
	err := dm.DB.First(&project, projectID).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates a project in the database
func (dm *DatabaseManager) UpdateProject(project *models.DigitalProject) error {
	return dm.DB.Save(project).Error
}

// RecordEvolution records an evolution event in the database
func (dm *DatabaseManager) RecordEvolution(evolution *models.DigitalEvolution) error {
	return dm.DB.Create(evolution).Error
}

// SeedInitialRoles seeds the initial digital team roles into the database
func (dm *DatabaseManager) SeedInitialRoles() error {
	// Check if roles already exist
	var count int64
	dm.DB.Model(&models.DigitalRole{}).Count(&count)
	if count > 0 {
		log.Println("Digital roles already exist in database")
		return nil
	}

	// Define initial roles
	roles := []models.DigitalRole{
		{
			RoleName:    "架构师",
			Description: "负责系统架构设计、技术栈选择和模块结构创建",
			Skills:      []string{"系统设计", "技术选型", "架构优化", "性能调优"},
			Experience:  100,
			Level:       5,
			Prompt:      `你是一名资深系统架构师，根据需求设计高质量的系统架构。`,
			Config:      map[string]interface{}{"max_complexity": "high", "framework": "microservices"},
			IsActive:    true,
		},
		{
			RoleName:    "程序员",
			Description: "负责代码生成、重构和bug修复",
			Skills:      []string{"Go", "Python", "JavaScript", "Java", "C++", "代码优化", "性能调优"},
			Experience:  200,
			Level:       8,
			Prompt:      `你是一名资深程序员，根据需求生成高质量、可维护的代码。`,
			Config:      map[string]interface{}{"code_quality": "high", "optimization_level": "aggressive"},
			IsActive:    true,
		},
		{
			RoleName:    "数据库专员",
			Description: "负责数据库设计、优化和管理",
			Skills:      []string{"SQL", "PostgreSQL", "MySQL", "MongoDB", "数据库优化", "索引设计"},
			Experience:  150,
			Level:       6,
			Prompt:      `你是一名资深数据库专员，设计高效、安全的数据库系统。`,
			Config:      map[string]interface{}{"performance": "high", "security": "strict"},
			IsActive:    true,
		},
		{
			RoleName:    "测试人员",
			Description: "负责自动化测试用例生成和执行",
			Skills:      []string{"单元测试", "集成测试", "自动化测试", "性能测试", "安全测试"},
			Experience:  120,
			Level:       4,
			Prompt:      `你是一名资深测试人员，设计全面的测试用例确保软件质量。`,
			Config:      map[string]interface{}{"coverage": "high", "automation": "full"},
			IsActive:    true,
		},
		{
			RoleName:    "审查员",
			Description: "负责代码审核和规范检查",
			Skills:      []string{"代码审核", "代码规范", "安全审查", "性能审查", "架构审查"},
			Experience:  180,
			Level:       7,
			Prompt:      `你是一名资深审查员，确保代码符合高质量标准和最佳实践。`,
			Config:      map[string]interface{}{"strictness": "high", "compliance": "full"},
			IsActive:    true,
		},
	}

	// Batch insert roles
	err := dm.DB.CreateInBatches(roles, len(roles)).Error
	if err != nil {
		return err
	}

	log.Printf("Successfully seeded %d digital roles into database", len(roles))
	return nil
}
