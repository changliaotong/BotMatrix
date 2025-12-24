package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"time"
)

// SignInPlugin 签到插件
type SignInPlugin struct {
	// 存储用户签到记录，key为用户ID，value为签到时间
	signInRecords map[string]time.Time
	// 存储用户连续签到天数，key为用户ID，value为连续天数
	continuousDays map[string]int
	// 积分插件引用
	pointsPlugin *PointsPlugin
	// 命令解析器
	cmdParser *CommandParser
}

// NewSignInPlugin 创建签到插件实例
func NewSignInPlugin(pointsPlugin *PointsPlugin) *SignInPlugin {
	return &SignInPlugin{
		signInRecords:  make(map[string]time.Time),
		continuousDays: make(map[string]int),
		pointsPlugin:   pointsPlugin,
		cmdParser:      NewCommandParser(),
	}
}

func (p *SignInPlugin) Name() string {
	return "sign_in"
}

func (p *SignInPlugin) Description() string {
	return common.T("", "signin_plugin_desc|📅 签到系统插件，支持每日签到和连续签到统计")
}

func (p *SignInPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *SignInPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "signin",
			Description: common.T("", "signin_skill_signin_desc|执行每日签到"),
			Usage:       "signin user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "signin_skill_param_userid|用户ID"),
			},
		},
		{
			Name:        "get_signin_stats",
			Description: common.T("", "signin_skill_stats_desc|获取签到统计信息"),
			Usage:       "get_signin_stats",
			Params:      map[string]string{},
		},
	}
}

// HandleSkill 处理技能调用
func (p *SignInPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "signin":
		userID := params["user_id"]
		if userID == "" {
			return "", fmt.Errorf(common.T("", "signin_missing_userid|❌ 缺少用户ID"))
		}
		msg := p.doSignIn(userID)
		p.sendMessage(robot, event, msg)
		return msg, nil
	case "get_signin_stats":
		msg := p.doGetSignInStats()
		p.sendMessage(robot, event, msg)
		return msg, nil
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

func (p *SignInPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "signin_plugin_loaded|✅ 签到系统插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理签到相关命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "signin") {
				HandleFeatureDisabled(robot, event, "signin")
				return nil
			}
		}

		userID := event.UserID
		if userID == 0 {
			return nil
		}
		userIDStr := fmt.Sprintf("%d", userID)

		// 1. 处理签到命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "signin_cmd_sign|签到|sign in|打卡|signin"), event.RawMessage); match {
			msg := p.doSignIn(userIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 2. 处理签到统计命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "signin_cmd_stats|签到统计|sign stats|sign_stats"), event.RawMessage); match {
			p.sendMessage(robot, event, p.doGetSignInStats())
			return nil
		}

		// 3. 自动签到逻辑
		now := time.Now()
		if lastSignIn, ok := p.signInRecords[userIDStr]; ok {
			if !isSameDay(lastSignIn, now) {
				p.doSignIn(userIDStr)
			}
		} else {
			p.doSignIn(userIDStr)
		}

		return nil
	})
}

// doSignIn 处理签到逻辑
func (p *SignInPlugin) doSignIn(userID string) string {
	now := time.Now()
	continuousDay := 1
	if lastSignIn, ok := p.signInRecords[userID]; ok {
		if isSameDay(lastSignIn, now) {
			continuousDay := p.continuousDays[userID]
			totalDays := 0
			for _, t := range p.signInRecords {
				if !t.IsZero() {
					totalDays++
				}
			}
			superPoints := 0
			if p.pointsPlugin != nil {
				superPoints = p.pointsPlugin.GetPoints(userID)
			}
			todaySignCount := p.getTodaySignInCount()
			return fmt.Sprintf(common.T("", "signin_already_signed|📅 您今天已经签到过了！\n💰 当前积分：%d\n📈 今日收益：+%d (%d)\n🔥 连续签到：%d 天\n📊 累计签到：%d 天\n🆙 当前等级：Lv.%d (%d/%d)\n🏆 今日第 %d 位签到者\n🔮 今日运势：%d"),
				superPoints,
				0, 0,
				continuousDay, totalDays,
				0, 0,
				todaySignCount, 0,
			)
		}
	}

	if lastSignIn, ok := p.signInRecords[userID]; ok {
		if isYesterday(lastSignIn, now) {
			continuousDay = p.continuousDays[userID] + 1
		} else {
			continuousDay = 1
		}
	}

	// 更新签到记录
	p.signInRecords[userID] = now
	p.continuousDays[userID] = continuousDay

	basePoints := 10
	extraPoints := 0
	if continuousDay > 1 {
		extraPoints = continuousDay - 1
	}
	totalPoints := basePoints + extraPoints

	if p.pointsPlugin != nil {
		p.pointsPlugin.AddPoints(userID, totalPoints, fmt.Sprintf(common.T("", "signin_reward_desc|🎁 第 %d 天连续签到奖励"), continuousDay), "sign_in")
	}

	currentPoints := 0
	if p.pointsPlugin != nil {
		currentPoints = p.pointsPlugin.GetPoints(userID)
	}
	todaySignCount := p.getTodaySignInCount()
	totalDays := 0
	for _, t := range p.signInRecords {
		if !t.IsZero() {
			totalDays++
		}
	}
	return fmt.Sprintf(common.T("", "signin_success_msg|✅ 签到成功！\n💰 获得积分：+%d\n💳 当前总积分：%d\n📈 今日收益：+%d (%d)\n🔥 连续签到：%d 天\n📊 累计签到：%d 天\n🆙 当前等级：Lv.%d (%d/%d)\n🏆 今日第 %d 位签到者\n🔮 今日运势：%d"),
		totalPoints, currentPoints,
		0, 0,
		continuousDay, totalDays,
		0, 0,
		todaySignCount, 0,
	)
}

// doGetSignInStats 获取签到统计信息
func (p *SignInPlugin) doGetSignInStats() string {
	return fmt.Sprintf(common.T("", "signin_stats_msg|📊 当前签到统计：\n👥 累计签到人数：%d\n📅 今日签到人数：%d"),
		len(p.signInRecords), p.getTodaySignInCount())
}

// sendMessage 发送消息
func (p *SignInPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "signin_send_failed_log|❌ 发送签到消息失败，机器人或事件为空"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "signin_send_failed_log|❌ 发送签到消息失败")+": %v", err)
	}
}

// isYesterday 检查t1是否是t2的前一天
func isYesterday(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()

	// 检查是否是前一天
	if y1 == y2 && m1 == m2 && d2 == d1+1 {
		return true
	}

	// 处理跨月的情况
	if m1 != m2 {
		// 检查是否是上个月的最后一天
		lastDayOfMonth := time.Date(y1, m1+1, 0, 0, 0, 0, 0, time.Local).Day()
		if d1 == lastDayOfMonth && d2 == 1 {
			return true
		}
	}

	// 处理跨年的情况
	if y1 != y2 {
		// 检查是否是去年的最后一天
		lastDayOfYear := time.Date(y1, 12, 31, 0, 0, 0, 0, time.Local).Day()
		if m1 == 12 && d1 == lastDayOfYear && y2 == y1+1 && m2 == 1 && d2 == 1 {
			return true
		}
	}

	return false
}

// getTodaySignInCount 获取今日签到人数
func (p *SignInPlugin) getTodaySignInCount() int {
	count := 0
	now := time.Now()
	for _, signInTime := range p.signInRecords {
		if isSameDay(signInTime, now) {
			count++
		}
	}
	return count
}
