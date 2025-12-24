package core

// OneBot Event Structure
type OneBotEvent struct {
	Time        int64       `json:"time"`
	SelfID      string      `json:"self_id"`
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type,omitempty"` // private, group
	SubType     string      `json:"sub_type,omitempty"`
	MessageID   string      `json:"message_id,omitempty"` // Int in OneBot, but String for us
	UserID      string      `json:"user_id,omitempty"`    // Int in OneBot, String for WeChat
	GroupID     string      `json:"group_id,omitempty"`
	Message     interface{} `json:"message,omitempty"` // String or Segment Array
	RawMessage  string      `json:"raw_message,omitempty"`
	Font        int         `json:"font,omitempty"`
	Sender      *Sender     `json:"sender,omitempty"`

	// Request Event
	RequestType string `json:"request_type,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Flag        string `json:"flag,omitempty"`

	// Notice Event
	NoticeType string `json:"notice_type,omitempty"`

	// Meta Event
	MetaEventType string `json:"meta_event_type,omitempty"`
	Platform      string `json:"platform,omitempty"`
	Status        string `json:"status,omitempty"`
}

type Sender struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex,omitempty"`
	Age      int    `json:"age,omitempty"`
}

// OneBot Action Request
type OneBotAction struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   interface{} `json:"echo"`
}

type ActionParams struct {
	UserID      string `json:"user_id"`
	GroupID     string `json:"group_id"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	// 管理类动作参数
	UserIDs  []string `json:"user_ids"`
	Duration int      `json:"duration"`
	Reason   string   `json:"reason"`
	// 信息查询类参数
	NoCache bool `json:"no_cache"`
}

// Action Response
type OneBotResponse struct {
	Status  string      `json:"status"` // ok, failed
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
	Echo    interface{} `json:"echo,omitempty"`
}
