package onebot

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type FlexibleInt64 int64

func (f *FlexibleInt64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = 0
		return nil
	}

	// 尝试作为数字解析
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexibleInt64(i)
		return nil
	}

	// 尝试作为带引号的字符串解析
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// 转换失败，按用户要求设为 0 而不返回错误
			*f = 0
			return nil
		}
		*f = FlexibleInt64(val)
		return nil
	}

	// 其他情况也默认设为 0
	*f = 0
	return nil
}

func (f FlexibleInt64) Int64() int64 {
	return int64(f)
}

func (f FlexibleInt64) String() string {
	return fmt.Sprintf("%d", int64(f))
}

func (f FlexibleInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(f))
}

type Event struct {
	Time          int64         `json:"time"`
	SelfID        FlexibleInt64 `json:"self_id"`
	PostType      string        `json:"post_type"`
	MessageType   string        `json:"message_type,omitempty"`
	SubType       string        `json:"sub_type,omitempty"`
	MessageID     FlexibleInt64 `json:"message_id,omitempty"`
	UserID        FlexibleInt64 `json:"user_id,omitempty"`
	GroupID       FlexibleInt64 `json:"group_id,omitempty"`
	Dice          int           `json:"dice,omitempty"`
	Anonymous     interface{}   `json:"anonymous,omitempty"`
	Message       interface{}   `json:"message,omitempty"`
	RawMessage    string        `json:"raw_message,omitempty"`
	Font          int           `json:"font,omitempty"`
	Sender        Sender        `json:"sender,omitempty"`
	NoticeType    string        `json:"notice_type,omitempty"`
	OperatorID    FlexibleInt64 `json:"operator_id,omitempty"`
	File          File          `json:"file,omitempty"`
	RequestType   string        `json:"request_type,omitempty"`
	Flag          string        `json:"flag,omitempty"`
	Comment       string        `json:"comment,omitempty"`
	Approved      bool          `json:"approved,omitempty"`
	EventName     string        `json:"event_name,omitempty"`
	Platform      string        `json:"platform,omitempty"`
	TargetUserID  string        `json:"target_user_id,omitempty"`
	TargetGroupID string        `json:"target_group_id,omitempty"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		UserID  interface{} `json:"user_id"`
		GroupID interface{} `json:"group_id"`
		SelfID  interface{} `json:"self_id"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 SelfID 和 TargetSelfID
	if aux.SelfID != nil {
		switch v := aux.SelfID.(type) {
		case string:
			e.TargetUserID = v // 注意：这里我们将 SelfID 也作为用户处理，共用 TargetUserID
			if val, err := strconv.ParseInt(v, 10, 64); err == nil {
				e.SelfID = FlexibleInt64(val)
			} else {
				e.SelfID = 0
			}
		case float64:
			e.SelfID = FlexibleInt64(int64(v))
			// 如果是数字，我们暂时不设置 TargetUserID，除非明确是 QQGuild 平台
		}
	}

	// 处理 UserID 和 TargetUserID
	if aux.UserID != nil {
		switch v := aux.UserID.(type) {
		case string:
			e.TargetUserID = v
			if val, err := strconv.ParseInt(v, 10, 64); err == nil {
				e.UserID = FlexibleInt64(val)
			} else {
				e.UserID = 0
			}
		case float64:
			e.UserID = FlexibleInt64(int64(v))
			e.TargetUserID = fmt.Sprintf("%.0f", v)
		}
	}

	// 处理 GroupID 和 TargetGroupID
	if aux.GroupID != nil {
		switch v := aux.GroupID.(type) {
		case string:
			e.TargetGroupID = v
			if val, err := strconv.ParseInt(v, 10, 64); err == nil {
				e.GroupID = FlexibleInt64(val)
			} else {
				e.GroupID = 0
			}
		case float64:
			e.GroupID = FlexibleInt64(int64(v))
			e.TargetGroupID = fmt.Sprintf("%.0f", v)
		}
	}

	_ = aux
	return nil
}

// EnsureIDs 为 QQGuild 平台自动生成 ID
func (e *Event) EnsureIDs(
	getUID func(string) (int64, error),
	getGID func(string) (int64, error),
	getMaxUID func() (int64, error),
	getMaxGID func() (int64, error),
	saveUser func(int64, int64, string, string, string) error,
	saveGroup func(int64, int64, string, string) error,
) {
	if e.Platform != "qqguild" {
		return
	}

	// 处理 SelfID (机器人自己)
	if e.SelfID == 0 && e.TargetUserID != "" {
		if uid, err := getUID(e.TargetUserID); err == nil && uid != 0 {
			e.SelfID = FlexibleInt64(uid)
		} else {
			// 生成新的 SelfID
			if id, err := getMaxUID(); err == nil {
				e.SelfID = FlexibleInt64(id)
				_ = saveUser(id, 0, e.TargetUserID, "Robot", "")
			}
		}
	}

	// 处理 UserID
	if e.UserID == 0 && e.TargetUserID != "" {
		if uid, err := getUID(e.TargetUserID); err == nil && uid != 0 {
			e.UserID = FlexibleInt64(uid)
		} else {
			// 生成新的 UserID
			if id, err := getMaxUID(); err == nil {
				e.UserID = FlexibleInt64(id)
				_ = saveUser(id, 0, e.TargetUserID, e.Sender.Nickname, "")
			}
		}
		if e.Sender.UserID == 0 {
			e.Sender.UserID = e.UserID
		}
	}

	// 处理 GroupID
	if e.GroupID == 0 && e.TargetGroupID != "" {
		if gid, err := getGID(e.TargetGroupID); err == nil && gid != 0 {
			e.GroupID = FlexibleInt64(gid)
		} else {
			// 生成新的 GroupID
			if id, err := getMaxGID(); err == nil {
				e.GroupID = FlexibleInt64(id)
				_ = saveGroup(id, 0, e.TargetGroupID, "Group_"+e.GroupID.String())
			}
		}
	}
}

type Sender struct {
	UserID   FlexibleInt64 `json:"user_id"`
	Nickname string        `json:"nickname"`
	Sex      string        `json:"sex"`
	Age      int           `json:"age"`
	Card     string        `json:"card,omitempty"`
	Area     string        `json:"area,omitempty"`
	Level    string        `json:"level,omitempty"`
	Role     string        `json:"role,omitempty"`
	Title    string        `json:"title,omitempty"`
}

type File struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Size  int64         `json:"size"`
	BusID FlexibleInt64 `json:"busid"`
}

type MessageSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
