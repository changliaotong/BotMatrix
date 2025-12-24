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

// LotteryPlugin 抽签插件
type LotteryPlugin struct {
	// 存储用户抽签记录，key为用户ID，value为上次抽签时间
	lastLotteryTime map[string]time.Time
	// 签文列表
	lotteries []Lottery
	// 命令解析器
	cmdParser *CommandParser
}

// Lottery 签文
type Lottery struct {
	Name           string // 签名
	Content        string // 签文内容
	Interpretation string // 解签内容
	Level          int    // 签的等级（1-5，1为上上签，5为下下签）
}

// NewLotteryPlugin 创建抽签插件实例
func NewLotteryPlugin() *LotteryPlugin {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 初始化签文列表
	lotteries := []Lottery{
		{
			Name:           common.T("", "lottery_level1_name|上上签"),
			Content:        common.T("", "lottery_level1_content|大吉大利，万事如意。"),
			Interpretation: common.T("", "lottery_level1_interpretation|这是一个非常好的签位，预示着你近期将会有非常好的运气。"),
			Level:          1,
		},
		{
			Name:           common.T("", "lottery_level2_name|大吉"),
			Content:        common.T("", "lottery_level2_content|顺风顺水，马到成功。"),
			Interpretation: common.T("", "lottery_level2_interpretation|这是一个大吉的签位，预示着你的事业或学业将会取得显著的进展。"),
			Level:          2,
		},
		{
			Name:           common.T("", "lottery_level3_name|中吉"),
			Content:        common.T("", "lottery_level3_content|平平安安，细水长流。"),
			Interpretation: common.T("", "lottery_level3_interpretation|这是一个中吉的签位，预示着你的生活将会非常平稳，没有什么大起大落。"),
			Level:          3,
		},
		{
			Name:           common.T("", "lottery_level4_name|小吉"),
			Content:        common.T("", "lottery_level4_content|小有收获，需多努力。"),
			Interpretation: common.T("", "lottery_level4_interpretation|这是一个小吉的签位，预示着你虽然会有一些小的收获，但仍需要付出努力。"),
			Level:          4,
		},
		{
			Name:           common.T("", "lottery_level5_name|末吉"),
			Content:        common.T("", "lottery_level5_content|守得云开见月明。"),
			Interpretation: common.T("", "lottery_level5_interpretation|这是一个末吉的签位，预示着你目前可能会遇到一些小困难，但只要坚持下去，最终会看到希望。"),
			Level:          5,
		},
	}

	return &LotteryPlugin{
		lastLotteryTime: make(map[string]time.Time),
		lotteries:       lotteries,
		cmdParser:       NewCommandParser(),
	}
}

func (p *LotteryPlugin) Name() string {
	return "lottery"
}

func (p *LotteryPlugin) Description() string {
	return common.T("", "lottery_plugin_desc|🔮 抽签插件，支持每日抽签和解签")
}

func (p *LotteryPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *LotteryPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "draw_lottery",
			Description: common.T("", "lottery_skill_draw_desc|抽取今日运势签"),
			Usage:       "draw_lottery user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "lottery_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "interpret_lottery",
			Description: common.T("", "lottery_skill_interpret_desc|解析已抽取的签文"),
			Usage:       "interpret_lottery user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "lottery_skill_param_user_id|用户ID"),
			},
		},
	}
}

func (p *LotteryPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "lottery_plugin_loaded|✅ 抽签插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return "", p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理抽签和解签命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "lottery") {
				HandleFeatureDisabled(robot, event, "lottery")
				return nil
			}
		}

		// 检查是否为抽签命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "lottery_cmd_draw|/抽签|抽签|draw|lottery"), event.RawMessage); match {
			userID := event.UserID
			if userID == 0 {
				p.sendMessage(robot, event, common.T("", "lottery_invalid_userid|❌ 无效的用户ID"))
				return nil
			}
			p.sendMessage(robot, event, p.doDrawLottery(fmt.Sprintf("%d", userID)))
			return nil
		}

		// 检查是否为解签命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "lottery_cmd_interpret|/解签|解签|interpret"), event.RawMessage); match {
			userID := event.UserID
			if userID == 0 {
				p.sendMessage(robot, event, common.T("", "lottery_invalid_userid|❌ 无效的用户ID"))
				return nil
			}
			p.sendMessage(robot, event, p.doInterpretLottery(fmt.Sprintf("%d", userID)))
			return nil
		}

		return nil
	})
}

// HandleSkill 处理技能调用
func (p *LotteryPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) error {
	userID := params["user_id"]
	if userID == "" {
		if event != nil {
			userID = fmt.Sprintf("%d", event.UserID)
		}
	}

	if userID == "" {
		return fmt.Errorf(common.T("", "lottery_missing_user_id|❌ 缺少用户ID参数"))
	}

	switch skillName {
	case "draw_lottery":
		p.sendMessage(robot, event, p.doDrawLottery(userID))
		return nil
	case "interpret_lottery":
		p.sendMessage(robot, event, p.doInterpretLottery(userID))
		return nil
	default:
		return fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doDrawLottery 执行抽签逻辑
func (p *LotteryPlugin) doDrawLottery(userID string) string {
	// 检查是否已经抽过签（每天限抽一次）
	now := time.Now()
	if lastLottery, ok := p.lastLotteryTime[userID]; ok {
		// 检查是否在同一天
		if isSameDay(lastLottery, now) {
			return fmt.Sprintf(common.T("", "lottery_already_drawn|⏳ 你今天已经在 %s 抽过签了，明天再来吧！"), lastLottery.Format("15:04:05"))
		}
	}

	// 随机抽取一个签
	lottery := p.lotteries[rand.Intn(len(p.lotteries))]

	// 更新抽签记录
	p.lastLotteryTime[userID] = now

	// 发送抽签结果
	msg := common.T("", "lottery_result_header|✨ 抽签结果 ✨\n")
	msg += fmt.Sprintf(common.T("", "lottery_result_name|【签名】：%s\n"), lottery.Name)
	msg += fmt.Sprintf(common.T("", "lottery_result_content|【签文】：%s\n"), lottery.Content)
	msg += fmt.Sprintf(common.T("", "lottery_result_interpretation|【解签】：%s"), lottery.Interpretation)

	return msg
}

// doInterpretLottery 执行解签逻辑
func (p *LotteryPlugin) doInterpretLottery(userID string) string {
	// 检查是否有抽签记录
	if _, ok := p.lastLotteryTime[userID]; !ok {
		return common.T("", "lottery_not_drawn|❌ 你今天还没有抽签，请先输入“/抽签”哦！")
	}

	// 重新抽取上次的签（模拟解签）
	lottery := p.lotteries[rand.Intn(len(p.lotteries))]

	// 发送解签结果
	msg := common.T("", "lottery_interpret_header|✨ 解签结果 ✨\n")
	msg += fmt.Sprintf(common.T("", "lottery_result_name|【签名】：%s\n"), lottery.Name)
	msg += fmt.Sprintf(common.T("", "lottery_result_content|【签文】：%s\n"), lottery.Content)
	msg += fmt.Sprintf(common.T("", "lottery_result_interpretation|【解签】：%s"), lottery.Interpretation)

	return msg
}

// sendMessage 发送消息
func (p *LotteryPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "lottery_send_failed|❌ 发送消息失败：%v"), err)
	}
}
