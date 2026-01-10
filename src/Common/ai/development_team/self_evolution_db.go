package development_team

import (
	"BotMatrix/common/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// DatabaseSelfEvolution implements self-evolution using database storage
// 数据库驱动的自我进化框架
type DatabaseSelfEvolution struct {
	DB *gorm.DB
}

// NewDatabaseSelfEvolution creates a new DatabaseSelfEvolution instance
func NewDatabaseSelfEvolution(db *gorm.DB) *DatabaseSelfEvolution {
	return &DatabaseSelfEvolution{DB: db}
}

// EvolveRole evolves a digital role based on performance data
// 根据性能数据进化数字角色
func (se *DatabaseSelfEvolution) EvolveRole(role *models.DigitalRole, performanceData map[string]interface{}) error {
	// Record the evolution before updating
	beforeData := map[string]interface{}{
		"experience": role.Experience,
		"level":      role.Level,
		"skills":     role.Skills,
		"config":     role.Config,
	}

	// Calculate new experience based on performance
	performance := performanceData["performance_score"].(float64)
	if performance > 80 {
		role.Experience += int(performance * 0.1)
	} else if performance > 50 {
		role.Experience += int(performance * 0.05)
	}

	// Level up if experience threshold is reached
	newLevel := 1 + role.Experience/100
	if newLevel > role.Level {
		role.Level = newLevel
		log.Printf("Role %s leveled up to %d", role.RoleName, role.Level)
	}

	// Update skills based on performance
	if performance > 90 {
		// Add new skills based on performance
		newSkill := performanceData["top_skill"].(string)
		role.Skills = append(role.Skills, newSkill)
	}

	// Update configuration based on performance
	if performance < 60 {
		// Adjust configuration to improve performance
		role.Config["optimization_level"] = "aggressive"
	}

	// Save the updated role
	err := se.DB.Save(role).Error
	if err != nil {
		return err
	}

	// Record the evolution event
	evolution := &models.DigitalEvolution{
		RoleID:        role.ID,
		RoleName:      role.RoleName,
		EvolutionType: "performance_based",
		Description:   "Role evolution based on performance data",
		BeforeData:    beforeData,
		AfterData: map[string]interface{}{
			"experience": role.Experience,
			"level":      role.Level,
			"skills":     role.Skills,
			"config":     role.Config,
		},
		Effectiveness: performance,
		CreatedAt:     time.Now(),
	}

	return se.DB.Create(evolution).Error
}

// AutoRepairCode automatically repairs code based on error data
// 自动修复代码
func (se *DatabaseSelfEvolution) AutoRepairCode(role *models.DigitalRole, errorData map[string]interface{}) error {
	// Record the repair before updating
	beforeData := map[string]interface{}{
		"experience": role.Experience,
		"skills":     role.Skills,
		"prompt":     role.Prompt,
	}

	// Increase experience for bug fixing
	role.Experience += 20

	// Add bug fixing skill if not present
	bugFixSkill := "bug_fixing"
	found := false
	for _, skill := range role.Skills {
		if skill == bugFixSkill {
			found = true
			break
		}
	}
	if !found {
		role.Skills = append(role.Skills, bugFixSkill)
	}

	// Update prompt to include bug fixing
	role.Prompt += "\n擅长自动识别和修复代码中的bug。"

	// Save the updated role
	err := se.DB.Save(role).Error
	if err != nil {
		return err
	}

	// Record the evolution event
	evolution := &models.DigitalEvolution{
		RoleID:        role.ID,
		RoleName:      role.RoleName,
		EvolutionType: "auto_repair",
		Description:   "Auto-repair code based on error data",
		BeforeData:    beforeData,
		AfterData: map[string]interface{}{
			"experience": role.Experience,
			"skills":     role.Skills,
			"prompt":     role.Prompt,
		},
		Effectiveness: 85,
		CreatedAt:     time.Now(),
	}

	return se.DB.Create(evolution).Error
}

// OptimizeCode automatically optimizes code based on performance metrics
// 自动优化代码
func (se *DatabaseSelfEvolution) OptimizeCode(role *models.DigitalRole, performanceMetrics map[string]interface{}) error {
	// Record the optimization before updating
	beforeData := map[string]interface{}{
		"experience": role.Experience,
		"config":     role.Config,
	}

	// Increase experience for optimization
	role.Experience += 15

	// Update optimization configuration
	role.Config["optimization_level"] = "aggressive"
	role.Config["performance_target"] = performanceMetrics["target"]

	// Save the updated role
	err := se.DB.Save(role).Error
	if err != nil {
		return err
	}

	// Record the evolution event
	evolution := &models.DigitalEvolution{
		RoleID:        role.ID,
		RoleName:      role.RoleName,
		EvolutionType: "code_optimization",
		Description:   "Optimize code based on performance metrics",
		BeforeData:    beforeData,
		AfterData: map[string]interface{}{
			"experience": role.Experience,
			"config":     role.Config,
		},
		Effectiveness: performanceMetrics["improvement"].(float64),
		CreatedAt:     time.Now(),
	}

	return se.DB.Create(evolution).Error
}

// LearnNewSkill makes a role learn a new skill
// 让角色学习新技能
func (se *DatabaseSelfEvolution) LearnNewSkill(role *models.DigitalRole, newSkill string) error {
	// Check if skill already exists
	for _, skill := range role.Skills {
		if skill == newSkill {
			return nil // Already knows this skill
		}
	}

	// Record before state
	beforeData := map[string]interface{}{
		"skills": role.Skills,
	}

	// Add the new skill
	role.Skills = append(role.Skills, newSkill)
	role.Experience += 10

	// Save the updated role
	err := se.DB.Save(role).Error
	if err != nil {
		return err
	}

	// Record the evolution event
	evolution := &models.DigitalEvolution{
		RoleID:        role.ID,
		RoleName:      role.RoleName,
		EvolutionType: "skill_learning",
		Description:   "Learned new skill: " + newSkill,
		BeforeData:    beforeData,
		AfterData: map[string]interface{}{
			"skills": role.Skills,
		},
		Effectiveness: 90,
		CreatedAt:     time.Now(),
	}

	return se.DB.Create(evolution).Error
}

// GetEvolutionHistory retrieves the evolution history for a role
// 获取角色的进化历史
func (se *DatabaseSelfEvolution) GetEvolutionHistory(roleID uint) ([]models.DigitalEvolution, error) {
	var evolutions []models.DigitalEvolution
	err := se.DB.Where("role_id = ?", roleID).Order("created_at DESC").Find(&evolutions).Error
	return evolutions, err
}

// GetPerformanceReport generates a performance report for a role
// 生成角色的性能报告
func (se *DatabaseSelfEvolution) GetPerformanceReport(roleID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	var evolutions []models.DigitalEvolution
	err := se.DB.Where("role_id = ? AND created_at BETWEEN ? AND ?", roleID, startDate, endDate).Find(&evolutions).Error
	if err != nil {
		return nil, err
	}

	// Calculate performance metrics
	totalEvolutions := len(evolutions)
	totalEffectiveness := 0.0
	for _, evo := range evolutions {
		totalEffectiveness += evo.Effectiveness
	}

	averageEffectiveness := 0.0
	if totalEvolutions > 0 {
		averageEffectiveness = totalEffectiveness / float64(totalEvolutions)
	}

	report := map[string]interface{}{
		"total_evolutions":      totalEvolutions,
		"average_effectiveness": averageEffectiveness,
		"evolution_types":       make(map[string]int),
		"evolution_history":     evolutions,
	}

	// Count evolution types
	for _, evo := range evolutions {
		report["evolution_types"].(map[string]int)[evo.EvolutionType]++
	}

	return report, nil
}
