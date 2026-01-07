package store

import (
	"BotMatrix/common/models"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Store provides access to all database operations
type Store struct {
	db *gorm.DB
	// specialized stores
	Users     *UserStore
	Groups    *GroupStore
	Members   *MemberStore
	Logs      *LogStore
	VIPs      *VIPStore
	Blacklist *BlacklistStore
	Whitelist *WhitelistStore
	Sessions  *SessionStore
	Messages  *MessageStore
	Caches    *CacheStore
	Robots    *RobotStore
	Friends   *FriendStore
	Greetings *GreetingStore
	Games     *GameStore
	Incomes   *IncomeStore
	Greylist  *GreylistStore
	Warns     *WarnStore
}

// NewStore creates a new Store instance
func NewStore(db *gorm.DB) *Store {
	s := &Store{db: db}
	s.Users = &UserStore{db: db}
	s.Groups = &GroupStore{db: db}
	s.Members = &MemberStore{db: db}
	s.Logs = &LogStore{db: db}
	s.VIPs = &VIPStore{db: db}
	s.Blacklist = &BlacklistStore{db: db}
	s.Whitelist = &WhitelistStore{db: db}
	s.Sessions = &SessionStore{db: db}
	s.Messages = &MessageStore{db: db}
	s.Caches = &CacheStore{db: db}
	s.Robots = &RobotStore{db: db}
	s.Friends = &FriendStore{db: db}
	s.Greetings = &GreetingStore{db: db}
	s.Games = &GameStore{db: db}
	s.Incomes = &IncomeStore{db: db}
	s.Greylist = &GreylistStore{db: db}
	s.Warns = &WarnStore{db: db}
	return s
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

// RunInTransaction executes the given function within a database transaction
func (s *Store) RunInTransaction(fn func(txStore *Store) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txStore := NewStore(tx)
		return fn(txStore)
	})
}

// UserStore handles user-related database operations
type UserStore struct {
	db *gorm.DB
}

func (s *UserStore) GetByID(id int64) (*models.UserInfo, error) {
	var user models.UserInfo
	err := s.db.First(&user, "Id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetOrCreate(id int64, name string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := s.db.First(&user, "Id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		user = models.UserInfo{
			Id:   id,
			Name: name,
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	return &user, err
}

func (s *UserStore) AddCredit(id int64, amount int64) error {
	return s.db.Model(&models.UserInfo{}).Where("Id = ?", id).Update("Credit", gorm.Expr("Credit + ?", amount)).Error
}

func (s *UserStore) Update(user *models.UserInfo) error {
	// 使用 Select 排除不需要更新的敏感字段，或者改用局部更新
	// 这里改为只更新积分、金币、算力等核心资产字段，避免覆盖其他重要元数据
	return s.db.Model(user).Updates(map[string]interface{}{
		"Credit": user.Credit,
		"Tokens": user.Tokens,
		"Coins":  user.Coins,
		"Name":   user.Name,
	}).Error
}

func (s *UserStore) UpdateField(id int64, field string, value interface{}) error {
	return s.db.Model(&models.UserInfo{}).Where("Id = ?", id).Update(field, value).Error
}

// GroupStore handles group-related database operations
type GroupStore struct {
	db *gorm.DB
}

func (s *GroupStore) GetByID(id int64) (*models.GroupInfo, error) {
	var group models.GroupInfo
	err := s.db.First(&group, "Id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupStore) GetOrCreate(id int64) (*models.GroupInfo, error) {
	var group models.GroupInfo
	err := s.db.First(&group, "Id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		group = models.GroupInfo{
			Id: id,
		}
		if err := s.db.Create(&group).Error; err != nil {
			return nil, err
		}
		return &group, nil
	}
	return &group, err
}

func (s *GroupStore) Update(group *models.GroupInfo) error {
	return s.db.Save(group).Error
}

func (s *GroupStore) GetRobotOwner(groupID int64) (int64, error) {
	var group models.GroupInfo
	err := s.db.Select("RobotOwner").First(&group, "Id = ?", groupID).Error
	return group.RobotOwner, err
}

func (s *GroupStore) UpdateField(id int64, field string, value interface{}) error {
	return s.db.Model(&models.GroupInfo{}).Where("Id = ?", id).Update(field, value).Error
}

func (s *GroupStore) GetAdmins(groupID int64) ([]models.GroupMember, error) {
	var admins []models.GroupMember
	err := s.db.Where("GroupId = ? AND IsAdmin = ?", groupID, true).Find(&admins).Error
	return admins, err
}

// MemberStore handles group member-related database operations
type MemberStore struct {
	db *gorm.DB
}

func (s *MemberStore) Get(groupID, userID int64) (*models.GroupMember, error) {
	var member models.GroupMember
	err := s.db.First(&member, "GroupId = ? AND UserId = ?", groupID, userID).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (s *MemberStore) GetOrCreate(groupID, userID int64, name string) (*models.GroupMember, error) {
	var member models.GroupMember
	err := s.db.First(&member, "GroupId = ? AND UserId = ?", groupID, userID).Error
	if err == gorm.ErrRecordNotFound {
		member = models.GroupMember{
			GroupId:  groupID,
			UserId:   userID,
			UserName: name,
		}
		if err := s.db.Create(&member).Error; err != nil {
			return nil, err
		}
		return &member, nil
	}
	return &member, err
}

func (s *MemberStore) AddCredit(groupID, userID int64, amount int64) error {
	return s.db.Model(&models.GroupMember{}).
		Where("GroupId = ? AND UserId = ?", groupID, userID).
		Update("Credit", gorm.Expr("Credit + ?", amount)).Error
}

func (s *MemberStore) Update(member *models.GroupMember) error {
	return s.db.Model(member).Updates(map[string]interface{}{
		"SignDate":     member.SignDate,
		"SignTimes":    member.SignTimes,
		"SignLevel":    member.SignLevel,
		"SignTimesAll": member.SignTimesAll,
		"Credit":       member.Credit,
		"UserName":     member.UserName,
	}).Error
}

func (s *MemberStore) GetTopByCredit(groupID int64, limit int) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := s.db.Where("GroupId = ?", groupID).Order("Credit DESC").Limit(limit).Find(&members).Error
	return members, err
}

func (s *MemberStore) GetTopByMsgCount(groupID int64, limit int) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := s.db.Where("GroupId = ?", groupID).Order("MsgCount DESC").Limit(limit).Find(&members).Error
	return members, err
}

func (s *MemberStore) GetAll(groupID int64) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := s.db.Where("GroupId = ?", groupID).Find(&members).Error
	return members, err
}

// LogStore handles log-related database operations
type LogStore struct {
	db *gorm.DB
}

func (s *LogStore) AddCreditLog(log *models.CreditLog) error {
	return s.db.Create(log).Error
}

func (s *LogStore) GetCreditLogs(groupID, userID int64, limit int) ([]*models.CreditLog, error) {
	var logs []*models.CreditLog
	err := s.db.Where("GroupId = ? AND UserId = ?", groupID, userID).Order("InsertDate DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (s *LogStore) AddTokensLog(log *models.TokensLog) error {
	return s.db.Create(log).Error
}

// RobotStore handles robot-related database operations
type RobotStore struct {
	db *gorm.DB
}

func (s *RobotStore) Get(botUin int64) (*models.BotInfo, error) {
	var bot models.BotInfo
	err := s.db.First(&bot, "BotUin = ?", botUin).Error
	return &bot, err
}

func (s *RobotStore) IsAdmin(botUin, userID int64) (bool, error) {
	// Super admins from C# BotInfo.cs
	if userID == 51437810 || userID == 1653346663 {
		return true, nil
	}
	bot, err := s.Get(botUin)
	if err != nil {
		return false, err
	}
	return bot.AdminId == userID, nil
}

func (s *RobotStore) Update(bot *models.BotInfo) error {
	return s.db.Save(bot).Error
}

// FriendStore handles friend-related database operations
type FriendStore struct {
	db *gorm.DB
}

func (s *FriendStore) GetByBot(botUin int64) ([]*models.Friend, error) {
	var friends []*models.Friend
	err := s.db.Where("BotUin = ?", botUin).Find(&friends).Error
	return friends, err
}

func (s *FriendStore) Get(botUin, userID int64) (*models.Friend, error) {
	var friend models.Friend
	err := s.db.Where("BotUin = ? AND UserId = ?", botUin, userID).First(&friend).Error
	return &friend, err
}

func (s *FriendStore) AddCredit(botUin, userID int64, amount int64) error {
	return s.db.Model(&models.Friend{}).
		Where("BotUin = ? AND UserId = ?", botUin, userID).
		Update("Credit", gorm.Expr("Credit + ?", amount)).Error
}

// IncomeStore handles income-related database operations
type IncomeStore struct {
	db *gorm.DB
}

func (s *IncomeStore) GetTotal(userID int64) (float64, error) {
	var result struct {
		Total float64
	}
	err := s.db.Model(&models.Income{}).Where("UserId = ?", userID).Select("SUM(IncomeMoney) as total").Scan(&result).Error
	return result.Total, err
}

func (s *LogStore) RecordMessageStat(botUin, groupID, userID int64, name string) error {
	now := time.Now()
	date := now.Truncate(24 * time.Hour)

	// Update MsgCount
	var msgCount models.MsgCount
	err := s.db.Where("BotUin = ? AND GroupId = ? AND UserId = ? AND CDate = ?", botUin, groupID, userID, date).First(&msgCount).Error
	if err == gorm.ErrRecordNotFound {
		msgCount = models.MsgCount{
			BotUin:   botUin,
			GroupId:  groupID,
			UserId:   userID,
			UserName: name,
			CDate:    date.Format("2006-01-02"),
			CMsg:     1,
			MsgDate:  now,
		}
		return s.db.Create(&msgCount).Error
	} else if err == nil {
		return s.db.Model(&msgCount).Updates(map[string]interface{}{
			"CMsg":    gorm.Expr("CMsg + 1"),
			"MsgDate": now,
		}).Error
	}
	return err
}

// GameStore handles game-related database operations
type GameStore struct {
	db *gorm.DB
}

func (s *GameStore) SaveShuffledDeck(groupId int64, deck []models.ShuffledDeck) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Where("GroupId = ?", groupId).Delete(&models.ShuffledDeck{}).Error; err != nil {
			return err
		}
		// Insert new
		if len(deck) > 0 {
			return tx.Create(&deck).Error
		}
		return nil
	})
}

func (s *GameStore) ReadShuffledDeck(groupId int64) ([]models.ShuffledDeck, error) {
	var deck []models.ShuffledDeck
	err := s.db.Where("GroupId = ?", groupId).Order("DeckOrder").Find(&deck).Error
	return deck, err
}

func (s *GameStore) ClearShuffledDeck(groupId int64) error {
	return s.db.Where("GroupId = ?", groupId).Delete(&models.ShuffledDeck{}).Error
}

func (s *GameStore) GetChengyuDict(chengyu string) (*models.ChengyuDict, error) {
	var dict models.ChengyuDict
	// Using simple query. NOTE: baseinfo table might need specific handling if cross-db.
	// But assuming same DB connection can access it or it is in the same DB.
	err := s.db.Where("chengyu = ?", chengyu).First(&dict).Error
	return &dict, err
}

func (s *GameStore) SearchChengyu(keyword string) ([]models.ChengyuDict, error) {
	var dicts []models.ChengyuDict
	err := s.db.Where("chengyu LIKE ?", "%"+keyword+"%").Limit(10).Find(&dicts).Error
	return dicts, err
}

func (s *GameStore) GetLastChengyu(groupId int64) (*models.ChengyuHistory, error) {
	var history models.ChengyuHistory
	err := s.db.Where("GroupId = ? AND GameNo = 1", groupId).Order("Id DESC").First(&history).Error
	// If not found, it means no game started?
	// But "GameNo=1" is the START marker.
	// To get the absolute LAST chengyu (any GameNo), we just order by Id DESC.
	// But we need the last chengyu of the CURRENT game.
	// First find start of current game.

	var startId int64
	err = s.db.Model(&models.ChengyuHistory{}).Select("Id").
		Where("GroupId = ? AND GameNo = 1", groupId).Order("Id DESC").Limit(1).Scan(&startId).Error

	if err != nil {
		return nil, err
	}
	if startId == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Now get last record >= startId
	err = s.db.Where("GroupId = ? AND Id >= ?", groupId, startId).Order("Id DESC").First(&history).Error
	return &history, err
}

func (s *GameStore) AppendChengyu(history *models.ChengyuHistory) error {
	return s.db.Create(history).Error
}

func (s *GameStore) IsChengyuUsed(groupId int64, chengyu string) (bool, error) {
	var startId int64
	err := s.db.Model(&models.ChengyuHistory{}).Select("Id").
		Where("GroupId = ? AND GameNo = 1", groupId).Order("Id DESC").Limit(1).Scan(&startId).Error
	if err != nil {
		return false, err
	}

	var count int64
	err = s.db.Model(&models.ChengyuHistory{}).
		Where("GroupId = ? AND chengyu = ? AND Id >= ?", groupId, chengyu, startId).Count(&count).Error
	return count > 0, err
}

func (s *GameStore) GetRandomChengyuByPinyin(pinyin string) (*models.ChengyuDict, error) {
	var dict models.ChengyuDict
	query := s.db.Model(&models.ChengyuDict{})

	if pinyin != "" {
		// Clean pinyin (remove tone numbers if user provided them, or handle the fact that DB has them)
		// Usually DB pinyin is like "yi1" or "yi".
		// If we use LIKE 'yi%', it matches "yi", "yi1", "yi2" etc.
		query = query.Where("pinyin LIKE ?", pinyin+"%")
	}

	// SQL Server: ORDER BY NEWID()
	// MySQL: ORDER BY RAND()
	// PostgreSQL/SQLite: ORDER BY RANDOM()
	// Let's use a more generic way or detect driver.
	// For now, sticking with RANDOM() as most common in dev environments.
	err := query.Order("RANDOM()").First(&dict).Error
	return &dict, err
}

// VIPStore handles VIP-related database operations
type VIPStore struct {
	db *gorm.DB
}

func (s *VIPStore) GetByGroupID(groupID int64) (*models.GroupVip, error) {
	var vip models.GroupVip
	err := s.db.First(&vip, "GroupId = ?", groupID).Error
	if err != nil {
		return nil, err
	}
	return &vip, nil
}

func (s *VIPStore) IsVIP(groupID int64) bool {
	var count int64
	s.db.Model(&models.GroupVip{}).Where("GroupId = ? AND EndDate > ?", groupID, time.Now()).Count(&count)
	return count > 0
}

// BlacklistStore handles blacklist-related database operations
type BlacklistStore struct {
	db *gorm.DB
}

func (s *BlacklistStore) IsBlacklisted(botUin, groupID, userID int64) bool {
	var count int64
	s.db.Model(&models.BlackList{}).Where("BotUin = ? AND GroupId = ? AND BlackId = ?", botUin, groupID, userID).Count(&count)
	return count > 0
}

// GreetingStore handles greeting-related database operations
type GreetingStore struct {
	db *gorm.DB
}

func (s *GreetingStore) Exists(groupID, userID int64, greetingType int) bool {
	var count int64
	// Calculate LogicalDate logic
	now := time.Now()
	offset := -3
	if greetingType != 0 {
		offset = -5
	}
	logicalTime := now.Add(time.Duration(offset) * time.Hour)
	logicalDate := time.Date(logicalTime.Year(), logicalTime.Month(), logicalTime.Day(), 0, 0, 0, 0, logicalTime.Location())

	s.db.Model(&models.GreetingRecord{}).
		Where("GroupId = ? AND QQ = ? AND GreetingType = ? AND LogicalDate = ?", groupID, userID, greetingType, logicalDate).
		Count(&count)
	return count > 0
}

func (s *GreetingStore) GetGroupCount(groupID int64, greetingType int) int64 {
	var count int64
	// Calculate LogicalDate logic
	now := time.Now()
	offset := -3
	if greetingType != 0 {
		offset = -5
	}
	logicalTime := now.Add(time.Duration(offset) * time.Hour)
	logicalDate := time.Date(logicalTime.Year(), logicalTime.Month(), logicalTime.Day(), 0, 0, 0, 0, logicalTime.Location())

	s.db.Model(&models.GreetingRecord{}).
		Where("GroupId = ? AND GreetingType = ? AND LogicalDate = ?", groupID, greetingType, logicalDate).
		Count(&count)
	return count
}

func (s *GreetingStore) GetGlobalCount(greetingType int) int64 {
	var count int64
	// Calculate LogicalDate logic
	now := time.Now()
	offset := -3
	if greetingType != 0 {
		offset = -5
	}
	logicalTime := now.Add(time.Duration(offset) * time.Hour)
	logicalDate := time.Date(logicalTime.Year(), logicalTime.Month(), logicalTime.Day(), 0, 0, 0, 0, logicalTime.Location())

	s.db.Model(&models.GreetingRecord{}).
		Where("GreetingType = ? AND LogicalDate = ?", greetingType, logicalDate).
		Count(&count)
	return count
}

func (s *GreetingStore) Add(record *models.GreetingRecord) error {
	return s.db.Create(record).Error
}

func (s *BlacklistStore) Add(botUin, groupID, userID int64, info string) error {
	item := models.BlackList{
		BotUin:     botUin,
		GroupId:    groupID,
		BlackID:    userID,
		BlackInfo:  info,
		InsertDate: time.Now(),
	}
	return s.db.Create(&item).Error
}

func (s *BlacklistStore) Remove(botUin, groupID, userID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ? AND BlackId = ?", botUin, groupID, userID).Delete(&models.BlackList{}).Error
}

func (s *BlacklistStore) Clear(botUin, groupID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Delete(&models.BlackList{}).Error
}

func (s *BlacklistStore) Count(botUin, groupID int64) (int64, error) {
	var count int64
	err := s.db.Model(&models.BlackList{}).Where("BotUin = ? AND GroupId = ?", botUin, groupID).Count(&count).Error
	return count, err
}

func (s *BlacklistStore) GetList(botUin, groupID int64, limit int) ([]models.BlackList, error) {
	var list []models.BlackList
	err := s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Order("Id DESC").Limit(limit).Find(&list).Error
	return list, err
}

// WhitelistStore handles whitelist-related database operations
type WhitelistStore struct {
	db *gorm.DB
}

func (s *WhitelistStore) IsWhitelisted(botUin, groupID, userID int64) bool {
	var count int64
	s.db.Model(&models.WhiteList{}).Where("BotUin = ? AND GroupId = ? AND WhiteId = ?", botUin, groupID, userID).Count(&count)
	return count > 0
}

func (s *WhitelistStore) Add(botUin, groupID, userID int64, info string) error {
	item := models.WhiteList{
		BotUin:     botUin,
		GroupId:    groupID,
		WhiteID:    userID,
		WhiteInfo:  info,
		InsertDate: time.Now(),
	}
	return s.db.Create(&item).Error
}

func (s *WhitelistStore) Remove(botUin, groupID, userID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ? AND WhiteId = ?", botUin, groupID, userID).Delete(&models.WhiteList{}).Error
}

func (s *WhitelistStore) Clear(botUin, groupID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Delete(&models.WhiteList{}).Error
}

func (s *WhitelistStore) Count(botUin, groupID int64) (int64, error) {
	var count int64
	err := s.db.Model(&models.WhiteList{}).Where("BotUin = ? AND GroupId = ?", botUin, groupID).Count(&count).Error
	return count, err
}

func (s *WhitelistStore) GetList(botUin, groupID int64, limit int) ([]models.WhiteList, error) {
	var list []models.WhiteList
	err := s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Order("Id DESC").Limit(limit).Find(&list).Error
	return list, err
}

// SessionStore handles session-related database operations
type SessionStore struct {
	db *gorm.DB
}

func (s *SessionStore) Get(sessionID string) (*models.Session, error) {
	var session models.Session
	err := s.db.First(&session, "SessionId = ?", sessionID).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionStore) Set(sessionID string, userID, groupID int64, state string, data interface{}) error {
	dataJSON := ""
	if data != nil {
		if b, err := json.Marshal(data); err == nil {
			dataJSON = string(b)
		}
	}

	var session models.Session
	err := s.db.Where("SessionId = ?", sessionID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		session = models.Session{
			SessionId: sessionID,
			UserId:    userID,
			GroupId:   groupID,
			State:     state,
			Data:      dataJSON,
		}
		return s.db.Create(&session).Error
	} else if err == nil {
		return s.db.Model(&session).Updates(map[string]interface{}{
			"State":     state,
			"Data":      dataJSON,
			"UpdatedAt": time.Now(),
		}).Error
	}
	return err
}

func (s *SessionStore) Delete(sessionID string) error {
	return s.db.Where("SessionId = ?", sessionID).Delete(&models.Session{}).Error
}

// MessageStore handles message-related database operations
type MessageStore struct {
	db *gorm.DB
}

func (s *MessageStore) LogMessage(log *models.MessageLog) error {
	return s.db.Create(log).Error
}

func (s *MessageStore) GetRecentLogs(groupID string, limit int) ([]models.MessageLog, error) {
	var logs []models.MessageLog
	err := s.db.Where("GroupId = ?", groupID).Order("CreatedAt DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (s *MessageStore) UpdateStat(groupID, userID string, date time.Time, count int64) error {
	dateStr := date.Format("2006-01-02")
	var stat models.MessageStat
	err := s.db.Where("GroupId = ? AND UserId = ? AND Date = ?", groupID, userID, dateStr).First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		stat = models.MessageStat{
			GroupID: groupID,
			UserID:  userID,
			Date:    date,
			Count:   count,
		}
		return s.db.Create(&stat).Error
	} else if err == nil {
		return s.db.Model(&stat).Update("Count", gorm.Expr("Count + ?", count)).Error
	}
	return err
}

func (s *MessageStore) GetTopStats(groupID string, date string, limit int) ([]models.MessageStat, error) {
	var stats []models.MessageStat
	err := s.db.Where("GroupId = ? AND Date = ?", groupID, date).Order("Count DESC").Limit(limit).Find(&stats).Error
	return stats, err
}

func (s *MessageStore) GetStatsRange(groupID string, startDate, endDate string, limit int) ([]models.MessageStat, error) {
	var stats []models.MessageStat
	err := s.db.Model(&models.MessageStat{}).
		Select("UserId, SUM(Count) as Count").
		Where("GroupId = ? AND Date >= ? AND Date <= ?", groupID, startDate, endDate).
		Group("UserId").
		Order("Count DESC").
		Limit(limit).
		Scan(&stats).Error
	return stats, err
}

// CacheStore handles cache-related database operations
type CacheStore struct {
	db *gorm.DB
}

func (s *CacheStore) GetMember(groupID, userID string) (*models.MemberCache, error) {
	var member models.MemberCache
	err := s.db.Where("GroupId = ? AND UserId = ?", groupID, userID).First(&member).Error
	return &member, err
}

func (s *CacheStore) GetGroup(groupID string) (*models.GroupCache, error) {
	var group models.GroupCache
	err := s.db.Where("GroupId = ?", groupID).First(&group).Error
	return &group, err
}

func (s *CacheStore) UpdateGroupCache(group *models.GroupCache) error {
	var existing models.GroupCache
	err := s.db.Where("GroupId = ?", group.GroupID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.Create(group).Error
	} else if err == nil {
		return s.db.Model(&existing).Updates(group).Error
	}
	return err
}

func (s *CacheStore) UpdateMemberCache(member *models.MemberCache) error {
	var existing models.MemberCache
	err := s.db.Where("GroupId = ? AND UserId = ?", member.GroupID, member.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.Create(member).Error
	} else if err == nil {
		return s.db.Model(&existing).Updates(member).Error
	}
	return err
}

// GreylistStore handles greylist-related database operations
type GreylistStore struct {
	db *gorm.DB
}

func (s *GreylistStore) Add(botUin, groupID, userID int64, info string) error {
	item := models.GreyList{
		BotUin:     botUin,
		GroupId:    groupID,
		GreyID:     userID,
		GreyInfo:   info,
		InsertDate: time.Now(),
	}
	return s.db.Create(&item).Error
}

func (s *GreylistStore) Remove(botUin, groupID, userID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ? AND GreyId = ?", botUin, groupID, userID).Delete(&models.GreyList{}).Error
}

func (s *GreylistStore) Clear(botUin, groupID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Delete(&models.GreyList{}).Error
}

func (s *GreylistStore) Count(botUin, groupID int64) (int64, error) {
	var count int64
	err := s.db.Model(&models.GreyList{}).Where("BotUin = ? AND GroupId = ?", botUin, groupID).Count(&count).Error
	return count, err
}

func (s *GreylistStore) GetList(botUin, groupID int64, limit int) ([]models.GreyList, error) {
	var list []models.GreyList
	err := s.db.Where("BotUin = ? AND GroupId = ?", botUin, groupID).Order("Id DESC").Limit(limit).Find(&list).Error
	return list, err
}

// WarnStore handles warn-related database operations
type WarnStore struct {
	db *gorm.DB
}

func (s *WarnStore) Add(botUin, groupID, userID, insertBy int64, info string) error {
	item := models.Warn{
		BotUin:     botUin,
		GroupId:    groupID,
		UserId:     userID,
		WarnInfo:   info,
		InsertBy:   insertBy,
		InsertDate: time.Now(),
	}
	return s.db.Create(&item).Error
}

func (s *WarnStore) Clear(groupID, userID int64) error {
	return s.db.Where("GroupId = ? AND UserId = ?", groupID, userID).Delete(&models.Warn{}).Error
}

func (s *WarnStore) Count(groupID, userID int64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Warn{}).Where("GroupId = ? AND UserId = ?", groupID, userID).Count(&count).Error
	return count, err
}
