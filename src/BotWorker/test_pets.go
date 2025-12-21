package main

import (
	"botworker/plugins"
	"fmt"
	"time"
)

// MockRobot 模拟机器人接口
type MockRobot struct {}

func (m *MockRobot) SendPrivateMessage(userID, message string) error {
	fmt.Printf("发送私聊消息给%s: %s\n", userID, message)
	return nil
}

func (m *MockRobot) SendGroupMessage(groupID, message string) error {
	fmt.Printf("发送群消息到%s: %s\n", groupID, message)
	return nil
}

func (m *MockRobot) OnMessage(handler func(event interface{}) error) error {
	// 模拟消息事件
	go func() {
		// 测试领养宠物
		fmt.Println("\n=== 测试领养宠物 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!领养"})
		
		// 等待1秒
		time.Sleep(1 * time.Second)
		
		// 测试查看宠物
		fmt.Println("\n=== 测试查看宠物 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!我的宠物"})
		
		// 等待1秒
		time.Sleep(1 * time.Second)
		
		// 测试喂食
		fmt.Println("\n=== 测试喂食 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!喂食 1"})
		
		// 等待1秒
		time.Sleep(1 * time.Second)
		
		// 测试玩耍
		fmt.Println("\n=== 测试玩耍 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!玩耍 1"})
		
		// 等待1秒
		time.Sleep(1 * time.Second)
		
		// 测试洗澡
		fmt.Println("\n=== 测试洗澡 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!洗澡 1"})
		
		// 等待1秒
		time.Sleep(1 * time.Second)
		
		// 再次查看宠物
		fmt.Println("\n=== 再次查看宠物 ===")
		handler(&MockEvent{MessageType: "private", UserID: "user123", RawMessage: "!我的宠物"})
	}()
	
	return nil
}

func (m *MockRobot) OnNotice(handler func(event interface{}) error) error {
	return nil
}

func (m *MockRobot) OnRequest(handler func(event interface{}) error) error {
	return nil
}

// MockEvent 模拟事件
type MockEvent struct {
	MessageType string
	UserID      string
	GroupID     string
	RawMessage  string
}

func main() {
	// 创建宠物插件
	petPlugin := plugins.NewPetPlugin()
	
	// 初始化插件
	petPlugin.Init(&MockRobot{})
	
	// 保持程序运行
	select {}
}