package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	common "BotMatrix/src/Common"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuConfig extends common.BotConfig with Feishu specific fields
type FeishuConfig struct {
	common.BotConfig
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	EncryptKey        string `json:"encrypt_key"`
	VerificationToken string `json:"verification_token"`
}

var (
	botService *common.BaseBot
	larkClient *lark.Client
	botCtx     context.Context
	botCancel  context.CancelFunc
	feishuCfg  FeishuConfig
)

func main() {
	botService = common.NewBaseBot(3135)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initial load of config into local feishuCfg
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

	// Sync common config to local feishuCfg
	botService.Mu.RLock()
	feishuCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Feishu specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &feishuCfg)
	}

	// Environment variable overrides for Feishu
	if envAppID := os.Getenv("APP_ID"); envAppID != "" {
		feishuCfg.AppID = envAppID
	}
	if envAppSecret := os.Getenv("APP_SECRET"); envAppSecret != "" {
		feishuCfg.AppSecret = envAppSecret
	}
	if envEncryptKey := os.Getenv("ENCRYPT_KEY"); envEncryptKey != "" {
		feishuCfg.EncryptKey = envEncryptKey
	}
	if envVerToken := os.Getenv("VERIFICATION_TOKEN"); envVerToken != "" {
		feishuCfg.VerificationToken = envVerToken
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	appID := feishuCfg.AppID
	appSecret := feishuCfg.AppSecret
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if appID == "" || appSecret == "" {
		log.Println("WARNING: Feishu AppID or AppSecret is not configured. Bot will not start until configured via Web UI.")
		return
	}

	// Initialize Lark API Client
	larkClient = lark.NewClient(appID, appSecret)

	botCtx, botCancel = context.WithCancel(context.Background())

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Feishu", appID, handleNexusCommand)

	// Start Lark WebSocket Client (Receive Events)
	go startLarkWS(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func startLarkWS(ctx context.Context) {
	botService.Mu.RLock()
	verToken := feishuCfg.VerificationToken
	encryptKey := feishuCfg.EncryptKey
	appID := feishuCfg.AppID
	appSecret := feishuCfg.AppSecret
	botService.Mu.RUnlock()

	// Register Event Handler
	eventHandler := larkevent.NewEventDispatcher(verToken, encryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			handleMessage(ctx, event)
			return nil
		})

	// Create WebSocket Client
	cli := larkws.NewClient(appID, appSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// Start WebSocket Client
	log.Println("Connecting to Feishu/Lark Gateway...")
	err := cli.Start(ctx)
	if err != nil {
		log.Printf("Failed to start Feishu WebSocket client: %v", err)
	}
}

func handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msg := event.Event.Message
	sender := event.Event.Sender

	// Extract Content (JSON string)
	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(*msg.Content), &contentMap); err != nil {
		log.Printf("Error parsing message content: %v", err)
		return
	}

	// Message Content Parsing
	text := ""
	switch *msg.MessageType {
	case "text":
		if t, ok := contentMap["text"].(string); ok {
			text = t
		}
	case "image":
		if key, ok := contentMap["image_key"].(string); ok {
			text = fmt.Sprintf("[CQ:image,file=%s]", key)
		}
	case "file":
		if key, ok := contentMap["file_key"].(string); ok {
			text = fmt.Sprintf("[CQ:file,file=%s]", key)
		}
	case "audio":
		if key, ok := contentMap["file_key"].(string); ok {
			text = fmt.Sprintf("[CQ:record,file=%s]", key)
		}
	case "post":
		// Simplified rich text handling
		text = "[RichText Message]"
		if title, ok := contentMap["title"].(string); ok {
			text += " " + title
		}
	default:
		text = fmt.Sprintf("[Unknown Type: %s]", *msg.MessageType)
	}

	// Determine Chat Type (group/p2p)
	msgType := "private"
	groupID := ""
	if *msg.ChatType == "group" {
		msgType = "group"
		groupID = *msg.ChatId
	}

	// Determine User ID (prefer open_id)
	userID := *sender.SenderId.OpenId

	log.Printf("Received Message: [%s] %s: %s", msgType, userID, text)

	// Construct OneBot Message
	obMsg := map[string]interface{}{
		"post_type":    "message",
		"message_type": msgType,
		"time":         time.Now().Unix(),
		"self_id":      feishuCfg.AppID,
		"sub_type":     "normal",
		"message_id":   *msg.MessageId,
		"user_id":      userID,
		"group_id":     groupID,
		"message":      text,
		"raw_message":  text,
		"sender": map[string]interface{}{
			"user_id":  userID,
			"nickname": "FeishuUser",
		},
	}

	botService.SendToNexus(obMsg)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(feishuCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig FeishuConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		feishuCfg = newConfig
		botService.Config = newConfig.BotConfig
		botService.Mu.Unlock()

		// Save to file
		data, _ := json.MarshalIndent(feishuCfg, "", "  ")
		os.WriteFile("config.json", data, 0644)

		go restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarted"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := feishuCfg
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FeishuBot 配置中心</title>
    <style>
        :root { --primary-color: #1a73e8; --success-color: #28a745; --danger-color: #dc3545; --bg-color: #f4f7f6; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: var(--bg-color); margin: 0; display: flex; height: 100vh; }
        .sidebar { width: 280px; background: #2c3e50; color: white; display: flex; flex-direction: column; }
        .sidebar-header { padding: 20px; font-size: 20px; font-weight: bold; border-bottom: 1px solid #34495e; }
        .nav-item { padding: 15px 20px; cursor: pointer; transition: background 0.2s; display: flex; align-items: center; gap: 10px; }
        .nav-item:hover { background: #34495e; }
        .nav-item.active { background: var(--primary-color); }
        .main-content { flex: 1; overflow-y: auto; padding: 30px; }
        .card { background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); padding: 25px; margin-bottom: 25px; }
        .section-title { font-size: 18px; font-weight: 600; margin-bottom: 20px; color: #2c3e50; display: flex; justify-content: space-between; align-items: center; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; color: #666; }
        input[type="text"], input[type="number"], input[type="password"], select { 
            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; 
        }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: 500; transition: opacity 0.2s; }
        .btn-primary { background: var(--primary-color); color: white; }
        .btn-success { background: var(--success-color); color: white; }
        .btn-danger { background: var(--danger-color); color: white; }
        .btn:hover { opacity: 0.9; }
        .logs-container { background: #1e1e1e; color: #d4d4d4; padding: 15px; border-radius: 6px; font-family: 'Consolas', monospace; height: 500px; overflow-y: auto; font-size: 13px; line-height: 1.5; }
        .log-line { margin-bottom: 4px; border-bottom: 1px solid #333; padding-bottom: 2px; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">FeishuBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">飞书 API 配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>App ID</label>
                        <input type="text" id="app_id" value="%s">
                    </div>
                    <div class="form-group">
                        <label>App Secret</label>
                        <input type="password" id="app_secret" value="%s">
                    </div>
                    <div class="form-group">
                        <label>Encrypt Key (可选)</label>
                        <input type="password" id="encrypt_key" value="%s">
                    </div>
                    <div class="form-group">
                        <label>Verification Token (可选)</label>
                        <input type="password" id="verification_token" value="%s">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">连接与服务配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>BotNexus 地址</label>
                        <input type="text" id="nexus_addr" value="%s">
                    </div>
                    <div class="form-group">
                        <label>Web UI 端口 (LogPort)</label>
                        <input type="number" id="log_port" value="%%d">
                    </div>
                </div>
            </div>

            <div style="text-align: center; margin-top: 30px;">
                <button class="btn btn-primary" style="padding: 15px 40px; font-size: 16px;" onclick="saveConfig()">保存配置并重启</button>
            </div>
        </div>

        <div id="logs-tab" style="display: none;">
            <div class="card">
                <div class="section-title">
                    系统日志
                    <button class="btn btn-danger" onclick="clearLogs()">清空显示</button>
                </div>
                <div id="logs" class="logs-container">正在加载日志...</div>
            </div>
        </div>
    </div>

    <script>
        let currentTab = 'config';
        function switchTab(tab) {
            document.getElementById(currentTab + '-tab').style.display = 'none';
            document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
            
            document.getElementById(tab + '-tab').style.display = 'block';
            event.currentTarget.classList.add('active');
            currentTab = tab;
            if (tab === 'logs') loadLogs();
        }

        async function saveConfig() {
            const cfg = {
                app_id: document.getElementById('app_id').value,
                app_secret: document.getElementById('app_secret').value,
                encrypt_key: document.getElementById('encrypt_key').value,
                verification_token: document.getElementById('verification_token').value,
                nexus_addr: document.getElementById('nexus_addr').value,
                log_port: parseInt(document.getElementById('log_port').value)
            };

            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(cfg)
            });

            if (resp.ok) {
                alert('配置已保存，机器人正在重启...');
                setTimeout(() => window.location.reload(), 3000);
            } else {
                const err = await resp.text();
                alert('保存失败: ' + err);
            }
        }

        async function loadLogs() {
            if (currentTab !== 'logs') return;
            try {
                const resp = await fetch('/logs?lines=200');
                const logs = await resp.json();
                const logsDiv = document.getElementById('logs');
                logsDiv.innerHTML = logs.map(line => `+"`"+`<div class="log-line">${line}</div>`+"`"+`).join('');
                logsDiv.scrollTop = logsDiv.scrollHeight;
            } catch (e) {}
            setTimeout(loadLogs, 2000);
        }

        function clearLogs() {
            document.getElementById('logs').innerText = '';
        }
    </script>
</body>
</html>
`, cfg.AppID, cfg.AppSecret, cfg.EncryptKey, cfg.VerificationToken, cfg.NexusAddr, cfg.LogPort)
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
		Echo   string                 `json:"echo"`
	}
	if err := json.Unmarshal(data, &cmd); err != nil {
		log.Printf("Invalid Nexus command: %v", err)
		return
	}

	log.Printf("Received Command: %s", cmd.Action)

	switch cmd.Action {
	case "send_group_msg", "send_msg":
		groupID, _ := cmd.Params["group_id"].(string)
		message, _ := cmd.Params["message"].(string)
		if groupID != "" && message != "" {
			sendFeishuMessage(groupID, "chat_id", message, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string) // open_id
		message, _ := cmd.Params["message"].(string)
		if userID != "" && message != "" {
			sendFeishuMessage(userID, "open_id", message, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteFeishuMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		botService.SendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  feishuCfg.AppID,
				"nickname": "FeishuBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendFeishuMessage(receiveID, receiveIDType, text, echo string) {
	if larkClient == nil {
		return
	}

	content := map[string]string{
		"text": text,
	}
	contentJSON, _ := json.Marshal(content)

	resp, err := larkClient.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeByMsgTypeText).
			ReceiveId(receiveID).
			Content(string(contentJSON)).
			Build()).
		Build())

	if err != nil {
		log.Printf("Failed to send Feishu message: %v", err)
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API error: %d - %s", resp.Code, resp.Msg)
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	botService.SendToNexus(map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"message_id": *resp.Data.MessageId,
		},
		"echo": echo,
	})
}

func deleteFeishuMessage(messageID, echo string) {
	if larkClient == nil {
		return
	}

	resp, err := larkClient.Im.Message.Delete(context.Background(), larkim.NewDeleteMessageReqBuilder().
		MessageId(messageID).
		Build())

	if err != nil {
		botService.LogManager.Error().Str("action", "delete_msg").Err(err).Str("message_id", messageID).Msg("Failed to delete Feishu message")
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	botService.SendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
}
