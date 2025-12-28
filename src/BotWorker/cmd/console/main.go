package main

import (
	"BotMatrix/common"
	"BotMatrix/common/log"
	"bufio"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/plugins"
	"fmt"
	"os"
	"strings"
	"time"
	"encoding/json"
)

// ConsoleRobot implements the plugin.Robot interface for terminal interaction
type ConsoleRobot struct {
	skills   map[string]plugin.Skill
	sessions map[string]*common.SessionContext
	states   map[string]*common.SessionState
}

func NewConsoleRobot() *ConsoleRobot {
	return &ConsoleRobot{
		skills:   make(map[string]plugin.Skill),
		sessions: make(map[string]*common.SessionContext),
		states:   make(map[string]*common.SessionState),
	}
}

// OnMessage implements plugin.Robot
func (r *ConsoleRobot) OnMessage(fn onebot.EventHandler) {}

// OnNotice implements plugin.Robot
func (r *ConsoleRobot) OnNotice(fn onebot.EventHandler) {}

// OnRequest implements plugin.Robot
func (r *ConsoleRobot) OnRequest(fn onebot.EventHandler) {}

// OnEvent implements plugin.Robot
func (r *ConsoleRobot) OnEvent(eventName string, fn onebot.EventHandler) {}

// HandleAPI implements plugin.Robot
func (r *ConsoleRobot) HandleAPI(action string, fn onebot.RequestHandler) {}

// SendMessage implements plugin.Robot
func (r *ConsoleRobot) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: %s\n", params.Message)
	return &onebot.Response{Status: "ok"}, nil
}

// DeleteMessage implements plugin.Robot
func (r *ConsoleRobot) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: (Deleted message %s)\n", params.MessageID)
	return &onebot.Response{Status: "ok"}, nil
}

// SendLike implements plugin.Robot
func (r *ConsoleRobot) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: (Sent like to %d)\n", params.UserID)
	return &onebot.Response{Status: "ok"}, nil
}

// SetGroupKick implements plugin.Robot
func (r *ConsoleRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: (Kicked %d from group %d)\n", params.UserID, params.GroupID)
	return &onebot.Response{Status: "ok"}, nil
}

// SetGroupBan implements plugin.Robot
func (r *ConsoleRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: (Banned %d in group %d for %d seconds)\n", params.UserID, params.GroupID, params.Duration)
	return &onebot.Response{Status: "ok"}, nil
}

// GetGroupMemberList implements plugin.Robot
func (r *ConsoleRobot) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	return &onebot.Response{Status: "ok", Data: []any{}}, nil
}

// GetGroupMemberInfo implements plugin.Robot
func (r *ConsoleRobot) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	return &onebot.Response{Status: "ok", Data: map[string]any{"user_id": params.UserID}}, nil
}

// SetGroupSpecialTitle implements plugin.Robot
func (r *ConsoleRobot) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	fmt.Printf("\n[Robot]: (Set title '%s' for %d in group %d)\n", params.SpecialTitle, params.UserID, params.GroupID)
	return &onebot.Response{Status: "ok"}, nil
}

// GetSelfID implements plugin.Robot
func (r *ConsoleRobot) GetSelfID() int64 {
	return 888888
}

// Session & State Management
func (r *ConsoleRobot) GetSessionContext(platform, userID string) (*common.SessionContext, error) {
	key := fmt.Sprintf("%s:%s", platform, userID)
	if ctx, ok := r.sessions[key]; ok {
		return ctx, nil
	}
	return &common.SessionContext{
		Platform:  platform,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *ConsoleRobot) SetSessionState(platform, userID string, state common.SessionState, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", platform, userID)
	state.UpdatedAt = time.Now()
	r.states[key] = &state
	return nil
}

func (r *ConsoleRobot) GetSessionState(platform, userID string) (*common.SessionState, error) {
	key := fmt.Sprintf("%s:%s", platform, userID)
	if state, ok := r.states[key]; ok {
		return state, nil
	}
	return nil, nil
}

func (r *ConsoleRobot) ClearSessionState(platform, userID string) error {
	key := fmt.Sprintf("%s:%s", platform, userID)
	delete(r.states, key)
	return nil
}

// HandleSkill implements plugin.Robot
func (r *ConsoleRobot) HandleSkill(skillName string, skill func(params map[string]string) (string, error)) {
	r.skills[skillName] = skill
}

func (r *ConsoleRobot) CallSkill(skillName string, params map[string]string) (string, error) {
	skill, ok := r.skills[skillName]
	if !ok {
		return "", fmt.Errorf("skill %s not found", skillName)
	}
	return skill(params)
}

func main() {
	// Initialize robot
	robot := NewConsoleRobot()

	// Initialize plugin manager
	pluginManager := plugin.NewManager(robot)

	// Load plugins
	pluginManager.LoadPlugin(&plugins.EchoPlugin{})
	pluginManager.LoadPlugin(plugins.NewTimePlugin())
	// WeatherPlugin might need an API key, but we'll load it anyway
	pluginManager.LoadPlugin(plugins.NewWeatherPlugin(nil))
	pluginManager.LoadPlugin(&plugins.WelcomePlugin{})

	fmt.Println("======================================")
	fmt.Println("   BotMatrix Console Robot Project    ")
	fmt.Println("======================================")
	fmt.Println("Type 'help' for available commands.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	platform := "console"
	userID := "user123"

	for {
		fmt.Printf("[%s@%s]> ", userID, platform)
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		trimmed := strings.TrimSpace(input)

		if trimmed == "" {
			continue
		}

		if trimmed == "exit" || trimmed == "quit" {
			break
		}

		if trimmed == "help" {
			fmt.Println("Commands:")
			fmt.Println("  help             - Show this help")
			fmt.Println("  list             - List loaded plugins")
			fmt.Println("  skills           - List available skills")
			fmt.Println("  call <skill> KV  - Call a skill (e.g., call echo message=hello)")
			fmt.Println("  session          - Show current session info")
			fmt.Println("  state            - Show current session state")
			fmt.Println("  clear            - Clear session state")
			fmt.Println("  exit             - Exit")
			continue
		}

		if trimmed == "list" {
			ps := pluginManager.GetPlugins()
			fmt.Printf("Loaded plugins (%d):\n", len(ps))
			for _, p := range ps {
				fmt.Printf("  - %s: %s (v%s)\n", p.Name(), p.Description(), p.Version())
			}
			continue
		}

		if trimmed == "skills" {
			fmt.Printf("Available skills (%d):\n", len(robot.skills))
			for name := range robot.skills {
				fmt.Printf("  - %s\n", name)
			}
			continue
		}

		if trimmed == "session" {
			ctx, _ := robot.GetSessionContext(platform, userID)
			data, _ := json.MarshalIndent(ctx, "", "  ")
			fmt.Printf("Current Session:\n%s\n", string(data))
			continue
		}

		if trimmed == "state" {
			state, _ := robot.GetSessionState(platform, userID)
			if state == nil {
				fmt.Println("No active session state.")
			} else {
				data, _ := json.MarshalIndent(state, "", "  ")
				fmt.Printf("Current State:\n%s\n", string(data))
			}
			continue
		}

		if trimmed == "clear" {
			robot.ClearSessionState(platform, userID)
			fmt.Println("Session state cleared.")
			continue
		}

		if strings.HasPrefix(trimmed, "call ") {
			parts := strings.SplitN(trimmed[5:], " ", 2)
			skillName := parts[0]
			params := make(map[string]string)

			if len(parts) > 1 {
				paramParts := strings.Split(parts[1], " ")
				for _, p := range paramParts {
					kv := strings.SplitN(p, "=", 2)
					if len(kv) == 2 {
						params[kv[0]] = kv[1]
					}
				}
			}

			result, err := robot.CallSkill(skillName, params)
			if err != nil {
				log.Errorf("Skill Error: %v", err)
			} else {
				fmt.Printf("Result: %s\n", result)
			}
			continue
		}

		// Default: try to treat as a message if no command matches
		fmt.Printf("Unknown command or message: %s\n", trimmed)
		fmt.Println("To call a skill, use: call <skill_name> key=value")
	}

	fmt.Println("Goodbye!")
}
