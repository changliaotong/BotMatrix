package common

import (
	"fmt"
	"log"
	"time"
)

// AddLog 添加日志条目
func (m *Manager) AddLog(level string, message string) {
	m.LogMutex.Lock()
	defer m.LogMutex.Unlock()

	now := time.Now()
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Time:      now.Format("15:04:05"),
		Timestamp: now,
	}

	m.LogBuffer = append(m.LogBuffer, entry)
	if len(m.LogBuffer) > 1000 {
		m.LogBuffer = m.LogBuffer[len(m.LogBuffer)-1000:]
	}

	// 广播给所有订阅者
	go func() {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()
		
		msg := map[string]interface{}{
			"post_type": "log",
			"data":      entry,
			"self_id":   "", // 系统日志没有 self_id
		}
		
		for _, sub := range m.Subscribers {
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
	m.LogMutex.RLock()
	defer m.LogMutex.RUnlock()

	if limit <= 0 || limit > len(m.LogBuffer) {
		limit = len(m.LogBuffer)
	}

	result := make([]LogEntry, limit)
	copy(result, m.LogBuffer[len(m.LogBuffer)-limit:])
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
