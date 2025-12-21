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

// GamesPlugin 游戏插件
type GamesPlugin struct{}

func (p *GamesPlugin) Name() string {
	return "games"
}

func (p *GamesPlugin) Description() string {
	return "游戏插件，支持猜拳、三公、梭哈、猜大小等游戏"
}

func (p *GamesPlugin) Version() string {
	return "1.0.0"
}

// NewGamesPlugin 创建游戏插件实例
func NewGamesPlugin() *GamesPlugin {
	return &GamesPlugin{}
}

func (p *GamesPlugin) Init(robot plugin.Robot) {
	log.Println("加载游戏插件")

	// 处理猜拳命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为猜拳命令
		msg := strings.TrimSpace(event.RawMessage)
		if !strings.HasPrefix(msg, "!猜拳 ") && !strings.HasPrefix(msg, "!rock ") {
			return nil
		}

		// 解析玩家选择
		var playerChoice string
		if strings.HasPrefix(msg, "!猜拳 ") {
			playerChoice = strings.TrimSpace(msg[3:])
		} else {
			playerChoice = strings.TrimSpace(msg[6:])
		}

		// 验证玩家选择
		validChoices := map[string]bool{"石头": true, "剪刀": true, "布": true, "rock": true, "paper": true, "scissors": true}
		if !validChoices[playerChoice] {
			p.sendMessage(robot, event, "无效选择，请选择石头、剪刀、布或rock、paper、scissors")
			return nil
		}

		// 机器人随机选择
		choices := []string{"石头", "剪刀", "布"}
		botChoice := choices[rand.Intn(len(choices))]

		// 判断胜负
		result := p.judgeRockPaperScissors(playerChoice, botChoice)

		// 发送结果
		resultMsg := fmt.Sprintf("你出了：%s\n机器人出了：%s\n结果：%s", playerChoice, botChoice, result)
		p.sendMessage(robot, event, resultMsg)

		return nil
	})

	// 处理猜大小命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为猜大小命令
		msg := strings.TrimSpace(event.RawMessage)
		if !strings.HasPrefix(msg, "!猜大小 ") && !strings.HasPrefix(msg, "!bigsmall ") {
			return nil
		}

		// 解析玩家选择
		var playerChoice string
		if strings.HasPrefix(msg, "!猜大小 ") {
			playerChoice = strings.TrimSpace(msg[4:])
		} else {
			playerChoice = strings.TrimSpace(msg[9:])
		}

		// 验证玩家选择
		validChoices := map[string]bool{"大": true, "小": true, "big": true, "small": true}
		if !validChoices[playerChoice] {
			p.sendMessage(robot, event, "无效选择，请选择大、小或big、small")
			return nil
		}

		// 生成随机数（1-100）
		num := rand.Intn(100) + 1
		actualResult := "大"
		if num <= 50 {
			actualResult = "小"
		}

		// 判断胜负
		result := "平局"
		if (playerChoice == "大" || playerChoice == "big") && actualResult == "大" {
			result = "你赢了！"
		} else if (playerChoice == "小" || playerChoice == "small") && actualResult == "小" {
			result = "你赢了！"
		} else {
			result = "你输了！"
		}

		// 发送结果
		resultMsg := fmt.Sprintf("你猜了：%s\n随机数：%d\n结果：%s", playerChoice, num, result)
		p.sendMessage(robot, event, resultMsg)

		return nil
	})

	// 处理抽奖命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为抽奖命令
		msg := strings.TrimSpace(event.RawMessage)
		if msg != "!抽奖" && msg != "!lottery" {
			return nil
		}

		// 生成随机奖品
		prizes := []string{"一等奖：100积分", "二等奖：50积分", "三等奖：10积分", "谢谢参与"}
		prize := prizes[rand.Intn(len(prizes))]

		// 发送结果
		resultMsg := fmt.Sprintf("🎁 抽奖结果：%s", prize)
		p.sendMessage(robot, event, resultMsg)

		return nil
	})
}

// judgeRockPaperScissors 判断猜拳胜负
func (p *GamesPlugin) judgeRockPaperScissors(player, bot string) string {
	// 统一转换为中文
	if player == "rock" {
		player = "石头"
	} else if player == "paper" {
		player = "布"
	} else if player == "scissors" {
		player = "剪刀"
	}

	if player == bot {
		return "平局！"
	}

	if (player == "石头" && bot == "剪刀") || (player == "剪刀" && bot == "布") || (player == "布" && bot == "石头") {
		return "你赢了！"
	}

	return "你输了！"
}

// sendMessage 发送消息
func (p *GamesPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}