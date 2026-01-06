package services

import (
	"time"

	"BotMatrix/common/models"
	"gorm.io/gorm"
)

type MsgCountService struct {
	db *gorm.DB
}

func NewMsgCountService(db *gorm.DB) *MsgCountService {
	return &MsgCountService{db: db}
}

// ExistToday checks if a record exists for today
func (s *MsgCountService) ExistToday(groupId int64, userId int64) (bool, error) {
	var count int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := s.db.Model(&models.MsgCount{}).
		Where("GroupId = ? AND UserId = ? AND CDate = ?", groupId, userId, startOfDay).
		Count(&count).Error
	return count > 0, err
}

// Append inserts a new record
func (s *MsgCountService) Append(botUin int64, groupId int64, groupName string, userId int64, name string) error {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	msgCount := models.MsgCount{
		BotUin:    botUin,
		GroupId:   groupId,
		GroupName: groupName,
		UserId:    userId,
		UserName:  name,
		CDate:     startOfDay,
		CMsg:      1,
	}
	return s.db.Create(&msgCount).Error
}

// Update increments message count or creates new record
func (s *MsgCountService) Update(botUin int64, groupId int64, groupName string, userId int64, name string) error {
	exists, err := s.ExistToday(groupId, userId)
	if err != nil {
		return err
	}

	if !exists {
		return s.Append(botUin, groupId, groupName, userId, name)
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return s.db.Model(&models.MsgCount{}).
		Where("GroupId = ? AND UserId = ? AND CDate = ?", groupId, userId, startOfDay).
		UpdateColumn("CMsg", gorm.Expr("CMsg + ?", 1)).Error
}

// GetMsgCount gets today's message count
func (s *MsgCountService) GetMsgCount(groupId int64, userId int64) (int, error) {
	var count int
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := s.db.Model(&models.MsgCount{}).
		Select("CMsg").
		Where("GroupId = ? AND UserId = ? AND CDate = ?", groupId, userId, startOfDay).
		Scan(&count).Error
	return count, err
}

// GetMsgCountY gets yesterday's message count
func (s *MsgCountService) GetMsgCountY(groupId int64, userId int64) (int, error) {
	var count int
	now := time.Now().Add(-24 * time.Hour)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := s.db.Model(&models.MsgCount{}).
		Select("CMsg").
		Where("GroupId = ? AND UserId = ? AND CDate = ?", groupId, userId, startOfDay).
		Scan(&count).Error
	return count, err
}

// GetCountOrder gets today's rank
func (s *MsgCountService) GetCountOrder(groupId int64, userId int64) (int64, error) {
	myMsgCount, err := s.GetMsgCount(groupId, userId)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var rank int64
	err = s.db.Model(&models.MsgCount{}).
		Where("GroupId = ? AND CDate = ? AND CMsg > ?", groupId, startOfDay, myMsgCount).
		Count(&rank).Error

	return rank + 1, err
}

// GetCountOrderY gets yesterday's rank
func (s *MsgCountService) GetCountOrderY(groupId int64, userId int64) (int64, error) {
	myMsgCount, err := s.GetMsgCountY(groupId, userId)
	if err != nil {
		return 0, err
	}

	now := time.Now().Add(-24 * time.Hour)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var rank int64
	err = s.db.Model(&models.MsgCount{}).
		Where("GroupId = ? AND CDate = ? AND CMsg > ?", groupId, startOfDay, myMsgCount).
		Count(&rank).Error

	return rank + 1, err
}

// GetCountList gets today's top N users
func (s *MsgCountService) GetCountList(groupId int64, top int) ([]models.MsgCount, error) {
	var list []models.MsgCount
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := s.db.Where("GroupId = ? AND CDate = ?", groupId, startOfDay).
		Order("CMsg desc").
		Limit(top).
		Find(&list).Error
	return list, err
}

// GetCountListY gets yesterday's top N users
func (s *MsgCountService) GetCountListY(groupId int64, top int) ([]models.MsgCount, error) {
	var list []models.MsgCount
	now := time.Now().Add(-24 * time.Hour)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := s.db.Where("GroupId = ? AND CDate = ?", groupId, startOfDay).
		Order("CMsg desc").
		Limit(top).
		Find(&list).Error
	return list, err
}
