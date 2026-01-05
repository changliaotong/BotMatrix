package onebot

import (
	"encoding/json"
)

// ==================== OneBot v11 Raw Types ====================

// V11RawMessage represents a raw OneBot v11 message/event
type V11RawMessage struct {
	PostType      string          `json:"post_type"`
	MessageType   string          `json:"message_type"`
	MessageID     any             `json:"message_id"`
	Time          any             `json:"time"`
	SelfID        any             `json:"self_id"` // Can be string or int64
	SubType       string          `json:"sub_type"`
	UserID        any             `json:"user_id"`  // Can be string or int64
	GroupID       any             `json:"group_id"` // Can be string or int64
	Message       any             `json:"message"`  // Can be string or []MessageSegment
	RawMessage    string          `json:"raw_message"`
	Font          int             `json:"font"`
	Sender        V11RawSender    `json:"sender"`
	MetaEventType string          `json:"meta_event_type"`
	NoticeType    string          `json:"notice_type"`
	RequestType   string          `json:"request_type"`
	Status        any             `json:"status"`
	Retcode       any             `json:"retcode"`
	Data          json.RawMessage `json:"data"`
	Echo          any             `json:"echo"`
	Msg           string          `json:"msg"`
	Wording       string          `json:"wording"`
}

// V11RawSender represents sender info in OneBot v11
type V11RawSender struct {
	UserID   any    `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
	Card     string `json:"card"`
	Level    string `json:"level"`
	Role     string `json:"role"`
	Title    string `json:"title"`
	Avatar   string `json:"avatar"`
}

// ==================== OneBot v12 Raw Types ====================

// V12RawMessage represents a raw OneBot v12 message/event
type V12RawMessage struct {
	ID            string          `json:"id"`
	Time          any             `json:"time"`
	Type          string          `json:"type"`
	DetailType    string          `json:"detail_type"`
	SubType       string          `json:"sub_type"`
	SelfID        string          `json:"self_id"`
	Platform      string          `json:"platform"`
	UserID        string          `json:"user_id"`
	GroupID       string          `json:"group_id"`
	GuildID       string          `json:"guild_id"`
	ChannelID     string          `json:"channel_id"`
	Message       json.RawMessage `json:"message"`
	AltMessage    string          `json:"alt_message"`
	User          V12RawUser      `json:"user"`
	MetaEventType string          `json:"meta_event_type"`
	Status        any             `json:"status"`
	Retcode       any             `json:"retcode"`
	Data          json.RawMessage `json:"data"`
	Echo          any             `json:"echo"`
	Msg           string          `json:"msg"`
}

// V12RawUser represents user info in OneBot v12
type V12RawUser struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}
