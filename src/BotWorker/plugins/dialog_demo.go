package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
)

type DialogDemoPlugin struct {
	cmdParser *CommandParser
}

func (p *DialogDemoPlugin) Name() string {
	return "dialog_demo"
}

func (p *DialogDemoPlugin) Description() string {
	return common.T("", "dialog_demo_plugin_desc|对话演示插件，展示多轮对话功能")
}

func (p *DialogDemoPlugin) Version() string {
	return "1.0.0"
}

func NewDialogDemoPlugin() *DialogDemoPlugin {
	return &DialogDemoPlugin{
		cmdParser: NewCommandParser(),
	}
}

// GetSkills 报备插件技能
func (p *DialogDemoPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "start_welcome_dialog",
			Description: common.T("", "dialog_demo_skill_welcome_desc|启动设置欢迎语的多轮对话"),
			Usage:       "start_welcome_dialog",
		},
	}
}

// HandleSkill 实现 SkillCapable 接口
func (p *DialogDemoPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}

	switch skillName {
	case "start_welcome_dialog":
		dialog := StartDialog("set_welcome", event, 5*60*1e9)
		if dialog == nil {
			return "", fmt.Errorf("failed to start dialog")
		}
		menu := common.T("", "dialog_demo_menu|请输入欢迎语模式：\n1. 简洁模式\n2. 详细模式\n3. 随机模式")
		p.sendMessage(robot, event, menu)
		return menu, nil
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

func (p *DialogDemoPlugin) Init(robot plugin.Robot) {
	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 统一处理对话逻辑
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 1. 处理已存在的对话
		dialog := GetDialog(event.GroupID, event.UserID)
		if dialog != nil && dialog.Type == "set_welcome" {
			text := event.RawMessage
			if text == "" {
				if msg, ok := event.Message.(string); ok {
					text = msg
				}
			}

			if text == "" {
				return nil
			}

			if dialog.Step == 1 {
				if text == "1" || text == "2" || text == "3" {
					dialog.Data["mode"] = text
					UpdateDialog(dialog, 2, 5*60*1e9)
					p.sendMessage(robot, event, common.T("", "dialog_demo_input_content|请输入欢迎语内容："))
				} else {
					p.sendMessage(robot, event, common.T("", "dialog_demo_invalid_option|无效的选项，请重新输入 1/2/3："))
				}
				return nil
			}

			if dialog.Step == 2 {
				dialog.Data["text"] = text
				mode := dialog.Data["mode"]
				msg := fmt.Sprintf(common.T("", "dialog_demo_updated|欢迎语设置成功！\n模式：%s\n内容：%s"), mode, text)
				p.sendMessage(robot, event, msg)
				EndDialog(event.GroupID, event.UserID)
				return nil
			}
			return nil
		}

		// 2. 匹配启动对话命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "dialog_demo_cmd_set_welcome|设置欢迎语"), event.RawMessage); match {
			_, err := p.HandleSkill(robot, event, "start_welcome_dialog", nil)
			return err
		}

		return nil
	})
}

func (p *DialogDemoPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if robot == nil || event == nil || msg == "" {
		return
	}
	_, _ = SendTextReply(robot, event, msg)
}

