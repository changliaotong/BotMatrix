package services

import (
	"fmt"
	"time"

	"BotMatrix/common/models"

	"gorm.io/gorm"
)

type GiftService struct {
	db *gorm.DB
}

func NewGiftService(db *gorm.DB) *GiftService {
	return &GiftService{db: db}
}

// SendGift handles the gift sending logic
func (s *GiftService) SendGift(botUin int64, groupId int64, groupName string, userId int64, userName string, targetId int64, giftName string, giftCount int) (string, error) {
	var resultMsg string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find Gift
		var gift models.Gift
		if err := tx.Where("GiftName = ?", giftName).First(&gift).Error; err != nil {
			return fmt.Errorf("不存在此礼物")
		}

		// 2. Calculate Costs
		totalCost := gift.GiftCredit * int64(giftCount)
		creditAdd := totalCost / 2
		creditOwner := creditAdd / 2

		// 3. Check Balance
		var user models.UserInfo
		if err := tx.Where("Id = ?", userId).First(&user).Error; err != nil {
			return fmt.Errorf("未找到用户信息")
		}
		if user.Credit < totalCost {
			return fmt.Errorf("您的积分%d不足%d", user.Credit, totalCost)
		}

		// 4. Get Owner
		var group models.GroupInfo
		tx.Where("Id = ?", groupId).First(&group)
		ownerId := group.RobotOwner
		var ownerName string
		var ownerUser models.UserInfo
		if tx.Where("Id = ?", ownerId).First(&ownerUser).Error == nil {
			ownerName = ownerUser.Name
		}

		// 5. Ensure Target
		var targetUser models.UserInfo
		if err := tx.Where("Id = ?", targetId).First(&targetUser).Error; err != nil {
			targetUser = models.UserInfo{Id: targetId}
			tx.Create(&targetUser)
		}

		// 6. Execute Updates
		// Deduct Sender
		if err := tx.Model(&models.UserInfo{}).Where("Id = ?", userId).UpdateColumn("Credit", gorm.Expr("Credit - ?", totalCost)).Error; err != nil {
			return err
		}
		tx.Create(&models.CreditLog{BotUin: botUin, GroupId: groupId, GroupName: groupName, UserId: userId, UserName: userName, CreditAdd: -totalCost, CreditInfo: "礼物扣分"})

		// Log Tokens change
		tx.Create(&models.TokensLog{BotUin: botUin, GroupId: groupId, GroupName: groupName, UserId: userId, UserName: userName, TokensAdd: -totalCost, TokensInfo: "礼物扣算力"})

		// Add Receiver
		if err := tx.Model(&models.UserInfo{}).Where("Id = ?", targetId).UpdateColumn("Credit", gorm.Expr("Credit + ?", creditAdd)).Error; err != nil {
			return err
		}
		tx.Create(&models.CreditLog{BotUin: botUin, GroupId: groupId, GroupName: groupName, UserId: targetId, UserName: "", CreditAdd: creditAdd, CreditInfo: "礼物加分"})

		// Add Owner
		if ownerId > 0 {
			tx.Model(&models.UserInfo{}).Where("Id = ?", ownerId).UpdateColumn("Credit", gorm.Expr("Credit + ?", creditOwner))
			tx.Create(&models.CreditLog{BotUin: botUin, GroupId: groupId, GroupName: groupName, UserId: ownerId, UserName: ownerName, CreditAdd: creditOwner, CreditInfo: "礼物加分"})
		}

		// Update Fans Value (Sender)
		fansAdd := totalCost / 20
		tx.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).UpdateColumn("FansValue", gorm.Expr("FansValue + ?", fansAdd))

		// Log Gift
		tx.Create(&models.GiftLog{
			BotUin:       botUin,
			GroupId:      groupId,
			GroupName:    groupName,
			UserId:       userId,
			UserName:     userName,
			GiftUserId:   targetId,
			GiftUserName: "",
			GiftId:       gift.Id,
			GiftName:     gift.GiftName,
			GiftCount:    giftCount,
			GiftCredit:   gift.GiftCredit,
		})

		resultMsg = fmt.Sprintf("赠送成功！亲密度+%d", fansAdd)
		return nil
	})

	return resultMsg, err
}

// JoinFans handles joining the fans club
func (s *GiftService) JoinFans(groupId int64, userId int64) error {
	var count int64
	s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).Count(&count)

	updates := map[string]interface{}{
		"IsFans":    true,
		"FansDate":  time.Now(),
		"FansLevel": 1,
		"FansValue": 100,
	}

	if count > 0 {
		return s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).Updates(updates).Error
	}

	// Should create if not exists? Usually JoinFans implies user is in group.
	// But to be safe, we can try to create.
	member := models.GroupMember{
		GroupId:   groupId,
		UserId:    userId,
		IsFans:    true,
		FansDate:  time.Now(),
		FansLevel: 1,
		FansValue: 100,
	}
	return s.db.Create(&member).Error
}

// LightLamp handles lighting the fans lamp
func (s *GiftService) LightLamp(groupId int64, userId int64) error {
	return s.db.Model(&models.GroupMember{}).
		Where("GroupId = ? AND UserId = ?", groupId, userId).
		Updates(map[string]interface{}{
			"LampDate":  time.Now(),
			"FansValue": gorm.Expr("FansValue + ?", 10),
		}).Error
}

// GetFansList returns top N fans
func (s *GiftService) GetFansList(groupId int64, top int) ([]models.GroupMember, error) {
	var list []models.GroupMember
	err := s.db.Where("GroupId = ? AND IsFans = ?", groupId, true).
		Order("FansValue desc").
		Limit(top).
		Find(&list).Error
	return list, err
}
