package models

import (
	"time"
)

// UserInfo represents the User table
type UserInfo struct {
	Id             int64     `gorm:"primaryKey;column:Id" json:"id"`
	Name           string    `gorm:"column:Name" json:"name"`
	UserOpenId     string    `gorm:"column:UserOpenId" json:"user_open_id"`
	InsertDate     time.Time `gorm:"column:InsertDate;autoCreateTime" json:"insert_date"`
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

	// Csz related fields from Cov mapping
	CszRes    int   `gorm:"column:CszRes" json:"csz_res"`
	CszCredit int64 `gorm:"column:CszCredit" json:"csz_credit"`
	CszTimes  int   `gorm:"column:CszTimes" json:"csz_times"`

	// OpenID related
	GroupOpenid string `gorm:"column:GroupOpenid" json:"group_openid"`
}

func (UserInfo) TableName() string {
	return "User"
}

// GroupInfo represents the Group table
type GroupInfo struct {
	Id                    int64     `gorm:"primaryKey;column:Id" json:"id"`
	GroupOpenId           string    `gorm:"-" json:"group_open_id"` // DbIgnore
	TargetGroup           int64     `gorm:"-" json:"target_group"`  // DbIgnore
	GroupName             string    `gorm:"column:GroupName" json:"group_name"`
	IsValid               bool      `gorm:"-" json:"is_valid"` // DbIgnore
	IsProxy               bool      `gorm:"-" json:"is_proxy"` // DbIgnore
	GroupMemo             string    `gorm:"column:GroupMemo" json:"group_memo"`
	GroupOwner            int64     `gorm:"-" json:"group_owner"` // DbIgnore
	GroupOwnerName        string    `gorm:"column:GroupOwnerName" json:"group_owner_name"`
	GroupOwnerNickname    string    `gorm:"column:GroupOwnerNickname" json:"group_owner_nickname"`
	GroupType             int       `gorm:"column:GroupType" json:"group_type"`
	RobotOwner            int64     `gorm:"-" json:"robot_owner"` // DbIgnore
	RobotOwnerName        string    `gorm:"column:RobotOwnerName" json:"robot_owner_name"`
	WelcomeMessage        string    `gorm:"column:WelcomeMessage" json:"welcome_message"`
	GroupState            int       `gorm:"column:GroupState" json:"group_state"`
	BotUin                int64     `gorm:"-" json:"bot_uin"` // DbIgnore
	BotName               string    `gorm:"column:BotName" json:"bot_name"`
	LastDate              time.Time `gorm:"column:LastDate" json:"last_date"`
	IsInGame              int       `gorm:"-" json:"is_in_game"` // DbIgnore
	IsOpen                bool      `gorm:"column:IsOpen" json:"is_open"`
	UseRight              int       `gorm:"column:UseRight" json:"use_right"`
	TeachRight            int       `gorm:"column:TeachRight" json:"teach_right"`
	AdminRight            int       `gorm:"column:AdminRight" json:"admin_right"`
	QuietTime             time.Time `gorm:"-" json:"quiet_time"` // DbIgnore
	IsCloseManager        bool      `gorm:"column:IsCloseManager" json:"is_close_manager"`
	IsAcceptNewMember     int       `gorm:"column:IsAcceptNewMember" json:"is_accept_new_member"`
	CloseRegex            string    `gorm:"-" json:"close_regex"` // DbIgnore
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
	IsSaveRecord          bool      `gorm:"-" json:"is_save_record"` // DbIgnore
	IsPause               bool      `gorm:"-" json:"is_pause"`       // DbIgnore
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
	LastAnswer            string    `gorm:"-" json:"last_answer"`         // DbIgnore
	LastChengyu           string    `gorm:"-" json:"last_chengyu"`        // DbIgnore
	LastChengyuDate       time.Time `gorm:"-" json:"last_chengyu_date"`   // DbIgnore
	TrialStartDate        time.Time `gorm:"-" json:"trial_start_date"`    // DbIgnore
	TrialEndDate          time.Time `gorm:"-" json:"trial_end_date"`      // DbIgnore
	LastExitHintDate      time.Time `gorm:"-" json:"last_exit_hint_date"` // DbIgnore
	BlockRes              string    `gorm:"-" json:"block_res"`           // DbIgnore
	BlockType             int       `gorm:"-" json:"block_type"`          // DbIgnore
	BlockMin              int       `gorm:"column:BlockMin" json:"block_min"`
	BlockFee              int       `gorm:"-" json:"block_fee"`  // DbIgnore
	GroupGuid             string    `gorm:"-" json:"group_guid"` // DbIgnore
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
	InsertDate            time.Time `gorm:"-" json:"insert_date"` // DbIgnore
	IsSendHelpInfo        bool      `gorm:"column:IsSendHelpInfo" json:"is_send_help_info"`
	IsRecall              bool      `gorm:"column:IsRecall" json:"is_recall"`
	IsCreditSystem        bool      `gorm:"column:IsCreditSystem" json:"is_credit_system"`
}

func (GroupInfo) TableName() string {
	return "Group"
}

// GroupMember represents the GroupMember table
type GroupMember struct {
	GroupId         int64      `gorm:"primaryKey;column:GroupId" json:"group_id"`
	UserId          int64      `gorm:"primaryKey;column:UserId" json:"user_id"`
	UserName        string     `gorm:"column:UserName" json:"user_name"`
	DisplayName     string     `gorm:"column:DisplayName" json:"display_name"`
	GroupCredit     int64      `gorm:"column:GroupCredit" json:"group_credit"`
	ConfirmCode     string     `gorm:"column:ConfirmCode" json:"confirm_code"`
	Status          int        `gorm:"column:Status" json:"status"`
	SignTimes       int        `gorm:"column:SignTimes" json:"sign_times"`
	SignLevel       int        `gorm:"column:SignLevel" json:"sign_level"`
	SignTimesAll    int        `gorm:"column:SignTimesAll" json:"sign_times_all"`
	SignDate        *time.Time `gorm:"column:SignDate" json:"sign_date"`
	IsFans          bool       `gorm:"column:IsFans" json:"is_fans"`
	FansDate        time.Time  `gorm:"column:FansDate" json:"fans_date"`
	FansLevel       int        `gorm:"column:FansLevel" json:"fans_level"`
	FansValue       int64      `gorm:"column:FansValue" json:"fans_value"`
	LampDate        time.Time  `gorm:"column:LampDate" json:"lamp_date"`
	InvitorUserId   int64      `gorm:"column:InvitorUserId" json:"invitor_user_id"`
	InviteCount     int        `gorm:"column:InviteCount" json:"invite_count"`
	InviteExitCount int        `gorm:"column:InviteExitCount" json:"invite_exit_count"`
	GoldCoins       int64      `gorm:"column:GoldCoins" json:"gold_coins"`
	PurpleCoins     int64      `gorm:"column:PurpleCoins" json:"purple_coins"`
	BlackCoins      int64      `gorm:"column:BlackCoins" json:"black_coins"`
	GameCoins       int64      `gorm:"column:GameCoins" json:"game_coins"`
	SaveCredit      int64      `gorm:"column:SaveCredit" json:"save_credit"`
}

func (GroupMember) TableName() string {
	return "GroupMember"
}

// BlackList represents the BlackList table
type BlackList struct {
	BotUin    int64  `gorm:"column:BotUin" json:"bot_uin"`
	GroupId   int64  `gorm:"primaryKey;column:GroupId" json:"group_id"`
	GroupName string `gorm:"column:GroupName" json:"group_name"`
	UserId    int64  `gorm:"column:UserId" json:"user_id"`
	UserName  string `gorm:"column:UserName" json:"user_name"`
	BlackId   int64  `gorm:"primaryKey;column:BlackId" json:"black_id"`
	BlackInfo string `gorm:"column:BlackInfo" json:"black_info"`
}

func (BlackList) TableName() string {
	return "BlackList"
}

// WhiteList represents the WhiteList table
type WhiteList struct {
	BotUin    int64  `gorm:"column:BotUin" json:"bot_uin"`
	GroupId   int64  `gorm:"primaryKey;column:GroupId" json:"group_id"`
	GroupName string `gorm:"column:GroupName" json:"group_name"`
	UserId    int64  `gorm:"column:UserId" json:"user_id"`
	UserName  string `gorm:"column:UserName" json:"user_name"`
	WhiteId   int64  `gorm:"primaryKey;column:WhiteId" json:"white_id"`
}

func (WhiteList) TableName() string {
	return "WhiteList"
}

// Friend represents the Friend table
type Friend struct {
	BotUin     int64  `gorm:"primaryKey;column:BotUin" json:"bot_uin"`
	UserId     int64  `gorm:"primaryKey;column:UserId" json:"user_id"`
	UserName   string `gorm:"column:UserName" json:"user_name"`
	Credit     int64  `gorm:"column:Credit" json:"credit"`
	SaveCredit int64  `gorm:"column:SaveCredit" json:"save_credit"`
}

func (Friend) TableName() string {
	return "Friend"
}

// BotCmd represents the Cmd table
type BotCmd struct {
	Id      int    `gorm:"primaryKey;column:Id" json:"id"`
	CmdName string `gorm:"column:CmdName" json:"cmd_name"`
	CmdText string `gorm:"column:CmdText" json:"cmd_text"`
	IsClose bool   `gorm:"column:IsClose" json:"is_close"`
}

func (BotCmd) TableName() string {
	return "Cmd"
}

// GroupVip represents the VIP table
type GroupVip struct {
	GroupId   int64     `gorm:"primaryKey;column:GroupId" json:"group_id"`
	GroupName string    `gorm:"column:GroupName" json:"group_name"`
	FirstPay  float64   `gorm:"column:FirstPay" json:"first_pay"`
	StartDate time.Time `gorm:"column:StartDate" json:"start_date"`
	EndDate   time.Time `gorm:"column:EndDate" json:"end_date"`
	VipInfo   string    `gorm:"column:VipInfo" json:"vip_info"`
	UserId    int64     `gorm:"column:UserId" json:"user_id"`
	IncomeDay float64   `gorm:"column:IncomeDay" json:"income_day"`
	IsYearVip bool      `gorm:"column:IsYearVip" json:"is_year_vip"`
	InsertBy  int       `gorm:"column:InsertBy" json:"insert_by"`
	IsGoon    *bool     `gorm:"column:IsGoon" json:"is_goon"` // Nullable
}

func (GroupVip) TableName() string {
	return "VIP"
}

// BotInfo represents the Member table (Bot information)
type BotInfo struct {
	BotUin         int64     `gorm:"primaryKey;column:BotUin" json:"bot_uin"`
	Password       string    `gorm:"column:Password" json:"password"`
	BotName        string    `gorm:"column:BotName" json:"bot_name"`
	BotType        int       `gorm:"column:BotType" json:"bot_type"`
	AdminId        int64     `gorm:"column:AdminId" json:"admin_id"`
	InsertDate     time.Time `gorm:"column:InsertDate" json:"insert_date"`
	BotMemo        string    `gorm:"column:BotMemo" json:"bot_memo"`
	WemcomeMessage string    `gorm:"column:WemcomeMessage" json:"wemcome_message"`
	ApiIP          string    `gorm:"column:ApiIP" json:"api_ip"`
	ApiPort        string    `gorm:"column:ApiPort" json:"api_port"`
	ApiKey         string    `gorm:"column:ApiKey" json:"api_key"`
	WebUIToken     string    `gorm:"column:WebUIToken" json:"web_ui_token"`
	WebUIPort      string    `gorm:"column:WebUIPort" json:"web_ui_port"`
	IsSignalR      bool      `gorm:"column:IsSignalR" json:"is_signal_r"`
	IsCredit       bool      `gorm:"column:IsCredit" json:"is_credit"`
	IsGroup        bool      `gorm:"column:IsGroup" json:"is_group"`
	IsPrivate      bool      `gorm:"column:IsPrivate" json:"is_private"`
	Valid          int       `gorm:"column:Valid" json:"valid"`
	IsFreeze       bool      `gorm:"column:IsFreeze" json:"is_freeze"`
	FreezeTimes    int       `gorm:"column:FreezeTimes" json:"freeze_times"`
	IsBlock        bool      `gorm:"column:IsBlock" json:"is_block"`
	IsVip          bool      `gorm:"column:IsVip" json:"is_vip"`

	// Ignored fields
	ValidDate     time.Time `gorm:"-" json:"valid_date"`
	LastDate      time.Time `gorm:"-" json:"last_date"`
	FreezeDate    time.Time `gorm:"-" json:"freeze_date"`
	BlockDate     time.Time `gorm:"-" json:"block_date"`
	HeartbeatDate time.Time `gorm:"-" json:"heartbeat_date"`
	ReceiveDate   time.Time `gorm:"-" json:"receive_date"`
}

func (BotInfo) TableName() string {
	return "Member"
}

// CoinsLog represents the Coins table
type CoinsLog struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin     int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId    int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName  string    `gorm:"column:GroupName" json:"group_name"`
	UserId     int64     `gorm:"column:UserId" json:"user_id"`
	UserName   string    `gorm:"column:UserName" json:"user_name"`
	CoinsType  int       `gorm:"column:CoinsType" json:"coins_type"`
	CoinsAdd   int64     `gorm:"column:CoinsAdd" json:"coins_add"`
	CoinsValue int64     `gorm:"column:CoinsValue" json:"coins_value"`
	CoinsInfo  string    `gorm:"column:CoinsInfo" json:"coins_info"`
	InsertDate time.Time `gorm:"column:InsertDate" json:"insert_date"`
}

func (CoinsLog) TableName() string {
	return "Coins"
}

// CreditLog represents the Credit table
type CreditLog struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin      int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId     int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName   string    `gorm:"column:GroupName" json:"group_name"`
	UserId      int64     `gorm:"column:UserId" json:"user_id"`
	UserName    string    `gorm:"column:UserName" json:"user_name"`
	CreditAdd   int64     `gorm:"column:CreditAdd" json:"credit_add"`
	CreditValue int64     `gorm:"column:CreditValue" json:"credit_value"`
	CreditInfo  string    `gorm:"column:CreditInfo" json:"credit_info"`
	InsertDate  time.Time `gorm:"column:InsertDate;autoCreateTime" json:"insert_date"`
}

func (CreditLog) TableName() string {
	return "Credit"
}

// TokensLog represents the Tokens table
type TokensLog struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin      int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId     int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName   string    `gorm:"column:GroupName" json:"group_name"`
	UserId      int64     `gorm:"column:UserId" json:"user_id"`
	UserName    string    `gorm:"column:UserName" json:"user_name"`
	TokensAdd   int64     `gorm:"column:TokensAdd" json:"tokens_add"`
	TokensValue int64     `gorm:"column:TokensValue" json:"tokens_value"`
	TokensInfo  string    `gorm:"column:TokensInfo" json:"tokens_info"`
	InsertDate  time.Time `gorm:"column:InsertDate;autoCreateTime" json:"insert_date"`
}

func (TokensLog) TableName() string {
	return "Tokens"
}

// MsgCount represents the MsgCount table
type MsgCount struct {
	Id        int       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin    int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId   int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName string    `gorm:"column:GroupName" json:"group_name"`
	UserId    int64     `gorm:"column:UserId" json:"user_id"`
	UserName  string    `gorm:"column:UserName" json:"user_name"`
	CDate     time.Time `gorm:"column:CDate" json:"c_date"`
	CMsg      int       `gorm:"column:CMsg" json:"c_msg"`
	MsgDate   time.Time `gorm:"column:MsgDate;autoCreateTime" json:"msg_date"`
}

func (MsgCount) TableName() string {
	return "MsgCount"
}

// Gift represents the Gift table
type Gift struct {
	Id         int64  `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	GiftName   string `gorm:"column:GiftName" json:"gift_name"`
	GiftCredit int64  `gorm:"column:GiftCredit" json:"gift_credit"`
	GiftUrl    string `gorm:"column:GiftUrl" json:"gift_url"`
	GiftImage  string `gorm:"column:GiftImage" json:"gift_image"`
	GiftType   int    `gorm:"column:GiftType" json:"gift_type"` // 1: normal, 2: advanced
	IsValid    bool   `gorm:"column:IsValid" json:"is_valid"`
}

func (Gift) TableName() string {
	return "Gift"
}

// GiftLog represents the GiftLog table
type GiftLog struct {
	Id           int64     `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotUin       int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId      int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName    string    `gorm:"column:GroupName" json:"group_name"`
	UserId       int64     `gorm:"column:UserId" json:"user_id"`
	UserName     string    `gorm:"column:UserName" json:"user_name"`
	RobotOwner   int64     `gorm:"column:RobotOwner" json:"robot_owner"`
	OwnerName    string    `gorm:"column:OwnerName" json:"owner_name"`
	GiftUserId   int64     `gorm:"column:GiftUserId" json:"gift_user_id"`
	GiftUserName string    `gorm:"column:GiftUserName" json:"gift_user_name"`
	GiftId       int64     `gorm:"column:GiftId" json:"gift_id"`
	GiftName     string    `gorm:"column:GiftName" json:"gift_name"`
	GiftCount    int       `gorm:"column:GiftCount" json:"gift_count"`
	GiftCredit   int64     `gorm:"column:GiftCredit" json:"gift_credit"`
	InsertDate   time.Time `gorm:"column:InsertDate;default:getdate()" json:"insert_date"`
}

func (GiftLog) TableName() string {
	return "GiftLog"
}

// RobotWeibo represents the robot_weibo table (Sign-in log)
type RobotWeibo struct {
	WeiboId    int       `gorm:"primaryKey;autoIncrement;column:weibo_id" json:"weibo_id"`
	RobotQQ    int64     `gorm:"column:robot_qq" json:"robot_qq"`
	WeiboQQ    int64     `gorm:"column:weibo_qq" json:"weibo_qq"`
	WeiboInfo  string    `gorm:"column:weibo_info" json:"weibo_info"`
	WeiboType  int       `gorm:"column:weibo_type" json:"weibo_type"`
	GroupId    int64     `gorm:"column:group_id" json:"group_id"`
	InsertDate time.Time `gorm:"column:insert_date;autoCreateTime" json:"insert_date"`
}

func (RobotWeibo) TableName() string {
	return "robot_weibo"
}

// Title represents the Title table
type Title struct {
	Id              string `gorm:"primaryKey;column:Id" json:"id"`
	Name            string `gorm:"column:Name" json:"name"`
	Description     string `gorm:"column:Description" json:"description"`
	UnlockCondition string `gorm:"column:UnlockCondition" json:"unlock_condition"`
	IsHidden        bool   `gorm:"column:IsHidden" json:"is_hidden"`
	IsExclusive     bool   `gorm:"column:IsExclusive" json:"is_exclusive"`
	Icon            string `gorm:"column:Icon" json:"icon"`
}

func (Title) TableName() string {
	return "Title"
}

// UserTitle represents the UserTitle table
type UserTitle struct {
	UserId     int64     `gorm:"primaryKey;column:UserId" json:"user_id"`
	TitleId    string    `gorm:"primaryKey;column:TitleId" json:"title_id"`
	UnlockTime time.Time `gorm:"column:UnlockTime" json:"unlock_time"`
	IsEquipped bool      `gorm:"column:IsEquipped" json:"is_equipped"`
}

func (UserTitle) TableName() string {
	return "UserTitle"
}
