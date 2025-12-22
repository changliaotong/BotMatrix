package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
)

type WelcomePlugin struct {
	// 命令解析器
	cmdParser *CommandParser
}

func (p *WelcomePlugin) Name() string {
	return "welcome"
}

func (p *WelcomePlugin) Description() string {
	return "发送欢迎语插件，在新成员入群和添加好友时发送欢迎消息"
}

func (p *WelcomePlugin) Version() string {
	return "1.0.0"
}

// NewWelcomePlugin 创建欢迎插件实例
func NewWelcomePlugin() *WelcomePlugin {
	return &WelcomePlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *WelcomePlugin) Init(robot plugin.Robot) {
	log.Println("加载欢迎语插件")

	// 处理群成员增加事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_member_increase" {
			log.Printf("新成员加入群 %d: 用户 %d\n", event.GroupID, event.UserID)

			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "welcome") {
				return nil
			}

			// 发送群欢迎消息
			welcomeMsg := fmt.Sprintf("欢迎新成员 @%d 加入本群！\n请遵守群规，文明交流。", event.UserID)

			params := &onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: welcomeMsg,
			}

			if _, err := robot.SendMessage(params); err != nil {
				log.Printf("发送群欢迎消息失败: %v\n", err)
				return err
			}

			log.Printf("已发送群欢迎消息给用户 %d\n", event.UserID)
		}
		return nil
	})

	// 处理好友添加事件
	robot.OnRequest(func(event *onebot.Event) error {
		if event.RequestType == "friend" && event.Approved {
			log.Printf("新好友添加: 用户 %d\n", event.UserID)

			// 发送好友欢迎消息
			welcomeMsg := fmt.Sprintf("你好！我是BotWorker机器人。\n\n我可以为你提供以下服务：\n- 发送消息\n- 管理群聊\n- 执行命令\n\n输入 'help' 查看更多帮助信息。")

			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: welcomeMsg,
			}

			if _, err := robot.SendMessage(params); err != nil {
				log.Printf("发送好友欢迎消息失败: %v\n", err)
				return err
			}

			log.Printf("已发送好友欢迎消息给用户 %d\n", event.UserID)
		}
		return nil
	})

	// 处理欢迎命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.RawMessage == "welcome" {
			log.Printf("用户 %d 请求欢迎信息\n", event.UserID)

			welcomeMsg := "欢迎使用BotWorker机器人！\n\n这是一个基于OneBot协议的机器人处理程序，支持多种插件扩展。\n\n当前加载的插件：\n- echo: 简单的回声插件\n- welcome: 发送欢迎语插件\n\n输入 'help' 查看更多帮助信息。"

			params := &onebot.SendMessageParams{
				UserID:  event.UserID,
				Message: welcomeMsg,
			}

			if _, err := robot.SendMessage(params); err != nil {
				log.Printf("发送欢迎命令回复失败: %v\n", err)
				return err
			}
		}
		return nil
	})
}
