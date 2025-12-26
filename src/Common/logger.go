package common

import (
	"fmt"
	"time"
)

// AddLog adds a log entry
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

	// Broadcast to all subscribers
	go func() {
		msg := map[string]interface{}{
			"post_type": "log",
			"data":      entry,
			"self_id":   "", // System logs have no self_id
		}
		m.BroadcastEvent(msg)
	}()

	// Also print to console without using log.Printf to avoid infinite loops if redirected
	fmt.Printf("[%s] [%s] %s\n", now.Format("2006-01-02 15:04:05"), level, message)
}

// GetLogs gets recent logs
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

// Helper log functions
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

// ClearLogs clears log buffer
func (m *Manager) ClearLogs() {
	m.LogMutex.Lock()
	defer m.LogMutex.Unlock()
	m.LogBuffer = make([]LogEntry, 0)
}
