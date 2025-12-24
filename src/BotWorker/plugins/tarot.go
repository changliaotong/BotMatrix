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

// TarotCard 代表一张塔罗牌
type TarotCard struct {
	Name        string // 牌名
	Type        string // 类型：大阿卡那/小阿卡那
	Suit        string // 花色（小阿卡那）
	Number      int    // 数字（小阿卡那）
	Description string // 描述
	Upright     string // 正位含义
	Reversed    string // 逆位含义
}

// TarotPlugin 塔罗牌插件
type TarotPlugin struct {
	cmdParser *CommandParser
	cards     []TarotCard
}

func (p *TarotPlugin) Name() string {
	return "tarot"
}

func (p *TarotPlugin) Description() string {
	return common.T("", "tarot_plugin_desc|塔罗牌占卜功能，可以抽取塔罗牌并查看解析")
}

func (p *TarotPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *TarotPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "draw_tarot",
			Description: common.T("", "tarot_skill_draw_desc|抽取一张塔罗牌并获得解析"),
			Usage:       "draw_tarot",
			Params:      map[string]string{},
		},
	}
}

// NewTarotPlugin 创建塔罗牌插件实例
func NewTarotPlugin() *TarotPlugin {
	plugin := &TarotPlugin{
		cmdParser: NewCommandParser(),
		cards:     make([]TarotCard, 0),
	}
	plugin.initCards()
	return plugin
}

// initCards 初始化塔罗牌数据
func (p *TarotPlugin) initCards() {
	// 初始化大阿卡那牌
	majorArcana := []TarotCard{
		{Name: common.T("", "tarot_card_0_name|愚人"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_0_desc|代表新的开始、冒险和自由"), Upright: common.T("", "tarot_card_0_upright|新的开始、勇气、冒险"), Reversed: common.T("", "tarot_card_0_reversed|鲁莽、盲目、逃避")},
		{Name: common.T("", "tarot_card_1_name|魔术师"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_1_desc|代表创造力、技能和自信"), Upright: common.T("", "tarot_card_1_upright|创造力、自信、成功"), Reversed: common.T("", "tarot_card_1_reversed|操纵、不诚实、缺乏自信")},
		{Name: common.T("", "tarot_card_2_name|女祭司"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_2_desc|代表直觉、智慧和神秘"), Upright: common.T("", "tarot_card_2_upright|直觉、智慧、神秘"), Reversed: common.T("", "tarot_card_2_reversed|秘密、沉默、孤立")},
		{Name: common.T("", "tarot_card_3_name|女皇"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_3_desc|代表母性、丰饶和爱"), Upright: common.T("", "tarot_card_3_upright|丰饶、爱、母性"), Reversed: common.T("", "tarot_card_3_reversed|依赖、过度保护、虚荣")},
		{Name: common.T("", "tarot_card_4_name|皇帝"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_4_desc|代表权威、控制和领导力"), Upright: common.T("", "tarot_card_4_upright|权威、控制、领导力"), Reversed: common.T("", "tarot_card_4_reversed|独裁、严格、缺乏弹性")},
		{Name: common.T("", "tarot_card_5_name|教皇"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_5_desc|代表传统、信仰和指导"), Upright: common.T("", "tarot_card_5_upright|传统、信仰、指导"), Reversed: common.T("", "tarot_card_5_reversed|教条、僵化、盲目信仰")},
		{Name: common.T("", "tarot_card_6_name|恋人"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_6_desc|代表爱情、选择和关系"), Upright: common.T("", "tarot_card_6_upright|爱情、选择、关系"), Reversed: common.T("", "tarot_card_6_reversed|分离、诱惑、错误的选择")},
		{Name: common.T("", "tarot_card_7_name|战车"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_7_desc|代表胜利、控制和决心"), Upright: common.T("", "tarot_card_7_upright|胜利、控制、决心"), Reversed: common.T("", "tarot_card_7_reversed|冲突、缺乏控制、失败")},
		{Name: common.T("", "tarot_card_8_name|力量"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_8_desc|代表勇气、力量和毅力"), Upright: common.T("", "tarot_card_8_upright|勇气、力量、毅力"), Reversed: common.T("", "tarot_card_8_reversed|软弱、恐惧、缺乏自信")},
		{Name: common.T("", "tarot_card_9_name|隐者"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_9_desc|代表智慧、孤独和内省"), Upright: common.T("", "tarot_card_9_upright|智慧、孤独、内省"), Reversed: common.T("", "tarot_card_9_reversed|孤立、退缩、缺乏方向")},
		{Name: common.T("", "tarot_card_10_name|命运之轮"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_10_desc|代表命运、变化和循环"), Upright: common.T("", "tarot_card_10_upright|命运、变化、循环"), Reversed: common.T("", "tarot_card_10_reversed|停滞、厄运、抵抗变化")},
		{Name: common.T("", "tarot_card_11_name|正义"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_11_desc|代表公正、平衡和法律"), Upright: common.T("", "tarot_card_11_upright|公正、平衡、法律"), Reversed: common.T("", "tarot_card_11_reversed|不公正、失衡、偏见")},
		{Name: common.T("", "tarot_card_12_name|倒吊人"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_12_desc|代表牺牲、等待和转变"), Upright: common.T("", "tarot_card_12_upright|牺牲、等待、转变"), Reversed: common.T("", "tarot_card_12_reversed|牺牲过度、缺乏耐心、徒劳")},
		{Name: common.T("", "tarot_card_13_name|死神"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_13_desc|代表结束、转变和重生"), Upright: common.T("", "tarot_card_13_upright|结束、转变、重生"), Reversed: common.T("", "tarot_card_13_reversed|抗拒改变、恐惧、停滞")},
		{Name: common.T("", "tarot_card_14_name|节制"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_14_desc|代表平衡、和谐和自我控制"), Upright: common.T("", "tarot_card_14_upright|平衡、和谐、自我控制"), Reversed: common.T("", "tarot_card_14_reversed|失衡、过度、缺乏控制")},
		{Name: common.T("", "tarot_card_15_name|恶魔"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_15_desc|代表欲望、诱惑和束缚"), Upright: common.T("", "tarot_card_15_upright|欲望、诱惑、束缚"), Reversed: common.T("", "tarot_card_15_reversed|摆脱束缚、拒绝诱惑、自由")},
		{Name: common.T("", "tarot_card_16_name|塔"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_16_desc|代表突然的变化、灾难和觉醒"), Upright: common.T("", "tarot_card_16_upright|突然的变化、灾难、觉醒"), Reversed: common.T("", "tarot_card_16_reversed|避免灾难、延迟变化、内部崩溃")},
		{Name: common.T("", "tarot_card_17_name|星星"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_17_desc|代表希望、灵感和指引"), Upright: common.T("", "tarot_card_17_upright|希望、灵感、指引"), Reversed: common.T("", "tarot_card_17_reversed|绝望、缺乏灵感、迷失方向")},
		{Name: common.T("", "tarot_card_18_name|月亮"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_18_desc|代表潜意识、恐惧和幻觉"), Upright: common.T("", "tarot_card_18_upright|潜意识、恐惧、幻觉"), Reversed: common.T("", "tarot_card_18_reversed|释放恐惧、看清真相、觉醒")},
		{Name: common.T("", "tarot_card_19_name|太阳"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_19_desc|代表成功、快乐和活力"), Upright: common.T("", "tarot_card_19_upright|成功、快乐、活力"), Reversed: common.T("", "tarot_card_19_reversed|暂时的失败、缺乏活力、悲伤")},
		{Name: common.T("", "tarot_card_20_name|审判"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_20_desc|代表重生、审判和觉醒"), Upright: common.T("", "tarot_card_20_upright|重生、审判、觉醒"), Reversed: common.T("", "tarot_card_20_reversed|自我否定、延迟、内疚")},
		{Name: common.T("", "tarot_card_21_name|世界"), Type: common.T("", "tarot_type_major|大阿卡那"), Description: common.T("", "tarot_card_21_desc|代表完成、圆满和统一"), Upright: common.T("", "tarot_card_21_upright|完成、圆满、统一"), Reversed: common.T("", "tarot_card_21_reversed|未完成、不圆满、分离")},
	}

	// 初始化小阿卡那牌
	suits := []string{
		common.T("", "tarot_suit_wands|权杖"),
		common.T("", "tarot_suit_cups|圣杯"),
		common.T("", "tarot_suit_swords|宝剑"),
		common.T("", "tarot_suit_pentacles|星币"),
	}
	suitNames := map[string]string{
		common.T("", "tarot_suit_wands|权杖"):     common.T("", "tarot_suit_wands_desc|火元素，代表行动、热情和创造力"),
		common.T("", "tarot_suit_cups|圣杯"):      common.T("", "tarot_suit_cups_desc|水元素，代表情感、爱和关系"),
		common.T("", "tarot_suit_swords|宝剑"):    common.T("", "tarot_suit_swords_desc|风元素，代表思想、沟通和挑战"),
		common.T("", "tarot_suit_pentacles|星币"): common.T("", "tarot_suit_pentacles_desc|土元素，代表物质、财富和现实"),
	}

	numbers := []string{
		"Ace", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		common.T("", "tarot_num_page|侍从"),
		common.T("", "tarot_num_knight|骑士"),
		common.T("", "tarot_num_queen|皇后"),
		common.T("", "tarot_num_king|国王"),
	}

	for _, suit := range suits {
		for i, number := range numbers {
			card := TarotCard{
				Name:        number + " of " + suit,
				Type:        common.T("", "tarot_type_minor|小阿卡那"),
				Suit:        suit,
				Number:      i + 1,
				Description: suitNames[suit],
				Upright:     common.T("", "tarot_msg_upright_prefix|正位含义：") + number + " of " + suit,
				Reversed:    common.T("", "tarot_msg_reversed_prefix|逆位含义：") + number + " of " + suit + common.T("", "tarot_msg_reversed_suffix|（逆）"),
			}
			p.cards = append(p.cards, card)
		}
	}

	// 添加大阿卡那牌
	p.cards = append(p.cards, majorArcana...)

	log.Printf("初始化塔罗牌完成，共 %d 张牌", len(p.cards))
}

// DrawCard 随机抽取一张塔罗牌
func (p *TarotPlugin) DrawCard() (TarotCard, bool) {
	rand.Seed(time.Now().UnixNano())
	card := p.cards[rand.Intn(len(p.cards))]
	isUpright := rand.Intn(2) == 0 // 50% 概率正位
	return card, isUpright
}

func (p *TarotPlugin) Init(robot plugin.Robot) {
	log.Println("加载塔罗牌插件")

	// 注册技能处理器
	robot.HandleSkill("draw_tarot", func(params map[string]string) (string, error) {
		if len(p.cards) == 0 {
			return "", fmt.Errorf("tarot cards not initialized")
		}

		rand.Seed(time.Now().UnixNano())
		cardIndex := rand.Intn(len(p.cards))
		card := p.cards[cardIndex]
		isUpright := rand.Intn(2) == 0

		direction := common.T("", "tarot_msg_upright|正位")
		meaning := card.Upright
		if !isUpright {
			direction = common.T("", "tarot_msg_reversed|逆位")
			meaning = card.Reversed
		}

		return fmt.Sprintf(common.T("", "tarot_msg_draw_result|你抽到了: %s (%s)\n类型: %s\n含义: %s"), card.Name, direction, card.Type, meaning), nil
	})

	// 处理抽塔罗牌命令
	robot.OnMessage(func(event *onebot.Event) error {
		if match, _ := p.cmdParser.MatchCommand(common.T("", "tarot_cmd_draw|抽塔罗牌"), event.RawMessage); match {
			card, isUpright := p.DrawCard()
			var result string

			header := common.T("", "tarot_msg_header|🎴 塔罗牌占卜结果 🎴")
			typeLabel := common.T("", "tarot_msg_type|类型：")
			suitLabel := common.T("", "tarot_msg_suit|花色：")

			if isUpright {
				result = header + "\n" +
					"\n【" + card.Name + "】" +
					"\n" + typeLabel + card.Type
				if card.Suit != "" {
					result += "\n" + suitLabel + card.Suit + "\n"
				}
				result += "\n" + card.Description +
					"\n\n✨ " + common.T("", "tarot_msg_upright_meaning|正位含义：") + card.Upright
			} else {
				result = header + "\n" +
					"\n【" + card.Name + "】" +
					"\n" + typeLabel + card.Type
				if card.Suit != "" {
					result += "\n" + suitLabel + card.Suit + "\n"
				}
				result += "\n" + card.Description +
					"\n\n🔄 " + common.T("", "tarot_msg_reversed_meaning|逆位含义：") + card.Reversed
			}

			// 发送结果
			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: result,
			}
			if event.GroupID != 0 {
				params.GroupID = event.GroupID
				params.MessageType = "group"
			} else {
				params.MessageType = "private"
			}
			robot.SendMessage(params)
		}
		return nil
	})
}
