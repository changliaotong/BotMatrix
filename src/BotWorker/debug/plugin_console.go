package main

import (
	"BotMatrix/common/log"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/types"
	"botworker/internal/onebot"
	"botworker/plugins"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type TestRobot struct {
	plugins []core.PluginModule
	skills  map[string]core.Skill
}

func NewTestRobot() *TestRobot {
	return &TestRobot{
		skills: make(map[string]core.Skill),
	}
}

func (r *TestRobot) OnMessage(fn func(event map[string]any)) {
	// 测试环境不实现消息监听
}

func (r *TestRobot) OnNotice(fn func(event map[string]any)) {
	// 测试环境不实现通知监听
}

func (r *TestRobot) OnRequest(fn func(event map[string]any)) {
	// 测试环境不实现请求监听
}

func (r *TestRobot) OnEvent(eventName string, fn func(event map[string]any)) {
	// 测试环境不实现事件监听
}

func (r *TestRobot) HandleAPI(action string, fn any) {
	// 测试环境不实现API处理
}

func (r *TestRobot) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	// 测试环境不实现消息发送
	return nil, nil
}

func (r *TestRobot) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	// 测试环境不实现消息删除
	return nil, nil
}

func (r *TestRobot) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	// 测试环境不实现点赞
	return nil, nil
}

func (r *TestRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	// 测试环境不实现踢人
	return nil, nil
}

func (r *TestRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	// 测试环境不实现禁言
	return nil, nil
}

func (r *TestRobot) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	// 测试环境不实现成员列表获取
	return nil, nil
}

func (r *TestRobot) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	// 测试环境不实现成员信息获取
	return nil, nil
}

func (r *TestRobot) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	// 测试环境不实现设置专属头衔
	return nil, nil
}

func (r *TestRobot) GetSelfID() int64 {
	return 123456
}

func (r *TestRobot) GetSessionContext(platform, userID string) (*types.SessionContext, error) {
	return nil, nil
}

func (r *TestRobot) SetSessionState(platform, userID string, state types.SessionState, ttl time.Duration) error {
	return nil
}

func (r *TestRobot) GetSessionState(platform, userID string) (*types.SessionState, error) {
	return nil, nil
}

func (r *TestRobot) ClearSessionState(platform, userID string) error {
	return nil
}

// HandleSkill implements plugin.Robot
func (r *TestRobot) HandleSkill(skillName string, skill func(ctx core.BaseContext, params map[string]string) (string, error)) {
	r.skills[skillName] = skill
}

func (r *TestRobot) RegisterSkill(capability core.SkillCapability, skill func(ctx core.BaseContext, params map[string]string) (string, error)) {
	r.skills[capability.Name] = skill
}

func (r *TestRobot) CallSkill(skillName string, params map[string]string) (string, error) {
	skill, ok := r.skills[skillName]
	if !ok {
		return "", fmt.Errorf("skill %s not found", skillName)
	}
	return skill(nil, params)
}

func (r *TestRobot) CallPluginAction(pluginID string, action string, payload map[string]any) (any, error) {
	// 测试环境不实现插件动作调用
	return nil, nil
}

func (r *TestRobot) CallBotAction(action string, params any) (any, error) {
	// 测试环境不实现机器人动作调用
	return nil, nil
}

func main() {
	// 初始化测试机器人
	robot := NewTestRobot()

	// 加载所有插件
	pm := core.NewPluginManager()

	// 加载 PointsProxy (模拟内部插件)
	pointsProxy := &plugins.PointsProxy{}
	pm.LoadPluginModule(pointsProxy, robot)

	fmt.Println("Plugin Console")
	fmt.Println("===================")
	fmt.Println("Available commands:")
	fmt.Println("  list - List all loaded plugins")
	fmt.Println("  skills - List all available skills")
	fmt.Println("  call <skill> [params] - Call a skill with parameters")
	fmt.Println("  exit - Exit the console")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		if input == "list" {
			// 内部插件
			internalPlugins := pm.GetInternalPlugins()
			log.Printf("Loaded internal plugins (%d):", len(internalPlugins))
			for name, p := range internalPlugins {
				log.Printf("  - %s: %s (v%s)", name, p.Description(), p.Version())
			}
			continue
		}

		if input == "skills" {
			log.Printf("Available skills (%d):", len(robot.skills))
			for skillName := range robot.skills {
				log.Printf("  - %s", skillName)
			}
			continue
		}

		if strings.HasPrefix(input, "call ") {
			parts := strings.SplitN(input[5:], " ", 2)
			if len(parts) < 1 {
				fmt.Println("Usage: call <skill> [params]")
				continue
			}

			skillName := parts[0]
			params := make(map[string]string)

			if len(parts) > 1 {
				paramParts := strings.Split(parts[1], " ")
				for _, param := range paramParts {
					kv := strings.SplitN(param, "=", 2)
					if len(kv) == 2 {
						params[kv[0]] = kv[1]
					}
				}
			}

			result, err := robot.CallSkill(skillName, params)
			if err != nil {
				log.Printf("Error calling skill %s: %v", skillName, err)
			} else {
				log.Printf("Result: %s", result)
			}
			continue
		}

		fmt.Println("Unknown command. Type 'list', 'skills', 'call', or 'exit'.")
	}

	fmt.Println("Exiting...")
}
