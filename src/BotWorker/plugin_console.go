package main

import (
	"bufio"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/plugins"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type TestRobot struct {
	plugins []plugin.Plugin
	skills  map[string]plugin.Skill
}

func NewTestRobot() *TestRobot {
	return &TestRobot{
		skills: make(map[string]plugin.Skill),
	}
}

func (r *TestRobot) OnMessage(fn onebot.EventHandler) {
	// 测试环境不实现消息监听
}

func (r *TestRobot) OnNotice(fn onebot.EventHandler) {
	// 测试环境不实现通知监听
}

func (r *TestRobot) OnRequest(fn onebot.EventHandler) {
	// 测试环境不实现请求监听
}

func (r *TestRobot) OnEvent(eventName string, fn onebot.EventHandler) {
	// 测试环境不实现事件监听
}

func (r *TestRobot) HandleAPI(action string, fn onebot.RequestHandler) {
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

func (r *TestRobot) GetSessionContext(platform, userID string) (map[string]interface{}, error) {
	return nil, nil
}

func (r *TestRobot) SetSessionState(platform, userID string, state map[string]interface{}, ttl time.Duration) error {
	return nil
}

func (r *TestRobot) GetSessionState(platform, userID string) (map[string]interface{}, error) {
	return nil, nil
}

func (r *TestRobot) HandleSkill(skillName string, skill plugin.Skill) {
	r.skills[skillName] = skill
}

func (r *TestRobot) CallSkill(skillName string, params map[string]string) (string, error) {
	skill, ok := r.skills[skillName]
	if !ok {
		return "", fmt.Errorf("skill %s not found", skillName)
	}
	return skill(params)
}

func main() {
	// 初始化测试机器人
	robot := NewTestRobot()

	// 加载所有插件
	pluginManager := plugin.NewManager(robot)

	// 注册插件
	pluginManager.LoadPlugin(&plugins.EchoPlugin{})
	pluginManager.LoadPlugin(&plugins.TimePlugin{})
	pluginManager.LoadPlugin(&plugins.WeatherPlugin{})
	pluginManager.LoadPlugin(&plugins.SystemInfoPlugin{})

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
			plugins := pluginManager.GetPlugins()
			fmt.Printf("Loaded plugins (%d):\n", len(plugins))
			for _, p := range plugins {
				fmt.Printf("  - %s: %s (v%s)\n", p.Name(), p.Description(), p.Version())
			}
			continue
		}

		if input == "skills" {
			fmt.Printf("Available skills (%d):\n", len(robot.skills))
			for skillName := range robot.skills {
				fmt.Printf("  - %s\n", skillName)
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
				fmt.Printf("Error calling skill %s: %v\n", skillName, err)
			} else {
				fmt.Printf("Result: %s\n", result)
			}
			continue
		}

		fmt.Println("Unknown command. Type 'list', 'skills', 'call', or 'exit'.")
	}

	fmt.Println("Exiting...")
}
