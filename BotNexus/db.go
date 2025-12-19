package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const DB_FILE = "data/botnexus.db"

// initDB 初始化数据库
func (m *Manager) initDB() error {
	// 确保目录存在
	dbDir := filepath.Dir(DB_FILE)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("无法创建数据库目录: %v", err)
	}

	db, err := sql.Open("sqlite", DB_FILE)
	if err != nil {
		return err
	}

	m.db = db

	// 设置繁忙超时，解决数据库锁定问题
	_, err = m.db.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		log.Printf("设置数据库繁忙超时失败: %v", err)
	}

	// 创建用户表
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT 0,
		session_version INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	);`

	_, err = m.db.Exec(query)
	if err != nil {
		log.Printf("创建用户表失败: %v", err)
		return err
	}

	// 创建路由规则表
	routingQuery := `
	CREATE TABLE IF NOT EXISTS routing_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT UNIQUE NOT NULL,
		target_worker_id TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	);`

	_, err = m.db.Exec(routingQuery)
	if err != nil {
		log.Printf("创建路由规则表失败: %v", err)
		return err
	}

	// 创建群组缓存表
	groupCacheQuery := `
	CREATE TABLE IF NOT EXISTS group_cache (
		group_id TEXT PRIMARY KEY,
		group_name TEXT,
		bot_id TEXT,
		last_seen DATETIME
	);`
	_, err = m.db.Exec(groupCacheQuery)
	if err != nil {
		log.Printf("创建群组缓存表失败: %v", err)
	}

	// 创建好友缓存表
	friendCacheQuery := `
	CREATE TABLE IF NOT EXISTS friend_cache (
		user_id TEXT PRIMARY KEY,
		nickname TEXT,
		last_seen DATETIME
	);`
	_, err = m.db.Exec(friendCacheQuery)
	if err != nil {
		log.Printf("创建好友缓存表失败: %v", err)
	}

	// 创建群成员缓存表
	memberCacheQuery := `
	CREATE TABLE IF NOT EXISTS member_cache (
		group_id TEXT,
		user_id TEXT,
		nickname TEXT,
		card TEXT,
		last_seen DATETIME,
		PRIMARY KEY (group_id, user_id)
	);`
	_, err = m.db.Exec(memberCacheQuery)
	if err != nil {
		log.Printf("创建群成员缓存表失败: %v", err)
	}

	// 创建系统统计表
	statsQuery := `
	CREATE TABLE IF NOT EXISTS system_stats (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME
	);`
	_, err = m.db.Exec(statsQuery)
	if err != nil {
		log.Printf("创建系统统计表失败: %v", err)
	}

	log.Printf("数据库初始化成功: %s", DB_FILE)
	return nil
}

// saveStatToDB 保存系统统计到数据库
func (m *Manager) saveStatToDB(key string, value interface{}) error {
	query := `
	INSERT INTO system_stats (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = excluded.updated_at;
	`
	now := time.Now().Format(time.RFC3339)
	_, err := m.db.Exec(query, key, fmt.Sprintf("%v", value), now)
	return err
}

// loadStatsFromDB 从数据库加载系统统计
func (m *Manager) loadStatsFromDB() error {
	rows, err := m.db.Query("SELECT key, value FROM system_stats")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			if key == "total_messages" {
				var count int64
				fmt.Sscanf(value, "%d", &count)
				m.TotalMessages = count
			}
		}
	}
	return nil
}

// saveGroupToDB 保存群组到数据库
func (m *Manager) saveGroupToDB(groupID, groupName, botID string) error {
	query := `
	INSERT INTO group_cache (group_id, group_name, bot_id, last_seen)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(group_id) DO UPDATE SET
		group_name = excluded.group_name,
		bot_id = excluded.bot_id,
		last_seen = excluded.last_seen;
	`
	now := time.Now().Format(time.RFC3339)
	_, err := m.db.Exec(query, groupID, groupName, botID, now)
	return err
}

// saveFriendToDB 保存好友到数据库
func (m *Manager) saveFriendToDB(userID, nickname string) error {
	query := `
	INSERT INTO friend_cache (user_id, nickname, last_seen)
	VALUES (?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		nickname = excluded.nickname,
		last_seen = excluded.last_seen;
	`
	now := time.Now().Format(time.RFC3339)
	_, err := m.db.Exec(query, userID, nickname, now)
	return err
}

// saveMemberToDB 保存群成员到数据库
func (m *Manager) saveMemberToDB(groupID, userID, nickname, card string) error {
	query := `
	INSERT INTO member_cache (group_id, user_id, nickname, card, last_seen)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(group_id, user_id) DO UPDATE SET
		nickname = excluded.nickname,
		card = excluded.card,
		last_seen = excluded.last_seen;
	`
	now := time.Now().Format(time.RFC3339)
	_, err := m.db.Exec(query, groupID, userID, nickname, card, now)
	return err
}

// loadCachesFromDB 从数据库加载所有缓存到内存
func (m *Manager) loadCachesFromDB() error {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// 1. 加载群组
	rows, err := m.db.Query("SELECT group_id, group_name, bot_id FROM group_cache")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var gID, name, botID string
			if err := rows.Scan(&gID, &name, &botID); err == nil {
				m.groupCache[gID] = map[string]interface{}{
					"group_id":   gID,
					"group_name": name,
					"bot_id":     botID,
					"is_cached":  true,
				}
			}
		}
	}

	// 2. 加载好友
	rowsF, err := m.db.Query("SELECT user_id, nickname FROM friend_cache")
	if err == nil {
		defer rowsF.Close()
		for rowsF.Next() {
			var uID, nickname string
			if err := rowsF.Scan(&uID, &nickname); err == nil {
				m.friendCache[uID] = map[string]interface{}{
					"user_id":   uID,
					"nickname":  nickname,
					"is_cached": true,
				}
			}
		}
	}

	// 3. 加载群成员
	rowsM, err := m.db.Query("SELECT group_id, user_id, nickname, card FROM member_cache")
	if err == nil {
		defer rowsM.Close()
		for rowsM.Next() {
			var gID, uID, nickname, card string
			if err := rowsM.Scan(&gID, &uID, &nickname, &card); err == nil {
				key := fmt.Sprintf("%s:%s", gID, uID)
				m.memberCache[key] = map[string]interface{}{
					"group_id":  gID,
					"user_id":   uID,
					"nickname":  nickname,
					"card":      card,
					"is_cached": true,
				}
			}
		}
	}

	log.Printf("[INFO] 从数据库加载了 %d 个群组, %d 个好友, %d 个成员缓存", len(m.groupCache), len(m.friendCache), len(m.memberCache))
	return nil
}

// loadRoutingRulesFromDB 从数据库加载所有路由规则到内存缓存
func (m *Manager) loadRoutingRulesFromDB() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rows, err := m.db.Query("SELECT pattern, target_worker_id FROM routing_rules")
	if err != nil {
		return err
	}
	defer rows.Close()

	m.routingRules = make(map[string]string)
	count := 0
	for rows.Next() {
		var pattern, target string
		if err := rows.Scan(&pattern, &target); err != nil {
			log.Printf("[ERROR] 解析路由规则行失败: %v", err)
			continue
		}
		m.routingRules[pattern] = target
		count++
	}
	log.Printf("[INFO] 从数据库加载了 %d 条路由规则", count)
	return nil
}

// saveRoutingRuleToDB 保存路由规则到数据库
func (m *Manager) saveRoutingRuleToDB(pattern, target string) error {
	query := `
	INSERT INTO routing_rules (pattern, target_worker_id, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(pattern) DO UPDATE SET
		target_worker_id = excluded.target_worker_id,
		updated_at = excluded.updated_at;
	`
	now := time.Now().Format(time.RFC3339)
	_, err := m.db.Exec(query, pattern, target, now, now)
	return err
}

// deleteRoutingRuleFromDB 从数据库删除路由规则
func (m *Manager) deleteRoutingRuleFromDB(pattern string) error {
	_, err := m.db.Exec("DELETE FROM routing_rules WHERE pattern = ?", pattern)
	return err
}

// loadUsersFromDB 从数据库加载所有用户到内存缓存
func (m *Manager) loadUsersFromDB() error {
	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()
	return m.loadUsersFromDBNoLock()
}

// loadUsersFromDBNoLock 从数据库加载所有用户到内存缓存 (无锁版本)
func (m *Manager) loadUsersFromDBNoLock() error {
	rows, err := m.db.Query("SELECT id, username, password_hash, is_admin, session_version, created_at, updated_at FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	// 清空当前内存缓存并重新加载
	m.users = make(map[string]*User)

	for rows.Next() {
		var user User
		var createdAt, updatedAt string
		err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.SessionVersion, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("解析用户行失败: %v", err)
			continue
		}

		if createdAt != "" {
			user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt != "" {
			user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}

		m.users[user.Username] = &user
	}

	log.Printf("从数据库加载了 %d 个用户", len(m.users))
	return nil
}

// saveUserToDB 保存或更新用户信息到数据库
func (m *Manager) saveUserToDB(user *User) error {
	query := `
	INSERT INTO users (username, password_hash, is_admin, session_version, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(username) DO UPDATE SET
		password_hash = excluded.password_hash,
		is_admin = excluded.is_admin,
		session_version = excluded.session_version,
		updated_at = excluded.updated_at;
	`

	_, err := m.db.Exec(query,
		user.Username,
		user.PasswordHash,
		user.IsAdmin,
		user.SessionVersion,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// deleteUserFromDB 从数据库删除用户
func (m *Manager) deleteUserFromDB(username string) error {
	_, err := m.db.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}
