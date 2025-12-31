package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"BotMatrix/common/bot"

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
	bot.BotConfig
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
	botService *bot.BaseBot
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
	botService = bot.NewBaseBot(3137)
	log.SetOutput(botService.LogManager)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("WebBot", &webCfg, restartBot, []bot.ConfigSection{
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: webCfg.NexusAddr},
				{Label: "服务端口 (LogPort)", ID: "log_port", Type: "number", Value: webCfg.LogPort},
			},
		},
		{
			Title: "安全与 TLS (WSS) 配置",
			Fields: []bot.ConfigField{
				{Label: "启用 TLS (HTTPS/WSS)", ID: "use_tls", Type: "checkbox", Value: webCfg.UseTLS},
				{Label: "证书文件路径 (Cert File)", ID: "cert_file", Type: "text", Value: webCfg.CertFile},
				{Label: "私钥文件路径 (Key File)", ID: "key_file", Type: "text", Value: webCfg.KeyFile},
			},
		},
	})

	botService.Mux.HandleFunc("/ws/widget", handleWidgetWebSocket)
	botService.Mux.Handle("/widget/", http.StripPrefix("/widget/", http.FileServer(http.Dir("widget"))))

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

	// Sync local config
	botService.Mu.RLock()
	webCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	appsMu.Lock()
	appsMap = make(map[string]WebAppConfig)
	for _, app := range webCfg.Apps {
		appsMap[app.AppKey] = app
	}
	appsMu.Unlock()
}

func restartBot() {
	botService.Info("Restarting WebBot Multi-tenant Hub...")

	botService.Mu.RLock()
	// Sync botService.Config from webCfg
	botService.Config = webCfg.BotConfig
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
