package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
)

// MenuPlugin 菜单插件
type MenuPlugin struct{}

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
	return &MenuPlugin{}
}

func (p *MenuPlugin) Init(robot plugin.Robot) {
	log.Println("加载菜单插件")

	// 处理菜单命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为菜单命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg == "!菜单" || msg == "!menu" || msg == "help" || msg == "!help" {
			// 显示菜单
			menuMsg := p.getMenu()
			p.sendMessage(robot, event, menuMsg)
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *MenuPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// GetMenu 获取菜单内容
func (p *MenuPlugin) GetMenu() string {
	menu := "🤖 机器人命令菜单\n"
	menu += "====================\n\n"
	menu += "📊 积分系统:\n"
	menu += "!积分 查询 - 查询当前积分\n"
	menu += "!积分排行 - 查看积分排行榜\n\n"
	menu += "📅 签到系统:\n"
	menu += "!签到 - 每日签到获取积分\n\n"
	menu += "🌤️ 天气查询:\n"
	menu += "!天气 <城市名> - 查询指定城市天气\n"
	menu += "!weather <城市名> - 查询指定城市天气\n\n"
	menu += "🎲 抽签功能:\n"
	menu += "!抽签 - 进行一次抽签\n"
	menu += "!解签 <签文> - 解析签文含义\n\n"
	menu += "🌐 翻译功能:\n"
	menu += "!翻译 <文本> - 翻译指定文本\n"
	menu += "!translate <文本> - 翻译指定文本\n\n"
	menu += "🎵 点歌功能:\n"
	menu += "!点歌 <歌曲名称> - 搜索并播放指定歌曲\n"
	menu += "!music <歌曲名称> - 搜索并播放指定歌曲\n\n"
	menu += "ℹ️ 其他命令:\n"
	menu += "!菜单 - 显示此帮助菜单\n"
	menu += "!help - 显示帮助信息\n"
	menu += "====================\n"
	menu += "💡 提示: 所有命令支持中文和英文两种格式"

	return menu
}