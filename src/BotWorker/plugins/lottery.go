package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
	"strings"
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
	Name        string // 签名
	Content     string // 签文内容
	Interpretation string // 解签内容
	Level       int    // 签的等级（1-5，1为上上签，5为下下签）
}

// NewLotteryPlugin 创建抽签插件实例
func NewLotteryPlugin() *LotteryPlugin {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 初始化签文列表
	lotteries := []Lottery{
		{
			Name:        "上上签",
			Content:     "久旱逢甘雨，他乡遇故知。洞房花烛夜，金榜题名时。",
			Interpretation: "此签为上上大吉，诸事顺遂，心想事成。",
			Level:       1,
		},
		{
			Name:        "上签",
			Content:     "春风得意马蹄疾，一日看尽长安花。",
			Interpretation: "此签为上吉，事业有成，前程似锦。",
			Level:       2,
		},
		{
			Name:        "中签",
			Content:     "行到水穷处，坐看云起时。",
			Interpretation: "此签为中平，遇事需耐心等待，转机将至。",
			Level:       3,
		},
		{
			Name:        "下签",
			Content:     "屋漏偏逢连夜雨，船迟又遇打头风。",
			Interpretation: "此签为下凶，诸事不顺，需谨慎行事。",
			Level:       4,
		},
		{
			Name:        "下下签",
			Content:     "福无双至，祸不单行。",
			Interpretation: "此签为下下大凶，遇事需格外小心，避免冲动。",
			Level:       5,
		},
	}

	return &LotteryPlugin{
		lastLotteryTime: make(map[string]time.Time),
		lotteries:       lotteries,
	}
}

func (p *LotteryPlugin) Name() string {
	return "lottery"
}

func (p *LotteryPlugin) Description() string {
	return "抽签插件，支持抽签和解签功能"
}

func (p *LotteryPlugin) Version() string {
	return "1.0.0"
}

func (p *LotteryPlugin) Init(robot plugin.Robot) {
	log.Println("加载抽签插件")

	// 处理抽签命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为抽签命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!lottery" && msg != "!抽签" {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，抽签失败")
			return nil
		}

		// 检查是否已经抽过签（每天限抽一次）
		now := time.Now()
		if lastLottery, ok := p.lastLotteryTime[userID]; ok {
			// 检查是否在同一天
			if isSameDay(lastLottery, now) {
				p.sendMessage(robot, event, fmt.Sprintf("你今天已经抽过签了！上次抽签时间：%s", lastLottery.Format("15:04:05")))
				return nil
			}
		}

		// 随机抽取一个签
		lottery := p.lotteries[rand.Intn(len(p.lotteries))]

		// 更新抽签记录
		p.lastLotteryTime[userID] = now

		// 发送抽签结果
		msg = fmt.Sprintf("🎐 抽签结果 🎐\n")
		msg += fmt.Sprintf("签名：%s\n", lottery.Name)
		msg += fmt.Sprintf("签文：%s\n", lottery.Content)
		msg += fmt.Sprintf("解签：%s\n", lottery.Interpretation)

		p.sendMessage(robot, event, msg)

		return nil
	})

	// 处理解签命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为解签命令
		msg := strings.TrimSpace(event.RawMessage)
		if !strings.HasPrefix(msg, "!interpret") && !strings.HasPrefix(msg, "!解签") {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，解签失败")
			return nil
		}

		// 检查是否有抽签记录
		if _, ok := p.lastLotteryTime[userID]; !ok {
			p.sendMessage(robot, event, "你还没有抽过签，请先抽签！")
			return nil
		}

		// 重新抽取上次的签（模拟解签）
		lottery := p.lotteries[rand.Intn(len(p.lotteries))]

		// 发送解签结果
		msg = fmt.Sprintf("📜 解签结果 📜\n")
		msg += fmt.Sprintf("签名：%s\n", lottery.Name)
		msg += fmt.Sprintf("签文：%s\n", lottery.Content)
		msg += fmt.Sprintf("解签：%s\n", lottery.Interpretation)

		p.sendMessage(robot, event, msg)

		return nil
	})
}

// sendMessage 发送消息
func (p *LotteryPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
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