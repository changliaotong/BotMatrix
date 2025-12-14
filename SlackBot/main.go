package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Config struct {
	BotToken  string `json:"bot_token"` // xoxb-...
	AppToken  string `json:"app_token"` // xapp-...
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	api       *slack.Client
	client    *socketmode.Client
	nexusConn *websocket.Conn
	selfID    string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	// Initialize Slack Client
	api = slack.New(
		config.BotToken,
		slack.OptionAppLevelToken(config.AppToken),
	)

	client = socketmode.New(
		api,
	)

	// Get Bot Info
	authTest, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Slack Auth failed: %v", err)
	}
	selfID = authTest.BotID
	log.Printf("Slack Bot Authorized: %s (ID: %s, User: %s)", authTest.User, selfID, authTest.UserID)

	// Connect to BotNexus
	go connectToNexus()

	// Start Socket Mode
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack Socket Mode...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack Socket Mode!")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					continue
				}
				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						handleMessage(ev)
					}
				}
			}
		}
	}()

	err = client.Run()
	if err != nil {
		log.Fatalf("Socket Mode failed: %v", err)
	}
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	}

	if envBot := os.Getenv("SLACK_BOT_TOKEN"); envBot != "" {
		config.BotToken = envBot
	}
	if envApp := os.Getenv("SLACK_APP_TOKEN"); envApp != "" {
		config.AppToken = envApp
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.BotToken == "" || config.AppToken == "" {
		log.Fatal("SLACK_BOT_TOKEN and SLACK_APP_TOKEN are required")
	}
}

func handleMessage(ev *slackevents.MessageEvent) {
	// Ignore bot messages
	if ev.BotID != "" && ev.BotID == selfID {
		return
	}
	// Also ignore if subtype is not empty (like message_changed, etc for now) unless we want to handle edits
	if ev.SubType != "" {
		// return // optional: strict handling
	}

	log.Printf("[%s] %s", ev.User, ev.Text)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  ev.ClientMsgID, // or ev.Ts
		"user_id":     ev.User,
		"message":     ev.Text,
		"raw_message": ev.Text,
		"sender": map[string]interface{}{
			"user_id":  ev.User,
			"nickname": "SlackUser", // We could fetch user info, but let's save API calls
		},
	}

	// Handle Channel vs DM
	// Slack Channel IDs start with C (Channel), D (Direct Message), G (Group)
	if strings.HasPrefix(ev.Channel, "D") {
		obMsg["message_type"] = "private"
	} else {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = ev.Channel
	}

	sendToNexus(obMsg)
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", selfID)
		header.Add("X-Platform", "Slack")

		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, header)
		if err != nil {
			log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		nexusConn = conn
		log.Println("Connected to BotNexus!")

		sendToNexus(map[string]interface{}{
			"post_type":       "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         selfID,
			"time":            time.Now().Unix(),
		})

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("BotNexus disconnected: %v", err)
				break
			}
			handleNexusCommand(message)
		}
		time.Sleep(1 * time.Second)
	}
}

func sendToNexus(msg interface{}) {
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send to Nexus: %v", err)
	}
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
		Echo   string                 `json:"echo"`
	}
	if err := json.Unmarshal(data, &cmd); err != nil {
		return
	}

	log.Printf("Received Command: %s", cmd.Action)

	switch cmd.Action {
	case "send_group_msg", "send_msg":
		channelID, _ := cmd.Params["group_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if channelID != "" && text != "" {
			sendSlackMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			// In Slack, sending to UserID directly usually works if it's a DM ID.
			// If it's a User ID (U...), we might need OpenConversation first.
			// Try sending directly, Slack API is smart.
			sendSlackMessage(userID, text, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": "SlackBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendSlackMessage(channelID, text, echo string) {
	// Convert CQ codes if necessary (basic image support)
	// [CQ:image,file=http://...] -> blocks?
	// For now, simple text.

	_, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
	)

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s", channelID)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": timestamp},
		"echo":   echo,
	})
}
