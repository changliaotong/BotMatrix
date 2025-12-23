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
	return common.T("", "greetings_plugin_desc")
}

func (p *GreetingsPlugin) Version() string {
	return "1.0.0"
}

// NewGreetingsPlugin 创建问候插件实例
func NewGreetingsPlugin() *GreetingsPlugin {
	return &GreetingsPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *GreetingsPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "greetings_plugin_loaded"))

	// 处理早安命令
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

		// 检查是否为早安命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "greetings_cmd_morning"), event.RawMessage); !match {
			return nil
		}

		// 发送早安问候
		morningMsg := common.T("", "greetings_morning_msg")
		p.sendMessage(robot, event, morningMsg)

		return nil
	})

	// 处理晚安命令
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

		// 检查是否为晚安命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "greetings_cmd_night"), event.RawMessage); !match {
			return nil
		}

		// 发送晚安问候
		nightMsg := common.T("", "greetings_night_msg")
		p.sendMessage(robot, event, nightMsg)

		return nil
	})

	// 处理欢迎语命令
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

		// 检查是否为欢迎语命令
		match, _, welcomeUser := p.cmdParser.MatchCommandWithSingleParam(common.T("", "greetings_cmd_welcome"), event.RawMessage)
		if !match {
			return nil
		}

		// 发送欢迎语
		welcomeMsg := fmt.Sprintf(common.T("", "greetings_welcome_msg"), welcomeUser)
		p.sendMessage(robot, event, welcomeMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *GreetingsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "greetings_send_failed"), err)
	}
}
