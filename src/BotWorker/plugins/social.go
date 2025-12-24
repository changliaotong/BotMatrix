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
	// 命令解析器
	cmdParser *CommandParser
}

func (p *SocialPlugin) Name() string {
	return "social"
}

func (p *SocialPlugin) Description() string {
	return common.T("", "social_plugin_desc|社交功能插件，提供变身、设置头衔等社交互动功能")
}

func (p *SocialPlugin) Version() string {
	return "1.0.0"
}

// NewSocialPlugin 创建social plugin实例
func NewSocialPlugin() *SocialPlugin {
	p := &SocialPlugin{
		cmdParser: NewCommandParser(),
	}
	p.initTitles()
	return p
}

func (p *SocialPlugin) initTitles() {
	p.titles = []string{
		common.T("", "social_title_owner|群主"), common.T("", "social_title_admin|管理员"), common.T("", "social_title_svip|超级会员"), common.T("", "social_title_vip|会员"),
		common.T("", "social_title_user|普通用户"), common.T("", "social_title_newbie|萌新"), common.T("", "social_title_master|大师"), common.T("", "social_title_scholar|学者"),
		common.T("", "social_title_loser|屌丝"), common.T("", "social_title_gamer|游戏玩家"), common.T("", "social_title_music_fan|音乐迷"), common.T("", "social_title_foodie|吃货"),
		common.T("", "social_title_traveler|旅行者"), common.T("", "social_title_photographer|摄影师"), common.T("", "social_title_writer|作家"), common.T("", "social_title_painter|画家"),
		common.T("", "social_title_designer|设计师"), common.T("", "social_title_coder|程序员"), common.T("", "social_title_engineer|工程师"), common.T("", "social_title_doctor|医生"),
		common.T("", "social_title_teacher|教师"), common.T("", "social_title_student|学生"), common.T("", "social_title_worker|工人"), common.T("", "social_title_freelancer|自由职业者"),
		common.T("", "social_title_entrepreneur|企业家"), common.T("", "social_title_investor|投资者"), common.T("", "social_title_collector|收藏家"), common.T("", "social_title_fitness|健身达人"),
		common.T("", "social_title_athlete|运动员"), common.T("", "social_title_eater|美食家"), common.T("", "social_title_sleeper|睡神"), common.T("", "social_title_procrastinator|拖延症患者"),
		common.T("", "social_title_ocd|强迫症"), common.T("", "social_title_indecisive|纠结帝"), common.T("", "social_title_lost|路痴"), common.T("", "social_title_blind|脸盲"),
		common.T("", "social_title_tone_deaf|音痴"), common.T("", "social_title_clumsy|手残党"), common.T("", "social_title_clean_freak|洁癖"), common.T("", "social_title_night_owl|熬夜冠军"),
		common.T("", "social_title_early_bird|早起鸟"), common.T("", "social_title_social_butterfly|交际花"), common.T("", "social_title_social_phobia|社交恐惧"), common.T("", "social_title_social_cow|社交牛逼症"),
		common.T("", "social_title_troll|喷子"), common.T("", "social_title_joker|小丑"), common.T("", "social_title_meme_master|表情包达人"), common.T("", "social_title_binge_watcher|刷剧狂人"),
		common.T("", "social_title_idol_fan|追星族"), common.T("", "social_title_2d|二次元"), common.T("", "social_title_3d|三次元"), common.T("", "social_title_4d|四次元"),
		common.T("", "social_title_coser|Coser"), common.T("", "social_title_editor|编辑"), common.T("", "social_title_video_editor|剪辑师"), common.T("", "social_title_screenwriter|编剧"),
		common.T("", "social_title_director|导演"), common.T("", "social_title_actor|演员"), common.T("", "social_title_singer|歌手"), common.T("", "social_title_dancer|舞者"),
		common.T("", "social_title_musician|音乐家"), common.T("", "social_title_producer|制作人"), common.T("", "social_title_streamer|主播"), common.T("", "social_title_up|UP主"),
		common.T("", "social_title_blogger|博主"), common.T("", "social_title_influencer|网红"), common.T("", "social_title_star|明星"), common.T("", "social_title_idol|偶像"),
		common.T("", "social_title_god|男神"), common.T("", "social_title_goddess|女神"), common.T("", "social_title_handsome|帅哥"), common.T("", "social_title_beauty|美女"),
		common.T("", "social_title_cute_girl|萌妹"), common.T("", "social_title_big_sister|御姐"), common.T("", "social_title_loli|萝莉"), common.T("", "social_title_shota|正太"),
		common.T("", "social_title_uncle|大叔"), common.T("", "social_title_aunt|阿姨"), common.T("", "social_title_little_sister|小姐姐"), common.T("", "social_title_little_brother|小哥哥"),
	}
}

func (p *SocialPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "social_plugin_loaded|加载社交功能插件"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 检查功能是否启用
		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "social") {
			HandleFeatureDisabled(robot, event, "social")
			return nil
		}

		// 处理爱群主命令
		if match, _ := p.cmdParser.MatchCommand("爱群主|loveowner", event.RawMessage); match {
			_, err := p.handleLoveOwnerLogic(robot, event)
			return err
		}

		// 处理变身命令
		if match, _ := p.cmdParser.MatchCommand("变身|transform", event.RawMessage); match {
			_, err := p.handleTransformLogic(robot, event)
			return err
		}

		// 处理头衔命令
		match, _, params := p.cmdParser.MatchCommandWithParams("头衔|title", "(.+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, common.T("", "social_title_usage|❌ 请输入想要设置的头衔，例如：头衔 大帅哥"))
				return nil
			}
			_, err := p.handleSetTitleLogic(robot, event, params[0])
			return err
		}

		return nil
	})
}

// GetSkills 实现 SkillCapable 接口
func (p *SocialPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "love_owner",
			Description: common.T("", "social_skill_love_owner_desc|向群主表达爱意"),
			Usage:       "love_owner",
		},
		{
			Name:        "transform",
			Description: common.T("", "social_skill_transform_desc|随机获得一个酷炫的头衔"),
			Usage:       "transform",
		},
		{
			Name:        "set_title",
			Description: common.T("", "social_skill_set_title_desc|自定义你的群头衔"),
			Usage:       "set_title <title>",
			Params: map[string]string{
				"title": common.T("", "social_skill_set_title_param_title|想要设置的头衔内容"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *SocialPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "love_owner":
		return p.handleLoveOwnerLogic(robot, event)
	case "transform":
		return p.handleTransformLogic(robot, event)
	case "set_title":
		title := params["title"]
		if title == "" {
			msg := common.T("", "social_title_usage|❌ 请输入想要设置的头衔，例如：头衔 大帅哥")
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		return p.handleSetTitleLogic(robot, event, title)
	}
	return "", nil
}

func (p *SocialPlugin) handleLoveOwnerLogic(robot plugin.Robot, event *onebot.Event) (string, error) {
	// 发送爱群主消息
	loveMessages := []string{
		common.T("", "social_love_msg1|群主大大我爱你，就像老鼠爱大米！"),
		common.T("", "social_love_msg2|群主最帅了，简直是本群的颜值担当！"),
		common.T("", "social_love_msg3|群主辛苦了，为您递上一杯热茶 🍵"),
		common.T("", "social_love_msg4|群主万岁，愿您的智慧照亮本群！"),
		common.T("", "social_love_msg5|群主，您就是我的偶像！"),
	}

	message := loveMessages[rand.Intn(len(loveMessages))]
	p.sendMessage(robot, event, message)

	return message, nil
}

func (p *SocialPlugin) handleTransformLogic(robot plugin.Robot, event *onebot.Event) (string, error) {
	// 确保头衔列表已初始化
	if len(p.titles) == 0 {
		p.initTitles()
	}

	// 随机选择一个头衔
	title := p.titles[rand.Intn(len(p.titles))]

	// 发送变身消息
	msg := fmt.Sprintf(common.T("", "social_transform_msg|✨ %s 华丽变身，获得了头衔：【%s】！"), event.Sender.Nickname, title)
	p.sendMessage(robot, event, msg)

	return msg, nil
}

func (p *SocialPlugin) handleSetTitleLogic(robot plugin.Robot, event *onebot.Event, customTitle string) (string, error) {
	customTitle = strings.TrimSpace(customTitle)
	if len(customTitle) > 10 {
		msg := common.T("", "social_title_too_long|❌ 头衔太长了，最多只能设置10个字符哦")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	// 设置群成员头衔
	_, err := robot.SetGroupSpecialTitle(&onebot.SetGroupSpecialTitleParams{
		GroupID:      event.GroupID,
		UserID:       event.UserID,
		SpecialTitle: customTitle,
		Duration:     -1, // 永久
	})

	if err != nil {
		log.Printf(common.T("", "social_set_title_failed_log|设置头衔失败: %v"), err)
		msg := common.T("", "social_set_title_failed|❌ 设置头衔失败，请检查机器人是否具有管理员权限")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	msg := fmt.Sprintf(common.T("", "social_set_title_success|✅ 成功为 %s 设置了头衔：【%s】"), event.Sender.Nickname, customTitle)
	p.sendMessage(robot, event, msg)

	return msg, nil
}

// sendMessage 发送消息
func (p *SocialPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
