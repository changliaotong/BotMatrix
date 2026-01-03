package app

import (
	"BotMatrix/common/ai"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EmployeeServiceImpl struct {
	db    *gorm.DB
	aiSvc AIIntegrationService
}

func NewEmployeeService(db *gorm.DB) *EmployeeServiceImpl {
	return &EmployeeServiceImpl{db: db}
}

func (s *EmployeeServiceImpl) SetAIService(aiSvc AIIntegrationService) {
	s.aiSvc = aiSvc
}

func (s *EmployeeServiceImpl) GetEmployeeByBotID(botID string) (*models.DigitalEmployeeGORM, error) {
	var employee models.DigitalEmployeeGORM
	if err := s.db.Where("bot_id = ?", botID).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (s *EmployeeServiceImpl) RecordKpi(employeeID uint, metric string, score float64) error {
	log := models.DigitalEmployeeKpiGORM{
		EmployeeID: employeeID,
		MetricName: metric,
		Score:      score,
	}
	if err := s.db.Create(&log).Error; err != nil {
		return err
	}

	// 同时更新员工的平均 KPI 分数 (简化逻辑：取平均值)
	var avgScore float64
	s.db.Model(&models.DigitalEmployeeKpiGORM{}).
		Where("employee_id = ?", employeeID).
		Select("AVG(score)").
		Scan(&avgScore)

	err := s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("id = ?", employeeID).
		Update("kpi_score", avgScore).Error

	if err != nil {
		return err
	}

	// 自动进化触发逻辑：如果平均分低于 85，且最近没有进行过进化，则尝试自动进化
	if avgScore < 85 {
		// 检查最近 24 小时内是否已经进化过，避免过度进化
		var recentEvolution int64
		s.db.Model(&models.DigitalEmployeeKpiGORM{}).
			Where("employee_id = ? AND metric_name = ? AND created_at > ?", employeeID, "auto_evolution", time.Now().Add(-24*time.Hour)).
			Count(&recentEvolution)

		if recentEvolution == 0 {
			go s.AutoEvolve(employeeID)
		}
	}

	return nil
}

func (s *EmployeeServiceImpl) UpdateOnlineStatus(botID string, status string) error {
	return s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("bot_id = ?", botID).
		Update("online_status", status).Error
}

func (s *EmployeeServiceImpl) ConsumeSalary(botID string, tokens int64) error {
	// 基础逻辑：累加消耗的 Token
	err := s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("bot_id = ?", botID).
		UpdateColumn("salary_token", gorm.Expr("salary_token + ?", tokens)).Error

	if err != nil {
		return err
	}

	// 进阶逻辑：记录流水日志（可选，用于后续审计和报表）
	// TODO: 实现 AIUsageLog 与 DigitalEmployee 的关联记录

	return nil
}

// CheckSalaryLimit 检查员工是否超过预算限制
func (s *EmployeeServiceImpl) CheckSalaryLimit(botID string) (bool, error) {
	var employee models.DigitalEmployeeGORM
	if err := s.db.Where("bot_id = ?", botID).First(&employee).Error; err != nil {
		return false, err
	}

	if employee.SalaryLimit > 0 && employee.SalaryToken > employee.SalaryLimit {
		return false, nil // 超过限制
	}

	return true, nil
}

func (s *EmployeeServiceImpl) UpdateSalary(botID string, salaryToken *int64, salaryLimit *int64) error {
	updates := make(map[string]interface{})
	if salaryToken != nil {
		updates["salary_token"] = *salaryToken
	}
	if salaryLimit != nil {
		updates["salary_limit"] = *salaryLimit
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("bot_id = ?", botID).
		Updates(updates).Error
}

func (s *EmployeeServiceImpl) AutoEvolve(employeeID uint) error {
	if s.aiSvc == nil {
		return fmt.Errorf("AI service not initialized")
	}

	// 1. 获取员工及关联的 Agent 信息
	var employee models.DigitalEmployeeGORM
	if err := s.db.Preload("Agent").First(&employee, employeeID).Error; err != nil {
		return err
	}

	if employee.AgentID == 0 {
		return fmt.Errorf("employee %d has no associated agent", employeeID)
	}

	// 2. 获取最近的绩效记录和差评记录
	var kpis []models.DigitalEmployeeKpiGORM
	s.db.Where("employee_id = ?", employeeID).Order("created_at desc").Limit(10).Find(&kpis)

	if len(kpis) == 0 {
		return nil // 无绩效记录，无需进化
	}

	// 汇总绩效反馈
	var feedback string
	var totalScore float64
	for _, k := range kpis {
		totalScore += k.Score
		if k.Detail != "" {
			feedback += fmt.Sprintf("- [%s] %s: %s\n", k.CreatedAt.Format("2006-01-02"), k.MetricName, k.Detail)
		}
	}
	avgScore := totalScore / float64(len(kpis))

	// 如果评分较高且没有负面反馈，跳过进化
	if avgScore >= 95 && feedback == "" {
		return nil
	}

	clog.Info("开始数字员工自动进化", zap.Uint("id", employeeID), zap.String("name", employee.Name), zap.Float64("avg_score", avgScore))

	// 3. 构建优化 Prompt 的指令
	systemPrompt := `你是一个资深的 AI 提示词架构师。你的任务是根据数字员工的当前系统提示词和最近的 KPI 绩效反馈，优化其提示词。
数字员工信息：
- 姓名：{{.Name}}
- 职位：{{.Title}}
- 部门：{{.Department}}
- 简介：{{.Bio}}

当前系统提示词：
"""
{{.CurrentPrompt}}
"""

最近的绩效反馈与评分（平均分：{{.AvgScore}}）：
"""
{{.Feedback}}
"""

请分析反馈中的不足（如：专业度不够、回复太慢、语气生硬、未遵循规范等），并输出一个优化后的、更强大的系统提示词。
要求：
1. 保持原有的人设特征。
2. 针对性地解决反馈中提到的问题。
3. 增强对复杂场景的处理能力。
4. 只输出优化后的系统提示词内容，不要包含其他解释。`

	// 填充模板（这里简单替换）
	instruction := systemPrompt
	instruction = strings.ReplaceAll(instruction, "{{.Name}}", employee.Name)
	instruction = strings.ReplaceAll(instruction, "{{.Title}}", employee.Title)
	instruction = strings.ReplaceAll(instruction, "{{.Department}}", employee.Department)
	instruction = strings.ReplaceAll(instruction, "{{.Bio}}", employee.Bio)
	instruction = strings.ReplaceAll(instruction, "{{.CurrentPrompt}}", employee.Agent.SystemPrompt)
	instruction = strings.ReplaceAll(instruction, "{{.AvgScore}}", fmt.Sprintf("%.2f", avgScore))
	instruction = strings.ReplaceAll(instruction, "{{.Feedback}}", feedback)

	// 4. 调用 AI 生成新 Prompt
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := s.aiSvc.Chat(ctx, employee.Agent.ModelID, []ai.Message{
		{Role: "user", Content: instruction},
	}, nil)

	if err != nil {
		return fmt.Errorf("AI optimization failed: %v", err)
	}

	newPrompt := ""
	if len(resp.Choices) > 0 {
		if content, ok := resp.Choices[0].Message.Content.(string); ok {
			newPrompt = strings.TrimSpace(content)
		}
	}

	if newPrompt == "" || newPrompt == employee.Agent.SystemPrompt {
		return nil // 无变化或生成失败
	}

	// 5. 更新 Agent 提示词
	if err := s.db.Model(&models.AIAgentGORM{}).Where("id = ?", employee.AgentID).Update("system_prompt", newPrompt).Error; err != nil {
		return err
	}

	// 6. 记录进化日志
	evolutionLog := models.DigitalEmployeeKpiGORM{
		EmployeeID: employeeID,
		MetricName: "auto_evolution",
		Score:      avgScore,
		Detail:     fmt.Sprintf("提示词已自动优化。旧评分: %.2f。反馈摘要: %d 条记录已处理。", avgScore, len(kpis)),
	}
	s.db.Create(&evolutionLog)

	clog.Info("数字员工进化成功", zap.Uint("id", employeeID), zap.String("name", employee.Name))

	return nil
}
