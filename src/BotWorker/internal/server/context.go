package server

import (
	"BotMatrix/common/models"
	"BotMatrix/common/plugin/core"
)

// DefaultBaseContext is a basic implementation of core.BaseContext
type DefaultBaseContext struct {
	botInfo *core.BotInfo
	group   *models.Sz84Group
	member  *models.Sz84GroupMember
	user    *models.Sz84User
	store   *models.Sz84Store
}

func NewBaseContext(botInfo *core.BotInfo, group *models.Sz84Group, member *models.Sz84GroupMember, user *models.Sz84User, store *models.Sz84Store) *DefaultBaseContext {
	return &DefaultBaseContext{
		botInfo: botInfo,
		group:   group,
		member:  member,
		user:    user,
		store:   store,
	}
}

func (c *DefaultBaseContext) BotInfo() *core.BotInfo {
	return c.botInfo
}

func (c *DefaultBaseContext) Group() *models.Sz84Group {
	return c.group
}

func (c *DefaultBaseContext) Member() *models.Sz84GroupMember {
	return c.member
}

func (c *DefaultBaseContext) User() *models.Sz84User {
	return c.user
}

func (c *DefaultBaseContext) Store() *models.Sz84Store {
	return c.store
}
