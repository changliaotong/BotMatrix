package common

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMManager GORM数据库管理器
type GORMManager struct {
	DB *gorm.DB
}

// NewGORMManager 创建新的GORM管理器
func NewGORMManager() *GORMManager {
	return &GORMManager{}
}

// InitGORM 初始化GORM数据库连接
func (gm *GORMManager) InitGORM() error {
	var err error
	var db *gorm.DB

	// 配置GORM日志级别
	logLevel := logger.Silent
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// PostgreSQL连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		PG_HOST, PG_USER, PG_PASSWORD, PG_DBNAME, PG_PORT, PG_SSLMODE)

	db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✅ GORM: Connected to PostgreSQL database")

	gm.DB = db

	// 自动迁移所有表结构
	if err := gm.autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate tables: %v", err)
	}

	return nil
}

// autoMigrate 自动迁移所有表结构
func (gm *GORMManager) autoMigrate() error {
	log.Println("🔄 GORM: Starting auto migration...")
	
	// 迁移所有表
	err := gm.DB.AutoMigrate(
		&UserGORM{},
		&RoutingRuleGORM{},
		&GroupCacheGORM{},
		&FriendCacheGORM{},
		&MemberCacheGORM{},
		&SystemStatGORM{},
		&GroupStatsGORM{},
		&UserStatsGORM{},
		&GroupStatsTodayGORM{},
		&UserStatsTodayGORM{},
		&FissionConfigGORM{},
		&InvitationGORM{},
		&FissionTaskGORM{},
		&UserFissionRecordGORM{},
		&FissionRewardLogGORM{},
	)
	
	if err != nil {
		return err
	}
	
	log.Println("✅ GORM: Auto migration completed successfully")
	return nil
}

// ==================== GORM CRUD操作 ====================

// GORMSaveUser 使用GORM保存用户
func (gm *GORMManager) GORMSaveUser(user *User) error {
	userGORM := &UserGORM{}
	userGORM.FromUser(user)
	
	// 检查是否已存在
	var existing UserGORM
	result := gm.DB.Where("username = ?", user.Username).First(&existing)
	
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		return gm.DB.Create(userGORM).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// 更新现有用户
		userGORM.ID = existing.ID
		return gm.DB.Save(userGORM).Error
	}
}

// GORMLoadUsers 使用GORM加载所有用户
func (gm *GORMManager) GORMLoadUsers() ([]*User, error) {
	var usersGORM []UserGORM
	result := gm.DB.Find(&usersGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	users := make([]*User, len(usersGORM))
	for i, userGORM := range usersGORM {
		users[i] = userGORM.ToUser()
	}
	
	return users, nil
}

// GORMSaveRoutingRule 使用GORM保存路由规则
func (gm *GORMManager) GORMSaveRoutingRule(rule *RoutingRule) error {
	ruleGORM := &RoutingRuleGORM{}
	ruleGORM.FromRoutingRule(rule)
	
	// 检查是否已存在
	var existing RoutingRuleGORM
	result := gm.DB.Where("pattern = ?", rule.Pattern).First(&existing)
	
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新规则
		return gm.DB.Create(ruleGORM).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// 更新现有规则
		return gm.DB.Model(&existing).Updates(ruleGORM).Error
	}
}

// GORMLoadRoutingRules 使用GORM加载所有路由规则
func (gm *GORMManager) GORMLoadRoutingRules() ([]*RoutingRule, error) {
	var rulesGORM []RoutingRuleGORM
	result := gm.DB.Find(&rulesGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	rules := make([]*RoutingRule, len(rulesGORM))
	for i, ruleGORM := range rulesGORM {
		rules[i] = ruleGORM.ToRoutingRule()
	}
	
	return rules, nil
}

// GORMSaveGroupCache 使用GORM保存群组缓存
func (gm *GORMManager) GORMSaveGroupCache(cache *GroupCache) error {
	cacheGORM := &GroupCacheGORM{}
	cacheGORM.FromGroupCache(cache)
	
	// 使用Upsert操作
	return gm.DB.Save(cacheGORM).Error
}

// GORMLoadGroupCaches 使用GORM加载所有群组缓存
func (gm *GORMManager) GORMLoadGroupCaches() ([]*GroupCache, error) {
	var cachesGORM []GroupCacheGORM
	result := gm.DB.Find(&cachesGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	caches := make([]*GroupCache, len(cachesGORM))
	for i, cacheGORM := range cachesGORM {
		caches[i] = cacheGORM.ToGroupCache()
	}
	
	return caches, nil
}

// GORMSaveFriendCache 使用GORM保存好友缓存
func (gm *GORMManager) GORMSaveFriendCache(cache *FriendCache) error {
	cacheGORM := &FriendCacheGORM{}
	cacheGORM.FromFriendCache(cache)
	
	// 使用Upsert操作
	return gm.DB.Save(cacheGORM).Error
}

// GORMLoadFriendCaches 使用GORM加载所有好友缓存
func (gm *GORMManager) GORMLoadFriendCaches() ([]*FriendCache, error) {
	var cachesGORM []FriendCacheGORM
	result := gm.DB.Find(&cachesGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	caches := make([]*FriendCache, len(cachesGORM))
	for i, cacheGORM := range cachesGORM {
		caches[i] = cacheGORM.ToFriendCache()
	}
	
	return caches, nil
}

// GORMSaveMemberCache 使用GORM保存群成员缓存
func (gm *GORMManager) GORMSaveMemberCache(cache *MemberCache) error {
	cacheGORM := &MemberCacheGORM{}
	cacheGORM.FromMemberCache(cache)
	
	// 使用Upsert操作
	return gm.DB.Save(cacheGORM).Error
}

// GORMLoadMemberCaches 使用GORM加载所有群成员缓存
func (gm *GORMManager) GORMLoadMemberCaches() ([]*MemberCache, error) {
	var cachesGORM []MemberCacheGORM
	result := gm.DB.Find(&cachesGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	caches := make([]*MemberCache, len(cachesGORM))
	for i, cacheGORM := range cachesGORM {
		caches[i] = cacheGORM.ToMemberCache()
	}
	
	return caches, nil
}

// GORMSaveSystemStat 使用GORM保存系统统计
func (gm *GORMManager) GORMSaveSystemStat(key string, value interface{}) error {
	statGORM := &SystemStatGORM{}
	statGORM.FromSystemStat(key, value)
	
	// 使用Upsert操作
	return gm.DB.Save(statGORM).Error
}

// GORMLoadSystemStats 使用GORM加载所有系统统计
func (gm *GORMManager) GORMLoadSystemStats() (map[string]interface{}, error) {
	var statsGORM []SystemStatGORM
	result := gm.DB.Find(&statsGORM)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	stats := make(map[string]interface{})
	for _, statGORM := range statsGORM {
		stats[statGORM.Key] = statGORM.Value
	}
	
	return stats, nil
}

// GORMLoadSystemStat 使用GORM加载单个系统统计
func (gm *GORMManager) GORMLoadSystemStat(key string) (interface{}, error) {
	var statGORM SystemStatGORM
	result := gm.DB.Where("key = ?", key).First(&statGORM)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	
	return statGORM.Value, nil
}

// GORMSaveGroupStats 使用GORM保存群组统计
func (gm *GORMManager) GORMSaveGroupStats(id string, count int64) error {
	statsGORM := &GroupStatsGORM{
		ID:    id,
		Count: count,
	}
	
	// 使用Upsert操作
	return gm.DB.Save(statsGORM).Error
}

// GORMLoadGroupStats 使用GORM加载群组统计
func (gm *GORMManager) GORMLoadGroupStats(id string) (int64, error) {
	var statsGORM GroupStatsGORM
	result := gm.DB.Where("id = ?", id).First(&statsGORM)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	
	return statsGORM.Count, nil
}

// GORMSaveUserStats 使用GORM保存用户统计
func (gm *GORMManager) GORMSaveUserStats(id string, count int64) error {
	statsGORM := &UserStatsGORM{
		ID:    id,
		Count: count,
	}
	
	// 使用Upsert操作
	return gm.DB.Save(statsGORM).Error
}

// GORMLoadUserStats 使用GORM加载用户统计
func (gm *GORMManager) GORMLoadUserStats(id string) (int64, error) {
	var statsGORM UserStatsGORM
	result := gm.DB.Where("id = ?", id).First(&statsGORM)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	
	return statsGORM.Count, nil
}

// GORMSaveGroupStatsToday 使用GORM保存群组每日统计
func (gm *GORMManager) GORMSaveGroupStatsToday(id string, day string, count int64) error {
	statsGORM := &GroupStatsTodayGORM{
		ID:    id,
		Day:   day,
		Count: count,
	}
	
	// 使用Upsert操作
	return gm.DB.Save(statsGORM).Error
}

// GORMLoadGroupStatsToday 使用GORM加载群组每日统计
func (gm *GORMManager) GORMLoadGroupStatsToday(id string, day string) (int64, error) {
	var statsGORM GroupStatsTodayGORM
	result := gm.DB.Where("id = ? AND day = ?", id, day).First(&statsGORM)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	
	return statsGORM.Count, nil
}

// GORMSaveUserStatsToday 使用GORM保存用户每日统计
func (gm *GORMManager) GORMSaveUserStatsToday(id string, day string, count int64) error {
	statsGORM := &UserStatsTodayGORM{
		ID:    id,
		Day:   day,
		Count: count,
	}
	
	// 使用Upsert操作
	return gm.DB.Save(statsGORM).Error
}

// GORMLoadUserStatsToday 使用GORM加载用户每日统计
func (gm *GORMManager) GORMLoadUserStatsToday(id string, day string) (int64, error) {
	var statsGORM UserStatsTodayGORM
	result := gm.DB.Where("id = ? AND day = ?", id, day).First(&statsGORM)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	
	return statsGORM.Count, nil
}

// GORMDeleteUser 使用GORM删除用户
func (gm *GORMManager) GORMDeleteUser(username string) error {
	return gm.DB.Where("username = ?", username).Delete(&UserGORM{}).Error
}

// GORMDeleteRoutingRule 使用GORM删除路由规则
func (gm *GORMManager) GORMDeleteRoutingRule(pattern string) error {
	return gm.DB.Where("pattern = ?", pattern).Delete(&RoutingRuleGORM{}).Error
}

// GORMDeleteGroupCache 使用GORM删除群组缓存
func (gm *GORMManager) GORMDeleteGroupCache(groupID string) error {
	return gm.DB.Where("group_id = ?", groupID).Delete(&GroupCacheGORM{}).Error
}

// GORMDeleteFriendCache 使用GORM删除好友缓存
func (gm *GORMManager) GORMDeleteFriendCache(userID string) error {
	return gm.DB.Where("user_id = ?", userID).Delete(&FriendCacheGORM{}).Error
}

// GORMDeleteMemberCache 使用GORM删除群成员缓存
func (gm *GORMManager) GORMDeleteMemberCache(groupID string, userID string) error {
	return gm.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&MemberCacheGORM{}).Error
}

// GORMDeleteSystemStat 使用GORM删除系统统计
func (gm *GORMManager) GORMDeleteSystemStat(key string) error {
	return gm.DB.Where("key = ?", key).Delete(&SystemStatGORM{}).Error
}

// GORMDeleteGroupStats 使用GORM删除群组统计
func (gm *GORMManager) GORMDeleteGroupStats(id string) error {
	return gm.DB.Where("id = ?", id).Delete(&GroupStatsGORM{}).Error
}

// GORMDeleteUserStats 使用GORM删除用户统计
func (gm *GORMManager) GORMDeleteUserStats(id string) error {
	return gm.DB.Where("id = ?", id).Delete(&UserStatsGORM{}).Error
}

// GORMDeleteGroupStatsToday 使用GORM删除群组今日统计
func (gm *GORMManager) GORMDeleteGroupStatsToday(id string, day string) error {
	return gm.DB.Where("id = ? AND day = ?", id, day).Delete(&GroupStatsTodayGORM{}).Error
}

// GORMDeleteUserStatsToday 使用GORM删除用户今日统计
func (gm *GORMManager) GORMDeleteUserStatsToday(id string, day string) error {
	return gm.DB.Where("id = ? AND day = ?", id, day).Delete(&UserStatsTodayGORM{}).Error
}

// GORMTransaction 使用GORM执行事务
func (gm *GORMManager) GORMTransaction(fn func(tx *gorm.DB) error) error {
	return gm.DB.Transaction(fn)
}