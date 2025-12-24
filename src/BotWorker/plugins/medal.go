package plugins

import (
	"BotMatrix/common"
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

// GetSkills 实现 SkillCapable 接口
func (p *MedalPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "my_medals",
			Description: common.T("", "medal_skill_my_desc|查看我的勋章列表"),
		},
		{
			Name:        "list_medals",
			Description: common.T("", "medal_skill_list_desc|查看系统所有勋章"),
		},
		{
			Name:        "medal_detail",
			Description: common.T("", "medal_skill_detail_desc|查看指定勋章详情"),
			Usage:       "medal_detail name=勋章名称",
			Params: map[string]string{
				"name": "勋章名称",
			},
		},
		{
			Name:        "grant_medal",
			Description: common.T("", "medal_skill_grant_desc|发放勋章给用户"),
			Usage:       "grant_medal user_id=123456 name=勋章名称",
			Params: map[string]string{
				"user_id": "用户ID",
				"name":    "勋章名称",
			},
		},
		{
			Name:        "remove_medal",
			Description: common.T("", "medal_skill_remove_desc|移除用户的勋章"),
			Usage:       "remove_medal user_id=123456 name=勋章名称",
			Params: map[string]string{
				"user_id": "用户ID",
				"name":    "勋章名称",
			},
		},
		{
			Name:        "upgrade_medal",
			Description: common.T("", "medal_skill_upgrade_desc|升级用户的勋章等级"),
			Usage:       "upgrade_medal user_id=123456 name=勋章名称 level=2",
			Params: map[string]string{
				"user_id": "用户ID",
				"name":    "勋章名称",
				"level":   "等级",
			},
		},
		{
			Name:        "enable_medal_system",
			Description: common.T("", "medal_skill_enable_desc|开启勋章系统"),
		},
		{
			Name:        "disable_medal_system",
			Description: common.T("", "medal_skill_disable_desc|关闭勋章系统"),
		},
	}
}

// HandleSkill 实现 SkillCapable 接口
func (p *MedalPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) error {
	var userID string
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if params["user_id"] != "" {
		userID = params["user_id"]
	}

	switch skillName {
	case "my_medals":
		msg, err := p.doMyMedals(userID)
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "list_medals":
		msg, err := p.doListMedals()
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "medal_detail":
		name := params["name"]
		msg, err := p.doMedalDetail(userID, name)
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "grant_medal":
		targetUserID := params["user_id"]
		name := params["name"]
		msg, err := p.doGrantMedal(userID, targetUserID, name)
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "remove_medal":
		targetUserID := params["user_id"]
		name := params["name"]
		msg, err := p.doRemoveMedal(userID, targetUserID, name)
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "upgrade_medal":
		targetUserID := params["user_id"]
		name := params["name"]
		level, _ := p.cmdParser.ParseInt(params["level"])
		msg, err := p.doUpgradeMedal(userID, targetUserID, name, level)
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "enable_medal_system":
		msg, err := p.doEnableSystem()
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	case "disable_medal_system":
		msg, err := p.doDisableSystem()
		if err != nil {
			p.sendMessage(robot, event, err.Error())
			return err
		}
		p.sendMessage(robot, event, msg)
	}
	return nil
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
	return common.T("", "medal_plugin_desc|勋章系统插件，提供勋章发放、查询和管理功能")
}

func (p *MedalPlugin) Version() string {
	return "1.0.0"
}

func (p *MedalPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "medal_plugin_loaded|勋章系统插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return "", p.HandleSkill(robot, nil, skillName, params)
		})
	}

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

		userID := fmt.Sprintf("%d", event.UserID)

		// 我的勋章
		if match, _ := p.cmdParser.MatchCommand("我的勋章", event.RawMessage); match {
			msg, err := p.doMyMedals(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 查看勋章
		if match, _ := p.cmdParser.MatchCommand("查看勋章", event.RawMessage); match {
			msg, err := p.doListMedals()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 查看勋章详情
		if match, _, params := p.cmdParser.MatchCommandWithParams("勋章详情", `(\S+)`, event.RawMessage); match && len(params) == 1 {
			medalName := params[0]
			msg, err := p.doMedalDetail(userID, medalName)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：发放勋章
		if match, _, params := p.cmdParser.MatchCommandWithParams("发放勋章", `(\S+)\s+(\S+)`, event.RawMessage); match && len(params) == 2 {
			targetUserID := params[0]
			medalName := params[1]
			msg, err := p.doGrantMedal(userID, targetUserID, medalName)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：移除勋章
		if match, _, params := p.cmdParser.MatchCommandWithParams("移除勋章", `(\S+)\s+(\S+)`, event.RawMessage); match && len(params) == 2 {
			targetUserID := params[0]
			medalName := params[1]
			msg, err := p.doRemoveMedal(userID, targetUserID, medalName)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：升级勋章
		if match, _, params := p.cmdParser.MatchCommandWithParams("升级勋章", `(\S+)\s+(\S+)\s+(\d+)`, event.RawMessage); match && len(params) == 3 {
			targetUserID := params[0]
			medalName := params[1]
			level, _ := p.cmdParser.ParseInt(params[2])
			msg, err := p.doUpgradeMedal(userID, targetUserID, medalName, level)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：开启勋章系统
		if match, _ := p.cmdParser.MatchCommand("开启勋章系统", event.RawMessage); match {
			msg, err := p.doEnableSystem()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 管理员命令：关闭勋章系统
		if match, _ := p.cmdParser.MatchCommand("关闭勋章系统", event.RawMessage); match {
			msg, err := p.doDisableSystem()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
func (p *MedalPlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
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
		log.Printf(common.T("", "medal_db_init_failed|勋章系统：数据库初始化失败：%v"), err)
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
		log.Printf(common.T("", "medal_db_init_failed|勋章系统：数据库初始化失败：%v"), err)
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
		log.Printf(common.T("", "medal_db_init_failed|勋章系统：数据库初始化失败：%v"), err)
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
		log.Printf(common.T("", "medal_db_init_failed|勋章系统：数据库初始化失败：%v"), err)
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
		log.Printf(common.T("", "medal_init_default_failed|勋章系统：初始化默认数据失败：%v"), err)
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
		log.Printf(common.T("", "medal_init_default_failed|勋章系统：初始化默认数据失败：%v"), err)
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
			log.Printf(common.T("", "medal_init_default_failed|勋章系统：初始化默认数据失败：%v"), err)
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
		return true // 默认开启
	}

	return isEnabled
}

// doEnableSystem 开启系统
func (p *MedalPlugin) doEnableSystem() (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	_, err := GlobalDB.Exec("UPDATE medal_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	return common.T("", "medal_system_enabled|勋章系统已开启"), nil
}

// doDisableSystem 关闭系统
func (p *MedalPlugin) doDisableSystem() (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	_, err := GlobalDB.Exec("UPDATE medal_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	return common.T("", "medal_system_disabled_msg|勋章系统已关闭"), nil
}

// doMyMedals 查看我的勋章
func (p *MedalPlugin) doMyMedals(userID string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	rows, err := GlobalDB.Query(`
		SELECT m.id, m.name, m.icon, m.type, um.level, um.progress 
		FROM user_medal um 
		JOIN medal m ON um.medal_id = m.id 
		WHERE um.user_id = $1 AND um.is_active = TRUE AND m.is_enabled = TRUE
		ORDER BY m.type, um.level DESC
	`, userID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}
	defer rows.Close()

	var medals []string
	for rows.Next() {
		var id uint
		var name, icon, medalType string
		var level, progress int
		if err := rows.Scan(&id, &name, &icon, &medalType, &level, &progress); err != nil {
			continue
		}
		medals = append(medals, common.T("", "medal_my_item|%s %s (等级: %d, 进度: %d)", icon, name, level, progress))
	}

	if len(medals) == 0 {
		return common.T("", "medal_my_empty|你目前还没有获得任何勋章哦，加油！"), nil
	}

	message := common.T("", "medal_my_title|📜 我的勋章库") + "\n" + strings.Join(medals, "\n")
	return message, nil
}

// doListMedals 查看所有勋章
func (p *MedalPlugin) doListMedals() (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	rows, err := GlobalDB.Query(`
		SELECT id, name, icon, type, description 
		FROM medal 
		WHERE is_enabled = TRUE 
		ORDER BY type
	`)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}
	defer rows.Close()

	var medals []string
	for rows.Next() {
		var id uint
		var name, icon, medalType, description string
		if err := rows.Scan(&id, &name, &icon, &medalType, &description); err != nil {
			continue
		}
		medals = append(medals, common.T("", "medal_list_item|%s %s [%s]: %s", icon, name, medalType, description))
	}

	if len(medals) == 0 {
		return common.T("", "medal_list_empty|系统目前没有任何勋章。"), nil
	}

	message := common.T("", "medal_list_title|🏅 勋章列表") + "\n" + strings.Join(medals, "\n")
	return message, nil
}

// doMedalDetail 查看勋章详情
func (p *MedalPlugin) doMedalDetail(userID string, medalName string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	var medal Medal
	err := GlobalDB.QueryRow(`
		SELECT id, name, description, icon, type, condition 
		FROM medal 
		WHERE name = $1 AND is_enabled = TRUE
	`, medalName).Scan(&medal.ID, &medal.Name, &medal.Description, &medal.Icon, &medal.Type, &medal.Condition)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_not_found|勋章系统：未找到勋章“%s”"), medalName)
	}

	// 查询用户是否拥有该勋章
	var hasMedal bool
	var level, progress int
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0, COALESCE(level, 0), COALESCE(progress, 0) FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medal.ID,
	).Scan(&hasMedal, &level, &progress)

	var userStatus string
	if hasMedal {
		userStatus = common.T("", "medal_detail_has|【我的状态】：已拥有 (等级: %d, 进度: %d)", level, progress)
	} else {
		userStatus = common.T("", "medal_detail_not_has|【我的状态】：尚未获得")
	}

	message := common.T("", "medal_detail_title|🔍 勋章详情") + "\n" +
		common.T("", "medal_detail_name|【勋章名称】：%s %s", medal.Icon, medal.Name) + "\n" +
		common.T("", "medal_detail_type|【勋章类型】：%s", medal.Type) + "\n" +
		common.T("", "medal_detail_desc|【勋章描述】：%s", medal.Description) + "\n" +
		common.T("", "medal_detail_condition|【获取条件】：%s", medal.Condition) + "\n" +
		userStatus

	return message, nil
}

// doGrantMedal 发放勋章
func (p *MedalPlugin) doGrantMedal(operatorID string, userID string, medalName string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_not_found|勋章系统：未找到勋章“%s”"), medalName)
	}

	// 检查用户是否已拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	if exists {
		return common.T("", "medal_grant_exists|用户已经拥有该勋章了。"), nil
	}

	// 发放勋章
	_, err = GlobalDB.Exec(
		"INSERT INTO user_medal (user_id, medal_id, grant_time) VALUES ($1, $2, CURRENT_TIMESTAMP)",
		userID, medalID,
	)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, operatorID, "管理员发放", 1,
	)
	if err != nil {
		log.Printf("记录勋章发放日志失败: %v\n", err)
	}

	return common.T("", "medal_grant_success|成功为用户 %s 发放了勋章“%s”！", userID, medalName), nil
}

// doRemoveMedal 移除勋章
func (p *MedalPlugin) doRemoveMedal(operatorID string, userID string, medalName string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_not_found|勋章系统：未找到勋章“%s”"), medalName)
	}

	// 检查用户是否拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	if !exists {
		return common.T("", "medal_remove_not_exists|该用户并未拥有此勋章。"), nil
	}

	// 移除勋章
	_, err = GlobalDB.Exec(
		"UPDATE user_medal SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND medal_id = $2",
		userID, medalID,
	)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, operatorID, "管理员移除", 0,
	)
	if err != nil {
		log.Printf("记录勋章移除日志失败: %v\n", err)
	}

	return common.T("", "medal_remove_success|成功为用户 %s 移除了勋章“%s”！", userID, medalName), nil
}

// doUpgradeMedal 升级勋章
func (p *MedalPlugin) doUpgradeMedal(operatorID string, userID string, medalName string, level int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "medal_db_not_init|勋章系统：数据库未初始化"))
	}

	if level <= 0 {
		return "", fmt.Errorf(common.T("", "medal_upgrade_level_invalid|勋章系统：等级必须大于0"))
	}

	// 查找勋章
	var medalID uint
	err := GlobalDB.QueryRow("SELECT id FROM medal WHERE name = $1 AND is_enabled = TRUE", medalName).Scan(&medalID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_not_found|勋章系统：未找到勋章“%s”"), medalName)
	}

	// 检查用户是否拥有
	var exists bool
	err = GlobalDB.QueryRow(
		"SELECT COUNT(*) > 0 FROM user_medal WHERE user_id = $1 AND medal_id = $2 AND is_active = TRUE",
		userID, medalID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	if !exists {
		return common.T("", "medal_remove_not_exists|该用户并未拥有此勋章。"), nil
	}

	// 升级勋章
	_, err = GlobalDB.Exec(
		"UPDATE user_medal SET level = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2 AND medal_id = $3",
		level, userID, medalID,
	)
	if err != nil {
		return "", fmt.Errorf(common.T("", "medal_op_failed|勋章系统：操作失败，请重试"))
	}

	// 记录日志
	_, err = GlobalDB.Exec(
		"INSERT INTO medal_grant_log (user_id, medal_id, operator, reason, level) VALUES ($1, $2, $3, $4, $5)",
		userID, medalID, operatorID, "管理员升级", level,
	)
	if err != nil {
		log.Printf("记录勋章升级日志失败: %v\n", err)
	}

	return common.T("", "medal_upgrade_success|成功将用户 %s 的勋章“%s”升级到第 %d 级！", userID, medalName, level), nil
}

// sendMessage 发送消息
func (p *MedalPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if robot == nil || event == nil || msg == "" {
		return
	}
	_, _ = SendTextReply(robot, event, msg)
}
