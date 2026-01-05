package plugins

import (
	"BotMatrix/common/log"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/plugin/generator"
	"fmt"
	"os"
	"strings"
)

type BuilderPlugin struct {
	robot      core.Robot
	pluginPath string
}

func NewBuilderPlugin(pluginPath string) *BuilderPlugin {
	return &BuilderPlugin{
		pluginPath: pluginPath,
	}
}

func (p *BuilderPlugin) Name() string        { return "PluginBuilder" }
func (p *BuilderPlugin) Description() string { return "Build plugins using natural language via chat" }
func (p *BuilderPlugin) Version() string     { return "1.0.0" }

func (p *BuilderPlugin) Init(robot core.Robot) {
	p.robot = robot
	// 我们通过 HandleAPI 来监听事件，或者通过技能。
	// 但内部插件最直接的方式是让 Robot 转发所有事件给它。
	// 在我们的实现中，PluginBridge 会分发事件。

	// 我们注册一个技能，或者直接监听消息。
	// 这里我们监听 API "on_message" (如果 Robot 支持)
	robot.HandleAPI("on_message", p.handleMessage)
}

func (p *BuilderPlugin) handleMessage(event map[string]any) {
	text, _ := event["text"].(string)
	if text == "" {
		return
	}

	// 匹配触发词
	var prompt string
	triggers := []string{"帮我写一个", "生成插件", "帮我做一个", "开发插件"}
	for _, t := range triggers {
		if strings.HasPrefix(text, t) {
			prompt = strings.TrimPrefix(text, t)
			break
		}
	}

	if prompt == "" {
		return
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return
	}

	// 安全检查：仅允许管理员使用
	userID, _ := event["user_id"].(string)
	if userID == "" {
		userID, _ = event["from"].(string) // 兼容性处理
	}

	adminEnv := strings.TrimSpace(os.Getenv("BM_ADMIN_USERS"))
	if adminEnv != "" {
		isAdmin := false
		admins := strings.Split(adminEnv, ",")
		for _, admin := range admins {
			if strings.TrimSpace(admin) == userID {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			log.Printf("[PluginBuilder] Security: Non-admin user %s tried to generate plugin", userID)
			p.reply(event, "抱歉，只有管理员才能使用插件生成功能。")
			return
		}
	}

	log.Printf("[PluginBuilder] User %s requested plugin: %s", userID, prompt)

	// 回复用户开始生成
	p.reply(event, "正在为你生成插件，请稍候... 🤖\n这可能需要 10-20 秒，我会使用 AI 为你编写代码并自动部署。")

	// 异步生成，避免阻塞主循环
	go func() {
		// 调用生成逻辑 (默认使用 python)
		result, err := generator.GeneratePlugin(prompt, "python")
		if err != nil {
			p.reply(event, fmt.Sprintf("抱歉，生成失败了: %v", err))
			return
		}

		// 保存插件
		dir, err := generator.SavePlugin(result, p.pluginPath)
		if err != nil {
			p.reply(event, fmt.Sprintf("保存插件失败: %v", err))
			return
		}

		p.reply(event, fmt.Sprintf("✨ 插件「%s」已生成并上线！\n\n版本: %s\n作者: %s\n\n你可以现在就开始测试它了。如果需要修改，请随时告诉我。",
			result.Manifest["name"], result.Manifest["version"], result.Manifest["author"]))
		log.Printf("[PluginBuilder] Plugin saved to %s", dir)
	}()
}

func (p *BuilderPlugin) reply(event map[string]any, text string) {
	target, _ := event["from"].(string)
	groupID, _ := event["group_id"].(string)

	params := map[string]any{
		"target":    target,
		"target_id": groupID,
		"text":      text,
	}

	p.robot.CallBotAction("send_msg", params)
}
