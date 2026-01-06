package services

import (
	"time"

	"BotMatrix/common/models"
	"gorm.io/gorm"
)

type LimiterService struct {
	store *models.Sz84Store
}

func NewLimiterService(store *models.Sz84Store) *LimiterService {
	return &LimiterService{store: store}
}

// HasUsed checks if the action has been performed today
func (s *LimiterService) HasUsed(groupID *int64, userID int64, actionKey string) (bool, error) {
	var record models.LimiterLog
	
	query := s.store.Meta().DB.Model(&models.LimiterLog{}).
		Where("UserId = ? AND ActionKey = ?", userID, actionKey)
	
	if groupID != nil {
		query = query.Where("GroupId = ?", *groupID)
	} else {
		query = query.Where("GroupId IS NULL")
	}

	err := query.First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	// Check if UsedAt is today
	now := time.Now()
	usedAt := record.UsedAt
	return usedAt.Year() == now.Year() && usedAt.Month() == now.Month() && usedAt.Day() == now.Day(), nil
}

// MarkUsed updates the last used time or creates a new record
func (s *LimiterService) MarkUsed(groupID *int64, userID int64, actionKey string) error {
	var record models.LimiterLog
	
	query := s.store.Meta().DB.Model(&models.LimiterLog{}).
		Where("UserId = ? AND ActionKey = ?", userID, actionKey)
	
	if groupID != nil {
		query = query.Where("GroupId = ?", *groupID)
	} else {
		query = query.Where("GroupId IS NULL")
	}

	err := query.First(&record).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new
		newRecord := models.LimiterLog{
			GroupID:   groupID,
			UserID:    userID,
			ActionKey: actionKey,
			UsedAt:    time.Now(),
		}
		return s.store.Meta().DB.Create(&newRecord).Error
	} else if err != nil {
		return err
	}

	// Update existing
	record.UsedAt = time.Now()
	return s.store.Meta().DB.Save(&record).Error
}

// GetLastUsed retrieves the last time the action was used
func (s *LimiterService) GetLastUsed(groupID *int64, userID int64, actionKey string) (*time.Time, error) {
	var record models.LimiterLog
	
	query := s.store.Meta().DB.Model(&models.LimiterLog{}).
		Where("UserId = ? AND ActionKey = ?", userID, actionKey)
	
	if groupID != nil {
		query = query.Where("GroupId = ?", *groupID)
	} else {
		query = query.Where("GroupId IS NULL")
	}

	err := query.First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &record.UsedAt, nil
}

// TryUse attempts to use an action if not used today
func (s *LimiterService) TryUse(groupID *int64, userID int64, actionKey string) (bool, error) {
	used, err := s.HasUsed(groupID, userID, actionKey)
	if err != nil {
		return false, err
	}
	if used {
		return false, nil
	}

	err = s.MarkUsed(groupID, userID, actionKey)
	return err == nil, err
}
