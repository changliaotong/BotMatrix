package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// initDB 初始化数据库
func (m *Manager) InitDB() error {
	// 初始化内存缓存 Map
	if m.Users == nil {
		m.Users = make(map[string]*User)
	}
	if m.RoutingRules == nil {
		m.RoutingRules = make(map[string]string)
	}
	if m.GroupStats == nil {
		m.GroupStats = make(map[string]int64)
	}
	if m.UserStats == nil {
		m.UserStats = make(map[string]int64)
	}
	if m.GroupStatsToday == nil {
		m.GroupStatsToday = make(map[string]int64)
	}
	if m.UserStatsToday == nil {
		m.UserStatsToday = make(map[string]int64)
	}

	var db *sql.DB
	var err error

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME, PG_SSLMODE)
	log.Printf("[DB] 正在连接 PostgreSQL: %s:%d/%s", PG_HOST, PG_PORT, PG_DBNAME)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("无法连接 PostgreSQL: %v", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("无法验证 PostgreSQL 连接: %v", err)
	}

	m.DB = db

	m.DB.Exec(m.PrepareQuery(`UPDATE users SET active = TRUE WHERE username = 'admin' AND active = FALSE`))
	m.InitDefaultAdmin()

	// PostgreSQL 建表逻辑
	idType := "SERIAL PRIMARY KEY"

	// 创建用户表
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS users (
		id %s,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT FALSE,
		active BOOLEAN DEFAULT TRUE,
		session_version INTEGER DEFAULT 1,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);`, idType)

	_, err = m.DB.Exec(m.PrepareQuery(query))
	if err != nil {
		log.Printf("创建用户表失败: %v", err)
		return err
	}

	// 兼容已有数据库：确保 active 字段存在
	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE users ADD COLUMN active BOOLEAN DEFAULT TRUE`))
	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "duplicate column name") && !strings.Contains(errMsg, "already exists") {
			log.Printf("为 users 表添加 active 字段失败: %v", err)
			return err
		}
	} else {
		// 如果是新添加的字段，确保所有现有用户都是启用状态
		m.DB.Exec(m.PrepareQuery(`UPDATE users SET active = TRUE WHERE active IS NULL`))
	}

	// 创建路由规则表
	routingQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS routing_rules (
		id %s,
		pattern TEXT UNIQUE NOT NULL,
		target_worker_id TEXT NOT NULL,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);`, idType)

	_, err = m.DB.Exec(m.PrepareQuery(routingQuery))
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
		last_seen TIMESTAMP
	);`
	_, err = m.DB.Exec(m.PrepareQuery(groupCacheQuery))
	if err != nil {
		log.Printf("创建群组缓存表失败: %v", err)
		return err
	}

	// 添加 bot_id 字段（如果不存在）
	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE group_cache ADD COLUMN IF NOT EXISTS bot_id TEXT`))
	if err != nil {
		log.Printf("为 group_cache 表添加 bot_id 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	// 创建好友缓存表
	friendCacheQuery := `
	CREATE TABLE IF NOT EXISTS friend_cache (
		user_id TEXT PRIMARY KEY,
		nickname TEXT,
		bot_id TEXT,
		last_seen TIMESTAMP
	);`
	_, err = m.DB.Exec(m.PrepareQuery(friendCacheQuery))
	if err != nil {
		log.Printf("创建好友缓存表失败: %v", err)
		return err
	}

	// 添加 bot_id 字段（如果不存在）
	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE friend_cache ADD COLUMN IF NOT EXISTS bot_id TEXT`))
	if err != nil {
		log.Printf("为 friend_cache 表添加 bot_id 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	// 创建群成员缓存表
	memberCacheQuery := `
	CREATE TABLE IF NOT EXISTS member_cache (
		group_id TEXT,
		user_id TEXT,
		nickname TEXT,
		card TEXT,
		last_seen TIMESTAMP,
		last_sign_time TIMESTAMP,
		streak INTEGER DEFAULT 0,
		total_sign_days INTEGER DEFAULT 0,
		total_points INTEGER DEFAULT 0,
		PRIMARY KEY (group_id, user_id)
	);`
	_, err = m.DB.Exec(m.PrepareQuery(memberCacheQuery))
	if err != nil {
		log.Printf("创建群成员缓存表失败: %v", err)
		return err
	}

	// 添加签到相关字段（如果不存在）
	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE member_cache ADD COLUMN IF NOT EXISTS last_sign_time TIMESTAMP`))
	if err != nil {
		log.Printf("为 member_cache 表添加 last_sign_time 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE member_cache ADD COLUMN IF NOT EXISTS streak INTEGER DEFAULT 0`))
	if err != nil {
		log.Printf("为 member_cache 表添加 streak 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE member_cache ADD COLUMN IF NOT EXISTS total_sign_days INTEGER DEFAULT 0`))
	if err != nil {
		log.Printf("为 member_cache 表添加 total_sign_days 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	_, err = m.DB.Exec(m.PrepareQuery(`ALTER TABLE member_cache ADD COLUMN IF NOT EXISTS total_points INTEGER DEFAULT 0`))
	if err != nil {
		log.Printf("为 member_cache 表添加 total_points 字段失败: %v", err)
		// 不返回错误，继续执行
	}

	// 创建系统统计表
	statsQuery := `
	CREATE TABLE IF NOT EXISTS system_stats (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at TIMESTAMP
	);`
	_, err = m.DB.Exec(m.PrepareQuery(statsQuery))
	if err != nil {
		log.Printf("创建系统统计表失败: %v", err)
		return err
	}

	// 创建详细统计表
	_, err = m.DB.Exec(m.PrepareQuery(`CREATE TABLE IF NOT EXISTS group_stats (id TEXT PRIMARY KEY, count BIGINT, updated_at TIMESTAMP)`))
	if err != nil {
		log.Printf("创建群组统计表失败: %v", err)
		return err
	}
	_, err = m.DB.Exec(m.PrepareQuery(`CREATE TABLE IF NOT EXISTS user_stats (id TEXT PRIMARY KEY, count BIGINT, updated_at TIMESTAMP)`))
	if err != nil {
		log.Printf("创建用户统计表失败: %v", err)
		return err
	}
	_, err = m.DB.Exec(m.PrepareQuery(`CREATE TABLE IF NOT EXISTS group_stats_today (id TEXT PRIMARY KEY, count BIGINT, day TEXT, updated_at TIMESTAMP)`))
	if err != nil {
		log.Printf("创建群组每日统计表失败: %v", err)
		return err
	}
	_, err = m.DB.Exec(m.PrepareQuery(`CREATE TABLE IF NOT EXISTS user_stats_today (id TEXT PRIMARY KEY, count BIGINT, day TEXT, updated_at TIMESTAMP)`))
	if err != nil {
		log.Printf("创建用户每日统计表失败: %v", err)
		return err
	}

	// 创建消息记录表
	messageTableQuery := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message_id TEXT,
		bot_id TEXT,
		user_id TEXT,
		group_id TEXT,
		type TEXT,
		content TEXT,
		raw_message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = m.DB.Exec(m.PrepareQuery(messageTableQuery))
	if err != nil {
		log.Printf("创建消息记录表失败: %v", err)
		return err
	}

	log.Printf("PostgreSQL 数据库初始化成功")

	// 初始化GORM（可选，如果USE_GORM环境变量设置为true）
	if os.Getenv("USE_GORM") == "true" {
		log.Println("🔄 正在初始化GORM...")
		m.GORMManager = NewGORMManager()
		if err := m.GORMManager.InitGORM(); err != nil {
			log.Printf("GORM初始化失败: %v", err)
			// 不返回错误，继续使用原生SQL
		} else {
			log.Println("✅ GORM初始化成功")
			m.GORMDB = m.GORMManager.DB
		}
	}

	return nil
}

// SaveMessageToDB 保存消息到数据库
func (m *Manager) SaveMessageToDB(messageID, botID, userID, groupID, msgType, content, rawMessage string) error {
	if m.DB == nil {
		return nil
	}

	query := `
	INSERT INTO messages (message_id, bot_id, user_id, group_id, type, content, raw_message)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := m.DB.Exec(m.PrepareQuery(query), messageID, botID, userID, groupID, msgType, content, rawMessage)
	if err != nil {
		log.Printf("保存消息到数据库失败: %v", err)
		return err
	}
	return nil
}

// PrepareQuery 根据数据库类型转换 SQL 语句 (目前固定为 PostgreSQL)
func (m *Manager) PrepareQuery(query string) string {
	// 1. 替换 ? 为 $1, $2, $3...
	// 注意：简单的字符串替换可能会有问题，如果 SQL 中包含问号（如 JSON 操作），
	// 但在这个项目中目前没有这种情况。
	n := 1
	for {
		newQuery := ""
		found := false
		for i := 0; i < len(query); i++ {
			if query[i] == '?' {
				newQuery = query[:i] + fmt.Sprintf("$%d", n) + query[i+1:]
				n++
				query = newQuery
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// 2. 统一使用 EXCLUDED (PostgreSQL 要求，SQLite 兼容)
	// query = strings.ReplaceAll(query, "excluded.", "EXCLUDED.")

	return query
}

// SaveStatToDB 保存系统统计到数据库
func (m *Manager) SaveStatToDB(key string, value any) error {
	query := `
	INSERT INTO system_stats (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		value = EXCLUDED.value,
		updated_at = EXCLUDED.updated_at;
	`
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(query), key, fmt.Sprintf("%v", value), now)
	return err
}

// LoadStatsFromDB 从数据库加载系统统计
func (m *Manager) LoadStatsFromDB() error {
	m.StatsMutex.Lock()
	defer m.StatsMutex.Unlock()

	// 初始化 Map
	m.GroupStats = make(map[string]int64)
	m.UserStats = make(map[string]int64)
	m.GroupStatsToday = make(map[string]int64)
	m.UserStatsToday = make(map[string]int64)

	// 1. 加载系统统计
	rows, err := m.DB.Query(m.PrepareQuery("SELECT key, value FROM system_stats"))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err == nil {
				if key == "total_messages" {
					fmt.Sscanf(value, "%d", &m.TotalMessages)
				} else if key == "sent_messages" {
					fmt.Sscanf(value, "%d", &m.SentMessages)
				}
			}
		}
	}

	// 2. 加载群组/用户全量统计
	rows, err = m.DB.Query(m.PrepareQuery("SELECT id, count FROM group_stats"))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var count int64
			if err := rows.Scan(&id, &count); err == nil {
				m.GroupStats[id] = count
			}
		}
	}

	rows, err = m.DB.Query(m.PrepareQuery("SELECT id, count FROM user_stats"))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var count int64
			if err := rows.Scan(&id, &count); err == nil {
				m.UserStats[id] = count
			}
		}
	}

	// 3. 加载今日统计
	today := time.Now().Format("2006-01-02")
	m.LastResetDate = today // 初始化重置日期
	rows, err = m.DB.Query(m.PrepareQuery("SELECT id, count FROM group_stats_today WHERE day = ?"), today)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var count int64
			if err := rows.Scan(&id, &count); err == nil {
				m.GroupStatsToday[id] = count
			}
		}
	}

	rows, err = m.DB.Query(m.PrepareQuery("SELECT id, count FROM user_stats_today WHERE day = ?"), today)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var count int64
			if err := rows.Scan(&id, &count); err == nil {
				m.UserStatsToday[id] = count
			}
		}
	}

	return nil
}

// SaveGroupToDB 保存群组到数据库
func (m *Manager) SaveGroupToDB(groupID, groupName, botID string) error {
	query := `
	INSERT INTO group_cache (group_id, group_name, bot_id, last_seen)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(group_id) DO UPDATE SET
		group_name = EXCLUDED.group_name,
		bot_id = EXCLUDED.bot_id,
		last_seen = EXCLUDED.last_seen;
	`
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(query), groupID, groupName, botID, now)
	return err
}

// SaveFriendToDB 保存好友到数据库
func (m *Manager) SaveFriendToDB(userID, nickname, botID string) error {
	query := `
	INSERT INTO friend_cache (user_id, nickname, bot_id, last_seen)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		bot_id = EXCLUDED.bot_id,
		last_seen = EXCLUDED.last_seen;
	`
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(query), userID, nickname, botID, now)
	return err
}

// SaveMemberToDB 保存群成员到数据库
func (m *Manager) SaveMemberToDB(groupID, userID, nickname, card string) error {
	query := `
	INSERT INTO member_cache (group_id, user_id, nickname, card, last_seen)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(group_id, user_id) DO UPDATE SET
		nickname = EXCLUDED.nickname,
		card = EXCLUDED.card,
		last_seen = EXCLUDED.last_seen;
	`
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(query), groupID, userID, nickname, card, now)
	return err
}

// LoadCachesFromDB 从数据库加载所有缓存到内存
func (m *Manager) LoadCachesFromDB() error {
	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// 1. 加载群组
	rows, err := m.DB.Query(m.PrepareQuery("SELECT group_id, group_name, bot_id FROM group_cache"))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var gID, name, botID string
			if err := rows.Scan(&gID, &name, &botID); err == nil {
				m.GroupCache[gID] = GroupInfo{
					GroupID:   gID,
					GroupName: name,
					BotID:     botID,
					IsCached:  true,
					LastSeen:  time.Now(),
				}
			}
		}
	}

	// 2. 加载好友
	rowsF, err := m.DB.Query(m.PrepareQuery("SELECT user_id, nickname, bot_id FROM friend_cache"))
	if err == nil {
		defer rowsF.Close()
		for rowsF.Next() {
			var uID, nickname, botID string
			if err := rowsF.Scan(&uID, &nickname, &botID); err == nil {
				m.FriendCache[uID] = FriendInfo{
					UserID:   uID,
					Nickname: nickname,
					BotID:    botID,
					IsCached: true,
					LastSeen: time.Now(),
				}
			}
		}
	}

	// 3. 加载群成员
	rowsM, err := m.DB.Query(m.PrepareQuery("SELECT group_id, user_id, nickname, card FROM member_cache"))
	if err == nil {
		defer rowsM.Close()
		for rowsM.Next() {
			var gID, uID, nickname, card string
			if err := rowsM.Scan(&gID, &uID, &nickname, &card); err == nil {
				key := fmt.Sprintf("%s:%s", gID, uID)
				m.MemberCache[key] = MemberInfo{
					GroupID:  gID,
					UserID:   uID,
					Nickname: nickname,
					Card:     card,
					IsCached: true,
					LastSeen: time.Now(),
				}
			}
		}
	}

	log.Printf("[INFO] 从数据库加载了 %d 个群组, %d 个好友, %d 个成员缓存", len(m.GroupCache), len(m.FriendCache), len(m.MemberCache))
	return nil
}

// LoadRoutingRulesFromDB 从数据库加载所有路由规则到内存缓存
func (m *Manager) LoadRoutingRulesFromDB() error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	rows, err := m.DB.Query(m.PrepareQuery("SELECT pattern, target_worker_id FROM routing_rules"))
	if err != nil {
		return err
	}
	defer rows.Close()

	m.RoutingRules = make(map[string]string)
	count := 0
	for rows.Next() {
		var pattern, target string
		if err := rows.Scan(&pattern, &target); err != nil {
			log.Printf("[ERROR] 解析路由规则行失败: %v", err)
			continue
		}
		m.RoutingRules[pattern] = target
		count++
	}
	log.Printf("[INFO] 从数据库加载了 %d 条路由规则", count)
	return nil
}

// SaveAllStatsToDB 保存所有内存中的统计数据到数据库
func (m *Manager) SaveAllStatsToDB() {
	m.StatsMutex.RLock()
	defer m.StatsMutex.RUnlock()

	tx, err := m.DB.Begin()
	if err != nil {
		log.Printf("[DB] 开始事务失败: %v", err)
		return
	}
	defer tx.Rollback()

	now := time.Now()
	today := time.Now().Format("2006-01-02")

	// 1. 保存全量群组统计
	for id, count := range m.GroupStats {
		_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO group_stats (id, count, updated_at) VALUES (?, ?, ?) 
			ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at`),
			id, count, now)
	}

	// 2. 保存全量用户统计
	for id, count := range m.UserStats {
		_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO user_stats (id, count, updated_at) VALUES (?, ?, ?) 
			ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at`),
			id, count, now)
	}

	// 3. 保存今日群组统计
	for id, count := range m.GroupStatsToday {
		_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO group_stats_today (id, count, day, updated_at) VALUES (?, ?, ?, ?) 
			ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at, day = EXCLUDED.day`),
			id, count, today, now)
	}

	// 4. 保存今日用户统计
	for id, count := range m.UserStatsToday {
		_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO user_stats_today (id, count, day, updated_at) VALUES (?, ?, ?, ?) 
			ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at, day = EXCLUDED.day`),
			id, count, today, now)
	}

	// 5. 保存基本统计
	_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO system_stats (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value`),
		"total_messages", fmt.Sprintf("%d", m.TotalMessages), now)
	_, _ = tx.Exec(m.PrepareQuery(`INSERT INTO system_stats (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value`),
		"sent_messages", fmt.Sprintf("%d", m.SentMessages), now)

	if err := tx.Commit(); err != nil {
		log.Printf("[DB] 提交事务失败: %v", err)
	}
}

// SaveRoutingRuleToDB 保存路由规则到数据库
func (m *Manager) SaveRoutingRuleToDB(pattern, target string) error {
	query := `
	INSERT INTO routing_rules (pattern, target_worker_id, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(pattern) DO UPDATE SET
		target_worker_id = EXCLUDED.target_worker_id,
		updated_at = EXCLUDED.updated_at;
	`
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(query), pattern, target, now, now)
	return err
}

// DeleteRoutingRuleFromDB 从数据库删除路由规则
func (m *Manager) DeleteRoutingRuleFromDB(pattern string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM routing_rules WHERE pattern = ?"), pattern)
	return err
}

// LoadUsersFromDB 从数据库加载所有用户到内存缓存
func (m *Manager) LoadUsersFromDB() error {
	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()
	return m.LoadUsersFromDBNoLock()
}

// LoadUsersFromDBNoLock 从数据库加载所有用户到内存缓存 (无锁版本)
func (m *Manager) LoadUsersFromDBNoLock() error {
	rows, err := m.DB.Query(m.PrepareQuery("SELECT id, username, password_hash, is_admin, session_version, active, created_at, updated_at FROM users"))
	if err != nil {
		return err
	}
	defer rows.Close()

	// 清空当前内存缓存并重新加载
	m.Users = make(map[string]*User)

	for rows.Next() {
		var user User
		var createdAt, updatedAt any
		err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.SessionVersion, &user.Active, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("解析用户行失败: %v", err)
			continue
		}

		// 处理时间字段，兼容不同数据库驱动返回的类型
		if createdAt != nil {
			switch v := createdAt.(type) {
			case time.Time:
				user.CreatedAt = v
			case string:
				user.CreatedAt, _ = time.Parse(time.RFC3339, v)
			}
		}
		if updatedAt != nil {
			switch v := updatedAt.(type) {
			case time.Time:
				user.UpdatedAt = v
			case string:
				user.UpdatedAt, _ = time.Parse(time.RFC3339, v)
			}
		}

		m.Users[user.Username] = &user
	}

	log.Printf("从数据库加载了 %d 个用户", len(m.Users))
	return nil
}

// SaveUserToDB 保存或更新用户信息到数据库
func (m *Manager) SaveUserToDB(user *User) error {
	query := `
	INSERT INTO users (username, password_hash, is_admin, session_version, active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(username) DO UPDATE SET
		password_hash = EXCLUDED.password_hash,
		is_admin = EXCLUDED.is_admin,
		session_version = EXCLUDED.session_version,
		active = EXCLUDED.active,
		updated_at = EXCLUDED.updated_at;
	`

	_, err := m.DB.Exec(m.PrepareQuery(query),
		user.Username,
		user.PasswordHash,
		user.IsAdmin,
		user.SessionVersion,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// DeleteUserFromDB 从数据库删除用户
func (m *Manager) DeleteUserFromDB(username string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM users WHERE username = ?"), username)
	return err
}

// DeleteUser 从数据库删除用户
func (m *Manager) DeleteUser(username string) error {
	return m.DeleteUserFromDB(username)
}

// DeleteRoutingRule 从数据库删除路由规则
func (m *Manager) DeleteRoutingRule(pattern string) error {
	return m.DeleteRoutingRuleFromDB(pattern)
}

// Transaction 原生SQL事务包装器
// SaveSystemStat 保存单个系统统计到数据库
func (m *Manager) SaveSystemStat(key string, value any) error {
	return m.SaveStatToDB(key, value)
}

// LoadSystemStat 从数据库加载单个系统统计
func (m *Manager) LoadSystemStat(key string) (any, error) {
	var value string
	err := m.DB.QueryRow(m.PrepareQuery("SELECT value FROM system_stats WHERE key = ?"), key).Scan(&value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// LoadSystemStatsFromDB 从数据库加载所有系统统计
func (m *Manager) LoadSystemStatsFromDB() (map[string]any, error) {
	rows, err := m.DB.Query(m.PrepareQuery("SELECT key, value FROM system_stats"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]any)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			stats[key] = value
		}
	}
	return stats, nil
}

// SaveGroupStats 保存群组统计
func (m *Manager) SaveGroupStats(id string, count int64) error {
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(`INSERT INTO group_stats (id, count, updated_at) VALUES (?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at`),
		id, count, now)
	return err
}

// LoadGroupStats 加载群组统计
func (m *Manager) LoadGroupStats(id string) (int64, error) {
	var count int64
	err := m.DB.QueryRow(m.PrepareQuery("SELECT count FROM group_stats WHERE id = ?"), id).Scan(&count)
	return count, err
}

// SaveUserStats 保存用户统计
func (m *Manager) SaveUserStats(id string, count int64) error {
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(`INSERT INTO user_stats (id, count, updated_at) VALUES (?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at`),
		id, count, now)
	return err
}

// LoadUserStats 加载用户统计
func (m *Manager) LoadUserStats(id string) (int64, error) {
	var count int64
	err := m.DB.QueryRow(m.PrepareQuery("SELECT count FROM user_stats WHERE id = ?"), id).Scan(&count)
	return count, err
}

// SaveGroupStatsToday 保存今日群组统计
func (m *Manager) SaveGroupStatsToday(id string, day string, count int64) error {
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(`INSERT INTO group_stats_today (id, count, day, updated_at) VALUES (?, ?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at, day = EXCLUDED.day`),
		id, count, day, now)
	return err
}

// LoadGroupStatsToday 加载今日群组统计
func (m *Manager) LoadGroupStatsToday(id string, day string) (int64, error) {
	var count int64
	err := m.DB.QueryRow(m.PrepareQuery("SELECT count FROM group_stats_today WHERE id = ? AND day = ?"), id, day).Scan(&count)
	return count, err
}

// SaveUserStatsToday 保存今日用户统计
func (m *Manager) SaveUserStatsToday(id string, day string, count int64) error {
	now := time.Now()
	_, err := m.DB.Exec(m.PrepareQuery(`INSERT INTO user_stats_today (id, count, day, updated_at) VALUES (?, ?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET count = EXCLUDED.count, updated_at = EXCLUDED.updated_at, day = EXCLUDED.day`),
		id, count, day, now)
	return err
}

// LoadUserStatsToday 加载今日用户统计
func (m *Manager) LoadUserStatsToday(id string, day string) (int64, error) {
	var count int64
	err := m.DB.QueryRow(m.PrepareQuery("SELECT count FROM user_stats_today WHERE id = ? AND day = ?"), id, day).Scan(&count)
	return count, err
}

// DeleteSystemStat 从数据库删除系统统计
func (m *Manager) DeleteSystemStat(key string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM system_stats WHERE key = ?"), key)
	return err
}

// DeleteUserStatsToday 从数据库删除今日用户统计
func (m *Manager) DeleteUserStatsToday(id string, day string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM user_stats_today WHERE id = ? AND day = ?"), id, day)
	return err
}

// DeleteGroupStats 从数据库删除群组统计
func (m *Manager) DeleteGroupStats(id string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM group_stats WHERE id = ?"), id)
	return err
}

// DeleteUserStats 从数据库删除用户统计
func (m *Manager) DeleteUserStats(id string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM user_stats WHERE id = ?"), id)
	return err
}

// DeleteGroupStatsToday 从数据库删除今日群组统计
func (m *Manager) DeleteGroupStatsToday(id string, day string) error {
	_, err := m.DB.Exec(m.PrepareQuery("DELETE FROM group_stats_today WHERE id = ? AND day = ?"), id, day)
	return err
}

func (m *Manager) Transaction(fn func(tx *Manager) error) error {
	if m.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	// 创建一个临时的 Manager 用于事务操作
	txManager := &Manager{
		DB:              m.DB, // 这里实际上应该用事务对象，但为了简化兼容性，原生SQL回退暂时不支持真事务嵌套
		Users:           m.Users,
		RoutingRules:    m.RoutingRules,
		GroupStats:      m.GroupStats,
		UserStats:       m.UserStats,
		GroupStatsToday: m.GroupStatsToday,
		UserStatsToday:  m.UserStatsToday,
	}

	err = fn(txManager)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// InitDefaultAdmin 初始化默认管理员账号
func (m *Manager) InitDefaultAdmin() {
	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()

	if _, ok := m.Users["admin"]; !ok {
		log.Printf("未找到管理员账号，正在创建默认管理员 admin...")
		now := time.Now()

		// 默认密码为 admin123
		hash, err := HashPassword("admin123")
		if err != nil {
			log.Printf("生成默认管理员密码哈希失败: %v", err)
			return
		}

		admin := &User{
			Username:       "admin",
			PasswordHash:   hash,
			IsAdmin:        true,
			Active:         true,
			SessionVersion: 1,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		m.Users["admin"] = admin
		if m.DB != nil {
			if err := m.SaveUserToDB(admin); err != nil {
				log.Printf("创建默认管理员失败: %v", err)
			} else {
				log.Printf("默认管理员账号 admin 创建成功 (默认密码: admin123)")
			}
		} else {
			log.Printf("数据库未初始化，默认管理员已存入内存")
		}
	}
}
