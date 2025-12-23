package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "music_plugin_desc")
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
	log.Println(common.T("", "music_plugin_loaded"))

	// 处理点歌命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "music") {
				HandleFeatureDisabled(robot, event, "music")
				return nil
			}
		}

		// 检查是否为点歌命令
		var songName string
		// 首先检查是否为带参数的点歌命令
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "music_cmd_play"), "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 解析歌曲名称
			songName = strings.TrimSpace(params[0])
		} else {
			// 检查是否为不带参数的点歌命令（显示帮助信息）
			matchHelp, _ := p.cmdParser.MatchCommand(common.T("", "music_cmd_play"), event.RawMessage)
			if !matchHelp {
				return nil
			}
			// 发送帮助信息
			helpMsg := common.T("", "music_help_msg")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 模拟点歌功能
		musicMsg := fmt.Sprintf(common.T("", "music_playing_msg"), songName, songName)
		p.sendMessage(robot, event, musicMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *MusicPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "music_send_failed_log"), err)
	}
}
