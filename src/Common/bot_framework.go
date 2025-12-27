package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// BotConfig defines basic configuration for any bot
type BotConfig struct {
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	LogPort   int    `json:"log_port"`
	UseTLS    bool   `json:"use_tls"`   // 是否启用 TLS (HTTPS/WSS)
	CertFile  string `json:"cert_file"` // 证书文件路径
	KeyFile   string `json:"key_file"`  // 私钥文件路径
}

// LogManager handles log rotation and retrieval
type LogManager struct {
	entries []LogEntry
	max     int
	mutex   sync.Mutex
	logChan chan LogEntry
}

func NewLogManager(max int) *LogManager {
	return &LogManager{
		entries: make([]LogEntry, 0, max),
		max:     max,
		logChan: make(chan LogEntry, 100),
	}
}

func (m *LogManager) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	entry := LogEntry{
		Level:     "INFO",
		Time:      now.Format("15:04:05"),
		Timestamp: now,
		Message:   string(p),
	}
	m.entries = append(m.entries, entry)
	if len(m.entries) > m.max {
		m.entries = m.entries[len(m.entries)-m.max:]
	}

	// 非阻塞发送到通道
	select {
	case m.logChan <- entry:
	default:
	}

	return os.Stderr.Write(p)
}

func (m *LogManager) Log(level, msg string) {
	m.mutex.Lock()
	now := time.Now()
	entry := LogEntry{
		Level:     level,
		Time:      now.Format("15:04:05"),
		Timestamp: now,
		Message:   msg,
	}
	m.entries = append(m.entries, entry)
	if len(m.entries) > m.max {
		m.entries = m.entries[len(m.entries)-m.max:]
	}
	m.mutex.Unlock()

	select {
	case m.logChan <- entry:
	default:
	}

	fmt.Fprintf(os.Stderr, "[%s] %s %s\n", entry.Timestamp.Format("15:04:05"), level, msg)
}

func (m *LogManager) GetLogs(lines int) []LogEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if lines > len(m.entries) {
		lines = len(m.entries)
	}
	result := make([]LogEntry, lines)
	copy(result, m.entries[len(m.entries)-lines:])
	return result
}

// BaseBot provides common functionality for all bots
type BaseBot struct {
	Config     BotConfig
	LogManager *LogManager
	Mu         sync.RWMutex
	Ctx        context.Context
	Cancel     context.CancelFunc
	Mux        *http.ServeMux
	NexusConn  *websocket.Conn
	ConnMu     sync.Mutex
}

func NewBaseBot(defaultLogPort int) *BaseBot {
	ctx, cancel := context.WithCancel(context.Background())
	b := &BaseBot{
		LogManager: NewLogManager(1000),
		Config: BotConfig{
			LogPort: defaultLogPort,
		},
		Ctx:    ctx,
		Cancel: cancel,
		Mux:    http.NewServeMux(),
	}

	// 启动后台日志推送
	go b.logPusher()

	return b
}

func (b *BaseBot) Info(format string, v ...interface{}) {
	b.LogManager.Log("INFO", fmt.Sprintf(format, v...))
}

func (b *BaseBot) Warn(format string, v ...interface{}) {
	b.LogManager.Log("WARN", fmt.Sprintf(format, v...))
}

func (b *BaseBot) Error(format string, v ...interface{}) {
	b.LogManager.Log("ERROR", fmt.Sprintf(format, v...))
}

func (b *BaseBot) logPusher() {
	for {
		select {
		case <-b.Ctx.Done():
			return
		case entry := <-b.LogManager.logChan:
			b.ConnMu.Lock()
			conn := b.NexusConn
			b.ConnMu.Unlock()

			if conn != nil {
				// 将日志包装为 OneBot 风格的消息上报给 Nexus
				b.SendToNexus(map[string]interface{}{
					"post_type": "log",
					"level":     entry.Level,
					"time":      entry.Timestamp.Unix(),
					"message":   entry.Message,
					"source":    b.Config.BotToken, // 暂时用 Token 标识来源，或后续用 SelfID
				})
			}
		}
	}
}

func (b *BaseBot) LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err == nil {
		b.Mu.Lock()
		if err := json.Unmarshal(file, &b.Config); err != nil {
			log.Printf("Error parsing %s: %v", path, err)
		}
		b.Mu.Unlock()
	}

	// Override with environment variables
	if envToken := os.Getenv("BOT_TOKEN"); envToken != "" {
		b.Config.BotToken = envToken
	}
	if envAddr := os.Getenv("NEXUS_ADDR"); envAddr != "" {
		b.Config.NexusAddr = envAddr
	}
	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		var v int
		if _, err := fmt.Sscanf(envLogPort, "%d", &v); err == nil {
			b.Config.LogPort = v
		}
	}

	if b.Config.NexusAddr == "" {
		b.Config.NexusAddr = "ws://bot-manager:3005"
	}

	return nil
}

func (b *BaseBot) SaveConfig(path string) error {
	b.Mu.RLock()
	data, err := json.MarshalIndent(b.Config, "", "  ")
	b.Mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (b *BaseBot) StartHTTPServer() {
	b.Mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		linesStr := r.URL.Query().Get("lines")
		lines := 100
		if linesStr != "" {
			fmt.Sscanf(linesStr, "%d", &lines)
		}
		logs := b.LogManager.GetLogs(lines)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})

	b.Mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", b.Config.LogPort)
	server := &http.Server{
		Addr:    addr,
		Handler: b.Mux,
	}

	go func() {
		b.Mu.RLock()
		useTLS := b.Config.UseTLS
		certFile := b.Config.CertFile
		keyFile := b.Config.KeyFile
		b.Mu.RUnlock()

		if useTLS && certFile != "" && keyFile != "" {
			log.Printf("Starting HTTPS/WSS server on %s", addr)
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS server failed: %v", err)
			}
		} else {
			log.Printf("Starting HTTP/WS server on %s", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server failed: %v", err)
			}
		}
	}()

	go func() {
		<-b.Ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
}

func (b *BaseBot) WaitExitSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down...")
	b.Cancel()
}

// StartNexusConnection connects to BotNexus and handles reconnection
func (b *BaseBot) StartNexusConnection(ctx context.Context, addr, platform, selfID string, commandHandler func([]byte)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Connecting to BotNexus at %s...", addr)
				header := http.Header{}
				header.Add("X-Self-ID", selfID)
				header.Add("X-Platform", platform)

				conn, _, err := websocket.DefaultDialer.Dial(addr, header)
				if err != nil {
					log.Printf("BotNexus connection failed: %v. Retrying in 5s...", err)
					select {
					case <-ctx.Done():
						return
					case <-time.After(5 * time.Second):
						continue
					}
				}

				b.ConnMu.Lock()
				b.NexusConn = conn
				b.ConnMu.Unlock()
				log.Println("Connected to BotNexus!")

				// Send Lifecycle Event
				b.SendToNexus(map[string]interface{}{
					"post_type":       "meta_event",
					"meta_event_type": "lifecycle",
					"sub_type":        "connect",
					"self_id":         selfID,
					"time":            time.Now().Unix(),
				})

				// Handle incoming commands
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("BotNexus disconnected: %v", err)
						b.ConnMu.Lock()
						if b.NexusConn == conn {
							b.NexusConn = nil
						}
						b.ConnMu.Unlock()
						break
					}
					commandHandler(message)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (b *BaseBot) SendToNexus(msg interface{}) {
	b.ConnMu.Lock()
	defer b.ConnMu.Unlock()
	if b.NexusConn == nil {
		return
	}
	if err := b.NexusConn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send to Nexus: %v", err)
		b.NexusConn.Close()
		b.NexusConn = nil
	}
}
