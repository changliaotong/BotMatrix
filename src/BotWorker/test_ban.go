package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.0.126:6379",
		Password: "redis_zsYik8",
		DB:       0,
	})

	event := map[string]interface{}{
		"post_type":    "message",
		"message_type": "group",
		"sub_type":     "normal",
		"message_id":   fmt.Sprintf("test_ban_%d", time.Now().Unix()),
		"group_id":     "123456",
		"user_id":      "10001",
		"message":      "禁言我",
		"raw_message":  "禁言我",
		"font":         0,
		"self_id":      "2958935140",
		"platform":     "worker",
		"time":         time.Now().Unix(),
		"sender": map[string]interface{}{
			"user_id":  "10001",
			"nickname": "Tester",
			"role":     "member",
		},
	}

	payload, _ := json.Marshal(event)

	msg := map[string]interface{}{
		"id":      fmt.Sprintf("ob_%d", time.Now().Unix()),
		"payload": string(payload),
	}

	err := rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "botmatrix:queue:default",
		Values: msg,
	}).Err()

	if err != nil {
		fmt.Printf("Error pushing message: %v\n", err)
	} else {
		fmt.Println("Message pushed successfully")
	}
}
