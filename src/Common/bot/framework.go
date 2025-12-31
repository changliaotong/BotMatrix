package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"BotMatrix/common/types"
	"BotMatrix/common/utils"

	"github.com/gorilla/websocket"
)

// BotConfig defines basic configuration for any bot
type BotConfig struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
	UseTLS    bool   `json:"use_tls"`   // Whether to use TLS (HTTPS/WSS)
	CertFile  string `json:"cert_file"` // Certificate file path
	KeyFile   string `json:"key_file"`  // Private key file path
}

// LogManager handles log rotation and retrieval
type LogManager struct {
	entries []types.LogEntry
	max     int
	mutex   sync.Mutex
	logChan chan types.LogEntry
}

func NewLogManager(max int) *LogManager {
	return &LogManager{
		entries: make([]types.LogEntry, 0, max),
		max:     max,
		logChan: make(chan types.LogEntry, 100),
	}
}

func (m *LogManager) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	entry := types.LogEntry{
		Level:     "INFO",
		Time:      now.Format("15:04:05"),
		Timestamp: now,
		Message:   string(p),
	}
	m.entries = append(m.entries, entry)
	if len(m.entries) > m.max {
		m.entries = m.entries[len(m.entries)-m.max:]
	}

	select {
	case m.logChan <- entry:
	default:
	}

	return os.Stderr.Write(p)
}

func (m *LogManager) Log(level, msg string) {
	m.mutex.Lock()
	now := time.Now()
	entry := types.LogEntry{
		Level:     level,
		Time:      now.Format("15:04:05"),
		Timestamp: now,
		Message:   msg,
	}
	m.entries = append(m.entries, entry)
	if len(m.entries) > m.max {
		m.entries = m.entries[len(m.entries)-m.max:]
	}
	m.mutex.Unlock()

	select {
	case m.logChan <- entry:
	default:
	}

	fmt.Fprintf(os.Stderr, "[%s] %s %s\n", entry.Timestamp.Format("15:04:05"), level, msg)
}

func (m *LogManager) GetLogs(lines int) []types.LogEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if lines > len(m.entries) {
		lines = len(m.entries)
	}
	result := make([]types.LogEntry, lines)
	copy(result, m.entries[len(m.entries)-lines:])
	return result
}

// ConfigField defines a field in the Config UI
type ConfigField struct {
	Label string `json:"label"`
	ID    string `json:"id"`
	Type  string `json:"type"` // "text", "password", "number"
	Value any    `json:"value"`
}

// ConfigSection defines a section in the Config UI
type ConfigSection struct {
	Title  string        `json:"title"`
	Fields []ConfigField `json:"fields"`
}

// BaseBot provides common functionality for all bots
type BaseBot struct {
	Config      BotConfig
	LogManager  *LogManager
	Mu          sync.RWMutex
	Ctx         context.Context
	Cancel      context.CancelFunc
	Mux         *http.ServeMux
	NexusConn   *websocket.Conn
	ConnMu      sync.Mutex
	SelfID      string
	BotName     string
	ConfigPtr   any
	RestartFunc func()
	Sections    []ConfigSection
}

func NewBaseBot(defaultLogPort int) *BaseBot {
	ctx, cancel := context.WithCancel(context.Background())
	b := &BaseBot{
		LogManager: NewLogManager(1000),
		Config: BotConfig{
			LogPort: defaultLogPort,
		},
		Ctx:    ctx,
		Cancel: cancel,
		Mux:    http.NewServeMux(),
	}

	// Start background log pusher
	go b.logPusher()

	return b
}

func (b *BaseBot) SetupStandardHandlers(botName string, configPtr any, restartFunc func(), sections []ConfigSection) {
	b.BotName = botName
	b.ConfigPtr = configPtr
	b.RestartFunc = restartFunc
	b.Sections = sections

	b.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config-ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	b.Mux.HandleFunc("/config", b.HandleConfig)
	b.Mux.HandleFunc("/config-ui", b.HandleConfigUI)
}

func (b *BaseBot) HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		b.Mu.RLock()
		defer b.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(b.ConfigPtr)
		return
	}

	if r.Method == http.MethodPost {
		b.Mu.Lock()
		if err := json.NewDecoder(r.Body).Decode(b.ConfigPtr); err != nil {
			b.Mu.Unlock()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save to file while still holding the lock
		b.SaveConfig("config.json")
		b.Mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated successfully"))

		if b.RestartFunc != nil {
			go b.RestartFunc()
		}
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (b *BaseBot) HandleConfigUI(w http.ResponseWriter, r *http.Request) {
	lang := utils.GetLangFromRequest(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create dynamic sections HTML
	var sectionsHTML string
	for _, section := range b.Sections {
		fieldsHTML := ""
		for _, field := range section.Fields {
			inputType := field.Type
			if inputType == "" {
				inputType = "text"
			}
			fieldsHTML += fmt.Sprintf(`
                <div class="form-group">
                    <label>%s</label>
                    <input type="%s" id="%s" value="%v">
                </div>`, field.Label, inputType, field.ID, field.Value)
		}
		sectionsHTML += fmt.Sprintf(`
            <div class="card">
                <div class="section-title">%s</div>
                %s
            </div>`, section.Title, fieldsHTML)
	}

	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - %s</title>
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
        input[type="text"], input[type="number"], input[type="password"] { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: 500; transition: opacity 0.2s; }
        .btn-primary { background: var(--primary-color); color: white; }
        .btn-danger { background: var(--danger-color); color: white; }
        .logs-container { background: #1e1e1e; color: #d4d4d4; padding: 15px; border-radius: 6px; font-family: 'Consolas', monospace; height: 500px; overflow-y: auto; font-size: 13px; line-height: 1.5; }
        .log-line { margin-bottom: 4px; border-bottom: 1px solid #333; padding-bottom: 2px; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">%s</div>
        <div class="nav-item active" onclick="switchTab('config')">%s</div>
        <div class="nav-item" onclick="switchTab('logs')">%s</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            %s
            <div style="text-align: center; margin-top: 30px;">
                <button class="btn btn-primary" style="padding: 15px 40px; font-size: 16px;" onclick="saveConfig()">%s</button>
            </div>
        </div>
        <div id="logs-tab" style="display: none;">
            <div class="card">
                <div class="section-title">%s <button class="btn btn-danger" onclick="clearLogs()">%s</button></div>
                <div id="logs" class="logs-container">Loading...</div>
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
            const cfg = {};
            document.querySelectorAll('#config-tab input').forEach(el => {
                let val = el.value;
                if (el.type === 'number') val = parseInt(val);
                cfg[el.id] = val;
            });
            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(cfg)
            });
            if (resp.ok) {
                alert('Success');
                setTimeout(() => window.location.reload(), 2000);
            } else {
                alert('Failed');
            }
        }
        async function loadLogs() {
            if (currentTab !== 'logs') return;
            try {
                const resp = await fetch('/logs?lines=200');
                const logs = await resp.json();
                const logsDiv = document.getElementById('logs');
                logsDiv.innerHTML = logs.map(line => `+"`"+`<div class="log-line">${line.time} [${line.level}] ${line.message}</div>`+"`"+`).join('');
                logsDiv.scrollTop = logsDiv.scrollHeight;
            } catch (e) {}
            setTimeout(loadLogs, 2000);
        }
        function clearLogs() { document.getElementById('logs').innerText = ''; }
    </script>
</body>
</html>
`, lang, b.BotName, utils.T(lang, "config_center|配置中心"),
		b.BotName, utils.T(lang, "core_config|核心配置"), utils.T(lang, "realtime_logs|实时日志"),
		sectionsHTML,
		utils.T(lang, "save_and_restart|保存配置并重启"),
		utils.T(lang, "system_logs|系统日志"), utils.T(lang, "clear_display|清空显示"))
}

func (b *BaseBot) Info(format string, v ...any) {
	b.LogManager.Log("INFO", fmt.Sprintf(format, v...))
}

func (b *BaseBot) Warn(format string, v ...any) {
	b.LogManager.Log("WARN", fmt.Sprintf(format, v...))
}

func (b *BaseBot) Error(format string, v ...any) {
	b.LogManager.Log("ERROR", fmt.Sprintf(format, v...))
}

func (b *BaseBot) logPusher() {
	for {
		select {
		case <-b.Ctx.Done():
			return
		case entry := <-b.LogManager.logChan:
			b.ConnMu.Lock()
			conn := b.NexusConn
			b.ConnMu.Unlock()

			if conn != nil {
				// Wrap log as OneBot style message for report to Nexus
				b.SendToNexus(map[string]any{
					"post_type": "log",
					"level":     entry.Level,
					"time":      entry.Timestamp.Unix(),
					"message":   entry.Message,
					"source":    b.Config.BotToken,
				})
			}
		}
	}
}

func (b *BaseBot) LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err == nil {
		b.Mu.Lock()
		if err := json.Unmarshal(file, &b.Config); err != nil {
			log.Printf("Error parsing %s: %v", path, err)
		}
		b.Mu.Unlock()
	}

	// Override with environment variables
	if envToken := os.Getenv("BOT_TOKEN"); envToken != "" {
		b.Config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		b.Config.NexusAddr = envAddr
	}
	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		var v int
		if _, err := fmt.Sscanf(envLogPort, "%d", &v); err == nil {
			b.Config.LogPort = v
		}
	}

	if b.Config.NexusAddr == "" {
		b.Config.NexusAddr = "ws://bot-manager:3005"
	}

	return nil
}

func (b *BaseBot) SaveConfig(path string) error {
	var data []byte
	var err error

	if b.ConfigPtr != nil {
		data, err = json.MarshalIndent(b.ConfigPtr, "", "  ")
	} else {
		data, err = json.MarshalIndent(b.Config, "", "  ")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (b *BaseBot) StartHTTPServer() {
	b.Mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		linesStr := r.URL.Query().Get("lines")
		lines := 100
		if linesStr != "" {
			fmt.Sscanf(linesStr, "%d", &lines)
		}
		logs := b.LogManager.GetLogs(lines)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})

	b.Mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", b.Config.LogPort)
	server := &http.Server{
		Addr:    addr,
		Handler: b.Mux,
	}

	go func() {
		b.Mu.RLock()
		useTLS := b.Config.UseTLS
		certFile := b.Config.CertFile
		keyFile := b.Config.KeyFile
		b.Mu.RUnlock()

		if useTLS && certFile != "" && keyFile != "" {
			log.Printf("Starting HTTPS/WSS server on %s", addr)
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS server failed: %v", err)
			}
		} else {
			log.Printf("Starting HTTP/WS server on %s", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server failed: %v", err)
			}
		}
	}()

	go func() {
		<-b.Ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
}

func (b *BaseBot) WaitExitSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down...")
	b.Cancel()
}

// StartNexusConnection connects to BotNexus and handles reconnection
func (b *BaseBot) StartNexusConnection(ctx context.Context, addr, platform, selfID string, commandHandler func([]byte)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Connecting to BotNexus at %s...", addr)
				header := http.Header{}
				header.Add("X-Self-ID", selfID)
				header.Add("X-Platform", platform)

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

				b.ConnMu.Lock()
				b.NexusConn = conn
				b.ConnMu.Unlock()
				log.Println("Connected to BotNexus!")

				// Send Lifecycle Event
				b.SendToNexus(map[string]any{
					"post_type":       "meta_event",
					"meta_event_type": "lifecycle",
					"sub_type":        "connect",
					"self_id":         selfID,
					"time":            time.Now().Unix(),
				})

				// Handle incoming commands
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("BotNexus disconnected: %v", err)
						b.ConnMu.Lock()
						if b.NexusConn == conn {
							b.NexusConn = nil
						}
						b.ConnMu.Unlock()
						break
					}
					commandHandler(message)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (b *BaseBot) SendToNexus(msg any) {
	b.ConnMu.Lock()
	defer b.ConnMu.Unlock()
	if b.NexusConn == nil {
		return
	}
	if err := b.NexusConn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send to Nexus: %v", err)
		b.NexusConn.Close()
		b.NexusConn = nil
	}
}
