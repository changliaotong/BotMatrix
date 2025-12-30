package core

import (
	log "BotMatrix/common/log"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PluginManager struct {
	plugins         map[string][]*Plugin // Changed: ID maps to a list of versions
	internalPlugins map[string]PluginModule
	pluginPath      string
	eventHandler    func(*EventMessage)
	actionHandler   func(*Plugin, *Action)
	mutex           sync.Mutex
}

type Plugin struct {
	ID            string
	Config        *PluginConfig
	Process       *os.Process
	Stdin         *os.File
	Stdout        *os.File
	State         string
	RestartCount  int
	LastRestart   time.Time
	Version       string
	Dir           string          // Added: The directory where the plugin is located
	RuntimeDir    string          // Added: The temporary directory where the plugin is running (Shadow Copy)
	MessageBuffer []*EventMessage // Added: Buffer messages during reload
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:         make(map[string][]*Plugin),
		internalPlugins: make(map[string]PluginModule),
	}
}

func (pm *PluginManager) SetPluginPath(path string) {
	pm.pluginPath = path
}

func (pm *PluginManager) GetPluginPath() string {
	return pm.pluginPath
}

func (pm *PluginManager) LoadPluginModule(m PluginModule, robot Robot) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if robot != nil {
		m.Init(robot)
	}
	pm.internalPlugins[m.Name()] = m
	return nil
}

func (pm *PluginManager) ScanPlugins(dir string) error {
	return pm.LoadPlugins(dir)
}

func (pm *PluginManager) LoadPlugins(dir string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Support both simple and versioned folders
	// 1. Simple: plugins/my_plugin/plugin.json
	// 2. Versioned: plugins/my_plugin/1.0.0/plugin.json

	// Check simple folders
	simpleFiles, _ := filepath.Glob(filepath.Join(dir, "*", "plugin.json"))
	for _, file := range simpleFiles {
		config, err := LoadPluginConfig(file)
		if err != nil {
			continue
		}
		pm.addPluginInternal(config, filepath.Dir(file))
	}

	// Check versioned folders
	versionedFiles, _ := filepath.Glob(filepath.Join(dir, "*", "*", "plugin.json"))
	for _, file := range versionedFiles {
		config, err := LoadPluginConfig(file)
		if err != nil {
			continue
		}
		pm.addPluginInternal(config, filepath.Dir(file))
	}

	return nil
}

func (pm *PluginManager) addPluginInternal(config *PluginConfig, dir string) {
	if config.ID == "" {
		log.Errorf("[PluginManager] 插件 ID 为空，跳过加载: %s", dir)
		return
	}
	// Avoid duplicates (but allow reloading the same version during development)
	versions := pm.plugins[config.ID]
	for i, v := range versions {
		if v.Config.Version == config.Version {
			if v.State == "stopped" || v.State == "crashed" {
				// If it's not running, we can safely replace it
				versions[i] = &Plugin{
					ID:     config.ID,
					Config: config,
					State:  "stopped",
					Dir:    dir,
				}
			}
			return
		}
	}

	p := &Plugin{
		ID:     config.ID,
		Config: config,
		State:  "stopped",
		Dir:    dir,
	}
	pm.plugins[config.ID] = append(pm.plugins[config.ID], p)
}

func (pm *PluginManager) GetPlugins() map[string][]*Plugin {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.plugins
}

func (pm *PluginManager) GetInternalPlugins() map[string]PluginModule {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.internalPlugins
}

func (pm *PluginManager) GetPlugin(id string, version string) *Plugin {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	versions, ok := pm.plugins[id]
	if !ok || len(versions) == 0 {
		return nil
	}

	if version == "" {
		return versions[len(versions)-1]
	}

	for _, p := range versions {
		if p.Config.Version == version {
			return p
		}
	}
	return nil
}

func (pm *PluginManager) RemovePlugin(id string, version string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	versions, ok := pm.plugins[id]
	if !ok {
		return
	}

	newVersions := make([]*Plugin, 0)
	for _, p := range versions {
		if p.Config.Version != version {
			newVersions = append(newVersions, p)
		}
	}

	if len(newVersions) == 0 {
		delete(pm.plugins, id)
	} else {
		pm.plugins[id] = newVersions
	}
}

func (pm *PluginManager) RegisterEventHandler(handler func(*EventMessage)) {
	pm.eventHandler = handler
}

func (pm *PluginManager) RegisterActionHandler(handler func(*Plugin, *Action)) {
	pm.actionHandler = handler
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
	if config.EntryPoint == "" {
		return fmt.Errorf("entry point is required")
	}
	return nil
}

func (pm *PluginManager) InstallPlugin(bmpkPath string, targetDir string) error {
	r, err := zip.OpenReader(bmpkPath)
	if err != nil {
		return fmt.Errorf("failed to open bmpk: %v", err)
	}
	defer r.Close()

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

	// Create target directory (Versioned)
	pluginDir := filepath.Join(targetDir, manifest.ID, manifest.Version)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

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

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	newPlugin := &Plugin{
		ID:     manifest.ID,
		Config: manifest,
		State:  "stopped",
		Dir:    pluginDir,
	}
	pm.plugins[manifest.ID] = append(pm.plugins[manifest.ID], newPlugin)

	log.Printf("Successfully installed plugin: %s (v%s)", manifest.Name, manifest.Version)
	return nil
}

func (pm *PluginManager) SyncFromMarket(url string, targetDir string) error {
	tmpFile, err := os.CreateTemp("", "*.bmpk")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("market returned status: %s", resp.Status)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save plugin: %v", err)
	}

	return pm.InstallPlugin(tmpFile.Name(), targetDir)
}

func VerifyPluginIntegrity(plugin *Plugin) error {
	return nil
}
