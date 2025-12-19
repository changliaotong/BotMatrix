package main

import (
	"fmt"
	"log"
	"time"
)

// AddLog 添加日志条目
func (m *Manager) AddLog(level string, message string) {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()

	now := time.Now()
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Time:      now.Format("15:04:05"),
		Timestamp: now,
	}

	m.logBuffer = append(m.logBuffer, entry)
	if len(m.logBuffer) > 1000 {
		m.logBuffer = m.logBuffer[len(m.logBuffer)-1000:]
	}

	// 广播给所有订阅者
	go func() {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		
		msg := map[string]interface{}{
			"post_type": "log",
			"data":      entry,
			"self_id":   "", // 系统日志没有 self_id
		}
		
		for _, sub := range m.subscribers {
			sub.Mutex.Lock()
			sub.Conn.WriteJSON(msg)
			sub.Mutex.Unlock()
		}
	}()

	// 同时打印到控制台
	log.Printf("[%s] %s", level, message)
}

// GetLogs 获取最近的日志
func (m *Manager) GetLogs(limit int) []LogEntry {
	m.logMutex.RLock()
	defer m.logMutex.RUnlock()

	if limit <= 0 || limit > len(m.logBuffer) {
		limit = len(m.logBuffer)
	}

	result := make([]LogEntry, limit)
	copy(result, m.logBuffer[len(m.logBuffer)-limit:])
	return result
}

// 快捷日志函数
func (m *Manager) LogDebug(format string, args ...interface{}) {
	m.AddLog("DEBUG", fmt.Sprintf(format, args...))
}

func (m *Manager) LogInfo(format string, args ...interface{}) {
	m.AddLog("INFO", fmt.Sprintf(format, args...))
}

func (m *Manager) LogWarn(format string, args ...interface{}) {
	m.AddLog("WARN", fmt.Sprintf(format, args...))
}

func (m *Manager) LogError(format string, args ...interface{}) {
	m.AddLog("ERROR", fmt.Sprintf(format, args...))
}
