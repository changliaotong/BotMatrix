package main

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

const DB_FILE = "botnexus.db"

// initDB 初始化数据库
func (m *Manager) initDB() error {
	db, err := sql.Open("sqlite", DB_FILE)
	if err != nil {
		return err
	}

	m.db = db

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

	log.Printf("数据库初始化成功: %s", DB_FILE)
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
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	return err
}

// deleteUserFromDB 从数据库删除用户
func (m *Manager) deleteUserFromDB(username string) error {
	_, err := m.db.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}
