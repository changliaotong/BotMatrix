package common

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// ==================== GORM兼容层 ====================
// 这些函数保持现有接口不变，但内部可以选择使用GORM或原生SQL

// SaveUserWithGORM 使用GORM保存用户（如果可用）
func (m *Manager) SaveUserWithGORM(user *User) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveUser(user)
	}
	// 回退到原生SQL
	return m.SaveUserToDB(user)
}

// LoadUsersWithGORM 使用GORM加载用户（如果可用）
func (m *Manager) LoadUsersWithGORM() ([]*User, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadUsers()
	}
	// 回退到原生SQL - 需要先加载到内存，然后返回用户列表
	err := m.LoadUsersFromDB()
	if err != nil {
		return nil, err
	}
	
	// 从内存缓存中获取用户列表
	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()
	
	users := make([]*User, 0, len(m.Users))
	for _, user := range m.Users {
		users = append(users, user)
	}
	return users, nil
}

// SaveRoutingRuleWithGORM 使用GORM保存路由规则（如果可用）
func (m *Manager) SaveRoutingRuleWithGORM(rule *RoutingRule) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveRoutingRule(rule)
	}
	// 回退到原生SQL
	return m.SaveRoutingRuleToDB(rule.Pattern, rule.TargetWorkerID)
}

// LoadRoutingRulesWithGORM 使用GORM加载路由规则（如果可用）
func (m *Manager) LoadRoutingRulesWithGORM() ([]*RoutingRule, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadRoutingRules()
	}
	// 回退到原生SQL - 需要先加载到内存，然后返回规则列表
	err := m.LoadRoutingRulesFromDB()
	if err != nil {
		return nil, err
	}
	
	// 从内存缓存中获取路由规则列表
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	
	rules := make([]*RoutingRule, 0, len(m.RoutingRules))
	for pattern, target := range m.RoutingRules {
		rules = append(rules, &RoutingRule{
			Pattern:        pattern,
			TargetWorkerID: target,
		})
	}
	return rules, nil
}

// SaveGroupCacheWithGORM 使用GORM保存群组缓存（如果可用）
func (m *Manager) SaveGroupCacheWithGORM(cache *GroupCache) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveGroupCache(cache)
	}
	// 回退到原生SQL
	return m.SaveGroupCache(cache)
}

// LoadGroupCachesWithGORM 使用GORM加载群组缓存（如果可用）
func (m *Manager) LoadGroupCachesWithGORM() ([]*GroupCache, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadGroupCaches()
	}
	// 回退到原生SQL
	return m.LoadGroupCachesFromDB()
}

// SaveFriendCacheWithGORM 使用GORM保存好友缓存（如果可用）
func (m *Manager) SaveFriendCacheWithGORM(cache *FriendCache) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveFriendCache(cache)
	}
	// 回退到原生SQL
	return m.SaveFriendCache(cache)
}

// LoadFriendCachesWithGORM 使用GORM加载好友缓存（如果可用）
func (m *Manager) LoadFriendCachesWithGORM() ([]*FriendCache, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadFriendCaches()
	}
	// 回退到原生SQL
	return m.LoadFriendCachesFromDB()
}

// SaveMemberCacheWithGORM 使用GORM保存群成员缓存（如果可用）
func (m *Manager) SaveMemberCacheWithGORM(cache *MemberCache) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveMemberCache(cache)
	}
	// 回退到原生SQL
	return m.SaveMemberCache(cache)
}

// LoadMemberCachesWithGORM 使用GORM加载群成员缓存（如果可用）
func (m *Manager) LoadMemberCachesWithGORM() ([]*MemberCache, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadMemberCaches()
	}
	// 回退到原生SQL
	return m.LoadMemberCachesFromDB()
}

// SaveSystemStatWithGORM 使用GORM保存系统统计（如果可用）
func (m *Manager) SaveSystemStatWithGORM(key string, value any) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveSystemStat(key, value)
	}
	// 回退到原生SQL
	return m.SaveSystemStat(key, value)
}

// LoadSystemStatsWithGORM 使用GORM加载系统统计（如果可用）
func (m *Manager) LoadSystemStatsWithGORM() (map[string]any, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadSystemStats()
	}
	// 回退到原生SQL
	return m.LoadSystemStatsFromDB()
}

// LoadSystemStatWithGORM 使用GORM加载单个系统统计（如果可用）
func (m *Manager) LoadSystemStatWithGORM(key string) (any, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadSystemStat(key)
	}
	// 回退到原生SQL
	return m.LoadSystemStat(key)
}

// SaveGroupStatsWithGORM 使用GORM保存群组统计（如果可用）
func (m *Manager) SaveGroupStatsWithGORM(id string, count int64) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveGroupStats(id, count)
	}
	// 回退到原生SQL
	return m.SaveGroupStats(id, count)
}

// LoadGroupStatsWithGORM 使用GORM加载群组统计（如果可用）
func (m *Manager) LoadGroupStatsWithGORM(id string) (int64, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadGroupStats(id)
	}
	// 回退到原生SQL
	return m.LoadGroupStats(id)
}

// SaveUserStatsWithGORM 使用GORM保存用户统计（如果可用）
func (m *Manager) SaveUserStatsWithGORM(id string, count int64) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveUserStats(id, count)
	}
	// 回退到原生SQL
	return m.SaveUserStats(id, count)
}

// LoadUserStatsWithGORM 使用GORM加载用户统计（如果可用）
func (m *Manager) LoadUserStatsWithGORM(id string) (int64, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadUserStats(id)
	}
	// 回退到原生SQL
	return m.LoadUserStats(id)
}

// SaveGroupStatsTodayWithGORM 使用GORM保存群组每日统计（如果可用）
func (m *Manager) SaveGroupStatsTodayWithGORM(id string, day string, count int64) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveGroupStatsToday(id, day, count)
	}
	// 回退到原生SQL
	return m.SaveGroupStatsToday(id, day, count)
}

// LoadGroupStatsTodayWithGORM 使用GORM加载群组每日统计（如果可用）
func (m *Manager) LoadGroupStatsTodayWithGORM(id string, day string) (int64, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadGroupStatsToday(id, day)
	}
	// 回退到原生SQL
	return m.LoadGroupStatsToday(id, day)
}

// SaveUserStatsTodayWithGORM 使用GORM保存用户每日统计（如果可用）
func (m *Manager) SaveUserStatsTodayWithGORM(id string, day string, count int64) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMSaveUserStatsToday(id, day, count)
	}
	// 回退到原生SQL
	return m.SaveUserStatsToday(id, day, count)
}

// LoadUserStatsTodayWithGORM 使用GORM加载用户每日统计（如果可用）
func (m *Manager) LoadUserStatsTodayWithGORM(id string, day string) (int64, error) {
	if m.GORMManager != nil {
		return m.GORMManager.GORMLoadUserStatsToday(id, day)
	}
	// 回退到原生SQL
	return m.LoadUserStatsToday(id, day)
}

// DeleteUserWithGORM 使用GORM删除用户（如果可用）
func (m *Manager) DeleteUserWithGORM(username string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteUser(username)
	}
	// 回退到原生SQL
	return m.DeleteUser(username)
}

// DeleteRoutingRuleWithGORM 使用GORM删除路由规则（如果可用）
func (m *Manager) DeleteRoutingRuleWithGORM(pattern string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteRoutingRule(pattern)
	}
	// 回退到原生SQL
	return m.DeleteRoutingRule(pattern)
}

// DeleteGroupCacheWithGORM 使用GORM删除群组缓存（如果可用）
func (m *Manager) DeleteGroupCacheWithGORM(groupID string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteGroupCache(groupID)
	}
	// 回退到原生SQL
	return m.DeleteGroupCache(groupID)
}

// DeleteFriendCacheWithGORM 使用GORM删除好友缓存（如果可用）
func (m *Manager) DeleteFriendCacheWithGORM(userID string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteFriendCache(userID)
	}
	// 回退到原生SQL
	return m.DeleteFriendCache(userID)
}

// DeleteMemberCacheWithGORM 使用GORM删除群成员缓存（如果可用）
func (m *Manager) DeleteMemberCacheWithGORM(groupID string, userID string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteMemberCache(groupID, userID)
	}
	// 回退到原生SQL
	return m.DeleteMemberCache(groupID, userID)
}

// DeleteSystemStatWithGORM 使用GORM删除系统统计（如果可用）
func (m *Manager) DeleteSystemStatWithGORM(key string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteSystemStat(key)
	}
	// 回退到原生SQL
	return m.DeleteSystemStat(key)
}

// DeleteUserStatsTodayWithGORM 使用GORM删除用户今日统计（如果可用）
func (m *Manager) DeleteUserStatsTodayWithGORM(id string, day string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteUserStatsToday(id, day)
	}
	// 回退到原生SQL
	return m.DeleteUserStatsToday(id, day)
}

// DeleteGroupStatsWithGORM 使用GORM删除群组统计（如果可用）
func (m *Manager) DeleteGroupStatsWithGORM(id string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteGroupStats(id)
	}
	// 回退到原生SQL
	return m.DeleteGroupStats(id)
}

// DeleteUserStatsWithGORM 使用GORM删除用户统计（如果可用）
func (m *Manager) DeleteUserStatsWithGORM(id string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteUserStats(id)
	}
	// 回退到原生SQL
	return m.DeleteUserStats(id)
}

// DeleteGroupStatsTodayWithGORM 使用GORM删除群组今日统计（如果可用）
func (m *Manager) DeleteGroupStatsTodayWithGORM(id string, day string) error {
	if m.GORMManager != nil {
		return m.GORMManager.GORMDeleteGroupStatsToday(id, day)
	}
	// 回退到原生SQL
	return m.DeleteGroupStatsToday(id, day)
}

// TransactionWithGORM 使用GORM执行事务（如果可用）
func (m *Manager) TransactionWithGORM(fn func(tx *Manager) error) error {
	if m.GORMManager != nil {
		return m.GORMManager.DB.Transaction(func(tx *gorm.DB) error {
			// 创建一个包装了GORM事务DB的临时Manager
			txGM := &GORMManager{DB: tx}
			txManager := &Manager{
				GORMManager: txGM,
				// 复制其他必要字段
				Users:           m.Users,
				RoutingRules:    m.RoutingRules,
				GroupStats:      m.GroupStats,
				UserStats:       m.UserStats,
				GroupStatsToday: m.GroupStatsToday,
				UserStatsToday:  m.UserStatsToday,
			}
			return fn(txManager)
		})
	}
	// 回退到原生SQL事务
	return m.Transaction(fn)
}

// IsGORMEnabled 检查GORM是否启用
func (m *Manager) IsGORMEnabled() bool {
	return m.GORMManager != nil
}

// GetGORMManager 获取GORM管理器
func (m *Manager) GetGORMManager() *GORMManager {
	return m.GORMManager
}

// SwitchToGORM 切换到GORM模式（如果可用）
func (m *Manager) SwitchToGORM() error {
	if !m.IsGORMEnabled() {
		return fmt.Errorf("GORM未启用，请设置USE_GORM=true环境变量")
	}
	
	log.Println("🔄 切换到GORM模式...")
	
	// 重新加载所有数据到内存
	if err := m.loadAllDataWithGORM(); err != nil {
		return fmt.Errorf("切换到GORM模式失败: %v", err)
	}
	
	log.Println("✅ 成功切换到GORM模式")
	return nil
}

// loadAllDataWithGORM 使用GORM重新加载所有数据
func (m *Manager) loadAllDataWithGORM() error {
	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()
	
	// 重新加载用户
	users, err := m.LoadUsersWithGORM()
	if err != nil {
		return fmt.Errorf("加载用户失败: %v", err)
	}
	
	m.Users = make(map[string]*User)
	for _, user := range users {
		m.Users[user.Username] = user
	}
	
	// 重新加载路由规则
	rules, err := m.LoadRoutingRulesWithGORM()
	if err != nil {
		return fmt.Errorf("加载路由规则失败: %v", err)
	}
	
	m.RoutingRules = make(map[string]string)
	for _, rule := range rules {
		m.RoutingRules[rule.Pattern] = rule.TargetWorkerID
	}
	
	return nil
}