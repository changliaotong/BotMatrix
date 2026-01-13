package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ISz84Store defines the interface for sz84 data operations
type ISz84Store interface {
	GetLimiterLogs(userID int64, actionKey string) ([]Sz84LimiterLog, error)
	AddLimiterLog(log *Sz84LimiterLog) error
	// Add other methods as needed
}

// Sz84Store implements ISz84Store with dialect-aware logic and Redis caching
type Sz84Store struct {
	db      *gorm.DB
	rdb     *redis.Client
	dialect string
}

func NewSz84Store(db *gorm.DB, rdb *redis.Client) *Sz84Store {
	dialect := db.Dialector.Name()
	return &Sz84Store{
		db:      db,
		rdb:     rdb,
		dialect: dialect,
	}
}

// SigninService handles signin related operations
type SigninService struct {
	store *Sz84Store
}

func NewSigninService(store *Sz84Store) *SigninService {
	return &SigninService{store: store}
}

// GetDateCondition returns a dialect-specific date comparison string
// daysAgo: 0 for today, 1 for yesterday, etc.
func (s *Sz84Store) GetDateCondition(columnName string, daysAgo int) string {
	if s.dialect == "sqlserver" {
		if daysAgo == 0 {
			return fmt.Sprintf("CAST(%s AS DATE) = CAST(GETDATE() AS DATE)", columnName)
		}
		return fmt.Sprintf("CAST(%s AS DATE) = CAST(DATEADD(day, -%d, GETDATE()) AS DATE)", columnName, daysAgo)
	}
	// Default to Postgres
	if daysAgo == 0 {
		return fmt.Sprintf("%s::date = CURRENT_DATE", columnName)
	}
	return fmt.Sprintf("%s::date = CURRENT_DATE - INTERVAL '%d day'", columnName, daysAgo)
}

func (s *Sz84Store) NewSigninService() *SigninService {
	return NewSigninService(s)
}

func (s *Sz84Store) GetLimiterLogs(userID int64, actionKey string) ([]Sz84LimiterLog, error) {
	var logs []Sz84LimiterLog
	err := s.db.Where("UserId = ? AND ActionKey = ?", userID, actionKey).Find(&logs).Error
	return logs, err
}

func (s *Sz84Store) AddLimiterLog(log *Sz84LimiterLog) error {
	return s.db.Create(log).Error
}

func (s *Sz84Store) GetGroup(groupID int64) (*Sz84Group, error) {
	key := fmt.Sprintf("sz84:group:%d", groupID)
	if s.rdb != nil {
		if val, err := s.rdb.Get(context.Background(), key).Result(); err == nil {
			var group Sz84Group
			if json.Unmarshal([]byte(val), &group) == nil {
				return &group, nil
			}
		}
	}

	var group Sz84Group
	err := s.db.Where("Id = ?", groupID).First(&group).Error
	if err == nil && s.rdb != nil {
		data, _ := json.Marshal(group)
		s.rdb.Set(context.Background(), key, data, time.Hour*24)
	}
	return &group, err
}

func (s *Sz84Store) GetMember(groupID, userID int64) (*Sz84GroupMember, error) {
	key := fmt.Sprintf("sz84:member:%d:%d", groupID, userID)
	if s.rdb != nil {
		if val, err := s.rdb.Get(context.Background(), key).Result(); err == nil {
			var member Sz84GroupMember
			if json.Unmarshal([]byte(val), &member) == nil {
				return &member, nil
			}
		}
	}

	var member Sz84GroupMember
	err := s.db.Where("GroupId = ? AND UserId = ?", groupID, userID).First(&member).Error
	if err == nil && s.rdb != nil {
		data, _ := json.Marshal(member)
		s.rdb.Set(context.Background(), key, data, time.Hour*2)
	}
	return &member, err
}

func (s *Sz84Store) GetUser(userID int64) (*Sz84User, error) {
	key := fmt.Sprintf("sz84:user:%d", userID)
	if s.rdb != nil {
		if val, err := s.rdb.Get(context.Background(), key).Result(); err == nil {
			var user Sz84User
			if json.Unmarshal([]byte(val), &user) == nil {
				return &user, nil
			}
		}
	}

	var user Sz84User
	err := s.db.Where("Id = ?", userID).First(&user).Error
	if err == nil && s.rdb != nil {
		data, _ := json.Marshal(user)
		s.rdb.Set(context.Background(), key, data, time.Hour*2)
	}
	return &user, err
}

// InvalidateMemberCache 清除特定成员的缓存，通常在数据更新后调用
func (s *Sz84Store) InvalidateMemberCache(groupID, userID int64) {
	if s.rdb != nil {
		s.rdb.Del(context.Background(), fmt.Sprintf("sz84:member:%d:%d", groupID, userID))
		s.rdb.Del(context.Background(), fmt.Sprintf("sz84:user:%d", userID))
	}
}

// Sz84LimiterLog represents the LimiterLog table migrated from sz84
type Sz84LimiterLog struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	GroupID   *int64    `gorm:"column:GroupId" json:"group_id"` // NULL for private chat
	UserID    int64     `gorm:"not null;column:UserId" json:"user_id"`
	ActionKey string    `gorm:"not null;column:ActionKey" json:"action_key"`
	UsedAt    time.Time `gorm:"not null;column:UsedAt" json:"used_at"`
}

func (Sz84LimiterLog) TableName() string {
	return "LimiterLog"
}

// Achievement represents the Achievements table migrated from sz84
type Achievement struct {
	ID            string `gorm:"primaryKey;column:Id" json:"id"`
	Title         string `gorm:"column:Title" json:"title"`
	Description   string `gorm:"column:Description" json:"description"`
	MaxLevel      int    `gorm:"column:MaxLevel" json:"max_level"`
	Category      string `gorm:"column:Category" json:"category"`
	IconUrl       string `gorm:"column:IconUrl" json:"icon_url"`
	Reward        string `gorm:"column:Reward" json:"reward"`
	RequiredCount int    `gorm:"column:RequiredCount" json:"required_count"`
	RewardCredit  int64  `gorm:"column:RewardCredit" json:"reward_credit"`
	CounterKey    string `gorm:"column:CounterKey" json:"counter_key"`
	// Rules is likely a JSON string or handled elsewhere in C#
	RulesRaw string `gorm:"column:Rules" json:"rules_raw"`
}

func (Achievement) TableName() string {
	return "Achievement"
}

// UserAchievement represents the UserAchievement table migrated from sz84
type UserAchievement struct {
	UserID             int64      `gorm:"primaryKey;column:UserId" json:"user_id"`
	AchievementId      string     `gorm:"primaryKey;column:AchievementId" json:"achievement_id"`
	CurrentLevel       int        `gorm:"column:CurrentLevel" json:"current_level"`
	CurrentValue       int        `gorm:"column:CurrentValue" json:"current_value"`
	LastActionDate     time.Time  `gorm:"column:LastActionDate" json:"last_action_date"`
	CurrentStreakDays  int        `gorm:"column:CurrentStreakDays" json:"current_streak_days"`
	LastActionDatePrev *time.Time `gorm:"column:LastActionDatePrev" json:"last_action_date_prev"`
	LastUpdated        time.Time  `gorm:"column:LastUpdated" json:"last_updated"`
}

func (UserAchievement) TableName() string {
	return "UserAchievement"
}

// UserTitle represents the Titles table migrated from sz84
type UserTitle struct {
	UserID     int64     `gorm:"primaryKey;column:UserId" json:"user_id"`
	TitleId    string    `gorm:"primaryKey;column:TitleId" json:"title_id"`
	UnlockTime time.Time `gorm:"column:UnlockTime" json:"unlock_time"`
	IsEquipped bool      `gorm:"column:IsEquipped" json:"is_equipped"`
	Title      string    `gorm:"column:Title" json:"title"`
}

func (UserTitle) TableName() string {
	return "UserTitle"
}

// Sz84GroupMember represents the GroupMember table with all fields from legacy C#
type Sz84GroupMember struct {
	GroupID       int64      `gorm:"primaryKey;column:GroupId" json:"group_id"`
	UserID        int64      `gorm:"primaryKey;column:UserId" json:"user_id"`
	UserName      string     `gorm:"column:UserName" json:"user_name"`
	DisplayName   string     `gorm:"column:DisplayName" json:"display_name"`
	GroupCredit   int64      `gorm:"column:GroupCredit" json:"group_credit"`
	GoldCoins     int64      `gorm:"column:GoldCoins" json:"gold_coins"`
	PurpleCoins   int64      `gorm:"column:PurpleCoins" json:"purple_coins"`
	BlackCoins    int64      `gorm:"column:BlackCoins" json:"black_coins"`
	GameCoins     int64      `gorm:"column:GameCoins" json:"game_coins"`
	SaveCredit    int64      `gorm:"column:SaveCredit" json:"save_credit"`
	Status        int        `gorm:"column:Status" json:"status"`
	ConfirmCode   string     `gorm:"column:ConfirmCode" json:"confirm_code"`
	SignDate      *time.Time `gorm:"column:SignDate" json:"sign_date"`
	SignTimes     int        `gorm:"column:SignTimes" json:"sign_times"`
	SignLevel     int        `gorm:"column:SignLevel" json:"sign_level"`
	SignTimesAll  int        `gorm:"column:SignTimesAll" json:"sign_times_all"`
	Title         string     `gorm:"column:Title" json:"title"`
	IsAdmin       bool       `gorm:"column:IsAdmin" json:"is_admin"`
	LastMsgDate   *time.Time `gorm:"column:LastMsgDate" json:"last_msg_date"`
	MsgCount      int        `gorm:"column:MsgCount" json:"msg_count"`
	JoinDate      *time.Time `gorm:"column:JoinDate" json:"join_date"`
	IsFans        bool       `gorm:"column:IsFans" json:"is_fans"`
	FansDate      *time.Time `gorm:"column:FansDate" json:"fans_date"`
	FansLevel     int        `gorm:"column:FansLevel" json:"fans_level"`
	FansValue     int64      `gorm:"column:FansValue" json:"fans_value"`
	LampDate      *time.Time `gorm:"column:LampDate" json:"lamp_date"`
	InvitorUserId int64      `gorm:"column:InvitorUserId" json:"invitor_user_id"`
	InviteCount   int        `gorm:"column:InviteCount" json:"invite_count"`
	InsertDate    *time.Time `gorm:"column:InsertDate" json:"insert_date"`
}

func (Sz84GroupMember) TableName() string {
	return "GroupMember"
}

// Sz84User represents the User table (UserInfo in C#) with all fields
type Sz84User struct {
	ID             int64     `gorm:"primaryKey;column:Id" json:"id"`
	Name           string    `gorm:"column:Name" json:"name"`
	UserOpenId     string    `gorm:"column:UserOpenId" json:"user_openid"`
	InsertDate     time.Time `gorm:"column:InsertDate" json:"insert_date"`
	Credit         int64     `gorm:"column:Credit" json:"credit"`
	CreditFreeze   int64     `gorm:"column:CreditFreeze" json:"credit_freeze"`
	Coins          int64     `gorm:"column:Coins" json:"coins"`
	CoinsFreeze    int64     `gorm:"column:CoinsFreeze" json:"coins_freeze"`
	Sz84Uid        int       `gorm:"column:Sz84Uid" json:"sz84_uid"`
	Sz84UserName   string    `gorm:"column:Sz84UserName" json:"sz84_user_name"`
	HomeUid        int       `gorm:"column:HomeUid" json:"home_uid"`
	HomeUserName   string    `gorm:"column:HomeUserName" json:"home_user_name"`
	HomeRealName   string    `gorm:"column:HomeRealName" json:"home_real_name"`
	BotUin         int64     `gorm:"column:BotUin" json:"bot_uin"`
	State          int       `gorm:"column:State" json:"state"`
	IsOpen         int       `gorm:"column:IsOpen" json:"is_open"`
	SzTong         string    `gorm:"column:SzTong" json:"sz_tong"`
	CityName       string    `gorm:"column:CityName" json:"city_name"`
	DefaultGroup   int64     `gorm:"column:DefaultGroup" json:"default_group"`
	IsDefaultHint  bool      `gorm:"column:IsDefaultHint" json:"is_default_hint"`
	BindDate       time.Time `gorm:"column:BindDate" json:"bind_date"`
	BindDateHome   time.Time `gorm:"column:BindDateHome" json:"bind_date_home"`
	IsBlack        bool      `gorm:"column:IsBlack" json:"is_black"`
	TeachLevel     int       `gorm:"column:TeachLevel" json:"teach_level"`
	IsCoins        bool      `gorm:"column:IsCoins" json:"is_coins"`
	XCredit        bool      `gorm:"column:XCredit" json:"x_credit"`
	RCq            int       `gorm:"column:RCq" json:"r_cq"`
	RefUserId      int64     `gorm:"column:RefUserId" json:"ref_user_id"`
	UserGuid       string    `gorm:"column:UserGuid" json:"user_guid"`
	IsBlock        bool      `gorm:"column:IsBlock" json:"is_block"`
	SaveCredit     int64     `gorm:"column:SaveCredit" json:"save_credit"`
	UpgradeDate    time.Time `gorm:"column:UpgradeDate" json:"upgrade_date"`
	PartnerUserId  int64     `gorm:"column:PartnerUserId" json:"partner_user_id"`
	FreezeCredit   int64     `gorm:"column:FreezeCredit" json:"freeze_credit"`
	VipStart       time.Time `gorm:"column:VipStart" json:"vip_start"`
	VipEnd         time.Time `gorm:"column:VipEnd" json:"vip_end"`
	LastChengyu    string    `gorm:"column:LastChengyu" json:"last_chengyu"`
	Balance        float64   `gorm:"column:Balance" json:"balance"`
	BalanceFreeze  float64   `gorm:"column:BalanceFreeze" json:"balance_freeze"`
	CreditGiving   int64     `gorm:"column:CreditGiving" json:"credit_giving"`
	LvValue        int       `gorm:"column:LvValue" json:"lv_value"`
	SumIncome      float64   `gorm:"column:SumIncome" json:"sum_income"`
	SuperDate      time.Time `gorm:"column:SuperDate" json:"super_date"`
	IsSuper        bool      `gorm:"column:IsSuper" json:"is_super"`
	IsFreeze       bool      `gorm:"column:IsFreeze" json:"is_freeze"`
	GroupId        int64     `gorm:"column:GroupId" json:"group_id"`
	IsSz84         bool      `gorm:"column:IsSz84" json:"is_sz84"`
	Sz84Date       time.Time `gorm:"column:Sz84Date" json:"sz84_date"`
	AnswerId       int64     `gorm:"column:AnswerId" json:"answer_id"`
	AnswerDate     time.Time `gorm:"column:AnswerDate" json:"answer_date"`
	IsTeach        bool      `gorm:"column:IsTeach" json:"is_teach"`
	IsShutup       bool      `gorm:"column:IsShutup" json:"is_shutup"`
	SystemPrompt   string    `gorm:"column:SystemPrompt" json:"system_prompt"`
	Tokens         int64     `gorm:"column:Tokens" json:"tokens"`
	IsAgent        bool      `gorm:"column:IsAgent" json:"is_agent"`
	IsAI           bool      `gorm:"column:IsAI" json:"is_ai"`
	Xxian          bool      `gorm:"column:Xxian" json:"xxian"`
	AgentId        int64     `gorm:"column:AgentId" json:"agent_id"`
	IsSendHelpInfo bool      `gorm:"column:IsSendHelpInfo" json:"is_send_help_info"`
	IsLog          bool      `gorm:"column:IsLog" json:"is_log"`
	IsMusicLogo    bool      `gorm:"column:IsMusicLogo" json:"is_music_logo"`
	CszRes         int       `gorm:"column:CszRes" json:"csz_res"`
	CszCredit      int64     `gorm:"column:CszCredit" json:"csz_credit"`
	CszTimes       int       `gorm:"column:CszTimes" json:"csz_times"`
	GroupOpenid    string    `gorm:"column:GroupOpenid" json:"group_openid"`
}

func (Sz84User) TableName() string {
	return "User"
}

// RobotWeibo represents the robot_weibo table used for sign-in logs
type RobotWeibo struct {
	WeiboID    int64     `gorm:"primaryKey;autoIncrement;column:Id" json:"weibo_id"`
	RobotQQ    int64     `gorm:"column:RobotQQ" json:"robot_qq"`
	WeiboQQ    int64     `gorm:"column:WeiboQQ" json:"weibo_qq"`
	WeiboInfo  string    `gorm:"column:WeiboInfo" json:"weibo_info"`
	WeiboType  int       `gorm:"column:WeiboType" json:"weibo_type"` // 1 for Sign-in
	GroupID    int64     `gorm:"column:GroupId" json:"group_id"`
	InsertDate time.Time `gorm:"column:InsertDate;default:CURRENT_TIMESTAMP" json:"insert_date"`
}

func (RobotWeibo) TableName() string {
	return "RobotWeibo"
}

// Sz84Group represents the GroupInfo table
type Sz84Group struct {
	Id                    int64     `gorm:"primaryKey;column:Id" json:"id"`
	GroupName             string    `gorm:"column:GroupName" json:"group_name"`
	GroupMemo             string    `gorm:"column:GroupMemo" json:"group_memo"`
	GroupOwnerName        string    `gorm:"column:GroupOwnerName" json:"group_owner_name"`
	GroupOwnerNickname    string    `gorm:"column:GroupOwnerNickname" json:"group_owner_nickname"`
	GroupType             int       `gorm:"column:GroupType" json:"group_type"`
	RobotOwnerName        string    `gorm:"column:RobotOwnerName" json:"robot_owner_name"`
	WelcomeMessage        string    `gorm:"column:WelcomeMessage" json:"welcome_message"`
	GroupState            int       `gorm:"column:GroupState" json:"group_state"`
	BotName               string    `gorm:"column:BotName" json:"bot_name"`
	LastDate              time.Time `gorm:"column:LastDate" json:"last_date"`
	IsOpen                bool      `gorm:"column:IsOpen" json:"is_open"`
	UseRight              int       `gorm:"column:UseRight" json:"use_right"`
	TeachRight            int       `gorm:"column:TeachRight" json:"teach_right"`
	AdminRight            int       `gorm:"column:AdminRight" json:"admin_right"`
	IsCloseManager        bool      `gorm:"column:IsCloseManager" json:"is_close_manager"`
	IsAcceptNewMember     int       `gorm:"column:IsAcceptNewMember" json:"is_accept_new_member"`
	RegexRequestJoin      string    `gorm:"column:RegexRequestJoin" json:"regex_request_join"`
	RejectMessage         string    `gorm:"column:RejectMessage" json:"reject_message"`
	IsWelcomeHint         bool      `gorm:"column:IsWelcomeHint" json:"is_welcome_hint"`
	IsExitHint            bool      `gorm:"column:IsExitHint" json:"is_exit_hint"`
	IsKickHint            bool      `gorm:"column:IsKickHint" json:"is_kick_hint"`
	IsChangeHint          bool      `gorm:"column:IsChangeHint" json:"is_change_hint"`
	IsRightHint           bool      `gorm:"column:IsRightHint" json:"is_right_hint"`
	IsCloudBlack          bool      `gorm:"column:IsCloudBlack" json:"is_cloud_black"`
	IsCloudAnswer         int       `gorm:"column:IsCloudAnswer" json:"is_cloud_answer"`
	IsRequirePrefix       bool      `gorm:"column:IsRequirePrefix" json:"is_require_prefix"`
	IsSz84                bool      `gorm:"column:IsSz84" json:"is_sz84"`
	IsWarn                bool      `gorm:"column:IsWarn" json:"is_warn"`
	IsBlackExit           bool      `gorm:"column:IsBlackExit" json:"is_black_exit"`
	IsBlackKick           bool      `gorm:"column:IsBlackKick" json:"is_black_kick"`
	IsBlackShare          bool      `gorm:"column:IsBlackShare" json:"is_black_share"`
	IsChangeEnter         bool      `gorm:"column:IsChangeEnter" json:"is_change_enter"`
	IsMuteEnter           bool      `gorm:"column:IsMuteEnter" json:"is_mute_enter"`
	IsChangeMessage       bool      `gorm:"column:IsChangeMessage" json:"is_change_message"`
	RecallKeyword         string    `gorm:"column:RecallKeyword" json:"recall_keyword"`
	WarnKeyword           string    `gorm:"column:WarnKeyword" json:"warn_keyword"`
	MuteKeyword           string    `gorm:"column:MuteKeyword" json:"mute_keyword"`
	KickKeyword           string    `gorm:"column:KickKeyword" json:"kick_keyword"`
	BlackKeyword          string    `gorm:"column:BlackKeyword" json:"black_keyword"`
	MuteEnterCount        int       `gorm:"column:MuteEnterCount" json:"mute_enter_count"`
	MuteKeywordCount      int       `gorm:"column:MuteKeywordCount" json:"mute_keyword_count"`
	KickCount             int       `gorm:"column:KickCount" json:"kick_count"`
	BlackCount            int       `gorm:"column:BlackCount" json:"black_count"`
	ParentGroup           int64     `gorm:"column:ParentGroup" json:"parent_group"`
	CardNamePrefixBoy     string    `gorm:"column:CardNamePrefixBoy" json:"card_name_prefix_boy"`
	CardNamePrefixGirl    string    `gorm:"column:CardNamePrefixGirl" json:"card_name_prefix_girl"`
	CardNamePrefixManager string    `gorm:"column:CardNamePrefixManager" json:"card_name_prefix_manager"`
	BlockMin              int       `gorm:"column:BlockMin" json:"block_min"`
	IsBlock               bool      `gorm:"column:IsBlock" json:"is_block"`
	IsWhite               bool      `gorm:"column:IsWhite" json:"is_white"`
	CityName              string    `gorm:"column:CityName" json:"city_name"`
	IsMuteRefresh         bool      `gorm:"column:IsMuteRefresh" json:"is_mute_refresh"`
	MuteRefreshCount      int       `gorm:"column:MuteRefreshCount" json:"mute_refresh_count"`
	IsProp                bool      `gorm:"column:IsProp" json:"is_prop"`
	IsPet                 bool      `gorm:"column:IsPet" json:"is_pet"`
	IsBlackRefresh        bool      `gorm:"column:IsBlackRefresh" json:"is_black_refresh"`
	FansName              string    `gorm:"column:FansName" json:"fans_name"`
	IsConfirmNew          bool      `gorm:"column:IsConfirmNew" json:"is_confirm_new"`
	WhiteKeyword          string    `gorm:"column:WhiteKeyword" json:"white_keyword"`
	CreditKeyword         string    `gorm:"column:CreditKeyword" json:"credit_keyword"`
	IsCredit              bool      `gorm:"column:IsCredit" json:"is_credit"`
	IsPowerOn             bool      `gorm:"column:IsPowerOn" json:"is_power_on"`
	IsHintClose           bool      `gorm:"column:IsHintClose" json:"is_hint_close"`
	RecallTime            int       `gorm:"column:RecallTime" json:"recall_time"`
	IsInvite              bool      `gorm:"column:IsInvite" json:"is_invite"`
	InviteCredit          int       `gorm:"column:InviteCredit" json:"invite_credit"`
	IsReplyImage          bool      `gorm:"column:IsReplyImage" json:"is_reply_image"`
	IsReplyRecall         bool      `gorm:"column:IsReplyRecall" json:"is_reply_recall"`
	IsVoiceReply          bool      `gorm:"column:IsVoiceReply" json:"is_voice_reply"`
	VoiceId               string    `gorm:"column:VoiceId" json:"voice_id"`
	IsAI                  bool      `gorm:"column:IsAI" json:"is_ai"`
	SystemPrompt          string    `gorm:"column:SystemPrompt" json:"system_prompt"`
	IsOwnerPay            bool      `gorm:"column:IsOwnerPay" json:"is_owner_pay"`
	ContextCount          int       `gorm:"column:ContextCount" json:"context_count"`
	IsMultAI              bool      `gorm:"column:IsMultAI" json:"is_mult_ai"`
	IsAutoSignin          bool      `gorm:"column:IsAutoSignin" json:"is_auto_signin"`
	IsUseKnowledgebase    bool      `gorm:"column:IsUseKnowledgebase" json:"is_use_knowledgebase"`
	IsSendHelpInfo        bool      `gorm:"column:IsSendHelpInfo" json:"is_send_help_info"`
	IsRecall              bool      `gorm:"column:IsRecall" json:"is_recall"`
	IsCreditSystem        bool      `gorm:"column:IsCreditSystem" json:"is_credit_system"`
	GroupOpenid           string    `gorm:"column:GroupOpenid" json:"group_openid"`
	GroupOwner            int64     `gorm:"column:GroupOwner" json:"group_owner"`
	RobotOwner            int64     `gorm:"column:RobotOwner" json:"robot_owner"`
	BotUin                int64     `gorm:"column:BotUin" json:"bot_uin"`
	IsValid               bool      `gorm:"column:IsValid" json:"is_valid"`
	IsProxy               bool      `gorm:"column:IsProxy" json:"is_proxy"`
	QuietTime             time.Time `gorm:"column:QuietTime" json:"quiet_time"`
	IsInGame              int       `gorm:"column:IsInGame" json:"is_in_game"`
	CloseRegex            string    `gorm:"column:CloseRegex" json:"close_regex"`
	IsSaveRecord          bool      `gorm:"column:IsSaveRecord" json:"is_save_record"`
	IsPause               bool      `gorm:"column:IsPause" json:"is_pause"`
	LastAnswer            string    `gorm:"column:LastAnswer" json:"last_answer"`
	LastChengyu           string    `gorm:"column:LastChengyu" json:"last_chengyu"`
	LastChengyuDate       time.Time `gorm:"column:LastChengyuDate" json:"last_chengyu_date"`
	TrialStartDate        time.Time `gorm:"column:TrialStartDate" json:"trial_start_date"`
	TrialEndDate          time.Time `gorm:"column:TrialEndDate" json:"trial_end_date"`
	LastExitHintDate      time.Time `gorm:"column:LastExitHintDate" json:"last_exit_hint_date"`
	BlockRes              string    `gorm:"column:BlockRes" json:"block_res"`
	BlockType             int       `gorm:"column:BlockType" json:"block_type"`
	BlockFee              int       `gorm:"column:BlockFee" json:"block_fee"`
	GroupGuid             string    `gorm:"column:GroupGuid" json:"group_guid"`
	InsertDate            time.Time `gorm:"column:InsertDate" json:"insert_date"`
}

func (Sz84Group) TableName() string {
	return "Group"
}

// Sz84CreditLog represents the Credit table (CreditLog in C#)
type Sz84CreditLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin      int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupID     int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName   string    `gorm:"column:GroupName" json:"group_name"`
	UserID      int64     `gorm:"column:UserId" json:"user_id"`
	UserName    string    `gorm:"column:UserName" json:"user_name"`
	CreditAdd   int64     `gorm:"column:CreditAdd" json:"credit_add"`
	CreditValue int64     `gorm:"column:CreditValue" json:"credit_value"`
	CreditInfo  string    `gorm:"column:CreditInfo" json:"credit_info"`
	InsertDate  time.Time `gorm:"column:InsertDate;default:CURRENT_TIMESTAMP" json:"insert_date"`
}

func (Sz84CreditLog) TableName() string {
	return "CreditLog"
}

// Sz84TokensLog represents the Tokens table (TokensLog in C#)
type Sz84TokensLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin      int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupID     int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName   string    `gorm:"column:GroupName" json:"group_name"`
	UserID      int64     `gorm:"column:UserId" json:"user_id"`
	UserName    string    `gorm:"column:UserName" json:"user_name"`
	TokensAdd   int64     `gorm:"column:TokensAdd" json:"tokens_add"`
	TokensValue int64     `gorm:"column:TokensValue" json:"tokens_value"`
	TokensInfo  string    `gorm:"column:TokensInfo" json:"tokens_info"`
	InsertDate  time.Time `gorm:"column:InsertDate;default:CURRENT_TIMESTAMP" json:"insert_date"`
}

func (Sz84TokensLog) TableName() string {
	return "TokensLog"
}

// Sz84MsgCount represents the MsgCount table
type Sz84MsgCount struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin    int64      `gorm:"column:BotUin" json:"bot_uin"`
	GroupID   int64      `gorm:"column:GroupId" json:"group_id"`
	GroupName string     `gorm:"column:GroupName" json:"group_name"`
	UserID    int64      `gorm:"column:UserId" json:"user_id"`
	UserName  string     `gorm:"column:UserName" json:"user_name"`
	CDate     time.Time  `gorm:"column:CDate;type:date" json:"c_date"`
	CMsg      int        `gorm:"column:CMsg" json:"c_msg"`
	MsgDate   *time.Time `gorm:"column:MsgDate" json:"msg_date"`
}

func (Sz84MsgCount) TableName() string {
	return "MsgCount"
}

// BlackList represents the BlackList table
type BlackList struct {
	BotUin     int64     `gorm:"primaryKey;column:BotUin"`
	GroupID    int64     `gorm:"primaryKey;column:GroupId"`
	BlackID    int64     `gorm:"primaryKey;column:BlackId"`
	BlackInfo  string    `gorm:"column:BlackInfo"`
	InsertDate time.Time `gorm:"column:InsertDate;default:CURRENT_TIMESTAMP"`
}

func (BlackList) TableName() string {
	return "BlackList"
}

// VIPInfo represents the Vips table
type VIPInfo struct {
	GroupID    int64     `gorm:"primaryKey;column:GroupId"`
	GroupName  string    `gorm:"column:GroupName"`
	FirstPay   float64   `gorm:"column:FirstPay"`
	StartDate  time.Time `gorm:"column:StartDate"`
	EndDate    time.Time `gorm:"column:EndDate"`
	VIPInfo    string    `gorm:"column:VipInfo"`
	UserID     int64     `gorm:"column:UserId"`
	IncomeDay  float64   `gorm:"column:IncomeDay"`
	IsYearVIP  bool      `gorm:"column:IsYearVip"`
	InsertBy   int       `gorm:"column:InsertBy"`
	InsertDate time.Time `gorm:"column:InsertDate;default:CURRENT_TIMESTAMP"`
}

func (VIPInfo) TableName() string {
	return "VIPInfo"
}
