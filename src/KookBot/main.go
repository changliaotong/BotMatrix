package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lonelyevil/kook"
)

type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	session   *kook.Session
	nexusConn *websocket.Conn
	selfID    string
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
func (e *ConsoleEntry) Msg(msg string)                                 { log.Println(msg) }
func (e *ConsoleEntry) Msgf(f string, i ...interface{})                { log.Printf(f, i...) }
func (e *ConsoleEntry) Str(key string, s string) kook.Entry            { return e }
func (e *ConsoleEntry) Strs(key string, s []string) kook.Entry         { return e }
func (e *ConsoleEntry) Time(key string, t time.Time) kook.Entry        { return e }

// ------------------------------------

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	session = kook.New(config.BotToken, &ConsoleLogger{})

	// Register Handlers
	session.AddHandler(textMessageHandler)
	session.AddHandler(imageMessageHandler)
	session.AddHandler(kmarkdownMessageHandler)

	// Open connection
	err := session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer session.Close()

	// Get Self Info
	me, err := session.UserMe()
	if err == nil {
		selfID = me.ID
		log.Printf("Bot logged in as %s (ID: %s)", me.Username, selfID)
	}

	// Connect to BotNexus
	go connectToNexus()

	// Wait for signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	}

	if envToken := os.Getenv("KOOK_BOT_TOKEN"); envToken != "" {
		config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.BotToken == "" {
		log.Fatal("KOOK_BOT_TOKEN is required")
	}
}

func handleCommon(common *kook.EventDataGeneral, author kook.User) {
	if author.Bot && author.ID == selfID {
		return
	}

	log.Printf("[%s] %s", author.Username, common.Content)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  common.MsgID,
		"user_id":     common.AuthorID,
		"message":     common.Content,
		"raw_message": common.Content,
		"sender": map[string]interface{}{
			"user_id":  common.AuthorID,
			"nickname": author.Username,
		},
	}

	if common.Type == kook.MessageTypeImage {
		obMsg["message"] = fmt.Sprintf("[CQ:image,file=%s]", common.Content)
		obMsg["raw_message"] = obMsg["message"]
	}

	if common.ChannelType == "GROUP" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = common.TargetID
	} else {
		obMsg["message_type"] = "private"
	}

	sendToNexus(obMsg)
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

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", selfID)
		header.Add("X-Platform", "Kook")

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
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
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
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s", msgID)
	sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
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
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s: %s", targetID, content)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": resp.MsgID},
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
		log.Printf("Failed to send private message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent private message to %s", targetID)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": resp.MsgID},
		"echo":   echo,
	})
}
