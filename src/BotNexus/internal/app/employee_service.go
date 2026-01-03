package app

import (
	"BotMatrix/common/models"
	"gorm.io/gorm"
)

type EmployeeServiceImpl struct {
	db *gorm.DB
}

func NewEmployeeService(db *gorm.DB) *EmployeeServiceImpl {
	return &EmployeeServiceImpl{db: db}
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

	return s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("id = ?", employeeID).
		Update("kpi_score", avgScore).Error
}

func (s *EmployeeServiceImpl) UpdateOnlineStatus(botID string, status string) error {
	return s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("bot_id = ?", botID).
		Update("online_status", status).Error
}

func (s *EmployeeServiceImpl) ConsumeSalary(botID string, tokens int64) error {
	return s.db.Model(&models.DigitalEmployeeGORM{}).
		Where("bot_id = ?", botID).
		UpdateColumn("salary_token", gorm.Expr("salary_token + ?", tokens)).Error
}
