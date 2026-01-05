package main

import (
	"BotMatrix/common/log"
	"BotMatrix/common/onebot"
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConsoleRobot implements a minimal OneBot client for terminal interaction
type ConsoleRobot struct {
	// OneBot Client fields
	conn     *websocket.Conn
	mu       sync.Mutex
	echoMap  map[string]chan *onebot.Response
	selfID   int64
	platform string
}

func NewConsoleRobot(selfID int64, platform string) *ConsoleRobot {
	return &ConsoleRobot{
		echoMap:  make(map[string]chan *onebot.Response),
		selfID:   selfID,
		platform: platform,
	}
}

func (r *ConsoleRobot) Connect(nexusAddr string) error {
	header := http.Header{}
	header.Add("X-Self-ID", fmt.Sprintf("%d", r.selfID))
	header.Add("X-Platform", r.platform)
	header.Add("X-Client-Role", "Bot")

	conn, _, err := websocket.DefaultDialer.Dial(nexusAddr, header)
	if err != nil {
		return err
	}
	r.conn = conn

	// Handle incoming messages (Events and API Responses)
	go r.readLoop()

	return nil
}

func (r *ConsoleRobot) readLoop() {
	for {
		_, message, err := r.conn.ReadMessage()
		if err != nil {
			log.Errorf("Read error: %v", err)
			return
		}

		// Debug: print raw message from Nexus
		// log.Printf("[ConsoleBot] Received raw from Nexus: %s", string(message))

		// Try parsing as response first (if echo is present and has status)
		var resp onebot.Response
		if err := json.Unmarshal(message, &resp); err == nil && resp.Echo != nil && resp.Status != "" {
			echoStr := fmt.Sprintf("%v", resp.Echo)
			// log.Printf("[ConsoleBot] Received API Response: echo=%s, status=%s", echoStr, resp.Status)
			fmt.Printf("\n[Nexus Response]: %s\n", resp.Status)
			r.mu.Lock()
			if ch, ok := r.echoMap[echoStr]; ok {
				ch <- &resp
				delete(r.echoMap, echoStr)
			}
			r.mu.Unlock()
			continue
		}

		// Try parsing as event
		var event onebot.Event
		if err := json.Unmarshal(message, &event); err == nil && event.PostType != "" {
			// log.Printf("[ConsoleBot] Received Event: post_type=%s, message=%s", event.PostType, event.RawMessage)
			fmt.Printf("\n[Nexus Event] %s: %s\n", event.PostType, event.RawMessage)
			// Handle different post types
			switch event.PostType {
			case "message":
				fmt.Printf("\n[Nexus -> Bot]: %s\n", event.RawMessage)
			case "request":
				fmt.Printf("\n[Nexus -> Bot] Request: %s from %v\n", event.RequestType, event.UserID)
			case "notice":
				fmt.Printf("\n[Nexus -> Bot] Notice: %s\n", event.NoticeType)
			}
			continue
		}

		// Try parsing as API request (from Nexus to Bot)
		var apiReq onebot.Request
		if err := json.Unmarshal(message, &apiReq); err == nil && apiReq.Action != "" {
			// fmt.Printf("\n[Nexus -> Bot Action]: %s, Params: %v\n", apiReq.Action, apiReq.Params)

			msgContent := ""
			// Params is any, need to cast to map
			if params, ok := apiReq.Params.(map[string]any); ok {
				if m, ok := params["message"].(string); ok {
					msgContent = m
				} else if segments, ok := params["message"].([]any); ok {
					// Handle segment list (v12 style or array of v11)
					for _, seg := range segments {
						if sMap, ok := seg.(map[string]any); ok {
							if sType, ok := sMap["type"].(string); ok && sType == "text" {
								if sData, ok := sMap["data"].(map[string]any); ok {
									if text, ok := sData["text"].(string); ok {
										msgContent += text
									}
								}
							}
						}
					}
				}
			}

			if msgContent != "" {
				fmt.Printf("\n🤖 [Bot -> %s]: %s\n", apiReq.Action, msgContent)
				if apiReq.Action == "send_group_msg" && strings.Contains(msgContent, "#确认") {
					parts := strings.Split(msgContent, "#确认 ")
					if len(parts) > 1 {
						idPart := strings.TrimSpace(parts[1])
						// 只取前 16 位字符作为 DraftID
						if len(idPart) >= 16 {
							draftID := idPart[:16]
							go func(id string) {
								time.Sleep(2 * time.Second)
								fmt.Printf("\n[Test] Auto-confirming draft: %s\n", id)
								confirmEvent := &onebot.Event{
									Time:        time.Now().Unix(),
									SelfID:      onebot.FlexibleInt64(r.selfID),
									PostType:    "message",
									MessageType: "group",
									UserID:      onebot.FlexibleInt64(1653346663), // Admin
									GroupID:     onebot.FlexibleInt64(527340256),  // Group
									Message:     "#确认 " + id,
									RawMessage:  "#确认 " + id,
									Platform:    r.platform,
									Sender: onebot.Sender{
										UserID:   onebot.FlexibleInt64(1653346663),
										Nickname: "ConsoleTest",
										Role:     "owner",
									},
								}
								r.SendEvent(confirmEvent)
							}(draftID)
						}
					}
				}
			} else {
				fmt.Printf("\n⚙️ [Bot Action]: %s, Params: %v\n", apiReq.Action, apiReq.Params)
			}

			// Send success response back to Nexus if echo is present
			if apiReq.Echo != nil {
				resp := onebot.Response{
					Status: "ok",
					Data:   map[string]any{"message_id": time.Now().Unix()},
					Echo:   apiReq.Echo,
				}
				r.mu.Lock()
				r.conn.WriteJSON(resp)
				r.mu.Unlock()
			}
			continue
		}

		// Unknown message type
		fmt.Printf("\n❓ [Unknown Msg]: %s\n", string(message))
	}
}

// CallAPI sends an API request to Nexus and waits for response
func (r *ConsoleRobot) CallAPI(action string, params any) (*onebot.Response, error) {
	echo := fmt.Sprintf("%d", time.Now().UnixNano())
	log.Printf("[ConsoleBot] Calling API: action=%s, echo=%s", action, echo)
	req := onebot.Request{
		Action: action,
		Params: params,
		Echo:   echo,
	}

	ch := make(chan *onebot.Response, 1)
	r.mu.Lock()
	r.echoMap[echo] = ch
	r.mu.Unlock()

	r.mu.Lock()
	err := r.conn.WriteJSON(req)
	r.mu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(5 * time.Second):
		r.mu.Lock()
		delete(r.echoMap, echo)
		r.mu.Unlock()
		return nil, fmt.Errorf("API call timeout")
	}
}

func (r *ConsoleRobot) SendEvent(event *onebot.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Printf("[ConsoleBot] Sending Event: post_type=%s, message=%s", event.PostType, event.RawMessage)
	return r.conn.WriteJSON(event)
}

func main() {
	// Configuration
	nexusAddr := "ws://localhost:3001/ws/bots" // Default Nexus Bot WebSocket address
	selfID := int64(51437810)
	platform := "qq"
	userID := int64(1653346663)
	groupID := int64(527340256)

	// Initialize robot (now as a clean OneBot Client)
	robot := NewConsoleRobot(selfID, platform)

	// Connect to Nexus
	fmt.Printf("Connecting to BotNexus at %s...\n", nexusAddr)
	err := robot.Connect(nexusAddr)
	if err != nil {
		fmt.Printf("Failed to connect to BotNexus: %v\n", err)
		fmt.Println("ConsoleBot must connect to Nexus to function.")
		os.Exit(1)
	} else {
		fmt.Println("Connected to BotNexus successfully.")
	}

	fmt.Println("======================================")
	fmt.Println("   BotMatrix Console Robot Project    ")
	fmt.Println("======================================")
	fmt.Println("Type 'help' for available commands.")
	fmt.Println()

	// Auto-send test command for AI task
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("\n[Test] Sending 'AI 随机禁言套餐'...")
		testEvent := &onebot.Event{
			Time:        time.Now().Unix(),
			SelfID:      onebot.FlexibleInt64(selfID),
			PostType:    "message",
			MessageType: "group",
			UserID:      onebot.FlexibleInt64(userID),
			GroupID:     onebot.FlexibleInt64(groupID),
			Message:     "AI 随机禁言套餐",
			RawMessage:  "AI 随机禁言套餐",
			Platform:    platform,
			Sender: onebot.Sender{
				UserID:   onebot.FlexibleInt64(userID),
				Nickname: "ConsoleTest",
				Role:     "owner",
			},
		}
		robot.SendEvent(testEvent)
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if groupID != 0 {
			fmt.Printf("[%d@%d]> ", userID, groupID)
		} else {
			fmt.Printf("[%d]> ", userID)
		}
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
			fmt.Println("  points/积分      - Check points (via PointsSystem plugin)")
			fmt.Println("  activate/激活    - Check/Activate points status")
			fmt.Println("  market/市场      - View market (/market list)")
			fmt.Println("  buy/买入         - Buy group points (/buy Q <amount> <price>)")
			fmt.Println("  sell/卖出        - Sell group points (/sell Q <amount> <price>)")
			fmt.Println("  tip/打赏/transfer/转账 - Tip/Transfer points (/tip @user amount)")
			fmt.Println("  rank/排行        - Show leaderboard")
			fmt.Println("  deposit/存       - Deposit points to bank (/deposit <amount>)")
			fmt.Println("  withdraw/取      - Withdraw points from bank (/withdraw <amount>)")
			fmt.Println("  group_activate/群激活 - Toggle group points mode (Owner only)")
			fmt.Println("  adjust/调整      - Adjust points (Owner only, /adjust @user amount)")
			fmt.Println("  freeze/冻结      - Freeze user points (Owner only, /freeze @user amount)")
			fmt.Println("  unfreeze/解冻    - Unfreeze user points (Owner only, /unfreeze @user amount)")
			fmt.Println("  exit/quit        - Exit")
			fmt.Println("  Anything else    - Sent as a message to Nexus")
			continue
		}

		if trimmed == "points" || trimmed == "积分" {
			trimmed = "积分"
		} else if trimmed == "activate" || trimmed == "激活" {
			trimmed = "激活"
		} else if trimmed == "market" || trimmed == "市场" {
			trimmed = "/market list"
		} else if trimmed == "rank" || trimmed == "排行" {
			trimmed = "/rank"
		} else if trimmed == "group_activate" || trimmed == "群激活" {
			trimmed = "/群激活"
		} else if strings.HasPrefix(trimmed, "tip ") || strings.HasPrefix(trimmed, "打赏 ") || strings.HasPrefix(trimmed, "transfer ") || strings.HasPrefix(trimmed, "转账 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				trimmed = fmt.Sprintf("/tip %s %s", parts[1], parts[2])
				if len(parts) > 3 {
					trimmed += " " + strings.Join(parts[3:], " ")
				}
			}
		} else if strings.HasPrefix(trimmed, "adjust ") || strings.HasPrefix(trimmed, "调整 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				trimmed = fmt.Sprintf("/adjust_points %s %s", parts[1], parts[2])
			}
		} else if strings.HasPrefix(trimmed, "deposit ") || strings.HasPrefix(trimmed, "存 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				trimmed = fmt.Sprintf("/deposit %s", parts[1])
			}
		} else if strings.HasPrefix(trimmed, "withdraw ") || strings.HasPrefix(trimmed, "取 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				trimmed = fmt.Sprintf("/withdraw %s", parts[1])
			}
		} else if strings.HasPrefix(trimmed, "freeze ") || strings.HasPrefix(trimmed, "冻结 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				trimmed = fmt.Sprintf("/freeze %s %s", parts[1], parts[2])
			}
		} else if strings.HasPrefix(trimmed, "unfreeze ") || strings.HasPrefix(trimmed, "解冻 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				trimmed = fmt.Sprintf("/unfreeze %s %s", parts[1], parts[2])
			}
		} else if strings.HasPrefix(trimmed, "buy ") || strings.HasPrefix(trimmed, "买入 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 4 {
				trimmed = fmt.Sprintf("/market buy %s %s %s", parts[1], parts[2], parts[3])
			}
		} else if strings.HasPrefix(trimmed, "sell ") || strings.HasPrefix(trimmed, "卖出 ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 4 {
				trimmed = fmt.Sprintf("/market sell %s %s %s", parts[1], parts[2], parts[3])
			}
		}

		// Send message as an EVENT to Nexus
		event := &onebot.Event{
			Time:        time.Now().Unix(),
			SelfID:      onebot.FlexibleInt64(selfID),
			PostType:    "message",
			MessageType: "group",
			UserID:      onebot.FlexibleInt64(userID),
			GroupID:     onebot.FlexibleInt64(groupID),
			Message:     trimmed,
			RawMessage:  trimmed,
			Platform:    platform,
			Sender: onebot.Sender{
				UserID:   onebot.FlexibleInt64(userID),
				Nickname: "ConsoleUser",
				Role:     "owner",
			},
		}

		if robot.conn != nil {
			err := robot.SendEvent(event)
			if err != nil {
				fmt.Printf("Error sending event: %v\n", err)
			}
		} else {
			fmt.Println("Not connected to Nexus. Message ignored.")
		}
	}

	fmt.Println("Goodbye!")
}
