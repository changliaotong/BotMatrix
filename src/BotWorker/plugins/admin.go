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
	return common.T("", "admin_plugin_desc|管理插件，支持群管理、功能设置等功能")
}

func (p *AdminPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *AdminPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "set_voice",
			Description: common.T("", "admin_skill_set_voice_desc|设置群聊语音包"),
			Usage:       "set_voice group_id=654321 voice=1",
			Params: map[string]string{
				"group_id": common.T("", "admin_skill_set_voice_param_group_id|群号"),
				"voice":    common.T("", "admin_skill_set_voice_param_voice|语音包名称或编号"),
			},
		},
		{
			Name:        "enable_feature",
			Description: common.T("", "admin_skill_enable_feature_desc|开启群功能"),
			Usage:       "enable_feature group_id=654321 feature=signin user_id=123456",
			Params: map[string]string{
				"group_id": common.T("", "admin_skill_enable_feature_param_group_id|群号"),
				"feature":  common.T("", "admin_skill_enable_feature_param_feature|功能名称"),
				"user_id":  common.T("", "admin_skill_enable_feature_param_user_id|操作用户ID"),
			},
		},
		{
			Name:        "disable_feature",
			Description: common.T("", "admin_skill_disable_feature_desc|关闭群功能"),
			Usage:       "disable_feature group_id=654321 feature=signin user_id=123456",
			Params: map[string]string{
				"group_id": common.T("", "admin_skill_disable_feature_param_group_id|群号"),
				"feature":  common.T("", "admin_skill_disable_feature_param_feature|功能名称"),
				"user_id":  common.T("", "admin_skill_disable_feature_param_user_id|操作用户ID"),
			},
		},
		{
			Name:        "set_qa_mode",
			Description: common.T("", "admin_skill_set_qa_mode_desc|设置群聊问答模式"),
			Usage:       "set_qa_mode group_id=654321 mode=chatty",
			Params: map[string]string{
				"group_id": common.T("", "admin_skill_set_qa_mode_param_group_id|群号"),
				"mode":     common.T("", "admin_skill_set_qa_mode_param_mode|模式名称（chatty, ultimate, agent, silent, group, official）"),
			},
		},
	}
}

// HandleSkill 实现 SkillCapable 接口
func (p *AdminPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	var userID string
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if params["user_id"] != "" {
		userID = params["user_id"]
	}

	var groupID string
	if event != nil && event.MessageType == "group" {
		groupID = fmt.Sprintf("%d", event.GroupID)
	} else if params["group_id"] != "" {
		groupID = params["group_id"]
	}

	switch skillName {
	case "set_voice":
		voice := params["voice"]
		if groupID == "" {
			return "", fmt.Errorf("missing parameter: group_id")
		}
		return p.doSetVoice(groupID, voice), nil
	case "enable_feature":
		feature := params["feature"]
		if groupID == "" || feature == "" || userID == "" {
			return "", fmt.Errorf("missing parameter: group_id, feature or user_id")
		}
		return p.doEnableFeature(groupID, feature, userID)
	case "disable_feature":
		feature := params["feature"]
		if groupID == "" || feature == "" || userID == "" {
			return "", fmt.Errorf("missing parameter: group_id, feature or user_id")
		}
		return p.doDisableFeature(groupID, feature, userID)
	case "set_qa_mode":
		mode := params["mode"]
		if groupID == "" || mode == "" {
			return "", fmt.Errorf("missing parameter: group_id or mode")
		}
		return p.doSetQAMode(groupID, mode), nil
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
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
	log.Println(common.T("", "admin_plugin_loaded|管理插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
	skillName := skill.Name
	robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 统一消息处理器
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 1. 设置语音
		matchNoArg, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_set_voice|设置语音"), event.RawMessage)
		matchWithArg, _, arg := p.cmdParser.MatchCommandWithSingleParam(common.T("", "admin_cmd_set_voice|设置语音"), event.RawMessage)
		if matchNoArg || matchWithArg {
			if event.MessageType != "group" {
				p.sendMessage(robot, event, common.T("", "admin_group_only_voice|❌ 该命令仅限群聊使用。"))
				return nil
			}
			groupID := fmt.Sprintf("%d", event.GroupID)
			p.sendMessage(robot, event, p.doSetVoice(groupID, arg))
			return nil
		}

		// 2. 后台菜单
		if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_admin|后台"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doShowAdminMenu())
			return nil
		}

		// 3. 开启功能
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "admin_cmd_enable|开启"), `(.*)`, event.RawMessage); match && len(params) >= 1 {
			if event.MessageType != "group" {
				p.sendMessage(robot, event, common.T("", "admin_group_only_feature|❌ 该功能设置仅限群聊使用。"))
				return nil
			}
			groupID := fmt.Sprintf("%d", event.GroupID)
			userID := fmt.Sprintf("%d", event.UserID)
			feature := strings.TrimSpace(params[0])
			msg, _ := p.doEnableFeature(groupID, feature, userID)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 4. 关闭功能
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "admin_cmd_disable|关闭"), `(.*)`, event.RawMessage); match && len(params) >= 1 {
			if event.MessageType != "group" {
				p.sendMessage(robot, event, common.T("", "admin_group_only_feature|❌ 该功能设置仅限群聊使用。"))
				return nil
			}
			groupID := fmt.Sprintf("%d", event.GroupID)
			userID := fmt.Sprintf("%d", event.UserID)
			feature := strings.TrimSpace(params[0])
			msg, _ := p.doDisableFeature(groupID, feature, userID)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 5. 设置参数
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "admin_cmd_set|设置"), `([^\s]+)\s+(.+)`, event.RawMessage); match && len(params) >= 2 {
			param := params[0]
			value := params[1]
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "admin_param_set_success|✅ 参数 %s 已设置为：%s"), param, value))
			return nil
		}

		// 6. 教学内容
		if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_help|帮助"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doShowTeaching())
			return nil
		}

		// 7. 本群信息
		if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_group|本群"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doShowGroupInfo())
			return nil
		}

		// 8. 问答模式
		var mode string
		if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_chatty|话唠模式"), event.RawMessage); match {
			mode = "chatty"
		} else if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_ultimate|终极模式"), event.RawMessage); match {
			mode = "ultimate"
		} else if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_agent|代理模式"), event.RawMessage); match {
			mode = "agent"
		} else if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_silent|静默模式"), event.RawMessage); match {
			mode = "silent"
		} else if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_group_qa|本群问答"), event.RawMessage); match {
			mode = "group"
		} else if match, _ := p.cmdParser.MatchCommand(common.T("", "admin_cmd_official_qa|官方问答"), event.RawMessage); match {
			mode = "official"
		}

		if mode != "" {
			if event.MessageType == "group" {
				groupID := fmt.Sprintf("%d", event.GroupID)
				p.sendMessage(robot, event, p.doSetQAMode(groupID, mode))
			} else {
				p.sendMessage(robot, event, common.T("", "admin_group_only_feature|❌ 该功能设置仅限群聊使用。"))
			}
			return nil
		}

		return nil
	})
}

// doSetVoice 执行设置语音逻辑
func (p *AdminPlugin) doSetVoice(groupID string, voice string) string {
	if p.db == nil {
		return common.T("", "admin_no_db_voice|数据库未连接，无法设置语音。")
	}

	input := strings.TrimSpace(voice)
	if input == "" {
		currentID, _ := db.GetGroupVoiceID(p.db, groupID)
		list := BuildVoiceList(currentID)
		return list + "\n" + common.T("", "admin_set_voice_usage|💡 使用方法：/设置语音 <名称/编号/随机>")
	}

	var item *VoiceItem
	var suffix string

	if num, err := strconv.Atoi(input); err == nil {
		item = FindVoiceByGlobalIndex(num)
	} else if strings.EqualFold(input, "随机") || strings.EqualFold(input, "random") {
		item = GetRandomVoice()
		suffix = "（" + common.T("", "admin_random|随机") + "）"
	} else {
		item = FindVoiceByName(input)
		if item == nil {
			item = FindVoiceFuzzy(input)
			if item != nil {
				suffix = "（" + common.T("", "admin_fuzzy_match|模糊匹配") + "）"
			}
		}
	}

	if item == nil {
		return "❌ " + common.T("", "admin_voice_not_found_hint|未找到匹配的语音，请输入正确的名称或编号。")
	}

	if err := db.SetGroupVoiceID(p.db, groupID, item.ID); err != nil {
		log.Printf("设置群语音失败: %v", err)
		return "❌ " + common.T("", "admin_set_voice_failed|设置语音失败，请稍后再试。")
	}

	categories := GetVoiceCategoriesForID(item.ID)
	categoryName := strings.Join(categories, "、")
	url := GetVoicePreviewURL(item.ID)

	msg := "✅ " + common.T("", "admin_set_voice_success|设置成功！当前语音包：") + item.Name
	if categoryName != "" {
		msg += "（" + categoryName + "）"
	}
	if suffix != "" {
		msg += suffix
	}
	if url != "" {
		msg += "\n" + common.T("", "admin_preview|预览") + "：" + url
	}

	return msg
}

// doEnableFeature 执行开启功能逻辑
func (p *AdminPlugin) doEnableFeature(groupID string, rawFeature string, userID string) (string, error) {
	if p.db == nil {
		return common.T("", "admin_no_db_feature|数据库未连接，无法设置功能。"), nil
	}

	feature, requireAdmin, requireSuperAdmin := normalizeFeatureName(rawFeature)
	if feature == "" {
		feature = rawFeature
	}

	// 权限检查
	uid, _ := strconv.ParseInt(userID, 10, 64)
	gid, _ := strconv.ParseInt(groupID, 10, 64)

	if requireSuperAdmin {
		if !isSuperAdmin(p.db, gid, uid) {
			return common.T("", "admin_insufficient_perms_super|❌ 只有超级管理员才能操作此功能。"), nil
		}
	} else if requireAdmin {
		if !isGroupAdmin(p.db, gid, uid) {
			return common.T("", "admin_insufficient_perms_admin|❌ 只有群管理员才能操作此功能。"), nil
		}
	}

	defaultEnabled, ok := FeatureDefaults[feature]
	if !ok {
		return fmt.Sprintf(common.T("", "admin_feature_not_found|❌ 未找到功能：%s"), feature), nil
	}

	var err error
	if defaultEnabled {
		err = db.DeleteGroupFeatureOverride(p.db, groupID, feature)
	} else {
		err = db.SetGroupFeatureOverride(p.db, groupID, feature, true)
	}

	if err != nil {
		log.Printf("设置功能开启失败: %v", err)
		return fmt.Sprintf("开启功能%s失败", feature), err
	}

	return fmt.Sprintf("功能%s已开启", feature), nil
}

// doDisableFeature 执行关闭功能逻辑
func (p *AdminPlugin) doDisableFeature(groupID string, rawFeature string, userID string) (string, error) {
	if p.db == nil {
		return common.T("", "admin_no_db_feature|数据库未连接，无法设置功能。"), nil
	}

	feature, requireAdmin, requireSuperAdmin := normalizeFeatureName(rawFeature)
	if feature == "" {
		feature = rawFeature
	}

	// 权限检查
	uid, _ := strconv.ParseInt(userID, 10, 64)
	gid, _ := strconv.ParseInt(groupID, 10, 64)

	if requireSuperAdmin {
		if !isSuperAdmin(p.db, gid, uid) {
			return common.T("", "admin_insufficient_perms_super|❌ 只有超级管理员才能操作此功能。"), nil
		}
	} else if requireAdmin {
		if !isGroupAdmin(p.db, gid, uid) {
			return common.T("", "admin_insufficient_perms_admin|❌ 只有群管理员才能操作此功能。"), nil
		}
	}

	defaultEnabled, ok := FeatureDefaults[feature]
	if !ok {
		return fmt.Sprintf(common.T("", "admin_feature_not_found|❌ 未找到功能：%s"), feature), nil
	}

	var err error
	if defaultEnabled {
		err = db.SetGroupFeatureOverride(p.db, groupID, feature, false)
	} else {
		err = db.DeleteGroupFeatureOverride(p.db, groupID, feature)
	}

	if err != nil {
		log.Printf("设置功能关闭失败: %v", err)
		return fmt.Sprintf(common.T("", "admin_disable_feature_failed|❌ 关闭功能 %s 失败"), feature), err
	}

	return fmt.Sprintf(common.T("", "admin_feature_disabled|✅ 功能 %s 已关闭"), feature), nil
}

// doSetQAMode 执行设置问答模式逻辑
func (p *AdminPlugin) doSetQAMode(groupID string, mode string) string {
	if p.db == nil {
		return common.T("", "admin_no_db_feature|数据库未连接，无法设置功能。")
	}

	if err := db.SetGroupQAMode(p.db, groupID, mode); err != nil {
		log.Printf("设置模式 %s 失败: %v", mode, err)
		return "设置失败"
	}

	switch mode {
	case "chatty":
		return common.T("", "admin_chatty_mode_enabled|✅ 已开启话唠模式")
	case "ultimate":
		return common.T("", "admin_ultimate_mode_enabled|✅ 已开启终极模式")
	case "agent":
		return common.T("", "admin_agent_mode_enabled|✅ 已开启代理模式")
	case "silent":
		return common.T("", "admin_silent_mode_enabled|✅ 已开启静默模式")
	case "group":
		return common.T("", "admin_group_mode_enabled|✅ 已开启本群问答模式")
	case "official":
		return common.T("", "admin_official_mode_enabled|✅ 已开启官方问答模式")
	}

	return "设置成功"
}

// doShowAdminMenu 获取后台菜单
func (p *AdminPlugin) doShowAdminMenu() string {
	adminMenu := "🔧 " + common.T("", "admin_menu_title|管理后台") + "\n"
	adminMenu += "====================\n"
	adminMenu += common.T("", "admin_menu_enable|1. /开启 <功能>") + "\n"
	adminMenu += common.T("", "admin_menu_disable|2. /关闭 <功能>") + "\n"
	adminMenu += common.T("", "admin_menu_set|3. /设置 <参数> <值>") + "\n"
	adminMenu += common.T("", "admin_menu_teach|4. /帮助 - 查看功能说明") + "\n"
	adminMenu += common.T("", "admin_menu_group_info|5. /本群 - 查看群信息") + "\n"
	adminMenu += common.T("", "admin_menu_chatty|6. /话唠模式 - 开启话唠模式") + "\n"
	adminMenu += common.T("", "admin_menu_ultimate|7. /终极模式 - 开启终极模式") + "\n"
	adminMenu += common.T("", "admin_menu_agent|8. /代理模式 - 开启代理模式") + "\n"
	return adminMenu
}

// doShowTeaching 获取教学内容
func (p *AdminPlugin) doShowTeaching() string {
	teaching := "📚 " + common.T("", "admin_tutorial_title|功能教学") + "\n"
	teaching += "====================\n"
	teaching += "/菜单 - " + common.T("", "admin_help_menu|查看所有功能菜单") + "\n"
	teaching += "/help - " + common.T("", "admin_help_help|查看帮助说明") + "\n"
	teaching += "/签到 - " + common.T("", "admin_help_signin|每日签到领取积分") + "\n"
	teaching += "/积分 - " + common.T("", "admin_help_points|查询自己的积分") + "\n"
	teaching += "/天气 <城市> - " + common.T("", "admin_help_weather|查询城市天气") + "\n"
	teaching += "/翻译 <文本> - " + common.T("", "admin_help_translate|中英文互译") + "\n"
	teaching += "/点歌 <歌曲> - " + common.T("", "admin_help_music|在线点歌") + "\n"
	teaching += "/猜拳 <选择> - " + common.T("", "admin_help_rps|和机器人猜拳") + "\n"
	teaching += "/猜大小 <选择> - " + common.T("", "admin_help_guess|猜大小游戏") + "\n"
	teaching += "/抽奖 - " + common.T("", "admin_help_lottery|积分抽奖") + "\n"
	teaching += "/早安 - " + common.T("", "admin_help_morning|早安打卡") + "\n"
	teaching += "/晚安 - " + common.T("", "admin_help_night|晚安打卡") + "\n"
	teaching += "/报时 - " + common.T("", "admin_help_time|当前时间报时") + "\n"
	teaching += "/计算 <表达式> - " + common.T("", "admin_help_calc|数学表达式计算") + "\n"
	teaching += "/笑话 - " + common.T("", "admin_help_joke|讲个笑话") + "\n"
	teaching += "/鬼故事 - " + common.T("", "admin_help_ghost|讲个鬼故事") + "\n"
	teaching += "/成语接龙 <成语> - " + common.T("", "admin_help_idiom|成语接龙游戏") + "\n"
	return teaching
}

// doShowGroupInfo 获取本群信息
func (p *AdminPlugin) doShowGroupInfo() string {
	groupInfo := "🏠 " + common.T("", "admin_group_info_title|本群信息") + "\n"
	groupInfo += "====================\n"
	groupInfo += common.T("", "admin_group_name|群名称") + "：" + common.T("", "admin_unknown|未知") + "\n"
	groupInfo += common.T("", "admin_group_member_count|成员数量") + "：" + common.T("", "admin_unknown|未知") + "\n"
	groupInfo += common.T("", "admin_group_create_time|创建时间") + "：" + common.T("", "admin_unknown|未知") + "\n"
	groupInfo += common.T("", "admin_group_notice|群公告") + "：" + common.T("", "admin_none|无") + "\n"
	return groupInfo
}

// sendMessage 发送消息
func (p *AdminPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

func (p *AdminPlugin) handleSaveGroupVoice(robot plugin.Robot, event *onebot.Event, groupID, voiceID, voiceName, suffix string) {
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "admin_no_db_voice|数据库未连接，无法设置语音。"))
		return
	}

	if err := db.SetGroupVoiceID(p.db, groupID, voiceID); err != nil {
		log.Printf("设置群语音失败: %v", err)
		p.sendMessage(robot, event, "❌ "+common.T("", "admin_set_voice_failed|设置语音失败，请稍后再试。"))
		return
	}

	categories := GetVoiceCategoriesForID(voiceID)
	categoryName := strings.Join(categories, "、")
	url := GetVoicePreviewURL(voiceID)

	msg := "✅ " + common.T("", "admin_set_voice_success|设置成功！当前语音包：") + voiceName
	if categoryName != "" {
		msg += "（" + categoryName + "）"
	}
	if suffix != "" {
		msg += suffix
	}
	if url != "" {
		msg += "\n" + common.T("", "admin_preview|预览") + "：" + url
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
