package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
)

type EchoPlugin struct {
	cmdParser *CommandParser
}

func (p *EchoPlugin) Name() string {
	return "echo"
}

func (p *EchoPlugin) Description() string {
	return common.T("", "echo_plugin_desc")
}

func (p *EchoPlugin) Version() string {
	return "1.0.0"
}

// NewEchoPlugin 创建回声插件实例
func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *EchoPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "echo_plugin_loaded"))

	// 响应消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		log.Printf(common.T("", "echo_msg_received"), event.RawMessage)

		// 只处理私聊消息
		if event.MessageType == "private" {
			// 发送回声消息
			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: common.T("", "echo_reply_prefix") + event.RawMessage,
			}
			robot.SendMessage(params)
		}

		return nil
	})

	// 处理帮助命令
	robot.OnMessage(func(event *onebot.Event) error {
		if match, _ := p.cmdParser.MatchCommand("help", event.RawMessage); match {
			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: common.T("", "echo_help_msg"),
			}
			robot.SendMessage(params)
		}

		return nil
	})
}
