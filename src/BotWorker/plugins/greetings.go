package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
)

// GreetingsPlugin 问候插件
type GreetingsPlugin struct{}

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
	return &GreetingsPlugin{}
}

func (p *GreetingsPlugin) Init(robot plugin.Robot) {
	log.Println("加载问候插件")

	// 处理早安命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为早安命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!早安" && msg != "!goodmorning" {
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

		// 检查是否为晚安命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!晚安" && msg != "!goodnight" {
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

		// 检查是否为欢迎语命令
		msg := strings.TrimSpace(event.RawMessage)
		if !strings.HasPrefix(msg, "!欢迎 ") && !strings.HasPrefix(msg, "!welcome ") {
			return nil
		}

		// 解析欢迎对象
		var welcomeUser string
		if strings.HasPrefix(msg, "!欢迎 ") {
			welcomeUser = strings.TrimSpace(msg[3:])
		} else {
			welcomeUser = strings.TrimSpace(msg[9:])
		}

		// 发送欢迎语
		welcomeMsg := fmt.Sprintf("🎉 欢迎%s加入本群！", welcomeUser)
		p.sendMessage(robot, event, welcomeMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *GreetingsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
