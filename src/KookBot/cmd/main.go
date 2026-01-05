package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"BotMatrix/common/bot"

	"github.com/lonelyevil/kook"
)

// KookConfig extends bot.BotConfig with Kook specific fields
type KookConfig struct {
	bot.BotConfig
}

var (
	botService *bot.BaseBot
	session    *kook.Session
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
	kookCfg    KookConfig
)

// --- Simple Logger Implementation ---

type ConsoleLogger struct{}

func (l *ConsoleLogger) Trace() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Debug() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Info() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Warn() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Error() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Fatal() kook.Entry { return &ConsoleEntry{} }

type ConsoleEntry struct{}

func (e *ConsoleEntry) Bool(key string, b bool) kook.Entry             { return e }
func (e *ConsoleEntry) Bytes(key string, val []byte) kook.Entry        { return e }
func (e *ConsoleEntry) Caller(depth int) kook.Entry                    { return e }
func (e *ConsoleEntry) Dur(key string, d time.Duration) kook.Entry     { return e }
func (e *ConsoleEntry) Err(key string, err error) kook.Entry           { return e }
func (e *ConsoleEntry) Float64(key string, f float64) kook.Entry       { return e }
func (e *ConsoleEntry) IPAddr(key string, ip net.IP) kook.Entry        { return e }
func (e *ConsoleEntry) Int(key string, i int) kook.Entry               { return e }
func (e *ConsoleEntry) Int64(key string, i int64) kook.Entry           { return e }
func (e *ConsoleEntry) Interface(key string, i interface{}) kook.Entry { return e }
func (e *ConsoleEntry) Msg(msg string)                                 { botService.Info(msg) }
func (e *ConsoleEntry) Msgf(f string, i ...interface{})                { botService.Info(f, i...) }
func (e *ConsoleEntry) Str(key string, s string) kook.Entry            { return e }
func (e *ConsoleEntry) Strs(key string, s []string) kook.Entry         { return e }
func (e *ConsoleEntry) Time(key string, t time.Time) kook.Entry        { return e }

// ------------------------------------

func main() {
	botService = bot.NewBaseBot(3136)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("KookBot", &kookCfg, restartBot, []bot.ConfigSection{
		{
			Title: "Kook API 配置",
			Fields: []bot.ConfigField{
				{Label: "Bot Token", ID: "bot_token", Type: "password", Value: kookCfg.BotToken},
			},
		},
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: kookCfg.NexusAddr},
				{Label: "Web UI 端口", ID: "log_port", Type: "number", Value: kookCfg.LogPort},
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

	// Sync common config to local kookCfg
	botService.Mu.RLock()
	kookCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Kook specific fields (if any in the future)
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &kookCfg)
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Update botService.Config from the potentially changed kookCfg
	botService.Config = kookCfg.BotConfig
	token := botService.Config.BotToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if token == "" {
		botService.Warn("Kook Bot Token is not configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	session = kook.New(token, &ConsoleLogger{})

	session.AddHandler(textMessageHandler)
	session.AddHandler(imageMessageHandler)
	session.AddHandler(kmarkdownMessageHandler)

	err := session.Open()
	if err != nil {
		botService.Error("Error opening Kook connection: %v", err)
		return
	}

	user, err := session.UserMe()
	if err == nil {
		selfID = user.ID
		botService.Info("Kook Bot started as %s (%s)", user.Username, selfID)
	}

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Kook", selfID, handleNexusCommand)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}

	if session != nil {
		session.Close()
		session = nil
	}
}

func handleCommon(commonData *kook.EventDataGeneral, author kook.User) {
	if author.Bot && author.ID == selfID {
		return
	}

	botService.Info("[%s] %s", author.Username, commonData.Content)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  commonData.MsgID,
		"user_id":     commonData.AuthorID,
		"message":     commonData.Content,
		"raw_message": commonData.Content,
		"sender": map[string]any{
			"user_id":  commonData.AuthorID,
			"nickname": author.Username,
		},
	}

	if commonData.Type == kook.MessageTypeImage {
		obMsg["message"] = fmt.Sprintf("[CQ:image,file=%s]", commonData.Content)
		obMsg["raw_message"] = obMsg["message"]
	}

	if commonData.ChannelType == "GROUP" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = commonData.TargetID
	} else {
		obMsg["message_type"] = "private"
	}

	botService.SendToNexus(obMsg)
}

func textMessageHandler(ctx *kook.TextMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
}

func imageMessageHandler(ctx *kook.ImageMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
}

func kmarkdownMessageHandler(ctx *kook.KmarkdownMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
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

	botService.Info("Received Command: %s", cmd.Action)

	switch cmd.Action {
	case "send_group_msg", "send_msg":
		channelID, _ := cmd.Params["group_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if channelID != "" && text != "" {
			sendKookMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			sendKookDirectMessage(userID, text, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteKookMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		botService.SendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  selfID,
				"nickname": "KookBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func deleteKookMessage(msgID, echo string) {
	err := session.MessageDelete(msgID)
	if err != nil {
		botService.Error("Failed to delete message %s: %v", msgID, err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.Info("Deleted message %s", msgID)
	botService.SendToNexus(map[string]any{"status": "ok", "echo": echo})
}

func sendKookMessage(targetID, content, echo string) {
	resp, err := session.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			TargetID: targetID,
			Content:  content,
			Type:     kook.MessageTypeText,
		},
	})

	if err != nil {
		botService.Error("Failed to send message to %s: %v", targetID, err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.Info("Sent message to %s", targetID)
	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": resp.MsgID},
		"echo":   echo,
	})
}

func sendKookDirectMessage(targetID, content, echo string) {
	resp, err := session.DirectMessageCreate(&kook.DirectMessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			TargetID: targetID,
			Content:  content,
			Type:     kook.MessageTypeText,
		},
	})

	if err != nil {
		botService.Error("Failed to send private message to %s: %v", targetID, err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.Info("Sent private message to %s", targetID)
	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": resp.MsgID},
		"echo":   echo,
	})
}

