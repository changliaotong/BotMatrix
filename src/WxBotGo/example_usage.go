package main

import (
	"BotMatrix/common"
	"fmt"
)

func main() {
	// 示例：使用 common 包进行连接配置管理
	
	// 1. 创建连接管理器
	manager := common.NewConnectionManager()
	
	// 2. 创建微信连接配置
	wxConfig := common.ConnectionConfig{
		ClientType: "wx",
		Name:       "微信机器人",
		Enabled:    true,
		HTTPConfig: common.HTTPConfig{
			Enabled: true,
			Port:    "8080",
			Host:    "0.0.0.0",
		},
		WebSocketConfig: common.WebSocketConfig{
			Enabled: true,
			Port:    "3001",
			Host:    "0.0.0.0",
			Path:    "/ws",
		},
		ClientSpecific: map[string]any{
			"self_id": "",
		},
	}
	
	// 3. 添加连接配置
	manager.AddConnection(wxConfig)
	
	// 4. 创建 QQ 连接配置
	qqConfig := common.ConnectionConfig{
		ClientType: "qq",
		Name:       "QQ机器人",
		Enabled:    true,
		HTTPConfig: common.HTTPConfig{
			Enabled: true,
			Port:    "8081",
			Host:    "0.0.0.0",
		},
		WebSocketConfig: common.WebSocketConfig{
			Enabled: true,
			Port:    "3002",
			Host:    "0.0.0.0",
			Path:    "/ws",
		},
		ClientSpecific: map[string]any{
			"self_id": "",
		},
	}
	
	// 5. 添加 QQ 连接配置
	manager.AddConnection(qqConfig)
	
	// 6. 获取所有连接配置
	connections := manager.GetConnections()
	fmt.Printf("共有 %d 个连接配置\n", len(connections))
	
	// 7. 遍历连接配置
	for i, conn := range connections {
		fmt.Printf("连接 %d: %s (%s)\n", i, conn.Name, conn.ClientType)
		fmt.Printf("  HTTP: %v, Port: %s\n", conn.HTTPConfig.Enabled, conn.HTTPConfig.Port)
		fmt.Printf("  WebSocket: %v, Port: %s\n", conn.WebSocketConfig.Enabled, conn.WebSocketConfig.Port)
	}
	
	// 8. 保存连接配置到文件
	configManager := common.ConnectionManager{
		Connections: connections,
	}
	err := configManager.SaveConnectionConfig("all_connections.json")
	if err != nil {
		fmt.Printf("保存配置失败: %v\n", err)
	} else {
		fmt.Println("配置保存成功")
	}
}
