package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"BotMatrix/common/bot"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuConfig extends bot.BotConfig with Feishu specific fields
type FeishuConfig struct {
	bot.BotConfig
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	EncryptKey        string `json:"encrypt_key"`
	VerificationToken string `json:"verification_token"`
}

var (
	botService *bot.BaseBot
	larkClient *lark.Client
	botCtx     context.Context
	botCancel  context.CancelFunc
	feishuCfg  FeishuConfig
)

func main() {
	botService = bot.NewBaseBot(3135)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initial load of config into local feishuCfg
	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("FeishuBot", &feishuCfg, restartBot, []bot.ConfigSection{
		{
			Title: "飞书 API 配置",
			Fields: []bot.ConfigField{
				{Label: "App ID", ID: "app_id", Type: "text", Value: feishuCfg.AppID},
				{Label: "App Secret", ID: "app_secret", Type: "password", Value: feishuCfg.AppSecret},
				{Label: "Encrypt Key (可选)", ID: "encrypt_key", Type: "password", Value: feishuCfg.EncryptKey},
				{Label: "Verification Token (可选)", ID: "verification_token", Type: "password", Value: feishuCfg.VerificationToken},
			},
		},
		{
			Title: "连接与服务配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: feishuCfg.NexusAddr},
				{Label: "Web UI 端口 (LogPort)", ID: "log_port", Type: "number", Value: feishuCfg.LogPort},
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

	// Sync common config to local feishuCfg
	botService.Mu.RLock()
	feishuCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Feishu specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &feishuCfg)
	}

	// Environment variable overrides for Feishu
	if envAppID := os.Getenv("APP_ID"); envAppID != "" {
		feishuCfg.AppID = envAppID
	}
	if envAppSecret := os.Getenv("APP_SECRET"); envAppSecret != "" {
		feishuCfg.AppSecret = envAppSecret
	}
	if envEncryptKey := os.Getenv("ENCRYPT_KEY"); envEncryptKey != "" {
		feishuCfg.EncryptKey = envEncryptKey
	}
	if envVerToken := os.Getenv("VERIFICATION_TOKEN"); envVerToken != "" {
		feishuCfg.VerificationToken = envVerToken
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Sync botService.Config from feishuCfg
	botService.Config = feishuCfg.BotConfig
	appID := feishuCfg.AppID
	appSecret := feishuCfg.AppSecret
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if appID == "" || appSecret == "" {
		botService.Warn("Feishu AppID or AppSecret is not configured.")
		return
	}

	// Initialize Lark API Client
	larkClient = lark.NewClient(appID, appSecret)

	botCtx, botCancel = context.WithCancel(context.Background())

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Feishu", appID, handleNexusCommand)

	// Start Lark WebSocket Client (Receive Events)
	go startLarkWS(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func startLarkWS(ctx context.Context) {
	botService.Mu.RLock()
	verToken := feishuCfg.VerificationToken
	encryptKey := feishuCfg.EncryptKey
	appID := feishuCfg.AppID
	appSecret := feishuCfg.AppSecret
	botService.Mu.RUnlock()

	// Register Event Handler
	eventHandler := larkevent.NewEventDispatcher(verToken, encryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			handleMessage(ctx, event)
			return nil
		})

	// Create WebSocket Client
	cli := larkws.NewClient(appID, appSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// Start WebSocket Client
	log.Println("Connecting to Feishu/Lark Gateway...")
	err := cli.Start(ctx)
	if err != nil {
		log.Printf("Failed to start Feishu WebSocket client: %v", err)
	}
}

func handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msg := event.Event.Message
	sender := event.Event.Sender

	// Extract Content (JSON string)
	var contentMap map[string]any
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
	obMsg := map[string]any{
		"post_type":    "message",
		"message_type": msgType,
		"time":         time.Now().Unix(),
		"self_id":      feishuCfg.AppID,
		"sub_type":     "normal",
		"message_id":   *msg.MessageId,
		"user_id":      userID,
		"group_id":     groupID,
		"message":      text,
		"raw_message":  text,
		"sender": map[string]any{
			"user_id":  userID,
			"nickname": "FeishuUser",
		},
	}

	botService.SendToNexus(obMsg)
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
	}
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
		botService.SendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  feishuCfg.AppID,
				"nickname": "FeishuBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendFeishuMessage(receiveID, receiveIDType, text, echo string) {
	if larkClient == nil {
		return
	}

	content := map[string]string{
		"text": text,
	}
	contentJSON, _ := json.Marshal(content)

	resp, err := larkClient.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(receiveID).
			Content(string(contentJSON)).
			Build()).
		Build())

	if err != nil {
		log.Printf("Failed to send Feishu message: %v", err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API error: %d - %s", resp.Code, resp.Msg)
		botService.SendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data": map[string]any{
			"message_id": *resp.Data.MessageId,
		},
		"echo": echo,
	})
}

func deleteFeishuMessage(messageID, echo string) {
	if larkClient == nil {
		return
	}

	resp, err := larkClient.Im.Message.Delete(context.Background(), larkim.NewDeleteMessageReqBuilder().
		MessageId(messageID).
		Build())

	if err != nil {
		botService.Error("Failed to delete Feishu message %s: %v", messageID, err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		botService.SendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	botService.SendToNexus(map[string]any{"status": "ok", "echo": echo})
}
