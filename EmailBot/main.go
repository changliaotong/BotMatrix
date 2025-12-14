package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gorilla/websocket"
	"github.com/jordan-wright/email"
)

// Config holds the bot configuration
type Config struct {
	ImapServer   string `json:"imap_server"`
	ImapPort     int    `json:"imap_port"`
	Username     string `json:"username"` // Email address
	Password     string `json:"password"` // App Password
	SmtpServer   string `json:"smtp_server"`
	SmtpPort     int    `json:"smtp_port"`
	SmtpUsername string `json:"smtp_username"` // Usually same as Username
	SmtpPassword string `json:"smtp_password"` // Usually same as Password
	PollInterval int    `json:"poll_interval"` // Seconds
	NexusAddr    string `json:"nexus_addr"`
}

var (
	config Config
	conn   *websocket.Conn
	selfID string
)

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config.json: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config.json: %v", err)
	}
	if config.SmtpUsername == "" {
		config.SmtpUsername = config.Username
	}
	if config.SmtpPassword == "" {
		config.SmtpPassword = config.Password
	}
	selfID = config.Username
}

func connectToNexus() {
	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		c, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, nil)
		if err != nil {
			log.Printf("Connection error: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn = c
		log.Println("Connected to BotNexus!")

		// Send Lifecycle Event: Connect
		sendEvent(map[string]interface{}{
			"post_type":       "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         selfID,
			"time":            time.Now().Unix(),
		})

		// Send Heartbeat
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if conn == nil {
					return
				}
				sendEvent(map[string]interface{}{
					"post_type":       "meta_event",
					"meta_event_type": "heartbeat",
					"self_id":         selfID,
					"time":            time.Now().Unix(),
					"status": map[string]interface{}{
						"online": true,
						"good":   true,
					},
				})
			}
		}()

		// Handle Incoming Actions
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				conn = nil
				break
			}
			handleAction(message)
		}
	}
}

func sendEvent(event map[string]interface{}) {
	if conn == nil {
		return
	}
	// Add platform info
	event["platform"] = "email"

	err := conn.WriteJSON(event)
	if err != nil {
		log.Printf("Write error: %v", err)
		conn = nil
	}
}

func handleAction(msg []byte) {
	var action map[string]interface{}
	if err := json.Unmarshal(msg, &action); err != nil {
		log.Printf("JSON Unmarshal error: %v", err)
		return
	}

	actionType, _ := action["action"].(string)
	params, _ := action["params"].(map[string]interface{})
	echo, _ := action["echo"]

	response := map[string]interface{}{
		"status":  "ok",
		"retcode": 0,
		"data":    nil,
		"echo":    echo,
	}

	switch actionType {
	case "send_private_msg":
		userID, _ := params["user_id"].(string) // Email address
		message, _ := params["message"].(string)

		// Attempt to parse subject from message (First line as subject?)
		// Simple strategy: Subject = "New Message", Body = message
		// Or try to detect if message starts with "Subject: "

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
			log.Printf("Failed to send email to %s: %v", userID, err)
			response["status"] = "failed"
			response["retcode"] = -1
		} else {
			log.Printf("Email sent to %s", userID)
		}
	}

	if conn != nil {
		conn.WriteJSON(response)
	}
}

func sendEmail(to, subject, body string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("BotMatrix <%s>", config.SmtpUsername)
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	addr := fmt.Sprintf("%s:%d", config.SmtpServer, config.SmtpPort)
	auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpServer)

	return e.Send(addr, auth)
}

func pollEmails() {
	log.Println("Connecting to IMAP server...")
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", config.ImapServer, config.ImapPort), nil)
	if err != nil {
		log.Fatalf("IMAP connection failed: %v", err)
	}
	log.Println("IMAP Connected")

	if err := c.Login(config.Username, config.Password); err != nil {
		log.Fatalf("IMAP Login failed: %v", err)
	}
	defer c.Logout()

	for {
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Printf("Select INBOX failed: %v", err)
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
			continue
		}

		// Search for UNSEEN messages
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.Seen}
		uids, err := c.Search(criteria)
		if err != nil {
			log.Printf("Search failed: %v", err)
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
			continue
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

		time.Sleep(time.Duration(config.PollInterval) * time.Second)
	}
}

func processMessage(msg *imap.Message, section *imap.BodySectionName) {
	r := msg.GetBody(section)
	if r == nil {
		return
	}

	// Simple body reading
	// For complex emails (MIME multipart), this might need parsing
	// But section usually gets the full body.
	// To do it right, we should use a parser.
	// But for now, let's try to read it as raw and maybe just extract text.

	// A better way is to use "github.com/emersion/go-message/mail" to parse the body
	// but I didn't import it. Let's do a basic read.

	bodyBytes, _ := ioutil.ReadAll(r)
	bodyStr := string(bodyBytes)

	// Construct OneBot Message
	sender := msg.Envelope.From[0]
	senderEmail := fmt.Sprintf("%s@%s", sender.MailboxName, sender.HostName)
	senderName := sender.PersonalName

	content := fmt.Sprintf("Subject: %s\n\n%s", msg.Envelope.Subject, bodyStr)

	log.Printf("Received email from %s: %s", senderEmail, msg.Envelope.Subject)

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

	sendEvent(event)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	loadConfig()

	go connectToNexus()
	go pollEmails()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
}
