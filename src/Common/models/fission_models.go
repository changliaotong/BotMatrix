package models

import "time"

// FissionUser represents a user in Fission platform
type FissionUser struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
}
