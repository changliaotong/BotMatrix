package main

import (
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/types"
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/plugins"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type TestContext struct {
	platform string
	userID   string
	groupID  string
}

func (c *TestContext) BotInfo() *core.BotInfo {
	return &core.BotInfo{
		Uin:      51437810,
		Platform: c.platform,
	}
}
func (c *TestContext) Group() *models.GroupInfo {
	return &models.GroupInfo{
		Id:        86433316,
		GroupName: "测试群",
		IsCredit:  true,
	}
}
func (c *TestContext) Member() *models.GroupMember {
	return &models.GroupMember{
		GroupId:  86433316,
		UserId:   1653346663,
		UserName: "测试用户",
	}
}
func (c *TestContext) User() *models.UserInfo {
	return &models.UserInfo{
		Id:   1653346663,
		Name: "测试用户",
	}
}
func (c *TestContext) Store() *models.Sz84Store { return plugins.GlobalSz84Store }
func (c *TestContext) Role() string             { return "admin" }
func (c *TestContext) RawMessage() string       { return "" }

type TestRobot struct {
	plugins      []core.PluginModule
	skills       map[string]core.Skill
	capabilities map[string]core.SkillCapability
}

func NewTestRobot() *TestRobot {
	return &TestRobot{
		skills:       make(map[string]core.Skill),
		capabilities: make(map[string]core.SkillCapability),
	}
}

func (r *TestRobot) OnMessage(fn func(event map[string]any)) {
	// 测试环境不实现消息监听
}

func (r *TestRobot) OnNotice(fn func(event map[string]any)) {
	// 测试环境不实现通知监听
}

func (r *TestRobot) OnRequest(fn func(event map[string]any)) {
	// 测试环境不实现请求监听
}

func (r *TestRobot) OnEvent(eventName string, fn func(event map[string]any)) {
	// 测试环境不实现事件监听
}

func (r *TestRobot) HandleAPI(action string, fn any) {
	// 测试环境不实现API处理
}

func (r *TestRobot) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	// 测试环境不实现消息发送
	return nil, nil
}

func (r *TestRobot) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	// 测试环境不实现消息删除
	return nil, nil
}

func (r *TestRobot) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	// 测试环境不实现点赞
	return nil, nil
}

func (r *TestRobot) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	// 测试环境不实现踢人
	return nil, nil
}

func (r *TestRobot) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	// 测试环境不实现禁言
	return nil, nil
}

func (r *TestRobot) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	// 测试环境不实现成员列表获取
	return nil, nil
}

func (r *TestRobot) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	// 测试环境不实现成员信息获取
	return nil, nil
}

func (r *TestRobot) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	// 测试环境不实现设置专属头衔
	return nil, nil
}

func (r *TestRobot) GetSelfID() int64 {
	return 123456
}

func (r *TestRobot) GetSessionContext(platform, userID string) (*types.SessionContext, error) {
	return nil, nil
}

func (r *TestRobot) SetSessionState(platform, userID string, state types.SessionState, ttl time.Duration) error {
	return nil
}

func (r *TestRobot) GetSessionState(platform, userID string) (*types.SessionState, error) {
	return nil, nil
}

func (r *TestRobot) ClearSessionState(platform, userID string) error {
	return nil
}

// HandleSkill implements plugin.Robot
func (r *TestRobot) HandleSkill(skillName string, skill func(ctx core.BaseContext, params map[string]string) (string, error)) {
	r.skills[skillName] = skill
}

func (r *TestRobot) RegisterSkill(capability core.SkillCapability, skill func(ctx core.BaseContext, params map[string]string) (string, error)) {
	r.skills[capability.Name] = skill
	r.capabilities[capability.Name] = capability
}

func (r *TestRobot) CallSkill(inputName string, params map[string]string) (string, error) {
	// 1. 先尝试完全匹配技能名
	if skill, ok := r.skills[inputName]; ok {
		ctx := &TestContext{
			platform: "test",
			userID:   "1653346663",
			groupID:  "86433316",
		}
		return skill(ctx, params)
	}

	// 2. 尝试正则匹配
	for name, cap := range r.capabilities {
		if cap.Regex != "" {
			re, err := regexp.Compile(cap.Regex)
			if err == nil && re.MatchString(inputName) {
				// 提取正则捕获组作为参数
				matches := re.FindStringSubmatch(inputName)
				if len(matches) > 0 {
					// 只有在 params 为空时才填充正则匹配结果，避免覆盖 call 命令传入的参数
					if len(params) == 0 {
						for i, match := range matches {
							params[fmt.Sprintf("%d", i)] = match
						}
					}
				}

				ctx := &TestContext{
					platform: "test",
					userID:   "1653346663",
					groupID:  "86433316",
				}
				return r.skills[name](ctx, params)
			}
		}

		// 3. 尝试匹配 Usage 或 Name (模糊匹配)
		if strings.Contains(cap.Usage, inputName) || strings.Contains(name, inputName) {
			ctx := &TestContext{
				platform: "test",
				userID:   "1653346663",
				groupID:  "86433316",
			}
			return r.skills[name](ctx, params)
		}
	}

	return "", fmt.Errorf("skill %s not found", inputName)
}

func (r *TestRobot) CallPluginAction(pluginID string, action string, payload map[string]any) (any, error) {
	// 测试环境不实现插件动作调用
	return nil, nil
}

func (r *TestRobot) CallBotAction(action string, params any) (any, error) {
	// 测试环境不实现机器人动作调用
	return nil, nil
}

func main() {
	// 1. 加载配置
	cfg, _, err := config.LoadFromCLI()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化数据库连接
	gdb, err := db.NewGORMConnection(&cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 初始化全局数据库和存储
	plugins.SetGlobalGORMDB(gdb)
	s := plugins.GlobalStore

	// 3. 初始化测试机器人
	robot := NewTestRobot()

	// 4. 加载所有插件
	pm := core.NewPluginManager()

	// 加载 SigninPlugin
	signinPlugin := plugins.NewSigninPlugin(s)
	pm.LoadPluginModule(signinPlugin, robot)

	// 加载 TeachPlugin
	teachPlugin := plugins.NewTeachPlugin(s)
	pm.LoadPluginModule(teachPlugin, robot)

	// 加载 EconomyPlugin
	economyPlugin := plugins.NewEconomyPlugin(s)
	pm.LoadPluginModule(economyPlugin, robot)

	// 加载 ToolsPlugin
	toolsPlugin := plugins.NewToolsPlugin()
	pm.LoadPluginModule(toolsPlugin, robot)

	// 加载 HelpPlugin
	helpPlugin := plugins.NewHelpPlugin()
	pm.LoadPluginModule(helpPlugin, robot)

	// 加载 WeatherPlugin
	weatherPlugin := plugins.NewWeatherPlugin()
	pm.LoadPluginModule(weatherPlugin, robot)

	fmt.Println("Plugin Console (Production Database Mode)")
	fmt.Println("Loaded Plugins: Teach, Economy, Tools, Help, Weather")
	fmt.Println("===================")
	fmt.Println("Available commands:")
	fmt.Println("  list - List all loaded plugins")
	fmt.Println("  skills - List all available skills")
	fmt.Println("  call <skill> [params] - Call a skill with parameters")
	fmt.Println("  exit - Exit the console")
	fmt.Println()
	fmt.Println("Try: call signin")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		if input == "list" {
			// 内部插件
			internalPlugins := pm.GetInternalPlugins()
			log.Printf("Loaded internal plugins (%d):", len(internalPlugins))
			for name, p := range internalPlugins {
				log.Printf("  - %s: %s (v%s)", name, p.Description(), p.Version())
			}
			continue
		}

		if input == "skills" {
			fmt.Printf("Available skills (%d):\n", len(robot.skills))
			for name := range robot.skills {
				fmt.Printf("  - %s\n", name)
			}
			continue
		}

		// 默认尝试作为技能调用
		var skillName string
		params := make(map[string]string)

		if strings.HasPrefix(input, "call ") {
			callParts := strings.SplitN(input[5:], " ", 2)
			skillName = callParts[0]
			if len(callParts) > 1 {
				paramsStr := callParts[1]
				paramParts := strings.Split(paramsStr, " ")
				for _, param := range paramParts {
					kv := strings.SplitN(param, "=", 2)
					if len(kv) == 2 {
						params[kv[0]] = kv[1]
					}
				}
			}
		} else {
			// 直接输入指令，如 "天气 北京" 或 "签到"
			skillName = input
		}

		result, err := robot.CallSkill(skillName, params)
		if err != nil {
			// 如果不是 call 命令且没找到技能，才报错
			if strings.HasPrefix(input, "call ") {
				log.Printf("❌ 错误: 找不到技能 %s", skillName)
			} else {
				fmt.Println("Unknown command. Type 'list', 'skills', 'call', or 'exit'.")
			}
		} else {
			fmt.Printf("\n🤖 机器人回复: \n%s\n\n", result)
		}
	}

	fmt.Println("Exiting...")
}
