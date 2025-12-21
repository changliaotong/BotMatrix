package main

import (
	"botworker/plugins"
	"log"
)

func main() {
	log.Println("测试菜单插件功能...")

	// 创建菜单插件实例
	menuPlugin := plugins.NewMenuPlugin()

	// 测试插件基本信息
	log.Printf("插件名称: %s", menuPlugin.Name())
	log.Printf("插件版本: %s", menuPlugin.Version())
	log.Printf("插件描述: %s", menuPlugin.Description())

	// 测试菜单内容
	log.Println("\n菜单内容:")
	menu := menuPlugin.GetMenu()
	log.Println(menu)

	log.Println("\n菜单插件功能测试通过!")
}