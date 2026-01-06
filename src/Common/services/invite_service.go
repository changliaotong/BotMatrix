package services

import (
	"fmt"
	"time"

	"BotMatrix/common/models"

	"gorm.io/gorm"
)

type InviteService struct {
	db          *gorm.DB
	userService *UserService
	listService *ListService
}

func NewInviteService(db *gorm.DB, userService *UserService, listService *ListService) *InviteService {
	return &InviteService{
		db:          db,
		userService: userService,
		listService: listService,
	}
}

// CheckJoinRequest checks if a user can join the group
func (s *InviteService) CheckJoinRequest(groupId int64, userId int64, password string) (int, string) {
	if s.listService.IsSystemBlack(userId) {
		return 0, "黑名单禁入"
	}

	// VIP check (placeholder, assume not VIP for now or implement if GroupVip migrated)
	// if GroupVip.IsClientVip(userId) ...

	var groupInfo models.GroupInfo
	if err := s.db.Where("Id = ?", groupId).First(&groupInfo).Error; err != nil {
		return 0, "群组不存在"
	}

	if groupInfo.IsAcceptNewMember == 3 {
		// Password check
		// Need regex match implementation or simple string check
		// For now simple equality check if RegexRequestJoin is simple
		// Or assume password param contains the message content to check against regex
		// This part depends on regex util.
		return 0, "密码验证未实现" // Placeholder
	}

	return groupInfo.IsAcceptNewMember, groupInfo.RejectMessage
}

// ProcessInvite handles invite credit logic when a user joins
func (s *InviteService) ProcessInvite(botUin int64, groupId int64, groupName string, userId int64, userName string, invitorQQ int64, invitorName string) (string, error) {
	if invitorQQ <= 0 {
		return "", nil
	}

	// 1. Ensure users exist
	err := s.userService.AppendUser(botUin, groupId, userId, userName, "", "")
	if err != nil {
		return "", err
	}
	err = s.userService.AppendUser(botUin, groupId, invitorQQ, invitorName, "", "")
	if err != nil {
		return "", err
	}

	// 2. Update GroupMember (InvitorUserId)
	err = s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).
		Update("InvitorUserId", invitorQQ).Error
	if err != nil {
		return "", err
	}

	// 3. Increment InviteCount for inviter
	err = s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, invitorQQ).
		UpdateColumn("InviteCount", gorm.Expr("InviteCount + ?", 1)).Error
	if err != nil {
		return "", err
	}

	// 4. Handle InviteCredit
	var groupInfo models.GroupInfo
	if err := s.db.Where("Id = ?", groupId).First(&groupInfo).Error; err != nil {
		return "", err
	}

	inviteCredit := groupInfo.InviteCredit
	if inviteCredit > 50 {
		minusCredit := int64(inviteCredit - 50)
		ownerCredit, _ := s.userService.GetCredit(groupId, groupInfo.RobotOwner)
		if ownerCredit >= minusCredit {
			// Deduct from RobotOwner
			s.userService.AddCredit(botUin, groupId, groupInfo.RobotOwner, -minusCredit)
			// Log
			s.logCredit(botUin, groupId, groupName, groupInfo.RobotOwner, groupInfo.RobotOwnerName, -minusCredit, fmt.Sprintf("邀人送分:%d邀请%d", invitorQQ, userId))
		} else {
			inviteCredit = 50
		}
	}

	// Add credit to inviter
	err = s.userService.AddCredit(botUin, groupId, invitorQQ, int64(inviteCredit))
	if err != nil {
		return "", err
	}
	// Log
	s.logCredit(botUin, groupId, groupName, invitorQQ, invitorName, int64(inviteCredit), fmt.Sprintf("邀人送分:邀请%d进群%d", userId, groupId))

	// Get updated invite count and credit
	var inviteCount int
	s.db.Model(&models.GroupMember{}).Select("InviteCount").Where("GroupId = ? AND UserId = ?", groupId, invitorQQ).Scan(&inviteCount)
	currentCredit, _ := s.userService.GetCredit(groupId, invitorQQ)

	answer := fmt.Sprintf("[@:%d] 邀请 [@:%d]进群\n累计已邀请%d人\n积分：+%d，累计：%d",
		invitorQQ, userId, inviteCount, inviteCredit, currentCredit)

	return answer, nil
}

// ProcessExit handles invite count deduction when a user leaves
func (s *InviteService) ProcessExit(groupId int64, userId int64) error {
	var invitorUserId int64
	err := s.db.Model(&models.GroupMember{}).Select("InvitorUserId").Where("GroupId = ? AND UserId = ?", groupId, userId).Scan(&invitorUserId).Error
	if err != nil {
		return err
	}

	if invitorUserId > 0 {
		return s.db.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, invitorUserId).
			UpdateColumn("InviteExitCount", gorm.Expr("InviteExitCount + ?", 1)).Error
	}
	return nil
}

func (s *InviteService) logCredit(botUin int64, groupId int64, groupName string, userId int64, userName string, credit int64, info string) {
	log := models.CreditLog{
		BotUin:      botUin,
		GroupId:     groupId,
		GroupName:   groupName,
		UserId:      userId,
		UserName:    userName,
		CreditAdd:   credit,
		CreditValue: 0, // Value usually not tracked in log insert for performance, or queried
		CreditInfo:  info,
		InsertDate:  time.Now(),
	}
	s.db.Create(&log)
}
