package core

import (
	"archive/zip"
	log "BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"io"
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

		pm.plugins[config.ID] = &Plugin{
			ID:     config.ID,
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
	if config.ID == "" {
		return fmt.Errorf("plugin id is required")
	}

	if config.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if config.EntryPoint == "" {
		return fmt.Errorf("entry point is required")
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

func (pm *PluginManager) InstallPlugin(bmpkPath string, targetDir string) error {
	// 1. Open .bmpk file (which is a zip)
	r, err := zip.OpenReader(bmpkPath)
	if err != nil {
		return fmt.Errorf("failed to open bmpk: %v", err)
	}
	defer r.Close()

	// 2. Find and read plugin.json first to get ID
	var manifest *PluginConfig
	for _, f := range r.File {
		if f.Name == "plugin.json" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(data, &manifest); err != nil {
				return fmt.Errorf("invalid manifest in bmpk: %v", err)
			}
			break
		}
	}

	if manifest == nil {
		return fmt.Errorf("plugin.json not found in bmpk")
	}

	// 3. Create target directory
	pluginDir := filepath.Join(targetDir, manifest.ID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// 4. Extract all files
	for _, f := range r.File {
		fpath := filepath.Join(pluginDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	// 5. Load the newly installed plugin
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.plugins[manifest.ID] = &Plugin{
		ID:     manifest.ID,
		Config: manifest,
		State:  "stopped",
	}

	log.Printf("Successfully installed plugin: %s (v%s)", manifest.Name, manifest.Version)
	return nil
}

func VerifyPluginIntegrity(plugin *Plugin) error {
	// TODO: Implement plugin integrity check
	// Could use hash verification of plugin files
	return nil
}
