package types

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// User represents a user with password hash
type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"` // 密码哈希，不序列化到JSON
	IsAdmin        bool      `json:"is_admin"`
	IsSuperPoints  bool      `json:"is_super_points"`
	QQ             string    `json:"qq"`
	SessionVersion int       `json:"session_version"` // 用于同步登录状态和强制登出
	Active         bool      `json:"active"`          // 用户状态：启用/禁用
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserClaims for JWT
type UserClaims struct {
	UserID         int64  `json:"user_id"`
	Username       string `json:"username"`
	IsAdmin        bool   `json:"is_admin"`
	SessionVersion int    `json:"session_version"`
	jwt.RegisteredClaims
}

// ContextKey is a custom type for context keys
type ContextKey string

// UserClaimsKey is the key for user claims in context
const UserClaimsKey ContextKey = "user_claims"

// Subscriber represents a UI or other consumer
type Subscriber struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	User  *User
}
