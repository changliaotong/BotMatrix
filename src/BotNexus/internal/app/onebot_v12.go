package app

import (
	"BotMatrix/common/onebot"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	clog "BotMatrix/common/log"
	"encoding/json"
	"strconv"
	"time"

	"github.com/botuniverse/go-libonebot"
	"go.uber.org/zap"
)

// v12ToInternal 将 OneBot v12 消息转换为内部标准格式
func (m *Manager) v12ToInternal(v12Msg onebot.V12RawMessage) types.InternalMessage {
	var segments []types.MessageSegment
	if len(v12Msg.Message) > 0 {
		var rawSegments []any
		if err := json.Unmarshal(v12Msg.Message, &rawSegments); err == nil {
			for _, s := range rawSegments {
				if segMap, ok := s.(map[string]any); ok {
					segType, _ := segMap["type"].(string)
					segData, _ := segMap["data"].(map[string]any)
					segments = append(segments, types.MessageSegment{
						Type: segType,
						Data: segData,
					})
				}
			}
		}
	}

	extras := make(map[string]any)
	if len(v12Msg.Data) > 0 {
		var data any
		if err := json.Unmarshal(v12Msg.Data, &data); err == nil {
			extras["data"] = data
		}
	}

	return types.InternalMessage{
		ID:          v12Msg.ID,
		Time:        utils.ToInt64(v12Msg.Time),
		Platform:    v12Msg.Platform,
		SelfID:      v12Msg.SelfID,
		Protocol:    "v12",
		PostType:    v12Msg.Type,
		MessageType: v12Msg.DetailType,
		UserID:      v12Msg.UserID,
		GroupID:     v12Msg.GroupID,
		Message:     segments,
		Echo:        utils.ToString(v12Msg.Echo),
		Status:      utils.ToString(v12Msg.Status),
		Msg:         v12Msg.Msg,
		Retcode:     int(utils.ToInt64(v12Msg.Retcode)),
		MetaType:    v12Msg.MetaEventType,
		SubType:     v12Msg.SubType,
		SenderName:  v12Msg.User.Nickname,
		Extras:      extras,
	}
}

// initOneBotActions 注册 OneBot v12 标准动作
func (m *Manager) initOneBotActions() {
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

// v11ToInternal 将 OneBot v11 消息转换为内部标准格式
func (m *Manager) v11ToInternal(v11Msg onebot.V11RawMessage, platform string) types.InternalMessage {
	userID := utils.ToString(v11Msg.UserID)
	groupID := utils.ToString(v11Msg.GroupID)
	selfID := utils.ToString(v11Msg.SelfID)

	var segments []types.MessageSegment
	var rawMessage string

	switch v := v11Msg.Message.(type) {
	case string:
		rawMessage = v
		// 使用解析器解析 CQ 码
		segments = onebot.ParseV11Message(rawMessage)
	case []any:
		// Handle array of segments if provided (some v11 implementations do this)
		if b, err := json.Marshal(v); err == nil {
			json.Unmarshal(b, &segments)
		}
	}

	return types.InternalMessage{
		ID:          utils.ToString(v11Msg.MessageID),
		Time:        timeToUnix(v11Msg.Time),
		Platform:    platform,
		SelfID:      selfID,
		Protocol:    "v11",
		PostType:    v11Msg.PostType,
		MessageType: v11Msg.MessageType,
		UserID:      userID,
		GroupID:     groupID,
		Message:     segments,
		RawMessage:  rawMessage,
		SenderName:  v11Msg.Sender.Nickname,
		SubType:     v11Msg.SubType,
	}
}

func timeToUnix(t any) int64 {
	switch v := t.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case string:
		ts, _ := strconv.ParseInt(v, 10, 64)
		return ts
	default:
		return time.Now().Unix()
	}
}
