package main

import (
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"botworker/plugins"
	"bufio"
	"database/sql"
	"encoding/json"
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

func printAPIRequest(action string, params interface{}) {
	req := onebot.Request{
		Action: action,
		Params: params,
	}
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		fmt.Printf("[Bot API] %s marshal error: %v\n", action, err)
		return
	}
	fmt.Printf("[Bot API] %s\n", string(data))
}

// SendMessage 发送消息
func (m *MockRobot) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	message := ""
	switch v := params.Message.(type) {
	case string:
		message = v
	case []onebot.MessageSegment:
		for _, segment := range v {
			if text, ok := segment.Data["text"].(string); ok && segment.Type == "text" {
				message += text
			} else {
				message += fmt.Sprintf("[%s]", segment.Type)
			}
		}
	case []interface{}:
		for _, seg := range v {
			if segment, ok := seg.(onebot.MessageSegment); ok {
				if text, ok := segment.Data["text"].(string); ok && segment.Type == "text" {
					message += text
				} else {
					message += fmt.Sprintf("[%s]", segment.Type)
				}
			} else {
				message += fmt.Sprintf("%v", seg)
			}
		}
	default:
		message = fmt.Sprintf("%v", params.Message)
	}

	target := ""
	if params.GroupID != 0 {
		target = fmt.Sprintf("群聊 %d", params.GroupID)
	} else {
		target = fmt.Sprintf("用户 %d", params.UserID)
	}

	fmt.Printf("[Bot -> %s] %s\n", target, message)
	return &onebot.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"message_id": 123456,
		},
	}, nil
}

// DeleteMessage 删除消息
func (m *MockRobot) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	printAPIRequest("delete_msg", params)
	return &onebot.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"message_id": params.MessageID,
		},
	}, nil
}

// SendLike 发送点赞
func (m *MockRobot) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	printAPIRequest("send_like", params)
	return &onebot.Response{
		Status: "ok",
	}, nil
}

// SetGroupKick 踢人
func (m *MockRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	printAPIRequest("set_group_kick", params)
	return &onebot.Response{
		Status: "ok",
	}, nil
}

// SetGroupBan 禁言
func (m *MockRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	printAPIRequest("set_group_ban", params)
	return &onebot.Response{
		Status: "ok",
	}, nil
}

func (m *MockRobot) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	printAPIRequest("get_group_member_list", params)
	return &onebot.Response{
		Status: "ok",
		Data:   []interface{}{},
	}, nil
}

func (m *MockRobot) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	printAPIRequest("get_group_member_info", params)
	role := "member"
	if params.UserID == 123456 {
		role = "owner"
	}
	data := map[string]interface{}{
		"user_id":        float64(params.UserID),
		"nickname":       "测试用户",
		"card":           "",
		"sex":            "unknown",
		"age":            float64(0),
		"join_time":      float64(0),
		"last_sent_time": float64(0),
		"level":          float64(1),
		"role":           role,
	}
	return &onebot.Response{
		Status: "ok",
		Data:   data,
	}, nil
}

func (m *MockRobot) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	printAPIRequest("set_group_special_title", params)
	return &onebot.Response{
		Status: "ok",
	}, nil
}

func (m *MockRobot) GetSelfID() int64 {
	return 123456789
}

// 模拟发送消息给插件
func simulateMessage(robot *MockRobot, userID, groupID int64, message string) {
	event := &onebot.Event{
		Time:        int64(strings.NewReader(message).Len()),
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
				Data: map[string]interface{}{
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
	var configPath string

	flag.Int64Var(&userID, "user", 123456789, "测试用户ID")
	flag.Int64Var(&groupID, "group", 987654321, "测试群组ID")
	flag.StringVar(&message, "msg", "", "测试消息内容")
	flag.BoolVar(&interactive, "i", false, "交互式测试模式")
	flag.StringVar(&configPath, "config", "config.json", "配置文件路径")
	flag.Parse()

	var database *sql.DB
	var cfg *config.Config
	if configPath != "" {
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			log.Printf("加载配置失败: %v", err)
		} else {
			cfg = loadedCfg
			database, err = db.NewDBConnection(&cfg.Database)
			if err != nil {
				log.Printf("警告: 无法连接到数据库: %v", err)
			} else {
				log.Println("成功连接到数据库")
				plugins.SetGlobalDB(database)
				if err := db.InitDatabase(database); err != nil {
					log.Printf("警告: 初始化数据库表失败: %v", err)
				}
			}
		}
	}
	if database != nil {
		defer database.Close()
	}

	// 创建模拟机器人
	mockRobot := NewMockRobot()

	// 创建插件管理器
	pluginManager := plugin.NewManager(mockRobot)

	// 加载所有插件（只加载不依赖外部服务的插件）
	log.Println("正在加载插件...")

	var pointsPlugin *plugins.PointsPlugin
	pointsPlugin = plugins.NewPointsPlugin(database)
	pluginManager.LoadPlugin(pointsPlugin)

	// 加载签到插件
	signInPlugin := plugins.NewSignInPlugin(pointsPlugin)
	pluginManager.LoadPlugin(signInPlugin)

	// 加载天气插件（如果配置可用）
	if cfg != nil {
		weatherPlugin := plugins.NewWeatherPlugin(&cfg.Weather)
		pluginManager.LoadPlugin(weatherPlugin)
	}

	// 加载翻译插件（如果配置可用）
	if cfg != nil {
		translatePlugin := plugins.NewTranslatePlugin(&cfg.Translate)
		pluginManager.LoadPlugin(translatePlugin)
	}

	// 加载游戏插件
	gamesPlugin := plugins.NewGamesPlugin()
	pluginManager.LoadPlugin(gamesPlugin)

	// 加载抽奖插件
	lotteryPlugin := plugins.NewLotteryPlugin()
	pluginManager.LoadPlugin(lotteryPlugin)

	// 加载点歌插件
	musicPlugin := plugins.NewMusicPlugin()
	pluginManager.LoadPlugin(musicPlugin)

	// 加载宠物插件
	petPlugin := plugins.NewPetPlugin(database, pointsPlugin)
	pluginManager.LoadPlugin(petPlugin)

	knowledgePlugin := plugins.NewKnowledgeBasePlugin(database, "")
	pluginManager.LoadPlugin(knowledgePlugin)

	// 加载欢迎插件
	welcomePlugin := plugins.NewWelcomePlugin()
	pluginManager.LoadPlugin(welcomePlugin)

	// 加载问候插件
	greetingsPlugin := plugins.NewGreetingsPlugin()
	pluginManager.LoadPlugin(greetingsPlugin)

	// 加载菜单插件
	menuPlugin := plugins.NewMenuPlugin()
	pluginManager.LoadPlugin(menuPlugin)

	// 加载回显插件
	echoPlugin := plugins.NewEchoPlugin()
	pluginManager.LoadPlugin(echoPlugin)

	// 加载群管插件
	var redisClient *redis.Client
	groupManagerPlugin := plugins.NewGroupManagerPlugin(database, redisClient)
	pluginManager.LoadPlugin(groupManagerPlugin)

	// 加载敏感词/风纪插件
	moderationPlugin := plugins.NewModerationPlugin()
	pluginManager.LoadPlugin(moderationPlugin)

	// 加载对话示例插件
	dialogDemoPlugin := plugins.NewDialogDemoPlugin()
	pluginManager.LoadPlugin(dialogDemoPlugin)

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
	fmt.Println("  撤回词+词1 词2        # 添加撤回词")
	fmt.Println("  撤回词-词1 词2        # 删除撤回词")
	fmt.Println("  扣分词+词1 词2        # 添加扣分词")
	fmt.Println("  扣分词-词1 词2        # 删除扣分词")
	fmt.Println("  警告词+词1 词2        # 添加警告词")
	fmt.Println("  警告词-词1 词2        # 删除警告词")
	fmt.Println("  禁言词+词1 词2        # 添加禁言词")
	fmt.Println("  禁言词-词1 词2        # 删除禁言词")
	fmt.Println("  踢出词+词1 词2        # 添加踢出词")
	fmt.Println("  踢出词-词1 词2        # 删除踢出词")
	fmt.Println("  拉黑词+词1 词2        # 添加拉黑词")
	fmt.Println("  拉黑词-词1 词2        # 删除拉黑词")
	fmt.Println()
}
