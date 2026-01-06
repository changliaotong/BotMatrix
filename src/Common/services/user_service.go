package services

import (
	"fmt"

	"BotMatrix/common/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// AppendUser adds a user to the system (User, GroupMember, Friend)
func (s *UserService) AppendUser(botUin int64, groupId int64, userId int64, name string, userOpenid string, groupOpenid string) error {
	// Skip specific users (blacklist/system users)
	ignoredUsers := []int64{2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
		3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880}
	for _, id := range ignoredUsers {
		if userId == id {
			return nil
		}
	}

	// 1. UserInfo.Append
	var count int64
	s.db.Model(&models.UserInfo{}).Where("Id = ?", userId).Count(&count)
	if count == 0 {
		// Get group owner
		var groupOwner int64
		s.db.Model(&models.GroupInfo{}).Select("GroupOwner").Where("Id = ?", groupId).Scan(&groupOwner)

		newUser := models.UserInfo{
			BotUin:      botUin,
			UserOpenId:  userOpenid,
			GroupOpenid: groupOpenid,
			GroupId:     groupId,
			Id:          userId,
			Credit:      50, // Default credit
			Name:        name,
			RefUserId:   groupOwner,
		}
		if userOpenid != "" {
			newUser.Credit = 5000
		}
		if err := s.db.Create(&newUser).Error; err != nil {
			return err
		}
	}

	// 2. GroupMember.Append
	s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).Count(&count)
	if count == 0 {
		newMember := models.GroupMember{
			GroupId:     groupId,
			UserId:      userId,
			UserName:    name,
			GroupCredit: 50,
		}
		if err := s.db.Create(&newMember).Error; err != nil {
			return err
		}
	} else {
		// If exists, update status
		s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).Updates(map[string]interface{}{
			"UserName": name,
			"Status":   1,
		})
	}

	// 3. Friend.Append (if IsCredit)
	var isCredit bool
	s.db.Model(&models.BotInfo{}).Select("IsCredit").Where("BotUin = ?", botUin).Scan(&isCredit)
	if isCredit {
		s.db.Model(&models.Friend{}).Where("BotUin = ? AND UserId = ?", botUin, userId).Count(&count)
		if count == 0 {
			newFriend := models.Friend{
				BotUin:   botUin,
				UserId:   userId,
				UserName: name,
			}
			if err := s.db.Create(&newFriend).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCredit gets the credit of a user in a group
func (s *UserService) GetCredit(groupId int64, userId int64) (int64, error) {
	// Check if group uses group credit
	var isCredit bool
	s.db.Model(&models.GroupInfo{}).Select("IsCredit").Where("Id = ?", groupId).Scan(&isCredit)

	if isCredit {
		// Return GroupCredit from GroupMember
		var credit int64
		err := s.db.Model(&models.GroupMember{}).Select("GroupCredit").Where("GroupId = ? AND UserId = ?", groupId, userId).Scan(&credit).Error
		return credit, err
	} else {
		// Return Credit from UserInfo
		var credit int64
		err := s.db.Model(&models.UserInfo{}).Select("Credit").Where("Id = ?", userId).Scan(&credit).Error
		return credit, err
	}
}

// AddCredit adds credit to a user
func (s *UserService) AddCredit(botUin int64, groupId int64, userId int64, amount int64) error {
	// Check if group uses group credit
	var isCredit bool
	s.db.Model(&models.GroupInfo{}).Select("IsCredit").Where("Id = ?", groupId).Scan(&isCredit)

	if isCredit {
		// Update GroupMember
		return s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).
			UpdateColumn("GroupCredit", gorm.Expr("GroupCredit + ?", amount)).Error
	} else {
		// Update UserInfo
		// Ensure user exists first? Assuming AppendUser is called before.
		// But UserInfo.SqlAddCredit checks exists.
		var count int64
		s.db.Model(&models.UserInfo{}).Where("Id = ?", userId).Count(&count)
		if count == 0 {
			// Insert with credit
			newUser := models.UserInfo{
				BotUin:  botUin,
				GroupId: groupId,
				Id:      userId,
				Credit:  amount,
			}
			return s.db.Create(&newUser).Error
		} else {
			return s.db.Model(&models.UserInfo{}).Where("Id = ?", userId).
				UpdateColumn("Credit", gorm.Expr("Credit + ?", amount)).Error
		}
	}
}
