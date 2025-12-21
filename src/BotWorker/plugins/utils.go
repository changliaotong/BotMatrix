package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// UtilsPlugin 工具插件
type UtilsPlugin struct {
	// 命令解析器
	cmdParser *CommandParser
}

func (p *UtilsPlugin) Name() string {
	return "utils"
}

func (p *UtilsPlugin) Description() string {
	return "工具插件，支持报时、计算、鬼故事、成语接龙、笑话等功能"
}

func (p *UtilsPlugin) Version() string {
	return "1.0.0"
}

// NewUtilsPlugin 创建工具插件实例
func NewUtilsPlugin() *UtilsPlugin {
	return &UtilsPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *UtilsPlugin) Init(robot plugin.Robot) {
	log.Println("加载工具插件")

	// 处理报时命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为报时命令
		if match, _ := p.cmdParser.MatchCommand("报时|time", event.RawMessage); !match {
			return nil
		}

		// 获取当前时间
		now := time.Now()
		timeMsg := fmt.Sprintf("🕐 当前时间：%s", now.Format("2006-01-02 15:04:05"))
		p.sendMessage(robot, event, timeMsg)

		return nil
	})

	// 处理计算命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为带参数的计算命令
		match, _, expr := p.cmdParser.MatchCommandWithSingleParam("计算|calc", event.RawMessage)
		if !match {
			// 检查是否为帮助请求（不带参数）
			if helpMatch, _ := p.cmdParser.MatchCommand("计算|calc", event.RawMessage); !helpMatch {
				return nil
			}
			// 发送帮助信息
			helpMsg := "计算命令格式：\n/计算 表达式\n/calc 表达式\n例如：/计算 1+2*3"
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 表达式已经被TrimSpace处理过

		// 简单计算（仅支持加减乘除）
		result, err := p.calculate(expr)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("计算失败：%v", err))
			return nil
		}

		// 发送结果
		resultMsg := fmt.Sprintf("%s = %.2f", expr, result)
		p.sendMessage(robot, event, resultMsg)

		return nil
	})

	// 处理笑话命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 使用命令解析器检查是否为笑话命令
		if match, _ := p.cmdParser.MatchCommand("笑话|joke", event.RawMessage); !match {
			return nil
		}

		// 随机选择笑话
		jokes := []string{
			"为什么程序员总是分不清万圣节和圣诞节？因为 Oct 31 = Dec 25！",
			"程序员的口头禅：这不可能啊！",
			"我问 Siri，‘你会说什么语言？’ Siri 回答：‘我会说多种语言，包括二进制。’",
			"为什么程序员喜欢用黑色背景？因为他们喜欢在黑暗中寻找光明！",
		}
		joke := jokes[rand.Intn(len(jokes))]

		// 发送笑话
		p.sendMessage(robot, event, joke)

		return nil
	})

	// 处理鬼故事命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 使用命令解析器检查是否为鬼故事命令
		if match, _ := p.cmdParser.MatchCommand("鬼故事|horror", event.RawMessage); !match {
			return nil
		}

		// 随机选择鬼故事
		stories := []string{
			"深夜，程序员在调试代码时，突然发现屏幕上出现了一行不属于自己的代码：// 我在看着你...",
			"小明在写代码时，突然听到身后传来键盘敲击声，回头却发现空无一人。",
			"程序员加班到凌晨，突然发现电脑屏幕上的光标自己在移动，输入了一行代码：// 该休息了...",
		}
		story := stories[rand.Intn(len(stories))]

		// 发送鬼故事
		p.sendMessage(robot, event, story)

		return nil
	})

	// 处理成语接龙命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为带参数的成语接龙命令
		match, _, idiom := p.cmdParser.MatchCommandWithSingleParam("成语接龙|idiom", event.RawMessage)
		if !match {
			// 检查是否为帮助请求（不带参数）
			if helpMatch, _ := p.cmdParser.MatchCommand("成语接龙|idiom", event.RawMessage); !helpMatch {
				return nil
			}
			// 发送帮助信息
			helpMsg := "成语接龙命令格式：\n/成语接龙 成语\n/idiom 成语\n例如：/成语接龙 一心一意"
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 成语已经被TrimSpace处理过

		// 随机选择接龙成语
		idioms := []string{
			"一心一意", "意气风发", "发扬光大", "大同小异", "异想天开",
			"开门见山", "山高水长", "长驱直入", "入木三分", "分秒必争",
		}
		response := idioms[rand.Intn(len(idioms))]

		// 发送结果
		resultMsg := fmt.Sprintf("你说：%s\n我说：%s", idiom, response)
		p.sendMessage(robot, event, resultMsg)

		return nil
	})
}

// calculate 简单计算
func (p *UtilsPlugin) calculate(expr string) (float64, error) {
	// 简单实现，仅支持加减乘除
	// 实际应用中应该使用更安全的表达式解析库
	// 这里仅做演示

	// 替换中文运算符
	expr = strings.ReplaceAll(expr, "加", "+")
	expr = strings.ReplaceAll(expr, "减", "-")
	expr = strings.ReplaceAll(expr, "乘", "*")
	expr = strings.ReplaceAll(expr, "除", "/")

	// 简单计算（仅支持两个操作数）
	// 实际应用中应该使用更复杂的解析
	// 这里仅做演示

	// 尝试解析加减乘除
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return 0, err
			}
			b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return 0, err
			}
			return a + b, nil
		}
	} else if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return 0, err
			}
			b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return 0, err
			}
			return a - b, nil
		}
	} else if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return 0, err
			}
			b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return 0, err
			}
			return a * b, nil
		}
	} else if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return 0, err
			}
			b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return 0, err
			}
			if b == 0 {
				return 0, fmt.Errorf("除数不能为零")
			}
			return a / b, nil
		}
	}

	return 0, fmt.Errorf("不支持的表达式格式")
}

// sendMessage 发送消息
func (p *UtilsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
