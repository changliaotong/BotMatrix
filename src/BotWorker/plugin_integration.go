package main

import (
	"botworker/internal/plugin"
	"botworker/internal/server"
	"fmt"

	"BotMatrix/common/plugin/core"
)

// PluginBridge 桥接我们的插件系统和BotWorker的插件系统
type PluginBridge struct {
	pluginManager *core.PluginManager
	server        *server.CombinedServer
}

func NewPluginBridge(server *server.CombinedServer) *PluginBridge {
	pm := core.NewPluginManager()
	// 配置插件路径
	pm.SetPluginPath("plugins")
	return &PluginBridge{
		pluginManager: pm,
		server:        server,
	}
}

func (pb *PluginBridge) LoadExternalPlugins() error {
	// 扫描插件
	if err := pb.pluginManager.ScanPlugins("plugins"); err != nil {
		return fmt.Errorf("扫描插件失败: %v", err)
	}

	// 启动所有插件
	for name := range pb.pluginManager.GetPlugins() {
		if err := pb.pluginManager.StartPlugin(name); err != nil {
			return fmt.Errorf("启动插件%s失败: %v", name, err)
		}
	}

	// 注册插件事件处理
	pb.pluginManager.RegisterEventHandler(func(event *core.EventMessage) {
		// 将插件事件转换为BotWorker事件
		// 这里需要实现事件转换逻辑
	})

	return nil
}

// 实现plugin.Plugin接口的包装器
type ExternalPluginWrapper struct {
	plugin *core.Plugin
}

func (w *ExternalPluginWrapper) Name() string {
	return w.plugin.Config.Name
}

func (w *ExternalPluginWrapper) Description() string {
	return w.plugin.Config.Description
}

func (w *ExternalPluginWrapper) Version() string {
	return w.plugin.Config.Version
}

func (w *ExternalPluginWrapper) Init(robot plugin.Robot) {
	// 初始化插件
	// 这里需要实现初始化逻辑
}
