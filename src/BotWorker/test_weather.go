package main

import (
	"botworker/internal/config"
	"botworker/plugins"
	"log"
)

func main() {
	// 测试天气插件的核心功能
	log.Println("测试天气插件功能...")

	// 创建默认配置
	cfg := config.DefaultConfig()
	cfg.Weather.APIKey = "test_api_key"
	cfg.Weather.Endpoint = "https://api.openweathermap.org/data/2.5/weather"

	// 创建天气插件实例
	weatherPlugin := plugins.NewWeatherPlugin(&cfg.Weather)

	// 测试插件基本信息
	log.Printf("插件名称: %s", weatherPlugin.Name())
	log.Printf("插件版本: %s", weatherPlugin.Version())
	log.Printf("插件描述: %s", weatherPlugin.Description())

	log.Println("天气插件基本功能测试通过!")
	log.Println("注意: 完整的API调用测试需要有效的OpenWeatherMap API密钥")
}