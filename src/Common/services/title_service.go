package services

import (
	"fmt"
	"time"

	"BotMatrix/common/models"
	"gorm.io/gorm"
)

type TitleService struct {
	db *gorm.DB
}

func NewTitleService(db *gorm.DB) *TitleService {
	return &TitleService{db: db}
}

// GetUserTitles returns all titles unlocked by a user
func (s *TitleService) GetUserTitles(userId int64) ([]models.UserTitle, error) {
	var titles []models.UserTitle
	err := s.db.Where("UserId = ?", userId).Find(&titles).Error
	return titles, err
}

// EquipTitle equips a specific title and un-equips others
func (s *TitleService) EquipTitle(userId int64, titleId string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Un-equip all titles for this user
		if err := tx.Model(&models.UserTitle{}).
			Where("UserId = ?", userId).
			Update("IsEquipped", false).Error; err != nil {
			return err
		}

		// 2. Equip the target title
		result := tx.Model(&models.UserTitle{}).
			Where("UserId = ? AND TitleId = ?", userId, titleId).
			Update("IsEquipped", true)

		if result.Error != nil {
			return result.Error
		}
		
		if result.RowsAffected == 0 {
			return fmt.Errorf("未拥有该头衔")
		}

		return nil
	})
}

// UnlockTitle grants a title to a user if they don't have it
func (s *TitleService) UnlockTitle(userId int64, titleId string) error {
	var count int64
	err := s.db.Model(&models.UserTitle{}).
		Where("UserId = ? AND TitleId = ?", userId, titleId).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		newTitle := models.UserTitle{
			UserId:     userId,
			TitleId:    titleId,
			UnlockTime: time.Now(),
			IsEquipped: false,
		}
		return s.db.Create(&newTitle).Error
	}

	return nil
}
