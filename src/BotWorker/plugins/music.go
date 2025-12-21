package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
)

// MusicPlugin 点歌插件
type MusicPlugin struct {
	// 命令解析器
	cmdParser *CommandParser
}

func (p *MusicPlugin) Name() string {
	return "music"
}

func (p *MusicPlugin) Description() string {
	return "点歌插件，支持搜索歌曲并播放"
}

func (p *MusicPlugin) Version() string {
	return "1.0.0"
}

// NewMusicPlugin 创建点歌插件实例
func NewMusicPlugin() *MusicPlugin {
	return &MusicPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MusicPlugin) Init(robot plugin.Robot) {
	log.Println("加载点歌插件")

	// 处理点歌命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为点歌命令
		var songName string
		// 首先检查是否为带参数的点歌命令
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams("点歌|music", "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 解析歌曲名称
			songName = strings.TrimSpace(params[0])
		} else {
			// 检查是否为不带参数的点歌命令（显示帮助信息）
			matchHelp, _ := p.cmdParser.MatchCommand("点歌|music", event.RawMessage)
			if !matchHelp {
				return nil
			}
			// 发送帮助信息
			helpMsg := "点歌命令格式：\n/点歌 <歌曲名称> - 搜索并播放指定歌曲\n/music <歌曲名称> - 搜索并播放指定歌曲\n例如：/点歌 晴天"
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 模拟点歌功能
		musicMsg := fmt.Sprintf("🎵 正在为您点歌：%s\n请点击链接播放：https://music.163.com/#/search/m=%s", songName, songName)
		p.sendMessage(robot, event, musicMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *MusicPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
