package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// handleLogin 处理登录请求
func (m *Manager) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   "test_token",
	})
}

// handleGetStats 处理获取统计信息的请求
func (m *Manager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取连接统计
	m.connectionStats.Mutex.Lock()
	botDurations := make(map[string]string)
	workerDurations := make(map[string]string)

	for k, v := range m.connectionStats.BotConnectionDurations {
		botDurations[k] = v.String()
	}
	for k, v := range m.connectionStats.WorkerConnectionDurations {
		workerDurations[k] = v.String()
	}

	stats := map[string]interface{}{
		"bots": map[string]interface{}{
			"count":       len(m.bots),
			"durations":   botDurations,
			"disconnects": m.connectionStats.BotDisconnectReasons,
		},
		"workers": map[string]interface{}{
			"count":       len(m.workers),
			"durations":   workerDurations,
			"disconnects": m.connectionStats.WorkerDisconnectReasons,
		},
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	m.connectionStats.Mutex.Unlock()

	json.NewEncoder(w).Encode(stats)
}

// handleGetLogs 处理获取日志的请求
func (m *Manager) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": []string{"日志功能开发中..."},
	})
}
