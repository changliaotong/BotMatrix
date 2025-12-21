package main

import (
	"fmt"
	"log"
	"time"
)

// PointsPlugin 积分系统插件
type PointsPlugin struct {
	// 存储用户积分，key为用户ID，value为积分数量
	points map[string]int
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin() *PointsPlugin {
	return &PointsPlugin{
		points: make(map[string]int),
	}
}

// addPoints 增加用户积分
func (p *PointsPlugin) addPoints(userID string, points int, reason string) {
	// 增加积分
	p.points[userID] += points
	log.Printf("用户 %s 获得%d积分，原因: %s", userID, points, reason)
}

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

	return false
}

func main() {
	log.Println("测试签到系统积分奖励功能...")

	// 创建积分插件实例
	pointsPlugin := NewPointsPlugin()

	// 创建签到插件实例（传递积分插件引用）
	signInPlugin := NewSignInPlugin(pointsPlugin)

	// 测试用户
	userID := "test_user_123"
	now := time.Now()

	// 第一次签到
	log.Println("第一次签到:")
	signInPlugin.signInRecords[userID] = now
	signInPlugin.continuousDays[userID] = 1

	basePoints := 10
	extraPoints := 0
	totalPoints := basePoints + extraPoints

	if signInPlugin.pointsPlugin != nil {
		signInPlugin.pointsPlugin.addPoints(userID, totalPoints, fmt.Sprintf("签到奖励（连续%d天）", 1))
	}

	log.Printf("签到成功！连续天数：1天，获得积分：%d（基础%d + 连续%d）", totalPoints, basePoints, extraPoints)
	log.Printf("当前积分：%d", pointsPlugin.points[userID])

	// 第二次签到（连续签到）
	log.Println("\n第二次签到（连续签到）:")
	nextDay := now.Add(24 * time.Hour)
	if isYesterday(signInPlugin.signInRecords[userID], nextDay) {
		signInPlugin.continuousDays[userID]++
	}

	signInPlugin.signInRecords[userID] = nextDay

	basePoints = 10
	extraPoints = signInPlugin.continuousDays[userID] - 1
	totalPoints = basePoints + extraPoints

	if signInPlugin.pointsPlugin != nil {
		signInPlugin.pointsPlugin.addPoints(userID, totalPoints, fmt.Sprintf("签到奖励（连续%d天）", signInPlugin.continuousDays[userID]))
	}

	log.Printf("签到成功！连续天数：%d天，获得积分：%d（基础%d + 连续%d）",
		signInPlugin.continuousDays[userID], totalPoints, basePoints, extraPoints)
	log.Printf("当前积分：%d", pointsPlugin.points[userID])

	log.Println("\n签到系统积分奖励功能测试通过!")
}
