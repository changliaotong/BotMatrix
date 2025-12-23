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

// FishingPlugin 钓鱼系统插件
type FishingPlugin struct {
	db *sql.DB
	// 存储用户上次钓鱼时间，key为用户ID，value为钓鱼时间
	lastFishingTime map[string]time.Time
	// 存储用户钓鱼技能等级，key为用户ID，value为等级
	fishingLevel map[string]int
	// 命令解析器
	cmdParser *CommandParser
}

// NewFishingPlugin 创建钓鱼系统插件实例
func NewFishingPlugin(database *sql.DB) *FishingPlugin {
	rand.Seed(time.Now().UnixNano())
	return &FishingPlugin{
		db:              database,
		lastFishingTime: make(map[string]time.Time),
		fishingLevel:    make(map[string]int),
		cmdParser:       NewCommandParser(),
	}
}

func (p *FishingPlugin) Name() string {
	return "fishing"
}

func (p *FishingPlugin) Description() string {
	return "钓鱼系统插件，支持钓鱼获取积分和提升钓鱼技能"
}

func (p *FishingPlugin) Version() string {
	return "1.0.0"
}

func (p *FishingPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("钓鱼系统插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载钓鱼系统插件")

	// 处理钓鱼命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "fishing") {
				HandleFeatureDisabled(robot, event, "fishing")
				return nil
			}
		}

		// 检查是否为钓鱼命令
		match, _ := p.cmdParser.MatchCommand("钓鱼", event.RawMessage)
		if !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		now := time.Now()

		// 检查钓鱼冷却时间（每10分钟只能钓鱼一次）
		if lastFishing, ok := p.lastFishingTime[userIDStr]; ok {
			if now.Sub(lastFishing) < 10*time.Minute {
				remainingTime := 10*time.Minute - now.Sub(lastFishing)
				p.sendMessage(robot, event, fmt.Sprintf("钓鱼冷却中，还需等待 %.0f 分钟才能再次钓鱼", remainingTime.Minutes()))
				return nil
			}
		}

		// 获取用户钓鱼等级
		level := p.getFishingLevel(userIDStr)

		// 钓鱼成功率（根据等级提升）
		successRate := 0.5 + float64(level)*0.05
		if successRate > 0.95 {
			successRate = 0.95
		}

		// 判断是否钓鱼成功
		if rand.Float64() > successRate {
			// 钓鱼失败
			p.sendMessage(robot, event, "🎣 钓鱼失败了！鱼跑掉了，再接再厉哦")
			p.lastFishingTime[userIDStr] = now
			return nil
		}

		// 钓鱼成功，随机获得积分
		basePoints := 10 + level*5
		bonusPoints := rand.Intn(20)
		totalPoints := basePoints + bonusPoints

		// 增加积分
		err := db.AddPoints(p.db, userIDStr, totalPoints, "钓鱼获得", "fishing")
		if err != nil {
			p.sendMessage(robot, event, "钓鱼成功，但积分增加失败")
			return nil
		}

		// 提升钓鱼技能经验
		expGain := rand.Intn(5) + 1
		newLevel := p.addFishingExperience(userIDStr, expGain)

		// 更新钓鱼时间
		p.lastFishingTime[userIDStr] = now

		// 发送成功消息
		message := fmt.Sprintf("🎣 钓鱼成功！获得了 %d 积分", totalPoints)
		if newLevel > level {
			message += fmt.Sprintf("\n✨ 恭喜！钓鱼技能提升到 %d 级", newLevel)
		}
		p.sendMessage(robot, event, message)

		return nil
	})

	// 处理查看钓鱼等级命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "fishing") {
				return nil
			}
		}

		// 检查是否为查看钓鱼等级命令
		if match, _ := p.cmdParser.MatchCommand("钓鱼等级", event.RawMessage); !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		level := p.getFishingLevel(userIDStr)

		p.sendMessage(robot, event, fmt.Sprintf("🎣 你的钓鱼技能等级：%d级", level))

		return nil
	})

	// 处理查看钓鱼冷却命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "fishing") {
				return nil
			}
		}

		// 检查是否为查看钓鱼冷却命令
		if match, _ := p.cmdParser.MatchCommand("钓鱼冷却", event.RawMessage); !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		now := time.Now()

		if lastFishing, ok := p.lastFishingTime[userIDStr]; ok {
			if now.Sub(lastFishing) < 10*time.Minute {
				remainingTime := 10*time.Minute - now.Sub(lastFishing)
				p.sendMessage(robot, event, fmt.Sprintf("钓鱼冷却中，还需等待 %.0f 分钟才能再次钓鱼", remainingTime.Minutes()))
			} else {
				p.sendMessage(robot, event, "钓鱼冷却已结束，可以再次钓鱼")
			}
		} else {
			p.sendMessage(robot, event, "你还没有钓鱼过，可以随时钓鱼")
		}

		return nil
	})
}

// getFishingLevel 获取用户钓鱼等级
func (p *FishingPlugin) getFishingLevel(userIDStr string) int {
	if level, ok := p.fishingLevel[userIDStr]; ok {
		return level
	}
	return 1 // 默认等级1
}

// addFishingExperience 增加钓鱼经验并提升等级
func (p *FishingPlugin) addFishingExperience(userIDStr string, exp int) int {
	level := p.getFishingLevel(userIDStr)
	expNeeded := level * 10 // 升级所需经验

	// 这里简化处理，实际应该存储经验值
	if rand.Intn(expNeeded) < exp {
		newLevel := level + 1
		p.fishingLevel[userIDStr] = newLevel
		return newLevel
	}

	return level
}

// sendMessage 发送消息
func (p *FishingPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}