package main

import (
	"BotMatrix/common/log"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lonelyevil/kook"
	"go.uber.org/zap"
)

type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	session     *kook.Session
	nexusConn   *websocket.Conn
	connMutex   sync.Mutex
	selfID      string

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

// --- Simple Logger Implementation ---

type ConsoleLogger struct{}

func (l *ConsoleLogger) Trace() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Debug() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Info() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Warn() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Error() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Fatal() kook.Entry { return &ConsoleEntry{} }

type ConsoleEntry struct{}

func (e *ConsoleEntry) Bool(key string, b bool) kook.Entry         { return e }
func (e *ConsoleEntry) Bytes(key string, val []byte) kook.Entry    { return e }
func (e *ConsoleEntry) Caller(depth int) kook.Entry                { return e }
func (e *ConsoleEntry) Dur(key string, d time.Duration) kook.Entry { return e }
func (e *ConsoleEntry) Err(key string, err error) kook.Entry       { return e }
func (e *ConsoleEntry) Float64(key string, f float64) kook.Entry   { return e }
func (e *ConsoleEntry) IPAddr(key string, ip net.IP) kook.Entry    { return e }
func (e *ConsoleEntry) Int(key string, i int) kook.Entry           { return e }
func (e *ConsoleEntry) Int64(key string, i int64) kook.Entry       { return e }
func (e *ConsoleEntry) Interface(key string, i any) kook.Entry     { return e }
func (e *ConsoleEntry) Msg(msg string)                             { log.Info(msg) }
func (e *ConsoleEntry) Msgf(f string, i ...any)                    { log.Info(fmt.Sprintf(f, i...)) }
func (e *ConsoleEntry) Str(key string, s string) kook.Entry        { return e }
func (e *ConsoleEntry) Strs(key string, s []string) kook.Entry     { return e }
func (e *ConsoleEntry) Time(key string, t time.Time) kook.Entry    { return e }

// ------------------------------------

func main() {
	// 初始化日志系统
	log.InitDefaultLogger()
	loadConfig()

	go startHTTPServer()

	// Initial start
	restartBot()

	// Wait for signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	stopBot()
}

func restartBot() {
	stopBot()

	configMutex.RLock()
	token := config.BotToken
	nexusAddr := config.NexusAddr
	configMutex.RUnlock()

	if token == "" {
		log.Warn("KOOK_BOT_TOKEN is not set, bot will not start")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())
	nexusCtx, nexusCancel = context.WithCancel(context.Background())

	session = kook.New(token, &ConsoleLogger{})

	// Register Handlers
	session.AddHandler(textMessageHandler)
	session.AddHandler(imageMessageHandler)
	session.AddHandler(kmarkdownMessageHandler)

	// Open the session
	err = session.Open()
	if err != nil {
		log.Error("Error opening connection", zap.Error(err))
		return
	}

	// Get Self Info
	me, err := session.Me()
	if err != nil {
		log.Error("Error getting self info", zap.Error(err))
		return
	}
	selfID = me.ID
	log.Info("Bot logged in", zap.String("username", me.Username), zap.String("id", selfID))

	// Connect to BotNexus
	go connectToNexus(nexusCtx, nexusAddr)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
	}
	if nexusCancel != nil {
		nexusCancel()
	}

	if session != nil {
		session.Close()
		session = nil
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

	if envToken := os.Getenv("KOOK_BOT_TOKEN"); envToken != "" {
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
		config.NexusAddr = "ws://bot-nexus:3001"
	}
	if config.LogPort == 0 {
		config.LogPort = 8085
	}
}

func handleCommon(common *kook.EventDataGeneral, author kook.User) {
	if author.Bot && author.ID == selfID {
		return
	}

	log.Printf("[%s] %s", author.Username, common.Content)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  common.MsgID,
		"user_id":     common.AuthorID,
		"message":     common.Content,
		"raw_message": common.Content,
		"sender": map[string]any{
			"user_id":  common.AuthorID,
			"nickname": author.Username,
		},
	}

	if common.Type == kook.MessageTypeImage {
		obMsg["message"] = fmt.Sprintf("[CQ:image,file=%s]", common.Content)
		obMsg["raw_message"] = obMsg["message"]
	}

	if common.ChannelType == "GROUP" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = common.TargetID
	} else {
		obMsg["message_type"] = "private"
	}

	sendToNexus(obMsg)
}

func textMessageHandler(ctx *kook.TextMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
}

func imageMessageHandler(ctx *kook.ImageMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
}

func kmarkdownMessageHandler(ctx *kook.KmarkdownMessageContext) {
	handleCommon(ctx.Common, ctx.Extra.Author)
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
			header.Add("X-Platform", "Kook")

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

			sendToNexus(map[string]any{
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

func sendToNexus(msg any) {
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
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
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
			sendKookMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			sendKookDirectMessage(userID, text, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteKookMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  selfID,
				"nickname": "KookBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func deleteKookMessage(msgID, echo string) {
	err := session.MessageDelete(msgID)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s", msgID)
	sendToNexus(map[string]any{"status": "ok", "echo": echo})
}

func sendKookMessage(targetID, content, echo string) {
	resp, err := session.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			TargetID: targetID,
			Content:  content,
			Type:     kook.MessageTypeText,
		},
	})

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s: %s", targetID, content)
	sendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": resp.MsgID},
		"echo":   echo,
	})
}

func sendKookDirectMessage(targetID, content, echo string) {
	resp, err := session.DirectMessageCreate(&kook.DirectMessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			TargetID: targetID,
			Content:  content,
			Type:     kook.MessageTypeText,
		},
	})

	if err != nil {
		log.Printf("Failed to send private message: %v", err)
		sendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent private message to %s", targetID)
	sendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": resp.MsgID},
		"echo":   echo,
	})
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
    <title>KookBot 配置中心</title>
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
        <div class="sidebar-header">KookBot</div>
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
                    机器人基础配置
                    <span class="status-badge">运行中</span>
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group" style="grid-column: span 2;">
                        <label>Bot Token</label>
                        <input type="password" id="bot_token" placeholder="请输入 Kook Bot Token">
                    </div>
                    <div class="form-group">
                        <label>Nexus 服务地址</label>
                        <input type="text" id="nexus_addr" placeholder="ws://localhost:3001">
                    </div>
                    <div class="form-group">
                        <label>Web UI/日志端口</label>
                        <input type="number" id="log_port" placeholder="8085">
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
                document.getElementById('nexus_addr').value = config.nexus_addr || '';
                document.getElementById('log_port').value = config.log_port || '';
            } catch (err) {
                console.error('Failed to load config:', err);
            }
        }

        async function saveConfig() {
            const config = {
                bot_token: document.getElementById('bot_token').value,
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
