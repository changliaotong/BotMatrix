package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	common "BotMatrix/src/Common"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
)

// TencentConfig extends common.BotConfig with Tencent specific fields
type TencentConfig struct {
	common.BotConfig
	AppID      uint64 `json:"app_id"`
	Token      string `json:"token"`
	Secret     string `json:"secret"`
	Sandbox    bool   `json:"sandbox"`
	FileHost   string `json:"file_host"`
	MediaRoute string `json:"media_route"`
}

// SessionCache to store last message ID for replying
type SessionCache struct {
	sync.RWMutex     `json:"-"`
	UserLastMsgID    map[string]string                   `json:"user_last_msg_id"`
	GroupLastMsgID   map[string]string                   `json:"group_last_msg_id"`
	ChannelLastMsgID map[string]string                   `json:"channel_last_msg_id"`
	LastMsgTime      map[string]int64                    `json:"last_msg_time"`
	PendingActions   map[string][]map[string]any `json:"pending_actions"`
}

var (
	botService     *common.BaseBot
	tencentCfg     TencentConfig
	api            openapi.OpenAPI
	botCtx         context.Context
	botCancel      context.CancelFunc
	msgSeq         int64
	accessToken    string
	tokenExpiresAt int64
	tokenMu        sync.Mutex
	sessionCache   = &SessionCache{
		UserLastMsgID:    make(map[string]string),
		GroupLastMsgID:   make(map[string]string),
		ChannelLastMsgID: make(map[string]string),
		LastMsgTime:      make(map[string]int64),
		PendingActions:   make(map[string][]map[string]any),
	}
)

func main() {
	botService = common.NewBaseBot(3133)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sessionCache.LoadDisk()
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

	// Sync common config to local tencentCfg
	botService.Mu.RLock()
	tencentCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Tencent specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &tencentCfg)
	}

	// Environment variable overrides
	if envAppID := os.Getenv("TENCENT_APP_ID"); envAppID != "" {
		fmt.Sscanf(envAppID, "%d", &tencentCfg.AppID)
	}
	if envToken := os.Getenv("TENCENT_TOKEN"); envToken != "" {
		tencentCfg.Token = envToken
	}
	if envSecret := os.Getenv("TENCENT_SECRET"); envSecret != "" {
		tencentCfg.Secret = envSecret
	}

	// Defaults
	if tencentCfg.MediaRoute == "" {
		tencentCfg.MediaRoute = "/media/"
	}
}

func restartBot() {
	stopBot()

	botCtx, botCancel = context.WithCancel(context.Background())

	// Initialize Tencent API
	botToken := token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     fmt.Sprintf("%d", tencentCfg.AppID),
			AppSecret: tencentCfg.Secret,
		},
	)

	if tencentCfg.Sandbox {
		api = botgo.NewSandboxOpenAPI(fmt.Sprintf("%d", tencentCfg.AppID), botToken).WithTimeout(3 * time.Second)
	} else {
		api = botgo.NewOpenAPI(fmt.Sprintf("%d", tencentCfg.AppID), botToken).WithTimeout(3 * time.Second)
	}

	// Get Bot Info
	go func() {
		for {
			me, err := api.Me(botCtx)
			if err != nil {
				log.Printf("Error getting bot info: %v. Retrying in 5s...", err)
				select {
				case <-botCtx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}
			botService.Mu.Lock()
			botService.SelfID = me.ID
			botService.Mu.Unlock()
			log.Printf("Bot Identity: %s (%s)", me.Username, me.ID)

			// Start Nexus connection
			botService.Mu.RLock()
			nexusAddr := botService.Config.NexusAddr
			botService.Mu.RUnlock()
			botService.StartNexusConnection(botCtx, nexusAddr, "Tencent", me.ID, handleNexusCommand)
			break
		}
	}()

	// Start Tencent WebSocket
	go startTencentWS(botCtx, botToken)

	log.Println("Tencent Bot started")
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
	log.Println("Tencent Bot stopped")
}

func startTencentWS(ctx context.Context, botToken token.TokenSource) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			wsInfo, err := api.WS(ctx, nil, "")
			if err != nil {
				log.Printf("Error getting WS info: %v. Retrying in 5s...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			intent := event.RegisterHandlers(
				event.ATMessageEventHandler(atMessageEventHandler),
				event.DirectMessageEventHandler(directMessageEventHandler),
				event.GroupATMessageEventHandler(groupATMessageEventHandler),
				event.C2CMessageEventHandler(c2cMessageEventHandler),
			)
			// Add mandatory intents
			intent = intent | (1 << 30) | (1 << 12) | (1 << 0) | (1 << 1) | (1 << 10) | (1 << 25)

			if err := botgo.NewSessionManager().Start(wsInfo, botToken, &intent); err != nil {
				log.Printf("Session Manager ended: %v. Retrying in 5s...", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// SessionCache methods
func (s *SessionCache) SaveDisk() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile("session_cache.json", data, 0644)
}

func (s *SessionCache) LoadDisk() {
	data, err := os.ReadFile("session_cache.json")
	if err != nil {
		return
	}
	s.Lock()
	defer s.Unlock()
	json.Unmarshal(data, s)
}

func (s *SessionCache) AddPending(keyType, key string, action map[string]any) {
	s.Lock()
	defer s.Unlock()
	compositeKey := keyType + ":" + key
	if s.PendingActions == nil {
		s.PendingActions = make(map[string][]map[string]any)
	}
	s.PendingActions[compositeKey] = append(s.PendingActions[compositeKey], action)
	go s.SaveDisk()
}

func (s *SessionCache) Save(keyType, key, msgID string) []map[string]any {
	s.Lock()
	defer s.Unlock()
	switch keyType {
	case "user":
		s.UserLastMsgID[key] = msgID
	case "group":
		s.GroupLastMsgID[key] = msgID
	case "channel":
		s.ChannelLastMsgID[key] = msgID
	}
	s.LastMsgTime[msgID] = time.Now().Unix()

	compositeKey := keyType + ":" + key
	var pending []map[string]any
	if actions, ok := s.PendingActions[compositeKey]; ok && len(actions) > 0 {
		pending = actions
		delete(s.PendingActions, compositeKey)
	}
	go s.SaveDisk()
	return pending
}

func (s *SessionCache) Get(keyType, key string) string {
	s.RLock()
	defer s.RUnlock()

	var msgID string
	switch keyType {
	case "user":
		msgID = s.UserLastMsgID[key]
	case "group":
		msgID = s.GroupLastMsgID[key]
	case "channel":
		msgID = s.ChannelLastMsgID[key]
	}

	if msgID == "" {
		return ""
	}

	if ts, ok := s.LastMsgTime[msgID]; ok {
		if time.Now().Unix()-ts > 290 {
			return ""
		}
		return msgID
	}
	return ""
}

// Event Handlers
func atMessageEventHandler(event *dto.WSPayload, data *dto.WSATMessageData) error {
	botService.Mu.RLock()
	selfID := botService.SelfID
	botService.Mu.RUnlock()

	content := data.Content
	// Remove bot mention
	re := regexp.MustCompile(`<@![0-9]+>`)
	content = re.ReplaceAllString(content, "")
	content = strings.TrimSpace(content)

	eventData := map[string]any{
		"post_type":    "message",
		"message_type": "guild",
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "channel",
		"message_id":   data.ID,
		"guild_id":     data.GuildID,
		"channel_id":   data.ChannelID,
		"user_id":      data.Author.ID,
		"message":      content,
		"raw_message":  data.Content,
		"sender": map[string]any{
			"user_id":  data.Author.ID,
			"nickname": data.Author.Username,
		},
	}

	botService.SendToNexus(eventData)

	// Save session and handle pending
	pending := sessionCache.Save("channel", data.ChannelID, data.ID)
	for _, action := range pending {
		handleAction(action)
	}

	return nil
}

func directMessageEventHandler(event *dto.WSPayload, data *dto.WSDirectMessageData) error {
	// Not fully implemented in SDK yet
	return nil
}

func groupATMessageEventHandler(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	botService.Mu.RLock()
	selfID := botService.SelfID
	botService.Mu.RUnlock()

	content := strings.TrimSpace(data.Content)

	eventData := map[string]any{
		"post_type":    "message",
		"message_type": "group",
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "normal",
		"message_id":   data.ID,
		"group_id":     data.GroupID,
		"user_id":      data.Author.MemberOpenID,
		"message":      content,
		"raw_message":  data.Content,
		"sender": map[string]any{
			"user_id": data.Author.MemberOpenID,
		},
	}

	botService.SendToNexus(eventData)

	pending := sessionCache.Save("group", data.GroupID, data.ID)
	for _, action := range pending {
		handleAction(action)
	}

	return nil
}

func c2cMessageEventHandler(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
	botService.Mu.RLock()
	selfID := botService.SelfID
	botService.Mu.RUnlock()

	content := strings.TrimSpace(data.Content)

	eventData := map[string]any{
		"post_type":    "message",
		"message_type": "private",
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"sub_type":     "friend",
		"message_id":   data.ID,
		"user_id":      data.Author.UserOpenID,
		"message":      content,
		"raw_message":  data.Content,
		"sender": map[string]any{
			"user_id": data.Author.UserOpenID,
		},
	}

	botService.SendToNexus(eventData)

	pending := sessionCache.Save("user", data.Author.UserOpenID, data.ID)
	for _, action := range pending {
		handleAction(action)
	}

	return nil
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

	log.Printf("Received Action: %s", cmd.Action)
	result, err := handleAction(map[string]any{
		"action": cmd.Action,
		"params": cmd.Params,
	})

	resp := map[string]any{
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

// Action Handling
func handleAction(action map[string]any) (any, error) {
	act, _ := action["action"].(string)
	params, _ := action["params"].(map[string]any)

	switch act {
	case "send_msg", "send_group_msg", "send_private_msg":
		messageType, _ := params["message_type"].(string)
		content, _ := params["message"].(string)
		if act == "send_group_msg" {
			messageType = "group"
		} else if act == "send_private_msg" {
			messageType = "private"
		}

		if messageType == "private" {
			userID := getString(params, "user_id")
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("user", userID)
				if msgID == "" {
					sessionCache.AddPending("user", userID, action)
					return nil, nil
				}
			}
			safeContent, imagePath, fileType := cleanContent(content)
			var media *dto.MediaInfo
			if imagePath != "" {
				fileInfo, err := uploadC2CFile(userID, imagePath, fileType)
				if err == nil {
					media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
				}
				if !strings.HasPrefix(imagePath, "http") {
					os.Remove(imagePath)
				}
			}
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
				Media:   media,
			}
			if media != nil {
				msgData.MsgType = 7
			}
			return api.PostC2CMessage(botCtx, userID, msgData)

		} else if messageType == "group" {
			groupID := getString(params, "group_id")
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("group", groupID)
				if msgID == "" {
					sessionCache.AddPending("group", groupID, action)
					return nil, nil
				}
			}
			safeContent, imagePath, fileType := cleanContent(content)
			var media *dto.MediaInfo
			if imagePath != "" {
				fileInfo, err := uploadGroupFile(groupID, imagePath, fileType)
				if err == nil {
					media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
				}
				if !strings.HasPrefix(imagePath, "http") {
					os.Remove(imagePath)
				}
			}
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
				Media:   media,
			}
			if media != nil {
				msgData.MsgType = 7
			}
			return api.PostGroupMessage(botCtx, groupID, msgData)

		} else {
			// Guild
			channelID := getString(params, "channel_id")
			if channelID == "" {
				channelID = getString(params, "group_id")
			}
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("channel", channelID)
				if msgID == "" {
					sessionCache.AddPending("channel", channelID, action)
					return nil, nil
				}
			}
			safeContent, _, _ := cleanContent(content)
			return api.PostMessage(botCtx, channelID, &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			})
		}

	case "get_login_info":
		me, err := api.Me(botCtx)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"user_id":  me.ID,
			"nickname": me.Username,
		}, nil
	}

	return nil, fmt.Errorf("unsupported action: %s", act)
}

// Helpers
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%.0f", f)
		}
	}
	return ""
}

func cleanContent(content string) (string, string, int) {
	// Simple CQ code parser
	reImg := regexp.MustCompile(`\[CQ:image,file=([^,\]]+)\]`)
	if matches := reImg.FindStringSubmatch(content); len(matches) > 1 {
		fileVal := matches[1]
		cleanMsg := reImg.ReplaceAllString(content, "")
		if strings.HasPrefix(fileVal, "base64://") {
			data, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(fileVal, "base64://"))
			tmpFile, _ := os.CreateTemp("", "tencent_media_*.png")
			tmpFile.Write(data)
			tmpFile.Close()
			return strings.TrimSpace(cleanMsg), tmpFile.Name(), 1
		}
		return strings.TrimSpace(cleanMsg), fileVal, 1
	}
	return content, "", 0
}

func uploadGroupFile(groupID string, filePath string, fileType int) (string, error) {
	// Implementation simplified for now
	return "", fmt.Errorf("uploadGroupFile not fully implemented")
}

func uploadC2CFile(userID string, filePath string, fileType int) (string, error) {
	// Implementation simplified for now
	return "", fmt.Errorf("uploadC2CFile not fully implemented")
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		json.NewEncoder(w).Encode(tencentCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newCfg TencentConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, _ := json.MarshalIndent(newCfg, "", "  ")
		if err := os.WriteFile("config.json", data, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		loadConfig()
		restartBot()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Config updated and bot restarted"))
		return
	}
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	defer botService.Mu.RUnlock()

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Tencent Bot Config</title>
    <style>
        body { font-family: sans-serif; margin: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="text"], input[type="number"], input[type="password"] { width: 100%; padding: 8px; box-sizing: border-box; }
        button { padding: 10px 15px; background-color: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background-color: #0056b3; }
    </style>
</head>
<body>
    <h1>Tencent Bot Configuration</h1>
    <form id="configForm">
        <div class="form-group">
            <label>Nexus Address:</label>
            <input type="text" name="nexus_addr" value="{{.NexusAddr}}">
        </div>
        <div class="form-group">
            <label>App ID:</label>
            <input type="number" name="app_id" value="{{.AppID}}">
        </div>
        <div class="form-group">
            <label>Token:</label>
            <input type="password" name="token" value="{{.Token}}">
        </div>
        <div class="form-group">
            <label>Secret:</label>
            <input type="password" name="secret" value="{{.Secret}}">
        </div>
        <div class="form-group">
            <label>Sandbox Mode:</label>
            <input type="checkbox" name="sandbox" {{if .Sandbox}}checked{{end}}>
        </div>
        <button type="submit">Save & Restart</button>
    </form>
    <script>
        document.getElementById('configForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const config = { sandbox: e.target.sandbox.checked };
            formData.forEach((value, key) => {
                if (key === 'app_id') config[key] = parseInt(value);
                else if (key !== 'sandbox') config[key] = value;
            });
            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            alert(await resp.text());
        };
    </script>
</body>
</html>
`
	t := template.Must(template.New("config").Parse(tmpl))
	t.Execute(w, tencentCfg)
}
