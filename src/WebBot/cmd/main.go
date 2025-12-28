package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	common "BotMatrix/src/Common"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebAppConfig 结构化存储每个 App 的配置
type WebAppConfig struct {
	AppKey     string `json:"app_key"`
	Title      string `json:"title"`
	ThemeColor string `json:"theme_color"`
	WelcomeMsg string `json:"welcome_msg"`
	BotSelfID  string `json:"bot_self_id"` // 该 App 对应的虚拟机器人 ID
}

// WebConfig 整体配置
type WebConfig struct {
	common.BotConfig
	Apps []WebAppConfig `json:"apps"`
}

type WebUser struct {
	ID       string          `json:"id"`
	Conn     *websocket.Conn `json:"-"`
	AppKey   string          `json:"app_key"`
	LastSeen time.Time       `json:"last_seen"`
	Nickname string          `json:"nickname"`
	Mu       sync.Mutex
}

var (
	botService *common.BaseBot
	webCfg     WebConfig
	appsMap    = make(map[string]WebAppConfig) // 快速查找
	appsMu     sync.RWMutex
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	users   = make(map[string]*WebUser)
	usersMu sync.RWMutex
)

func main() {
	botService = common.NewBaseBot(3137)
	log.SetOutput(botService.LogManager)

	loadConfig()

	botService.Mux.HandleFunc("/ws/widget", handleWidgetWebSocket)
	botService.Mux.Handle("/widget/", http.StripPrefix("/widget/", http.FileServer(http.Dir("widget"))))
	botService.Mux.HandleFunc("/config", handleConfig)
	botService.Mux.HandleFunc("/config-ui", handleConfigUI)

	go botService.StartHTTPServer()

	restartBot()

	botService.WaitExitSignal()
}

func loadConfig() {
	botService.LoadConfig("config.json")

	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &webCfg)
	}

	appsMu.Lock()
	appsMap = make(map[string]WebAppConfig)
	for _, app := range webCfg.Apps {
		appsMap[app.AppKey] = app
	}
	appsMu.Unlock()
}

func restartBot() {
	botService.LogManager.Info().Msg("Restarting WebBot Multi-tenant Hub...")

	botService.Mu.RLock()
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if nexusAddr == "" {
		return
	}

	ctx := context.Background()
	botService.StartNexusConnection(ctx, nexusAddr, "Web", "web-gateway-hub", handleNexusCommand)
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
		SelfID string         `json:"self_id"`
	}

	if err := json.Unmarshal(data, &cmd); err != nil {
		return
	}

	switch cmd.Action {
	case "send_msg":
		userID, _ := cmd.Params["user_id"].(string)
		content, _ := cmd.Params["content"].(string)
		sendToWebUser(userID, content, cmd.Echo)
	}
}

func handleWidgetWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	appKey := r.URL.Query().Get("app_key")
	userID := r.URL.Query().Get("user_id")

	appsMu.RLock()
	app, ok := appsMap[appKey]
	appsMu.RUnlock()

	if !ok {
		conn.WriteJSON(map[string]string{"error": "Invalid AppKey"})
		conn.Close()
		return
	}

	if userID == "" {
		userID = uuid.New().String()
	}

	user := &WebUser{
		ID:       userID,
		Conn:     conn,
		AppKey:   appKey,
		LastSeen: time.Now(),
		Nickname: "Visitor_" + userID[:4],
	}

	usersMu.Lock()
	users[userID] = user
	usersMu.Unlock()

	conn.WriteJSON(map[string]any{
		"type": "init",
		"data": map[string]any{
			"user_id": userID,
			"title":   app.Title,
			"theme":   app.ThemeColor,
			"welcome": app.WelcomeMsg,
		},
	})

	defer func() {
		conn.Close()
		usersMu.Lock()
		delete(users, userID)
		usersMu.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var webMsg struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(message, &webMsg); err != nil {
			continue
		}

		botService.SendToNexus(map[string]any{
			"post_type":    "message",
			"message_type": "private",
			"user_id":      userID,
			"message":      webMsg.Content,
			"self_id":      app.BotSelfID,
			"app_key":      appKey,
			"nickname":     user.Nickname,
			"time":         time.Now().Unix(),
		})
	}
}

func sendToWebUser(userID, content, echo string) {
	usersMu.RLock()
	user, ok := users[userID]
	usersMu.RUnlock()

	if !ok {
		return
	}

	msg, _ := json.Marshal(map[string]any{
		"type":    "text",
		"content": content,
		"from":    "bot",
		"time":    time.Now().Unix(),
	})

	user.Mu.Lock()
	user.Conn.WriteMessage(websocket.TextMessage, msg)
	user.Mu.Unlock()
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(webCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig WebConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		webCfg = newConfig
		botService.Config = newConfig.BotConfig
		botService.Mu.Unlock()

		data, _ := json.MarshalIndent(webCfg, "", "  ")
		os.WriteFile("config.json", data, 0644)

		loadConfig()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated"))
		go restartBot()
	}
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := webCfg
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
	<title>WebBot Gateway Config</title>
	<style>
		body { font-family: sans-serif; margin: 20px; background: #f0f2f5; }
		.card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); max-width: 800px; margin: auto; }
		h2 { margin-top: 0; color: #1a73e8; }
		.field { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
		input[type="text"], input[type="number"], textarea { width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		.checkbox-field { display: flex; align-items: center; gap: 10px; }
		.checkbox-field input { width: auto; }
		button { background: #1a73e8; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; font-weight: bold; }
		button:hover { background: #1557b0; }
		.section { border-top: 1px solid #eee; margin-top: 20px; pt: 20px; }
	</style>
</head>
<body>
	<div class="card">
		<h2>WebBot 网关配置</h2>
		<form id="configForm">
			<div class="field">
				<label>Nexus 地址</label>
				<input type="text" name="nexus_addr" value="%s">
			</div>
			<div class="field">
				<label>服务端口 (LogPort)</label>
				<input type="number" name="log_port" value="%d">
			</div>

			<div class="section">
				<h3>安全与 TLS (WSS)</h3>
				<div class="field checkbox-field">
					<input type="checkbox" name="use_tls" id="use_tls" %s>
					<label for="use_tls">启用 TLS (HTTPS/WSS)</label>
				</div>
				<div class="field">
					<label>证书文件路径 (Cert File)</label>
					<input type="text" name="cert_file" value="%s" placeholder="path/to/cert.pem">
				</div>
				<div class="field">
					<label>私钥文件路径 (Key File)</label>
					<input type="text" name="key_file" value="%s" placeholder="path/to/key.pem">
				</div>
			</div>

			<button type="submit">保存并重启</button>
		</form>
	</div>
	<script>
		document.getElementById('configForm').onsubmit = async (e) => {
			e.preventDefault();
			const formData = new FormData(e.target);
			const config = {
				nexus_addr: formData.get('nexus_addr'),
				log_port: parseInt(formData.get('log_port')),
				use_tls: formData.get('use_tls') === 'on',
				cert_file: formData.get('cert_file'),
				key_file: formData.get('key_file'),
				apps: %s // 保持原有的 Apps 配置
			};
			const resp = await fetch('/config', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(config)
			});
			if (resp.ok) alert('保存成功！');
			else alert('保存失败：' + await resp.text());
		};
	</script>
</body>
</html>
	`, cfg.NexusAddr, cfg.LogPort, func() string {
		if cfg.UseTLS {
			return "checked"
		}
		return ""
	}(), cfg.CertFile, cfg.KeyFile, func() string {
		data, _ := json.Marshal(cfg.Apps)
		return string(data)
	}())
}
