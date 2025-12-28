package main

import (
	"BotMatrix/common"
	"fmt"
)

// WxConnectionHandler 微信连接处理器
type WxConnectionHandler struct {
	config common.ConnectionConfig
	running bool
}

// NewWxConnectionHandler 创建微信连接处理器
func NewWxConnectionHandler(config common.ConnectionConfig) *WxConnectionHandler {
	return &WxConnectionHandler{
		config: config,
	}
}

// Start 启动微信连接
func (h *WxConnectionHandler) Start(config common.ConnectionConfig) error {
	fmt.Printf("Starting WeChat connection: %s\n", config.Name)
	fmt.Printf("Client Type: %s\n", config.ClientType)
	fmt.Printf("HTTP Enabled: %v, Port: %s\n", config.HTTPConfig.Enabled, config.HTTPConfig.Port)
	fmt.Printf("WebSocket Enabled: %v, Port: %s\n", config.WebSocketConfig.Enabled, config.WebSocketConfig.Port)
	
	// 实际启动逻辑
	h.running = true
	return nil
}

// Stop 停止微信连接
func (h *WxConnectionHandler) Stop() error {
	fmt.Printf("Stopping WeChat connection: %s\n", h.config.Name)
	h.running = false
	return nil
}

// IsRunning 检查连接是否运行
func (h *WxConnectionHandler) IsRunning() bool {
	return h.running
}

// InitConnectionManager 初始化连接管理器
func InitConnectionManager() (*common.ConnectionManager, error) {
	// 创建连接管理器
	manager := common.NewConnectionManager()
	
	// 加载连接配置
	configs, err := common.LoadConnectionConfig("connections.json")
	if err != nil {
		return nil, err
	}
	
	// 添加所有连接配置
	for _, conn := range configs.Connections {
		manager.AddConnection(conn)
	}
	
	return manager, nil
}

// StartAllConnections 启动所有连接
func StartAllConnections(manager *common.ConnectionManager) error {
	connections := manager.GetConnections()
	
	for _, conn := range connections {
		if conn.Enabled {
			handler := NewWxConnectionHandler(conn)
			err := manager.StartConnection(conn.Name, handler)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}
