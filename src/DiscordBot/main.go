package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	dg        *discordgo.Session
	nexusConn *websocket.Conn
	selfID    string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	var err error
	dg, err = discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	selfID = dg.State.User.ID
	log.Printf("Bot is now running. Logged in as %s#%s (%s)", dg.State.User.Username, dg.State.User.Discriminator, selfID)

	go connectToNexus()

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	}

	if envToken := os.Getenv("DISCORD_BOT_TOKEN"); envToken != "" {
		config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.BotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Handle Attachments (Images)
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 || attachment.Height > 0 { // Simple check if it's an image
			m.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		}
	}

	log.Printf("[%s] %s", m.Author.Username, m.Content)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  m.ID,
		"user_id":     m.Author.ID,
		"message":     m.Content,
		"raw_message": m.Content,
		"sender": map[string]interface{}{
			"user_id":  m.Author.ID,
			"nickname": m.Author.Username,
		},
	}

	if m.GuildID != "" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = m.ChannelID // OneBot group_id maps to Discord Channel ID for simplicity
		// Or we could use GuildID, but messages happen in channels.
		// For OneBot compatibility, mapping ChannelID to GroupID is more practical for chat bots.
	} else {
		obMsg["message_type"] = "private"
	}

	sendToNexus(obMsg)
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", selfID)
		header.Add("X-Platform", "Discord")

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
			sendDiscordMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			// Create DM channel first
			ch, err := dg.UserChannelCreate(userID)
			if err == nil {
				sendDiscordMessage(ch.ID, text, cmd.Echo)
			} else {
				log.Printf("Failed to create DM: %v", err)
			}
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteDiscordMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": dg.State.User.Username,
			},
			"echo": cmd.Echo,
		})
	}
}

func sendDiscordMessage(channelID, text, echo string) {
	msg, err := dg.ChannelMessageSend(channelID, text)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}
	log.Printf("Sent message to %s: %s", channelID, text)
	// Return composite ID: "channelID:messageID"
	compositeID := fmt.Sprintf("%s:%s", channelID, msg.ID)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteDiscordMessage(compositeID, echo string) {
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		sendToNexus(map[string]interface{}{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}
	channelID := parts[0]
	messageID := parts[1]

	err := dg.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s in channel %s", messageID, channelID)
	sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
}
