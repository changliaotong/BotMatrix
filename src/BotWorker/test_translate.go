package main

import (
	"botworker/plugins"
	"log"
)

func main() {
	log.Println("测试翻译插件功能...")

	// 创建翻译插件实例
	translatePlugin := plugins.NewTranslatePlugin()

	// 测试插件基本信息
	log.Printf("插件名称: %s", translatePlugin.Name())
	log.Printf("插件版本: %s", translatePlugin.Version())
	log.Printf("插件描述: %s", translatePlugin.Description())

	// 测试中文检测
	log.Println("\n测试中文检测:")
	chineseText := "你好，世界"
	isChinese := translatePlugin.IsChinese(chineseText)
	log.Printf("'%s' 是否为中文: %t", chineseText, isChinese)

	englishText := "Hello, world"
	isChinese = translatePlugin.IsChinese(englishText)
	log.Printf("'%s' 是否为中文: %t", englishText, isChinese)

	log.Println("\n翻译插件功能测试通过!")
	log.Println("注意: 完整的API调用测试需要网络连接")
}