package main

import (
	"BotMatrix/common/bot"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/silenceper/wechat/v2/work"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	workMessage "github.com/silenceper/wechat/v2/work/message"
)

// WeComConfig extends bot.BotConfig with WeCom specific fields
type WeComConfig struct {
	bot.BotConfig
	CorpID         string `json:"corp_id"`
	AgentID        string `json:"agent_id"`
	Secret         string `json:"secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	ListenPort     int    `json:"listen_port"` // Callback listen port
}

var (
	wecomCfg   WeComConfig
	botService *bot.BaseBot
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

func handleAction(action map[string]any) (any, error) {
	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]any)

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

		req := workMessage.SendTextRequest{
			SendRequestCommon: &workMessage.SendRequestCommon{
				ToUser:  userID,
				MsgType: "text",
				AgentID: agentID,
			},
			Text: workMessage.TextField{
				Content: msgContent,
			},
		}

		msgManager := currentWC.GetMessage()
		msgID, err := msgManager.SendText(req)
		if err != nil {
			return nil, err
		}
		return map[string]any{"message_id": msgID}, nil

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
		return map[string]any{
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

	var result map[string]any
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
	// Sync botService.Config from wecomCfg
	botService.Config = wecomCfg.BotConfig
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
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
		Echo   string         `json:"echo"`
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

func handleCallback(c *gin.Context) {
	mu.RLock()
	// currentWC is not needed for callback handling when using officialaccount shim
	mu.RUnlock()

	// We use officialaccount server for WeCom callback because they are compatible in XML/Encryption
	offCfg := &offConfig.Config{
		AppID:          wecomCfg.CorpID,
		AppSecret:      wecomCfg.Secret,
		Token:          wecomCfg.Token,
		EncodingAESKey: wecomCfg.EncodingAESKey,
		Cache:          cache.NewMemory(),
	}
	off := wechat.NewWechat().GetOfficialAccount(offCfg)
	server := off.GetServer(c.Request, c.Writer)

	// Set message handler
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		log.Printf("Received WeCom message: %s from %s", msg.Content, msg.FromUserName)

		// Broadcast to Nexus
		botService.SendToNexus(map[string]any{
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
	botService = bot.NewBaseBot(8083)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("WeComBot", &wecomCfg, restartBot, []bot.ConfigSection{
		{
			Title: "企业微信 API 配置",
			Fields: []bot.ConfigField{
				{Label: "Corp ID", ID: "corp_id", Type: "text", Value: wecomCfg.CorpID},
				{Label: "Agent ID", ID: "agent_id", Type: "text", Value: wecomCfg.AgentID},
				{Label: "Secret", ID: "secret", Type: "password", Value: wecomCfg.Secret},
				{Label: "Token", ID: "token", Type: "text", Value: wecomCfg.Token},
				{Label: "Encoding AES Key", ID: "encoding_aes_key", Type: "text", Value: wecomCfg.EncodingAESKey},
				{Label: "回调监听端口", ID: "listen_port", Type: "number", Value: wecomCfg.ListenPort},
			},
		},
		{
			Title: "连接与服务配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: wecomCfg.NexusAddr},
				{Label: "Web UI 端口 (LogPort)", ID: "log_port", Type: "number", Value: wecomCfg.LogPort},
			},
		},
	})

	go botService.StartHTTPServer()

	restartBot()

	botService.WaitExitSignal()
	stopBot()
}
