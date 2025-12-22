package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type AdminPlugin struct {
	admins []string
	db     *sql.DB
	cmdParser *CommandParser
}

func (p *AdminPlugin) Name() string {
	return "admin"
}

func (p *AdminPlugin) Description() string {
	return "admin plugin，支持后台设置、功能开关、教学等功能"
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
	log.Println("加载admin plugin")

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
		adminMenu := "🔧 后台管理菜单\n"
		adminMenu += "====================\n"
		adminMenu += "/开启 <功能> - 开启指定功能\n"
		adminMenu += "/关闭 <功能> - 关闭指定功能\n"
		adminMenu += "/设置 <参数> <值> - 设置参数\n"
		adminMenu += "/教学 - 查看使用教程\n"
		adminMenu += "/本群 - 查看本群信息\n"
		adminMenu += "/话唠 - 开启话唠模式\n"
		adminMenu += "/终极 - 开启终极模式\n"
		adminMenu += "/智能体 - 开启智能体模式\n"
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
					p.sendMessage(robot, event, "权限不足，只有群主或机器人主人可以操作该功能")
					return nil
				}
			} else if requireAdmin {
				if !isGroupAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, "权限不足，只有管理员可以操作该功能")
					return nil
				}
			}
		}
		defaultEnabled, ok := FeatureDefaults[feature]
		if !ok {
			p.sendMessage(robot, event, fmt.Sprintf("功能%s不存在", feature))
			return nil
		}

		if event.MessageType != "group" {
			p.sendMessage(robot, event, "仅支持在群聊中设置功能开关")
			return nil
		}

		if p.db == nil {
			p.sendMessage(robot, event, "数据库未配置，无法保存功能开关")
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
					p.sendMessage(robot, event, "权限不足，只有群主或机器人主人可以操作该功能")
					return nil
				}
			} else if requireAdmin {
				if !isGroupAdmin(p.db, event.GroupID, event.UserID) {
					p.sendMessage(robot, event, "权限不足，只有管理员可以操作该功能")
					return nil
				}
			}
		}
		defaultEnabled, ok := FeatureDefaults[feature]
		if !ok {
			p.sendMessage(robot, event, fmt.Sprintf("功能%s不存在", feature))
			return nil
		}

		if event.MessageType != "group" {
			p.sendMessage(robot, event, "仅支持在群聊中设置功能开关")
			return nil
		}

		if p.db == nil {
			p.sendMessage(robot, event, "数据库未配置，无法保存功能开关")
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
			p.sendMessage(robot, event, fmt.Sprintf("关闭功能%s失败", feature))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("功能%s已关闭", feature))

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
		p.sendMessage(robot, event, fmt.Sprintf("参数%s已设置为%s", param, value))

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
		teaching := "📚 使用教程\n"
		teaching += "====================\n"
		teaching += "/菜单 - 查看所有命令\n"
		teaching += "/help - 查看帮助信息\n"
		teaching += "/签到 - 每日签到\n"
		teaching += "/积分 - 查询积分\n"
		teaching += "/天气 <城市> - 查询天气\n"
		teaching += "/翻译 <文本> - 翻译文本\n"
		teaching += "/点歌 <歌曲> - 点歌\n"
		teaching += "/猜拳 <选择> - 猜拳\n"
		teaching += "/猜大小 <选择> - 猜大小\n"
		teaching += "/抽奖 - 抽奖\n"
		teaching += "/早安 - 早安问候\n"
		teaching += "/晚安 - 晚安问候\n"
		teaching += "/报时 - 查看当前时间\n"
		teaching += "/计算 <表达式> - 计算\n"
		teaching += "/笑话 - 讲笑话\n"
		teaching += "/鬼故事 - 讲鬼故事\n"
		teaching += "/成语接龙 <成语> - 成语接龙\n"
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
		groupInfo := "🏠 本群信息\n"
		groupInfo += "====================\n"
		groupInfo += "群名称：未知\n"
		groupInfo += "群人数：未知\n"
		groupInfo += "群创建时间：未知\n"
		groupInfo += "群公告：无\n"
		p.sendMessage(robot, event, groupInfo)

		return nil
	})

	// 处理话唠命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为话唠命令
		if match, _ := p.cmdParser.MatchCommand("话唠|chatty", event.RawMessage); !match {
			return nil
		}

		// 开启话唠模式
		p.sendMessage(robot, event, "话唠模式已开启！我会更积极地回复消息哦！")

		return nil
	})

	// 处理终极命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为终极命令
		if match, _ := p.cmdParser.MatchCommand("终极|ultimate", event.RawMessage); !match {
			return nil
		}

		// 开启终极模式
		p.sendMessage(robot, event, "终极模式已开启！我会释放全部能力！")

		return nil
	})

	// 处理智能体命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为智能体命令
		if match, _ := p.cmdParser.MatchCommand("智能体|agent", event.RawMessage); !match {
			return nil
		}

		// 开启智能体模式
		p.sendMessage(robot, event, "智能体模式已开启！我会更智能地回复消息！")

		return nil
	})
}

// sendMessage 发送消息
func (p *AdminPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
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
