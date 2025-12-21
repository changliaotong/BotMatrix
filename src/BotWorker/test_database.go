package main

import (
	"botworker/internal/config"
	"fmt"
	"log"
)

func main() {
	// 创建默认配置
	cfg := config.DefaultConfig()
	
	// 打印默认数据库配置
	fmt.Println("默认数据库配置:")
	fmt.Printf("  主机: %s\n", cfg.Database.Host)
	fmt.Printf("  端口: %d\n", cfg.Database.Port)
	fmt.Printf("  用户名: %s\n", cfg.Database.User)
	fmt.Printf("  密码: %s\n", cfg.Database.Password)
	fmt.Printf("  数据库名: %s\n", cfg.Database.DBName)
	fmt.Printf("  SSL模式: %s\n", cfg.Database.SSLMode)
	
	// 测试从文件加载配置
	fmt.Println("\n从文件加载配置...")
	cfgFromFile, err := config.LoadConfig("test_config.json")
	if err != nil {
		log.Printf("警告: 无法从文件加载配置: %v\n", err)
	} else {
		fmt.Println("从文件加载的数据库配置:")
		fmt.Printf("  主机: %s\n", cfgFromFile.Database.Host)
		fmt.Printf("  端口: %d\n", cfgFromFile.Database.Port)
		fmt.Printf("  用户名: %s\n", cfgFromFile.Database.User)
		fmt.Printf("  密码: %s\n", cfgFromFile.Database.Password)
		fmt.Printf("  数据库名: %s\n", cfgFromFile.Database.DBName)
		fmt.Printf("  SSL模式: %s\n", cfgFromFile.Database.SSLMode)
	}
	
	fmt.Println("\n数据库配置功能测试完成!")
}
