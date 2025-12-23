package plugins

import (
	"BotMatrix/common"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type AdminPlugin struct {
	admins    []string
	db        *sql.DB
	cmdParser *CommandParser
}

func (p *AdminPlugin) Name() string {
	return "admin"
}

func (p *AdminPlugin) Description() string {
	return common.T("", "admin_plugin_desc")
}

func (p *AdminPlugin) Version() string {
	return "1.0.0"
}

// NewAdminPlugin 创建admin plugin实例
func NewAdminPlugin(database *sql.DB) *AdminPlugin {
	return &AdminPlugin{
		admins:    []string{},
		db:        database,
		cmdParser: NewCommandParser(),
	}
}

func (p *AdminPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "admin_plugin_loaded"))

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		matchNoArg, _ := p.cmdParser.MatchCommand("设置语音|setvoice", event.RawMessage)
		matchWithArg, _, arg := p.cmdParser.MatchCommandWithSingleParam("设置语音|setvoice", event.RawMessage)

		if !matchNoArg && !matchWithArg {
			return nil
		}

		if event.MessageType != "group" {
			p.sendMessage(robot, event, common.T("", "admin_group_only_voice"))
			return nil
		}

		if p.db == nil {
			p.sendMessage(robot, event, common.T("", "admin_no_db_voice"))
			return nil
		}

		groupID := fmt.Sprintf("%d", event.GroupID)

		if matchNoArg && !matchWithArg {
			currentID, _ := db.GetGroupVoiceID(p.db, groupID)
			list := BuildVoiceList(currentID)
			msg := list + "\n" + common.T("", "admin_set_voice_usage")
			p.sendMessage(robot, event, msg)
			return nil
		}

		if !matchWithArg {
			return nil
		}

		input := strings.TrimSpace(arg)
		if input == "" {
			currentID, _ := db.GetGroupVoiceID(p.db, groupID)
			list := BuildVoiceList(currentID)
			msg := list + "\n" + common.T("", "admin_set_voice_usage")
			p.sendMessage(robot, event, msg)
			return nil
		}

		if num, err := strconv.Atoi(input); err == nil {
			item := FindVoiceByGlobalIndex(num)
			if item == nil {
				p.sendMessage(robot, event, "❌ "+common.T("", "admin_voice_not_found"))
				return nil
			}
			p.handleSaveGroupVoice(robot, event, groupID, item.ID, item.Name, "")
			return nil
		}

		if strings.EqualFold(input, "随机") || strings.EqualFold(input, "random") {
			item := GetRandomVoice()
			if item == nil {
				p.sendMessage(robot, event, "❌ "+common.T("", "admin_voice_list_not_supported"))
				return nil
			}
			p.handleSaveGroupVoice(robot, event, groupID, item.ID, item.Name, "（"+common.T("", "admin_random")+"）")
			return nil
		}

		if item := FindVoiceByName(input); item != nil {
			p.handleSaveGroupVoice(robot, event, groupID, item.ID, item.Name, "")
			return nil
		}

		if item := FindVoiceFuzzy(input); item != nil {
			p.handleSaveGroupVoice(robot, event, groupID, item.ID, item.Name, "（"+common.T("", "admin_fuzzy_match")+"）")
			return nil
		}

		p.sendMessage(robot, event, "❌ "+common.T("", "admin_voice_not_found_hint"))

		return nil
	})

	// 处理后台命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为后台命令
		if match, _ := p.cmdParser.MatchCommand("后台|admin", event.RawMessage); !match {
			return nil
		}

		// 发送后台菜单
		adminMenu := "🔧 " + common.T("", "admin_menu_title") + "\n"
		adminMenu += "====================\n"
		adminMenu += common.T("", "admin_menu_enable") + "\n"
		adminMenu += common.T("", "admin_menu_disable") + "\n"
		adminMenu += common.T("", "admin_menu_set") + "\n"
		adminMenu += common.T("", "admin_menu_teach") + "\n"
		adminMenu += common.T("", "admin_menu_group_info") + "\n"
		adminMenu += common.T("", "admin_menu_chatty") + "\n"
		adminMenu += common.T("", "admin_menu_ultimate") + "\n"
		adminMenu += common.T("", "admin_menu_agent") + "\n"
		p.sendMessage(robot, event, adminMenu)

		return nil
	})

	// 处理开启命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		match, _, params := p.cmdParser.MatchCommandWithParams("开启|enable", `(.*)`, event.RawMessage)
		if !match || len(params) < 1 {
			return nil
		}

		rawFeature := strings.TrimSpace(params[0])
		feature, requireAdmin, requireSuperAdmin := normalizeFeatureName(rawFeature)
		if feature == "" {
			feature = rawFeature
		}

		if event.MessageType == "group" && p.db != nil {
			if requireSuperAdmin {
				if !isSuperAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, common.T("", "admin_insufficient_perms_super"))
					return nil
				}
			} else if requireAdmin {
				if !isGroupAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, common.T("", "admin_insufficient_perms_admin"))
					return nil
				}
			}
		}
		defaultEnabled, ok := FeatureDefaults[feature]
		if !ok {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_feature_not_found"), feature))
			return nil
		}

		if event.MessageType != "group" {
			p.sendMessage(robot, event, common.T("", "admin_group_only_feature"))
			return nil
		}

		if p.db == nil {
			p.sendMessage(robot, event, common.T("", "admin_no_db_feature"))
			return nil
		}

		groupID := fmt.Sprintf("%d", event.GroupID)
		var err error
		if defaultEnabled {
			err = db.DeleteGroupFeatureOverride(p.db, groupID, feature)
		} else {
			err = db.SetGroupFeatureOverride(p.db, groupID, feature, true)
		}
		if err != nil {
			log.Printf("设置功能开启失败: %v", err)
			p.sendMessage(robot, event, fmt.Sprintf("开启功能%s失败", feature))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("功能%s已开启", feature))

		return nil
	})

	// 处理关闭命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		match, _, params := p.cmdParser.MatchCommandWithParams("关闭|disable", `(.*)`, event.RawMessage)
		if !match || len(params) < 1 {
			return nil
		}

		rawFeature := strings.TrimSpace(params[0])
		feature, requireAdmin, requireSuperAdmin := normalizeFeatureName(rawFeature)
		if feature == "" {
			feature = rawFeature
		}

		if event.MessageType == "group" && p.db != nil {
			if requireSuperAdmin {
				if !isSuperAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, common.T("", "admin_insufficient_perms_super"))
					return nil
				}
			} else if requireAdmin {
				if !isGroupAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, common.T("", "admin_insufficient_perms_admin"))
					return nil
				}
			}
		}
		defaultEnabled, ok := FeatureDefaults[feature]
		if !ok {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_feature_not_found"), feature))
			return nil
		}

		if event.MessageType != "group" {
			p.sendMessage(robot, event, common.T("", "admin_group_only_feature"))
			return nil
		}

		if p.db == nil {
			p.sendMessage(robot, event, common.T("", "admin_no_db_feature"))
			return nil
		}

		groupID := fmt.Sprintf("%d", event.GroupID)
		var err error
		if defaultEnabled {
			err = db.SetGroupFeatureOverride(p.db, groupID, feature, false)
		} else {
			err = db.DeleteGroupFeatureOverride(p.db, groupID, feature)
		}
		if err != nil {
			log.Printf("设置功能关闭失败: %v", err)
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_disable_feature_failed"), feature))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_feature_disabled"), feature))

		return nil
	})

	// 处理设置命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为设置命令
		match, _, params := p.cmdParser.MatchCommandWithParams("设置|set", `([^\s]+)\s+(.+)`, event.RawMessage)
		if !match || len(params) < 2 {
			return nil
		}

		// 解析参数和值
		param := params[0]
		value := params[1]

		// 模拟设置
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_param_set_success"), param, value))

		return nil
	})

	// 处理教学命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为教学命令
		if match, _ := p.cmdParser.MatchCommand("教学|help", event.RawMessage); !match {
			return nil
		}

		// 发送教学内容
		teaching := "📚 " + common.T("", "admin_tutorial_title") + "\n"
		teaching += "====================\n"
		teaching += "/菜单 - " + common.T("", "admin_help_menu") + "\n"
		teaching += "/help - " + common.T("", "admin_help_help") + "\n"
		teaching += "/签到 - " + common.T("", "admin_help_signin") + "\n"
		teaching += "/积分 - " + common.T("", "admin_help_points") + "\n"
		teaching += "/天气 <城市> - " + common.T("", "admin_help_weather") + "\n"
		teaching += "/翻译 <文本> - " + common.T("", "admin_help_translate") + "\n"
		teaching += "/点歌 <歌曲> - " + common.T("", "admin_help_music") + "\n"
		teaching += "/猜拳 <选择> - " + common.T("", "admin_help_rps") + "\n"
		teaching += "/猜大小 <选择> - " + common.T("", "admin_help_guess") + "\n"
		teaching += "/抽奖 - " + common.T("", "admin_help_lottery") + "\n"
		teaching += "/早安 - " + common.T("", "admin_help_morning") + "\n"
		teaching += "/晚安 - " + common.T("", "admin_help_night") + "\n"
		teaching += "/报时 - " + common.T("", "admin_help_time") + "\n"
		teaching += "/计算 <表达式> - " + common.T("", "admin_help_calc") + "\n"
		teaching += "/笑话 - " + common.T("", "admin_help_joke") + "\n"
		teaching += "/鬼故事 - " + common.T("", "admin_help_ghost") + "\n"
		teaching += "/成语接龙 <成语> - " + common.T("", "admin_help_idiom") + "\n"
		p.sendMessage(robot, event, teaching)

		return nil
	})

	// 处理本群命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为本群命令
		if match, _ := p.cmdParser.MatchCommand("本群|group", event.RawMessage); !match {
			return nil
		}

		// 发送本群信息
		groupInfo := "🏠 " + common.T("", "admin_group_info_title") + "\n"
		groupInfo += "====================\n"
		groupInfo += common.T("", "admin_group_name") + "：" + common.T("", "admin_unknown") + "\n"
		groupInfo += common.T("", "admin_group_member_count") + "：" + common.T("", "admin_unknown") + "\n"
		groupInfo += common.T("", "admin_group_create_time") + "：" + common.T("", "admin_unknown") + "\n"
		groupInfo += common.T("", "admin_group_notice") + "：" + common.T("", "admin_none") + "\n"
		p.sendMessage(robot, event, groupInfo)

		return nil
	})

	// 处理话唠命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("话唠|chatty", event.RawMessage); !match {
			return nil
		}

		if event.MessageType == "group" && p.db != nil {
			groupID := fmt.Sprintf("%d", event.GroupID)
			if err := db.SetGroupQAMode(p.db, groupID, "chatty"); err != nil {
				log.Printf("设置话唠模式失败: %v", err)
			}
		}

		p.sendMessage(robot, event, common.T("", "admin_chatty_mode_enabled"))

		return nil
	})

	// 处理终极命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("终极|ultimate", event.RawMessage); !match {
			return nil
		}

		if event.MessageType == "group" && p.db != nil {
			groupID := fmt.Sprintf("%d", event.GroupID)
			if err := db.SetGroupQAMode(p.db, groupID, "ultimate"); err != nil {
				log.Printf("设置终极模式失败: %v", err)
			}
		}

		p.sendMessage(robot, event, common.T("", "admin_ultimate_mode_enabled"))

		return nil
	})

	// 处理智能体命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("智能体|agent", event.RawMessage); !match {
			return nil
		}

		p.sendMessage(robot, event, common.T("", "admin_agent_mode_enabled"))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("闭嘴|silent", event.RawMessage); match {
			if event.MessageType == "group" && p.db != nil {
				groupID := fmt.Sprintf("%d", event.GroupID)
				if err := db.SetGroupQAMode(p.db, groupID, "silent"); err != nil {
					log.Printf("设置闭嘴模式失败: %v", err)
				}
			}
			p.sendMessage(robot, event, common.T("", "admin_silent_mode_enabled"))
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("本群模式|本群问答|本群", event.RawMessage); match {
			if event.MessageType == "group" && p.db != nil {
				groupID := fmt.Sprintf("%d", event.GroupID)
				if err := db.SetGroupQAMode(p.db, groupID, "group"); err != nil {
					log.Printf("设置本群模式失败: %v", err)
				}
			}
			p.sendMessage(robot, event, common.T("", "admin_group_mode_enabled"))
			return nil
		}

		if match, _ := p.cmdParser.MatchCommand("官方模式|官方问答|官方", event.RawMessage); match {
			if event.MessageType == "group" && p.db != nil {
				groupID := fmt.Sprintf("%d", event.GroupID)
				if err := db.SetGroupQAMode(p.db, groupID, "official"); err != nil {
					log.Printf("设置官方模式失败: %v", err)
				}
			}
			p.sendMessage(robot, event, common.T("", "admin_official_mode_enabled"))
			return nil
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *AdminPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

func (p *AdminPlugin) handleSaveGroupVoice(robot plugin.Robot, event *onebot.Event, groupID, voiceID, voiceName, suffix string) {
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "admin_no_db_voice"))
		return
	}

	if err := db.SetGroupVoiceID(p.db, groupID, voiceID); err != nil {
		log.Printf("设置群语音失败: %v", err)
		p.sendMessage(robot, event, "❌ "+common.T("", "admin_set_voice_failed"))
		return
	}

	categories := GetVoiceCategoriesForID(voiceID)
	categoryName := strings.Join(categories, "、")
	url := GetVoicePreviewURL(voiceID)

	msg := "✅ " + common.T("", "admin_set_voice_success") + voiceName
	if categoryName != "" {
		msg += "（" + categoryName + "）"
	}
	if suffix != "" {
		msg += suffix
	}
	if url != "" {
		msg += "\n" + common.T("", "admin_preview") + "：" + url
	}

	p.sendMessage(robot, event, msg)
}

func normalizeFeatureName(name string) (string, bool, bool) {
	n := strings.TrimSpace(name)
	if n == "" {
		return "", false, false
	}

	n = strings.ReplaceAll(n, "话痨", "话唠")
	n = strings.ReplaceAll(n, "加黑", "拉黑")
	n = strings.ReplaceAll(n, "模式", "")

	lower := strings.ToLower(n)
	if n == "语音" || strings.EqualFold(n, "AI声聊") || n == "声聊" || n == "声音" || lower == "voice" {
		n = "语音回复"
	}

	if n == "自动撤回" {
		n = "阅后即焚"
	}
	if n == "积分系统" {
		n = "积分"
	}
	if n == "积分" {
		n = "积分系统"
	}
	if n == "回复图片" {
		n = "图片回复"
	}
	if n == "回复撤回" {
		n = "撤回回复"
	}

	requireAdmin := false
	requireSuperAdmin := false
	featureID := ""

	switch n {
	case "功能关闭提示", "关闭提示", "功能提示":
		featureID = "feature_disabled_notice"
	case "欢迎语":
		featureID = "welcome"
	case "退群提示":
		featureID = "leave_notify"
	case "改名提示":
		featureID = "rename_notify"
	case "命令前缀":
		featureID = "command_prefix"
	case "进群改名":
		featureID = "join_rename"
	case "退群拉黑":
		featureID = "leave_to_black"
	case "被踢拉黑", "踢出拉黑":
		featureID = "kick_to_black"
	case "被踢提示":
		featureID = "kick_notify"
	case "进群禁言":
		featureID = "join_mute"
	case "道具系统":
		featureID = "props"
	case "宠物系统":
		featureID = "pets"
	case "群管系统", "敏感词", "敏感词系统":
		featureID = "moderation"
	case "简洁":
		featureID = "simple_mode"
	case "进群确认":
		featureID = "join_confirm"
	case "群链":
		featureID = "group_link"
	case "邀请统计":
		featureID = "invite_stats"
	case "AI":
		featureID = "ai"
	case "群主付":
		featureID = "owner_pay"
		requireSuperAdmin = true
	case "自动签到":
		featureID = "signin"
	case "权限提示":
		featureID = "permission_hint"
	case "云黑名单":
		featureID = "cloud_blacklist"
	case "管理加白":
		featureID = "admin_whitelist"
	case "多人互动":
		featureID = "multi_interaction"
	case "知识库":
		featureID = "knowledge_base"
	case "图片回复":
		featureID = "image_reply"
	case "撤回回复":
		featureID = "recall_reply"
	case "语音回复":
		featureID = "voice_reply"
	case "阅后即焚":
		featureID = "burn_after_reading"
	case "积分系统":
		featureID = "points"
	case "本群积分":
		featureID = "points"
		requireAdmin = true
	}

	return featureID, requireAdmin, requireSuperAdmin
}

func isGroupAdmin(database *sql.DB, groupID, userID int64) bool {
	if database == nil {
		return false
	}

	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	isAdmin, err := db.IsGroupAdmin(database, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("检查群 %d 中用户 %d 的管理员状态失败: %v", groupID, userID, err)
		return false
	}

	return isAdmin
}

func isSuperAdmin(database *sql.DB, groupID, userID int64) bool {
	if database == nil {
		return false
	}

	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	ok, err := db.IsSuperAdmin(database, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("检查群 %d 中用户 %d 的超级管理员状态失败: %v", groupID, userID, err)
		return false
	}

	return ok
}
