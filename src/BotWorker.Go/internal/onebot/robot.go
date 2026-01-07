package onebot

type Robot interface {
	OnMessage(fn func(event *Event) error)
	OnNotice(fn func(event *Event) error)
	OnRequest(fn func(event *Event) error)
	OnEvent(eventName string, fn func(event *Event) error)
	SendMessage(params *SendMessageParams) (*Response, error)
	DeleteMessage(params *DeleteMessageParams) (*Response, error)
	SendLike(params *SendLikeParams) (*Response, error)
	SetGroupKick(params *SetGroupKickParams) (*Response, error)
	SetGroupBan(params *SetGroupBanParams) (*Response, error)
	GetGroupMemberList(params *GetGroupMemberListParams) (*Response, error)
	GetGroupMemberInfo(params *GetGroupMemberInfoParams) (*Response, error)
	SetGroupSpecialTitle(params *SetGroupSpecialTitleParams) (*Response, error)
	Run() error
	Stop() error
}

type EventHandler func(event *Event) error

type RequestHandler func(request *Request) (*Response, error)
