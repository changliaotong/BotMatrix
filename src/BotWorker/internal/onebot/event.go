package onebot

type Event struct {
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type,omitempty"`
	SubType     string      `json:"sub_type,omitempty"`
	MessageID   int64       `json:"message_id,omitempty"`
	UserID      int64       `json:"user_id,omitempty"`
	GroupID     int64       `json:"group_id,omitempty"`
	Dice        int         `json:"dice,omitempty"`
	Anonymous   interface{} `json:"anonymous,omitempty"`
	Message     interface{} `json:"message,omitempty"`
	RawMessage  string      `json:"raw_message,omitempty"`
	Font        int         `json:"font,omitempty"`
	Sender      Sender      `json:"sender,omitempty"`
	NoticeType  string      `json:"notice_type,omitempty"`
	OperatorID  int64       `json:"operator_id,omitempty"`
	File        File        `json:"file,omitempty"`
	RequestType string      `json:"request_type,omitempty"`
	Flag        string      `json:"flag,omitempty"`
	Comment     string      `json:"comment,omitempty"`
	Approved    bool        `json:"approved,omitempty"`
	EventName   string      `json:"event_name,omitempty"`
}

type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
	Card     string `json:"card,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Role     string `json:"role,omitempty"`
	Title    string `json:"title,omitempty"`
}

type File struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	BusID int64  `json:"busid"`
}

type MessageSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
