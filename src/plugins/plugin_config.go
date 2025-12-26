package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// PluginConfig 插件配置结构
type PluginConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	EntryPoint  string `json:"entry_point"` // 插件可执行文件路径
	TimeoutMS   int    `json:"timeout_ms"`  // 超时时间，默认5000ms
	MaxRestarts int    `json:"max_restarts"` // 最大重启次数，默认3次
	Type        string `json:"type"`        // 插件类型："master" 或 "feature"
}

// LoadPluginConfig 加载插件配置
func LoadPluginConfig(configPath string) (*PluginConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer file.Close()

	var config PluginConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if config.TimeoutMS == 0 {
		config.TimeoutMS = 5000
	}
	if config.MaxRestarts == 0 {
		config.MaxRestarts = 3
	}
	if config.Type == "" {
		config.Type = "feature" // 默认类型为功能插件
	}

	return &config, nil
}

// SavePluginConfig 保存插件配置
func SavePluginConfig(configPath string, config *PluginConfig) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}