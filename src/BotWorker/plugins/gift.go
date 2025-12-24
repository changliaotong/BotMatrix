package plugins

import (
	"BotMatrix/common"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// GiftPlugin 送礼物插件
type GiftPlugin struct {
	db        *sql.DB
	cmdParser *CommandParser
}

// GiftItem 礼物物品
type GiftItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Emoji    string `json:"emoji"`
	Rarity   string `json:"rarity"`
	DropRate float64 `json:"drop_rate"`
}

// NewGiftPlugin 创建送礼物插件实例
func NewGiftPlugin(database *sql.DB) *GiftPlugin {
	return &GiftPlugin{
		db:        database,
		cmdParser: NewCommandParser(),
	}
}

func (p *GiftPlugin) Name() string {
	return "gift"
}

func (p *GiftPlugin) Description() string {
	return common.T("", "gift_plugin_desc|送礼物插件，可以消耗积分给他人送礼，也可以抽奖获得礼物")
}

func (p *GiftPlugin) Version() string {
	return "1.0.0"
}

func (p *GiftPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "gift_db_not_configured|礼物插件初始化失败：数据库未连接"))
		return
	}
	log.Println(common.T("", "gift_plugin_loaded|礼物系统插件已加载"))

	// 报备技能
	robot.HandleSkill("send_gift", func(params map[string]string) (string, error) {
		toUserID := params["to_user_id"]
		giftName := params["gift_name"]
		if toUserID == "" || giftName == "" {
			return "", fmt.Errorf(common.T("", "gift_param_error|❌ 参数错误，请输入正确的用户ID和礼物名称"))
		}
		err := p.sendGift(robot, nil, toUserID, giftName)
		return "", err
	})
	robot.HandleSkill("gift_list", func(params map[string]string) (string, error) {
		err := p.showGiftList(robot, nil)
		return "", err
	})
	robot.HandleSkill("draw_gift", func(params map[string]string) (string, error) {
		err := p.drawGiftLogic(robot, nil)
		return "", err
	})

	// 处理送礼物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "gift") {
				HandleFeatureDisabled(robot, event, "gift")
				return nil
			}
		}

		// 检查是否为送礼物命令
		match, cmd, params := p.cmdParser.MatchCommandWithParams("送礼物|gift|礼物", "(\\d+)\\s+(\\w+)", event.RawMessage)
		if !match || len(params) != 2 {
			if match {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "gift_cmd_usage|%s 命令用法：%s <用户ID> <礼物名称>"), cmd, cmd))
				p.sendMessage(robot, event, common.T("", "gift_available_list|可送礼物：鲜花、蛋糕、巧克力、钻石"))
			}
			return nil
		}

		return p.sendGift(robot, event, params[0], params[1])
	})

	// 处理礼物列表命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为礼物列表命令
		if match, _ := p.cmdParser.MatchCommand("礼物列表|giftlist", event.RawMessage); match {
			return p.showGiftList(robot, event)
		}

		return nil
	})

	// 处理抽礼物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "gift") {
				HandleFeatureDisabled(robot, event, "gift")
				return nil
			}
		}

		// 检查是否为抽礼物命令
		if match, _ := p.cmdParser.MatchCommand("抽礼物|drawgift|抽奖", event.RawMessage); match {
			return p.drawGiftLogic(robot, event)
		}

		return nil
	})
}

// GetSkills 实现 SkillCapable 接口
func (p *GiftPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability {
		{
			Name:        "send_gift",
			Description: common.T("", "gift_skill_send_desc|给指定用户赠送礼物"),
			Usage:       common.T("", "gift_skill_send_usage|send_gift to_user_id=123456 gift_name=鲜花"),
			Params: map[string]string{
				"to_user_id": common.T("", "gift_skill_send_param_to_user_id|接收礼物的用户ID"),
				"gift_name":  common.T("", "gift_skill_send_param_gift_name|礼物名称"),
			},
		},
		{
			Name:        "gift_list",
			Description: common.T("", "gift_skill_list_desc|查看所有可用礼物及价格"),
			Usage:       common.T("", "gift_skill_list_usage|gift_list"),
		},
		{
			Name:        "draw_gift",
			Description: common.T("", "gift_skill_draw_desc|随机抽取一份礼物"),
			Usage:       common.T("", "gift_skill_draw_usage|draw_gift"),
		},
	}
}

func (p *GiftPlugin) sendGift(robot plugin.Robot, event *onebot.Event, toUserID string, giftName string) error {
	// 获取礼物价格
	giftPrice := p.getGiftPrice(giftName)
	if giftPrice == 0 {
		if event != nil {
			p.sendMessage(robot, event, common.T("", "gift_invalid_name|❌ 礼物名称无效，请输入正确的礼物名称"))
		}
		return nil
	}

	// 获取操作者ID
	var fromUserID int64
	if event != nil {
		fromUserID = event.UserID
	}
	fromUserIDStr := fmt.Sprintf("%d", fromUserID)

	if fromUserIDStr == toUserID {
		if event != nil {
			p.sendMessage(robot, event, common.T("", "gift_send_self|❌ 不能给自己送礼物哦"))
		}
		return nil
	}

	// 检查积分是否足够
	fromUserPoints, err := db.GetPoints(p.db, fromUserIDStr)
	if err != nil {
		log.Printf("获取积分失败: %v", err)
		if event != nil {
			p.sendMessage(robot, event, common.T("", "gift_query_points_failed|❌ 查询积分失败，请稍后再试"))
		}
		return nil
	}

	if fromUserPoints < giftPrice {
		if event != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "gift_insufficient_points|❌ 积分不足，送出该礼物需要 %d 积分，你当前只有 %d 积分"), giftPrice, fromUserPoints))
		}
		return nil
	}

	// 执行送礼物操作
	reason := fmt.Sprintf(common.T("", "gift_send_reason|送礼物给他人：%s"), giftName)
	err = db.TransferPoints(p.db, fromUserIDStr, toUserID, giftPrice, reason, "gift")
	if err != nil {
		if event != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "gift_send_failed|❌ 送礼物失败：%v"), err))
		}
		return nil
	}

	// 发送成功消息
	giftEmoji := p.getGiftEmoji(giftName)
	if event != nil {
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "gift_send_success|✅ 成功送出礼物 %s 给用户 %s (%s)"), giftEmoji, toUserID, giftName))
	}
	return nil
}

func (p *GiftPlugin) showGiftList(robot plugin.Robot, event *onebot.Event) error {
	if event != nil {
		p.sendMessage(robot, event, p.getGiftList())
	}
	return nil
}

func (p *GiftPlugin) drawGiftLogic(robot plugin.Robot, event *onebot.Event) error {
	// 抽取礼物
	gift := p.drawGift()

	// 发送抽礼物结果
	message := fmt.Sprintf(common.T("", "gift_draw_result|🎁 抽奖结果：恭喜你抽中了 %s %s (%s)！"), gift.Emoji, gift.Name, gift.Name)

	if event != nil {
		p.sendMessage(robot, event, message)
	}
	return nil
}

// getGiftPrice 获取礼物价格
func (p *GiftPlugin) getGiftPrice(giftName string) int {
	giftName = strings.ToLower(giftName)
	giftPrices := map[string]int{
		common.T("", "gift_flower|鲜花"):    5,
		"flower":                       5,
		common.T("", "gift_cake|蛋糕"):      10,
		"cake":                         10,
		common.T("", "gift_chocolate|巧克力"): 15,
		"chocolate":                    15,
		common.T("", "gift_diamond|钻石"):   50,
		"diamond":                      50,
	}
	return giftPrices[giftName]
}

// getGiftEmoji 获取礼物对应的表情
func (p *GiftPlugin) getGiftEmoji(giftName string) string {
	giftName = strings.ToLower(giftName)
	giftEmojis := map[string]string{
		common.T("", "gift_flower|鲜花"):    "🌸",
		"flower":                       "🌸",
		common.T("", "gift_cake|蛋糕"):      "🍰",
		"cake":                         "🍰",
		common.T("", "gift_chocolate|巧克力"): "🍫",
		"chocolate":                    "🍫",
		common.T("", "gift_diamond|钻石"):   "💎",
		"diamond":                      "💎",
	}
	return giftEmojis[giftName]
}

// getGiftList 获取礼物列表
func (p *GiftPlugin) getGiftList() string {
	list := common.T("", "gift_list_title|🎁 可用礼物列表：\n")
	list += fmt.Sprintf(common.T("", "gift_list_item|%s %s - 价格：%d 积分\n"), "🌸", common.T("", "gift_flower|鲜花"), 5)
	list += fmt.Sprintf(common.T("", "gift_list_item|%s %s - 价格：%d 积分\n"), "🍰", common.T("", "gift_cake|蛋糕"), 10)
	list += fmt.Sprintf(common.T("", "gift_list_item|%s %s - 价格：%d 积分\n"), "🍫", common.T("", "gift_chocolate|巧克力"), 15)
	list += fmt.Sprintf(common.T("", "gift_list_item|%s %s - 价格：%d 积分\n"), "💎", common.T("", "gift_diamond|钻石"), 50)
	list += common.T("", "gift_list_footer|\n使用方法：送礼物 <用户ID> <礼物名称>")
	return list
}

// getGiftPool 获取抽奖礼物池
func (p *GiftPlugin) getGiftPool() []GiftItem {
	return []GiftItem{
		{ID: 1, Name: common.T("", "gift_flower|鲜花"), Price: 5, Emoji: "🌸", Rarity: "common", DropRate: 0.5},
		{ID: 2, Name: common.T("", "gift_cake|蛋糕"), Price: 10, Emoji: "🍰", Rarity: "uncommon", DropRate: 0.3},
		{ID: 3, Name: common.T("", "gift_chocolate|巧克力"), Price: 15, Emoji: "🍫", Rarity: "rare", DropRate: 0.15},
		{ID: 4, Name: common.T("", "gift_diamond|钻石"), Price: 50, Emoji: "💎", Rarity: "epic", DropRate: 0.05},
	}
}

// drawGift 抽取礼物
func (p *GiftPlugin) drawGift() *GiftItem {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 获取礼物池
	giftPool := p.getGiftPool()

	// 生成随机数
	randomValue := rand.Float64()

	// 根据掉落率选择礼物
	cumulativeRate := 0.0
	for _, gift := range giftPool {
		cumulativeRate += gift.DropRate
		if randomValue <= cumulativeRate {
			return &gift
		}
	}

	// 默认返回第一个礼物
	return &giftPool[0]
}

// sendMessage 发送消息
func (p *GiftPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
	params := &onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     message,
	}
	_, err := robot.SendMessage(params)
	if err != nil {
		log.Printf("发送消息失败: %v", err)
	}
}