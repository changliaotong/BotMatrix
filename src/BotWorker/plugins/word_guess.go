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

// WordGuessPlugin 猜单词插件
type WordGuessPlugin struct {
	cmdParser *CommandParser
	// 存储当前正在进行的游戏，key为用户ID，value为游戏数据
	games map[string]*WordGame
	// 单词列表
	wordList []string
}

// WordGame 游戏数据结构
type WordGame struct {
	UserID     string
	Word       string
	Hint       string
	Guessed    string
	Attempts   int
	MaxAttempts int
	StartTime  time.Time
}

// NewWordGuessPlugin 创建猜单词插件实例
func NewWordGuessPlugin() *WordGuessPlugin {
	rand.Seed(time.Now().UnixNano())
	plugin := &WordGuessPlugin{
		cmdParser: NewCommandParser(),
		games:     make(map[string]*WordGame),
		wordList:  []string{},
	}
	plugin.initWordList()
	return plugin
}

func (p *WordGuessPlugin) Name() string {
	return "word_guess"
}

func (p *WordGuessPlugin) Description() string {
	return "猜单词游戏，可以随机选择单词让用户猜测"
}

func (p *WordGuessPlugin) Version() string {
	return "1.0.0"
}

// initWordList 初始化单词列表
func (p *WordGuessPlugin) initWordList() {
	// 初始化简单的英语单词列表
	p.wordList = []string{
		"apple", "banana", "orange", "grape", "melon",
		"book", "pencil", "pen", "paper", "ruler",
		"cat", "dog", "bird", "fish", "rabbit",
		"car", "bus", "train", "plane", "bike",
		"house", "school", "park", "shop", "hospital",
		"sun", "moon", "star", "sky", "cloud",
		"red", "blue", "green", "yellow", "black",
		"happy", "sad", "angry", "excited", "tired",
		"run", "walk", "jump", "swim", "fly",
		"big", "small", "long", "short", "tall",
	}
}

// Init 初始化插件
func (p *WordGuessPlugin) Init(robot plugin.Robot) {
	log.Println("加载猜单词插件")

	// 处理消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查功能是否启用
		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "word_guess") {
				HandleFeatureDisabled(robot, event, "word_guess")
				return nil
			}
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)

		// 检查是否为开始猜单词命令
		if match, _ := p.cmdParser.MatchCommand("猜单词|开始猜单词", event.RawMessage); match {
			p.startNewGame(robot, event, userIDStr)
			return nil
		}

		// 检查是否为提交答案命令
		match, _, params := p.cmdParser.MatchCommandWithParams("提交", "(.+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, "提交命令格式：提交 <答案>")
				return nil
			}
			answer := strings.TrimSpace(params[0])
			p.submitAnswer(robot, event, userIDStr, answer)
			return nil
		}

		// 检查是否为查看当前游戏状态命令
		if match, _ := p.cmdParser.MatchCommand("查看游戏|游戏状态", event.RawMessage); match {
			p.showGameStatus(robot, event, userIDStr)
			return nil
		}

		// 检查是否为放弃游戏命令
		if match, _ := p.cmdParser.MatchCommand("放弃游戏|结束游戏", event.RawMessage); match {
			p.giveUpGame(robot, event, userIDStr)
			return nil
		}

		return nil
	})
}

// startNewGame 开始新游戏
func (p *WordGuessPlugin) startNewGame(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否已有正在进行的游戏
	if _, exists := p.games[userID]; exists {
		p.sendMessage(robot, event, "您已经有一个正在进行的猜单词游戏，请先完成当前游戏或放弃游戏")
		return
	}

	// 随机选择一个单词
	word := p.wordList[rand.Intn(len(p.wordList))]
	
	// 生成提示
	hint := p.generateHint(word)
	
	// 生成已猜字母字符串（初始全为下划线）
	guessed := strings.Repeat("_", len(word))

	// 创建新游戏
	game := &WordGame{
		UserID:      userID,
		Word:        word,
		Hint:        hint,
		Guessed:     guessed,
		Attempts:    0,
		MaxAttempts: 6,
		StartTime:   time.Now(),
	}

	p.games[userID] = game

	// 发送游戏开始消息
	p.sendMessage(robot, event, fmt.Sprintf(
		"🎮 猜单词游戏开始！\n"+
		"提示：%s\n"+
		"单词：%s\n"+
		"剩余次数：%d\n"+
		"输入 '提交 <答案>' 来猜测",
		game.Hint, game.Guessed, game.MaxAttempts
	))
}

// generateHint 生成单词提示
func (p *WordGuessPlugin) generateHint(word string) string {
	hints := map[string]string{
		"apple": "一种红色或绿色的水果",
		"banana": "一种黄色的弯曲水果",
		"orange": "一种橙色的水果",
		"grape": "一种紫色或绿色的小水果",
		"melon": "一种大型的瓜类水果",
		"book": "用来阅读的物品",
		"pencil": "用来写字的文具",
		"pen": "用来写字的工具",
		"paper": "用来书写的纸张",
		"ruler": "用来测量的工具",
		"cat": "一种小型的宠物",
		"dog": "一种忠诚的宠物",
		"bird": "一种会飞的动物",
		"fish": "一种生活在水中的动物",
		"rabbit": "一种长耳朵的动物",
		"car": "一种交通工具",
		"bus": "一种公共交通工具",
		"train": "一种在轨道上行驶的交通工具",
		"plane": "一种在天空中飞行的交通工具",
		"bike": "一种两轮的交通工具",
		"house": "人们居住的地方",
		"school": "学习的地方",
		"park": "休闲的地方",
		"shop": "购物的地方",
		"hospital": "看病的地方",
		"sun": "白天发光的天体",
		"moon": "晚上发光的天体",
		"star": "天空中的星星",
		"sky": "天空",
		"cloud": "天空中的云朵",
		"red": "一种颜色",
		"blue": "一种颜色",
		"green": "一种颜色",
		"yellow": "一种颜色",
		"black": "一种颜色",
		"happy": "一种情绪",
		"sad": "一种情绪",
		"angry": "一种情绪",
		"excited": "一种情绪",
		"tired": "一种情绪",
		"run": "一种运动",
		"walk": "一种运动",
		"jump": "一种运动",
		"swim": "一种运动",
		"fly": "一种运动",
		"big": "一种形容词",
		"small": "一种形容词",
		"long": "一种形容词",
		"short": "一种形容词",
		"tall": "一种形容词",
	}

	if hint, ok := hints[word]; ok {
		return hint
	}
	return "未知提示"
}

// submitAnswer 提交答案
func (p *WordGuessPlugin) submitAnswer(robot plugin.Robot, event *onebot.Event, userID string, answer string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, "您还没有开始猜单词游戏，请先输入 '猜单词' 开始游戏")
		return
	}

	// 增加尝试次数
	game.Attempts++

	// 检查答案是否正确
	if strings.EqualFold(answer, game.Word) {
		// 猜对了
		duration := time.Since(game.StartTime)
		p.sendMessage(robot, event, fmt.Sprintf(
			"🎉 恭喜您猜对了！\n"+
			"单词：%s\n"+
			"用时：%v\n"+
			"尝试次数：%d/%d",
			game.Word, duration, game.Attempts, game.MaxAttempts
		))
		// 删除游戏
		delete(p.games, userID)
		return
	}

	// 检查是否还有剩余次数
	remaining := game.MaxAttempts - game.Attempts
	if remaining <= 0 {
		// 游戏结束
		p.sendMessage(robot, event, fmt.Sprintf(
			"😔 游戏结束，您没有猜对！\n"+
			"正确答案：%s",
			game.Word
		))
		// 删除游戏
		delete(p.games, userID)
		return
	}

	// 显示当前状态
	p.sendMessage(robot, event, fmt.Sprintf(
		"❌ 猜测错误！\n"+
		"单词：%s\n"+
		"剩余次数：%d\n"+
		"请继续猜测",
		game.Guessed, remaining
	))
}

// showGameStatus 显示当前游戏状态
func (p *WordGuessPlugin) showGameStatus(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, "您还没有开始猜单词游戏，请先输入 '猜单词' 开始游戏")
		return
	}

	remaining := game.MaxAttempts - game.Attempts
	duration := time.Since(game.StartTime)

	p.sendMessage(robot, event, fmt.Sprintf(
		"🎮 当前游戏状态\n"+
		"提示：%s\n"+
		"单词：%s\n"+
		"尝试次数：%d/%d\n"+
		"剩余次数：%d\n"+
		"游戏时长：%v",
		game.Hint, game.Guessed, game.Attempts, game.MaxAttempts, remaining, duration
	))
}

// giveUpGame 放弃游戏
func (p *WordGuessPlugin) giveUpGame(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, "您还没有开始猜单词游戏，请先输入 '猜单词' 开始游戏")
		return
	}

	// 显示放弃消息
	p.sendMessage(robot, event, fmt.Sprintf(
		"😔 您放弃了游戏！\n"+
		"正确答案：%s",
		game.Word
	))

	// 删除游戏
	delete(p.games, userID)
}

// sendMessage 发送消息
func (p *WordGuessPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		MessageType: event.MessageType,
		UserID:      event.UserID,
		GroupID:     event.GroupID,
		Message:     message,
	}
	_, err := robot.SendMessage(params)
	if err != nil {
		log.Printf("发送消息失败: %v", err)
	}
}
