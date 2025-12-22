package plugins

import (
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
	return "问候插件，支持早安、晚安、欢迎语等功能"
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
	log.Println("加载问候插件")

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
		if match, _ := p.cmdParser.MatchCommand("早安|goodmorning", event.RawMessage); !match {
			return nil
		}

		// 发送早安问候
		morningMsg := "☀️ 早安！美好的一天开始了！"
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
		if match, _ := p.cmdParser.MatchCommand("晚安|goodnight", event.RawMessage); !match {
			return nil
		}

		// 发送晚安问候
		nightMsg := "🌙 晚安！祝你做个好梦！"
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
		match, _, welcomeUser := p.cmdParser.MatchCommandWithSingleParam("欢迎|welcome", event.RawMessage)
		if !match {
			return nil
		}

		// 发送欢迎语
		welcomeMsg := fmt.Sprintf("🎉 欢迎%s加入本群！", welcomeUser)
		p.sendMessage(robot, event, welcomeMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *GreetingsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
