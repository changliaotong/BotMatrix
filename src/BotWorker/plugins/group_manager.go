package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"BotMatrix/common"
)

type GroupManagerPlugin struct {
	// 数据库连接
	db *sql.DB
	// Redis客户端
	redisClient *redis.Client
	// 命令解析器
	cmdParser *CommandParser
}

func NewGroupManagerPlugin(database *sql.DB, redisClient *redis.Client) *GroupManagerPlugin {
	return &GroupManagerPlugin{
		db:          database,
		redisClient: redisClient,
		cmdParser:   NewCommandParser(),
	}
}

func (p *GroupManagerPlugin) Name() string {
	return "group_manager"
}

func (p *GroupManagerPlugin) Description() string {
	return common.T("", "group_manager_plugin_desc|群管理插件，提供踢人、禁言、设置群规等功能")
}

func (p *GroupManagerPlugin) Version() string {
	return "1.0.0"
}

func (p *GroupManagerPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "group_manager_plugin_loaded|群管理插件已加载"))

	// 报备技能
	robot.HandleSkill("love_owner", func(params map[string]string) (string, error) {
		err := p.handleLoveOwnerLogic(robot, nil)
		return "", err
	})
	robot.HandleSkill("fan_rank", func(params map[string]string) (string, error) {
		err := p.handleFanRankLogic(robot, nil)
		return "", err
	})
	robot.HandleSkill("kick", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		refuse := params["refuse_rejoin"]
		err := p.handleKickLogic(robot, nil, []string{userID, refuse})
		return "", err
	})
	robot.HandleSkill("ban", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		duration := params["duration_minutes"]
		err := p.handleBanLogic(robot, nil, []string{userID, duration})
		return "", err
	})
	robot.HandleSkill("unban", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		err := p.handleUnbanLogic(robot, nil, []string{userID})
		return "", err
	})
	robot.HandleSkill("add_admin", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		err := p.handleAddAdminLogic(robot, nil, []string{userID})
		return "", err
	})
	robot.HandleSkill("del_admin", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		err := p.handleDelAdminLogic(robot, nil, []string{userID})
		return "", err
	})
	robot.HandleSkill("set_rules", func(params map[string]string) (string, error) {
		rules := params["rules"]
		err := p.handleSetRulesLogic(robot, nil, []string{rules})
		return "", err
	})
	robot.HandleSkill("add_word", func(params map[string]string) (string, error) {
		word := params["word"]
		err := p.handleAddWordLogic(robot, nil, []string{word})
		return "", err
	})
	robot.HandleSkill("del_word", func(params map[string]string) (string, error) {
		word := params["word"]
		err := p.handleDelWordLogic(robot, nil, []string{word})
		return "", err
	})
	robot.HandleSkill("get_members", func(params map[string]string) (string, error) {
		err := p.handleGetMembersLogic(robot, nil, nil)
		return "", err
	})
	robot.HandleSkill("member_info", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		err := p.handleGetMemberInfoLogic(robot, nil, []string{userID})
		return "", err
	})
	robot.HandleSkill("set_title", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		title := params["title"]
		err := p.handleSetTitleLogic(robot, nil, []string{userID, title})
		return "", err
	})
	robot.HandleSkill("invitation_stats", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		var err error
		if userID != "" {
			err = p.handleInvitationStatsLogic(robot, nil, []string{userID})
		} else {
			err = p.handleInvitationStatsLogic(robot, nil, nil)
		}
		return "", err
	})
	robot.HandleSkill("invite_rank", func(params map[string]string) (string, error) {
		err := p.handleInviteRankLogic(robot, nil, nil)
		return "", err
	})

	// 处理爱群主命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 检查是否为爱群主命令
		if match, _ := p.cmdParser.MatchCommand("爱群主|loveowner|loveadmin", event.RawMessage); match {
			return p.handleLoveOwnerLogic(robot, event)
		}

		return nil
	})

	// 处理粉丝团排行榜命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 检查是否为粉丝团排行榜命令
		if match, _ := p.cmdParser.MatchCommand("粉丝团排行|fanrank|intimacyrank", event.RawMessage); match {
			return p.handleFanRankLogic(robot, event)
		}

		return nil
	})

	// 如果数据库连接可用，添加默认敏感词
	if p.db != nil {
		defaultSensitiveWords := []string{"敏感词1", "敏感词2", "敏感词3"}
		for _, word := range defaultSensitiveWords {
			if err := db.AddSensitiveWord(p.db, word, 3); err != nil {
				log.Printf(common.T("", "group_manager_add_default_sensitive_failed|添加默认敏感词失败"), err)
			}
		}

		// 设置默认群规（如果不存在）
		defaultRules := common.T("", "group_manager_default_rules|1. 禁止发布广告\n2. 禁止人身攻击\n3. 请遵守群规")
		if err := db.SetGroupRules(p.db, "0", defaultRules); err != nil {
			log.Printf(common.T("", "group_manager_set_default_rules_failed|设置默认群规失败"), err)
		}
	}

	// 处理群消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(p.db, groupIDStr, "moderation") {
			HandleFeatureDisabled(robot, event, "moderation")
			return nil
		}

		// 检查是否为管理员命令
		if p.isAdminCommand(event) {
			return p.handleAdminCommand(robot, event)
		}

		// 关键词过滤
		if p.containsSensitiveWords(event.RawMessage) {
			// 警告用户
			warningMsg := fmt.Sprintf("@%d 请注意你的发言，包含敏感词汇！", event.UserID)
			p.sendMessage(robot, event, warningMsg)

			// 记录日志
			log.Printf("用户 %d 在群 %d 发送了敏感消息: %s", event.UserID, event.GroupID, event.RawMessage)
		}

		// 检查是否是命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "group_manager_cmd_rules|群规|rules"), event.RawMessage); match {
			p.sendGroupRules(robot, event)
		} else if match, _ := p.cmdParser.MatchCommand("help", event.RawMessage); match {
			p.sendHelp(robot, event)
		}

		return nil
	})

	// 处理群成员增加事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_member_increase" {
			// 发送欢迎消息和群规
			p.sendWelcomeAndRules(robot, event)
		}
		return nil
	})

	// 定期检查禁言时间
	go p.checkBanExpiration(robot)
}

// GetSkills 实现 SkillCapable 接口
func (p *GroupManagerPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "love_owner",
			Description: "爱群主，增加亲密度和积分",
			Usage:       "love_owner",
		},
		{
			Name:        "fan_rank",
			Description: "查看粉丝团排行榜",
			Usage:       "fan_rank",
		},
		{
			Name:        "kick",
			Description: "移除群成员",
			Usage:       "kick <user_id> [refuse_rejoin]",
			Params: map[string]string{
				"user_id":       "要移除的用户ID",
				"refuse_rejoin": "是否拒绝再次加入（true/false）",
			},
		},
		{
			Name:        "ban",
			Description: "禁言群成员",
			Usage:       "ban <user_id> [duration_minutes]",
			Params: map[string]string{
				"user_id":          "要禁言的用户ID",
				"duration_minutes": "禁言时长（分钟）",
			},
		},
		{
			Name:        "unban",
			Description: "解除禁言",
			Usage:       "unban <user_id>",
			Params: map[string]string{
				"user_id": "要解除禁言的用户ID",
			},
		},
		{
			Name:        "add_admin",
			Description: "添加群管理员",
			Usage:       "add_admin <user_id>",
			Params: map[string]string{
				"user_id": "要添加的用户ID",
			},
		},
		{
			Name:        "del_admin",
			Description: "删除群管理员",
			Usage:       "del_admin <user_id>",
			Params: map[string]string{
				"user_id": "要删除的用户ID",
			},
		},
		{
			Name:        "set_rules",
			Description: "设置群规",
			Usage:       "set_rules <rules_content>",
			Params: map[string]string{
				"rules": "群规内容",
			},
		},
		{
			Name:        "add_word",
			Description: "添加敏感词",
			Usage:       "add_word <word>",
			Params: map[string]string{
				"word": "敏感词内容",
			},
		},
		{
			Name:        "del_word",
			Description: "删除敏感词",
			Usage:       "del_word <word>",
			Params: map[string]string{
				"word": "敏感词内容",
			},
		},
		{
			Name:        "get_members",
			Description: "获取群成员列表",
			Usage:       "get_members",
		},
		{
			Name:        "member_info",
			Description: "获取群成员信息",
			Usage:       "member_info <user_id>",
			Params: map[string]string{
				"user_id": "用户ID",
			},
		},
		{
			Name:        "set_title",
			Description: "设置群头衔",
			Usage:       "set_title <user_id> <title>",
			Params: map[string]string{
				"user_id": "用户ID",
				"title":   "头衔内容",
			},
		},
		{
			Name:        "invitation_stats",
			Description: "查看邀请统计",
			Usage:       "invitation_stats [user_id]",
			Params: map[string]string{
				"user_id": "可选，用户ID",
			},
		},
		{
			Name:        "invite_rank",
			Description: "查看邀请排行榜",
			Usage:       "invite_rank",
		},
	}
}

// sendMessage 发送消息
func (p *GroupManagerPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID:    event.GroupID,
		UserID:     event.UserID,
		Message:    message,
		AutoEscape: false,
	})
}

// 检查是否是管理员命令
func (p *GroupManagerPlugin) isAdminCommand(event *onebot.Event) bool {
	if event.MessageType != "group" {
		return false
	}

	// 使用CommandParser检查是否是命令，支持可选的/前缀
	return p.cmdParser.IsCommand("\\w+", event.RawMessage)
}

// 处理管理员命令
func (p *GroupManagerPlugin) handleAdminCommand(robot plugin.Robot, event *onebot.Event) error {
	// 检查是否为管理员
	if !p.isAdmin(event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_admin|权限不足，只有管理员可以使用此命令"))
		return nil
	}

	// 提取命令和参数 - 使用CommandParser的通用模式匹配
	pattern := `(\w+)`     // 匹配命令名
	paramPattern := `(.*)` // 匹配所有参数
	match, command, paramMatches := p.cmdParser.MatchCommandWithParams(pattern, paramPattern, event.RawMessage)
	if !match || len(command) == 0 {
		return nil
	}

	command = strings.ToLower(command)
	args := strings.Fields(paramMatches[0])

	// 处理不同的命令
	switch command {
	case "kick":
		return p.handleKickLogic(robot, event, args)
	case "ban":
		return p.handleBanLogic(robot, event, args)
	case "unban":
		return p.handleUnbanLogic(robot, event, args)
	case "addadmin":
		return p.handleAddAdminLogic(robot, event, args)
	case "deladmin":
		return p.handleDelAdminLogic(robot, event, args)
	case "setrules":
		return p.handleSetRulesLogic(robot, event, args)
	case "addword":
		return p.handleAddWordLogic(robot, event, args)
	case "delword":
		return p.handleDelWordLogic(robot, event, args)
	case "members":
		return p.handleGetMembersLogic(robot, event, args)
	case "memberinfo":
		return p.handleGetMemberInfoLogic(robot, event, args)
	case "settitle":
		return p.handleSetTitleLogic(robot, event, args)
	case "invitationstats":
		return p.handleInvitationStatsLogic(robot, event, args)
	case "inviterank":
		return p.handleInviteRankLogic(robot, event, args)
	}

	return nil
}

// 处理邀请统计命令
func (p *GroupManagerPlugin) handleInvitationStatsLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if p.db == nil {
		p.sendMessage(robot, event, "数据库未配置，无法查看邀请统计！")
		return nil
	}

	var targetUserID string
	if len(args) > 0 {
		targetUserID = args[0]
	} else {
		targetUserID = fmt.Sprintf("%d", event.UserID)
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 查询邀请次数
	var count int
	query := "SELECT COALESCE(invitation_count, 0) FROM group_invitation_stats WHERE group_id = ? AND user_id = ?"
	err := p.db.QueryRow(query, groupIDStr, targetUserID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有邀请记录
			p.sendMessage(robot, event, fmt.Sprintf("用户 %s 暂无邀请记录！", targetUserID))
		} else {
			log.Printf("[GroupManager] 查询邀请统计失败: %v", err)
			p.sendMessage(robot, event, "查询邀请统计失败，请稍后重试！")
		}
		return nil
	}

	// 查询邀请的具体用户
	inviteesQuery := "SELECT invitee_id FROM group_invitations WHERE group_id = ? AND inviter_id = ? ORDER BY invite_time DESC"
	rows, err := p.db.Query(inviteesQuery, groupIDStr, targetUserID)
	if err != nil {
		log.Printf("[GroupManager] 查询邀请用户列表失败: %v", err)
		return nil
	}
	defer rows.Close()

	var invitees []string
	for rows.Next() {
		var inviteeID string
		if err := rows.Scan(&inviteeID); err != nil {
			log.Printf("[GroupManager] 扫描邀请用户失败: %v", err)
			continue
		}
		invitees = append(invitees, inviteeID)
	}

	// 发送统计信息
	message := fmt.Sprintf("用户 %s 的邀请统计：\n", targetUserID)
	message += fmt.Sprintf("邀请人数：%d\n", count)
	if len(invitees) > 0 {
		message += fmt.Sprintf("邀请的用户：%s\n", strings.Join(invitees, ", "))
	}

	p.sendMessage(robot, event, message)
	return nil
}

// 处理邀请排行榜命令
func (p *GroupManagerPlugin) handleInviteRankLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if p.db == nil {
		p.sendMessage(robot, event, "数据库未配置，无法查看邀请排行榜！")
		return nil
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 查询邀请排行榜
	query := "SELECT user_id, invitation_count FROM group_invitation_stats WHERE group_id = ? ORDER BY invitation_count DESC LIMIT 10"
	rows, err := p.db.Query(query, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] 查询邀请排行榜失败: %v", err)
		p.sendMessage(robot, event, "查询邀请排行榜失败，请稍后重试！")
		return nil
	}
	defer rows.Close()

	// 构建排行榜信息
	var rankMsg strings.Builder
	rankMsg.WriteString("邀请排行榜（前10名）：\n\n")

	rank := 1
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			log.Printf("[GroupManager] 扫描排行榜数据失败: %v", err)
			continue
		}
		rankMsg.WriteString(fmt.Sprintf("%d. 用户 %s：%d 人\n", rank, userID, count))
		rank++
	}

	if rank == 1 {
		rankMsg.WriteString("暂无邀请记录！")
	}

	// 发送排行榜信息
	p.sendMessage(robot, event, rankMsg.String())
	return nil
}

// 处理爱群主操作
func (p *GroupManagerPlugin) handleLoveOwnerLogic(robot plugin.Robot, event *onebot.Event) error {
	if p.db == nil {
		p.sendMessage(robot, event, "数据库未配置，无法使用爱群主功能！")
		return fmt.Errorf("数据库未配置")
	}

	userIDStr := fmt.Sprintf("%d", event.UserID)
	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 检查冷却时间
	coolKey := fmt.Sprintf("love_owner_cool:%s:%s", groupIDStr, userIDStr)
	coolExpire, err := p.redisClient.TTL(context.Background(), coolKey).Result()
	if err != nil && err != redis.Nil {
		log.Printf("[GroupManager] 检查冷却时间失败: %v", err)
	} else if coolExpire > 0 {
		remaining := time.Duration(coolExpire) * time.Second
		message := fmt.Sprintf("💖 爱群主功能冷却中，剩余时间：%.0f分钟", remaining.Minutes())
		p.sendMessage(robot, event, message)
		return nil
	}

	// 检查是否已经加入粉丝团
	var isMember bool
	query := "SELECT EXISTS(SELECT 1 FROM fan_group_members WHERE group_id = ? AND user_id = ?)"
	err = p.db.QueryRow(query, groupIDStr, userIDStr).Scan(&isMember)
	if err != nil {
		log.Printf("[GroupManager] 检查粉丝团成员失败: %v", err)
		return err
	}

	if !isMember {
		// 自动加入粉丝团
		insertQuery := "INSERT INTO fan_group_members (group_id, user_id, join_time) VALUES (?, ?, ?)"
		_, err = p.db.Exec(insertQuery, groupIDStr, userIDStr, time.Now())
		if err != nil {
			log.Printf("[GroupManager] 加入粉丝团失败: %v", err)
			return err
		}
	}

	// 增加亲密度和积分
	intimacyPoints := 10
	pointReward := 50

	// 更新亲密度
	updateIntimacyQuery := "INSERT INTO fan_group_intimacy (group_id, user_id, intimacy_points, last_love_time) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE intimacy_points = intimacy_points + ?, last_love_time = ?"
	_, err = p.db.Exec(updateIntimacyQuery, groupIDStr, userIDStr, intimacyPoints, time.Now(), intimacyPoints, time.Now())
	if err != nil {
		log.Printf("[GroupManager] 更新亲密度失败: %v", err)
		return err
	}

	// 发放积分奖励
	// 这里假设存在points表，需要根据实际情况调整
	updatePointsQuery := "INSERT INTO user_points (user_id, points) VALUES (?, ?) ON DUPLICATE KEY UPDATE points = points + ?"
	_, err = p.db.Exec(updatePointsQuery, userIDStr, pointReward, pointReward)
	if err != nil {
		log.Printf("[GroupManager] 发放积分奖励失败: %v", err)
		return err
	}

	// 设置冷却时间（10分钟）
	coolKey = fmt.Sprintf("love_owner_cool:%s:%s", groupIDStr, userIDStr)
	_, err = p.redisClient.SetEX(context.Background(), coolKey, "1", 10*time.Minute).Result()
	if err != nil {
		log.Printf("[GroupManager] 设置冷却时间失败: %v", err)
		return err
	}

	// 发送成功消息
	message := fmt.Sprintf("💖 爱群主成功！\n")
	message += fmt.Sprintf("获得亲密度：+%d\n", intimacyPoints)
	message += fmt.Sprintf("获得积分奖励：+%d\n", pointReward)
	message += "每10分钟可以爱一次群主哦～"

	p.sendMessage(robot, event, message)

	return nil
}

// 处理粉丝团排行榜
func (p *GroupManagerPlugin) handleFanRankLogic(robot plugin.Robot, event *onebot.Event) error {
	if p.db == nil {
		p.sendMessage(robot, event, "数据库未配置，无法查看粉丝团排行！")
		return fmt.Errorf("数据库未配置")
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 查询粉丝团排行榜
	query := "SELECT user_id, intimacy_points FROM fan_group_intimacy WHERE group_id = ? ORDER BY intimacy_points DESC LIMIT 10"
	rows, err := p.db.Query(query, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] 查询粉丝团排行失败: %v", err)
		p.sendMessage(robot, event, "查询粉丝团排行失败，请稍后重试！")
		return err
	}
	defer rows.Close()

	// 构建排行榜信息
	var rankMsg strings.Builder
	rankMsg.WriteString("粉丝团亲密度排行榜（前10名）：\n\n")

	rank := 1
	for rows.Next() {
		var userID string
		var intimacyPoints int
		if err := rows.Scan(&userID, &intimacyPoints); err != nil {
			log.Printf("[GroupManager] 扫描粉丝团排行数据失败: %v", err)
			continue
		}
		rankMsg.WriteString(fmt.Sprintf("%d. 用户 %s：%d 亲密度\n", rank, userID, intimacyPoints))
		rank++
	}

	if rank == 1 {
		rankMsg.WriteString("暂无粉丝团成员！")
	}

	// 发送排行榜信息
	p.sendMessage(robot, event, rankMsg.String())

	return nil
}

// 处理踢人操作
func (p *GroupManagerPlugin) handleKickLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_kick_usage|用法：kick <user_id> [refuse_rejoin]"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 检查是否是管理员
	if p.isAdmin(event.GroupID, userID) {
		p.sendMessage(robot, event, common.T("", "group_manager_kick_admin_denied|不能移除管理员"))
		log.Printf("[GroupManager] %s %d %s %d, %s", common.T("", "group_manager_try_kick_admin|尝试移除管理员"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "group_manager_op_denied|操作被拒绝"))
		return nil
	}

	// 执行踢人操作
	refuse := false
	if len(args) > 1 && (args[1] == "true" || args[1] == "1") {
		refuse = true
	}

	// 记录踢人操作
	log.Printf("[GroupManager] %s %d %s %d, %s: %v", common.T("", "group_manager_try_kick_user|尝试移除用户"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "group_manager_refuse_rejoin|是否拒绝再次加入"), refuse)

	if robot != nil {
		_, err = robot.SetGroupKick(&onebot.SetGroupKickParams{
			GroupID:   event.GroupID,
			UserID:    userID,
			RejectAdd: refuse,
		})

		if err != nil {
			log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_kick_user|移除用户"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "failed|失败"), err)
			p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_kick_failed|移除失败"), err))
			return nil
		}
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %s %d", common.T("", "group_manager_kick_success|成功移除用户"), userID, common.T("", "from_group|从群"), event.GroupID)
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_kick_success_msg|已成功将用户 %d 移出本群"), userID))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "kick",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_kick|移除用户 %d，拒绝再次加入：%v"), userID, refuse),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理禁言命令
func (p *GroupManagerPlugin) handleBanLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_ban_usage|用法：ban <user_id> [duration_minutes]"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 检查是否是管理员
	if p.isAdmin(event.GroupID, userID) {
		p.sendMessage(robot, event, common.T("", "group_manager_ban_admin_denied|不能禁言管理员"))
		log.Printf("[GroupManager] %s %d %s %d, %s", common.T("", "group_manager_try_ban_admin|尝试禁言管理员"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "group_manager_op_denied|操作被拒绝"))
		return nil
	}

	// 解析禁言时长
	duration := 30 * time.Minute // 默认30分钟
	if len(args) > 1 {
		minutes, err := parseDuration(args[1])
		if err == nil && minutes > 0 {
			duration = time.Duration(minutes) * time.Minute
		} else {
			log.Printf("[GroupManager] %s '%s' %s, %s", common.T("", "group_manager_parse_duration_failed|解析时长失败"), args[1], common.T("", "failed|失败"), common.T("", "group_manager_use_default_duration|使用默认时长"))
		}
	}

	// 执行禁言操作
	log.Printf("[GroupManager] %s %d %s %d, %s %d %s", common.T("", "group_manager_try_ban_user|尝试禁言用户"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "duration|时长"), int(duration.Minutes()), common.T("", "minutes|分钟"))

	if robot != nil {
		_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
			GroupID:  event.GroupID,
			UserID:   userID,
			Duration: int(duration.Seconds()),
		})

		if err != nil {
			log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_ban_user|禁言用户"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "failed|失败"), err)
			p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_ban_failed|禁言失败"), err))
			return nil
		}
	}

	// 存储禁言信息到Redis
	if p.redisClient != nil {
		ctx := context.Background()
		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		userIDStr := fmt.Sprintf("%d", userID)
		banKey := fmt.Sprintf("group:%s:ban:%s", groupIDStr, userIDStr)

		// 设置禁言记录，带过期时间
		if err := p.redisClient.Set(ctx, banKey, time.Now().Add(duration).Unix(), duration).Err(); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_redis_save_ban_failed|保存禁言记录到Redis失败"), err)
			// 回退到数据库存储
			if p.db != nil {
				banEndTime := time.Now().Add(duration)
				if err := db.BanUser(p.db, groupIDStr, userIDStr, banEndTime); err != nil {
					log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_db_save_ban_failed|保存禁言记录到数据库失败"), err)
				} else {
					log.Printf("[GroupManager] %s", common.T("", "group_manager_fallback_db_save|已回退到数据库存储"))
				}
			}
		} else {
			log.Printf("[GroupManager] %s", common.T("", "group_manager_redis_save_ban_success|禁言记录已保存到Redis"))
		}
	} else if p.db != nil {
		// Redis不可用时，使用数据库存储
		banEndTime := time.Now().Add(duration)
		if err := db.BanUser(p.db, fmt.Sprintf("%d", event.GroupID), fmt.Sprintf("%d", userID), banEndTime); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_db_save_ban_failed|保存禁言记录到数据库失败"), err)
		} else {
			log.Printf("[GroupManager] %s", common.T("", "group_manager_db_save_ban_success|禁言记录已保存到数据库"))
		}
	} else {
		log.Printf("[GroupManager] %s", common.T("", "group_manager_persistence_unavailable|无可用存储引擎（Redis/DB）"))
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %d %s", common.T("", "group_manager_ban_success|成功禁言用户"), userID, int(duration.Minutes()), common.T("", "minutes|分钟"))
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_ban_success_msg|已成功禁言用户 %d，时长 %d 分钟"), userID, int(duration.Minutes())))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "ban",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_ban|禁言用户 %d，时长 %d 分钟"), userID, int(duration.Minutes())),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理解除禁言命令
func (p *GroupManagerPlugin) handleUnbanLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_unban_usage|用法：unban <user_id>"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 执行解除禁言操作
	log.Printf("[GroupManager] %s %d %s %d", common.T("", "group_manager_try_unban_user|尝试解除禁言用户"), event.GroupID, common.T("", "in_group|在群"), userID)

	if robot != nil {
		_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
			GroupID:  event.GroupID,
			UserID:   userID,
			Duration: 0,
		})

		if err != nil {
			log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_unban_user|解除禁言用户"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "failed|失败"), err)
			p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_unban_failed|解除禁言失败"), err))
			return nil
		}
	}

	// 从Redis移除禁言记录
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	if p.redisClient != nil {
		ctx := context.Background()
		banKey := fmt.Sprintf("group:%s:ban:%s", groupIDStr, userIDStr)

		if err := p.redisClient.Del(ctx, banKey).Err(); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_redis_del_ban_failed|从Redis删除禁言记录失败"), err)
		} else {
			log.Printf("[GroupManager] %s", common.T("", "group_manager_redis_del_ban_success|从Redis删除禁言记录成功"))
		}
	}

	// 同时从数据库移除禁言记录，确保数据一致性
	if p.db != nil {
		if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_db_del_ban_failed|从数据库删除禁言记录失败"), err)
		} else {
			log.Printf("[GroupManager] %s", common.T("", "group_manager_db_del_ban_success|从数据库删除禁言记录成功"))
		}
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d", common.T("", "group_manager_unban_success|成功解除禁言用户"), userID)
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_unban_success_msg|已成功解除用户 %d 的禁言"), userID))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "unban",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_unban|解除用户 %d 的禁言"), userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理添加管理员命令
func (p *GroupManagerPlugin) handleAddAdminLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	// 只有超级管理员可以添加管理员
	if !p.isSuperAdmin(event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_superadmin|权限不足，只有超级管理员可以使用此命令"))
		log.Printf("[GroupManager] %s %d %s %d %s, %s", common.T("", "user|用户"), event.UserID, common.T("", "try_add_admin_in_group|尝试在群添加管理员"), event.GroupID, common.T("", "but_not_superadmin|但不是超级管理员"), common.T("", "group_manager_op_denied|操作被拒绝"))
		return nil
	}

	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_addadmin_usage|用法：addadmin <user_id>"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "group_manager_db_unavailable|数据库服务不可用"))
		log.Printf("[GroupManager] %s", common.T("", "group_manager_db_unavailable_log|数据库未配置，无法添加管理员"))
		return nil
	}

	// 添加到管理员列表（数据库）
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	// 检查是否已经是管理员
	isAdmin, err := db.IsGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_check_admin_status|检查管理员状态"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, common.T("", "group_manager_op_failed_retry|操作失败，请重试"))
		return nil
	}

	if isAdmin {
		p.sendMessage(robot, event, common.T("", "group_manager_already_admin|该用户已经是管理员"))
		return nil
	}

	// 添加管理员，默认权限级别为1（普通管理员）
	if err := db.AddGroupAdmin(p.db, groupIDStr, userIDStr, 1); err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_add_admin|添加管理员"), event.GroupID, common.T("", "to_group|到群"), userID, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, common.T("", "group_manager_op_failed_retry|操作失败，请重试"))
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %s %d %s", common.T("", "group|群"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "group_manager_add_admin_success|成功添加管理员"))
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_add_admin_success_msg|已成功将用户 %d 设为本群管理员"), userID))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "add_admin",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_add_admin|添加用户 %d 为管理员"), userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理删除管理员命令
func (p *GroupManagerPlugin) handleDelAdminLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	// 只有超级管理员可以删除管理员
	if !p.isSuperAdmin(event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_superadmin|权限不足，只有超级管理员可以使用此命令"))
		log.Printf("[GroupManager] %s %d %s %d %s", common.T("", "user|用户"), event.UserID, common.T("", "group_manager_try_del_admin|尝试删除群管理员"), event.GroupID, common.T("", "group_manager_op_denied|操作被拒绝"))
		return nil
	}

	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_deladmin_usage|用法：deladmin <user_id>"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "group_manager_db_unavailable|数据库服务不可用"))
		log.Printf("[GroupManager] %s", common.T("", "group_manager_db_unavailable_log|数据库未配置，无法操作"))
		return nil
	}

	// 从管理员列表中删除（数据库）
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	// 移除管理员
	err = db.RemoveGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_del_admin|删除管理员"), event.GroupID, common.T("", "from_group|从群"), userID, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, common.T("", "group_manager_not_admin|该用户不是管理员"))
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %s %d %s", common.T("", "group|群"), event.GroupID, common.T("", "in_group|在群"), userID, common.T("", "group_manager_del_admin_success|成功删除管理员"))
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_del_admin_success_msg|已成功移除用户 %d 的管理员权限"), userID))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "del_admin",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_del_admin|删除用户 %d 的管理员权限"), userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理设置群规命令
func (p *GroupManagerPlugin) handleSetRulesLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_setrules_usage|用法：setrules <rules>"))
		return nil
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "group_manager_db_unavailable|数据库服务不可用"))
		log.Printf("[GroupManager] %s", common.T("", "group_manager_db_unavailable_log|数据库未配置，无法操作"))
		return nil
	}

	// 设置群规
	rules := strings.Join(args, " ")
	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	if err := db.SetGroupRules(p.db, groupIDStr, rules); err != nil {
		log.Printf("[GroupManager] %s %d %s %s: %v", common.T("", "group_manager_set_rules|设置群规"), event.GroupID, common.T("", "failed|失败"), common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, common.T("", "group_manager_set_rules_failed|设置群规失败"))
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %s", common.T("", "group|群"), event.GroupID, common.T("", "group_manager_rules_updated|群规已更新"))
	p.sendMessage(robot, event, common.T("", "group_manager_rules_updated_msg|群规已成功更新"))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:     fmt.Sprintf("%d", event.GroupID),
			AdminID:     fmt.Sprintf("%d", event.UserID),
			Action:      "set_rules",
			Description: fmt.Sprintf(common.T("", "group_manager_audit_set_rules|设置群规：%s"), rules),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理添加敏感词命令
func (p *GroupManagerPlugin) handleAddWordLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_addword_usage|用法：addword [level] <word>"))
		return nil
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "group_manager_db_unavailable|数据库服务不可用"))
		log.Printf("[GroupManager] %s", common.T("", "group_manager_db_unavailable_log|数据库未配置，无法操作"))
		return nil
	}

	level := 3
	startIndex := 0

	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[0]); err == nil && v >= 1 && v <= 6 {
			level = v
			startIndex = 1
		}
	}

	if startIndex >= len(args) {
		p.sendMessage(robot, event, common.T("", "group_manager_provide_sensitive_word|请提供要添加的敏感词"))
		return nil
	}

	word := strings.Join(args[startIndex:], " ")

	// 添加到数据库
	if err := db.AddSensitiveWord(p.db, word, level); err != nil {
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_add_sensitive|添加敏感词"), word, common.T("", "failed|失败"), err)
		// 检查是否为重复添加
		if strings.Contains(err.Error(), "duplicate key") {
			p.sendMessage(robot, event, common.T("", "group_manager_sensitive_exists|该敏感词已存在"))
		} else {
			p.sendMessage(robot, event, common.T("", "group_manager_add_sensitive_failed_msg|添加敏感词失败，请重试"))
		}
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s '%s' %s", common.T("", "group_manager_sensitive|敏感词"), word, common.T("", "group_manager_sensitive_added|已添加"))
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_sensitive_added_msg|已成功添加敏感词：%s"), word))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:     fmt.Sprintf("%d", event.GroupID),
			AdminID:     fmt.Sprintf("%d", event.UserID),
			Action:      "add_sensitive_word",
			Description: fmt.Sprintf(common.T("", "group_manager_audit_add_sensitive|添加敏感词：%s"), word),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 处理删除敏感词命令
func (p *GroupManagerPlugin) handleDelWordLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_delword_usage|用法：delword <word>"))
		return nil
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		p.sendMessage(robot, event, common.T("", "group_manager_db_unavailable|数据库服务不可用"))
		log.Printf("[GroupManager] %s", common.T("", "group_manager_db_unavailable_log|数据库未配置，无法操作"))
		return nil
	}

	// 删除敏感词
	word := strings.Join(args, " ")

	// 从数据库删除
	if err := db.RemoveSensitiveWord(p.db, word); err != nil {
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_del_sensitive|删除敏感词"), word, common.T("", "failed|失败"), err)
		// 检查是否为不存在的敏感词
		if strings.Contains(err.Error(), "no rows in result set") || strings.Contains(err.Error(), "not found") {
			p.sendMessage(robot, event, common.T("", "group_manager_sensitive_not_exists|该敏感词不存在"))
		} else {
			p.sendMessage(robot, event, common.T("", "group_manager_op_failed_retry|操作失败，请重试"))
		}
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s '%s' %s", common.T("", "group_manager_sensitive|敏感词"), word, common.T("", "group_manager_del_success|删除成功"))
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_del_sensitive_success_msg|已成功删除敏感词：%s"), word))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:     fmt.Sprintf("%d", event.GroupID),
			AdminID:     fmt.Sprintf("%d", event.UserID),
			Action:      "del_sensitive_word",
			Description: fmt.Sprintf(common.T("", "group_manager_audit_del_sensitive|删除敏感词：%s"), word),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 检查是否包含敏感词
func (p *GroupManagerPlugin) containsSensitiveWords(message string) bool {
	if p.db == nil {
		return false
	}

	words, err := db.GetAllSensitiveWords(p.db)
	if err != nil {
		log.Printf("[GroupManager] 获取敏感词失败: %v", err)
		return false
	}

	for _, w := range words {
		if strings.Contains(strings.ToLower(message), strings.ToLower(w.Word)) {
			return true
		}
	}

	return false
}

// 检查是否为管理员
func (p *GroupManagerPlugin) isAdmin(groupID, userID int64) bool {
	// 从数据库检查是否为群管理员
	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	isAdmin, err := db.IsGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_check_admin_status|检查管理员状态"), groupID, common.T("", "of_user|用户"), userID, common.T("", "failed|失败"), err)
		return false
	}

	return isAdmin
}

// 检查是否为超级管理员
func (p *GroupManagerPlugin) isSuperAdmin(groupID, userID int64) bool {
	// 从数据库检查是否为超级管理员
	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	isSuperAdmin, err := db.IsSuperAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_check_superadmin_status|检查超级管理员状态"), groupID, common.T("", "of_user|用户"), userID, common.T("", "failed|失败"), err)
		return false
	}

	return isSuperAdmin
}

// 处理设置头衔命令
func (p *GroupManagerPlugin) handleSetTitleLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	// 只有群主可以设置头衔
	if !p.isOwner(robot, event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_owner|权限不足，只有群主可以使用此命令"))
		log.Printf("[GroupManager] %s %d %s %d %s", common.T("", "user|用户"), event.UserID, common.T("", "group_manager_try_set_title_not_owner|尝试在非群主身份下设置头衔"), event.GroupID, common.T("", "group_manager_op_denied|操作被拒绝"))
		return nil
	}

	if len(args) < 2 {
		p.sendMessage(robot, event, common.T("", "group_manager_settitle_usage|用法：settitle <user_id> <title>"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	// 检查目标用户是否存在
	_, err = robot.GetGroupMemberInfo(&onebot.GetGroupMemberInfoParams{
		GroupID: event.GroupID,
		UserID:  userID,
	})
	if err != nil {
		p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_get_member_info_failed|获取群成员信息失败"), err))
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_get_member_info_failed_log|获取群成员信息失败"), event.GroupID, userID, common.T("", "failed|失败"), err)
		return nil
	}

	// 解析头衔
	title := strings.Join(args[1:], " ")
	if len(title) > 12 {
		p.sendMessage(robot, event, common.T("", "group_manager_title_too_long|头衔长度不能超过12个字符"))
		return nil
	}

	// 执行设置头衔操作
	_, err = robot.SetGroupSpecialTitle(&onebot.SetGroupSpecialTitleParams{
		GroupID:      event.GroupID,
		UserID:       userID,
		SpecialTitle: title,
	})

	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s '%s' %s: %v", common.T("", "group_manager_set_title|设置群头衔"), event.GroupID, common.T("", "of_user|用户"), userID, common.T("", "to|为"), title, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_set_title_failed|设置群头衔失败"), err))
		return nil
	}

	// 记录成功操作
	log.Printf("[GroupManager] %s %d %s '%s'", common.T("", "group_manager_set_title_success_log|成功设置用户头衔"), userID, common.T("", "to|为"), title)
	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "group_manager_set_title_success_msg|已成功将用户 %d 的头衔设置为：%s"), userID, title))

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "set_title",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf(common.T("", "group_manager_audit_set_title|设置用户 %d 的头衔为：%s"), userID, title),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_add_audit_failed|添加审核日志失败"), err)
		}
	}
	return nil
}

// 检查是否为群主
func (p *GroupManagerPlugin) isOwner(robot plugin.Robot, groupID, userID int64) bool {
	if robot == nil {
		return false
	}
	// 获取用户的群成员信息
	memberInfo, err := robot.GetGroupMemberInfo(&onebot.GetGroupMemberInfoParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_get_member_info_failed_user_log|获取用户群成员信息失败"), userID, common.T("", "in_group|在群"), groupID, common.T("", "failed|失败"), err)
		return false
	}

	// 检查用户是否为群主
	if memberData, ok := memberInfo.Data.(map[string]interface{}); ok {
		if role, ok := memberData["role"].(string); ok {
			return role == "owner"
		}
	}

	// 如果无法获取角色信息，返回false
	return false
}

// 发送欢迎消息和群规
func (p *GroupManagerPlugin) sendWelcomeAndRules(robot plugin.Robot, event *onebot.Event) {
	// 发送欢迎消息
	welcomeMsg := fmt.Sprintf(common.T("", "group_manager_welcome_member|欢迎新成员 %d 加入本群！"), event.UserID)

	// 从数据库获取群规
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	rules, err := db.GetGroupRules(p.db, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %s: %v", common.T("", "group_manager_get_rules|获取群规"), event.GroupID, common.T("", "failed|失败"), common.T("", "failed|失败"), err)
		// 使用默认群规
		if err == sql.ErrNoRows {
			defaultRules, err := db.GetGroupRules(p.db, "0")
			if err != nil {
				log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_get_default_rules_failed|获取默认群规失败"), err)
				rules = ""
			} else {
				rules = defaultRules
			}
		}
	}

	if rules == "" {
		// 如果数据库中没有群规，使用默认群规
		rules = common.T("", "group_manager_default_rules|暂无群规")
		log.Printf("[GroupManager] %s", common.T("", "group_manager_use_builtin_rules|使用内置默认群规"))
	}

	// 合并消息
	fullMsg := welcomeMsg + "\n" + rules

	// 发送消息
	p.sendMessage(robot, event, fullMsg)

	// 记录邀请统计
	if event.OperatorID != 0 && event.OperatorID != event.UserID {
		// 邀请者ID和被邀请者ID different，说明是邀请加入
		inviterIDStr := fmt.Sprintf("%d", event.OperatorID)
		inviteeIDStr := fmt.Sprintf("%d", event.UserID)

		// 更新邀请统计
		err = p.updateInvitationCount(groupIDStr, inviterIDStr, inviteeIDStr)
		if err != nil {
			log.Printf("[GroupManager] 更新邀请统计失败: %v", err)
		}
	}
}

// 更新邀请统计
func (p *GroupManagerPlugin) updateInvitationCount(groupID, inviterID, inviteeID string) error {
	if p.db == nil {
		return fmt.Errorf("数据库未配置")
	}

	// 检查是否已经记录过该邀请
	var count int
	query := "SELECT COUNT(*) FROM group_invitations WHERE group_id = ? AND inviter_id = ? AND invitee_id = ?"
	err := p.db.QueryRow(query, groupID, inviterID, inviteeID).Scan(&count)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("检查邀请记录失败: %v", err)
		}
	}

	if count > 0 {
		// 已经记录过，不重复记录
		return nil
	}

	// 插入新的邀请记录
	insertQuery := "INSERT INTO group_invitations (group_id, inviter_id, invitee_id, invite_time) VALUES (?, ?, ?, ?)"
	_, err = p.db.Exec(insertQuery, groupID, inviterID, inviteeID, time.Now())
	if err != nil {
		return fmt.Errorf("插入邀请记录失败: %v", err)
	}

	// 更新邀请者的邀请次数
	updateQuery := "INSERT INTO group_invitation_stats (group_id, user_id, invitation_count) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE invitation_count = invitation_count + 1"
	_, err = p.db.Exec(updateQuery, groupID, inviterID)
	if err != nil {
		return fmt.Errorf("更新邀请统计失败: %v", err)
	}

	log.Printf("[GroupManager] 邀请统计更新成功: 群 %s, 邀请者 %s, 被邀请者 %s", groupID, inviterID, inviteeID)
	return nil
}

// 发送群规
func (p *GroupManagerPlugin) sendGroupRules(robot plugin.Robot, event *onebot.Event) {
	// 从数据库获取群规
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	rules, err := db.GetGroupRules(p.db, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] %s %d %s %s: %v", common.T("", "group_manager_get_rules|获取群规"), event.GroupID, common.T("", "failed|失败"), common.T("", "failed|失败"), err)
		// 使用默认群规
		if err == sql.ErrNoRows {
			defaultRules, err := db.GetGroupRules(p.db, "0")
			if err != nil {
				log.Printf("[GroupManager] %s: %v", common.T("", "group_manager_get_default_rules_failed|获取默认群规失败"), err)
				rules = ""
			} else {
				rules = defaultRules
			}
		}
	}

	if rules == "" {
		// 如果数据库中没有群规，使用默认群规
		rules = common.T("", "group_manager_default_rules|暂无群规")
		log.Printf("[GroupManager] %s", common.T("", "group_manager_use_builtin_rules|使用内置默认群规"))
	}

	// 发送群规
	p.sendMessage(robot, event, common.T("", "group_manager_rules_prefix|本群群规如下：")+"\n"+rules)
}

// 发送帮助信息
func (p *GroupManagerPlugin) sendHelp(robot plugin.Robot, event *onebot.Event) {
	helpMsg := common.T("", "group_manager_help_msg|群管理帮助信息：\nkick <user_id> - 踢人\nban <user_id> [minutes] - 禁言\nunban <user_id> - 解除禁言\naddadmin <user_id> - 添加管理员\ndeladmin <user_id> - 删除管理员\nsetrules <rules> - 设置群规\naddword [level] <word> - 添加敏感词\ndelword <word> - 删除敏感词\nmembers - 查看成员列表\nmemberinfo <user_id> - 查看成员信息\nsettitle <user_id> <title> - 设置头衔")
	p.sendMessage(robot, event, helpMsg)
}

// 定期检查禁言时间
func (p *GroupManagerPlugin) checkBanExpiration(robot plugin.Robot) {
	if robot == nil {
		return
	}
	for {
		// 每隔1分钟检查一次
		time.Sleep(1 * time.Minute)

		// 检查Redis中的禁言记录
		if p.redisClient != nil {
			ctx := context.Background()
			var cursor uint64 = 0

			for {
				// 使用SCAN命令遍历所有禁言记录
				keys, nextCursor, err := p.redisClient.Scan(ctx, cursor, "group:*:ban:*", 10).Result()
				if err != nil {
					log.Printf("%s: %v", common.T("", "group_manager_redis_get_ban_failed|从Redis获取禁言记录失败"), err)
					break
				}

				// 处理每个禁言记录
				for _, key := range keys {
					// 获取禁言过期时间
					banEndTimeStr, err := p.redisClient.Get(ctx, key).Result()
					if err != nil {
						log.Printf("%s: %v", common.T("", "group_manager_redis_get_ban_key_failed|从Redis获取禁言键失败"), err)
						continue
					}

					banEndTime, err := strconv.ParseInt(banEndTimeStr, 10, 64)
					if err != nil {
						log.Printf("%s: %v", common.T("", "group_manager_redis_parse_ban_time_failed|解析Redis禁言时间失败"), err)
						continue
					}

					// 检查是否过期
					if time.Now().Unix() >= banEndTime {
						// 解析groupID和userID
						parts := strings.Split(key, ":")
						if len(parts) != 4 {
							log.Printf("%s: %s", common.T("", "group_manager_invalid_ban_key|无效的禁言键"), key)
							continue
						}

						groupIDStr := parts[1]
						userIDStr := parts[3]
						groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
						if err != nil {
							log.Printf("%s: %v", common.T("", "group_manager_convert_groupid_failed|转换群ID失败"), err)
							continue
						}
						userID, err := strconv.ParseInt(userIDStr, 10, 64)
						if err != nil {
							log.Printf("%s: %v", common.T("", "group_manager_convert_userid_failed|转换用户ID失败"), err)
							continue
						}

						// 解除禁言
						_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
							GroupID:  groupID,
							UserID:   userID,
							Duration: 0,
						})
						if err != nil {
							log.Printf("%s: %v", common.T("", "group_manager_unban_failed_log|解除禁言失败"), err)
							continue
						}

						// 从Redis移除禁言记录
						if err := p.redisClient.Del(ctx, key).Err(); err != nil {
							log.Printf("%s: %v", common.T("", "group_manager_redis_del_ban_failed|从Redis删除禁言记录失败"), err)
						}

						// 同时从数据库移除禁言记录（如果存在）
						if p.db != nil {
							if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
								log.Printf("%s: %v", common.T("", "group_manager_db_del_ban_failed|从数据库删除禁言记录失败"), err)
							}
						}

						// 发送通知
						p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf(common.T("", "group_manager_ban_expired_msg|用户 %d 的禁言已到期，已自动解除"), userID))
					}
				}

				// 检查是否遍历完毕
				if nextCursor == 0 {
					break
				}
				cursor = nextCursor
			}
		}

		// 同时检查数据库中的禁言记录（作为后备）
		if p.db != nil {
			// 从数据库获取所有过期的禁言记录
			expiredBans, err := db.GetExpiredBans(p.db)
			if err != nil {
				log.Printf("%s: %v", common.T("", "group_manager_get_expired_bans_failed|获取过期禁言记录失败"), err)
				continue
			}

			// 遍历所有过期的禁言记录
			for _, ban := range expiredBans {
				// 转换groupID和userID为int64
				groupIDStr := ban["group_id"].(string)
				userIDStr := ban["user_id"].(string)
				groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
				if err != nil {
					log.Printf("%s: %v", common.T("", "group_manager_convert_groupid_failed|转换群ID失败"), err)
					continue
				}
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err != nil {
					log.Printf("%s: %v", common.T("", "group_manager_convert_userid_failed|转换用户ID失败"), err)
					continue
				}

				// 解除禁言
				_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
					GroupID:  groupID,
					UserID:   userID,
					Duration: 0,
				})
				if err != nil {
					log.Printf("%s: %v", common.T("", "group_manager_unban_failed_log|解除禁言失败"), err)
					continue
				}

				// 从数据库移除禁言记录
				if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
					log.Printf("%s: %v", common.T("", "group_manager_db_del_ban_failed|从数据库删除禁言记录失败"), err)
					continue
				}

				// 发送通知
				p.sendMessage(robot, &onebot.Event{GroupID: groupID}, fmt.Sprintf(common.T("", "group_manager_ban_expired_msg|用户 %d 的禁言已到期，已自动解除"), userID))
			}
		}
	}
}

// 解析用户ID
func parseUserID(str string) (int64, error) {
	// 处理 @ 开头的用户ID
	if strings.HasPrefix(str, "@") {
		str = str[1:]
	}

	// 提取数字
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(str)
	if numStr == "" {
		return 0, fmt.Errorf(common.T("", "group_manager_invalid_userid_err|无效的用户ID格式"))
	}

	// 转换为int64
	userID := int64(0)
	for _, c := range numStr {
		userID = userID*10 + int64(c-'0')
	}

	return userID, nil
}

// 解析时长
func parseDuration(str string) (int, error) {
	// 提取数字
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(str)
	if numStr == "" {
		return 0, fmt.Errorf(common.T("", "group_manager_invalid_duration_err|无效的时长格式"))
	}

	// 转换为int
	duration := 0
	for _, c := range numStr {
		duration = duration*10 + int(c-'0')
	}

	return duration, nil
}

// 处理获取群成员列表命令
func (p *GroupManagerPlugin) handleGetMembersLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	// 只有管理员可以查看群成员列表
	if !p.isAdmin(event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_admin_view_members|权限不足，只有管理员可以查看群成员列表"))
		return nil
	}

	if robot == nil {
		return nil
	}

	// 调用OneBot API获取群成员列表
	resp, err := robot.GetGroupMemberList(&onebot.GetGroupMemberListParams{
		GroupID: event.GroupID,
		NoCache: true,
	})

	if err != nil {
		log.Printf("[GroupManager] %s %d %s: %v", common.T("", "group_manager_get_member_list_failed_log|获取群成员列表失败"), event.GroupID, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_get_member_list_failed|获取群成员列表失败"), err))
		return nil
	}

	// 解析返回数据
	memberList, ok := resp.Data.([]interface{})
	if !ok {
		log.Printf("[GroupManager] %s: %T", common.T("", "group_manager_parse_member_list_failed_log|解析群成员列表数据失败"), resp.Data)
		p.sendMessage(robot, event, common.T("", "group_manager_parse_member_list_failed|解析群成员列表失败"))
		return nil
	}

	// 格式化群成员信息
	var membersInfo strings.Builder
	membersInfo.WriteString(fmt.Sprintf(common.T("", "group_manager_member_list_title|群 %d 成员列表（共 %d 人）：\n"), event.GroupID, len(memberList)))

	for i, member := range memberList {
		memberMap, ok := member.(map[string]interface{})
		if !ok {
			continue
		}

		userID, _ := memberMap["user_id"].(float64)
		nickname, _ := memberMap["nickname"].(string)
		card, _ := memberMap["card"].(string)
		sex, _ := memberMap["sex"].(string)
		joinTime, _ := memberMap["join_time"].(float64)

		// 显示群名片或昵称
		name := nickname
		if card != "" {
			name = card
		}

		// 格式化加入时间
		joinDate := time.Unix(int64(joinTime), 0).Format("2006-01-02")

		// 添加到信息字符串
		membersInfo.WriteString(fmt.Sprintf(common.T("", "group_manager_member_list_item|%d. %d (%s) [%s] 加入时间: %s\n"),
			i+1, int64(userID), name, sex, joinDate))

		// 每50个成员发送一次消息，避免消息过长
		if (i+1)%50 == 0 || i == len(memberList)-1 {
			p.sendMessage(robot, event, membersInfo.String())
			membersInfo.Reset()
			membersInfo.WriteString(fmt.Sprintf(common.T("", "group_manager_member_list_cont|群 %d 成员列表（续）：\n"), event.GroupID))
		}
	}
	return nil
}

// 处理获取群成员信息命令
func (p *GroupManagerPlugin) handleGetMemberInfoLogic(robot plugin.Robot, event *onebot.Event, args []string) error {
	// 只有管理员可以查看群成员信息
	if !p.isAdmin(event.GroupID, event.UserID) {
		p.sendMessage(robot, event, common.T("", "group_manager_insufficient_perms_admin_view_info|权限不足，只有管理员可以查看群成员信息"))
		return nil
	}

	if len(args) < 1 {
		p.sendMessage(robot, event, common.T("", "group_manager_memberinfo_usage|用法：memberinfo <user_id>"))
		return nil
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		p.sendMessage(robot, event, common.T("", "group_manager_invalid_userid|无效的用户ID"))
		log.Printf("[GroupManager] %s '%s' %s: %v", common.T("", "group_manager_parse_userid_failed|解析用户ID失败"), args[0], common.T("", "failed|失败"), err)
		return nil
	}

	if robot == nil {
		return nil
	}

	// 调用OneBot API获取群成员信息
	resp, err := robot.GetGroupMemberInfo(&onebot.GetGroupMemberInfoParams{
		GroupID: event.GroupID,
		UserID:  userID,
		NoCache: true,
	})

	if err != nil {
		log.Printf("[GroupManager] %s %d %s %d %s: %v", common.T("", "group_manager_get_member_info_failed_member_log|获取用户群成员信息失败"), event.GroupID, common.T("", "of_user|用户"), userID, common.T("", "failed|失败"), err)
		p.sendMessage(robot, event, fmt.Sprintf("%s: %v", common.T("", "group_manager_get_member_info_failed|获取群成员信息失败"), err))
		return nil
	}

	// 解析返回数据
	memberInfo, ok := resp.Data.(map[string]interface{})
	if !ok {
		log.Printf("[GroupManager] %s: %T", common.T("", "group_manager_parse_member_info_failed_log|解析群成员信息失败"), resp.Data)
		p.sendMessage(robot, event, common.T("", "group_manager_parse_member_info_failed|解析群成员信息失败"))
		return nil
	}

	// 提取成员信息
	userIDFloat, _ := memberInfo["user_id"].(float64)
	nickname, _ := memberInfo["nickname"].(string)
	card, _ := memberInfo["card"].(string)
	sex, _ := memberInfo["sex"].(string)
	age, _ := memberInfo["age"].(float64)
	joinTime, _ := memberInfo["join_time"].(float64)
	lastSentTime, _ := memberInfo["last_sent_time"].(float64)
	level, _ := memberInfo["level"].(float64)
	role, _ := memberInfo["role"].(string)

	// 显示群名片或昵称
	name := nickname
	if card != "" {
		name = card
	}

	// 格式化时间
	joinDate := time.Unix(int64(joinTime), 0).Format("2006-01-02 15:04:05")
	lastSentDate := time.Unix(int64(lastSentTime), 0).Format("2006-01-02 15:04:05")

	// 格式化成员信息
	memberDetail := fmt.Sprintf(
		common.T("", "group_manager_member_detail|用户详细信息：\nID: %d\n昵称: %s\n名片: %s\n性别: %s\n年龄: %d\n加入时间: %s\n最后发言: %s\n等级: %d\n角色: %s"),
		int64(userIDFloat), name, card, sex, int(age), joinDate, lastSentDate, int(level), role)

	// 发送成员信息
	p.sendMessage(robot, event, memberDetail)
	return nil
}
