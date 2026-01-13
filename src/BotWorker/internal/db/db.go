package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"botworker/internal/config"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewGORMConnection 创建一个新的 GORM 数据库连接
func NewGORMConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("无法打开 GORM 数据库连接: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	return db, nil
}

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

const (
	TableUser        = "users"
	TableGroup       = "groups"
	TableGroupMember = "group_members"
	TableQuestion    = "questions"
	TableAnswer      = "answers"
	TableBlackList   = "black_list"
	TableWhiteList   = "white_list"
	TableGreyList    = "grey_list"
	TableVIP         = "vips"
	TableCredit      = "credits"
	TableSavings     = "user_savings_metadata"
	TableFriend      = "friends"
	TableConsumption = "user_consumptions"
	TableAgent       = "agents"
)

// InitDatabase 初始化数据库 (已按照要求禁用自动创建表功能)
func InitDatabase(db *sql.DB) error {
	/*
		// 此处省略大量 CREATE TABLE 和 ALTER TABLE 逻辑，因为用户要求不在代码中修改数据库。
		// 原逻辑包含 messages, sessions, group_admins, group_rules, audit_logs, pets, points_logs, black_list, vips 等表的初始化。
	*/
	return nil
}

// Agent 定义智能体模型
type Agent struct {
	Id               int64     `json:"id"`
	Guid             string    `json:"guid"`
	Name             string    `json:"name"`
	Prompt           string    `json:"prompt"`
	Model            string    `json:"model"`
	Temperature      float64   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             float64   `json:"top_p"`
	FrequencyPenalty float64   `json:"frequency_penalty"`
	PresencePenalty  float64   `json:"presence_penalty"`
	Stop             string    `json:"stop"`
	Private          int       `json:"private"`
	OwnerId          int64     `json:"owner_id"`
	InsertDate       time.Time `json:"insert_date"`
	UpdateDate       time.Time `json:"update_date"`
}

// GetMaxUserIDPlusOne 获取用户表中最大的 UserID + 1，如果为空则返回 980000000000
func GetMaxUserIDPlusOne(db *sql.DB) (int64, error) {
	var maxID sql.NullInt64
	query := fmt.Sprintf(`SELECT MAX(id) FROM %s`, TableUser)
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
	query := fmt.Sprintf(`SELECT MAX(oid) FROM %s`, TableGroup)
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
	query := fmt.Sprintf(`SELECT id FROM %s WHERE target_user_id = $1`, TableUser)
	err := db.QueryRow(query, targetID).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

// GetGroupIDByTargetID 根据 TargetGroupID 获取 GroupID (int64)
func GetGroupIDByTargetID(db *sql.DB, targetID int64) (int64, error) {
	var groupID int64
	query := fmt.Sprintf(`SELECT oid FROM %s WHERE target_group = $1`, TableGroup)
	err := db.QueryRow(query, targetID).Scan(&groupID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return groupID, err
}

// GetUserIDByOpenID 根据 UserOpenID 获取 UserID (int64)
func GetUserIDByOpenID(db *sql.DB, openID string) (int64, error) {
	var userID int64
	query := fmt.Sprintf(`SELECT id FROM %s WHERE user_openid = $1`, TableUser)
	err := db.QueryRow(query, openID).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

// GetGroupIDByOpenID 根据 GroupOpenID 获取 GroupID (int64)
func GetGroupIDByOpenID(db *sql.DB, openID string) (int64, error) {
	var groupID int64
	query := fmt.Sprintf(`SELECT oid FROM %s WHERE group_openid = $1`, TableGroup)
	err := db.QueryRow(query, openID).Scan(&groupID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return groupID, err
}

// CreateUserWithTargetID 创建带有 TargetUserID 和 UserOpenID 的用户
func CreateUserWithTargetID(db *sql.DB, userID int64, targetID int64, openID string, nickname, avatar string) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (id, target_user_id, user_openid, name)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE
	SET target_user_id = $2, user_openid = $3, name = $4
	`, TableUser)
	_, err := db.Exec(query, userID, targetID, openID, nickname)
	return err
}

// UpdateUserSuperPoints 更新用户超级积分状态
func UpdateUserSuperPoints(db *sql.DB, userID int64, isSuperPoints bool) error {
	query := fmt.Sprintf(`
	UPDATE %s
	SET is_super = $2, upgrade_date = CURRENT_TIMESTAMP
	WHERE id = $1
	`, TableUser)
	result, err := db.Exec(query, userID, isSuperPoints)
	if err != nil {
		return fmt.Errorf("更新用户超级积分状态失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在: %d", userID)
	}

	return nil
}

// UpdateUserPoints 更新用户通用积分
func UpdateUserPoints(db *sql.DB, userID int64, points int) error {
	query := fmt.Sprintf(`
	UPDATE %s
	SET credit = $2
	WHERE id = $1
	`, TableUser)
	result, err := db.Exec(query, userID, points)
	if err != nil {
		return fmt.Errorf("更新用户积分失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在: %d", userID)
	}

	return nil
}

// CreateGroupWithTargetID 创建带有 TargetGroupID 和 GroupOpenID 的群组
func CreateGroupWithTargetID(db *sql.DB, groupID int64, targetID int64, openID string, name string) error {
	query := fmt.Sprintf(`
	INSERT INTO %s ("Oid", "TargetGroup", "GroupOpenid", "GroupName", "LastDate")
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT ("Oid") DO UPDATE
	SET "TargetGroup" = $2, "GroupOpenid" = $3, "GroupName" = $4, "LastDate" = CURRENT_TIMESTAMP
	`, TableGroup)
	_, err := db.Exec(query, groupID, targetID, openID, name)
	return err
}

// SetGroupValue 设置群组字段值
func SetGroupValue(db *sql.DB, groupID int64, fieldName string, value interface{}) error {
	query := fmt.Sprintf(`UPDATE %s SET "%s" = $2, "LastDate" = CURRENT_TIMESTAMP WHERE "Oid" = $1`, TableGroup, fieldName)
	_, err := db.Exec(query, groupID, value)
	return err
}

// GetGroupValue 获取群组字段值
func GetGroupValue(db *sql.DB, groupID int64, fieldName string) (string, error) {
	var val string
	query := fmt.Sprintf(`SELECT COALESCE("%s", '') FROM %s WHERE "Oid" = $1`, fieldName, TableGroup)
	err := db.QueryRow(query, groupID).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// User 定义用户模型
type User struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TargetUserID  int64     `json:"target_user_id"`
	UserOpenID    string    `json:"user_openid"`
	Nickname      string    `json:"nickname"`
	IsSuperPoints bool      `json:"is_super_points"`
	Points        int64     `json:"points"`
	SavingsPoints int64     `json:"savings_points"`
	FrozenPoints  int64     `json:"frozen_points"`
	IsAI          bool      `json:"is_ai"`
	AgentId       int64     `json:"agent_id"`
	Tokens        int64     `json:"tokens"`
	DayTokens     int64     `json:"day_tokens"`
	SystemPrompt  string    `json:"system_prompt"`
	LastSignIn    time.Time `json:"last_sign_in"`
	SignInDays    int       `json:"sign_in_days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Group 定义群组模型
type Group struct {
	ID              int64     `json:"id"`
	GroupID         int64     `json:"group_id"`
	TargetGroupID   int64     `json:"target_group_id"`
	GroupOpenID     string    `json:"group_openid"`
	Name            string    `json:"name"`
	RecallKeyword   string    `json:"recall_keyword"`
	WarnKeyword     string    `json:"warn_keyword"`
	MuteKeyword     string    `json:"mute_keyword"`
	KickKeyword     string    `json:"kick_keyword"`
	BlackKeyword    string    `json:"black_keyword"`
	WhiteKeyword    string    `json:"white_keyword"`
	IsPowerOn       bool      `json:"is_power_on"`
	IsWelcomeHint   bool      `json:"is_welcome_hint"`
	WelcomeMessage  string    `json:"welcome_message"`
	IsMuteEnter     bool      `json:"is_mute_enter"`
	MuteEnterCount  uint32    `json:"mute_enter_count"`
	IsConfirmNew    bool      `json:"is_confirm_new"`
	IsRequirePrefix bool      `json:"is_require_prefix"`
	IsAI            bool      `json:"is_ai"`
	IsOwnerPay      bool      `json:"is_owner_pay"`
	RobotOwner      int64     `json:"robot_owner"`
	ContextCount    int       `json:"context_count"`
	SystemPrompt    string    `json:"system_prompt"`
	IsMultAI        bool      `json:"is_mult_ai"`
	IsVoiceReply    bool      `json:"is_voice_reply"`
	VoiceId         string    `json:"voice_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetGroupByGroupID 根据群组ID获取群组信息
func GetGroupByGroupID(db *sql.DB, groupID int64) (*Group, error) {
	query := fmt.Sprintf(`
	SELECT "Oid", "Oid", "TargetGroup", "GroupOpenid", "GroupName", 
	       COALESCE("RecallKeyword", ''), COALESCE("WarnKeyword", ''), COALESCE("MuteKeyword", ''), 
	       COALESCE("KickKeyword", ''), COALESCE("BlackKeyword", ''), COALESCE("WhiteKeyword", ''),
	       COALESCE("IsPowerOn", true), COALESCE("IsWelcomeHint", false), COALESCE("WelcomeMessage", ''),
	       COALESCE("IsMuteEnter", false), COALESCE("MuteEnterCount", 0), COALESCE("IsConfirmNew", false),
	       COALESCE("IsRequirePrefix", false),
	       "InsertDate", "LastDate"
	FROM %s
	WHERE "Oid" = $1
	`, TableGroup)

	g := &Group{}
	err := db.QueryRow(query, groupID).Scan(
		&g.ID, &g.GroupID, &g.TargetGroupID, &g.GroupOpenID, &g.Name,
		&g.RecallKeyword, &g.WarnKeyword, &g.MuteKeyword,
		&g.KickKeyword, &g.BlackKeyword, &g.WhiteKeyword,
		&g.IsPowerOn, &g.IsWelcomeHint, &g.WelcomeMessage,
		&g.IsMuteEnter, &g.MuteEnterCount, &g.IsConfirmNew,
		&g.IsRequirePrefix,
		&g.CreatedAt, &g.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取群组失败: %w", err)
	}

	return g, nil
}

// GroupMember 定义群成员模型
type GroupMember struct {
	UserID    int64     `json:"user_id"`
	GroupID   int64     `json:"group_id"`
	IsShutup  bool      `json:"is_shutup"`
	CreatedAt time.Time `json:"created_at"`
}

// GetGroupMember 获取群成员信息
func GetGroupMember(db *sql.DB, userID int64, groupID int64) (*GroupMember, error) {
	query := fmt.Sprintf(`
	SELECT "UserId", "GroupId", "IsShutup", "InsertDate"
	FROM %s
	WHERE "UserId" = $1 AND "GroupId" = $2
	`, TableGroupMember)

	gm := &GroupMember{}
	err := db.QueryRow(query, userID, groupID).Scan(
		&gm.UserID, &gm.GroupID, &gm.IsShutup, &gm.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取群成员失败: %w", err)
	}

	return gm, nil
}

// CreateGroupMember 创建群成员
func CreateGroupMember(db *sql.DB, userID int64, groupID int64) error {
	query := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "IsShutup", "InsertDate")
	VALUES ($1, $2, false, CURRENT_TIMESTAMP)
	ON CONFLICT ("UserId", "GroupId") DO NOTHING
	`, TableGroupMember)

	_, err := db.Exec(query, userID, groupID)
	return err
}

// UpdateGroupMemberShutup 更新群成员禁言状态
func UpdateGroupMemberShutup(db *sql.DB, userID int64, groupID int64, isShutup bool) error {
	query := fmt.Sprintf(`
	UPDATE %s
	SET "IsShutup" = $3
	WHERE "UserId" = $1 AND "GroupId" = $2
	`, TableGroupMember)

	_, err := db.Exec(query, userID, groupID, isShutup)
	return err
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, user *User) error {
	query := fmt.Sprintf(`
	INSERT INTO %s ("Id", "UserOpenid", "Name", "IsSuper", "Credit", "SaveCredit")
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT ("Id") DO UPDATE
	SET "UserOpenid" = $2, "Name" = $3, "IsSuper" = $4, "Credit" = $5, "SaveCredit" = $6
	`, TableUser)

	_, err := db.Exec(query, user.UserID, user.UserOpenID, user.Nickname, user.IsSuperPoints, user.Points, user.SavingsPoints)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByUserID 根据用户ID获取用户信息
func GetUserByUserID(db *sql.DB, userID int64) (*User, error) {
	query := fmt.Sprintf(`
	SELECT "Id", "Id", "UserOpenid", "Name", "IsSuper", "Credit", "SaveCredit", "FreezeCredit", "InsertDate", "UpgradeDate"
	FROM %s
	WHERE "Id" = $1
	`, TableUser)

	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.UserID, &user.UserOpenID, &user.Nickname, &user.IsSuperPoints, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
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
	query := fmt.Sprintf(`
	UPDATE %s
	SET "Name" = $2, "IsSuper" = $3, "Credit" = $4, "SaveCredit" = $5
	WHERE "Id" = $1
	`, TableUser)

	result, err := db.Exec(query, user.UserID, user.Nickname, user.IsSuperPoints, user.Points, user.SavingsPoints)
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
	query := fmt.Sprintf(`
	INSERT INTO %s ("Id", "UserOpenid", "Name", "IsSuper", "Credit", "SaveCredit")
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT ("Id") DO UPDATE
	SET "UserOpenid" = $2, "Name" = $3, "IsSuper" = $4, "Credit" = $5, "SaveCredit" = $6
	`, TableUser)

	_, err := tx.Exec(query, user.UserID, user.UserOpenID, user.Nickname, user.IsSuperPoints, user.Points, user.SavingsPoints)
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
	query := fmt.Sprintf(`
	SELECT "Id", "Id", "UserOpenid", "Name", '', '', "IsSuper", "Credit", "SaveCredit", "FreezeCredit", "InsertDate", "UpgradeDate"
	FROM %s
	ORDER BY "InsertDate" DESC
	`, TableUser)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取所有用户失败: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		var avatar, gender string
		err := rows.Scan(
			&user.ID, &user.UserID, &user.UserOpenID, &user.Nickname, &avatar, &gender, &user.IsSuperPoints, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
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
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, TableUser)
	if err := db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取分页数据
	query := fmt.Sprintf(`
	SELECT "Id", "Id", "UserOpenid", "Name", '', '', "IsSuper", "Credit", "SaveCredit", "FreezeCredit", "InsertDate", "UpgradeDate"
	FROM %s
	ORDER BY "InsertDate" DESC
	LIMIT $1 OFFSET $2
	`, TableUser)

	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("分页获取用户失败: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		var avatar, gender string
		err := rows.Scan(
			&user.ID, &user.UserID, &user.UserOpenID, &user.Nickname, &avatar, &gender, &user.IsSuperPoints, &user.Points, &user.SavingsPoints, &user.FrozenPoints, &user.CreatedAt, &user.UpdatedAt,
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
	query := fmt.Sprintf(`
	DELETE FROM %s
	WHERE "Id" = $1
	`, TableUser)

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

	query := fmt.Sprintf(`
	SELECT "Id", "GroupId", "Question", "Question", 'approved', "UserId", 0, "InsertDate", "InsertDate"
	FROM %s
	WHERE "Question" = $1
	`, TableQuestion)

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

	query := fmt.Sprintf(`
	INSERT INTO %s ("GroupId", "Question", "UserId", "InsertDate")
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	RETURNING "Id", "GroupId", "Question", "Question", 'approved', "UserId", 0, "InsertDate", "InsertDate"
	`, TableQuestion)

	row := dbConn.QueryRow(query, q.GroupID, q.QuestionRaw, q.CreatedBy)

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

	query := fmt.Sprintf(`
	INSERT INTO %s ("QuestionId", "Answer", "InsertDate", "UpdateDate", "UsedTimes")
	VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0)
	RETURNING "Id", "QuestionId", "Answer", 'approved', 0, "InsertDate", "UpdateDate"
	`, TableAnswer)

	row := dbConn.QueryRow(query, a.QuestionID, a.Answer)

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

	query := fmt.Sprintf(`
	SELECT "Id", "QuestionId", "Answer", 'approved', 0, "InsertDate", "UpdateDate"
	FROM %s
	WHERE "QuestionId" = $1
	ORDER BY "Id" ASC
	`, TableAnswer)

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
	// sz84_robot 中 Question 表没有引用计数字段，此处不做操作
	return nil
}

func IncrementAnswerUsage(dbConn *sql.DB, answerID int) error {
	if dbConn == nil || answerID == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE %s
	SET "UsedTimes" = "UsedTimes" + 1,
	    "UpdateDate" = NOW()
	WHERE "Id" = $1
	`, TableAnswer)

	if _, err := dbConn.Exec(query, answerID); err != nil {
		return fmt.Errorf("更新答案使用次数失败: %w", err)
	}

	return nil
}

func IncrementAnswerShortIntervalUsageIfRecent(dbConn *sql.DB, answerID int) error {
	// sz84_robot 中 Answer 表没有短间隔使用计数字段，此处不做操作
	return nil
}

// BlackList 定义黑名单模型
type BlackList struct {
	ID         int64     `json:"id"`
	GroupID    int64     `json:"group_id"`
	GroupName  string    `json:"group_name"`
	BlackID    int64     `json:"black_id"`
	IsBlack    bool      `json:"is_black"`
	InsertDate time.Time `json:"insert_date"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	BlackInfo  string    `json:"black_info"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	BotUin     int64     `json:"bot_uin"`
	BlackTimes int64     `json:"black_times"`
}

// GetBlackList 获取黑名单记录
func GetBlackList(db *sql.DB, groupID int64, blackID int64) (*BlackList, error) {
	query := fmt.Sprintf(`
	SELECT "Id", "GroupId", "GroupName", "BlackId", "IsBlack", "InsertDate", "StartDate", "EndDate", "BlackInfo", "UserId", "UserName", "BotUin", "BlackTimes"
	FROM %s
	WHERE "GroupId" = $1 AND "BlackId" = $2
	`, TableBlackList)

	bl := &BlackList{}
	err := db.QueryRow(query, groupID, blackID).Scan(
		&bl.ID, &bl.GroupID, &bl.GroupName, &bl.BlackID, &bl.IsBlack, &bl.InsertDate, &bl.StartDate, &bl.EndDate, &bl.BlackInfo, &bl.UserID, &bl.UserName, &bl.BotUin, &bl.BlackTimes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取黑名单失败: %w", err)
	}

	return bl, nil
}

// WhiteList 定义白名单模型
type WhiteList struct {
	ID         int64     `json:"id"`
	GroupID    int64     `json:"group_id"`
	GroupName  string    `json:"group_name"`
	WhiteID    int64     `json:"white_id"`
	InsertDate time.Time `json:"insert_date"`
	WhiteInfo  string    `json:"white_info"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	BotUin     int64     `json:"bot_uin"`
}

// GetWhiteList 获取白名单记录
func GetWhiteList(db *sql.DB, groupID int64, whiteID int64) (*WhiteList, error) {
	query := fmt.Sprintf(`
	SELECT "Id", "GroupId", "GroupName", "WhiteId", "InsertDate", "WhiteInfo", "UserId", "UserName", "BotUin"
	FROM %s
	WHERE "GroupId" = $1 AND "WhiteId" = $2
	`, TableWhiteList)

	wl := &WhiteList{}
	err := db.QueryRow(query, groupID, whiteID).Scan(
		&wl.ID, &wl.GroupID, &wl.GroupName, &wl.WhiteID, &wl.InsertDate, &wl.WhiteInfo, &wl.UserID, &wl.UserName, &wl.BotUin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取白名单失败: %w", err)
	}

	return wl, nil
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

// IsGroupCreditSystemEnabled 检查群积分系统是否开启
func IsGroupCreditSystemEnabled(db *sql.DB, groupID int64) (bool, error) {
	if groupID <= 0 {
		return false, nil
	}
	var enabled sql.NullBool
	query := fmt.Sprintf(`SELECT "IsCreditSystem" FROM %s WHERE "Oid" = $1`, TableGroup)
	err := db.QueryRow(query, groupID).Scan(&enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return enabled.Bool, nil
}

// IsRobotOwner 检查用户是否为机器人主人
func IsRobotOwner(db *sql.DB, groupID int64, userID int64) (bool, error) {
	if groupID <= 0 {
		return false, nil
	}
	var ownerID sql.NullInt64
	query := fmt.Sprintf(`SELECT "RobotOwner" FROM %s WHERE "Oid" = $1`, TableGroup)
	err := db.QueryRow(query, groupID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return ownerID.Int64 == userID, nil
}

// GetPoints 获取用户积分 (自动路由全局或群积分)
func GetPoints(db *sql.DB, userID int64, groupID int64) (int64, error) {
	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return 0, fmt.Errorf("检查群积分模式失败: %w", err)
	}

	var points int64
	var query string
	if isGroupActive {
		query = fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember)
		err := db.QueryRow(query, userID, groupID).Scan(&points)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, fmt.Errorf("获取群积分失败: %w", err)
		}
	} else {
		query = fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser)
		err := db.QueryRow(query, userID).Scan(&points)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, fmt.Errorf("获取全局积分失败: %w", err)
		}
	}
	return points, nil
}

func GetFrozenPoints(db *sql.DB, userID int64, groupID int64) (int64, error) {
	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return 0, fmt.Errorf("检查群积分模式失败: %w", err)
	}

	var points int64
	var query string
	if isGroupActive {
		query = fmt.Sprintf(`SELECT "FreezeCredit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember)
		err := db.QueryRow(query, userID, groupID).Scan(&points)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, fmt.Errorf("获取群冻结积分失败: %w", err)
		}
	} else {
		query = fmt.Sprintf(`SELECT "FreezeCredit" FROM %s WHERE "Id" = $1`, TableUser)
		err := db.QueryRow(query, userID).Scan(&points)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, fmt.Errorf("获取全局冻结积分失败: %w", err)
		}
	}
	return points, nil
}

// AddPoints 增加或扣除用户积分 (自动路由全局或群积分)
func AddPoints(db *sql.DB, botUin int64, userID int64, groupID int64, amount int64, reason string, category string) error {
	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return fmt.Errorf("检查群积分模式失败: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var logType string

	if isGroupActive {
		logType = "group"

		// 更新 GroupMember 表积分
		query := fmt.Sprintf(`
		UPDATE %s
		SET "Credit" = "Credit" + $3
		WHERE "UserId" = $1 AND "GroupId" = $2
		`, TableGroupMember)
		result, err := tx.Exec(query, userID, groupID, amount)
		if err != nil {
			return fmt.Errorf("更新 GroupMember 积分失败: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// 如果记录不存在，不自动插入，因为这通常由加群事件触发
			return fmt.Errorf("群成员记录不存在: user=%d, group=%d", userID, groupID)
		}
	} else {
		logType = "global"

		// 更新 User 表积分
		query := fmt.Sprintf(`
		UPDATE %s
		SET "Credit" = "Credit" + $2
		WHERE "Id" = $1
		`, TableUser)
		result, err := tx.Exec(query, userID, amount)
		if err != nil {
			return fmt.Errorf("更新 User 积分失败: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			insertQuery := fmt.Sprintf(`
			INSERT INTO %s ("Id", "Credit", "InsertDate")
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			`, TableUser)
			if _, err := tx.Exec(insertQuery, userID, amount); err != nil {
				return fmt.Errorf("插入 User 并初始化积分失败: %w", err)
			}
		}
	}

	// 记录日志到 Credit 表
	var newPoints int64
	if isGroupActive {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember), userID, groupID).Scan(&newPoints)
	} else {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)
	}

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, groupID, botUin, amount, newPoints, fmt.Sprintf("%s [%s] (%s)", reason, category, logType))
	if err != nil {
		return fmt.Errorf("记录积分日志到 Credit 表失败: %w", err)
	}

	return tx.Commit()
}

func applySavingsInterestTx(tx *sql.Tx, botUin int64, userID int64) (int, error) {
	var savings int
	var lastInterest sql.NullTime

	// 1. 获取存款金额
	query := fmt.Sprintf(`SELECT "SaveCredit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
	err := tx.QueryRow(query, userID).Scan(&savings)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("查询存款失败: %w", err)
	}

	// 2. 获取上次结息时间
	err = tx.QueryRow(fmt.Sprintf(`SELECT "LastInterestAt" FROM %s WHERE "UserId" = $1 FOR UPDATE`, TableSavings), userID).Scan(&lastInterest)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("查询结息元数据失败: %w", err)
	}

	now := time.Now()
	if savings <= 0 {
		// 如果没有存款，只更新结息时间
		upsertQuery := fmt.Sprintf(`
		INSERT INTO %s ("UserId", "LastInterestAt", "UpdateDate")
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT ("UserId") DO UPDATE
		SET "LastInterestAt" = $2, "UpdateDate" = CURRENT_TIMESTAMP
		`, TableSavings)
		_, err = tx.Exec(upsertQuery, userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新结息时间失败: %w", err)
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
		// 如果不足一天，只更新结息时间（或者不更新也可以，这里选择更新）
		upsertQuery := fmt.Sprintf(`
		INSERT INTO %s ("UserId", "LastInterestAt", "UpdateDate")
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT ("UserId") DO UPDATE
		SET "LastInterestAt" = $2, "UpdateDate" = CURRENT_TIMESTAMP
		`, TableSavings)
		_, err = tx.Exec(upsertQuery, userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新结息时间失败: %w", err)
		}
		return 0, nil
	}

	dailyRate := 0.0005
	interest := int(float64(savings) * dailyRate * float64(days))
	if interest <= 0 {
		upsertQuery := fmt.Sprintf(`
		INSERT INTO %s ("UserId", "LastInterestAt", "UpdateDate")
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT ("UserId") DO UPDATE
		SET "LastInterestAt" = $2, "UpdateDate" = CURRENT_TIMESTAMP
		`, TableSavings)
		_, err = tx.Exec(upsertQuery, userID, now)
		if err != nil {
			return 0, fmt.Errorf("更新结息时间失败: %w", err)
		}
		return 0, nil
	}

	newSavings := savings + interest

	// 3. 更新存款和结息时间
	updateUserQuery := fmt.Sprintf(`UPDATE %s SET "SaveCredit" = $1 WHERE "Id" = $2`, TableUser)
	_, err = tx.Exec(updateUserQuery, newSavings, userID)
	if err != nil {
		return 0, fmt.Errorf("更新存款利息失败: %w", err)
	}

	upsertQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "LastInterestAt", "UpdateDate")
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT ("UserId") DO UPDATE
	SET "LastInterestAt" = $2, "UpdateDate" = CURRENT_TIMESTAMP
	`, TableSavings)
	_, err = tx.Exec(upsertQuery, userID, now)
	if err != nil {
		return 0, fmt.Errorf("更新结息元数据失败: %w", err)
	}

	// 4. 记录日志到 Credit 表
	var newPoints int64
	tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, 0, botUin, int64(interest), newPoints, "存积分利息 [saving_interest]")
	if err != nil {
		return 0, fmt.Errorf("记录利息日志失败: %w", err)
	}

	return interest, nil
}

func DepositPointsToSavings(db *sql.DB, botUin int64, userID int64, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("存入积分必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var points int
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
	err = tx.QueryRow(query, userID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户不存在或没有积分")
		}
		return fmt.Errorf("查询用户积分失败: %w", err)
	}

	if points < amount {
		return fmt.Errorf("积分不足，当前积分为: %d", points)
	}

	updateQuery := fmt.Sprintf(`UPDATE %s SET "Credit" = "Credit" - $1, "SaveCredit" = "SaveCredit" + $1 WHERE "Id" = $2`, TableUser)
	_, err = tx.Exec(updateQuery, amount, userID)
	if err != nil {
		return fmt.Errorf("存款失败: %w", err)
	}

	_, err = applySavingsInterestTx(tx, botUin, userID)
	if err != nil {
		return err
	}

	// 记录日志到 Credit 表
	var newPoints int64
	tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, 0, botUin, -amount, newPoints, "存入积分 [saving_deposit]")
	if err != nil {
		return fmt.Errorf("记录存入积分日志失败: %w", err)
	}

	return tx.Commit()
}

func WithdrawPointsFromSavings(db *sql.DB, botUin int64, userID int64, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("取出积分必须大于0")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	_, err = applySavingsInterestTx(tx, botUin, userID)
	if err != nil {
		return err
	}

	var balance int
	query := fmt.Sprintf(`SELECT "SaveCredit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
	err = tx.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("没有存款记录")
		}
		return fmt.Errorf("查询存款失败: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("存款余额不足，当前余额为: %d", balance)
	}

	updateQuery := fmt.Sprintf(`UPDATE %s SET "SaveCredit" = "SaveCredit" - $1, "Credit" = "Credit" + $1 WHERE "Id" = $2`, TableUser)
	_, err = tx.Exec(updateQuery, amount, userID)
	if err != nil {
		return fmt.Errorf("取款失败: %w", err)
	}

	// 记录日志到 Credit 表
	var newPoints int64
	tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, 0, botUin, amount, newPoints, "取出积分 [saving_withdraw]")
	if err != nil {
		return fmt.Errorf("记录取出积分日志失败: %w", err)
	}

	return tx.Commit()
}

func GetSavingsPoints(db *sql.DB, botUin int64, userID int64) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	_, err = applySavingsInterestTx(tx, botUin, userID)
	if err != nil {
		return 0, err
	}

	var balance int
	query := fmt.Sprintf(`SELECT "SaveCredit" FROM %s WHERE "Id" = $1`, TableUser)
	err = tx.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			balance = 0
		} else {
			return 0, fmt.Errorf("查询存款失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return balance, nil
}

func FreezePoints(db *sql.DB, botUin int64, userID int64, groupID int64, amount int64, reason string) error {
	if amount <= 0 {
		return fmt.Errorf("冻结积分数量必须大于0")
	}

	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return fmt.Errorf("检查群积分模式失败: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var logType string

	if isGroupActive {
		logType = "group"

		var available int64
		query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2 FOR UPDATE`, TableGroupMember)
		err = tx.QueryRow(query, userID, groupID).Scan(&available)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("群成员不存在或没有积分")
			}
			return fmt.Errorf("查询群成员积分失败: %w", err)
		}

		if available < amount {
			return fmt.Errorf("群可用积分不足，当前积分为: %d", available)
		}

		updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET "Credit" = "Credit" - $3, "FreezeCredit" = "FreezeCredit" + $3
		WHERE "UserId" = $1 AND "GroupId" = $2
		`, TableGroupMember)
		_, err = tx.Exec(updateQuery, userID, groupID, amount)
		if err != nil {
			return fmt.Errorf("更新群成员冻结积分失败: %w", err)
		}
	} else {
		logType = "global"

		var available int64
		query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
		err = tx.QueryRow(query, userID).Scan(&available)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("用户不存在或没有积分")
			}
			return fmt.Errorf("查询全局用户积分失败: %w", err)
		}

		if available < amount {
			return fmt.Errorf("全局可用积分不足，当前积分为: %d", available)
		}

		updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET "Credit" = "Credit" - $2, "FreezeCredit" = "FreezeCredit" + $2, "UpgradeDate" = CURRENT_TIMESTAMP
		WHERE "Id" = $1
		`, TableUser)
		_, err = tx.Exec(updateQuery, userID, amount)
		if err != nil {
			return fmt.Errorf("更新全局冻结积分失败: %w", err)
		}
	}

	// 记录日志到 Credit 表
	var newPoints int64
	if isGroupActive {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember), userID, groupID).Scan(&newPoints)
	} else {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)
	}

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, groupID, botUin, -amount, newPoints, reason+fmt.Sprintf(" [freeze] (%s)", logType))
	if err != nil {
		return fmt.Errorf("记录冻结积分日志失败: %w", err)
	}

	return tx.Commit()
}

func UnfreezePoints(db *sql.DB, botUin int64, userID int64, groupID int64, amount int64, reason string) error {
	if amount <= 0 {
		return fmt.Errorf("解冻积分数量必须大于0")
	}

	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return fmt.Errorf("检查群积分模式失败: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var logType string

	if isGroupActive {
		logType = "group"

		var frozen int64
		query := fmt.Sprintf(`SELECT "FreezeCredit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2 FOR UPDATE`, TableGroupMember)
		err = tx.QueryRow(query, userID, groupID).Scan(&frozen)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("没有群成员冻结积分记录")
			}
			return fmt.Errorf("查询群成员冻结积分失败: %w", err)
		}

		if frozen < amount {
			return fmt.Errorf("群成员冻结积分不足，当前冻结积分为: %d", frozen)
		}

		updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET "FreezeCredit" = "FreezeCredit" - $3, "Credit" = "Credit" + $3
		WHERE "UserId" = $1 AND "GroupId" = $2
		`, TableGroupMember)
		_, err = tx.Exec(updateQuery, userID, groupID, amount)
		if err != nil {
			return fmt.Errorf("更新群成员解冻积分失败: %w", err)
		}
	} else {
		logType = "global"

		var frozen int64
		query := fmt.Sprintf(`SELECT "FreezeCredit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
		err = tx.QueryRow(query, userID).Scan(&frozen)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("没有全局冻结积分记录")
			}
			return fmt.Errorf("查询全局冻结积分失败: %w", err)
		}

		if frozen < amount {
			return fmt.Errorf("全局冻结积分不足，当前冻结积分为: %d", frozen)
		}

		updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET "FreezeCredit" = "FreezeCredit" - $2, "Credit" = "Credit" + $2, "UpgradeDate" = CURRENT_TIMESTAMP
		WHERE "Id" = $1
		`, TableUser)
		_, err = tx.Exec(updateQuery, userID, amount)
		if err != nil {
			return fmt.Errorf("更新全局解冻积分失败: %w", err)
		}
	}

	// 记录日志到 Credit 表
	var newPoints int64
	if isGroupActive {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember), userID, groupID).Scan(&newPoints)
	} else {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), userID).Scan(&newPoints)
	}

	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, groupID, botUin, amount, newPoints, reason+fmt.Sprintf(" [unfreeze] (%s)", logType))
	if err != nil {
		return fmt.Errorf("记录解冻积分日志失败: %w", err)
	}

	return tx.Commit()
}

// TransferPoints 积分转账
func TransferPoints(db *sql.DB, botUin int64, fromUserID, toUserID int64, groupID int64, amount int64, reason string, category string) error {
	if amount <= 0 {
		return fmt.Errorf("转账金额必须大于0")
	}

	isGroupActive, err := IsGroupCreditSystemEnabled(db, groupID)
	if err != nil {
		return fmt.Errorf("检查群积分模式失败: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var logType string

	if isGroupActive {
		logType = "group"

		// 1. 检查并锁定转出者积分
		var fromPoints int64
		queryFrom := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2 FOR UPDATE`, TableGroupMember)
		err = tx.QueryRow(queryFrom, fromUserID, groupID).Scan(&fromPoints)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("转出用户在群内不存在或没有积分")
			}
			return fmt.Errorf("查询转出用户群积分失败: %w", err)
		}

		if fromPoints < amount {
			return fmt.Errorf("群积分不足，当前积分为: %d", fromPoints)
		}

		// 2. 更新转出者积分
		updateFrom := fmt.Sprintf(`UPDATE %s SET "Credit" = "Credit" - $3 WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember)
		_, err = tx.Exec(updateFrom, fromUserID, groupID, amount)
		if err != nil {
			return fmt.Errorf("扣除群积分失败: %w", err)
		}

		// 3. 增加接收者积分
		updateTo := fmt.Sprintf(`UPDATE %s SET "Credit" = "Credit" + $3 WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember)
		result, err := tx.Exec(updateTo, toUserID, groupID, amount)
		if err != nil {
			return fmt.Errorf("增加接收者群积分失败: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// 群模式下，如果接收者不在群里，通常不自动创建记录
			return fmt.Errorf("接收者不在群内，无法完成转账")
		}
	} else {
		logType = "global"

		// 1. 检查并锁定转出者积分
		var fromPoints int64
		queryFrom := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
		err = tx.QueryRow(queryFrom, fromUserID).Scan(&fromPoints)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("转出用户不存在或没有积分")
			}
			return fmt.Errorf("查询转出用户全局积分失败: %w", err)
		}

		if fromPoints < amount {
			return fmt.Errorf("全局积分不足，当前积分为: %d", fromPoints)
		}

		// 2. 更新转出者积分
		updateFrom := fmt.Sprintf(`UPDATE %s SET "Credit" = "Credit" - $1 WHERE "Id" = $2`, TableUser)
		_, err = tx.Exec(updateFrom, amount, fromUserID)
		if err != nil {
			return fmt.Errorf("扣除全局积分失败: %w", err)
		}

		// 3. 增加接收者积分（如果不存在则插入）
		updateTo := fmt.Sprintf(`UPDATE %s SET "Credit" = "Credit" + $1 WHERE "Id" = $2`, TableUser)
		result, err := tx.Exec(updateTo, amount, toUserID)
		if err != nil {
			return fmt.Errorf("增加接收者全局积分失败: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			insertQuery := fmt.Sprintf(`
			INSERT INTO %s ("Id", "Credit", "InsertDate")
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			`, TableUser)
			if _, err := tx.Exec(insertQuery, toUserID, amount); err != nil {
				return fmt.Errorf("插入接收者全局积分失败: %w", err)
			}
		}
	}

	// 4. 记录日志到 Credit 表 (两条记录)
	creditLogQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, TableCredit)

	// 为转出者记录
	var fromNewPoints int64
	if isGroupActive {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember), fromUserID, groupID).Scan(&fromNewPoints)
	} else {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), fromUserID).Scan(&fromNewPoints)
	}
	_, err = tx.Exec(creditLogQuery, fromUserID, groupID, botUin, -amount, fromNewPoints, fmt.Sprintf("转账给 %d: %s [%s] (%s)", toUserID, reason, category, logType))
	if err != nil {
		return fmt.Errorf("记录转出日志失败: %w", err)
	}

	// 为接收者记录
	var toNewPoints int64
	if isGroupActive {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember), toUserID, groupID).Scan(&toNewPoints)
	} else {
		tx.QueryRow(fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1`, TableUser), toUserID).Scan(&toNewPoints)
	}
	_, err = tx.Exec(creditLogQuery, toUserID, groupID, botUin, amount, toNewPoints, fmt.Sprintf("来自 %d 的转账: %s [%s] (%s)", fromUserID, reason, category, logType))
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

// AddGroupWhitelistUser 添加群白名单用户
func AddGroupWhitelistUser(db *sql.DB, groupID, userID int64) error {
	// 先检查是否存在，因为没有唯一索引，直接插入会导致重复，但 ON CONFLICT 又报错
	var exists bool
	checkQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE "GroupId" = $1 AND "WhiteId" = $2)`, TableWhiteList)
	err := db.QueryRow(checkQuery, groupID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	query := fmt.Sprintf(`
	INSERT INTO %s ("GroupId", "WhiteId", "UserId", "InsertDate")
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, TableWhiteList)

	_, err = db.Exec(query, groupID, userID, userID)
	if err != nil {
		return fmt.Errorf("添加群白名单用户失败: %w", err)
	}

	return nil
}

// RemoveGroupWhitelistUser 移除群白名单用户
func RemoveGroupWhitelistUser(db *sql.DB, groupID, userID int64) error {
	query := fmt.Sprintf(`
	DELETE FROM %s
	WHERE "GroupId" = $1 AND "WhiteId" = $2
	`, TableWhiteList)

	_, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("移除群白名单用户失败: %w", err)
	}

	return nil
}

// IsUserInGroupWhitelist 检查用户是否在群白名单中
func IsUserInGroupWhitelist(db *sql.DB, groupID, userID int64) (bool, error) {
	query := fmt.Sprintf(`
	SELECT EXISTS(SELECT 1 FROM %s WHERE "GroupId" = $1 AND "WhiteId" = $2)
	`, TableWhiteList)

	var exists bool
	err := db.QueryRow(query, groupID, userID).Scan(&exists)
	return exists, err
}

// AddGroupBlacklistUser 添加群黑名单用户
func AddGroupBlacklistUser(db *sql.DB, groupID, userID int64, reason string) error {
	// 先检查是否存在
	var exists bool
	checkQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE "GroupId" = $1 AND "BlackId" = $2)`, TableBlackList)
	err := db.QueryRow(checkQuery, groupID, userID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		query := fmt.Sprintf(`
		UPDATE %s
		SET "IsBlack" = true, "BlackInfo" = $3, "InsertDate" = CURRENT_TIMESTAMP
		WHERE "GroupId" = $1 AND "BlackId" = $2
		`, TableBlackList)
		_, err = db.Exec(query, groupID, userID, reason)
	} else {
		query := fmt.Sprintf(`
		INSERT INTO %s ("GroupId", "BlackId", "UserId", "IsBlack", "BlackInfo", "InsertDate")
		VALUES ($1, $2, $2, true, $3, CURRENT_TIMESTAMP)
		`, TableBlackList)
		_, err = db.Exec(query, groupID, userID, reason)
	}

	if err != nil {
		return fmt.Errorf("添加群黑名单用户失败: %w", err)
	}

	return nil
}

// RemoveGroupBlacklistUser 移除群黑名单用户
func RemoveGroupBlacklistUser(db *sql.DB, groupID, userID int64) error {
	query := fmt.Sprintf(`
	UPDATE %s
	SET "IsBlack" = false
	WHERE "GroupId" = $1 AND "BlackId" = $2
	`, TableBlackList)

	_, err := db.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("移除群黑名单用户失败: %w", err)
	}

	return nil
}

// IsUserInGroupBlacklist 检查用户是否在群黑名单中
func IsUserInGroupBlacklist(db *sql.DB, groupID, userID int64) (bool, error) {
	query := fmt.Sprintf(`
	SELECT EXISTS(SELECT 1 FROM %s WHERE "GroupId" = $1 AND "BlackId" = $2 AND "IsBlack" = true)
	`, TableBlackList)

	var exists bool
	err := db.QueryRow(query, groupID, userID).Scan(&exists)
	return exists, err
}

// UpdateGroupKeyword 更新群组关键词（增量更新）
func UpdateGroupKeyword(db *sql.DB, groupID int64, column string, word string, action string) error {
	// 获取当前关键词
	var currentKeywords string
	query := fmt.Sprintf(`SELECT COALESCE("%s", '') FROM %s WHERE "Oid" = $1`, column, TableGroup)
	err := db.QueryRow(query, groupID).Scan(&currentKeywords)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("群组不存在: %d", groupID)
		}
		return fmt.Errorf("获取群组关键词失败: %w", err)
	}

	keywords := strings.Split(currentKeywords, ",")
	newKeywords := []string{}
	found := false

	for _, k := range keywords {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if k == word {
			found = true
			if action == "remove" {
				continue
			}
		}
		newKeywords = append(newKeywords, k)
	}

	if action == "add" && !found {
		newKeywords = append(newKeywords, word)
	}

	updatedKeywords := strings.Join(newKeywords, ",")
	updateQuery := fmt.Sprintf(`UPDATE %s SET "%s" = $1, "LastDate" = CURRENT_TIMESTAMP WHERE "Oid" = $2`, TableGroup, column)
	_, err = db.Exec(updateQuery, updatedKeywords, groupID)
	if err != nil {
		return fmt.Errorf("更新群组关键词失败: %w", err)
	}

	return nil
}

func ClearGroupWhitelist(db *sql.DB, groupID int64) error {
	query := fmt.Sprintf(`
	DELETE FROM %s
	WHERE "GroupId" = $1
	`, TableWhiteList)

	_, err := db.Exec(query, groupID)
	if err != nil {
		return fmt.Errorf("清空群白名单失败: %w", err)
	}

	return nil
}

func ClearGroupBlacklist(db *sql.DB, groupID int64) error {
	query := fmt.Sprintf(`
	UPDATE %s
	SET "IsBlack" = false
	WHERE "GroupId" = $1
	`, TableBlackList)

	_, err := db.Exec(query, groupID)
	if err != nil {
		return fmt.Errorf("清空群黑名单失败: %w", err)
	}

	return nil
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

// ------------------- 本机积分 (Local Points) 相关操作 -------------------

// GetLocalPoints 获取本机积分
func GetLocalPoints(db *sql.DB, botUin int64, userID int64) (int64, error) {
	var points int64
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "BotUin" = $1 AND "UserId" = $2`, TableFriend)
	err := db.QueryRow(query, botUin, userID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("获取本机积分失败: %w", err)
	}
	return points, nil
}

// UpdateLocalPoints 更新本机积分
func UpdateLocalPoints(db *sql.DB, botUin int64, userID int64, amount int64) error {
	query := fmt.Sprintf(`
	INSERT INTO %s ("BotUin", "UserId", "Credit", "UpdateDate")
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	ON CONFLICT ("BotUin", "UserId") DO UPDATE
	SET "Credit" = %s."Credit" + $3, "UpdateDate" = CURRENT_TIMESTAMP
	`, TableFriend, TableFriend)

	_, err := db.Exec(query, botUin, userID, amount)
	if err != nil {
		return fmt.Errorf("更新本机积分失败: %w", err)
	}
	return nil
}

// ------------------- 消费统计与个人激活相关操作 -------------------

// RecordConsumption 记录消费
func RecordConsumption(db *sql.DB, botUin int64, userID int64, amount int64) error {
	if amount <= 0 {
		return nil
	}
	query := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "BotUin", "Amount", "ConsumeDate")
	VALUES ($1, $2, $3, CURRENT_DATE)
	ON CONFLICT ("UserId", "BotUin", "ConsumeDate") DO UPDATE
	SET "Amount" = %s."Amount" + $3
	`, TableConsumption, TableConsumption)

	_, err := db.Exec(query, userID, botUin, amount)
	if err != nil {
		return fmt.Errorf("记录消费失败: %w", err)
	}
	return nil
}

// GetTodayConsumption 获取用户今日在该机器人的消费总额
func GetTodayConsumption(db *sql.DB, botUin int64, userID int64) (int64, error) {
	var total int64
	query := fmt.Sprintf(`SELECT COALESCE(SUM("Amount"), 0) FROM %s WHERE "UserId" = $1 AND "BotUin" = $2 AND "ConsumeDate" = CURRENT_DATE`, TableConsumption)
	err := db.QueryRow(query, userID, botUin).Scan(&total)
	return total, err
}

// GetRolling12MConsumption 获取用户最近12个月在该机器人的消费总额
func GetRolling12MConsumption(db *sql.DB, botUin int64, userID int64) (int64, error) {
	var total int64
	query := fmt.Sprintf(`SELECT COALESCE(SUM("Amount"), 0) FROM %s WHERE "UserId" = $1 AND "BotUin" = $2 AND "ConsumeDate" >= CURRENT_DATE - INTERVAL '12 months'`, TableConsumption)
	err := db.QueryRow(query, userID, botUin).Scan(&total)
	return total, err
}

// CheckPersonalActivation 检查个人激活状态
func CheckPersonalActivation(db *sql.DB, botUin int64, userID int64) (bool, error) {
	todayTotal, err := GetTodayConsumption(db, botUin, userID)
	if err != nil {
		return false, err
	}
	if todayTotal >= 500 { // 对应 C# 中的 TODAY_TOTAL_THRESHOLD
		return true, nil
	}

	rollingTotal, err := GetRolling12MConsumption(db, botUin, userID)
	if err != nil {
		return false, err
	}
	if rollingTotal >= 1000 { // 对应 C# 中的 ROLLING_12M_THRESHOLD
		return true, nil
	}

	return false, nil
}

// ------------------- 打赏与手动调整 -------------------

// updatePointsTx 在事务中更新全局积分
func updatePointsTx(tx *sql.Tx, userID int64, amount int64, reason string, category string, botUin int64) error {
	var current int64
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "Id" = $1 FOR UPDATE`, TableUser)
	err := tx.QueryRow(query, userID).Scan(&current)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("查询全局积分失败: %w", err)
	}

	if amount < 0 && current < -amount {
		return fmt.Errorf("全局积分不足，当前积分为: %d", current)
	}

	updateQuery := fmt.Sprintf(`
	INSERT INTO %s ("Id", "Credit", "InsertDate", "UpgradeDate")
	VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT ("Id") DO UPDATE
	SET "Credit" = %s."Credit" + $2, "UpgradeDate" = CURRENT_TIMESTAMP
	RETURNING "Credit"
	`, TableUser, TableUser)

	var newBalance int64
	err = tx.QueryRow(updateQuery, userID, amount).Scan(&newBalance)
	if err != nil {
		return fmt.Errorf("更新全局积分失败: %w", err)
	}

	// 记录日志
	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate", "Category")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, $7)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, 0, botUin, amount, newBalance, reason, category)
	return err
}

// GetGroupPoints 获取用户在本群的积分
func GetGroupPoints(db *sql.DB, userID int64, groupID int64) (int64, error) {
	var points int64
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2`, TableGroupMember)
	err := db.QueryRow(query, userID, groupID).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("获取群积分失败: %w", err)
	}
	return points, nil
}

// UpdateGroupPoints 更新用户在本群的积分
func UpdateGroupPoints(db *sql.DB, userID int64, groupID int64, amount int64) error {
	return updateGroupPoints(db, userID, groupID, amount, "插件操作", "group")
}

// updateGroupPointsTx 在事务中更新群积分
func updateGroupPointsTx(tx *sql.Tx, userID int64, groupID int64, amount int64, reason string, category string) error {
	var current int64
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "UserId" = $1 AND "GroupId" = $2 FOR UPDATE`, TableGroupMember)
	err := tx.QueryRow(query, userID, groupID).Scan(&current)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("查询群积分失败: %w", err)
	}

	if amount < 0 && current < -amount {
		return fmt.Errorf("群积分不足，当前积分为: %d", current)
	}

	updateQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "Credit", "InsertDate", "UpdateDate")
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT ("UserId", "GroupId") DO UPDATE
	SET "Credit" = %s."Credit" + $3, "UpdateDate" = CURRENT_TIMESTAMP
	RETURNING "Credit"
	`, TableGroupMember, TableGroupMember)

	var newBalance int64
	err = tx.QueryRow(updateQuery, userID, groupID, amount).Scan(&newBalance)
	if err != nil {
		return fmt.Errorf("更新群积分失败: %w", err)
	}

	// 记录日志
	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "GroupId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate", "Category")
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, $7)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, groupID, 0, amount, newBalance, reason, category)
	return err
}

// updateGroupPoints 非事务版本更新群积分
func updateGroupPoints(db *sql.DB, userID int64, groupID int64, amount int64, reason string, category string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := updateGroupPointsTx(tx, userID, groupID, amount, reason, category); err != nil {
		return err
	}

	return tx.Commit()
}

// TipPoints 打赏积分
func TipPoints(db *sql.DB, botUin int64, fromUserID, toUserID int64, groupID int64, amount int64, tier string) error {
	if amount <= 0 {
		return fmt.Errorf("打赏金额必须大于0")
	}

	// 1. 获取转出者是否为超级用户（免手续费）
	user, err := GetUserByUserID(db, fromUserID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	isSuper := user != nil && user.IsSuperPoints

	// 2. 计算手续费 (20%)
	fee := int64(0)
	if !isSuper {
		fee = int64(float64(amount) * 0.2)
		if fee < 1 && amount > 0 {
			fee = 1 // 最小手续费
		}
	}
	netAmount := amount - fee

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	reason := fmt.Sprintf("打赏给用户 %d", toUserID)
	receiveReason := fmt.Sprintf("收到来自用户 %d 的打赏", fromUserID)

	switch tier {
	case "global":
		// 扣除转出者
		if err := updatePointsTx(tx, fromUserID, -amount, reason, "tip_out", 0); err != nil {
			return err
		}
		// 增加接收者
		if err := updatePointsTx(tx, toUserID, netAmount, receiveReason, "tip_in", 0); err != nil {
			return err
		}
	case "group":
		// 扣除转出者
		if err := updateGroupPointsTx(tx, fromUserID, groupID, -amount, reason, "tip_out"); err != nil {
			return err
		}
		// 增加接收者
		if err := updateGroupPointsTx(tx, toUserID, groupID, netAmount, receiveReason, "tip_in"); err != nil {
			return err
		}
	case "local":
		// 扣除转出者
		if err := updateLocalPointsTx(tx, botUin, fromUserID, -amount, reason, "tip_out"); err != nil {
			return err
		}
		// 增加接收者
		if err := updateLocalPointsTx(tx, botUin, toUserID, netAmount, receiveReason, "tip_in"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("无效的积分层级: %s", tier)
	}

	return tx.Commit()
}

// updateLocalPointsTx 在事务中更新本机积分
func updateLocalPointsTx(tx *sql.Tx, botUin int64, userID int64, amount int64, reason string, category string) error {
	// 检查余额
	var current int64
	query := fmt.Sprintf(`SELECT "Credit" FROM %s WHERE "BotUin" = $1 AND "UserId" = $2 FOR UPDATE`, TableFriend)
	err := tx.QueryRow(query, botUin, userID).Scan(&current)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("查询本机积分失败: %w", err)
	}

	if amount < 0 && current < -amount {
		return fmt.Errorf("本机积分不足，当前积分为: %d", current)
	}

	// 更新积分
	updateQuery := fmt.Sprintf(`
	INSERT INTO %s ("BotUin", "UserId", "Credit", "UpdateDate")
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	ON CONFLICT ("BotUin", "UserId") DO UPDATE
	SET "Credit" = %s."Credit" + $3, "UpdateDate" = CURRENT_TIMESTAMP
	RETURNING "Credit"
	`, TableFriend, TableFriend)

	var newBalance int64
	err = tx.QueryRow(updateQuery, botUin, userID, amount).Scan(&newBalance)
	if err != nil {
		return fmt.Errorf("更新本机积分失败: %w", err)
	}

	// 记录日志
	logQuery := fmt.Sprintf(`
	INSERT INTO %s ("UserId", "BotUin", "CreditAdd", "CreditValue", "CreditInfo", "InsertDate", "Category")
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, $6)
	`, TableCredit)
	_, err = tx.Exec(logQuery, userID, botUin, amount, newBalance, reason, category)
	return err
}

// AdjustPoints 手动调整积分
func AdjustPoints(db *sql.DB, botUin int64, userID int64, groupID int64, amount int64, tier string, reason string) error {
	switch tier {
	case "global":
		return UpdateUserPoints(db, userID, int(amount)) // 注意：UpdateUserPoints 是覆盖式的，而通常调整应该是增量式的。
		// 应该使用增量式更新，这里我重新写一个增量式的。
	case "group":
		return updateGroupPoints(db, userID, groupID, amount, reason, "admin_adjust")
	case "local":
		return UpdateLocalPoints(db, botUin, userID, amount)
	default:
		return fmt.Errorf("无效的积分层级: %s", tier)
	}
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
			_ = AddPoints(db, 0, userID, 0, int64(points), reason, "fission_task")
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
