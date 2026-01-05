package types

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// BotClient represents a connected OneBot client
type BotClient struct {
	Conn          *websocket.Conn `json:"-"`
	SelfID        string          `json:"self_id"`
	Nickname      string          `json:"nickname"`
	GroupCount    int             `json:"group_count"`
	FriendCount   int             `json:"friend_count"`
	Connected     time.Time       `json:"connected"`
	Platform      string          `json:"platform"`
	Protocol      string          `json:"protocol"` // "v11" or "v12"
	Mutex         sync.Mutex      `json:"-"`
	SentCount     int64           `json:"sent_count"`     // Track sent messages per bot session
	RecvCount     int64           `json:"recv_count"`     // Track received messages per bot session
	LastHeartbeat time.Time       `json:"last_heartbeat"` // Track last heartbeat for timeout detection
}

// GroupInfo represents cached group information
type GroupInfo struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	BotID     string    `json:"bot_id"`
	IsCached  bool      `json:"is_cached"`
	LastSeen  time.Time `json:"last_seen"`
	Avatar    string    `json:"avatar"`
}

// FriendInfo represents cached friend information
type FriendInfo struct {
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	BotID    string    `json:"bot_id"`
	IsCached bool      `json:"is_cached"`
	LastSeen time.Time `json:"last_seen"`
	Avatar   string    `json:"avatar"`
}

// MemberInfo represents cached group member information
type MemberInfo struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Card     string    `json:"card"`
	Role     string    `json:"role"`
	BotID    string    `json:"bot_id"`
	Avatar   string    `json:"avatar"`
	IsCached bool      `json:"is_cached"`
	LastSeen time.Time `json:"last_seen"`
}

// GroupListItem represents an item in the list returned by get_group_list
type GroupListItem struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// LoginInfo represents the data returned by get_login_info
type LoginInfo struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}
