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
	return s.db.Create(&log).Error
}
