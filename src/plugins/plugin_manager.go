package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PluginManager 插件管理器
type PluginManager struct {
	masterPlugins map[string]*Plugin   // 总控插件
	featurePlugins map[string]*Plugin  // 功能插件
	mutex          sync.Mutex
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		masterPlugins: make(map[string]*Plugin),
		featurePlugins: make(map[string]*Plugin),
	}
}

// LoadPlugins 加载插件目录下的所有插件
func (pm *PluginManager) LoadPlugins(dir string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 扫描master插件目录
	masterDir := filepath.Join(dir, "master")
	if _, err := os.Stat(masterDir); err == nil {
		masterPluginDirs, err := filepath.Glob(filepath.Join(masterDir, "*"))
		if err == nil {
			for _, pluginDir := range masterPluginDirs {
				pm.loadSinglePlugin(pluginDir, "master")
			}
		}
	}

	// 扫描feature插件目录
	featureDir := filepath.Join(dir, "features")
	if _, err := os.Stat(featureDir); err == nil {
		featurePluginDirs, err := filepath.Glob(filepath.Join(featureDir, "*"))
		if err == nil {
			for _, pluginDir := range featurePluginDirs {
				pm.loadSinglePlugin(pluginDir, "feature")
			}
		}
	}

	// 兼容旧的插件目录结构
	legacyPluginDirs, err := filepath.Glob(filepath.Join(dir, "*"))
	if err == nil {
		for _, pluginDir := range legacyPluginDirs {
			info, err := os.Stat(pluginDir)
			if err != nil || !info.IsDir() {
				continue
			}
			if pluginDir == masterDir || pluginDir == featureDir {
				continue
			}
			pm.loadSinglePlugin(pluginDir, "feature")
		}
	}

	return nil
}

// loadSinglePlugin 加载单个插件
func (pm *PluginManager) loadSinglePlugin(pluginDir string, expectedType string) {
	// 加载插件配置
	configPath := filepath.Join(pluginDir, "plugin.json")
	config, err := LoadPluginConfig(configPath)
	if err != nil {
		fmt.Printf("加载插件配置失败: %v\n", err)
		return
	}

	// 检查插件类型是否匹配
	if expectedType != "" && config.Type != expectedType {
		fmt.Printf("插件类型不匹配: %s 应该是 %s 类型\n", config.Name, expectedType)
		return
	}

	// 创建插件实例
	pluginID := config.Name
	plugin := NewPlugin(pluginID, config)
	
	// 启动插件
	if err := plugin.Start(); err != nil {
		fmt.Printf("启动插件%s失败: %v\n", pluginID, err)
		return
	}

	// 根据类型添加到不同的插件映射
	if config.Type == "master" {
		pm.masterPlugins[pluginID] = plugin
	} else {
		pm.featurePlugins[pluginID] = plugin
	}

	fmt.Printf("%s插件%s加载成功\n", config.Type, pluginID)
}

// StartPlugin 启动指定插件
func (pm *PluginManager) StartPlugin(id string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 先在总控插件中查找
	if plugin, ok := pm.masterPlugins[id]; ok {
		return plugin.Start()
	}
	
	// 再在功能插件中查找
	if plugin, ok := pm.featurePlugins[id]; ok {
		return plugin.Start()
	}

	return fmt.Errorf("插件%s不存在", id)
}

// StopPlugin 停止指定插件
func (pm *PluginManager) StopPlugin(id string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 先在总控插件中查找
	if plugin, ok := pm.masterPlugins[id]; ok {
		return plugin.Stop()
	}
	
	// 再在功能插件中查找
	if plugin, ok := pm.featurePlugins[id]; ok {
		return plugin.Stop()
	}

	return fmt.Errorf("插件%s不存在", id)
}

// DispatchEvent 分发事件到所有插件
func (pm *PluginManager) DispatchEvent(eventName string, payload any) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 构造事件消息
	event := EventMessage{
		ID:      generateUUID(),
		Type:    "event",
		Name:    eventName,
		Payload: payload,
	}

	// 发送到总控插件
	for _, plugin := range pm.masterPlugins {
		if plugin.State != StateRunning {
			continue
		}

		// 序列化事件
		jsonData, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("序列化事件失败: %v\n", err)
			continue
		}

		// 发送到插件stdin
		if _, err := plugin.Stdin.Write(jsonData); err != nil {
			fmt.Printf("发送事件到总控插件%s失败: %v\n", plugin.ID, err)
			continue
		}
	}

	// 发送到功能插件
	for _, plugin := range pm.featurePlugins {
		if plugin.State != StateRunning {
			continue
		}

		// 序列化事件
		jsonData, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("序列化事件失败: %v\n", err)
			continue
		}

		// 发送到插件stdin
		if _, err := plugin.Stdin.Write(jsonData); err != nil {
			fmt.Printf("发送事件到功能插件%s失败: %v\n", plugin.ID, err)
			continue
		}
	}

	return nil
}