package main

import (
	"BotMatrix/common/log"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"go.uber.org/zap"
)

// Config holds the bot configuration
type Config struct {
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	EncryptKey        string `json:"encrypt_key"`
	VerificationToken string `json:"verification_token"`
	NexusAddr         string `json:"nexus_addr"`
	LogPort           int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	larkClient  *lark.Client
	nexusConn   *websocket.Conn
	connMutex   sync.Mutex
	selfID      string // AppID as SelfID

	nexusCtx    context.Context
	nexusCancel context.CancelFunc
	botCtx      context.Context
	botCancel   context.CancelFunc

	logManager = NewLogManager(1000)
)

// LogManager handles log rotation and retrieval
type LogManager struct {
	logs  []string
	max   int
	mutex sync.Mutex
}

func NewLogManager(max int) *LogManager {
	return &LogManager{
		logs: make([]string, 0, max),
		max:  max,
	}
}

func (m *LogManager) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	line := string(p)
	m.logs = append(m.logs, line)
	if len(m.logs) > m.max {
		m.logs = m.logs[len(m.logs)-m.max:]
	}
	return os.Stderr.Write(p)
}

func (m *LogManager) GetLogs(lines int) []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if lines > len(m.logs) {
		lines = len(m.logs)
	}
	result := make([]string, lines)
	copy(result, m.logs[len(m.logs)-lines:])
	return result
}

// OneBotMessage represents a OneBot 11 message
type OneBotMessage struct {
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type"`
	Time        int64       `json:"time"`
	SelfID      string      `json:"self_id"` // OneBot uses int64 usually, but string is safer for AppID
	SubType     string      `json:"sub_type"`
	MessageID   string      `json:"message_id"`
	UserID      string      `json:"user_id"`
	GroupID     string      `json:"group_id,omitempty"`
	Message     any         `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Sender      Sender      `json:"sender"`
}

type Sender struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

func main() {
	// 初始化日志系统
	log.InitDefaultLogger()

	// Load Configuration
	loadConfig()

	// Ensure LogPort is set
	configMutex.Lock()
	if config.LogPort == 0 {
		config.LogPort = 3135 // Default for FeishuBot
	}
	configMutex.Unlock()

	startBot()

	// Start HTTP Server for Web UI and Logs
	go startHTTPServer()

	// Wait for signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	stopBot()
}

func loadConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.ReadFile("config.json")
	if err == nil {
		if err := json.Unmarshal(file, &config); err != nil {
			log.Printf("Error parsing config.json: %v", err)
		}
	} else {
		log.Println("config.json not found, using environment variables or defaults")
	}

	if envAppID := os.Getenv("APP_ID"); envAppID != "" {
		config.AppID = envAppID
	}
	if envAppSecret := os.Getenv("APP_SECRET"); envAppSecret != "" {
		config.AppSecret = envAppSecret
	}
	if envNexusAddr := os.Getenv("NEXUS_ADDR"); envNexusAddr != "" {
		config.NexusAddr = envNexusAddr
	}
	if envEncryptKey := os.Getenv("ENCRYPT_KEY"); envEncryptKey != "" {
		config.EncryptKey = envEncryptKey
	}
	if envVerToken := os.Getenv("VERIFICATION_TOKEN"); envVerToken != "" {
		config.VerificationToken = envVerToken
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	selfID = config.AppID
}

func startBot() {
	botCtx, botCancel = context.WithCancel(context.Background())

	configMutex.RLock()
	appID := config.AppID
	appSecret := config.AppSecret
	configMutex.RUnlock()

	if config.AppID == "" || config.AppSecret == "" {
		log.Warn("Feishu AppID or AppSecret is not configured. Bot will not start until configured via Web UI.")
		return
	}

	// Initialize Lark API Client
	larkClient = lark.NewClient(appID, appSecret)

	startNexus()

	// Start Lark WebSocket Client (Receive Events)
	go startLarkWS(botCtx)
}

func stopBot() {
	stopNexus()
	if botCancel != nil {
		botCancel()
	}
}

func startNexus() {
	nexusCtx, nexusCancel = context.WithCancel(botCtx)
	go connectToNexus(nexusCtx)
}

func stopNexus() {
	if nexusCancel != nil {
		nexusCancel()
	}
	connMutex.Lock()
	if nexusConn != nil {
		nexusConn.Close()
		nexusConn = nil
	}
	connMutex.Unlock()
}

func startLarkWS(ctx context.Context) {
	configMutex.RLock()
	verToken := config.VerificationToken
	encryptKey := config.EncryptKey
	appID := config.AppID
	appSecret := config.AppSecret
	configMutex.RUnlock()

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
	log.Info("Connecting to Feishu/Lark Gateway...")
	err := cli.Start(ctx)
	if err != nil {
		log.Error("Failed to start Feishu WebSocket client", zap.Error(err))
	}
}

func handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msg := event.Event.Message
	sender := event.Event.Sender

	// Extract Content (JSON string)
	var contentMap map[string]any
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
	obMsg := OneBotMessage{
		PostType:    "message",
		MessageType: msgType,
		Time:        time.Now().Unix(),
		SelfID:      selfID,
		SubType:     "normal",
		MessageID:   *msg.MessageId,
		UserID:      userID,
		GroupID:     groupID,
		Message:     text,
		RawMessage:  text,
		Sender: Sender{
			UserID:   userID,
			Nickname: "FeishuUser", // Feishu doesn't provide nickname in message event directly usually
		},
	}

	sendToNexus(obMsg)
}

func connectToNexus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Nexus connection stopped.")
			return
		default:
			configMutex.RLock()
			addr := config.NexusAddr
			configMutex.RUnlock()

			log.Printf("Connecting to BotNexus at %s...", addr)
			header := http.Header{}
			header.Add("X-Self-ID", selfID)
			header.Add("X-Platform", "Feishu")

			conn, _, err := websocket.DefaultDialer.Dial(addr, header)
			if err != nil {
				log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			connMutex.Lock()
			nexusConn = conn
			connMutex.Unlock()
			log.Println("Connected to BotNexus!")

			// Send Lifecycle Event
			sendToNexus(map[string]any{
				"post_type":       "meta_event",
				"meta_event_type": "lifecycle",
				"sub_type":        "connect",
				"self_id":         selfID,
				"time":            time.Now().Unix(),
			})

			// Handle incoming commands from Nexus
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("BotNexus disconnected: %v", err)
					connMutex.Lock()
					if nexusConn == conn {
						nexusConn = nil
					}
					connMutex.Unlock()
					break
				}
				handleNexusCommand(message)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func sendToNexus(msg any) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send to Nexus: %v", err)
		nexusConn = nil
	}
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config-ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		lines := 100
		if l := r.URL.Query().Get("lines"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				lines = v
			}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		logs := logManager.GetLogs(lines)
		for _, line := range logs {
			fmt.Fprint(w, line)
		}
	})

	mux.HandleFunc("/config", handleConfig)
	mux.HandleFunc("/config-ui", handleConfigUI)

	configMutex.RLock()
	port := config.LogPort
	configMutex.RUnlock()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP Server at http://localhost%s (UI: /config-ui, Logs: /logs)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Failed to start HTTP Server: %v", err)
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		configMutex.RLock()
		defer configMutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save to file
		data, err := json.MarshalIndent(newConfig, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := os.WriteFile("config.json", data, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Restart bot
		stopBot()
		configMutex.Lock()
		config = newConfig
		configMutex.Unlock()
		startBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarted"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
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
                        <input type="text" id="app_id">
                    </div>
                    <div class="form-group">
                        <label>App Secret</label>
                        <input type="password" id="app_secret">
                    </div>
                    <div class="form-group">
                        <label>Encrypt Key (可选)</label>
                        <input type="password" id="encrypt_key">
                    </div>
                    <div class="form-group">
                        <label>Verification Token (可选)</label>
                        <input type="password" id="verification_token">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">连接与服务配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>BotNexus 地址</label>
                        <input type="text" id="nexus_addr">
                    </div>
                    <div class="form-group">
                        <label>Web UI 端口 (LogPort)</label>
                        <input type="number" id="log_port">
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
                <div id="logs" class="logs-container"></div>
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
        }

        async function loadConfig() {
            const resp = await fetch('/config');
            const cfg = await resp.json();
            
            document.getElementById('app_id').value = cfg.app_id || '';
            document.getElementById('app_secret').value = cfg.app_secret || '';
            document.getElementById('encrypt_key').value = cfg.encrypt_key || '';
            document.getElementById('verification_token').value = cfg.verification_token || '';
            document.getElementById('nexus_addr').value = cfg.nexus_addr || '';
            document.getElementById('log_port').value = cfg.log_port || 0;
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

        async function updateLogs() {
            if (currentTab !== 'logs') return;
            try {
                const resp = await fetch('/logs?lines=100');
                const text = await resp.text();
                const logsDiv = document.getElementById('logs');
                logsDiv.innerText = text;
                logsDiv.scrollTop = logsDiv.scrollHeight;
            } catch (e) {}
        }

        function clearLogs() {
            document.getElementById('logs').innerText = '';
        }

        setInterval(updateLogs, 2000);
        loadConfig();
    </script>
</body>
</html>
	`)
}

// NexusCommand represents a command from BotNexus (OneBot Action)
type NexusCommand struct {
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
	Echo   string         `json:"echo"`
}

func handleNexusCommand(data []byte) {
	var cmd NexusCommand
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
		sendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  selfID,
				"nickname": "FeishuBot",
			},
			"echo": cmd.Echo,
		})
	case "get_group_list":
		getGroupList(cmd.Echo)
	case "get_group_member_list":
		groupID, _ := cmd.Params["group_id"].(string)
		if groupID != "" {
			getGroupMemberList(groupID, cmd.Echo)
		}
	}
}

func sendFeishuMessage(receiveID, receiveIDType, text, echo string) {
	// Simple text vs image handling could be added here
	// For now, assuming text
	content := map[string]any{
		"text": text,
	}
	contentJSON, _ := json.Marshal(content)
	contentStr := string(contentJSON)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveID).
			MsgType(larkim.MsgTypeText).
			Content(contentStr).
			Build()).
		Build()

	resp, err := larkClient.Im.Message.Create(context.Background(), req)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
	} else {
		log.Printf("Sent message to %s: %s", receiveID, text)
		sendToNexus(map[string]any{"status": "ok", "data": map[string]any{"message_id": *resp.Data.MessageId}, "echo": echo})
	}
}

func deleteFeishuMessage(messageID, echo string) {
	req := larkim.NewDeleteMessageReqBuilder().
		MessageId(messageID).
		Build()

	resp, err := larkClient.Im.Message.Delete(context.Background(), req)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Delete Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
	} else {
		log.Printf("Deleted message: %s", messageID)
		sendToNexus(map[string]any{"status": "ok", "echo": echo})
	}
}

func getGroupMemberList(chatID, echo string) {
	req := larkim.NewGetChatMembersReqBuilder().
		ChatId(chatID).
		PageSize(100).
		Build()

	resp, err := larkClient.Im.ChatMembers.Get(context.Background(), req)
	if err != nil {
		log.Printf("Failed to get group member list: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Group Member List Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	var members []map[string]any
	for _, member := range resp.Data.Items {
		members = append(members, map[string]any{
			"user_id":  *member.MemberId,
			"nickname": "FeishuMember", // Names need a separate lookup or GetChat
			"role":     "member",
		})
	}

	sendToNexus(map[string]any{
		"status": "ok",
		"data":   members,
		"echo":   echo,
	})
}

func getGroupList(echo string) {
	// Iterate pages if needed, but for now just one page
	req := larkim.NewListChatReqBuilder().
		SortType("ByCreateTimeAsc").
		PageSize(20).
		Build()

	resp, err := larkClient.Im.Chat.List(context.Background(), req)
	if err != nil {
		log.Printf("Failed to get group list: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	if !resp.Success() {
		log.Printf("Feishu API Group List Error: %d %s", resp.Code, resp.Msg)
		sendToNexus(map[string]any{"status": "failed", "message": resp.Msg, "echo": echo})
		return
	}

	var groups []map[string]any
	for _, chat := range resp.Data.Items {
		groups = append(groups, map[string]any{
			"group_id":   *chat.ChatId,
			"group_name": *chat.Name,
		})
	}

	sendToNexus(map[string]any{
		"status": "ok",
		"data":   groups,
		"echo":   echo,
	})
}
