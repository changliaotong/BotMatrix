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

	common "BotMatrix/src/Common"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
)

// DingTalkConfig extends common.BotConfig with DingTalk specific fields
type DingTalkConfig struct {
	common.BotConfig
	// Webhook Mode (Custom Robot)
	AccessToken string `json:"access_token"`
	Secret      string `json:"secret"` // Optional: for HMAC signature

	// Stream Mode (Enterprise Robot)
	ClientID     string `json:"client_id"`     // AppKey
	ClientSecret string `json:"client_secret"` // AppSecret

	SelfID string `json:"self_id"` // Optional: manually set SelfID
}

var (
	botService   *common.BaseBot
	dingTalkCfg  DingTalkConfig
	streamClient *client.StreamClient
	botCtx       context.Context
	botCancel    context.CancelFunc
	httpClient   = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	botService = common.NewBaseBot(8088)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	botService.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config-ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	botService.Mux.HandleFunc("/config", handleConfig)
	botService.Mux.HandleFunc("/config-ui", handleConfigUI)

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

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dingTalkCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig DingTalkConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		dingTalkCfg = newConfig
		botService.Config = newConfig.BotConfig
		botService.Mu.Unlock()

		botService.SaveConfig("config.json")
		go restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarted"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := dingTalkCfg
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>DingTalkBot 配置中心</title>
    <style>
        :root { --primary-color: #007bff; --bg-color: #f4f7f6; }
        body { font-family: sans-serif; background: var(--bg-color); margin: 0; padding: 20px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); max-width: 600px; margin: 0 auto; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; color: #666; }
        input { width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { background: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; width: 100%%; }
        h1 { font-size: 1.5rem; text-align: center; }
    </style>
</head>
<body>
    <div class="card">
        <h1>DingTalkBot 配置</h1>
        <div class="form-group">
            <label>Access Token (Webhook)</label>
            <input type="text" id="access_token" value="%s">
        </div>
        <div class="form-group">
            <label>Secret (Webhook Signature)</label>
            <input type="text" id="secret" value="%s">
        </div>
        <div class="form-group">
            <label>AppKey (Stream Mode)</label>
            <input type="text" id="client_id" value="%s">
        </div>
        <div class="form-group">
            <label>AppSecret (Stream Mode)</label>
            <input type="password" id="client_secret" value="%s">
        </div>
        <div class="form-group">
            <label>Nexus 地址</label>
            <input type="text" id="nexus_addr" value="%s">
        </div>
        <div class="form-group">
            <label>SelfID</label>
            <input type="text" id="self_id" value="%s">
        </div>
        <button onclick="saveConfig()">保存并重启</button>
    </div>
    <script>
        async function saveConfig() {
            const cfg = {
                access_token: document.getElementById('access_token').value,
                secret: document.getElementById('secret').value,
                client_id: document.getElementById('client_id').value,
                client_secret: document.getElementById('client_secret').value,
                nexus_addr: document.getElementById('nexus_addr').value,
                self_id: document.getElementById('self_id').value,
                log_port: %d
            };
            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(cfg)
            });
            if (resp.ok) alert('配置已保存，机器人正在重启...');
            else alert('保存失败');
        }
    </script>
</body>
</html>`, cfg.AccessToken, cfg.Secret, cfg.ClientID, cfg.ClientSecret, cfg.NexusAddr, cfg.SelfID, cfg.LogPort)
}
