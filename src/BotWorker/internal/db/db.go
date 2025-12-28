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
		user_id BIGINT NOT NULL UNIQUE,
		target_user_id BIGINT,
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

	// 创建群组表
	groupTableSQL := `
	CREATE TABLE IF NOT EXISTS groups (
		id SERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL UNIQUE,
		target_group_id BIGINT,
		name VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建消息记录表
	messageTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message_id VARCHAR(255) NOT NULL UNIQUE,
		user_id BIGINT,
		group_id BIGINT,
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
		user_id BIGINT,
		group_id BIGINT,
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
		group_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		level INTEGER NOT NULL DEFAULT 1, -- 权限级别: 1=普通管理员, 2=超级管理员
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	// 创建群规表
	groupRulesTableSQL := `
	CREATE TABLE IF NOT EXISTS group_rules (
		id SERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL UNIQUE,
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
		group_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		ban_end_time TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	// 创建审核日志表
	auditLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL,
		admin_id BIGINT NOT NULL,
		action VARCHAR(50) NOT NULL,
		target_user_id BIGINT,
		target_group_id BIGINT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	groupFeaturesTableSQL := `
	CREATE TABLE IF NOT EXISTS group_features (
		id SERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL,
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
		group_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, user_id)
	);
	`

	petsTableSQL := `
	CREATE TABLE IF NOT EXISTS pets (
		id SERIAL PRIMARY KEY,
		pet_id VARCHAR(255) NOT NULL UNIQUE,
		user_id BIGINT NOT NULL,
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
		user_id BIGINT NOT NULL,
		amount INTEGER NOT NULL,
		reason VARCHAR(255),
		category VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	questionsTableSQL := `
	CREATE TABLE IF NOT EXISTS questions (
		id SERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL,
		question_raw TEXT NOT NULL,
		question_normalized TEXT NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'approved',
		created_by BIGINT,
		source_group_id BIGINT,
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
		created_by BIGINT,
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
		group_id BIGINT NOT NULL UNIQUE,
		qa_mode VARCHAR(50) NOT NULL,
		last_answer_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 裂变系统表
	fissionConfigsTableSQL := `
	CREATE TABLE IF NOT EXISTS fission_configs (
		id SERIAL PRIMARY KEY,
		enabled BOOLEAN NOT NULL DEFAULT FALSE,
		invite_reward_points INTEGER NOT NULL DEFAULT 0,
		invite_reward_duration INTEGER NOT NULL DEFAULT 24,
		anti_fraud_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		welcome_message TEXT,
		invite_code_template VARCHAR(255) DEFAULT 'INV-{RAND}',
		rules TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	fissionTasksTableSQL := `
	CREATE TABLE IF NOT EXISTS fission_tasks (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		task_type VARCHAR(50) NOT NULL, -- registration, usage, group_join
		target_count INTEGER NOT NULL DEFAULT 1,
		reward_points INTEGER NOT NULL DEFAULT 0,
		reward_duration INTEGER NOT NULL DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	invitationsTableSQL := `
	CREATE TABLE IF NOT EXISTS invitations (
		id SERIAL PRIMARY KEY,
		inviter_id BIGINT NOT NULL,
		invitee_id BIGINT NOT NULL,
		platform VARCHAR(50),
		invite_code VARCHAR(50),
		status VARCHAR(50) DEFAULT 'pending', -- pending, completed, invalid
		ip_address VARCHAR(50),
		device_id VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(invitee_id)
	);
	`

	userFissionRecordsTableSQL := `
	CREATE TABLE IF NOT EXISTS user_fission_records (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL UNIQUE,
		platform VARCHAR(50),
		invite_count INTEGER NOT NULL DEFAULT 0,
		points INTEGER NOT NULL DEFAULT 0,
		level INTEGER NOT NULL DEFAULT 1,
		total_rewards DOUBLE PRECISION NOT NULL DEFAULT 0,
		invite_code VARCHAR(50) UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	fissionRewardLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS fission_reward_logs (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		type VARCHAR(50) NOT NULL, -- points, duration, item
		amount INTEGER NOT NULL DEFAULT 0,
		reason VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(userTableSQL); err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	if _, err := db.Exec(fissionConfigsTableSQL); err != nil {
		return fmt.Errorf("创建裂变配置表失败: %w", err)
	}

	if _, err := db.Exec(fissionTasksTableSQL); err != nil {
		return fmt.Errorf("创建裂变任务表失败: %w", err)
	}

	if _, err := db.Exec(invitationsTableSQL); err != nil {
		return fmt.Errorf("创建邀请记录表失败: %w", err)
	}

	if _, err := db.Exec(userFissionRecordsTableSQL); err != nil {
		return fmt.Errorf("创建用户裂变记录表失败: %w", err)
	}

	if _, err := db.Exec(fissionRewardLogsTableSQL); err != nil {
		return fmt.Errorf("创建裂变奖励日志表失败: %w", err)
	}

	// 尝试更改列类型（如果已存在）
	_, _ = db.Exec(`ALTER TABLE users ALTER COLUMN user_id TYPE BIGINT USING user_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE users ALTER COLUMN target_user_id TYPE BIGINT USING target_user_id::BIGINT`)

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS target_user_id BIGINT`); err != nil {
		return fmt.Errorf("为用户表添加 TargetUserID 字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS user_openid VARCHAR(255)`); err != nil {
		return fmt.Errorf("为用户表添加 UserOpenID 字段失败: %w", err)
	}

	if _, err := db.Exec(groupTableSQL); err != nil {
		return fmt.Errorf("创建群组表失败: %w", err)
	}

	// 尝试更改列类型（如果已存在）
	_, _ = db.Exec(`ALTER TABLE groups ALTER COLUMN group_id TYPE BIGINT USING group_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE groups ALTER COLUMN target_group_id TYPE BIGINT USING target_group_id::BIGINT`)

	if _, err := db.Exec(`ALTER TABLE groups ADD COLUMN IF NOT EXISTS target_group_id BIGINT`); err != nil {
		return fmt.Errorf("为群组表添加 TargetGroupID 字段失败: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE groups ADD COLUMN IF NOT EXISTS group_openid VARCHAR(255)`); err != nil {
		return fmt.Errorf("为群组表添加 GroupOpenID 字段失败: %w", err)
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

	// 尝试更改列类型
	_, _ = db.Exec(`ALTER TABLE group_admins ALTER COLUMN group_id TYPE BIGINT USING group_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE group_admins ALTER COLUMN user_id TYPE BIGINT USING user_id::BIGINT`)

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

	// 统一更改所有表中的 user_id 和 group_id 为 BIGINT
	tables := []string{"messages", "sessions", "group_rules", "banned_users", "audit_logs", "group_features", "group_whitelist", "pets", "points_logs", "questions"}
	for _, table := range tables {
		_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN user_id TYPE BIGINT USING user_id::BIGINT`, table))
		_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN group_id TYPE BIGINT USING group_id::BIGINT`, table))
	}
	_, _ = db.Exec(`ALTER TABLE audit_logs ALTER COLUMN admin_id TYPE BIGINT USING admin_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ALTER COLUMN target_user_id TYPE BIGINT USING target_user_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ALTER COLUMN target_group_id TYPE BIGINT USING target_group_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE questions ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE questions ALTER COLUMN source_group_id TYPE BIGINT USING source_group_id::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE answers ALTER COLUMN created_by TYPE BIGINT USING created_by::BIGINT`)
	_, _ = db.Exec(`ALTER TABLE group_ai_settings ALTER COLUMN group_id TYPE BIGINT USING group_id::BIGINT`)

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

// GetMaxUserIDPlusOne 获取用户表中最大的 UserID + 1，如果为空则返回 980000000000
func GetMaxUserIDPlusOne(db *sql.DB) (int64, error) {
	var maxID sql.NullInt64
	query := `SELECT MAX(user_id) FROM users`
	err := db.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 980000000000, nil
	}
	if !maxID.Valid {
		return 980000000000, nil
	}
	if maxID.Int64 < 980000000000 {
		return 980000000000, nil
	}
	return maxID.Int64 + 1, nil
}

// GetMaxGroupIDPlusOne 获取群组表中最大的 GroupID + 1，如果为空则返回 990000000000
func GetMaxGroupIDPlusOne(db *sql.DB) (int64, error) {
	var maxID sql.NullInt64
	query := `SELECT MAX(group_id) FROM groups`
	err := db.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 990000000000, nil
	}
	if !maxID.Valid {
		return 990000000000, nil
	}
	if maxID.Int64 < 990000000000 {
		return 990000000000, nil
	}
	return maxID.Int64 + 1, nil
}

// GetUserIDByTargetID 根据 TargetUserID 获取 UserID (int64)
func GetUserIDByTargetID(db *sql.DB, targetID int64) (int64, error) {
	var userID int64
	query := `SELECT user_id FROM users WHERE target_user_id = $1`
	err := db.QueryRow(query, targetID).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

// GetGroupIDByTargetID 根据 TargetGroupID 获取 GroupID (int64)
func GetGroupIDByTargetID(db *sql.DB, targetID int64) (int64, error) {
	var groupID int64
	query := `SELECT group_id FROM groups WHERE target_group_id = $1`
	err := db.QueryRow(query, targetID).Scan(&groupID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return groupID, err
}

// GetUserIDByOpenID 根据 UserOpenID 获取 UserID (int64)
func GetUserIDByOpenID(db *sql.DB, openID string) (int64, error) {
	var userID int64
	query := `SELECT user_id FROM users WHERE user_openid = $1`
	err := db.QueryRow(query, openID).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

// GetGroupIDByOpenID 根据 GroupOpenID 获取 GroupID (int64)
func GetGroupIDByOpenID(db *sql.DB, openID string) (int64, error) {
	var groupID int64
	query := `SELECT group_id FROM groups WHERE group_openid = $1`
	err := db.QueryRow(query, openID).Scan(&groupID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return groupID, err
}

// CreateUserWithTargetID 创建带有 TargetUserID 和 UserOpenID 的用户
func CreateUserWithTargetID(db *sql.DB, userID int64, targetID int64, openID string, nickname, avatar string) error {
	query := `
	INSERT INTO users (user_id, target_user_id, user_openid, nickname, avatar, updated_at)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET target_user_id = $2, user_openid = $3, nickname = $4, avatar = $5, updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, userID, targetID, openID, nickname, avatar)
	return err
}

// CreateGroupWithTargetID 创建带有 TargetGroupID 和 GroupOpenID 的群组
func CreateGroupWithTargetID(db *sql.DB, groupID int64, targetID int64, openID string, name string) error {
	query := `
	INSERT INTO groups (group_id, target_group_id, group_openid, name, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (group_id) DO UPDATE
	SET target_group_id = $2, group_openid = $3, name = $4, updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, groupID, targetID, openID, name)
	return err
}

// User 定义用户模型
type User struct {
	ID            int       `json:"id"`
	UserID        int64     `json:"user_id"`
	TargetUserID  int64     `json:"target_user_id"`
	UserOpenID    string    `json:"user_openid"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	Gender        string    `json:"gender"`
	Points        int       `json:"points"`
	SavingsPoints int       `json:"savings_points"`
	FrozenPoints  int       `json:"frozen_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Group 定义群组模型
type Group struct {
	ID            int       `json:"id"`
	GroupID       int64     `json:"group_id"`
	TargetGroupID int64     `json:"target_group_id"`
	GroupOpenID   string    `json:"group_openid"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, user *User) error {
	query := `
	INSERT INTO users (user_id, user_openid, nickname, avatar, gender, points, savings_points, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET user_openid = $2, nickname = $3, avatar = $4, gender = $5, points = $6, savings_points = $7, updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, user.UserID, user.UserOpenID, user.Nickname, user.Avatar, user.Gender, user.Points, user.SavingsPoints)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByUserID 根据用户ID获取用户信息
func GetUserByUserID(db *sql.DB, userID int64) (*User, error) {
	query := `
	SELECT id, user_id, user_openid, nickname, avatar, gender, points, savings_points, frozen_points, created_at, updated_at
	FROM users
	WHERE user_id = $1
	`

	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.UserID, &user.UserOpenID, &user.Nickname, &user.Avatar, &user.Gender, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
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
		return fmt.Errorf("用户不存在: %d", user.UserID)
	}

	return nil
}

// Message 定义消息记录模型
type Message struct {
	ID        int       `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    int64     `json:"user_id"`
	GroupID   int64     `json:"group_id"`
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
func GetMessagesByUserID(db *sql.DB, userID int64, limit int) ([]*Message, error) {
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
	ID        int            `json:"id"`
	SessionID string         `json:"session_id"`
	UserID    int64          `json:"user_id"`
	GroupID   int64          `json:"group_id"`
	State     string         `json:"state"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
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
func DeleteUser(db *sql.DB, userID int64) error {
	query := `
	DELETE FROM users
	WHERE user_id = $1
	`

	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
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
func GetMessagesByGroupID(db *sql.DB, groupID int64, limit int) ([]*Message, error) {
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
func GetSessionsByUserID(db *sql.DB, userID int64) ([]*Session, error) {
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
func GetSessionsByGroupID(db *sql.DB, groupID int64) ([]*Session, error) {
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
	UserID    int64     `json:"user_id"`
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
	GroupID      int64     `json:"group_id"`
	QAMode       string    `json:"qa_mode"`
	LastAnswerID int       `json:"last_answer_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Question struct {
	ID                 int       `json:"id"`
	GroupID            int64     `json:"group_id"`
	QuestionRaw        string    `json:"question_raw"`
	QuestionNormalized string    `json:"question_normalized"`
	Status             string    `json:"status"`
	CreatedBy          int64     `json:"created_by"`
	SourceGroupID      int64     `json:"source_group_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Answer struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question_id"`
	Answer     string    `json:"answer"`
	Status     string    `json:"status"`
	CreatedBy  int64     `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func GetGroupQAMode(db *sql.DB, groupID int64) (string, error) {
	if db == nil || groupID == 0 {
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

func SetGroupQAMode(db *sql.DB, groupID int64, mode string) error {
	if db == nil || groupID == 0 {
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

func GetGroupLastAnswerID(db *sql.DB, groupID int64) (int, error) {
	if db == nil || groupID == 0 {
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

func SetGroupLastAnswerID(db *sql.DB, groupID int64, answerID int) error {
	if db == nil || groupID == 0 {
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

func GetQuestionByGroupAndNormalized(dbConn *sql.DB, groupID int64, normalized string) (*Question, error) {
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
func GetPetsByUserID(db *sql.DB, userID int64) ([]*PetModel, error) {
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
func GetPoints(db *sql.DB, userID int64) (int, error) {
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

func GetFrozenPoints(db *sql.DB, userID int64) (int, error) {
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
func AddPoints(db *sql.DB, userID int64, amount int, reason string, category string) error {
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

func applySavingsInterestTx(tx *sql.Tx, userID int64) (int, error) {
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

func DepositPointsToSavings(db *sql.DB, userID int64, amount int) error {
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

func WithdrawPointsFromSavings(db *sql.DB, userID int64, amount int) error {
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

func GetSavingsPoints(db *sql.DB, userID int64) (int, error) {
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

func FreezePoints(db *sql.DB, userID int64, amount int, reason string) error {
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

func UnfreezePoints(db *sql.DB, userID int64, amount int, reason string) error {
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
func TransferPoints(db *sql.DB, fromUserID, toUserID int64, amount int, reason string, category string) error {
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
func AddGroupAdmin(db *sql.DB, groupID, userID int64, level int) error {
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
func RemoveGroupAdmin(db *sql.DB, groupID, userID int64) error {
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
func GetGroupAdmins(db *sql.DB, groupID int64) ([]map[string]any, error) {
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

	admins := []map[string]any{}
	for rows.Next() {
		var userID int64
		var level int
		if err := rows.Scan(&userID, &level); err != nil {
			return nil, fmt.Errorf("扫描群管理员失败: %w", err)
		}
		admins = append(admins, map[string]any{
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
func IsGroupAdmin(db *sql.DB, groupID, userID int64) (bool, error) {
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
func GetAdminLevel(db *sql.DB, groupID, userID int64) (int, error) {
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
func IsSuperAdmin(db *sql.DB, groupID, userID int64) (bool, error) {
	level, err := GetAdminLevel(db, groupID, userID)
	if err != nil {
		return false, err
	}
	return level >= 2, nil
}

// ------------------- 群规相关操作 -------------------

// SetGroupRules 设置群规
func SetGroupRules(db *sql.DB, groupID int64, rules string) error {
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
func GetGroupRules(db *sql.DB, groupID int64) (string, error) {
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

func SetGroupVoiceID(db *sql.DB, groupID int64, voiceID string) error {
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

func GetGroupVoiceID(db *sql.DB, groupID int64) (string, error) {
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

func SetGroupFeatureOverride(db *sql.DB, groupID int64, featureID string, enabled bool) error {
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

func DeleteGroupFeatureOverride(db *sql.DB, groupID int64, featureID string) error {
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

func GetGroupFeatureOverride(db *sql.DB, groupID int64, featureID string) (bool, bool, error) {
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

func AddGroupWhitelistUser(db *sql.DB, groupID, userID int64) error {
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

func RemoveGroupWhitelistUser(db *sql.DB, groupID, userID int64) error {
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

func ClearGroupWhitelist(db *sql.DB, groupID int64) error {
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

func IsUserInGroupWhitelist(db *sql.DB, groupID, userID int64) (bool, error) {
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
func BanUser(db *sql.DB, groupID, userID int64, banEndTime time.Time) error {
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
func UnbanUser(db *sql.DB, groupID, userID int64) error {
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
func GetBannedUsersByGroup(db *sql.DB, groupID int64) ([]map[string]any, error) {
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

	bannedUsers := []map[string]any{}
	for rows.Next() {
		var userID int64
		var banEndTime time.Time
		if err := rows.Scan(&userID, &banEndTime); err != nil {
			return nil, fmt.Errorf("扫描禁言用户失败: %w", err)
		}

		bannedUsers = append(bannedUsers, map[string]any{
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
func GetExpiredBans(db *sql.DB) ([]map[string]any, error) {
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

	expiredBans := []map[string]any{}
	for rows.Next() {
		var groupID, userID int64
		var banEndTime time.Time
		if err := rows.Scan(&groupID, &userID, &banEndTime); err != nil {
			return nil, fmt.Errorf("扫描过期禁言记录失败: %w", err)
		}

		expiredBans = append(expiredBans, map[string]any{
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
func IsUserBanned(db *sql.DB, groupID, userID int64) (bool, time.Time, error) {
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
	GroupID       int64     `json:"group_id"`
	AdminID       int64     `json:"admin_id"`
	Action        string    `json:"action"`
	TargetUserID  int64     `json:"target_user_id"`
	TargetGroupID int64     `json:"target_group_id"`
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
func GetAuditLogsByGroup(db *sql.DB, groupID int64, limit, offset int) ([]AuditLog, error) {
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

// ------------------- 裂变系统相关操作 -------------------

// FissionConfig 定义裂变配置模型
type FissionConfig struct {
	ID                  int       `json:"id"`
	Enabled             bool      `json:"enabled"`
	InviteRewardPoints  int       `json:"invite_reward_points"`
	NewUserRewardPoints int       `json:"new_user_reward_points"`
	MinLevelRequired    int       `json:"min_level_required"`
	MaxDailyInvites     int       `json:"max_daily_invites"`
	AntiFraudEnabled    bool      `json:"anti_fraud_enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GetFissionConfig 获取裂变配置
func GetFissionConfig(db *sql.DB) (*FissionConfig, error) {
	query := `
	SELECT id, enabled, invite_reward_points, new_user_reward_points, min_level_required, max_daily_invites, anti_fraud_enabled, created_at, updated_at
	FROM fission_configs
	LIMIT 1
	`
	config := &FissionConfig{}
	err := db.QueryRow(query).Scan(
		&config.ID, &config.Enabled, &config.InviteRewardPoints, &config.NewUserRewardPoints, &config.MinLevelRequired, &config.MaxDailyInvites, &config.AntiFraudEnabled, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取裂变配置失败: %w", err)
	}
	return config, nil
}

// CreateInvitation 创建邀请记录
func CreateInvitation(db *sql.DB, inviterID, inviteeID int64, inviteCode string, ipAddress, deviceID string) error {
	query := `
	INSERT INTO invitations (inviter_id, invitee_id, invite_code, status, ip_address, device_id, updated_at)
	VALUES ($1, $2, $3, 'pending', $4, $5, CURRENT_TIMESTAMP)
	ON CONFLICT (invitee_id) DO NOTHING
	`
	_, err := db.Exec(query, inviterID, inviteeID, inviteCode, ipAddress, deviceID)
	if err != nil {
		return fmt.Errorf("创建邀请记录失败: %w", err)
	}
	return nil
}

// CheckInvitationFraud 检查是否存在作弊风险
func CheckInvitationFraud(db *sql.DB, ipAddress, deviceID string) (bool, string, error) {
	if ipAddress != "" {
		var ipCount int
		ipQuery := `SELECT COUNT(*) FROM invitations WHERE ip_address = $1 AND status = 'completed'`
		err := db.QueryRow(ipQuery, ipAddress).Scan(&ipCount)
		if err == nil && ipCount >= 3 {
			return true, "IP 绑定次数过多", nil
		}
	}

	if deviceID != "" {
		var deviceCount int
		deviceQuery := `SELECT COUNT(*) FROM invitations WHERE device_id = $1 AND status = 'completed'`
		err := db.QueryRow(deviceQuery, deviceID).Scan(&deviceCount)
		if err == nil && deviceCount >= 2 {
			return true, "设备绑定次数过多", nil
		}
	}

	return false, "", nil
}

// GetDailyInviteCount 获取用户当日邀请数量
func GetDailyInviteCount(db *sql.DB, inviterID int64) (int, error) {
	query := `
	SELECT COUNT(*) 
	FROM invitations 
	WHERE inviter_id = $1 AND created_at >= CURRENT_DATE
	`
	var count int
	err := db.QueryRow(query, inviterID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取当日邀请数量失败: %w", err)
	}
	return count, nil
}

// FissionTask 定义裂变任务模型
type FissionTask struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	TaskType       string    `json:"task_type"`
	TargetCount    int       `json:"target_count"`
	RewardPoints   int       `json:"reward_points"`
	RewardDuration int       `json:"reward_duration"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetActiveFissionTasks 获取进行中的裂变任务
func GetActiveFissionTasks(db *sql.DB) ([]FissionTask, error) {
	query := `
	SELECT id, name, description, task_type, target_count, reward_points, reward_duration, status, created_at
	FROM fission_tasks
	WHERE status = 'active'
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取裂变任务失败: %w", err)
	}
	defer rows.Close()

	tasks := []FissionTask{}
	for rows.Next() {
		var t FissionTask
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.TaskType, &t.TargetCount, &t.RewardPoints, &t.RewardDuration, &t.Status, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描裂变任务失败: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// CreateFissionRewardLog 记录裂变奖励日志
func CreateFissionRewardLog(db *sql.DB, userID int64, rewardType string, amount int, reason string) error {
	query := `
	INSERT INTO fission_reward_logs (user_id, type, amount, reason, created_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`
	_, err := db.Exec(query, userID, rewardType, amount, reason)
	if err != nil {
		return fmt.Errorf("记录裂变奖励日志失败: %w", err)
	}
	return nil
}

// CompleteFissionTask 完成裂变任务并分发奖励
func CompleteFissionTask(db *sql.DB, userID int64, taskType string) error {
	// 1. 获取所有进行中的该类型的任务
	query := `
	SELECT id, name, reward_points, reward_duration 
	FROM fission_tasks 
	WHERE task_type = $1 AND status = 'active'
	`
	rows, err := db.Query(query, taskType)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID int
		var taskName string
		var points, duration int
		if err := rows.Scan(&taskID, &taskName, &points, &duration); err != nil {
			continue
		}

		// 2. 检查用户是否已经完成过此任务
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM fission_reward_logs WHERE user_id = $1 AND reason LIKE $2)`
		_ = db.QueryRow(checkQuery, userID, "%"+taskName+"%").Scan(&exists)
		if exists {
			continue
		}

		// 3. 发放奖励
		reason := fmt.Sprintf("完成裂变任务: %s", taskName)
		if points > 0 {
			_ = AddPoints(db, userID, points, reason, "fission_task")
			_ = CreateFissionRewardLog(db, userID, "points", points, reason)
		}

		// 4. 如果是 register 任务，且该用户是被邀请者，也要给邀请者额外奖励（如果配置了）
		if taskType == "register" {
			inviterID, _, status, err := GetInvitationByInviteeID(db, userID)
			if err == nil && inviterID != 0 && status == "completed" {
				// 可以根据需要增加邀请者的额外奖励逻辑
			}
		}
	}

	return nil
}

// GetInvitationByInviteeID 根据被邀请者ID获取邀请记录
func GetInvitationByInviteeID(db *sql.DB, inviteeID int64) (int64, string, string, error) {
	query := `SELECT inviter_id, invite_code, status FROM invitations WHERE invitee_id = $1`
	var inviterID int64
	var inviteCode string
	var status string
	err := db.QueryRow(query, inviteeID).Scan(&inviterID, &inviteCode, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", nil
		}
		return 0, "", "", fmt.Errorf("获取邀请记录失败: %w", err)
	}
	return inviterID, inviteCode, status, nil
}

// UpdateInvitationStatus 更新邀请状态
func UpdateInvitationStatus(db *sql.DB, inviteeID int64, status string) error {
	query := `UPDATE invitations SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE invitee_id = $2`
	_, err := db.Exec(query, status, inviteeID)
	if err != nil {
		return fmt.Errorf("更新邀请状态失败: %w", err)
	}
	return nil
}

// UserFissionRecord 定义用户裂变记录模型
type UserFissionRecord struct {
	UserID      int64  `json:"user_id"`
	InviteCount int    `json:"invite_count"`
	Points      int    `json:"points"`
	Level       int    `json:"level"`
	InviteCode  string `json:"invite_code"`
}

// GetUserFissionRecord 获取用户裂变记录
func GetUserFissionRecord(db *sql.DB, userID int64) (*UserFissionRecord, error) {
	query := `SELECT invite_count, points, level, invite_code FROM user_fission_records WHERE user_id = $1`
	record := &UserFissionRecord{UserID: userID}
	err := db.QueryRow(query, userID).Scan(&record.InviteCount, &record.Points, &record.Level, &record.InviteCode)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果不存在，返回默认记录（邀请码逻辑会在插件层处理）
			return record, nil
		}
		return nil, fmt.Errorf("获取用户裂变记录失败: %w", err)
	}
	return record, nil
}

// UpdateUserFissionRecord 更新用户裂变记录
func UpdateUserFissionRecord(db *sql.DB, userID int64, inviteIncr, rewardIncr, pointsIncr int) error {
	query := `
	INSERT INTO user_fission_records (user_id, invite_count, total_rewards, points, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET invite_count = user_fission_records.invite_count + $2,
	    total_rewards = user_fission_records.total_rewards + $3,
	    points = user_fission_records.points + $4,
	    updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, userID, inviteIncr, rewardIncr, pointsIncr)
	if err != nil {
		return fmt.Errorf("更新用户裂变记录失败: %w", err)
	}
	return nil
}

// GetFissionRank 获取裂变排行榜
func GetFissionRank(db *sql.DB, limit int) ([]map[string]any, error) {
	query := `
	SELECT user_id, invite_count, points
	FROM user_fission_records
	ORDER BY invite_count DESC, points DESC
	LIMIT $1
	`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("获取裂变排行榜失败: %w", err)
	}
	defer rows.Close()

	rank := []map[string]any{}
	for rows.Next() {
		var userID int64
		var inviteCount, points int
		if err := rows.Scan(&userID, &inviteCount, &points); err != nil {
			return nil, fmt.Errorf("扫描裂变排行失败: %w", err)
		}
		rank = append(rank, map[string]any{
			"user_id":      userID,
			"invite_count": inviteCount,
			"points":       points,
		})
	}
	return rank, nil
}

// GetAuditLogsByAdmin 获取管理员的审核日志
func GetAuditLogsByAdmin(db *sql.DB, adminID int64, limit, offset int) ([]AuditLog, error) {
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
