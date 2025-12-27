package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	common "BotMatrix/src/Common"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jordan-wright/email"
)

// EmailConfig extends common.BotConfig with Email specific fields
type EmailConfig struct {
	common.BotConfig
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
	botService *common.BaseBot
	emailCfg   EmailConfig
	botCtx     context.Context
	botCancel  context.CancelFunc
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

func handleAction(action map[string]interface{}) (interface{}, error) {
	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]interface{})

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
				case <-time.After(30 * time.After(30*time.Second)):
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

	event := map[string]interface{}{
		"post_type":    "message",
		"message_type": "private", // Treat all emails as private messages
		"time":         msg.Envelope.Date.Unix(),
		"self_id":      selfID,
		"sub_type":     "friend",
		"message_id":   fmt.Sprintf("%d", msg.SeqNum),
		"user_id":      senderEmail,
		"message":      content,
		"raw_message":  content,
		"sender": map[string]interface{}{
			"user_id":  senderEmail,
			"nickname": senderName,
		},
	}

	botService.SendToNexus(event)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		botService.Mu.RLock()
		defer botService.Mu.RUnlock()
		json.NewEncoder(w).Encode(emailCfg)
		return
	}

	if r.Method == http.MethodPost {
		var newCfg EmailConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update file
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

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleConfigUI(w http.ResponseWriter, r *http.Request) {
	botService.Mu.RLock()
	defer botService.Mu.RUnlock()

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Email Bot Config</title>
    <style>
        body { font-family: sans-serif; margin: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="text"], input[type="number"], input[type="password"] { width: 100%; padding: 8px; box-sizing: border-box; }
        button { padding: 10px 15px; background-color: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background-color: #0056b3; }
        pre { background: #f4f4f4; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Email Bot Configuration</h1>
    <form id="configForm">
        <div class="form-group">
            <label>Nexus Address:</label>
            <input type="text" name="nexus_addr" value="{{.NexusAddr}}">
        </div>
        <div class="form-group">
            <label>IMAP Server:</label>
            <input type="text" name="imap_server" value="{{.ImapServer}}">
        </div>
        <div class="form-group">
            <label>IMAP Port:</label>
            <input type="number" name="imap_port" value="{{.ImapPort}}">
        </div>
        <div class="form-group">
            <label>Username (Email):</label>
            <input type="text" name="username" value="{{.Username}}">
        </div>
        <div class="form-group">
            <label>Password (App Password):</label>
            <input type="password" name="password" value="{{.Password}}">
        </div>
        <div class="form-group">
            <label>SMTP Server:</label>
            <input type="text" name="smtp_server" value="{{.SmtpServer}}">
        </div>
        <div class="form-group">
            <label>SMTP Port:</label>
            <input type="number" name="smtp_port" value="{{.SmtpPort}}">
        </div>
        <div class="form-group">
            <label>SMTP Username:</label>
            <input type="text" name="smtp_username" value="{{.SmtpUsername}}">
        </div>
        <div class="form-group">
            <label>SMTP Password:</label>
            <input type="password" name="smtp_password" value="{{.SmtpPassword}}">
        </div>
        <div class="form-group">
            <label>Poll Interval (seconds):</label>
            <input type="number" name="poll_interval" value="{{.PollInterval}}">
        </div>
        <button type="submit">Save & Restart</button>
    </form>

    <h2>Logs</h2>
    <pre id="logs">Loading logs...</pre>

    <script>
        document.getElementById('configForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const config = {};
            formData.forEach((value, key) => {
                if (key.includes('port') || key.includes('interval')) {
                    config[key] = parseInt(value);
                } else {
                    config[key] = value;
                }
            });

            const resp = await fetch('/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            alert(await resp.text());
        };

        async function updateLogs() {
            try {
                const resp = await fetch('/logs?lines=50');
                const text = await resp.text();
                document.getElementById('logs').textContent = text;
            } catch (e) {}
        }
        setInterval(updateLogs, 2000);
        updateLogs();
    </script>
</body>
</html>
`
	t := template.Must(template.New("config").Parse(tmpl))
	t.Execute(w, emailCfg)
}
