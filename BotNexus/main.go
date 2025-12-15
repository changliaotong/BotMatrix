package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Config
var (
	WS_PORT    = ":3001"
	WEBUI_PORT = ":5000"
	STATS_FILE = "stats.json"
	REDIS_ADDR = "192.168.0.126:6379"
	REDIS_PWD  = "redis_zsYik8"
)

func init() {
	if v := os.Getenv("WS_PORT"); v != "" {
		WS_PORT = v
	}
	if v := os.Getenv("WEBUI_PORT"); v != "" {
		WEBUI_PORT = v
	}
	if v := os.Getenv("STATS_FILE"); v != "" {
		STATS_FILE = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		REDIS_ADDR = v
	}
	if v := os.Getenv("REDIS_PWD"); v != "" {
		REDIS_PWD = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		JWT_SECRET = []byte(v)
	}
}

// --- JWT & Magic Link Helpers ---

var JWT_SECRET = []byte("botmatrix_secret_key_change_me_in_prod")

type UserClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(username, role string) (string, error) {
	claims := UserClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24h expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET)
}

func ValidateJWT(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWT_SECRET, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn        *websocket.Conn
	SelfID      string
	Nickname    string
	GroupCount  int
	FriendCount int
	Connected   time.Time
	Platform    string
	Mutex       sync.Mutex
	SentCount   int64 // New: Track sent messages per bot session
	RecvCount   int64 // New: Track received messages per bot session
}

type WorkerClient struct {
	Conn         *websocket.Conn
	Mutex        sync.Mutex
	Connected    time.Time
	HandledCount int64
}

type Subscriber struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	User  *User
}

// Manager holds the state
type Manager struct {
	bots        map[string]*BotClient
	subscribers map[*websocket.Conn]*Subscriber // UI or other consumers (Broadcast)
	workers     []*WorkerClient                 // Business logic workers (Round-Robin)
	workerIndex int                             // For Round-Robin
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
	logBuffer   []LogEntry
	logMutex    sync.RWMutex

	// Pending Requests (Echo -> Channel)
	pendingRequests map[string]chan map[string]interface{}
	pendingMutex    sync.Mutex

	// Redis
	rdb *redis.Client

	// Chat Stats
	statsMutex      sync.RWMutex
	TotalMessages   int64            `json:"total_messages"`    // Global counter
	SentMessages    int64            `json:"sent_messages"`     // Global sent counter
	UserStats       map[int64]int64  `json:"user_stats"`        // UserID -> Count (Total)
	GroupStats      map[int64]int64  `json:"group_stats"`       // GroupID -> Count (Total)
	BotStats        map[string]int64 `json:"bot_stats"`         // BotID -> Count (Total Recv)
	BotStatsSent    map[string]int64 `json:"bot_stats_sent"`    // BotID -> Count (Total Sent)
	UserStatsToday  map[int64]int64  `json:"user_stats_today"`  // UserID -> Count (Today)
	GroupStatsToday map[int64]int64  `json:"group_stats_today"` // GroupID -> Count (Today)
	BotStatsToday   map[string]int64 `json:"bot_stats_today"`   // BotID -> Count (Today)
	LastResetDate   string           `json:"last_reset_date"`   // YYYY-MM-DD

	// Granular Stats (Per Bot)
	BotDetailedStats map[string]*BotStatDetail `json:"bot_detailed_stats"` // BotID -> Detail

	// Time Series Stats (New)
	HistoryMutex sync.RWMutex
	CPUTrend     []float64 `json:"cpu_trend"`
	MemTrend     []float64 `json:"mem_trend"`
	MsgTrend     []float64 `json:"msg_trend"` // Msg count per interval
	SentTrend    []float64 `json:"sent_trend"`
	RecvTrend    []float64 `json:"recv_trend"`
	CurrentCPU   float64   `json:"-"`

	UserNames  map[int64]string `json:"-"` // Cache names
	GroupNames map[int64]string `json:"-"` // Cache names

	// Deduplication
	DedupMutex sync.Mutex
	DedupCache map[string]int64 `json:"-"` // MessageID -> Timestamp (Unix)

	// Active Sessions (Contacts)
	SessionMutex sync.RWMutex
	Sessions     map[string]*ContactSession `json:"-"` // Key: BotID:Type:ID

	// Auto Recall
	AutoRecallMutex sync.RWMutex
	AutoRecallMap   map[string]AutoRecallTask `json:"-"` // Echo -> Task
}

type AutoRecallTask struct {
	Delay int // Seconds
	BotID string
}

type ContactSession struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"` // "private", "group", "guild"
	BotID       string             `json:"bot_id"`
	GuildID     string             `json:"guild_id,omitempty"`
	LastActive  int64              `json:"last_active"`
	LastMsgID   string             `json:"last_msg_id"`
	LastMsgTime int64              `json:"last_msg_time"`
	ActiveBots  map[string]BotInfo `json:"active_bots"` // Other bots seen in this group
}

type BotInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Platform string `json:"platform"`
}

type BotStatDetail struct {
	UserStats       map[int64]int64 `json:"user_stats"`
	GroupStats      map[int64]int64 `json:"group_stats"`
	UserStatsToday  map[int64]int64 `json:"user_stats_today"`
	GroupStatsToday map[int64]int64 `json:"group_stats_today"`
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	BotID   string `json:"bot_id,omitempty"`
}

func NewManager() *Manager {
	m := &Manager{
		bots:             make(map[string]*BotClient),
		subscribers:      make(map[*websocket.Conn]*Subscriber),
		workers:          make([]*WorkerClient, 0),
		UserStats:        make(map[int64]int64),
		GroupStats:       make(map[int64]int64),
		BotStats:         make(map[string]int64),
		BotStatsSent:     make(map[string]int64),
		UserStatsToday:   make(map[int64]int64),
		GroupStatsToday:  make(map[int64]int64),
		BotStatsToday:    make(map[string]int64),
		BotDetailedStats: make(map[string]*BotStatDetail),
		CPUTrend:         make([]float64, 0),
		MemTrend:         make([]float64, 0),
		MsgTrend:         make([]float64, 0),
		SentTrend:        make([]float64, 0),
		RecvTrend:        make([]float64, 0),
		UserNames:        make(map[int64]string),
		GroupNames:       make(map[int64]string),
		DedupCache:       make(map[string]int64),
		LastResetDate:    time.Now().Format("2006-01-02"),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logBuffer:       make([]LogEntry, 0, 2000),
		Sessions:        make(map[string]*ContactSession),
		AutoRecallMap:   make(map[string]AutoRecallTask),
		pendingRequests: make(map[string]chan map[string]interface{}),
	}

	// Initialize Redis
	m.rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PWD,
		DB:       0, // use default DB
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] Failed to connect to Redis at %s: %v. Running without Redis persistence.", REDIS_ADDR, err)
		m.rdb = nil
	} else {
		log.Printf("[INFO] Connected to Redis at %s", REDIS_ADDR)
		// Clear previous session data
		m.rdb.Del(context.Background(), "bots:online")
	}

	m.LoadStats()
	return m
}

func (m *Manager) AddLog(level, message string, botID ...string) {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()

	var bid string
	if len(botID) > 0 {
		bid = botID[0]
	}

	entry := LogEntry{
		Time:    time.Now().Format("15:04:05"),
		Level:   level,
		Message: message,
		BotID:   bid,
	}

	if len(m.logBuffer) >= 2000 {
		m.logBuffer = m.logBuffer[1:]
	}
	m.logBuffer = append(m.logBuffer, entry)
	log.Printf("[%s] %s", level, message)

	// Broadcast log to subscribers (wrapped in event)
	go m.broadcastToSubscribers(map[string]interface{}{
		"post_type": "log",
		"data":      entry,
		"self_id":   bid,
	})
}

func (m *Manager) SaveStats() {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	data := map[string]interface{}{
		"total_messages":     m.TotalMessages,
		"sent_messages":      m.SentMessages,
		"user_stats":         m.UserStats,
		"group_stats":        m.GroupStats,
		"bot_stats":          m.BotStats,
		"bot_stats_sent":     m.BotStatsSent,
		"user_stats_today":   m.UserStatsToday,
		"group_stats_today":  m.GroupStatsToday,
		"bot_stats_today":    m.BotStatsToday,
		"last_reset_date":    m.LastResetDate,
		"bot_detailed_stats": m.BotDetailedStats,
		"cpu_trend":          m.CPUTrend,
		"mem_trend":          m.MemTrend,
		"msg_trend":          m.MsgTrend,
		"sent_trend":         m.SentTrend,
		"recv_trend":         m.RecvTrend,
		"user_names":         m.UserNames,
		"group_names":        m.GroupNames,
	}

	file, err := os.Create(STATS_FILE)
	if err != nil {
		log.Printf("Error creating stats file: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Printf("Error encoding stats: %v", err)
	}
}

func (m *Manager) LoadStats() {
	file, err := os.Open(STATS_FILE)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error opening stats file: %v", err)
		}
		return
	}
	defer file.Close()

	var data struct {
		TotalMessages    int64                     `json:"total_messages"`
		SentMessages     int64                     `json:"sent_messages"`
		UserStats        map[int64]int64           `json:"user_stats"`
		GroupStats       map[int64]int64           `json:"group_stats"`
		BotStats         map[string]int64          `json:"bot_stats"`
		BotStatsSent     map[string]int64          `json:"bot_stats_sent"`
		UserStatsToday   map[int64]int64           `json:"user_stats_today"`
		GroupStatsToday  map[int64]int64           `json:"group_stats_today"`
		BotStatsToday    map[string]int64          `json:"bot_stats_today"`
		LastResetDate    string                    `json:"last_reset_date"`
		BotDetailedStats map[string]*BotStatDetail `json:"bot_detailed_stats"`
		CPUTrend         []float64                 `json:"cpu_trend"`
		MemTrend         []float64                 `json:"mem_trend"`
		MsgTrend         []float64                 `json:"msg_trend"` // Total (Sent + Recv)
		SentTrend        []float64                 `json:"sent_trend"`
		RecvTrend        []float64                 `json:"recv_trend"`

		UserNames  map[int64]string `json:"user_names"`
		GroupNames map[int64]string `json:"group_names"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Printf("Error decoding stats: %v", err)
		return
	}

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.TotalMessages = data.TotalMessages
	m.SentMessages = data.SentMessages

	if data.UserStats != nil {
		m.UserStats = data.UserStats
	}
	if data.GroupStats != nil {
		m.GroupStats = data.GroupStats
	}
	if data.BotStats != nil {
		m.BotStats = data.BotStats
	}
	if data.BotStatsSent != nil {
		m.BotStatsSent = data.BotStatsSent
	}
	if data.UserStatsToday != nil {
		m.UserStatsToday = data.UserStatsToday
	}
	if data.GroupStatsToday != nil {
		m.GroupStatsToday = data.GroupStatsToday
	}
	if data.BotStatsToday != nil {
		m.BotStatsToday = data.BotStatsToday
	}
	if data.LastResetDate != "" {
		m.LastResetDate = data.LastResetDate
	}
	if data.BotDetailedStats != nil {
		m.BotDetailedStats = data.BotDetailedStats
	} else {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}

	if data.CPUTrend != nil {
		m.CPUTrend = data.CPUTrend
	}
	if data.MemTrend != nil {
		m.MemTrend = data.MemTrend
	}
	if data.MsgTrend != nil {
		m.MsgTrend = data.MsgTrend
	}
	if data.SentTrend != nil {
		m.SentTrend = data.SentTrend
	}
	if data.RecvTrend != nil {
		m.RecvTrend = data.RecvTrend
	}

	if data.UserNames != nil {
		m.UserNames = data.UserNames
	}
	if data.GroupNames != nil {
		m.GroupNames = data.GroupNames
	}
	log.Printf("Loaded stats: %d users, %d groups (Last Reset: %s)", len(m.UserStats), len(m.GroupStats), m.LastResetDate)
}

func (m *Manager) GetLogs() []LogEntry {
	m.logMutex.RLock()
	defer m.logMutex.RUnlock()
	// Return a copy
	logs := make([]LogEntry, len(m.logBuffer))
	copy(logs, m.logBuffer)
	return logs
}

func main() {
	manager := NewManager()
	manager.LoadStats()

	// 1. WebSocket Server Mux
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWS(manager, w, r)
	})

	go func() {
		manager.AddLog("INFO", fmt.Sprintf("Starting OneBot Gateway on %s", WS_PORT))
		if err := http.ListenAndServe(WS_PORT, wsMux); err != nil {
			log.Fatal("WS Server error:", err)
		}
	}()

	// Periodic Save
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			manager.SaveStats()
		}
	}()

	// Periodic Bot Info Refresh (Every 1 hour to ensure data consistency)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			manager.mutex.RLock()
			for _, bot := range manager.bots {
				go func(client *BotClient) {
					client.Mutex.Lock()
					defer client.Mutex.Unlock()

					// Group & Friend Count
					// Check Platform: Guild bots use custom count actions; QQ/Others use list fetching
					isGuild := strings.Contains(strings.ToLower(client.Platform), "guild")

					if isGuild {
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_group_count",
							"echo":   "internal_get_group_count",
						})
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_friend_count",
							"echo":   "internal_get_friend_count",
						})
					} else {
						// Standard OneBot / QQ: Fetch full lists to count
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_group_list",
							"echo":   "internal_get_group_list_count",
						})
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_friend_list",
							"echo":   "internal_get_friend_list_count",
						})
					}
				}(bot)
			}
			manager.mutex.RUnlock()
		}
	}()

	// Periodic Trend Collection (2s interval)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		lastRecvCount := manager.TotalMessages
		lastSentCount := manager.SentMessages
		for range ticker.C {
			// CPU
			c, err := cpu.Percent(0, false) // Instant if possible, or use short interval?
			// cpu.Percent(0, false) returns error if interval is 0? No, it returns since last call.
			// But first call returns 0.
			// Better: cpu.Percent(time.Second, false) but that blocks.
			// Let's use a non-blocking approach if possible or just accept the block in this goroutine.
			cpuVal := 0.0
			if err == nil && len(c) > 0 {
				cpuVal = c[0]
			}

			// Mem
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			memVal := float64(mem.Alloc)

			// Msg Throughput (msgs per 2s)
			manager.statsMutex.RLock()
			currentRecvCount := manager.TotalMessages
			currentSentCount := manager.SentMessages
			manager.statsMutex.RUnlock()

			recvDelta := float64(currentRecvCount - lastRecvCount)
			if recvDelta < 0 {
				recvDelta = 0
			}
			lastRecvCount = currentRecvCount

			sentDelta := float64(currentSentCount - lastSentCount)
			if sentDelta < 0 {
				sentDelta = 0
			}
			lastSentCount = currentSentCount

			totalDelta := recvDelta + sentDelta

			// Update Trends
			manager.HistoryMutex.Lock()

			manager.CurrentCPU = cpuVal

			// CPU Trend
			manager.CPUTrend = append(manager.CPUTrend, cpuVal)
			if len(manager.CPUTrend) > 1800 {
				manager.CPUTrend = manager.CPUTrend[1:]
			}

			// Mem Trend
			manager.MemTrend = append(manager.MemTrend, memVal)
			if len(manager.MemTrend) > 1800 {
				manager.MemTrend = manager.MemTrend[1:]
			}

			// Msg Trend (Total)
			manager.MsgTrend = append(manager.MsgTrend, totalDelta)
			if len(manager.MsgTrend) > 1800 {
				manager.MsgTrend = manager.MsgTrend[1:]
			}

			// Sent Trend
			manager.SentTrend = append(manager.SentTrend, sentDelta)
			if len(manager.SentTrend) > 1800 {
				manager.SentTrend = manager.SentTrend[1:]
			}

			// Recv Trend
			manager.RecvTrend = append(manager.RecvTrend, recvDelta)
			if len(manager.RecvTrend) > 1800 {
				manager.RecvTrend = manager.RecvTrend[1:]
			}
			manager.HistoryMutex.Unlock()
		}
	}()

	// 2. Web UI Server Mux
	uiMux := http.NewServeMux()

	// API Endpoints
	uiMux.HandleFunc("/api/bots", manager.handleGetBots)
	uiMux.HandleFunc("/api/workers", manager.handleGetWorkers)
	uiMux.HandleFunc("/api/logs", manager.handleGetLogs)
	uiMux.HandleFunc("/api/stats", manager.handleGetStats)
	uiMux.HandleFunc("/api/stats/chat", manager.handleGetChatStats)
	uiMux.HandleFunc("/api/system/stats", manager.handleSystemStats)
	uiMux.HandleFunc("/api/contacts", manager.handleGetContacts)
	uiMux.HandleFunc("/api/action", manager.handleAction)
	uiMux.HandleFunc("/api/smart_action", manager.handleSmartAction)

	// Auth API
	uiMux.HandleFunc("/api/login", manager.handleLogin)
	uiMux.HandleFunc("/api/login/magic", manager.handleMagicLogin)
	uiMux.HandleFunc("/api/admin/magic_link", manager.handleGenerateMagicLink)
	uiMux.HandleFunc("/api/me", manager.handleMe)
	uiMux.HandleFunc("/api/user/password", manager.handleUpdatePassword)
	uiMux.HandleFunc("/api/admin/assign", manager.handleAssignBot) // Admin only

	// Static Files
	fs := http.FileServer(http.Dir("."))
	uiMux.Handle("/", fs)

	manager.AddLog("INFO", fmt.Sprintf("Starting Web UI on %s", WEBUI_PORT))
	if err := http.ListenAndServe(WEBUI_PORT, uiMux); err != nil {
		log.Fatal("WebUI Server error:", err)
	}
}

func (m *Manager) broadcastToSubscribers(data interface{}) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Extract self_id for filtering
	var selfID string
	if msgMap, ok := data.(map[string]interface{}); ok {
		if id, ok := msgMap["self_id"]; ok {
			selfID = fmt.Sprintf("%v", id)
		}
	}

	// 1. Broadcast to passive subscribers (UI, Monitors)
	for conn, sub := range m.subscribers {
		// Filter
		if sub.User != nil && sub.User.Role != "admin" {
			// If message has self_id, check ownership
			if selfID != "" && !sub.User.OwnedBots[selfID] {
				// Special case: "meta_event" lifecycle might be relevant?
				// Usually strict filtering is better.
				continue
			}
		}

		sub.Mutex.Lock()
		err := sub.Conn.WriteJSON(data)
		sub.Mutex.Unlock()
		if err != nil {
			go func(c *websocket.Conn) {
				c.Close()
				m.mutex.Lock()
				delete(m.subscribers, c)
				m.mutex.Unlock()
			}(conn)
		}
	}

	// 2. Load Balance to Workers (Business Logic)
	// ... (Existing Worker Logic) ...
	if len(m.workers) > 0 {
		var eventSummary string
		if msgMap, ok := data.(map[string]interface{}); ok {
			if pt, ok := msgMap["post_type"].(string); ok {
				// Prevent infinite loop: Don't log "log" events
				if pt == "log" {
					// Just dispatch, don't log to avoid recursion
				} else {
					eventSummary = fmt.Sprintf("Type: %s", pt)
					if sub, ok := msgMap["sub_type"].(string); ok {
						eventSummary += fmt.Sprintf(", Sub: %s", sub)
					}
					if msg, ok := msgMap["raw_message"].(string); ok {
						if len(msg) > 50 {
							eventSummary += fmt.Sprintf(", Msg: %s...", msg[:50])
						} else {
							eventSummary += fmt.Sprintf(", Msg: %s", msg)
						}
					}
					// Use log.Printf instead of m.AddLog to avoid infinite recursion loop
					// m.AddLog triggers broadcastToSubscribers which triggers m.AddLog...
					if eventSummary != "" {
						// log.Printf("[DEBUG] Dispatching event to worker: %s", eventSummary)
					}
				}
			}
		}
		targetIndex := int(time.Now().UnixNano()) % len(m.workers)
		worker := m.workers[targetIndex]

		worker.Mutex.Lock()
		err := worker.Conn.WriteJSON(data)
		worker.HandledCount++
		worker.Mutex.Unlock()

		if err != nil {
			go func(w *WorkerClient) {
				m.removeWorker(w)
			}(worker)
			for i, w := range m.workers {
				if i == targetIndex {
					continue
				}
				w.Mutex.Lock()
				e := w.Conn.WriteJSON(data)
				w.Mutex.Unlock()
				if e == nil {
					break
				}
			}
		}
	} else {
		// Only log if it's a message event to avoid noise
		if msgMap, ok := data.(map[string]interface{}); ok {
			if pt, ok := msgMap["post_type"].(string); ok && pt == "message" {
				m.AddLog("WARN", "No workers available to handle message event!")
			}
		}
	}

	// 3. Broadcast to other Bots (Universal Clients / Controllers)
	// This allows C# clients connecting as standard Bots (without role=worker) to receive events.
	// We act as a message broker/router here.

	// FIX: Don't send logs or meta_events to other bots to avoid infinite loops with OneBot implementations (like NapCat)
	// that might treat incoming JSON as API requests and error out (triggering more logs).
	// ALSO: Don't send API Responses (which have no post_type) to other bots, as they might be confused.
	// We ONLY want to broadcast "message", "notice", or "request" events to other bots (like C# clients).
	shouldBroadcastToBots := false
	if msgMap, ok := data.(map[string]interface{}); ok {
		if pt, ok := msgMap["post_type"].(string); ok {
			// Allow message, notice, request
			if pt == "message" || pt == "notice" || pt == "request" {
				shouldBroadcastToBots = true
			}
		}
	}

	if !shouldBroadcastToBots {
		return
	}

	for id, bot := range m.bots {
		// Don't send back to self (sender)
		if selfID != "" && id == selfID {
			continue
		}

		// Fix: Don't broadcast to standard OneBot implementations (like NapCat/LLOneBot)
		// that don't support incoming events and treat them as "undefined" API actions.
		// These usually default to "QQ" platform.
		if bot.Platform == "QQ" {
			continue
		}

		// Avoid sending to unauthenticated bots if strictly required?
		// For now, trust internal network.

		bot.Mutex.Lock()
		err := bot.Conn.WriteJSON(data)
		bot.Mutex.Unlock()
		if err != nil {
			// Just log, don't remove from map here (Read Loop will handle disconnect)
			// log.Printf("Error broadcasting to bot %s: %v", id, err)
		}
	}
}

func (m *Manager) removeWorker(target *WorkerClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	newWorkers := make([]*WorkerClient, 0)
	for _, w := range m.workers {
		if w != target {
			newWorkers = append(newWorkers, w)
		}
	}
	m.workers = newWorkers
	target.Conn.Close()
	m.AddLog("INFO", "Worker removed due to error")
}

func serveWS(m *Manager, w http.ResponseWriter, r *http.Request) {
	// Check role
	role := r.URL.Query().Get("role")
	if role == "subscriber" {
		serveSubscriber(m, w, r)
		return
	} else if role == "worker" {
		serveWorker(m, w, r)
		return
	}

	// Check if it's a bot or a client
	// For now, we assume everything connecting to 3001 is a bot/client complying with OneBot
	// Headers: X-Self-ID, X-Client-Role, X-Platform
	platform := r.Header.Get("X-Platform")
	if platform == "" {
		platform = r.URL.Query().Get("platform")
	}
	if platform == "" {
		platform = "QQ" // Default to QQ
	}

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Upgrade error: %v", err))
		return
	}

	// Read first message to identify or wait for lifecycle event
	// Or use headers if available. OneBot 11 uses headers.
	selfID := r.Header.Get("X-Self-ID")
	if selfID == "" {
		// Fallback: wait for identification?
		// For simplicity, we assign a temp ID or wait for first event
		selfID = fmt.Sprintf("unknown-%d", time.Now().UnixNano())
	}

	client := &BotClient{
		Conn:      conn,
		SelfID:    selfID,
		Connected: time.Now(),
		Platform:  platform,
	}

	m.mutex.Lock()
	m.bots[selfID] = client
	m.mutex.Unlock()

	m.AddLog("INFO", fmt.Sprintf("Client connected: %s (%s) [Platform: %s]", selfID, r.RemoteAddr, platform))

	defer func() {
		m.mutex.Lock()
		delete(m.bots, selfID)
		m.mutex.Unlock()

		if m.rdb != nil {
			ctx := context.Background()
			m.rdb.SRem(ctx, "bots:online", selfID)
			// Don't delete info, keep it for offline history
			// m.rdb.Del(ctx, fmt.Sprintf("bot:info:%s", selfID))

			// Mark as disconnected
			m.rdb.HSet(ctx, fmt.Sprintf("bot:info:%s", selfID), "disconnected_at", time.Now().Format(time.RFC3339))
			m.rdb.HSet(ctx, fmt.Sprintf("bot:info:%s", selfID), "platform", platform)
		}

		conn.Close()
		m.AddLog("INFO", fmt.Sprintf("Client disconnected: %s [Platform: %s]", selfID, platform))
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Try to parse message to update SelfID if it's a lifecycle event
		var msgMap map[string]interface{}
		if err := json.Unmarshal(message, &msgMap); err != nil {
			m.AddLog("ERROR", fmt.Sprintf("Failed to parse message from %s: %v | Content: %s", selfID, err, string(message)))
			continue
		}

		// Handle Auto Recall Response
		if echo, ok := msgMap["echo"].(string); ok && echo != "" {
			m.AutoRecallMutex.RLock()
			task, exists := m.AutoRecallMap[echo]
			m.AutoRecallMutex.RUnlock()

			if exists {
				// Remove task
				m.AutoRecallMutex.Lock()
				delete(m.AutoRecallMap, echo)
				m.AutoRecallMutex.Unlock()

				// Check status
				status, _ := msgMap["status"].(string)
				retcode, _ := msgMap["retcode"].(float64)

				if status == "ok" || retcode == 0 {
					// Extract message_id
					var msgID string
					if data, ok := msgMap["data"].(map[string]interface{}); ok {
						msgID = getString(data, "message_id")
					}

					if msgID != "" {
						go func(botID string, msgID string, delay int) {
							if delay > 0 {
								time.Sleep(time.Duration(delay) * time.Second)
							}

							m.mutex.RLock()
							targetBot, ok := m.bots[botID]
							m.mutex.RUnlock()

							if ok {
								m.AddLog("INFO", fmt.Sprintf("Auto Recall: Deleting message %s from Bot %s after %ds", msgID, botID, delay))

								targetBot.Mutex.Lock()
								targetBot.Conn.WriteJSON(map[string]interface{}{
									"action": "delete_msg",
									"params": map[string]interface{}{
										"message_id": msgID,
									},
								})
								targetBot.Mutex.Unlock()
							}
						}(task.BotID, msgID, task.Delay)
					}
				}
			}
		}

		// Update SelfID if needed
		if id, ok := msgMap["self_id"]; ok {
			var newID string
			switch v := id.(type) {
			case float64:
				newID = fmt.Sprintf("%.0f", v)
			case string:
				newID = v
			default:
				newID = fmt.Sprintf("%v", v)
			}

			if newID != "" && newID != "0" && newID != selfID {
				m.mutex.Lock()
				// Check if we are renaming or just updating
				// If selfID is unknown-..., we remove it and add new key
				// But we need to make sure we don't overwrite an existing connection if duplicate?
				// For now simple rename logic:
				delete(m.bots, selfID)

				selfID = newID
				client.SelfID = selfID
				m.bots[selfID] = client
				m.mutex.Unlock()
				m.AddLog("INFO", fmt.Sprintf("Client identified as: %s", selfID))

				// Update Redis
				if m.rdb != nil {
					ctx := context.Background()
					m.rdb.SAdd(ctx, "bots:online", selfID)
					m.rdb.SAdd(ctx, "bots:all", selfID) // Track all bots ever connected
					m.rdb.HSet(ctx, fmt.Sprintf("bot:info:%s", selfID), map[string]interface{}{
						"connected_at": client.Connected.Format(time.RFC3339),
						"remote_addr":  client.Conn.RemoteAddr().String(),
						"is_alive":     true, // Explicitly mark as alive
						"platform":     client.Platform,
					})
					m.rdb.HDel(ctx, fmt.Sprintf("bot:info:%s", selfID), "disconnected_at") // Clear disconnect time
				}

				// Trigger get_login_info, get_group_list, get_friend_list
				go func() {
					client.Mutex.Lock()
					defer client.Mutex.Unlock()

					// Nickname
					client.Conn.WriteJSON(map[string]interface{}{
						"action": "get_login_info",
						"echo":   "internal_get_login_info",
					})

					// Group & Friend Count
					// Check Platform: Guild bots use custom count actions; QQ/Others use list fetching
					isGuild := strings.Contains(strings.ToLower(client.Platform), "guild")

					if isGuild {
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_group_count",
							"echo":   "internal_get_group_count",
						})
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_friend_count",
							"echo":   "internal_get_friend_count",
						})
					} else {
						// Standard OneBot / QQ: Fetch full lists to count
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_group_list",
							"echo":   "internal_get_group_list_count",
						})
						client.Conn.WriteJSON(map[string]interface{}{
							"action": "get_friend_list",
							"echo":   "internal_get_friend_list_count",
						})
					}
				}()
			}
		}

		// --- Log Forwarding ---
		if pt, ok := msgMap["post_type"].(string); ok && pt == "log" {
			level, _ := msgMap["level"].(string)
			message, _ := msgMap["message"].(string)
			if level == "" {
				level = "INFO"
			}
			m.AddLog(level, message, selfID)
			continue
		}

		// --- Magic Link Logic ---
		if pt, ok := msgMap["post_type"].(string); ok && pt == "message" {
			raw, _ := msgMap["raw_message"].(string)
			if raw == "后台" || strings.ToLower(raw) == "login" {
				// Get Sender ID
				var userIDStr string
				if uid, ok := msgMap["user_id"]; ok {
					userIDStr = fmt.Sprintf("%v", uid)
				}

				if userIDStr != "" {
					// Generate Token
					token := GenerateRandomToken(32)
					// Save to Redis (User ID as Username)
					if m.rdb != nil {
						key := fmt.Sprintf("auth:magic:%s", token)
						m.rdb.Set(context.Background(), key, userIDStr, 5*time.Minute)

						// Construct URL
						// Use localhost for local demo
						link := fmt.Sprintf("http://localhost%s/?magic_token=%s", WEBUI_PORT, token)

						reply := map[string]interface{}{
							"action": "send_msg",
							"params": map[string]interface{}{
								"user_id": msgMap["user_id"],
								"message": fmt.Sprintf("免密码登录链接 (5分钟有效):\n%s", link),
							},
						}

						// Support Group
						if mt, ok := msgMap["message_type"].(string); ok && mt == "group" {
							reply["params"].(map[string]interface{})["group_id"] = msgMap["group_id"]
						}

						conn.WriteJSON(reply)
					}
				}
			}
		}

		// Update Info from Internal Requests
		if echo, ok := msgMap["echo"].(string); ok {
			switch echo {
			case "internal_get_login_info":
				if data, ok := msgMap["data"].(map[string]interface{}); ok {
					if nick, ok := data["nickname"].(string); ok {
						client.Nickname = nick
						if m.rdb != nil {
							m.rdb.HSet(context.Background(), fmt.Sprintf("bot:info:%s", selfID), "nickname", nick)
						}
					}
				}
			case "internal_get_group_count":
				// Debug Logging for Group Count Issue
				if data, ok := msgMap["data"].(map[string]interface{}); ok {
					if countVal, ok := data["count"]; ok {
						var count int
						switch v := countVal.(type) {
						case float64:
							count = int(v)
						case int:
							count = v
						case int64:
							count = int(v)
						}
						client.GroupCount = count
						m.AddLog("DEBUG", fmt.Sprintf("Bot %s Group Count: %d", selfID, client.GroupCount))
						if m.rdb != nil {
							m.rdb.HSet(context.Background(), fmt.Sprintf("bot:info:%s", selfID), "group_count", client.GroupCount)
						}
					}
				} else {
					m.AddLog("WARN", fmt.Sprintf("Bot %s returned invalid group_count data: %v", selfID, msgMap["data"]))
				}

			case "internal_get_friend_count":
				if data, ok := msgMap["data"].(map[string]interface{}); ok {
					if countVal, ok := data["count"]; ok {
						var count int
						switch v := countVal.(type) {
						case float64:
							count = int(v)
						case int:
							count = v
						case int64:
							count = int(v)
						}
						client.FriendCount = count
						m.AddLog("DEBUG", fmt.Sprintf("Bot %s Friend Count: %d", selfID, client.FriendCount))
						if m.rdb != nil {
							m.rdb.HSet(context.Background(), fmt.Sprintf("bot:info:%s", selfID), "friend_count", client.FriendCount)
						}
					}
				} else {
					m.AddLog("WARN", fmt.Sprintf("Bot %s returned invalid friend_count data: %v", selfID, msgMap["data"]))
				}

			case "internal_get_group_list_count":
				if data, ok := msgMap["data"].([]interface{}); ok {
					count := len(data)
					client.GroupCount = count
					m.AddLog("DEBUG", fmt.Sprintf("Bot %s Group Count (from list): %d", selfID, client.GroupCount))
					if m.rdb != nil {
						m.rdb.HSet(context.Background(), fmt.Sprintf("bot:info:%s", selfID), "group_count", client.GroupCount)
					}
				}

			case "internal_get_friend_list_count":
				if data, ok := msgMap["data"].([]interface{}); ok {
					count := len(data)
					client.FriendCount = count
					m.AddLog("DEBUG", fmt.Sprintf("Bot %s Friend Count (from list): %d", selfID, client.FriendCount))
					if m.rdb != nil {
						m.rdb.HSet(context.Background(), fmt.Sprintf("bot:info:%s", selfID), "friend_count", client.FriendCount)
					}
				}
			}
		}
		// Fallback: check lifecycle meta_event for nickname? (OneBot 11 doesn't specify it usually)

		// Broadcast to subscribers
		m.broadcastToSubscribers(msgMap)

		// Log API response
		if _, ok := msgMap["echo"]; ok {
			// Don't log internal login info echo to avoid clutter if frequent? Actually it's once.
			m.AddLog("DEBUG", fmt.Sprintf("Recv API Resp from %s: %s", selfID, string(message)))
		}

		// Update Recv Count (Session)
		if pt, ok := msgMap["post_type"].(string); ok && pt == "message" {
			client.Mutex.Lock()
			client.RecvCount++
			client.Mutex.Unlock()
		}

		// Record Stats
		go m.recordStats(selfID, msgMap)

		// Log heartbeat only occasionally or filter it
		if msgMap != nil {
			if pt, ok := msgMap["post_type"].(string); !ok || pt != "meta_event" {
				// Don't log full content of huge messages (like get_group_list response)
				// But do log that we received SOMETHING
				msgStr := string(message)
				if len(msgStr) > 1000 {
					msgStr = msgStr[:1000] + "...(truncated)"
				}
				m.AddLog("DEBUG", fmt.Sprintf("Recv from %s: %s", selfID, msgStr))
			}
		}
	}
}

func serveSubscriber(m *Manager, w http.ResponseWriter, r *http.Request) {
	// Auth
	token := r.URL.Query().Get("token")
	var user *User
	if token != "" {
		claims, err := ValidateJWT(token)
		if err == nil {
			user = m.getUserFromClaims(claims)
		}
	}

	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Subscriber upgrade error: %v", err))
		return
	}

	sub := &Subscriber{Conn: conn, User: user}

	m.mutex.Lock()
	m.subscribers[conn] = sub
	m.mutex.Unlock()
	// m.AddLog("INFO", "Subscriber connected")

	defer func() {
		m.mutex.Lock()
		delete(m.subscribers, conn)
		m.mutex.Unlock()
		conn.Close()
		// m.AddLog("INFO", "Subscriber disconnected")
	}()

	for {
		// Read messages from subscriber (e.g. actions)
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// TODO: Handle actions from subscriber via WS if needed
	}
}

func serveWorker(m *Manager, w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Worker upgrade error: %v", err))
		return
	}

	worker := &WorkerClient{
		Conn:      conn,
		Connected: time.Now(),
	}

	m.mutex.Lock()
	m.workers = append(m.workers, worker)
	m.mutex.Unlock()
	m.AddLog("INFO", "New Worker connected (Competing Consumer)")

	// Keep alive / Read loop (to detect close)
	defer func() {
		m.removeWorker(worker)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Handle API requests from Worker
		var req map[string]interface{}
		if err := json.Unmarshal(message, &req); err == nil {
			m.dispatchAPIRequest(req)
		}
	}
}

func (m *Manager) dispatchAPIRequest(req map[string]interface{}) {
	// 1. Determine Target Bot ID
	var targetID string

	// Extract auto_recall
	autoRecall := 0
	if ar, ok := req["auto_recall"]; ok {
		switch v := ar.(type) {
		case float64:
			autoRecall = int(v)
		case int:
			autoRecall = v
		}
		delete(req, "auto_recall")
	}

	// Ensure echo is present if auto_recall is used
	if autoRecall > 0 {
		if _, ok := req["echo"]; !ok {
			req["echo"] = fmt.Sprintf("api_%d", time.Now().UnixNano())
		}
	}

	// Check top-level "self_id" (Best practice for routing)
	if id, ok := req["self_id"]; ok {
		targetID = fmt.Sprintf("%v", id)
	}

	// Fallback: Check "params.self_id" (Some implementations put it here)
	if targetID == "" {
		if params, ok := req["params"].(map[string]interface{}); ok {
			if id, ok := params["self_id"]; ok {
				targetID = fmt.Sprintf("%v", id)
			}
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var targetBot *BotClient

	if targetID != "" {
		if bot, ok := m.bots[targetID]; ok {
			targetBot = bot
		}
	}

	// 2. If no specific bot or not found, pick any (First one)
	if targetBot == nil {
		if len(m.bots) > 0 {
			// Just pick the first one
			for _, bot := range m.bots {
				targetBot = bot
				break
			}
		}
	}

	// 3. Send to Bot
	if targetBot != nil {
		if autoRecall > 0 {
			// Validate Auto Recall Delay
			if autoRecall > 120 {
				// Check if it's a Guild/Channel Bot
				// Assuming Platform string contains "guild" or "channel" or "qqguild"
				isGuildBot := false
				if strings.Contains(strings.ToLower(targetBot.Platform), "guild") ||
					strings.Contains(strings.ToLower(targetBot.Platform), "channel") {
					isGuildBot = true
				}

				if isGuildBot {
					autoRecall = 120
					m.AddLog("WARN", fmt.Sprintf("Auto Recall delay capped to 120s for Guild Bot %s", targetBot.SelfID))
				}
			}

			m.AutoRecallMutex.Lock()
			m.AutoRecallMap[getString(req, "echo")] = AutoRecallTask{
				Delay: autoRecall,
				BotID: targetBot.SelfID,
			}
			m.AutoRecallMutex.Unlock()
		}

		targetBot.Mutex.Lock()
		err := targetBot.Conn.WriteJSON(req)

		// Update Sent Count (Session)
		if action, ok := req["action"].(string); ok && strings.HasPrefix(action, "send_") {
			targetBot.SentCount++
		}

		targetBot.Mutex.Unlock()

		if err != nil {
			m.AddLog("ERROR", fmt.Sprintf("Failed to send API to bot %s: %v", targetBot.SelfID, err))
		} else {
			// Update Global Sent Stats (Persistent)
			if action, ok := req["action"].(string); ok && strings.HasPrefix(action, "send_") {
				m.statsMutex.Lock()
				m.SentMessages++
				if m.BotStatsSent == nil {
					m.BotStatsSent = make(map[string]int64)
				}
				m.BotStatsSent[targetBot.SelfID]++
				m.statsMutex.Unlock()

				if m.rdb != nil {
					ctx := context.Background()
					m.rdb.Incr(ctx, "stats:msg:sent")
				}
			}
			// Enable this log for debugging
			m.AddLog("DEBUG", fmt.Sprintf("Forwarded API to bot %s: %v | Params: %v", targetBot.SelfID, req["action"], req["params"]))
		}
	} else {
		m.AddLog("WARN", fmt.Sprintf("No bot available to handle API request. TargetID: %s", targetID))
	}
}

func (m *Manager) recordStats(botID string, msg map[string]interface{}) {
	postType, ok := msg["post_type"].(string)
	if !ok || postType != "message" {
		return
	}

	// Parse User
	var userID int64
	var userName string
	if uidVal, ok := msg["user_id"]; ok {
		switch v := uidVal.(type) {
		case float64:
			userID = int64(v)
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				userID = parsed
			}
		}
	}
	if sender, ok := msg["sender"].(map[string]interface{}); ok {
		if card, ok := sender["card"].(string); ok && card != "" {
			userName = card
		} else if nick, ok := sender["nickname"].(string); ok {
			userName = nick
		}
	}
	if userName == "" {
		userName = fmt.Sprintf("%d", userID)
	}

	// Parse Group
	var groupID int64
	var groupName string
	if gidVal, ok := msg["group_id"]; ok {
		switch v := gidVal.(type) {
		case float64:
			groupID = int64(v)
		case int64:
			groupID = v
		case int:
			groupID = int64(v)
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				groupID = parsed
			}
		}
	}
	if gn, ok := msg["group_name"].(string); ok && gn != "" {
		groupName = gn
	}

	// Deduplication Logic
	isDuplicate := false
	if mid, ok := msg["message_id"]; ok {
		msgID := fmt.Sprintf("%v", mid)
		// Composite key to avoid collisions if IDs are not globally unique
		key := msgID
		if groupID != 0 {
			key = fmt.Sprintf("g:%d:%s", groupID, msgID)
		} else if userID != 0 {
			key = fmt.Sprintf("u:%d:%s", userID, msgID)
		}

		m.DedupMutex.Lock()
		if ts, exists := m.DedupCache[key]; exists && time.Now().Unix()-ts < 10 {
			isDuplicate = true
		} else {
			m.DedupCache[key] = time.Now().Unix()
			// Periodic cleanup (simple reset if too large)
			if len(m.DedupCache) > 10000 {
				m.DedupCache = make(map[string]int64)
				m.DedupCache[key] = time.Now().Unix()
			}
		}
		m.DedupMutex.Unlock()
	}

	// Update Stats
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// Check for daily reset
	today := time.Now().Format("2006-01-02")
	if m.LastResetDate != today {
		m.UserStatsToday = make(map[int64]int64)
		m.GroupStatsToday = make(map[int64]int64)
		m.BotStatsToday = make(map[string]int64)
		// Reset granular daily stats
		if m.BotDetailedStats != nil {
			for _, detail := range m.BotDetailedStats {
				detail.UserStatsToday = make(map[int64]int64)
				detail.GroupStatsToday = make(map[int64]int64)
			}
		}
		m.LastResetDate = today
		// Optional: We could save the "yesterday" stats to history here if we wanted
	}

	if !isDuplicate {
		m.TotalMessages++
	}
	if m.BotStats == nil {
		m.BotStats = make(map[string]int64)
	}
	if m.BotStatsToday == nil {
		m.BotStatsToday = make(map[string]int64)
	}
	m.BotStats[botID]++
	m.BotStatsToday[botID]++

	// Ensure Detail Exists
	if m.BotDetailedStats == nil {
		m.BotDetailedStats = make(map[string]*BotStatDetail)
	}
	detail, exists := m.BotDetailedStats[botID]
	if !exists {
		detail = &BotStatDetail{
			UserStats:       make(map[int64]int64),
			GroupStats:      make(map[int64]int64),
			UserStatsToday:  make(map[int64]int64),
			GroupStatsToday: make(map[int64]int64),
		}
		m.BotDetailedStats[botID] = detail
	} else {
		// Ensure maps are not nil (compatibility)
		if detail.UserStats == nil {
			detail.UserStats = make(map[int64]int64)
		}
		if detail.GroupStats == nil {
			detail.GroupStats = make(map[int64]int64)
		}
		if detail.UserStatsToday == nil {
			detail.UserStatsToday = make(map[int64]int64)
		}
		if detail.GroupStatsToday == nil {
			detail.GroupStatsToday = make(map[int64]int64)
		}
	}

	if !isDuplicate && m.rdb != nil {
		ctx := context.Background()
		m.rdb.Incr(ctx, "stats:msg:total")
	}

	if userID != 0 {
		// Exclude self (Bot) from User Stats (Dragon King)
		isSelf := false
		if selfIDInt, err := strconv.ParseInt(botID, 10, 64); err == nil && selfIDInt == userID {
			isSelf = true
		}

		if !isDuplicate && !isSelf {
			m.UserStats[userID]++
			m.UserStatsToday[userID]++
			// Granular
			detail.UserStats[userID]++
			detail.UserStatsToday[userID]++
		}

		if !isSelf {
			m.UserNames[userID] = userName
			if !isDuplicate && m.rdb != nil {
				m.rdb.HIncrBy(context.Background(), "stats:user", fmt.Sprintf("%d", userID), 1)
			}
		}
	}
	if groupID != 0 {
		if !isDuplicate {
			m.GroupStats[groupID]++
			m.GroupStatsToday[groupID]++
			// Granular
			detail.GroupStats[groupID]++
			detail.GroupStatsToday[groupID]++
		}

		if groupName != "" {
			m.GroupNames[groupID] = groupName
		}
		if !isDuplicate && m.rdb != nil {
			m.rdb.HIncrBy(context.Background(), "stats:group", fmt.Sprintf("%d", groupID), 1)
		}
	}

	// --- Session Tracking ---
	go m.updateSession(botID, msg, userID, userName, groupID, groupName)
}

func (m *Manager) updateSession(botID string, msg map[string]interface{}, userID int64, userName string, groupID int64, groupName string) {
	messageType, _ := msg["message_type"].(string)

	// Extract Message ID
	var msgID string
	if idVal, ok := msg["message_id"]; ok {
		switch v := idVal.(type) {
		case string:
			msgID = v
		case float64:
			msgID = fmt.Sprintf("%.0f", v)
		case int64:
			msgID = fmt.Sprintf("%d", v)
		default:
			msgID = fmt.Sprintf("%v", v)
		}
	}

	m.SessionMutex.Lock()
	defer m.SessionMutex.Unlock()

	now := time.Now().Unix()

	// 1. Group Session
	if groupID != 0 {
		key := fmt.Sprintf("%s:group:%d", botID, groupID)
		if s, ok := m.Sessions[key]; ok {
			s.LastActive = now
			s.LastMsgID = msgID
			s.LastMsgTime = now
			if groupName != "" {
				s.Name = groupName
			}
		} else {
			name := groupName
			if name == "" {
				name = fmt.Sprintf("Group %d", groupID)
			}
			m.Sessions[key] = &ContactSession{
				ID:          fmt.Sprintf("%d", groupID),
				Name:        name,
				Type:        "group",
				BotID:       botID,
				LastActive:  now,
				LastMsgID:   msgID,
				LastMsgTime: now,
				ActiveBots:  make(map[string]BotInfo),
			}
		}

		// Update ActiveBots for ALL sessions of this GroupID
		targetGroupIDStr := fmt.Sprintf("%d", groupID)
		m.mutex.RLock()
		currentBotClient, exists := m.bots[botID]
		m.mutex.RUnlock()

		if exists {
			info := BotInfo{
				ID:       botID,
				Nickname: currentBotClient.Nickname,
				Platform: currentBotClient.Platform,
			}
			for _, s := range m.Sessions {
				if s.Type == "group" && s.ID == targetGroupIDStr {
					if s.ActiveBots == nil {
						s.ActiveBots = make(map[string]BotInfo)
					}
					s.ActiveBots[botID] = info
				}
			}
		}
	}

	// 2. Guild/Channel Session
	if messageType == "guild" {
		guildID, _ := msg["guild_id"].(string)
		channelID, _ := msg["channel_id"].(string)

		if channelID != "" {
			key := fmt.Sprintf("%s:guild:%s", botID, channelID)
			// Try to get guild name from message if available (not standard OneBot but useful)
			guildName, _ := msg["guild_name"].(string)
			channelName, _ := msg["channel_name"].(string)

			name := channelName
			if name == "" {
				name = fmt.Sprintf("Channel %s", channelID)
			}
			if guildName != "" {
				name = fmt.Sprintf("[%s] %s", guildName, name)
			} else if guildID != "" {
				name = fmt.Sprintf("[%s] %s", guildID, name)
			}

			if s, ok := m.Sessions[key]; ok {
				s.LastActive = now
				s.LastMsgID = msgID
				s.LastMsgTime = now
				if channelName != "" {
					s.Name = name // Update name if we got better info
				}
			} else {
				m.Sessions[key] = &ContactSession{
					ID:          channelID,
					Name:        name,
					Type:        "guild",
					BotID:       botID,
					GuildID:     guildID,
					LastActive:  now,
					LastMsgID:   msgID,
					LastMsgTime: now,
				}
			}
		}
	}

	// 3. User Session (Private)
	// Only track if it's explicitly a private message OR we want to track individual users in groups too?
	// Usually "contacts" implies people we can DM.
	// If message_type is private, track it.
	if messageType == "private" && userID != 0 {
		key := fmt.Sprintf("%s:private:%d", botID, userID)
		if s, ok := m.Sessions[key]; ok {
			s.LastActive = now
			s.LastMsgID = msgID
			s.LastMsgTime = now
			if userName != "" {
				s.Name = userName
			}
		} else {
			name := userName
			if name == "" {
				name = fmt.Sprintf("User %d", userID)
			}
			m.Sessions[key] = &ContactSession{
				ID:          fmt.Sprintf("%d", userID),
				Name:        name,
				Type:        "private",
				BotID:       botID,
				LastActive:  now,
				LastMsgID:   msgID,
				LastMsgTime: now,
			}
		}
	}
}

// Auth Handlers & Logic

type User struct {
	Username  string          `json:"username"`
	Role      string          `json:"role"` // "admin" or "user"
	OwnedBots map[string]bool `json:"-"`
}

func (m *Manager) authenticate(r *http.Request) *User {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}
	tokenStr := parts[1]

	// Validate JWT
	claims, err := ValidateJWT(tokenStr)
	if err != nil {
		return nil
	}

	return m.getUserFromClaims(claims)
}

func (m *Manager) getUserFromClaims(claims *UserClaims) *User {
	if claims.Username == "admin" {
		return &User{Username: "admin", Role: "admin"}
	}

	if m.rdb != nil {
		ctx := context.Background()
		ownedBots := make(map[string]bool)
		bots, _ := m.rdb.SMembers(ctx, fmt.Sprintf("auth:user:%s:bots", claims.Username)).Result()
		for _, b := range bots {
			ownedBots[b] = true
		}
		return &User{Username: claims.Username, Role: claims.Role, OwnedBots: ownedBots}
	}

	return &User{Username: claims.Username, Role: claims.Role, OwnedBots: map[string]bool{}}
}

func (m *Manager) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&creds)

	// Validate
	valid := false
	role := "user"
	passwordOverridden := false

	// Check Redis first for any user (including admin)
	if m.rdb != nil {
		ctx := context.Background()
		storedPwd, err := m.rdb.HGet(ctx, fmt.Sprintf("auth:user:%s:pwd", creds.Username), "password").Result()
		if err == nil && storedPwd != "" {
			passwordOverridden = true
			if storedPwd == creds.Password {
				valid = true
				if creds.Username == "admin" {
					role = "admin"
				}
			}
		}
	}

	if !passwordOverridden {
		if creds.Username == "admin" && creds.Password == "admin888" {
			valid = true
			role = "admin"
		} else if creds.Username == "test" && creds.Password == "test" {
			valid = true
		}
	}

	if !valid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := GenerateJWT(creds.Username, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
		"role":  role,
	})
}

func (m *Manager) handleGetContacts(w http.ResponseWriter, r *http.Request) {
	// Simple auth check (optional, but good practice)
	// user := m.authenticate(r)
	// if user == nil { http.Error(...) }

	m.SessionMutex.RLock()
	defer m.SessionMutex.RUnlock()

	sessions := make([]*ContactSession, 0, len(m.Sessions))
	for _, s := range m.Sessions {
		sessions = append(sessions, s)
	}

	// Sort by LastActive Desc
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive > sessions[j].LastActive
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// Magic Link Handlers

func (m *Manager) handleGenerateMagicLink(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil || user.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// Generate Magic Token
	magicToken := GenerateRandomToken(32) // 32 bytes -> 64 hex chars

	// Store in Redis with 5m TTL
	if m.rdb == nil {
		http.Error(w, "Redis required for magic links", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("auth:magic:%s", magicToken)
	err := m.rdb.Set(ctx, key, req.Username, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}

	// Construct Link
	// Use Referer or Host header to build absolute URL
	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	link := fmt.Sprintf("%s://%s/?magic_token=%s", scheme, host, magicToken)

	json.NewEncoder(w).Encode(map[string]string{
		"url":   link,
		"token": magicToken,
	})
}

func (m *Manager) handleMagicLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if m.rdb == nil {
		http.Error(w, "Redis unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("auth:magic:%s", req.Token)

	username, err := m.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}

	// Token valid! Delete it (One-time use)
	m.rdb.Del(ctx, key)

	// Determine role
	role := "user"
	if username == "admin" {
		role = "admin"
	}

	// Generate JWT
	jwtToken, err := GenerateJWT(username, role)
	if err != nil {
		http.Error(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token":    jwtToken,
		"role":     role,
		"username": username,
	})
}

func (m *Manager) handleMe(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (m *Manager) handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password cannot be empty", http.StatusBadRequest)
		return
	}

	// 1. Check if Redis is enabled
	if m.rdb == nil {
		// If using hardcoded admin, we can't really update it persistently
		// But for user experience, we can return error or fake it if it matches hardcoded
		if user.Username == "admin" && req.OldPassword == "admin888" {
			http.Error(w, "Cannot update password in default mode (Redis required)", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "Persistence service unavailable", http.StatusServiceUnavailable)
		return
	}

	// 2. Verify Old Password
	ctx := context.Background()
	storedPwd, err := m.rdb.HGet(ctx, fmt.Sprintf("auth:user:%s:pwd", user.Username), "password").Result()

	// If not found in Redis (e.g. admin first time), maybe we allow if it matches default "admin888"?
	if err != nil {
		if user.Username == "admin" && req.OldPassword == "admin888" {
			// Allow proceeding to set new password in Redis
		} else {
			http.Error(w, "Invalid old password", http.StatusUnauthorized)
			return
		}
	} else {
		if storedPwd != req.OldPassword {
			http.Error(w, "Invalid old password", http.StatusUnauthorized)
			return
		}
	}

	// 3. Update Password
	if err := m.rdb.HSet(ctx, fmt.Sprintf("auth:user:%s:pwd", user.Username), "password", req.NewPassword).Err(); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *Manager) handleAssignBot(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil || user.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		BotID    string `json:"bot_id"`
		Action   string `json:"action"` // "add" or "remove"
	}
	json.NewDecoder(r.Body).Decode(&req)

	if m.rdb == nil {
		http.Error(w, "Redis required for persistence", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("auth:user:%s:bots", req.Username)

	if req.Action == "remove" {
		m.rdb.SRem(ctx, key, req.BotID)
		m.rdb.HDel(ctx, "auth:bot_owners", req.BotID)
	} else {
		m.rdb.SAdd(ctx, key, req.BotID)
		m.rdb.HSet(ctx, "auth:bot_owners", req.BotID, req.Username)

		// Ensure user exists in user list
		m.rdb.SAdd(ctx, "auth:users", req.Username)
		// Set default pwd for new user if not exists
		m.rdb.HSetNX(ctx, fmt.Sprintf("auth:user:%s:pwd", req.Username), "password", "123456")
	}

	w.WriteHeader(http.StatusOK)
}

// API Handlers

func (m *Manager) handleGetBots(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get all known bot IDs
	var allBotIDs []string
	if m.rdb != nil {
		ctx := context.Background()
		ids, err := m.rdb.SMembers(ctx, "bots:all").Result()
		if err == nil {
			allBotIDs = ids
		}
	}

	// Fallback or merge with current memory bots (in case Redis missed something or is down)
	// Use a map to dedup
	botIDSet := make(map[string]bool)
	for _, id := range allBotIDs {
		botIDSet[id] = true
	}
	for id := range m.bots {
		botIDSet[id] = true
	}

	botList := make([]map[string]interface{}, 0)
	for id := range botIDSet {
		// Filter by ownership
		if user.Role != "admin" {
			if !user.OwnedBots[id] {
				continue
			}
		}

		// Check if Online
		client, isOnline := m.bots[id]

		// Fetch Info
		var info map[string]interface{}
		info = make(map[string]interface{})

		info["self_id"] = id
		info["is_alive"] = isOnline
		info["platform"] = "QQ" // Default to QQ

		// Owner Info
		owner := "admin" // Default or None
		if m.rdb != nil {
			ctx := context.Background()
			o, _ := m.rdb.HGet(ctx, "auth:bot_owners", id).Result()
			if o != "" {
				owner = o
			}
			// Get platform from Redis if available (fallback or override)
			p, _ := m.rdb.HGet(ctx, fmt.Sprintf("bot:info:%s", id), "platform").Result()
			if p != "" {
				info["platform"] = p
			}
		}
		info["owner"] = owner

		if isOnline {
			// Use Memory Data
			info["remote_addr"] = client.Conn.RemoteAddr().String()
			info["connected"] = client.Connected.Format(time.RFC3339)
			info["nickname"] = client.Nickname
			info["group_count"] = client.GroupCount
			info["friend_count"] = client.FriendCount
			if client.Platform != "" {
				info["platform"] = client.Platform
			}
		} else {
			// Use Redis Data
			if m.rdb != nil {
				ctx := context.Background()
				redisInfo, _ := m.rdb.HGetAll(ctx, fmt.Sprintf("bot:info:%s", id)).Result()

				info["remote_addr"] = redisInfo["remote_addr"]
				// Use disconnected_at if available, otherwise connected_at
				if disconnectedAt, ok := redisInfo["disconnected_at"]; ok {
					info["connected"] = disconnectedAt // Show when it went offline? Or add a separate field?
					info["disconnected_at"] = disconnectedAt
				} else {
					info["connected"] = redisInfo["connected_at"]
				}

				info["nickname"] = redisInfo["nickname"]

				gc, _ := strconv.Atoi(redisInfo["group_count"])
				info["group_count"] = gc

				fc, _ := strconv.Atoi(redisInfo["friend_count"])
				info["friend_count"] = fc

				if p, ok := redisInfo["platform"]; ok && p != "" {
					info["platform"] = p
				}
			}
		}

		// Inject Stats
		m.statsMutex.RLock()
		recvCount := int64(0)
		sentCount := int64(0)

		if m.BotStats != nil {
			recvCount = m.BotStats[id]
		}
		if m.BotStatsSent != nil {
			sentCount = m.BotStatsSent[id]
		}

		info["recv_count"] = recvCount
		info["sent_count"] = sentCount
		info["msg_count"] = recvCount + sentCount

		if m.BotStatsToday != nil {
			info["msg_count_today"] = m.BotStatsToday[id]
		} else {
			info["msg_count_today"] = 0
		}
		m.statsMutex.RUnlock()

		// Generate Avatar URL based on Platform
		// QQ / Android / Guild / Tencent: Use QQ Avatar
		platform := fmt.Sprintf("%v", info["platform"])
		if platform == "QQ" || platform == "Android" || platform == "Guild" || platform == "Tencent" {
			info["avatar_url"] = fmt.Sprintf("http://q1.qlogo.cn/g?b=qq&nk=%s&s=640", id)
		} else if platform == "DingTalk" {
			info["avatar_url"] = "https://img.alicdn.com/tfs/TB19Z7Kj4z1gK0jSZSgXXavwpXa-1024-1024.png"
		} else if platform == "Lark" {
			info["avatar_url"] = "https://sf3-cn.feishucdn.com/obj/eden-cn/ul_j_ul/feishu-logo.png"
		} else if platform == "Telegram" {
			info["avatar_url"] = "https://telegram.org/img/t_logo.png"
		} else {
			// Fallback
			info["avatar_url"] = "https://ui-avatars.com/api/?name=" + fmt.Sprintf("%v", info["nickname"])
		}

		botList = append(botList, info)
	}

	// Sort: 1. Nickname, 2. SelfID (Numerically)
	sort.Slice(botList, func(i, j int) bool {
		n1 := ""
		if v, ok := botList[i]["nickname"].(string); ok {
			n1 = v
		}
		n2 := ""
		if v, ok := botList[j]["nickname"].(string); ok {
			n2 = v
		}

		if n1 != n2 {
			return n1 < n2
		}

		id1 := ""
		if v, ok := botList[i]["self_id"].(string); ok {
			id1 = v
		}
		id2 := ""
		if v, ok := botList[j]["self_id"].(string); ok {
			id2 = v
		}

		// Sort numerically by length then string
		if len(id1) != len(id2) {
			return len(id1) < len(id2)
		}
		return id1 < id2
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botList)
}

func (m *Manager) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	workerList := make([]map[string]interface{}, 0)
	for _, w := range m.workers {
		workerList = append(workerList, map[string]interface{}{
			"remote_addr":   w.Conn.RemoteAddr().String(),
			"connected":     w.Connected.Format(time.RFC3339),
			"status":        "active",
			"handled_count": w.HandledCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workerList)
}

func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	botID := r.URL.Query().Get("bot_id")
	logs := m.GetLogs()

	if botID != "" {
		filtered := make([]LogEntry, 0)
		for _, l := range logs {
			if botID == "system" {
				if l.BotID == "" {
					filtered = append(filtered, l)
				}
			} else {
				if l.BotID == botID {
					filtered = append(filtered, l)
				}
			}
		}
		logs = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (m *Manager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var runtimeMem runtime.MemStats
	runtime.ReadMemStats(&runtimeMem)

	m.mutex.RLock()
	botCount := len(m.bots)
	subCount := len(m.subscribers)
	m.mutex.RUnlock()

	m.statsMutex.RLock()
	totalMsgs := m.TotalMessages
	sentMsgs := m.SentMessages
	activeGroups := len(m.GroupStats)
	activeUsers := len(m.UserStats)

	activeGroupsToday := 0
	activeUsersToday := 0
	today := time.Now().Format("2006-01-02")
	if m.LastResetDate == today {
		activeGroupsToday = len(m.GroupStatsToday)
		activeUsersToday = len(m.UserStatsToday)
	}

	botTotal := len(m.BotStats)
	m.statsMutex.RUnlock()

	// System Info
	var cpuModel string
	var cpuFreq float64
	var memTotal uint64

	cInfos, err := cpu.Info()
	if err == nil && len(cInfos) > 0 {
		cpuModel = cInfos[0].ModelName
		cpuFreq = cInfos[0].Mhz
	}
	physicalCores, _ := cpu.Counts(false)
	logicalCores, _ := cpu.Counts(true)

	vMem, err := mem.VirtualMemory()
	if err == nil {
		memTotal = vMem.Total
	}

	m.HistoryMutex.RLock()
	currentCPU := m.CurrentCPU
	cpuTrend := make([]float64, len(m.CPUTrend))
	copy(cpuTrend, m.CPUTrend)
	memTrend := make([]float64, len(m.MemTrend))
	copy(memTrend, m.MemTrend)
	msgTrend := make([]float64, len(m.MsgTrend))
	copy(msgTrend, m.MsgTrend)
	sentTrend := make([]float64, len(m.SentTrend))
	copy(sentTrend, m.SentTrend)
	recvTrend := make([]float64, len(m.RecvTrend))
	copy(recvTrend, m.RecvTrend)
	m.HistoryMutex.RUnlock()

	// Calculate Bot Stats
	if botTotal < botCount {
		botTotal = botCount // Should not happen if logic is correct, but safe guard
	}

	stats := map[string]interface{}{
		"cpu_usage":           currentCPU,
		"cpu_model":           cpuModel,
		"cpu_cores_physical":  physicalCores,
		"cpu_cores_logical":   logicalCores,
		"cpu_freq":            cpuFreq,
		"goroutines":          runtime.NumGoroutine(),
		"memory_alloc":        runtimeMem.Alloc,
		"memory_sys":          runtimeMem.Sys,
		"memory_total":        memTotal,
		"uptime":              "N/A", // TODO: Implement uptime
		"bot_count":           botCount,
		"bot_count_total":     botTotal,
		"bot_count_offline":   botTotal - botCount,
		"subscriber_count":    subCount,
		"message_count":       totalMsgs,
		"sent_message_count":  sentMsgs,
		"active_groups":       activeGroups,
		"active_users":        activeUsers,
		"active_groups_today": activeGroupsToday,
		"active_users_today":  activeUsersToday,
		"cpu_trend":           cpuTrend,
		"mem_trend":           memTrend,
		"msg_trend":           msgTrend,
		"sent_trend":          sentTrend,
		"recv_trend":          recvTrend,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (m *Manager) handleGetChatStats(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	var resp map[string]interface{}
	today := time.Now().Format("2006-01-02")

	// Prepare effective today stats (handle stale data if no msg received yet today)
	userStatsToday := m.UserStatsToday
	groupStatsToday := m.GroupStatsToday
	if m.LastResetDate != today {
		userStatsToday = make(map[int64]int64)
		groupStatsToday = make(map[int64]int64)
	}

	if user.Role == "admin" {
		// Admin sees global stats
		resp = map[string]interface{}{
			"user_stats":        m.UserStats,
			"group_stats":       m.GroupStats,
			"user_stats_today":  userStatsToday,
			"group_stats_today": groupStatsToday,
			"last_reset_date":   today,
			"user_names":        m.UserNames,
			"group_names":       m.GroupNames,
		}
	} else {
		// User sees aggregated stats from their owned bots
		aggUserStats := make(map[int64]int64)
		aggGroupStats := make(map[int64]int64)
		aggUserStatsToday := make(map[int64]int64)
		aggGroupStatsToday := make(map[int64]int64)

		if m.BotDetailedStats != nil {
			for botID, detail := range m.BotDetailedStats {
				if user.OwnedBots[botID] {
					// Aggregate
					for k, v := range detail.UserStats {
						aggUserStats[k] += v
					}
					for k, v := range detail.GroupStats {
						aggGroupStats[k] += v
					}

					// Only aggregate today stats if date matches
					if m.LastResetDate == today {
						for k, v := range detail.UserStatsToday {
							aggUserStatsToday[k] += v
						}
						for k, v := range detail.GroupStatsToday {
							aggGroupStatsToday[k] += v
						}
					}
				}
			}
		}

		resp = map[string]interface{}{
			"user_stats":        aggUserStats,
			"group_stats":       aggGroupStats,
			"user_stats_today":  aggUserStatsToday,
			"group_stats_today": aggGroupStatsToday,
			"last_reset_date":   today,
			"user_names":        m.UserNames, // Names are global cache, safe to share
			"group_names":       m.GroupNames,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (m *Manager) handleSmartAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := m.authenticate(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	action, _ := req["action"].(string)
	params, _ := req["params"].(map[string]interface{})

	// Only handle smart group messages for now
	if action == "send_group_msg" {
		botID := getString(req, "self_id") // Optional target bot
		groupID := getString(params, "group_id")
		msgContent := getString(params, "message")

		if groupID == "" || msgContent == "" {
			http.Error(w, "Missing group_id or message", http.StatusBadRequest)
			return
		}

		m.SessionMutex.RLock()
		// Try to find a suitable bot session for this group
		// Priority:
		// 1. Specified BotID with valid MsgID
		// 2. Any Bot with valid MsgID
		// 3. Specified BotID (requires WakeUp)
		// 4. Any Bot (Direct send)

		// Find the session for this bot/group
		var targetSession *ContactSession
		var helperBotID string

		// If botID is specified, find its session
		if botID != "" {
			key := fmt.Sprintf("%s:group:%s", botID, groupID)
			if s, ok := m.Sessions[key]; ok {
				targetSession = s
			}
		}

		// Check if we need to wake up
		needsWakeUp := false
		if targetSession != nil {
			// Check if message_id is valid (within 290s to be safe)
			if time.Now().Unix()-targetSession.LastMsgTime > 290 {
				needsWakeUp = true
			}
		}

		if needsWakeUp && targetSession != nil {
			// Find a helper bot in the SAME group
			for bID := range targetSession.ActiveBots {
				if bID != botID {
					// Found a helper!
					helperBotID = bID
					break
				}
			}

			if helperBotID != "" {
				// 1. Send WakeUp command via Helper Bot
				m.AddLog("INFO", fmt.Sprintf("SmartSend: Waking up Bot %s via Helper %s in Group %s", botID, helperBotID, groupID))

				// Get Target Bot Nickname
				targetNick := "Bot" // Default
				m.mutex.RLock()
				if tb, ok := m.bots[botID]; ok {
					targetNick = tb.Nickname
				}
				m.mutex.RUnlock()

				// Send WakeUp Message
				wakeUpMsg := fmt.Sprintf("@%s [WakeUp]", targetNick)
				// Or use CQ Code if needed: [CQ:at,qq=BOT_ID]
				// wakeUpMsg = fmt.Sprintf("[CQ:at,qq=%s] [WakeUp]", botID) // Better reliability

				wakeUpReq := map[string]interface{}{
					"action": "send_group_msg",
					"params": map[string]interface{}{
						"group_id": groupID,
						"message":  wakeUpMsg,
					},
					"self_id":     helperBotID,
					"auto_recall": 5, // Auto delete after 5s
				}

				// Send directly to helper
				m.dispatchAPIRequest(wakeUpReq)

				// 2. Wait a bit for the event to propagate (Hack/Simplification)
				// Ideally we should wait for the event, but for now a short sleep might work
				// since local network is fast.
				time.Sleep(2 * time.Second)

				// 3. Re-check session (it should be updated by the incoming message event)
				m.SessionMutex.RUnlock()           // Release lock to allow update
				time.Sleep(100 * time.Millisecond) // Yield
				m.SessionMutex.RLock()             // Re-acquire

				if s, ok := m.Sessions[fmt.Sprintf("%s:group:%s", botID, groupID)]; ok {
					targetSession = s
					// Update params with new message_id
					params["message_id"] = s.LastMsgID
					m.AddLog("INFO", fmt.Sprintf("SmartSend: WakeUp successful? Using MsgID: %s", s.LastMsgID))
				}
			} else {
				m.AddLog("WARN", fmt.Sprintf("SmartSend: No helper bot found for Group %s to wake up %s", groupID, botID))
			}
		}
		m.SessionMutex.RUnlock()

		// Inject message_id if available from session (even if not waking up, maybe it's still valid)
		if targetSession != nil && targetSession.LastMsgID != "" {
			// Only inject if not already present
			if _, ok := params["message_id"]; !ok {
				params["message_id"] = targetSession.LastMsgID
			}
		}
	}

	// Dispatch original request (with potentially updated params)
	// If it was a "Smart" send, we might have added message_id
	m.dispatchAPIRequest(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"detail": "Request dispatched (Smart Logic Applied)",
	})
}

func (m *Manager) handleAction(w http.ResponseWriter, r *http.Request) {
	user := m.authenticate(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BotID      string                 `json:"bot_id"`
		Action     string                 `json:"action"`
		Params     map[string]interface{} `json:"params"`
		AutoRecall int                    `json:"auto_recall"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Permission Check
	if user.Role != "admin" {
		if req.BotID == "" || !user.OwnedBots[req.BotID] {
			http.Error(w, "Forbidden: You do not own this bot", http.StatusForbidden)
			return
		}
	}

	if req.Action == "" {
		http.Error(w, "Action is required", http.StatusBadRequest)
		return
	}

	// Track Sent Messages
	if len(req.Action) > 5 && req.Action[:5] == "send_" {
		m.statsMutex.Lock()
		m.SentMessages++
		m.statsMutex.Unlock()
	}

	m.mutex.RLock()
	client, ok := m.bots[req.BotID]
	m.mutex.RUnlock()

	if !ok {
		// If no specific bot, maybe broadcast or pick first?
		// For now, fail
		if req.BotID == "" && len(m.bots) > 0 {
			// Pick first
			for _, c := range m.bots {
				client = c
				break
			}
		} else {
			http.Error(w, "Bot not found", http.StatusNotFound)
			return
		}
	}

	if req.Params == nil {
		req.Params = make(map[string]interface{})
	}

	// Construct OneBot Action Frame
	echo := fmt.Sprintf("api_%d", time.Now().UnixNano())
	actionFrame := map[string]interface{}{
		"action": req.Action,
		"params": req.Params,
		"echo":   echo,
	}

	// Register Auto Recall if needed
	if req.AutoRecall > 0 {
		m.AutoRecallMutex.Lock()
		m.AutoRecallMap[echo] = AutoRecallTask{
			Delay: req.AutoRecall,
			BotID: client.SelfID,
		}
		m.AutoRecallMutex.Unlock()
	}

	client.Mutex.Lock()
	err := client.Conn.WriteJSON(actionFrame)
	client.Mutex.Unlock()

	m.AddLog("DEBUG", fmt.Sprintf("Sent API to %s: %s (echo: %s)", req.BotID, req.Action, actionFrame["echo"]))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "echo": actionFrame["echo"].(string)})
}

type SystemStats struct {
	CPUUsage      float64              `json:"cpu_usage"`
	HostInfo      *host.InfoStat       `json:"host_info"`
	Processes     []ProcessInfo        `json:"processes"`
	DiskUsage     []*disk.UsageStat    `json:"disk_usage"`
	NetIO         []net.IOCountersStat `json:"net_io"`
	NetInterfaces []net.InterfaceStat  `json:"net_interfaces"`
}

type ProcessInfo struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"` // RSS in bytes
}

func (m *Manager) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Get total CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Get processes
	procs, err := process.Processes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var procInfos []ProcessInfo
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}

		c, err := p.CPUPercent()
		if err != nil {
			// Some processes might return error for CPU, but we still might want them if sorting by Mem?
			// For now, continue but maybe set to 0
			c = 0
		}

		m, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		// Always add, we will sort and slice later
		procInfos = append(procInfos, ProcessInfo{
			PID:    p.Pid,
			Name:   n,
			CPU:    c,
			Memory: m.RSS,
		})
	}

	// Sort by CPU desc
	sort.Slice(procInfos, func(i, j int) bool {
		return procInfos[i].CPU > procInfos[j].CPU
	})

	// Limit to top 10 (or 20 for better view)
	if len(procInfos) > 20 {
		procInfos = procInfos[:20]
	}

	// 3. Get Host Info
	hostInfo, _ := host.Info()
	if hostInfo == nil {
		hostInfo = &host.InfoStat{}
	}
	// Hardcode for display as requested
	hostInfo.OS = "linux"
	hostInfo.Platform = "alpine"
	hostInfo.PlatformVersion = "3.23.0"
	hostInfo.KernelVersion = "6.8.0-86-generic"

	// 4. Disk Usage
	var diskUsages []*disk.UsageStat
	parts, err := disk.Partitions(false)
	if err == nil {
		for _, part := range parts {
			u, err := disk.Usage(part.Mountpoint)
			if err == nil {
				diskUsages = append(diskUsages, u)
			}
		}
	}

	// 5. Net IO
	netIO, _ := net.IOCounters(false) // Total
	netInterfaces, _ := net.Interfaces()

	resp := SystemStats{
		CPUUsage:      cpuPercent[0],
		HostInfo:      hostInfo,
		Processes:     procInfos,
		DiskUsage:     diskUsages,
		NetIO:         netIO,
		NetInterfaces: netInterfaces,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
