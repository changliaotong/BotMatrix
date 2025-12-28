package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	common "BotMatrix/src/Common"

	"github.com/bwmarrin/discordgo"
)

var (
	botService *common.BaseBot
	dg         *discordgo.Session
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
)

func main() {
	// Initialize base bot with default log port
	botService = common.NewBaseBot(3134)

	// Setup logging to use the common LogManager
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	botService.LoadConfig("config.json")

	// Setup HTTP handlers for config and UI
	botService.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config-ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	botService.Mux.HandleFunc("/config", handleConfig)
	botService.Mux.HandleFunc("/config-ui", handleConfigUI)

	// Start common HTTP services (health, logs, config)
	go botService.StartHTTPServer()

	// Start platform specific bot
	restartBot()

	// Wait for exit signal and handle cleanup
	botService.WaitExitSignal()
	stopBot()
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	token := botService.Config.BotToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if token == "" {
		log.Println("WARNING: Discord Bot Token is not configured. Bot will not start until configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

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

	// Connect to Nexus for central management
	botService.StartNexusConnection(botCtx, nexusAddr, "Discord", selfID, handleNexusCommand)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
	if dg != nil {
		log.Println("Stopping Discord session...")
		dg.Close()
		dg = nil
	}
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
		botService.Config = newConfig
		botService.Mu.Unlock()

		// Save to file
		botService.SaveConfig("config.json")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated successfully"))

		// Restart bot with new config
		go restartBot()
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	cfg := botService.Config
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DiscordBot 配置中心</title>
    <style>
        :root { --primary-color: #5865F2; --success-color: #28a745; --danger-color: #dc3545; --bg-color: #f4f7f6; }
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
        <div class="sidebar-header">DiscordBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">Discord API 配置</div>
                <div class="form-group">
                    <label>Bot Token</label>
                    <input type="password" id="bot_token" value="%s">
                </div>
            </div>
            <div class="card">
                <div class="section-title">连接配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>BotNexus 地址</label>
                        <input type="text" id="nexus_addr" value="%s">
                    </div>
                    <div class="form-group">
                        <label>Web UI 端口</label>
                        <input type="number" id="log_port" value="%d">
                    </div>
                </div>
            </div>
            <div style="text-align: center; margin-top: 30px;">
                <button class="btn btn-primary" style="padding: 15px 40px; font-size: 16px;" onclick="saveConfig()">保存配置并重启</button>
            </div>
        </div>
        <div id="logs-tab" style="display: none;">
            <div class="card">
                <div class="section-title">系统日志 <button class="btn btn-danger" onclick="clearLogs()">清空显示</button></div>
                <div id="logs" class="logs-container">正在加载日志...</div>
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
            const cfg = {
                bot_token: document.getElementById('bot_token').value,
                nexus_addr: document.getElementById('nexus_addr').value,
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
                alert('保存失败');
            }
        }
        async function loadLogs() {
            if (currentTab !== 'logs') return;
            try {
                const resp = await fetch('/logs?lines=200');
                const logs = await resp.json();
                const logsDiv = document.getElementById('logs');
                logsDiv.innerHTML = logs.map(line => `+"`"+`<div class="log-line">${line}</div>`+"`"+`).join('');
                logsDiv.scrollTop = logsDiv.scrollHeight;
            } catch (e) {}
            setTimeout(loadLogs, 2000);
        }
        function clearLogs() { document.getElementById('logs').innerText = ''; }
    </script>
</body>
</html>
`, cfg.BotToken, cfg.NexusAddr, cfg.LogPort)
}

func handleNexusCommand(data []byte) {
	var req struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return
	}

	log.Printf("Received action from Nexus: %s", req.Action)

	var resp map[string]any
	switch req.Action {
	case "send_msg", "send_group_msg", "send_private_msg":
		msg, _ := req.Params["message"].(string)
		targetID := ""
		if req.Action == "send_group_msg" {
			targetID, _ = req.Params["group_id"].(string)
		} else {
			targetID, _ = req.Params["user_id"].(string)
		}

		if targetID != "" && msg != "" {
			_, err := dg.ChannelMessageSend(targetID, msg)
			if err != nil {
				resp = map[string]any{"status": "failed", "retcode": 500, "msg": err.Error()}
			} else {
				resp = map[string]any{"status": "ok", "retcode": 0, "data": map[string]any{"message_id": "discord_" + time.Now().String()}}
			}
		}
	case "get_status":
		resp = map[string]interface{}{
			"status":  "ok",
			"retcode": 0,
			"data": map[string]interface{}{
				"online": true,
				"good":   true,
			},
		}
	case "get_login_info":
		resp = map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"user_id":  selfID,
				"nickname": dg.State.User.Username,
			},
		}
	}

	if resp != nil && req.Echo != "" {
		resp["echo"] = req.Echo
		botService.SendToNexus(resp)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Handle images
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 || attachment.Height > 0 {
			m.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		}
	}

	log.Printf("[%s] %s", m.Author.Username, m.Content)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  m.ID,
		"user_id":     m.Author.ID,
		"message":     m.Content,
		"raw_message": m.Content,
		"sender": map[string]any{
			"user_id":  m.Author.ID,
			"nickname": m.Author.Username,
		},
	}

	if m.GuildID != "" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = m.ChannelID
	} else {
		obMsg["message_type"] = "private"
	}

	botService.SendToNexus(obMsg)
}
