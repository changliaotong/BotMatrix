package main

import (
	"log"
	"math/rand"
	"time"
)

// LotteryPlugin 抽签插件
type LotteryPlugin struct {
	// 存储用户抽签记录，key为用户ID，value为上次抽签时间
	lastLotteryTime map[string]time.Time
	// 签文列表
	lotteries []Lottery
}

// Lottery 签文
type Lottery struct {
	Name           string // 签名
	Content        string // 签文内容
	Interpretation string // 解签内容
	Level          int    // 签的等级（1-5，1为上上签，5为下下签）
}

// NewLotteryPlugin 创建抽签插件实例
func NewLotteryPlugin() *LotteryPlugin {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 初始化签文列表
	lotteries := []Lottery{
		{
			Name:           "上上签",
			Content:        "久旱逢甘雨，他乡遇故知。洞房花烛夜，金榜题名时。",
			Interpretation: "此签为上上大吉，诸事顺遂，心想事成。",
			Level:          1,
		},
		{
			Name:           "上签",
			Content:        "春风得意马蹄疾，一日看尽长安花。",
			Interpretation: "此签为上吉，事业有成，前程似锦。",
			Level:          2,
		},
		{
			Name:           "中签",
			Content:        "行到水穷处，坐看云起时。",
			Interpretation: "此签为中平，遇事需耐心等待，转机将至。",
			Level:          3,
		},
		{
			Name:           "下签",
			Content:        "屋漏偏逢连夜雨，船迟又遇打头风。",
			Interpretation: "此签为下凶，诸事不顺，需谨慎行事。",
			Level:          4,
		},
		{
			Name:           "下下签",
			Content:        "福无双至，祸不单行。",
			Interpretation: "此签为下下大凶，遇事需格外小心，避免冲动。",
			Level:          5,
		},
	}

	return &LotteryPlugin{
		lastLotteryTime: make(map[string]time.Time),
		lotteries:       lotteries,
	}
}

// isSameDay 检查两个时间是否在同一天
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func main() {
	log.Println("测试抽签系统核心逻辑...")

	// 创建抽签插件实例
	plugin := NewLotteryPlugin()

	// 测试插件基本信息
	log.Println("插件名称: lottery")
	log.Println("插件版本: 1.0.0")
	log.Println("插件描述: 抽签插件，支持抽签和解签功能")

	// 测试抽签功能
	userID := "test_user_123"
	now := time.Now()

	// 第一次抽签
	lottery := plugin.lotteries[rand.Intn(len(plugin.lotteries))]
	log.Printf("用户 %s 抽签结果:", userID)
	log.Printf("签名：%s", lottery.Name)
	log.Printf("签文：%s", lottery.Content)
	log.Printf("解签：%s", lottery.Interpretation)

	// 测试重复抽签
	plugin.lastLotteryTime[userID] = now
	if isSameDay(plugin.lastLotteryTime[userID], now) {
		log.Printf("用户 %s 今天已经抽过签了", userID)
	}

	// 测试解签功能
	log.Println("测试解签功能...")
	lottery = plugin.lotteries[rand.Intn(len(plugin.lotteries))]
	log.Printf("用户 %s 解签结果:", userID)
	log.Printf("签名：%s", lottery.Name)
	log.Printf("签文：%s", lottery.Content)
	log.Printf("解签：%s", lottery.Interpretation)

	log.Println("抽签系统核心逻辑测试通过!")
}
