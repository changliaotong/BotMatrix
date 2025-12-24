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
	SelfID    string `json:"self_id"`  // Optional: manually set SelfID
	LogPort   int    `json:"log_port"` // Port for HTTP Log Viewer
}

var (
	config         Config
	nexusConn      *websocket.Conn
	nexusMu        sync.Mutex // Added mutex
	httpClient     = &http.Client{Timeout: 10 * time.Second}
	streamClient   *client.StreamClient
	logManager     *LogManager
	botCtx         context.Context
	botCancel      context.CancelFunc
	botMu          sync.Mutex
	nexusCtx       context.Context
	nexusCancel    context.CancelFunc
)

type LogManager struct {
	mu     sync.RWMutex
	buffer []string
	size   int
}

func NewLogManager(size int) *LogManager {
	return &LogManager{
		buffer: make([]string, 0, size),
		size:   size,
	}
}

func (l *LogManager) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := string(p)
	// Simple rotation
	if len(l.buffer) >= l.size {
		l.buffer = l.buffer[1:]
	}
	l.buffer = append(l.buffer, strings.TrimRight(msg, "\n"))

	// Stream to Nexus
	go func(m string) {
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

func (l *LogManager) GetLogs(lines int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if lines > len(l.buffer) {
		lines = len(l.buffer)
	}
	return l.buffer[len(l.buffer)-lines:]
}

// stopNexus stops only the WebSocket connection to BotNexus
func stopNexus() {
	nexusMu.Lock()
	if nexusCancel != nil {
		log.Println("Stopping Nexus connection...")
		nexusCancel()
		nexusCancel = nil
	}
	if nexusConn != nil {
		nexusConn.Close()
		nexusConn = nil
	}
	nexusMu.Unlock()
}

// startNexus starts WebSocket connection to BotNexus
func startNexus() {
	nexusMu.Lock()
	defer nexusMu.Unlock()

	if nexusCancel != nil {
		nexusCancel()
	}

	nexusCtx, nexusCancel = context.WithCancel(botCtx)
	log.Println("Starting Nexus connection...")
	go connectNexus(nexusCtx)
}

// stopBot stops all bot-related goroutines and connections
func stopBot() {
	botMu.Lock()
	defer botMu.Unlock()

	stopNexus()

	if botCancel != nil {
		log.Println("Stopping bot services...")
		botCancel()
		botCancel = nil
	}
}

// startBot starts all bot-related goroutines
func startBot() {
	botMu.Lock()
	defer botMu.Unlock()

	// Ensure previous bot is stopped
	if botCancel != nil {
		botCancel()
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	log.Println("Starting bot services...")

	// Connect to BotNexus
	startNexus()

	// Start Stream Client if configured
	if config.ClientID != "" && config.ClientSecret != "" {
		go startStreamClient(botCtx)
	} else {
		log.Println("Stream Mode not configured (missing client_id/client_secret). Running in Webhook Send-Only mode.")
	}
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
	logManager = NewLogManager(2000)
	log.SetOutput(logManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	// Start HTTP Server for Config UI
	if config.LogPort > 0 {
		go startHTTPServer()
	}

	startBot()

	// Keep alive
	select {}
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
			fmt.Fprintln(w, line)
		}
	})

	mux.HandleFunc("/config", handleConfig)
	mux.HandleFunc("/config-ui", handleConfigUI)

	addr := fmt.Sprintf(":%d", config.LogPort)
	log.Printf("Starting HTTP Server at http://localhost%s (UI: /config-ui, Logs: /logs)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Failed to start HTTP Server: %v", err)
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(config)
		return
	}
	if r.Method == "POST" {
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botMu.Lock()
		config = newConfig
		botMu.Unlock()

		// Save to file
		file, _ := os.Create("config.json")
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(newConfig)

		// Restart bot to apply changes
		go func() {
			stopBot()
			startBot()
		}()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Configuration updated and bot restarted"})
		return
	}
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DingTalkBot 配置中心</title>
    <style>
        :root { --primary-color: #007bff; --success-color: #28a745; --danger-color: #dc3545; --bg-color: #f4f7f6; }
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
        <div class="sidebar-header">DingTalkBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">Webhook 模式 (自定义机器人)</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>Access Token</label>
                        <input type="text" id="access_token">
                    </div>
                    <div class="form-group">
                        <label>Secret (可选)</label>
                        <input type="text" id="secret">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">Stream 模式 (企业机器人)</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>AppKey (ClientID)</label>
                        <input type="text" id="client_id">
                    </div>
                    <div class="form-group">
                        <label>AppSecret (ClientSecret)</label>
                        <input type="password" id="client_secret">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">连接配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>BotNexus 地址</label>
                        <input type="text" id="nexus_addr">
                    </div>
                    <div class="form-group">
                        <label>机器人 SelfID</label>
                        <input type="text" id="self_id">
                    </div>
                    <div class="form-group">
                        <label>Web UI 端口</label>
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
            
            document.getElementById('access_token').value = cfg.access_token || '';
            document.getElementById('secret').value = cfg.secret || '';
            document.getElementById('client_id').value = cfg.client_id || '';
            document.getElementById('client_secret').value = cfg.client_secret || '';
            document.getElementById('nexus_addr').value = cfg.nexus_addr || '';
            document.getElementById('self_id').value = cfg.self_id || '';
            document.getElementById('log_port').value = cfg.log_port || 0;
        }

        async function saveConfig() {
            const cfg = {
                access_token: document.getElementById('access_token').value,
                secret: document.getElementById('secret').value,
                client_id: document.getElementById('client_id').value,
                client_secret: document.getElementById('client_secret').value,
                nexus_addr: document.getElementById('nexus_addr').value,
                self_id: document.getElementById('self_id').value,
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

// --- Stream SDK Integration ---

func startStreamClient(ctx context.Context) {
	// logger.SetLogger(logger.NewStdLogger(os.Stdout)) // Use default logger or implement ILogger if needed

	cli := client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(config.ClientID, config.ClientSecret)),
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
	log.Println("Stream Client started successfully! Listening for events...")

	// Block until close or context done
	<-ctx.Done()
	log.Println("Stream Client stopping...")
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

func connectNexus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Nexus connection loop stopped by context.")
			return
		default:
			log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
			header := http.Header{}
			header.Add("X-Self-ID", config.SelfID)
			header.Add("X-Client-Role", "Universal") // Generic OneBot Client

			conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, header)
			if err != nil {
				log.Printf("Connection failed: %v. Retrying in 5s...", err)
				select {
				case <-time.After(5 * time.Second):
					continue
				case <-ctx.Done():
					return
				}
			}

			nexusMu.Lock()
			nexusConn = conn
			nexusMu.Unlock()
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
			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Println("Read error:", err)
						return
					}
					handleNexusMessage(message)
				}
			}()

			select {
			case <-ctx.Done():
				conn.Close()
				<-done
				return
			case <-done:
				conn.Close()
				log.Println("Disconnected from BotNexus. Reconnecting...")
				select {
				case <-time.After(3 * time.Second):
				case <-ctx.Done():
					return
				}
			}
		}
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
