package main

import (
	"log"
	"os"
)

func main() {
	log.Println("=== 测试日志输出 ===")
	log.Printf("当前工作目录: %s", func() string { dir, _ := os.Getwd(); return dir }())
	log.Println("环境变量PATH:", os.Getenv("PATH"))
	log.Println("=== 测试完成 ===")
}