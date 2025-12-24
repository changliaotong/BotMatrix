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

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

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

		_, err := p.playRockPaperScissors(robot, event, playerChoice)
		return err
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

		_, err := p.playBigSmall(robot, event, playerChoice)
		return err
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

		_, err := p.playLottery(robot, event)
		return err
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
			_, err := p.handleIdiomContinue(robot, event, idiom)
			return err
		}

		if matchStart, _ := p.cmdParser.MatchCommand(common.T("", "games_cmd_idiom"), event.RawMessage); matchStart {
			_, err := p.handleIdiomStart(robot, event)
			return err
		}

		return nil
	})
}

// GetSkills 实现 SkillCapable 接口
func (p *GamesPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "rock_paper_scissors",
			Description: common.T("", "games_skill_rps_desc"),
			Usage:       common.T("", "games_skill_rps_usage"),
			Params: map[string]string{
				"choice": common.T("", "games_skill_rps_param_choice"),
			},
		},
		{
			Name:        "big_small",
			Description: common.T("", "games_skill_bigsmall_desc"),
			Usage:       common.T("", "games_skill_bigsmall_usage"),
			Params: map[string]string{
				"choice": common.T("", "games_skill_bigsmall_param_choice"),
			},
		},
		{
			Name:        "lottery",
			Description: common.T("", "games_skill_lottery_desc"),
			Usage:       common.T("", "games_skill_lottery_usage"),
			Params:      map[string]string{},
		},
		{
			Name:        "idiom_start",
			Description: common.T("", "games_skill_idiom_start_desc"),
			Usage:       common.T("", "games_skill_idiom_start_usage"),
			Params:      map[string]string{},
		},
		{
			Name:        "idiom_continue",
			Description: common.T("", "games_skill_idiom_continue_desc"),
			Usage:       common.T("", "games_skill_idiom_continue_usage"),
			Params: map[string]string{
				"idiom": common.T("", "games_skill_idiom_continue_param_idiom"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *GamesPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "rock_paper_scissors":
		choice := params["choice"]
		if choice == "" {
			msg := common.T("", "games_rock_invalid")
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		return p.playRockPaperScissors(robot, event, choice)
	case "big_small":
		choice := params["choice"]
		if choice == "" {
			msg := common.T("", "games_bigsmall_invalid")
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		return p.playBigSmall(robot, event, choice)
	case "lottery":
		return p.playLottery(robot, event)
	case "idiom_start":
		return p.handleIdiomStart(robot, event)
	case "idiom_continue":
		idiom := params["idiom"]
		if idiom == "" {
			msg := common.T("", "games_idiom_invalid")
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		return p.handleIdiomContinue(robot, event, idiom)
	}
	return "", nil
}

func (p *GamesPlugin) playRockPaperScissors(robot plugin.Robot, event *onebot.Event, playerChoice string) (string, error) {
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
		msg := common.T("", "games_rock_invalid")
		p.sendMessage(robot, event, msg)
		return msg, nil
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

	return resultMsg, nil
}

func (p *GamesPlugin) playBigSmall(robot plugin.Robot, event *onebot.Event, playerChoice string) (string, error) {
	// 验证玩家选择
	validChoices := map[string]bool{
		common.T("", "games_big"):   true,
		common.T("", "games_small"): true,
		"big":                       true,
		"small":                     true,
	}
	if !validChoices[playerChoice] {
		msg := common.T("", "games_bigsmall_invalid")
		p.sendMessage(robot, event, msg)
		return msg, nil
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

	return resultMsg, nil
}

func (p *GamesPlugin) playLottery(robot plugin.Robot, event *onebot.Event) (string, error) {
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

	return resultMsg, nil
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

func (p *GamesPlugin) handleIdiomStart(robot plugin.Robot, event *onebot.Event) (string, error) {
	if len(p.idioms) == 0 {
		msg := common.T("", "games_idiom_empty")
		p.sendMessage(robot, event, msg)
		return msg, nil
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
	return msg, nil
}

func (p *GamesPlugin) handleIdiomContinue(robot plugin.Robot, event *onebot.Event, idiom string) (string, error) {
	key := p.getIdiomGameKey(event)
	state, ok := p.idiomGames[key]
	if !ok || state.CurrentIdiom == "" {
		msg := common.T("", "games_idiom_not_started")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	idiomRunes := []rune(idiom)
	if len(idiomRunes) < 2 {
		msg := common.T("", "games_idiom_invalid")
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	prevRunes := []rune(state.CurrentIdiom)
	if len(prevRunes) == 0 {
		state.CurrentIdiom = idiom
	} else {
		last := prevRunes[len(prevRunes)-1]
		first := idiomRunes[0]
		if last != first {
			msg := fmt.Sprintf(common.T("", "games_idiom_wrong_char"), last)
			p.sendMessage(robot, event, msg)
			return msg, nil
		}
		state.CurrentIdiom = idiom
	}

	botIdiom, ok := p.findNextIdiom(idiom)
	if !ok {
		delete(p.idiomGames, key)
		msg := fmt.Sprintf(common.T("", "games_idiom_win"), idiom)
		p.sendMessage(robot, event, msg)
		return msg, nil
	}

	state.CurrentIdiom = botIdiom
	nextRunes := []rune(botIdiom)
	nextLast := ' '
	if len(nextRunes) > 0 {
		nextLast = nextRunes[len(nextRunes)-1]
	}

	msg := fmt.Sprintf(common.T("", "games_idiom_continue"), idiom, botIdiom, nextLast)
	p.sendMessage(robot, event, msg)
	return msg, nil
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
	if robot == nil || event == nil {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "games_send_failed"), err)
	}
}
