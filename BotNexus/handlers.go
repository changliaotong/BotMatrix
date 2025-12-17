package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 日志条目
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// 添加日志
func (m *Manager) AddLog(level string, message string) {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()
	
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}
	
	m.logBuffer = append(m.logBuffer, entry)
	if len(m.logBuffer) > 1000 {
		m.logBuffer = m.logBuffer[len(m.logBuffer)-1000:]
	}
	
	// 同时打印到控制台
	log.Printf("[%s] %s", level, message)
}

// 启动Worker超时检测
func (m *Manager) StartWorkerTimeoutDetection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.checkWorkerTimeouts()
	}
}

// 启动Bot超时检测
func (m *Manager) StartBotTimeoutDetection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.checkBotTimeouts()
	}
}

// 启动定期统计保存
func (m *Manager) StartPeriodicStatsSave() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		m.saveStats()
	}
}

// 启动趋势收集
func (m *Manager) StartTrendCollection() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.collectTrends()
	}
}

// 检查Worker超时
func (m *Manager) checkWorkerTimeouts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	now := time.Now()
	activeWorkers := make([]*WorkerClient, 0)
	
	for _, worker := range m.workers {
		worker.Mutex.Lock()
		if now.Sub(worker.LastHeartbeat) < 2*time.Minute {
			activeWorkers = append(activeWorkers, worker)
		} else {
			// 超时Worker，关闭连接
			worker.Conn.Close()
			m.AddLog("WARN", fmt.Sprintf("[Worker Heartbeat Timeout] ID: %s, Timeout: %v", worker.ID, now.Sub(worker.LastHeartbeat)))
		}
		worker.Mutex.Unlock()
	}
	
	m.workers = activeWorkers
}

// 检查Bot超时
func (m *Manager) checkBotTimeouts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	now := time.Now()
	activeBots := make(map[string]*BotClient)
	
	for botID, bot := range m.bots {
		bot.Mutex.Lock()
		lastActive := bot.LastHeartbeat
		if lastActive.IsZero() {
			lastActive = bot.Connected
		}
		
		if now.Sub(lastActive) < 2*time.Minute {
			activeBots[botID] = bot
		} else {
			// 超时Bot，关闭连接
			bot.Conn.Close()
			m.AddLog("WARN", fmt.Sprintf("[Bot Heartbeat Timeout] ID: %s, Timeout: %v", botID, now.Sub(lastActive)))
		}
		bot.Mutex.Unlock()
	}
	
	m.bots = activeBots
}

// 收集趋势数据
func (m *Manager) collectTrends() {
	// 实现趋势收集逻辑
	m.AddLog("DEBUG", "[Trend Collection] Collecting trends...")
}

// WebSocket处理函数
func (m *Manager) handleBotWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Bot WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()
	
	m.AddLog("INFO", "[Bot WebSocket] New connection established")
	// 实现Bot WebSocket处理逻辑
}

func (m *Manager) handleWorkerWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Worker WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()
	
	m.AddLog("INFO", "[Worker WebSocket] New connection established")
	// 实现Worker WebSocket处理逻辑
}

func (m *Manager) handleSubscriberWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.AddLog("ERROR", fmt.Sprintf("Subscriber WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()
	
	m.AddLog("INFO", "[Subscriber WebSocket] New connection established")
	// 实现Subscriber WebSocket处理逻辑
}

// API处理函数
func (m *Manager) handleLogin(w http.ResponseWriter, r *http.Request) {
	// 实现登录逻辑
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login endpoint",
	})
}

func (m *Manager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// 实现获取统计信息逻辑
	w.Header().Set("Content-Type", "application/json")
	
	stats := m.GetConnectionStats()
	stats["total_messages"] = m.TotalMessages
	stats["sent_messages"] = m.SentMessages
	
	json.NewEncoder(w).Encode(stats)
}

func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	// 实现获取日志逻辑
	m.logMutex.RLock()
	defer m.logMutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.logBuffer)
}