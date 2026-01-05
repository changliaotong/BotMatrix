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
	UserID      any    `json:"user_id,omitempty"`
	GroupID     any    `json:"group_id,omitempty"`
	Message     any    `json:"message"`
	AutoEscape  bool   `json:"auto_escape,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	ID          any    `json:"id,omitempty"`
	Platform    string `json:"platform,omitempty"`
	SelfID      string `json:"self_id,omitempty"`
}

type DeleteMessageParams struct {
	MessageID FlexibleInt64 `json:"message_id"`
	Platform  string        `json:"platform,omitempty"`
	SelfID    string        `json:"self_id,omitempty"`
}

type SendLikeParams struct {
	UserID   FlexibleInt64 `json:"user_id"`
	Times    int           `json:"times"`
	Platform string        `json:"platform,omitempty"`
	SelfID   string        `json:"self_id,omitempty"`
}

type SetGroupKickParams struct {
	GroupID   FlexibleInt64 `json:"group_id"`
	UserID    FlexibleInt64 `json:"user_id"`
	RejectAdd bool          `json:"reject_add_request,omitempty"`
	Platform  string        `json:"platform,omitempty"`
	SelfID    string        `json:"self_id,omitempty"`
}

type SetGroupBanParams struct {
	GroupID  FlexibleInt64 `json:"group_id"`
	UserID   FlexibleInt64 `json:"user_id"`
	Duration int           `json:"duration,omitempty"`
	Platform string        `json:"platform,omitempty"`
	SelfID   string        `json:"self_id,omitempty"`
}

type GetGroupMemberListParams struct {
	GroupID  FlexibleInt64 `json:"group_id"`
	NoCache  bool          `json:"no_cache,omitempty"`
	Platform string        `json:"platform,omitempty"`
	SelfID   string        `json:"self_id,omitempty"`
}

type GetGroupMemberInfoParams struct {
	GroupID  FlexibleInt64 `json:"group_id"`
	UserID   FlexibleInt64 `json:"user_id"`
	NoCache  bool          `json:"no_cache,omitempty"`
	Platform string        `json:"platform,omitempty"`
	SelfID   string        `json:"self_id,omitempty"`
}

type SetGroupSpecialTitleParams struct {
	GroupID      FlexibleInt64 `json:"group_id"`
	UserID       FlexibleInt64 `json:"user_id"`
	SpecialTitle string        `json:"special_title"`
	Duration     int           `json:"duration,omitempty"`
	Platform     string        `json:"platform,omitempty"`
	SelfID       string        `json:"self_id,omitempty"`
}
