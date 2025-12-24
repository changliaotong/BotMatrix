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
	return common.T("", "menu_plugin_desc|菜单插件，提供命令导航和帮助信息")
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
	log.Println(common.T("", "menu_plugin_loaded|菜单插件已加载"))

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
	menu := common.T("", "menu_title|🤖 BotMatrix 菜单") + "\n"
	menu += "====================\n\n"
	menu += common.T("", "menu_section_points|💰 积分系统") + ":\n"
	menu += common.T("", "menu_cmd_points_query|/points - 查询积分") + "\n"
	menu += common.T("", "menu_cmd_points_rank|/rank - 积分排行榜") + "\n\n"
	menu += common.T("", "menu_section_signin|📅 签到系统") + ":\n"
	menu += common.T("", "menu_cmd_signin|/signin - 每日签到") + "\n\n"
	menu += common.T("", "menu_section_weather|🌤️ 天气查询") + ":\n"
	menu += common.T("", "menu_cmd_weather|/weather - 查询默认城市天气") + "\n"
	menu += "/weather <city> - " + common.T("", "menu_help_weather|查询指定城市天气") + "\n\n"
	menu += common.T("", "menu_section_lottery|🧧 抽签解签") + ":\n"
	menu += common.T("", "menu_cmd_lottery|/lottery - 每日抽签") + "\n"
	menu += common.T("", "menu_cmd_lottery_explain|/explain - 解签") + "\n\n"
	menu += common.T("", "menu_section_translate|🔤 翻译系统") + ":\n"
	menu += common.T("", "menu_cmd_translate|/translate - 翻译指定文本") + "\n"
	menu += "/translate <text> - " + common.T("", "menu_help_translate|翻译指定文本") + "\n\n"
	menu += common.T("", "menu_section_music|🎵 音乐搜索") + ":\n"
	menu += common.T("", "menu_cmd_music|/music - 搜索并点歌") + "\n"
	menu += "/music <name> - " + common.T("", "menu_help_music|搜索并点歌") + "\n\n"
	menu += common.T("", "menu_section_pets|🐾 宠物系统") + ":\n"
	menu += common.T("", "menu_cmd_pets_adopt|/adopt - 领养宠物") + "\n"
	menu += common.T("", "menu_cmd_pets_my|/mypet - 我的宠物") + "\n"
	menu += common.T("", "menu_cmd_pets_feed|/feed - 喂食宠物") + "\n"
	menu += common.T("", "menu_cmd_pets_play|/play - 陪宠物玩耍") + "\n"
	menu += common.T("", "menu_cmd_pets_bath|/bath - 给宠物洗澡") + "\n\n"
	menu += "🎮 小型游戏:\n"
	menu += "小游戏\n\n"
	menu += common.T("", "menu_section_other|⚙️ 其他功能") + ":\n"
	menu += common.T("", "menu_cmd_menu|菜单 - 显示此菜单") + "\n"
	menu += common.T("", "menu_cmd_help|帮助 - 显示此菜单") + "\n"
	menu += "====================\n"
	menu += common.T("", "menu_tip|💡 提示: 发送命令即可触发功能") + "\n"
	menu += "💡 提示: 发送 '小游戏' 查看所有游戏列表"

	return menu
}
