package main

import (
	"bufio"
	"botworker/internal/plugin"
	"botworker/plugins"
	"fmt"
	log "BotMatrix/common/log"
	"os"
	"strings"
)

type TestRobot struct {
	skills map[string]plugin.Skill
}

func NewTestRobot() *TestRobot {
	return &TestRobot{
		skills: make(map[string]plugin.Skill),
	}
}

func (r *TestRobot) OnMessage(fn func(map[string]any) error) {}
func (r *TestRobot) OnNotice(fn func(map[string]any) error) {}
func (r *TestRobot) OnRequest(fn func(map[string]any) error) {}
func (r *TestRobot) OnEvent(eventName string, fn func(map[string]any) error) {}
func (r *TestRobot) HandleAPI(action string, fn func(map[string]any) (map[string]any, error)) {}
func (r *TestRobot) SendMessage(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) DeleteMessage(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) SendLike(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) SetGroupKick(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) SetGroupBan(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) GetGroupMemberList(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) GetGroupMemberInfo(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) SetGroupSpecialTitle(params map[string]any) (map[string]any, error) { return nil, nil }
func (r *TestRobot) GetSelfID() int64 { return 123456 }
func (r *TestRobot) GetSessionContext(platform, userID string) (map[string]any, error) { return nil, nil }
func (r *TestRobot) SetSessionState(platform, userID string, state map[string]any, ttl int) error { return nil }
func (r *TestRobot) GetSessionState(platform, userID string) (map[string]any, error) { return nil, nil }
func (r *TestRobot) HandleSkill(skillName string, skill plugin.Skill) {
	r.skills[skillName] = skill
}

func main() {
	robot := NewTestRobot()
	pluginManager := plugin.NewManager(robot)
	
	// 加载echo插件
	pluginManager.LoadPlugin(&plugins.EchoPlugin{})
	
	fmt.Println("Simple Plugin Test")
	fmt.Println("==================")
	fmt.Println("Available commands:")
	fmt.Println("  call echo message=<text> - Test echo plugin")
	fmt.Println("  exit - Exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		
		if input == "exit" {
			break
		}
		
		if strings.HasPrefix(input, "call echo") {
			parts := strings.SplitN(input, "=", 2)
			if len(parts) < 2 {
				fmt.Println("Usage: call echo message=<text>")
				continue
			}
			
			message := parts[1]
			params := map[string]string{"message": message}
			
			// 调用echo技能
			if skill, ok := robot.skills["echo"]; ok {
				result, err := skill(params)
				if err != nil {
					log.Printf("Error: %v", err)
				} else {
					log.Printf("Result: %s", result)
				}
			} else {
				fmt.Println("echo skill not found")
			}
			continue
		}
		
		fmt.Println("Unknown command")
	}
	
	fmt.Println("Exiting...")
}