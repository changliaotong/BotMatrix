package types

import (
	"fmt"
	"strings"
)

// MessageSegment represents a structured message segment (OneBot v12 compatible)
type MessageSegment struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// TextSegmentData represents the data for a text message segment
type TextSegmentData struct {
	Text string `json:"text"`
}

// ImageSegmentData represents the data for an image message segment
type ImageSegmentData struct {
	File string `json:"file"`
	URL  string `json:"url,omitempty"`
}

// InternalMessage is the unified message format used within BotMatrix
type InternalMessage struct {
	ID          string           `json:"id"`           // Message ID
	Time        int64            `json:"time"`         // Timestamp
	Platform    string           `json:"platform"`     // qq, wechat, etc.
	SelfID      string           `json:"self_id"`      // Bot ID
	Protocol    string           `json:"protocol"`     // v11, v12, etc.
	PostType    string           `json:"post_type"`    // message, notice, request, meta_event
	MessageType string           `json:"message_type"` // private, group
	SubType     string           `json:"sub_type"`     // friend, normal, etc.
	UserID      string           `json:"user_id"`      // Sender ID
	GroupID     string           `json:"group_id"`     // Group ID (if applicable)
	GroupName   string           `json:"group_name"`   // Group Name (if applicable)
	Message     []MessageSegment `json:"message"`      // Structured message
	RawMessage  string           `json:"raw_message"`  // Original raw message string
	SenderName  string           `json:"sender_name"`  // Sender nickname
	SenderCard  string           `json:"sender_card"`  // Sender card/alias in group
	UserAvatar  string           `json:"user_avatar"`  // User avatar URL
	Echo        string           `json:"echo"`         // Echo for tracking
	Status      any              `json:"status"`       // ok, failed
	Retcode     int              `json:"retcode"`      // OneBot return code
	Msg         string           `json:"msg"`          // Error message or info
	MetaType    string           `json:"meta_type"`    // heartbeat, lifecycle
	Extras      map[string]any   `json:"extras"`       // Additional platform-specific fields
}

// InternalAction is the unified action format used within BotMatrix
type InternalAction struct {
	Action   string         `json:"action"`
	Params   map[string]any `json:"params"`
	Echo     string         `json:"echo"`
	SelfID   string         `json:"self_id,omitempty"`
	Platform string         `json:"platform,omitempty"`

	// Common fields to avoid map usage
	UserID      string `json:"user_id,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	DetailType  string `json:"detail_type,omitempty"`
	Message     any    `json:"message,omitempty"`
}

// ToV11Map converts internal message to OneBot v11 compatible map
func (m *InternalMessage) ToV11Map() map[string]any {
	res := make(map[string]any)

	// 1. Start with extras
	if m.Extras != nil {
		for k, v := range m.Extras {
			res[k] = v
		}
	}

	// 2. Core fields
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

		msgStr := m.BuildV11String()
		if msgStr == "" && m.RawMessage != "" {
			msgStr = m.RawMessage
		}
		res["message"] = msgStr
		res["raw_message"] = m.RawMessage
		res["message_id"] = m.ID
		if m.GroupID != "" {
			res["group_id"] = m.GroupID
		}
		if m.ID != "" {
			res["message_id"] = m.ID
		}

		// Sender info
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
		res["sender"] = sender
	}

	return res
}

// ToV12Map converts internal message to OneBot v12 compatible map
func (m *InternalMessage) ToV12Map() map[string]any {
	res := make(map[string]any)

	// 1. Join Extras
	if m.Extras != nil {
		for k, v := range m.Extras {
			res[k] = v
		}
	}

	// 2. Core fields
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

// BuildV11String converts structured segments to CQ code string
func (m *InternalMessage) BuildV11String() string {
	var sb strings.Builder
	for _, seg := range m.Message {
		if seg.Type == "text" {
			if d, ok := seg.Data.(TextSegmentData); ok {
				sb.WriteString(escapeCQ(d.Text))
			} else if d, ok := seg.Data.(*TextSegmentData); ok {
				sb.WriteString(escapeCQ(d.Text))
			} else if d, ok := seg.Data.(map[string]any); ok {
				if text, ok := d["text"].(string); ok {
					sb.WriteString(escapeCQ(text))
				}
			} else if s, ok := seg.Data.(string); ok {
				sb.WriteString(escapeCQ(s))
			}
		} else {
			sb.WriteString(fmt.Sprintf("[CQ:%s", seg.Type))
			if d, ok := seg.Data.(map[string]any); ok {
				for k, v := range d {
					sb.WriteString(fmt.Sprintf(",%s=%v", k, escapeCQValue(fmt.Sprint(v))))
				}
			} else if seg.Type == "image" {
				if d, ok := seg.Data.(ImageSegmentData); ok {
					sb.WriteString(fmt.Sprintf(",file=%s", escapeCQValue(d.File)))
				} else if d, ok := seg.Data.(*ImageSegmentData); ok {
					sb.WriteString(fmt.Sprintf(",file=%s", escapeCQValue(d.File)))
				}
			}
			sb.WriteString("]")
		}
	}
	return sb.String()
}

func escapeCQ(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "[", "&#91;")
	s = strings.ReplaceAll(s, "]", "&#93;")
	return s
}

func escapeCQValue(s string) string {
	s = escapeCQ(s)
	s = strings.ReplaceAll(s, ",", "&#44;")
	return s
}
