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

// CultivationPlugin 修炼系统插件
type CultivationPlugin struct {
	db *sql.DB
	// 存储用户上次修炼时间，key为用户ID，value为修炼时间
	lastCultivationTime map[string]time.Time
	// 存储用户修炼等级，key为用户ID，value为等级
	cultivationLevel map[string]int
	// 存储用户灵力值，key为用户ID，value为灵力值
	energy map[string]int
	// 命令解析器
	cmdParser *CommandParser
}

// NewCultivationPlugin 创建修炼系统插件实例
func NewCultivationPlugin(database *sql.DB) *CultivationPlugin {
	rand.Seed(time.Now().UnixNano())
	return &CultivationPlugin{
		db:                  database,
		lastCultivationTime: make(map[string]time.Time),
		cultivationLevel:    make(map[string]int),
		energy:              make(map[string]int),
		cmdParser:           NewCommandParser(),
	}
}

func (p *CultivationPlugin) Name() string {
	return "cultivation"
}

func (p *CultivationPlugin) Description() string {
	return "修炼系统插件，支持修炼提升境界和获得奖励"
}

func (p *CultivationPlugin) Version() string {
	return "1.0.0"
}

func (p *CultivationPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("修炼系统插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载修炼系统插件")

	// 处理修炼命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "cultivation") {
				HandleFeatureDisabled(robot, event, "cultivation")
				return nil
			}
		}

		// 检查是否为修炼命令
		match, _ := p.cmdParser.MatchCommand("修炼|修行", event.RawMessage)
		if !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		now := time.Now()

		// 检查修炼冷却时间（每30分钟只能修炼一次）
		if lastCultivation, ok := p.lastCultivationTime[userIDStr]; ok {
			if now.Sub(lastCultivation) < 30*time.Minute {
				remainingTime := 30*time.Minute - now.Sub(lastCultivation)
				p.sendMessage(robot, event, fmt.Sprintf("修炼冷却中，还需等待 %.0f 分钟才能再次修炼", remainingTime.Minutes()))
				return nil
			}
		}

		// 获取用户当前等级
		level := p.getCultivationLevel(userIDStr)
		currentEnergy := p.getEnergy(userIDStr)

		// 计算本次修炼获得的灵力
		energyGain := 10 + level*2 + rand.Intn(10)
		newEnergy := currentEnergy + energyGain

		// 检查是否可以突破境界
		requiredEnergy := level * 100
		var breakthrough bool
		var newLevel int

		if newEnergy >= requiredEnergy {
			// 突破境界
			breakthrough = true
			newLevel = level + 1
			newEnergy = newEnergy - requiredEnergy
			p.cultivationLevel[userIDStr] = newLevel
			
			// 突破奖励积分
			rewardPoints := level * 50
			err := db.AddPoints(p.db, userIDStr, rewardPoints, "突破境界奖励", "cultivation_breakthrough")
			if err != nil {
				log.Printf("突破奖励积分增加失败: %v", err)
			}
		} else {
			newLevel = level
		}

		// 更新灵力值
		p.energy[userIDStr] = newEnergy

		// 更新修炼时间
		p.lastCultivationTime[userIDStr] = now

		// 发送修炼结果消息
		message := fmt.Sprintf("🧘 修炼完成！获得了 %d 灵力", energyGain)
		message += fmt.Sprintf("\n当前灵力: %d/%d", newEnergy, requiredEnergy)
		message += fmt.Sprintf("\n当前境界: %d 级", newLevel)

		if breakthrough {
			message += fmt.Sprintf("\n🎉 恭喜！成功突破到 %d 级！获得 %d 积分奖励", newLevel, level*50)
		}

		p.sendMessage(robot, event, message)

		return nil
	})

	// 处理查看境界命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "cultivation") {
				return nil
			}
		}

		// 检查是否为查看境界命令
		match, _ := p.cmdParser.MatchCommand("境界|修炼等级", event.RawMessage)
		if !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		level := p.getCultivationLevel(userIDStr)
		currentEnergy := p.getEnergy(userIDStr)
		requiredEnergy := level * 100

		message := fmt.Sprintf("🧘 你的当前境界: %d 级", level)
		message += fmt.Sprintf("\n当前灵力: %d/%d", currentEnergy, requiredEnergy)
		message += fmt.Sprintf("\n下一级突破需要: %d 灵力", requiredEnergy-currentEnergy)

		p.sendMessage(robot, event, message)

		return nil
	})

	// 处理查看修炼冷却命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "cultivation") {
				return nil
			}
		}

		// 检查是否为查看修炼冷却命令
		match, _ := p.cmdParser.MatchCommand("修炼冷却", event.RawMessage)
		if !match {
			return nil
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		now := time.Now()

		if lastCultivation, ok := p.lastCultivationTime[userIDStr]; ok {
			if now.Sub(lastCultivation) < 30*time.Minute {
				remainingTime := 30*time.Minute - now.Sub(lastCultivation)
				p.sendMessage(robot, event, fmt.Sprintf("修炼冷却中，还需等待 %.0f 分钟才能再次修炼", remainingTime.Minutes()))
			} else {
				p.sendMessage(robot, event, "修炼冷却已结束，可以再次修炼")
			}
		} else {
			p.sendMessage(robot, event, "你还没有修炼过，可以随时开始修炼")
		}

		return nil
	})
}

// getCultivationLevel 获取用户修炼等级
func (p *CultivationPlugin) getCultivationLevel(userIDStr string) int {
	if level, ok := p.cultivationLevel[userIDStr]; ok {
		return level
	}
	return 1 // 默认等级1
}

// getEnergy 获取用户灵力值
func (p *CultivationPlugin) getEnergy(userIDStr string) int {
	if energy, ok := p.energy[userIDStr]; ok {
		return energy
	}
	return 0 // 默认灵力0
}

// sendMessage 发送消息
func (p *CultivationPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}