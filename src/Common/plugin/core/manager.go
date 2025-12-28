package core

import (
	log "BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PluginManager struct {
	plugins      map[string]*Plugin
	pluginPath   string
	eventHandler func(*EventMessage)
	mutex        sync.Mutex
}

func (pm *PluginManager) SetPluginPath(path string) {
	pm.pluginPath = path
}

func (pm *PluginManager) ScanPlugins(dir string) error {
	return pm.LoadPlugins(dir)
}

func (pm *PluginManager) GetPlugins() map[string]*Plugin {
	return pm.plugins
}

func (pm *PluginManager) RegisterEventHandler(handler func(*EventMessage)) {
	pm.eventHandler = handler
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

func (pm *PluginManager) LoadPlugins(dir string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	files, err := filepath.Glob(filepath.Join(dir, "*", "plugin.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		config, err := LoadPluginConfig(file)
		if err != nil {
			log.Printf("Invalid plugin config %s: %v", file, err)
			continue
		}

		if err := ValidatePluginConfig(config); err != nil {
			log.Printf("Plugin config validation failed %s: %v", file, err)
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
	// Support both "1.0" and "1.0.0"
	if config.APIVersion != "1.0" && config.APIVersion != "1.0.0" {
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
	return nil
}

func VerifyPluginIntegrity(plugin *Plugin) error {
	// TODO: Implement plugin integrity check
	return nil
}
