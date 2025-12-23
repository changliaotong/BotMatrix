package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
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
	return "送礼物插件，支持给其他用户发送虚拟礼物"
}

func (p *GiftPlugin) Version() string {
	return "1.0.0"
}

func (p *GiftPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("送礼物插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载送礼物插件")

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
		match, cmd, params := p.cmdParser.MatchCommandWithParams("送礼物|gift|礼物", "(\d+)\s+(\w+)", event.RawMessage)
		if !match || len(params) != 2 {
			if match {
				p.sendMessage(robot, event, fmt.Sprintf("%s命令格式：%s <用户ID> <礼物名称>", cmd, cmd))
				p.sendMessage(robot, event, "可用礼物：鲜花(5积分)、蛋糕(10积分)、巧克力(15积分)、钻石(50积分)")
			}
			return nil
		}

		// 解析礼物信息
		toUserID := params[0]
		giftName := params[1]

		// 获取礼物价格
		giftPrice := p.getGiftPrice(giftName)
		if giftPrice == 0 {
			p.sendMessage(robot, event, "无效的礼物名称！可用礼物：鲜花(5积分)、蛋糕(10积分)、巧克力(15积分)、钻石(50积分)")
			return nil
		}

		// 获取操作者ID
		fromUserID := event.UserID
		fromUserIDStr := fmt.Sprintf("%d", fromUserID)

		if fromUserIDStr == toUserID {
			p.sendMessage(robot, event, "不能给自己送礼物哦")
			return nil
		}

		// 检查积分是否足够
		fromUserPoints, err := db.GetPoints(p.db, fromUserIDStr)
		if err != nil {
			log.Printf("获取积分失败: %v", err)
			p.sendMessage(robot, event, "查询积分失败，请稍后再试")
			return nil
		}

		if fromUserPoints < giftPrice {
			p.sendMessage(robot, event, fmt.Sprintf("你的积分不足！需要 %d 积分，当前只有 %d 积分", giftPrice, fromUserPoints))
			return nil
		}

		// 执行送礼物操作
		reason := fmt.Sprintf("送礼物：%s", giftName)
		err = db.TransferPoints(p.db, fromUserIDStr, toUserID, giftPrice, reason, "gift")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("送礼物失败: %v", err))
			return nil
		}

		// 发送成功消息
		giftEmoji := p.getGiftEmoji(giftName)
		p.sendMessage(robot, event, fmt.Sprintf("%s 送礼物成功！你给用户 %s 送了一份%s", giftEmoji, toUserID, giftName))
		return nil
	})

	// 处理礼物列表命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为礼物列表命令
		if match, _ := p.cmdParser.MatchCommand("礼物列表|giftlist", event.RawMessage); match {
			p.sendMessage(robot, event, p.getGiftList())
			return nil
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
			// 抽取礼物
			gift := p.drawGift()
			
			// 发送抽礼物结果
			message := fmt.Sprintf("🎉 恭喜你抽到了%s %s！", gift.Emoji, gift.Name)
			message += fmt.Sprintf("\n💡 你可以使用 /送礼物 <用户ID> %s 将礼物送给其他用户", gift.Name)
			
			p.sendMessage(robot, event, message)
			return nil
		}

		return nil
	})
}

// getGiftPrice 获取礼物价格
func (p *GiftPlugin) getGiftPrice(giftName string) int {
	giftName = strings.ToLower(giftName)
	giftPrices := map[string]int{
		"鲜花": 5,
		"flower": 5,
		"蛋糕": 10,
		"cake": 10,
		"巧克力": 15,
		"chocolate": 15,
		"钻石": 50,
		"diamond": 50,
	}
	return giftPrices[giftName]
}

// getGiftEmoji 获取礼物对应的表情
func (p *GiftPlugin) getGiftEmoji(giftName string) string {
	giftName = strings.ToLower(giftName)
	giftEmojis := map[string]string{
		"鲜花": "🌸",
		"flower": "🌸",
		"蛋糕": "🍰",
		"cake": "🍰",
		"巧克力": "🍫",
		"chocolate": "🍫",
		"钻石": "💎",
		"diamond": "💎",
	}
	return giftEmojis[giftName]
}

// getGiftList 获取礼物列表
func (p *GiftPlugin) getGiftList() string {
	list := "🎁 可用礼物列表\n"
	list += "====================\n\n"
	list += "🌸 鲜花 - 5积分\n"
	list += "🍰 蛋糕 - 10积分\n"
	list += "🍫 巧克力 - 15积分\n"
	list += "💎 钻石 - 50积分\n\n"
	list += "💡 使用方法：/送礼物 <用户ID> <礼物名称>"
	return list
}

// getGiftPool 获取抽奖礼物池
func (p *GiftPlugin) getGiftPool() []GiftItem {
	return []GiftItem{
		{ID: 1, Name: "鲜花", Price: 5, Emoji: "🌸", Rarity: "common", DropRate: 0.5},
		{ID: 2, Name: "蛋糕", Price: 10, Emoji: "🍰", Rarity: "uncommon", DropRate: 0.3},
		{ID: 3, Name: "巧克力", Price: 15, Emoji: "🍫", Rarity: "rare", DropRate: 0.15},
		{ID: 4, Name: "钻石", Price: 50, Emoji: "💎", Rarity: "epic", DropRate: 0.05},
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
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}