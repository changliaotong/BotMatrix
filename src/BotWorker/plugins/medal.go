package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
	"time"
)

// MedalPlugin 勋章系统插件
type MedalPlugin struct {
	cmdParser *CommandParser
}

// Medal 勋章定义
type Medal struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Icon        string    `gorm:"size:100" json:"icon"`
	Type        string    `gorm:"size:20" json:"type"`       // honor, achievement, rank
	Condition   string    `gorm:"size:255" json:"condition"` // 获取条件描述
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserMedal 用户持有勋章记录
type UserMedal struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	MedalID   uint      `json:"medal_id"`
	GrantTime time.Time `json:"grant_time"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	Level     int       `gorm:"default:1" json:"level"`    // 勋章等级
	Progress  int       `gorm:"default:0" json:"progress"` // 升级进度
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MedalGrantLog 勋章发放日志
type MedalGrantLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	MedalID   uint      `json:"medal_id"`
	Operator  string    `gorm:"size:20" json:"operator"` // system, admin, event
	Reason    string    `gorm:"size:255" json:"reason"`
	Level     int       `json:"level"`
	CreatedAt time.Time `json:"created_at"`
}

// MedalConfig 勋章系统配置
type MedalConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IsEnabled bool      `gorm:"default:true" json:"is_enabled"`
	UpdateAt  time.Time `json:"update_at"`
}

// NewMedalPlugin 创建勋章系统插件实例
func NewMedalPlugin() *MedalPlugin {
	return &MedalPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MedalPlugin) Name() string {
	return "medal"
}

func (p *MedalPlugin) Description() string {
	return "勋章系统插件，提供勋章发放、查询和管理功能"
}

func (p *MedalPlugin) Version() string {
	return "1.0.0"
}

func (p *MedalPlugin) Init(robot plugin.Robot) {
	log.Println("加载勋章系统插件")

	// 初始化数据库
	p.initDatabase()

	// 初始化默认勋章
	p.initDefaultMedals()

	// 处理勋章系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		// 我的勋章
		if match, _ := p.cmdParser.MatchCommand("我的勋章", event.RawMessage); match {
			p.myMedals(robot, event)
			return nil
		}

		// 查看勋章
		if match, _ := p.cmdParser.MatchCommand("查看勋章", event.RawMessage); match {
			p.listMedals(robot, event)
			return nil
		}

		// 查看勋章详情
		if match, params := p.cmdParser.MatchCommandWithParams("勋章详情", `(\S+)`, event.RawMessage); match && len(params) == 1 {
			medalName := params[0]
			p.medalDetail(robot, event, medalName)
			return nil
		}

		// 管理员命令：发放勋章
		if match, params := p.cmdParser.MatchCommandWithParams("发放勋章", `(\S+)\s+(\S+)`, event.RawMessage); match && len(params) == 2 {
			userID := params[0]
			medalName := params[1]
			p.grantMedal(robot, event, userID, medalName)
			return nil
		}

		// 管理员命令：移除勋章
		if match, params := p.cmdParser.MatchCommandWithParams("移除勋章", `(\S+)\s+(\S+)`, event.RawMessage); match && len(params) == 2 {
			userID := params[0]
			medalName := params[1]
			p.removeMedal(robot, event, userID, medalName)
			return nil
		}

		// 管理员命令：升级勋章
		if match, params := p.cmdParser.MatchCommandWithParams("升级勋章", `(\S+)\s+(\S+)\s+(\d+)`, event.RawMessage); match && len(params) == 3 {
			userID := params[0]
			medalName := params[1]
			level, _ := p.cmdParser.ParseInt(params[2])
			p.upgradeMedal(robot, event, userID, medalName, level)
			return nil
		}

		// 管理员命令：开启勋章系统
		if match, _ := p.cmdParser.MatchCommand("开启勋章系统", event.RawMessage); match {
			p.enableSystem(robot, event)
			return nil
		}

		// 管理员命令：关闭勋章系统
		if match, _ := p.cmdParser.MatchCommand("关闭勋章系统", event.RawMessage); match {
			p.disableSystem(robot, event)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
func (p *MedalPlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，勋章系统将无法正常工作")
		return
	}

	// 创建勋章表
	createMedalTable := `
	CREATE TABLE IF NOT EXISTS medal (
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
	_, err := GlobalDB.Exec(createMedalTable)
	if err != nil {
		log.Printf("创建勋章表失败: %v\n", err)
		return
	}

	// 创建用户勋章表
	createUserMedalTable := `
	CREATE TABLE IF NOT EXISTS user_medal (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		medal_id INT NOT NULL REFERENCES medal(id) ON DELETE CASCADE,
		grant_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		level INT NOT NULL DEFAULT 1,
		progress INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, medal_id)
	)
	`
	_, err = GlobalDB.Exec(createUserMedalTable)
	if err != nil {
		log.Printf("创建用户勋章表失败: %v\n", err)
		return
	}

	// 创建勋章发放日志表
	createMedalGrantLogTable := `
	CREATE TABLE IF NOT EXISTS medal_grant_log (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		medal_id INT NOT NULL REFERENCES medal(id) ON DELETE CASCADE,
		operator VARCHAR(20) NOT NULL,
		reason VARCHAR(255) NOT NULL,
		level INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createMedalGrantLogTable)
	if err != nil {
		log.Printf("创建勋章发放日志表失败: %v\n", err)
		return
	}

	// 创建勋章系统配置表
	createMedalConfigTable := `
	CREATE TABLE IF NOT EXISTS medal_config (
		id SERIAL PRIMARY KEY,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createMedalConfigTable)
	if err != nil {
		log.Printf("创建勋章系统配置表失败: %v\n", err)
		return
	}

	// 初始化默认配置
	insertDefaultConfig := `
	INSERT INTO medal_config (is_enabled) 
	SELECT TRUE 
	WHERE NOT EXISTS (SELECT 1 FROM medal_config)
	`
	_, err = GlobalDB.Exec(insertDefaultConfig)
	if err != nil {
		log.Printf("初始化勋章系统配置失败: %v\n", err)
		return
	}
}

// initDefaultMedals 初始化默认勋章
func (p *MedalPlugin) initDefaultMedals() {
	if GlobalDB == nil {
		return
	}

	// 检查是否已有勋章
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM medal").Scan(&count)
	if err != nil {
		log.Printf("查询勋章数量失败: %v\n", err)
		return
	}

	if count > 0 {
		return // 已有勋章，跳过初始化
	}

	// 默认勋章列表
	defaultMedals := []Medal{
		{
			Name:        "新人勋章",
			Description: "欢迎加入的新成员",
			Icon:        "🏅",
			Type:        "honor",
			Condition:   "新用户注册",
			IsEnabled:   true,
		},
		{
			Name:        "活跃用户",
			Description: "积极参与群聊的用户",
			Icon:        "⭐",
			Type:        "achievement",
			Condition:   "发言超过100次",
			IsEnabled:   true,
		},
		{
			Name:        "贡献者",
			Description: "为群聊做出贡献的用户",
			Icon:        "💎",
			Type:        "rank",
			Condition:   "帮助他人解决问题",
			IsEnabled:   true,
		},
	}

	// 插入默认勋章
	for _, medal := range defaultMedals {
		_, err := GlobalDB.Exec(
			"INSERT INTO medal (name, description, icon, type, condition, is_enabled) VALUES ($1, $2, $3, $4, $5, $6)",
			medal.Name, medal.Description, medal.Icon, medal.Type, medal.Condition, medal.IsEnabled,
		)
		if err != nil {
			log.Printf("插入默认勋章失败: %v\n", err)
		}
	}
}

// isSystemEnabled 检查系统是否开启
func (p *MedalPlugin) isSystemEnabled() bool {
	if GlobalDB == nil {
		return true // 默认开启
	}

	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM medal_config LIMIT 1").Scan(&isEnabled)
	if err != nil {
		log.Printf("查询勋章系统配置失败: %v\n", err)
		return true // 默认开启
	}

	return isEnabled
}

// enableSystem 开启系统
func (p *MedalPlugin) enableSystem(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法操作")
		return
	}

	_, err := GlobalDB.Exec("UPDATE medal_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("开启勋章系统失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	p.sendMessage(robot, event, "勋章系统已开启")
}

// disableSystem 关闭系统
func (p *MedalPlugin) disableSystem(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法操作")
		return
	}

	_, err := GlobalDB.Exec("UPDATE medal_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("关闭勋章系统失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	p.sendMessage(robot, event, "勋章系统已关闭")
}

// myMedals 查看我的勋章
func (p *MedalPlugin) myMedals(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法查询")
		return
	}

	userID := fmt.Sprintf("%d", event.UserID)
	rows, err := GlobalDB.Query(`
		SELECT m.id, m.name, m.icon, m.type, um.level, um.progress 
		FROM user_medal um 
		JOIN medal m ON um.medal_id = m.id 
		WHERE um.user_id = $1 AND um.is_active = TRUE AND m.is_enabled = TRUE
		ORDER BY m.type, um.level DESC
	`, userID)
	if err != nil {
		log.Printf("查询用户勋章失败: %v\n", err)
		p.sendMessage(robot, event, "查询失败")
		return
	}
	defer rows.Close()

	var medals []string
	for rows.Next() {
		var id uint
		var name, icon, medalType string
		var level, progress int
		if err := rows.Scan(&id, &name, &icon, &medalType, &level, &progress); err != nil {
			log.Printf("扫描用户勋章数据失败: %v\n", err)
			continue
		}
		medals = append(medals, fmt.Sprintf("%s %s (等级 %d, 进度 %d)", icon, name, level, progress))
	}

	if len(medals) == 0 {
		p.sendMessage(robot, event, "您还没有获得任何勋章")
		return
	}

	message := "🏅 我的勋章\n" + strings.Join(medals, "\n")
	p.sendMessage(robot, event, message)
}

// listMedals 查看所有勋章
func (p *MedalPlugin) listMedals(robot plugin.Robot, event *onebot.Event) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法查询")
		return
	}

	rows, err := GlobalDB.Query(`
		SELECT id, name, icon, type, description 
		FROM medal 
		WHERE is_enabled = TRUE 
		ORDER BY type
	`)
	if err != nil {
		log.Printf("查询所有勋章失败: %v\n", err)
		p.sendMessage(robot, event, "查询失败")
		return
	}
	defer rows.Close()

	var medals []string
	for rows.Next() {
		var id uint
		var name, icon, medalType, description string
		if err := rows.Scan(&id, &name, &icon, &medalType, &description); err != nil {
			log.Printf("扫描勋章数据失败: %v\n", err)
			continue
		}
		medals = append(medals, fmt.Sprintf("%s %s (%s): %s", icon, name, medalType, description))
	}

	if len(medals) == 0 {
		p.sendMessage(robot, event, "当前没有可用的勋章")
		return
	}

	message := "🏅 所有勋章\n" + strings.Join(medals, "\n")
	p.sendMessage(robot, event, message)
}

// medalDetail 查看勋章详情
func (p *MedalPlugin) medalDetail(robot plugin.Robot, event *onebot.Event, medalName string) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法查询")
		return
	}

	var medal Medal
	err := GlobalDB.QueryRow(`
		SELECT id, name, description, icon, type, condition 
		FROM medal 
		WHERE name = $1 AND is_enabled = TRUE
	`, medalName).Scan(&medal.ID, &medal.Name, &medal.Description, &medal.Icon, &medal.Type, &medal.Condition)
	if err != nil {
		log.Printf("查询勋章详情失败: %v\n", err)
		p.sendMessage(robot, event, "勋章不存在或已关闭")
		return
	}

	// 查询用户是否拥有该勋章
	userID := fmt.Sprintf("%d", event.UserID)
	var hasMedal bool
	var level, progress int
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0, COALESCE(level, 0), COALESCE(progress, 0) FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medal.ID,
	).Scan(&hasMedal, &level, &progress)

	var userStatus string
	if hasMedal {
		userStatus = fmt.Sprintf("\n🔹 您已拥有该勋章 (等级 %d, 进度 %d)", level, progress)
	} else {
		userStatus = "\n🔹 您尚未获得该勋章"
	}

	message := fmt.Sprintf(
		"🏅 勋章详情\n"+
			"名称：%s %s\n"+
			"类型：%s\n"+
			"描述：%s\n"+
			"获取条件：%s\n"+
			"%s",
		medal.Icon, medal.Name, medal.Type, medal.Description, medal.Condition, userStatus,
	)
	p.sendMessage(robot, event, message)
}

// grantMedal 发放勋章
func (p *MedalPlugin) grantMedal(robot plugin.Robot, event *onebot.Event, userID string, medalName string) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法操作")
		return
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		log.Printf("查询勋章失败: %v\n", err)
		p.sendMessage(robot, event, "勋章不存在或已关闭")
		return
	}

	// 检查用户是否已拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		log.Printf("检查用户勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	if exists {
		p.sendMessage(robot, event, "该用户已拥有此勋章")
		return
	}

	// 发放勋章
	_, err = GlobalDB.Exec(
		"INSERT INTO user_medal (user_id, medal_id, grant_time) VALUES ($1, $2, CURRENT_TIMESTAMP)",
		userID, medalID,
	)
	if err != nil {
		log.Printf("发放勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, fmt.Sprintf("%d", event.UserID), "管理员发放", 1,
	)
	if err != nil {
		log.Printf("记录勋章发放日志失败: %v\n", err)
	}

	p.sendMessage(robot, event, fmt.Sprintf("成功为用户 %s 发放勋章 %s", userID, medalName))
}

// removeMedal 移除勋章
func (p *MedalPlugin) removeMedal(robot plugin.Robot, event *onebot.Event, userID string, medalName string) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法操作")
		return
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		log.Printf("查询勋章失败: %v\n", err)
		p.sendMessage(robot, event, "勋章不存在或已关闭")
		return
	}

	// 检查用户是否拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		log.Printf("检查用户勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	if !exists {
		p.sendMessage(robot, event, "该用户未拥有此勋章")
		return
	}

	// 移除勋章
	_, err = GlobalDB.Exec(
		"UPDATE user_medal SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND medal_id = $2",
		userID, medalID,
	)
	if err != nil {
		log.Printf("移除勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, fmt.Sprintf("%d", event.UserID), "管理员移除", 0,
	)
	if err != nil {
		log.Printf("记录勋章移除日志失败: %v\n", err)
	}

	p.sendMessage(robot, event, fmt.Sprintf("成功为用户 %s 移除勋章 %s", userID, medalName))
}

// upgradeMedal 升级勋章
func (p *MedalPlugin) upgradeMedal(robot plugin.Robot, event *onebot.Event, userID string, medalName string, level int) {
	if GlobalDB == nil {
		p.sendMessage(robot, event, "数据库未初始化，无法操作")
		return
	}

	if level <= 0 {
		p.sendMessage(robot, event, "等级必须大于0")
		return
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		log.Printf("查询勋章失败: %v\n", err)
		p.sendMessage(robot, event, "勋章不存在或已关闭")
		return
	}

	// 检查用户是否拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		log.Printf("检查用户勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	if !exists {
		p.sendMessage(robot, event, "该用户未拥有此勋章")
		return
	}

	// 升级勋章
	_, err = GlobalDB.Exec(
		"UPDATE user_medal SET level = $3, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND medal_id = $2",
		userID, medalID, level,
	)
	if err != nil {
		log.Printf("升级勋章失败: %v\n", err)
		p.sendMessage(robot, event, "操作失败")
		return
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, fmt.Sprintf("%d", event.UserID), "管理员升级", level,
	)
	if err != nil {
		log.Printf("记录勋章升级日志失败: %v\n", err)
	}

	p.sendMessage(robot, event, fmt.Sprintf("成功将用户 %s 的勋章 %s 升级到等级 %d", userID, medalName, level))
}

// sendMessage 发送消息
func (p *MedalPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if event.MessageType == "group" {
		robot.SendGroupMessage(event.GroupID, message)
	} else {
		robot.SendPrivateMessage(event.UserID, message)
	}
}
