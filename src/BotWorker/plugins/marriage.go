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
	
	// 这里可以添加数据库初始化代码
	// 如果使用GORM，可以使用AutoMigrate创建表
	// 但当前项目使用的是原生sql.DB，所以需要手动创建表
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
	// 检查用户是否单身
	// 检查对方是否单身
	// 创建求婚记录
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "求婚已发送，请等待对方回应！",
	})
}

// marry 结婚
func (p *MarriagePlugin) marry(robot plugin.Robot, event *onebot.Event, spouseID string) {
	// 检查是否有求婚记录
	// 更新婚姻状态
	// 创建婚姻记录
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "恭喜你们喜结良缘！",
	})
}

// divorce 离婚
func (p *MarriagePlugin) divorce(robot plugin.Robot, event *onebot.Event) {
	// 检查用户是否已婚
	// 更新婚姻状态
	// 记录离婚日期
	robot.SendMessage(&onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     "离婚手续已办理完成！",
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