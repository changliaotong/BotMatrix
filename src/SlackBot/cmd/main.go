package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"BotMatrix/common/bot"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackConfig extends bot.BotConfig with Slack specific fields
type SlackConfig struct {
	bot.BotConfig
	AppToken string `json:"app_token"` // xapp-...
}

var (
	botService *bot.BaseBot
	api        *slack.Client
	client     *socketmode.Client
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
	slackCfg   SlackConfig
)

func main() {
	botService = bot.NewBaseBot(8086)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("SlackBot", &slackCfg, restartBot, []bot.ConfigSection{
		{
			Title: "Slack API 配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Token (xoxb-...)", ID: "bot_token", Type: "password", Value: slackCfg.BotToken},
				{Label: "App Token (xapp-...)", ID: "app_token", Type: "password", Value: slackCfg.AppToken},
			},
		},
		{
			Title: "连接与服务配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: slackCfg.NexusAddr},
				{Label: "Web UI 端口 (LogPort)", ID: "log_port", Type: "number", Value: slackCfg.LogPort},
			},
		},
	})

	go botService.StartHTTPServer()

	restartBot()

	botService.WaitExitSignal()
	stopBot()
}

func loadConfig() {
	botService.LoadConfig("config.json")

	// Sync common config to local slackCfg
	botService.Mu.RLock()
	slackCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Slack specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &slackCfg)
	}

	// Environment variable overrides
	if envAppToken := os.Getenv("SLACK_APP_TOKEN"); envAppToken != "" {
		slackCfg.AppToken = envAppToken
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Sync botService.Config from slackCfg
	botService.Config = slackCfg.BotConfig
	botToken := slackCfg.BotToken
	appToken := slackCfg.AppToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if botToken == "" || appToken == "" {
		log.Println("WARNING: Slack BotToken or AppToken is not configured. Bot will not start until configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	// Initialize Slack Client
	api = slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	client = socketmode.New(api)

	// Get Bot Info
	authTest, err := api.AuthTest()
	if err != nil {
		log.Printf("Slack Auth failed: %v", err)
		return
	}
	selfID = authTest.BotID
	log.Printf("Slack Bot Authorized: %s (ID: %s, User: %s)", authTest.User, selfID, authTest.UserID)

	// Connect to BotNexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Slack", selfID, handleNexusCommand)

	// Start Socket Mode
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-client.Events:
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
		}
	}(botCtx)

	go func() {
		if err := client.Run(); err != nil {
			log.Printf("Socket Mode failed: %v", err)
		}
	}()
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func handleMessage(ev *slackevents.MessageEvent) {
	if ev.BotID != "" && ev.BotID == selfID {
		return
	}

	log.Printf("[%s] %s", ev.User, ev.Text)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  ev.ClientMsgID,
		"user_id":     ev.User,
		"message":     ev.Text,
		"raw_message": ev.Text,
		"sender": map[string]any{
			"user_id":  ev.User,
			"nickname": "SlackUser",
		},
	}

	if strings.HasPrefix(ev.Channel, "D") {
		obMsg["message_type"] = "private"
	} else {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = ev.Channel
	}

	botService.SendToNexus(obMsg)
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string                 `json:"action"`
		Params map[string]any         `json:"params"`
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
			sendSlackMessage(userID, text, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteSlackMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		botService.SendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  selfID,
				"nickname": "SlackBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendSlackMessage(channelID, text, echo string) {
	if api == nil {
		return
	}
	_, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
	)

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s", channelID)
	compositeID := fmt.Sprintf("%s:%s", channelID, timestamp)
	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteSlackMessage(compositeID, echo string) {
	if api == nil {
		return
	}
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		botService.SendToNexus(map[string]any{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}
	channelID := parts[0]
	timestamp := parts[1]

	_, _, err := api.DeleteMessage(channelID, timestamp)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s in channel %s", timestamp, channelID)
	botService.SendToNexus(map[string]any{"status": "ok", "echo": echo})
}
