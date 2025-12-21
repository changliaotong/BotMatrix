package main

import (
	"botworker/internal/config"
	"botworker/internal/plugin"
	"botworker/internal/server"
	"botworker/plugins"
	"log"
	"os"
)

func main() {
	log.Println("=== 简单测试启动 ===")
	log.Println("当前工作目录:", func() string { dir, _ := os.Getwd(); return dir }())

	// 加载配置
	cfg, _, err := config.LoadFromCLI()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建服务器
	combinedServer := server.NewCombinedServer(cfg)

	// 获取插件管理器
	pluginManager := combinedServer.GetPluginManager()

	// 只加载群管理插件
	groupManagerPlugin := plugins.NewGroupManagerPlugin(nil, nil)
	if err := pluginManager.LoadPlugin(groupManagerPlugin); err != nil {
		log.Fatalf("加载群管理插件失败: %v", err)
	}

	// 打印已加载的插件
	log.Println("已加载的插件:")
	for _, plugin := range pluginManager.GetPlugins() {
		log.Printf("- %s v%s: %s", plugin.Name(), plugin.Version(), plugin.Description())
	}

	// 启动服务器
	log.Println("启动简化版服务器...")
	if err := combinedServer.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}