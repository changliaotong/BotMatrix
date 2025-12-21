package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("=== 测试输出程序 ===")
	fmt.Printf("当前时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("当前工作目录: %s\n", func() string { dir, _ := os.Getwd(); return dir }())
	fmt.Println("输出测试完成")
}
