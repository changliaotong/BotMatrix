package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"time"
)

// BabyProduct 宝宝用品定义
type BabyProduct struct {
	ID          string
	Name        string
	Price       int
	GrowthValue int
}

var babyProducts = map[string]*BabyProduct{
	"milk": {ID: "milk", Name: "奶粉", Price: 100, GrowthValue: 50},
	"toy":  {ID: "toy", Name: "玩具", Price: 200, GrowthValue: 100},
	"book": {ID: "book", Name: "绘本", Price: 500, GrowthValue: 300},
}

// BabyPlugin 宝宝系统插件
type BabyPlugin struct {
	cmdParser *CommandParser
}

// Baby 宝宝数据模型
type Baby struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"size:20;index" json:"user_id"`
	Name        string    `gorm:"size:50" json:"name"`
	Birthday    time.Time `json:"birthday"`
	GrowthValue int       `json:"growth_value"`
	DaysOld     int       `json:"days_old"`
	Level       int       `json:"level"`
	Status      string    `gorm:"size:20;default:active" json:"status"` // active, abandoned
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BabyEvent 宝宝事件记录
type BabyEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BabyID    uint      `json:"baby_id"`
	EventType string    `gorm:"size:50" json:"event_type"` // birthday, learn, work, interact
	Content   string    `gorm:"size:255" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// BabyConfig 宝宝系统配置
type BabyConfig struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	IsEnabled  bool      `gorm:"default:true" json:"is_enabled"`
	GrowthRate int       `gorm:"default:1000" json:"growth_rate"` // 每1000成长值增加1天
	UpdateAt   time.Time `json:"update_at"`
}

// NewBabyPlugin 创建宝宝系统插件实例
func NewBabyPlugin() *BabyPlugin {
	return &BabyPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *BabyPlugin) Name() string {
	return "baby"
}

func (p *BabyPlugin) Description() string {
	return common.T("", "baby_plugin_desc|宝宝系统插件，可以领养、培养和互动的小生命")
}

func (p *BabyPlugin) Version() string {
	return "1.0.0"
}

// HandleSkill 处理技能调用
func (p *BabyPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	var userID string
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if params["user_id"] != "" {
		userID = params["user_id"]
	}

	switch skillName {
	case "baby_birth":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		msg, err := p.doBabyBirth(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "my_baby":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		msg, err := p.doMyBaby(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_learn":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		msg, err := p.doBabyLearn(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_mall":
		msg, err := p.doBabyMall()
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "buy_product":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		productID := params["product_id"]
		if productID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_product_id|缺少商品ID"))
		}
		msg, err := p.doBuyProduct(userID, productID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_interact":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		msg, err := p.doBabyInteract(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_work":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		msg, err := p.doBabyWork(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_rename":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_user_id|缺少用户ID"))
		}
		newName := params["new_name"]
		if newName == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_new_name|请输入新名字"))
		}
		msg, err := p.doBabyRename(userID, newName)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "enable_baby_system":
		msg, err := p.doEnableSystem(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "disable_baby_system":
		msg, err := p.doDisableSystem(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "abandon_baby":
		adminID := params["admin_id"]
		if adminID == "" {
			adminID = userID
		}
		targetUserID := params["target_user_id"]
		if targetUserID == "" {
			return "", fmt.Errorf(common.T("", "baby_missing_target_user_id|缺少目标用户ID"))
		}
		msg, err := p.doAbandonBaby(adminID, targetUserID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "baby_abandon_info":
		msg, err := p.doBabyAbandonInfo()
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil
	}
	return "", nil
}

// GetSkills 返回插件提供的技能列表
func (p *BabyPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "baby_birth",
			Description: common.T("", "baby_skill_birth_desc|让一个新的宝宝降临到你身边"),
			Usage:       "baby_birth user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "my_baby",
			Description: common.T("", "baby_skill_my_baby_desc|查看我的宝宝详细信息"),
			Usage:       "my_baby user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "baby_learn",
			Description: common.T("", "baby_skill_learn_desc|让宝宝学习，增加成长值"),
			Usage:       "baby_learn user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "baby_mall",
			Description: common.T("", "baby_skill_mall_desc|查看宝宝用品商城"),
			Usage:       "baby_mall",
			Params:      map[string]string{},
		},
		{
			Name:        "buy_product",
			Description: common.T("", "baby_skill_buy_product_desc|为宝宝购买商品"),
			Usage:       "buy_product user_id=123456 product_id=1",
			Params: map[string]string{
				"user_id":    common.T("", "baby_skill_param_user_id|用户ID"),
				"product_id": common.T("", "baby_skill_param_product_id|商品ID"),
			},
		},
		{
			Name:        "baby_interact",
			Description: common.T("", "baby_skill_interact_desc|与宝宝进行互动，增加成长值"),
			Usage:       "baby_interact user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "baby_work",
			Description: common.T("", "baby_skill_work_desc|让宝宝去打工，增加成长值和积分"),
			Usage:       "baby_work user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "baby_rename",
			Description: common.T("", "baby_skill_rename_desc|给宝宝改一个新的名字"),
			Usage:       "baby_rename user_id=123456 new_name=小可爱",
			Params: map[string]string{
				"user_id":  common.T("", "baby_skill_param_user_id|用户ID"),
				"new_name": common.T("", "baby_skill_param_new_name|新名字"),
			},
		},
		{
			Name:        "enable_baby_system",
			Description: common.T("", "baby_skill_enable_desc|开启宝宝系统（仅限管理员）"),
			Usage:       "enable_baby_system user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_admin_id|管理员ID"),
			},
		},
		{
			Name:        "disable_baby_system",
			Description: common.T("", "baby_skill_disable_desc|关闭宝宝系统（仅限管理员）"),
			Usage:       "disable_baby_system user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "baby_skill_param_admin_id|管理员ID"),
			},
		},
		{
			Name:        "abandon_baby",
			Description: common.T("", "baby_skill_abandon_desc|抛弃指定用户的宝宝（仅限管理员）"),
			Usage:       "abandon_baby admin_id=123456 target_user_id=654321",
			Params: map[string]string{
				"admin_id":       common.T("", "baby_skill_param_admin_id|管理员ID"),
				"target_user_id": common.T("", "baby_skill_param_target_user_id|目标用户ID"),
			},
		},
		{
			Name:        "baby_abandon_info",
			Description: common.T("", "baby_skill_abandon_info_desc|查看宝宝拐卖（抛弃）系统说明"),
			Usage:       "baby_abandon_info",
			Params:      map[string]string{},
		},
	}
}

func (p *BabyPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "baby_plugin_loaded|宝宝系统插件已加载"))

	// 初始化数据库
	p.initDatabase()

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理宝宝系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		// 宝宝降临
		if match, _ := p.cmdParser.MatchCommand("宝宝降临", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBabyBirth(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 我的宝宝
		if match, _ := p.cmdParser.MatchCommand("我的宝宝", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doMyBaby(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 宝宝学习
		if match, _ := p.cmdParser.MatchCommand("宝宝学习", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBabyLearn(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 宝宝商城
		if match, _ := p.cmdParser.MatchCommand("宝宝商城", event.RawMessage); match {
			msg, _ := p.doBabyMall()
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 购买商品
		if match, params := p.cmdParser.MatchRegex("购买(\\d+)", event.RawMessage); match && len(params) > 1 {
			productID := params[1]
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBuyProduct(userID, productID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 宝宝互动
		if match, _ := p.cmdParser.MatchCommand("宝宝互动", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBabyInteract(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 宝宝打工
		if match, _ := p.cmdParser.MatchCommand("宝宝打工", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBabyWork(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 宝宝改名
		if match, params := p.cmdParser.MatchRegex("宝宝改名\\+(\\S+)", event.RawMessage); match && len(params) > 1 {
			newName := params[1]
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doBabyRename(userID, newName)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 开启宝宝系统
		if match, _ := p.cmdParser.MatchCommand("开启宝宝系统", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doEnableSystem(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 关闭宝宝系统
		if match, _ := p.cmdParser.MatchCommand("关闭宝宝系统", event.RawMessage); match {
			userID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doDisableSystem(userID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 超管抛弃宝宝功能
		if match, params := p.cmdParser.MatchRegex("抛弃宝宝(\\d+)", event.RawMessage); match && len(params) > 1 {
			targetUserID := params[1]
			adminID := fmt.Sprintf("%d", event.UserID)
			msg, err := p.doAbandonBaby(adminID, targetUserID)
			if err != nil {
				p.sendMessage(robot, event, "❌ "+err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 拐卖宝宝说明
		if match, _ := p.cmdParser.MatchCommand("拐卖宝宝说明", event.RawMessage); match {
			msg, _ := p.doBabyAbandonInfo()
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
func (p *BabyPlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，宝宝系统将使用模拟数据")
		return
	}

	// 创建宝宝表
	createBabyTable := `
	CREATE TABLE IF NOT EXISTS baby (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		name VARCHAR(50) NOT NULL,
		birthday TIMESTAMP NOT NULL,
		growth_value INT NOT NULL DEFAULT 0,
		days_old INT NOT NULL DEFAULT 0,
		level INT NOT NULL DEFAULT 1,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := GlobalDB.Exec(createBabyTable)
	if err != nil {
		log.Printf("创建宝宝表失败: %v\n", err)
		return
	}

	// 创建宝宝事件表
	createBabyEventTable := `
	CREATE TABLE IF NOT EXISTS baby_event (
		id SERIAL PRIMARY KEY,
		baby_id INT NOT NULL REFERENCES baby(id) ON DELETE CASCADE,
		event_type VARCHAR(50) NOT NULL,
		content VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBabyEventTable)
	if err != nil {
		log.Printf("创建宝宝事件表失败: %v\n", err)
		return
	}

	// 创建宝宝系统配置表
	createBabyConfigTable := `
	CREATE TABLE IF NOT EXISTS baby_config (
		id SERIAL PRIMARY KEY,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		growth_rate INT NOT NULL DEFAULT 1000,
		update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBabyConfigTable)
	if err != nil {
		log.Printf("创建宝宝系统配置表失败: %v\n", err)
		return
	}

	// 初始化配置
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM baby_config").Scan(&count)
	if err != nil {
		log.Printf("查询宝宝系统配置失败: %v\n", err)
		return
	}

	if count == 0 {
		_, err = GlobalDB.Exec("INSERT INTO baby_config (is_enabled, growth_rate) VALUES (TRUE, 1000)")
		if err != nil {
			log.Printf("初始化宝宝系统配置失败: %v\n", err)
			return
		}
	}

	log.Println("宝宝系统数据库初始化完成")
}

// isSystemEnabled 检查宝宝系统是否开启
func (p *BabyPlugin) isSystemEnabled() bool {
	if GlobalDB == nil {
		// 如果没有数据库连接，默认返回开启状态
		return true
	}

	// 查询系统配置
	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM baby_config LIMIT 1").Scan(&isEnabled)
	if err != nil {
		// 如果查询失败，默认返回开启状态
		log.Printf("查询宝宝系统配置失败: %v\n", err)
		return true
	}

	return isEnabled
}

// sendMessage 发送消息并进行 nil 检查
func (p *BabyPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	SendTextReply(robot, event, message)
}

// babyBirth 宝宝降临功能
func (p *BabyPlugin) babyBirth(robot plugin.Robot, event *onebot.Event) {
	userID := fmt.Sprintf("%d", event.UserID)
	msg, err := p.doBabyBirth(userID)
	if err != nil {
		p.sendMessage(robot, event, "❌ "+err.Error())
		return
	}
	p.sendMessage(robot, event, msg)
}

// doBabyBirth 执行宝宝降临逻辑
func (p *BabyPlugin) doBabyBirth(userID string) (string, error) {
	// 检查用户是否已有宝宝
	if GlobalDB != nil {
		var count int
		err := GlobalDB.QueryRow("SELECT COUNT(*) FROM baby WHERE user_id = ? AND status = 'active'", userID).Scan(&count)
		if err != nil {
			log.Printf("查询用户宝宝失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_db_query_failed|数据库查询失败"))
		}

		if count > 0 {
			return "", fmt.Errorf(common.T("", "baby_already_has|你已经有一个宝宝了，不能再领养了哦"))
		}
	}

	// 创建新宝宝
	baby := Baby{
		UserID:      userID,
		Name:        "小宝宝",
		Birthday:    time.Now(),
		GrowthValue: 0,
		DaysOld:     0,
		Level:       1,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存宝宝数据到数据库
	if GlobalDB != nil {
		insertQuery := `
		INSERT INTO baby (user_id, name, birthday, growth_value, days_old, level, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := GlobalDB.Exec(insertQuery,
			baby.UserID, baby.Name, baby.Birthday, baby.GrowthValue, baby.DaysOld,
			baby.Level, baby.Status, baby.CreatedAt, baby.UpdatedAt)
		if err != nil {
			log.Printf("创建宝宝失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_birth_failed|宝宝降临失败，请稍后再试"))
		}
	}

	msg := common.T("", "baby_birth_success|🎊 恭喜！一个新的小生命降临了！")
	msg += fmt.Sprintf(common.T("", "baby_info_name_val|\n宝宝名字：%s"), baby.Name)
	msg += fmt.Sprintf(common.T("", "baby_info_birthday_val|\n出生日期：%s"), baby.Birthday.Format("2006-01-02"))
	msg += common.T("", "baby_birth_tip|\n记得要好好照顾他/她哦！")
	msg += common.T("", "baby_my_baby_tip|\n发送“我的宝宝”查看详情。")

	return msg, nil
}

// myBaby 我的宝宝功能
func (p *BabyPlugin) myBaby(robot plugin.Robot, event *onebot.Event) {
	userID := fmt.Sprintf("%d", event.UserID)
	msg, err := p.doMyBaby(userID)
	if err != nil {
		SendTextReply(robot, event, "❌ "+err.Error())
		return
	}
	SendTextReply(robot, event, msg)
}

// doMyBaby 执行我的宝宝详情查询逻辑
func (p *BabyPlugin) doMyBaby(userID string) (string, error) {
	// 查询用户的宝宝
	var baby Baby
	if GlobalDB != nil {
		row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", userID)
		err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
		if err != nil {
			return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
		}
	} else {
		// 如果没有数据库连接，使用模拟数据
		baby = Baby{
			Name:        "小宝宝",
			Birthday:    time.Now().AddDate(0, 0, -10),
			GrowthValue: 5000,
			DaysOld:     5,
			Level:       1,
		}
	}

	msg := common.T("", "baby_info_title|🍼 我的宝宝详情\n")
	msg += "================================\n"
	msg += fmt.Sprintf(common.T("", "baby_info_name|名字：%s\n"), baby.Name)
	msg += fmt.Sprintf(common.T("", "baby_info_birthday|生日：%s\n"), baby.Birthday.Format("2006-01-02"))
	msg += fmt.Sprintf(common.T("", "baby_info_age|年龄：%s\n"), p.getBabyAge(baby))
	msg += fmt.Sprintf(common.T("", "baby_info_growth|成长值：%s\n"), fmt.Sprintf("%d", baby.GrowthValue))
	msg += fmt.Sprintf(common.T("", "baby_info_level|等级：Lv.%s\n"), fmt.Sprintf("%d", baby.Level))
	msg += "================================\n"
	msg += common.T("", "baby_commands_hint|💡 提示：你可以通过“宝宝学习”、“宝宝互动”、“宝宝打工”来培养他/她。")

	return msg, nil
}

// babyLearn 宝宝学习功能
func (p *BabyPlugin) babyLearn(robot plugin.Robot, event *onebot.Event) {
	userID := fmt.Sprintf("%d", event.UserID)
	msg, err := p.doBabyLearn(userID)
	if err != nil {
		SendTextReply(robot, event, "❌ "+err.Error())
		return
	}
	SendTextReply(robot, event, msg)
}

// doBabyLearn 执行宝宝学习逻辑
func (p *BabyPlugin) doBabyLearn(userID string) (string, error) {
	// 查询用户的宝宝
	var baby Baby
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", userID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 增加成长值
	growthAdd := 100
	newGrowthValue := baby.GrowthValue + growthAdd

	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	if newDays > baby.DaysOld {
		// 更新天数和等级
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ? WHERE id = ?",
			newGrowthValue, newDays, newDays/30+1, baby.ID)
		if err != nil {
			log.Printf("更新宝宝学习数据失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_learn_failed|更新宝宝学习数据失败"))
		}

		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
		baby.DaysOld = newDays
		baby.Level = newDays/30 + 1
	} else {
		// 只更新成长值
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ? WHERE id = ?", newGrowthValue, baby.ID)
		if err != nil {
			log.Printf("更新宝宝学习数据失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_learn_failed|更新宝宝学习数据失败"))
		}

		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
	}

	// 记录学习事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
		baby.ID, "learn", fmt.Sprintf(common.T("", "baby_event_learn|宝宝努力学习，成长值增加了 %d"), growthAdd))
	if err != nil {
		log.Printf("记录宝宝学习事件失败: %v\n", err)
	}

	return fmt.Sprintf(common.T("", "baby_learn_success|📖 学习使人进步！宝宝成长值增加了 %d，当前成长值：%d，当前等级：%d"), growthAdd, baby.GrowthValue, baby.Level), nil
}

// doBabyMall 执行获取商城信息逻辑
func (p *BabyPlugin) doBabyMall() (string, error) {
	return common.T("", "baby_mall_title|🛒 宝宝用品商城"), nil
}

// doBuyProduct 执行购买商品逻辑
func (p *BabyPlugin) doBuyProduct(userID string, productID string) (string, error) {
	// 检查商品是否存在
	product, ok := babyProducts[productID]
	if !ok {
		return "", fmt.Errorf(common.T("", "baby_product_not_found|抱歉，没有找到该商品"))
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 查询用户的宝宝
	var baby Baby
	row := GlobalDB.QueryRow("SELECT id, user_id, name, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", userID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 增加宝宝成长值
	growthAdd := product.GrowthValue
	newGrowthValue := baby.GrowthValue + growthAdd

	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	newLevel := baby.Level
	if newDays > baby.DaysOld {
		newLevel = newDays/30 + 1 // 每30天升1级
	}

	// 更新宝宝信息
	_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newGrowthValue, newDays, newLevel, baby.ID)
	if err != nil {
		log.Printf("更新宝宝购买数据失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_buy_failed|购买失败，请稍后再试"))
	}

	// 记录购买事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
		baby.ID, "buy", fmt.Sprintf(common.T("", "baby_event_buy|给宝宝购买了 %s，成长值增加了 %d"), product.Name, growthAdd))
	if err != nil {
		log.Printf("记录宝宝购买事件失败: %v\n", err)
	}

	return fmt.Sprintf(common.T("", "baby_buy_success|🎁 购买成功！宝宝使用了 %s，成长值增加了 %d"), product.Name, growthAdd), nil
}

// doBabyInteract 执行宝宝互动逻辑
func (p *BabyPlugin) doBabyInteract(userID string) (string, error) {
	// 查询用户的宝宝
	var baby Baby
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", userID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 增加成长值
	growthAdd := 50
	newGrowthValue := baby.GrowthValue + growthAdd

	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	if newDays > baby.DaysOld {
		// 更新天数和等级
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ? WHERE id = ?",
			newGrowthValue, newDays, newDays/30+1, baby.ID)
		if err != nil {
			log.Printf("更新宝宝互动数据失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_interact_failed|互动失败，请稍后再试"))
		}

		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
		baby.DaysOld = newDays
		baby.Level = newDays/30 + 1
	} else {
		// 只更新成长值
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ? WHERE id = ?", newGrowthValue, baby.ID)
		if err != nil {
			log.Printf("更新宝宝互动数据失败: %v\n", err)
			return "", fmt.Errorf(common.T("", "baby_interact_failed|互动失败，请稍后再试"))
		}

		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
	}

	// 记录互动事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
		baby.ID, "interact", fmt.Sprintf(common.T("", "baby_event_interact|与宝宝进行了亲密互动，成长值增加了 %d"), growthAdd))
	if err != nil {
		log.Printf("记录宝宝互动事件失败: %v\n", err)
	}

	return fmt.Sprintf(common.T("", "baby_interact_success|😊 互动成功！宝宝很开心，成长值增加了 %d，当前总成长值：%d"), growthAdd, baby.GrowthValue), nil
}

// doBabyWork 执行宝宝打工逻辑
func (p *BabyPlugin) doBabyWork(userID string) (string, error) {
	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 查询用户的宝宝
	var baby Baby
	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", userID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 检查宝宝年龄是否足够打工（至少30天）
	if baby.DaysOld < 30 {
		return "", fmt.Errorf(common.T("", "baby_too_young_to_work|你的宝宝太小了（%s），还不满30天，不能去打工赚钱哦"), p.getBabyAge(baby))
	}

	// 增加成长值和积分
	growthAdd := 150
	pointsAdd := 50
	newGrowthValue := baby.GrowthValue + growthAdd

	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	newLevel := baby.Level
	if newDays > baby.DaysOld {
		newLevel = newDays/30 + 1 // 每30天升1级
	}

	// 更新宝宝信息
	_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newGrowthValue, newDays, newLevel, baby.ID)
	if err != nil {
		log.Printf("更新宝宝打工数据失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_work_failed|打工失败，请稍后再试"))
	}

	// 记录宝宝打工事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
		baby.ID, "work", fmt.Sprintf(common.T("", "baby_event_work|宝宝勤劳打工，赚取了 %d 积分，成长值增加了 %d"), pointsAdd, growthAdd))
	if err != nil {
		log.Printf("记录宝宝打工事件失败: %v\n", err)
	}

	return fmt.Sprintf(common.T("", "baby_work_success|💰 打工成功！宝宝赚取了 %d 积分，成长值增加了 %d，当前总成长值：%d"), pointsAdd, growthAdd, newGrowthValue), nil
}

// doBabyRename 执行宝宝改名逻辑
func (p *BabyPlugin) doBabyRename(userID string, newName string) (string, error) {
	if len(newName) < 2 || len(newName) > 10 {
		return "", fmt.Errorf(common.T("", "baby_name_length_error|宝宝名字长度必须在 2 到 10 个字符之间"))
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 查询用户的宝宝
	var oldName string
	row := GlobalDB.QueryRow("SELECT name FROM baby WHERE user_id = ? AND status = 'active'", userID)
	err := row.Scan(&oldName)
	if err != nil {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 更新宝宝名字
	_, err = GlobalDB.Exec("UPDATE baby SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND status = 'active'", newName, userID)
	if err != nil {
		log.Printf("更新宝宝名字失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_rename_failed|修改宝宝名字失败"))
	}

	// 记录改名事件
	var babyID int
	row = GlobalDB.QueryRow("SELECT id FROM baby WHERE user_id = ? AND status = 'active'", userID)
	row.Scan(&babyID)

	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
		babyID, "rename", fmt.Sprintf(common.T("", "baby_event_rename|宝宝改名了，从 %s 改为了 %s"), oldName, newName))
	if err != nil {
		log.Printf("记录宝宝改名事件失败: %v\n", err)
	}

	return fmt.Sprintf(common.T("", "baby_rename_success|✅ 修改成功！宝宝现在叫 %s 啦"), newName), nil
}

// doEnableSystem 执行开启系统逻辑
func (p *BabyPlugin) doEnableSystem(userID string) (string, error) {
	// 检查用户权限
	if !p.isSuperAdmin(userID) {
		return "", fmt.Errorf(common.T("", "baby_not_admin|抱歉，该操作仅限超级管理员使用"))
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 更新系统配置为开启
	_, err := GlobalDB.Exec("UPDATE baby_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("开启宝宝系统失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_db_query_failed|数据库查询失败"))
	}

	return common.T("", "baby_system_enabled|✅ 宝宝系统已成功开启"), nil
}

// doDisableSystem 执行关闭系统逻辑
func (p *BabyPlugin) doDisableSystem(userID string) (string, error) {
	// 检查用户权限
	if !p.isSuperAdmin(userID) {
		return "", fmt.Errorf(common.T("", "baby_not_admin|抱歉，该操作仅限超级管理员使用"))
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 更新系统配置为关闭
	_, err := GlobalDB.Exec("UPDATE baby_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("关闭宝宝系统失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_db_query_failed|数据库查询失败"))
	}

	return common.T("", "baby_system_disabled_msg|✅ 宝宝系统已成功关闭"), nil
}

// doAbandonBaby 执行抛弃宝宝逻辑
func (p *BabyPlugin) doAbandonBaby(adminID string, targetUserID string) (string, error) {
	// 检查用户权限
	if !p.isSuperAdmin(adminID) {
		return "", fmt.Errorf(common.T("", "baby_not_admin|抱歉，该操作仅限超级管理员使用"))
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "baby_db_conn_failed|数据库连接失败，请稍后再试"))
	}

	// 查询用户的宝宝
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM baby WHERE user_id = ? AND status = 'active'", targetUserID).Scan(&count)
	if err != nil {
		log.Printf("查询用户宝宝失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_db_query_failed|数据库查询失败"))
	}

	if count == 0 {
		return "", fmt.Errorf(common.T("", "baby_no_baby|你还没有领养宝宝呢，发送“宝宝降临”来领养一个吧"))
	}

	// 标记宝宝为已抛弃
	_, err = GlobalDB.Exec("UPDATE baby SET status = 'abandoned', updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND status = 'active'", targetUserID)
	if err != nil {
		log.Printf("抛弃宝宝失败: %v\n", err)
		return "", fmt.Errorf(common.T("", "baby_abandon_failed|抛弃宝宝失败"))
	}

	return fmt.Sprintf(common.T("", "baby_abandon_success|✅ 已成功抛弃用户 %s 的宝宝"), targetUserID), nil
}

// doBabyAbandonInfo 执行获取拐卖说明逻辑
func (p *BabyPlugin) doBabyAbandonInfo() (string, error) {
	return common.T("", "baby_abandon_info_content|宝宝拐卖（抛弃）系统说明：\n1. 仅管理员可操作\n2. 抛弃后宝宝将处于 abandoned 状态"), nil
}

// getBabyAge 获取宝宝年龄描述
func (p *BabyPlugin) getBabyAge(baby Baby) string {
	duration := time.Since(baby.Birthday)
	days := int(duration.Hours() / 24)
	years := days / 365
	remainingDays := days % 365

	if years > 0 {
		return fmt.Sprintf(common.T("", "baby_age_format|%d岁%d天"), years, remainingDays)
	}
	return fmt.Sprintf(common.T("", "baby_age_days|%d天"), days)
}

// isSuperAdmin 检查是否为超级管理员
func (p *BabyPlugin) isSuperAdmin(userID string) bool {
	// 超级管理员列表（实际使用时应从配置或数据库读取）
	// 这里暂时硬编码几个示例ID用于测试
	superAdmins := []string{
		"123456789", // 示例超级管理员ID
		"987654321", // 示例超级管理员ID
	}

	// 检查用户ID是否在超级管理员列表中
	for _, adminID := range superAdmins {
		if userID == adminID {
			return true
		}
	}

	return false
}

// updateGrowthValue 更新宝宝成长值
func (p *BabyPlugin) updateGrowthValue() {
	log.Println(common.T("", "baby_log_start_update|开始更新宝宝每日成长值..."))

	// 检查全局数据库连接
	if GlobalDB == nil {
		log.Println(common.T("", "baby_log_db_not_init|全局数据库未初始化，停止更新宝宝数据"))
		return
	}

	// 查询所有活跃状态的宝宝
	rows, err := GlobalDB.Query("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE status = 'active'")
	if err != nil {
		log.Printf("查询活跃宝宝失败: %v\n", err)
		return
	}
	defer rows.Close()

	// 遍历所有宝宝，更新成长值
	for rows.Next() {
		var baby Baby
		err := rows.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
		if err != nil {
			log.Printf("扫描宝宝数据失败: %v\n", err)
			continue
		}

		growthAdd := 50 // 每日自动增加50成长值
		newGrowthValue := baby.GrowthValue + growthAdd

		// 计算应该增加的天数（每1000成长值=1天）
		newDays := newGrowthValue / 1000
		if newDays > baby.DaysOld {
			newLevel := newDays/30 + 1 // 每30天升1级

			// 更新宝宝数据到数据库
			_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
				newGrowthValue, newDays, newLevel, baby.ID)
			if err != nil {
				log.Printf("更新宝宝 %s 数据失败: %v\n", baby.Name, err)
				continue
			}

			// 更新本地变量用于后续处理
			baby.GrowthValue = newGrowthValue
			baby.DaysOld = newDays
			baby.Level = newLevel

			// 检查是否过生日
			p.checkBirthday(baby)
			log.Printf("宝宝 %s 更新完成：成长值=%d, 天数=%d, 等级=%d\n", baby.Name, baby.GrowthValue, baby.DaysOld, baby.Level)
		} else {
			// 只更新成长值
			_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
				newGrowthValue, baby.ID)
			if err != nil {
				log.Printf("更新宝宝 %s 成长值失败: %v\n", baby.Name, err)
				continue
			}
			log.Printf("宝宝 %s 更新完成：成长值=%d\n", baby.Name, newGrowthValue)
		}

		// 检查是否达到宝宝达人徽章条件（成长值达到10000）
		if newGrowthValue >= 10000 && baby.GrowthValue < 10000 {
			// 获取徽章插件实例
			badgePlugin := GetBadgePluginInstance()
			// 发放宝宝达人徽章
			err := badgePlugin.GrantBadgeToUser(baby.UserID, "宝宝达人", "system", "宝宝成长值达到10000")
			if err != nil {
				log.Printf("给宝宝 %s 的用户 %s 发放宝宝达人徽章失败: %v\n", baby.Name, baby.UserID, err)
			} else {
				log.Printf("给宝宝 %s 的用户 %s 成功发放宝宝达人徽章\n", baby.Name, baby.UserID)
			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("遍历宝宝数据失败: %v\n", err)
	}

	log.Println(common.T("", "baby_log_update_finished|宝宝每日成长值更新任务完成"))
}

// checkBirthday 检查宝宝是否过生日
func (p *BabyPlugin) checkBirthday(baby Baby) {
	now := time.Now()
	birthMonth := baby.Birthday.Month()
	birthDay := baby.Birthday.Day()

	// 检查是否是生日
	if now.Month() == birthMonth && now.Day() == birthDay {
		// 如果是生日，记录生日事件
		_, err := GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)",
			baby.ID, "birthday", fmt.Sprintf(common.T("", "baby_event_birthday|宝宝今天过生日啦！现在 %d 天了"), baby.DaysOld))
		if err != nil {
			log.Printf("记录宝宝 %s 生日事件失败: %v\n", baby.Name, err)
			return
		}

		log.Printf("🎉 宝宝 %s 今天过生日了！现在 %d 天了\n", baby.Name, baby.DaysOld)
	}
}
