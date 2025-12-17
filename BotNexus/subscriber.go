package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Subscriber represents a UI or other consumer
type Subscriber struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	User  *User
}

// User represents a user
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}