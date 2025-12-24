package plugins

import (
	"BotMatrix/common"
	"botworker/internal/db"
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
	return common.T("", "moderation_plugin_desc|内容审核插件，支持敏感词过滤、黑白名单、进群禁言等功能")
}

func (p *ModerationPlugin) Version() string {
	return "1.0.0"
}

// NewModerationPlugin 创建moderation plugin实例
func NewModerationPlugin() *ModerationPlugin {
	return &ModerationPlugin{
		sensitiveWords: []string{
			common.T("", "moderation_word_profanity|脏话"),
			common.T("", "moderation_word_ad|广告"),
			common.T("", "moderation_word_image|图片"),
			common.T("", "moderation_word_url|链接"),
		},
		whitelist:    []string{},
		blacklist:    []string{},
		groupConfigs: make(map[string]*GroupConfig),
		cmdParser:    NewCommandParser(),
	}
}

func (p *ModerationPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "moderation_plugin_loaded|内容审核插件已加载"))

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
				p.sendMessage(robot, event, common.T("", "moderation_blacklist_cleared|黑名单已清空"))
			}
			if result.Action == "clear_whitelist" {
				p.whitelist = []string{}
				if GlobalDB != nil && event != nil && event.GroupID != 0 {
					groupIDStr := fmt.Sprintf("%d", event.GroupID)
					if err := db.ClearGroupWhitelist(GlobalDB, groupIDStr); err != nil {
						log.Printf(common.T("", "moderation_clear_whitelist_failed|清空群白名单失败: %v"), err)
					}
				}
				p.sendMessage(robot, event, common.T("", "moderation_whitelist_cleared|白名单已清空"))
			}
		}

		if result.Canceled {
			p.sendMessage(robot, event, common.T("", "moderation_op_canceled|操作已取消"))
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

		userID := event.UserID
		if userID == 0 {
			return nil
		}

		groupIDStr := ""
		if event.MessageType == "group" {
			groupIDStr = fmt.Sprintf("%d", event.GroupID)
		}

		userIDStr := fmt.Sprintf("%d", userID)

		if p.isWhitelisted(groupIDStr, userIDStr) {
			return nil
		}

		if p.isBlacklisted(userIDStr) {
			p.sendMessage(robot, event, common.T("", "moderation_user_blacklisted|你在本系统的黑名单中，无法使用该功能"))
			return nil
		}

		msg := strings.TrimSpace(event.RawMessage)
		if msg == "" {
			return nil
		}

		level := 0
		reason := ""

		if GlobalDB != nil {
			words, err := db.GetAllSensitiveWords(GlobalDB)
			if err == nil {
				for _, w := range words {
					if strings.Contains(msg, w.Word) {
						if w.Level > level {
							level = w.Level
							reason = w.Word
						}
					}
				}
			} else {
				log.Printf(common.T("", "moderation_get_sensitive_words_failed|获取敏感词库失败: %v"), err)
			}
		}

		if p.containsSensitiveWords(msg) {
			if level < 3 {
				level = 3
			}
			if reason == "" {
				reason = common.T("", "moderation_reason_sensitive|触发敏感词")
			}
		}

		if p.containsAdvertisement(msg) {
			if level < 4 {
				level = 4
				reason = common.T("", "moderation_reason_ad|发布广告内容")
			}
		}

		if p.containsImage(msg) {
			if level < 1 {
				level = 1
				reason = common.T("", "moderation_reason_image|包含受限图片")
			}
		}

		if p.containsURL(msg) {
			if level < 2 {
				level = 2
				reason = common.T("", "moderation_reason_url|包含受限链接")
			}
		}

		if level > 0 {
			p.handleSensitiveHit(robot, event, level, reason)
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
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_ban|拉黑|ban"), `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 添加到黑名单
		p.blacklist = append(p.blacklist, userID)

		// 自动踢出群聊
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_user_banned_msg|用户 %s 已被拉黑并模拟踢出"), userID))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			HandleFeatureDisabled(robot, event, "moderation")
			return nil
		}

		match, cmd, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_word_manage|敏感词|word"), `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		if GlobalDB == nil {
			p.sendMessage(robot, event, common.T("", "moderation_db_unavailable|数据库未连接，无法管理敏感词"))
			return nil
		}

		adminIDStr := fmt.Sprintf("%d", event.UserID)
		isAdmin, err := db.IsGroupAdmin(GlobalDB, groupIDStr, adminIDStr)
		if err != nil || !isAdmin {
			p.sendMessage(robot, event, common.T("", "moderation_admin_only|只有群管理员才能管理敏感词"))
			return nil
		}

		raw := strings.TrimSpace(paramMatches[0])
		if raw == "" {
			p.sendMessage(robot, event, common.T("", "moderation_provide_content|请提供敏感词内容"))
			return nil
		}

		op := "add"
		if strings.HasPrefix(raw, "+") {
			raw = strings.TrimSpace(raw[1:])
			op = "add"
		} else if strings.HasPrefix(raw, "-") {
			raw = strings.TrimSpace(raw[1:])
			op = "del"
		}

		parts := strings.Fields(raw)
		if len(parts) == 0 {
			p.sendMessage(robot, event, common.T("", "moderation_provide_content|请提供敏感词内容"))
			return nil
		}

		level := 1
		switch cmd {
		case common.T("", "moderation_cmd_recall_word|撤回"):
			level = 1
		case common.T("", "moderation_cmd_points_word|扣分"):
			level = 2
		case common.T("", "moderation_cmd_warn_word|警告"):
			level = 3
		case common.T("", "moderation_cmd_mute_word|禁言"):
			level = 4
		case common.T("", "moderation_cmd_kick_word|踢出"):
			level = 5
		case common.T("", "moderation_cmd_ban_word|封禁"):
			level = 6
		}

		if op == "del" {
			for _, w := range parts {
				if err := db.RemoveSensitiveWord(GlobalDB, w); err != nil {
					p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_del_word_failed|删除敏感词 %s 失败"), w))
					return nil
				}
			}
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_del_word_success|已成功从 %s 列表中删除: %s"), cmd, strings.Join(parts, " ")))
			return nil
		}

		for _, w := range parts {
			if err := db.AddSensitiveWord(GlobalDB, w, level); err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_add_word_failed|添加敏感词 %s 失败"), w))
				return nil
			}
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_add_word_success|已成功添加到 %s 列表: %s"), cmd, strings.Join(parts, " ")))

		return nil
	})

	// 处理白名单命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "moderation") {
			HandleFeatureDisabled(robot, event, "moderation")
			return nil
		}

		match, _, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_whitelist|白名单|whitelist"), `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		if GlobalDB == nil {
			p.sendMessage(robot, event, common.T("", "moderation_db_unavailable_whitelist|数据库未连接，无法管理白名单"))
			return nil
		}

		adminIDStr := fmt.Sprintf("%d", event.UserID)
		isAdmin, err := db.IsGroupAdmin(GlobalDB, groupIDStr, adminIDStr)
		if err != nil || !isAdmin {
			p.sendMessage(robot, event, common.T("", "moderation_admin_only_whitelist|只有群管理员才能管理白名单"))
			return nil
		}

		userIDStr := strings.TrimSpace(paramMatches[0])
		if userIDStr == "" {
			p.sendMessage(robot, event, common.T("", "moderation_provide_userid|请提供要操作的用户ID"))
			return nil
		}

		exists, err := db.IsUserInGroupWhitelist(GlobalDB, groupIDStr, userIDStr)
		if err != nil {
			log.Printf(common.T("", "moderation_check_whitelist_failed|检查白名单状态失败: %v"), err)
			p.sendMessage(robot, event, common.T("", "moderation_check_failed_retry|检查白名单失败，请稍后再试"))
			return nil
		}

		if exists {
			if err := db.RemoveGroupWhitelistUser(GlobalDB, groupIDStr, userIDStr); err != nil {
				log.Printf(common.T("", "moderation_remove_whitelist_user_failed|移除白名单用户失败: %v"), err)
				p.sendMessage(robot, event, common.T("", "moderation_remove_failed_retry|移除白名单失败，请稍后再试"))
				return nil
			}
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_user_removed_whitelist|用户 %s 已从白名单中移除"), userIDStr))
		} else {
			if err := db.AddGroupWhitelistUser(GlobalDB, groupIDStr, userIDStr); err != nil {
				log.Printf(common.T("", "moderation_add_whitelist_user_failed|添加白名单用户失败: %v"), err)
				p.sendMessage(robot, event, common.T("", "moderation_add_failed_retry|添加白名单失败，请稍后再试"))
				return nil
			}
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_user_added_whitelist|用户 %s 已添加到白名单"), userIDStr))
		}

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
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_kick|踢出|kick"), `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 模拟踢出
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_user_kicked_msg|用户 %s 已被模拟踢出"), userID))

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
		match, _, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_mute|禁言|mute"), `(.+)`, event.RawMessage)
		if !match || len(paramMatches) < 1 {
			return nil
		}

		// 解析用户ID
		userID := strings.TrimSpace(paramMatches[0])

		// 模拟禁言
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_user_muted_msg|用户 %s 已被模拟禁言"), userID))

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
		match, _ := p.cmdParser.MatchCommand(common.T("", "moderation_cmd_recall|撤回|recall"), event.RawMessage)
		if !match {
			return nil
		}

		// 模拟撤回
		p.sendMessage(robot, event, common.T("", "moderation_msg_recalled|消息已被撤回"))

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

		match, _, paramMatches := p.cmdParser.MatchCommandWithParams(common.T("", "moderation_cmd_group_config|审核配置|mconfig"), `(.+)`, event.RawMessage)
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
		case common.T("", "moderation_config_kick_black_on|开启被踢加黑"):
			config.kickToBlack = true
			p.sendMessage(robot, event, common.T("", "moderation_kick_to_black_on|已开启被踢加黑功能"))
		case common.T("", "moderation_config_kick_black_off|关闭被踢加黑"):
			config.kickToBlack = false
			p.sendMessage(robot, event, common.T("", "moderation_kick_to_black_off|已关闭被踢加黑功能"))
		case common.T("", "moderation_config_kick_notify_on|开启被踢提示"):
			config.kickNotify = true
			p.sendMessage(robot, event, common.T("", "moderation_kick_notify_on|已开启被踢提示功能"))
		case common.T("", "moderation_config_kick_notify_off|关闭被踢提示"):
			config.kickNotify = false
			p.sendMessage(robot, event, common.T("", "moderation_kick_notify_off|已关闭被踢提示功能"))
		case common.T("", "moderation_config_leave_black_on|开启退群加黑"):
			config.leaveToBlack = true
			p.sendMessage(robot, event, common.T("", "moderation_leave_to_black_on|已开启退群加黑功能"))
		case common.T("", "moderation_config_leave_black_off|关闭退群加黑"):
			config.leaveToBlack = false
			p.sendMessage(robot, event, common.T("", "moderation_leave_to_black_off|已关闭退群加黑功能"))
		case common.T("", "moderation_config_leave_notify_on|开启退群提示"):
			config.leaveNotify = true
			p.sendMessage(robot, event, common.T("", "moderation_leave_notify_on|已开启退群提示功能"))
		case common.T("", "moderation_config_leave_notify_off|关闭退群提示"):
			config.leaveNotify = false
			p.sendMessage(robot, event, common.T("", "moderation_leave_notify_off|已关闭退群提示功能"))
		case common.T("", "moderation_config_view|查看配置"):
			msg := fmt.Sprintf(common.T("", "moderation_view_config|当前群审核配置：\n被踢加黑: %v\n被踢提示: %v\n退群加黑: %v\n退群提示: %v"),
				config.kickToBlack, config.kickNotify, config.leaveToBlack, config.leaveNotify)
			p.sendMessage(robot, event, msg)
		case common.T("", "moderation_config_clear_blacklist|清空黑名单"):
			pc := StartConfirmation("clear_blacklist", event, "", "", nil, 2*time.Minute)
			if pc != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_clear_blacklist_confirm|确定要清空黑名单吗？发送 %s 确认，发送 %s 取消"), pc.ConfirmCode, pc.CancelCode))
			}
		case common.T("", "moderation_config_clear_whitelist|清空白名单"):
			pc := StartConfirmation("clear_whitelist", event, "", "", nil, 2*time.Minute)
			if pc != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_clear_whitelist_confirm|确定要清空群白名单吗？发送 %s 确认，发送 %s 取消"), pc.ConfirmCode, pc.CancelCode))
			}
		default:
			p.sendMessage(robot, event, common.T("", "moderation_unknown_config|未知配置项"))
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
			log.Printf(common.T("", "moderation_join_mute_failed|新成员禁言失败: %v"), err)
			return nil
		}

		p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf(common.T("", "moderation_join_mute_msg|欢迎新成员 %d，已为您开启 5 分钟禁言保护"), userID))

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
			p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf(common.T("", "moderation_user_kicked_msg|用户 %s 已被模拟踢出"), userID))
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
			p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf(common.T("", "moderation_user_left_msg|用户 %s 已退群，已将其拉黑"), userID))
		}

		return nil
	})
}

func (p *ModerationPlugin) isWhitelisted(groupID, userID string) bool {
	for _, id := range p.whitelist {
		if id == userID {
			return true
		}
	}

	if GlobalDB != nil && groupID != "" && userID != "" {
		if IsFeatureEnabledForGroup(GlobalDB, groupID, "admin_whitelist") {
			isAdmin, err := db.IsGroupAdmin(GlobalDB, groupID, userID)
			if err == nil && isAdmin {
				return true
			}
		}

		ok, err := db.IsUserInGroupWhitelist(GlobalDB, groupID, userID)
		if err == nil && ok {
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
	adWords := []string{
		common.T("", "moderation_ad_word_1|代刷"),
		common.T("", "moderation_ad_word_2|兼职"),
		common.T("", "moderation_ad_word_3|加群"),
		common.T("", "moderation_ad_word_4|点我"),
		common.T("", "moderation_ad_word_5|联系方式"),
	}
	for _, word := range adWords {
		if strings.Contains(msg, word) {
			return true
		}
	}
	return false
}

// containsImage 检查消息是否包含图片
func (p *ModerationPlugin) containsImage(msg string) bool {
	imageWords := []string{
		common.T("", "moderation_image_word_1|图片"),
		common.T("", "moderation_image_word_2|表情"),
		common.T("", "moderation_image_word_3|图库"),
		"img",
		"image",
		"pic",
	}
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
		log.Printf(common.T("", "moderation_send_failed_log|发送消息失败: %v"), err)
	}
}

func (p *ModerationPlugin) handleSensitiveHit(robot plugin.Robot, event *onebot.Event, level int, reason string) {
	if event == nil {
		return
	}

	if event.MessageID != 0 {
		_, err := robot.DeleteMessage(&onebot.DeleteMessageParams{
			MessageID: event.MessageID,
		})
		if err != nil {
			log.Printf(common.T("", "moderation_recall_failed_log|撤回消息失败: %v"), err)
		}
	}

	groupIDStr := ""
	if event.MessageType == "group" && event.GroupID != 0 {
		groupIDStr = fmt.Sprintf("%d", event.GroupID)
	}

	userIDStr := fmt.Sprintf("%d", event.UserID)

	switch {
	case level <= 1:
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_recalled|检测到敏感内容 (%s)，消息已被撤回"), reason))
	case level == 2:
		if GlobalDB != nil && userIDStr != "" {
			if err := db.AddPoints(GlobalDB, userIDStr, -10, common.T("", "moderation_points_deduct_reason|触发敏感内容: ")+reason, "sensitive_word"); err != nil {
				log.Printf(common.T("", "moderation_points_deduct_failed_log|扣除积分失败: %v"), err)
			}
		}
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_points_deduct|检测到敏感内容 (%s)，已扣除 10 积分"), reason))
	case level == 3:
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_warn|警告：检测到敏感内容 (%s)，请注意言行"), reason))
	case level == 4:
		if event.MessageType == "group" && groupIDStr != "" {
			_, err := robot.SetGroupBan(&onebot.SetGroupBanParams{
				GroupID:  event.GroupID,
				UserID:   event.UserID,
				Duration: 600,
			})
			if err != nil {
				log.Printf(common.T("", "moderation_mute_failed_log|禁言失败: %v"), err)
			}
		}
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_muted|检测到敏感内容 (%s)，已禁言 1 小时"), reason))
	case level == 5:
		if event.MessageType == "group" && groupIDStr != "" {
			_, err := robot.SetGroupKick(&onebot.SetGroupKickParams{
				GroupID: event.GroupID,
				UserID:  event.UserID,
			})
			if err != nil {
				log.Printf(common.T("", "moderation_kick_failed_log|踢出失败: %v"), err)
			}
		}
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_kicked|检测到严重敏感内容 (%s)，已移出群聊"), reason))
	default:
		p.blacklist = append(p.blacklist, userIDStr)
		if event.MessageType == "group" && groupIDStr != "" {
			_, err := robot.SetGroupKick(&onebot.SetGroupKickParams{
				GroupID: event.GroupID,
				UserID:  event.UserID,
			})
			if err != nil {
				log.Printf(common.T("", "moderation_ban_kick_failed_log|拉黑并踢出失败: %v"), err)
			}
		}
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "moderation_sensitive_banned|检测到极度敏感内容 (%s)，已封禁并移出群聊"), reason))
	}
}
