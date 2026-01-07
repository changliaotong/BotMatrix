package server

import (
	"BotMatrix/common/models"
	"BotMatrix/common/plugin/core"
)

type WorkerContext struct {
	botUin   int64
	groupID  int64
	userID   int64
	platform string
	role     string
	rawMsg   string
}

func NewWorkerContext(botUin, groupID, userID int64, platform, role, rawMsg string) *WorkerContext {
	return &WorkerContext{
		botUin:   botUin,
		groupID:  groupID,
		userID:   userID,
		platform: platform,
		role:     role,
		rawMsg:   rawMsg,
	}
}

func (c *WorkerContext) BotInfo() *core.BotInfo {
	return nil // TODO: 实现
}

func (c *WorkerContext) Group() *models.GroupInfo {
	return nil // TODO: 实现
}

func (c *WorkerContext) Member() *models.GroupMember {
	return nil // TODO: 实现
}

func (c *WorkerContext) User() *models.UserInfo {
	return nil // TODO: 实现
}

func (c *WorkerContext) Store() *models.Sz84Store {
	return nil // TODO: 实现
}

func (c *WorkerContext) Role() string {
	return c.role
}

func (c *WorkerContext) RawMessage() string {
	return c.rawMsg
}
