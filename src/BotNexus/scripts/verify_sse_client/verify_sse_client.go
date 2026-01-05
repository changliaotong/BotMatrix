package main

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

func main() {
	// 这是一个简单的 SSE 客户端，用于验证 BotNexus 的 MCP SSE 接口
	// 假设 BotNexus 运行在 localhost:8080
	url := "http://localhost:8080/api/mcp/v1/sse"
	fmt.Printf("Connecting to SSE endpoint: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Connected! Waiting for events...")
	scanner := bufio.NewScanner(resp.Body)
	
	// 设置超时，防止挂起
	timer := time.After(10 * time.Second)

	for {
		select {
		case <-timer:
			fmt.Println("Test finished (timeout).")
			return
		default:
			if scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					fmt.Printf("Received: %s\n", line)
				}
			} else {
				if err := scanner.Err(); err != nil {
					fmt.Printf("Scanner error: %v\n", err)
				}
				return
			}
		}
	}
}
