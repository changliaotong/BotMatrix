package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
	"time"
)

// BadgePlugin 徽章系统插件
type BadgePlugin struct {
	cmdParser *CommandParser
}

// Badge 徽章定义
type Badge struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Icon        string    `gorm:"size:100" json:"icon"`
	Type        string    `gorm:"size:20" json:"type"` // system, achievement, event
	Condition   string    `gorm:"size:255" json:"condition"` // 获取条件描述
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserBadge 用户持有徽章记录
type UserBadge struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	BadgeID   uint      `json:"badge_id"`
	GrantTime time.Time `json:"grant_time"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BadgeGrantLog 徽章发放日志
type BadgeGrantLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	BadgeID   uint      `json:"badge_id"`
	Operator  string    `gorm:"size:20" json:"operator"` // system, admin, event
	Reason    string    `gorm:"size:255" json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// BadgeConfig 徽章系统配置
type BadgeConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	UpdateAt    time.Time `json:"update_at"`
}

// NewBadgePlugin 创建徽章系统插件实例
func NewBadgePlugin() *BadgePlugin {
	return &BadgePlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *BadgePlugin) Name() string {
	return "badge"
}

func (p *BadgePlugin) Description() string {
	return "徽章系统插件，提供徽章发放、查询和管理功能"
}

func (p *BadgePlugin) Version() string {
	return "1.0.0"
}

func (p *BadgePlugin) Init(robot plugin.Robot) {
	log.Println("加载徽章系统插件")

	// 初始化数据库
	p.initDatabase()

	// 初始化默认徽章
	p.initDefaultBadges()

	// 处理徽章系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		// 我的徽章
		if match, _ := p.cmdParser.MatchCommand("我的徽章", event.RawMessage); match {
			p.myBadges(robot, event)
			return nil
		}

		// 查看徽章
		if match, _ := p.cmdParser.MatchCommand("查看徽章", event.RawMessage); match {
			p.listBadges(robot, event)
			return nil
		}

		// 查看徽章详情
		if match, params := p.cmdParser.MatchCommandWithParams("徽章详情(d+)", event.RawMessage); match && len(params) > 0 {
			badgeID := params[1]
			p.badgeDetail(robot, event, badgeID)
			return nil
		}

		// 管理员命令：发放徽章
		if match, params := p.cmdParser.MatchCommandWithParams("发放徽章(d+)(\d+)", event.RawMessage); match && len(params) > 0 {
			userID := params[1]
			badgeID := params[2]
			p.grantBadge(robot, event, userID, badgeID)
			return nil
		}

		// 管理员命令：移除徽章
		if match, params := p.cmdParser.MatchCommandWithParams("移除徽章(d+)(\d+)", event.RawMessage); match && len(params) > 0 {
			userID := params[1]
			badgeID := params[2]
			p.removeBadge(robot, event, userID, badgeID)
			return nil
		}

		// 管理员命令：开启徽章系统
		if match, _ := p.cmdParser.MatchCommand("开启徽章系统", event.RawMessage); match {
			p.enableSystem(robot, event)
			return nil
		}

		// 管理员命令：关闭徽章系统
		if match, _ := p.cmdParser.MatchCommand("关闭徽章系统", event.RawMessage); match {
			p.disableSystem(robot, event)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
func (p *BadgePlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，徽章系统将无法正常工作")
		return
	}
	
	// 创建徽章表
	createBadgeTable := `
	CREATE TABLE IF NOT EXISTS badge (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) NOT NULL UNIQUE,
		description VARCHAR(255) NOT NULL,
		icon VARCHAR(100) NOT NULL,
		type VARCHAR(20) NOT NULL,
		condition VARCHAR(255) NOT NULL,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := GlobalDB.Exec(createBadgeTable)
	if err != nil {
		log.Printf("创建徽章表失败: %v\n", err)
		return
	}
	
	// 创建用户徽章表
	createUserBadgeTable := `
	CREATE TABLE IF NOT EXISTS user_badge (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		badge_id INT NOT NULL REFERENCES badge(id) ON DELETE CASCADE,
		grant_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, badge_id)
	)
	`
	_, err = GlobalDB.Exec(createUserBadgeTable)
	if err != nil {
		log.Printf("创建用户徽章表失败: %v\n", err)
		return
	}
	
	// 创建徽章发放日志表
	createBadgeGrantLogTable := `
	CREATE TABLE IF NOT EXISTS badge_grant_log (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		badge_id INT NOT NULL REFERENCES badge(id) ON DELETE CASCADE,
		operator VARCHAR(20) NOT NULL,
		reason VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBadgeGrantLogTable)
	if err != nil {
		log.Printf("创建徽章发放日志表失败: %v\n", err)
		return
	}
	
	// 创建徽章系统配置表
	createBadgeConfigTable := `
	CREATE TABLE IF NOT EXISTS badge_config (
		id SERIAL PRIMARY KEY,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBadgeConfigTable)
	if err != nil {
		log.Printf("创建徽章系统配置表失败: %v\n", err)
		return
	}
	
	// 初始化配置
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM badge_config").Scan(&count)
	if err != nil {
		log.Printf("查询徽章系统配置失败: %v\n", err)
		return
	}
	
	if count == 0 {
		_, err = GlobalDB.Exec("INSERT INTO badge_config (is_enabled) VALUES (TRUE)")
		if err != nil {
			log.Printf("初始化徽章系统配置失败: %v\n", err)
			return
		}
	}
	
	log.Println("徽章系统数据库初始化完成")
}

// initDefaultBadges 初始化默认徽章
func (p *BadgePlugin) initDefaultBadges() {
	if GlobalDB == nil {
		return
	}
	
	// 检查是否已有徽章
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM badge").Scan(&count)
	if err != nil {
		log.Printf("查询徽章数量失败: %v\n", err)
		return
	}
	
	if count > 0 {
		return // 已有徽章，不需要初始化
	}
	
	// 初始化默认徽章
	defaultBadges := []Badge{
		{
			Name:        "新手徽章",
			Description: "欢迎加入的证明",
			Icon:        "🎟️",
			Type:        "system",
			Condition:   "新用户注册自动获得",
			IsEnabled:   true,
		},
		{
			Name:        "宝宝达人",
			Description: "宝宝系统忠实用户",
			Icon:        "👶",
			Type:        "achievement",
			Condition:   "宝宝成长值达到10000",
			IsEnabled:   true,
		},
		{
			Name:        "婚姻伴侣",
			Description: "步入婚姻殿堂的证明",
			Icon:        "💍",
			Type:        "achievement",
			Condition:   "成功结婚",
			IsEnabled:   true,
		},
		{
			Name:        "活动参与者",
			Description: "积极参与活动的证明",
			Icon:        "🎉",
			Type:        "event",
			Condition:   "参与指定活动获得",
			IsEnabled:   true,
		},
	}
	
	for _, badge := range defaultBadges {
		_, err := GlobalDB.Exec(
			"INSERT INTO badge (name, description, icon, type, condition, is_enabled) VALUES ($1, $2, $3, $4, $5, $6)",
			badge.Name, badge.Description, badge.Icon, badge.Type, badge.Condition, badge.IsEnabled,
		)
		if err != nil {
			log.Printf("初始化默认徽章失败: %v\n", err)
		}
	}
	
	log.Println("默认徽章初始化完成")
}

// isSystemEnabled 检查徽章系统是否开启
func (p *BadgePlugin) isSystemEnabled() bool {
	if GlobalDB == nil {
		// 如果没有数据库连接，默认返回开启状态
		return true
	}
	
	// 查询系统配置
	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM badge_config LIMIT 1").Scan(&isEnabled)
	if err != nil {
		// 如果查询失败，默认返回开启状态
		log.Printf("查询徽章系统配置失败: %v\n", err)
		return true
	}
	
	return isEnabled
}

// myBadges 我的徽章功能
func (p *BadgePlugin) myBadges(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 查询用户的徽章
	rows, err := GlobalDB.Query(`
		SELECT b.id, b.name, b.description, b.icon, ub.grant_time 
		FROM badge b 
		JOIN user_badge ub ON b.id = ub.badge_id 
		WHERE ub.user_id = ? AND ub.is_active = TRUE AND b.is_enabled = TRUE
	`, event.UserID)
	if err != nil {
		log.Printf("查询用户徽章失败: %v\n", err)
		SendTextReply(robot, event, "❌ 查询失败，请稍后重试")
		return
	}
	defer rows.Close()
	
	var badges []Badge
	var grantTimes []time.Time
	
	for rows.Next() {
		var badge Badge
		var grantTime time.Time
		err := rows.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &grantTime)
		if err != nil {
			log.Printf("扫描用户徽章失败: %v\n", err)
			continue
		}
		badges = append(badges, badge)
		grantTimes = append(grantTimes, grantTime)
	}
	
	if len(badges) == 0 {
		SendTextReply(robot, event, "❌ 您还没有获得任何徽章哦~ 继续努力吧！")
		return
	}
	
	// 构建回复消息
	msg := "🏆 我的徽章\n"
	msg += "================================\n"
	
	for i, badge := range badges {
		msg += badge.Icon + " " + badge.Name + "\n"
		msg += "   " + badge.Description + "\n"
		msg += "   获得时间: " + grantTimes[i].Format("2006-01-02") + "\n"
	}
	
	msg += "================================\n"
	msg += "💡 发送【查看徽章】了解更多徽章信息"
	
	SendTextReply(robot, event, msg)
}

// listBadges 查看所有徽章功能
func (p *BadgePlugin) listBadges(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 查询所有启用的徽章
	rows, err := GlobalDB.Query("SELECT id, name, description, icon, type, condition FROM badge WHERE is_enabled = TRUE")
	if err != nil {
		log.Printf("查询徽章列表失败: %v\n", err)
		SendTextReply(robot, event, "❌ 查询失败，请稍后重试")
		return
	}
	defer rows.Close()
	
	var badges []Badge
	
	for rows.Next() {
		var badge Badge
		err := rows.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition)
		if err != nil {
			log.Printf("扫描徽章列表失败: %v\n", err)
			continue
		}
		badges = append(badges, badge)
	}
	
	if len(badges) == 0 {
		SendTextReply(robot, event, "❌ 暂无可用徽章")
		return
	}
	
	// 构建回复消息
	msg := "🏅 徽章列表\n"
	msg += "================================\n"
	
	for _, badge := range badges {
		msg += badge.Icon + " " + badge.Name + "\n"
		msg += "   " + badge.Description + "\n"
		msg += "   类型: " + badge.Type + "\n"
		msg += "   条件: " + badge.Condition + "\n"
		msg += "\n"
	}
	
	msg += "================================\n"
	msg += "💡 发送【徽章详情+徽章ID】查看徽章详细信息"
	
	SendTextReply(robot, event, msg)
}

// badgeDetail 徽章详情功能
func (p *BadgePlugin) badgeDetail(robot plugin.Robot, event *onebot.Event, badgeID string) {
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 查询徽章详情
	var badge Badge
	row := GlobalDB.QueryRow("SELECT id, name, description, icon, type, condition, is_enabled FROM badge WHERE id = ?", badgeID)
	err := row.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition, &badge.IsEnabled)
	if err != nil {
		SendTextReply(robot, event, "❌ 徽章不存在或已被禁用")
		return
	}
	
	if !badge.IsEnabled {
		SendTextReply(robot, event, "❌ 该徽章已被禁用")
		return
	}
	
	// 构建回复消息
	msg := "🏅 徽章详情\n"
	msg += "================================\n"
	msg += "ID: " + IntToString(int(badge.ID)) + "\n"
	msg += "名称: " + badge.Icon + " " + badge.Name + "\n"
	msg += "描述: " + badge.Description + "\n"
	msg += "类型: " + badge.Type + "\n"
	msg += "获取条件: " + badge.Condition + "\n"
	msg += "状态: " + func() string { if badge.IsEnabled { return "启用" } else { return "禁用" } }() + "\n"
	msg += "================================\n"
	
	SendTextReply(robot, event, msg)
}

// grantBadge 发放徽章功能（管理员命令）
func (p *BadgePlugin) grantBadge(robot plugin.Robot, event *onebot.Event, userID string, badgeID string) {
	// TODO: 这里应该检查用户是否为管理员
	
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 检查徽章是否存在且启用
	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM badge WHERE id = ?", badgeID).Scan(&isEnabled)
	if err != nil {
		SendTextReply(robot, event, "❌ 徽章不存在")
		return
	}
	
	if !isEnabled {
		SendTextReply(robot, event, "❌ 该徽章已被禁用")
		return
	}
	
	// 检查用户是否已获得该徽章
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM user_badge WHERE user_id = ? AND badge_id = ?", userID, badgeID).Scan(&count)
	if err != nil {
		log.Printf("查询用户徽章失败: %v\n", err)
		SendTextReply(robot, event, "❌ 发放失败，请稍后重试")
		return
	}
	
	if count > 0 {
		SendTextReply(robot, event, "❌ 该用户已获得此徽章")
		return
	}
	
	// 开始事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf("开启事务失败: %v\n", err)
		SendTextReply(robot, event, "❌ 发放失败，请稍后重试")
		return
	}
	
	// 发放徽章
	_, err = tx.Exec("INSERT INTO user_badge (user_id, badge_id, grant_time) VALUES (?, ?, CURRENT_TIMESTAMP)", userID, badgeID)
	if err != nil {
		tx.Rollback()
		log.Printf("发放徽章失败: %v\n", err)
		SendTextReply(robot, event, "❌ 发放失败，请稍后重试")
		return
	}
	
	// 记录发放日志
	_, err = tx.Exec("INSERT INTO badge_grant_log (user_id, badge_id, operator, reason) VALUES (?, ?, ?, ?)", 
		userID, badgeID, "admin", "管理员手动发放")
	if err != nil {
		tx.Rollback()
		log.Printf("记录发放日志失败: %v\n", err)
		SendTextReply(robot, event, "❌ 发放失败，请稍后重试")
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf("提交事务失败: %v\n", err)
		SendTextReply(robot, event, "❌ 发放失败，请稍后重试")
		return
	}
	
	SendTextReply(robot, event, "✅ 徽章发放成功")
}

// removeBadge 移除徽章功能（管理员命令）
func (p *BadgePlugin) removeBadge(robot plugin.Robot, event *onebot.Event, userID string, badgeID string) {
	// TODO: 这里应该检查用户是否为管理员
	
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 检查用户是否持有该徽章
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM user_badge WHERE user_id = ? AND badge_id = ? AND is_active = TRUE", userID, badgeID).Scan(&count)
	if err != nil {
		log.Printf("查询用户徽章失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}
	
	if count == 0 {
		SendTextReply(robot, event, "❌ 该用户未获得此徽章")
		return
	}
	
	// 移除徽章
	_, err = GlobalDB.Exec("UPDATE user_badge SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND badge_id = ?", userID, badgeID)
	if err != nil {
		log.Printf("移除徽章失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}
	
	SendTextReply(robot, event, "✅ 徽章移除成功")
}

// enableSystem 开启徽章系统
func (p *BadgePlugin) enableSystem(robot plugin.Robot, event *onebot.Event) {
	// TODO: 这里应该检查用户是否为管理员
	
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	_, err := GlobalDB.Exec("UPDATE badge_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("开启徽章系统失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}
	
	SendTextReply(robot, event, "✅ 徽章系统已开启")
}

// disableSystem 关闭徽章系统
func (p *BadgePlugin) disableSystem(robot plugin.Robot, event *onebot.Event) {
	// TODO: 这里应该检查用户是否为管理员
	
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	_, err := GlobalDB.Exec("UPDATE badge_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("关闭徽章系统失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}
	
	SendTextReply(robot, event, "✅ 徽章系统已关闭")
}

// 全局徽章插件实例
var globalBadgePlugin *BadgePlugin

// GetBadgePluginInstance 获取徽章插件实例
func GetBadgePluginInstance() *BadgePlugin {
	if globalBadgePlugin == nil {
		globalBadgePlugin = NewBadgePlugin()
	}
	return globalBadgePlugin
}

// GrantBadgeToUser 外部调用接口：给用户发放徽章
func (p *BadgePlugin) GrantBadgeToUser(userID string, badgeName string, operator string, reason string) error {
	if GlobalDB == nil {
		return nil
	}
	
	// 查找徽章ID
	var badgeID uint
	err := GlobalDB.QueryRow("SELECT id FROM badge WHERE name = ? AND is_enabled = TRUE", badgeName).Scan(&badgeID)
	if err != nil {
		return err
	}
	
	// 检查用户是否已获得该徽章
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM user_badge WHERE user_id = ? AND badge_id = ? AND is_active = TRUE", userID, badgeID).Scan(&count)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return nil // 用户已获得，不需要重复发放
	}
	
	// 开始事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		return err
	}
	
	// 发放徽章
	_, err = tx.Exec("INSERT INTO user_badge (user_id, badge_id, grant_time) VALUES (?, ?, CURRENT_TIMESTAMP)", userID, badgeID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// 记录发放日志
	_, err = tx.Exec("INSERT INTO badge_grant_log (user_id, badge_id, operator, reason) VALUES (?, ?, ?, ?)", 
		userID, badgeID, operator, reason)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// 提交事务
	return tx.Commit()
}

// GetUserBadges 获取用户的徽章列表
func (p *BadgePlugin) GetUserBadges(userID string) ([]struct {
	ID        uint      `json:"id"`
	BadgeID   uint      `json:"badge_id"`
	BadgeName string    `json:"badge_name"`
	Icon      string    `json:"icon"`
	AcquiredAt time.Time `json:"acquired_at"`
}, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	
	rows, err := GlobalDB.Query(`
		SELECT ub.id, ub.badge_id, b.name, b.icon, ub.grant_time 
		FROM user_badge ub 
		JOIN badge b ON ub.badge_id = b.id 
		WHERE ub.user_id = ? AND ub.is_active = TRUE AND b.is_enabled = TRUE
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var userBadges []struct {
		ID        uint      `json:"id"`
		BadgeID   uint      `json:"badge_id"`
		BadgeName string    `json:"badge_name"`
		Icon      string    `json:"icon"`
		AcquiredAt time.Time `json:"acquired_at"`
	}
	
	for rows.Next() {
		var ub struct {
			ID        uint      `json:"id"`
			BadgeID   uint      `json:"badge_id"`
			BadgeName string    `json:"badge_name"`
			Icon      string    `json:"icon"`
			AcquiredAt time.Time `json:"acquired_at"`
		}
		err := rows.Scan(&ub.ID, &ub.BadgeID, &ub.BadgeName, &ub.Icon, &ub.AcquiredAt)
		if err != nil {
			return nil, err
		}
		userBadges = append(userBadges, ub)
	}
	
	return userBadges, nil
}

// GetBadgeByName 根据名称获取徽章信息
func (p *BadgePlugin) GetBadgeByName(name string) (*Badge, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	
	var badge Badge
	err := GlobalDB.QueryRow("SELECT id, name, description, icon, type, condition, is_enabled, created_at, updated_at FROM badge WHERE name = ?", name).Scan(
		&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition, &badge.IsEnabled, &badge.CreatedAt, &badge.UpdatedAt
	)
	if err != nil {
		return nil, err
	}
	
	return &badge, nil
}