package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"BotMatrix/common/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botService *bot.BaseBot
	tgBot      *tgbotapi.BotAPI
	botCtx     context.Context
	botCancel  context.CancelFunc
)

func main() {
	botService = bot.NewBaseBot(8087)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	botService.LoadConfig("config.json")

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("TelegramBot", &botService.Config, restartBot, []bot.ConfigSection{
		{
			Title: "Telegram API 配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Token", ID: "bot_token", Type: "password", Value: botService.Config.BotToken},
			},
		},
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: botService.Config.NexusAddr},
				{Label: "Web UI 端口", ID: "log_port", Type: "number", Value: botService.Config.LogPort},
			},
		},
	})

	go botService.StartHTTPServer()

	restartBot()

	botService.WaitExitSignal()
	stopBot()
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	botToken := botService.Config.BotToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if botToken == "" {
		log.Println("Telegram bot token is not set, bot will not start")
		return
	}

	var err error
	tgBot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("Failed to create Telegram Bot: %v", err)
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	tgBot.Debug = true
	selfID := fmt.Sprintf("%d", tgBot.Self.ID)
	log.Printf("Authorized on account %s (ID: %s)", tgBot.Self.UserName, selfID)

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Telegram", selfID, handleNexusCommand)

	// Start Polling
	go func(ctx context.Context) {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := tgBot.GetUpdatesChan(u)

		for {
			select {
			case <-ctx.Done():
				tgBot.StopReceivingUpdates()
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				if update.Message == nil {
					continue
				}
				handleMessage(update.Message, selfID)
			}
		}
	}(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func handleMessage(msg *tgbotapi.Message, selfID string) {
	// Handle Multimedia
	if msg.Photo != nil && len(msg.Photo) > 0 {
		photo := msg.Photo[len(msg.Photo)-1]
		fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
		file, err := tgBot.GetFile(fileConfig)
		if err == nil {
			botService.Mu.RLock()
			token := botService.Config.BotToken
			botService.Mu.RUnlock()
			url := file.Link(token)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	} else if msg.Sticker != nil {
		fileConfig := tgbotapi.FileConfig{FileID: msg.Sticker.FileID}
		file, err := tgBot.GetFile(fileConfig)
		if err == nil {
			botService.Mu.RLock()
			token := botService.Config.BotToken
			botService.Mu.RUnlock()
			url := file.Link(token)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	}

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	obMsg := map[string]any{
		"post_type":    "message",
		"message_type": "group",
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "normal",
		"message_id":   fmt.Sprintf("%d", msg.MessageID),
		"user_id":      fmt.Sprintf("%d", msg.From.ID),
		"message":      msg.Text,
		"raw_message":  msg.Text,
		"sender": map[string]any{
			"user_id":  fmt.Sprintf("%d", msg.From.ID),
			"nickname": msg.From.UserName,
		},
	}

	if msg.Chat.IsPrivate() {
		obMsg["message_type"] = "private"
	} else {
		obMsg["group_id"] = fmt.Sprintf("%d", msg.Chat.ID)
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
		return
	}

	log.Printf("Received Action: %s", cmd.Action)
	result, err := handleAction(cmd.Action, cmd.Params)

	resp := map[string]any{
		"echo": cmd.Echo,
	}
	if err != nil {
		resp["status"] = "failed"
		resp["msg"] = err.Error()
	} else {
		resp["status"] = "ok"
		resp["data"] = result
	}
	botService.SendToNexus(resp)
}

func handleAction(action string, params map[string]any) (any, error) {
	switch action {
	case "send_msg":
		return sendTelegramMessage(params)
	case "delete_msg":
		return deleteTelegramMessage(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func sendTelegramMessage(params map[string]any) (any, error) {
	if tgBot == nil {
		return nil, fmt.Errorf("bot is not running")
	}

	chatIDStr, ok := params["chat_id"].(string)
	if !ok {
		// Try group_id or user_id as fallback
		if gid, ok := params["group_id"].(string); ok {
			chatIDStr = gid
		} else if uid, ok := params["user_id"].(string); ok {
			chatIDStr = uid
		} else {
			return nil, fmt.Errorf("missing chat_id")
		}
	}

	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message")
	}

	var chatID int64
	fmt.Sscanf(chatIDStr, "%d", &chatID)

	msg := tgbotapi.NewMessage(chatID, message)
	sent, err := tgBot.Send(msg)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message_id": fmt.Sprintf("%d", sent.MessageID),
	}, nil
}

func deleteTelegramMessage(params map[string]any) (any, error) {
	if tgBot == nil {
		return nil, fmt.Errorf("bot is not running")
	}

	chatIDStr, _ := params["chat_id"].(string)
	messageIDStr, ok := params["message_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message_id")
	}

	var chatID int64
	var messageID int
	fmt.Sscanf(chatIDStr, "%d", &chatID)
	fmt.Sscanf(messageIDStr, "%d", &messageID)

	delMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := tgBot.Request(delMsg)
	return nil, err
}