package main

import (
	"fmt"
	"time"
)

// HandlePluginError 处理插件错误
func (pm *PluginManager) HandlePluginError(plugin *Plugin, err error) {
	plugin.Mutex.Lock()
	defer plugin.Mutex.Unlock()

	fmt.Printf("插件%s错误: %v\n", plugin.ID, err)

	// 检查是否需要重启
	if plugin.RestartCount < plugin.Config.MaxRestarts {
		// 检查重启频率
		if time.Since(plugin.LastRestart) > 5*time.Second {
			plugin.RestartCount = 0 // 重置重启计数
		}

		fmt.Printf("正在重启插件%s (第%d次)\n", plugin.ID, plugin.RestartCount+1)
		plugin.State = StateRestarting

		// 异步重启
		go func() {
			if err := plugin.Restart(); err != nil {
				fmt.Printf("重启插件%s失败: %v\n", plugin.ID, err)
				plugin.State = StateCrashed
			} else {
				fmt.Printf("插件%s重启成功\n", plugin.ID)
				plugin.State = StateRunning
			}
		}()
	} else {
		fmt.Printf("插件%s已达到最大重启次数，停止重启\n", plugin.ID)
		plugin.State = StateCrashed
	}
}

// CheckPluginTimeouts 检查插件超时
func (pm *PluginManager) CheckPluginTimeouts() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查总控插件
	for _, plugin := range pm.masterPlugins {
		plugin.Mutex.Lock()
		defer plugin.Mutex.Unlock()

		if plugin.State == StateRunning {
			// 检查是否超时
			// 这里可以根据实际情况实现超时逻辑
		}
	}

	// 检查功能插件
	for _, plugin := range pm.featurePlugins {
		plugin.Mutex.Lock()
		defer plugin.Mutex.Unlock()

		if plugin.State == StateRunning {
			// 检查是否超时
			// 这里可以根据实际情况实现超时逻辑
		}
	}
}
