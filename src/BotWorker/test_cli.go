package main

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/plugins"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// MockRobot 用于测试的模拟机器人
type MockRobot struct {
	messageHandlers []onebot.EventHandler
	eventHandlers   map[string][]onebot.EventHandler
}

// NewMockRobot 创建新的模拟机器人
func NewMockRobot() *MockRobot {
	return &MockRobot{
		messageHandlers: []onebot.EventHandler{},
		eventHandlers:   make(map[string][]onebot.EventHandler),
	}
}

// OnMessage 注册消息处理函数
func (m *MockRobot) OnMessage(fn onebot.EventHandler) {
	m.messageHandlers = append(m.messageHandlers, fn)
}

// OnNotice 注册通知处理函数
func (m *MockRobot) OnNotice(fn onebot.EventHandler) {
	m.eventHandlers["notice"] = append(m.eventHandlers["notice"], fn)
}

// OnRequest 注册请求处理函数
func (m *MockRobot) OnRequest(fn onebot.EventHandler) {
	m.eventHandlers["request"] = append(m.eventHandlers["request"], fn)
}

// OnEvent 注册事件处理函数
func (m *MockRobot) OnEvent(eventName string, fn onebot.EventHandler) {
	m.eventHandlers[eventName] = append(m.eventHandlers[eventName], fn)
}

// HandleAPI 注册API处理函数
func (m *MockRobot) HandleAPI(action string, fn onebot.RequestHandler) {
	// 不需要实现
}

// SendMessage 发送消息
func (m *MockRobot) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	message := ""
	for _, segment := range params.Message {
		if segment.Type == "text" {
			message += segment.Data["text"]
		} else {
			message += fmt.Sprintf("[%s]", segment.Type)
		}
	}
	
	target := ""
	if params.GroupID != 0 {
		target = fmt.Sprintf("群聊 %d", params.GroupID)
	} else {
		target = fmt.Sprintf("用户 %d", params.UserID)
	}
	
	fmt.Printf("[Bot -> %s] %s\n", target, message)
	return &onebot.Response{Retcode: 0, Status: "ok"}, nil
}

// DeleteMessage 删除消息
func (m *MockRobot) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	return &onebot.Response{Retcode: 0, Status: "ok"}, nil
}

// SendLike 发送点赞
func (m *MockRobot) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	return &onebot.Response{Retcode: 0, Status: "ok"}, nil
}

// SetGroupKick 踢人
func (m *MockRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	return &onebot.Response{Retcode: 0, Status: "ok"}, nil
}

// SetGroupBan 禁言
func (m *MockRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	return &onebot.Response{Retcode: 0, Status: "ok"}, nil
}

// 模拟发送消息给插件
func simulateMessage(robot *MockRobot, userID, groupID, message string) {
	// 创建模拟事件
	event := &onebot.Event{
		Time:        int64(strings.NewReader(message).Len()), // 简单使用消息长度作为时间戳
		SelfID:      123456,
		PostType:    "message",
		MessageType: "group",
		SubType:     "normal",
		MessageID:   123456789,
		GroupID:     groupID,
		UserID:      userID,
		RawMessage:  message,
		Message: []onebot.MessageSegment{
			{
				Type: "text",
				Data: map[string]string{
					"text": message,
				},
			},
		},
		Sender: onebot.Sender{
			UserID:   userID,
			Nickname: "测试用户",
			Card:     "测试用户",
			Sex:      "male",
			Age:      0,
			Area:     "",
			Level:    "",
			Role:     "member",
			Title:    "",
		},
	}

	fmt.Printf("[User -> Bot] %s\n", message)

	// 调用所有注册的消息处理函数
	for _, handler := range robot.messageHandlers {
		err := handler(event)
		if err != nil {
			log.Printf("处理消息时出错: %v", err)
		}
	}
}

func main() {
	// 解析命令行参数
	var userID int64
	var groupID int64
	var message string
	var interactive bool

	flag.Int64Var(&userID, "user", 123456789, "测试用户ID")
	flag.Int64Var(&groupID, "group", 987654321, "测试群组ID")
	flag.StringVar(&message, "msg", "", "测试消息内容")
	flag.BoolVar(&interactive, "i", false, "交互式测试模式")
	flag.Parse()

	// 创建模拟机器人
	mockRobot := NewMockRobot()

	// 创建插件管理器
	pluginManager := plugin.NewManager(mockRobot)

	// 加载所有插件（只加载不依赖外部服务的插件）
	log.Println("正在加载插件...")

	// 加载积分插件（使用内存实现）
	pointsPlugin := plugins.NewPointsPlugin()
	pluginManager.LoadPlugin(pointsPlugin)

	// 加载游戏插件
	gamesPlugin := plugins.NewGamesPlugin()
	pluginManager.LoadPlugin(gamesPlugin)

	// 加载工具插件
	utilsPlugin := plugins.NewUtilsPlugin()
	pluginManager.LoadPlugin(utilsPlugin)

	// 加载抽奖插件
	lotteryPlugin := plugins.NewLotteryPlugin()
	pluginManager.LoadPlugin(lotteryPlugin)

	// 加载宠物插件
	petPlugin := plugins.NewPetPlugin()
	pluginManager.LoadPlugin(petPlugin)

	// 加载问候插件
	greetingsPlugin := plugins.NewGreetingsPlugin()
	pluginManager.LoadPlugin(greetingsPlugin)

	// 加载菜单插件
	menuPlugin := plugins.NewMenuPlugin()
	pluginManager.LoadPlugin(menuPlugin)

	// 加载回显插件
	echoPlugin := plugins.NewEchoPlugin()
	pluginManager.LoadPlugin(echoPlugin)

	log.Printf("成功加载 %d 个插件\n", len(pluginManager.GetPlugins()))

	// 显示帮助信息
	showHelp()

	if message != "" {
		// 单次测试模式
		fmt.Printf("\n[测试消息] 用户ID: %d, 群组ID: %d\n", userID, groupID)
		simulateMessage(mockRobot, userID, groupID, message)
		return
	}

	if interactive {
		// 交互式测试模式
		fmt.Printf("\n[交互式测试模式] 用户ID: %d, 群组ID: %d\n", userID, groupID)
		fmt.Println("输入 'exit' 退出，'help' 显示帮助信息")
		
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			
			msg := scanner.Text()
			if strings.ToLower(msg) == "exit" {
				break
			}
			if strings.ToLower(msg) == "help" {
				showHelp()
				continue
			}
			
			simulateMessage(mockRobot, userID, groupID, msg)
		}
		
		if err := scanner.Err(); err != nil {
			log.Printf("读取输入失败: %v", err)
		}
		
		return
	}

	// 默认显示帮助信息
	fmt.Println("请指定测试消息或使用 -i 进入交互式测试模式")
	fmt.Println("使用 'go run test_cli.go -help' 查看所有选项")
}

// 显示帮助信息
func showHelp() {
	fmt.Println("\n=== 机器人插件测试工具 ===")
	fmt.Println("支持的命令示例：")
	fmt.Println("  / 积分                # 查询积分")
	fmt.Println("  / 签到                # 签到")
	fmt.Println("  / 猜拳 石头           # 猜拳游戏")
	fmt.Println("  / 天气 北京           # 查询天气")
	fmt.Println("  / 翻译 hello          # 翻译文本")
	fmt.Println("  / 点歌 周杰伦         # 点歌")
	fmt.Println("  / 计算 1+2*3          # 计算表达式")
	fmt.Println("  / 抽奖                # 抽奖")
	fmt.Println("  / 宠物领养 小狗       # 领养宠物")
	fmt.Println("  / 报时                # 查看当前时间")
	fmt.Println()
}
