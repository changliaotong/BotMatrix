package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PluginManager struct {
	plugins map[string]*Plugin
	mutex   sync.Mutex
}

type Plugin struct {
	ID           string
	Config       *PluginConfig
	Process      *os.Process
	Stdin        *os.File
	Stdout       *os.File
	State        string
	RestartCount int
	LastRestart  time.Time
	Version      string
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*Plugin),
	}
}

func (pm *PluginManager) ScanPlugins(dir string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	files, err := filepath.Glob(filepath.Join(dir, "*", "plugin.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		config, err := LoadPluginConfig(file)
		if err != nil {
			fmt.Printf("Invalid plugin config %s: %v\n", file, err)
			continue
		}

		if err := ValidatePluginConfig(config); err != nil {
			fmt.Printf("Plugin config validation failed %s: %v\n", file, err)
			continue
		}

		pm.plugins[config.Name] = &Plugin{
			ID:     config.Name,
			Config: config,
			State:  "stopped",
		}
	}

	return nil
}

func LoadPluginConfig(file string) (*PluginConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config PluginConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ValidatePluginConfig(config *PluginConfig) error {
	if config.APIVersion != "1.0" {
		return fmt.Errorf("unsupported API version %s", config.APIVersion)
	}

	if config.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if config.EntryPoint == "" {
		return fmt.Errorf("entry point is required")
	}

	// Validate plugin level
	if config.PluginLevel != "master" && config.PluginLevel != "feature" {
		return fmt.Errorf("invalid plugin level %s", config.PluginLevel)
	}

	// Validate signature if present
	if config.Signature != "" {
		if err := VerifyPluginSignature(config); err != nil {
			return fmt.Errorf("invalid plugin signature: %v", err)
		}
	}

	return nil
}

func VerifyPluginSignature(config *PluginConfig) error {
	// TODO: Implement actual signature verification logic
	// This is a placeholder for future implementation
	// Could use RSA, ECDSA, or other signature algorithms
	return nil
}

func VerifyPluginIntegrity(plugin *Plugin) error {
	// TODO: Implement plugin integrity check
	// Could use hash verification of plugin files
	return nil
}
