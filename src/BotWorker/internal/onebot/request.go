package onebot

type Request struct {
	Action  string `json:"action"`
	Params  any    `json:"params,omitempty"`
	Echo    any    `json:"echo,omitempty"`
	Request any    `json:"request,omitempty"`
}

type Response struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Echo    any    `json:"echo,omitempty"`
}

type SendMessageParams struct {
	UserID      FlexibleInt64 `json:"user_id,omitempty"`
	GroupID     FlexibleInt64 `json:"group_id,omitempty"`
	Message     any           `json:"message"`
	AutoEscape  bool          `json:"auto_escape,omitempty"`
	MessageType string        `json:"message_type,omitempty"`
	ID          FlexibleInt64 `json:"id,omitempty"`
}

type DeleteMessageParams struct {
	MessageID FlexibleInt64 `json:"message_id"`
}

type SendLikeParams struct {
	UserID FlexibleInt64 `json:"user_id"`
	Times  int           `json:"times"`
}

type SetGroupKickParams struct {
	GroupID   FlexibleInt64 `json:"group_id"`
	UserID    FlexibleInt64 `json:"user_id"`
	RejectAdd bool          `json:"reject_add_request,omitempty"`
}

type SetGroupBanParams struct {
	GroupID  FlexibleInt64 `json:"group_id"`
	UserID   FlexibleInt64 `json:"user_id"`
	Duration int           `json:"duration,omitempty"`
}

// GetGroupMemberListParams 获取群成员列表参数
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#get_group_member_list-%E8%8E%B7%E5%8F%96%E7%BE%A4%E6%88%90%E5%91%98%E5%88%97%E8%A1%A8

type GetGroupMemberListParams struct {
	GroupID FlexibleInt64 `json:"group_id"`
	NoCache bool          `json:"no_cache,omitempty"`
}

// GetGroupMemberInfoParams 获取群成员信息参数
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#get_group_member_info-%E8%8E%B7%E5%8F%96%E7%BE%A4%E6%88%90%E5%91%98%E4%BF%A1%E6%81%AF

type GetGroupMemberInfoParams struct {
	GroupID FlexibleInt64 `json:"group_id"`
	UserID  FlexibleInt64 `json:"user_id"`
	NoCache bool          `json:"no_cache,omitempty"`
}

// SetGroupSpecialTitleParams 设置群成员头衔参数
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#set_group_special_title-%E8%AE%BE%E7%BD%AE%E7%BE%A4%E6%88%90%E5%91%98%E5%A4%B4%E8%A1%94

type SetGroupSpecialTitleParams struct {
	GroupID      FlexibleInt64 `json:"group_id"`
	UserID       FlexibleInt64 `json:"user_id"`
	SpecialTitle string        `json:"special_title"`
	Duration     int           `json:"duration,omitempty"`
}
