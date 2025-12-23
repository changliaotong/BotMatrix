package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
	"time"
)

// TimePlugin 报时插件
type TimePlugin struct {
	cmdParser *CommandParser
}

func (p *TimePlugin) Name() string {
	return "time"
}

func (p *TimePlugin) Description() string {
	return "报时插件，显示当前时间"
}

func (p *TimePlugin) Version() string {
	return "1.0.0"
}

// NewTimePlugin 创建报时插件实例
func NewTimePlugin() *TimePlugin {
	return &TimePlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *TimePlugin) Init(robot plugin.Robot) {
	log.Println("加载报时插件")

	// 处理报时命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为报时命令
		if match, _ := p.cmdParser.MatchCommand("时间|time|now", event.RawMessage); match {
			// 获取当前时间
			currentTime := time.Now().Format("2006-01-02 15:04:05")
			message := "🕐 当前时间: " + currentTime
			
			// 发送时间消息
			p.sendMessage(robot, event, message)
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *TimePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}