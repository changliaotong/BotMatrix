package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// 直接使用标准库日志
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("=== 超简单测试程序开始 ===")
	log.Printf("当前工作目录: %s", func() string { dir, _ := os.Getwd(); return dir }())
	log.Printf("当前程序路径: %s", os.Args[0])
	log.Printf("环境变量GOOS: %s", os.Getenv("GOOS"))
	log.Printf("环境变量GOARCH: %s", os.Getenv("GOARCH"))

	// 检查文件是否存在
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("读取当前目录失败: %v", err)
	}

	log.Println("当前目录文件:")
	for _, file := range files {
		if !file.IsDir() {
			log.Printf("- %s (大小: %d字节)", file.Name(), func() int64 { info, _ := file.Info(); return info.Size() }())
		}
	}

	fmt.Println("\n=== 测试完成 ===")
	log.Println("=== 超简单测试程序结束 ===")
}