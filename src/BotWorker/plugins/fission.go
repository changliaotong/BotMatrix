package plugins

import (
	"BotMatrix/common"
	"botworker/internal/db"
	"botworker/internal/fission"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// FissionPlugin 裂变系统插件
type FissionPlugin struct {
	db        *sql.DB
	service   *fission.Service
	cmdParser *CommandParser
}

// NewFissionPlugin 创建裂变系统插件实例
func NewFissionPlugin(database *sql.DB) *FissionPlugin {
	return &FissionPlugin{
		db:        database,
		cmdParser: NewCommandParser(),
	}
}

func (p *FissionPlugin) Name() string {
	return "fission"
}

func (p *FissionPlugin) Description() string {
	return common.T("", "fission_plugin_desc|裂变系统插件，支持邀请、奖励、排行榜等功能")
}

func (p *FissionPlugin) Version() string {
	return "1.0.0"
}

func (p *FissionPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "get_invite_code",
			Description: common.T("", "fission_skill_get_invite_code_desc|获取邀请码"),
			Usage:       "get_invite_code",
			Params:      map[string]string{},
		},
		{
			Name:        "get_fission_stats",
			Description: common.T("", "fission_skill_get_stats_desc|获取裂变统计"),
			Usage:       "get_fission_stats",
			Params:      map[string]string{},
		},
	}
}

func (p *FissionPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "fission_db_not_configured|裂变插件初始化失败：数据库未配置"))
		return
	}
	p.service = fission.NewService(p.db)
	log.Println(common.T("", "fission_plugin_loaded|裂变系统插件已加载"))

	// 统一处理裂变相关命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		userID := event.UserID.Int64()
		if userID == 0 {
			return nil
		}

		// 1. 触发任务逻辑 (解耦到 Service)
		p.service.TriggerTask(userID, "usage")

		// 1. 获取邀请码
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fission_cmd_invite|邀请|邀请码|invite"), event.RawMessage); match {
			msg, err := p.doGetInviteCode(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 2. 查看奖励/进度
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fission_cmd_rewards|奖励|进度|rewards"), event.RawMessage); match {
			msg, err := p.doGetFissionStats(userID)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 3. 裂变排行榜
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fission_cmd_rank|裂变榜|邀请榜|fissionrank"), event.RawMessage); match {
			msg, err := p.doGetFissionRank()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 4. 绑定邀请码
		matchBind, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "fission_cmd_bind|绑定|填写邀请码|bind"), `([a-zA-Z0-9]+)`, event.RawMessage)
		if matchBind && len(params) == 1 {
			msg, err := p.doBindInviteCode(userID, params[0], event.Platform)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 5. 查看任务
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fission_cmd_tasks|任务|裂变任务|tasks"), event.RawMessage); match {
			msg, err := p.doGetFissionTasks()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})

	// 处理通知事件 (如进群)
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_increase" {
			userID := event.UserID.Int64()
			if userID != 0 {
				// 触发进群任务奖励
				_ = db.CompleteFissionTask(p.db, userID, "group_join")
			}
		}
		return nil
	})
}

// sendMessage 发送消息
func (p *FissionPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送裂变消息失败: %v", err)
	}
}

// doGetInviteCode 获取或生成邀请码
func (p *FissionPlugin) doGetInviteCode(userID int64) (string, error) {
	// 1. 获取配置以获取链接模板
	config, _ := db.GetFissionConfig(p.db)

	// 2. 生成邀请码
	inviteCode := fmt.Sprintf("U%X", userID)

	// 3. 生成邀请链接 (如果有模板)
	inviteLink := ""
	if config.InviteCodeTemplate != "" {
		// 假设模板中有 {CODE} 占位符
		inviteLink = strings.Replace(config.InviteCodeTemplate, "{CODE}", inviteCode, -1)
		// 如果模板只是一个前缀，直接拼接
		if !strings.Contains(config.InviteCodeTemplate, "{CODE}") {
			inviteLink = config.InviteCodeTemplate + inviteCode
		}
	}

	msg := fmt.Sprintf(common.T("", "fission_invite_code_msg|🎁 您的专属邀请码为：【%s】\n"), inviteCode)
	if inviteLink != "" {
		msg += fmt.Sprintf(common.T("", "fission_invite_link_msg|🔗 专属推广链接：%s\n"), inviteLink)
	}
	msg += common.T("", "fission_invite_guide|发送给好友，让好友发送“绑定 %s”即可完成绑定！\n每成功邀请一位新用户，您将获得积分奖励。")
	msg = strings.Replace(msg, "%s", inviteCode, -1)

	return msg, nil
}

// doGetFissionStats 查看奖励/进度
func (p *FissionPlugin) doGetFissionStats(userID int64) (string, error) {
	stats, err := p.service.GetUserStats(userID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "fission_get_stats_failed|❌ 获取统计信息失败"))
	}

	msg := common.T("", "fission_stats_header|📊 您的裂变进度：\n")
	msg += fmt.Sprintf(common.T("", "fission_stats_invite_code|🔹 我的邀请码: %v\n"), stats["invite_code"])
	msg += fmt.Sprintf(common.T("", "fission_stats_invite_count|🔹 累计邀请: %v 人\n"), stats["invite_count"])
	msg += fmt.Sprintf(common.T("", "fission_stats_points|🔹 累计获得积分: %v\n"), stats["points"])
	msg += fmt.Sprintf(common.T("", "fission_stats_level|🔹 推广等级: LV%v\n"), stats["level"])
	msg += "------------------------\n"
	msg += common.T("", "fission_stats_footer|发送“任务”查看更多裂变任务奖励")

	return msg, nil
}

// doGetFissionRank 获取裂变排行榜
func (p *FissionPlugin) doGetFissionRank() (string, error) {
	rank, err := db.GetFissionRank(p.db, 10)
	if err != nil {
		return "", err
	}

	if len(rank) == 0 {
		return common.T("", "fission_rank_empty|暂无邀请排行数据"), nil
	}

	msg := "🏆 邀请达人榜 (Top 10)\n"
	msg += "------------------------\n"
	for i, item := range rank {
		medal := fmt.Sprintf("%d.", i+1)
		if i == 0 {
			medal = "🥇"
		} else if i == 1 {
			medal = "🥈"
		} else if i == 2 {
			medal = "🥉"
		}
		msg += fmt.Sprintf("%s 用户(%v): 邀请 %v 人\n", medal, item["user_id"], item["invite_count"])
	}
	msg += "------------------------"
	return msg, nil
}

// doBindInviteCode 绑定邀请码
func (p *FissionPlugin) doBindInviteCode(userID int64, code string, platform string) (string, error) {
	// 1. 解析邀请码获取邀请者ID
	if !strings.HasPrefix(code, "U") {
		return "", fmt.Errorf(common.T("", "fission_invalid_code|❌ 邀请码格式错误"))
	}

	inviterIDStr := code[1:]
	var inviterID int64
	_, err := fmt.Sscanf(inviterIDStr, "%X", &inviterID)
	if err != nil {
		return "", fmt.Errorf(common.T("", "fission_invalid_code|❌ 邀请码格式错误"))
	}

	// 2. 调用核心服务处理绑定逻辑
	// 注意：这里暂时没有 IP 和 DeviceID，传空字符串
	msg, err := p.service.ProcessBind(inviterID, userID, platform, code, "", "")
	if err != nil {
		return "", err
	}

	return common.T("", "fission_bind_success_custom|✅ %s", msg), nil
}

// doGetFissionTasks 获取裂变任务列表
func (p *FissionPlugin) doGetFissionTasks() (string, error) {
	tasks, err := db.GetActiveFissionTasks(p.db)
	if err != nil {
		return "", err
	}

	if len(tasks) == 0 {
		return common.T("", "fission_tasks_empty|🎁 当前暂无可领取的裂变任务"), nil
	}

	msg := "🎁 裂变任务列表：\n"
	msg += "------------------------\n"
	for _, t := range tasks {
		reward := ""
		if t.RewardPoints > 0 {
			reward += fmt.Sprintf("%d 积分 ", t.RewardPoints)
		}
		if t.RewardDuration > 0 {
			reward += fmt.Sprintf("%d 小时时长 ", t.RewardDuration)
		}
		msg += fmt.Sprintf("【%s】\n内容：%s\n奖励：%s\n\n", t.Name, t.Description, reward)
	}
	msg += "------------------------\n"
	msg += "快去邀请好友完成任务获取奖励吧！"
	return msg, nil
}

func init() {
	// 这里可以添加随机数种子初始化
	rand.Seed(time.Now().UnixNano())
}
