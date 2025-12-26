package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Skill func(params map[string]string) (string, error)

type Plugin interface {
	Name() string
	Description() string
	Version() string
	Init(robot Robot)
}

type Robot interface {
	HandleSkill(skillName string, skill Skill)
}

type TestRobot struct {
	skills map[string]Skill
}

func NewTestRobot() *TestRobot {
	return &TestRobot{
		skills: make(map[string]Skill),
	}
}

func (r *TestRobot) HandleSkill(skillName string, skill Skill) {
	r.skills[skillName] = skill
}

// EchoPlugin 实现了一个简单的echo插件
type EchoPlugin struct {}

func (p *EchoPlugin) Name() string {
	return "echo"
}

func (p *EchoPlugin) Description() string {
	return "复读插件，可以重复用户说的话"
}

func (p *EchoPlugin) Version() string {
	return "1.0.0"
}

func (p *EchoPlugin) Init(robot Robot) {
	robot.HandleSkill("echo", func(params map[string]string) (string, error) {
		message, ok := params["message"]
		if !ok || message == "" {
			return "", fmt.Errorf("missing message parameter")
		}
		return message, nil
	})
}

func main() {
	robot := NewTestRobot()
	
	// 加载echo插件
	echoPlugin := &EchoPlugin{}
	echoPlugin.Init(robot)
	
	fmt.Println("Plugin Test Console")
	fmt.Println("===================")
	fmt.Println("Available commands:")
	fmt.Println("  call echo message=<text> - Test echo plugin")
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
		
		if strings.HasPrefix(input, "call echo") {
			parts := strings.SplitN(input, "=", 2)
			if len(parts) < 2 {
				fmt.Println("Usage: call echo message=<text>")
				continue
			}
			
			message := parts[1]
			params := map[string]string{"message": message}
			
			if skill, ok := robot.skills["echo"]; ok {
				result, err := skill(params)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Result: %s\n", result)
				}
			} else {
				fmt.Println("echo skill not found")
			}
			continue
		}
		
		fmt.Println("Unknown command. Type 'call echo message=<text>' or 'exit'.")
	}
	
	fmt.Println("Exiting...")
}