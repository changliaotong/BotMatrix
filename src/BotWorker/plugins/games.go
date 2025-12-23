package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
)

type GamesPlugin struct {
	cmdParser  *CommandParser
	idiomGames map[string]*IdiomGameState
	idioms     []string
}

type IdiomGameState struct {
	CurrentIdiom string
}

func (p *GamesPlugin) Name() string {
	return "games"
}

func (p *GamesPlugin) Description() string {
	return common.T("", "games_plugin_desc")
}

func (p *GamesPlugin) Version() string {
	return "1.0.0"
}

func NewGamesPlugin() *GamesPlugin {
	return &GamesPlugin{
		cmdParser:  NewCommandParser(),
		idiomGames: make(map[string]*IdiomGameState),
		idioms: []string{
			common.T("", "idiom_1"),
			common.T("", "idiom_2"),
			common.T("", "idiom_3"),
			common.T("", "idiom_4"),
			common.T("", "idiom_5"),
			common.T("", "idiom_6"),
			common.T("", "idiom_7"),
			common.T("", "idiom_8"),
			common.T("", "idiom_9"),
			common.T("", "idiom_10"),
			common.T("", "idiom_11"),
			common.T("", "idiom_12"),
			common.T("", "idiom_13"),
			common.T("", "idiom_14"),
			common.T("", "idiom_15"),
			common.T("", "idiom_16"),
			common.T("", "idiom_17"),
			common.T("", "idiom_18"),
			common.T("", "idiom_19"),
			common.T("", "idiom_20"),
		},
	}
}

func (p *GamesPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "games_plugin_loaded"))

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
		match, _, playerChoice := p.cmdParser.MatchCommandWithSingleParam(common.T("", "games_cmd_rock"), event.RawMessage)
		if !match {
			return nil
		}

		// 验证玩家选择
		validChoices := map[string]bool{
			common.T("", "games_rock"):     true,
			common.T("", "games_paper"):    true,
			common.T("", "games_scissors"): true,
			"rock":                         true,
			"paper":                        true,
			"scissors":                     true,
		}
		if !validChoices[playerChoice] {
			p.sendMessage(robot, event, common.T("", "games_rock_invalid"))
			return nil
		}

		// 机器人随机选择
		choices := []string{
			common.T("", "games_rock"),
			common.T("", "games_scissors"),
			common.T("", "games_paper"),
		}
		botChoice := choices[rand.Intn(len(choices))]

		// 判断胜负
		result := p.judgeRockPaperScissors(playerChoice, botChoice)

		// 发送结果
		resultMsg := fmt.Sprintf(common.T("", "games_rock_result"), playerChoice, botChoice, result)
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
		match, _, playerChoice := p.cmdParser.MatchCommandWithSingleParam(common.T("", "games_cmd_bigsmall"), event.RawMessage)
		if !match {
			return nil
		}

		// 验证玩家选择
		validChoices := map[string]bool{
			common.T("", "games_big"):   true,
			common.T("", "games_small"): true,
			"big":                       true,
			"small":                     true,
		}
		if !validChoices[playerChoice] {
			p.sendMessage(robot, event, common.T("", "games_bigsmall_invalid"))
			return nil
		}

		// 生成随机数（1-100）
		num := rand.Intn(100) + 1
		actualResult := common.T("", "games_big")
		if num <= 50 {
			actualResult = common.T("", "games_small")
		}

		// 判断胜负
		result := common.T("", "games_draw")
		if (playerChoice == common.T("", "games_big") || playerChoice == "big") && actualResult == common.T("", "games_big") {
			result = common.T("", "games_win")
		} else if (playerChoice == common.T("", "games_small") || playerChoice == "small") && actualResult == common.T("", "games_small") {
			result = common.T("", "games_win")
		} else {
			result = common.T("", "games_lose")
		}

		// 发送结果
		resultMsg := fmt.Sprintf(common.T("", "games_bigsmall_result"), playerChoice, num, result)
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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "games_cmd_lottery"), event.RawMessage); !match {
			return nil
		}

		// 生成随机奖品
		prizes := []string{
			common.T("", "games_lottery_prize1"),
			common.T("", "games_lottery_prize2"),
			common.T("", "games_lottery_prize3"),
			common.T("", "games_lottery_thanks"),
		}
		prize := prizes[rand.Intn(len(prizes))]

		// 发送结果
		resultMsg := fmt.Sprintf(common.T("", "games_lottery_result"), prize)
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

		matchContinue, _, idiom := p.cmdParser.MatchCommandWithSingleParam(common.T("", "games_cmd_idiom"), event.RawMessage)
		if matchContinue && idiom != "" {
			return p.handleIdiomContinue(robot, event, idiom)
		}

		if matchStart, _ := p.cmdParser.MatchCommand(common.T("", "games_cmd_idiom"), event.RawMessage); matchStart {
			return p.handleIdiomStart(robot, event)
		}

		return nil
	})
}

// judgeRockPaperScissors 判断猜拳胜负
func (p *GamesPlugin) judgeRockPaperScissors(player, bot string) string {
	// 统一转换为中文
	if player == "rock" {
		player = common.T("", "games_rock")
	} else if player == "paper" {
		player = common.T("", "games_paper")
	} else if player == "scissors" {
		player = common.T("", "games_scissors")
	}

	if player == bot {
		return common.T("", "games_draw")
	}

	if (player == common.T("", "games_rock") && bot == common.T("", "games_scissors")) ||
		(player == common.T("", "games_scissors") && bot == common.T("", "games_paper")) ||
		(player == common.T("", "games_paper") && bot == common.T("", "games_rock")) {
		return common.T("", "games_win")
	}

	return common.T("", "games_lose")
}

func (p *GamesPlugin) getIdiomGameKey(event *onebot.Event) string {
	if event.MessageType == "group" {
		return fmt.Sprintf("group:%d", event.GroupID)
	}
	return fmt.Sprintf("user:%d", event.UserID)
}

func (p *GamesPlugin) handleIdiomStart(robot plugin.Robot, event *onebot.Event) error {
	if len(p.idioms) == 0 {
		p.sendMessage(robot, event, common.T("", "games_idiom_empty"))
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

	msg := fmt.Sprintf(common.T("", "games_idiom_start"), start, last)
	p.sendMessage(robot, event, msg)
	return nil
}

func (p *GamesPlugin) handleIdiomContinue(robot plugin.Robot, event *onebot.Event, idiom string) error {
	key := p.getIdiomGameKey(event)
	state, ok := p.idiomGames[key]
	if !ok || state.CurrentIdiom == "" {
		p.sendMessage(robot, event, common.T("", "games_idiom_not_started"))
		return nil
	}

	idiomRunes := []rune(idiom)
	if len(idiomRunes) < 2 {
		p.sendMessage(robot, event, common.T("", "games_idiom_invalid"))
		return nil
	}

	prevRunes := []rune(state.CurrentIdiom)
	if len(prevRunes) == 0 {
		state.CurrentIdiom = idiom
	} else {
		last := prevRunes[len(prevRunes)-1]
		first := idiomRunes[0]
		if last != first {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "games_idiom_wrong_char"), last))
			return nil
		}
		state.CurrentIdiom = idiom
	}

	botIdiom, ok := p.findNextIdiom(idiom)
	if !ok {
		delete(p.idiomGames, key)
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "games_idiom_win"), idiom))
		return nil
	}

	state.CurrentIdiom = botIdiom
	nextRunes := []rune(botIdiom)
	nextLast := ' '
	if len(nextRunes) > 0 {
		nextLast = nextRunes[len(nextRunes)-1]
	}

	msg := fmt.Sprintf(common.T("", "games_idiom_continue"), idiom, botIdiom, nextLast)
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
		log.Printf(common.T("", "games_send_failed"), err)
	}
}
