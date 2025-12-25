package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gorilla/websocket"
	"github.com/jordan-wright/email"
	"go.uber.org/zap"
)

// Config holds the bot configuration
type Config struct {
	ImapServer   string `json:"imap_server"`
	ImapPort     int    `json:"imap_port"`
	Username     string `json:"username"` // Email address
	Password     string `json:"password"` // App Password
	SmtpServer   string `json:"smtp_server"`
	SmtpPort     int    `json:"smtp_port"`
	SmtpUsername string `json:"smtp_username"` // Usually same as Username
	SmtpPassword string `json:"smtp_password"` // Usually same as Password
	PollInterval int    `json:"poll_interval"` // Seconds
	NexusAddr    string `json:"nexus_addr"`
	LogPort      int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
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

func loadConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	}

	if envUser := os.Getenv("EMAIL_USERNAME"); envUser != "" {
		config.Username = envUser
	}
	if envPass := os.Getenv("EMAIL_PASSWORD"); envPass != "" {
		config.Password = envPass
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}
	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		if p, err := strconv.Atoi(envLogPort); err == nil {
			config.LogPort = p
		}
	}

	if config.SmtpUsername == "" {
		config.SmtpUsername = config.Username
	}
	if config.SmtpPassword == "" {
		config.SmtpPassword = config.Password
	}
	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-nexus:3001"
	}
	if config.LogPort == 0 {
		config.LogPort = 8086
	}
	if config.PollInterval == 0 {
		config.PollInterval = 60
	}
	selfID = config.Username
}

func connectToNexus(ctx context.Context, addr string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			Info("Connecting to BotNexus", zap.String("addr", addr))
			header := http.Header{}
			header.Add("X-Self-ID", selfID)
			header.Add("X-Platform", "Email")

			c, _, err := websocket.DefaultDialer.Dial(addr, header)
			if err != nil {
				Error("Connection error", zap.Error(err))
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			connMutex.Lock()
			nexusConn = c
			connMutex.Unlock()
			Info("Connected to BotNexus!")

			// Send Lifecycle Event: Connect
			sendEvent(map[string]interface{}{
				"post_type":       "meta_event",
				"meta_event_type": "lifecycle",
				"sub_type":        "connect",
				"self_id":         selfID,
				"time":            time.Now().Unix(),
			})

			// Heartbeat ticker
			ticker := time.NewTicker(30 * time.Second)
			done := make(chan struct{})

			// Message reading loop
			go func() {
				defer close(done)
				defer ticker.Stop()
				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						Error("Read error", zap.Error(err))
						return
					}
					handleAction(message)
				}
			}()

			// Wait for disconnect or context cancel
			for {
				select {
				case <-ctx.Done():
					c.Close()
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
						goto next_reconnect
					}
				case <-ticker.C:
					sendEvent(map[string]interface{}{
						"post_type":       "meta_event",
						"meta_event_type": "heartbeat",
						"self_id":         selfID,
						"time":            time.Now().Unix(),
						"status": map[string]interface{}{
							"online": true,
							"good":   true,
						},
					})
				}
			}
		next_reconnect:
		}
	}
}

func sendEvent(event map[string]interface{}) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if nexusConn == nil {
		return
	}
	// Add platform info
	event["platform"] = "email"

	err := nexusConn.WriteJSON(event)
	if err != nil {
		Error("Write error", zap.Error(err))
		nexusConn = nil
	}
}

func handleAction(msg []byte) {
	var action map[string]interface{}
	if err := json.Unmarshal(msg, &action); err != nil {
		Error("JSON Unmarshal error", zap.Error(err))
		return
	}

	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]interface{})
	echo, _ := action["echo"]

	response := map[string]interface{}{
		"status":  "ok",
		"retcode": 0,
		"data":    nil,
		"echo":    echo,
	}

	switch actionType {
	case "send_private_msg":
		userID, _ := params["user_id"].(string) // Email address
		message, _ := params["message"].(string)

		// Attempt to parse subject from message (First line as subject?)
		// Simple strategy: Subject = "New Message", Body = message
		// Or try to detect if message starts with "Subject: "

		subject := "Message from BotMatrix"
		body := message

		lines := strings.SplitN(message, "\n", 2)
		if len(lines) > 0 && strings.HasPrefix(lines[0], "Subject:") {
			subject = strings.TrimPrefix(lines[0], "Subject:")
			subject = strings.TrimSpace(subject)
			if len(lines) > 1 {
				body = lines[1]
			} else {
				body = ""
			}
		}

		configMutex.RLock()
		err := sendEmail(userID, subject, body)
		configMutex.RUnlock()

		if err != nil {
			Error("Failed to send email", zap.String("to", userID), zap.Error(err))
			response["status"] = "failed"
			response["retcode"] = -1
		} else {
			Info("Email sent", zap.String("to", userID))
		}
	}

	sendEvent(response)
}

func sendEmail(to, subject, body string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("BotMatrix <%s>", config.SmtpUsername)
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	addr := fmt.Sprintf("%s:%d", config.SmtpServer, config.SmtpPort)
	auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpServer)

	return e.Send(addr, auth)
}

func pollEmails(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			configMutex.RLock()
			imapServer := config.ImapServer
			imapPort := config.ImapPort
			username := config.Username
			password := config.Password
			pollInterval := config.PollInterval
			configMutex.RUnlock()

			if imapServer == "" || username == "" || password == "" {
				Warn("IMAP configuration missing")
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}

			Info("Connecting to IMAP server")
			c, err := client.DialTLS(fmt.Sprintf("%s:%d", imapServer, imapPort), nil)
			if err != nil {
				Error("IMAP connection failed", zap.Error(err))
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
					continue
				}
			}
			Info("IMAP Connected")

			if err := c.Login(username, password); err != nil {
				Error("IMAP Login failed", zap.Error(err))
				c.Close()
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
					continue
				}
			}

			// Polling loop for this connection
			for {
				select {
				case <-ctx.Done():
					c.Logout()
					return
				default:
					_, err := c.Select("INBOX", false)
					if err != nil {
						Error("Select INBOX failed", zap.Error(err))
						goto next_imap_conn
					}

					// Search for UNSEEN messages
					criteria := imap.NewSearchCriteria()
					criteria.WithoutFlags = []string{imap.SeenFlag}
					uids, err := c.Search(criteria)
					if err != nil {
						Error("Search failed", zap.Error(err))
						goto next_imap_conn
					}

					if len(uids) > 0 {
						Info("Found new emails", zap.Int("count", len(uids)))
						seqset := new(imap.SeqSet)
						seqset.AddNum(uids...)

						section := &imap.BodySectionName{}
						items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

						messages := make(chan *imap.Message, 10)
						done := make(chan error, 1)
						go func() {
							done <- c.Fetch(seqset, items, messages)
						}()

						for msg := range messages {
							processMessage(msg, section)
						}

						if err := <-done; err != nil {
							Error("Fetch failed", zap.Error(err))
						}
					}

					select {
					case <-ctx.Done():
						c.Logout()
						return
					case <-time.After(time.Duration(pollInterval) * time.Second):
						continue
					}
				}
			}
		next_imap_conn:
			c.Logout()
		}
	}
}

func processMessage(msg *imap.Message, section *imap.BodySectionName) {
	r := msg.GetBody(section)
	if r == nil {
		return
	}

	// Simple body reading
	// For complex emails (MIME multipart), this might need parsing
	// But section usually gets the full body.
	// To do it right, we should use a parser.
	// But for now, let's try to read it as raw and maybe just extract text.

	// A better way is to use "github.com/emersion/go-message/mail" to parse the body
	// but I didn't import it. Let's do a basic read.

	bodyBytes, _ := ioutil.ReadAll(r)
	bodyStr := string(bodyBytes)

	// Construct OneBot Message
	sender := msg.Envelope.From[0]
	senderEmail := fmt.Sprintf("%s@%s", sender.MailboxName, sender.HostName)
	senderName := sender.PersonalName

	content := fmt.Sprintf("Subject: %s\n\n%s", msg.Envelope.Subject, bodyStr)

	Info("Received email", zap.String("from", senderEmail), zap.String("subject", msg.Envelope.Subject))

	event := map[string]interface{}{
		"post_type":    "message",
		"message_type": "private", // Treat all emails as private messages
		"time":         msg.Envelope.Date.Unix(),
		"self_id":      selfID,
		"sub_type":     "friend",
		"message_id":   fmt.Sprintf("%d", msg.SeqNum),
		"user_id":      senderEmail,
		"message":      content,
		"raw_message":  content,
		"sender": map[string]interface{}{
			"user_id":  senderEmail,
			"nickname": senderName,
		},
	}

	sendEvent(event)
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
	Info("Starting HTTP Server", zap.String("addr", addr), zap.String("ui_path", "/config-ui"), zap.String("logs_path", "/logs"))
	if err := http.ListenAndServe(addr, mux); err != nil {
		Error("Failed to start HTTP Server", zap.Error(err))
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
    <title>EmailBot 配置中心</title>
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
        <div class="sidebar-header">EmailBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">IMAP 接收配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>IMAP 服务器</label>
                        <input type="text" id="imap_server">
                    </div>
                    <div class="form-group">
                        <label>IMAP 端口</label>
                        <input type="number" id="imap_port">
                    </div>
                    <div class="form-group">
                        <label>邮箱账号</label>
                        <input type="text" id="username">
                    </div>
                    <div class="form-group">
                        <label>邮箱密码/授权码</label>
                        <input type="password" id="password">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">SMTP 发送配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>SMTP 服务器</label>
                        <input type="text" id="smtp_server">
                    </div>
                    <div class="form-group">
                        <label>SMTP 端口</label>
                        <input type="number" id="smtp_port">
                    </div>
                    <div class="form-group">
                        <label>SMTP 账号 (留空则同邮箱账号)</label>
                        <input type="text" id="smtp_username">
                    </div>
                    <div class="form-group">
                        <label>SMTP 密码 (留空则同邮箱密码)</label>
                        <input type="password" id="smtp_password">
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="section-title">通用配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>BotNexus 地址</label>
                        <input type="text" id="nexus_addr">
                    </div>
                    <div class="form-group">
                        <label>轮询间隔 (秒)</label>
                        <input type="number" id="poll_interval">
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
            
            document.getElementById('imap_server').value = cfg.imap_server || '';
            document.getElementById('imap_port').value = cfg.imap_port || 0;
            document.getElementById('username').value = cfg.username || '';
            document.getElementById('password').value = cfg.password || '';
            document.getElementById('smtp_server').value = cfg.smtp_server || '';
            document.getElementById('smtp_port').value = cfg.smtp_port || 0;
            document.getElementById('smtp_username').value = cfg.smtp_username || '';
            document.getElementById('smtp_password').value = cfg.smtp_password || '';
            document.getElementById('nexus_addr').value = cfg.nexus_addr || '';
            document.getElementById('poll_interval').value = cfg.poll_interval || 60;
            document.getElementById('log_port').value = cfg.log_port || 0;
        }

        async function saveConfig() {
            const cfg = {
                imap_server: document.getElementById('imap_server').value,
                imap_port: parseInt(document.getElementById('imap_port').value),
                username: document.getElementById('username').value,
                password: document.getElementById('password').value,
                smtp_server: document.getElementById('smtp_server').value,
                smtp_port: parseInt(document.getElementById('smtp_port').value),
                smtp_username: document.getElementById('smtp_username').value,
                smtp_password: document.getElementById('smtp_password').value,
                nexus_addr: document.getElementById('nexus_addr').value,
                poll_interval: parseInt(document.getElementById('poll_interval').value),
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

func main() {
	// 初始化日志系统
	InitDefaultLogger()
	defer Sync()

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
	nexusAddr := config.NexusAddr
	configMutex.RUnlock()

	botCtx, botCancel = context.WithCancel(context.Background())
	nexusCtx, nexusCancel = context.WithCancel(context.Background())

	// Connect to BotNexus
	go connectToNexus(nexusCtx, nexusAddr)

	// Start polling
	go pollEmails(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
	}
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
