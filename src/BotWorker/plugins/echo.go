package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
)

type EchoPlugin struct{}

func (p *EchoPlugin) Name() string {
	return "Echo"
}

func (p *EchoPlugin) Description() string {
	return "复读消息插件"
}

func (p *EchoPlugin) Version() string {
	return "1.0.0"
}

func (p *EchoPlugin) Init(robot plugin.Robot) {
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType == "private" || event.MessageType == "group" {
			params := &onebot.SendMessageParams{
				MessageType: event.MessageType,
				UserID:      event.UserID,
				GroupID:     event.GroupID,
				Message:     event.Message,
			}
			robot.SendMessage(params)
		}
		return nil
	})
}
