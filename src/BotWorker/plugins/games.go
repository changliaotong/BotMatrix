package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
)

type GamesPlugin struct {
	cmdParser   *CommandParser
	idiomGames  map[string]*IdiomGameState
	idioms      []string
}

type IdiomGameState struct {
	CurrentIdiom string
}

func (p *GamesPlugin) Name() string {
	return "games"
}

func (p *GamesPlugin) Description() string {
	return "游戏插件，支持猜拳、三公、梭哈、猜大小等游戏"
}

func (p *GamesPlugin) Version() string {
	return "1.0.0"
}

func NewGamesPlugin() *GamesPlugin {
	return &GamesPlugin{
		cmdParser:  NewCommandParser(),
		idiomGames: make(map[string]*IdiomGameState),
		idioms: []string{
			"画蛇添足",
			"足智多谋",
			"谋事在人",
			"人山人海",
			"海阔天空",
			"空前绝后",
			"后来居上",
			"上行下效",
			"效颦学步",
			"步步高升",
			"升堂入室",
			"室雅人和",
			"和气致祥",
			"祥风时雨",
			"雨过天晴",
			"晴空万里",
			"里应外合",
			"合情合理",
			"理直气壮",
			"壮志凌云",
		},
	}
}

func (p *GamesPlugin) Init(robot plugin.Robot) {
	log.Println("加载游戏插件")

	// 处理猜拳命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "games") {
				HandleFeatureDisabled(robot, event, "games")
				return nil
			}
		}

		// 检查是否为猜拳命令
		match, _, playerChoice := p.cmdParser.MatchCommandWithSingleParam("猜拳|rock", event.RawMessage)
		if !match {
			return nil
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

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "games") {
				HandleFeatureDisabled(robot, event, "games")
				return nil
			}
		}

		// 检查是否为猜大小命令
		match, _, playerChoice := p.cmdParser.MatchCommandWithSingleParam("猜大小|bigsmall", event.RawMessage)
		if !match {
			return nil
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

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "games") {
				HandleFeatureDisabled(robot, event, "games")
				return nil
			}
		}

		// 检查是否为抽奖命令
		if match, _ := p.cmdParser.MatchCommand("抽奖|lottery", event.RawMessage); !match {
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

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "games") {
				HandleFeatureDisabled(robot, event, "games")
				return nil
			}
		}

		matchContinue, _, idiom := p.cmdParser.MatchCommandWithSingleParam("成语接龙|idiom", event.RawMessage)
		if matchContinue && idiom != "" {
			return p.handleIdiomContinue(robot, event, idiom)
		}

		if matchStart, _ := p.cmdParser.MatchCommand("成语接龙|idiom", event.RawMessage); matchStart {
			return p.handleIdiomStart(robot, event)
		}

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

func (p *GamesPlugin) getIdiomGameKey(event *onebot.Event) string {
	if event.MessageType == "group" {
		return fmt.Sprintf("group:%d", event.GroupID)
	}
	return fmt.Sprintf("user:%d", event.UserID)
}

func (p *GamesPlugin) handleIdiomStart(robot plugin.Robot, event *onebot.Event) error {
	if len(p.idioms) == 0 {
		p.sendMessage(robot, event, "成语库为空，暂时无法开始成语接龙")
		return nil
	}

	key := p.getIdiomGameKey(event)
	start := p.idioms[rand.Intn(len(p.idioms))]
	p.idiomGames[key] = &IdiomGameState{CurrentIdiom: start}

	runes := []rune(start)
	last := ""
	if len(runes) > 0 {
		last = string(runes[len(runes)-1])
	}

	msg := fmt.Sprintf("成语接龙开始！第一个成语：%s\n请接下一个成语，要求首字为「%s」", start, last)
	p.sendMessage(robot, event, msg)
	return nil
}

func (p *GamesPlugin) handleIdiomContinue(robot plugin.Robot, event *onebot.Event, idiom string) error {
	key := p.getIdiomGameKey(event)
	state, ok := p.idiomGames[key]
	if !ok || state.CurrentIdiom == "" {
		p.sendMessage(robot, event, "你还没有开始成语接龙，请先发送「/ 成语接龙」")
		return nil
	}

	idiomRunes := []rune(idiom)
	if len(idiomRunes) < 2 {
		p.sendMessage(robot, event, "请输入正确的成语")
		return nil
	}

	prevRunes := []rune(state.CurrentIdiom)
	if len(prevRunes) == 0 {
		state.CurrentIdiom = idiom
	} else {
		last := prevRunes[len(prevRunes)-1]
		first := idiomRunes[0]
		if last != first {
			p.sendMessage(robot, event, fmt.Sprintf("不对哦，新成语必须以「%c」开头", last))
			return nil
		}
		state.CurrentIdiom = idiom
	}

	botIdiom, ok := p.findNextIdiom(idiom)
	if !ok {
		delete(p.idiomGames, key)
		p.sendMessage(robot, event, fmt.Sprintf("你接得很好：%s\n我一时想不出下一个了，这局你赢了！", idiom))
		return nil
	}

	state.CurrentIdiom = botIdiom
	nextRunes := []rune(botIdiom)
	nextLast := ' '
	if len(nextRunes) > 0 {
		nextLast = nextRunes[len(nextRunes)-1]
	}

	msg := fmt.Sprintf("你接了：%s\n我接：%s\n继续，请接首字为「%c」的成语", idiom, botIdiom, nextLast)
	p.sendMessage(robot, event, msg)
	return nil
}

func (p *GamesPlugin) findNextIdiom(prev string) (string, bool) {
	runes := []rune(prev)
	if len(runes) == 0 {
		return "", false
	}
	last := runes[len(runes)-1]

	candidates := make([]string, 0)
	for _, item := range p.idioms {
		ir := []rune(item)
		if len(ir) == 0 {
			continue
		}
		if ir[0] == last && item != prev {
			candidates = append(candidates, item)
		}
	}

	if len(candidates) == 0 {
		return "", false
	}

	return candidates[rand.Intn(len(candidates))], true
}

// sendMessage 发送消息
func (p *GamesPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
