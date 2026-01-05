package onebot

import (
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// V12ToInternal 将 OneBot v12 消息转换为内部标准格式
func V12ToInternal(v12Msg V12RawMessage) types.InternalMessage {
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

// V11ToInternal 将 OneBot v11 消息转换为内部标准格式
func V11ToInternal(v11Msg V11RawMessage, platform string) types.InternalMessage {
	userID := utils.ToString(v11Msg.UserID)
	groupID := utils.ToString(v11Msg.GroupID)
	selfID := utils.ToString(v11Msg.SelfID)

	var segments []types.MessageSegment
	var rawMessage string

	switch v := v11Msg.Message.(type) {
	case string:
		rawMessage = v
		// 使用解析器解析 CQ 码
		segments = ParseV11Message(rawMessage)
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

// ParseV11Message 解析 OneBot v11 的 CQ 码消息为结构化段
func ParseV11Message(raw string) []types.MessageSegment {
	var segments []types.MessageSegment

	// 正则匹配 CQ 码: [CQ:type,key=value,key2=value2]
	re := regexp.MustCompile(`\[CQ:([a-zA-Z0-9]+)((?:,[a-zA-Z0-9_\-\.]+=?[^,\]]*)*)\]`)

	lastIndex := 0
	matches := re.FindAllStringSubmatchIndex(raw, -1)

	for _, match := range matches {
		// 处理 CQ 码之前的文本
		if match[0] > lastIndex {
			text := raw[lastIndex:match[0]]
			segments = append(segments, types.MessageSegment{
				Type: "text",
				Data: types.TextSegmentData{Text: UnescapeCQ(text)},
			})
		}

		// 处理 CQ 码
		cqType := raw[match[2]:match[3]]
		paramsStr := raw[match[4]:match[5]]

		data := make(map[string]any)
		if paramsStr != "" {
			// 分解参数 ,key=value
			paramPairs := strings.Split(paramsStr, ",")
			for _, pair := range paramPairs {
				if pair == "" {
					continue
				}
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					data[kv[0]] = UnescapeCQValue(kv[1])
				} else if len(kv) == 1 {
					data[kv[0]] = ""
				}
			}
		}

		segments = append(segments, types.MessageSegment{
			Type: cqType,
			Data: data,
		})

		lastIndex = match[1]
	}

	// 处理剩余的文本
	if lastIndex < len(raw) {
		text := raw[lastIndex:]
		segments = append(segments, types.MessageSegment{
			Type: "text",
			Data: types.TextSegmentData{Text: UnescapeCQ(text)},
		})
	}

	return segments
}

// UnescapeCQ 反转义 CQ 码中的特殊字符
func UnescapeCQ(s string) string {
	s = strings.ReplaceAll(s, "&#91;", "[")
	s = strings.ReplaceAll(s, "&#93;", "]")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}

// UnescapeCQValue 反转义 CQ 码参数值中的特殊字符
func UnescapeCQValue(s string) string {
	s = strings.ReplaceAll(s, "&#44;", ",")
	return UnescapeCQ(s)
}

// EscapeCQ 转义 CQ 码中的特殊字符
func EscapeCQ(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "[", "&#91;")
	s = strings.ReplaceAll(s, "]", "&#93;")
	return s
}

// EscapeCQValue 转义 CQ 码参数值中的特殊字符
func EscapeCQValue(s string) string {
	s = EscapeCQ(s)
	s = strings.ReplaceAll(s, ",", "&#44;")
	return s
}

// BuildV11Message 将结构化段转换为 OneBot v11 的 CQ 码字符串
func BuildV11Message(segments []types.MessageSegment) string {
	var sb strings.Builder
	for _, seg := range segments {
		data, ok := seg.Data.(map[string]any)
		if !ok {
			// Check for structured types
			switch v := seg.Data.(type) {
			case types.TextSegmentData:
				sb.WriteString(EscapeCQ(v.Text))
				continue
			case *types.TextSegmentData:
				sb.WriteString(EscapeCQ(v.Text))
				continue
			case types.ImageSegmentData:
				sb.WriteString(fmt.Sprintf("[CQ:image,file=%s]", EscapeCQValue(v.File)))
				continue
			case *types.ImageSegmentData:
				sb.WriteString(fmt.Sprintf("[CQ:image,file=%s]", EscapeCQValue(v.File)))
				continue
			}

			// 如果不是 map，尝试作为 string 处理 (v12 可能有简单形式)
			if s, ok := seg.Data.(string); ok && seg.Type == "text" {
				sb.WriteString(EscapeCQ(s))
			}
			continue
		}

		if seg.Type == "text" {
			if text, ok := data["text"].(string); ok {
				sb.WriteString(EscapeCQ(text))
			}
		} else {
			sb.WriteString(fmt.Sprintf("[CQ:%s", seg.Type))
			for k, v := range data {
				sb.WriteString(fmt.Sprintf(",%s=%s", k, EscapeCQValue(fmt.Sprint(v))))
			}
			sb.WriteString("]")
		}
	}
	return sb.String()
}

// ConvertLegacyPlaceholders 将旧版的占位符（如 [Face6.gif], [@:12345], [Image:xxx]）转换为标准 CQ 码
func ConvertLegacyPlaceholders(raw string) string {
	if raw == "" {
		return raw
	}

	// 1. [Face6.gif] -> [CQ:face,id=6]
	reFace := regexp.MustCompile(`\[Face(\d+)\.gif\]`)
	res := reFace.ReplaceAllString(raw, "[CQ:face,id=$1]")

	// 2. [@:12345] -> [CQ:at,qq=12345]
	reAt := regexp.MustCompile(`\[@:(\d+)\]`)
	res = reAt.ReplaceAllString(res, "[CQ:at,qq=$1]")

	// 3. [Image:xxx] -> [CQ:image,file=xxx]
	reImage := regexp.MustCompile(`\[Image:([^\]]+)\]`)
	res = reImage.ReplaceAllString(res, "[CQ:image,file=$1]")

	return res
}

// ToV11Map 将内部消息转换为 OneBot v11 兼容的 map 格式
func ToV11Map(m *types.InternalMessage) map[string]any {
	res := make(map[string]any)

	// 1. 加入 Extras 中的字段
	for k, v := range m.Extras {
		res[k] = v
	}

	// 2. 用强类型字段覆盖/补充核心字段
	res["time"] = m.Time
	res["self_id"] = m.SelfID
	if m.Platform != "" {
		res["platform"] = m.Platform
	}
	res["post_type"] = m.PostType
	if m.PostType == "" {
		res["post_type"] = "message"
	}

	if m.PostType == "message" {
		res["message_type"] = m.MessageType
		res["user_id"] = m.UserID

		msg := BuildV11Message(m.Message)
		if msg == "" && m.RawMessage != "" {
			msg = m.RawMessage
		}
		res["message"] = msg
		res["raw_message"] = m.RawMessage
		if m.GroupID != "" {
			res["group_id"] = m.GroupID
		}
		if m.ID != "" {
			res["message_id"] = m.ID
		}
		if m.GroupName != "" {
			res["group_name"] = m.GroupName
		}

		// 补充发送者信息
		sender := make(map[string]any)
		if s, ok := res["sender"].(map[string]any); ok {
			for k, v := range s {
				sender[k] = v
			}
		}
		if m.SenderName != "" {
			sender["nickname"] = m.SenderName
		}
		if m.SenderCard != "" {
			sender["card"] = m.SenderCard
		}
		if m.UserAvatar != "" {
			sender["avatar"] = m.UserAvatar
		}
		if len(sender) > 0 {
			res["sender"] = sender
		}
	} else if m.PostType == "meta_event" {
		res["meta_event_type"] = m.MetaType
	}

	if m.SubType != "" {
		res["sub_type"] = m.SubType
	}

	if m.Echo != "" {
		res["echo"] = m.Echo
	}

	if m.Status != "" {
		res["status"] = m.Status
	}

	return res
}

// ToV12Map 将内部消息转换为 OneBot v12 兼容的 map 格式
func ToV12Map(m *types.InternalMessage) map[string]any {
	res := make(map[string]any)

	// 1. 加入 Extras 中的字段
	for k, v := range m.Extras {
		res[k] = v
	}

	// 2. 用强类型字段覆盖
	res["id"] = m.ID
	res["time"] = float64(m.Time)
	res["type"] = m.PostType
	if m.PostType == "" {
		res["type"] = "message"
	}

	if res["type"] == "message" {
		res["detail_type"] = m.MessageType
		res["platform"] = m.Platform
		res["self_id"] = m.SelfID
		res["user_id"] = m.UserID
		if m.GroupID != "" {
			res["group_id"] = m.GroupID
		}

		segments := make([]map[string]any, len(m.Message))
		for i, seg := range m.Message {
			segments[i] = map[string]any{
				"type": seg.Type,
				"data": seg.Data,
			}
		}
		res["message"] = segments
	}

	if m.Echo != "" {
		res["echo"] = m.Echo
	}

	return res
}
