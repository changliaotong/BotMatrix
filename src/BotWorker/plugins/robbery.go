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
	return common.T("", "robbery_plugin_desc|打劫系统插件，支持用户之间的积分抢夺功能")
}

func (p *RobberyPlugin) Version() string {
	return "1.0.0"
}

func (p *RobberyPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "robbery_db_not_configured|打劫系统插件未配置数据库，功能将不可用"))
		return
	}
	log.Println(common.T("", "robbery_plugin_loaded|加载打劫系统插件"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 统一消息处理器
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

		// 1. 打劫命令
		if match, cmd, params := p.cmdParser.MatchCommandWithParams(common.T("", "robbery_cmd_rob|打劫|抢劫"), "(\\d+)", event.RawMessage); match {
			if len(params) != 1 {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "robbery_cmd_usage|%s命令格式：%s <用户ID>"), cmd, cmd))
				return nil
			}
			msg, err := p.handleRobLogic(robot, event, params[0])
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 2. 查看打劫冷却时间命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "robbery_cmd_cooling|打劫冷却"), event.RawMessage); match {
			msg, err := p.handleCoolingLogic(robot, event)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})
}

// GetSkills 实现 SkillCapable 接口
func (p *RobberyPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "rob",
			Description: common.T("", "robbery_skill_rob_desc|打劫指定用户的积分"),
			Usage:       "rob to_user_id=123456",
			Params: map[string]string{
				"to_user_id": common.T("", "robbery_skill_rob_param_to_user_id|被打劫的用户ID"),
			},
		},
		{
			Name:        "cooling",
			Description: common.T("", "robbery_skill_cooling_desc|查看打劫冷却时间"),
			Usage:       "cooling",
		},
	}
}

// HandleSkill 处理技能调用
func (p *RobberyPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "rob":
		toUserID := params["to_user_id"]
		if toUserID == "" {
			return "", fmt.Errorf(common.T("", "robbery_cmd_usage|%s命令格式：%s <用户ID>"), "rob", "rob")
		}
		return p.handleRobLogic(robot, event, toUserID)
	case "cooling":
		return p.handleCoolingLogic(robot, event)
	}
	return "", fmt.Errorf("unknown skill: %s", skillName)
}

func (p *RobberyPlugin) handleRobLogic(robot plugin.Robot, event *onebot.Event, victimUserIDStr string) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}
	robberUserID := event.UserID
	robberUserIDStr := fmt.Sprintf("%d", robberUserID)

	// 不能打劫自己
	if robberUserIDStr == victimUserIDStr {
		return common.T("", "robbery_self|不能打劫自己哦"), nil
	}

	// 检查打劫冷却时间（每小时只能打劫一次）
	now := time.Now()
	if lastRobbery, ok := p.lastRobberyTime[robberUserIDStr]; ok {
		if now.Sub(lastRobbery) < 1*time.Hour {
			remainingTime := 1*time.Hour - now.Sub(lastRobbery)
			return fmt.Sprintf(common.T("", "robbery_cooling|打劫冷却中，还需等待 %.0f 分钟才能再次打劫"), remainingTime.Minutes()), nil
		}
	}

	// 获取被打劫用户的积分
	victimPoints, err := db.GetPoints(p.db, victimUserIDStr)
	if err != nil {
		return common.T("", "robbery_query_failed|查询被打劫用户信息失败"), nil
	}

	// 检查被打劫用户积分是否足够
	if victimPoints < 10 {
		return common.T("", "robbery_victim_too_poor|这个用户太穷了，不值得打劫"), nil
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
	err = db.TransferPoints(p.db, victimUserIDStr, robberUserIDStr, robberyAmount, common.T("", "robbery_reason_victim|被打劫"), "robbery")
	if err != nil {
		return fmt.Sprintf(common.T("", "robbery_failed|打劫失败: %v"), err), nil
	}

	// 更新打劫时间
	p.lastRobberyTime[robberUserIDStr] = now

	// 发送成功消息
	return fmt.Sprintf(common.T("", "robbery_success|✅ 打劫成功！从用户 %s 处抢到了 %d 积分"), victimUserIDStr, robberyAmount), nil
}

func (p *RobberyPlugin) handleCoolingLogic(robot plugin.Robot, event *onebot.Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}
	userIDStr := fmt.Sprintf("%d", event.UserID)
	now := time.Now()

	if lastRobbery, ok := p.lastRobberyTime[userIDStr]; ok {
		if now.Sub(lastRobbery) < 1*time.Hour {
			remainingTime := 1*time.Hour - now.Sub(lastRobbery)
			return fmt.Sprintf(common.T("", "robbery_cooling_status|打劫冷却中，还需等待 %.0f 分钟才能再次打劫"), remainingTime.Minutes()), nil
		} else {
			return common.T("", "robbery_cooling_finished|打劫冷却已结束，可以再次打劫"), nil
		}
	} else {
		return common.T("", "robbery_no_record|你还没有打劫过，可以随时打劫"), nil
	}
}

// sendMessage 发送消息
func (p *RobberyPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
