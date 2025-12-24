package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
)

// GreetingsPlugin 问候插件
type GreetingsPlugin struct {
	cmdParser *CommandParser
}

func (p *GreetingsPlugin) Name() string {
	return "greetings"
}

func (p *GreetingsPlugin) Description() string {
	return common.T("", "greetings_plugin_desc|问候插件，支持早安、晚安、欢迎语等功能")
}

func (p *GreetingsPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *GreetingsPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "morning_greeting",
			Description: common.T("", "greetings_skill_morning_desc|发送早安问候语"),
			Usage:       "morning_greeting",
			Params:      map[string]string{},
		},
		{
			Name:        "night_greeting",
			Description: common.T("", "greetings_skill_night_desc|发送晚安问候语"),
			Usage:       "night_greeting",
			Params:      map[string]string{},
		},
		{
			Name:        "welcome_greeting",
			Description: common.T("", "greetings_skill_welcome_desc|发送欢迎新成员问候语"),
			Usage:       "welcome_greeting user=张三",
			Params: map[string]string{
				"user": common.T("", "greetings_skill_param_user|被欢迎的用户名"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *GreetingsPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "morning_greeting":
		msg := p.doMorningGreeting()
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "night_greeting":
		msg := p.doNightGreeting()
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "welcome_greeting":
		user := params["user"]
		msg := p.doWelcomeGreeting(user)
		p.sendMessage(robot, event, msg)
		return msg, nil
	}
	return "", nil
}

// NewGreetingsPlugin 创建问候插件实例
func NewGreetingsPlugin() *GreetingsPlugin {
	return &GreetingsPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *GreetingsPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "greetings_plugin_loaded|加载问候插件"))

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
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "greetings") {
				HandleFeatureDisabled(robot, event, "greetings")
				return nil
			}
		}

		// 1. 早安命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "greetings_cmd_morning|早安|goodmorning"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doMorningGreeting())
			return nil
		}

		// 2. 晚安命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "greetings_cmd_night|晚安|goodnight"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doNightGreeting())
			return nil
		}

		// 3. 欢迎语命令
		match, _, welcomeUser := p.cmdParser.MatchCommandWithSingleParam(common.T("", "greetings_cmd_welcome|欢迎|welcome"), event.RawMessage)
		if match {
			p.sendMessage(robot, event, p.doWelcomeGreeting(welcomeUser))
			return nil
		}

		return nil
	})
}

// doMorningGreeting 执行早安问候逻辑
func (p *GreetingsPlugin) doMorningGreeting() string {
	return common.T("", "greetings_morning_msg|☀️ 早安！美好的一天开始了！")
}

// doNightGreeting 执行晚安问候逻辑
func (p *GreetingsPlugin) doNightGreeting() string {
	return common.T("", "greetings_night_msg|🌙 晚安！祝你做个好梦！")
}

// doWelcomeGreeting 执行欢迎问候逻辑
func (p *GreetingsPlugin) doWelcomeGreeting(user string) string {
	return fmt.Sprintf(common.T("", "greetings_welcome_msg|🎉 欢迎%s加入本群！"), user)
}

// sendMessage 发送消息
func (p *GreetingsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "greetings_send_failed|发送消息失败: %v"), err)
	}
}
