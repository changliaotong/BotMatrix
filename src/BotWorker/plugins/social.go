package plugins

import (
	"BotMatrix/common"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"botworker/internal/onebot"
	"botworker/internal/plugin"
)

// SocialPlugin  social plugin
type SocialPlugin struct {
	// 头衔列表
	titles []string
}

func (p *SocialPlugin) Name() string {
	return "social"
}

func (p *SocialPlugin) Description() string {
	return common.T("", "social_plugin_desc")
}

func (p *SocialPlugin) Version() string {
	return "1.0.0"
}

// NewSocialPlugin 创建social plugin实例
func NewSocialPlugin() *SocialPlugin {
	return &SocialPlugin{
		titles: []string{
			common.T("", "social_title_owner"), common.T("", "social_title_admin"), common.T("", "social_title_svip"), common.T("", "social_title_vip"),
			common.T("", "social_title_user"), common.T("", "social_title_newbie"), common.T("", "social_title_master"), common.T("", "social_title_scholar"),
			common.T("", "social_title_loser"), common.T("", "social_title_gamer"), common.T("", "social_title_music_fan"), common.T("", "social_title_foodie"),
			common.T("", "social_title_traveler"), common.T("", "social_title_photographer"), common.T("", "social_title_writer"), common.T("", "social_title_painter"),
			common.T("", "social_title_designer"), common.T("", "social_title_coder"), common.T("", "social_title_engineer"), common.T("", "social_title_doctor"),
			common.T("", "social_title_teacher"), common.T("", "social_title_student"), common.T("", "social_title_worker"), common.T("", "social_title_freelancer"),
			common.T("", "social_title_entrepreneur"), common.T("", "social_title_investor"), common.T("", "social_title_collector"), common.T("", "social_title_fitness"),
			common.T("", "social_title_athlete"), common.T("", "social_title_eater"), common.T("", "social_title_sleeper"), common.T("", "social_title_procrastinator"),
			common.T("", "social_title_ocd"), common.T("", "social_title_indecisive"), common.T("", "social_title_lost"), common.T("", "social_title_blind"),
			common.T("", "social_title_tone_deaf"), common.T("", "social_title_clumsy"), common.T("", "social_title_clean_freak"), common.T("", "social_title_night_owl"),
			common.T("", "social_title_early_bird"), common.T("", "social_title_social_butterfly"), common.T("", "social_title_social_phobia"), common.T("", "social_title_social_cow"),
			common.T("", "social_title_troll"), common.T("", "social_title_joker"), common.T("", "social_title_meme_master"), common.T("", "social_title_binge_watcher"),
			common.T("", "social_title_idol_fan"), common.T("", "social_title_2d"), common.T("", "social_title_3d"), common.T("", "social_title_4d"),
			common.T("", "social_title_coser"), common.T("", "social_title_editor"), common.T("", "social_title_video_editor"), common.T("", "social_title_screenwriter"),
			common.T("", "social_title_director"), common.T("", "social_title_actor"), common.T("", "social_title_singer"), common.T("", "social_title_dancer"),
			common.T("", "social_title_musician"), common.T("", "social_title_producer"), common.T("", "social_title_streamer"), common.T("", "social_title_up"),
			common.T("", "social_title_blogger"), common.T("", "social_title_influencer"), common.T("", "social_title_star"), common.T("", "social_title_idol"),
			common.T("", "social_title_god"), common.T("", "social_title_goddess"), common.T("", "social_title_handsome"), common.T("", "social_title_beauty"),
			common.T("", "social_title_cute_girl"), common.T("", "social_title_big_sister"), common.T("", "social_title_loli"), common.T("", "social_title_shota"),
			common.T("", "social_title_uncle"), common.T("", "social_title_aunt"), common.T("", "social_title_little_sister"), common.T("", "social_title_little_brother"),
		},
	}
}

func (p *SocialPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "social_plugin_loaded"))

	// 处理群消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 处理爱群主命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, common.T("", "social_cmd_love_owner")) {
			return p.handleLoveOwnerCommand(robot, event)
		}

		// 处理变身命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, common.T("", "social_cmd_transform")) {
			return p.handleTransformCommand(robot, event)
		}

		// 处理头衔命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, common.T("", "social_cmd_title")) {
			return p.handleTitleCommand(robot, event)
		}

		return nil
	})
}

func (p *SocialPlugin) handleLoveOwnerCommand(robot plugin.Robot, event *onebot.Event) error {
	// 获取群主信息
	memberInfo, err := robot.GetGroupMemberInfo(&onebot.GetGroupMemberInfoParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		NoCache: true,
	})

	if err != nil {
		log.Printf(common.T("", "social_get_member_failed_log"), err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: common.T("", "social_love_failed"),
		})
		return nil
	}

	if memberInfo == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: common.T("", "social_love_failed_no_info"),
		})
		return nil
	}

	// 发送爱群主消息
	loveMessages := []string{
		common.T("", "social_love_msg1"),
		common.T("", "social_love_msg2"),
		common.T("", "social_love_msg3"),
		common.T("", "social_love_msg4"),
		common.T("", "social_love_msg5"),
	}

	message := loveMessages[rand.Intn(len(loveMessages))]
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: message,
	})

	return nil
}

func (p *SocialPlugin) handleTransformCommand(robot plugin.Robot, event *onebot.Event) error {
	// 随机选择一个头衔
	title := p.titles[rand.Intn(len(p.titles))]

	// 发送变身消息
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf(common.T("", "social_transform_msg"), event.Sender.Nickname, title),
	})

	return nil
}

func (p *SocialPlugin) handleTitleCommand(robot plugin.Robot, event *onebot.Event) error {
	// 解析命令参数
	msgStr, ok := event.Message.(string)
	if !ok {
		return nil
	}
	parts := strings.Fields(msgStr)
	if len(parts) < 2 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: common.T("", "social_title_usage"),
		})
		return nil
	}

	// 获取自定义头衔
	customTitle := strings.Join(parts[1:], " ")
	if len(customTitle) > 10 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: common.T("", "social_title_too_long"),
		})
		return nil
	}

	// 设置群成员头衔
	_, err := robot.SetGroupSpecialTitle(&onebot.SetGroupSpecialTitleParams{
		GroupID:      event.GroupID,
		UserID:       event.UserID,
		SpecialTitle: customTitle,
		Duration:     -1, // 永久
	})

	if err != nil {
		log.Printf(common.T("", "social_set_title_failed_log"), err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: common.T("", "social_set_title_failed"),
		})
		return nil
	}

	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf(common.T("", "social_set_title_success"), customTitle),
	})

	return nil
}
