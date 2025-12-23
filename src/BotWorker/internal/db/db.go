package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"botworker/internal/config"

	_ "github.com/lib/pq"
)

// NewDBConnection 创建一个新的数据库连接
func NewDBConnection(cfg *config.DatabaseConfig) (*sql.DB, error) {
	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// 打开数据库连接
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库连接: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("无法连接到数据库: %w", err)
	}

	return db, nil
}

// InitDatabase 初始化数据库，创建必要的表
func InitDatabase(db *sql.DB) error {
	// 创建用户表（将积分字段直接存储在用户表中）
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL UNIQUE,
		nickname VARCHAR(255),
		avatar VARCHAR(255),
		gender VARCHAR(10),
		points INTEGER NOT NULL DEFAULT 0,
		savings_points INTEGER NOT NULL DEFAULT 0,
		savings_last_interest_at TIMESTAMP,
		frozen_points INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建消息记录表
	messageTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message_id VARCHAR(255) NOT NULL UNIQUE,
		user_id VARCHAR(255),
		group_id VARCHAR(255),
		type VARCHAR(50) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建会话状态表
	sessionTableSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		session_id VARCHAR(255) NOT NULL UNIQUE,
		user_id VARCHAR(255),
		group_id VARCHAR(255),
		state VARCHAR(255),
		data JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建群管理员表
	groupAdminsTableSQL := `
	CREATE TABLE IF NOT EXISTS group_admins (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		level INTEGER NOT NULL DEFAULT 1, -- 权限级别: 1=普通管理员, 2=超级管理员
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	// 创建群规表
	groupRulesTableSQL := `
	CREATE TABLE IF NOT EXISTS group_rules (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL UNIQUE,
		rules TEXT NOT NULL,
		voice_id VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	sensitiveWordsTableSQL := `
	CREATE TABLE IF NOT EXISTS sensitive_words (
		id SERIAL PRIMARY KEY,
		word VARCHAR(255) NOT NULL UNIQUE,
		level INTEGER NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建禁言记录表
	bannedUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS banned_users (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		ban_end_time TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	// 创建审核日志表
	auditLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		admin_id VARCHAR(255) NOT NULL,
		action VARCHAR(50) NOT NULL,
		target_user_id VARCHAR(255),
		target_group_id VARCHAR(255),
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	groupFeaturesTableSQL := `
	CREATE TABLE IF NOT EXISTS group_features (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		feature_id VARCHAR(100) NOT NULL,
		enabled BOOLEAN NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, feature_id)
	);
	`

	groupWhitelistTableSQL := `
	CREATE TABLE IF NOT EXISTS group_whitelist (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	petsTableSQL := `
	CREATE TABLE IF NOT EXISTS pets (
		id SERIAL PRIMARY KEY,
		pet_id VARCHAR(255) NOT NULL UNIQUE,
		user_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		level INTEGER NOT NULL DEFAULT 1,
		exp INTEGER NOT NULL DEFAULT 0,
		hunger INTEGER NOT NULL DEFAULT 100,
		happiness INTEGER NOT NULL DEFAULT 100,
		health INTEGER NOT NULL DEFAULT 100,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	pointsLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS points_logs (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		amount INTEGER NOT NULL,
		reason VARCHAR(255),
		category VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	questionsTableSQL := `
	CREATE TABLE IF NOT EXISTS questions (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL,
		question_raw TEXT NOT NULL,
		question_normalized TEXT NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'approved',
		created_by VARCHAR(255),
		source_group_id VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usage_count INTEGER NOT NULL DEFAULT 0,
		UNIQUE (question_normalized)
	);
	`

	answersTableSQL := `
	CREATE TABLE IF NOT EXISTS answers (
		id SERIAL PRIMARY KEY,
		question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
		answer TEXT NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'approved',
		created_by VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usage_count INTEGER NOT NULL DEFAULT 0,
		short_interval_usage_count INTEGER NOT NULL DEFAULT 0,
		last_used_at TIMESTAMP
	);
	`

	groupAISettingsTableSQL := `
	CREATE TABLE IF NOT EXISTS group_ai_settings (
		id SERIAL PRIMARY KEY,
		group_id VARCHAR(255) NOT NULL UNIQUE,
		qa_mode VARCHAR(50) NOT NULL,
		last_answer_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 执行建表语句
	if _, err := db.Exec(userTableSQL); err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS savings_points INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("为用户表添加存积分字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS savings_last_interest_at TIMESTAMP`); err != nil {
		return fmt.Errorf("为用户表添加存积分利息时间字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS frozen_points INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("为用户表添加冻结积分字段失败: %w", err)
	}

	if _, err := db.Exec(messageTableSQL); err != nil {
		return fmt.Errorf("创建消息记录表失败: %w", err)
	}

	if _, err := db.Exec(sessionTableSQL); err != nil {
		return fmt.Errorf("创建会话状态表失败: %w", err)
	}

	if _, err := db.Exec(groupAdminsTableSQL); err != nil {
		return fmt.Errorf("创建群管理员表失败: %w", err)
	}

	if _, err := db.Exec(groupRulesTableSQL); err != nil {
		return fmt.Errorf("创建群规表失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE group_rules ADD COLUMN IF NOT EXISTS voice_id VARCHAR(255)`); err != nil {
		return fmt.Errorf("为群规表添加语音配置字段失败: %w", err)
	}

	if _, err := db.Exec(sensitiveWordsTableSQL); err != nil {
		return fmt.Errorf("创建敏感词表失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE sensitive_words ADD COLUMN IF NOT EXISTS level INTEGER NOT NULL DEFAULT 1`); err != nil {
		return fmt.Errorf("为敏感词表添加级别字段失败: %w", err)
	}

	if _, err := db.Exec(bannedUsersTableSQL); err != nil {
		return fmt.Errorf("创建禁言记录表失败: %w", err)
	}

	if _, err := db.Exec(auditLogsTableSQL); err != nil {
		return fmt.Errorf("创建审核日志表失败: %w", err)
	}

	if _, err := db.Exec(petsTableSQL); err != nil {
		return fmt.Errorf("创建宠物表失败: %w", err)
	}

	if _, err := db.Exec(pointsLogsTableSQL); err != nil {
		return fmt.Errorf("创建积分记录表失败: %w", err)
	}

	if _, err := db.Exec(groupFeaturesTableSQL); err != nil {
		return fmt.Errorf("创建群功能开关表失败: %w", err)
	}

	if _, err := db.Exec(groupWhitelistTableSQL); err != nil {
		return fmt.Errorf("创建群白名单表失败: %w", err)
	}

	if _, err := db.Exec(questionsTableSQL); err != nil {
		return fmt.Errorf("创建问题表失败: %w", err)
	}

	if _, err := db.Exec(answersTableSQL); err != nil {
		return fmt.Errorf("创建答案表失败: %w", err)
	}

	if _, err := db.Exec(groupAISettingsTableSQL); err != nil {
		return fmt.Errorf("创建群AI设置表失败: %w", err)
	}

	// 兼容旧表结构，补充缺失的分类字段
	if _, err := db.Exec(`ALTER TABLE points_logs ADD COLUMN IF NOT EXISTS category VARCHAR(100)`); err != nil {
		return fmt.Errorf("为积分记录表添加分类字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE questions ADD COLUMN IF NOT EXISTS usage_count INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("为问题表添加使用次数字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE answers ADD COLUMN IF NOT EXISTS usage_count INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("为答案表添加使用次数字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE answers ADD COLUMN IF NOT EXISTS short_interval_usage_count INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("为答案表添加短间隔使用次数字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE answers ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP`); err != nil {
		return fmt.Errorf("为答案表添加最后使用时间字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE group_ai_settings ADD COLUMN IF NOT EXISTS last_answer_id INTEGER`); err != nil {
		return fmt.Errorf("为群AI设置表添加最后答案ID字段失败: %w", err)
	}

	return nil
}

// User 定义用户模型
type User struct {
	ID            int       `json:"id"`
	UserID        string    `json:"user_id"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	Gender        string    `json:"gender"`
	Points        int       `json:"points"`
	SavingsPoints int       `json:"savings_points"`
	FrozenPoints  int       `json:"frozen_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, user *User) error {
	query := `
	INSERT INTO users (user_id, nickname, avatar, gender, points, savings_points, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET nickname = $2, avatar = $3, gender = $4, points = $5, savings_points = $6, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, user.UserID, user.Nickname, user.Avatar, user.Gender, user.Points, user.SavingsPoints)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByUserID 根据用户ID获取用户信息
func GetUserByUserID(db *sql.DB, userID string) (*User, error) {
	query := `
	SELECT id, user_id, nickname, avatar, gender, points, savings_points, frozen_points, created_at, updated_at
	FROM users
	WHERE user_id = $1
	`

	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(db *sql.DB, user *User) error {
	query := `
	UPDATE users
	SET nickname = $2, avatar = $3, gender = $4, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = $1
	`

	result, err := db.Exec(query, user.UserID, user.Nickname, user.Avatar, user.Gender)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在: %s", user.UserID)
	}

	return nil
}

// Message 定义消息记录模型
type Message struct {
	ID        int       `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	GroupID   string    `json:"group_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateMessage 创建消息记录
func CreateMessage(db *sql.DB, message *Message) error {
	query := `
	INSERT INTO messages (message_id, user_id, group_id, type, content)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (message_id) DO NOTHING
	`

	_, err := db.Exec(query, message.MessageID, message.UserID, message.GroupID, message.Type, message.Content)
	if err != nil {
		return fmt.Errorf("创建消息记录失败: %w", err)
	}

	return nil
}

// GetMessagesByUserID 根据用户ID获取消息记录
func GetMessagesByUserID(db *sql.DB, userID string, limit int) ([]*Message, error) {
	query := `
	SELECT id, message_id, user_id, group_id, type, content, created_at
	FROM messages
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2
	`

	rows, err := db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取消息记录失败: %w", err)
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.UserID, &message.GroupID, &message.Type, &message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描消息记录失败: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息记录失败: %w", err)
	}

	return messages, nil
}

// Session 定义会话状态模型
type Session struct {
	ID        int                    `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	GroupID   string                 `json:"group_id"`
	State     string                 `json:"state"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateOrUpdateSession 创建或更新会话状态
func CreateOrUpdateSession(db *sql.DB, session *Session) error {
	query := `
	INSERT INTO sessions (session_id, user_id, group_id, state, data, updated_at)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	ON CONFLICT (session_id) DO UPDATE
	SET user_id = $2, group_id = $3, state = $4, data = $5, updated_at = CURRENT_TIMESTAMP
	`

	// 将Data转换为JSONB
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	_, err = db.Exec(query, session.SessionID, session.UserID, session.GroupID, session.State, dataJSON)
	if err != nil {
		return fmt.Errorf("创建或更新会话失败: %w", err)
	}

	return nil
}

// CreateUserTx 在事务中创建新用户
func CreateUserTx(tx *sql.Tx, user *User) error {
	query := `
	INSERT INTO users (user_id, nickname, avatar, gender, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET nickname = $2, avatar = $3, gender = $4, updated_at = CURRENT_TIMESTAMP
	`

	_, err := tx.Exec(query, user.UserID, user.Nickname, user.Avatar, user.Gender)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// CreateOrUpdateSessionTx 在事务中创建或更新会话状态
func CreateOrUpdateSessionTx(tx *sql.Tx, session *Session) error {
	query := `
	INSERT INTO sessions (session_id, user_id, group_id, state, data, updated_at)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	ON CONFLICT (session_id) DO UPDATE
	SET user_id = $2, group_id = $3, state = $4, data = $5, updated_at = CURRENT_TIMESTAMP
	`

	// 将Data转换为JSONB
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	_, err = tx.Exec(query, session.SessionID, session.UserID, session.GroupID, session.State, dataJSON)
	if err != nil {
		return fmt.Errorf("创建或更新会话失败: %w", err)
	}

	return nil
}

// GetSessionBySessionID 根据会话ID获取会话状态
func GetSessionBySessionID(db *sql.DB, sessionID string) (*Session, error) {
	query := `
	SELECT id, session_id, user_id, group_id, state, data, created_at, updated_at
	FROM sessions
	WHERE session_id = $1
	`

	session := &Session{}
	var dataJSON []byte

	err := db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.GroupID, &session.State, &dataJSON, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 会话不存在
		}
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 解析JSONB数据
	if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	return session, nil
}

// DeleteSession 删除会话状态
func DeleteSession(db *sql.DB, sessionID string) error {
	query := `
	DELETE FROM sessions
	WHERE session_id = $1
	`

	result, err := db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	return nil
}

// ------------------- 用户相关扩展操作 -------------------

// GetAllUsers 获取所有用户
func GetAllUsers(db *sql.DB) ([]*User, error) {
	query := `
	SELECT id, user_id, nickname, avatar, gender, points, savings_points, frozen_points, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取所有用户失败: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描用户失败: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户失败: %w", err)
	}

	return users, nil
}

// GetUsersWithPagination 分页获取用户
func GetUsersWithPagination(db *sql.DB, page, pageSize int) ([]*User, int, error) {
	// 获取总记录数
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	if err := db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取分页数据
	query := `
	SELECT id, user_id, nickname, avatar, gender, points, savings_points, frozen_points, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("分页获取用户失败: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描用户失败: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历用户失败: %w", err)
	}

	return users, total, nil
}

// DeleteUser 删除用户
func DeleteUser(db *sql.DB, userID string) error {
	query := `
	DELETE FROM users
	WHERE user_id = $1
	`

	result, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在: %s", userID)
	}

	return nil
}

// ------------------- 消息相关扩展操作 -------------------

// GetMessageByID 根据消息ID获取消息
func GetMessageByID(db *sql.DB, messageID string) (*Message, error) {
	query := `
	SELECT id, message_id, user_id, group_id, type, content, created_at
	FROM messages
	WHERE message_id = $1
	`

	message := &Message{}
	err := db.QueryRow(query, messageID).Scan(
		&message.ID, &message.MessageID, &message.UserID, &message.GroupID, &message.Type, &message.Content, &message.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 消息不存在
		}
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	return message, nil
}

// GetMessagesByGroupID 根据群组ID获取消息
func GetMessagesByGroupID(db *sql.DB, groupID string, limit int) ([]*Message, error) {
	query := `
	SELECT id, message_id, user_id, group_id, type, content, created_at
	FROM messages
	WHERE group_id = $1
	ORDER BY created_at DESC
	LIMIT $2
	`

	rows, err := db.Query(query, groupID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取群组消息失败: %w", err)
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.UserID, &message.GroupID, &message.Type, &message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描消息失败: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息失败: %w", err)
	}

	return messages, nil
}

// GetLatestMessages 获取最新消息
func GetLatestMessages(db *sql.DB, limit int) ([]*Message, error) {
	query := `
	SELECT id, message_id, user_id, group_id, type, content, created_at
	FROM messages
	ORDER BY created_at DESC
	LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("获取最新消息失败: %w", err)
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.UserID, &message.GroupID, &message.Type, &message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描消息失败: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息失败: %w", err)
	}

	return messages, nil
}

// DeleteMessage 删除消息
func DeleteMessage(db *sql.DB, messageID string) error {
	query := `
	DELETE FROM messages
	WHERE message_id = $1
	`

	result, err := db.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("消息不存在: %s", messageID)
	}

	return nil
}

// ------------------- 会话相关扩展操作 -------------------

// GetSessionsByUserID 根据用户ID获取会话
func GetSessionsByUserID(db *sql.DB, userID string) ([]*Session, error) {
	query := `
	SELECT id, session_id, user_id, group_id, state, data, created_at, updated_at
	FROM sessions
	WHERE user_id = $1
	ORDER BY updated_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户会话失败: %w", err)
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		session := &Session{}
		var dataJSON []byte
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.UserID, &session.GroupID, &session.State, &dataJSON, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描会话失败: %w", err)
		}

		// 解析JSONB数据
		if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
			return nil, fmt.Errorf("解析会话数据失败: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话失败: %w", err)
	}

	return sessions, nil
}

// GetSessionsByGroupID 根据群组ID获取会话
func GetSessionsByGroupID(db *sql.DB, groupID string) ([]*Session, error) {
	query := `
	SELECT id, session_id, user_id, group_id, state, data, created_at, updated_at
	FROM sessions
	WHERE group_id = $1
	ORDER BY updated_at DESC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("获取群组会话失败: %w", err)
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		session := &Session{}
		var dataJSON []byte
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.UserID, &session.GroupID, &session.State, &dataJSON, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描会话失败: %w", err)
		}

		// 解析JSONB数据
		if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
			return nil, fmt.Errorf("解析会话数据失败: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话失败: %w", err)
	}

	return sessions, nil
}

// GetAllSessions 获取所有会话
func GetAllSessions(db *sql.DB) ([]*Session, error) {
	query := `
	SELECT id, session_id, user_id, group_id, state, data, created_at, updated_at
	FROM sessions
	ORDER BY updated_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取所有会话失败: %w", err)
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		session := &Session{}
		var dataJSON []byte
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.UserID, &session.GroupID, &session.State, &dataJSON, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描会话失败: %w", err)
		}

		// 解析JSONB数据
		if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
			return nil, fmt.Errorf("解析会话数据失败: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话失败: %w", err)
	}

	return sessions, nil
}

// ------------------- 统计相关操作 -------------------

// GetUserCount 获取用户数量
func GetUserCount(db *sql.DB) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users"
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取用户数量失败: %w", err)
	}
	return count, nil
}

// ------------------- 宠物相关操作 -------------------

// Pet 定义宠物模型（与 plugins/pets.go 中的 Pet 保持一致，但可能需要根据 db 需求调整）
type PetModel struct {
	ID        int       `json:"id"`
	PetID     string    `json:"pet_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Level     int       `json:"level"`
	Exp       int       `json:"exp"`
	Hunger    int       `json:"hunger"`
	Happiness int       `json:"happiness"`
	Health    int       `json:"health"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupAISettings struct {
	ID           int       `json:"id"`
	GroupID      string    `json:"group_id"`
	QAMode       string    `json:"qa_mode"`
	LastAnswerID int       `json:"last_answer_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Question struct {
	ID                 int       `json:"id"`
	GroupID            string    `json:"group_id"`
	QuestionRaw        string    `json:"question_raw"`
	QuestionNormalized string    `json:"question_normalized"`
	Status             string    `json:"status"`
	CreatedBy          string    `json:"created_by"`
	SourceGroupID      string    `json:"source_group_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Answer struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question_id"`
	Answer     string    `json:"answer"`
	Status     string    `json:"status"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func GetGroupQAMode(db *sql.DB, groupID string) (string, error) {
	if db == nil || groupID == "" {
		return "", nil
	}

	query := `
	SELECT qa_mode
	FROM group_ai_settings
	WHERE group_id = $1
	`

	var mode string
	err := db.QueryRow(query, groupID).Scan(&mode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("获取群问答模式失败: %w", err)
	}

	return mode, nil
}

func SetGroupQAMode(db *sql.DB, groupID string, mode string) error {
	if db == nil || groupID == "" {
		return fmt.Errorf("数据库或群ID为空")
	}

	query := `
	INSERT INTO group_ai_settings (group_id, qa_mode, created_at, updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id) DO UPDATE
	SET qa_mode = EXCLUDED.qa_mode,
	    updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, groupID, mode)
	if err != nil {
		return fmt.Errorf("设置群问答模式失败: %w", err)
	}

	return nil
}

func GetGroupLastAnswerID(db *sql.DB, groupID string) (int, error) {
	if db == nil || groupID == "" {
		return 0, nil
	}

	query := `
	SELECT COALESCE(last_answer_id, 0)
	FROM group_ai_settings
	WHERE group_id = $1
	`

	var lastID int
	err := db.QueryRow(query, groupID).Scan(&lastID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("获取群最后答案ID失败: %w", err)
	}

	return lastID, nil
}

func SetGroupLastAnswerID(db *sql.DB, groupID string, answerID int) error {
	if db == nil || groupID == "" {
		return fmt.Errorf("数据库或群ID为空")
	}

	query := `
	INSERT INTO group_ai_settings (group_id, qa_mode, last_answer_id, created_at, updated_at)
	VALUES ($1, 'group', $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id) DO UPDATE
	SET last_answer_id = EXCLUDED.last_answer_id,
	    updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, groupID, answerID)
	if err != nil {
		return fmt.Errorf("设置群最后答案ID失败: %w", err)
	}

	return nil
}

func GetQuestionByGroupAndNormalized(dbConn *sql.DB, groupID string, normalized string) (*Question, error) {
	if dbConn == nil || normalized == "" {
		return nil, nil
	}

	query := `
	SELECT id, group_id, question_raw, question_normalized, status, created_by, source_group_id, created_at, updated_at
	FROM questions
	WHERE question_normalized = $1
	`

	q := &Question{}
	err := dbConn.QueryRow(query, normalized).Scan(
		&q.ID,
		&q.GroupID,
		&q.QuestionRaw,
		&q.QuestionNormalized,
		&q.Status,
		&q.CreatedBy,
		&q.SourceGroupID,
		&q.CreatedAt,
		&q.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取问题失败: %w", err)
	}

	return q, nil
}

func CreateQuestion(dbConn *sql.DB, q *Question) (*Question, error) {
	if dbConn == nil || q == nil {
		return nil, fmt.Errorf("数据库或问题为空")
	}

	if q.Status == "" {
		q.Status = "approved"
	}

	query := `
	INSERT INTO questions (group_id, question_raw, question_normalized, status, created_by, source_group_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (question_normalized) DO UPDATE
	SET question_raw = EXCLUDED.question_raw,
	    status = EXCLUDED.status,
	    created_by = EXCLUDED.created_by,
	    source_group_id = EXCLUDED.source_group_id,
	    updated_at = CURRENT_TIMESTAMP
	RETURNING id, group_id, question_raw, question_normalized, status, created_by, source_group_id, created_at, updated_at
	`

	row := dbConn.QueryRow(query, q.GroupID, q.QuestionRaw, q.QuestionNormalized, q.Status, q.CreatedBy, q.SourceGroupID)

	var result Question
	if err := row.Scan(
		&result.ID,
		&result.GroupID,
		&result.QuestionRaw,
		&result.QuestionNormalized,
		&result.Status,
		&result.CreatedBy,
		&result.SourceGroupID,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("创建或更新问题失败: %w", err)
	}

	return &result, nil
}

func AddAnswer(dbConn *sql.DB, a *Answer) (*Answer, error) {
	if dbConn == nil || a == nil {
		return nil, fmt.Errorf("数据库或答案为空")
	}

	if a.Status == "" {
		a.Status = "approved"
	}

	query := `
	INSERT INTO answers (question_id, answer, status, created_by, created_at, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING id, question_id, answer, status, created_by, created_at, updated_at
	`

	row := dbConn.QueryRow(query, a.QuestionID, a.Answer, a.Status, a.CreatedBy)

	var result Answer
	if err := row.Scan(
		&result.ID,
		&result.QuestionID,
		&result.Answer,
		&result.Status,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("创建答案失败: %w", err)
	}

	return &result, nil
}

func GetApprovedAnswersByQuestionID(dbConn *sql.DB, questionID int) ([]*Answer, error) {
	if dbConn == nil || questionID == 0 {
		return nil, nil
	}

	query := `
	SELECT id, question_id, answer, status, created_by, created_at, updated_at
	FROM answers
	WHERE question_id = $1 AND status = 'approved'
	ORDER BY id ASC
	`

	rows, err := dbConn.Query(query, questionID)
	if err != nil {
		return nil, fmt.Errorf("获取答案列表失败: %w", err)
	}
	defer rows.Close()

	var answers []*Answer
	for rows.Next() {
		var a Answer
		if err := rows.Scan(
			&a.ID,
			&a.QuestionID,
			&a.Answer,
			&a.Status,
			&a.CreatedBy,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描答案失败: %w", err)
		}
		answers = append(answers, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历答案失败: %w", err)
	}

	return answers, nil
}

func GetRandomApprovedAnswer(dbConn *sql.DB, questionID int) (*Answer, error) {
	answers, err := GetApprovedAnswersByQuestionID(dbConn, questionID)
	if err != nil {
		return nil, err
	}
	if len(answers) == 0 {
		return nil, nil
	}
	if len(answers) == 1 {
		return answers[0], nil
	}

	index := time.Now().UnixNano() % int64(len(answers))
	if index < 0 {
		index = -index
	}
	return answers[index], nil
}

func IncrementQuestionUsage(dbConn *sql.DB, questionID int) error {
	if dbConn == nil || questionID == 0 {
		return nil
	}

	query := `
	UPDATE questions
	SET usage_count = usage_count + 1
	WHERE id = $1
	`

	if _, err := dbConn.Exec(query, questionID); err != nil {
		return fmt.Errorf("更新问题使用次数失败: %w", err)
	}

	return nil
}

func IncrementAnswerUsage(dbConn *sql.DB, answerID int) error {
	if dbConn == nil || answerID == 0 {
		return nil
	}

	query := `
	UPDATE answers
	SET usage_count = usage_count + 1,
	    last_used_at = NOW()
	WHERE id = $1
	`

	if _, err := dbConn.Exec(query, answerID); err != nil {
		return fmt.Errorf("更新答案使用次数失败: %w", err)
	}

	return nil
}

func IncrementAnswerShortIntervalUsageIfRecent(dbConn *sql.DB, answerID int) error {
	if dbConn == nil || answerID == 0 {
		return nil
	}

	query := `
	UPDATE answers
	SET short_interval_usage_count = short_interval_usage_count + 1
	WHERE id = $1
	  AND last_used_at IS NOT NULL
	  AND last_used_at >= NOW() - INTERVAL '5 minutes'
	`

	if _, err := dbConn.Exec(query, answerID); err != nil {
		return fmt.Errorf("更新答案短间隔使用次数失败: %w", err)
	}

	return nil
}

// CreatePet 创建宠物
func CreatePet(db *sql.DB, pet *PetModel) error {
	query := `
	INSERT INTO pets (pet_id, user_id, name, type, level, exp, hunger, happiness, health, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP)
	ON CONFLICT (pet_id) DO UPDATE
	SET name = $3, level = $5, exp = $6, hunger = $7, happiness = $8, health = $9, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, pet.PetID, pet.UserID, pet.Name, pet.Type, pet.Level, pet.Exp, pet.Hunger, pet.Happiness, pet.Health)
	if err != nil {
		return fmt.Errorf("创建宠物失败: %w", err)
	}

	return nil
}

// GetPetByPetID 根据宠物ID获取宠物
func GetPetByPetID(db *sql.DB, petID string) (*PetModel, error) {
	query := `
	SELECT id, pet_id, user_id, name, type, level, exp, hunger, happiness, health, created_at, updated_at
	FROM pets
	WHERE pet_id = $1
	`

	pet := &PetModel{}
	err := db.QueryRow(query, petID).Scan(
		&pet.ID, &pet.PetID, &pet.UserID, &pet.Name, &pet.Type, &pet.Level, &pet.Exp, &pet.Hunger, &pet.Happiness, &pet.Health, &pet.CreatedAt, &pet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取宠物失败: %w", err)
	}

	return pet, nil
}

// GetPetsByUserID 获取用户的所有宠物
func GetPetsByUserID(db *sql.DB, userID string) ([]*PetModel, error) {
	query := `
	SELECT id, pet_id, user_id, name, type, level, exp, hunger, happiness, health, created_at, updated_at
	FROM pets
	WHERE user_id = $1
	ORDER BY created_at ASC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户宠物失败: %w", err)
	}
	defer rows.Close()

	var pets []*PetModel
	for rows.Next() {
		pet := &PetModel{}
		err := rows.Scan(
			&pet.ID, &pet.PetID, &pet.UserID, &pet.Name, &pet.Type, &pet.Level, &pet.Exp, &pet.Hunger, &pet.Happiness, &pet.Health, &pet.CreatedAt, &pet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描宠物失败: %w", err)
		}
		pets = append(pets, pet)
	}

	return pets, nil
}

// UpdatePet 更新宠物状态
func UpdatePet(db *sql.DB, pet *PetModel) error {
	query := `
	UPDATE pets
	SET name = $2, level = $3, exp = $4, hunger = $5, happiness = $6, health = $7, updated_at = CURRENT_TIMESTAMP
	WHERE pet_id = $1
	`

	_, err := db.Exec(query, pet.PetID, pet.Name, pet.Level, pet.Exp, pet.Hunger, pet.Happiness, pet.Health)
	if err != nil {
		return fmt.Errorf("更新宠物失败: %w", err)
	}

	return nil
}

// DeletePet 删除宠物
func DeletePet(db *sql.DB, petID string) error {
	query := "DELETE FROM pets WHERE pet_id = $1"
	_, err := db.Exec(query, petID)
	if err != nil {
		return fmt.Errorf("删除宠物失败: %w", err)
	}
	return nil
}

// GetAllPets 获取所有宠物（用于定时更新状态）
func GetAllPets(db *sql.DB) ([]*PetModel, error) {
	query := `
	SELECT id, pet_id, user_id, name, type, level, exp, hunger, happiness, health, created_at, updated_at
	FROM pets
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取所有宠物失败: %w", err)
	}
	defer rows.Close()

	var pets []*PetModel
	for rows.Next() {
		pet := &PetModel{}
		err := rows.Scan(
			&pet.ID, &pet.PetID, &pet.UserID, &pet.Name, &pet.Type, &pet.Level, &pet.Exp, &pet.Hunger, &pet.Happiness, &pet.Health, &pet.CreatedAt, &pet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描宠物失败: %w", err)
		}
		pets = append(pets, pet)
	}

	return pets, nil
}

// GetMessageCount 获取消息数量
func GetMessageCount(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages`
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, fmt.Errorf("获取消息数量失败: %w", err)
	}
	return count, nil
}

// GetSessionCount 获取会话数量
func GetSessionCount(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions`
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, fmt.Errorf("获取会话数量失败: %w", err)
	}
	return count, nil
}

// ------------------- 事务相关操作示例 -------------------

// RegisterUserWithSession 用户注册并创建会话（事务示例）
func RegisterUserWithSession(db *sql.DB, user *User, session *Session) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	// 确保事务最终会提交或回滚
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 在事务中执行用户创建
	if err = CreateUserTx(tx, user); err != nil {
		return err
	}

	// 在事务中执行会话创建
	if err = CreateOrUpdateSessionTx(tx, session); err != nil {
		return err
	}

	return err
}

// ------------------- 积分系统相关操作 -------------------

// GetPoints 获取用户积分
func GetPoints(db *sql.DB, userID string) (int, error) {
	query := `SELECT points FROM users WHERE user_id = $1`
	var points int
	err := db.QueryRow(query, userID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // 用户还没有积分记录
		}
		return 0, fmt.Errorf("获取积分失败: %w", err)
	}
	return points, nil
}

func GetFrozenPoints(db *sql.DB, userID string) (int, error) {
	query := `SELECT frozen_points FROM users WHERE user_id = $1`
	var points int
	err := db.QueryRow(query, userID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("获取冻结积分失败: %w", err)
	}
	return points, nil
}

// AddPoints 增加或扣除用户积分
func AddPoints(db *sql.DB, userID string, amount int, reason string, category string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 1. 更新用户积分
	query := `
	UPDATE users
	SET points = points + $2, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = $1
	`
	result, err := tx.Exec(query, userID, amount)
	if err != nil {
		return fmt.Errorf("更新积分失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if rowsAffected == 0 {
		// 如果用户不存在，则创建用户并设置初始积分
		insertQuery := `
		INSERT INTO users (user_id, points, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		if _, err := tx.Exec(insertQuery, userID, amount); err != nil {
			return fmt.Errorf("插入用户并初始化积分失败: %w", err)
		}
	}

	// 2. 记录日志
	logQuery := `INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(logQuery, userID, amount, reason, category)
	if err != nil {
		return fmt.Errorf("记录积分日志失败: %w", err)
	}

	return tx.Commit()
}

func applySavingsInterestTx(tx *sql.Tx, userID string) (int, error) {
	var savings int
	var lastInterest sql.NullTime

	err := tx.QueryRow("SELECT savings_points, savings_last_interest_at FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&savings, &lastInterest)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("查询存积分失败: %w", err)
	}

	now := time.Now()
	if savings <= 0 {
		_, err = tx.Exec("UPDATE users SET savings_last_interest_at = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新存积分时间失败: %w", err)
		}
		return 0, nil
	}

	var fromTime time.Time
	if lastInterest.Valid {
		fromTime = lastInterest.Time
	} else {
		fromTime = now
	}

	days := int(now.Sub(fromTime).Hours() / 24)
	if days <= 0 {
		_, err = tx.Exec("UPDATE users SET savings_last_interest_at = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新存积分时间失败: %w", err)
		}
		return 0, nil
	}

	dailyRate := 0.0005
	interest := int(float64(savings) * dailyRate * float64(days))
	if interest <= 0 {
		_, err = tx.Exec("UPDATE users SET savings_last_interest_at = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新存积分时间失败: %w", err)
		}
		return 0, nil
	}

	newSavings := savings + interest

	_, err = tx.Exec("UPDATE users SET savings_points = $2, savings_last_interest_at = $3, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, newSavings, now)
	if err != nil {
		return 0, fmt.Errorf("更新存积分利息失败: %w", err)
	}

	_, err = tx.Exec("INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)", userID, interest, "存积分利息", "saving_interest")
	if err != nil {
		return 0, fmt.Errorf("记录存积分利息失败: %w", err)
	}

	return interest, nil
}

func DepositPointsToSavings(db *sql.DB, userID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("存入积分必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var points int
	err = tx.QueryRow("SELECT points FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户不存在或没有积分")
		}
		return fmt.Errorf("查询用户积分失败: %w", err)
	}

	if points < amount {
		return fmt.Errorf("积分不足，当前积分为: %d", points)
	}

	_, err = tx.Exec("UPDATE users SET points = points - $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("扣除积分失败: %w", err)
	}

	_, err = applySavingsInterestTx(tx, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE users SET savings_points = savings_points + $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, amount)
	if err != nil {
		return fmt.Errorf("更新存积分失败: %w", err)
	}

	_, err = tx.Exec("INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)", userID, -amount, "存入积分", "saving_deposit")
	if err != nil {
		return fmt.Errorf("记录存入积分日志失败: %w", err)
	}

	return tx.Commit()
}

func WithdrawPointsFromSavings(db *sql.DB, userID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("取出积分必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	_, err = applySavingsInterestTx(tx, userID)
	if err != nil {
		return err
	}

	var balance int
	err = tx.QueryRow("SELECT savings_points FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("没有存积分记录")
		}
		return fmt.Errorf("查询存积分失败: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("存积分余额不足，当前余额为: %d", balance)
	}

	_, err = tx.Exec("UPDATE users SET savings_points = savings_points - $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID, amount)
	if err != nil {
		return fmt.Errorf("更新存积分失败: %w", err)
	}

	_, err = tx.Exec("UPDATE users SET points = points + $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("增加用户积分失败: %w", err)
	}

	_, err = tx.Exec("INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)", userID, amount, "取出积分", "saving_withdraw")
	if err != nil {
		return fmt.Errorf("记录取出积分日志失败: %w", err)
	}

	return tx.Commit()
}

func GetSavingsPoints(db *sql.DB, userID string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	_, err = applySavingsInterestTx(tx, userID)
	if err != nil {
		return 0, err
	}

	var balance int
	err = tx.QueryRow("SELECT savings_points FROM users WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			balance = 0
		} else {
			return 0, fmt.Errorf("查询存积分失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return balance, nil
}

func FreezePoints(db *sql.DB, userID string, amount int, reason string) error {
	if amount <= 0 {
		return fmt.Errorf("冻结积分数量必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var available int
	err = tx.QueryRow("SELECT points FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&available)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户不存在或没有积分")
		}
		return fmt.Errorf("查询用户积分失败: %w", err)
	}

	if available < amount {
		return fmt.Errorf("可用积分不足，当前积分为: %d", available)
	}

	_, err = tx.Exec("UPDATE users SET points = points - $1, frozen_points = frozen_points + $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("更新冻结积分失败: %w", err)
	}

	_, err = tx.Exec("INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)", userID, -amount, reason, "freeze")
	if err != nil {
		return fmt.Errorf("记录冻结积分日志失败: %w", err)
	}

	return tx.Commit()
}

func UnfreezePoints(db *sql.DB, userID string, amount int, reason string) error {
	if amount <= 0 {
		return fmt.Errorf("解冻积分数量必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var frozen int
	err = tx.QueryRow("SELECT frozen_points FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&frozen)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("没有冻结积分记录")
		}
		return fmt.Errorf("查询冻结积分失败: %w", err)
	}

	if frozen < amount {
		return fmt.Errorf("冻结积分不足，当前冻结积分为: %d", frozen)
	}

	_, err = tx.Exec("UPDATE users SET frozen_points = frozen_points - $1, points = points + $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("更新解冻积分失败: %w", err)
	}

	_, err = tx.Exec("INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)", userID, amount, reason, "unfreeze")
	if err != nil {
		return fmt.Errorf("记录解冻积分日志失败: %w", err)
	}

	return tx.Commit()
}

// TransferPoints 积分转账
func TransferPoints(db *sql.DB, fromUserID, toUserID string, amount int, reason string, category string) error {
	if amount <= 0 {
		return fmt.Errorf("转账金额必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 1. 检查并锁定转出者积分
	var fromPoints int
	err = tx.QueryRow("SELECT points FROM users WHERE user_id = $1 FOR UPDATE", fromUserID).Scan(&fromPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("转出用户不存在或没有积分")
		}
		return fmt.Errorf("查询转出用户积分失败: %w", err)
	}

	if fromPoints < amount {
		return fmt.Errorf("积分不足，当前积分为: %d", fromPoints)
	}

	_, err = tx.Exec("UPDATE users SET points = points - $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, fromUserID)
	if err != nil {
		return fmt.Errorf("扣除积分失败: %w", err)
	}

	// 2. 增加接收者积分（如果不存在则插入）
	result, err := tx.Exec("UPDATE users SET points = points + $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", amount, toUserID)
	if err != nil {
		return fmt.Errorf("增加接收者积分失败: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if rowsAffected == 0 {
		insertQuery := `
		INSERT INTO users (user_id, points, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		if _, err := tx.Exec(insertQuery, toUserID, amount); err != nil {
			return fmt.Errorf("插入接收者用户并初始化积分失败: %w", err)
		}
	}

	// 3. 记录日志 (两条记录)
	logQuery := `INSERT INTO points_logs (user_id, amount, reason, category) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(logQuery, fromUserID, -amount, fmt.Sprintf("转账给 %s: %s", toUserID, reason), category)
	if err != nil {
		return fmt.Errorf("记录转出日志失败: %w", err)
	}
	_, err = tx.Exec(logQuery, toUserID, amount, fmt.Sprintf("来自 %s 的转账: %s", fromUserID, reason), category)
	if err != nil {
		return fmt.Errorf("记录转入日志失败: %w", err)
	}

	return tx.Commit()
}

// ------------------- 群管理员相关操作 -------------------

// AddGroupAdmin 添加群管理员
func AddGroupAdmin(db *sql.DB, groupID, userID string, level int) error {
	query := `
	INSERT INTO group_admins (group_id, user_id, level)
	VALUES ($1, $2, $3)
	ON CONFLICT (group_id, user_id) DO UPDATE
	SET level = $3
	`

	_, err := db.Exec(query, groupID, userID, level)
	if err != nil {
		return fmt.Errorf("添加群管理员失败: %w", err)
	}

	return nil
}

// RemoveGroupAdmin 移除群管理员
func RemoveGroupAdmin(db *sql.DB, groupID, userID string) error {
	query := `
	DELETE FROM group_admins
	WHERE group_id = $1 AND user_id = $2
	`

	result, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("移除群管理员失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("该用户不是群管理员: %s", userID)
	}

	return nil
}

// GetGroupAdmins 获取群管理员列表
func GetGroupAdmins(db *sql.DB, groupID string) ([]map[string]interface{}, error) {
	query := `
	SELECT user_id, level
	FROM group_admins
	WHERE group_id = $1
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("获取群管理员列表失败: %w", err)
	}
	defer rows.Close()

	admins := []map[string]interface{}{}
	for rows.Next() {
		var userID string
		var level int
		if err := rows.Scan(&userID, &level); err != nil {
			return nil, fmt.Errorf("扫描群管理员失败: %w", err)
		}
		admins = append(admins, map[string]interface{}{
			"user_id": userID,
			"level":   level,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历群管理员失败: %w", err)
	}

	return admins, nil
}

// IsGroupAdmin 检查用户是否为群管理员
func IsGroupAdmin(db *sql.DB, groupID, userID string) (bool, error) {
	query := `
	SELECT COUNT(*) > 0
	FROM group_admins
	WHERE group_id = $1 AND user_id = $2
	`

	var isAdmin bool
	if err := db.QueryRow(query, groupID, userID).Scan(&isAdmin); err != nil {
		return false, fmt.Errorf("检查群管理员失败: %w", err)
	}

	return isAdmin, nil
}

// GetAdminLevel 获取管理员权限级别
func GetAdminLevel(db *sql.DB, groupID, userID string) (int, error) {
	query := `
	SELECT level
	FROM group_admins
	WHERE group_id = $1 AND user_id = $2
	`

	var level int
	if err := db.QueryRow(query, groupID, userID).Scan(&level); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // 用户不是管理员
		}
		return 0, fmt.Errorf("获取管理员权限级别失败: %w", err)
	}

	return level, nil
}

// IsSuperAdmin 检查用户是否为超级管理员
func IsSuperAdmin(db *sql.DB, groupID, userID string) (bool, error) {
	level, err := GetAdminLevel(db, groupID, userID)
	if err != nil {
		return false, err
	}
	return level >= 2, nil
}

// ------------------- 群规相关操作 -------------------

// SetGroupRules 设置群规
func SetGroupRules(db *sql.DB, groupID, rules string) error {
	query := `
	INSERT INTO group_rules (group_id, rules, updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id) DO UPDATE
	SET rules = $2, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, groupID, rules)
	if err != nil {
		return fmt.Errorf("设置群规失败: %w", err)
	}

	return nil
}

// GetGroupRules 获取群规
func GetGroupRules(db *sql.DB, groupID string) (string, error) {
	query := `
	SELECT rules
	FROM group_rules
	WHERE group_id = $1
	`

	var rules string
	if err := db.QueryRow(query, groupID).Scan(&rules); err != nil {
		if err == sql.ErrNoRows {
			return "", nil // 群规不存在
		}
		return "", fmt.Errorf("获取群规失败: %w", err)
	}

	return rules, nil
}

func SetGroupVoiceID(db *sql.DB, groupID, voiceID string) error {
	query := `
	INSERT INTO group_rules (group_id, rules, voice_id, updated_at)
	VALUES ($1, '', $2, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id) DO UPDATE
	SET voice_id = $2, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, groupID, voiceID)
	if err != nil {
		return fmt.Errorf("设置群语音配置失败: %w", err)
	}

	return nil
}

func GetGroupVoiceID(db *sql.DB, groupID string) (string, error) {
	query := `
	SELECT voice_id
	FROM group_rules
	WHERE group_id = $1
	`

	var voiceID sql.NullString
	if err := db.QueryRow(query, groupID).Scan(&voiceID); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("获取群语音配置失败: %w", err)
	}

	if !voiceID.Valid {
		return "", nil
	}

	return voiceID.String, nil
}

func SetGroupFeatureOverride(db *sql.DB, groupID, featureID string, enabled bool) error {
	query := `
	INSERT INTO group_features (group_id, feature_id, enabled, updated_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id, feature_id) DO UPDATE
	SET enabled = $3, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, groupID, featureID, enabled)
	if err != nil {
		return fmt.Errorf("设置群功能开关失败: %w", err)
	}

	return nil
}

func DeleteGroupFeatureOverride(db *sql.DB, groupID, featureID string) error {
	query := `
	DELETE FROM group_features
	WHERE group_id = $1 AND feature_id = $2
	`

	_, err := db.Exec(query, groupID, featureID)
	if err != nil {
		return fmt.Errorf("删除群功能开关失败: %w", err)
	}

	return nil
}

func GetGroupFeatureOverride(db *sql.DB, groupID, featureID string) (bool, bool, error) {
	query := `
	SELECT enabled
	FROM group_features
	WHERE group_id = $1 AND feature_id = $2
	`

	var enabled bool
	err := db.QueryRow(query, groupID, featureID).Scan(&enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, fmt.Errorf("获取群功能开关失败: %w", err)
	}

	return enabled, true, nil
}

type SensitiveWord struct {
	Word  string
	Level int
}

func AddSensitiveWord(db *sql.DB, word string, level int) error {
	if level <= 0 {
		level = 1
	}

	query := `
	INSERT INTO sensitive_words (word, level)
	VALUES ($1, $2)
	ON CONFLICT (word) DO UPDATE
	SET level = EXCLUDED.level
	`

	_, err := db.Exec(query, word, level)
	if err != nil {
		return fmt.Errorf("添加敏感词失败: %w", err)
	}

	return nil
}

func RemoveSensitiveWord(db *sql.DB, word string) error {
	query := `
	DELETE FROM sensitive_words
	WHERE word = $1
	`

	result, err := db.Exec(query, word)
	if err != nil {
		return fmt.Errorf("移除敏感词失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("该敏感词不存在: %s", word)
	}

	return nil
}

func GetAllSensitiveWords(db *sql.DB) ([]SensitiveWord, error) {
	query := `
	SELECT word, level
	FROM sensitive_words
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取敏感词列表失败: %w", err)
	}
	defer rows.Close()

	words := []SensitiveWord{}
	for rows.Next() {
		var word SensitiveWord
		if err := rows.Scan(&word.Word, &word.Level); err != nil {
			return nil, fmt.Errorf("扫描敏感词失败: %w", err)
		}
		words = append(words, word)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历敏感词失败: %w", err)
	}

	return words, nil
}

func AddGroupWhitelistUser(db *sql.DB, groupID, userID string) error {
	query := `
	INSERT INTO group_whitelist (group_id, user_id)
	VALUES ($1, $2)
	ON CONFLICT (group_id, user_id) DO NOTHING
	`

	_, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("添加群白名单用户失败: %w", err)
	}

	return nil
}

func RemoveGroupWhitelistUser(db *sql.DB, groupID, userID string) error {
	query := `
	DELETE FROM group_whitelist
	WHERE group_id = $1 AND user_id = $2
	`

	_, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("移除群白名单用户失败: %w", err)
	}

	return nil
}

func ClearGroupWhitelist(db *sql.DB, groupID string) error {
	query := `
	DELETE FROM group_whitelist
	WHERE group_id = $1
	`

	_, err := db.Exec(query, groupID)
	if err != nil {
		return fmt.Errorf("清空群白名单失败: %w", err)
	}

	return nil
}

func IsUserInGroupWhitelist(db *sql.DB, groupID, userID string) (bool, error) {
	query := `
	SELECT COUNT(*) > 0
	FROM group_whitelist
	WHERE group_id = $1 AND user_id = $2
	`

	var exists bool
	if err := db.QueryRow(query, groupID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("检查群白名单用户失败: %w", err)
	}

	return exists, nil
}

// ------------------- 禁言记录相关操作 -------------------

// BanUser 禁言用户
func BanUser(db *sql.DB, groupID, userID string, banEndTime time.Time) error {
	query := `
	INSERT INTO banned_users (group_id, user_id, ban_end_time)
	VALUES ($1, $2, $3)
	ON CONFLICT (group_id, user_id) DO UPDATE
	SET ban_end_time = $3
	`

	_, err := db.Exec(query, groupID, userID, banEndTime)
	if err != nil {
		return fmt.Errorf("禁言用户失败: %w", err)
	}

	return nil
}

// UnbanUser 解除禁言
func UnbanUser(db *sql.DB, groupID, userID string) error {
	query := `
	DELETE FROM banned_users
	WHERE group_id = $1 AND user_id = $2
	`

	result, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("解除禁言失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("该用户未被禁言: %s", userID)
	}

	return nil
}

// GetBannedUsersByGroup 获取群内禁言用户列表
func GetBannedUsersByGroup(db *sql.DB, groupID string) ([]map[string]interface{}, error) {
	query := `
	SELECT user_id, ban_end_time
	FROM banned_users
	WHERE group_id = $1
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("获取禁言用户列表失败: %w", err)
	}
	defer rows.Close()

	bannedUsers := []map[string]interface{}{}
	for rows.Next() {
		var userID string
		var banEndTime time.Time
		if err := rows.Scan(&userID, &banEndTime); err != nil {
			return nil, fmt.Errorf("扫描禁言用户失败: %w", err)
		}

		bannedUsers = append(bannedUsers, map[string]interface{}{
			"user_id":      userID,
			"ban_end_time": banEndTime,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历禁言用户失败: %w", err)
	}

	return bannedUsers, nil
}

// GetExpiredBans 获取过期的禁言记录
func GetExpiredBans(db *sql.DB) ([]map[string]interface{}, error) {
	query := `
	SELECT group_id, user_id, ban_end_time
	FROM banned_users
	WHERE ban_end_time <= CURRENT_TIMESTAMP
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取过期禁言记录失败: %w", err)
	}
	defer rows.Close()

	expiredBans := []map[string]interface{}{}
	for rows.Next() {
		var groupID, userID string
		var banEndTime time.Time
		if err := rows.Scan(&groupID, &userID, &banEndTime); err != nil {
			return nil, fmt.Errorf("扫描过期禁言记录失败: %w", err)
		}

		expiredBans = append(expiredBans, map[string]interface{}{
			"group_id":     groupID,
			"user_id":      userID,
			"ban_end_time": banEndTime,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历过期禁言记录失败: %w", err)
	}

	return expiredBans, nil
}

// IsUserBanned 检查用户是否被禁言
func IsUserBanned(db *sql.DB, groupID, userID string) (bool, time.Time, error) {
	query := `
	SELECT ban_end_time
	FROM banned_users
	WHERE group_id = $1 AND user_id = $2
	`

	var banEndTime time.Time
	err := db.QueryRow(query, groupID, userID).Scan(&banEndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, time.Time{}, nil // 用户未被禁言
		}
		return false, time.Time{}, fmt.Errorf("检查用户禁言状态失败: %w", err)
	}

	// 检查禁言是否已过期
	if time.Now().After(banEndTime) {
		// 自动解除过期禁言
		if err := UnbanUser(db, groupID, userID); err != nil {
			return false, time.Time{}, fmt.Errorf("自动解除禁言失败: %w", err)
		}
		return false, time.Time{}, nil
	}

	return true, banEndTime, nil
}

// ------------------- 审核日志相关操作 -------------------

// AuditLog 定义审核日志模型
type AuditLog struct {
	ID            int       `json:"id"`
	GroupID       string    `json:"group_id"`
	AdminID       string    `json:"admin_id"`
	Action        string    `json:"action"`
	TargetUserID  string    `json:"target_user_id"`
	TargetGroupID string    `json:"target_group_id"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// AddAuditLog 添加审核日志
func AddAuditLog(db *sql.DB, log *AuditLog) error {
	query := `
	INSERT INTO audit_logs (group_id, admin_id, action, target_user_id, target_group_id, description)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.Exec(query, log.GroupID, log.AdminID, log.Action, log.TargetUserID, log.TargetGroupID, log.Description)
	if err != nil {
		return fmt.Errorf("添加审核日志失败: %w", err)
	}

	return nil
}

// GetAuditLogsByGroup 获取群的审核日志
func GetAuditLogsByGroup(db *sql.DB, groupID string, limit, offset int) ([]AuditLog, error) {
	query := `
	SELECT id, group_id, admin_id, action, target_user_id, target_group_id, description, created_at
	FROM audit_logs
	WHERE group_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取审核日志失败: %w", err)
	}
	defer rows.Close()

	logs := []AuditLog{}
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.GroupID, &log.AdminID, &log.Action, &log.TargetUserID, &log.TargetGroupID, &log.Description, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描审核日志失败: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历审核日志失败: %w", err)
	}

	return logs, nil
}

// GetAuditLogsByAdmin 获取管理员的审核日志
func GetAuditLogsByAdmin(db *sql.DB, adminID string, limit, offset int) ([]AuditLog, error) {
	query := `
	SELECT id, group_id, admin_id, action, target_user_id, target_group_id, description, created_at
	FROM audit_logs
	WHERE admin_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, adminID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取审核日志失败: %w", err)
	}
	defer rows.Close()

	logs := []AuditLog{}
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.GroupID, &log.AdminID, &log.Action, &log.TargetUserID, &log.TargetGroupID, &log.Description, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描审核日志失败: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历审核日志失败: %w", err)
	}

	return logs, nil
}
