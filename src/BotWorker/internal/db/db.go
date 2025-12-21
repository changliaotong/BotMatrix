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
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL UNIQUE,
		nickname VARCHAR(255),
		avatar VARCHAR(255),
		gender VARCHAR(10),
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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建敏感词表
	sensitiveWordsTableSQL := `
	CREATE TABLE IF NOT EXISTS sensitive_words (
		id SERIAL PRIMARY KEY,
		word VARCHAR(255) NOT NULL UNIQUE,
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

	// 执行建表语句
	if _, err := db.Exec(userTableSQL); err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
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

	if _, err := db.Exec(sensitiveWordsTableSQL); err != nil {
		return fmt.Errorf("创建敏感词表失败: %w", err)
	}

	if _, err := db.Exec(bannedUsersTableSQL); err != nil {
		return fmt.Errorf("创建禁言记录表失败: %w", err)
	}

	if _, err := db.Exec(auditLogsTableSQL); err != nil {
		return fmt.Errorf("创建审核日志表失败: %w", err)
	}

	return nil
}

// User 定义用户模型
type User struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Gender    string    `json:"gender"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, user *User) error {
	query := `
	INSERT INTO users (user_id, nickname, avatar, gender, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET nickname = $2, avatar = $3, gender = $4, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, user.UserID, user.Nickname, user.Avatar, user.Gender)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByUserID 根据用户ID获取用户信息
func GetUserByUserID(db *sql.DB, userID string) (*User, error) {
	query := `
	SELECT id, user_id, nickname, avatar, gender, created_at, updated_at
	FROM users
	WHERE user_id = $1
	`

	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.CreatedAt, &user.UpdatedAt,
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
	SELECT id, user_id, nickname, avatar, gender, created_at, updated_at
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
			&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.CreatedAt, &user.UpdatedAt,
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
	SELECT id, user_id, nickname, avatar, gender, created_at, updated_at
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
			&user.ID, &user.UserID, &user.Nickname, &user.Avatar, &user.Gender, &user.CreatedAt, &user.UpdatedAt,
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
	query := `SELECT COUNT(*) FROM users`
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, fmt.Errorf("获取用户数量失败: %w", err)
	}
	return count, nil
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

// ------------------- 敏感词相关操作 -------------------

// AddSensitiveWord 添加敏感词
func AddSensitiveWord(db *sql.DB, word string) error {
	query := `
	INSERT INTO sensitive_words (word)
	VALUES ($1)
	ON CONFLICT (word) DO NOTHING
	`

	_, err := db.Exec(query, word)
	if err != nil {
		return fmt.Errorf("添加敏感词失败: %w", err)
	}

	return nil
}

// RemoveSensitiveWord 移除敏感词
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

// GetAllSensitiveWords 获取所有敏感词
func GetAllSensitiveWords(db *sql.DB) ([]string, error) {
	query := `
	SELECT word
	FROM sensitive_words
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取敏感词列表失败: %w", err)
	}
	defer rows.Close()

	words := []string{}
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, fmt.Errorf("扫描敏感词失败: %w", err)
		}
		words = append(words, word)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历敏感词失败: %w", err)
	}

	return words, nil
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
