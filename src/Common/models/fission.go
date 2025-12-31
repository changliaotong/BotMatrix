package models

import (
	"time"
)

// FissionConfigGORM 裂变配置
type FissionConfigGORM struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Enabled              bool      `gorm:"default:false;column:enabled" json:"enabled"`
	InviteRewardPoints   int       `gorm:"default:10;column:invite_reward_points" json:"invite_reward_points"`
	NewUserRewardPoints  int       `gorm:"default:5;column:new_user_reward_points" json:"new_user_reward_points"`
	InviteRewardDuration int       `gorm:"default:24;column:invite_reward_duration" json:"invite_reward_duration"` // 小时
	MinLevelRequired     int       `gorm:"default:0;column:min_level_required" json:"min_level_required"`
	MaxDailyInvites      int       `gorm:"default:0;column:max_daily_invites" json:"max_daily_invites"`
	AntiFraudEnabled     bool      `gorm:"default:true;column:anti_fraud_enabled" json:"anti_fraud_enabled"`
	WelcomeMessage       string    `gorm:"size:1000;column:welcome_message" json:"welcome_message"`
	InviteCodeTemplate   string    `gorm:"size:255;default:'INV-{RAND}';column:invite_code_template" json:"invite_code_template"`
	Rules                string    `gorm:"type:text;column:rules" json:"rules"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FissionConfigGORM) TableName() string {
	return "fission_configs"
}

// InvitationGORM 邀请记录
type InvitationGORM struct {
	ID          uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	InviterID   int64      `gorm:"index;column:inviter_id" json:"inviter_id"`       // 邀请人 ID
	InviteeID   int64      `gorm:"uniqueIndex;column:invitee_id" json:"invitee_id"` // 被邀请人 ID (唯一，一个用户只能被邀请一次)
	Platform    string     `gorm:"size:50;column:platform" json:"platform"`
	InviteCode  string     `gorm:"index;size:50;column:invite_code" json:"invite_code"`
	Status      string     `gorm:"size:20;default:'pending';column:status" json:"status"` // pending, completed, invalid
	IPAddress   string     `gorm:"column:ip_address;size:100" json:"ip_address"`
	DeviceID    string     `gorm:"size:255;column:device_id" json:"device_id"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (InvitationGORM) TableName() string {
	return "invitations"
}

// FissionTaskGORM 裂变任务
type FissionTaskGORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name           string    `gorm:"size:255;column:name" json:"name"`
	Description    string    `gorm:"size:500;column:description" json:"description"`
	TaskType       string    `gorm:"column:task_type;size:50" json:"task_type"` // register, use_bot, join_group
	TargetCount    int       `gorm:"default:1;column:target_count" json:"target_count"`
	RewardPoints   int       `gorm:"default:0;column:reward_points" json:"reward_points"`
	RewardDuration int       `gorm:"default:0;column:reward_duration" json:"reward_duration"`
	Status         string    `gorm:"size:20;default:'active';column:status" json:"status"` // active, inactive
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FissionTaskGORM) TableName() string {
	return "fission_tasks"
}

// UserFissionRecordGORM 用户裂变数据
type UserFissionRecordGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID       int64     `gorm:"uniqueIndex;column:user_id" json:"user_id"`
	Platform     string    `gorm:"size:50;column:platform" json:"platform"`
	InviteCount  int       `gorm:"default:0;column:invite_count" json:"invite_count"`
	Points       int       `gorm:"default:0;column:points" json:"points"`
	Level        int       `gorm:"default:1;column:level" json:"level"`
	TotalRewards float64   `gorm:"default:0;column:total_rewards" json:"total_rewards"`
	InviteCode   string    `gorm:"uniqueIndex;size:50;column:invite_code" json:"invite_code"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserFissionRecordGORM) TableName() string {
	return "user_fission_records"
}

// FissionRewardLogGORM 奖励记录
type FissionRewardLogGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"index;column:user_id" json:"user_id"`
	Type      string    `gorm:"size:50;column:type" json:"type"` // points, duration, item
	Amount    int       `gorm:"default:0;column:amount" json:"amount"`
	Reason    string    `gorm:"size:255;column:reason" json:"reason"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (FissionRewardLogGORM) TableName() string {
	return "fission_reward_logs"
}
