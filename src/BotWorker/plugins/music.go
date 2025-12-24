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
	return common.T("", "music_plugin_desc|点歌插件，支持搜索并分享音乐")
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
	log.Println(common.T("", "music_plugin_loaded|点歌插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

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
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "music_cmd_play|点歌|播放|play music"), "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 解析歌曲名称
			songName = strings.TrimSpace(params[0])
		} else {
			// 检查是否为不带参数的点歌命令（显示帮助信息）
			matchHelp, _ := p.cmdParser.MatchCommand(common.T("", "music_cmd_play|点歌|播放|play music"), event.RawMessage)
			if !matchHelp {
				return nil
			}
			// 发送帮助信息
			helpMsg := common.T("", "music_help_msg|请在点歌命令后加上歌曲名称，例如：点歌 晴天")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 模拟点歌功能
		musicMsg, _ := p.doPlayMusic(songName)
		p.sendMessage(robot, event, musicMsg)

		return nil
	})
}

// GetSkills 报备插件技能
func (p *MusicPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "play_music",
			Description: common.T("", "music_skill_play_desc|播放指定名称的歌曲"),
			Usage:       "play_music song_name=晴天",
			Params: map[string]string{
				"song_name": common.T("", "music_skill_param_song_name|歌曲名称"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *MusicPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "play_music":
		songName := params["song_name"]
		if songName == "" {
			return "", fmt.Errorf(common.T("", "music_missing_param_song_name|缺少歌曲名称参数"))
		}
		return p.doPlayMusic(songName)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doPlayMusic 执行点歌逻辑
func (p *MusicPlugin) doPlayMusic(songName string) (string, error) {
	return fmt.Sprintf(common.T("", "music_playing_msg|正在为您播放歌曲：%s\n[点击播放](https://music.163.com/search/#/m/?s=%s)"), songName, songName), nil
}

// sendMessage 发送消息
func (p *MusicPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "music_send_failed_log|发送点歌消息失败，机器人或事件为空"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "music_send_failed_log|发送点歌消息失败")+": %v", err)
	}
}
