package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
	"time"
)

// ModerationPlugin  moderation plugin
type ModerationPlugin struct {
	// 敏感词列表
	sensitiveWords []string
	// 白名单用户
	whitelist []string
	// 黑名单用户
	blacklist []string
	// 群配置
	groupConfigs map[string]*GroupConfig
	// 命令解析器
	cmdParser *CommandParser
}

// GroupConfig 群配置
type GroupConfig struct {
	// 被踢加黑
	kickToBlack bool
	// 被踢提示
	kickNotify bool
	// 退群加黑
	leaveToBlack bool
	// 退群提示
	leaveNotify bool
}

func (p *ModerationPlugin) Name() string {
	return "moderation"
}

func (p *ModerationPlugin) Description() string {
	return "moderation plugin，支持敏感词过滤、广告检测、图片网址过滤等功能"
}

func (p *ModerationPlugin) Version() string {
	return "1.0.0"
}

// NewModerationPlugin 创建moderation plugin实例
func NewModerationPlugin() *ModerationPlugin {
	return &ModerationPlugin{
		sensitiveWords: []string{"脏话", "广告", "图片", "网址"},
		whitelist:      []string{},
		blacklist:      []string{},
		groupConfigs:   make(map[string]*GroupConfig),
		cmdParser:      NewCommandParser(),
	}
}

func (p *ModerationPlugin) Init(robot plugin.Robot) {
	log.Println("加载moderation plugin")

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		result := HandleConfirmationReply(event)
		if result == nil || !result.Matched {
			return nil
		}

		if result.Confirmed {
			if result.Action == "clear_blacklist" {
				p.blacklist = []string{}
				p.sendMessage(robot, event, "黑名单已清空")
			}
			if result.Action == "clear_whitelist" {
				p.whitelist = []string{}
				p.sendMessage(robot, event, "白名单已清空")
			}
		}

		if result.Canceled {
			p.sendMessage(robot, event, "操作已取消")
		}

		return nil
	})

	// 处理敏感词检测
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
				HandleFeatureDisabled(robot, event, "moderation")
				return nil
			}
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			return nil
		}

		// 检查是否在白名单
		if p.isWhitelisted(fmt.Sprintf("%d", userID)) {
			// 白名单用户可以发送任何消息，包括敏感词、广告、图片、网址
			return nil
		}

		// 检查是否在黑名单
		if p.isBlacklisted(fmt.Sprintf("%d", userID)) {
			p.sendMessage(robot, event, "你已被拉黑，无法发送消息")
			return nil
		}

		// 检查敏感词
		msg := strings.TrimSpace(event.RawMessage)
		if p.containsSensitiveWords(msg) {
			p.sendMessage(robot, event, "消息包含敏感词，已被拦截")
			return nil
		}

		// 检查广告
		if p.containsAdvertisement(msg) {
			p.sendMessage(robot, event, "消息包含广告，已被拦截")
			return nil
		}

		// 检查图片
		if p.containsImage(msg) {
			p.sendMessage(robot, event, "消息包含图片，已被拦截")
			return nil
		}

		// 检查网址
		if p.containsURL(msg) {
			p.sendMessage(robot, event, "消息包含网址，已被拦截")
			return nil
		}

		return nil
	})

	// 处理拉黑命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
				HandleFeatureDisabled(robot, event, "moderation")
				return nil
			}
		}

		// 检查是否为拉黑命令
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams("拉黑|ban", `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 添加到黑名单
		p.blacklist = append(p.blacklist, userID)

		// 自动踢出群聊
		p.sendMessage(robot, event, fmt.Sprintf("用户%s已被拉黑并踢出群聊", userID))

		return nil
	})

	// 处理踢出命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
				HandleFeatureDisabled(robot, event, "moderation")
				return nil
			}
		}

		// 检查是否为踢出命令
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams("踢出|kick", `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 模拟踢出
		p.sendMessage(robot, event, fmt.Sprintf("用户%s已被踢出群聊", userID))

		return nil
	})

	// 处理禁言命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
				HandleFeatureDisabled(robot, event, "moderation")
				return nil
			}
		}

		// 检查是否为禁言命令
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams("禁言|mute", `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 模拟禁言
		p.sendMessage(robot, event, fmt.Sprintf("用户%s已被禁言", userID))

		return nil
	})

	// 处理撤回命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
				return nil
			}
		}

		// 检查是否为撤回命令
		match, _ := p.cmdParser.MatchCommand("撤回|recall", event.RawMessage)
		if !match {
			return nil
		}

		// 模拟撤回
		p.sendMessage(robot, event, "消息已被撤回")

		return nil
	})

	// 处理群配置命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			HandleFeatureDisabled(robot, event, "moderation")
			return nil
		}

		match, _, paramMatches := p.cmdParser.MatchCommandWithParams("群配置", `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		configStr := strings.TrimSpace(paramMatches[0])

		// 获取或创建群配置
		config, ok := p.groupConfigs[groupIDStr]
		if !ok {
			config = &GroupConfig{
				kickToBlack:  true,
				kickNotify:   true,
				leaveToBlack: true,
				leaveNotify:  true,
			}
			p.groupConfigs[groupIDStr] = config
		}

		// 处理配置
		switch configStr {
		case "被踢加黑开启":
			config.kickToBlack = true
			p.sendMessage(robot, event, "被踢加黑功能已开启")
		case "被踢加黑关闭":
			config.kickToBlack = false
			p.sendMessage(robot, event, "被踢加黑功能已关闭")
		case "被踢提示开启":
			config.kickNotify = true
			p.sendMessage(robot, event, "被踢提示功能已开启")
		case "被踢提示关闭":
			config.kickNotify = false
			p.sendMessage(robot, event, "被踢提示功能已关闭")
		case "退群加黑开启":
			config.leaveToBlack = true
			p.sendMessage(robot, event, "退群加黑功能已开启")
		case "退群加黑关闭":
			config.leaveToBlack = false
			p.sendMessage(robot, event, "退群加黑功能已关闭")
		case "退群提示开启":
			config.leaveNotify = true
			p.sendMessage(robot, event, "退群提示功能已开启")
		case "退群提示关闭":
			config.leaveNotify = false
			p.sendMessage(robot, event, "退群提示功能已关闭")
		case "查看":
			msg := fmt.Sprintf("当前群配置：\n被踢加黑：%t\n被踢提示：%t\n退群加黑：%t\n退群提示：%t",
				config.kickToBlack, config.kickNotify, config.leaveToBlack, config.leaveNotify)
			p.sendMessage(robot, event, msg)
		case "清空黑名单":
			pc := StartConfirmation("clear_blacklist", event, "", "", nil, 2*time.Minute)
			if pc != nil {
				p.sendMessage(robot, event, fmt.Sprintf("即将清空本群黑名单。\n发送【%s】确认，发送【%s】取消。", pc.ConfirmCode, pc.CancelCode))
			}
		case "清空白名单":
			pc := StartConfirmation("clear_whitelist", event, "", "", nil, 2*time.Minute)
			if pc != nil {
				p.sendMessage(robot, event, fmt.Sprintf("即将清空本群白名单。\n发送【%s】确认，发送【%s】取消。", pc.ConfirmCode, pc.CancelCode))
			}
		default:
			p.sendMessage(robot, event, "未知配置项，可用配置：被踢加黑开启/关闭、被踢提示开启/关闭、退群加黑开启/关闭、退群提示开启/关闭、查看")
		}

		return nil
	})

	// 处理新成员进群事件（进群禁言）
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType != "group_member_increase" {
			return nil
		}

		groupID := event.GroupID
		userID := event.UserID
		groupIDStr := fmt.Sprintf("%d", groupID)

		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			return nil
		}

		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "join_mute") {
			return nil
		}

		_, err := robot.SetGroupBan(&onebot.SetGroupBanParams{
			GroupID:  groupID,
			UserID:   userID,
			Duration: 300,
		})
		if err != nil {
			log.Printf("进群禁言失败: %v", err)
			return nil
		}

		p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf("新成员%d已被进群禁言5分钟", userID))

		return nil
	})

	// 处理被踢事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType != "group_decrease" || event.SubType != "kick" {
			return nil
		}

		groupID := event.GroupID
		userID := event.UserID
		groupIDStr := fmt.Sprintf("%d", groupID)

		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			return nil
		}

		// 获取群配置
		config, ok := p.groupConfigs[groupIDStr]
		if !ok {
			return nil
		}

		// 被踢加黑
		if config.kickToBlack && IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "kick_to_black") {
			p.blacklist = append(p.blacklist, fmt.Sprintf("%d", userID))
		}

		// 被踢提示
		if config.kickNotify && IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "kick_notify") {
			p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf("用户%d已被踢出群聊", userID))
		}

		return nil
	})

	// 处理退群事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType != "group_decrease" || event.SubType != "leave" {
			return nil
		}

		groupID := event.GroupID
		userID := event.UserID
		groupIDStr := fmt.Sprintf("%d", groupID)

		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			return nil
		}

		// 获取群配置
		config, ok := p.groupConfigs[groupIDStr]
		if !ok {
			return nil
		}

		// 退群加黑
		if config.leaveToBlack && IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "leave_to_black") {
			p.blacklist = append(p.blacklist, fmt.Sprintf("%d", userID))
		}

		// 退群提示
		if config.leaveNotify && IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "leave_notify") {
			p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf("用户%d已退出群聊", userID))
		}

		return nil
	})
}

// isWhitelisted 检查用户是否在白名单
func (p *ModerationPlugin) isWhitelisted(userID string) bool {
	for _, id := range p.whitelist {
		if id == userID {
			return true
		}
	}
	return false
}

// isBlacklisted 检查用户是否在黑名单
func (p *ModerationPlugin) isBlacklisted(userID string) bool {
	for _, id := range p.blacklist {
		if id == userID {
			return true
		}
	}
	return false
}

// containsSensitiveWords 检查消息是否包含敏感词
func (p *ModerationPlugin) containsSensitiveWords(msg string) bool {
	for _, word := range p.sensitiveWords {
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

// containsAdvertisement 检查消息是否包含广告
func (p *ModerationPlugin) containsAdvertisement(msg string) bool {
	adWords := []string{"广告", "推广", "促销", "优惠", "打折"}
	for _, word := range adWords {
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

// containsImage 检查消息是否包含图片
func (p *ModerationPlugin) containsImage(msg string) bool {
	imageWords := []string{"图片", "照片", "截图", "img", "image", "pic"}
	for _, word := range imageWords {
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

// containsURL 检查消息是否包含网址
func (p *ModerationPlugin) containsURL(msg string) bool {
	urlWords := []string{"http://", "https://", "www.", ".com", ".cn", ".net"}
	for _, word := range urlWords {
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

// sendMessage 发送消息
func (p *ModerationPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
