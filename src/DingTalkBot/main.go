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
	"strings"
	"sync"
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
	SelfID    string `json:"self_id"` // Optional: manually set SelfID
}

var (
	config       Config
	nexusConn    *websocket.Conn
	nexusMu      sync.Mutex // Added mutex
	httpClient   = &http.Client{Timeout: 10 * time.Second}
	streamClient *client.StreamClient
)

type LogManager struct {
	mu     sync.Mutex
	buffer []string
	size   int
}

func (l *LogManager) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := string(p)
	// Stream to Nexus
	go func(m string) {
		// Use a temporary map to avoid concurrent map writes if sendToNexus modifies it
		// But sendToNexus is locked now.
		sendToNexus(map[string]interface{}{
			"post_type": "log",
			"level":     "INFO",
			"message":   strings.TrimSpace(m),
			"time":      time.Now().Format("15:04:05"),
			"self_id":   config.SelfID,
		})
	}(msg)

	return os.Stdout.Write(p)
}

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
	if config.SelfID == "" {
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
		config.SelfID = fmt.Sprintf("%x", bs[:4])
		log.Printf("Auto-generated SelfID: %s", config.SelfID)
	}
}

func main() {
	// Initialize Log Manager
	logManager := &LogManager{
		buffer: make([]string, 0, 100),
		size:   100,
	}
	log.SetOutput(logManager)
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
	log.Printf("Received Stream Event: Type=%s", df.Type)

	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(df.Data), &eventData); err != nil {
		log.Println("Error parsing event data:", err)
		return
	}

	// Extract Event Type
	eventType, _ := eventData["type"].(string) // e.g., "im.message.receive_v1"

	if eventType == "im.message.receive_v1" {
		// Handle Message Event
		// Note: The actual structure is deeply nested
		// eventData["data"] -> {"content": "...", "sender_id": ...} (Simplified assumption)
		// DingTalk Stream V2 event structure:
		// {
		//   "specVersion": "1.0",
		//   "type": "im.message.receive_v1",
		//   "headers": {...},
		//   "data": { ... message details ... }
		// }

		if data, ok := eventData["data"].(map[string]interface{}); ok {
			contentStr, _ := data["content"].(string)
			// Content is often a JSON string itself
			var contentMap map[string]interface{}
			json.Unmarshal([]byte(contentStr), &contentMap)

			text := ""
			if t, ok := contentMap["text"].(string); ok {
				text = t
			}

			// Sender
			senderID := ""
			if sender, ok := data["sender"].(map[string]interface{}); ok {
				senderID, _ = sender["sender_id"].(string) // UnionID or StaffID
			}

			// Conversation
			groupID := ""
			if cid, ok := data["conversation_id"].(string); ok {
				groupID = cid
			}

			log.Printf("Parsed Message: [%s] %s", senderID, text)

			sendToNexus(map[string]interface{}{
				"post_type":    "message",
				"message_type": "group", // Default to group/chat
				"time":         time.Now().Unix(),
				"self_id":      config.SelfID,
				"sub_type":     "normal",
				"message_id":   getString(data, "message_id"),
				"user_id":      senderID,
				"group_id":     groupID,
				"message":      text,
				"raw_message":  text,
				"sender": map[string]interface{}{
					"user_id":  senderID,
					"nickname": "DingTalkUser",
				},
			})
			return
		}
	}

	// Forward other events or if parsing failed
	sendToNexus(map[string]interface{}{
		"post_type":   "notice",
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
		header.Add("X-Self-ID", config.SelfID)
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
	nexusMu.Lock()
	defer nexusMu.Unlock()

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
		groupID := getString(params, "group_id")

		if msg != "" {
			var err error
			var msgID string

			if config.AccessToken != "" {
				err = sendDingTalkMessage(msg)
			} else if config.ClientID != "" {
				// Enterprise Mode
				if groupID == "" {
					err = fmt.Errorf("group_id required for enterprise group message")
				} else {
					msgID, err = sendEnterpriseGroupMessage(groupID, msg)
				}
			} else {
				err = fmt.Errorf("webhook access_token not configured")
			}

			if err == nil {
				data := map[string]interface{}{}
				if msgID != "" {
					data["message_id"] = msgID
				}
				sendToNexus(map[string]interface{}{"status": "ok", "data": data, "echo": action["echo"]})
			} else {
				sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": action["echo"]})
			}
		}

	case "delete_msg":
		params, _ := action["params"].(map[string]interface{})
		msgID := getString(params, "message_id")
		// DingTalk recall requires conversation ID too?
		// "recall/group/message" needs "openConversationId" and "processQueryKey" (msgID)
		// We don't have conversation ID in delete_msg params from generic logic.
		// However, processQueryKey might be unique enough or we might need to store it?
		// Actually, let's see if we can recall with just msgID or if we need to encode groupID in msgID.

		// Strategy: Encode groupID in msgID -> "groupID|processQueryKey"
		if msgID != "" && config.ClientID != "" {
			// Try to split
			// If simpler approach: user passes group_id in params? No, BotNexus generic logic doesn't send it.
			// So we MUST encode it.

			// If encoded "groupID|msgID"
			// But wait, if I change return ID format, I need to ensure it doesn't break anything.
			// It should be fine as it's just a string token.

			err := recallEnterpriseMessage(msgID)
			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
			} else {
				sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": action["echo"]})
			}
		} else {
			sendToNexus(map[string]interface{}{"status": "failed", "message": "recall not supported or invalid id", "echo": action["echo"]})
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
			} else if config.ClientID != "" {
				err = sendEnterprisePrivateMessage(userID, msg)
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

// --- Enterprise Robot API (Stream Mode) ---

type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

var (
	enterpriseToken string
	tokenExpiry     time.Time
)

func getEnterpriseAccessToken() (string, error) {
	if enterpriseToken != "" && time.Now().Before(tokenExpiry) {
		return enterpriseToken, nil
	}

	url := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", config.ClientID, config.ClientSecret)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("get token error: %s", result.ErrMsg)
	}

	enterpriseToken = result.AccessToken
	tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-200) * time.Second)
	return enterpriseToken, nil
}

func sendEnterpriseGroupMessage(conversationID, content string) (string, error) {
	token, err := getEnterpriseAccessToken()
	if err != nil {
		return "", err
	}

	url := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"

	msgParam := map[string]string{"content": content}
	msgParamBytes, _ := json.Marshal(msgParam)

	payload := map[string]interface{}{
		"robotCode":          config.ClientID,
		"openConversationId": conversationID,
		"msgKey":             "sampleText",
		"msgParam":           string(msgParamBytes),
	}

	resp, err := postToDingTalkAPI(url, token, payload)
	if err != nil {
		return "", err
	}

	// Get processQueryKey
	if key, ok := resp["processQueryKey"].(string); ok {
		// Encode conversationID for recall: "cid|key"
		return fmt.Sprintf("%s|%s", conversationID, key), nil
	}
	return "", nil
}

func recallEnterpriseMessage(encodedID string) error {
	// Split "cid|key"
	parts := strings.Split(encodedID, "|")
	if len(parts) != 2 {
		return fmt.Errorf("invalid message_id format for recall")
	}
	conversationID := parts[0]
	processQueryKey := parts[1]

	token, err := getEnterpriseAccessToken()
	if err != nil {
		return err
	}

	url := "https://api.dingtalk.com/v1.0/robot/groupMessages/recall"

	payload := map[string]interface{}{
		"robotCode":          config.ClientID,
		"openConversationId": conversationID,
		"processQueryKey":    processQueryKey,
	}

	_, err = postToDingTalkAPI(url, token, payload)
	return err
}

func sendEnterprisePrivateMessage(userID, content string) error {
	token, err := getEnterpriseAccessToken()
	if err != nil {
		return err
	}

	// Using batchSend for single user
	url := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"

	msgParam := map[string]string{"content": content}
	msgParamBytes, _ := json.Marshal(msgParam)

	payload := map[string]interface{}{
		"robotCode": config.ClientID,
		"userIds":   []string{userID},
		"msgKey":    "sampleText",
		"msgParam":  string(msgParamBytes),
	}

	_, err = postToDingTalkAPI(url, token, payload)
	return err
}

func postToDingTalkAPI(url, token string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result, nil
}
