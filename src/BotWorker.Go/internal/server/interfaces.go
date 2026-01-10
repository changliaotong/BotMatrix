package server

import (
	"botworker/internal/onebot"
)

// ActionCaller defines the interface for sending actions to the OneBot implementation
type ActionCaller interface {
	SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error)
	DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error)
	SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error)
	SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error)
}
