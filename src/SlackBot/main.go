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

	"github.com/gorilla/websocket"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Config struct {
	BotToken  string `json:"bot_token"` // xoxb-...
	AppToken  string `json:"app_token"` // xapp-...
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	api         *slack.Client
	client      *socketmode.Client
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
	appToken := config.AppToken
	nexusAddr := config.NexusAddr
	configMutex.RUnlock()

	if botToken == "" || appToken == "" {
		log.Println("Slack tokens are not set, bot will not start")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	// Initialize Slack Client
	api = slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	client = socketmode.New(api)

	// Get Bot Info
	authTest, err := api.AuthTest()
	if err != nil {
		log.Printf("Slack Auth failed: %v", err)
		return
	}
	selfID = authTest.BotID
	log.Printf("Slack Bot Authorized: %s (ID: %s, User: %s)", authTest.User, selfID, authTest.UserID)

	// Connect to BotNexus
	go connectToNexus(botCtx, nexusAddr)

	// Start Socket Mode
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-client.Events:
				switch evt.Type {
				case socketmode.EventTypeConnecting:
					log.Println("Connecting to Slack Socket Mode...")
				case socketmode.EventTypeConnectionError:
					log.Println("Connection failed. Retrying later...")
				case socketmode.EventTypeConnected:
					log.Println("Connected to Slack Socket Mode!")
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
					if !ok {
						continue
					}
					client.Ack(*evt.Request)

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innerEvent := eventsAPIEvent.InnerEvent
						switch ev := innerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							handleMessage(ev)
						}
					}
				}
			}
		}
	}(botCtx)

	go func() {
		if err := client.Run(); err != nil {
			log.Printf("Socket Mode failed: %v", err)
		}
	}()
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
		json.Unmarshal(file, &config)
	}

	if envBot := os.Getenv("SLACK_BOT_TOKEN"); envBot != "" {
		config.BotToken = envBot
	}
	if envApp := os.Getenv("SLACK_APP_TOKEN"); envApp != "" {
		config.AppToken = envApp
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
		config.LogPort = 8086
	}
}

func handleMessage(ev *slackevents.MessageEvent) {
	// Ignore bot messages
	if ev.BotID != "" && ev.BotID == selfID {
		return
	}
	// Also ignore if subtype is not empty (like message_changed, etc for now) unless we want to handle edits
	if ev.SubType != "" {
		// return // optional: strict handling
	}

	log.Printf("[%s] %s", ev.User, ev.Text)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  ev.ClientMsgID, // or ev.Ts
		"user_id":     ev.User,
		"message":     ev.Text,
		"raw_message": ev.Text,
		"sender": map[string]interface{}{
			"user_id":  ev.User,
			"nickname": "SlackUser", // We could fetch user info, but let's save API calls
		},
	}

	// Handle Channel vs DM
	// Slack Channel IDs start with C (Channel), D (Direct Message), G (Group)
	if strings.HasPrefix(ev.Channel, "D") {
		obMsg["message_type"] = "private"
	} else {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = ev.Channel
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
			header.Add("X-Platform", "Slack")

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

			sendToNexus(map[string]interface{}{
				"post_type":       "meta_event",
				"meta_event_type": "lifecycle",
				"sub_type":        "connect",
				"self_id":         selfID,
				"time":            time.Now().Unix(),
			})

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
				conn.Close()
				<-done
				return
			case <-done:
				connMutex.Lock()
				nexusConn = nil
				connMutex.Unlock()
				select {
				case <-ctx.Done():
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}
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
		channelID, _ := cmd.Params["group_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if channelID != "" && text != "" {
			sendSlackMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			sendSlackMessage(userID, text, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteSlackMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": "SlackBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendSlackMessage(channelID, text, echo string) {
	_, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
	)

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s", channelID)
	compositeID := fmt.Sprintf("%s:%s", channelID, timestamp)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteSlackMessage(compositeID, echo string) {
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		sendToNexus(map[string]interface{}{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}
	channelID := parts[0]
	timestamp := parts[1]

	_, _, err := api.DeleteMessage(channelID, timestamp)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s in channel %s", timestamp, channelID)
	sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
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

		configMutex.Lock()
		config = newConfig
		configMutex.Unlock()

		// Save to file
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile("config.json", data, 0644)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated successfully"))

		// Restart bot with new config
		go restartBot()
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
    <title>SlackBot 配置中心</title>
    <style>
        :root {
            --primary-color: #1a73e8;
            --success-color: #28a745;
            --danger-color: #dc3545;
            --bg-color: #f4f7f6;
            --text-color: #333;
            --sidebar-width: 280px;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background-color: var(--bg-color);
            margin: 0;
            display: flex;
            height: 100vh;
            color: var(--text-color);
        }

        .sidebar {
            width: var(--sidebar-width);
            background: #2c3e50;
            color: white;
            display: flex;
            flex-direction: column;
            box-shadow: 2px 0 5px rgba(0,0,0,0.1);
        }

        .sidebar-header {
            padding: 20px;
            font-size: 20px;
            font-weight: bold;
            text-align: center;
            border-bottom: 1px solid #34495e;
            background: #1a252f;
        }

        .nav-item {
            padding: 15px 20px;
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 10px;
            border-left: 4px solid transparent;
        }

        .nav-item:hover {
            background: #34495e;
        }

        .nav-item.active {
            background: #34495e;
            border-left-color: var(--primary-color);
            color: var(--primary-color);
        }

        .main-content {
            flex: 1;
            overflow-y: auto;
            padding: 30px;
        }

        .tab-content {
            display: none;
            max-width: 900px;
            margin: 0 auto;
        }

        .tab-content.active {
            display: block;
        }

        .card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
            padding: 25px;
            margin-bottom: 25px;
        }

        .section-title {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 20px;
            color: #2c3e50;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f2f5;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .form-group {
            margin-bottom: 20px;
        }

        .form-group label {
            display: block;
            margin-bottom: 8px;
            font-weight: 500;
            color: #555;
        }

        .form-group input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
            font-size: 14px;
            transition: border-color 0.2s;
        }

        .form-group input:focus {
            outline: none;
            border-color: var(--primary-color);
        }

        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 500;
            transition: all 0.2s;
            font-size: 14px;
        }

        .btn-primary {
            background: var(--primary-color);
            color: white;
        }

        .btn-primary:hover {
            background: #1557b0;
            box-shadow: 0 2px 5px rgba(26,115,232,0.3);
        }

        .logs-container {
            background: #1e1e1e;
            color: #d4d4d4;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            height: 600px;
            overflow-y: auto;
            font-size: 13px;
            line-height: 1.5;
            white-space: pre-wrap;
            word-break: break-all;
        }

        .status-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
            background: #e8f0fe;
            color: var(--primary-color);
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">SlackBot</div>
        <div class="nav-item active" onclick="switchTab('config')">
            <span>⚙️</span> 核心配置
        </div>
        <div class="nav-item" onclick="switchTab('logs')">
            <span>📝</span> 实时日志
        </div>
    </div>

    <div class="main-content">
        <!-- 配置页 -->
        <div id="config-tab" class="tab-content active">
            <div class="card">
                <div class="section-title">
                    Slack 机器人配置
                    <span class="status-badge">运行中</span>
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>Bot Token (xoxb-...)</label>
                        <input type="password" id="bot_token" placeholder="请输入 Slack Bot Token">
                    </div>
                    <div class="form-group">
                        <label>App Token (xapp-...)</label>
                        <input type="password" id="app_token" placeholder="请输入 Slack App Token">
                    </div>
                    <div class="form-group">
                        <label>Nexus 服务地址</label>
                        <input type="text" id="nexus_addr" placeholder="ws://localhost:3005">
                    </div>
                    <div class="form-group">
                        <label>Web UI/日志端口</label>
                        <input type="number" id="log_port" placeholder="8086">
                    </div>
                </div>
                <div style="margin-top: 10px;">
                    <button class="btn btn-primary" onclick="saveConfig()">保存并重启机器人</button>
                </div>
            </div>
        </div>

        <!-- 日志页 -->
        <div id="logs-tab" class="tab-content">
            <div class="card">
                <div class="section-title">
                    实时运行日志
                    <button class="btn" style="background: #f0f2f5;" onclick="fetchLogs()">🔄 刷新</button>
                </div>
                <div id="logs" class="logs-container">正在加载日志...</div>
            </div>
        </div>
    </div>

    <script>
        function switchTab(tabId) {
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
            
            document.getElementById(tabId + '-tab').classList.add('active');
            event.currentTarget.classList.add('active');

            if (tabId === 'logs') {
                fetchLogs();
            }
        }

        async function loadConfig() {
            try {
                const resp = await fetch('/config');
                const config = await resp.json();
                document.getElementById('bot_token').value = config.bot_token || '';
                document.getElementById('app_token').value = config.app_token || '';
                document.getElementById('nexus_addr').value = config.nexus_addr || '';
                document.getElementById('log_port').value = config.log_port || '';
            } catch (err) {
                console.error('Failed to load config:', err);
            }
        }

        async function saveConfig() {
            const config = {
                bot_token: document.getElementById('bot_token').value,
                app_token: document.getElementById('app_token').value,
                nexus_addr: document.getElementById('nexus_addr').value,
                log_port: parseInt(document.getElementById('log_port').value)
            };

            try {
                const resp = await fetch('/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config)
                });

                if (resp.ok) {
                    alert('配置已保存，机器人正在重启...');
                    const currentPort = window.location.port;
                    if (config.log_port && config.log_port.toString() !== currentPort) {
                        setTimeout(() => {
                            window.location.href = 'http://' + window.location.hostname + ':' + config.log_port + '/config-ui';
                        }, 2000);
                    }
                } else {
                    alert('保存失败: ' + await resp.text());
                }
            } catch (err) {
                alert('保存出错: ' + err);
            }
        }

        async function fetchLogs() {
            try {
                const resp = await fetch('/logs?lines=200');
                const text = await resp.text();
                const logDiv = document.getElementById('logs');
                if (logDiv) {
                    logDiv.textContent = text;
                    logDiv.scrollTop = logDiv.scrollHeight;
                }
            } catch (err) {
                console.error('Failed to fetch logs:', err);
            }
        }

        // 初始化
        loadConfig();
        setInterval(fetchLogs, 2000);
    </script>
</body>
</html>
	`)
}
