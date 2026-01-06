package services

import (
	"fmt"

	"BotMatrix/common/models"

	"gorm.io/gorm"
)

type ListService struct {
	db *gorm.DB
}

func NewListService(db *gorm.DB) *ListService {
	return &ListService{db: db}
}

const GroupIdDef = 86433316

// GetSystemBlackListAsync gets the system black list
func (s *ListService) GetSystemBlackListAsync() ([]int64, error) {
	var list []int64
	err := s.db.Model(&models.BlackList{}).Select("BlackId").Where("GroupId = ?", GroupIdDef).Scan(&list).Error
	return list, err
}

// IsSystemBlack checks if a user is in the system black list
func (s *ListService) IsSystemBlack(userId int64) bool {
	var count int64
	s.db.Model(&models.BlackList{}).Where("GroupId = ? AND BlackId = ?", GroupIdDef, userId).Count(&count)
	return count > 0
}

// AddBlackList adds a user to the black list
func (s *ListService) AddBlackList(botUin int64, groupId int64, groupName string, qq int64, name string, blackQQ int64, blackInfo string) error {
	var count int64
	s.db.Model(&models.BlackList{}).Where("GroupId = ? AND BlackId = ?", groupId, blackQQ).Count(&count)
	if count > 0 {
		return fmt.Errorf("user %d is already in black list of group %d", blackQQ, groupId)
	}

	blackList := models.BlackList{
		BotUin:    botUin,
		GroupId:   groupId,
		GroupName: groupName,
		UserId:    qq,
		UserName:  name,
		BlackId:   blackQQ,
		BlackInfo: blackInfo,
	}

	return s.db.Create(&blackList).Error
}

// RemoveBlackList removes a user from the black list
func (s *ListService) RemoveBlackList(groupId int64, blackQQ int64) error {
	return s.db.Where("GroupId = ? AND BlackId = ?", groupId, blackQQ).Delete(&models.BlackList{}).Error
}

// AppendWhiteList adds a user to the white list
func (s *ListService) AppendWhiteList(botUin int64, groupId int64, groupName string, qq int64, name string, whiteQQ int64) error {
	var count int64
	s.db.Model(&models.WhiteList{}).Where("GroupId = ? AND WhiteId = ?", groupId, whiteQQ).Count(&count)
	if count > 0 {
		return fmt.Errorf("user %d is already in white list of group %d", whiteQQ, groupId)
	}

	whiteList := models.WhiteList{
		BotUin:    botUin,
		GroupId:   groupId,
		GroupName: groupName,
		UserId:    qq,
		UserName:  name,
		WhiteId:   whiteQQ,
	}

	return s.db.Create(&whiteList).Error
}

// RemoveWhiteList removes a user from the white list
func (s *ListService) RemoveWhiteList(groupId int64, whiteQQ int64) error {
	return s.db.Where("GroupId = ? AND WhiteId = ?", groupId, whiteQQ).Delete(&models.WhiteList{}).Error
}
