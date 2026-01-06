package models

import (
	"time"

	"gorm.io/gorm"
)

// Member represents the bot (Member table in legacy C#)
type Member struct {
	BotUin         int64     `gorm:"primaryKey;column:BotUin" json:"bot_uin"`
	Password       string    `gorm:"column:Password" json:"password"`
	BotName        string    `gorm:"column:BotName" json:"bot_name"`
	BotType        int       `gorm:"column:BotType" json:"bot_type"`
	AdminId        int64     `gorm:"column:AdminId" json:"admin_id"`
	InsertDate     time.Time `gorm:"column:InsertDate" json:"insert_date"`
	BotMemo        string    `gorm:"column:BotMemo" json:"bot_memo"`
	WemcomeMessage string    `gorm:"column:WemcomeMessage" json:"welcome_message"`
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
	ValidDate      time.Time `gorm:"column:ValidDate" json:"valid_date"`
	LastDate       time.Time `gorm:"column:LastDate" json:"last_date"`
	IsFreeze       bool      `gorm:"column:IsFreeze" json:"is_freeze"`
	FreezeDate     time.Time `gorm:"column:FreezeDate" json:"freeze_date"`
	FreezeTimes    int       `gorm:"column:FreezeTimes" json:"freeze_times"`
	IsBlock        bool      `gorm:"column:IsBlock" json:"is_block"`
	BlockDate      time.Time `gorm:"column:BlockDate" json:"block_date"`
	HeartbeatDate  time.Time `gorm:"column:HeartbeatDate" json:"heartbeat_date"`
	ReceiveDate    time.Time `gorm:"column:ReceiveDate" json:"receive_date"`
	IsVip          bool      `gorm:"column:IsVip" json:"is_vip"`
	BotGuid        string    `gorm:"column:BotGuid" json:"bot_guid"`
	// Keep these for system status
	Connected bool           `gorm:"-" json:"connected"`
	LastSeen  time.Time      `gorm:"-" json:"last_seen"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (Member) TableName() string {
	return "Member"
}

// MessageLog represents a message log in the database
type MessageLog struct {
	Id        uint      `gorm:"primaryKey;column:Id" json:"id"`
	BotId     string    `gorm:"index;size:64;column:BotId" json:"bot_id"`
	UserId    string    `gorm:"index;size:64;column:UserId" json:"user_id"`
	GroupId   string    `gorm:"index;size:64;column:GroupId" json:"group_id"`
	Content   string    `gorm:"type:text;column:Content" json:"content"`
	RawData   string    `gorm:"type:text;column:RawData" json:"raw_data"`
	Platform  string    `gorm:"size:32;column:Platform" json:"platform"`
	Direction string    `gorm:"size:16;column:Direction" json:"direction"` // incoming, outgoing
	CreatedAt time.Time `gorm:"index;column:CreatedAt" json:"created_at"`
}

func (MessageLog) TableName() string {
	return "MessageLog"
}

// Session represents a session in the database
type Session struct {
	Id        uint      `gorm:"primaryKey;column:Id" json:"id"`
	SessionId string    `gorm:"uniqueIndex;size:128;column:SessionId" json:"session_id"`
	UserId    int64     `gorm:"index;column:UserId" json:"user_id"`
	GroupId   int64     `gorm:"index;column:GroupId" json:"group_id"`
	State     string    `gorm:"size:64;column:State" json:"state"`
	Data      string    `gorm:"type:text;column:Data" json:"data"`
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (Session) TableName() string {
	return "Session"
}

// RoutingRule 路由规则表模型
type RoutingRule struct {
	Id             uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Pattern        string    `gorm:"uniqueIndex;not null;size:500;column:Pattern" json:"pattern"`
	TargetWorkerId string    `gorm:"not null;size:255;column:TargetWorkerId" json:"target_worker_id"`
	CreatedAt      time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (RoutingRule) TableName() string {
	return "RoutingRule"
}

// GroupCache 群组缓存表模型
type GroupCache struct {
	GroupID   string    `gorm:"primaryKey;size:255;column:GroupId" json:"group_id"`
	GroupName string    `gorm:"size:255;column:GroupName" json:"group_name"`
	BotID     string    `gorm:"size:255;column:BotId" json:"bot_id"`
	LastSeen  time.Time `gorm:"column:LastSeen" json:"last_seen"`
}

// TableName 设置表名
func (GroupCache) TableName() string {
	return "GroupCache"
}

// MemberCache 群成员缓存表模型
type MemberCache struct {
	ID       uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	GroupID  string    `gorm:"index:idx_group_user,unique;size:255;column:GroupId" json:"group_id"`
	UserID   string    `gorm:"index:idx_group_user,unique;size:255;column:UserId" json:"user_id"`
	Nickname string    `gorm:"size:255;column:Nickname" json:"nickname"`
	Card     string    `gorm:"size:255;column:Card" json:"card"`
	Role     string    `gorm:"size:50;column:Role" json:"role"`
	LastSeen time.Time `gorm:"column:LastSeen" json:"last_seen"`
}

func (MemberCache) TableName() string {
	return "MemberCache"
}

// MessageStat 消息统计表
type MessageStat struct {
	ID      uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Date    time.Time `gorm:"index:idx_date_group_user;type:date;column:Date" json:"date"`
	GroupID string    `gorm:"index:idx_date_group_user;size:64;column:GroupId" json:"group_id"`
	UserID  string    `gorm:"index:idx_date_group_user;size:64;column:UserId" json:"user_id"`
	Count   int64     `gorm:"column:Count" json:"count"`
}

func (MessageStat) TableName() string {
	return "MessageStat"
}

// UserLoginToken 临时登录验证码表
type UserLoginToken struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Platform   string    `gorm:"size:32;not null;column:Platform" json:"platform"`
	PlatformID string    `gorm:"index;size:64;not null;column:PlatformId" json:"platform_id"`
	Token      string    `gorm:"size:16;not null;column:Token" json:"token"`
	ExpiresAt  time.Time `gorm:"column:ExpiresAt" json:"expires_at"`
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

func (UserLoginToken) TableName() string {
	return "UserLoginToken"
}

// FriendCache 好友缓存表模型
type FriendCache struct {
	UserID   string    `gorm:"primaryKey;size:255;column:UserId" json:"user_id"`
	Nickname string    `gorm:"size:255;column:Nickname" json:"nickname"`
	Remark   string    `gorm:"size:255;column:Remark" json:"remark"`
	BotID    string    `gorm:"size:255;column:BotId" json:"bot_id"`
	LastSeen time.Time `gorm:"column:LastSeen" json:"last_seen"`
}

// TableName 设置表名
func (FriendCache) TableName() string {
	return "FriendCache"
}

// AIProvider AI 提供商配置
type AIProvider struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name      string         `gorm:"size:100;not null;column:Name" json:"name"`
	Type      string         `gorm:"size:50;not null;column:Type" json:"type"` // openai, ollama, deepseek, etc.
	BaseURL   string         `gorm:"size:255;column:BaseUrl" json:"base_url"`
	APIKey    string         `gorm:"size:500;column:ApiKey" json:"api_key"` // 存储时加密
	IsEnabled bool           `gorm:"default:true;column:IsEnabled" json:"is_enabled"`
	Priority  int            `gorm:"default:1;column:Priority" json:"priority"`
	UserID    *uint          `gorm:"index;column:UserId" json:"user_id"` // 为空表示系统公共配置，不为空表示用户私有配置
	CreatedAt time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

// TableName 设置表名
func (AIProvider) TableName() string {
	return "AIProvider"
}

// AIModel AI 具体模型
type AIModel struct {
	ID           uint       `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	ProviderID   uint       `gorm:"index;column:ProviderId" json:"provider_id"`
	Provider     AIProvider `gorm:"foreignKey:ProviderID" json:"provider"`
	ApiModelID   string     `gorm:"type:varchar(100);column:ApiModelId" json:"model_id"` // 实际 API 调用的模型 ID (如 gpt-4)
	ModelName    string     `gorm:"size:100;column:ModelName" json:"model_name"`         // 展示名称 (如 GPT-4 Turbo)
	Capabilities string     `gorm:"size:255;column:Capabilities" json:"capabilities"`    // JSON array: ["chat", "vision"]
	BaseURL      string     `gorm:"size:255;column:BaseUrl" json:"base_url"`             // 模型级别 BaseURL 覆盖
	APIKey       string     `gorm:"size:500;column:ApiKey" json:"api_key"`               // 模型级别 APIKey 覆盖
	ContextSize  int        `gorm:"default:4096;column:ContextSize" json:"context_size"`
	IsDefault    bool       `gorm:"default:false;column:IsDefault" json:"is_default"`
	CreatedAt    time.Time  `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AIModel) TableName() string {
	return "AIModel"
}

// AIPromptTemplate 提示词模板
type AIPromptTemplate struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Scene     string    `gorm:"size:100;uniqueIndex;column:Scene" json:"scene"` // 使用场景: intent_recognition, reply_gen, etc.
	Content   string    `gorm:"type:text;column:Content" json:"content"`
	Version   string    `gorm:"size:20;column:Version" json:"version"`
	IsActive  bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AIPromptTemplate) TableName() string {
	return "AIPromptTemplate"
}

// AIKnowledgeBase 知识库配置 (RAG)
type AIKnowledgeBase struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name        string    `gorm:"size:100;not null;column:Name" json:"name"`
	Description string    `gorm:"size:255;column:Description" json:"description"`
	OwnerID     uint      `gorm:"index;column:OwnerId" json:"owner_id"` // 所属用户
	IsPublic    bool      `gorm:"default:false;column:IsPublic" json:"is_public"`
	EmbedModel  string    `gorm:"size:100;column:EmbedModel" json:"embed_model"` // 使用的向量化模型
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AIKnowledgeBase) TableName() string {
	return "AIKnowledgeBase"
}

// AIUsageLog AI 使用日志
type AIUsageLog struct {
	ID              uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	UserID          uint      `gorm:"index;column:UserId" json:"user_id"`
	AgentID         uint      `gorm:"index;column:AgentId" json:"agent_id"` // 关联智能体 ID
	ModelName       string    `gorm:"size:100;column:ModelName" json:"model_name"`
	ProviderType    string    `gorm:"size:50;column:ProviderType" json:"provider_type"`
	InputTokens     int       `gorm:"column:InputTokens" json:"input_tokens"`
	OutputTokens    int       `gorm:"column:OutputTokens" json:"output_tokens"`
	TotalTokens     int       `gorm:"column:TotalTokens" json:"total_tokens"`
	DurationMS      int       `gorm:"column:DurationMs" json:"duration_ms"`
	Status          string    `gorm:"size:20;column:Status" json:"status"`
	RevenueDeducted int       `gorm:"default:0;column:RevenueDeducted" json:"revenue_deducted"`
	CreatedAt       time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

// TableName 设置表名
func (AIUsageLog) TableName() string {
	return "AIUsageLog"
}

// CognitiveMemory 认知记忆表
type CognitiveMemory struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	UserID     string    `gorm:"index;size:64;column:UserId" json:"user_id"`
	BotID      string    `gorm:"index;size:64;column:BotId" json:"bot_id"`
	Category   string    `gorm:"size:32;column:Category" json:"category"` // profile, preference, event, fact
	Content    string    `gorm:"type:text;column:Content" json:"content"` // 记忆内容
	Importance int       `gorm:"default:1;column:Importance" json:"importance"`
	Metadata   string    `gorm:"type:text;column:Metadata" json:"metadata"`   // 额外元数据 (JSON)
	Embedding  string    `gorm:"type:vector(1536);column:Embedding" json:"-"` // 向量存储 (支持 pgvector)
	LastSeen   time.Time `gorm:"column:LastSeen" json:"last_seen"`
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

// TableName 设置表名
func (CognitiveMemory) TableName() string {
	return "CognitiveMemory"
}

// MCPServer MCP 服务器配置模型
type MCPServer struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name        string         `gorm:"size:100;not null;column:Name" json:"name"`
	Description string         `gorm:"type:text;column:Description" json:"description"`
	Type        string         `gorm:"size:20;not null;column:Type" json:"type"` // http, sse, webhook, internal
	Endpoint    string         `gorm:"size:500;not null;column:Endpoint" json:"endpoint"`
	APIKey      string         `gorm:"size:500;column:ApiKey" json:"api_key"`
	Scope       string         `gorm:"size:20;default:'user';column:Scope" json:"scope"` // global, org, user
	OwnerID     uint           `gorm:"index;column:OwnerId" json:"owner_id"`             // 所属用户或组织 ID
	Status      string         `gorm:"size:20;default:'active';column:Status" json:"status"`
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (MCPServer) TableName() string {
	return "MCPServer"
}

// AISkill 技能定义
type AISkill struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null;column:Name" json:"name"`
	Description string    `gorm:"size:255;column:Description" json:"description"`
	Category    string    `gorm:"size:50;column:Category" json:"category"` // tools, entertainment, business, etc.
	Prompt      string    `gorm:"type:text;column:Prompt" json:"prompt"`
	Config      string    `gorm:"type:text;column:Config" json:"config"` // JSON 扩展配置
	OwnerID     uint      `gorm:"index;column:OwnerId" json:"owner_id"`
	IsPublic    bool      `gorm:"default:false;column:IsPublic" json:"is_public"`
	Status      string    `gorm:"size:20;default:'active';column:Status" json:"status"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AISkill) TableName() string {
	return "AISkill"
}

// AITrainingData 技能训练数据 (Few-shot / Fine-tuning)
type AITrainingData struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SkillID   uint      `gorm:"index;column:SkillId" json:"skill_id"`
	Input     string    `gorm:"type:text;not null;column:Input" json:"input"`
	Output    string    `gorm:"type:text;not null;column:Output" json:"output"`
	Source    string    `gorm:"size:50;column:Source" json:"source"` // manual, feedback, auto
	IsUsed    bool      `gorm:"default:true;column:IsUsed" json:"is_used"`
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AITrainingData) TableName() string {
	return "AITrainingData"
}

// AIIntent 意图定义
type AIIntent struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null;column:Name" json:"name"` // 意图名称: complaint, feedback, inquiry
	DisplayName string    `gorm:"size:100;column:DisplayName" json:"display_name"`       // 展示名称: 用户投诉, 意见反馈, 售前咨询
	Description string    `gorm:"size:500;column:Description" json:"description"`        // 详细描述该意图的判定标准
	Keywords    string    `gorm:"size:500;column:Keywords" json:"keywords"`              // 辅助判定的关键词 (JSON array)
	IsActive    bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AIIntent) TableName() string {
	return "AIIntent"
}

// AIIntentRouting 意图路由规则
type AIIntentRouting struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	IntentID   uint      `gorm:"index;column:IntentId" json:"intent_id"`
	TargetType string    `gorm:"size:20;column:TargetType" json:"target_type"` // skill, worker, plugin, bot_redirect
	TargetID   string    `gorm:"size:100;column:TargetId" json:"target_id"`    // AISkill.ID 或 WorkerID 或 BotID
	Priority   int       `gorm:"default:1;column:Priority" json:"priority"`
	Condition  string    `gorm:"type:text;column:Condition" json:"condition"` // 额外的过滤条件 (JSON)
	IsActive   bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (AIIntentRouting) TableName() string {
	return "AIIntentRouting"
}

// GroupBotRole 群组机器人角色配置 (用于用户侧多机协作)
type GroupBotRole struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	GroupID   string    `gorm:"index:idx_group_bot,unique;size:255;column:GroupId" json:"group_id"`
	BotID     string    `gorm:"index:idx_group_bot,unique;size:255;column:BotId" json:"bot_id"`
	Role      string    `gorm:"size:100;column:Role" json:"role"`           // 角色: dispatcher(调度官), handler(执行者), monitor(监控)
	Specialty string    `gorm:"size:255;column:Specialty" json:"specialty"` // 擅长领域: tech, finance, chat, etc.
	IsActive  bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (GroupBotRole) TableName() string {
	return "GroupBotRole"
}

// MCPTool MCP 工具缓存模型
type MCPTool struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	ServerID    uint      `gorm:"index;column:ServerId" json:"server_id"`
	Name        string    `gorm:"size:100;not null;column:Name" json:"name"`
	Description string    `gorm:"type:text;column:Description" json:"description"`
	InputSchema string    `gorm:"type:text;column:InputSchema" json:"input_schema"` // JSON Schema
	IsActive    bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (MCPTool) TableName() string {
	return "MCPTool"
}

// Enterprise 企业/组织模型
type Enterprise struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name        string         `gorm:"size:255;uniqueIndex;not null;column:Name" json:"name"` // 企业名称
	Code        string         `gorm:"size:100;uniqueIndex;not null;column:Code" json:"code"` // 企业唯一代码 (用于 B2B 通信)
	Description string         `gorm:"type:text;column:Description" json:"description"`
	OwnerID     uint           `gorm:"index;column:OwnerId" json:"owner_id"`                 // 企业所有者 (关联 UserGORM)
	Config      string         `gorm:"type:text;column:Config" json:"config"`                // 企业级全局配置 (JSON)
	Status      string         `gorm:"size:20;default:'active';column:Status" json:"status"` // active, suspended
	PublicKey   string         `gorm:"type:text;column:PublicKey" json:"public_key"`         // 用于 B2B 安全认证的公钥
	PrivateKey  string         `gorm:"type:text;column:PrivateKey" json:"private_key"`       // 用于 B2B 安全认证的私钥 (加密存储)
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

// TableName 设置表名
func (Enterprise) TableName() string {
	return "Enterprise"
}

// PlatformAccount 第三方平台账号配置 (公众号, 抖音等)
type PlatformAccount struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EnterpriseID uint      `gorm:"index;column:EnterpriseId" json:"enterprise_id"`
	Platform     string    `gorm:"size:50;not null;column:Platform" json:"platform"` // wechat_mp, tiktok, weibo, etc.
	AccountName  string    `gorm:"size:100;column:AccountName" json:"account_name"`  // 账号名称
	AccountID    string    `gorm:"size:100;column:AccountId" json:"account_id"`      // 平台内部 ID (如 AppID)
	Config       string    `gorm:"type:text;column:Config" json:"config"`            // 平台配置 (JSON: AppSecret, Token, AESKey 等)
	Status       string    `gorm:"size:20;default:'active';column:Status" json:"status"`
	CreatedAt    time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (PlatformAccount) TableName() string {
	return "PlatformAccount"
}

// DigitalEmployee 数字员工模型 (Bot 的高级拟人化抽象)
type DigitalEmployee struct {
	ID                uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EnterpriseID      uint      `gorm:"index;column:EnterpriseId" json:"enterprise_id"`                     // 所属企业
	BotID             string    `gorm:"uniqueIndex;size:64;column:BotId" json:"bot_id"`                     // 关联的底层机器人
	EmployeeID        string    `gorm:"uniqueIndex;size:64;column:EmployeeId" json:"employee_id"`           // 工号
	RoleTemplateID    uint      `gorm:"index;column:RoleTemplateId" json:"role_template_id"`                // 关联岗位模板 ID
	Name              string    `gorm:"size:100;not null;column:Name" json:"name"`                          // 姓名 (如: 张三)
	Title             string    `gorm:"size:100;column:Title" json:"title"`                                 // 职位 (如: 高级售后工程师)
	Level             string    `gorm:"size:50;column:Level" json:"level"`                                  // 职级 (如: P3, M2)
	Department        string    `gorm:"size:100;column:Department" json:"department"`                       // 部门 (如: 技术部)
	SupervisorID      uint      `gorm:"index;column:SupervisorId" json:"supervisor_id"`                     // 直属上级员工 ID
	Bio               string    `gorm:"type:text;column:Bio" json:"bio"`                                    // 个人简介/人设定义
	AgentID           uint      `gorm:"index;column:AgentId" json:"agent_id"`                               // 关联的 AI 智能体 ID
	Agent             AIAgent   `gorm:"foreignKey:AgentID" json:"agent"`                                    // 关联的 AI 智能体详情
	Skills            string    `gorm:"type:text;column:Skills" json:"skills"`                              // 技能列表 (JSON: ["complaint_handling", "log_analysis"])
	Permissions       string    `gorm:"type:text;column:Permissions" json:"permissions"`                    // 核心权限配置 (JSON)
	SecurityPolicy    string    `gorm:"type:text;column:SecurityPolicy" json:"security_policy"`             // 安全与脱敏策略 (JSON)
	OnboardingAt      time.Time `gorm:"column:OnboardingAt" json:"onboarding_at"`                           // 入职时间
	Status            string    `gorm:"size:20;default:'active';column:Status" json:"status"`               // 状态: active(在职), training(培训中), retired(离职)
	OnlineStatus      string    `gorm:"size:20;default:'offline';column:OnlineStatus" json:"online_status"` // online, offline, busy
	SalaryToken       int64     `gorm:"default:0;column:SalaryToken" json:"salary_token"`                   // 累计消耗 Token (作为薪资统计)
	SalaryLimit       int64     `gorm:"default:1000000;column:SalaryLimit" json:"salary_limit"`             // Token 预算限制
	KpiScore          float64   `gorm:"default:100;column:KpiScore" json:"kpi_score"`                       // KPI 评分 (基于满意度打分)
	ExternalCommLevel int       `gorm:"default:0;column:ExternalCommLevel" json:"external_comm_level"`      // 外部通信等级: 0(禁止), 1(仅限白名单企业), 2(公开)
	CreatedAt         time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// TableName 设置表名
func (DigitalEmployee) TableName() string {
	return "DigitalEmployee"
}

// DigitalRoleTemplate 岗位标准模板
type DigitalRoleTemplate struct {
	ID            uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name          string         `gorm:"size:100;uniqueIndex;not null;column:Name" json:"name"` // 模板名称: 行政助理, 技术支持等
	Description   string         `gorm:"type:text;column:Description" json:"description"`
	DefaultBio    string         `gorm:"type:text;column:DefaultBio" json:"default_bio"`
	DefaultSkills string         `gorm:"type:text;column:DefaultSkills" json:"default_skills"` // JSON array
	BasePrompt    string         `gorm:"type:text;column:BasePrompt" json:"base_prompt"`       // 岗位基础 Prompt
	SuggestedKPI  string         `gorm:"type:text;column:SuggestedKpi" json:"suggested_kpi"`   // 建议的 KPI 指标 (JSON)
	CreatedAt     time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

// TableName 设置表名
func (DigitalRoleTemplate) TableName() string {
	return "DigitalRoleTemplate"
}

// EnterpriseMember 企业成员表
type EnterpriseMember struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EnterpriseID uint      `gorm:"index:idx_ent_user,unique;column:EnterpriseId" json:"enterprise_id"`
	UserID       uint      `gorm:"index:idx_ent_user,unique;column:UserId" json:"user_id"`
	Role         string    `gorm:"size:50;default:'member';column:Role" json:"role"` // admin, hr, supervisor, member
	JoinedAt     time.Time `gorm:"column:JoinedAt" json:"joined_at"`
}

// TableName 设置表名
func (EnterpriseMember) TableName() string {
	return "EnterpriseMember"
}

// B2BConnection 企业间 B2B 连接
type B2BConnection struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SourceEntID  uint      `gorm:"index:idx_b2b_conn;column:SourceEntId" json:"source_ent_id"`      // 发起方企业
	TargetEntID  uint      `gorm:"index:idx_b2b_conn;column:TargetEntId" json:"target_ent_id"`      // 接收方企业
	Status       string    `gorm:"size:20;default:'pending';column:Status" json:"status"`           // pending, active, blocked
	AuthProtocol string    `gorm:"size:50;default:'mtls';column:AuthProtocol" json:"auth_protocol"` // mtls, oauth2, custom
	Config       string    `gorm:"type:text;column:Config" json:"config"`                           // 连接特定配置
	CreatedAt    time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

// B2BSkillSharing B2B 技能共享授权表
type B2BSkillSharing struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SourceEntID uint      `gorm:"index;column:SourceEntId" json:"source_ent_id"`         // 提供技能的企业
	TargetEntID uint      `gorm:"index;column:TargetEntId" json:"target_ent_id"`         // 使用技能的企业
	SkillName   string    `gorm:"size:100;column:SkillName" json:"skill_name"`           // 共享的技能名称
	AliasName   string    `gorm:"size:100;column:AliasName" json:"alias_name"`           // 在目标企业的别名 (可选)
	Status      string    `gorm:"size:20;default:'pending';column:Status" json:"status"` // pending, approved, rejected, blocked
	IsActive    bool      `gorm:"default:true;column:IsActive" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (B2BSkillSharing) TableName() string {
	return "B2BSkillSharing"
}

// DigitalEmployeeDispatch 数字员工外派授权表
type DigitalEmployeeDispatch struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EmployeeID  uint      `gorm:"index;column:EmployeeId" json:"employee_id"`            // 外派的员工 ID
	SourceEntID uint      `gorm:"index;column:SourceEntId" json:"source_ent_id"`         // 所属企业
	TargetEntID uint      `gorm:"index;column:TargetEntId" json:"target_ent_id"`         // 接收企业
	Status      string    `gorm:"size:20;default:'pending';column:Status" json:"status"` // pending, approved, rejected, recalled
	Permissions string    `gorm:"type:text;column:Permissions" json:"permissions"`       // 授予的权限列表 (JSON: ["chat", "skill_call"])
	DispatchAt  time.Time `gorm:"column:DispatchAt" json:"dispatch_at"`
	ExpireAt    time.Time `gorm:"column:ExpireAt" json:"expire_at"` // 有效期 (可选)
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (DigitalEmployeeDispatch) TableName() string {
	return "DigitalEmployeeDispatch"
}

// TableName 设置表名
func (B2BConnection) TableName() string {
	return "B2BConnection"
}

// DigitalEmployeeKpi 数字员工考核日志
type DigitalEmployeeKpi struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EmployeeID uint      `gorm:"index;column:EmployeeId" json:"employee_id"`
	MetricName string    `gorm:"size:100;column:MetricName" json:"metric_name"` // 考核项: response_speed, accuracy, satisfaction
	Score      float64   `gorm:"column:Score" json:"score"`
	Detail     string    `gorm:"type:text;column:Detail" json:"detail"` // 考核详情 (关联的消息 ID 或评价内容)
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

// DigitalEmployeeTodo 数字员工待办事项
type DigitalEmployeeTodo struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	EmployeeID  uint           `gorm:"index;column:EmployeeId" json:"employee_id"`  // 所属员工 ID
	Title       string         `gorm:"size:255;not null;column:Title" json:"title"` // 待办事项标题
	Description string         `gorm:"type:text;column:Description" json:"description"`
	Priority    string         `gorm:"size:20;default:'medium';column:Priority" json:"priority"` // low, medium, high
	Status      string         `gorm:"size:20;default:'pending';column:Status" json:"status"`    // pending, in_progress, completed, cancelled
	DueDate     *time.Time     `gorm:"column:DueDate" json:"due_date"`                           // 截止日期
	CompletedAt *time.Time     `gorm:"column:CompletedAt" json:"completed_at"`                   // 完成时间
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (DigitalEmployeeTodo) TableName() string {
	return "DigitalEmployeeTodo"
}

// DigitalEmployeeTask 数字员工任务全生命周期管理
type DigitalEmployeeTask struct {
	ID               uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	ExecutionID      string         `gorm:"size:100;uniqueIndex;column:ExecutionId" json:"execution_id"`      // 全局唯一执行 ID
	ParentTaskID     uint           `gorm:"index;column:ParentTaskId" json:"parent_task_id"`                  // 父任务 ID (用于子任务拆解)
	CreatorID        string         `gorm:"size:64;column:CreatorId" json:"creator_id"`                       // 创建者 (可以是 UserID 或 EmployeeID)
	AssigneeID       uint             `gorm:"index;column:AssigneeId" json:"assignee_id"`                       // 负责人 (数字员工 ID)
	Assignee         *DigitalEmployee `gorm:"foreignKey:AssigneeID" json:"assignee"`                            // 关联负责人详情
	TaskType         string           `gorm:"size:20;not null;default:'rule';column:TaskType" json:"task_type"` // rule, ai, hybrid
	Title            string         `gorm:"size:255;not null;column:Title" json:"title"`
	Description      string         `gorm:"type:text;column:Description" json:"description"`
	Context          string         `gorm:"type:text;column:Context" json:"context"`               // 任务上下文 (JSON)
	Steps            string         `gorm:"type:text;column:Steps" json:"steps"`                   // 执行步骤记录 (JSON)
	Result           string         `gorm:"type:text;column:Result" json:"result"`                 // 最终结果
	Status           string         `gorm:"size:20;default:'pending';column:Status" json:"status"` // pending, running, paused, done, failed
	Priority         string         `gorm:"size:20;default:'medium';column:Priority" json:"priority"`
	Progress         int            `gorm:"default:0;column:Progress" json:"progress"`                   // 0-100
	PlanRaw          string         `gorm:"type:text;column:PlanRaw" json:"plan_raw"`                    // AI 生成的任务计划 (JSON)
	CurrentStepIndex int            `gorm:"default:0;column:CurrentStepIndex" json:"current_step_index"` // 当前正在执行的步骤索引
	ResultRaw        string         `gorm:"type:text;column:ResultRaw" json:"result_raw"`                // 最终执行结果 (JSON)
	ErrorMsg         string         `gorm:"type:text;column:ErrorMsg" json:"error_msg"`                  // 错误信息
	TokenUsage       int            `gorm:"default:0;column:TokenUsage" json:"token_usage"`              // 任务总消耗 Token
	Duration         int            `gorm:"default:0;column:Duration" json:"duration"`                   // 执行耗时 (秒)
	StartTime        *time.Time     `gorm:"column:StartTime" json:"start_time"`
	EndTime          *time.Time     `gorm:"column:EndTime" json:"end_time"`
	Deadline         *time.Time     `gorm:"column:Deadline" json:"deadline"`
	CreatedAt        time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (DigitalEmployeeTask) TableName() string {
	return "DigitalEmployeeTask"
}

// TableName 设置表名
func (DigitalEmployeeKpi) TableName() string {
	return "DigitalEmployeeKpi"
}

// AIAgent 智能体定义 (LLM 核心配置)
type AIAgent struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Name         string    `gorm:"size:100;not null;column:Name" json:"name"`
	Description  string    `gorm:"type:text;column:Description" json:"description"`
	SystemPrompt string    `gorm:"type:text;column:SystemPrompt" json:"system_prompt"`
	ModelID      uint      `gorm:"index;column:ModelId" json:"model_id"`
	Model        AIModel   `gorm:"foreignKey:ModelID" json:"model"`
	Temperature  float32   `gorm:"default:0.7;column:Temperature" json:"temperature"`
	MaxTokens    int       `gorm:"default:2048;column:MaxTokens" json:"max_tokens"`
	Tools        string    `gorm:"type:text;column:Tools" json:"tools"` // JSON array of tool names or IDs
	IsVoice      bool      `gorm:"default:false;column:IsVoice" json:"is_voice"`
	VoiceID      string    `gorm:"size:100;column:VoiceId" json:"voice_id"`
	VoiceName    string    `gorm:"size:100;column:VoiceName" json:"voice_name"`
	VoiceLang    string    `gorm:"size:50;column:VoiceLang" json:"voice_lang"`
	VoiceRate    float32   `gorm:"default:1.0;column:VoiceRate" json:"voice_rate"`                      // 语速 0.1 - 10
	OwnerID      uint      `gorm:"index;column:OwnerId" json:"owner_id"`                                // 所属用户 ID
	Visibility   string    `gorm:"size:20;default:'public';column:Visibility" json:"visibility"`        // public, private, link_only
	RevenueRate  float64   `gorm:"type:decimal(10,4);default:0;column:RevenueRate" json:"revenue_rate"` // 收益率 (每 1k tokens 收益多少算力)
	CallCount    int       `gorm:"default:0;column:CallCount" json:"call_count"`                        // 使用次数 (热门排序依据)
	CreatedAt    time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (AIAgent) TableName() string {
	return "AIAgent"
}

// AISession AI 对话会话
type AISession struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SessionID  string    `gorm:"size:100;uniqueIndex;not null;column:SessionId" json:"session_id"` // 会话唯一 ID
	UserID     uint      `gorm:"index;column:UserId" json:"user_id"`                               // 所属用户
	AgentID    uint      `gorm:"index;column:AgentId" json:"agent_id"`                             // 关联智能体
	Agent      AIAgent   `gorm:"foreignKey:AgentID" json:"agent"`                                  // 关联智能体详情
	Topic      string    `gorm:"size:200;column:Topic" json:"topic"`                               // 对话主题
	LastMsg    string    `gorm:"type:text;column:LastMsg" json:"last_msg"`                         // 最后一条消息预览
	Platform   string    `gorm:"size:50;column:Platform" json:"platform"`
	Status     string    `gorm:"size:20;default:'active';column:Status" json:"status"`
	ContextRaw string    `gorm:"type:text;column:ContextRaw" json:"context_raw"` // 额外的上下文元数据 (JSON)
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (AISession) TableName() string {
	return "AISession"
}

// AIChatMessage AI 对话历史消息
type AIChatMessage struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SessionID  string    `gorm:"size:100;index;column:SessionId" json:"session_id"`
	UserID     uint      `gorm:"index;column:UserId" json:"user_id"` // 所属用户
	Role       string    `gorm:"size:20;column:Role" json:"role"`    // system, user, assistant, tool
	Content    string    `gorm:"type:text;column:Content" json:"content"`
	ToolCalls  string    `gorm:"type:text;column:ToolCalls" json:"tool_calls"` // JSON 存储 ToolCall 详情
	UsageToken int       `gorm:"default:0;column:UsageToken" json:"usage_token"`
	CreatedAt  time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

func (AIChatMessage) TableName() string {
	return "AIChatMessage"
}

// BotSkillPermission 机器人技能授权表
type BotSkillPermission struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	BotID     string    `gorm:"index;size:64;column:BotId" json:"bot_id"`
	SkillName string    `gorm:"index;size:100;column:SkillName" json:"skill_name"`
	IsAllowed bool      `gorm:"default:true;column:IsAllowed" json:"is_allowed"`
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

func (BotSkillPermission) TableName() string {
	return "BotSkillPermission"
}

// SessionGORM represents general plugin session state
type SessionGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID string    `gorm:"size:255;uniqueIndex;not null;column:session_id" json:"session_id"`
	UserID    int64     `gorm:"index;column:user_id" json:"user_id"`
	GroupID   int64     `gorm:"index;column:group_id" json:"group_id"`
	State     string    `gorm:"size:255;column:state" json:"state"`
	Data      string    `gorm:"type:jsonb;column:data" json:"data"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SessionGORM) TableName() string {
	return "sessions"
}

// AIAgentTrace AI Agent 执行追踪日志
type AIAgentTrace struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	SessionID   string    `gorm:"index;size:64;column:SessionId" json:"session_id"`
	ExecutionID string    `gorm:"index;size:100;column:ExecutionId" json:"execution_id"` // 关联的任务执行 ID
	TaskID      uint      `gorm:"index;column:TaskId" json:"task_id"`                    // 关联的任务 ID
	BotID       string    `gorm:"index;size:64;column:BotId" json:"bot_id"`
	Step        int       `gorm:"column:Step" json:"step"`
	Type        string    `gorm:"size:32;column:Type" json:"type"` // reasoning, tool_call, tool_result, llm_response
	Content     string    `gorm:"type:text;column:Content" json:"content"`
	Metadata    string    `gorm:"type:text;column:Metadata" json:"metadata"` // JSON 格式的额外信息
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
}

// TableName 设置表名
func (AIAgentTrace) TableName() string {
	return "AIAgentTrace"
}

// DigitalFactoryGoal 数字工厂战略目标
type DigitalFactoryGoal struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	Title       string         `gorm:"size:255;not null;column:Title" json:"title"`
	Description string         `gorm:"type:text;column:Description" json:"description"`
	Status      string         `gorm:"size:20;default:'pending';column:Status" json:"status"` // pending, in_progress, completed, failed
	Progress    float64        `gorm:"default:0;column:Progress" json:"progress"`             // 0-100
	Priority    int            `gorm:"default:1;column:Priority" json:"priority"`
	Deadline    *time.Time     `gorm:"column:Deadline" json:"deadline"`
	CreatedAt   time.Time      `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:UpdatedAt" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:DeletedAt" json:"-"`
}

func (DigitalFactoryGoal) TableName() string {
	return "DigitalFactoryGoal"
}

// DigitalFactoryMilestone 目标里程碑
type DigitalFactoryMilestone struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:Id" json:"id"`
	GoalID      uint      `gorm:"index;column:GoalId" json:"goal_id"`
	Title       string    `gorm:"size:255;not null;column:Title" json:"title"`
	Description string    `gorm:"type:text;column:Description" json:"description"`
	Status      string    `gorm:"size:20;default:'pending';column:Status" json:"status"` // pending, in_progress, completed
	Weight      float64   `gorm:"default:10;column:Weight" json:"weight"`                // 权重，用于计算目标进度
	Order       int       `gorm:"default:0;column:Order" json:"order"`
	CreatedAt   time.Time `gorm:"column:CreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (DigitalFactoryMilestone) TableName() string {
	return "DigitalFactoryMilestone"
}
