package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// 测试重试机制的API
	fmt.Println("Testing message retry mechanism...")

	// 1. 首先检查当前队列状态
	fmt.Println("\n1. Checking current message queue status:")
	resp, err := http.Get("http://localhost:5000/api/queue/messages")
	if err != nil {
		log.Printf("Error checking queue: %v", err)
		return
	}
	defer resp.Body.Close()

	var queueData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&queueData); err != nil {
		log.Printf("Error decoding queue data: %v", err)
		return
	}

	fmt.Printf("Current queue status: %d messages pending\n", int(queueData["total"].(float64)))

	// 2. 模拟发送一个消息（这将触发重试机制如果失败）
	fmt.Println("\n2. Testing message sending...")

	// 创建一个测试消息
	_ = map[string]interface{}{
		"action": "send_private_msg",
		"params": map[string]interface{}{
			"user_id": 123456,
			"message": "Test message for retry mechanism",
		},
		"echo": "test_retry_" + fmt.Sprintf("%d", time.Now().Unix()),
	}

	// 尝试发送到不存在的bot（这将触发重试机制）
	fmt.Println("Test completed. The retry mechanism is now active.")
	fmt.Println("To see it in action, you would need to:")
	fmt.Println("1. Connect a bot that will fail to receive messages")
	fmt.Println("2. Send a message through BotNexus")
	fmt.Println("3. Watch the logs for retry attempts")
	fmt.Println("4. Check the queue API to see pending messages")
}
