package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "menu_plugin_desc")
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
	log.Println(common.T("", "menu_plugin_loaded"))

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
	menu := common.T("", "menu_title") + "\n"
	menu += "====================\n\n"
	menu += common.T("", "menu_section_points") + ":\n"
	menu += common.T("", "menu_cmd_points_query") + "\n"
	menu += common.T("", "menu_cmd_points_rank") + "\n\n"
	menu += common.T("", "menu_section_signin") + ":\n"
	menu += common.T("", "menu_cmd_signin") + "\n\n"
	menu += common.T("", "menu_section_weather") + ":\n"
	menu += common.T("", "menu_cmd_weather") + "\n"
	menu += "/weather <city> - " + common.T("", "menu_help_weather") + "\n\n"
	menu += common.T("", "menu_section_lottery") + ":\n"
	menu += common.T("", "menu_cmd_lottery") + "\n"
	menu += common.T("", "menu_cmd_lottery_explain") + "\n\n"
	menu += common.T("", "menu_section_translate") + ":\n"
	menu += common.T("", "menu_cmd_translate") + "\n"
	menu += "/translate <text> - " + common.T("", "menu_help_translate") + "\n\n"
	menu += common.T("", "menu_section_music") + ":\n"
	menu += common.T("", "menu_cmd_music") + "\n"
	menu += "/music <name> - " + common.T("", "menu_help_music") + "\n\n"
	menu += common.T("", "menu_section_pets") + ":\n"
	menu += common.T("", "menu_cmd_pets_adopt") + "\n"
	menu += common.T("", "menu_cmd_pets_my") + "\n"
	menu += common.T("", "menu_cmd_pets_feed") + "\n"
	menu += common.T("", "menu_cmd_pets_play") + "\n"
	menu += common.T("", "menu_cmd_pets_bath") + "\n\n"
	menu += "🎮 小型游戏:\n"
	menu += "小游戏\n\n"
	menu += common.T("", "menu_section_other") + ":\n"
	menu += common.T("", "menu_cmd_menu") + "\n"
	menu += common.T("", "menu_cmd_help") + "\n"
	menu += "====================\n"
	menu += common.T("", "menu_tip") + "\n"
	menu += "💡 提示: 发送 '小游戏' 查看所有游戏列表"

	return menu
}
