package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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
}

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn      *websocket.Conn
	SelfID    string
	Nickname  string
	Connected time.Time
	Mutex     sync.Mutex
}

type WorkerClient struct {
	Conn      *websocket.Conn
	Mutex     sync.Mutex
	Connected time.Time
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

	// Redis
	rdb *redis.Client

	// Chat Stats
	statsMutex sync.RWMutex
	UserStats  map[int64]int64  `json:"user_stats"`  // UserID -> Count
	GroupStats map[int64]int64  `json:"group_stats"` // GroupID -> Count
	UserNames  map[int64]string `json:"-"`           // Cache names
	GroupNames map[int64]string `json:"-"`           // Cache names
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func NewManager() *Manager {
	m := &Manager{
		bots:        make(map[string]*BotClient),
		subscribers: make(map[*websocket.Conn]*Subscriber),
		workers:     make([]*WorkerClient, 0),
		UserStats:   make(map[int64]int64),
		GroupStats:  make(map[int64]int64),
		UserNames:   make(map[int64]string),
		GroupNames:  make(map[int64]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logBuffer: make([]LogEntry, 0, 200),
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

func (m *Manager) AddLog(level, message string) {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()

	entry := LogEntry{
		Time:    time.Now().Format("15:04:05"),
		Level:   level,
		Message: message,
	}

	if len(m.logBuffer) >= 200 {
		m.logBuffer = m.logBuffer[1:]
	}
	m.logBuffer = append(m.logBuffer, entry)
	log.Printf("[%s] %s", level, message)

	// Broadcast log to subscribers (wrapped in event)
	go m.broadcastToSubscribers(map[string]interface{}{
		"post_type": "log",
		"data":      entry,
	})
}

func (m *Manager) SaveStats() {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	data := map[string]interface{}{
		"user_stats":  m.UserStats,
		"group_stats": m.GroupStats,
		"user_names":  m.UserNames,
		"group_names": m.GroupNames,
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
		UserStats  map[int64]int64  `json:"user_stats"`
		GroupStats map[int64]int64  `json:"group_stats"`
		UserNames  map[int64]string `json:"user_names"`
		GroupNames map[int64]string `json:"group_names"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Printf("Error decoding stats: %v", err)
		return
	}

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if data.UserStats != nil {
		m.UserStats = data.UserStats
	}
	if data.GroupStats != nil {
		m.GroupStats = data.GroupStats
	}
	if data.UserNames != nil {
		m.UserNames = data.UserNames
	}
	if data.GroupNames != nil {
		m.GroupNames = data.GroupNames
	}
	log.Printf("Loaded stats: %d users, %d groups", len(m.UserStats), len(m.GroupStats))
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

	// 2. Web UI Server Mux
	uiMux := http.NewServeMux()

	// API Endpoints
	uiMux.HandleFunc("/api/bots", manager.handleGetBots)
	uiMux.HandleFunc("/api/logs", manager.handleGetLogs)
	uiMux.HandleFunc("/api/stats", manager.handleGetStats)
	uiMux.HandleFunc("/api/stats/chat", manager.handleGetChatStats)
	uiMux.HandleFunc("/api/action", manager.handleAction)

	// Auth API
	uiMux.HandleFunc("/api/login", manager.handleLogin)
	uiMux.HandleFunc("/api/me", manager.handleMe)
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
	// Only for message events (or maybe all events? usually all events need processing)
	// But heartbeats should probably be ignored or broadcast?
	// For now, we distribute everything.
	if len(m.workers) > 0 {
		// Round-Robin
		// Note: We need to upgrade lock to Write if we modify workerIndex?
		// But here we are in RLock. We can use atomic or just ignore race condition for index as it's not critical.
		// Or better: use a separate mutex for load balancing index, or just pick random?
		// Random is easier and stateless for this lock scope.
		// Let's use simple modification of index assuming single threaded dispatch or acceptable race.
		// Actually, let's just pick one.

		// To do strict Round-Robin safely:
		// We can't modify m.workerIndex under RLock.
		// Let's assume we want to avoid WLock for every message.
		// We can use a local counter or random.
		// Let's use Random for now, it's good enough for load balancing.
		// Or upgrade to Lock just for the index update? No, too heavy.
		// Let's use a try-lock or just iterate and find first available?
		// "Competition" means we just need to send to ONE worker.

		// Implementation: Send to m.workers[m.workerIndex % len]
		// We will cast RLock to Lock momentarily? No.
		// Let's just use a global atomic counter?
		// For simplicity in this pair programming session:
		// We'll iterate until we find a worker that accepts the write.

		// IMPROVED STRATEGY:
		// Since we are in RLock, we can't modify m.workerIndex.
		// Let's just pick based on time (random-ish) or a separate atomic counter.
		// But actually, we need to handle potential dead workers here too.

		// Simple approach: Try the first one, if fails, try next.
		// But we want LB.
		// Let's select one worker based on a hash of something or random.
		targetIndex := int(time.Now().UnixNano()) % len(m.workers)
		worker := m.workers[targetIndex]

		worker.Mutex.Lock()
		err := worker.Conn.WriteJSON(data)
		worker.Mutex.Unlock()

		if err != nil {
			// This worker is dead. We need to remove it.
			// But we are in RLock.
			// We can spawn a goroutine to remove it, and try sending to another worker?
			// For now, let's just trigger cleanup.
			go func(w *WorkerClient) {
				m.removeWorker(w)
			}(worker)

			// Fallback: try to send to ANY other worker
			for i, w := range m.workers {
				if i == targetIndex {
					continue
				}
				w.Mutex.Lock()
				e := w.Conn.WriteJSON(data)
				w.Mutex.Unlock()
				if e == nil {
					break // Sent successfully
				}
			}
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
	// Headers: X-Self-ID, X-Client-Role

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
	}

	m.mutex.Lock()
	m.bots[selfID] = client
	m.mutex.Unlock()

	m.AddLog("INFO", fmt.Sprintf("Client connected: %s (%s)", selfID, r.RemoteAddr))

	defer func() {
		m.mutex.Lock()
		delete(m.bots, selfID)
		m.mutex.Unlock()

		if m.rdb != nil {
			ctx := context.Background()
			m.rdb.SRem(ctx, "bots:online", selfID)
			m.rdb.Del(ctx, fmt.Sprintf("bot:info:%s", selfID))
		}

		conn.Close()
		m.AddLog("INFO", fmt.Sprintf("Client disconnected: %s", selfID))
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Try to parse message to update SelfID if it's a lifecycle event
		var msgMap map[string]interface{}
		if err := json.Unmarshal(message, &msgMap); err == nil {
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
						m.rdb.HSet(ctx, fmt.Sprintf("bot:info:%s", selfID), map[string]interface{}{
							"connected_at": client.Connected.Format(time.RFC3339),
							"remote_addr":  client.Conn.RemoteAddr().String(),
						})
					}

					// Trigger get_login_info to fetch nickname
					go func() {
						req := map[string]interface{}{
							"action": "get_login_info",
							"echo":   "internal_get_login_info",
						}
						client.Mutex.Lock()
						client.Conn.WriteJSON(req)
						client.Mutex.Unlock()
					}()
				}
			}

			// Update Nickname from get_login_info response or event
			if echo, ok := msgMap["echo"].(string); ok && echo == "internal_get_login_info" {
				if data, ok := msgMap["data"].(map[string]interface{}); ok {
					if nick, ok := data["nickname"].(string); ok {
						client.Nickname = nick
					}
				}
			}
			// Fallback: check lifecycle meta_event for nickname? (OneBot 11 doesn't specify it usually)

			// Broadcast to subscribers
			m.broadcastToSubscribers(msgMap)

			// Record Stats
			go m.recordStats(msgMap)
		}

		// Log heartbeat only occasionally or filter it
		// m.AddLog("DEBUG", fmt.Sprintf("Recv from %s: %s", selfID, string(message)))
	}
}

func serveSubscriber(m *Manager, w http.ResponseWriter, r *http.Request) {
	// Auth
	token := r.URL.Query().Get("token")
	var user *User
	if token != "" {
		// Re-use authenticate logic logic (simplified)
		// Since we don't have header, we construct a fake request or just check token manually
		// For simplicity, token IS username in this demo
		username := token
		if username == "admin" {
			user = &User{Username: "admin", Role: "admin"}
		} else if m.rdb != nil {
			ctx := context.Background()
			exists, _ := m.rdb.SIsMember(ctx, "auth:users", username).Result()
			if exists {
				// Load Owned Bots
				ownedBots := make(map[string]bool)
				bots, _ := m.rdb.SMembers(ctx, fmt.Sprintf("auth:user:%s:bots", username)).Result()
				for _, b := range bots {
					ownedBots[b] = true
				}
				user = &User{Username: username, Role: "user", OwnedBots: ownedBots}
			}
		} else if username == "test" {
			user = &User{Username: "test", Role: "user", OwnedBots: map[string]bool{}}
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
		targetBot.Mutex.Lock()
		err := targetBot.Conn.WriteJSON(req)
		targetBot.Mutex.Unlock()
		if err != nil {
			m.AddLog("ERROR", fmt.Sprintf("Failed to send API to bot %s: %v", targetBot.SelfID, err))
		} else {
			// m.AddLog("DEBUG", fmt.Sprintf("Forwarded API to bot %s: %v", targetBot.SelfID, req["action"]))
		}
	} else {
		m.AddLog("WARN", "No bot available to handle API request")
	}
}

func (m *Manager) recordStats(msg map[string]interface{}) {
	postType, ok := msg["post_type"].(string)
	if !ok || postType != "message" {
		return
	}

	// Parse User
	var userID int64
	var userName string
	if uid, ok := msg["user_id"].(float64); ok {
		userID = int64(uid)
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
	// var groupName string // msg usually doesn't have group_name
	if gid, ok := msg["group_id"].(float64); ok {
		groupID = int64(gid)
		// groupName = fmt.Sprintf("Group %d", groupID)
	}

	// Update Stats
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if m.rdb != nil {
		ctx := context.Background()
		m.rdb.Incr(ctx, "stats:msg:total")
	}

	if userID != 0 {
		m.UserStats[userID]++
		m.UserNames[userID] = userName
		if m.rdb != nil {
			m.rdb.HIncrBy(context.Background(), "stats:user", fmt.Sprintf("%d", userID), 1)
		}
	}
	if groupID != 0 {
		m.GroupStats[groupID]++
		// Only update group name if we really have it (future improvement)
		// m.GroupNames[groupID] = groupName
		if m.rdb != nil {
			m.rdb.HIncrBy(context.Background(), "stats:group", fmt.Sprintf("%d", groupID), 1)
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
	// Simple Bearer Token: "Bearer <username>" (In production use real tokens)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Format: "Bearer <username>"
	// For this demo, the token IS the username.
	// In production, check Redis for session: token -> username
	var username string
	fmt.Sscanf(authHeader, "Bearer %s", &username)

	if username == "" {
		return nil
	}

	// Check Redis or Hardcoded
	// Admin is hardcoded
	if username == "admin" {
		return &User{Username: "admin", Role: "admin"}
	}

	// Other users: Check Redis if they exist
	if m.rdb != nil {
		ctx := context.Background()
		exists, _ := m.rdb.SIsMember(ctx, "auth:users", username).Result()
		if exists {
			// Load Owned Bots
			ownedBots := make(map[string]bool)
			bots, _ := m.rdb.SMembers(ctx, fmt.Sprintf("auth:user:%s:bots", username)).Result()
			for _, b := range bots {
				ownedBots[b] = true
			}
			return &User{Username: username, Role: "user", OwnedBots: ownedBots}
		}
	} else {
		// Fallback for demo without Redis: Allow any user "test"
		if username == "test" {
			return &User{Username: "test", Role: "user", OwnedBots: map[string]bool{}}
		}
	}

	return nil
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

	if creds.Username == "admin" && creds.Password == "admin888" {
		valid = true
		role = "admin"
	} else if m.rdb != nil {
		// Check Redis
		// In prod: Hash password
		ctx := context.Background()
		storedPwd, _ := m.rdb.HGet(ctx, fmt.Sprintf("auth:user:%s:pwd", creds.Username), "password").Result()
		if storedPwd == creds.Password {
			valid = true
			// role = fetch from redis
		}
	} else if creds.Username == "test" && creds.Password == "test" {
		valid = true
	}

	if !valid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Return Token (Username as token for simplicity)
	json.NewEncoder(w).Encode(map[string]string{
		"token": creds.Username,
		"role":  role,
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

	botList := make([]map[string]interface{}, 0)
	for id, client := range m.bots {
		// Filter
		if user.Role != "admin" {
			if !user.OwnedBots[id] {
				continue
			}
		}

		owner := "admin" // Default or None
		if m.rdb != nil {
			ctx := context.Background()
			o, _ := m.rdb.HGet(ctx, "auth:bot_owners", id).Result()
			if o != "" {
				owner = o
			}
		}

		botList = append(botList, map[string]interface{}{
			"self_id":     id,
			"remote_addr": client.Conn.RemoteAddr().String(),
			"connected":   client.Connected.Format(time.RFC3339),
			"is_alive":    true,
			"nickname":    client.Nickname,
			"owner":       owner,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botList)
}

func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetLogs())
}

func (m *Manager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	m.mutex.RLock()
	botCount := len(m.bots)
	subCount := len(m.subscribers)
	m.mutex.RUnlock()

	stats := map[string]interface{}{
		"goroutines":       runtime.NumGoroutine(),
		"memory_alloc":     mem.Alloc,
		"memory_sys":       mem.Sys,
		"uptime":           "N/A", // TODO: Implement uptime
		"bot_count":        botCount,
		"subscriber_count": subCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (m *Manager) handleGetChatStats(w http.ResponseWriter, r *http.Request) {
	if m.authenticate(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	// Convert maps to lists for sorting (frontend can sort too, but let's return raw map for now)
	// Actually, let's return top 10 to save bandwidth if maps are huge.
	// For simplicity, returning full maps.

	resp := map[string]interface{}{
		"user_stats":  m.UserStats,
		"group_stats": m.GroupStats,
		"user_names":  m.UserNames,
		"group_names": m.GroupNames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
		BotID  string                 `json:"bot_id"`
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
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

	// Construct OneBot Action Frame
	actionFrame := map[string]interface{}{
		"action": req.Action,
		"params": req.Params,
		"echo":   fmt.Sprintf("api_%d", time.Now().UnixNano()),
	}

	client.Mutex.Lock()
	err := client.Conn.WriteJSON(actionFrame)
	client.Mutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "echo": actionFrame["echo"].(string)})
}
