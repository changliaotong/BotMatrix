package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/websocket"
)

// Config holds the bot configuration
type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	bot         *tgbotapi.BotAPI
	nexusConn   *websocket.Conn
	connMutex   sync.Mutex
	selfID      string

	botCtx    context.Context
	botCancel context.CancelFunc

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

func main() {
	log.SetOutput(logManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	go startHTTPServer()

	// Initial start
	restartBot()

	// Wait for signal
	select {}
}

func restartBot() {
	stopBot()

	configMutex.RLock()
	botToken := config.BotToken
	nexusAddr := config.NexusAddr
	configMutex.RUnlock()

	if botToken == "" {
		log.Println("Telegram bot token is not set, bot will not start")
		return
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("Failed to create Telegram Bot: %v", err)
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	bot.Debug = true
	selfID = fmt.Sprintf("%d", bot.Self.ID)
	log.Printf("Authorized on account %s (ID: %s)", bot.Self.UserName, selfID)

	// Connect to Nexus
	go connectToNexus(botCtx, nexusAddr)

	// Start Polling
	go func(ctx context.Context) {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := bot.GetUpdatesChan(u)

		for {
			select {
			case <-ctx.Done():
				bot.StopReceivingUpdates()
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				if update.Message == nil {
					continue
				}
				handleMessage(update.Message)
			}
		}
	}(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}

	connMutex.Lock()
	if nexusConn != nil {
		nexusConn.Close()
		nexusConn = nil
	}
	connMutex.Unlock()
}

func loadConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.ReadFile("config.json")
	if err == nil {
		if err := json.Unmarshal(file, &config); err != nil {
			log.Printf("Error parsing config.json: %v", err)
		}
	}

	if envToken := os.Getenv("TELEGRAM_BOT_TOKEN"); envToken != "" {
		config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}
	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		if p, err := strconv.Atoi(envLogPort); err == nil {
			config.LogPort = p
		}
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
	if config.LogPort == 0 {
		config.LogPort = 8087
	}
}

func handleMessage(msg *tgbotapi.Message) {
	// Handle Multimedia
	if msg.Photo != nil && len(msg.Photo) > 0 {
		// Get the largest photo
		photo := msg.Photo[len(msg.Photo)-1]
		fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
		file, err := bot.GetFile(fileConfig)
		if err == nil {
			// Construct direct URL: https://api.telegram.org/file/bot<token>/<file_path>
			// tgbotapi helper:
			url := file.Link(config.BotToken)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	} else if msg.Sticker != nil {
		// Handle Sticker as image
		fileConfig := tgbotapi.FileConfig{FileID: msg.Sticker.FileID}
		file, err := bot.GetFile(fileConfig)
		if err == nil {
			url := file.Link(config.BotToken)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	}

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	// OneBot Message
	obMsg := map[string]interface{}{
		"post_type":    "message",
		"message_type": "group", // Telegram doesn't strictly distinguish, but group chats exist
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "normal",
		"message_id":   fmt.Sprintf("%d", msg.MessageID),
		"user_id":      fmt.Sprintf("%d", msg.From.ID),
		"message":      msg.Text,
		"raw_message":  msg.Text,
		"sender": map[string]interface{}{
			"user_id":  fmt.Sprintf("%d", msg.From.ID),
			"nickname": msg.From.UserName,
		},
	}

	if msg.Chat.IsPrivate() {
		obMsg["message_type"] = "private"
	} else {
		obMsg["group_id"] = fmt.Sprintf("%d", msg.Chat.ID)
	}

	sendToNexus(obMsg)
}

func connectToNexus(ctx context.Context, addr string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("Connecting to BotNexus at %s...", addr)
			header := http.Header{}
			header.Add("X-Self-ID", selfID)
			header.Add("X-Platform", "Telegram")

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

			// Lifecycle Event
			sendToNexus(map[string]interface{}{
				"post_type":       "meta_event",
				"meta_event_type": "lifecycle",
				"sub_type":        "connect",
				"self_id":         selfID,
				"time":            time.Now().Unix(),
			})

			// Handle incoming commands
			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("BotNexus disconnected: %v", err)
						return
					}
					handleNexusCommand(message)
				}
			}()

			select {
			case <-ctx.Done():
				connMutex.Lock()
				if nexusConn != nil {
					nexusConn.Close()
					nexusConn = nil
				}
				connMutex.Unlock()
				return
			case <-done:
				connMutex.Lock()
				nexusConn = nil
				connMutex.Unlock()
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func sendToNexus(msg interface{}) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send to Nexus: %v", err)
	}
}

func startHTTPServer() {
	configMutex.RLock()
	port := config.LogPort
	configMutex.RUnlock()

	http.HandleFunc("/", handleConfigUI)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		lines := 100
		if l := r.URL.Query().Get("lines"); l != "" {
			if i, err := strconv.Atoi(l); err == nil {
				lines = i
			}
		}
		logs := logManager.GetLogs(lines)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})

	log.Printf("Admin UI started on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("HTTP server failed: %v", err)
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

		configMutex.Lock()
		config.BotToken = newConfig.BotToken
		config.NexusAddr = newConfig.NexusAddr
		// LogPort update requires restart of HTTP server, usually we don't do it on the fly
		configMutex.Unlock()

		// Save to file
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile("config.json", data, 0644)

		// Restart Bot
		go restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarting"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	cfg := config
	configMutex.RUnlock()

	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>TelegramBot 控制面板</title>
    <style>
        :root {
            --primary-color: #0088cc;
            --bg-color: #f4f7f9;
            --sidebar-width: 240px;
        }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; display: flex; height: 100vh; background: var(--bg-color); }
        .sidebar { width: var(--sidebar-width); background: #2c3e50; color: white; display: flex; flex-direction: column; }
        .sidebar-header { padding: 20px; font-size: 1.2em; font-weight: bold; border-bottom: 1px solid #34495e; text-align: center; }
        .nav-item { padding: 15px 20px; cursor: pointer; transition: background 0.3s; display: flex; align-items: center; }
        .nav-item:hover { background: #34495e; }
        .nav-item.active { background: var(--primary-color); }
        .nav-item span { margin-right: 10px; }
        
        .main-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
        .header { background: white; padding: 15px 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); display: flex; justify-content: space-between; align-items: center; }
        .status-badge { padding: 5px 12px; border-radius: 15px; font-size: 0.9em; background: #e8f5e9; color: #2e7d32; }
        
        .content-area { padding: 30px; overflow-y: auto; flex: 1; }
        .card { background: white; padding: 25px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); max-width: 800px; margin: 0 auto; }
        
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 8px; font-weight: 600; color: #333; }
        input { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; font-size: 14px; }
        input:focus { outline: none; border-color: var(--primary-color); box-shadow: 0 0 0 2px rgba(0,136,204,0.1); }
        
        .btn { background: var(--primary-color); color: white; border: none; padding: 12px 24px; border-radius: 4px; cursor: pointer; font-weight: 600; transition: background 0.3s; }
        .btn:hover { background: #0077b3; }
        
        #log-container { background: #1e1e1e; color: #d4d4d4; padding: 15px; border-radius: 4px; font-family: 'Consolas', monospace; height: 600px; overflow-y: auto; font-size: 13px; line-height: 1.5; }
        .log-line { margin-bottom: 4px; border-bottom: 1px solid #333; padding-bottom: 2px; white-space: pre-wrap; word-break: break-all; }
        
        .tab-content { display: none; }
        .tab-content.active { display: block; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">TelegramBot</div>
        <div class="nav-item active" onclick="switchTab('config')">
            <span>⚙️</span> 核心配置
        </div>
        <div class="nav-item" onclick="switchTab('logs')">
            <span>📝</span> 实时日志
        </div>
    </div>
    
    <div class="main-content">
        <div class="header">
            <h2 id="page-title">核心配置</h2>
            <div class="status-badge">运行中</div>
        </div>
        
        <div class="content-area">
            <div id="config-tab" class="tab-content active">
                <div class="card">
                    <div class="form-group">
                        <label>Bot Token</label>
                        <input type="password" id="botToken" value="` + cfg.BotToken + `" placeholder="Enter Telegram Bot Token">
                    </div>
                    <div class="form-group">
                        <label>Nexus 地址</label>
                        <input type="text" id="nexusAddr" value="` + cfg.NexusAddr + `" placeholder="ws://localhost:3005">
                    </div>
                    <div class="form-group">
                        <label>管理端口 (需重启生效)</label>
                        <input type="number" id="logPort" value="` + fmt.Sprint(cfg.LogPort) + `">
                    </div>
                    <button class="btn" onclick="saveConfig()">保存并重启 Bot</button>
                </div>
            </div>
            
            <div id="logs-tab" class="tab-content">
                <div id="log-container">正在加载日志...</div>
            </div>
        </div>
    </div>

    <script>
        function switchTab(tab) {
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
            
            if(tab === 'config') {
                document.getElementById('config-tab').classList.add('active');
                document.querySelector('.nav-item:nth-child(2)').classList.add('active');
                document.getElementById('page-title').innerText = '核心配置';
            } else {
                document.getElementById('logs-tab').classList.add('active');
                document.querySelector('.nav-item:nth-child(3)').classList.add('active');
                document.getElementById('page-title').innerText = '实时日志';
                loadLogs();
            }
        }

        async function saveConfig() {
            const config = {
                bot_token: document.getElementById('botToken').value,
                nexus_addr: document.getElementById('nexusAddr').value,
                log_port: parseInt(document.getElementById('logPort').value)
            };
            
            try {
                const resp = await fetch('/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config)
                });
                if(resp.ok) {
                    alert('配置已保存，Bot 正在重启...');
                } else {
                    alert('保存失败: ' + await resp.text());
                }
            } catch(e) {
                alert('请求失败: ' + e);
            }
        }

        async function loadLogs() {
            if(!document.getElementById('logs-tab').classList.contains('active')) return;
            
            try {
                const resp = await fetch('/logs?lines=200');
                const logs = await resp.json();
                const container = document.getElementById('log-container');
                container.innerHTML = logs.map(line => ` + "`" + `<div class="log-line">${line}</div>` + "`" + `).join('');
                container.scrollTop = container.scrollHeight;
            } catch(e) {}
            
            setTimeout(loadLogs, 2000);
        }
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
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
		chatIDStr, _ := cmd.Params["group_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if chatIDStr != "" && text != "" {
			sendTelegramMessage(chatIDStr, text, cmd.Echo)
		}
	case "send_private_msg":
		chatIDStr, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if chatIDStr != "" && text != "" {
			sendTelegramMessage(chatIDStr, text, cmd.Echo)
		}
	case "delete_msg":
		msgIDStr, _ := cmd.Params["message_id"].(string)
		// Telegram requires ChatID to delete a message.
		// Standard OneBot delete_msg only provides message_id.
		// However, we implemented sendTelegramMessage to return message_id as just ID.
		// To support deletion, we need to know the ChatID.
		// We can't know ChatID from just MessageID in Telegram API (unlike Discord/Slack where we might need it too).
		// Wait, Telegram `deleteMessage` takes `chat_id` and `message_id`.
		// If we don't store the mapping, we can't delete it.
		// BUT, maybe we can change the returned message_id to be "chat_id:message_id"?
		// Yes, let's do that in sendTelegramMessage first.

		if msgIDStr != "" {
			deleteTelegramMessage(msgIDStr, cmd.Echo)
		}

	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": bot.Self.UserName,
			},
			"echo": cmd.Echo,
		})
	}
}

func sendTelegramMessage(chatIDStr, text, echo string) {
	// Parse chatID (int64)
	var chatID int64
	fmt.Sscanf(chatIDStr, "%d", &chatID)

	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %d: %s", chatID, text)
	// Return composite ID: "chat_id:message_id"
	compositeID := fmt.Sprintf("%d:%d", chatID, sentMsg.MessageID)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteTelegramMessage(compositeID, echo string) {
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		sendToNexus(map[string]interface{}{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}

	var chatID int64
	var messageID int
	fmt.Sscanf(parts[0], "%d", &chatID)
	fmt.Sscanf(parts[1], "%d", &messageID)

	delCfg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := bot.Request(delCfg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %d in chat %d", messageID, chatID)
	sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
}
