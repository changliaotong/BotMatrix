package main

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/plugins"
	"fmt"
	"log"
	"os"
)

// MockRobot 实现插件所需的Robot接口
// 用于测试群管理命令的处理逻辑

type MockRobot struct{}

func (m *MockRobot) SendMessage(event *onebot.Event, message string) error {
	fmt.Printf("[发送消息] %s\n", message)
	return nil
}

func (m *MockRobot) SendMessageToGroup(groupID int64, message string) error {
	fmt.Printf("[发送群消息] 群ID: %d, 消息: %s\n", groupID, message)
	return nil
}

func (m *MockRobot) SendMessageToPrivate(userID int64, message string) error {
	fmt.Printf("[发送私聊消息] 用户ID: %d, 消息: %s\n", userID, message)
	return nil
}

func (m *MockRobot) OnMessage(handler func(event *onebot.Event) error) {
	fmt.Println("[注册消息处理函数]")
}

func (m *MockRobot) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.CommonResponse, error) {
	// 模拟群主信息
	if params.UserID == 123456 {
		return &onebot.CommonResponse{
			Status: "ok",
			Data: map[string]interface{}{
				"role": "owner", // 群主
			},
		}, nil
	}
	// 模拟普通成员
	return &onebot.CommonResponse{
		Status: "ok",
		Data: map[string]interface{}{
			"role": "member", // 普通成员
		},
	}, nil
}

func (m *MockRobot) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.CommonResponse, error) {
	fmt.Printf("[设置群头衔] 群ID: %d, 用户ID: %d, 头衔: %s\n", params.GroupID, params.UserID, params.SpecialTitle)
	return &onebot.CommonResponse{Status: "ok"}, nil
}

func (m *MockRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.CommonResponse, error) {
	fmt.Printf("[踢人] 群ID: %d, 用户ID: %d, 拒绝加群: %t\n", params.GroupID, params.UserID, params.RejectAddRequest)
	return &onebot.CommonResponse{Status: "ok"}, nil
}

func (m *MockRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.CommonResponse, error) {
	fmt.Printf("[禁言] 群ID: %d, 用户ID: %d, 时长: %d秒\n", params.GroupID, params.UserID, params.Duration)
	return &onebot.CommonResponse{Status: "ok"}, nil
}

func (m *MockRobot) SetGroupUnban(params *onebot.SetGroupUnbanParams) (*onebot.CommonResponse, error) {
	fmt.Printf("[解除禁言] 群ID: %d, 用户ID: %d\n", params.GroupID, params.UserID)
	return &onebot.CommonResponse{Status: "ok"}, nil
}

func (m *MockRobot) GetConfig() plugin.RobotConfig {
	return plugin.RobotConfig{}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("=== 群管理插件测试开始 ===")

	// 创建群管理插件实例
	groupManagerPlugin := plugins.NewGroupManagerPlugin(nil, nil)

	// 初始化插件
	mockRobot := &MockRobot{}
	groupManagerPlugin.Init(mockRobot)

	log.Println("群管理插件初始化完成")

	// 测试命令解析
	fmt.Println("\n=== 命令解析测试 ===")

	// 测试1: 设置头衔命令
	fmt.Println("\n测试1: 设置头衔命令")
	testCommand("/settitle 123456 测试头衔", 123456, 789012)

	// 测试2: 踢人命令
	fmt.Println("\n测试2: 踢人命令")
	testCommand("/kick 123456", 789012, 123456)

	// 测试3: 禁言命令
	fmt.Println("\n测试3: 禁言命令")
	testCommand("/ban 123456 60", 789012, 123456)

	// 测试4: 解除禁言命令
	fmt.Println("\n测试4: 解除禁言命令")
	testCommand("/unban 123456", 789012, 123456)

	// 测试5: 添加管理员命令
	fmt.Println("\n测试5: 添加管理员命令")
	testCommand("/addadmin 123456", 789012, 123456)

	// 测试6: 删除管理员命令
	fmt.Println("\n测试6: 删除管理员命令")
	testCommand("/deladmin 123456", 789012, 123456)

	log.Println("\n=== 群管理插件测试结束 ===")
}

// testCommand 测试命令解析
func testCommand(message string, userID, groupID int64) {
	// 创建模拟事件
	event := &onebot.Event{
		Type:        "message",
		MessageType: "group",
		UserID:      userID,
		GroupID:     groupID,
		RawMessage:  message,
		Message:     message,
	}

	fmt.Printf("输入命令: %s\n", message)

	// 简单的命令解析测试
	// 由于我们没有运行完整的插件系统，这里只测试基本的命令格式
	if len(message) > 0 && message[0] == '/' {
		parts := splitCommand(message[1:])
		if len(parts) > 0 {
			cmd := parts[0]
			params := parts[1:]
			fmt.Printf("解析结果: 命令=%s, 参数=%v\n", cmd, params)
		}
	}
}

// splitCommand 简单的命令解析函数
func splitCommand(cmd string) []string {
	var parts []string
	var current string
	inQuotes := false
	
	for i, char := range cmd {
		switch char {
		case ' ':
			if !inQuotes {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		case '"':
			inQuotes = !inQuotes
		default:
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}