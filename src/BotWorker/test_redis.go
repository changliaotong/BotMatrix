package main

import (
	"botworker/internal/config"
	"botworker/internal/redis"
	"context"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 设置命令行参数
	redisHost := flag.String("redis-host", "localhost", "Redis服务器地址")
	redisPort := flag.Int("redis-port", 6379, "Redis服务器端口")
	redisPassword := flag.String("redis-password", "", "Redis密码")
	redisDB := flag.Int("redis-db", 0, "Redis数据库编号")
	flag.Parse()

	// 创建测试配置
	cfg := &config.RedisConfig{
		Host:     *redisHost,
		Port:     *redisPort,
		Password: *redisPassword,
		DB:       *redisDB,
	}

	fmt.Printf("测试Redis配置: Host=%s, Port=%d, Password='%s', DB=%d\n", cfg.Host, cfg.Port, cfg.Password, cfg.DB)

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 打印更多调试信息
	fmt.Printf("尝试连接到Redis服务器: %s:%d\n", cfg.Host, cfg.Port)

	// 尝试连接Redis
	startTime := time.Now()
	client, err := redis.NewClient(cfg)
	elapsedTime := time.Since(startTime)

	fmt.Printf("连接尝试耗时: %v\n", elapsedTime)

	if err != nil {
		log.Printf("错误: 无法连接到Redis服务器: %v\n", err)
		log.Println("诊断信息:")
		log.Println("- 请确保Redis服务器正在运行")
		log.Println("- 检查Redis服务器地址和端口是否正确")
		log.Println("- 检查Redis密码是否正确（如果设置了密码）")
		log.Println("- 检查防火墙设置是否允许连接")
		return
	}

	log.Println("成功: 连接到Redis服务器")

	// 测试基本操作
	fmt.Println("测试基本Redis操作...")
	key := "test_key"
	value := "test_value"

	// 设置键值对
	if err := client.Set(ctx, key, value, 10*time.Second).Err(); err != nil {
		log.Printf("错误: 设置键值对失败: %v\n", err)
	} else {
		log.Printf("成功: 设置键 %s = %s\n", key, value)

		// 获取键值对
		if val, err := client.Get(ctx, key).Result(); err != nil {
			log.Printf("错误: 获取键值对失败: %v\n", err)
		} else {
			log.Printf("成功: 获取键 %s = %s\n", key, val)
		}
	}

	// 关闭连接
	fmt.Println("关闭Redis连接...")
	if err := client.Close(); err != nil {
		log.Printf("错误: 关闭Redis连接时出错: %v\n", err)
	} else {
		log.Println("成功: 关闭Redis连接")
	}

	fmt.Println("Redis配置测试完成")
}
