package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"BotMatrix/common/bot"

	"github.com/bwmarrin/discordgo"
)

type DiscordConfig struct {
	bot.BotConfig
}

var (
	botService *bot.BaseBot
	discordCfg DiscordConfig
	dg         *discordgo.Session
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
)

func main() {
	// Initialize base bot with default log port
	botService = bot.NewBaseBot(3134)

	// Setup logging to use the common LogManager
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("DiscordBot", &discordCfg, restartBot, []bot.ConfigSection{
		{
			Title: "Discord API 配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Token", ID: "bot_token", Type: "password", Value: discordCfg.BotToken},
			},
		},
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Nexus 地址", ID: "nexus_addr", Type: "text", Value: discordCfg.NexusAddr},
				{Label: "Web UI 端口", ID: "log_port", Type: "number", Value: discordCfg.LogPort},
			},
		},
	})

	// Start common HTTP services (health, logs, config)
	go botService.StartHTTPServer()

	// Start platform specific bot
	restartBot()

	// Wait for exit signal and handle cleanup
	botService.WaitExitSignal()
	stopBot()
}

func loadConfig() {
	botService.LoadConfig("config.json")

	// Sync local config
	botService.Mu.RLock()
	discordCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Handle environment variables if any (optional, but good for consistency)
	if envToken := os.Getenv("DISCORD_TOKEN"); envToken != "" {
		discordCfg.BotToken = envToken
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Sync botService.Config from discordCfg
	botService.Config = discordCfg.BotConfig
	token := discordCfg.BotToken
	nexusAddr := discordCfg.NexusAddr
	botService.Mu.RUnlock()

	if token == "" {
		log.Println("WARNING: Discord Bot Token is not configured. Bot will not start until configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Printf("Error creating Discord session: %v", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		log.Printf("Error opening Discord connection: %v", err)
		return
	}

	selfID = dg.State.User.ID
	log.Printf("Bot is now running. Logged in as %s#%s (%s)", dg.State.User.Username, dg.State.User.Discriminator, selfID)

	// Connect to Nexus for central management
	botService.StartNexusConnection(botCtx, nexusAddr, "Discord", selfID, handleNexusCommand)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
	if dg != nil {
		log.Println("Stopping Discord session...")
		dg.Close()
		dg = nil
	}
}

func handleNexusCommand(data []byte) {
	var req struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return
	}

	log.Printf("Received action from Nexus: %s", req.Action)

	var resp map[string]any
	switch req.Action {
	case "send_msg", "send_group_msg", "send_private_msg":
		msg, _ := req.Params["message"].(string)
		targetID := ""
		if req.Action == "send_group_msg" {
			targetID, _ = req.Params["group_id"].(string)
		} else {
			targetID, _ = req.Params["user_id"].(string)
		}

		if targetID != "" && msg != "" {
			_, err := dg.ChannelMessageSend(targetID, msg)
			if err != nil {
				resp = map[string]any{"status": "failed", "retcode": 500, "msg": err.Error()}
			} else {
				resp = map[string]any{"status": "ok", "retcode": 0, "data": map[string]any{"message_id": "discord_" + time.Now().String()}}
			}
		}
	case "get_status":
		resp = map[string]interface{}{
			"status":  "ok",
			"retcode": 0,
			"data": map[string]interface{}{
				"online": true,
				"good":   true,
			},
		}
	case "get_login_info":
		resp = map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": dg.State.User.Username,
			},
		}
	}

	if resp != nil && req.Echo != "" {
		resp["echo"] = req.Echo
		botService.SendToNexus(resp)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Handle images
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 || attachment.Height > 0 {
			m.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		}
	}

	log.Printf("[%s] %s", m.Author.Username, m.Content)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  m.ID,
		"user_id":     m.Author.ID,
		"message":     m.Content,
		"raw_message": m.Content,
		"sender": map[string]any{
			"user_id":  m.Author.ID,
			"nickname": m.Author.Username,
		},
	}

	if m.GuildID != "" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = m.ChannelID
	} else {
		obMsg["message_type"] = "private"
	}

	botService.SendToNexus(obMsg)
}
