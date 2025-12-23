package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// PluginManagerPlugin 插件管理插件
type PluginManagerPlugin struct {
	db        *sql.DB
	cmdParser *CommandParser
}

func (p *PluginManagerPlugin) Name() string {
	return "plugin_manager"
}

func (p *PluginManagerPlugin) Description() string {
	return "插件管理功能，帮助用户选择开启哪些功能，类似应用商店"
}

func (p *PluginManagerPlugin) Version() string {
	return "1.0.0"
}

// NewPluginManagerPlugin 创建插件管理插件实例
func NewPluginManagerPlugin(database *sql.DB) *PluginManagerPlugin {
	return &PluginManagerPlugin{
		db:        database,
		cmdParser: NewCommandParser(),
	}
}

func (p *PluginManagerPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("插件管理插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载插件管理插件")

	// 处理插件商店命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "plugin_manager") {
				HandleFeatureDisabled(robot, event, "plugin_manager")
				return nil
			}
		}

		// 检查是否为插件商店命令
		if match, _ := p.cmdParser.MatchCommand("插件商店|功能商店", event.RawMessage); match {
			p.showPluginStore(robot, event)
			return nil
		}

		// 检查是否为启用插件命令
		match, cmd, params := p.cmdParser.MatchCommandWithParams("启用插件|开启插件", `(\S+)`, event.RawMessage)
		if match && len(params) == 1 {
			p.enablePlugin(robot, event, params[0])
			return nil
		}

		// 检查是否为禁用插件命令
		match, cmd, params = p.cmdParser.MatchCommandWithParams("禁用插件|关闭插件", `(\S+)`, event.RawMessage)
		if match && len(params) == 1 {
			p.disablePlugin(robot, event, params[0])
			return nil
		}

		// 检查是否为查看已启用插件命令
		if match, _ := p.cmdParser.MatchCommand("已启用插件|已开启功能", event.RawMessage); match {
			p.showEnabledPlugins(robot, event)
			return nil
		}

		return nil
	})
}

// showPluginStore 显示插件商店
func (p *PluginManagerPlugin) showPluginStore(robot plugin.Robot, event *onebot.Event) {
	if event.MessageType != "group" && event.MessageType != "private" {
		return
	}

	var groupIDStr string
	if event.MessageType == "group" {
		groupIDStr = fmt.Sprintf("%d", event.GroupID)
	}

	// 构建插件列表
	var pluginList strings.Builder
	pluginList.WriteString("📱 插件商店 📱\n")
	pluginList.WriteString("------------------------\n")

	// 按功能类型分类显示
	pluginList.WriteString("🎮 娱乐功能\n")
	entertainmentFeatures := []string{"tarot", "games", "music", "lottery", "pets", "fishing", "farm", "robbery", "cultivation", "gift"}
	for _, featureID := range entertainmentFeatures {
		p.addPluginToStoreList(&pluginList, groupIDStr, featureID)
	}

	pluginList.WriteString("\n💼 实用功能\n")
	utilityFeatures := []string{"weather", "translate", "points", "signin", "utils", "welcome", "greetings"}
	for _, featureID := range utilityFeatures {
		p.addPluginToStoreList(&pluginList, groupIDStr, featureID)
	}

	pluginList.WriteString("\n🛡️ 群管功能\n")
	moderationFeatures := []string{"moderation", "kick_to_black", "kick_notify", "leave_to_black", "leave_notify", "join_mute"}
	for _, featureID := range moderationFeatures {
		p.addPluginToStoreList(&pluginList, groupIDStr, featureID)
	}

	pluginList.WriteString("\n⚙️ 系统功能\n")
	systemFeatures := []string{"plugin_manager", "feature_disabled_notice", "voice_reply", "burn_after_reading"}
	for _, featureID := range systemFeatures {
		p.addPluginToStoreList(&pluginList, groupIDStr, featureID)
	}

	pluginList.WriteString("\n------------------------\n")
	pluginList.WriteString("命令格式：\n")
	pluginList.WriteString("- 启用插件 <功能名称>\n")
	pluginList.WriteString("- 禁用插件 <功能名称>\n")
	pluginList.WriteString("- 已启用插件\n")

	// 发送插件列表
	p.sendMessage(robot, event, pluginList.String())
}

// addPluginToStoreList 将插件添加到商店列表
func (p *PluginManagerPlugin) addPluginToStoreList(list *strings.Builder, groupIDStr, featureID string) {
	// 检查功能是否存在
	displayName, ok := FeatureDisplayNames[featureID]
	if !ok {
		return
	}

	// 检查功能是否有默认设置
	_, hasDefault := FeatureDefaults[featureID]
	if !hasDefault {
		return
	}

	// 获取当前状态
	enabled := IsFeatureEnabledForGroup(p.db, groupIDStr, featureID)
	status := "❌ 已禁用"
	if enabled {
		status = "✅ 已启用"
	}

	list.WriteString(fmt.Sprintf("%s %s\n", status, displayName))
}

// enablePlugin 启用插件
func (p *PluginManagerPlugin) enablePlugin(robot plugin.Robot, event *onebot.Event, featureName string) {
	if event.MessageType != "group" {
		p.sendMessage(robot, event, "插件管理功能仅支持群聊使用")
		return
	}

	// 查找功能ID
	featureID := p.findFeatureIDByName(featureName)
	if featureID == "" {
		p.sendMessage(robot, event, fmt.Sprintf("未找到功能：%s", featureName))
		return
	}

	// 检查功能是否可配置
	_, hasDefault := FeatureDefaults[featureID]
	if !hasDefault {
		p.sendMessage(robot, event, fmt.Sprintf("功能 %s 不支持配置", featureName))
		return
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 设置功能为启用
	err := db.SetGroupFeatureOverride(p.db, groupIDStr, featureID, true)
	if err != nil {
		log.Printf("启用功能失败: %v", err)
		p.sendMessage(robot, event, fmt.Sprintf("启用功能 %s 失败，请稍后再试", featureName))
		return
	}

	p.sendMessage(robot, event, fmt.Sprintf("✅ 已成功启用功能：%s", featureName))
}

// disablePlugin 禁用插件
func (p *PluginManagerPlugin) disablePlugin(robot plugin.Robot, event *onebot.Event, featureName string) {
	if event.MessageType != "group" {
		p.sendMessage(robot, event, "插件管理功能仅支持群聊使用")
		return
	}

	// 查找功能ID
	featureID := p.findFeatureIDByName(featureName)
	if featureID == "" {
		p.sendMessage(robot, event, fmt.Sprintf("未找到功能：%s", featureName))
		return
	}

	// 检查功能是否可配置
	_, hasDefault := FeatureDefaults[featureID]
	if !hasDefault {
		p.sendMessage(robot, event, fmt.Sprintf("功能 %s 不支持配置", featureName))
		return
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	// 设置功能为禁用
	err := db.SetGroupFeatureOverride(p.db, groupIDStr, featureID, false)
	if err != nil {
		log.Printf("禁用功能失败: %v", err)
		p.sendMessage(robot, event, fmt.Sprintf("禁用功能 %s 失败，请稍后再试", featureName))
		return
	}

	p.sendMessage(robot, event, fmt.Sprintf("✅ 已成功禁用功能：%s", featureName))
}

// showEnabledPlugins 显示已启用的插件
func (p *PluginManagerPlugin) showEnabledPlugins(robot plugin.Robot, event *onebot.Event) {
	if event.MessageType != "group" && event.MessageType != "private" {
		return nil
	}

	var groupIDStr string
	if event.MessageType == "group" {
		groupIDStr = fmt.Sprintf("%d", event.GroupID)
	}

	// 构建已启用插件列表
	var enabledList strings.Builder
	enabledList.WriteString("✅ 已启用功能列表 ✅\n")
	enabledList.WriteString("------------------------\n")

	// 按功能类型分类显示
	enabledList.WriteString("🎮 娱乐功能\n")
	entertainmentFeatures := []string{"tarot", "games", "music", "lottery", "pets", "fishing", "farm", "robbery", "cultivation", "gift"}
	p.addEnabledPluginToList(&enabledList, groupIDStr, entertainmentFeatures)

	enabledList.WriteString("\n💼 实用功能\n")
	utilityFeatures := []string{"weather", "translate", "points", "signin", "utils", "welcome", "greetings"}
	p.addEnabledPluginToList(&enabledList, groupIDStr, utilityFeatures)

	enabledList.WriteString("\n🛡️ 群管功能\n")
	moderationFeatures := []string{"moderation", "kick_to_black", "kick_notify", "leave_to_black", "leave_notify", "join_mute"}
	p.addEnabledPluginToList(&enabledList, groupIDStr, moderationFeatures)

	enabledList.WriteString("\n⚙️ 系统功能\n")
	systemFeatures := []string{"plugin_manager", "feature_disabled_notice", "voice_reply", "burn_after_reading"}
	p.addEnabledPluginToList(&enabledList, groupIDStr, systemFeatures)

	enabledList.WriteString("\n------------------------\n")
	enabledList.WriteString("使用命令管理功能：\n")
	enabledList.WriteString("- 插件商店：查看所有可用功能\n")
	enabledList.WriteString("- 启用插件 <功能名称>：开启功能\n")
	enabledList.WriteString("- 禁用插件 <功能名称>：关闭功能\n")

	// 发送已启用插件列表
	p.sendMessage(robot, event, enabledList.String())

	return nil
}

// addEnabledPluginToList 将已启用的插件添加到列表
func (p *PluginManagerPlugin) addEnabledPluginToList(list *strings.Builder, groupIDStr string, featureIDs []string) {
	for _, featureID := range featureIDs {
		// 检查功能是否存在
		displayName, ok := FeatureDisplayNames[featureID]
		if !ok {
			continue
		}

		// 检查功能是否有默认设置
		_, hasDefault := FeatureDefaults[featureID]
		if !hasDefault {
			continue
		}

		// 检查功能是否已启用
		if IsFeatureEnabledForGroup(p.db, groupIDStr, featureID) {
			list.WriteString(fmt.Sprintf("✅ %s\n", displayName))
		}
	}
}

// findFeatureIDByName 根据功能名称查找功能ID
func (p *PluginManagerPlugin) findFeatureIDByName(featureName string) string {
	featureName = strings.TrimSpace(featureName)
	if featureName == "" {
		return ""
	}

	// 直接匹配功能ID
	if _, ok := FeatureDefaults[featureName]; ok {
		return featureName
	}

	// 匹配功能显示名称
	for featureID, displayName := range FeatureDisplayNames {
		if strings.EqualFold(displayName, featureName) {
			return featureID
		}
	}

	return ""
}

// sendMessage 发送消息
func (p *PluginManagerPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
