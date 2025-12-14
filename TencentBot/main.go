package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	botws "github.com/tencent-connect/botgo/websocket"
)

// Config holds the configuration
type Config struct {
	AppID     uint64 `json:"app_id"` // AppID is uint64 in SDK
	Token     string `json:"token"`
	Secret    string `json:"secret"`
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	nexusConn *websocket.Conn
	api       openapi.OpenAPI
	ctx       context.Context
)

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Println("config.json not found, creating from sample...")
		// Create a dummy config if not exists, but better to just fail or use env
		if os.IsNotExist(err) {
			sampleData, _ := os.ReadFile("config.sample.json")
			os.WriteFile("config.json", sampleData, 0644)
			log.Fatal("Please edit config.json and restart.")
		}
		log.Fatal("Error reading config:", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Error decoding config:", err)
	}
}

// NexusConnect connects to BotNexus
func NexusConnect() {
	headers := http.Header{}
	headers.Add("X-Self-ID", fmt.Sprintf("%d", config.AppID))
	headers.Add("X-Platform", "QQOfficial")

	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, headers)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		nexusConn = conn
		log.Println("Connected to BotNexus!")

		// Handle incoming messages from BotNexus (Actions)
		go handleNexusMessages()

		return
	}
}

func handleNexusMessages() {
	defer nexusConn.Close()
	for {
		_, message, err := nexusConn.ReadMessage()
		if err != nil {
			log.Println("BotNexus connection lost:", err)
			// Reconnect logic could be here, but for now we just exit/restart
			os.Exit(1)
			return
		}

		var actionMap map[string]interface{}
		if err := json.Unmarshal(message, &actionMap); err != nil {
			log.Println("Error parsing action:", err)
			continue
		}

		// Handle Actions (e.g. send_msg)
		// This is where we translate OneBot actions to Tencent SDK calls
		handleAction(actionMap)
	}
}

func handleAction(action map[string]interface{}) {
	act, ok := action["action"].(string)
	if !ok {
		return
	}

	log.Printf("Received action: %s", act)

	switch act {
	case "send_msg":
		// Example: Send message
		// Params: channel_id (group_id), content
		params, _ := action["params"].(map[string]interface{})
		channelID, _ := params["group_id"].(string) // Map group_id to channel_id
		if channelID == "" {
			// Try user_id as channel_id for direct messages?
			// In Guild, it's mostly channel_id
			channelID, _ = params["user_id"].(string)
		}

		content, _ := params["message"].(string)
		if content == "" {
			content = "Empty message"
		}

		if channelID != "" {
			// Call SDK
			_, err := api.PostMessage(ctx, channelID, &dto.MessageToCreate{
				Content: content,
			})
			if err != nil {
				log.Println("Error sending message:", err)
			}
		}

	case "get_login_info":
		// Return bot info
		me, err := api.Me(ctx)
		if err == nil {
			resp := map[string]interface{}{
				"status": "ok",
				"data": map[string]interface{}{
					"user_id":  me.ID,
					"nickname": me.Username,
				},
				"echo": action["echo"],
			}
			sendToNexus(resp)
		}
	}
}

func sendToNexus(data interface{}) {
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(data); err != nil {
		log.Println("Error sending to Nexus:", err)
	}
}

// Event Handlers

func atMessageEventHandler(event *dto.WSPayload, data *dto.WSATMessageData) error {
	log.Printf("Received AT Message from %s: %s", data.Author.Username, data.Content)

	// Translate to OneBot Message Event
	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "group", // Guild messages are like group messages
		"sub_type":     "normal",
		"message_id":   data.ID,
		"user_id":      data.Author.ID, // String ID
		"group_id":     data.ChannelID, // Use ChannelID as GroupID
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": data.Author.Username,
		},
		"time":    time.Now().Unix(),
		"self_id": fmt.Sprintf("%d", config.AppID),
	}

	sendToNexus(obEvent)
	return nil
}

func main() {
	loadConfig()
	ctx = context.Background()

	// Initialize Bot Token
	botToken := token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     fmt.Sprintf("%d", config.AppID),
			AppSecret: config.Secret,
		},
	)

	// Initialize API
	api = botgo.NewOpenAPI(fmt.Sprintf("%d", config.AppID), botToken).WithTimeout(3 * time.Second)

	// Connect to BotNexus
	go NexusConnect()

	// Connect to Tencent WebSocket
	wsInfo, err := api.WS(ctx, nil, "")
	if err != nil {
		log.Fatal("Error getting WS info:", err)
	}

	intent := botws.RegisterHandlers(
		// Register handlers
		atMessageEventHandler,
		// Add more handlers as needed
	)

	log.Println("Starting Tencent Bot Session Manager...")
	if err := botgo.NewSessionManager().Start(wsInfo, botToken, &intent); err != nil {
		log.Fatal(err)
	}
}
