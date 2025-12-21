package main

import (
	"flag"
	"fmt"
	"log"
	"plugins"
)

func main() {
	// 解析命令行参数
	msg := flag.String("msg", "/ 积分", "要测试的消息")
	flag.Parse()

	// 创建命令解析器实例
	cmdParser := plugins.NewCommandParser()

	// 测试MatchCommand方法
	fmt.Println("=== 测试MatchCommand方法 ===")
	match, cmd := cmdParser.MatchCommand("points|积分", *msg)
	fmt.Printf("测试消息: '%s'\n", *msg)
	fmt.Printf("是否匹配积分命令: %t\n", match)
	fmt.Printf("匹配的命令: '%s'\n", cmd)

	// 测试MatchCommandWithSingleParam方法
	fmt.Println("\n=== 测试MatchCommandWithSingleParam方法 ===")
	match2, cmd2, param := cmdParser.MatchCommandWithSingleParam("猜拳|rock", *msg)
	fmt.Printf("测试消息: '%s'\n", *msg)
	fmt.Printf("是否匹配带参数命令: %t\n", match2)
	fmt.Printf("匹配的命令: '%s'\n", cmd2)
	fmt.Printf("参数: '%s'\n", param)

	// 测试IsCommand方法
	fmt.Println("\n=== 测试IsCommand方法 ===")
	isCmd := cmdParser.IsCommand("points|积分|猜拳|rock", *msg)
	fmt.Printf("测试消息: '%s'\n", *msg)
	fmt.Printf("是否为命令: %t\n", isCmd)

	// 测试GetCommandPrefix方法
	fmt.Println("\n=== 测试GetCommandPrefix方法 ===")
	prefix := cmdParser.GetCommandPrefix(*msg)
	fmt.Printf("测试消息: '%s'\n", *msg)
	fmt.Printf("命令前缀: '%s'\n", prefix)

	// 测试ExtractCommand方法
	fmt.Println("\n=== 测试ExtractCommand方法 ===")
	extractedCmd := cmdParser.ExtractCommand(*msg)
	fmt.Printf("测试消息: '%s'\n", *msg)
	fmt.Printf("提取的命令: '%s'\n", extractedCmd)
}
