package core

import (
	"BotMatrix/common/models"
)

// BaseContext 封装了插件执行时的基础上下文信息
// 插件无需自己去查询这些基础数据，由系统统一注入
type BaseContext interface {
	// BotInfo 获取当前机器人信息
	BotInfo() *BotInfo
	// Group 获取当前群组信息（如果是私聊则返回 nil）
	Group() *models.Sz84Group
	// Member 获取当前发送者在群内的成员信息
	Member() *models.Sz84GroupMember
	// User 获取当前发送者的全局用户信息（含积分）
	User() *models.Sz84User
	// Store 获取底层存储，用于执行高级数据库操作
	Store() *models.Sz84Store
}

// BotInfo 描述机器人的基础信息
type BotInfo struct {
	Uin      int64
	Nickname string
	Platform string // "qq", "wechat", etc.
}

// ContextProvider 定义了如何获取上述上下文
type ContextProvider interface {
	GetBaseContext(botUin int64, groupID int64, userID int64) (BaseContext, error)
}
