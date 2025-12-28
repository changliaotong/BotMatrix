package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
)

type WelcomePlugin struct{}

func (p *WelcomePlugin) Name() string {
	return "Welcome"
}

func (p *WelcomePlugin) Description() string {
	return "欢迎新成员插件"
}

func (p *WelcomePlugin) Version() string {
	return "1.0.0"
}

func (p *WelcomePlugin) Init(robot plugin.Robot) {
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_increase" {
			params := &onebot.SendMessageParams{
				MessageType: "group",
				GroupID:     event.GroupID,
				Message:     fmt.Sprintf("欢迎新成员 [CQ:at,qq=%d] 加入本群！", event.UserID),
			}
			robot.SendMessage(params)
		}
		return nil
	})
}
