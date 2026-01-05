package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	"BotMatrix/common/bot"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jordan-wright/email"
)

// EmailConfig extends bot.BotConfig with Email specific fields
type EmailConfig struct {
	bot.BotConfig
	ImapServer   string `json:"imap_server"`
	ImapPort     int    `json:"imap_port"`
	Username     string `json:"username"` // Email address
	Password     string `json:"password"` // App Password
	SmtpServer   string `json:"smtp_server"`
	SmtpPort     int    `json:"smtp_port"`
	SmtpUsername string `json:"smtp_username"` // Usually same as Username
	SmtpPassword string `json:"smtp_password"` // Usually same as Password
	PollInterval int    `json:"poll_interval"` // Seconds
}

var (
	botService *bot.BaseBot
	emailCfg   EmailConfig
	botCtx     context.Context
	botCancel  context.CancelFunc
)

func main() {
	botService = bot.NewBaseBot(8086)
	log.SetOutput(botService.LogManager)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loadConfig()

	// Setup standard handlers using the new abstracted logic
	botService.SetupStandardHandlers("EmailBot", &emailCfg, restartBot, []bot.ConfigSection{
		{
			Title: "IMAP 配置 (接收邮件)",
			Fields: []bot.ConfigField{
				{Label: "IMAP 服务器", ID: "imap_server", Type: "text", Value: emailCfg.ImapServer},
				{Label: "IMAP 端口", ID: "imap_port", Type: "number", Value: emailCfg.ImapPort},
				{Label: "用户名 (邮箱)", ID: "username", Type: "text", Value: emailCfg.Username},
				{Label: "密码 (应用专用密码)", ID: "password", Type: "password", Value: emailCfg.Password},
				{Label: "轮询间隔 (秒)", ID: "poll_interval", Type: "number", Value: emailCfg.PollInterval},
			},
		},
		{
			Title: "SMTP 配置 (发送邮件)",
			Fields: []bot.ConfigField{
				{Label: "SMTP 服务器", ID: "smtp_server", Type: "text", Value: emailCfg.SmtpServer},
				{Label: "SMTP 端口", ID: "smtp_port", Type: "number", Value: emailCfg.SmtpPort},
				{Label: "SMTP 用户名", ID: "smtp_username", Type: "text", Value: emailCfg.SmtpUsername},
				{Label: "SMTP 密码", ID: "smtp_password", Type: "password", Value: emailCfg.SmtpPassword},
			},
		},
		{
			Title: "连接配置",
			Fields: []bot.ConfigField{
				{Label: "BotNexus 地址", ID: "nexus_addr", Type: "text", Value: emailCfg.NexusAddr},
				{Label: "Web UI 端口", ID: "log_port", Type: "number", Value: emailCfg.LogPort},
			},
		},
	})

	go botService.StartHTTPServer()

	restartBot()

	botService.WaitExitSignal()
	stopBot()
}

func loadConfig() {
	botService.LoadConfig("config.json")

	// Sync common config to local emailCfg
	botService.Mu.RLock()
	emailCfg.BotConfig = botService.Config
	botService.Mu.RUnlock()

	// Load Email specific fields
	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, &emailCfg)
	}

	// Environment variable overrides
	if envUser := os.Getenv("EMAIL_USERNAME"); envUser != "" {
		emailCfg.Username = envUser
	}
	if envPass := os.Getenv("EMAIL_PASSWORD"); envPass != "" {
		emailCfg.Password = envPass
	}

	if emailCfg.SmtpUsername == "" {
		emailCfg.SmtpUsername = emailCfg.Username
	}
	if emailCfg.SmtpPassword == "" {
		emailCfg.SmtpPassword = emailCfg.Password
	}
	if emailCfg.PollInterval == 0 {
		emailCfg.PollInterval = 60
	}

	// Set SelfID for the framework
	botService.Mu.Lock()
	botService.SelfID = emailCfg.Username
	botService.Mu.Unlock()
}

func restartBot() {
	stopBot()

	botService.Mu.RLock()
	// Sync botService.Config from emailCfg
	botService.Config = emailCfg.BotConfig
	nexusAddr := botService.Config.NexusAddr
	selfID := emailCfg.Username
	botService.Mu.RUnlock()

	botCtx, botCancel = context.WithCancel(context.Background())

	// Start Nexus connection using framework
	botService.StartNexusConnection(botCtx, nexusAddr, "Email", selfID, handleNexusCommand)

	// Start Email polling
	go pollEmails(botCtx)

	log.Println("Email Bot started")
}

func stopBot() {
	if botCancel != nil {
		botCancel()
		botCancel = nil
	}
	log.Println("Email Bot stopped")
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

func handleAction(action map[string]any) (any, error) {
	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]any)

	switch actionType {
	case "send_private_msg":
		userID, _ := params["user_id"].(string) // Email address
		message, _ := params["message"].(string)

		subject := "Message from BotMatrix"
		body := message

		lines := strings.SplitN(message, "\n", 2)
		if len(lines) > 0 && strings.HasPrefix(lines[0], "Subject:") {
			subject = strings.TrimPrefix(lines[0], "Subject:")
			subject = strings.TrimSpace(subject)
			if len(lines) > 1 {
				body = lines[1]
			} else {
				body = ""
			}
		}

		err := sendEmail(userID, subject, body)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return nil, fmt.Errorf("unsupported action: %s", actionType)
}

func sendEmail(to, subject, body string) error {
	botService.Mu.RLock()
	smtpServer := emailCfg.SmtpServer
	smtpPort := emailCfg.SmtpPort
	smtpUsername := emailCfg.SmtpUsername
	smtpPassword := emailCfg.SmtpPassword
	botService.Mu.RUnlock()

	e := email.NewEmail()
	e.From = fmt.Sprintf("BotMatrix <%s>", smtpUsername)
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	addr := fmt.Sprintf("%s:%d", smtpServer, smtpPort)
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)

	return e.Send(addr, auth)
}

func pollEmails(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			botService.Mu.RLock()
			imapServer := emailCfg.ImapServer
			imapPort := emailCfg.ImapPort
			username := emailCfg.Username
			password := emailCfg.Password
			pollInterval := emailCfg.PollInterval
			botService.Mu.RUnlock()

			if imapServer == "" || username == "" || password == "" {
				log.Println("IMAP configuration missing, waiting 10s...")
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}

			log.Println("Connecting to IMAP server...")
			// DialTLS usually needs a TLS config, but we can pass nil for default
			c, err := client.DialTLS(fmt.Sprintf("%s:%d", imapServer, imapPort), nil)
			if err != nil {
				log.Printf("IMAP connection failed: %v, retrying in 30s...", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
					continue
				}
			}
			log.Println("IMAP Connected")

			if err := c.Login(username, password); err != nil {
				log.Printf("IMAP Login failed: %v, retrying in 30s...", err)
				c.Close()
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
					continue
				}
			}

			// Polling loop for this connection
			for {
				select {
				case <-ctx.Done():
					c.Logout()
					return
				default:
					_, err := c.Select("INBOX", false)
					if err != nil {
						log.Printf("Select INBOX failed: %v", err)
						goto next_imap_conn
					}

					// Search for UNSEEN messages
					criteria := imap.NewSearchCriteria()
					criteria.WithoutFlags = []string{imap.SeenFlag}
					uids, err := c.Search(criteria)
					if err != nil {
						log.Printf("Search failed: %v", err)
						goto next_imap_conn
					}

					if len(uids) > 0 {
						log.Printf("Found %d new emails", len(uids))
						seqset := new(imap.SeqSet)
						seqset.AddNum(uids...)

						section := &imap.BodySectionName{}
						items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

						messages := make(chan *imap.Message, 10)
						done := make(chan error, 1)
						go func() {
							done <- c.Fetch(seqset, items, messages)
						}()

						for msg := range messages {
							processMessage(msg, section)
						}

						if err := <-done; err != nil {
							log.Printf("Fetch failed: %v", err)
						}
					}

					select {
					case <-ctx.Done():
						c.Logout()
						return
					case <-time.After(time.Duration(pollInterval) * time.Second):
						continue
					}
				}
			}
		next_imap_conn:
			c.Logout()
		}
	}
}

func processMessage(msg *imap.Message, section *imap.BodySectionName) {
	r := msg.GetBody(section)
	if r == nil {
		return
	}

	// Read body
	buf := new(strings.Builder)
	_, err := fmt.Fprint(buf, r) // Simple read
	if err != nil {
		log.Printf("Failed to read body: %v", err)
	}
	bodyStr := buf.String()

	// Construct OneBot Message
	sender := msg.Envelope.From[0]
	senderEmail := fmt.Sprintf("%s@%s", sender.MailboxName, sender.HostName)
	senderName := sender.PersonalName

	content := fmt.Sprintf("Subject: %s\n\n%s", msg.Envelope.Subject, bodyStr)

	log.Printf("Received email from %s: %s", senderEmail, msg.Envelope.Subject)

	botService.Mu.RLock()
	selfID := emailCfg.Username
	botService.Mu.RUnlock()

	event := map[string]any{
		"post_type":    "message",
		"message_type": "private", // Treat all emails as private messages
		"time":         msg.Envelope.Date.Unix(),
		"self_id":      selfID,
		"sub_type":     "friend",
		"message_id":   fmt.Sprintf("%d", msg.SeqNum),
		"user_id":      senderEmail,
		"message":      content,
		"raw_message":  content,
		"sender": map[string]any{
			"user_id":  senderEmail,
			"nickname": senderName,
		},
	}

	botService.SendToNexus(event)
}
