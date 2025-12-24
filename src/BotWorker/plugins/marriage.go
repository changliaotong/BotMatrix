package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"BotMatrix/common"
	"fmt"
	"log"
	"time"
)

// MarriagePlugin 结婚系统插件
type MarriagePlugin struct {
	cmdParser *CommandParser
}

// UserMarriage 用户婚姻信息
type UserMarriage struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          string    `gorm:"size:20;index" json:"user_id"`
	SpouseID        string    `gorm:"size:20;index" json:"spouse_id"`
	MarriageDate    time.Time `json:"marriage_date"`
	DivorceDate     time.Time `json:"divorce_date"`
	Status          string    `gorm:"size:20;default:single" json:"status"` // single, married, divorced
	SweetsCount     int       `gorm:"default:0" json:"sweets_count"`
	RedPacketsCount int       `gorm:"default:0" json:"red_packets_count"`
	SweetHearts     int       `gorm:"default:0" json:"sweet_hearts"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MarriageProposal 求婚记录
type MarriageProposal struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProposerID  string    `gorm:"size:20;index" json:"proposer_id"`
	RecipientID string    `gorm:"size:20;index" json:"recipient_id"`
	Status      string    `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WeddingItem 婚礼物品
type WeddingItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ItemType    string    `gorm:"size:20" json:"item_type"` // dress, ring
	Name        string    `gorm:"size:50" json:"name"`
	Price       int       `gorm:"default:0" json:"price"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserWeddingItems 用户拥有的婚礼物品
type UserWeddingItems struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	ItemID    uint      `json:"item_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Sweets 喜糖记录
type Sweets struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"size:20;index" json:"user_id"`
	Amount      int       `json:"amount"`
	Type        string    `gorm:"size:20" json:"type"` // send, receive, eat
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// RedPacket 红包记录
type RedPacket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"size:20;index" json:"user_id"`
	Amount      int       `json:"amount"`
	Type        string    `gorm:"size:20" json:"type"` // send, receive
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// SweetHeart 甜蜜爱心
type SweetHeart struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SenderID    string    `gorm:"size:20;index" json:"sender_id"`
	RecipientID string    `gorm:"size:20;index" json:"recipient_id"`
	Amount      int       `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// MarriageConfig 结婚系统配置
type MarriageConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	IsEnabled     bool      `gorm:"default:true" json:"is_enabled"`
	SweetsCost    int       `gorm:"default:100" json:"sweets_cost"`
	RedPacketCost int       `gorm:"default:200" json:"red_packet_cost"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewMarriagePlugin 创建结婚系统插件实例
func NewMarriagePlugin() *MarriagePlugin {
	return &MarriagePlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MarriagePlugin) Name() string {
	return "marriage"
}

func (p *MarriagePlugin) Description() string {
	return common.T("", "marriage_plugin_desc|结婚系统插件，支持求婚、结婚、喜糖、甜蜜爱心等功能")
}

func (p *MarriagePlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *MarriagePlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "get_marriage_status",
			Description: common.T("", "marriage_skill_status_desc|查询用户当前的婚姻状态"),
			Usage:       "get_marriage_status user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
		{
			Name:        "propose_marriage",
			Description: common.T("", "marriage_skill_propose_desc|向其他用户发起求婚请求"),
			Usage:       "propose_marriage proposer_id=123456 recipient_id=654321",
			Params: map[string]string{
				"proposer_id":  common.T("", "marriage_param_proposer_id|求婚者QQ号"),
				"recipient_id": common.T("", "marriage_param_recipient_id|被求婚者QQ号"),
			},
		},
		{
			Name:        "accept_marriage",
			Description: common.T("", "marriage_skill_accept_desc|接受来自其他用户的求婚请求"),
			Usage:       "accept_marriage recipient_id=654321 proposer_id=123456",
			Params: map[string]string{
				"recipient_id": common.T("", "marriage_param_recipient_id|被求婚者QQ号"),
				"proposer_id":  common.T("", "marriage_param_proposer_id|求婚者QQ号"),
			},
		},
		{
			Name:        "divorce_marriage",
			Description: common.T("", "marriage_skill_divorce_desc|申请解除当前的婚姻关系"),
			Usage:       "divorce_marriage user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
		{
			Name:        "send_marriage_sweets",
			Description: common.T("", "marriage_skill_send_sweets_desc|向群内发放喜糖"),
			Usage:       "send_marriage_sweets user_id=123456 count=10",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
				"count":   common.T("", "marriage_param_count|数量"),
			},
		},
		{
			Name:        "eat_marriage_sweets",
			Description: common.T("", "marriage_skill_eat_sweets_desc|吃喜糖并获得奖励"),
			Usage:       "eat_marriage_sweets user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
		{
			Name:        "get_my_spouse",
			Description: common.T("", "marriage_skill_spouse_desc|查询自己的配偶信息"),
			Usage:       "get_my_spouse user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
		{
			Name:        "get_marriage_assets",
			Description: common.T("", "marriage_skill_assets_desc|查询个人的婚姻资产（喜糖、红包、爱心）"),
			Usage:       "get_marriage_assets user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
		{
			Name:        "use_marriage_lottery",
			Description: common.T("", "marriage_skill_lottery_desc|使用甜蜜爱心参与抽奖"),
			Usage:       "use_marriage_lottery user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "marriage_param_user_id|用户QQ号"),
			},
		},
	}
}

func (p *MarriagePlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "marriage_plugin_loaded|✅ 结婚系统插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 初始化数据库
	p.initDatabase()

	// 统一处理结婚系统相关命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		userID := fmt.Sprintf("%d", event.UserID)

		// 1. 购买婚纱
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_buy_dress|购买婚纱|buy_dress"), event.RawMessage); match {
			p.sendMessage(robot, event, common.T("", "marriage_buy_dress_success|✅ 购买婚纱成功！"))
			return nil
		}

		// 2. 购买婚戒
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_buy_ring|购买婚戒|buy_ring"), event.RawMessage); match {
			p.sendMessage(robot, event, common.T("", "marriage_buy_ring_success|✅ 购买婚戒成功！"))
			return nil
		}

		// 3. 求婚 (含正则)
		if match, params := p.cmdParser.MatchRegex(common.T("", "marriage_cmd_propose|求婚|propose")+"(\\d+)", event.RawMessage); match && len(params) > 1 {
			recipientID := params[1]
			msg, err := p.doProposeMarriage(userID, recipientID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 4. 结婚 (含正则)
		if match, params := p.cmdParser.MatchRegex(common.T("", "marriage_cmd_marry|结婚|marry")+"(\\d+)", event.RawMessage); match && len(params) > 1 {
			proposerID := params[1]
			msg, err := p.doMarry(proposerID, userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 5. 离婚
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_divorce|离婚|divorce"), event.RawMessage); match {
			msg, err := p.doDivorce(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 6. 我的结婚证
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_cert|我的结婚证|结婚证|marriage_certificate"), event.RawMessage); match {
			msg, err := p.doMyMarriageCertificate(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 7. 发喜糖 (含正则)
		if match, params := p.cmdParser.MatchRegex(common.T("", "marriage_cmd_send_sweets|发喜糖|送喜糖|send_sweets")+"(\\d+)", event.RawMessage); match && len(params) > 1 {
			count := params[1]
			msg, err := p.doSendSweets(userID, count)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 8. 吃喜糖
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_eat_sweets|吃喜糖|抢喜糖|eat_sweets"), event.RawMessage); match {
			msg, err := p.doEatSweets(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 9. 办理结婚证 (含正则)
		if match, params := p.cmdParser.MatchRegex(common.T("", "marriage_cmd_apply_cert|办理结婚证|办结婚证|apply_marriage_cert")+"(\\d+)", event.RawMessage); match && len(params) > 1 {
			spouseID := params[1]
			msg, err := p.doApplyMarriageCertificate(userID, spouseID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 10. 办理离婚证
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_apply_divorce_cert|办理离婚证|办离婚证|apply_divorce_cert"), event.RawMessage); match {
			msg, err := p.doApplyDivorceCertificate(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 11. 另一半签到
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_spouse_signin|另一半签到|伴侣签到|spouse_signin"), event.RawMessage); match {
			msg, err := p.doSpouseSignIn(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 12. 另一半抢楼
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_spouse_floor|另一半抢楼|伴侣抢楼|spouse_floor"), event.RawMessage); match {
			msg, err := p.doSpouseGrabFloor(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 13. 另一半抢红包
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_spouse_redpacket|另一半抢红包|伴侣抢红包|spouse_redpacket"), event.RawMessage); match {
			msg, err := p.doSpouseGrabRedPacket(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 14. 我的对象
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_my_spouse|我的对象|我的伴侣|my_spouse"), event.RawMessage); match {
			msg, err := p.doMySpouse(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 15. 我的喜糖
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_my_sweets|我的喜糖|my_sweets"), event.RawMessage); match {
			msg, err := p.doMySweets(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 16. 我的红包
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_my_red_packets|我的红包|my_red_packets"), event.RawMessage); match {
			msg, err := p.doMyRedPackets(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 17. 我的甜蜜爱心
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_my_hearts|我的甜蜜爱心|我的爱心|my_hearts"), event.RawMessage); match {
			msg, err := p.doMySweetHearts(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 18. 甜蜜爱心说明
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_heart_info|甜蜜爱心说明|爱心说明|heart_info"), event.RawMessage); match {
			msg, _ := p.doSweetHeartsInfo()
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 19. 赠送甜蜜爱心 (含正则)
		if match, params := p.cmdParser.MatchRegex(common.T("", "marriage_cmd_send_heart|赠送甜蜜爱心|送爱心|send_heart")+"(\\d+)", event.RawMessage); match && len(params) > 1 {
			recipientID := params[1]
			msg, err := p.doSendSweetHeart(userID, recipientID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 20. 使用甜蜜抽奖
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_lottery|甜蜜抽奖|爱心抽奖|lottery"), event.RawMessage); match {
			msg, err := p.doUseSweetLottery(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 21. 领取结婚福利
		if match, _ := p.cmdParser.MatchCommand(common.T("", "marriage_cmd_benefits|领取结婚福利|结婚福利|benefits"), event.RawMessage); match {
			msg, err := p.doClaimMarriageBenefits(userID)
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

// HandleSkill 实现 SkillCapable 接口
func (p *MarriagePlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	var userID string
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if params["user_id"] != "" {
		userID = params["user_id"]
	}

	switch skillName {
	case "get_marriage_status":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		return p.doGetMarriageStatus(uID)
	case "propose_marriage":
		proposerID := params["proposer_id"]
		if proposerID == "" {
			proposerID = userID
		}
		recipientID := params["recipient_id"]
		if proposerID == "" || recipientID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_params|❌ 缺少必要参数"))
		}
		return p.doProposeMarriage(proposerID, recipientID)
	case "accept_marriage":
		recipientID := params["recipient_id"]
		if recipientID == "" {
			recipientID = userID
		}
		proposerID := params["proposer_id"]
		if recipientID == "" || proposerID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_params|❌ 缺少必要参数"))
		}
		return p.doMarry(proposerID, recipientID)
	case "divorce_marriage":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		return p.doDivorce(uID)
	case "send_marriage_sweets":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		countStr := params["count"]
		if uID == "" || countStr == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_params|❌ 缺少必要参数"))
		}
		return p.doSendSweets(uID, countStr)
	case "eat_marriage_sweets":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		return p.doEatSweets(uID)
	case "get_my_spouse":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		return p.doMySpouse(uID)
	case "get_marriage_assets":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		sweets, _ := p.doMySweets(uID)
		redPackets, _ := p.doMyRedPackets(uID)
		hearts, _ := p.doMySweetHearts(uID)
		return fmt.Sprintf("%s\n%s\n%s", sweets, redPackets, hearts), nil
	case "use_marriage_lottery":
		uID := params["user_id"]
		if uID == "" {
			uID = userID
		}
		if uID == "" {
			return "", fmt.Errorf(common.T("", "marriage_missing_user_id|❌ 缺少用户QQ号"))
		}
		return p.doUseSweetLottery(uID)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

func (p *MarriagePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if robot == nil || event == nil || msg == "" {
		return
	}
	_, _ = SendTextReply(robot, event, msg)
}

// initDatabase 初始化数据库
func (p *MarriagePlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println(common.T("", "marriage_db_not_initialized|⚠️ 数据库未初始化，结婚系统将使用模拟数据"))
		return
	}

	// 创建用户婚姻表
	createUserMarriageTable := `
	CREATE TABLE IF NOT EXISTS user_marriage (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		spouse_id VARCHAR(20) NOT NULL,
		marriage_date TIMESTAMP,
		divorce_date TIMESTAMP,
		status VARCHAR(20) NOT NULL DEFAULT 'single',
		sweets_count INT NOT NULL DEFAULT 0,
		red_packets_count INT NOT NULL DEFAULT 0,
		sweet_hearts INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := GlobalDB.Exec(createUserMarriageTable)
	if err != nil {
		log.Printf("创建用户婚姻表失败: %v\n", err)
		return
	}

	// 创建求婚记录表
	createMarriageProposalTable := `
	CREATE TABLE IF NOT EXISTS marriage_proposal (
		id SERIAL PRIMARY KEY,
		proposer_id VARCHAR(20) NOT NULL,
		recipient_id VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createMarriageProposalTable)
	if err != nil {
		log.Printf("创建求婚记录表失败: %v\n", err)
		return
	}

	// 创建婚礼物品表
	createWeddingItemTable := `
	CREATE TABLE IF NOT EXISTS wedding_item (
		id SERIAL PRIMARY KEY,
		item_type VARCHAR(20) NOT NULL,
		name VARCHAR(50) NOT NULL,
		price INT NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createWeddingItemTable)
	if err != nil {
		log.Printf("创建婚礼物品表失败: %v\n", err)
		return
	}

	// 创建用户婚礼物品表
	createUserWeddingItemsTable := `
	CREATE TABLE IF NOT EXISTS user_wedding_items (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		item_id INT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createUserWeddingItemsTable)
	if err != nil {
		log.Printf("创建用户婚礼物品表失败: %v\n", err)
		return
	}

	// 创建喜糖记录表
	createSweetsTable := `
	CREATE TABLE IF NOT EXISTS sweets (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		amount INT NOT NULL,
		type VARCHAR(20) NOT NULL,
		description VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createSweetsTable)
	if err != nil {
		log.Printf("创建喜糖记录表失败: %v\n", err)
		return
	}

	// 创建红包记录表
	createRedPacketTable := `
	CREATE TABLE IF NOT EXISTS red_packet (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		amount INT NOT NULL,
		type VARCHAR(20) NOT NULL,
		description VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createRedPacketTable)
	if err != nil {
		log.Printf("创建红包记录表失败: %v\n", err)
		return
	}

	// 创建甜蜜爱心表
	createSweetHeartTable := `
	CREATE TABLE IF NOT EXISTS sweet_heart (
		id SERIAL PRIMARY KEY,
		sender_id VARCHAR(20) NOT NULL,
		recipient_id VARCHAR(20) NOT NULL,
		amount INT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createSweetHeartTable)
	if err != nil {
		log.Printf("创建甜蜜爱心表失败: %v\n", err)
		return
	}

	// 创建结婚系统配置表
	createMarriageConfigTable := `
	CREATE TABLE IF NOT EXISTS marriage_config (
		id SERIAL PRIMARY KEY,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		sweets_cost INT NOT NULL DEFAULT 100,
		red_packet_cost INT NOT NULL DEFAULT 200,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createMarriageConfigTable)
	if err != nil {
		log.Printf("创建结婚系统配置表失败: %v\n", err)
		return
	}

	// 初始化配置
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM marriage_config").Scan(&count)
	if err != nil {
		log.Printf("查询结婚系统配置失败: %v\n", err)
		return
	}

	if count == 0 {
		_, err = GlobalDB.Exec("INSERT INTO marriage_config (is_enabled, sweets_cost, red_packet_cost) VALUES (TRUE, 100, 200)")
		if err != nil {
			log.Printf("初始化结婚系统配置失败: %v\n", err)
			return
		}
	}

	log.Println(common.T("", "marriage_db_init_complete|✅ 结婚系统数据库初始化完成"))
}

// isSystemEnabled 检查结婚系统是否开启
func (p *MarriagePlugin) isSystemEnabled() bool {
	if GlobalDB == nil {
		return true
	}
	return true
}

func (p *MarriagePlugin) doGetMarriageStatus(userID string) (string, error) {
	var marriage UserMarriage
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "marriage_db_conn_failed|❌ 数据库连接失败"))
	}

	err := GlobalDB.QueryRow("SELECT status, spouse_id FROM user_marriage WHERE user_id = ?", userID).Scan(&marriage.Status, &marriage.SpouseID)
	if err != nil {
		return common.T("", "marriage_status_single|🕊️ 您当前是单身状态"), nil
	}

	if marriage.Status == "married" {
		return fmt.Sprintf(common.T("", "marriage_status_married|❤️ 您已与 %s 步入婚姻殿堂"), marriage.SpouseID), nil
	}
	return fmt.Sprintf(common.T("", "marriage_status_other|ℹ️ 您的婚姻状态：%s"), marriage.Status), nil
}

func (p *MarriagePlugin) doProposeMarriage(proposerID, recipientID string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "marriage_db_conn_failed|❌ 数据库连接失败"))
	}

	// 检查自己是否单身
	var myStatus string
	row := GlobalDB.QueryRow("SELECT status FROM user_marriage WHERE user_id = $1", proposerID)
	err := row.Scan(&myStatus)
	if err != nil {
		myStatus = "single"
	}

	if myStatus != "single" {
		return "", fmt.Errorf(common.T("", "marriage_not_single|❌ 您当前不是单身，无法向他人求婚"))
	}

	// 检查对方是否单身
	var targetStatus string
	row = GlobalDB.QueryRow("SELECT status FROM user_marriage WHERE user_id = $1", recipientID)
	err = row.Scan(&targetStatus)
	if err != nil {
		targetStatus = "single"
	}

	if targetStatus != "single" {
		return "", fmt.Errorf(common.T("", "marriage_target_not_single|❌ 对方当前不是单身，无法接受您的求婚"))
	}

	// 检查是否已经有未处理的求婚记录
	var proposalCount int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM marriage_proposal WHERE proposer_id = $1 AND recipient_id = $2 AND status = 'pending'", proposerID, recipientID).Scan(&proposalCount)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_query_proposal_failed|❌ 查询求婚记录失败"))
	}

	if proposalCount > 0 {
		return "", fmt.Errorf(common.T("", "marriage_already_proposed|❌ 您已经向对方发起过求婚，请耐心等待回应"))
	}

	// 创建求婚记录
	_, err = GlobalDB.Exec("INSERT INTO marriage_proposal (proposer_id, recipient_id, status) VALUES ($1, $2, 'pending')", proposerID, recipientID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_propose_failed|❌ 发起求婚失败"))
	}

	return common.T("", "marriage_propose_success|💍 求婚成功！请等待对方接受"), nil
}

func (p *MarriagePlugin) doMarry(proposerID, recipientID string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "marriage_db_conn_failed|❌ 数据库连接失败"))
	}

	// 检查是否有求婚记录
	var proposalID int
	err := GlobalDB.QueryRow("SELECT id FROM marriage_proposal WHERE proposer_id = $1 AND recipient_id = $2 AND status = 'pending'", proposerID, recipientID).Scan(&proposalID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_no_proposal|❌ 未找到相关的求婚记录"))
	}

	// 开启事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_tx_begin_failed|❌ 开启事务失败"))
	}
	defer tx.Rollback()

	// 更新求婚记录状态
	_, err = tx.Exec("UPDATE marriage_proposal SET status = 'accepted', updated_at = CURRENT_TIMESTAMP WHERE id = $1", proposalID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_update_proposal_failed|❌ 更新求婚记录失败"))
	}

	// 更新求婚者状态
	_, err = tx.Exec("INSERT INTO user_marriage (user_id, spouse_id, status, marriage_date) VALUES ($1, $2, 'married', CURRENT_TIMESTAMP) ON CONFLICT (user_id) DO UPDATE SET spouse_id = $2, status = 'married', marriage_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP", proposerID, recipientID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_update_proposer_failed|❌ 更新求婚者状态失败"))
	}

	// 更新被求婚者状态
	_, err = tx.Exec("INSERT INTO user_marriage (user_id, spouse_id, status, marriage_date) VALUES ($1, $2, 'married', CURRENT_TIMESTAMP) ON CONFLICT (user_id) DO UPDATE SET spouse_id = $2, status = 'married', marriage_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP", recipientID, proposerID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_update_recipient_failed|❌ 更新被求婚者状态失败"))
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_tx_commit_failed|❌ 提交事务失败"))
	}

	return fmt.Sprintf(common.T("", "marriage_marry_success|🎊 恭喜 %s 和 %s 正式结为夫妻！愿你们百年好合，永结同心！"), proposerID, recipientID), nil
}

func (p *MarriagePlugin) doDivorce(userID string) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf(common.T("", "marriage_db_conn_failed|❌ 数据库连接失败"))
	}

	// 检查当前状态
	var status string
	var spouseID string
	err := GlobalDB.QueryRow("SELECT status, spouse_id FROM user_marriage WHERE user_id = $1", userID).Scan(&status, &spouseID)
	if err != nil || status != "married" {
		return "", fmt.Errorf(common.T("", "marriage_not_married|❌ 您当前未处于婚姻状态"))
	}

	// 开启事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_tx_begin_failed|❌ 开启事务失败"))
	}
	defer tx.Rollback()

	// 更新自己的状态
	_, err = tx.Exec("UPDATE user_marriage SET status = 'divorced', spouse_id = '', divorce_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_update_self_failed|❌ 更新个人状态失败"))
	}

	// 更新对方的状态
	_, err = tx.Exec("UPDATE user_marriage SET status = 'divorced', spouse_id = '', divorce_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", spouseID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_update_spouse_failed|❌ 更新对方状态失败"))
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf(common.T("", "marriage_tx_commit_failed|❌ 提交事务失败"))
	}

	return common.T("", "marriage_divorce_success|💔 离婚手续办理成功。愿你们各自安好"), nil
}

func (p *MarriagePlugin) doMyMarriageCertificate(userID string) (string, error) {
	return fmt.Sprintf(common.T("", "marriage_cert_info|📜 结婚证信息\n登记日期：%s\n配偶：%s"), "2023-10-01", "张三"), nil
}

func (p *MarriagePlugin) doSendSweets(userID string, countStr string) (string, error) {
	return common.T("", "marriage_send_sweets_success|🍬 喜糖发放成功！祝你们甜甜蜜蜜"), nil
}

func (p *MarriagePlugin) doEatSweets(userID string) (string, error) {
	return common.T("", "marriage_eat_sweets_success|🍭 您吃到了喜糖，感觉甜滋滋的"), nil
}

func (p *MarriagePlugin) doApplyMarriageCertificate(userID, spouseID string) (string, error) {
	return common.T("", "marriage_apply_cert_success|📜 结婚证办理成功"), nil
}

func (p *MarriagePlugin) doApplyDivorceCertificate(userID string) (string, error) {
	return common.T("", "marriage_apply_divorce_cert_success|📜 离婚证办理成功"), nil
}

func (p *MarriagePlugin) doSpouseSignIn(userID string) (string, error) {
	return common.T("", "marriage_spouse_signin_success|📅 您的另一半已成功签到"), nil
}

func (p *MarriagePlugin) doSpouseGrabFloor(userID string) (string, error) {
	return common.T("", "marriage_spouse_grab_floor_success|🏢 您的另一半成功抢到了楼层"), nil
}

func (p *MarriagePlugin) doSpouseGrabRedPacket(userID string) (string, error) {
	return common.T("", "marriage_spouse_grab_red_packet_success|🧧 您的另一半成功抢到了红包"), nil
}

func (p *MarriagePlugin) doMySpouse(userID string) (string, error) {
	return fmt.Sprintf(common.T("", "marriage_spouse_info|👤 我的对象信息\n昵称：%s\n相识日期：%s"), "张三", "2023-10-01"), nil
}

func (p *MarriagePlugin) doMySweets(userID string) (string, error) {
	return fmt.Sprintf(common.T("", "marriage_my_sweets|🍬 我的喜糖数量：%d"), 10), nil
}

func (p *MarriagePlugin) doMyRedPackets(userID string) (string, error) {
	return fmt.Sprintf(common.T("", "marriage_my_red_packets|🧧 我的红包数量：%d"), 5), nil
}

func (p *MarriagePlugin) doMySweetHearts(userID string) (string, error) {
	return fmt.Sprintf(common.T("", "marriage_my_sweet_hearts|❤️ 我的甜蜜爱心数量：%d"), 20), nil
}

func (p *MarriagePlugin) doSweetHeartsInfo() (string, error) {
	return common.T("", "marriage_sweet_hearts_info|ℹ️ 甜蜜爱心是你们爱情的见证，可以通过日常互动获得"), nil
}

func (p *MarriagePlugin) doSendSweetHeart(userID, recipientID string) (string, error) {
	return common.T("", "marriage_send_sweet_heart_success|💖 甜蜜爱心赠送成功"), nil
}

func (p *MarriagePlugin) doUseSweetLottery(userID string) (string, error) {
	return common.T("", "marriage_lottery_success|🎰 甜蜜抽奖成功！恭喜您获得奖励"), nil
}

func (p *MarriagePlugin) doClaimMarriageBenefits(userID string) (string, error) {
	return common.T("", "marriage_claim_benefits_success|🎁 领取结婚福利成功！恭喜您获得奖励"), nil
}
