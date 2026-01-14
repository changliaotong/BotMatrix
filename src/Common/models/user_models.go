package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user for the management portal
type User struct {
	ID             uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username       string         `gorm:"uniqueIndex;not null;size:255;column:username" json:"username"`
	PasswordHash   string         `gorm:"not null;size:255;column:password_hash" json:"-"`
	Platform       string         `gorm:"size:32;column:platform" json:"platform"`
	PlatformID     string         `gorm:"size:64;column:platform_id" json:"platform_id"`
	IsAdmin        bool           `gorm:"default:false;column:is_admin" json:"is_admin"`
	QQ             string         `gorm:"size:20;column:qq" json:"qq"`
	Active         bool           `gorm:"default:true;column:active" json:"active"`
	SessionVersion int            `gorm:"default:1;column:session_version" json:"session_version"`
	CreatedAt      time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// UserGORM is an alias for User for compatibility
type UserGORM = User

// UserLoginTokenGORM is an alias for UserLoginToken for compatibility
type UserLoginTokenGORM struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Platform   string    `gorm:"size:32;not null;column:Platform" json:"platform"`
	PlatformID string    `gorm:"index;size:64;not null;column:PlatformId" json:"platform_id"`
	Token      string    `gorm:"size:16;not null;column:Token" json:"token"`
	ExpiresAt  time.Time `gorm:"column:ExpiresAt" json:"expires_at"`
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

func (UserLoginTokenGORM) TableName() string {
	return "UserLoginToken"
}
