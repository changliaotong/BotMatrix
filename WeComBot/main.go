package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		req := message.SendTextRequest{
			SendRequestCommon: &message.SendRequestCommon{
				ToUser:  userID,
				MsgType: "text",
				AgentID: strconv.FormatInt(config.AgentID, 10),
			},
			Text: message.TextField{
				Content: msgContent,
			},
		}

		msgManager := wc.GetMessage()
		msgID, err := msgManager.SendText(req)
		if err != nil {
			log.Printf("Failed to send WeCom message: %v", err)
			response["status"] = "failed"
			response["retcode"] = -1
		} else {
			response["data"] = map[string]interface{}{
				"message_id": msgID,
			}
		}

	case "delete_msg":
		msgID, _ := params["message_id"].(string)
		if msgID != "" {
			err := recallMessage(msgID)
			if err != nil {
				log.Printf("Failed to recall WeCom message: %v", err)
				response["status"] = "failed"
				response["retcode"] = -1
			}
		}
	}

	if conn != nil {
		conn.WriteJSON(response)
	}
}

func recallMessage(msgID string) error {
	token, err := wc.GetContext().GetAccessToken()
	if err != nil {
		return err
	}

	url := "https://qyapi.weixin.qq.com/cgi-bin/message/recall?access_token=" + token
	payload := map[string]string{"msgid": msgID}
	jsonBody, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("wecom recall error: %v", result)
	}
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	// Initialize WeCom SDK
	wcConfig := &workConfig.Config{
		CorpID:         config.CorpID,
		AgentID:        strconv.FormatInt(config.AgentID, 10),
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
	// TODO: Work module in v2.1.11 does not have GetServer.
	// We need to implement manual callback handling (signature verification, decryption, XML parsing).
	// For now, we just acknowledge the request to avoid timeout errors on WeCom side.

	echoStr := c.Query("echostr")
	if echoStr != "" {
		// Verify URL (GET request)
		// signature := c.Query("msg_signature")
		// timestamp := c.Query("timestamp")
		// nonce := c.Query("nonce")
		// if util.Signature(config.Token, timestamp, nonce, echoStr) == signature { ... }

		// For verification, we usually need to decrypt the echostr using EncodingAESKey.
		// Since we don't have the full logic yet, we might fail the verification step in WeCom admin panel.
		// But if this is already configured, we just need to handle POST messages.

		// We can use util.DecryptMsg if needed.
		// But for now, let's just log.
		log.Printf("Received verification request: %s", echoStr)
	}

	// Handle POST messages
	if c.Request.Method == "POST" {
		log.Println("Received WeCom callback (POST)")
		// TODO: Parse body, decrypt, broadcast to Nexus
	}

	c.String(200, "success")
}
