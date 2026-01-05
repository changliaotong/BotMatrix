package employee

import (
	"BotMatrix/common/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DigitalEmployeeFactory 是数字员工的“制造工厂”
// 负责根据 DigitalJob (职位) 动态组装出 DigitalEmployee (员工)
type DigitalEmployeeFactory struct {
	db *gorm.DB
}

func NewDigitalEmployeeFactory(db *gorm.DB) *DigitalEmployeeFactory {
	return &DigitalEmployeeFactory{db: db}
}

// RecruitParams 招聘参数
type RecruitParams struct {
	JobID        uint   // 职位 ID
	EnterpriseID uint   // 所属企业 ID
	Name         string // 员工姓名 (可选，不填则自动生成)
	BotID        string // 绑定的底层 Bot 账号 (QQ/WeChat/Discord ID)

	// A/B Testing
	VariantID *uint // 如果属于某个 A/B 测试变体
}

// Recruit 招聘一名新员工 (实例化)
func (f *DigitalEmployeeFactory) Recruit(ctx context.Context, params RecruitParams) (*models.DigitalEmployee, error) {
	// 1. 获取职位定义 (DNA)
	var job models.DigitalJob
	if err := f.db.WithContext(ctx).First(&job, params.JobID).Error; err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	// 2. 检查是否有 Variant (基因突变/特定配置)
	var variantConfig map[string]interface{}
	if params.VariantID != nil {
		var variant models.ABVariant
		if err := f.db.WithContext(ctx).First(&variant, *params.VariantID).Error; err == nil {
			// 解析配置覆盖
			json.Unmarshal([]byte(variant.ConfigOverride), &variantConfig)
		}
	}

	// 3. 自动生成姓名 (如果未提供)
	name := params.Name
	if name == "" {
		name = fmt.Sprintf("%s-%s", job.Name, time.Now().Format("060102"))
		if params.VariantID != nil {
			name += fmt.Sprintf("-V%d", *params.VariantID)
		}
	}

	// 4. 构建 System Prompt (灵魂注入)
	// System Prompt = 职位描述 + 企业文化 + 基础规则 + Variant Override
	systemPrompt := f.buildSystemPrompt(job, variantConfig)

	// 5. 确定模型参数
	modelID := uint(1) // Default
	temperature := float32(0.7)
	if val, ok := variantConfig["temperature"].(float64); ok {
		temperature = float32(val)
	}
	if val, ok := variantConfig["model_id"].(float64); ok {
		modelID = uint(val)
	}

	// 6. 创建 Agent 实体 (大脑)
	agent := models.AIAgent{
		Name:        name + "_Brain",
		Description: fmt.Sprintf("AI Brain for %s (%s)", name, job.Name),
		Type:        "digital_employee",
		Prompt:      systemPrompt,
		ModelID:     modelID,
		Temperature: temperature,
		Tools:       "[]", // 初始为空，稍后填充
		IsPublic:    false,
	}
	if err := f.db.WithContext(ctx).Create(&agent).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent brain: %w", err)
	}

	// 7. 创建 DigitalEmployee 档案 (肉身)
	employeeID := fmt.Sprintf("EMP-%d-%d", params.EnterpriseID, time.Now().UnixNano())
	employee := models.DigitalEmployee{
		EnterpriseID: params.EnterpriseID,
		BotID:        params.BotID,
		EmployeeID:   employeeID,
		Name:         name,
		Title:        job.Name,
		Level:        fmt.Sprintf("P%d", job.Level),
		Department:   job.Department,
		AgentID:      agent.ID,
		Status:       "onboarding",        // 入职中
		SalaryLimit:  job.BaseSalary * 30, // 初始预算
		OnboardingAt: time.Now(),
	}

	if err := f.db.WithContext(ctx).Create(&employee).Error; err != nil {
		return nil, fmt.Errorf("failed to create employee record: %w", err)
	}

	// 8. 记录 A/B 测试关系
	if params.VariantID != nil {
		relation := models.EmployeeExperimentRelation{
			EmployeeID: employee.ID,
			VariantID:  *params.VariantID,
			JoinedAt:   time.Now(),
		}
		// 获取 ExperimentID (稍微绕一下，为了严谨应该在 params 里传，这里简化处理)
		var v models.ABVariant
		if err := f.db.Select("experiment_id").First(&v, *params.VariantID).Error; err == nil {
			relation.ExperimentID = v.ExperimentID
		}
		f.db.WithContext(ctx).Create(&relation)
	}

	// 9. 赋予能力 (装备工具)
	if err := f.equipCapabilities(ctx, &employee, job.ID); err != nil {
		return nil, fmt.Errorf("failed to equip capabilities: %w", err)
	}

	// 10. 赋予初始技能 (注入知识)
	// 如果 Variant 指定了 "memory_snapshot_id"，则进行记忆移植
	if val, ok := variantConfig["memory_snapshot_id"].(float64); ok {
		f.graftMemory(ctx, &employee, uint(val))
	} else {
		f.initSkills(ctx, &employee, job.ID)
	}

	// 11. 激活员工
	employee.Status = "active"
	f.db.WithContext(ctx).Save(&employee)

	return &employee, nil
}

// buildSystemPrompt 根据职位构建 Prompt
func (f *DigitalEmployeeFactory) buildSystemPrompt(job models.DigitalJob, override map[string]interface{}) string {
	var sb strings.Builder

	// 允许 Variant 完全覆盖 Base Prompt
	if val, ok := override["prompt_base"].(string); ok {
		sb.WriteString(val)
	} else {
		sb.WriteString(fmt.Sprintf("You are a %s in the %s department.\n", job.Name, job.Department))
		sb.WriteString(fmt.Sprintf("Your primary responsibility is: %s\n", job.Description))
	}

	sb.WriteString("\nCore Principles:\n")
	sb.WriteString("1. Be professional and efficient.\n")

	// 允许 Variant 添加额外的 Instruction
	if val, ok := override["prompt_suffix"].(string); ok {
		sb.WriteString(val + "\n")
	}

	return sb.String()
}

// equipCapabilities 为员工配置 MCP 工具
func (f *DigitalEmployeeFactory) equipCapabilities(ctx context.Context, employee *models.DigitalEmployee, jobID uint) error {
	// 1. 查找职位要求的能力
	var relations []models.JobCapabilityRelation
	if err := f.db.WithContext(ctx).Where("job_id = ? AND is_required = ?", jobID, true).Find(&relations).Error; err != nil {
		return err
	}

	var capabilityIDs []uint
	for _, r := range relations {
		capabilityIDs = append(capabilityIDs, r.CapabilityID)
	}

	if len(capabilityIDs) == 0 {
		return nil
	}

	// 2. 查找能力详情
	var caps []models.DigitalCapability
	if err := f.db.WithContext(ctx).Where("id IN ?", capabilityIDs).Find(&caps).Error; err != nil {
		return err
	}

	// 3. 将能力转换为 Agent 的 Tools 配置 (JSON)
	toolConfig, _ := json.Marshal(capabilityIDs)

	// 更新 Agent 的 Tools
	return f.db.WithContext(ctx).Model(&models.AIAgent{}).Where("id = ?", employee.AgentID).Update("Tools", string(toolConfig)).Error
}

// initSkills 初始化技能树
func (f *DigitalEmployeeFactory) initSkills(ctx context.Context, employee *models.DigitalEmployee, jobID uint) error {
	// 这里未来可以扩展：根据 Job 查找“推荐技能”，并在 EmployeeSkillRelation 中创建初始记录
	return nil
}

// graftMemory 记忆移植：将快照中的记忆复制给新员工
func (f *DigitalEmployeeFactory) graftMemory(ctx context.Context, target *models.DigitalEmployee, snapshotID uint) error {
	// 这是一个高级功能：
	// 1. 找到 Snapshot 对应的原始记忆
	// 2. 批量复制 CognitiveMemory 记录，将 BotID 替换为 target.BotID
	// 3. 这样新员工一出生就“记得”老员工处理过的经典案例
	// TODO: 实现具体的 SQL 复制逻辑
	return nil
}

// CloneEmployee 克隆一个现有员工 (用于批量复制高绩效模板)
func (f *DigitalEmployeeFactory) CloneEmployee(ctx context.Context, sourceEmployeeID uint, count int) ([]*models.DigitalEmployee, error) {
	var source models.DigitalEmployee
	if err := f.db.WithContext(ctx).Preload("Agent").First(&source, sourceEmployeeID).Error; err != nil {
		return nil, err
	}

	var clones []*models.DigitalEmployee
	for i := 0; i < count; i++ {
		// 实际上是基于 Source 的属性作为 Template 再次 Recruit
		// 但这里简化为直接复制属性
		// TODO: 真正的 Clone 应该创建一个新的 Job 模板或者 Variant，然后基于那个来 Recruit
		// 这里暂留接口
	}
	return clones, nil
}
