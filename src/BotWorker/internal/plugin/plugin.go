package plugin

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Plugin interface {
	Name() string
	Description() string
	Version() string
	Init(robot Robot)
}

// Skill 插件提供的技能函数类型
type Skill func(params map[string]string) (string, error)

// SkillCapability 描述插件提供的技能
type SkillCapability struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Usage       string            `json:"usage"`
	Params      map[string]string `json:"params"`
}

// SkillCapable 插件可选实现的接口，用于报备技能
type SkillCapable interface {
	GetSkills() []SkillCapability
}

type Robot interface {
	OnMessage(fn onebot.EventHandler)
	OnNotice(fn onebot.EventHandler)
	OnRequest(fn onebot.EventHandler)
	OnEvent(eventName string, fn onebot.EventHandler)
	HandleAPI(action string, fn onebot.RequestHandler)
	SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error)
	DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error)
	SendLike(params *onebot.SendLikeParams) (*onebot.Response, error)
	SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error)
	SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error)
	GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error)
	GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error)
	SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error)
	GetSelfID() int64

	// Session & State Management
	GetSessionContext(platform, userID string) (*common.SessionContext, error)
	SetSessionState(platform, userID string, state common.SessionState, ttl time.Duration) error
	GetSessionState(platform, userID string) (*common.SessionState, error)
	ClearSessionState(platform, userID string) error

	// Task & Skill Management
	HandleSkill(skillName string, fn func(params map[string]string) (string, error))
}

type Manager struct {
	plugins []Plugin
	robot   Robot
}

func NewManager(robot Robot) *Manager {
	return &Manager{
		plugins: []Plugin{},
		robot:   robot,
	}
}

func (m *Manager) LoadPlugin(plugin Plugin) error {
	plugin.Init(m.robot)
	m.plugins = append(m.plugins, plugin)
	return nil
}

func (m *Manager) LoadPlugins(dir string) error {
	log.Printf("正在从目录加载插件: %s", dir)

	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // 目录不存在不报错，直接返回
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(dir, entry.Name())
		configPath := filepath.Join(pluginDir, "plugin.json")

		// 检查是否有 plugin.json
		if _, err := os.Stat(configPath); err == nil {
			log.Printf("发现插件目录: %s", entry.Name())
			// 这里未来可以扩展为加载动态链接库 (.so) 或 脚本插件
			// 目前原生插件主要通过 main.go 中的 loadAllPlugins 硬编码加载
		}
	}

	return nil
}

func (m *Manager) HandleEvent(msg map[string]any) {
	// 将 map 转换为 onebot.Event
	// 这里可以根据 msg 的内容构建 onebot.Event 对象
	// 然后分发给 robot (CombinedServer) 的 handleEvent 方法
	// 注意：CombinedServer 内部的 wsServer 也有 handleEvent，
	// 我们需要确保 Redis 队列的消息也能通过同样的逻辑分发。

	// 由于 CombinedServer 实现了 Robot 接口，但 HandleEvent 不是 Robot 接口的一部分
	// 我们直接在 CombinedServer 中实现 handleEvent 逻辑，或者让 Manager 负责分发。

	// 这里的 msg 是从 Redis 队列出来的原始 json map
	// 我们 need 将其传递给 CombinedServer 来处理
	if s, ok := m.robot.(interface{ HandleQueueEvent(map[string]any) }); ok {
		s.HandleQueueEvent(msg)
	}
}

func (m *Manager) GetPlugins() []Plugin {
	return m.plugins
}

func (m *Manager) GetPlugin(name string) Plugin {
	for _, plugin := range m.plugins {
		if plugin.Name() == name {
			return plugin
		}
	}
	return nil
}
