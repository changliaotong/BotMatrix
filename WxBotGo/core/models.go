package core

// OneBot Event Structure
type OneBotEvent struct {
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
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
}

// Action Response
type OneBotResponse struct {
	Status  string      `json:"status"` // ok, failed
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
	Echo    interface{} `json:"echo,omitempty"`
}
