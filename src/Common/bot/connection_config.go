package bot

import (
	"encoding/json"
	"os"
)

// ConnectionConfig Unified connection configuration
type ConnectionConfig struct {
	ClientType string `json:"client_type"` // "wx", "qq", "wecom", "telegram"
	Name       string `json:"name"`        // Connection name
	Enabled    bool   `json:"enabled"`     // Whether enabled

	HTTPConfig      HTTPConfig      `json:"http_config"`
	WebSocketConfig WebSocketConfig `json:"websocket_config"`

	ClientSpecific map[string]any `json:"client_specific"`
}

// HTTPConfig HTTP service configuration
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
}

// WebSocketConfig WebSocket service configuration
type WebSocketConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
	Path    string `json:"path"`
}

// ConnectionConfigManager manages connection configurations
type ConnectionConfigManager struct {
	Connections []ConnectionConfig `json:"connections"`
}

// LoadConnectionConfig loads connection configuration
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

// SaveConnectionConfig saves connection configuration
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

// DefaultConnectionConfig returns default connection configuration
func DefaultConnectionConfig() *ConnectionConfigManager {
	return &ConnectionConfigManager{
		Connections: []ConnectionConfig{
			{
				ClientType: "wx",
				Name:       "Default Bot",
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
				ClientSpecific: map[string]any{
					"self_id":         "",
					"report_self_msg": true,
				},
			},
		},
	}
}
