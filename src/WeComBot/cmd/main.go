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
	"strconv"
	"sync"
	"time"

	"BotMatrix/src/Common"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/work"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
)

// WeComConfig extends common.BotConfig with WeCom specific fields
type WeComConfig struct {
	common.BotConfig
	CorpID         string `json:"corp_id"`
	AgentID        string `json:"agent_id"`
	Secret         string `json:"secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	ListenPort     int    `json:"listen_port"` // Callback listen port
}

var (
	wecomCfg   WeComConfig
	botService *common.BaseBot
	wc         *work.Work
	botCtx     context.Context
	botCancel  context.CancelFunc
	mu         sync.RWMutex
)

func loadConfig() {
	botService.LoadConfig("config.json")

	// Sync to wecomCfg
	botService.Mu.RLock()
	wecomCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load extra fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &wecomCfg)
	}

	if wecomCfg.LogPort == 0 {
		wecomCfg.LogPort = 8083 // Default for WeComBot
	}
}

func handleAction(action map[string]interface{}) (interface{}, error) {
	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]interface{})

	mu.RLock()
	currentWC := wc
	agentID := wecomCfg.AgentID
	mu.RUnlock()

	switch actionType {
	case "send_private_msg", "send_msg":
		userID, _ := params["user_id"].(string)
		msgContent, _ := params["message"].(string)

		if currentWC == nil {
			return nil, fmt.Errorf("WeCom client not initialized")
		}

		req := message.SendTextRequest{
			SendRequestCommon: &message.SendRequestCommon{
				ToUser:  userID,
				MsgType: "text",
				AgentID: agentID,
			},
			Text: message.TextField{
				Content: msgContent,
			},
		}

		msgManager := currentWC.GetMessage()
		msgID, err := msgManager.SendText(req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"message_id": msgID}, nil

	case "delete_msg":
		msgID, _ := params["message_id"].(string)
		if msgID != "" && currentWC != nil {
			err := recallMessage(currentWC, msgID)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return nil, fmt.Errorf("invalid message_id or client not initialized")

	case "get_login_info":
		return map[string]interface{}{
			"user_id":  agentID,
			"nickname": "WeCom Agent",
		}, nil
	}

	return nil, fmt.Errorf("unsupported action: %s", actionType)
}

func recallMessage(w *work.Work, msgID string) error {
	token, err := w.GetContext().GetAccessToken()
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

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	botCtx, botCancel = context.WithCancel(context.Background())

	mu.Lock()
	wcConfig := &workConfig.Config{
		CorpID:         wecomCfg.CorpID,
		AgentID:        wecomCfg.AgentID,
		CorpSecret:     wecomCfg.Secret,
		Token:          wecomCfg.Token,
		EncodingAESKey: wecomCfg.EncodingAESKey,
		Cache:          cache.NewMemory(),
	}
	wc = wechat.NewWechat().GetWork(wcConfig)
	listenPort := wecomCfg.ListenPort
	agentID := wecomCfg.AgentID
	mu.Unlock()

	// Start Nexus connection
	go botService.StartNexusConnection(botCtx, nexusAddr, "WeCom", agentID, handleNexusCommand)

	// Start WeCom Callback Server
	if listenPort > 0 {
		go startCallbackServer(botCtx, listenPort)
	}

	log.Println("WeCom Bot started")
}

func stopBot() {
	if botCancel != nil {
		botCancel()
	}
	mu.Lock()
	wc = nil
	mu.Unlock()
}

func startCallbackServer(ctx context.Context, port int) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/callback", handleCallback)
	r.POST("/callback", handleCallback)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	log.Printf("Starting WeCom Callback Server on :%d", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Callback server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	log.Println("Callback server stopped.")
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

	log.Printf("Received Action: %s", cmd.Action)
	result, err := handleAction(map[string]interface{}{
		"action": cmd.Action,
		"params": cmd.Params,
	})

	resp := map[string]interface{}{
		"echo": cmd.Echo,
	}
	if err != nil {
		resp["status"] = "failed"
		resp["msg"] = err.Error()
	} else {
		resp["status"] = "ok"
		resp["data"] = result
	}
	botService.SendToNexus(resp)
}

func handleCallback(c *gin.Context) {
	mu.RLock()
	currentWC := wc
	mu.RUnlock()

	if currentWC == nil {
		c.String(http.StatusInternalServerError, "WeCom client not initialized")
		return
	}

	server := currentWC.GetServer(c.Request, c.Writer)

	// Set message handler
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		log.Printf("Received WeCom message: %s from %s", msg.Content, msg.FromUserName)

		// Broadcast to Nexus
		botService.SendToNexus(map[string]interface{}{
			"post_type":    "message",
			"message_type": "private",
			"user_id":      msg.FromUserName,
			"message":      msg.Content,
			"raw_message":  msg.Content,
			"self_id":      wecomCfg.AgentID,
			"platform":     "wecom",
			"time":         time.Now().Unix(),
		})

		return nil // No auto reply
	})

	err := server.Serve()
	if err != nil {
		log.Printf("WeCom server serve error: %v", err)
		return
	}
}

func main() {
	botService = common.NewBaseBot(8083)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wecomCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newCfg WeComConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, _ := json.MarshalIndent(newCfg, "", "  ")
		if err := os.WriteFile("config.json", data, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		botService.Mu.Lock()
		wecomCfg = newCfg
		botService.Config = wecomCfg.BotConfig
		botService.Mu.Unlock()

		restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarted"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := wecomCfg
	botService.Mu.RUnlock()

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>WeComBot Configuration</title>
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
        <h1>WeComBot Configuration</h1>
        <form id="configForm">
            <div class="field">
                <label>Nexus Address:</label>
                <input type="text" name="nexus_addr" value="{{.NexusAddr}}">
            </div>
            <div class="field">
                <label>Web UI Port (LogPort):</label>
                <input type="number" name="log_port" value="{{.LogPort}}">
            </div>
            <hr>
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
            <button type="submit">Save & Restart</button>
        </form>

        <h2>Recent Logs</h2>
        <div class="logs" id="logBox">Loading logs...</div>
    </div>

    <script>
        document.getElementById('configForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = {
                nexus_addr: formData.get('nexus_addr'),
                log_port: parseInt(formData.get('log_port')),
                corp_id: formData.get('corp_id'),
                agent_id: formData.get('agent_id'),
                secret: formData.get('secret'),
                token: formData.get('token'),
                encoding_aes_key: formData.get('encoding_aes_key'),
                listen_port: parseInt(formData.get('listen_port'))
            };

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
