package models

import (
	"time"

	"gorm.io/gorm"
)

// BotEntity represents a bot entity in the system
type BotEntity struct {
	Id        uint           `gorm:"primaryKey;column:Id" json:"id"`
	SelfID    string         `gorm:"uniqueIndex;size:64;column:SelfId" json:"self_id"`
	Nickname  string         `gorm:"size:255;column:Nickname" json:"nickname"`
	Platform  string         `gorm:"size:32;column:Platform" json:"platform"`
	Status    string         `gorm:"size:32;column:Status" json:"status"`
	Connected bool           `gorm:"-" json:"connected"`
	LastSeen  time.Time      `gorm:"-" json:"last_seen"`
	CreatedAt time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (BotEntity) TableName() string {
	return "BotEntity"
}

// BotEntityGORM is an alias or compatible struct for BotEntity
type BotEntityGORM struct {
	Id        uint           `gorm:"primaryKey;column:Id" json:"id"`
	SelfID    string         `gorm:"uniqueIndex;size:64;column:SelfId" json:"self_id"`
	Nickname  string         `gorm:"size:255;column:Nickname" json:"nickname"`
	Platform  string         `gorm:"size:32;column:Platform" json:"platform"`
	Status    string         `gorm:"size:32;column:Status" json:"status"`
	Connected bool           `gorm:"-" json:"connected"`
	LastSeen  time.Time      `gorm:"-" json:"last_seen"`
	CreatedAt time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (BotEntityGORM) TableName() string {
	return "BotEntity"
}
