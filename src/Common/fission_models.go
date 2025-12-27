package common

import (
	"time"
)

// FissionConfigGORM 裂变配置
type FissionConfigGORM struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement"`
	Enabled              bool   `gorm:"default:false"`
	InviteRewardPoints   int    `gorm:"default:10"`
	NewUserRewardPoints  int    `gorm:"default:5"`
	InviteRewardDuration int    `gorm:"default:24"` // 小时
	MinLevelRequired     int    `gorm:"default:0"`
	MaxDailyInvites      int    `gorm:"default:0"`
	AntiFraudEnabled     bool   `gorm:"default:true"`
	WelcomeMessage       string `gorm:"size:1000"`
	InviteCodeTemplate   string `gorm:"size:255;default:'INV-{RAND}'"`
	Rules                string `gorm:"type:text"`
	UpdatedAt            time.Time
}

func (FissionConfigGORM) TableName() string {
	return "fission_configs"
}

// InvitationGORM 邀请记录
type InvitationGORM struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	InviterID   int64  `gorm:"index"`       // 邀请人 ID
	InviteeID   int64  `gorm:"uniqueIndex"` // 被邀请人 ID (唯一，一个用户只能被邀请一次)
	Platform    string `gorm:"size:50"`
	InviteCode  string `gorm:"index;size:50"`
	Status      string `gorm:"size:20;default:'pending'"` // pending, completed, invalid
	IPAddress   string `gorm:"column:ip_address;size:100"`
	DeviceID    string `gorm:"size:255"`
	CreatedAt   time.Time
	CompletedAt *time.Time
	UpdatedAt   time.Time
}

func (InvitationGORM) TableName() string {
	return "invitations"
}

// FissionTaskGORM 裂变任务
type FissionTaskGORM struct {
	ID             uint   `gorm:"primaryKey;autoIncrement"`
	Name           string `gorm:"size:255"`
	Description    string `gorm:"size:500"`
	TaskType       string `gorm:"column:task_type;size:50"` // register, use_bot, join_group
	TargetCount    int    `gorm:"default:1"`
	RewardPoints   int    `gorm:"default:0"`
	RewardDuration int    `gorm:"default:0"`
	Status         string `gorm:"size:20;default:'active'"` // active, inactive
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (FissionTaskGORM) TableName() string {
	return "fission_tasks"
}

// UserFissionRecordGORM 用户裂变数据
type UserFissionRecordGORM struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	UserID       int64   `gorm:"uniqueIndex"`
	Platform     string  `gorm:"size:50"`
	InviteCount  int     `gorm:"default:0"`
	Points       int     `gorm:"default:0"`
	Level        int     `gorm:"default:1"`
	TotalRewards float64 `gorm:"default:0"`
	InviteCode   string  `gorm:"uniqueIndex;size:50"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserFissionRecordGORM) TableName() string {
	return "user_fission_records"
}

// FissionRewardLogGORM 奖励记录
type FissionRewardLogGORM struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    int64  `gorm:"index"`
	Type      string `gorm:"size:50"` // points, duration, item
	Amount    int    `gorm:"default:0"`
	Reason    string `gorm:"size:255"`
	CreatedAt time.Time
}

func (FissionRewardLogGORM) TableName() string {
	return "fission_reward_logs"
}
