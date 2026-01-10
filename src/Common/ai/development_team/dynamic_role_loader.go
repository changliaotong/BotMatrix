package development_team

import (
	"BotMatrix/common/models"
	"gorm.io/gorm"
	"log"
	"errors"
)

// DynamicRoleLoader dynamically loads digital roles from the database
// 动态角色加载器，从数据库加载数字角色
type DynamicRoleLoader struct {
	DB *gorm.DB
}

// NewDynamicRoleLoader creates a new DynamicRoleLoader instance
func NewDynamicRoleLoader(db *gorm.DB) *DynamicRoleLoader {
	return &DynamicRoleLoader{DB: db}
}

// LoadRole loads a digital role from the database and creates a role instance
// 从数据库加载数字角色并创建角色实例
func (dl *DynamicRoleLoader) LoadRole(roleName string) (DeveloperRole, error) {
	// Get the role from database
	roleModel, err := dl.getRoleModel(roleName)
	if err != nil {
		return nil, err
	}
	
	// Create the appropriate role instance based on role name
	switch roleName {
	case "架构师":
		return dl.loadArchitect(roleModel)
	case "程序员":
		return dl.loadProgrammer(roleModel)
	case "数据库专员":
		return dl.loadDatabaseSpecialist(roleModel)
	case "测试人员":
		return dl.loadTester(roleModel)
	case "审查员":
		return dl.loadReviewer(roleModel)
	default:
		return nil, errors.New("unknown role type: " + roleName)
	}
}

// getRoleModel retrieves a role model from the database
func (dl *DynamicRoleLoader) getRoleModel(roleName string) (*models.DigitalRole, error) {
	var role models.DigitalRole
	err := dl.DB.Where("role_name = ? AND is_active = ?", roleName, true).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// loadArchitect creates an Architect instance from a role model
func (dl *DynamicRoleLoader) loadArchitect(roleModel *models.DigitalRole) (DeveloperRole, error) {
	// Create AI service
	aiSvc, err := NewAIService()
	if err != nil {
		return nil, err
	}
	
	// Create architect instance
	architect := &ArchitectImpl{
		aiSvc:       aiSvc,
		role:        roleModel.RoleName,
		skills:      roleModel.Skills,
		experience:  roleModel.Experience,
		prompt:      roleModel.Prompt,
		config:      roleModel.Config,
	}
	
	return architect, nil
}

// loadProgrammer creates a Programmer instance from a role model
func (dl *DynamicRoleLoader) loadProgrammer(roleModel *models.DigitalRole) (DeveloperRole, error) {
	// Create AI service
	aiSvc, err := NewAIService()
	if err != nil {
		return nil, err
	}
	
	// Create programmer instance
	programmer := &ProgrammerImpl{
		aiSvc:       aiSvc,
		role:        roleModel.RoleName,
		skills:      roleModel.Skills,
		experience:  roleModel.Experience,
		prompt:      roleModel.Prompt,
		config:      roleModel.Config,
	}
	
	return programmer, nil
}

// loadDatabaseSpecialist creates a DatabaseSpecialist instance from a role model
func (dl *DynamicRoleLoader) loadDatabaseSpecialist(roleModel *models.DigitalRole) (DeveloperRole, error) {
	// Create AI service
	aiSvc, err := NewAIService()
	if err != nil {
		return nil, err
	}
	
	// Create database specialist instance
	dbSpecialist := &DatabaseSpecialistImpl{
		aiSvc:       aiSvc,
		role:        roleModel.RoleName,
		skills:      roleModel.Skills,
		experience:  roleModel.Experience,
		prompt:      roleModel.Prompt,
		config:      roleModel.Config,
	}
	
	return dbSpecialist, nil
}

// loadTester creates a Tester instance from a role model
func (dl *DynamicRoleLoader) loadTester(roleModel *models.DigitalRole) (DeveloperRole, error) {
	// Create AI service
	aiSvc, err := NewAIService()
	if err != nil {
		return nil, err
	}
	
	// Create tester instance
	tester := &TesterImpl{
		aiSvc:       aiSvc,
		role:        roleModel.RoleName,
		skills:      roleModel.Skills,
		experience:  roleModel.Experience,
		prompt:      roleModel.Prompt,
		config:      roleModel.Config,
	}
	
	return tester, nil
}

// loadReviewer creates a Reviewer instance from a role model
func (dl *DynamicRoleLoader) loadReviewer(roleModel *models.DigitalRole) (DeveloperRole, error) {
	// Create AI service
	aiSvc, err := NewAIService()
	if err != nil {
		return nil, err
	}
	
	// Create reviewer instance
	reviewer := &ReviewerImpl{
		aiSvc:       aiSvc,
		role:        roleModel.RoleName,
		skills:      roleModel.Skills,
		experience:  roleModel.Experience,
		prompt:      roleModel.Prompt,
		config:      roleModel.Config,
	}
	
	return reviewer, nil
}

// LoadAllRoles loads all active digital roles from the database
// 加载所有激活的数字角色
func (dl *DynamicRoleLoader) LoadAllRoles() ([]DeveloperRole, error) {
	var roleModels []models.DigitalRole
	err := dl.DB.Where("is_active = ?", true).Find(&roleModels).Error
	if err != nil {
		return nil, err
	}
	
	var roles []DeveloperRole
	for _, roleModel := range roleModels {
		role, err := dl.LoadRole(roleModel.RoleName)
		if err != nil {
			log.Printf("Failed to load role %s: %v", roleModel.RoleName, err)
			continue
		}
		roles = append(roles, role)
	}
	
	return roles, nil
}

// UpdateRole updates a role instance with the latest data from the database
// 更新角色实例，使用数据库中的最新数据
func (dl *DynamicRoleLoader) UpdateRole(role DeveloperRole) error {
	roleName := role.GetRole()
	roleModel, err := dl.getRoleModel(roleName)
	if err != nil {
		return err
	}
	
	// Update the role instance with the latest data
	switch r := role.(type) {
	case *ArchitectImpl:
		r.skills = roleModel.Skills
		r.experience = roleModel.Experience
		r.prompt = roleModel.Prompt
		r.config = roleModel.Config
	case *ProgrammerImpl:
		r.skills = roleModel.Skills
		r.experience = roleModel.Experience
		r.prompt = roleModel.Prompt
		r.config = roleModel.Config
	case *DatabaseSpecialistImpl:
		r.skills = roleModel.Skills
		r.experience = roleModel.Experience
		r.prompt = roleModel.Prompt
		r.config = roleModel.Config
	case *TesterImpl:
		r.skills = roleModel.Skills
		r.experience = roleModel.Experience
		r.prompt = roleModel.Prompt
		r.config = roleModel.Config
	case *ReviewerImpl:
		r.skills = roleModel.Skills
		r.experience = roleModel.Experience
		r.prompt = roleModel.Prompt
		r.config = roleModel.Config
	}
	
	return nil
}

// ReloadRole reloads a role from the database
// 重新加载角色（从数据库获取最新数据）
func (dl *DynamicRoleLoader) ReloadRole(role DeveloperRole) (DeveloperRole, error) {
	roleName := role.GetRole()
	return dl.LoadRole(roleName)
}