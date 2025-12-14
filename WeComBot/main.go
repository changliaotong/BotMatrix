package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/work"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
)

// Config holds the bot configuration
type Config struct {
	CorpID         string `json:"corp_id"`
	AgentID        int64  `json:"agent_id"`
	Secret         string `json:"secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	ListenPort     int    `json:"listen_port"`
	NexusAddr      string `json:"nexus_addr"`
}

var (
	config Config
	conn   *websocket.Conn
	wc     *work.Work
	selfID string
)

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config.json: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config.json: %v", err)
	}
	selfID = fmt.Sprintf("%d", config.AgentID)
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		c, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, nil)
		if err != nil {
			log.Printf("Connection error: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn = c
		log.Println("Connected to BotNexus!")

		// Send Lifecycle Event
		sendEvent(map[string]interface{}{
			"post_type":       "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         selfID,
			"time":            time.Now().Unix(),
		})

		// Send Heartbeat
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if conn == nil {
					return
				}
				sendEvent(map[string]interface{}{
					"post_type":       "meta_event",
					"meta_event_type": "heartbeat",
					"self_id":         selfID,
					"time":            time.Now().Unix(),
					"status": map[string]interface{}{
						"online": true,
						"good":   true,
					},
				})
			}
		}()

		// Handle Incoming Actions
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				conn = nil
				break
			}
			handleAction(msg)
		}
	}
}

func sendEvent(event map[string]interface{}) {
	if conn == nil {
		return
	}
	event["platform"] = "wecom"
	err := conn.WriteJSON(event)
	if err != nil {
		log.Printf("Write error: %v", err)
		conn = nil
	}
}

func handleAction(msg []byte) {
	var action map[string]interface{}
	if err := json.Unmarshal(msg, &action); err != nil {
		log.Printf("JSON Unmarshal error: %v", err)
		return
	}

	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]interface{})
	echo, _ := action["echo"]

	response := map[string]interface{}{
		"status":  "ok",
		"retcode": 0,
		"data":    nil,
		"echo":    echo,
	}

	switch actionType {
	case "send_private_msg":
		userID, _ := params["user_id"].(string)
		msgContent, _ := params["message"].(string)

		// Send text message
		msg := message.Message{
			ToUser:  userID,
			MsgType: "text",
			Text: message.Text{
				Content: msgContent,
			},
			Safe: 0,
		}

		msgManager := wc.GetMessage()
		_, err := msgManager.Send(msg)
		if err != nil {
			log.Printf("Failed to send WeCom message: %v", err)
			response["status"] = "failed"
			response["retcode"] = -1
		}
	}

	if conn != nil {
		conn.WriteJSON(response)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	// Initialize WeCom SDK
	wcConfig := &workConfig.Config{
		CorpID:         config.CorpID,
		AgentID:        config.AgentID,
		CorpSecret:     config.Secret,
		Token:          config.Token,
		EncodingAESKey: config.EncodingAESKey,
		Cache:          cache.NewMemory(),
	}
	wc = wechat.NewWechat().GetWork(wcConfig)

	// Connect to BotNexus
	go connectToNexus()

	// Start HTTP Server for Callbacks
	r := gin.Default()
	r.GET("/callback", handleCallback)
	r.POST("/callback", handleCallback)

	log.Printf("Starting WeCom Callback Server on :%d", config.ListenPort)
	go func() {
		if err := r.Run(fmt.Sprintf(":%d", config.ListenPort)); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
}

func handleCallback(c *gin.Context) {
	server := wc.GetServer(c.Request, c.Writer)

	// Set Message Handler
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		// Log message
		log.Printf("Received WeCom Message: %+v", msg)

		// Construct OneBot Message
		// MsgType: text, image, voice, video, location, link

		var content string
		if msg.MsgType == "text" {
			content = msg.Content
		} else if msg.MsgType == "image" {
			content = fmt.Sprintf("[CQ:image,file=%s]", msg.PicURL)
		} else {
			content = fmt.Sprintf("[Unsupported MsgType: %s]", msg.MsgType)
		}

		event := map[string]interface{}{
			"post_type":    "message",
			"message_type": "private", // Internal App usually treats messages as private
			"time":         msg.CreateTime,
			"self_id":      selfID,
			"sub_type":     "friend",
			"message_id":   strconv.FormatInt(msg.MsgID, 10),
			"user_id":      msg.FromUserName, // This is the UserID in WeCom
			"message":      content,
			"raw_message":  content,
			"sender": map[string]interface{}{
				"user_id":  msg.FromUserName,
				"nickname": msg.FromUserName, // We don't have nickname yet
			},
		}

		sendEvent(event)

		return nil // Don't reply directly via XML to avoid timeout issues with complex logic
	})

	// Process Request
	if err := server.Serve(); err != nil {
		log.Printf("WeCom Serve Error: %v", err)
		return
	}

	// In gin, we don't need to do anything else as server.Serve() handles the response
}
