package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WorkerClient represents a business logic worker
type WorkerClient struct {
	ID            string // Worker标识
	Conn          *websocket.Conn
	Mutex         sync.Mutex
	Connected     time.Time
	HandledCount  int64
	LastHeartbeat time.Time
}

// Worker连接管理相关的方法

// 添加Worker连接统计
func (m *Manager) TrackWorkerConnection(workerID string) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()
	
	m.connectionStats.TotalWorkerConnections++
	m.connectionStats.LastWorkerActivity[workerID] = time.Now()
	m.AddLog("INFO", fmt.Sprintf("[Worker Connected] ID: %s, Total: %d", workerID, m.connectionStats.TotalWorkerConnections))
}

// 记录Worker断开连接
func (m *Manager) TrackWorkerDisconnection(workerID string, reason string, duration time.Duration) {
	m.connectionStats.Mutex.Lock()
	defer m.connectionStats.Mutex.Unlock()
	
	m.connectionStats.WorkerConnectionDurations[workerID] = duration
	m.connectionStats.WorkerDisconnectReasons[reason]++
	delete(m.connectionStats.LastWorkerActivity, workerID)
	
	m.AddLog("INFO", fmt.Sprintf("[Worker Disconnected] ID: %s, Reason: %s, Duration: %v", workerID, reason, duration))
}

// Worker心跳处理
func (m *Manager) updateWorkerHeartbeat(workerID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, worker := range m.workers {
		if worker.ID == workerID {
			worker.Mutex.Lock()
			worker.LastHeartbeat = time.Now()
			worker.Mutex.Unlock()
			break
		}
	}
}