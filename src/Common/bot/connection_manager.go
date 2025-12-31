package bot

import (
	"fmt"
	"sync"
)

// ConnectionHandler Connection handler interface
type ConnectionHandler interface {
	Start(config ConnectionConfig) error
	Stop() error
	IsRunning() bool
}

// ConnectionManager manages multiple bot connections
type ConnectionManager struct {
	connections map[string]ConnectionConfig
	handlers    map[string]ConnectionHandler
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]ConnectionConfig),
		handlers:    make(map[string]ConnectionHandler),
	}
}

// AddConnection adds a connection configuration
func (m *ConnectionManager) AddConnection(config ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[config.Name]; exists {
		return fmt.Errorf("connection with name %s already exists", config.Name)
	}
	m.connections[config.Name] = config
	return nil
}

// RemoveConnection removes a connection configuration
func (m *ConnectionManager) RemoveConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection with name %s not found", name)
	}
	delete(m.connections, name)
	if handler, exists := m.handlers[name]; exists {
		_ = handler.Stop()
		delete(m.handlers, name)
	}
	return nil
}

// UpdateConnection updates a connection configuration
func (m *ConnectionManager) UpdateConnection(name string, config ConnectionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection with name %s not found", name)
	}
	config.Name = name
	m.connections[name] = config
	if handler, exists := m.handlers[name]; exists && handler.IsRunning() {
		_ = handler.Stop()
		_ = handler.Start(config)
	}
	return nil
}

// GetConnections returns all connection configurations
func (m *ConnectionManager) GetConnections() []ConnectionConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	configs := make([]ConnectionConfig, 0, len(m.connections))
	for _, config := range m.connections {
		configs = append(configs, config)
	}
	return configs
}

// StartConnection starts a connection with a handler
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
	m.handlers[name] = handler
	return nil
}
