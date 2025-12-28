package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	common "BotMatrix/src/Common"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackConfig extends common.BotConfig with Slack specific fields
type SlackConfig struct {
	common.BotConfig
	AppToken string `json:"app_token"` // xapp-...
}

var (
	botService *common.BaseBot
	api        *slack.Client
	client     *socketmode.Client
	selfID     string
	botCtx     context.Context
	botCancel  context.CancelFunc
	slackCfg   SlackConfig
)

func main() {
	botService = common.NewBaseBot(8086)
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

	// Sync common config to local slackCfg
	botService.Mu.RLock()
	slackCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Slack specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &slackCfg)
	}

	// Environment variable overrides
	if envAppToken := os.Getenv("SLACK_APP_TOKEN"); envAppToken != "" {
		slackCfg.AppToken = envAppToken
	}
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	botToken := slackCfg.BotToken
	appToken := slackCfg.AppToken
	nexusAddr := botService.Config.NexusAddr
	botService.Mu.RUnlock()

	if botToken == "" || appToken == "" {
		log.Println("WARNING: Slack BotToken or AppToken is not configured. Bot will not start until configured.")
		return
	}

	botCtx, botCancel = context.WithCancel(context.Background())

	// Initialize Slack Client
	api = slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	client = socketmode.New(api)

	// Get Bot Info
	authTest, err := api.AuthTest()
	if err != nil {
		log.Printf("Slack Auth failed: %v", err)
		return
	}
	selfID = authTest.BotID
	log.Printf("Slack Bot Authorized: %s (ID: %s, User: %s)", authTest.User, selfID, authTest.UserID)

	// Connect to BotNexus
	botService.StartNexusConnection(botCtx, nexusAddr, "Slack", selfID, handleNexusCommand)

	// Start Socket Mode
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-client.Events:
				switch evt.Type {
				case socketmode.EventTypeConnecting:
					log.Println("Connecting to Slack Socket Mode...")
				case socketmode.EventTypeConnectionError:
					log.Println("Connection failed. Retrying later...")
				case socketmode.EventTypeConnected:
					log.Println("Connected to Slack Socket Mode!")
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
					if !ok {
						continue
					}
					client.Ack(*evt.Request)

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innerEvent := eventsAPIEvent.InnerEvent
						switch ev := innerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							handleMessage(ev)
						}
					}
				}
			}
		}
	}(botCtx)

	go func() {
		if err := client.Run(); err != nil {
			log.Printf("Socket Mode failed: %v", err)
		}
	}()
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
}

func handleMessage(ev *slackevents.MessageEvent) {
	if ev.BotID != "" && ev.BotID == selfID {
		return
	}

	log.Printf("[%s] %s", ev.User, ev.Text)

	obMsg := map[string]any{
		"post_type":   "message",
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"sub_type":    "normal",
		"message_id":  ev.ClientMsgID,
		"user_id":     ev.User,
		"message":     ev.Text,
		"raw_message": ev.Text,
		"sender": map[string]any{
			"user_id":  ev.User,
			"nickname": "SlackUser",
		},
	}

	if strings.HasPrefix(ev.Channel, "D") {
		obMsg["message_type"] = "private"
	} else {
		obMsg["message_type"] = "group"
		obMsg["group_id"] = ev.Channel
	}

	botService.SendToNexus(obMsg)
}

func handleNexusCommand(data []byte) {
	var cmd struct {
		Action string                 `json:"action"`
		Params map[string]any         `json:"params"`
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
			sendSlackMessage(channelID, text, cmd.Echo)
		}
	case "send_private_msg":
		userID, _ := cmd.Params["user_id"].(string)
		text, _ := cmd.Params["message"].(string)
		if userID != "" && text != "" {
			sendSlackMessage(userID, text, cmd.Echo)
		}
	case "delete_msg":
		msgID, _ := cmd.Params["message_id"].(string)
		if msgID != "" {
			deleteSlackMessage(msgID, cmd.Echo)
		}
	case "get_login_info":
		botService.SendToNexus(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"user_id":  selfID,
				"nickname": "SlackBot",
			},
			"echo": cmd.Echo,
		})
	}
}

func sendSlackMessage(channelID, text, echo string) {
	if api == nil {
		return
	}
	_, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
	)

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Sent message to %s", channelID)
	compositeID := fmt.Sprintf("%s:%s", channelID, timestamp)
	botService.SendToNexus(map[string]any{
		"status": "ok",
		"data":   map[string]any{"message_id": compositeID},
		"echo":   echo,
	})
}

func deleteSlackMessage(compositeID, echo string) {
	if api == nil {
		return
	}
	parts := strings.Split(compositeID, ":")
	if len(parts) != 2 {
		botService.SendToNexus(map[string]any{"status": "failed", "message": "invalid message_id format", "echo": echo})
		return
	}
	channelID := parts[0]
	timestamp := parts[1]

	_, _, err := api.DeleteMessage(channelID, timestamp)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		botService.SendToNexus(map[string]any{"status": "failed", "message": err.Error(), "echo": echo})
		return
	}

	log.Printf("Deleted message %s in channel %s", timestamp, channelID)
	botService.SendToNexus(map[string]any{"status": "ok", "echo": echo})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slackCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig SlackConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		botService.Mu.Lock()
		slackCfg = newConfig
		botService.Config = newConfig.BotConfig
		botService.Mu.Unlock()

		// Save to file
		data, _ := json.MarshalIndent(slackCfg, "", "  ")
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
	cfg := slackCfg
	botService.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SlackBot 配置中心</title>
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
        <div class="sidebar-header">SlackBot</div>
        <div class="nav-item active" onclick="switchTab('config')">核心配置</div>
        <div class="nav-item" onclick="switchTab('logs')">实时日志</div>
    </div>
    <div class="main-content">
        <div id="config-tab">
            <div class="card">
                <div class="section-title">Slack API 配置</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                    <div class="form-group">
                        <label>Bot Token (xoxb-...)</label>
                        <input type="password" id="bot_token" value="%s">
                    </div>
                    <div class="form-group">
                        <label>App Token (xapp-...)</label>
                        <input type="password" id="app_token" value="%s">
                    </div>
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
                app_token: document.getElementById('app_token').value,
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
`, cfg.BotToken, cfg.AppToken, cfg.NexusAddr, cfg.LogPort)
}
