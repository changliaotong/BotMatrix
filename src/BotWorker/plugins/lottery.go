package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
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
	// 命令解析器
	cmdParser *CommandParser
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
			Name:           common.T("", "lottery_level1_name"),
			Content:        common.T("", "lottery_level1_content"),
			Interpretation: common.T("", "lottery_level1_interpretation"),
			Level:          1,
		},
		{
			Name:           common.T("", "lottery_level2_name"),
			Content:        common.T("", "lottery_level2_content"),
			Interpretation: common.T("", "lottery_level2_interpretation"),
			Level:          2,
		},
		{
			Name:           common.T("", "lottery_level3_name"),
			Content:        common.T("", "lottery_level3_content"),
			Interpretation: common.T("", "lottery_level3_interpretation"),
			Level:          3,
		},
		{
			Name:           common.T("", "lottery_level4_name"),
			Content:        common.T("", "lottery_level4_content"),
			Interpretation: common.T("", "lottery_level4_interpretation"),
			Level:          4,
		},
		{
			Name:           common.T("", "lottery_level5_name"),
			Content:        common.T("", "lottery_level5_content"),
			Interpretation: common.T("", "lottery_level5_interpretation"),
			Level:          5,
		},
	}

	return &LotteryPlugin{
		lastLotteryTime: make(map[string]time.Time),
		lotteries:       lotteries,
		cmdParser:       NewCommandParser(),
	}
}

func (p *LotteryPlugin) Name() string {
	return "lottery"
}

func (p *LotteryPlugin) Description() string {
	return common.T("", "lottery_plugin_desc")
}

func (p *LotteryPlugin) Version() string {
	return "1.0.0"
}

func (p *LotteryPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "lottery_plugin_loaded"))

	// 处理抽签命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "lottery") {
				HandleFeatureDisabled(robot, event, "lottery")
				return nil
			}
		}

		// 检查是否为抽签命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "lottery_cmd_draw"), event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "lottery_invalid_userid"))
			return nil
		}

		// 检查是否已经抽过签（每天限抽一次）
		now := time.Now()
		if lastLottery, ok := p.lastLotteryTime[fmt.Sprintf("%d", userID)]; ok {
			// 检查是否在同一天
			if isSameDay(lastLottery, now) {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "lottery_already_drawn"), lastLottery.Format("15:04:05")))
				return nil
			}
		}

		// 随机抽取一个签
		lottery := p.lotteries[rand.Intn(len(p.lotteries))]

		// 更新抽签记录
		p.lastLotteryTime[fmt.Sprintf("%d", userID)] = now

		// 发送抽签结果
		msg := common.T("", "lottery_result_header")
		msg += fmt.Sprintf(common.T("", "lottery_result_name"), lottery.Name)
		msg += fmt.Sprintf(common.T("", "lottery_result_content"), lottery.Content)
		msg += fmt.Sprintf(common.T("", "lottery_result_interpretation"), lottery.Interpretation)

		p.sendMessage(robot, event, msg)

		return nil
	})

	// 处理解签命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "lottery") {
				HandleFeatureDisabled(robot, event, "lottery")
				return nil
			}
		}

		// 检查是否为解签命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "lottery_cmd_interpret"), event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "lottery_invalid_userid"))
			return nil
		}

		// 检查是否有抽签记录
		if _, ok := p.lastLotteryTime[fmt.Sprintf("%d", userID)]; !ok {
			p.sendMessage(robot, event, common.T("", "lottery_not_drawn"))
			return nil
		}

		// 重新抽取上次的签（模拟解签）
		lottery := p.lotteries[rand.Intn(len(p.lotteries))]

		// 发送解签结果
		msg := common.T("", "lottery_interpret_header")
		msg += fmt.Sprintf(common.T("", "lottery_result_name"), lottery.Name)
		msg += fmt.Sprintf(common.T("", "lottery_result_content"), lottery.Content)
		msg += fmt.Sprintf(common.T("", "lottery_result_interpretation"), lottery.Interpretation)

		p.sendMessage(robot, event, msg)

		return nil
	})
}

// sendMessage 发送消息
func (p *LotteryPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "lottery_send_failed"), err)
	}
}
