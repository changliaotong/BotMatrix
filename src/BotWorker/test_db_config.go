package main

import (
	"botworker/internal/config"
	"botworker/internal/db"
	"flag"
	"fmt"
	"log"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "test_config.json", "配置文件路径")
	flag.Parse()
	
	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 打印数据库配置
	fmt.Println("数据库配置:")
	fmt.Printf("  主机: %s\n", cfg.Database.Host)
	fmt.Printf("  端口: %d\n", cfg.Database.Port)
	fmt.Printf("  用户名: %s\n", cfg.Database.User)
	fmt.Printf("  密码: %s\n", cfg.Database.Password)
	fmt.Printf("  数据库名: %s\n", cfg.Database.DBName)
	fmt.Printf("  SSL模式: %s\n", cfg.Database.SSLMode)
	
	// 测试数据库连接
	fmt.Println("\n测试数据库连接...")
	database, err := db.NewDBConnection(&cfg.Database)
	if err != nil {
		fmt.Printf("警告: 无法连接到数据库: %v\n", err)
		fmt.Println("数据库功能将不可用")
	} else {
		fmt.Println("成功连接到数据库!")
		// 关闭数据库连接
		database.Close()
	}
}
