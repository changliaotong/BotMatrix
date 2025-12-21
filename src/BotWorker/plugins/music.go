package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
)

// MusicPlugin 点歌插件
type MusicPlugin struct{}

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
	return &MusicPlugin{}
}

func (p *MusicPlugin) Init(robot plugin.Robot) {
	log.Println("加载点歌插件")

	// 处理点歌命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为点歌命令
		msg := strings.TrimSpace(event.RawMessage)
		if strings.HasPrefix(msg, "!点歌 ") || strings.HasPrefix(msg, "!music ") {
			// 解析歌曲名称
			var songName string
			if strings.HasPrefix(msg, "!点歌 ") {
				songName = strings.TrimSpace(msg[4:])
			} else {
				songName = strings.TrimSpace(msg[7:])
			}

			if songName == "" {
				// 发送帮助信息
				helpMsg := "点歌命令格式：\n!点歌 <歌曲名称> - 搜索并播放指定歌曲\n!music <歌曲名称> - 搜索并播放指定歌曲\n例如：!点歌 晴天"
				p.sendMessage(robot, event, helpMsg)
				return nil
			}

			// 模拟点歌功能
			musicMsg := fmt.Sprintf("🎵 正在为您点歌：%s\n请点击链接播放：https://music.163.com/#/search/m=%s", songName, songName)
			p.sendMessage(robot, event, musicMsg)
		}

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
