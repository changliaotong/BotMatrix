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
	return "简单的回声插件，回复收到的消息"
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
	log.Println("加载回声插件")

	// 响应消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		log.Printf("收到消息: %s", event.RawMessage)

		// 只处理私聊消息
		if event.MessageType == "private" {
			// 发送回声消息
			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: "你说: " + event.RawMessage,
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
				Message: "可用命令:\n- /help: 显示帮助信息\n- /echo [内容]: 回复相同内容",
			}
			robot.SendMessage(params)
		}

		return nil
	})
}
