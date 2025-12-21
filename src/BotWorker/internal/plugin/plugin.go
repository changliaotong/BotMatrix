package plugin

import (
	"botworker/internal/onebot"
)

type Plugin interface {
	Name() string
	Description() string
	Version() string
	Init(robot Robot)
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
