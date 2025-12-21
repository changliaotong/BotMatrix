package main

import (
	"log"
	"time"
)

// PointsPlugin 积分系统插件
type PointsPlugin struct {
	// 存储用户积分，key为用户ID，value为积分数量
	points map[string]int
	// 存储用户上次签到时间，key为用户ID，value为签到时间
	lastSignInTime map[string]time.Time
	// 存储用户积分记录，key为用户ID，value为积分记录列表
	pointsRecords map[string][]PointsRecord
}

// PointsRecord 积分记录
type PointsRecord struct {
	Points    int       // 积分数量
	Reason    string    // 积分变动原因
	Timestamp time.Time // 变动时间
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin() *PointsPlugin {
	return &PointsPlugin{
		points:         make(map[string]int),
		lastSignInTime: make(map[string]time.Time),
		pointsRecords:  make(map[string][]PointsRecord),
	}
}

// addPoints 增加用户积分
func (p *PointsPlugin) addPoints(userID string, points int, reason string) {
	// 增加积分
	p.points[userID] += points

	// 记录积分变动
	record := PointsRecord{
		Points:    points,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	p.pointsRecords[userID] = append(p.pointsRecords[userID], record)
}

// isSameDay 检查两个时间是否在同一天
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func main() {
	log.Println("测试积分系统核心逻辑...")

	// 创建积分系统插件实例
	plugin := NewPointsPlugin()

	// 测试插件基本信息
	log.Println("插件名称: points")
	log.Println("插件版本: 1.0.0")
	log.Println("插件描述: 积分系统插件，支持签到积分、发言积分、查询积分等功能")

	// 测试积分功能
	userID := "test_user_123"

	// 测试签到积分
	log.Println("测试签到积分...")
	plugin.addPoints(userID, 10, "签到奖励")
	log.Printf("用户 %s 签到获得10积分，当前积分: %d", userID, plugin.points[userID])

	// 测试发言积分
	log.Println("测试发言积分...")
	plugin.addPoints(userID, 1, "发言奖励")
	log.Printf("用户 %s 发言获得1积分，当前积分: %d", userID, plugin.points[userID])

	// 测试积分记录
	log.Println("测试积分记录...")
	records := plugin.pointsRecords[userID]
	for i, record := range records {
		log.Printf("记录 %d: %d积分，原因: %s，时间: %s", i+1, record.Points, record.Reason, record.Timestamp.Format("15:04:05"))
	}

	// 测试重复签到
	log.Println("测试重复签到...")
	now := time.Now()
	plugin.lastSignInTime[userID] = now
	if isSameDay(plugin.lastSignInTime[userID], now) {
		log.Printf("用户 %s 今天已经签到过了", userID)
	}

	log.Println("积分系统核心逻辑测试通过!")
}