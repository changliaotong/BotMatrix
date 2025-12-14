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
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// Config holds the bot configuration
type Config struct {
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	EncryptKey        string `json:"encrypt_key"`
	VerificationToken string `json:"verification_token"`
	NexusAddr         string `json:"nexus_addr"`
}

var (
	config     Config
	larkClient *lark.Client
	nexusConn  *websocket.Conn
	selfID     string // AppID as SelfID
)

// OneBotMessage represents a OneBot 11 message
type OneBotMessage struct {
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type"`
	Time        int64       `json:"time"`
	SelfID      string      `json:"self_id"` // OneBot uses int64 usually, but string is safer for AppID
	SubType     string      `json:"sub_type"`
	MessageID   string      `json:"message_id"`
	UserID      string      `json:"user_id"`
	GroupID     string      `json:"group_id,omitempty"`
	Message     interface{} `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Sender      Sender      `json:"sender"`
}

type Sender struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

func main() {
	// Load Configuration
	loadConfig()
	selfID = config.AppID

	// Initialize Lark API Client
	larkClient = lark.NewClient(config.AppID, config.AppSecret)

	// Start BotNexus Connection
	go connectToNexus()

	// Start Lark WebSocket Client (Receive Events)
	startLarkWS()
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Println("config.json not found, using environment variables or defaults")
		config = Config{
			AppID:             os.Getenv("APP_ID"),
			AppSecret:         os.Getenv("APP_SECRET"),
			NexusAddr:         os.Getenv("NEXUS_ADDR"),
			EncryptKey:        os.Getenv("ENCRYPT_KEY"),
			VerificationToken: os.Getenv("VERIFICATION_TOKEN"),
		}
	} else {
		if err := json.Unmarshal(file, &config); err != nil {
			log.Fatalf("Error parsing config.json: %v", err)
		}
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.AppID == "" || config.AppSecret == "" {
		log.Fatal("AppID and AppSecret are required")
	}
}

func startLarkWS() {
	// Register Event Handler
	eventHandler := larkevent.NewEventDispatcher(config.VerificationToken, config.EncryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			handleMessage(ctx, event)
			return nil
		})

	// Create WebSocket Client
	cli := larkws.NewClient(config.AppID, config.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// Start WebSocket Client
	log.Println("Connecting to Feishu/Lark Gateway...")
	err := cli.Start(context.Background())
	if err != nil {
		log.Fatalf("Failed to start Feishu WebSocket client: %v", err)
	}
}

func handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msg := event.Event.Message
	sender := event.Event.Sender

	// Extract Content (JSON string)
	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(*msg.Content), &contentMap); err != nil {
		log.Printf("Error parsing message content: %v", err)
		return
	}

	// Message Content Parsing
	text := ""
	switch *msg.MessageType {
	case "text":
		if t, ok := contentMap["text"].(string); ok {
			text = t
		}
	case "image":
		if key, ok := contentMap["image_key"].(string); ok {
			text = fmt.Sprintf("[CQ:image,file=%s]", key)
		}
	case "file":
		if key, ok := contentMap["file_key"].(string); ok {
			text = fmt.Sprintf("[CQ:file,file=%s]", key)
		}
	case "audio":
		if key, ok := contentMap["file_key"].(string); ok {
			text = fmt.Sprintf("[CQ:record,file=%s]", key)
		}
	case "post":
		// Simplified rich text handling
		text = "[RichText Message]"
		if title, ok := contentMap["title"].(string); ok {
			text += " " + title
		}
	default:
		text = fmt.Sprintf("[Unknown Type: %s]", *msg.MessageType)
	}

	// Determine Chat Type (group/p2p)
	msgType := "private"
	groupID := ""
	if *msg.ChatType == "group" {
		msgType = "group"
		groupID = *msg.ChatId
	}

	// Determine User ID (prefer open_id)
	userID := *sender.SenderId.OpenId

	log.Printf("Received Message: [%s] %s: %s", msgType, userID, text)

	// Construct OneBot Message
	obMsg := OneBotMessage{
		PostType:    "message",
		MessageType: msgType,
		Time:        time.Now().Unix(),
		SelfID:      selfID,
		SubType:     "normal",
		MessageID:   *msg.MessageId,
		UserID:      userID,
		GroupID:     groupID,
		Message:     text,
		RawMessage:  text,
		Sender: Sender{
			UserID:   userID,
			Nickname: "FeishuUser", // Feishu doesn't provide nickname in message event directly usually
		},
	}

	sendToNexus(obMsg)
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", selfID)
		header.Add("X-Platform", "Feishu")

		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, header)
		if err != nil {
			log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		nexusConn = conn
		log.Println("Connected to BotNexus!")

		// Send Lifecycle Event
		sendToNexus(map[string]interface{}{
			"post_type":       "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         selfID,
			"time":            time.Now().Unix(),
		})

		// Handle incoming commands from Nexus
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

// NexusCommand represents a command from BotNexus (OneBot Action)
type NexusCommand struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   string                 `json:"echo"`
}

func handleNexusCommand(data []byte) {
	var cmd NexusCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		log.Printf("Invalid Nexus command: %v", err)
		return
	}

	log.Printf("Received Command: %s", cmd.Action)

	switch cmd.Action {
	case "send_group_msg", "send_msg":
		groupID, _ := cmd.Params["group_id"].(string)
		message, _ := cmd.Params["message"].(string)
		if groupID != "" && message != "" {
			sendFeishuMessage(groupID, "chat_id", message, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string) // open_id
		message, _ := cmd.Params["message"].(string)
		if userID != "" && message != "" {
			sendFeishuMessage(userID, "open_id", message, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteFeishuMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": "FeishuBot",
			},
			"echo": cmd.Echo,
		})
	case "get_group_list":
		getGroupList(cmd.Echo)
	}
}

func sendFeishuMessage(receiveID, receiveIDType, text, echo string) {
	// Simple text vs image handling could be added here
	// For now, assuming text
	content := map[string]interface{}{
		"text": text,
	}
	contentJSON, _ := json.Marshal(content)
	contentStr := string(contentJSON)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveID).
			MsgType(larkim.MsgTypeText).
			Content(contentStr).
			Build()).
		Build()

	resp, err := larkClient.Im.Message.Create(context.Background(), req)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]interface{}{"status": "failed", "message": resp.Msg, "echo": echo})
	} else {
		log.Printf("Sent message to %s: %s", receiveID, text)
		sendToNexus(map[string]interface{}{"status": "ok", "data": map[string]interface{}{"message_id": *resp.Data.MessageId}, "echo": echo})
	}
}

func deleteFeishuMessage(messageID, echo string) {
	req := larkim.NewDeleteMessageReqBuilder().
		MessageId(messageID).
		Build()

	resp, err := larkClient.Im.Message.Delete(context.Background(), req)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Delete Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]interface{}{"status": "failed", "message": resp.Msg, "echo": echo})
	} else {
		log.Printf("Deleted message: %s", messageID)
		sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
	}
}

func getGroupList(echo string) {
	// Iterate pages if needed, but for now just one page
	req := larkim.NewListChatReqBuilder().
		SortType("ByCreateTimeAsc").
		PageSize(20).
		Build()

	resp, err := larkClient.Im.Chat.List(context.Background(), req)
	if err != nil {
		log.Printf("Failed to get group list: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Group List Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]interface{}{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	var groups []map[string]interface{}
	for _, chat := range resp.Data.Items {
		groups = append(groups, map[string]interface{}{
			"group_id":   *chat.ChatId,
			"group_name": *chat.Name,
		})
	}

	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   groups,
		"echo":   echo,
	})
}
