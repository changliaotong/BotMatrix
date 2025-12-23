package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
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
	return "塔罗牌占卜功能，可以抽取塔罗牌并查看解析"
}

func (p *TarotPlugin) Version() string {
	return "1.0.0"
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
		{Name: "愚人", Type: "大阿卡那", Description: "代表新的开始、冒险和自由", Upright: "新的开始、勇气、冒险", Reversed: "鲁莽、盲目、逃避"},
		{Name: "魔术师", Type: "大阿卡那", Description: "代表创造力、技能和自信", Upright: "创造力、自信、成功", Reversed: "操纵、不诚实、缺乏自信"},
		{Name: "女祭司", Type: "大阿卡那", Description: "代表直觉、智慧和神秘", Upright: "直觉、智慧、神秘", Reversed: "秘密、沉默、孤立"},
		{Name: "女皇", Type: "大阿卡那", Description: "代表母性、丰饶和爱", Upright: "丰饶、爱、母性", Reversed: "依赖、过度保护、虚荣"},
		{Name: "皇帝", Type: "大阿卡那", Description: "代表权威、控制和领导力", Upright: "权威、控制、领导力", Reversed: "独裁、严格、缺乏弹性"},
		{Name: "教皇", Type: "大阿卡那", Description: "代表传统、信仰和指导", Upright: "传统、信仰、指导", Reversed: "教条、僵化、盲目信仰"},
		{Name: "恋人", Type: "大阿卡那", Description: "代表爱情、选择和关系", Upright: "爱情、选择、关系", Reversed: "分离、诱惑、错误的选择"},
		{Name: "战车", Type: "大阿卡那", Description: "代表胜利、控制和决心", Upright: "胜利、控制、决心", Reversed: "冲突、缺乏控制、失败"},
		{Name: "力量", Type: "大阿卡那", Description: "代表勇气、力量和毅力", Upright: "勇气、力量、毅力", Reversed: "软弱、恐惧、缺乏自信"},
		{Name: "隐者", Type: "大阿卡那", Description: "代表智慧、孤独和内省", Upright: "智慧、孤独、内省", Reversed: "孤立、退缩、缺乏方向"},
		{Name: "命运之轮", Type: "大阿卡那", Description: "代表命运、变化和循环", Upright: "命运、变化、循环", Reversed: "停滞、厄运、抵抗变化"},
		{Name: "正义", Type: "大阿卡那", Description: "代表公正、平衡和法律", Upright: "公正、平衡、法律", Reversed: "不公正、失衡、偏见"},
		{Name: "倒吊人", Type: "大阿卡那", Description: "代表牺牲、等待和转变", Upright: "牺牲、等待、转变", Reversed: "牺牲过度、缺乏耐心、徒劳"},
		{Name: "死神", Type: "大阿卡那", Description: "代表结束、转变和重生", Upright: "结束、转变、重生", Reversed: "抗拒改变、恐惧、停滞"},
		{Name: "节制", Type: "大阿卡那", Description: "代表平衡、和谐和自我控制", Upright: "平衡、和谐、自我控制", Reversed: "失衡、过度、缺乏控制"},
		{Name: "恶魔", Type: "大阿卡那", Description: "代表欲望、诱惑和束缚", Upright: "欲望、诱惑、束缚", Reversed: "摆脱束缚、拒绝诱惑、自由"},
		{Name: "塔", Type: "大阿卡那", Description: "代表突然的变化、灾难和觉醒", Upright: "突然的变化、灾难、觉醒", Reversed: "避免灾难、延迟变化、内部崩溃"},
		{Name: "星星", Type: "大阿卡那", Description: "代表希望、灵感和指引", Upright: "希望、灵感、指引", Reversed: "绝望、缺乏灵感、迷失方向"},
		{Name: "月亮", Type: "大阿卡那", Description: "代表潜意识、恐惧和幻觉", Upright: "潜意识、恐惧、幻觉", Reversed: "释放恐惧、看清真相、觉醒"},
		{Name: "太阳", Type: "大阿卡那", Description: "代表成功、快乐和活力", Upright: "成功、快乐、活力", Reversed: "暂时的失败、缺乏活力、悲伤"},
		{Name: "审判", Type: "大阿卡那", Description: "代表重生、审判和觉醒", Upright: "重生、审判、觉醒", Reversed: "自我否定、延迟、内疚"},
		{Name: "世界", Type: "大阿卡那", Description: "代表完成、圆满和统一", Upright: "完成、圆满、统一", Reversed: "未完成、不圆满、分离"},
	}

	// 初始化小阿卡那牌
	 suits := []string{"权杖", "圣杯", "宝剑", "星币"}
	suitNames := map[string]string{
		"权杖": "火元素，代表行动、热情和创造力",
		"圣杯": "水元素，代表情感、爱和关系",
		"宝剑": "风元素，代表思想、沟通和挑战",
		"星币": "土元素，代表物质、财富和现实",
	}

	numbers := []string{"Ace", "2", "3", "4", "5", "6", "7", "8", "9", "10", "侍从", "骑士", "皇后", "国王"}

	for _, suit := range suits {
		for i, number := range numbers {
			card := TarotCard{
				Name:        number + " of " + suit,
				Type:        "小阿卡那",
				Suit:        suit,
				Number:      i + 1,
				Description: suitNames[suit],
				Upright:     "正位含义：" + number + " of " + suit,
				Reversed:    "逆位含义：" + number + " of " + suit + "（逆）",
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

	// 处理抽塔罗牌命令
	robot.OnMessage(func(event *onebot.Event) error {
		if match, _ := p.cmdParser.MatchCommand("抽塔罗牌", event.RawMessage); match {
			card, isUpright := p.DrawCard()
			var result string

			if isUpright {
				result = "🎴 塔罗牌占卜结果 🎴\n" +
					"\n【" + card.Name + "】" +
					"\n类型：" + card.Type +
					("\n花色：" + card.Suit + "\n").IfNotEmpty(card.Suit) +
					"\n" + card.Description +
					"\n\n✨ 正位含义：" + card.Upright
			} else {
				result = "🎴 塔罗牌占卜结果 🎴\n" +
					"\n【" + card.Name + "】" +
					"\n类型：" + card.Type +
					("\n花色：" + card.Suit + "\n").IfNotEmpty(card.Suit) +
					"\n" + card.Description +
					"\n\n🔄 逆位含义：" + card.Reversed
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

// IfNotEmpty 如果字符串非空则返回该字符串，否则返回空字符串
func (s string) IfNotEmpty(condition string) string {
	if condition != "" {
		return s
	}
	return ""
}
