package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/work"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
)

// Config holds the bot configuration
type Config struct {
	CorpID         string `json:"corp_id"`
	AgentID        string `json:"agent_id"`
	Secret         string `json:"secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	ListenPort     int    `json:"listen_port"`
	NexusAddr      string `json:"nexus_addr"`
	LogPort        int    `json:"log_port"`
}

var (
	config      Config
	configMutex sync.RWMutex
	conn        *websocket.Conn
	connMutex   sync.Mutex
	wc          *work.Work
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

	file, err := os.Open("config.json")
	if err != nil {
		log.Printf("Error opening config.json: %v. Using default values.", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Error decoding config.json: %v", err)
		return
	}
	selfID = config.AgentID
}

func connectToNexus() {
	for {
		select {
		case <-nexusCtx.Done():
			log.Println("Nexus connection stopped.")
			return
		default:
			configMutex.RLock()
			addr := config.NexusAddr
			configMutex.RUnlock()

			log.Printf("Connecting to BotNexus at %s...", addr)
			c, _, err := websocket.DefaultDialer.Dial(addr, nil)
			if err != nil {
				log.Printf("Connection error: %v. Retrying in 5s...", err)
				select {
				case <-nexusCtx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			connMutex.Lock()
			conn = c
			connMutex.Unlock()
			log.Println("Connected to BotNexus!")

			// Send Lifecycle Event
			sendEvent(map[string]interface{}{
				"post_type":       "meta_event",
				"meta_event_type": "lifecycle",
				"sub_type":        "connect",
				"self_id":         selfID,
				"time":            time.Now().Unix(),
			})

			// Send Heartbeat
			heartbeatCtx, heartbeatCancel := context.WithCancel(nexusCtx)
			go func(ctx context.Context) {
				ticker := time.NewTicker(10 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						connMutex.Lock()
						if conn == nil {
							connMutex.Unlock()
							return
						}
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
						connMutex.Unlock()
					}
				}
			}(heartbeatCtx)

			// Handle Incoming Actions
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					log.Printf("Read error: %v", err)
					connMutex.Lock()
					if conn == c {
						conn = nil
					}
					connMutex.Unlock()
					heartbeatCancel()
					break
				}
				handleAction(msg)
			}
		}
	}
}

func sendEvent(event map[string]interface{}) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if conn == nil {
		return
	}
	event["platform"] = "wework"
	err := conn.WriteJSON(event)
	if err != nil {
		log.Printf("Write error: %v", err)
		conn = nil
	}
}

func handleAction(msg []byte) {
	var action map[string]interface{}
	if err := json.Unmarshal(msg, &action); err != nil {
		log.Printf("JSON Unmarshal error: %v", err)
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
		userID, _ := params["user_id"].(string)
		msgContent, _ := params["message"].(string)

		// Send text message
		req := message.SendTextRequest{
			SendRequestCommon: &message.SendRequestCommon{
				ToUser:  userID,
				MsgType: "text",
				AgentID: config.AgentID,
			},
			Text: message.TextField{
				Content: msgContent,
			},
		}

		if wc != nil {
			msgManager := wc.GetMessage()
			msgID, err := msgManager.SendText(req)
			if err != nil {
				log.Printf("Failed to send WeCom message: %v", err)
				response["status"] = "failed"
				response["retcode"] = -1
			} else {
				response["data"] = map[string]interface{}{
					"message_id": msgID,
				}
			}
		} else {
			response["status"] = "failed"
			response["retcode"] = -1
		}

	case "delete_msg":
		msgID, _ := params["message_id"].(string)
		if msgID != "" && wc != nil {
			err := recallMessage(msgID)
			if err != nil {
				log.Printf("Failed to recall WeCom message: %v", err)
				response["status"] = "failed"
				response["retcode"] = -1
			}
		}
	}

	connMutex.Lock()
	if conn != nil {
		conn.WriteJSON(response)
	}
	connMutex.Unlock()
}

func recallMessage(msgID string) error {
	token, err := wc.GetContext().GetAccessToken()
	if err != nil {
		return err
	}

	url := "https://qyapi.weixin.qq.com/cgi-bin/message/recall?access_token=" + token
	payload := map[string]string{"msgid": msgID}
	jsonBody, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("wecom recall error: %v", result)
	}
	return nil
}

func startBot() {
	botCtx, botCancel = context.WithCancel(context.Background())

	configMutex.RLock()
	wcConfig := &workConfig.Config{
		CorpID:         config.CorpID,
		AgentID:        config.AgentID,
		CorpSecret:     config.Secret,
		Token:          config.Token,
		EncodingAESKey: config.EncodingAESKey,
		Cache:          cache.NewMemory(),
	}
	port := config.ListenPort
	configMutex.RUnlock()

	wc = wechat.NewWechat().GetWork(wcConfig)

	// Start HTTP Server for Callbacks
	r := gin.Default()
	r.GET("/callback", handleCallback)
	r.POST("/callback", handleCallback)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	log.Printf("Starting WeWork Callback Server on :%d", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Callback server failed: %v", err)
		}
	}()

	go func() {
		<-botCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
		log.Println("Callback server stopped.")
	}()

	startNexus()
}

func stopBot() {
	stopNexus()
	if botCancel != nil {
		botCancel()
	}
}

func startNexus() {
	nexusCtx, nexusCancel = context.WithCancel(botCtx)
	go connectToNexus()
}

func stopNexus() {
	if nexusCancel != nil {
		nexusCancel()
	}
	connMutex.Lock()
	if conn != nil {
		conn.Close()
		conn = nil
	}
	connMutex.Unlock()
}

func main() {
	log.SetOutput(logManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Ensure LogPort is set
	configMutex.Lock()
	if config.LogPort == 0 {
		config.LogPort = 8083 // Default for WeWorkBot
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
    <title>WeWorkBot Configuration</title>
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
        <h1>WeWorkBot Configuration</h1>
        <form id="configForm">
            <div class="field">
                <label>Corp ID:</label>
                <input type="text" name="corp_id" value="{{.CorpID}}">
            </div>
            <div class="field">
                <label>Agent ID:</label>
                <input type="text" name="agent_id" value="{{.AgentID}}">
            </div>
            <div class="field">
                <label>Secret:</label>
                <input type="password" name="secret" value="{{.Secret}}">
            </div>
            <div class="field">
                <label>Token:</label>
                <input type="text" name="token" value="{{.Token}}">
            </div>
            <div class="field">
                <label>Encoding AES Key:</label>
                <input type="text" name="encoding_aes_key" value="{{.EncodingAESKey}}">
            </div>
            <div class="field">
                <label>Callback Listen Port:</label>
                <input type="number" name="listen_port" value="{{.ListenPort}}">
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
                if (key === 'listen_port' || key === 'log_port') {
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

func handleCallback(c *gin.Context) {
	// TODO: Work module in v2.1.11 does not have GetServer.
	// We need to implement manual callback handling (signature verification, decryption, XML parsing).
	// For now, we just acknowledge the request to avoid timeout errors on WeCom side.

	echoStr := c.Query("echostr")
	if echoStr != "" {
		// Verify URL (GET request)
		// signature := c.Query("msg_signature")
		// timestamp := c.Query("timestamp")
		// nonce := c.Query("nonce")
		// if util.Signature(config.Token, timestamp, nonce, echoStr) == signature { ... }

		// For verification, we usually need to decrypt the echostr using EncodingAESKey.
		// Since we don't have the full logic yet, we might fail the verification step in WeCom admin panel.
		// But if this is already configured, we just need to handle POST messages.

		// We can use util.DecryptMsg if needed.
		// But for now, let's just log.
		log.Printf("Received verification request: %s", echoStr)
	}

	// Handle POST messages
	if c.Request.Method == "POST" {
		log.Println("Received WeCom callback (POST)")
		// TODO: Parse body, decrypt, broadcast to Nexus
	}

	c.String(200, "success")
}
