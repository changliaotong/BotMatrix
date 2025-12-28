package common

import (
	"fmt"
	"sync"
)

// ConnectionHandler 连接处理器接口
type ConnectionHandler interface {
	Start(config ConnectionConfig) error
	Stop() error
	IsRunning() bool
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	connections map[string]ConnectionConfig
	handlers    map[string]ConnectionHandler
	mu          sync.RWMutex
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]ConnectionConfig),
		handlers: make(map[string]ConnectionHandler),
	}
}

// AddConnection 添加连接配置
func (m *ConnectionManager) AddConnection(config ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[config.Name]; exists {
		return fmt.Errorf("connection with name %s already exists", config.Name)
	}
	m.connections[config.Name] = config
	return nil
}

// RemoveConnection 移除连接配置
func (m *ConnectionManager) RemoveConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection with name %s not found", name)
	}
	delete(m.connections, name)
	// Stop and clean up running handler if exists
	if handler, exists := m.handlers[name]; exists {
		_ = handler.Stop()
		delete(m.handlers, name)
	}
	return nil
}

// UpdateConnection 更新连接配置
func (m *ConnectionManager) UpdateConnection(name string, config ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection with name %s not found", name)
	}
	// Preserve original name to avoid mismatches
	config.Name = name
	m.connections[name] = config
	// Restart running connection to apply new config
	if handler, exists := m.handlers[name]; exists && handler.IsRunning() {
		_ = handler.Stop()
		_ = handler.Start(config)
	}
	return nil
}

// GetConnections 获取所有连接配置
func (m *ConnectionManager) GetConnections() []ConnectionConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	configs := make([]ConnectionConfig, 0, len(m.connections))
	for _, config := range m.connections {
		configs = append(configs, config)
	}
	return configs
}

// StartConnection 启动连接
func (m *ConnectionManager) StartConnection(name string, handler ConnectionHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	config, exists := m.connections[name]
	if !exists {
		return fmt.Errorf("connection not found: %s", name)
	}
	if !config.Enabled {
		return fmt.Errorf("connection is not enabled")
	}
	err := handler.Start(config)
	if err != nil {
		return err
	}
	m.handlers[config.Name] = handler
	return nil
}

// StopConnection 停止连接
func (m *ConnectionManager) StopConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	handler, exists := m.handlers[name]
	if !exists {
		return fmt.Errorf("connection handler not found: %s", name)
	}
	err := handler.Stop()
	if err != nil {
		return err
	}
	delete(m.handlers, name)
	return nil
}

// StopAllConnections 停止所有连接
func (m *ConnectionManager) StopAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, handler := range m.handlers {
		if handler.IsRunning() {
			handler.Stop()
		}
		delete(m.handlers, name)
	}
}
