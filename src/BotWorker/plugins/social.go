package plugins

import (
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
	return "social plugin，支持爱群主、变身、头衔等功能"
}

func (p *SocialPlugin) Version() string {
	return "1.0.0"
}

// NewSocialPlugin 创建social plugin实例
func NewSocialPlugin() *SocialPlugin {
	return &SocialPlugin{
		titles: []string{
			"群主大大", "管理员", "超级会员", "VIP", "普通用户", "萌新", "大佬", "学霸", "学渣",
			"游戏高手", "音乐达人", "美食家", "旅行家", "摄影师", "作家", "画家", "设计师",
			"程序员", "工程师", "医生", "老师", "学生", "上班族", "自由职业者", "创业者",
			"投资者", "收藏家", "健身达人", "运动健将", "吃货", "睡神", "拖延症患者",
			"强迫症患者", "选择困难症患者", "路痴", "脸盲", "音痴", "手残党", "强迫症",
			"洁癖", "夜猫子", "早起鸟", "社交达人", "社恐", "社牛", "吐槽帝", "段子手",
			"表情包达人", "追剧狂魔", "追星族", "二次元", "三次元", "四次元", "coser",
			"后期", "剪辑师", "编剧", "导演", "演员", "歌手", "舞者", "音乐人", "制作人",
			"主播", "UP主", "博主", "网红", "明星", "偶像", "男神", "女神", "帅哥", "美女",
			"萌妹", "御姐", "萝莉", "正太", "大叔", "阿姨", "小姐姐", "小哥哥",
		},
	}
}

func (p *SocialPlugin) Init(robot plugin.Robot) {
	log.Println("加载social插件")

	// 处理群消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 处理爱群主命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, "爱群主") {
			return p.handleLoveOwnerCommand(robot, event)
		}

		// 处理变身命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, "变身") {
			return p.handleTransformCommand(robot, event)
		}

		// 处理头衔命令
		if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, "头衔") {
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
		log.Printf("[Social] 获取群成员信息失败: %v", err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "❤️ 爱群主失败，请稍后重试！",
		})
		return nil
	}

	if memberInfo == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "❤️ 爱群主失败，无法获取群成员信息！",
		})
		return nil
	}

	// 发送爱群主消息
	loveMessages := []string{
		"❤️ 群主大大最棒了！",
		"❤️ 爱群主，群主最帅！",
		"❤️ 群主是大家的榜样！",
		"❤️ 感谢群主的辛勤付出！",
		"❤️ 群主威武霸气！",
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
		Message: fmt.Sprintf("✨ %s 变身为：%s", event.Sender.Nickname, title),
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
			Message: "用法：头衔 [自定义头衔]\n例如：头衔 游戏大神",
		})
		return nil
	}

	// 获取自定义头衔
	customTitle := strings.Join(parts[1:], " ")
	if len(customTitle) > 10 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "头衔长度不能超过10个字符！",
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
		log.Printf("[Social] 设置群成员头衔失败: %v", err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "❌ 设置头衔失败，请稍后重试！",
		})
		return nil
	}

	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("✅ 已成功设置头衔：%s", customTitle),
	})

	return nil
}
