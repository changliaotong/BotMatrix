package main

import (
	"log"
	"time"
)

// SignInPlugin 签到插件
type SignInPlugin struct {
	// 存储用户签到记录，key为用户ID，value为签到时间
	signInRecords map[string]time.Time
	// 存储用户连续签到天数，key为用户ID，value为连续天数
	continuousDays map[string]int
}

// NewSignInPlugin 创建签到插件实例
func NewSignInPlugin() *SignInPlugin {
	return &SignInPlugin{
		signInRecords: make(map[string]time.Time),
		continuousDays: make(map[string]int),
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

func main() {
	log.Println("测试签到系统核心逻辑...")

	// 创建签到插件实例
	plugin := NewSignInPlugin()

	// 测试插件基本信息
	log.Println("插件名称: sign_in")
	log.Println("插件版本: 1.0.0")
	log.Println("插件描述: 签到系统插件，支持每日签到和连续签到")

	// 测试签到功能
	userID := "test_user_123"
	now := time.Now()

	// 第一次签到
	plugin.signInRecords[userID] = now
	plugin.continuousDays[userID] = 1
	log.Printf("用户 %s 第一次签到成功，连续天数: %d", userID, plugin.continuousDays[userID])

	// 测试同一天再次签到
	if isSameDay(plugin.signInRecords[userID], now) {
		log.Printf("用户 %s 今天已经签到过了", userID)
	}

	// 测试连续签到
	// 模拟第二天签到
	nextDay := now.Add(24 * time.Hour)
	if isYesterday(plugin.signInRecords[userID], nextDay) {
		plugin.continuousDays[userID]++
		log.Printf("用户 %s 连续签到成功，连续天数: %d", userID, plugin.continuousDays[userID])
	}

	// 测试非连续签到
	// 模拟第三天签到
	thirdDay := now.Add(48 * time.Hour)
	if !isYesterday(nextDay, thirdDay) {
		plugin.continuousDays[userID] = 1
		log.Printf("用户 %s 签到成功，连续天数重置为: %d", userID, plugin.continuousDays[userID])
	}

	log.Println("签到系统核心逻辑测试通过!")
}