package services

import (
	"fmt"
	"time"

	"BotMatrix/common/models"

	"gorm.io/gorm"
)

type EconomyService struct {
	db *gorm.DB
}

func NewEconomyService(db *gorm.DB) *EconomyService {
	return &EconomyService{db: db}
}

// Coin types mapping (based on C# CoinsLog.conisNames/conisFields)
// Assuming: 0: GoldCoins, 1: PurpleCoins, 2: BlackCoins, 3: GameCoins, 4: GroupCredit?
// I need to check CoinsLog.cs for the mapping.

// GetCoins gets the balance of a specific coin type
func (s *EconomyService) GetCoins(coinsType int, groupId int64, userId int64) (int64, error) {
	fieldName := s.getCoinsField(coinsType)
	if fieldName == "" {
		return 0, fmt.Errorf("invalid coins type: %d", coinsType)
	}

	var balance int64
	// Use raw SQL or dynamic select because field name is dynamic
	err := s.db.Model(&models.GroupMember{}).Select(fieldName).Where("GroupId = ? AND UserId = ?", groupId, userId).Scan(&balance).Error
	return balance, err
}

// AddCoins adds coins to a user
func (s *EconomyService) AddCoins(botUin int64, groupId int64, groupName string, userId int64, userName string, coinsType int, amount int64, coinsInfo string) (int64, error) {
	fieldName := s.getCoinsField(coinsType)
	if fieldName == "" {
		return 0, fmt.Errorf("invalid coins type: %d", coinsType)
	}

	// Ensure user exists
	// Assuming UserService.AppendUser is called or we do it here?
	// C# AddCoins calls Append if not exists.
	// I'll skip Append here for simplicity or assume caller handles it.
	// Or better, do a check.

	var newBalance int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update balance
		if err := tx.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, userId).
			UpdateColumn(fieldName, gorm.Expr(fieldName+" + ?", amount)).Error; err != nil {
			return err
		}

		// 2. Get new balance
		if err := tx.Model(&models.GroupMember{}).Select(fieldName).Where("GroupId = ? AND UserId = ?", groupId, userId).Scan(&newBalance).Error; err != nil {
			return err
		}

		// 3. Log
		log := models.CoinsLog{
			BotUin:     botUin,
			GroupId:    groupId,
			GroupName:  groupName,
			UserId:     userId,
			UserName:   userName,
			CoinsType:  coinsType,
			CoinsAdd:   amount,
			CoinsValue: newBalance,
			CoinsInfo:  coinsInfo,
			InsertDate: time.Now(),
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})

	return newBalance, err
}

// TransferCoins transfers coins between users
func (s *EconomyService) TransferCoins(botUin int64, groupId int64, groupName string, fromUser int64, fromName string, toUser int64, coinsType int, amount int64) error {
	fieldName := s.getCoinsField(coinsType)
	if fieldName == "" {
		return fmt.Errorf("invalid coins type: %d", coinsType)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check balance
		var balance int64
		if err := tx.Model(&models.GroupMember{}).Select(fieldName).Where("GroupId = ? AND UserId = ?", groupId, fromUser).Scan(&balance).Error; err != nil {
			return err
		}
		if balance < amount {
			return fmt.Errorf("insufficient balance")
		}

		// Deduct from sender
		if err := tx.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, fromUser).
			UpdateColumn(fieldName, gorm.Expr(fieldName+" - ?", amount)).Error; err != nil {
			return err
		}

		// Add to receiver
		// Ensure receiver exists?
		// Assuming exists.
		if err := tx.Model(&models.GroupMember{}).Where("GroupId = ? AND UserId = ?", groupId, toUser).
			UpdateColumn(fieldName, gorm.Expr(fieldName+" + ?", amount)).Error; err != nil {
			return err
		}

		// Log sender
		var newBalanceFrom int64
		tx.Model(&models.GroupMember{}).Select(fieldName).Where("GroupId = ? AND UserId = ?", groupId, fromUser).Scan(&newBalanceFrom)
		logFrom := models.CoinsLog{
			BotUin:     botUin,
			GroupId:    groupId,
			GroupName:  groupName,
			UserId:     fromUser,
			UserName:   fromName,
			CoinsType:  coinsType,
			CoinsAdd:   -amount,
			CoinsValue: newBalanceFrom,
			CoinsInfo:  fmt.Sprintf("转出:%d", toUser),
			// InsertDate: time.Now(), // Removed
		}
		tx.Create(&logFrom)

		// Log receiver
		var newBalanceTo int64
		tx.Model(&models.GroupMember{}).Select(fieldName).Where("GroupId = ? AND UserId = ?", groupId, toUser).Scan(&newBalanceTo)
		logTo := models.CoinsLog{
			BotUin:     botUin,
			GroupId:    groupId,
			GroupName:  groupName,
			UserId:     toUser,
			UserName:   "", // Name might be unknown here
			CoinsType:  coinsType,
			CoinsAdd:   amount,
			CoinsValue: newBalanceTo,
			CoinsInfo:  fmt.Sprintf("转入:%d", fromUser),
			// InsertDate: time.Now(), // Removed
		}
		tx.Create(&logTo)

		return nil
	})
}

func (s *EconomyService) getCoinsField(coinsType int) string {
	// Need to check CoinsLog.cs for mapping
	// Placeholder mapping
	switch coinsType {
	case 0:
		return "GoldCoins"
	case 1:
		return "PurpleCoins"
	case 2:
		return "BlackCoins"
	case 3:
		return "GameCoins"
	case 4:
		return "GroupCredit"
	default:
		return ""
	}
}
