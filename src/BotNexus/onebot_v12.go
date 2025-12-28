package main

import (
	"BotMatrix/common"
	clog "BotMatrix/common/log"
	"strconv"
	"time"

	"github.com/botuniverse/go-libonebot"
	"go.uber.org/zap"
)

// v12ToInternal 将 OneBot v12 消息转换为内部标准格式
func (m *Manager) v12ToInternal(v12Msg common.V12RawMessage) common.InternalMessage {
	var segments []common.MessageSegment
	if len(v12Msg.Message) > 0 {
		var rawSegments []any
		if err := json.Unmarshal(v12Msg.Message, &rawSegments); err == nil {
			for _, s := range rawSegments {
				if segMap, ok := s.(map[string]any); ok {
					segType, _ := segMap["type"].(string)
					segData, _ := segMap["data"].(map[string]any)
					segments = append(segments, common.MessageSegment{
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

	return common.InternalMessage{
		ID:          v12Msg.ID,
		Time:        int64(v12Msg.Time),
		Platform:    v12Msg.Platform,
		SelfID:      v12Msg.SelfID,
		Protocol:    "v12",
		PostType:    v12Msg.Type,
		MessageType: v12Msg.DetailType,
		UserID:      v12Msg.UserID,
		GroupID:     v12Msg.GroupID,
		Message:     segments,
		Echo:        common.ToString(v12Msg.Echo),
		Status:      v12Msg.Status,
		Msg:         v12Msg.Msg,
		Retcode:     v12Msg.Retcode,
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
func (m *Manager) v11ToInternal(v11Msg common.V11RawMessage) common.InternalMessage {
	userID := common.ToString(v11Msg.UserID)
	groupID := common.ToString(v11Msg.GroupID)
	selfID := common.ToString(v11Msg.SelfID)

	var rawMessage string
	switch v := v11Msg.Message.(type) {
	case string:
		rawMessage = v
	case []any:
		// Handle array of segments if provided (some v11 implementations do this)
		if b, err := json.Marshal(v); err == nil {
			rawMessage = string(b)
		}
	}

	// 使用解析器解析 CQ 码
	segments := common.ParseV11Message(rawMessage)

	extras := make(map[string]any)
	if len(v11Msg.Data) > 0 {
		var data any
		if err := json.Unmarshal(v11Msg.Data, &data); err == nil {
			extras["data"] = data
		}
	}

	return common.InternalMessage{
		ID:          common.ToString(v11Msg.Echo), // v11 sometimes uses echo as ID for responses
		Time:        timeToUnix(v11Msg.Time),
		Platform:    "qq", // v11 默认为 qq
		SelfID:      selfID,
		Protocol:    "v11",
		PostType:    v11Msg.PostType,
		MessageType: v11Msg.MessageType,
		UserID:      userID,
		GroupID:     groupID,
		Message:     segments,
		RawMessage:  rawMessage,
		Echo:        common.ToString(v11Msg.Echo),
		Status:      v11Msg.Status,
		Msg:         v11Msg.Msg,
		Retcode:     v11Msg.Retcode,
		MetaType:    v11Msg.MetaEventType,
		SubType:     v11Msg.SubType,
		SenderName:  v11Msg.Sender.Nickname,
		SenderCard:  v11Msg.Sender.Card,
		UserAvatar:  v11Msg.Sender.Avatar,
		Extras:      extras,
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
