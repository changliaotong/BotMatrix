package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strconv"
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
	Type        string    `gorm:"size:20" json:"type"`       // system, achievement, event
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
	ID        uint      `gorm:"primaryKey" json:"id"`
	IsEnabled bool      `gorm:"default:true" json:"is_enabled"`
	UpdateAt  time.Time `json:"update_at"`
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
	return common.T("", "badge_plugin_desc|徽章系统插件，用于管理和展示用户的荣誉徽章")
}

func (p *BadgePlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *BadgePlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "grant_badge",
			Description: common.T("", "badge_skill_grant_desc|给用户发放指定徽章"),
			Usage:       common.T("", "badge_skill_grant_usage|grant_badge [user_id] [badge_id]"),
			Params: map[string]string{
				"user_id":  common.T("", "badge_param_user_id|用户ID"),
				"badge_id": common.T("", "badge_param_badge_id|徽章ID"),
			},
		},
		{
			Name:        "remove_badge",
			Description: common.T("", "badge_skill_remove_desc|移除用户持有的指定徽章"),
			Usage:       common.T("", "badge_skill_remove_usage|remove_badge [user_id] [badge_id]"),
			Params: map[string]string{
				"user_id":  common.T("", "badge_param_user_id|用户ID"),
				"badge_id": common.T("", "badge_param_badge_id|徽章ID"),
			},
		},
		{
			Name:        "get_user_badges",
			Description: common.T("", "badge_skill_my_desc|获取用户持有的所有徽章"),
			Usage:       common.T("", "badge_skill_my_usage|get_user_badges [user_id]"),
			Params: map[string]string{
				"user_id": common.T("", "badge_param_user_id|用户ID"),
			},
		},
		{
			Name:        "list_badges",
			Description: common.T("", "badge_skill_list_desc|列出系统中所有可用的徽章"),
			Usage:       common.T("", "badge_skill_list_usage|list_badges"),
			Params:      map[string]string{},
		},
		{
			Name:        "badge_detail",
			Description: common.T("", "badge_skill_detail_desc|获取指定徽章的详细信息"),
			Usage:       common.T("", "badge_skill_detail_usage|badge_detail [badge_id]"),
			Params: map[string]string{
				"badge_id": common.T("", "badge_param_badge_id|徽章ID"),
			},
		},
		{
			Name:        "enable_badge_system",
			Description: common.T("", "badge_skill_enable_desc|开启徽章系统"),
			Usage:       common.T("", "badge_skill_enable_usage|enable_badge_system"),
			Params:      map[string]string{},
		},
		{
			Name:        "disable_badge_system",
			Description: common.T("", "badge_skill_disable_desc|关闭徽章系统"),
			Usage:       common.T("", "badge_skill_disable_usage|disable_badge_system"),
			Params:      map[string]string{},
		},
	}
}

func (p *BadgePlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "badge_plugin_loaded|徽章系统插件已加载"))

	// 初始化数据库
	p.initDatabase()

	// 初始化默认徽章
	p.initDefaultBadges()

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "badge_cmd_my_badges|我的徽章"), event.RawMessage); match {
			msg, _ := p.doMyBadges(fmt.Sprintf("%d", event.UserID))
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 查看徽章
		if match, _ := p.cmdParser.MatchCommand(common.T("", "badge_cmd_list_badges|查看徽章"), event.RawMessage); match {
			msg, _ := p.doListBadges()
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 查看徽章详情
		if match, params := p.cmdParser.MatchRegex(common.T("", "badge_cmd_detail_regex|^徽章详情\\s+(\\d+)$"), event.RawMessage); match && len(params) > 1 {
			badgeID := params[1]
			msg, _ := p.doBadgeDetail(badgeID)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：发放徽章
		if match, params := p.cmdParser.MatchRegex(common.T("", "badge_cmd_grant_regex|^发放徽章\\s+(\\d+)\\s+(\\d+)$"), event.RawMessage); match && len(params) > 2 {
			userID := params[1]
			badgeID := params[2]

			// 权限检查
			isAdmin := isSuperAdmin(GlobalDB, event.GroupID, event.UserID)
			if !isAdmin && event.MessageType == "group" {
				isAdmin = isGroupAdmin(GlobalDB, event.GroupID, event.UserID)
			}
			if !isAdmin {
				p.sendMessage(robot, event, common.T("", "badge_admin_only_grant|抱歉，您没有权限执行此操作。"))
				return nil
			}

			msg, _ := p.doGrantBadge(userID, badgeID, "admin", common.T("", "badge_grant_reason_admin|管理员手动发放"))
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：移除徽章
		if match, params := p.cmdParser.MatchRegex(common.T("", "badge_cmd_remove_regex|^移除徽章\\s+(\\d+)\\s+(\\d+)$"), event.RawMessage); match && len(params) > 2 {
			userID := params[1]
			badgeID := params[2]

			// 权限检查
			isAdmin := isSuperAdmin(GlobalDB, event.GroupID, event.UserID)
			if !isAdmin && event.MessageType == "group" {
				isAdmin = isGroupAdmin(GlobalDB, event.GroupID, event.UserID)
			}
			if !isAdmin {
				p.sendMessage(robot, event, common.T("", "badge_admin_only_remove|抱歉，您没有权限执行此操作。"))
				return nil
			}

			msg, _ := p.doRemoveBadge(userID, badgeID)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：开启徽章系统
		if match, _ := p.cmdParser.MatchCommand(common.T("", "badge_cmd_enable_system|开启徽章系统"), event.RawMessage); match {
			// 权限检查
			isAdmin := isSuperAdmin(GlobalDB, event.GroupID, event.UserID)
			if !isAdmin && event.MessageType == "group" {
				isAdmin = isGroupAdmin(GlobalDB, event.GroupID, event.UserID)
			}
			if !isAdmin {
				p.sendMessage(robot, event, common.T("", "badge_admin_only_enable|抱歉，您没有权限执行此操作。"))
				return nil
			}

			msg, _ := p.doEnableSystem()
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：关闭徽章系统
		if match, _ := p.cmdParser.MatchCommand(common.T("", "badge_cmd_disable_system|关闭徽章系统"), event.RawMessage); match {
			// 权限检查
			isAdmin := isSuperAdmin(GlobalDB, event.GroupID, event.UserID)
			if !isAdmin && event.MessageType == "group" {
				isAdmin = isGroupAdmin(GlobalDB, event.GroupID, event.UserID)
			}
			if !isAdmin {
				p.sendMessage(robot, event, common.T("", "badge_admin_only_disable|抱歉，您没有权限执行此操作。"))
				return nil
			}

			msg, _ := p.doDisableSystem()
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})
}

func (p *BadgePlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userID := ""
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if uid, ok := params["user_id"]; ok {
		userID = uid
	}

	badgeID := params["badge_id"]

	switch skillName {
	case "grant_badge":
		if userID == "" || badgeID == "" {
			return common.T("", "badge_missing_params|缺少必要参数"), nil
		}
		return p.doGrantBadge(userID, badgeID, "system", "skill_call")
	case "remove_badge":
		if userID == "" || badgeID == "" {
			return common.T("", "badge_missing_params|缺少必要参数"), nil
		}
		return p.doRemoveBadge(userID, badgeID)
	case "get_user_badges":
		if userID == "" {
			return common.T("", "badge_missing_params|缺少必要参数"), nil
		}
		return p.doMyBadges(userID)
	case "list_badges":
		return p.doListBadges()
	case "badge_detail":
		if badgeID == "" {
			return common.T("", "badge_missing_params|缺少必要参数"), nil
		}
		return p.doBadgeDetail(badgeID)
	case "enable_badge_system":
		return p.doEnableSystem()
	case "disable_badge_system":
		return p.doDisableSystem()
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// initDatabase 初始化数据库
func (p *BadgePlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println(common.T("", "badge_db_init_warn|全局数据库未初始化，徽章系统部分功能可能受限"))
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
		log.Printf(common.T("", "badge_create_table_failed|创建徽章表失败: %v"), err)
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
		log.Printf(common.T("", "badge_create_user_badge_table_failed|创建用户徽章表失败: %v"), err)
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
		log.Printf(common.T("", "badge_create_grant_log_table_failed|创建徽章发放日志表失败: %v"), err)
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
		log.Printf(common.T("", "badge_create_config_table_failed|创建徽章系统配置表失败: %v"), err)
		return
	}

	// 初始化配置
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM badge_config").Scan(&count)
	if err != nil {
		log.Printf(common.T("", "badge_query_config_failed|查询徽章系统配置失败: %v"), err)
		return
	}

	if count == 0 {
		_, err = GlobalDB.Exec("INSERT INTO badge_config (is_enabled) VALUES (TRUE)")
		if err != nil {
			log.Printf(common.T("", "badge_init_config_failed|初始化徽章系统配置失败: %v"), err)
			return
		}
	}

	log.Println(common.T("", "badge_db_init_done|徽章系统数据库初始化完成"))
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
		log.Printf(common.T("", "badge_query_count_failed|查询徽章数量失败: %v"), err)
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
			log.Printf(common.T("", "badge_init_default_failed|初始化默认徽章 [%s] 失败: %v"), badge.Name, err)
		}
	}

	log.Println(common.T("", "badge_default_init_done|默认徽章初始化完成"))
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
		log.Printf(common.T("", "badge_query_config_failed|查询徽章系统配置失败: %v"), err)
		return true
	}

	return isEnabled
}

// doMyBadges 我的徽章逻辑
func (p *BadgePlugin) doMyBadges(userID string) (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	// 查询用户的徽章
	rows, err := GlobalDB.Query(`
		SELECT b.id, b.name, b.description, b.icon, ub.grant_time 
		FROM badge b 
		JOIN user_badge ub ON b.id = ub.badge_id 
		WHERE ub.user_id = ? AND ub.is_active = TRUE AND b.is_enabled = TRUE
	`, userID)
	if err != nil {
		log.Printf(common.T("", "badge_query_user_badges_failed|查询用户徽章失败: %v"), err)
		return common.T("", "badge_query_error|查询过程中出现错误，请稍后再试"), err
	}
	defer rows.Close()

	var badges []Badge
	var grantTimes []time.Time

	for rows.Next() {
		var badge Badge
		var grantTime time.Time
		err := rows.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &grantTime)
		if err != nil {
			log.Printf(common.T("", "badge_scan_user_badges_failed|扫描用户徽章数据失败: %v"), err)
			continue
		}
		badges = append(badges, badge)
		grantTimes = append(grantTimes, grantTime)
	}

	if len(badges) == 0 {
		return common.T("", "badge_no_badges|你目前还没有获得任何徽章，继续加油哦！"), nil
	}

	// 构建回复消息
	msg := common.T("", "badge_my_title|🎖️ 我的徽章库") + "\n"
	msg += "================================\n"

	for i, badge := range badges {
		msg += badge.Icon + " " + badge.Name + "\n"
		msg += "   " + badge.Description + "\n"
		msg += "   " + common.T("", "badge_get_time|获得时间") + ": " + grantTimes[i].Format("2006-01-02") + "\n"
	}

	msg += "================================\n"
	msg += common.T("", "badge_footer_list|使用 [查看徽章] 了解更多，[徽章详情 ID] 查看详细。")

	return msg, nil
}

// doListBadges 查看所有徽章逻辑
func (p *BadgePlugin) doListBadges() (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	// 查询所有启用的徽章
	rows, err := GlobalDB.Query("SELECT id, name, description, icon, type, condition FROM badge WHERE is_enabled = TRUE")
	if err != nil {
		log.Printf(common.T("", "badge_query_list_failed|查询徽章列表失败: %v"), err)
		return common.T("", "badge_query_error|查询过程中出现错误，请稍后再试"), err
	}
	defer rows.Close()

	var badges []Badge

	for rows.Next() {
		var badge Badge
		err := rows.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition)
		if err != nil {
			log.Printf(common.T("", "badge_scan_list_failed|扫描徽章列表数据失败: %v"), err)
			continue
		}
		badges = append(badges, badge)
	}

	if len(badges) == 0 {
		return common.T("", "badge_list_empty|系统中目前没有任何可用的徽章。"), nil
	}

	// 构建回复消息
	msg := common.T("", "badge_list_title|📜 全服徽章一览") + "\n"
	msg += "================================\n"

	for _, badge := range badges {
		msg += badge.Icon + " " + badge.Name + "\n"
		msg += "   " + common.T("", "badge_id|徽章ID") + ": " + strconv.Itoa(int(badge.ID)) + "\n"
		msg += "   " + badge.Description + "\n"
		msg += "   " + common.T("", "badge_type|徽章类型") + ": " + badge.Type + "\n"
		msg += "   " + common.T("", "badge_condition|获取条件") + ": " + badge.Condition + "\n"
		msg += "\n"
	}

	msg += "================================\n"
	msg += common.T("", "badge_footer_detail|提示：输入 [徽章详情 ID] 查看具体获取方式。")

	return msg, nil
}

// doBadgeDetail 徽章详情逻辑
func (p *BadgePlugin) doBadgeDetail(badgeID string) (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	// 查询徽章详情
	var badge Badge
	row := GlobalDB.QueryRow("SELECT id, name, description, icon, type, condition, is_enabled FROM badge WHERE id = ?", badgeID)
	err := row.Scan(&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition, &badge.IsEnabled)
	if err != nil {
		return common.T("", "badge_not_found|抱歉，未找到该徽章或用户未持有。"), nil
	}

	if !badge.IsEnabled {
		return common.T("", "badge_disabled|该徽章目前已被系统禁用。"), nil
	}

	// 构建回复消息
	msg := common.T("", "badge_detail_title|🔍 徽章详细资料") + "\n"
	msg += "================================\n"
	msg += common.T("", "badge_id|徽章ID") + ": " + strconv.Itoa(int(badge.ID)) + "\n"
	msg += common.T("", "badge_name|徽章名称") + ": " + badge.Icon + " " + badge.Name + "\n"
	msg += common.T("", "badge_desc|徽章描述") + ": " + badge.Description + "\n"
	msg += common.T("", "badge_type|徽章类型") + ": " + badge.Type + "\n"
	msg += common.T("", "badge_condition|获取条件") + ": " + badge.Condition + "\n"
	msg += common.T("", "badge_status|当前状态") + ": " + func() string {
		if badge.IsEnabled {
			return common.T("", "badge_enabled|已启用")
		} else {
			return common.T("", "badge_disabled_text|已禁用")
		}
	}() + "\n"
	msg += "================================\n"

	return msg, nil
}

// doGrantBadge 发放徽章逻辑
func (p *BadgePlugin) doGrantBadge(userID string, badgeID string, operator string, reason string) (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	// 检查徽章是否存在且启用
	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM badge WHERE id = ?", badgeID).Scan(&isEnabled)
	if err != nil {
		return common.T("", "badge_not_found|抱歉，未找到该徽章或用户未持有。"), nil
	}

	if !isEnabled {
		return common.T("", "badge_disabled|该徽章目前已被系统禁用。"), nil
	}

	// 检查用户是否已获得该徽章
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM user_badge WHERE user_id = ? AND badge_id = ?", userID, badgeID).Scan(&count)
	if err != nil {
		log.Printf(common.T("", "badge_query_user_badges_failed|查询用户徽章失败: %v"), err)
		return common.T("", "badge_grant_failed|发放徽章失败，请稍后再试。"), err
	}

	if count > 0 {
		return common.T("", "badge_already_have|该用户已经拥有这个徽章了。"), nil
	}

	// 开始事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf(common.T("", "badge_tx_begin_failed|启动发放事务失败: %v"), err)
		return common.T("", "badge_grant_failed|发放徽章失败，请稍后再试。"), err
	}

	// 发放徽章
	_, err = tx.Exec("INSERT INTO user_badge (user_id, badge_id, grant_time) VALUES (?, ?, CURRENT_TIMESTAMP)", userID, badgeID)
	if err != nil {
		tx.Rollback()
		log.Printf(common.T("", "badge_grant_failed_log|发放徽章记录插入失败: %v"), err)
		return common.T("", "badge_grant_failed|发放徽章失败，请稍后再试。"), err
	}

	// 记录发放日志
	_, err = tx.Exec("INSERT INTO badge_grant_log (user_id, badge_id, operator, reason) VALUES (?, ?, ?, ?)",
		userID, badgeID, operator, reason)
	if err != nil {
		tx.Rollback()
		log.Printf(common.T("", "badge_log_failed|记录徽章发放日志失败: %v"), err)
		return common.T("", "badge_grant_failed|发放徽章失败，请稍后再试。"), err
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf(common.T("", "badge_tx_commit_failed|提交发放事务失败: %v"), err)
		return common.T("", "badge_grant_failed|发放徽章失败，请稍后再试。"), err
	}

	return common.T("", "badge_grant_success|🎉 徽章发放成功！"), nil
}

// doRemoveBadge 移除徽章逻辑
func (p *BadgePlugin) doRemoveBadge(userID string, badgeID string) (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	// 检查用户是否持有该徽章
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM user_badge WHERE user_id = ? AND badge_id = ? AND is_active = TRUE", userID, badgeID).Scan(&count)
	if err != nil {
		log.Printf(common.T("", "badge_query_user_badges_failed|查询用户徽章失败: %v"), err)
		return common.T("", "badge_op_failed|操作失败，请联系管理员。"), err
	}

	if count == 0 {
		return common.T("", "badge_not_found|抱歉，未找到该徽章或用户未持有。"), nil
	}

	// 移除徽章
	_, err = GlobalDB.Exec("UPDATE user_badge SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND badge_id = ?", userID, badgeID)
	if err != nil {
		log.Printf(common.T("", "badge_remove_failed_log|移除用户徽章失败: %v"), err)
		return common.T("", "badge_remove_failed|移除徽章失败，请稍后再试。"), err
	}

	return common.T("", "badge_remove_success|✅ 徽章已成功移除。"), nil
}

// doEnableSystem 开启系统逻辑
func (p *BadgePlugin) doEnableSystem() (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	_, err := GlobalDB.Exec("UPDATE badge_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf(common.T("", "badge_enable_failed_log|开启徽章系统失败: %v"), err)
		return common.T("", "badge_op_failed|操作失败，请联系管理员。"), err
	}

	return common.T("", "badge_system_enabled|✅ 徽章系统已成功开启。"), nil
}

// doDisableSystem 关闭系统逻辑
func (p *BadgePlugin) doDisableSystem() (string, error) {
	if GlobalDB == nil {
		return common.T("", "badge_db_error|数据库连接异常，请联系管理员"), nil
	}

	_, err := GlobalDB.Exec("UPDATE badge_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf(common.T("", "badge_disable_failed_log|关闭徽章系统失败: %v"), err)
		return common.T("", "badge_op_failed|操作失败，请联系管理员。"), err
	}

	return common.T("", "badge_system_disabled|✅ 徽章系统已成功关闭。"), nil
}

// sendMessage 发送消息
func (p *BadgePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "badge_send_failed_log|发送徽章回复消息失败: %v"), err)
	}
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

	_, err = p.doGrantBadge(userID, strconv.Itoa(int(badgeID)), operator, reason)
	return err
}

// GetUserBadges 获取用户的徽章列表
func (p *BadgePlugin) GetUserBadges(userID string) ([]struct {
	ID         uint      `json:"id"`
	BadgeID    uint      `json:"badge_id"`
	BadgeName  string    `json:"badge_name"`
	Icon       string    `json:"icon"`
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
		ID         uint      `json:"id"`
		BadgeID    uint      `json:"badge_id"`
		BadgeName  string    `json:"badge_name"`
		Icon       string    `json:"icon"`
		AcquiredAt time.Time `json:"acquired_at"`
	}

	for rows.Next() {
		var ub struct {
			ID         uint      `json:"id"`
			BadgeID    uint      `json:"badge_id"`
			BadgeName  string    `json:"badge_name"`
			Icon       string    `json:"icon"`
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
		&badge.ID, &badge.Name, &badge.Description, &badge.Icon, &badge.Type, &badge.Condition, &badge.IsEnabled, &badge.CreatedAt, &badge.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &badge, nil
}
