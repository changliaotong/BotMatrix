package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
)

// Config holds the configuration
type Config struct {
	// Webhook Mode (Custom Robot)
	AccessToken string `json:"access_token"`
	Secret      string `json:"secret"` // Optional: for HMAC signature

	// Stream Mode (Enterprise Robot)
	ClientID     string `json:"client_id"`     // AppKey
	ClientSecret string `json:"client_secret"` // AppSecret

	NexusAddr string `json:"nexus_addr"`
	SelfID    int64  `json:"self_id"` // Optional: manually set SelfID
}

var (
	config       Config
	nexusConn    *websocket.Conn
	httpClient   = &http.Client{Timeout: 10 * time.Second}
	streamClient *client.StreamClient
)

func loadConfig() {
	file, err := os.Open("config.json")
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			log.Println("Error decoding config.json:", err)
		}
	} else {
		log.Println("config.json not found, please create one.")
	}

	// Environment variable overrides
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}
	if config.NexusAddr == "" {
		config.NexusAddr = "ws://localhost:3005"
	}
	// Webhook envs
	if envToken := os.Getenv("DINGTALK_TOKEN"); envToken != "" {
		config.AccessToken = envToken
	}
	if envSecret := os.Getenv("DINGTALK_SECRET"); envSecret != "" {
		config.Secret = envSecret
	}
	// Stream envs
	if envClientID := os.Getenv("DINGTALK_CLIENT_ID"); envClientID != "" {
		config.ClientID = envClientID
	}
	if envClientSecret := os.Getenv("DINGTALK_CLIENT_SECRET"); envClientSecret != "" {
		config.ClientSecret = envClientSecret
	}

	// Generate a SelfID if not set
	if config.SelfID == 0 {
		// Use a hash of the token or client_id as a pseudo ID
		key := config.AccessToken
		if key == "" {
			key = config.ClientID
		}
		if key == "" {
			key = "dingtalk_bot"
		}
		h := sha256.New()
		h.Write([]byte(key))
		bs := h.Sum(nil)
		// Take first 4 bytes
		config.SelfID = int64(bs[0])<<24 | int64(bs[1])<<16 | int64(bs[2])<<8 | int64(bs[3])
		if config.SelfID < 0 {
			config.SelfID = -config.SelfID
		}
		log.Printf("Auto-generated SelfID: %d", config.SelfID)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	if config.AccessToken == "" && config.ClientID == "" {
		log.Fatal("Either DINGTALK_TOKEN (Webhook) or DINGTALK_CLIENT_ID (Stream Mode) is required")
	}

	// Connect to BotNexus
	go connectNexus()

	// Start Stream Client if configured
	if config.ClientID != "" && config.ClientSecret != "" {
		go startStreamClient()
	} else {
		log.Println("Stream Mode not configured (missing client_id/client_secret). Running in Webhook Send-Only mode.")
	}

	// Keep alive
	select {}
}

// --- Stream SDK Integration ---

func startStreamClient() {
	// logger.SetLogger(logger.NewStdLogger(os.Stdout)) // Use default logger or implement ILogger if needed

	cli := client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(config.ClientID, config.ClientSecret)),
		client.WithUserAgent(client.NewDingtalkGoSDKUserAgent()),
		client.WithSubscription("EVENT", "*", func(ctx context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
			handleStreamEvent(df)
			return payload.NewSuccessDataFrameResponse(), nil
		}),
	)

	err := cli.Start(context.Background())
	if err != nil {
		log.Printf("Stream Client failed to start: %v", err)
		return
	}
	streamClient = cli
	log.Println("Stream Client started successfully! Listening for events...")

	// Block until close
	select {}
}

func handleStreamEvent(df *payload.DataFrame) {
	// Parse the event content
	// DingTalk Stream events usually contain a JSON payload in the Data field
	log.Printf("Received Stream Event: Type=%s, Data=%s", df.Type, df.Data)

	// Basic OneBot Message Event Construction
	// Note: We need to parse the specific DingTalk event format.
	// This is a simplified example assuming we receive a message event.

	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(df.Data), &eventData); err != nil {
		log.Println("Error parsing event data:", err)
		return
	}

	// Check event type from data if available, or assume based on subscription
	eventType, _ := eventData["type"].(string)

	// Map to OneBot Message
	// This mapping depends heavily on the actual JSON structure of DingTalk Stream events
	// For now, we log it. To make it functional, we'd need the exact spec.
	// But simply receiving it proves the SDK works.

	// Example: Forwarding a raw event to Nexus for debugging/handling
	sendToNexus(map[string]interface{}{
		"post_type":   "notice", // or message
		"notice_type": "dingtalk_event",
		"sub_type":    eventType,
		"raw_data":    eventData,
		"self_id":     config.SelfID,
		"time":        time.Now().Unix(),
	})
}

// --- BotNexus Integration ---

func connectNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		header := http.Header{}
		header.Add("X-Self-ID", fmt.Sprintf("%d", config.SelfID))
		header.Add("X-Client-Role", "Universal") // Generic OneBot Client

		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, header)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5s...", err)
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
			"self_id":         config.SelfID,
			"time":            time.Now().Unix(),
		})

		// Handle messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}
			handleNexusMessage(message)
		}

		conn.Close()
		log.Println("Disconnected from BotNexus. Reconnecting...")
		time.Sleep(3 * time.Second)
	}
}

func sendToNexus(data map[string]interface{}) {
	if nexusConn == nil {
		return
	}
	// Ensure self_id is present
	if _, ok := data["self_id"]; !ok {
		data["self_id"] = config.SelfID
	}

	err := nexusConn.WriteJSON(data)
	if err != nil {
		log.Println("Error sending to Nexus:", err)
	}
}

func handleNexusMessage(message []byte) {
	var action map[string]interface{}
	if err := json.Unmarshal(message, &action); err != nil {
		log.Println("Invalid JSON:", err)
		return
	}

	// Only handle action requests
	actionName, ok := action["action"].(string)
	if !ok {
		return
	}

	log.Printf("Received Action: %s", actionName)

	switch actionName {
	case "send_group_msg", "send_msg":
		params, _ := action["params"].(map[string]interface{})
		msg := getString(params, "message")
		if msg != "" {
			var err error
			if config.AccessToken != "" {
				err = sendDingTalkMessage(msg)
			} else {
				// Fallback to Enterprise sending if needed (not implemented yet)
				err = fmt.Errorf("webhook access_token not configured")
			}

			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
			} else {
				sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": action["echo"]})
			}
		}

	case "send_private_msg":
		params, _ := action["params"].(map[string]interface{})
		msg := getString(params, "message")
		userID := getString(params, "user_id") // Can be mobile or DingTalk UserID

		if msg != "" {
			var err error
			if config.AccessToken != "" {
				// Simulate private msg via @mention in group
				err = sendDingTalkMessageWithAt(msg, []string{userID})
			} else {
				err = fmt.Errorf("webhook access_token not configured")
			}

			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
			} else {
				sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": action["echo"]})
			}
		}

	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  config.SelfID,
				"nickname": "DingTalk Bot",
			},
			"echo": action["echo"],
		})
	}
}

// Helper to get string from map safely
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// --- DingTalk Webhook API ---

func getWebhookURL() string {
	baseURL := "https://oapi.dingtalk.com/robot/send?access_token=" + config.AccessToken
	if config.Secret == "" {
		return baseURL
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	stringToSign := timestamp + "\n" + config.Secret

	h := hmac.New(sha256.New, []byte(config.Secret))
	h.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	return fmt.Sprintf("%s&timestamp=%s&sign=%s", baseURL, timestamp, sign)
}

func sendDingTalkMessage(content string) error {
	return sendDingTalkMessageWithAt(content, nil)
}

func sendDingTalkMessageWithAt(content string, atMobiles []string) error {
	url := getWebhookURL()

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	if len(atMobiles) > 0 {
		payload["at"] = map[string]interface{}{
			"atMobiles": atMobiles,
		}
	}

	jsonBody, _ := json.Marshal(payload)

	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("dingtalk api error: %v", result)
	}

	return nil
}
