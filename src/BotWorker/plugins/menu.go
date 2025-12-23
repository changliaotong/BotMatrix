package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
)

// MenuPlugin 菜单插件
type MenuPlugin struct {
	// 命令解析器
	cmdParser *CommandParser
}

func (p *MenuPlugin) Name() string {
	return "menu"
}

func (p *MenuPlugin) Description() string {
	return "菜单插件，显示所有可用命令"
}

func (p *MenuPlugin) Version() string {
	return "1.0.0"
}

// NewMenuPlugin 创建菜单插件实例
func NewMenuPlugin() *MenuPlugin {
	return &MenuPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MenuPlugin) Init(robot plugin.Robot) {
	log.Println("加载菜单插件")

	// 处理菜单命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为菜单命令
		if match, _ := p.cmdParser.MatchCommand("菜单|menu|help", event.RawMessage); match {
			// 显示菜单
			menuMsg := p.GetMenu()
			p.sendMessage(robot, event, menuMsg)
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *MenuPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// GetMenu 获取菜单内容
func (p *MenuPlugin) GetMenu() string {
	menu := "🤖 机器人命令菜单\n"
	menu += "====================\n\n"
	menu += "🎮 小型游戏:\n"
	menu += "小游戏\n\n"
	menu += "====================\n"
	menu += "💡 提示: 发送 '小游戏' 查看所有游戏列表"

	return menu
}
