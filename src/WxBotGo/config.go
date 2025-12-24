package core

import (
	"encoding/json"
	"os"
)

// Config 配置结构
type Config struct {
	Networks  []NetworkConfig  `json:"networks"`
	HTTPs     []HTTPConfig     `json:"https"`
	WebSockets []WebSocketConfig `json:"websockets"`
	Logging   LoggingConfig   `json:"logging"`
	Features  FeaturesConfig  `json:"features"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ManagerURL string `json:"manager_url"`
	SelfID     string `json:"self_id"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
}

// WebSocketConfig WebSocket 服务器配置
type WebSocketConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
	Path    string `json:"path"`
}

// FeaturesConfig 功能配置
type FeaturesConfig struct {
	AutoLogin     bool `json:"auto_login"`
	QRCodeSave    bool `json:"qr_code_save"`
	AutoReconnect bool `json:"auto_reconnect"`
	ReportSelfMsg bool `json:"report_self_msg"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(filename string, config *Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Networks: []NetworkConfig{
			{
				ManagerURL: "ws://localhost:3001",
				SelfID:     "", // Will be set by server
			},
		},
		HTTPs: []HTTPConfig{
			{
				Enabled: true,
				Port:    "8080",
				Host:    "0.0.0.0",
			},
		},
		WebSockets: []WebSocketConfig{
			{
				Enabled: true,
				Port:    "3001",
				Host:    "0.0.0.0",
				Path:    "/ws",
			},
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "wxbotgo.log",
		},
		Features: FeaturesConfig{
		AutoLogin:     true,
		QRCodeSave:    true,
		AutoReconnect: true,
		ReportSelfMsg: true,
	},
	}
}
