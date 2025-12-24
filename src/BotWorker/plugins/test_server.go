package plugins

import (
	"database/sql"
	"fmt"
	"time"

	"botworker/internal/onebot"
	"botworker/internal/plugin"
)

// TestServerPlugin 测试服插件
// 用于管理用户的测试服状态和新功能访问权限
type TestServerPlugin struct {
	db        *sql.DB
	robot     plugin.Robot
	cmdParser *CommandParser
}

// UserTestServerStatus 用户测试服状态
// 记录用户是否启用测试服功能
type UserTestServerStatus struct {
	UserID        string    `json:"user_id"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// NewTestServerPlugin 创建新的测试服插件实例
func NewTestServerPlugin() *TestServerPlugin {
	return &TestServerPlugin{
		db:        nil,
		robot:     nil,
		cmdParser: NewCommandParser(),
	}
}

// Name 获取插件名称
func (p *TestServerPlugin) Name() string {
	return "TestServer"
}

// Description 获取插件描述
func (p *TestServerPlugin) Description() string {
	return "测试服功能，允许用户体验机器人新功能"
}

// Version 获取插件版本
func (p *TestServerPlugin) Version() string {
	return "1.0.0"
}

// Init 初始化插件
func (p *TestServerPlugin) Init(robot plugin.Robot) {
	p.robot = robot
	p.db = GlobalDB

	// 初始化数据库表
	p.initDatabase()

	// 注册消息处理
	p.robot.OnMessage(p.handleMessage)

	// 记录插件加载
	if p.db != nil {
		p.logAction("plugin", "TestServer插件已初始化", "system")
	}
}

// initDatabase 初始化测试服相关数据库表
func (p *TestServerPlugin) initDatabase() {
	if p.db == nil {
		return
	}

	// 创建用户测试服状态表
	query := `
	CREATE TABLE IF NOT EXISTS user_test_server_status (
		user_id VARCHAR(255) PRIMARY KEY,
		enabled BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := p.db.Exec(query)
	if err != nil {
		fmt.Printf("初始化测试服状态表失败: %v\n", err)
	}
}

// handleMessage 处理消息事件
func (p *TestServerPlugin) handleMessage(event *onebot.Event) error {
	if event.MessageType != "private" && event.MessageType != "group" {
		return nil
	}

	// 获取用户ID
	userID := event.UserID
	if userID == 0 {
		return nil
	}

	// 检查是否为启用测试服命令
	if match, _ := p.cmdParser.MatchCommand("开启测试服|启用测试服", event.RawMessage); match {
		p.toggleTestServerStatus(event, fmt.Sprintf("%d", userID), true)
		return nil
	}

	// 检查是否为禁用测试服命令
	if match, _ := p.cmdParser.MatchCommand("关闭测试服|禁用测试服", event.RawMessage); match {
		p.toggleTestServerStatus(event, fmt.Sprintf("%d", userID), false)
		return nil
	}

	// 检查是否为查看测试服状态命令
	if match, _ := p.cmdParser.MatchCommand("测试服状态", event.RawMessage); match {
		p.checkTestServerStatus(event, fmt.Sprintf("%d", userID))
		return nil
	}

	// 检查是否为查看测试服帮助命令
	if match, _ := p.cmdParser.MatchCommand("测试服帮助|测试服说明", event.RawMessage); match {
		p.showTestServerHelp(event)
		return nil
	}

	return nil
}

// toggleTestServerStatus 切换用户测试服状态
func (p *TestServerPlugin) toggleTestServerStatus(event *onebot.Event, userID string, enabled bool) {
	if p.db == nil {
		p.sendMessage(p.robot, event, "⚠️ 数据库未连接，无法使用测试服功能")
		return
	}

	// 更新数据库
	query := `
	INSERT INTO user_test_server_status (user_id, enabled, last_updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT (user_id) DO UPDATE
	SET enabled = $2, last_updated_at = CURRENT_TIMESTAMP;
	`

	_, err := p.db.Exec(query, userID, enabled)
	if err != nil {
		p.sendMessage(p.robot, event, "⚠️ 更新测试服状态失败")
		return
	}

	// 发送结果消息
	if enabled {
		p.sendMessage(p.robot, event, "✅ 测试服功能已启用！您现在可以体验机器人的最新功能")
	} else {
		p.sendMessage(p.robot, event, "✅ 测试服功能已关闭！您将使用机器人的稳定版本")
	}

	// 记录操作
	p.logAction(userID, fmt.Sprintf("切换测试服状态: %t", enabled), "user")
}

// checkTestServerStatus 查看用户测试服状态
func (p *TestServerPlugin) checkTestServerStatus(event *onebot.Event, userID string) {
	if p.db == nil {
		p.sendMessage(p.robot, event, "⚠️ 数据库未连接，无法查询测试服状态")
		return
	}

	// 获取当前状态
	status, err := p.getUserTestServerStatus(userID)
	if err != nil {
		p.sendMessage(p.robot, event, "⚠️ 查询测试服状态失败")
		return
	}

	// 发送状态消息
	statusText := "关闭"
	if status.Enabled {
		statusText = "启用"
	}

	response := fmt.Sprintf("📋 您的测试服状态：%s\n", statusText)
	response += fmt.Sprintf("📅 上次更新：%s", status.LastUpdatedAt.Format("2006-01-02 15:04:05"))

	p.sendMessage(p.robot, event, response)
}

// showTestServerHelp 显示测试服功能说明
func (p *TestServerPlugin) showTestServerHelp(event *onebot.Event) {
	helpMsg := `📚 测试服功能说明

🔹 测试服是机器人新功能的体验环境，您可以在这里率先体验最新开发的功能
🔹 测试服功能可能不稳定，随时可能更新或调整
🔹 您可以随时切换测试服状态

📌 可用命令：
🔸 开启测试服/启用测试服 - 开启测试服功能
🔸 关闭测试服/禁用测试服 - 关闭测试服功能
🔸 测试服状态 - 查看当前测试服状态
🔸 测试服说明 - 查看本说明

💡 提示：新功能会在测试服中优先发布，欢迎您提供反馈！`

	p.sendMessage(p.robot, event, helpMsg)
}

// getUserTestServerStatus 获取用户测试服状态
func (p *TestServerPlugin) getUserTestServerStatus(userID string) (*UserTestServerStatus, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `
	SELECT user_id, enabled, created_at, last_updated_at
	FROM user_test_server_status
	WHERE user_id = $1;
	`

	var status UserTestServerStatus
	err := p.db.QueryRow(query, userID).Scan(
		&status.UserID,
		&status.Enabled,
		&status.CreatedAt,
		&status.LastUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在，返回默认状态
			return &UserTestServerStatus{
				UserID:        userID,
				Enabled:       false,
				CreatedAt:     time.Now(),
				LastUpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}

	return &status, nil
}

// IsUserInTestServer 检查用户是否启用了测试服功能
// 供其他插件调用，用于决定是否向用户开放新功能
func (p *TestServerPlugin) IsUserInTestServer(userID string) bool {
	if p.db == nil {
		return false
	}

	status, err := p.getUserTestServerStatus(userID)
	if err != nil {
		return false
	}

	return status.Enabled
}

// GetAllTestServerUsers 获取所有启用测试服的用户
func (p *TestServerPlugin) GetAllTestServerUsers() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `
	SELECT user_id
	FROM user_test_server_status
	WHERE enabled = true;
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}

	return users, nil
}

// sendMessage 发送消息
func (p *TestServerPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, content string) {
	params := &onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     content,
	}

	robot.SendMessage(params)
}

// logAction 记录操作日志
func (p *TestServerPlugin) logAction(userID, action, actionType string) {
	if p.db == nil {
		return
	}

	// 使用现有的日志记录功能或创建新的日志条目
	// 这里简单地记录到控制台
	fmt.Printf("[%s] %s: %s\n", actionType, userID, action)
}

// SetDB 设置数据库连接
func (p *TestServerPlugin) SetDB(db *sql.DB) {
	p.db = db
}
