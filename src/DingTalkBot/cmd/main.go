package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"BotMatrix/common/bot"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
)

// DingTalkConfig extends bot.BotConfig with DingTalk specific fields
type DingTalkConfig struct {
	bot.BotConfig
	// Webhook Mode (Custom Robot)
	AccessToken string `json:"access_token"`
	Secret      string `json:"secret"` // Optional: for HMAC signature

	// Stream Mode (Enterprise Robot)
	ClientID     string `json:"client_id"`     // AppKey
	ClientSecret string `json:"client_secret"` // AppSecret

	SelfID string `json:"self_id"` // Optional: manually set SelfID
}

var (
	botService   *bot.BaseBot
	dingTalkCfg  DingTalkConfig
	streamClient *client.StreamClient
	botCtx       context.Context
	botCancel    context.CancelFunc
	httpClient   = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	botService = bot.NewBaseBot(8088)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("DingTalkBot", &dingTalkCfg, restartBot, []bot.ConfigSection{
		{
			Title: "Webhook 模式 (自定义机器人)",
			Fields: []bot.ConfigField{
				{Label: "Access Token", ID: "access_token", Type: "text", Value: dingTalkCfg.AccessToken},
				{Label: "Secret (可选)", ID: "secret", Type: "password", Value: dingTalkCfg.Secret},
			},
		},
		{
			Title: "Stream 模式 (企业机器人)",
			Fields: []bot.ConfigField{
				{Label: "AppKey (Client ID)", ID: "client_id", Type: "text", Value: dingTalkCfg.ClientID},
				{Label: "AppSecret (Client Secret)", ID: "client_secret", Type: "password", Value: dingTalkCfg.ClientSecret},
			},
		},
		{
			Title: "连接与服务配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: dingTalkCfg.NexusAddr},
				{Label: "Web UI 端口 (LogPort)", ID: "log_port", Type: "number", Value: dingTalkCfg.LogPort},
				{Label: "SelfID (可选)", ID: "self_id", Type: "text", Value: dingTalkCfg.SelfID},
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

	// Sync common config to local dingTalkCfg
	botService.Mu.RLock()
	dingTalkCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load DingTalk specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &dingTalkCfg)
	}

	// Environment variable overrides
	if envToken := os.Getenv("DINGTALK_TOKEN"); envToken != "" {
		dingTalkCfg.AccessToken = envToken
	}
	if envSecret := os.Getenv("DINGTALK_SECRET"); envSecret != "" {
		dingTalkCfg.Secret = envSecret
	}
	if envClientID := os.Getenv("DINGTALK_CLIENT_ID"); envClientID != "" {
		dingTalkCfg.ClientID = envClientID
	}
	if envClientSecret := os.Getenv("DINGTALK_CLIENT_SECRET"); envClientSecret != "" {
		dingTalkCfg.ClientSecret = envClientSecret
	}

	// Generate a SelfID if not set
	if dingTalkCfg.SelfID == "" {
		key := dingTalkCfg.AccessToken
		if key == "" {
			key = dingTalkCfg.ClientID
		}
		if key == "" {
			key = "dingtalk_bot"
		}
		h := sha256.New()
		h.Write([]byte(key))
		bs := h.Sum(nil)
		dingTalkCfg.SelfID = fmt.Sprintf("%x", bs[:4])
		log.Printf("Auto-generated SelfID: %s", dingTalkCfg.SelfID)
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Sync botService.Config from dingTalkCfg
	botService.Config = dingTalkCfg.BotConfig
	clientID := dingTalkCfg.ClientID
	clientSecret := dingTalkCfg.ClientSecret
	nexusAddr := botService.Config.NexusAddr
	selfID := dingTalkCfg.SelfID
	botService.Mu.RUnlock()

	botCtx, botCancel = context.WithCancel(context.Background())

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "DingTalk", selfID, handleNexusCommand)

	// Start Stream Client if configured
	if clientID != "" && clientSecret != "" {
		go startStreamClient(botCtx)
	} else {
		log.Println("Stream Mode not configured. Running in Webhook Send-Only mode.")
	}
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func startStreamClient(ctx context.Context) {
	cli := client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(dingTalkCfg.ClientID, dingTalkCfg.ClientSecret)),
		client.WithUserAgent(client.NewDingtalkGoSDKUserAgent()),
		client.WithSubscription("EVENT", "*", func(ctx context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
			handleStreamEvent(df)
			return payload.NewSuccessDataFrameResponse(), nil
		}),
	)

	err := cli.Start(ctx)
	if err != nil {
		log.Printf("Stream Client failed to start: %v", err)
		return
	}
	streamClient = cli
	log.Println("Stream Client started successfully!")

	<-ctx.Done()
	log.Println("Stream Client stopping...")
}

func handleStreamEvent(df *payload.DataFrame) {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(df.Data), &eventData); err != nil {
		return
	}

	eventType, _ := eventData["type"].(string)

	if eventType == "im.message.receive_v1" {
		if data, ok := eventData["data"].(map[string]interface{}); ok {
			contentStr, _ := data["content"].(string)
			var contentMap map[string]interface{}
			json.Unmarshal([]byte(contentStr), &contentMap)

			text := ""
			if t, ok := contentMap["text"].(string); ok {
				text = t
			}

			senderID := ""
			if sender, ok := data["sender"].(map[string]interface{}); ok {
				senderID, _ = sender["sender_id"].(string)
			}

			groupID := ""
			if cid, ok := data["conversation_id"].(string); ok {
				groupID = cid
			}

			obMsg := map[string]interface{}{
				"post_type":    "message",
				"message_type": "group",
				"time":         time.Now().Unix(),
				"self_id":      dingTalkCfg.SelfID,
				"sub_type":     "normal",
				"message_id":   data["message_id"],
				"user_id":      senderID,
				"group_id":     groupID,
				"message":      text,
				"raw_message":  text,
				"sender": map[string]interface{}{
					"user_id":  senderID,
					"nickname": "DingTalkUser",
				},
			}
			botService.SendToNexus(obMsg)
			return
		}
	}

	botService.SendToNexus(map[string]interface{}{
		"post_type":   "notice",
		"notice_type": "dingtalk_event",
		"sub_type":    eventType,
		"raw_data":    eventData,
		"self_id":     dingTalkCfg.SelfID,
		"time":        time.Now().Unix(),
	})
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

	switch cmd.Action {
	case "send_msg", "send_group_msg", "send_private_msg":
		text, _ := cmd.Params["message"].(string)
		if text != "" {
			sendDingTalkMessage(text, cmd.Echo)
		}
	case "get_login_info":
		botService.SendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  dingTalkCfg.SelfID,
				"nickname": "DingTalkBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendDingTalkMessage(text, echo string) {
	botService.Mu.RLock()
	accessToken := dingTalkCfg.AccessToken
	secret := dingTalkCfg.Secret
	botService.Mu.RUnlock()

	if accessToken == "" {
		return
	}

	apiURL := "https://oapi.dingtalk.com/robot/send?access_token=" + accessToken
	if secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(stringToSign))
		signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
		apiURL += fmt.Sprintf("&timestamp=%d&sign=%s", timestamp, url.QueryEscape(signature))
	}

	msg := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}
	payload, _ := json.Marshal(msg)

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to send DingTalk message: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data":   result,
		"echo":   echo,
	})
}
