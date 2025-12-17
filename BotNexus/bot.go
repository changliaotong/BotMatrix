package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn          *websocket.Conn
	SelfID        string
	Nickname      string
	GroupCount    int
	FriendCount   int
	Connected     time.Time
	Platform      string
	Mutex         sync.Mutex
	SentCount     int64     // Track sent messages per bot session
	RecvCount     int64     // Track received messages per bot session
	LastHeartbeat time.Time // Track last heartbeat for timeout detection
}

// Bot连接管理相关的方法

// 添加Bot连接统计
func (m *Manager) TrackBotConnection(botID string) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.TotalBotConnections++
	m.connectionStats.LastBotActivity[botID] = time.Now()
	m.AddLog("INFO", fmt.Sprintf("[Bot Connected] ID: %s, Total: %d", botID, m.connectionStats.TotalBotConnections))
}

// 记录Bot断开连接
func (m *Manager) TrackBotDisconnection(botID string, reason string, duration time.Duration) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()

	m.connectionStats.BotConnectionDurations[botID] = duration
	m.connectionStats.BotDisconnectReasons[reason]++
	delete(m.connectionStats.LastBotActivity, botID)

	m.AddLog("INFO", fmt.Sprintf("[Bot Disconnected] ID: %s, Reason: %s, Duration: %v", botID, reason, duration))
}

// Bot心跳处理
func (m *Manager) updateBotHeartbeat(botID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if bot, exists := m.bots[botID]; exists {
		bot.Mutex.Lock()
		bot.LastHeartbeat = time.Now()
		bot.Mutex.Unlock()
	}
}
