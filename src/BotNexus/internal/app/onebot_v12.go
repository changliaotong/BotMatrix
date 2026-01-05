package app

import (
	clog "BotMatrix/common/log"

	"github.com/botuniverse/go-libonebot"
	"go.uber.org/zap"
)

// initOneBotActions 注册 OneBot v12 标准动作
func (m *Manager) initOneBotActions() {
	if m.OneBot == nil {
		return
	}
	// 注册核心动作处理器
	m.OneBot.HandleFunc(func(w libonebot.ResponseWriter, r *libonebot.Request) {
		clog.Debug("Received OneBot v12 Action", zap.String("action", r.Action))

		// 这里将来对接 BotMatrix 的核心调度逻辑
		// 目前先简单记录日志
		switch r.Action {
		case "send_message":
			m.handleV12SendMessage(w, r)
		default:
			clog.Warn("Unsupported OneBot v12 Action", zap.String("action", r.Action))
		}
	})
}

// handleV12SendMessage 处理 v12 的发消息动作
func (m *Manager) handleV12SendMessage(w libonebot.ResponseWriter, r *libonebot.Request) {
	// 具体的 v12 发送逻辑
}
