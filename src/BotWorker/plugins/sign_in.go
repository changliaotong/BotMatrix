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
	return common.T("", "signin_plugin_desc")
}

func (p *SignInPlugin) Version() string {
	return "1.0.0"
}

func (p *SignInPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "signin_plugin_loaded"))

	// 处理签到命令
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

		// 检查是否为签到命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "signin_cmd_sign"), event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "signin_no_userid"))
			return nil
		}

		// 执行签到
		userIDStr := fmt.Sprintf("%d", userID)
		p.processSignIn(robot, event, userIDStr)

		return nil
	})

	// 自动签到功能：群或私聊发言时自动签到
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

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			return nil
		}

		// 检查是否已经签到
		now := time.Now()
		userIDStr := fmt.Sprintf("%d", userID)
		if lastSignIn, ok := p.signInRecords[userIDStr]; ok {
			// 检查是否在同一天
			if isSameDay(lastSignIn, now) {
				return nil // 已经签到过了
			}
		}

		// 执行自动签到
		p.processSignIn(robot, event, userIDStr)

		return nil
	})

	// 处理签到统计命令
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

		if match, _ := p.cmdParser.MatchCommand(common.T("", "signin_cmd_stats"), event.RawMessage); !match {
			return nil
		}

		// 发送签到统计信息
		statsMsg := fmt.Sprintf(common.T("", "signin_stats_msg"),
			len(p.signInRecords), p.getTodaySignInCount())
		p.sendMessage(robot, event, statsMsg)

		return nil
	})
}

// processSignIn 处理签到逻辑
func (p *SignInPlugin) processSignIn(robot plugin.Robot, event *onebot.Event, userID string) {
	now := time.Now()
	continuousDay := 1
	if lastSignIn, ok := p.signInRecords[userID]; ok {
		if isSameDay(lastSignIn, now) {
			continuousDay := p.continuousDays[userID]
			totalDays := continuousDay
			superPoints := 0
			if p.pointsPlugin != nil {
				superPoints = p.pointsPlugin.GetPoints(userID)
			}
			todaySignCount := p.getTodaySignInCount()
			msg := fmt.Sprintf(common.T("", "signin_already_signed"),
				superPoints,
				0, 0,
				continuousDay, totalDays,
				0, 0,
				todaySignCount, 0,
			)
			p.sendMessage(robot, event, msg)
			return
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
		p.pointsPlugin.AddPoints(userID, totalPoints, fmt.Sprintf(common.T("", "signin_reward_desc"), continuousDay), "sign_in")
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
	msg := fmt.Sprintf(common.T("", "signin_success_msg"),
		totalPoints, currentPoints,
		0, 0,
		continuousDay, totalDays,
		0, 0,
		todaySignCount, 0,
	)
	p.sendMessage(robot, event, msg)
}

// sendMessage 发送消息
func (p *SignInPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "signin_send_failed_log"), err)
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
