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
	Sessions  *SessionStore
	Messages  *MessageStore
	Caches    *CacheStore
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
	s.Sessions = &SessionStore{db: db}
	s.Messages = &MessageStore{db: db}
	s.Caches = &CacheStore{db: db}
	return s
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
	return s.db.Save(user).Error
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
		Update("GroupCredit", gorm.Expr("GroupCredit + ?", amount)).Error
}

func (s *MemberStore) Update(member *models.GroupMember) error {
	return s.db.Save(member).Error
}

func (s *MemberStore) GetTopByCredit(groupID int64, limit int) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := s.db.Where("GroupId = ?", groupID).Order("GroupCredit DESC").Limit(limit).Find(&members).Error
	return members, err
}

func (s *MemberStore) GetTopByMsgCount(groupID int64, limit int) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := s.db.Where("GroupId = ?", groupID).Order("MsgCount DESC").Limit(limit).Find(&members).Error
	return members, err
}

// LogStore handles log-related database operations
type LogStore struct {
	db *gorm.DB
}

func (s *LogStore) AddCreditLog(log *models.CreditLog) error {
	return s.db.Create(log).Error
}

func (s *LogStore) AddTokensLog(log *models.TokensLog) error {
	return s.db.Create(log).Error
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
			GroupID:  groupID,
			UserID:   userID,
			UserName: name,
			CDate:    date,
			CMsg:     1,
			MsgDate:  &now,
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

func (s *BlacklistStore) Add(botUin, groupID, userID int64, info string) error {
	item := models.BlackList{
		BotUin:     botUin,
		GroupID:    groupID,
		BlackID:    userID,
		BlackInfo:  info,
		InsertDate: time.Now(),
	}
	return s.db.Create(&item).Error
}

func (s *BlacklistStore) Remove(botUin, groupID, userID int64) error {
	return s.db.Where("BotUin = ? AND GroupId = ? AND BlackId = ?", botUin, groupID, userID).Delete(&models.BlackList{}).Error
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
