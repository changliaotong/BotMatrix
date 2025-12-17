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
