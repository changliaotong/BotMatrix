package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/websocket"
)

// Config holds the bot configuration
type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	bot       *tgbotapi.BotAPI
	nexusConn *websocket.Conn
	selfID    string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	var err error
	bot, err = tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram Bot: %v", err)
	}

	bot.Debug = true
	selfID = fmt.Sprintf("%d", bot.Self.ID)
	log.Printf("Authorized on account %s (ID: %s)", bot.Self.UserName, selfID)

	// Connect to Nexus
	go connectToNexus()

	// Start Polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleMessage(update.Message)
	}
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err == nil {
		if err := json.Unmarshal(file, &config); err != nil {
			log.Printf("Error parsing config.json: %v", err)
		}
	}

	if envToken := os.Getenv("TELEGRAM_BOT_TOKEN"); envToken != "" {
		config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}
}

func handleMessage(msg *tgbotapi.Message) {
	// Handle Multimedia
	if msg.Photo != nil && len(msg.Photo) > 0 {
		// Get the largest photo
		photo := msg.Photo[len(msg.Photo)-1]
		fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
		file, err := bot.GetFile(fileConfig)
		if err == nil {
			// Construct direct URL: https://api.telegram.org/file/bot<token>/<file_path>
			// tgbotapi helper:
			url := file.Link(config.BotToken)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	} else if msg.Sticker != nil {
		// Handle Sticker as image
		fileConfig := tgbotapi.FileConfig{FileID: msg.Sticker.FileID}
		file, err := bot.GetFile(fileConfig)
		if err == nil {
			url := file.Link(config.BotToken)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	}

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	// OneBot Message
	obMsg := map[string]interface{}{
		"post_type":    "message",
		"message_type": "group", // Telegram doesn't strictly distinguish, but group chats exist
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "normal",
		"message_id":   fmt.Sprintf("%d", msg.MessageID),
		"user_id":      fmt.Sprintf("%d", msg.From.ID),
		"message":      msg.Text,
		"raw_message":  msg.Text,
		"sender": map[string]interface{}{
			"user_id":  fmt.Sprintf("%d", msg.From.ID),
			"nickname": msg.From.UserName,
		},
	}

	if msg.Chat.IsPrivate() {
		obMsg["message_type"] = "private"
	} else {
		obMsg["group_id"] = fmt.Sprintf("%d", msg.Chat.ID)
	}

	sendToNexus(obMsg)
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", selfID)
		header.Add("X-Platform", "Telegram")

		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, header)
		if err != nil {
			log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		nexusConn = conn
		log.Println("Connected to BotNexus!")

		// Lifecycle Event
		sendToNexus(map[string]interface{}{
			"post_type":       "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         selfID,
			"time":            time.Now().Unix(),
		})

		// Handle incoming commands
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
		chatIDStr, _ := cmd.Params["group_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if chatIDStr != "" && text != "" {
			sendTelegramMessage(chatIDStr, text, cmd.Echo)
		}
	case "send_private_msg":
		chatIDStr, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if chatIDStr != "" && text != "" {
			sendTelegramMessage(chatIDStr, text, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": bot.Self.UserName,
			},
			"echo": cmd.Echo,
		})
	}
}

func sendTelegramMessage(chatIDStr, text, echo string) {
	// Parse chatID (int64)
	var chatID int64
	fmt.Sscanf(chatIDStr, "%d", &chatID)

	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %d: %s", chatID, text)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": fmt.Sprintf("%d", sentMsg.MessageID)},
		"echo":   echo,
	})
}
