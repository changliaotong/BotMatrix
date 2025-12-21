package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strings"
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
}

// NewSignInPlugin 创建签到插件实例
func NewSignInPlugin(pointsPlugin *PointsPlugin) *SignInPlugin {
	return &SignInPlugin{
		signInRecords:  make(map[string]time.Time),
		continuousDays: make(map[string]int),
		pointsPlugin:   pointsPlugin,
	}
}

func (p *SignInPlugin) Name() string {
	return "sign_in"
}

func (p *SignInPlugin) Description() string {
	return "签到系统插件，支持每日签到和连续签到"
}

func (p *SignInPlugin) Version() string {
	return "1.0.0"
}

func (p *SignInPlugin) Init(robot plugin.Robot) {
	log.Println("加载签到系统插件")

	// 处理签到命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为签到命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!sign" && msg != "!签到" {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，签到失败")
			return nil
		}

		// 执行签到
		p.processSignIn(robot, event, userID)

		return nil
	})

	// 自动签到功能：群或私聊发言时自动签到
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			return nil
		}

		// 检查是否已经签到
		now := time.Now()
		if lastSignIn, ok := p.signInRecords[userID]; ok {
			// 检查是否在同一天
			if isSameDay(lastSignIn, now) {
				return nil // 已经签到过了
			}
		}

		// 执行自动签到
		p.processSignIn(robot, event, userID)

		return nil
	})

	// 处理签到统计命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!signstats" && msg != "!签到统计" {
			return nil
		}

		// 发送签到统计信息
		statsMsg := fmt.Sprintf("签到系统统计信息：\n总签到人数：%d\n今日签到人数：%d",
			len(p.signInRecords), p.getTodaySignInCount())
		p.sendMessage(robot, event, statsMsg)

		return nil
	})
}

// processSignIn 处理签到逻辑
func (p *SignInPlugin) processSignIn(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否已经签到
	now := time.Now()
	if lastSignIn, ok := p.signInRecords[userID]; ok {
		// 检查是否在同一天
		if isSameDay(lastSignIn, now) {
			return // 已经签到过了
		}
	}

	// 计算连续签到天数
	continuousDay := 1
	if lastSignIn, ok := p.signInRecords[userID]; ok {
		// 检查是否连续
		if isYesterday(lastSignIn, now) {
			continuousDay = p.continuousDays[userID] + 1
		} else {
			continuousDay = 1
		}
	}

	// 更新签到记录
	p.signInRecords[userID] = now
	p.continuousDays[userID] = continuousDay

	// 奖励积分（基础10积分，连续签到额外奖励）
	basePoints := 10
	extraPoints := 0
	if continuousDay > 1 {
		extraPoints = continuousDay - 1 // 连续签到每天额外奖励1积分
	}
	totalPoints := basePoints + extraPoints

	// 添加积分
	if p.pointsPlugin != nil {
		p.pointsPlugin.addPoints(userID, totalPoints, fmt.Sprintf("签到奖励（连续%d天）", continuousDay))
	}

	// 发送签到成功消息
	msg := fmt.Sprintf("签到成功！\n今日签到时间：%s\n连续签到天数：%d天\n获得积分：%d（基础%d + 连续%d）",
		now.Format("15:04:05"), continuousDay, totalPoints, basePoints, extraPoints)
	p.sendMessage(robot, event, msg)
}

// sendMessage 发送消息
func (p *SignInPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// isSameDay 检查两个时间是否在同一天
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
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
