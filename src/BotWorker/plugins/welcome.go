package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "welcome_plugin_desc")
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
	log.Println(common.T("", "welcome_plugin_loaded"))

	// 处理群成员增加事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_member_increase" {
			log.Printf("新成员加入群 %d: 用户 %d\n", event.GroupID, event.UserID)

			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "welcome") {
				return nil
			}

			// 发送群欢迎消息
			welcomeMsg := fmt.Sprintf(common.T("", "welcome_group_new_member"), event.UserID)

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
			welcomeMsg := common.T("", "welcome_friend_new")

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

			welcomeMsg := common.T("", "welcome_cmd_msg")

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
