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
	return common.T("", "cultivation_plugin_desc|修炼系统插件，支持闭关修炼、境界突破和灵力查看")
}

func (p *CultivationPlugin) Version() string {
	return "1.0.0"
}

func (p *CultivationPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "cultivation_no_db|数据库未初始化，修炼插件功能受限"))
		return
	}
	log.Println(common.T("", "cultivation_plugin_loaded|修炼系统插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理修炼相关命令
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

		userIDStr := fmt.Sprintf("%d", event.UserID)

		// 检查是否为修炼命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "cultivation_cmd_cultivate|修炼|闭关|开始修炼"), event.RawMessage); match {
			resp, err := p.doCultivate(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为查看境界命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "cultivation_cmd_status|境界|我的境界|查看修为"), event.RawMessage); match {
			resp, err := p.doGetStatus(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为查看修炼冷却命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "cultivation_cmd_cooldown|修炼冷却|修炼时间"), event.RawMessage); match {
			resp, err := p.doGetCooldown(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		return nil
	})
}

// GetSkills 报备插件技能
func (p *CultivationPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "start_cultivation",
			Description: common.T("", "cultivation_skill_start_desc|开始一次闭关修炼"),
			Usage:       "start_cultivation",
			Params:      map[string]string{},
		},
		{
			Name:        "get_cultivation_status",
			Description: common.T("", "cultivation_skill_status_desc|获取当前修炼境界和灵力信息"),
			Usage:       "get_cultivation_status",
			Params:      map[string]string{},
		},
		{
			Name:        "get_cultivation_cooldown",
			Description: common.T("", "cultivation_skill_cooldown_desc|查询下一次修炼所需的等待时间"),
			Usage:       "get_cultivation_cooldown",
			Params:      map[string]string{},
		},
	}
}

// HandleSkill 处理技能调用
func (p *CultivationPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userID := ""
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if uid, ok := params["user_id"]; ok {
		userID = uid
	}

	if userID == "" {
		return common.T("", "cultivation_missing_user_id|未找到用户ID，无法执行修炼相关操作"), nil
	}

	switch skillName {
	case "start_cultivation":
		return p.doCultivate(userID)
	case "get_cultivation_status":
		return p.doGetStatus(userID)
	case "get_cultivation_cooldown":
		return p.doGetCooldown(userID)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doCultivate 修炼逻辑
func (p *CultivationPlugin) doCultivate(userIDStr string) (string, error) {
	now := time.Now()

	// 检查修炼冷却时间（每30分钟只能修炼一次）
	if lastCultivation, ok := p.lastCultivationTime[userIDStr]; ok {
		if now.Sub(lastCultivation) < 30*time.Minute {
			remainingTime := 30*time.Minute - now.Sub(lastCultivation)
			return fmt.Sprintf(common.T("", "cultivation_cooldown_msg|道友莫急，你刚修炼完不久，神识尚未恢复。还需等待 %.1f 分钟方可再次修炼。"), remainingTime.Minutes()), nil
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
		uid, _ := strconv.ParseInt(userIDStr, 10, 64)
		err := db.AddPoints(p.db, uid, rewardPoints, common.T("", "cultivation_breakthrough_reason|突破境界奖励"), "cultivation_breakthrough")
		if err != nil {
			log.Printf(common.T("", "cultivation_reward_failed_log|发放修炼突破奖励失败")+": %v", err)
		}
	} else {
		newLevel = level
	}

	// 更新灵力值
	p.energy[userIDStr] = newEnergy

	// 更新修炼时间
	p.lastCultivationTime[userIDStr] = now

	// 构建修炼结果消息
	message := fmt.Sprintf(common.T("", "cultivation_result_gain|你沉浸在玄妙的感悟中，本次修炼获得了 %d 点灵力。"), energyGain)
	message += "\n" + fmt.Sprintf(common.T("", "cultivation_result_status|当前灵力：%d/%d，当前境界：第 %d 阶"), newEnergy, requiredEnergy, newLevel)

	if breakthrough {
		message += "\n" + fmt.Sprintf(common.T("", "cultivation_result_breakthrough|恭喜道友成功突破至第 %d 阶！由于境界提升，你获得了 %d 积分奖励。"), newLevel, level*50)
	}

	return message, nil
}

// doGetStatus 查看境界逻辑
func (p *CultivationPlugin) doGetStatus(userIDStr string) (string, error) {
	level := p.getCultivationLevel(userIDStr)
	currentEnergy := p.getEnergy(userIDStr)
	requiredEnergy := level * 100

	return fmt.Sprintf(common.T("", "cultivation_status_msg|道友当前的修为信息：\n当前境界：第 %d 阶\n当前灵力：%d/%d\n距离下一境界还需灵力：%d"), level, currentEnergy, requiredEnergy, requiredEnergy-currentEnergy), nil
}

// doGetCooldown 查看冷却逻辑
func (p *CultivationPlugin) doGetCooldown(userIDStr string) (string, error) {
	now := time.Now()

	if lastCultivation, ok := p.lastCultivationTime[userIDStr]; ok {
		if now.Sub(lastCultivation) < 30*time.Minute {
			remainingTime := 30*time.Minute - now.Sub(lastCultivation)
			return fmt.Sprintf(common.T("", "cultivation_cooldown_msg|道友莫急，你刚修炼完不久，神识尚未恢复。还需等待 %.1f 分钟方可再次修炼。"), remainingTime.Minutes()), nil
		} else {
			return common.T("", "cultivation_cooldown_ready|道友神识已完全恢复，随时可以再次进入修炼状态。"), nil
		}
	} else {
		return common.T("", "cultivation_no_history|道友尚未开始修炼，不如现在就开始闭关感悟？"), nil
	}
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
	if robot == nil || event == nil {
		log.Printf(common.T("", "cultivation_send_failed_log|发送修炼消息失败，机器人或事件为空"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "cultivation_send_failed_log|发送修炼消息失败")+": %v", err)
	}
}
