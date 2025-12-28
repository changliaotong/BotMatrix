package common

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseV11Message 解析 OneBot v11 的 CQ 码消息为结构化段
func ParseV11Message(raw string) []MessageSegment {
	var segments []MessageSegment
	
	// 正则匹配 CQ 码: [CQ:type,key=value,key2=value2]
	re := regexp.MustCompile(`\[CQ:([a-zA-Z0-9]+)((?:,[a-zA-Z0-9_\-\.]+=?[^,\]]*)*)\]`)
	
	lastIndex := 0
	matches := re.FindAllStringSubmatchIndex(raw, -1)
	
	for _, match := range matches {
		// 处理 CQ 码之前的文本
		if match[0] > lastIndex {
			text := raw[lastIndex:match[0]]
			segments = append(segments, MessageSegment{
				Type: "text",
				Data: TextSegmentData{Text: UnescapeCQ(text)},
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
		
		segments = append(segments, MessageSegment{
			Type: cqType,
			Data: data,
		})
		
		lastIndex = match[1]
	}
	
	// 处理剩余的文本
	if lastIndex < len(raw) {
		text := raw[lastIndex:]
		segments = append(segments, MessageSegment{
			Type: "text",
			Data: TextSegmentData{Text: UnescapeCQ(text)},
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
func BuildV11Message(segments []MessageSegment) string {
	var sb strings.Builder
	for _, seg := range segments {
		if seg.Type == "text" {
			if text, ok := seg.Data["text"].(string); ok {
				sb.WriteString(EscapeCQ(text))
			}
		} else {
			sb.WriteString(fmt.Sprintf("[CQ:%s", seg.Type))
			for k, v := range seg.Data {
				sb.WriteString(fmt.Sprintf(",%s=%s", k, EscapeCQValue(fmt.Sprint(v))))
			}
			sb.WriteString("]")
		}
	}
	return sb.String()
}

// ToV11Map 将内部消息转换为 OneBot v11 兼容的 map 格式
func (m *InternalMessage) ToV11Map() map[string]any {
	res := make(map[string]any)

	// 1. 加入 Extras 中的字段
	for k, v := range m.Extras {
		res[k] = v
	}

	// 2. 用强类型字段覆盖/补充核心字段
	res["time"] = m.Time
	res["self_id"] = m.SelfID
	res["post_type"] = m.PostType
	if m.PostType == "" {
		res["post_type"] = "message"
	}

	if m.PostType == "message" {
		res["message_type"] = m.MessageType
		res["user_id"] = m.UserID
		res["message"] = BuildV11Message(m.Message)
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
func (m *InternalMessage) ToV12Map() map[string]any {
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

