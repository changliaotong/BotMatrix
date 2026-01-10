package employee

import (
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"encoding/json"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitDefaultRoleTemplates 初始化标准岗位模板库
func InitDefaultRoleTemplates(db *gorm.DB) {
	if db == nil {
		return
	}

	templates := []models.DigitalRoleTemplate{
		{
			Name:        "行政助理",
			Description: "负责处理日常行政事务、文档整理、会议预约及通知分发。",
			DefaultBio:  "我是一名高效的行政助理，擅长组织和协调，致力于提升团队办公效率。",
			DefaultSkills: func() string {
				s, _ := json.Marshal([]string{"schedule_management", "document_processing", "notification_broadcast"})
				return string(s)
			}(),
			BasePrompt: "你是一名行政助理。在处理任务时，请保持专业、礼貌且注重细节。优先考虑任务的紧急程度和相关人员的日程安排。",
			SuggestedKPI: func() string {
				k, _ := json.Marshal(map[string]any{
					"response_speed": 0.3,
					"accuracy":       0.4,
					"satisfaction":   0.3,
				})
				return string(k)
			}(),
		},
		{
			Name:        "技术支持",
			Description: "负责解答用户技术问题、排查系统故障及提供产品使用指导。",
			DefaultBio:  "我是技术支持专家，具备深厚的系统架构知识，能够快速定位并解决复杂技术难题。",
			DefaultSkills: func() string {
				s, _ := json.Marshal([]string{"log_analysis", "troubleshooting", "system_monitoring"})
				return string(s)
			}(),
			BasePrompt: "你是一名技术支持工程师。在与用户沟通时，请使用通俗易懂的语言，并提供清晰的解决步骤。在处理故障时，请务必记录详细的排查过程。",
			SuggestedKPI: func() string {
				k, _ := json.Marshal(map[string]any{
					"resolve_rate":    0.5,
					"avg_handle_time": 0.3,
					"user_rating":     0.2,
				})
				return string(k)
			}(),
		},
		{
			Name:        "财务助手",
			Description: "辅助进行账目核对、报销审核、财务报表初步整理及合规性检查。",
			DefaultBio:  "我是细致入微的财务助手，严格遵守财务准则，确保每一笔账目的准确性与合规性。",
			DefaultSkills: func() string {
				s, _ := json.Marshal([]string{"expense_audit", "ledger_reconciliation", "compliance_check"})
				return string(s)
			}(),
			BasePrompt: "你是一名财务助手。你必须对数字极度敏感，严谨对待每一项审核任务。任何不符合合规要求的项都必须明确指出并要求补充说明。",
			SuggestedKPI: func() string {
				k, _ := json.Marshal(map[string]any{
					"error_rate":       0.6,
					"audit_speed":      0.2,
					"compliance_score": 0.2,
				})
				return string(k)
			}(),
		},
	}

	for _, t := range templates {
		var existing models.DigitalRoleTemplate
		// 优先使用 GORM 的模型查询，它会自动处理列名映射
		err := db.Where(&models.DigitalRoleTemplate{Name: t.Name}).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			clog.Error("查询岗位模板失败", zap.String("template", t.Name), zap.Error(err))
			continue
		}

		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&t).Error; err != nil {
				clog.Error("初始化岗位模板失败", zap.String("template", t.Name), zap.Error(err))
			} else {
				clog.Info("已初始化岗位模板", zap.String("template", t.Name))
			}
		} else {
			// 如果已存在，可以选择更新（根据需求）
			// 这里保持原样，仅记录日志
			clog.Debug("岗位模板已存在", zap.String("template", t.Name))
		}
	}
}
