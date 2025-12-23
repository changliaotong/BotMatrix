package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
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
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProposerID   string    `gorm:"size:20;index" json:"proposer_id"`
	RecipientID  string    `gorm:"size:20;index" json:"recipient_id"`
	Status       string    `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	return "结婚系统插件，提供求婚、结婚、离婚、喜糖、红包等功能"
}

func (p *MarriagePlugin) Version() string {
	return "1.0.0"
}

func (p *MarriagePlugin) Init(robot plugin.Robot) {
	log.Println("加载结婚系统插件")

	// 初始化数据库
	p.initDatabase()

	// 处理结婚系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		// 购买婚纱
		if match, _ := p.cmdParser.MatchCommand("购买婚纱", event.RawMessage); match {
			p.buyWeddingDress(robot, event)
			return nil
		}

		// 购买婚戒
		if match, _ := p.cmdParser.MatchCommand("购买婚戒", event.RawMessage); match {
			p.buyWeddingRing(robot, event)
			return nil
		}

		// 求婚
		if match, params := p.cmdParser.MatchCommandWithParams("求婚(\d+)", event.RawMessage); match && len(params) > 0 {
			spouseID := params[1]
			p.proposeMarriage(robot, event, spouseID)
			return nil
		}

		// 结婚
		if match, params := p.cmdParser.MatchCommandWithParams("结婚(\d+)", event.RawMessage); match && len(params) > 0 {
			spouseID := params[1]
			p.marry(robot, event, spouseID)
			return nil
		}

		// 离婚
		if match, _ := p.cmdParser.MatchCommand("离婚", event.RawMessage); match {
			p.divorce(robot, event)
			return nil
		}

		// 我的结婚证
		if match, _ := p.cmdParser.MatchCommand("我的结婚证", event.RawMessage); match {
			p.myMarriageCertificate(robot, event)
			return nil
		}

		// 发喜糖
		if match, params := p.cmdParser.MatchCommandWithParams("发喜糖(\d+)", event.RawMessage); match && len(params) > 0 {
			count := params[1]
			p.sendSweets(robot, event, count)
			return nil
		}

		// 吃喜糖
		if match, _ := p.cmdParser.MatchCommand("吃喜糖", event.RawMessage); match {
			p.eatSweets(robot, event)
			return nil
		}

		// 办理结婚证
		if match, params := p.cmdParser.MatchCommandWithParams("办理结婚证(\d+)", event.RawMessage); match && len(params) > 0 {
			spouseID := params[1]
			p.applyMarriageCertificate(robot, event, spouseID)
			return nil
		}

		// 办理离婚证
		if match, _ := p.cmdParser.MatchCommand("办理离婚证", event.RawMessage); match {
			p.applyDivorceCertificate(robot, event)
			return nil
		}

		// 另一半签到
		if match, _ := p.cmdParser.MatchCommand("另一半签到", event.RawMessage); match {
			p.spouseSignIn(robot, event)
			return nil
		}

		// 另一半抢楼
		if match, _ := p.cmdParser.MatchCommand("另一半抢楼", event.RawMessage); match {
			p.spouseGrabFloor(robot, event)
			return nil
		}

		// 另一半抢红包
		if match, _ := p.cmdParser.MatchCommand("另一半抢红包", event.RawMessage); match {
			p.spouseGrabRedPacket(robot, event)
			return nil
		}

		// 我的对象
		if match, _ := p.cmdParser.MatchCommand("我的对象", event.RawMessage); match {
			p.mySpouse(robot, event)
			return nil
		}

		// 我的喜糖
		if match, _ := p.cmdParser.MatchCommand("我的喜糖", event.RawMessage); match {
			p.mySweets(robot, event)
			return nil
		}

		// 我的红包
		if match, _ := p.cmdParser.MatchCommand("我的红包", event.RawMessage); match {
			p.myRedPackets(robot, event)
			return nil
		}

		// 我的甜蜜爱心
		if match, _ := p.cmdParser.MatchCommand("我的甜蜜爱心", event.RawMessage); match {
			p.mySweetHearts(robot, event)
			return nil
		}

		// 甜蜜爱心说明
		if match, _ := p.cmdParser.MatchCommand("甜蜜爱心说明", event.RawMessage); match {
			p.sweetHeartsInfo(robot, event)
			return nil
		}

		// 赠送甜蜜爱心
		if match, params := p.cmdParser.MatchCommandWithParams("赠送甜蜜爱心(\d+)", event.RawMessage); match && len(params) > 0 {
			recipientID := params[1]
			p.sendSweetHeart(robot, event, recipientID)
			return nil
		}

		// 使用甜蜜抽奖
		if match, _ := p.cmdParser.MatchCommand("使用甜蜜抽奖", event.RawMessage); match {
			p.useSweetLottery(robot, event)
			return nil
		}

		// 领取结婚福利
		if match, _ := p.cmdParser.MatchCommand("领取结婚福利", event.RawMessage); match {
			p.claimMarriageBenefits(robot, event)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
func (p *MarriagePlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，结婚系统将使用模拟数据")
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
	
	log.Println("结婚系统数据库初始化完成")
}

// isSystemEnabled 检查结婚系统是否开启
func (p *MarriagePlugin) isSystemEnabled() bool {
	// 检查全局数据库连接
	if GlobalDB == nil {
		// 如果没有数据库连接，默认返回开启状态
		return true
	}
	
	// 这里可以添加从数据库获取系统配置的代码
	// 例如：SELECT is_enabled FROM marriage_config WHERE id = 1
	// 现在默认返回开启状态
	return true
}

// buyWeddingDress 购买婚纱
func (p *MarriagePlugin) buyWeddingDress(robot plugin.Robot, event *onebot.Event) {
	// 检查用户积分
	// 扣除积分
	// 记录购买的婚纱
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "婚纱购买成功！",
	})
}

// buyWeddingRing 购买婚戒
func (p *MarriagePlugin) buyWeddingRing(robot plugin.Robot, event *onebot.Event) {
	// 检查用户积分
	// 扣除积分
	// 记录购买的婚戒
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "婚戒购买成功！",
	})
}

// proposeMarriage 求婚
func (p *MarriagePlugin) proposeMarriage(robot plugin.Robot, event *onebot.Event, spouseID string) {
	// 检查全局数据库连接
	if GlobalDB == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 数据库连接失败，请稍后重试",
		})
		return
	}
	
	// 检查自己是否单身
	var myStatus string
	row := GlobalDB.QueryRow("SELECT status FROM user_marriage WHERE user_id = $1", event.UserID)
	err := row.Scan(&myStatus)
	if err != nil {
		// 如果没有记录，默认是单身
		myStatus = "single"
	}
	
	if myStatus != "single" {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 您当前不是单身状态，无法求婚",
		})
		return
	}
	
	// 检查对方是否单身
	var targetStatus string
	row = GlobalDB.QueryRow("SELECT status FROM user_marriage WHERE user_id = $1", spouseID)
	err = row.Scan(&targetStatus)
	if err != nil {
		// 如果没有记录，默认是单身
		targetStatus = "single"
	}
	
	if targetStatus != "single" {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 对方当前不是单身状态，无法求婚",
		})
		return
	}
	
	// 检查是否已经有未处理的求婚记录
	var proposalCount int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM marriage_proposal WHERE proposer_id = $1 AND recipient_id = $2 AND status = 'pending'", event.UserID, spouseID).Scan(&proposalCount)
	if err != nil {
		log.Printf("查询求婚记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 查询求婚记录失败，请稍后重试",
		})
		return
	}
	
	if proposalCount > 0 {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 您已经向对方发送过求婚请求，请等待对方回应",
		})
		return
	}
	
	// 创建求婚记录
	_, err = GlobalDB.Exec("INSERT INTO marriage_proposal (proposer_id, recipient_id, status) VALUES ($1, $2, 'pending')", event.UserID, spouseID)
	if err != nil {
		log.Printf("创建求婚记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 发送求婚请求失败，请稍后重试",
		})
		return
	}
	
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "💍 求婚请求已发送，请等待对方回应！",
	})
}

// marry 结婚
func (p *MarriagePlugin) marry(robot plugin.Robot, event *onebot.Event, spouseID string) {
	// 检查全局数据库连接
	if GlobalDB == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 数据库连接失败，请稍后重试",
		})
		return
	}
	
	// 检查是否有对方的求婚记录
	var proposalID int
	row := GlobalDB.QueryRow("SELECT id FROM marriage_proposal WHERE proposer_id = $1 AND recipient_id = $2 AND status = 'pending'", spouseID, event.UserID)
	err := row.Scan(&proposalID)
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 未找到对方的求婚记录，请确认对方已向您求婚",
		})
		return
	}
	
	// 开始事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf("开启事务失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 更新求婚记录状态为已接受
	_, err = tx.Exec("UPDATE marriage_proposal SET status = 'accepted' WHERE id = $1", proposalID)
	if err != nil {
		tx.Rollback()
		log.Printf("更新求婚记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 处理自己的婚姻记录
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM user_marriage WHERE user_id = $1", event.UserID).Scan(&count)
	if err != nil {
		tx.Rollback()
		log.Printf("查询婚姻记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	if count > 0 {
		// 更新现有记录
		_, err = tx.Exec("UPDATE user_marriage SET spouse_id = $1, status = 'married', marriage_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", spouseID, event.UserID)
	} else {
		// 创建新记录
		_, err = tx.Exec("INSERT INTO user_marriage (user_id, spouse_id, status, marriage_date) VALUES ($1, $2, 'married', CURRENT_TIMESTAMP)", event.UserID, spouseID)
	}
	
	if err != nil {
		tx.Rollback()
		log.Printf("更新自己婚姻记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 处理对方的婚姻记录
	err = tx.QueryRow("SELECT COUNT(*) FROM user_marriage WHERE user_id = $1", spouseID).Scan(&count)
	if err != nil {
		tx.Rollback()
		log.Printf("查询对方婚姻记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	if count > 0 {
		// 更新现有记录
		_, err = tx.Exec("UPDATE user_marriage SET spouse_id = $1, status = 'married', marriage_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2", event.UserID, spouseID)
	} else {
		// 创建新记录
		_, err = tx.Exec("INSERT INTO user_marriage (user_id, spouse_id, status, marriage_date) VALUES ($1, $2, 'married', CURRENT_TIMESTAMP)", spouseID, event.UserID)
	}
	
	if err != nil {
		tx.Rollback()
		log.Printf("更新对方婚姻记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf("提交事务失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 发放婚姻伴侣徽章给双方用户
	badgePlugin := GetBadgePluginInstance()
	// 给当前用户发放徽章
	err = badgePlugin.GrantBadgeToUser(event.UserID, "婚姻伴侣", "system", "成功结婚")
	if err != nil {
		log.Printf("给用户 %s 发放婚姻伴侣徽章失败: %v\n", event.UserID, err)
	} else {
		log.Printf("给用户 %s 成功发放婚姻伴侣徽章\n", event.UserID)
	}
	// 给配偶用户发放徽章
	err = badgePlugin.GrantBadgeToUser(spouseID, "婚姻伴侣", "system", "成功结婚")
	if err != nil {
		log.Printf("给用户 %s 发放婚姻伴侣徽章失败: %v\n", spouseID, err)
	} else {
		log.Printf("给用户 %s 成功发放婚姻伴侣徽章\n", spouseID)
	}
	
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "🎎 恭喜你们喜结良缘！祝你们百年好合，永结同心！\n💍 你们已获得【婚姻伴侣】徽章！",
	})
}

// divorce 离婚
func (p *MarriagePlugin) divorce(robot plugin.Robot, event *onebot.Event) {
	// 检查全局数据库连接
	if GlobalDB == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 数据库连接失败，请稍后重试",
		})
		return
	}
	
	// 检查用户是否已婚
	var spouseID string
	var status string
	row := GlobalDB.QueryRow("SELECT spouse_id, status FROM user_marriage WHERE user_id = $1", event.UserID)
	err := row.Scan(&spouseID, &status)
	if err != nil || status != "married" {
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 您当前不是已婚状态，无法办理离婚",
		})
		return
	}
	
	// 开始事务
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf("开启事务失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 更新自己的婚姻状态
	_, err = tx.Exec("UPDATE user_marriage SET status = 'divorced', divorce_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", event.UserID)
	if err != nil {
		tx.Rollback()
		log.Printf("更新自己离婚记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 更新对方的婚姻状态
	_, err = tx.Exec("UPDATE user_marriage SET status = 'divorced', divorce_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", spouseID)
	if err != nil {
		tx.Rollback()
		log.Printf("更新对方离婚记录失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf("提交事务失败: %v\n", err)
		robot.SendMessage(&onebot.SendMessageParams{
			MessageType: event.MessageType,
			UserID:      event.UserID,
			GroupID:     event.GroupID,
			Message:     "❌ 系统错误，请稍后重试",
		})
		return
	}
	
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "📝 离婚手续已办理完成，祝您未来生活幸福！",
	})
}

// myMarriageCertificate 我的结婚证
func (p *MarriagePlugin) myMarriageCertificate(robot plugin.Robot, event *onebot.Event) {
	// 查询用户婚姻信息
	// 返回结婚证信息
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "您的结婚证信息：\n婚姻状态：已婚\n结婚日期：2023-10-01\n配偶：张三",
	})
}

// sendSweets 发喜糖
func (p *MarriagePlugin) sendSweets(robot plugin.Robot, event *onebot.Event, count string) {
	// 检查用户喜糖数量
	// 扣除喜糖
	// 发送喜糖给群成员
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "喜糖已发送！",
	})
}

// eatSweets 吃喜糖
func (p *MarriagePlugin) eatSweets(robot plugin.Robot, event *onebot.Event) {
	// 随机获得积分或其他奖励
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "恭喜你获得了10个积分！",
	})
}

// applyMarriageCertificate 办理结婚证
func (p *MarriagePlugin) applyMarriageCertificate(robot plugin.Robot, event *onebot.Event, spouseID string) {
	// 检查求婚记录
	// 更新婚姻状态
	// 创建婚姻记录
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "结婚证办理成功！",
	})
}

// applyDivorceCertificate 办理离婚证
func (p *MarriagePlugin) applyDivorceCertificate(robot plugin.Robot, event *onebot.Event) {
	// 检查婚姻状态
	// 更新婚姻状态
	// 记录离婚日期
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "离婚证办理成功！",
	})
}

// spouseSignIn 另一半签到
func (p *MarriagePlugin) spouseSignIn(robot plugin.Robot, event *onebot.Event) {
	// 检查婚姻状态
	// 为配偶添加积分或其他奖励
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "另一半签到成功，为配偶获得了5个积分！",
	})
}

// spouseGrabFloor 另一半抢楼
func (p *MarriagePlugin) spouseGrabFloor(robot plugin.Robot, event *onebot.Event) {
	// 检查婚姻状态
	// 为配偶添加积分或其他奖励
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "另一半抢楼成功，为配偶获得了10个积分！",
	})
}

// spouseGrabRedPacket 另一半抢红包
func (p *MarriagePlugin) spouseGrabRedPacket(robot plugin.Robot, event *onebot.Event) {
	// 检查婚姻状态
	// 为配偶添加积分或其他奖励
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "另一半抢红包成功，为配偶获得了15个积分！",
	})
}

// mySpouse 我的对象
func (p *MarriagePlugin) mySpouse(robot plugin.Robot, event *onebot.Event) {
	// 查询用户婚姻信息
	// 返回配偶信息
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "您的配偶是：张三",
	})
}

// mySweets 我的喜糖
func (p *MarriagePlugin) mySweets(robot plugin.Robot, event *onebot.Event) {
	// 查询用户喜糖数量
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "您当前有10个喜糖！",
	})
}

// myRedPackets 我的红包
func (p *MarriagePlugin) myRedPackets(robot plugin.Robot, event *onebot.Event) {
	// 查询用户红包数量
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "您当前有5个红包！",
	})
}

// mySweetHearts 我的甜蜜爱心
func (p *MarriagePlugin) mySweetHearts(robot plugin.Robot, event *onebot.Event) {
	// 查询用户甜蜜爱心数量
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "您当前有20个甜蜜爱心！",
	})
}

// sweetHeartsInfo 甜蜜爱心说明
func (p *MarriagePlugin) sweetHeartsInfo(robot plugin.Robot, event *onebot.Event) {
	// 返回甜蜜爱心的说明
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "甜蜜爱心是结婚系统的虚拟货币，可以通过签到、抢楼、抢红包等方式获得，用于抽奖和购买特殊物品！",
	})
}

// sendSweetHeart 赠送甜蜜爱心
func (p *MarriagePlugin) sendSweetHeart(robot plugin.Robot, event *onebot.Event, recipientID string) {
	// 检查用户甜蜜爱心数量
	// 扣除甜蜜爱心
	// 增加接收者的甜蜜爱心
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "甜蜜爱心赠送成功！",
	})
}

// useSweetLottery 使用甜蜜抽奖
func (p *MarriagePlugin) useSweetLottery(robot plugin.Robot, event *onebot.Event) {
	// 检查用户甜蜜爱心数量
	// 扣除甜蜜爱心
	// 随机获得奖励
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "抽奖成功！您获得了50个积分！",
	})
}

// claimMarriageBenefits 领取结婚福利
func (p *MarriagePlugin) claimMarriageBenefits(robot plugin.Robot, event *onebot.Event) {
	// 检查婚姻状态
	// 发放结婚福利
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "结婚福利领取成功！您获得了100个积分和5个甜蜜爱心！",
	})
}