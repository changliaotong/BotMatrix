package models

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// SigninResult represents the result of a sign-in operation
type SigninResult struct {
	Success       bool
	Message       string
	CreditAdd     int64
	TokensAdd     int64
	NewCredit     int64
	NewTokens     int64
	SignTimes     int
	SignTimesAll  int
	SignLevel     int
	AlreadySigned bool
}

// SigninService handles the replicated C# sign-in logic
type SigninService struct {
	store *Sz84Store
}

func NewSigninService(store *Sz84Store) *SigninService {
	return &SigninService{store: store}
}

// TrySignIn replicates the C# BotMessage.TrySignIn logic
func (s *SigninService) TrySignIn(botUin int64, groupID int64, groupName string, userID int64, userName string, isAuto bool) (*SigninResult, error) {
	db := s.store.db

	// 1. Get Group Info
	var group Group
	if err := db.Where("GroupId = ?", groupID).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If group doesn't exist, we might want to create it or handle default
			group = Group{GroupID: groupID, IsCredit: true, IsAutoSignin: true}
		} else {
			return nil, err
		}
	}

	if isAuto && !group.IsAutoSignin {
		return &SigninResult{Success: false, Message: ""}, nil
	}

	// 1.5 Get member info to check today's status
	var member GroupMember
	err := db.Where("GroupId = ? AND UserId = ?", groupID, userID).First(&member).Error

	// Check if already signed in today
	now := time.Now()
	if err == nil && member.SignDate != nil {
		if s.getDateDiffInDays(*member.SignDate, now) == 0 {
			// Already signed in today
			var user User
			db.Where("Id = ?", userID).First(&user)
			res := s.buildSignedMessage(&member, &user, true, group.IsCredit)
			return &SigninResult{
				Success:       true,
				AlreadySigned: true,
				Message:       res,
				SignTimes:     member.SignTimes,
				SignTimesAll:  member.SignTimesAll,
				SignLevel:     member.SignLevel,
			}, nil
		}
	}

	// 2. Ensure Member exists and get info (Handle first-time users)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Add member if not exists
			member = GroupMember{
				GroupID:     groupID,
				UserID:      userID,
				UserName:    userName,
				GroupCredit: 50,
			}
			if err := db.Create(&member).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 4. Calculate Streak and Level
	newSignTimes := 1
	newSignLevel := 1

	if member.SignDate != nil {
		daysDiff := s.getDateDiffInDays(*member.SignDate, now)
		if daysDiff <= 1 {
			newSignTimes = member.SignTimes + 1
			newSignLevel = s.calculateSignLevel(newSignTimes)
		}
	}

	creditAdd := int64(newSignLevel * 50)

	// 5. Get User Info for Super status and global credits
	var user User
	if err := db.Where("Id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create user if not exists
			user = User{
				ID:      userID,
				Name:    userName,
				Credit:  50,
				BotUin:  botUin,
				GroupID: groupID,
			}
			if err := db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if user.IsSuper {
		creditAdd *= 2
	}

	tokensAdd := creditAdd

	// 6. Transactional Update
	err = db.Transaction(func(tx *gorm.DB) error {
		// Update GroupMember
		if err := tx.Model(&member).Updates(map[string]interface{}{
			"SignDate":     now,
			"SignTimes":    newSignTimes,
			"SignLevel":    newSignLevel,
			"SignTimesAll": member.SignTimesAll + 1,
			"GroupCredit":  member.GroupCredit + creditAdd,
		}).Error; err != nil {
			return err
		}

		// Update User Info
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"Credit": user.Credit + creditAdd,
			"Tokens": user.Tokens + tokensAdd,
		}).Error; err != nil {
			return err
		}

		// Log in RobotWeibo
		log := RobotWeibo{
			RobotQQ:    botUin,
			WeiboQQ:    userID,
			WeiboInfo:  "", // CmdPara in C#
			WeiboType:  1,
			GroupID:    groupID,
			InsertDate: now,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		// Log Credit change
		if group.IsCredit {
			creditLog := CreditLog{
				BotUin:      botUin,
				GroupID:     groupID,
				GroupName:   groupName,
				UserID:      userID,
				UserName:    userName,
				CreditAdd:   creditAdd,
				CreditValue: user.Credit + creditAdd,
				CreditInfo:  "签到加分",
				InsertDate:  now,
			}
			if err := tx.Create(&creditLog).Error; err != nil {
				return err
			}
		}

		// Log Tokens change
		tokensLog := TokensLog{
			BotUin:      botUin,
			GroupID:     groupID,
			GroupName:   groupName,
			UserID:      userID,
			UserName:    userName,
			TokensAdd:   tokensAdd,
			TokensValue: user.Tokens + tokensAdd,
			TokensInfo:  "签到加算力",
			InsertDate:  now,
		}
		if err := tx.Create(&tokensLog).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %v", err)
	}

	// 8. Build Response
	resultMsg := fmt.Sprintf("✅ %s签到成功！\n", func() string {
		if isAuto {
			return "自动"
		}
		return ""
	}())
	if group.IsCredit {
		resultMsg += fmt.Sprintf("💎 积分：+%d→%d\n", creditAdd, user.Credit+creditAdd)
	}

	// Update local objects for accurate message building
	member.SignTimes = newSignTimes
	member.SignLevel = newSignLevel
	member.SignTimesAll += 1
	user.Credit += creditAdd
	user.Tokens += tokensAdd

	resultMsg += s.buildSignedMessage(&member, &user, false, group.IsCredit)

	return &SigninResult{
		Success:      true,
		Message:      resultMsg,
		CreditAdd:    creditAdd,
		TokensAdd:    tokensAdd,
		NewCredit:    user.Credit,
		NewTokens:    user.Tokens,
		SignTimes:    newSignTimes,
		SignTimesAll: member.SignTimesAll,
		SignLevel:    newSignLevel,
	}, nil
}

func (s *SigninService) calculateSignLevel(days int) int {
	switch {
	case days >= 230:
		return 10
	case days >= 170:
		return 9
	case days >= 120:
		return 8
	case days >= 80:
		return 7
	case days >= 50:
		return 6
	case days >= 30:
		return 5
	case days >= 14:
		return 4
	case days >= 7:
		return 3
	case days >= 3:
		return 2
	default:
		return 1
	}
}

func (s *SigninService) getDateDiffInDays(t1, t2 time.Time) int {
	// Normalize to midnight for day difference
	d1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	d2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
	diff := d2.Sub(d1).Hours() / 24
	return int(math.Abs(diff))
}

func (s *SigninService) buildSignedMessage(member *GroupMember, user *User, alreadySigned bool, isCreditSystem bool) string {
	res := ""
	if alreadySigned {
		res = "✅ 今天签过了，明天再来！\n"
		if isCreditSystem {
			res += fmt.Sprintf("💎 积分：%d\n", user.Credit)
		}
	}

	signTimes := member.SignTimes
	signLevel := member.SignLevel
	signTimesAll := member.SignTimesAll

	nextLevelDays := 0
	switch signLevel {
	case 10:
		nextLevelDays = 0
	case 9:
		nextLevelDays = 230 - signTimes
	case 8:
		nextLevelDays = 170 - signTimes
	case 7:
		nextLevelDays = 120 - signTimes
	case 6:
		nextLevelDays = 80 - signTimes
	case 5:
		nextLevelDays = 50 - signTimes
	case 4:
		nextLevelDays = 30 - signTimes
	case 3:
		nextLevelDays = 14 - signTimes
	case 2:
		nextLevelDays = 7 - signTimes
	case 1:
		nextLevelDays = 3 - signTimes
	}

	if isCreditSystem {
		groupRank := s.getCreditRanking(member.GroupID, user.Credit)
		worldRank := s.getCreditRankingAll(user.Credit + user.SaveCredit)
		res += fmt.Sprintf("🏆 积分排名：本群%d 世界%d\n", groupRank, worldRank)
	}

	res += fmt.Sprintf("📅 签到天数：连签%d 累计%d ✨\n", signTimes, signTimesAll)

	todayMsgCount := s.getMsgCount(member.GroupID, member.UserID, 0)
	yesterdayMsgCount := s.getMsgCount(member.GroupID, member.UserID, 1)
	res += fmt.Sprintf("🗣️ 发言次数：今天%d 昨天%d\n", todayMsgCount, yesterdayMsgCount)

	todaySignCount := s.getTodaySignCount(member.GroupID)
	yesterdaySignCount := s.getYesterdaySignCount(member.GroupID)
	res += fmt.Sprintf("👥 签到人次：今天%d 昨日%d", todaySignCount, yesterdaySignCount)

	if nextLevelDays > 0 {
		res += fmt.Sprintf("\n📈 升级进度：还差%d天升级", nextLevelDays)
	}

	return res
}

func (s *SigninService) getCreditRanking(groupID int64, credit int64) int64 {
	var count int64
	db := s.store.db
	db.Model(&GroupMember{}).Where("GroupId = ? AND GroupCredit > ?", groupID, credit).Count(&count)
	return count + 1
}

func (s *SigninService) getCreditRankingAll(totalCredit int64) int64 {
	var count int64
	db := s.store.db
	db.Model(&User{}).Where("Credit + SaveCredit > ?", totalCredit).Count(&count)
	return count + 1
}

func (s *SigninService) getTodaySignCount(groupID int64) int64 {
	var count int64
	db := s.store.db
	db.Model(&RobotWeibo{}).Where("GroupId = ? AND WeiboType = 1 AND "+s.store.GetDateCondition("InsertDate", 0), groupID).Count(&count)
	return count
}

func (s *SigninService) getYesterdaySignCount(groupID int64) int64 {
	var count int64
	db := s.store.db
	db.Model(&RobotWeibo{}).Where("GroupId = ? AND WeiboType = 1 AND "+s.store.GetDateCondition("InsertDate", 1), groupID).Count(&count)
	return count
}

func (s *SigninService) getMsgCount(groupID int64, userID int64, daysAgo int) int {
	var count int
	db := s.store.db
	db.Model(&MsgCount{}).
		Where("GroupId = ? AND UserId = ? AND "+s.store.GetDateCondition("CDate", daysAgo), groupID, userID).
		Select("CMsg").
		Scan(&count)
	return count
}
