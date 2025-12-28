package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	common "BotMatrix/src/Common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botService *common.BaseBot
	tgBot      *tgbotapi.BotAPI
	botCtx     context.Context
	botCancel  context.CancelFunc
)

func main() {
	botService = common.NewBaseBot(8087)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	botService.LoadConfig("config.json")

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

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	botToken := botService.Config.BotToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if botToken == "" {
		log.Println("Telegram bot token is not set, bot will not start")
		return
	}

	var err error
	tgBot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("Failed to create Telegram Bot: %v", err)
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	tgBot.Debug = true
	selfID := fmt.Sprintf("%d", tgBot.Self.ID)
	log.Printf("Authorized on account %s (ID: %s)", tgBot.Self.UserName, selfID)

	// Connect to Nexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Telegram", selfID, handleNexusCommand)

	// Start Polling
	go func(ctx context.Context) {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := tgBot.GetUpdatesChan(u)

		for {
			select {
			case <-ctx.Done():
				tgBot.StopReceivingUpdates()
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				if update.Message == nil {
					continue
				}
				handleMessage(update.Message, selfID)
			}
		}
	}(botCtx)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func handleMessage(msg *tgbotapi.Message, selfID string) {
	// Handle Multimedia
	if msg.Photo != nil && len(msg.Photo) > 0 {
		photo := msg.Photo[len(msg.Photo)-1]
		fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
		file, err := tgBot.GetFile(fileConfig)
		if err == nil {
			botService.Mu.RLock()
			token := botService.Config.BotToken
			botService.Mu.RUnlock()
			url := file.Link(token)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	} else if msg.Sticker != nil {
		fileConfig := tgbotapi.FileConfig{FileID: msg.Sticker.FileID}
		file, err := tgBot.GetFile(fileConfig)
		if err == nil {
			botService.Mu.RLock()
			token := botService.Config.BotToken
			botService.Mu.RUnlock()
			url := file.Link(token)
			msg.Text += fmt.Sprintf("[CQ:image,file=%s]", url)
		}
	}

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	obMsg := map[string]any{
		"post_type":    "message",
		"message_type": "group",
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "normal",
		"message_id":   fmt.Sprintf("%d", msg.MessageID),
		"user_id":      fmt.Sprintf("%d", msg.From.ID),
		"message":      msg.Text,
		"raw_message":  msg.Text,
		"sender": map[string]any{
			"user_id":  fmt.Sprintf("%d", msg.From.ID),
			"nickname": msg.From.UserName,
		},
	}

	if msg.Chat.IsPrivate() {
		obMsg["message_type"] = "private"
	} else {
		obMsg["group_id"] = fmt.Sprintf("%d", msg.Chat.ID)
	}

	botService.SendToNexus(obMsg)
}

func handleNexusCommand(action string, params map[string]any) (any, error) {
	switch action {
	case "send_msg":
		return sendTelegramMessage(params)
	case "delete_msg":
		return deleteTelegramMessage(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func sendTelegramMessage(params map[string]any) (any, error) {
	if tgBot == nil {
		return nil, fmt.Errorf("bot is not running")
	}

	chatIDStr, ok := params["chat_id"].(string)
	if !ok {
		// Try group_id or user_id as fallback
		if gid, ok := params["group_id"].(string); ok {
			chatIDStr = gid
		} else if uid, ok := params["user_id"].(string); ok {
			chatIDStr = uid
		} else {
			return nil, fmt.Errorf("missing chat_id")
		}
	}

	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message")
	}

	var chatID int64
	fmt.Sscanf(chatIDStr, "%d", &chatID)

	msg := tgbotapi.NewMessage(chatID, message)
	sent, err := tgBot.Send(msg)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message_id": fmt.Sprintf("%d", sent.MessageID),
	}, nil
}

func deleteTelegramMessage(params map[string]any) (any, error) {
	if tgBot == nil {
		return nil, fmt.Errorf("bot is not running")
	}

	chatIDStr, _ := params["chat_id"].(string)
	messageIDStr, ok := params["message_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message_id")
	}

	var chatID int64
	var messageID int
	fmt.Sscanf(chatIDStr, "%d", &chatID)
	fmt.Sscanf(messageIDStr, "%d", &messageID)

	delMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := tgBot.Request(delMsg)
	return nil, err
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(botService.Config)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig common.BotConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		botService.Config.BotToken = newConfig.BotToken
		botService.Config.NexusAddr = newConfig.NexusAddr
		botService.Mu.Unlock()

		botService.SaveConfig("config.json")

		go restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarting"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := botService.Config
	botService.Mu.RUnlock()

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>TelegramBot 控制面板</title>
    <style>
        :root {
            --primary-color: #0088cc;
            --bg-color: #f4f7f9;
        }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background: var(--bg-color); margin: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .card { background: white; padding: 2rem; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); width: 100%%; max-width: 400px; }
        h1 { margin-top: 0; color: #333; font-size: 1.5rem; text-align: center; }
        .form-group { margin-bottom: 1.5rem; }
        label { display: block; margin-bottom: 0.5rem; color: #666; font-size: 0.9rem; }
        input { width: 100%%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 6px; box-sizing: border-box; font-size: 1rem; }
        button { width: 100%%; padding: 0.75rem; background: var(--primary-color); color: white; border: none; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; transition: opacity 0.2s; }
        button:hover { opacity: 0.9; }
        .footer { margin-top: 1.5rem; text-align: center; font-size: 0.8rem; color: #999; }
    </style>
</head>
<body>
    <div class="card">
        <h1>TelegramBot 配置</h1>
        <div class="form-group">
            <label>Bot Token</label>
            <input type="password" id="botToken" value="%s" placeholder="输入 Telegram Bot Token">
        </div>
        <div class="form-group">
            <label>Nexus 地址</label>
            <input type="text" id="nexusAddr" value="%s" placeholder="例如 ws://localhost:8000/bot/ws">
        </div>
        <button onclick="saveConfig()">保存并重启</button>
        <div class="footer">BotMatrix Ecosystem</div>
    </div>
    <script>
        function saveConfig() {
            const botToken = document.getElementById('botToken').value;
            const nexusAddr = document.getElementById('nexusAddr').value;
            fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ bot_token: botToken, nexus_addr: nexusAddr })
            }).then(res => {
                if (res.ok) alert('配置已保存，机器人正在重启...');
                else alert('保存失败');
            });
        }
    </script>
</body>
</html>`, cfg.BotToken, cfg.NexusAddr)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
