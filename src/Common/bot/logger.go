package bot

import (
	"fmt"
	"sync"
	"time"

	"BotMatrix/common/types"
)

// AddLog adds a log entry to the buffer and handles broadcasting
func AddLog(
	logBuffer []types.LogEntry,
	logMutex *sync.RWMutex,
	level string,
	message string,
	broadcast func(any),
	source ...string,
) []types.LogEntry {
	logMutex.Lock()
	defer logMutex.Unlock()

	logSource := ""
	if len(source) > 0 {
		logSource = source[0]
	}

	now := time.Now()
	entry := types.LogEntry{
		Level:     level,
		Message:   message,
		Time:      now.Format("15:04:05"),
		Timestamp: now,
		Source:    logSource,
	}

	logBuffer = append(logBuffer, entry)
	if len(logBuffer) > 1000 {
		logBuffer = logBuffer[len(logBuffer)-1000:]
	}

	go func() {
		msg := map[string]any{
			"post_type": "log",
			"data":      entry,
			"self_id":   "",
		}
		broadcast(msg)
	}()

	sourceStr := ""
	if logSource != "" {
		sourceStr = fmt.Sprintf("[%s] ", logSource)
	}
	fmt.Printf("[%s] %s[%s] %s\n", now.Format("2006-01-02 15:04:05"), sourceStr, level, message)

	return logBuffer
}

// AddLog adds a log entry
func (m *Manager) AddLog(level string, message string, source ...string) {
	m.LogBuffer = AddLog(
		m.LogBuffer,
		&m.LogMutex,
		level,
		message,
		m.BroadcastEvent,
		source...,
	)
}

// GetLogs gets recent logs
func (m *Manager) GetLogs(limit int) []types.LogEntry {
	return GetLogs(m.LogBuffer, &m.LogMutex, limit)
}

// Helper log functions
func (m *Manager) LogTrace(format string, args ...any) {
	m.AddLog("TRACE", fmt.Sprintf(format, args...))
}

func (m *Manager) LogDebug(format string, args ...any) {
	m.AddLog("DEBUG", fmt.Sprintf(format, args...))
}

func (m *Manager) LogInfo(format string, args ...any) {
	m.AddLog("INFO", fmt.Sprintf(format, args...))
}

func (m *Manager) LogWarn(format string, args ...any) {
	m.AddLog("WARN", fmt.Sprintf(format, args...))
}

func (m *Manager) LogError(format string, args ...any) {
	m.AddLog("ERROR", fmt.Sprintf(format, args...))
}

// ClearLogs clears log buffer
func (m *Manager) ClearLogs() {
	m.LogMutex.Lock()
	defer m.LogMutex.Unlock()
	m.LogBuffer = make([]types.LogEntry, 0)
}

// GetLogs gets recent logs from the buffer
func GetLogs(logBuffer []types.LogEntry, logMutex *sync.RWMutex, limit int) []types.LogEntry {
	logMutex.RLock()
	defer logMutex.RUnlock()

	if limit <= 0 || limit > len(logBuffer) {
		limit = len(logBuffer)
	}

	result := make([]types.LogEntry, limit)
	copy(result, logBuffer[len(logBuffer)-limit:])
	return result
}
