package common

import (
	"encoding/json"
	"os"
)

// ConnectionConfig 统一的连接配置
// 适用于所有客户端类型：QQ、微信、企业微信、电报等
type ConnectionConfig struct {
	ClientType string `json:"client_type"` // "wx", "qq", "wecom", "telegram"
	Name       string `json:"name"`        // 连接名称
	Enabled    bool   `json:"enabled"`     // 是否启用

	// HTTP 配置
	HTTPConfig HTTPConfig `json:"http_config"`

	// WebSocket 配置
	WebSocketConfig WebSocketConfig `json:"websocket_config"`

	// 客户端特定配置
	ClientSpecific map[string]interface{} `json:"client_specific"`
}

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
}

// WebSocketConfig WebSocket 服务配置
type WebSocketConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
	Path    string `json:"path"`
}

// ConnectionConfigManager 连接配置管理器
type ConnectionConfigManager struct {
	Connections []ConnectionConfig `json:"connections"`
}

// LoadConnectionConfig 加载连接配置
func LoadConnectionConfig(filePath string) (*ConnectionConfigManager, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return DefaultConnectionConfig(), nil
	}
	defer file.Close()

	var manager ConnectionConfigManager
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&manager)
	if err != nil {
		return DefaultConnectionConfig(), err
	}

	return &manager, nil
}

// SaveConnectionConfig 保存连接配置
func (m *ConnectionConfigManager) SaveConnectionConfig(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(m)
	return err
}

// DefaultConnectionConfig 默认连接配置
func DefaultConnectionConfig() *ConnectionConfigManager {
	return &ConnectionConfigManager{
		Connections: []ConnectionConfig{
			{
				ClientType: "wx",
				Name:       "微信机器人",
				Enabled:    true,
				HTTPConfig: HTTPConfig{
					Enabled: true,
					Port:    "8080",
					Host:    "0.0.0.0",
				},
				WebSocketConfig: WebSocketConfig{
					Enabled: true,
					Port:    "3001",
					Host:    "0.0.0.0",
					Path:    "/ws",
				},
				ClientSpecific: map[string]interface{}{
					"self_id": "",
					"report_self_msg": true, // 默认上报自身消息
				},
			},
		},
	}
}
