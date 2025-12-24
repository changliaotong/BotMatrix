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

// IdiomGuessPlugin 猜成语插件
type IdiomGuessPlugin struct {
	cmdParser *CommandParser
	// 存储当前正在进行的游戏，key为用户ID，value为游戏数据
	games map[string]*IdiomGame
	// 成语列表
	idiomList []Idiom
}

// Idiom 成语数据结构
type Idiom struct {
	Word        string
	Pinyin      string
	Explanation string
	Example     string
}

// IdiomGame 游戏数据结构
type IdiomGame struct {
	UserID      string
	Idiom       Idiom
	Hint        string
	Guessed     string
	Attempts    int
	MaxAttempts int
	StartTime   time.Time
}

// NewIdiomGuessPlugin 创建猜成语插件实例
func NewIdiomGuessPlugin() *IdiomGuessPlugin {
	rand.Seed(time.Now().UnixNano())
	plugin := &IdiomGuessPlugin{
		cmdParser: NewCommandParser(),
		games:     make(map[string]*IdiomGame),
		idiomList: []Idiom{},
	}
	plugin.initIdiomList()
	return plugin
}

func (p *IdiomGuessPlugin) Name() string {
	return "idiom_guess"
}

func (p *IdiomGuessPlugin) Description() string {
	return common.T("", "idiom_guess_plugin_desc|猜成语游戏插件")
}

func (p *IdiomGuessPlugin) Version() string {
	return "1.0.0"
}

// initIdiomList 初始化成语列表
func (p *IdiomGuessPlugin) initIdiomList() {
	// 初始化常用成语列表
	p.idiomList = []Idiom{
		{Word: "一心一意", Pinyin: "yī xīn yī yì", Explanation: "形容做事专心一意，一门心思地只做一件事", Example: "他一心一意地学习，终于取得了好成绩"},
		{Word: "十全十美", Pinyin: "shí quán shí měi", Explanation: "十分完美，毫无欠缺", Example: "世界上没有十全十美的人"},
		{Word: "三心二意", Pinyin: "sān xīn èr yì", Explanation: "形容犹豫不决，意志不坚定或用心不专一", Example: "做事不能三心二意，否则什么都做不好"},
		{Word: "四面八方", Pinyin: "sì miàn bā fāng", Explanation: "指各个方面或各个地方", Example: "来自四面八方的朋友汇聚在一起"},
		{Word: "五颜六色", Pinyin: "wǔ yán liù sè", Explanation: "形容色彩复杂或花样繁多", Example: "公园里开满了五颜六色的花朵"},
		{Word: "六神无主", Pinyin: "liù shén wú zhǔ", Explanation: "形容心慌意乱，拿不定主意", Example: "面对突然的变故，他显得六神无主"},
		{Word: "七上八下", Pinyin: "qī shàng bā xià", Explanation: "形容心里慌乱不安，心神不定", Example: "考试成绩公布前，他心里七上八下的"},
		{Word: "八仙过海", Pinyin: "bā xiān guò hǎi", Explanation: "比喻各自拿出本领或办法，互相竞赛", Example: "在这次比赛中，选手们八仙过海，各显神通"},
		{Word: "九牛一毛", Pinyin: "jiǔ niú yī máo", Explanation: "比喻极大数量中极微小的数量，微不足道", Example: "这点损失对他来说只是九牛一毛"},
		{Word: "十拿九稳", Pinyin: "shí ná jiǔ wěn", Explanation: "比喻很有把握，十分可靠", Example: "这次考试他准备得很充分，十拿九稳能通过"},
		{Word: "百年好合", Pinyin: "bǎi nián hǎo hé", Explanation: "夫妻永远和好之意", Example: "祝福这对新人百年好合，永结同心"},
		{Word: "千方百计", Pinyin: "qiān fāng bǎi jì", Explanation: "想尽或用尽一切办法", Example: "他千方百计地寻找解决问题的方法"},
		{Word: "万紫千红", Pinyin: "wàn zǐ qiān hóng", Explanation: "形容百花齐放，色彩艳丽", Example: "春天的花园里万紫千红，美不胜收"},
		{Word: "亡羊补牢", Pinyin: "wáng yáng bǔ láo", Explanation: "比喻出了问题以后想办法补救，可以防止继续受损失", Example: "虽然犯了错误，但亡羊补牢，为时不晚"},
		{Word: "守株待兔", Pinyin: "shǒu zhū dài tù", Explanation: "比喻不主动努力，而存万一的侥幸心理，希望得到意外的收获", Example: "我们不能守株待兔，应该主动寻找机会"},
		{Word: "画龙点睛", Pinyin: "huà lóng diǎn jīng", Explanation: "比喻写文章或讲话时，在关键处用几句话点明实质，使内容更加生动有力", Example: "这篇文章的结尾起到了画龙点睛的作用"},
		{Word: "叶公好龙", Pinyin: "yè gōng hào lóng", Explanation: "比喻口头上说爱好某事物，实际上并不真爱好", Example: "他只是叶公好龙，并不是真正喜欢读书"},
		{Word: "井底之蛙", Pinyin: "jǐng dǐ zhī wā", Explanation: "比喻见识短浅的人", Example: "我们要多读书，多出去看看，不要做井底之蛙"},
		{Word: "掩耳盗铃", Pinyin: "yǎn ěr dào líng", Explanation: "比喻自己欺骗自己，明明掩盖不住的事情偏要想法子掩盖", Example: "这种做法无异于掩耳盗铃，自欺欺人"},
		{Word: "刻舟求剑", Pinyin: "kè zhōu qiú jiàn", Explanation: "比喻拘泥不知变通，不懂得根据实际情况处理问题", Example: "我们要学会灵活变通，不能刻舟求剑"},
	}
}

// GetSkills 实现 SkillCapable 接口
func (p *IdiomGuessPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "start",
			Description: common.T("", "idiom_guess_skill_start_desc|开始一个新的猜成语游戏"),
			Usage:       "start",
		},
		{
			Name:        "submit",
			Description: common.T("", "idiom_guess_skill_submit_desc|提交你的成语答案"),
			Usage:       "submit <answer>",
			Params: map[string]string{
				"answer": common.T("", "idiom_guess_skill_submit_param_answer|你猜的成语答案"),
			},
		},
		{
			Name:        "status",
			Description: common.T("", "idiom_guess_skill_status_desc|查看当前游戏的状态和进度"),
			Usage:       "status",
		},
		{
			Name:        "give_up",
			Description: common.T("", "idiom_guess_skill_giveup_desc|放弃当前游戏并查看正确答案"),
			Usage:       "give_up",
		},
	}
}

// HandleSkill 处理技能调用
func (p *IdiomGuessPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) error {
	userIDStr := fmt.Sprintf("%d", event.UserID)
	switch skillName {
	case "start":
		p.startNewGameLogic(robot, event, userIDStr)
	case "submit":
		answer := params["answer"]
		if answer == "" {
			p.sendMessage(robot, event, common.T("", "idiom_guess_enter_answer|请输入要提交的答案"))
			return nil
		}
		p.submitAnswerLogic(robot, event, userIDStr, answer)
	case "status":
		p.showGameStatusLogic(robot, event, userIDStr)
	case "give_up":
		p.giveUpGameLogic(robot, event, userIDStr)
	}
	return nil
}

// Init 初始化插件
func (p *IdiomGuessPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "idiom_guess_plugin_loaded|猜成语插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return "", p.HandleSkill(robot, nil, skillName, params)
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
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "idiom_guess") {
				HandleFeatureDisabled(robot, event, "idiom_guess")
				return nil
			}
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)

		// 检查是否为开始猜成语命令
		if match, _ := p.cmdParser.MatchCommand("猜成语|开始猜成语", event.RawMessage); match {
			p.startNewGameLogic(robot, event, userIDStr)
			return nil
		}

		// 检查是否为提交答案命令
		match, _, params := p.cmdParser.MatchCommandWithParams("提交", "(.+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, common.T("", "idiom_guess_cmd_submit_usage|用法：提交 <成语>"))
				return nil
			}
			answer := strings.TrimSpace(params[0])
			p.submitAnswerLogic(robot, event, userIDStr, answer)
			return nil
		}

		// 检查是否为查看当前游戏状态命令
		if match, _ := p.cmdParser.MatchCommand("查看游戏|游戏状态", event.RawMessage); match {
			p.showGameStatusLogic(robot, event, userIDStr)
			return nil
		}

		// 检查是否为放弃游戏命令
		if match, _ := p.cmdParser.MatchCommand("放弃游戏|结束游戏", event.RawMessage); match {
			p.giveUpGameLogic(robot, event, userIDStr)
			return nil
		}

		return nil
	})
}

// startNewGameLogic 开始新游戏逻辑
func (p *IdiomGuessPlugin) startNewGameLogic(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否已有正在进行的游戏
	if _, exists := p.games[userID]; exists {
		p.sendMessage(robot, event, common.T("", "idiom_guess_already_started|你已经有一个正在进行的猜成语游戏了！"))
		return
	}

	// 随机选择一个成语
	idiom := p.idiomList[rand.Intn(len(p.idiomList))]

	// 生成提示
	hint := fmt.Sprintf("解释：%s\n示例：%s", idiom.Explanation, idiom.Example)

	// 生成已猜字符串（初始全为下划线）
	guessed := strings.Repeat("_", len(idiom.Word))

	// 创建新游戏
	game := &IdiomGame{
		UserID:      userID,
		Idiom:       idiom,
		Hint:        hint,
		Guessed:     guessed,
		Attempts:    0,
		MaxAttempts: 6,
		StartTime:   time.Now(),
	}

	p.games[userID] = game

	// 发送游戏开始消息
	p.sendMessage(robot, event, fmt.Sprintf(
		common.T("", "idiom_guess_start_msg|猜成语游戏开始！\n%s\n当前：\n%s\n你有 %d 次尝试机会。"),
		game.Hint, game.Guessed, game.MaxAttempts,
	))
}

// submitAnswerLogic 提交答案逻辑
func (p *IdiomGuessPlugin) submitAnswerLogic(robot plugin.Robot, event *onebot.Event, userID string, answer string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, common.T("", "idiom_guess_no_game|你当前没有正在进行的猜成语游戏。"))
		return
	}

	// 增加尝试次数
	game.Attempts++

	// 检查答案是否正确
	if strings.EqualFold(answer, game.Idiom.Word) {
		// 猜对了
		duration := time.Since(game.StartTime)
		p.sendMessage(robot, event, fmt.Sprintf(
			common.T("", "idiom_guess_correct|恭喜你猜对了！\n正确答案是：%s (%s)\n用时：%v\n尝试次数：%d/%d"),
			game.Idiom.Word, game.Idiom.Pinyin, duration.Round(time.Second), game.Attempts, game.MaxAttempts,
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
			common.T("", "idiom_guess_game_over|游戏结束！机会已用完。\n正确答案是：%s (%s)\n释义：%s"),
			game.Idiom.Word, game.Idiom.Pinyin, game.Idiom.Explanation,
		))
		// 删除游戏
		delete(p.games, userID)
		return
	}

	// 显示当前状态
	p.sendMessage(robot, event, fmt.Sprintf(
		common.T("", "idiom_guess_wrong|很遗憾，猜错了。\n当前：%s\n剩余机会：%d"),
		game.Guessed, remaining,
	))
}

// showGameStatusLogic 显示当前游戏状态逻辑
func (p *IdiomGuessPlugin) showGameStatusLogic(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, common.T("", "idiom_guess_no_game|你当前没有正在进行的猜成语游戏。"))
		return
	}

	remaining := game.MaxAttempts - game.Attempts
	duration := time.Since(game.StartTime)

	p.sendMessage(robot, event, fmt.Sprintf(
		common.T("", "idiom_guess_status|当前猜成语游戏状态：\n%s\n进度：%s\n已尝试：%d/%d 次\n剩余机会：%d 次\n已用时间：%v"),
		game.Hint, game.Guessed, game.Attempts, game.MaxAttempts, remaining, duration.Round(time.Second),
	))
}

// giveUpGameLogic 放弃游戏逻辑
func (p *IdiomGuessPlugin) giveUpGameLogic(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查是否有正在进行的游戏
	game, exists := p.games[userID]
	if !exists {
		p.sendMessage(robot, event, common.T("", "idiom_guess_no_game|你当前没有正在进行的猜成语游戏。"))
		return
	}

	// 显示放弃消息
	p.sendMessage(robot, event, fmt.Sprintf(
		common.T("", "idiom_guess_give_up|好吧，你选择了放弃。游戏已结束。\n正确答案是：%s (%s)\n释义：%s"),
		game.Idiom.Word, game.Idiom.Pinyin, game.Idiom.Explanation,
	))

	// 删除游戏
	delete(p.games, userID)
}

// sendMessage 发送消息
func (p *IdiomGuessPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		return
	}
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
