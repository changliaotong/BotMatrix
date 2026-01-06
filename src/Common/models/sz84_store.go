package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Sz84Store struct {
	db *gorm.DB
}

func NewSz84Store(db *gorm.DB) *Sz84Store {
	return &Sz84Store{db: db}
}

// Helpers moved from sz84_models.go

func (s *Sz84Store) GetGroup(groupId int64) *GroupInfo {
	var group GroupInfo
	if err := s.db.Where("Id = ?", groupId).First(&group).Error; err != nil {
		return nil
	}
	return &group
}

func (s *Sz84Store) GetUser(userId int64) *UserInfo {
	var user UserInfo
	if err := s.db.Where("Id = ?", userId).First(&user).Error; err != nil {
		return nil
	}
	return &user
}

func (s *Sz84Store) GetMember(groupId int64, userId int64) *GroupMember {
	var member GroupMember
	if err := s.db.Where("GroupId = ? AND UserId = ?", groupId, userId).First(&member).Error; err != nil {
		return nil
	}
	return &member
}

func (s *Sz84Store) SaveGroup(group *GroupInfo) error {
	return s.db.Save(group).Error
}

func (s *Sz84Store) SaveUser(user *UserInfo) error {
	return s.db.Save(user).Error
}

func (s *Sz84Store) SaveMember(member *GroupMember) error {
	return s.db.Save(member).Error
}

// CoinsType enum
const (
	CoinsType_Coins       = 0
	CoinsType_GoldCoins   = 1
	CoinsType_PurpleCoins = 2
	CoinsType_BlackCoins  = 3
	CoinsType_GameCoins   = 4
	CoinsType_GroupCredit = 5
)

// AddCoins logic
func (s *Sz84Store) AddCoins(botUin int64, groupId int64, groupName string, userId int64, name string, coinsType int, coinsAdd int64, coinsInfo string) (int64, error) {
	var finalCoins int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Determine target table and field
		var tableName string
		var fieldName string

		switch coinsType {
		case CoinsType_Coins:
			tableName = "User"
			fieldName = "Coins"
		case CoinsType_GoldCoins:
			tableName = "GroupMember"
			fieldName = "GoldCoins"
		case CoinsType_PurpleCoins:
			tableName = "GroupMember"
			fieldName = "PurpleCoins"
		case CoinsType_BlackCoins:
			tableName = "GroupMember"
			fieldName = "BlackCoins"
		case CoinsType_GameCoins:
			tableName = "GroupMember"
			fieldName = "GameCoins"
		case CoinsType_GroupCredit:
			tableName = "GroupMember"
			fieldName = "GroupCredit"
		default:
			return fmt.Errorf("unknown coins type: %d", coinsType)
		}

		// 2. Update value
		if tableName == "User" {
			var user UserInfo
			// Use FirstOrCreate to ensure user exists
			if err := tx.Where("Id = ?", userId).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					user = UserInfo{
						Id:   userId,
						Name: name,
					}
					if err := tx.Create(&user).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Model(&user).UpdateColumn(fieldName, gorm.Expr(fieldName+" + ?", coinsAdd)).Error; err != nil {
				return err
			}

			// Get updated value
			if err := tx.Where("Id = ?", userId).First(&user).Error; err != nil {
				return err
			}
			finalCoins = user.Coins

		} else { // GroupMember
			var member GroupMember
			// Use FirstOrCreate to ensure member exists
			if err := tx.Where("GroupId = ? AND UserId = ?", groupId, userId).First(&member).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					member = GroupMember{
						GroupId:  groupId,
						UserId:   userId,
						UserName: name,
					}
					if err := tx.Create(&member).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Model(&member).UpdateColumn(fieldName, gorm.Expr(fieldName+" + ?", coinsAdd)).Error; err != nil {
				return err
			}

			// Get updated value
			if err := tx.Where("GroupId = ? AND UserId = ?", groupId, userId).First(&member).Error; err != nil {
				return err
			}

			switch coinsType {
			case CoinsType_GoldCoins:
				finalCoins = member.GoldCoins
			case CoinsType_PurpleCoins:
				finalCoins = member.PurpleCoins
			case CoinsType_BlackCoins:
				finalCoins = member.BlackCoins
			case CoinsType_GameCoins:
				finalCoins = member.GameCoins
			case CoinsType_GroupCredit:
				finalCoins = member.GroupCredit
			}
		}

		// 3. Log to CoinsLog
		log := CoinsLog{
			BotUin:     botUin,
			GroupId:    groupId,
			GroupName:  groupName,
			UserId:     userId,
			UserName:   name,
			CoinsType:  coinsType,
			CoinsAdd:   coinsAdd,
			CoinsValue: finalCoins,
			CoinsInfo:  coinsInfo,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})

	return finalCoins, err
}
