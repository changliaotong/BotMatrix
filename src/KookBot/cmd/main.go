package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	common "BotMatrix/src/Common"

	"github.com/lonelyevil/kook"
)

// KookConfig extends common.BotConfig with Kook specific fields
type KookConfig struct {
	common.BotConfig
}

var (
	botService *common.BaseBot
	session    *kook.Session
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
	kookCfg    KookConfig
)

// --- Simple Logger Implementation ---

type ConsoleLogger struct{}

func (l *ConsoleLogger) Trace() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Debug() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Info() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Warn() kook.Entry  { return &ConsoleEntry{} }
func (l *ConsoleLogger) Error() kook.Entry { return &ConsoleEntry{} }
func (l *ConsoleLogger) Fatal() kook.Entry { return &ConsoleEntry{} }

type ConsoleEntry struct{}

func (e *ConsoleEntry) Bool(key string, b bool) kook.Entry             { return e }
func (e *ConsoleEntry) Bytes(key string, val []byte) kook.Entry        { return e }
func (e *ConsoleEntry) Caller(depth int) kook.Entry                    { return e }
func (e *ConsoleEntry) Dur(key string, d time.Duration) kook.Entry     { return e }
func (e *ConsoleEntry) Err(key string, err error) kook.Entry           { return e }
func (e *ConsoleEntry) Float64(key string, f float64) kook.Entry       { return e }
func (e *ConsoleEntry) IPAddr(key string, ip net.IP) kook.Entry        { return e }
func (e *ConsoleEntry) Int(key string, i int) kook.Entry               { return e }
func (e *ConsoleEntry) Int64(key string, i int64) kook.Entry           { return e }
func (e *ConsoleEntry) Interface(key string, i interface{}) kook.Entry { return e }
func (e *ConsoleEntry) Msg(msg string)                                 { botService.LogManager.Info().Msg(msg) }
func (e *ConsoleEntry) Msgf(f string, i ...interface{})                { botService.LogManager.Info().Msgf(f, i...) }
func (e *ConsoleEntry) Str(key string, s string) kook.Entry            { return e }
func (e *ConsoleEntry) Strs(key string, s []string) kook.Entry         { return e }
func (e *ConsoleEntry) Time(key string, t time.Time) kook.Entry        { return e }

// ------------------------------------

func main() {
	botService = common.NewBaseBot(3136)
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

func loadConfig() {
	botService.LoadConfig("config.json")

	// Sync common config to local kookCfg
	botService.Mu.RLock()
	kookCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Kook specific fields (if any in the future)
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &kookCfg)
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	token := botService.Config.BotToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if token == "" {
		botService.LogManager.Warn().Msg("KOOK Bot Token is not configured. Bot will not start until configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	session = kook.New(token, &ConsoleLogger{})

	// Register Handlers
	session.AddHandler(textMessageHandler)
	session.AddHandler(imageMessageHandler)
	session.AddHandler(kmarkdownMessageHandler)

	// Open connection
	err := session.Open()
	if err != nil {
		log.Printf("Error opening connection: %v", err)
		return
	}

	// Get Self Info
	me, err := session.UserMe()
	if err == nil {
		selfID = me.ID
		log.Printf("Bot logged in as %s (ID: %s)", me.Username, selfID)
	}

	// Connect to BotNexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Kook", selfID, handleNexusCommand)
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}

	if session != nil {
		session.Close()
		session = nil
	}
}

func handleCommon(commonData *kook.EventDataGeneral, author kook.User) {
	if author.Bot && author.ID == selfID {
		return
	}

	botService.LogManager.Info().Str("username", author.Username).Msg(commonData.Content)

	obMsg := map[string]interface{}{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  commonData.MsgID,
		"user_id":     commonData.AuthorID,
		"message":     commonData.Content,
		"raw_message": commonData.Content,
		"sender": map[string]interface{}{
			"user_id":  commonData.AuthorID,
			"nickname": author.Username,
		},
	}

	if commonData.Type == kook.MessageTypeImage {
		obMsg["message"] = fmt.Sprintf("[CQ:image,file=%s]", commonData.Content)
		obMsg["raw_message"] = obMsg["message"]
	}

	if commonData.ChannelType == "GROUP" {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = commonData.TargetID
	} else {
		obMsg["message_type"] = "private"
	}

	botService.SendToNexus(obMsg)
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

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
		Echo   string                 `json:"echo"`
	}
	if err := json.Unmarshal(data, &cmd); err != nil {
		return
	}

	botService.LogManager.Info().Str("action", cmd.Action).Msg("Received Command")

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
		botService.SendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
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
		botService.LogManager.Error().Str("message_id", msgID).Err(err).Msg("Failed to delete message")
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.LogManager.Info().Str("message_id", msgID).Msg("Deleted message")
	botService.SendToNexus(map[string]interface{}{"status": "ok", "echo": echo})
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
		botService.LogManager.Error().Str("target_id", targetID).Str("content", content).Err(err).Msg("Failed to send message")
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.LogManager.Info().Str("target_id", targetID).Str("content", content).Msg("Sent message")
	botService.SendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": resp.MsgID},
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
		botService.LogManager.Error().Str("target_id", targetID).Str("content", content).Err(err).Msg("Failed to send private message")
		botService.SendToNexus(map[string]interface{}{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	botService.LogManager.Info().Str("target_id", targetID).Str("content", content).Msg("Sent private message")
	botService.SendToNexus(map[string]interface{}{
		"status": "ok",
		"data":   map[string]interface{}{"message_id": resp.MsgID},
		"echo":   echo,
	})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kookCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig KookConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		kookCfg = newConfig
		botService.Config = newConfig.BotConfig
		botService.Mu.Unlock()

		// Save to file
		data, _ := json.MarshalIndent(kookCfg, "", "  ")
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
	botService.Mu.RLock()
	cfg := kookCfg
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KookBot 配置中心</title>
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
        <div class="sidebar-header">KookBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">Kook API 配置</div>
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
