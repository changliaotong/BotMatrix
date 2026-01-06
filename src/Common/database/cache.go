package database

import (
	"database/sql"
	"time"

	"BotMatrix/common/models"
)

// SaveGroupCache saves group cache to database
func SaveGroupCache(db *sql.DB, prepareQuery func(string) string, cache *models.GroupCache) error {
	query := `
	INSERT INTO group_cache (group_id, group_name, bot_id, last_seen)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(group_id) DO UPDATE SET
		group_name = EXCLUDED.group_name,
		bot_id = EXCLUDED.bot_id,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := db.Exec(prepareQuery(query), cache.GroupID, cache.GroupName, cache.BotID, cache.LastSeen)
	return err
}

// SaveFriendCache saves friend cache to database
func SaveFriendCache(db *sql.DB, prepareQuery func(string) string, cache *models.FriendCache) error {
	query := `
	INSERT INTO friend_cache (user_id, nickname, last_seen)
	VALUES (?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := db.Exec(prepareQuery(query), cache.UserID, cache.Nickname, cache.LastSeen)
	return err
}

// SaveMemberCache saves member cache to database
func SaveMemberCache(db *sql.DB, prepareQuery func(string) string, cache *models.MemberCache) error {
	query := `
	INSERT INTO member_cache (group_id, user_id, nickname, card, role, last_seen)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(group_id, user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		card = EXCLUDED.card,
		role = EXCLUDED.role,
		last_seen = EXCLUDED.last_seen;
	`
	_, err := db.Exec(prepareQuery(query), cache.GroupID, cache.UserID, cache.Nickname, cache.Card, cache.Role, cache.LastSeen)
	return err
}

// LoadGroupCachesFromDB loads all group caches from database
func LoadGroupCachesFromDB(db *sql.DB, prepareQuery func(string) string) ([]*models.GroupCache, error) {
	rows, err := db.Query(prepareQuery("SELECT group_id, group_name, bot_id, last_seen FROM group_cache"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*models.GroupCache
	for rows.Next() {
		var cache models.GroupCache
		var lastSeen time.Time
		if err := rows.Scan(&cache.GroupID, &cache.GroupName, &cache.BotID, &lastSeen); err != nil {
			continue
		}
		cache.LastSeen = lastSeen
		caches = append(caches, &cache)
	}
	return caches, nil
}

// DeleteGroupCache deletes group cache from database
func DeleteGroupCache(db *sql.DB, prepareQuery func(string) string, groupID string) error {
	_, err := db.Exec(prepareQuery("DELETE FROM group_cache WHERE group_id = ?"), groupID)
	return err
}

// DeleteFriendCache deletes friend cache from database
func DeleteFriendCache(db *sql.DB, prepareQuery func(string) string, userID string) error {
	_, err := db.Exec(prepareQuery("DELETE FROM friend_cache WHERE user_id = ?"), userID)
	return err
}

// DeleteMemberCache deletes member cache from database
func DeleteMemberCache(db *sql.DB, prepareQuery func(string) string, groupID, userID string) error {
	_, err := db.Exec(prepareQuery("DELETE FROM member_cache WHERE group_id = ? AND user_id = ?"), groupID, userID)
	return err
}

// LoadFriendCachesFromDB loads all friend caches from database
func LoadFriendCachesFromDB(db *sql.DB, prepareQuery func(string) string) ([]*models.FriendCache, error) {
	rows, err := db.Query(prepareQuery("SELECT user_id, nickname, last_seen FROM friend_cache"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*models.FriendCache
	for rows.Next() {
		var cache models.FriendCache
		var lastSeen time.Time
		if err := rows.Scan(&cache.UserID, &cache.Nickname, &lastSeen); err != nil {
			continue
		}
		cache.LastSeen = lastSeen
		caches = append(caches, &cache)
	}
	return caches, nil
}
