package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
)

type EchoPlugin struct {
	cmdParser *CommandParser
}

func (p *EchoPlugin) Name() string {
	return "echo"
}

func (p *EchoPlugin) Description() string {
	return common.T("", "echo_plugin_desc|复读插件，可以重复用户说的话")
}

func (p *EchoPlugin) Version() string {
	return "1.0.0"
}

func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *EchoPlugin) Init(robot plugin.Robot) {
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 复读命令: /echo <message>
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "echo_cmd|复读"), `(.*)`, event.RawMessage); match && len(params) > 0 {
			message := params[0]
			if message == "" {
				return nil
			}

			log.Printf("Echoing message: %s\n", message)
			_, err := SendTextReply(robot, event, message)
			return err
		}

		return nil
	})

	// 报备技能
	robot.HandleSkill("echo", func(params map[string]string) (string, error) {
		message, ok := params["message"]
		if !ok || message == "" {
			return "", fmt.Errorf("missing message parameter")
		}
		return message, nil
	})
}
