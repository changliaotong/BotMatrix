package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// EventMessage 事件消息结构
type EventMessage struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

// Action 动作结构
type Action struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	TargetID string `json:"target_id"`
	Text     string `json:"text"`
}

// ResponseMessage 响应消息结构
type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

// ReadMessage 从stdin读取消息
func ReadMessage() (*EventMessage, error) {
	decoder := json.NewDecoder(os.Stdin)
	var msg EventMessage
	if err := decoder.Decode(&msg); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("连接关闭")
		}
		return nil, fmt.Errorf("解析消息失败: %v", err)
	}
	return &msg, nil
}

// WriteMessage 向stdout写入消息
func WriteMessage(msg *ResponseMessage) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	return nil
}

// generateUUID 生成简单的UUID（简化版）
func generateUUID() string {
	// 这里实现一个简单的UUID生成器
	// 实际项目中应使用标准库
	return "temp-uuid-1234"
}