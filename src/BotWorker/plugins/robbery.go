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
	"time"
)

// RobberyPlugin 打劫系统插件
type RobberyPlugin struct {
	db *sql.DB
	// 存储用户上次打劫时间，key为用户ID，value为打劫时间
	lastRobberyTime map[string]time.Time
	// 命令解析器
	cmdParser *CommandParser
}

// NewRobberyPlugin 创建打劫系统插件实例
func NewRobberyPlugin(database *sql.DB) *RobberyPlugin {
	rand.Seed(time.Now().UnixNano())
	return &RobberyPlugin{
		db:              database,
		lastRobberyTime: make(map[string]time.Time),
		cmdParser:       NewCommandParser(),
	}
}

func (p *RobberyPlugin) Name() string {
	return "robbery"
}

func (p *RobberyPlugin) Description() string {
	return "打劫系统插件，支持用户之间的积分抢夺功能"
}

func (p *RobberyPlugin) Version() string {
	return "1.0.0"
}

func (p *RobberyPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("打劫系统插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载打劫系统插件")

	// 处理打劫命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "robbery") {
				HandleFeatureDisabled(robot, event, "robbery")
				return nil
			}
		}

		// 检查是否为打劫命令
		match, cmd, params := p.cmdParser.MatchCommandWithParams("打劫|抢劫", "(\d+)", event.RawMessage)
		if !match {
			return nil
		}

		if len(params) != 1 {
			p.sendMessage(robot, event, fmt.Sprintf("%s命令格式：%s <用户ID>", cmd, cmd))
			return nil
		}

		// 解析被打劫用户ID
		victimUserIDStr := params[0]
		robberUserID := event.UserID
		robberUserIDStr := fmt.Sprintf("%d", robberUserID)

		// 不能打劫自己
		if robberUserIDStr == victimUserIDStr {
			p.sendMessage(robot, event, "不能打劫自己哦")
			return nil
		}

		// 检查打劫冷却时间（每小时只能打劫一次）
		now := time.Now()
		if lastRobbery, ok := p.lastRobberyTime[robberUserIDStr]; ok {
			if now.Sub(lastRobbery) < 1*time.Hour {
				remainingTime := 1*time.Hour - now.Sub(lastRobbery)
				p.sendMessage(robot, event, fmt.Sprintf("打劫冷却中，还需等待 %.0f 分钟才能再次打劫", remainingTime.Minutes()))
				return nil
			}
		}

		// 获取被打劫用户的积分
		victimPoints, err := db.GetPoints(p.db, victimUserIDStr)
		if err != nil {
			p.sendMessage(robot, event, "查询被打劫用户信息失败")
			return nil
		}

		// 检查被打劫用户积分是否足够
		if victimPoints < 10 {
			p.sendMessage(robot, event, "这个用户太穷了，不值得打劫")
			return nil
		}

		// 计算可打劫的积分范围（10% - 30%，最少10积分，最多100积分）
		robberyPercentage := 0.1 + rand.Float64()*0.2
		robberyAmount := int(float64(victimPoints) * robberyPercentage)
		if robberyAmount < 10 {
			robberyAmount = 10
		}
		if robberyAmount > 100 {
			robberyAmount = 100
		}

		// 执行打劫（使用数据库事务）
		err = db.TransferPoints(p.db, victimUserIDStr, robberUserIDStr, robberyAmount, "被打劫", "robbery")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("打劫失败: %v", err))
			return nil
		}

		// 更新打劫时间
		p.lastRobberyTime[robberUserIDStr] = now

		// 发送成功消息
		p.sendMessage(robot, event, fmt.Sprintf("✅ 打劫成功！从用户 %s 处抢到了 %d 积分", victimUserIDStr, robberyAmount))
		return nil
	})

	// 处理查看打劫冷却时间命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "robbery") {
				return nil
			}
		}

		// 检查是否为查看打劫冷却时间命令
		if match, _ := p.cmdParser.MatchCommand("打劫冷却", event.RawMessage); !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		now := time.Now()

		if lastRobbery, ok := p.lastRobberyTime[userIDStr]; ok {
			if now.Sub(lastRobbery) < 1*time.Hour {
				remainingTime := 1*time.Hour - now.Sub(lastRobbery)
				p.sendMessage(robot, event, fmt.Sprintf("打劫冷却中，还需等待 %.0f 分钟才能再次打劫", remainingTime.Minutes()))
			} else {
				p.sendMessage(robot, event, "打劫冷却已结束，可以再次打劫")
			}
		} else {
			p.sendMessage(robot, event, "你还没有打劫过，可以随时打劫")
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *RobberyPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}