package common

import (
	"time"
)

// ==================== 原生SQL缓存操作函数 ====================

// SaveGroupCache 保存群组缓存到数据库
func (m *Manager) SaveGroupCache(cache *GroupCache) error {
	query := `
	INSERT INTO group_cache (group_id, group_name, bot_id, last_seen)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(group_id) DO UPDATE SET
		group_name = EXCLUDED.group_name,
		bot_id = EXCLUDED.bot_id,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := m.DB.Exec(m.PrepareQuery(query), cache.GroupID, cache.GroupName, cache.BotID, cache.LastSeen)
	return err
}

// SaveFriendCache 保存好友缓存到数据库
func (m *Manager) SaveFriendCache(cache *FriendCache) error {
	query := `
	INSERT INTO friend_cache (user_id, nickname, last_seen)
	VALUES (?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := m.DB.Exec(m.PrepareQuery(query), cache.UserID, cache.Nickname, cache.LastSeen)
	return err
}

// SaveMemberCache 保存群成员缓存到数据库
func (m *Manager) SaveMemberCache(cache *MemberCache) error {
	query := `
	INSERT INTO member_cache (group_id, user_id, nickname, card, last_seen)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(group_id, user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		card = EXCLUDED.card,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := m.DB.Exec(m.PrepareQuery(query), cache.GroupID, cache.UserID, cache.Nickname, cache.Card, cache.LastSeen)
	return err
}

// LoadGroupCachesFromDB 从数据库加载所有群组缓存
func (m *Manager) LoadGroupCachesFromDB() ([]*GroupCache, error) {
	rows, err := m.DB.Query(m.PrepareQuery("SELECT group_id, group_name, bot_id, last_seen FROM group_cache"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*GroupCache
	for rows.Next() {
		var cache GroupCache
		var lastSeen time.Time
		if err := rows.Scan(&cache.GroupID, &cache.GroupName, &cache.BotID, &lastSeen); err != nil {
			continue
		}
		cache.LastSeen = lastSeen
		caches = append(caches, &cache)
	}
	return caches, nil
}

// DeleteGroupCache 从数据库删除群组缓存
func (m *Manager) DeleteGroupCache(groupID string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM group_cache WHERE group_id = ?"), groupID)
	return err
}

// DeleteFriendCache 从数据库删除好友缓存
func (m *Manager) DeleteFriendCache(userID string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM friend_cache WHERE user_id = ?"), userID)
	return err
}

// DeleteMemberCache 从数据库删除群成员缓存
func (m *Manager) DeleteMemberCache(groupID, userID string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM member_cache WHERE group_id = ? AND user_id = ?"), groupID, userID)
	return err
}

// LoadFriendCachesFromDB 从数据库加载所有好友缓存
func (m *Manager) LoadFriendCachesFromDB() ([]*FriendCache, error) {
	rows, err := m.DB.Query(m.PrepareQuery("SELECT user_id, nickname, last_seen FROM friend_cache"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*FriendCache
	for rows.Next() {
		var cache FriendCache
		var lastSeen time.Time
		if err := rows.Scan(&cache.UserID, &cache.Nickname, &lastSeen); err != nil {
			continue
		}
		cache.LastSeen = lastSeen
		caches = append(caches, &cache)
	}
	return caches, nil
}

// LoadMemberCachesFromDB 从数据库加载所有群成员缓存
func (m *Manager) LoadMemberCachesFromDB() ([]*MemberCache, error) {
	rows, err := m.DB.Query(m.PrepareQuery("SELECT group_id, user_id, nickname, card, last_seen FROM member_cache"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*MemberCache
	for rows.Next() {
		var cache MemberCache
		var lastSeen time.Time
		if err := rows.Scan(&cache.GroupID, &cache.UserID, &cache.Nickname, &cache.Card, &lastSeen); err != nil {
			continue
		}
		cache.LastSeen = lastSeen
		caches = append(caches, &cache)
	}
	return caches, nil
}