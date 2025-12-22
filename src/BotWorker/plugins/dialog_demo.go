package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
)

type DialogDemoPlugin struct {
	cmdParser *CommandParser
}

func (p *DialogDemoPlugin) Name() string {
	return "dialog_demo"
}

func (p *DialogDemoPlugin) Description() string {
	return "对话流程示例插件，演示多级菜单与多步输入"
}

func (p *DialogDemoPlugin) Version() string {
	return "1.0.0"
}

func NewDialogDemoPlugin() *DialogDemoPlugin {
	return &DialogDemoPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *DialogDemoPlugin) Init(robot plugin.Robot) {
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		dialog := GetDialog(event.GroupID, event.UserID)
		if dialog != nil && dialog.Type == "set_welcome" {
			text := event.RawMessage
			if text == "" {
				if msg, ok := event.Message.(string); ok {
					text = msg
				}
			}

			if text == "" {
				return nil
			}

			if dialog.Step == 1 {
				if text == "1" || text == "2" || text == "3" {
					dialog.Data["mode"] = text
					UpdateDialog(dialog, 2, 5*60*1e9)
					SendTextReply(robot, event, "请输入欢迎语内容：")
				} else {
					SendTextReply(robot, event, "无效选项，请回复 1、2 或 3。")
				}
				return nil
			}

			if dialog.Step == 2 {
				dialog.Data["text"] = text
				mode := dialog.Data["mode"]
				msg := fmt.Sprintf("欢迎语已更新。\n模式: %s\n内容: %s", mode, text)
				SendTextReply(robot, event, msg)
				EndDialog(event.GroupID, event.UserID)
				return nil
			}
		}

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		match, _ := p.cmdParser.MatchCommand("设置欢迎语", event.RawMessage)
		if !match {
			return nil
		}

		dialog := StartDialog("set_welcome", event, 5*60*1e9)
		if dialog == nil {
			return nil
		}

		menu := "请选择欢迎语模式：\n1. 简短欢迎\n2. 详细欢迎\n3. 自定义文案\n请回复数字选择。"
		SendTextReply(robot, event, menu)

		return nil
	})
}

