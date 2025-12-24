package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

type Config struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	dg          *discordgo.Session
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

func main() {
	log.SetOutput(logManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Ensure LogPort is set
	configMutex.Lock()
	if config.LogPort == 0 {
		config.LogPort = 3134 // Default for DiscordBot
	}
	configMutex.Unlock()

	startBot()

	// Start HTTP Server for Web UI and Logs
	go startHTTPServer()

	// Wait for signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	stopBot()
}

func loadConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	} else {
		log.Printf("config.json not found: %v", err)
	}

	if envToken := os.Getenv("DISCORD_BOT_TOKEN"); envToken != "" {
		config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		config.NexusAddr = envAddr
	}
	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		if v, err := strconv.Atoi(envLogPort); err == nil {
			config.LogPort = v
		}
	}

	if config.NexusAddr == "" {
		config.NexusAddr = "ws://bot-manager:3005"
	}
}

func startBot() {
	botCtx, botCancel = context.WithCancel(context.Background())

	configMutex.RLock()
	token := config.BotToken
	configMutex.RUnlock()

	if token == "" {
		log.Println("WARNING: Discord Bot Token is not configured. Bot will not start until configured via Web UI.")
		return
	}

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Printf("Error creating Discord session: %v", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		log.Printf("Error opening Discord connection: %v", err)
		return
	}

	selfID = dg.State.User.ID
	log.Printf("Bot is now running. Logged in as %s#%s (%s)", dg.State.User.Username, dg.State.User.Discriminator, selfID)

	startNexus()

	go func() {
		<-botCtx.Done()
		log.Println("Stopping Discord session...")
		dg.Close()
	}()
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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Handle Attachments (Images)
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 || attachment.Height > 0 { // Simple check if it's an image
			m.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		}
	}

	log.Printf("[%s] %s", m.Author.Username, m.Content)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  m.ID,
		"user_id":     m.Author.ID,
		"message":     m.Content,
		"raw_message": m.Content,
		"sender": map[string]interface{}{
			"user_id":  m.Author.ID,
			"nickname": m.Author.Username,
		},
	}

	if m.GuildID != "" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = m.ChannelID // OneBot group_id maps to Discord Channel ID for simplicity
		// Or we could use GuildID, but messages happen in channels.
		// For OneBot compatibility, mapping ChannelID to GroupID is more practical for chat bots.
	} else {
		obMsg["message_type"] = "private"
	}

	sendToNexus(obMsg)
}

func connectToNexus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			configMutex.RLock()
			addr := config.NexusAddr
			configMutex.RUnlock()

			log.Printf("Connecting to BotNexus at %s...", addr)
			header := http.Header{}
			header.Add("X-Self-ID", selfID)
			header.Add("X-Platform", "Discord")

			conn, _, err := websocket.DefaultDialer.Dial(addr, header)
			if err != nil {
				log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
				select {
				case <-time.After(5 * time.Second):
					continue
				case <-ctx.Done():
					return
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

			// Message reading loop
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
	configMutex.RLock()
	cfg := config
	configMutex.RUnlock()

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>DiscordBot Configuration</title>
    <style>
        body { font-family: sans-serif; margin: 20px; background: #f0f2f5; }
        .container { max-width: 600px; margin: auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #1a73e8; }
        .field { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="text"], input[type="number"], input[type="password"] {
            width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;
        }
        button {
            background: #1a73e8; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer;
        }
        button:hover { background: #1557b0; }
        .logs { margin-top: 20px; background: #202124; color: #f1f3f4; padding: 15px; border-radius: 4px; font-family: monospace; height: 300px; overflow-y: auto; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <h1>DiscordBot Configuration</h1>
        <form id="configForm">
            <div class="field">
                <label>Discord Bot Token:</label>
                <input type="password" name="bot_token" value="{{.BotToken}}">
            </div>
            <div class="field">
                <label>Nexus Address:</label>
                <input type="text" name="nexus_addr" value="{{.NexusAddr}}">
            </div>
            <div class="field">
                <label>Web UI Port (LogPort):</label>
                <input type="number" name="log_port" value="{{.LogPort}}">
            </div>
            <button type="submit">Save & Restart</button>
        </form>

        <h2>Recent Logs</h2>
        <div class="logs" id="logBox">Loading logs...</div>
    </div>

    <script>
        document.getElementById('configForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = {};
            formData.forEach((value, key) => {
                if (key === 'log_port') {
                    data[key] = parseInt(value);
                } else {
                    data[key] = value;
                }
            });

            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            if (resp.ok) {
                alert('Configuration saved and bot restarting...');
                if (data.log_port !== {{.LogPort}}) {
                    window.location.href = 'http://' + window.location.hostname + ':' + data.log_port + '/config-ui';
                }
            } else {
                alert('Error: ' + await resp.text());
            }
        };

        async function fetchLogs() {
            try {
                const resp = await fetch('/logs?lines=50');
                const text = await resp.text();
                const logBox = document.getElementById('logBox');
                const isScrolledToBottom = logBox.scrollHeight - logBox.clientHeight <= logBox.scrollTop + 1;
                logBox.textContent = text;
                if (isScrolledToBottom) {
                    logBox.scrollTop = logBox.scrollHeight;
                }
            } catch (e) {}
        }

        setInterval(fetchLogs, 2000);
        fetchLogs();
    </script>
</body>
</html>
`
	t, err := template.New("ui").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, cfg)
}

func sendToNexus(msg interface{}) {
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
			sendDiscordMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			// Create DM channel first
			ch, err := dg.UserChannelCreate(userID)
			if err == nil {
				sendDiscordMessage(ch.ID, text, cmd.Echo)
			} else {
				log.Printf("Failed to create DM: %v", err)
			}
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteDiscordMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": dg.State.User.Username,
			},
			"echo": cmd.Echo,
		})
	}
}

func sendDiscordMessage(channelID, text, echo string) {
	msg, err := dg.ChannelMessageSend(channelID, text)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}
	log.Printf("Sent message to %s: %s", channelID, text)
	// Return composite ID: "channelID:messageID"
	compositeID := fmt.Sprintf("%s:%s", channelID, msg.ID)
	sendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteDiscordMessage(compositeID, echo string) {
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		sendToNexus(map[string]interface{}{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}
	channelID := parts[0]
	messageID := parts[1]

	err := dg.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		sendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s in channel %s", messageID, channelID)
	sendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
}
