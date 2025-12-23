package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "dialog_demo_plugin_desc")
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
					SendTextReply(robot, event, common.T("", "dialog_demo_input_content"))
				} else {
					SendTextReply(robot, event, common.T("", "dialog_demo_invalid_option"))
				}
				return nil
			}

			if dialog.Step == 2 {
				dialog.Data["text"] = text
				mode := dialog.Data["mode"]
				msg := fmt.Sprintf(common.T("", "dialog_demo_updated"), mode, text)
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

		match, _ := p.cmdParser.MatchCommand(common.T("", "dialog_demo_cmd_set_welcome"), event.RawMessage)
		if !match {
			return nil
		}

		dialog := StartDialog("set_welcome", event, 5*60*1e9)
		if dialog == nil {
			return nil
		}

		menu := common.T("", "dialog_demo_menu")
		SendTextReply(robot, event, menu)

		return nil
	})
}

