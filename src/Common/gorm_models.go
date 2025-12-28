package common

import (
	"fmt"
	"time"
)

// ==================== GORM 模型定义 ====================

// UserGORM 用户表GORM模型
type UserGORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Username       string    `gorm:"uniqueIndex;not null;size:255"`
	PasswordHash   string    `gorm:"not null;size:255"`
	IsAdmin        bool      `gorm:"default:false"`
	Active         bool      `gorm:"default:true"`
	SessionVersion int       `gorm:"default:1"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName 设置表名
func (UserGORM) TableName() string {
	return "users"
}

// ToUser 转换为普通User结构体
func (u *UserGORM) ToUser() *User {
	return &User{
		ID:             int64(u.ID),
		Username:       u.Username,
		PasswordHash:   u.PasswordHash,
		IsAdmin:        u.IsAdmin,
		Active:         u.Active,
		SessionVersion: u.SessionVersion,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

// FromUser 从普通User结构体转换
func (u *UserGORM) FromUser(user *User) {
	if user.ID > 0 {
		u.ID = uint(user.ID)
	}
	u.Username = user.Username
	u.PasswordHash = user.PasswordHash
	u.IsAdmin = user.IsAdmin
	u.Active = user.Active
	u.SessionVersion = user.SessionVersion
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
}

// RoutingRuleGORM 路由规则表GORM模型
type RoutingRuleGORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Pattern        string    `gorm:"uniqueIndex;not null;size:500"`
	TargetWorkerID string    `gorm:"not null;size:255"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName 设置表名
func (RoutingRuleGORM) TableName() string {
	return "routing_rules"
}

// ToRoutingRule 转换为普通RoutingRule结构体
func (r *RoutingRuleGORM) ToRoutingRule() *RoutingRule {
	return &RoutingRule{
		Pattern:        r.Pattern,
		TargetWorkerID: r.TargetWorkerID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// FromRoutingRule 从普通RoutingRule结构体转换
func (r *RoutingRuleGORM) FromRoutingRule(rule *RoutingRule) {
	r.Pattern = rule.Pattern
	r.TargetWorkerID = rule.TargetWorkerID
	r.CreatedAt = rule.CreatedAt
	r.UpdatedAt = rule.UpdatedAt
}

// GroupCacheGORM 群组缓存表GORM模型
type GroupCacheGORM struct {
	GroupID   string    `gorm:"primaryKey;size:255"`
	GroupName string    `gorm:"size:255"`
	BotID     string    `gorm:"size:255"`
	LastSeen  time.Time
}

// TableName 设置表名
func (GroupCacheGORM) TableName() string {
	return "group_cache"
}

// ToGroupCache 转换为普通GroupCache结构体
func (g *GroupCacheGORM) ToGroupCache() *GroupCache {
	return &GroupCache{
		GroupID:   g.GroupID,
		GroupName: g.GroupName,
		BotID:     g.BotID,
		LastSeen:  g.LastSeen,
	}
}

// FromGroupCache 从普通GroupCache结构体转换
func (g *GroupCacheGORM) FromGroupCache(cache *GroupCache) {
	g.GroupID = cache.GroupID
	g.GroupName = cache.GroupName
	g.BotID = cache.BotID
	g.LastSeen = cache.LastSeen
}

// FriendCacheGORM 好友缓存表GORM模型
type FriendCacheGORM struct {
	UserID   string    `gorm:"primaryKey;size:255"`
	Nickname string    `gorm:"size:255"`
	BotID    string    `gorm:"size:255"`
	LastSeen time.Time
}

// TableName 设置表名
func (FriendCacheGORM) TableName() string {
	return "friend_cache"
}

// ToFriendCache 转换为普通FriendCache结构体
func (f *FriendCacheGORM) ToFriendCache() *FriendCache {
	return &FriendCache{
		UserID:   f.UserID,
		Nickname: f.Nickname,
		BotID:    f.BotID,
		LastSeen: f.LastSeen,
	}
}

// FromFriendCache 从普通FriendCache结构体转换
func (f *FriendCacheGORM) FromFriendCache(cache *FriendCache) {
	f.UserID = cache.UserID
	f.Nickname = cache.Nickname
	f.BotID = cache.BotID
	f.LastSeen = cache.LastSeen
}

// MemberCacheGORM 群成员缓存表GORM模型
type MemberCacheGORM struct {
	GroupID  string    `gorm:"primaryKey;size:255"`
	UserID   string    `gorm:"primaryKey;size:255"`
	Nickname string    `gorm:"size:255"`
	Card     string    `gorm:"size:255"`
	LastSeen time.Time
}

// TableName 设置表名
func (MemberCacheGORM) TableName() string {
	return "member_cache"
}

// ToMemberCache 转换为普通MemberCache结构体
func (m *MemberCacheGORM) ToMemberCache() *MemberCache {
	return &MemberCache{
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		Nickname: m.Nickname,
		Card:     m.Card,
		LastSeen: m.LastSeen,
	}
}

// FromMemberCache 从普通MemberCache结构体转换
func (m *MemberCacheGORM) FromMemberCache(cache *MemberCache) {
	m.GroupID = cache.GroupID
	m.UserID = cache.UserID
	m.Nickname = cache.Nickname
	m.Card = cache.Card
	m.LastSeen = cache.LastSeen
}

// SystemStatGORM 系统统计表GORM模型
type SystemStatGORM struct {
	Key       string    `gorm:"primaryKey;size:255"`
	Value     string    `gorm:"size:1000"`
	UpdatedAt time.Time
}

// TableName 设置表名
func (SystemStatGORM) TableName() string {
	return "system_stats"
}

// ToSystemStat 转换为普通map格式
func (s *SystemStatGORM) ToSystemStat() map[string]interface{} {
	return map[string]interface{}{
		"key":   s.Key,
		"value": s.Value,
	}
}

// FromSystemStat 从普通map格式转换
func (s *SystemStatGORM) FromSystemStat(key string, value interface{}) {
	s.Key = key
	s.Value = fmt.Sprintf("%v", value)
	s.UpdatedAt = time.Now()
}

// GroupStatsGORM 群组统计表GORM模型
type GroupStatsGORM struct {
	ID        string    `gorm:"primaryKey;size:255"`
	Count     int64     `gorm:"default:0"`
	UpdatedAt time.Time
}

// TableName 设置表名
func (GroupStatsGORM) TableName() string {
	return "group_stats"
}

// UserStatsGORM 用户统计表GORM模型
type UserStatsGORM struct {
	ID        string    `gorm:"primaryKey;size:255"`
	Count     int64     `gorm:"default:0"`
	UpdatedAt time.Time
}

// TableName 设置表名
func (UserStatsGORM) TableName() string {
	return "user_stats"
}

// GroupStatsTodayGORM 群组每日统计表GORM模型
type GroupStatsTodayGORM struct {
	ID        string    `gorm:"primaryKey;size:255"`
	Count     int64     `gorm:"default:0"`
	Day       string    `gorm:"size:10"`
	UpdatedAt time.Time
}

// TableName 设置表名
func (GroupStatsTodayGORM) TableName() string {
	return "group_stats_today"
}

// UserStatsTodayGORM 用户每日统计表GORM模型
type UserStatsTodayGORM struct {
	ID        string    `gorm:"primaryKey;size:255"`
	Count     int64     `gorm:"default:0"`
	Day       string    `gorm:"size:10"`
	UpdatedAt time.Time
}

// TableName 设置表名
func (UserStatsTodayGORM) TableName() string {
	return "user_stats_today"
}
