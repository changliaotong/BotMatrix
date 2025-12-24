package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strconv"
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

// GetSkills 报备插件技能
func (p *WelcomePlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "welcome_user",
			Description: "发送欢迎消息给新用户",
			Usage:       "welcome_user user_id=123456 group_id=654321",
			Params: map[string]string{
				"user_id":  "用户QQ号",
				"group_id": "群号（可选，如果不提供则发送私聊欢迎）",
			},
		},
	}
}

// NewWelcomePlugin 创建欢迎插件实例
func NewWelcomePlugin() *WelcomePlugin {
	return &WelcomePlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *WelcomePlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "welcome_plugin_loaded"))

	// 注册技能处理器
	robot.HandleSkill("welcome_user", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		groupID := params["group_id"]
		if userID == "" {
			return "", fmt.Errorf("missing parameter: user_id")
		}

		if groupID != "" {
			return p.doWelcomeGroup(robot, userID, groupID)
		}
		return p.doWelcomeFriend(robot, userID)
	})

	// 处理群成员增加事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_member_increase" {
			log.Printf("新成员加入群 %d: 用户 %d\n", event.GroupID, event.UserID)

			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "welcome") {
				return nil
			}

			p.doWelcomeGroup(robot, fmt.Sprintf("%d", event.UserID), groupIDStr)
		}
		return nil
	})

	// 处理好友添加事件
	robot.OnRequest(func(event *onebot.Event) error {
		if event.RequestType == "friend" && event.Approved {
			log.Printf("新好友添加: 用户 %d\n", event.UserID)

			p.doWelcomeFriend(robot, fmt.Sprintf("%d", event.UserID))
		}
		return nil
	})

	// 处理欢迎命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.RawMessage == "welcome" {
			log.Printf("用户 %d 请求欢迎信息\n", event.UserID)

			welcomeMsg := common.T("", "welcome_cmd_msg")
			p.sendMessage(robot, event, welcomeMsg)
		}

		return nil
	})
}

// doWelcomeGroup 执行群欢迎逻辑
func (p *WelcomePlugin) doWelcomeGroup(robot plugin.Robot, userID string, groupID string) (string, error) {
	welcomeMsg := fmt.Sprintf(common.T("", "welcome_group_new_member"), userID)
	groupIDUint, _ := strconv.ParseUint(groupID, 10, 64)
	params := &onebot.SendMessageParams{
		GroupID: int64(groupIDUint),
		Message: welcomeMsg,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送群欢迎消息失败: %v\n", err)
		return "", err
	}

	log.Printf("已发送群欢迎消息给用户 %s (群 %s)\n", userID, groupID)
	return "已成功发送群欢迎消息", nil
}

// doWelcomeFriend 执行好友欢迎逻辑
func (p *WelcomePlugin) doWelcomeFriend(robot plugin.Robot, userID string) (string, error) {
	welcomeMsg := common.T("", "welcome_friend_new")
	userIDUint, _ := strconv.ParseUint(userID, 10, 64)
	params := &onebot.SendMessageParams{
		UserID:  int64(userIDUint),
		Message: welcomeMsg,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送好友欢迎消息失败: %v\n", err)
		return "", err
	}

	log.Printf("已发送好友欢迎消息给用户 %s\n", userID)
	return "已成功发送好友欢迎消息", nil
}

// sendMessage 发送消息
func (p *WelcomePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		UserID:  event.UserID,
		Message: message,
	}
	if event.MessageType == "group" {
		params.GroupID = event.GroupID
	}
	robot.SendMessage(params)
}
