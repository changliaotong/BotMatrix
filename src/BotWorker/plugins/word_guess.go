package plugins

import (
	"BotMatrix/common"
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
	UserID      string
	Word        string
	Hint        string
	Guessed     string
	Attempts    int
	MaxAttempts int
	StartTime   time.Time
}

// NewWordGuessPlugin 创建猜单词插件实例
func NewWordGuessPlugin() *WordGuessPlugin {
	p := &WordGuessPlugin{
		cmdParser: NewCommandParser(),
		games:     make(map[string]*WordGame),
		wordList:  []string{},
	}
	p.initWordList()
	return p
}

func (p *WordGuessPlugin) Name() string {
	return "word_guess"
}

func (p *WordGuessPlugin) Description() string {
	return common.T("", "word_guess_plugin_desc")
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
	log.Println(common.T("", "word_guess_plugin_loaded"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

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

		// 检查是否为开始游戏命令
		if match, _ := p.cmdParser.MatchCommand("单词猜谜|猜单词|wordguess", event.RawMessage); match {
			_, err := p.handleStartGameLogic(robot, event, userIDStr)
			return err
		}

		// 检查是否为提交答案命令
		match, _, params := p.cmdParser.MatchCommandWithParams("猜单词提交|提交单词", "(.+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, common.T("", "idiom_guess_enter_answer"))
				return nil
			}
			_, err := p.handleSubmitAnswerLogic(robot, event, userIDStr, params[0])
			return err
		}

		// 检查是否为查看当前游戏状态命令
		if match, _ := p.cmdParser.MatchCommand("查看游戏|游戏状态", event.RawMessage); match {
			_, err := p.handleShowStatusLogic(robot, event, userIDStr)
			return err
		}

		// 检查是否为放弃游戏命令
		if match, _ := p.cmdParser.MatchCommand("放弃游戏|结束游戏", event.RawMessage); match {
			_, err := p.handleGiveUpLogic(robot, event, userIDStr)
			return err
		}

		return nil
	})
}

// GetSkills 实现 SkillCapable 接口
func (p *WordGuessPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "start",
			Description: common.T("", "word_guess_skill_start_desc"),
			Usage:       "start",
		},
		{
			Name:        "submit",
			Description: common.T("", "word_guess_skill_submit_desc"),
			Usage:       "submit <answer>",
			Params: map[string]string{
				"answer": common.T("", "word_guess_skill_submit_param_answer"),
			},
		},
		{
			Name:        "status",
			Description: common.T("", "word_guess_skill_status_desc"),
			Usage:       "status",
		},
		{
			Name:        "giveup",
			Description: common.T("", "word_guess_skill_giveup_desc"),
			Usage:       "giveup",
		},
	}
}

// HandleSkill 处理技能调用
func (p *WordGuessPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userIDStr := fmt.Sprintf("%d", event.UserID)
	switch skillName {
	case "start":
		return p.handleStartGameLogic(robot, event, userIDStr)
	case "submit":
		answer := params["answer"]
		if answer == "" {
			msg := common.T("", "idiom_guess_enter_answer")
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		return p.handleSubmitAnswerLogic(robot, event, userIDStr, answer)
	case "status":
		return p.handleShowStatusLogic(robot, event, userIDStr)
	case "giveup":
		return p.handleGiveUpLogic(robot, event, userIDStr)
	}
	return "", nil
}

func (p *WordGuessPlugin) handleStartGameLogic(robot plugin.Robot, event *onebot.Event, userID string) (string, error) {
	// 检查是否已有正在进行的游戏
	if _, exists := p.games[userID]; exists {
		msg := common.T("", "word_guess_already_started")
		p.sendMessage(robot, event, msg)
		return msg, nil
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
	msg := fmt.Sprintf(
		common.T("", "word_guess_start_msg"),
		game.Hint, game.Guessed, game.MaxAttempts,
	)
	p.sendMessage(robot, event, msg)
	return msg, nil
}

func (p *WordGuessPlugin) handleSubmitAnswerLogic(robot plugin.Robot, event *onebot.Event, userID string, answer string) (string, error) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		msg := common.T("", "word_guess_no_game")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	// 增加尝试次数
	game.Attempts++

	// 检查答案是否正确
	if strings.EqualFold(answer, game.Word) {
		// 猜对了
		duration := time.Since(game.StartTime)
		msg := fmt.Sprintf(
			common.T("", "word_guess_correct"),
			game.Word, duration, game.Attempts, game.MaxAttempts,
		)
		p.sendMessage(robot, event, msg)
		// 删除游戏
		delete(p.games, userID)
		return msg, nil
	}

	// 检查是否还有剩余次数
	remaining := game.MaxAttempts - game.Attempts
	if remaining <= 0 {
		// 游戏结束
		msg := fmt.Sprintf(
			common.T("", "word_guess_game_over"),
			game.Word, game.Attempts,
		)
		p.sendMessage(robot, event, msg)
		// 删除游戏
		delete(p.games, userID)
		return msg, nil
	}

	// 继续游戏，给出反馈
	msg := fmt.Sprintf(
		common.T("", "word_guess_incorrect"),
		answer, remaining,
	)
	p.sendMessage(robot, event, msg)
	return msg, nil
}

func (p *WordGuessPlugin) handleShowStatusLogic(robot plugin.Robot, event *onebot.Event, userID string) (string, error) {
	game, exists := p.games[userID]
	if !exists {
		msg := common.T("", "word_guess_no_game")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	msg := fmt.Sprintf(
		common.T("", "word_guess_status_msg"),
		game.Hint, game.Guessed, game.Attempts, game.MaxAttempts,
	)
	p.sendMessage(robot, event, msg)
	return msg, nil
}

func (p *WordGuessPlugin) handleGiveUpLogic(robot plugin.Robot, event *onebot.Event, userID string) (string, error) {
	game, exists := p.games[userID]
	if !exists {
		msg := common.T("", "word_guess_no_game")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	msg := fmt.Sprintf(
		common.T("", "word_guess_give_up"),
		game.Word,
	)
	p.sendMessage(robot, event, msg)
	delete(p.games, userID)
	return msg, nil
}

// generateHint 生成单词提示
func (p *WordGuessPlugin) generateHint(word string) string {
	key := "word_guess_hint_" + word
	hint := common.T("", key)
	if hint == key {
		return common.T("", "word_guess_hint_unknown")
	}
	return hint
}

// sendMessage 发送消息
func (p *WordGuessPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
