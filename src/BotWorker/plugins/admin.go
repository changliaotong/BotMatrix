package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
)

// AdminPlugin  admin plugin
type AdminPlugin struct {
	// 管理员列表
	admins []string
	// 功能开关
	featureSwitches map[string]bool
	// 命令解析器
	cmdParser *CommandParser
}

func (p *AdminPlugin) Name() string {
	return "admin"
}

func (p *AdminPlugin) Description() string {
	return "admin plugin，支持后台设置、功能开关、教学等功能"
}

func (p *AdminPlugin) Version() string {
	return "1.0.0"
}

// NewAdminPlugin 创建admin plugin实例
func NewAdminPlugin() *AdminPlugin {
	return &AdminPlugin{
		admins: []string{},
		featureSwitches: map[string]bool{
			"weather":    true,
			"points":     true,
			"signin":     true,
			"lottery":    true,
			"translate":  true,
			"music":      true,
			"games":      true,
			"greetings":  true,
			"utils":      true,
			"moderation": true,
		},
		cmdParser: NewCommandParser(),
	}
}

func (p *AdminPlugin) Init(robot plugin.Robot) {
	log.Println("加载admin plugin")

	// 处理后台命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为后台命令
		if match, _ := p.cmdParser.MatchCommand("后台|admin", event.RawMessage); !match {
			return nil
		}

		// 发送后台菜单
		adminMenu := "🔧 后台管理菜单\n"
		adminMenu += "====================\n"
		adminMenu += "/开启 <功能> - 开启指定功能\n"
		adminMenu += "/关闭 <功能> - 关闭指定功能\n"
		adminMenu += "/设置 <参数> <值> - 设置参数\n"
		adminMenu += "/教学 - 查看使用教程\n"
		adminMenu += "/本群 - 查看本群信息\n"
		adminMenu += "/话唠 - 开启话唠模式\n"
		adminMenu += "/终极 - 开启终极模式\n"
		adminMenu += "/智能体 - 开启智能体模式\n"
		p.sendMessage(robot, event, adminMenu)

		return nil
	})

	// 处理开启命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为开启命令
		match, _, params := p.cmdParser.MatchCommandWithParams("开启|enable", `(.*)`, event.RawMessage)
		if !match || len(params) < 1 {
			return nil
		}

		// 解析功能名称
		feature := params[0]

		// 检查功能是否存在
		if _, ok := p.featureSwitches[feature]; !ok {
			p.sendMessage(robot, event, fmt.Sprintf("功能%s不存在", feature))
			return nil
		}

		// 开启功能
		p.featureSwitches[feature] = true
		p.sendMessage(robot, event, fmt.Sprintf("功能%s已开启", feature))

		return nil
	})

	// 处理关闭命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为关闭命令
		match, _, params := p.cmdParser.MatchCommandWithParams("关闭|disable", `(.*)`, event.RawMessage)
		if !match || len(params) < 1 {
			return nil
		}

		// 解析功能名称
		feature := params[0]

		// 检查功能是否存在
		if _, ok := p.featureSwitches[feature]; !ok {
			p.sendMessage(robot, event, fmt.Sprintf("功能%s不存在", feature))
			return nil
		}

		// 关闭功能
		p.featureSwitches[feature] = false
		p.sendMessage(robot, event, fmt.Sprintf("功能%s已关闭", feature))

		return nil
	})

	// 处理设置命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为设置命令
		match, _, params := p.cmdParser.MatchCommandWithParams("设置|set", `([^\s]+)\s+(.+)`, event.RawMessage)
		if !match || len(params) < 2 {
			return nil
		}

		// 解析参数和值
		param := params[0]
		value := params[1]

		// 模拟设置
		p.sendMessage(robot, event, fmt.Sprintf("参数%s已设置为%s", param, value))

		return nil
	})

	// 处理教学命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为教学命令
		if match, _ := p.cmdParser.MatchCommand("教学|help", event.RawMessage); !match {
			return nil
		}

		// 发送教学内容
		teaching := "📚 使用教程\n"
		teaching += "====================\n"
		teaching += "/菜单 - 查看所有命令\n"
		teaching += "/help - 查看帮助信息\n"
		teaching += "/签到 - 每日签到\n"
		teaching += "/积分 - 查询积分\n"
		teaching += "/天气 <城市> - 查询天气\n"
		teaching += "/翻译 <文本> - 翻译文本\n"
		teaching += "/点歌 <歌曲> - 点歌\n"
		teaching += "/猜拳 <选择> - 猜拳\n"
		teaching += "/猜大小 <选择> - 猜大小\n"
		teaching += "/抽奖 - 抽奖\n"
		teaching += "/早安 - 早安问候\n"
		teaching += "/晚安 - 晚安问候\n"
		teaching += "/报时 - 查看当前时间\n"
		teaching += "/计算 <表达式> - 计算\n"
		teaching += "/笑话 - 讲笑话\n"
		teaching += "/鬼故事 - 讲鬼故事\n"
		teaching += "/成语接龙 <成语> - 成语接龙\n"
		p.sendMessage(robot, event, teaching)

		return nil
	})

	// 处理本群命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为本群命令
		if match, _ := p.cmdParser.MatchCommand("本群|group", event.RawMessage); !match {
			return nil
		}

		// 发送本群信息
		groupInfo := "🏠 本群信息\n"
		groupInfo += "====================\n"
		groupInfo += "群名称：未知\n"
		groupInfo += "群人数：未知\n"
		groupInfo += "群创建时间：未知\n"
		groupInfo += "群公告：无\n"
		p.sendMessage(robot, event, groupInfo)

		return nil
	})

	// 处理话唠命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为话唠命令
		if match, _ := p.cmdParser.MatchCommand("话唠|chatty", event.RawMessage); !match {
			return nil
		}

		// 开启话唠模式
		p.sendMessage(robot, event, "话唠模式已开启！我会更积极地回复消息哦！")

		return nil
	})

	// 处理终极命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为终极命令
		if match, _ := p.cmdParser.MatchCommand("终极|ultimate", event.RawMessage); !match {
			return nil
		}

		// 开启终极模式
		p.sendMessage(robot, event, "终极模式已开启！我会释放全部能力！")

		return nil
	})

	// 处理智能体命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为智能体命令
		if match, _ := p.cmdParser.MatchCommand("智能体|agent", event.RawMessage); !match {
			return nil
		}

		// 开启智能体模式
		p.sendMessage(robot, event, "智能体模式已开启！我会更智能地回复消息！")

		return nil
	})
}

// sendMessage 发送消息
func (p *AdminPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
