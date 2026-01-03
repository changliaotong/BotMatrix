package models

import (
	"time"

	"gorm.io/gorm"
)

// BotEntityGORM represents a bot in the database
type BotEntityGORM struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	SelfID    string         `gorm:"uniqueIndex;size:64;column:self_id" json:"self_id"`
	Nickname  string         `gorm:"size:128;column:nickname" json:"nickname"`
	Platform  string         `gorm:"size:32;column:platform" json:"platform"`
	Status    string         `gorm:"size:32;column:status" json:"status"`
	Connected bool           `gorm:"column:connected" json:"connected"`
	LastSeen  time.Time      `gorm:"column:last_seen" json:"last_seen"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (BotEntityGORM) TableName() string {
	return "bot_entities"
}

// MessageLogGORM represents a message log in the database
type MessageLogGORM struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	BotID     string    `gorm:"index;size:64;column:bot_id" json:"bot_id"`
	UserID    string    `gorm:"index;size:64;column:user_id" json:"user_id"`
	GroupID   string    `gorm:"index;size:64;column:group_id" json:"group_id"`
	Content   string    `gorm:"type:text;column:content" json:"content"`
	RawData   string    `gorm:"type:text;column:raw_data" json:"raw_data"`
	Platform  string    `gorm:"size:32;column:platform" json:"platform"`
	Direction string    `gorm:"size:16;column:direction" json:"direction"` // incoming, outgoing
	CreatedAt time.Time `gorm:"index;column:created_at" json:"created_at"`
}

func (MessageLogGORM) TableName() string {
	return "message_logs"
}

// UserGORM 用户表GORM模型
type UserGORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null;size:255;column:username" json:"username"`
	PasswordHash   string    `gorm:"not null;size:255;column:password_hash" json:"password_hash"`
	IsAdmin        bool      `gorm:"default:false;column:is_admin" json:"is_admin"`
	Active         bool      `gorm:"default:true;column:active" json:"active"`
	SessionVersion int       `gorm:"default:1;column:session_version" json:"session_version"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (UserGORM) TableName() string {
	return "users"
}

// RoutingRuleGORM 路由规则表GORM模型
type RoutingRuleGORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Pattern        string    `gorm:"uniqueIndex;not null;size:500;column:pattern" json:"pattern"`
	TargetWorkerID string    `gorm:"not null;size:255;column:target_worker_id" json:"target_worker_id"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (RoutingRuleGORM) TableName() string {
	return "routing_rules"
}

// GroupCacheGORM 群组缓存表GORM模型
type GroupCacheGORM struct {
	GroupID   string    `gorm:"primaryKey;size:255;column:group_id" json:"group_id"`
	GroupName string    `gorm:"size:255;column:group_name" json:"group_name"`
	BotID     string    `gorm:"size:255;column:bot_id" json:"bot_id"`
	LastSeen  time.Time `gorm:"column:last_seen" json:"last_seen"`
}

// TableName 设置表名
func (GroupCacheGORM) TableName() string {
	return "group_cache"
}

// MemberCacheGORM 群成员缓存表GORM模型
type MemberCacheGORM struct {
	ID       uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID  string    `gorm:"index:idx_group_user,unique;size:255;column:group_id" json:"group_id"`
	UserID   string    `gorm:"index:idx_group_user,unique;size:255;column:user_id" json:"user_id"`
	Nickname string    `gorm:"size:255;column:nickname" json:"nickname"`
	Card     string    `gorm:"size:255;column:card" json:"card"`
	Role     string    `gorm:"size:50;column:role" json:"role"`
	LastSeen time.Time `gorm:"column:last_seen" json:"last_seen"`
}

// TableName 设置表名
func (MemberCacheGORM) TableName() string {
	return "member_cache"
}

// FriendCacheGORM 好友缓存表GORM模型
type FriendCacheGORM struct {
	UserID   string    `gorm:"primaryKey;size:255;column:user_id" json:"user_id"`
	Nickname string    `gorm:"size:255;column:nickname" json:"nickname"`
	Remark   string    `gorm:"size:255;column:remark" json:"remark"`
	BotID    string    `gorm:"size:255;column:bot_id" json:"bot_id"`
	LastSeen time.Time `gorm:"column:last_seen" json:"last_seen"`
}

// TableName 设置表名
func (FriendCacheGORM) TableName() string {
	return "friend_cache"
}

// AIProviderGORM AI 提供商配置
type AIProviderGORM struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string         `gorm:"size:100;not null;column:name" json:"name"`
	Type      string         `gorm:"size:50;not null;column:type" json:"type"` // openai, ollama, deepseek, etc.
	BaseURL   string         `gorm:"size:255;column:base_url" json:"base_url"`
	APIKey    string         `gorm:"size:500;column:api_key" json:"api_key"` // 存储时加密
	IsEnabled bool           `gorm:"default:true;column:is_enabled" json:"is_enabled"`
	Priority  int            `gorm:"default:1;column:priority" json:"priority"`
	UserID    *uint          `gorm:"index;column:user_id" json:"user_id"` // 为空表示系统公共配置，不为空表示用户私有配置
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

// TableName 设置表名
func (AIProviderGORM) TableName() string {
	return "ai_providers"
}

// AIModelGORM AI 具体模型
type AIModelGORM struct {
	ID           uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProviderID   uint           `gorm:"index;column:provider_id" json:"provider_id"`
	Provider     AIProviderGORM `gorm:"foreignKey:ProviderID" json:"provider"`
	ModelID      string         `gorm:"size:100;not null;column:model_name" json:"model_id"` // 实际 API 调用的模型 ID (如 gpt-4)
	ModelName    string         `gorm:"size:100;column:display_name" json:"model_name"`      // 展示名称 (如 GPT-4 Turbo)
	Capabilities string         `gorm:"size:255;column:capabilities" json:"capabilities"`    // JSON array: ["chat", "vision"]
	BaseURL      string         `gorm:"size:255;column:base_url" json:"base_url"`            // 模型级别 BaseURL 覆盖
	APIKey       string         `gorm:"size:500;column:api_key" json:"api_key"`              // 模型级别 APIKey 覆盖
	ContextSize  int            `gorm:"default:4096;column:context_size" json:"context_size"`
	IsDefault    bool           `gorm:"default:false;column:is_default" json:"is_default"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AIModelGORM) TableName() string {
	return "ai_models"
}

// AIPromptTemplateGORM 提示词模板
type AIPromptTemplateGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Scene     string    `gorm:"size:100;uniqueIndex;column:scene" json:"scene"` // 使用场景: intent_recognition, reply_gen, etc.
	Content   string    `gorm:"type:text;column:content" json:"content"`
	Version   string    `gorm:"size:20;column:version" json:"version"`
	IsActive  bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AIPromptTemplateGORM) TableName() string {
	return "ai_prompt_templates"
}

// AIKnowledgeBaseGORM 知识库配置 (RAG)
type AIKnowledgeBaseGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"size:100;not null;column:name" json:"name"`
	Description string    `gorm:"size:255;column:description" json:"description"`
	OwnerID     uint      `gorm:"index;column:owner_id" json:"owner_id"` // 所属用户
	IsPublic    bool      `gorm:"default:false;column:is_public" json:"is_public"`
	EmbedModel  string    `gorm:"size:100;column:embed_model" json:"embed_model"` // 使用的向量化模型
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AIKnowledgeBaseGORM) TableName() string {
	return "ai_knowledge_bases"
}

// AIUsageLogGORM AI 使用日志
type AIUsageLogGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID       uint      `gorm:"index;column:user_id" json:"user_id"`
	AgentID      uint      `gorm:"index;column:agent_id" json:"agent_id"` // 关联智能体 ID
	ModelName    string    `gorm:"size:100;column:model_name" json:"model_name"`
	ProviderType string    `gorm:"size:50;column:provider_type" json:"provider_type"`
	InputTokens  int       `gorm:"column:input_tokens" json:"input_tokens"`
	OutputTokens int       `gorm:"column:output_tokens" json:"output_tokens"`
	TotalTokens  int       `gorm:"column:total_tokens" json:"total_tokens"`
	DurationMS   int       `gorm:"column:duration_ms" json:"duration_ms"`
	Status       string    `gorm:"size:20;column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 设置表名
func (AIUsageLogGORM) TableName() string {
	return "ai_usage_logs"
}

// CognitiveMemoryGORM 认知记忆表
type CognitiveMemoryGORM struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     string    `gorm:"index;size:64;column:user_id" json:"user_id"`
	BotID      string    `gorm:"index;size:64;column:bot_id" json:"bot_id"`
	Category   string    `gorm:"size:32;column:category" json:"category"` // profile, preference, event, fact
	Content    string    `gorm:"type:text;column:content" json:"content"` // 记忆内容
	Importance int       `gorm:"default:1;column:importance" json:"importance"`
	Embedding  string    `gorm:"type:vector(1536);column:embedding" json:"-"` // 向量存储 (支持 pgvector)
	LastSeen   time.Time `gorm:"column:last_seen" json:"last_seen"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 设置表名
func (CognitiveMemoryGORM) TableName() string {
	return "cognitive_memories"
}

// MCPServerGORM MCP 服务器配置模型
type MCPServerGORM struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string         `gorm:"size:100;not null;column:name" json:"name"`
	Type      string         `gorm:"size:20;not null;column:type" json:"type"` // http, sse, webhook, internal
	Endpoint  string         `gorm:"size:500;not null;column:endpoint" json:"endpoint"`
	APIKey    string         `gorm:"size:500;column:api_key" json:"api_key"`
	Scope     string         `gorm:"size:20;default:'user';column:scope" json:"scope"` // global, org, user
	OwnerID   uint           `gorm:"index;column:owner_id" json:"owner_id"`            // 所属用户或组织 ID
	Status    string         `gorm:"size:20;default:'active';column:status" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (MCPServerGORM) TableName() string {
	return "mcp_servers"
}

// AISkillGORM 技能定义
type AISkillGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null;column:name" json:"name"`
	Description string    `gorm:"size:255;column:description" json:"description"`
	Category    string    `gorm:"size:50;column:category" json:"category"` // tools, entertainment, business, etc.
	Prompt      string    `gorm:"type:text;column:prompt" json:"prompt"`
	Config      string    `gorm:"type:text;column:config" json:"config"` // JSON 扩展配置
	OwnerID     uint      `gorm:"index;column:owner_id" json:"owner_id"`
	IsPublic    bool      `gorm:"default:false;column:is_public" json:"is_public"`
	Status      string    `gorm:"size:20;default:'active';column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AISkillGORM) TableName() string {
	return "ai_skills"
}

// AITrainingDataGORM 技能训练数据 (Few-shot / Fine-tuning)
type AITrainingDataGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SkillID   uint      `gorm:"index;column:skill_id" json:"skill_id"`
	Input     string    `gorm:"type:text;not null;column:input" json:"input"`
	Output    string    `gorm:"type:text;not null;column:output" json:"output"`
	Source    string    `gorm:"size:50;column:source" json:"source"` // manual, feedback, auto
	IsUsed    bool      `gorm:"default:true;column:is_used" json:"is_used"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AITrainingDataGORM) TableName() string {
	return "ai_training_data"
}

// AIIntentGORM 意图定义
type AIIntentGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null;column:name" json:"name"` // 意图名称: complaint, feedback, inquiry
	DisplayName string    `gorm:"size:100;column:display_name" json:"display_name"`      // 展示名称: 用户投诉, 意见反馈, 售前咨询
	Description string    `gorm:"size:500;column:description" json:"description"`        // 详细描述该意图的判定标准
	Keywords    string    `gorm:"size:500;column:keywords" json:"keywords"`              // 辅助判定的关键词 (JSON array)
	IsActive    bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AIIntentGORM) TableName() string {
	return "ai_intents"
}

// AIIntentRoutingGORM 意图路由规则
type AIIntentRoutingGORM struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	IntentID   uint      `gorm:"index;column:intent_id" json:"intent_id"`
	TargetType string    `gorm:"size:20;column:target_type" json:"target_type"` // skill, worker, plugin, bot_redirect
	TargetID   string    `gorm:"size:100;column:target_id" json:"target_id"`    // AISkill.ID 或 WorkerID 或 BotID
	Priority   int       `gorm:"default:1;column:priority" json:"priority"`
	Condition  string    `gorm:"type:text;column:condition" json:"condition"` // 额外的过滤条件 (JSON)
	IsActive   bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (AIIntentRoutingGORM) TableName() string {
	return "ai_intent_routings"
}

// GroupBotRoleGORM 群组机器人角色配置 (用于用户侧多机协作)
type GroupBotRoleGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID   string    `gorm:"index:idx_group_bot,unique;size:255;column:group_id" json:"group_id"`
	BotID     string    `gorm:"index:idx_group_bot,unique;size:255;column:bot_id" json:"bot_id"`
	Role      string    `gorm:"size:100;column:role" json:"role"`           // 角色: dispatcher(调度官), handler(执行者), monitor(监控)
	Specialty string    `gorm:"size:255;column:specialty" json:"specialty"` // 擅长领域: tech, finance, chat, etc.
	IsActive  bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (GroupBotRoleGORM) TableName() string {
	return "group_bot_roles"
}

// MCPToolGORM MCP 工具缓存模型
type MCPToolGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ServerID    uint      `gorm:"index;column:server_id" json:"server_id"`
	Name        string    `gorm:"size:100;not null;column:name" json:"name"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	InputSchema string    `gorm:"type:text;column:input_schema" json:"input_schema"` // JSON Schema
	IsActive    bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MCPToolGORM) TableName() string {
	return "mcp_tools"
}

// EnterpriseGORM 企业/组织模型
type EnterpriseGORM struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"size:255;uniqueIndex;not null;column:name" json:"name"` // 企业名称
	Code        string         `gorm:"size:100;uniqueIndex;not null;column:code" json:"code"` // 企业唯一代码 (用于 B2B 通信)
	Description string         `gorm:"type:text;column:description" json:"description"`
	OwnerID     uint           `gorm:"index;column:owner_id" json:"owner_id"`                // 企业所有者 (关联 UserGORM)
	Config      string         `gorm:"type:text;column:config" json:"config"`                // 企业级全局配置 (JSON)
	Status      string         `gorm:"size:20;default:'active';column:status" json:"status"` // active, suspended
	PublicKey   string         `gorm:"type:text;column:public_key" json:"public_key"`        // 用于 B2B 安全认证的公钥
	PrivateKey  string         `gorm:"type:text;column:private_key" json:"private_key"`      // 用于 B2B 安全认证的私钥 (加密存储)
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

// TableName 设置表名
func (EnterpriseGORM) TableName() string {
	return "enterprises"
}

// PlatformAccountGORM 第三方平台账号配置 (公众号, 抖音等)
type PlatformAccountGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EnterpriseID uint      `gorm:"index;column:enterprise_id" json:"enterprise_id"`
	Platform     string    `gorm:"size:50;not null;column:platform" json:"platform"` // wechat_mp, tiktok, weibo, etc.
	AccountName  string    `gorm:"size:100;column:account_name" json:"account_name"` // 账号名称
	AccountID    string    `gorm:"size:100;column:account_id" json:"account_id"`     // 平台内部 ID (如 AppID)
	Config       string    `gorm:"type:text;column:config" json:"config"`            // 平台配置 (JSON: AppSecret, Token, AESKey 等)
	Status       string    `gorm:"size:20;default:'active';column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (PlatformAccountGORM) TableName() string {
	return "platform_accounts"
}

// DigitalEmployeeGORM 数字员工模型 (Bot 的高级拟人化抽象)
type DigitalEmployeeGORM struct {
	ID                uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EnterpriseID      uint      `gorm:"index;column:enterprise_id" json:"enterprise_id"`                     // 所属企业
	BotID             string    `gorm:"uniqueIndex;size:64;column:bot_id" json:"bot_id"`                     // 关联的底层机器人
	EmployeeID        string    `gorm:"uniqueIndex;size:64;column:employee_id" json:"employee_id"`           // 工号
	Name              string    `gorm:"size:100;not null;column:name" json:"name"`                           // 姓名 (如: 张三)
	Title             string    `gorm:"size:100;column:title" json:"title"`                                  // 职位 (如: 高级售后工程师)
	Department        string    `gorm:"size:100;column:department" json:"department"`                        // 部门 (如: 技术部)
	Bio               string    `gorm:"type:text;column:bio" json:"bio"`                                     // 个人简介/人设定义
	AgentID           uint      `gorm:"index;column:agent_id" json:"agent_id"`                               // 关联的 AI 智能体 ID
	Skills            string    `gorm:"type:text;column:skills" json:"skills"`                               // 技能列表 (JSON: ["complaint_handling", "log_analysis"])
	OnboardingAt      time.Time `gorm:"column:onboarding_at" json:"onboarding_at"`                           // 入职时间
	Status            string    `gorm:"size:20;default:'active';column:status" json:"status"`                // 状态: active(在职), training(培训中), retired(离职)
	OnlineStatus      string    `gorm:"size:20;default:'offline';column:online_status" json:"online_status"` // online, offline, busy
	SalaryToken       int64     `gorm:"default:0;column:salary_token" json:"salary_token"`                   // 累计消耗 Token (作为薪资统计)
	SalaryLimit       int64     `gorm:"default:1000000;column:salary_limit" json:"salary_limit"`             // Token 预算限制
	KpiScore          float64   `gorm:"default:100;column:kpi_score" json:"kpi_score"`                       // KPI 评分 (基于满意度打分)
	ExternalCommLevel int       `gorm:"default:0;column:external_comm_level" json:"external_comm_level"`     // 外部通信等级: 0(禁止), 1(仅限白名单企业), 2(公开)
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (DigitalEmployeeGORM) TableName() string {
	return "digital_employees"
}

// EnterpriseMemberGORM 企业成员表
type EnterpriseMemberGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EnterpriseID uint      `gorm:"index:idx_ent_user,unique;column:enterprise_id" json:"enterprise_id"`
	UserID       uint      `gorm:"index:idx_ent_user,unique;column:user_id" json:"user_id"`
	Role         string    `gorm:"size:50;default:'member';column:role" json:"role"` // admin, hr, supervisor, member
	JoinedAt     time.Time `gorm:"column:joined_at" json:"joined_at"`
}

// TableName 设置表名
func (EnterpriseMemberGORM) TableName() string {
	return "enterprise_members"
}

// B2BConnectionGORM 企业间 B2B 连接
type B2BConnectionGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SourceEntID  uint      `gorm:"index:idx_b2b_conn;column:source_ent_id" json:"source_ent_id"`     // 发起方企业
	TargetEntID  uint      `gorm:"index:idx_b2b_conn;column:target_ent_id" json:"target_ent_id"`     // 接收方企业
	Status       string    `gorm:"size:20;default:'pending';column:status" json:"status"`            // pending, active, blocked
	AuthProtocol string    `gorm:"size:50;default:'mtls';column:auth_protocol" json:"auth_protocol"` // mtls, oauth2, custom
	Config       string    `gorm:"type:text;column:config" json:"config"`                            // 连接特定配置
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// B2BSkillSharingGORM B2B 技能共享授权表
type B2BSkillSharingGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SourceEntID uint      `gorm:"index;column:source_ent_id" json:"source_ent_id"`       // 提供技能的企业
	TargetEntID uint      `gorm:"index;column:target_ent_id" json:"target_ent_id"`       // 使用技能的企业
	SkillName   string    `gorm:"size:100;column:skill_name" json:"skill_name"`          // 共享的技能名称
	AliasName   string    `gorm:"size:100;column:alias_name" json:"alias_name"`          // 在目标企业的别名 (可选)
	Status      string    `gorm:"size:20;default:'pending';column:status" json:"status"` // pending, approved, rejected, blocked
	IsActive    bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (B2BSkillSharingGORM) TableName() string {
	return "b2b_skill_sharings"
}

// DigitalEmployeeDispatchGORM 数字员工外派授权表
type DigitalEmployeeDispatchGORM struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EmployeeID  uint      `gorm:"index;column:employee_id" json:"employee_id"`           // 外派的员工 ID
	SourceEntID uint      `gorm:"index;column:source_ent_id" json:"source_ent_id"`       // 所属企业
	TargetEntID uint      `gorm:"index;column:target_ent_id" json:"target_ent_id"`       // 接收企业
	Status      string    `gorm:"size:20;default:'pending';column:status" json:"status"` // pending, approved, rejected, recalled
	Permissions string    `gorm:"type:text;column:permissions" json:"permissions"`      // 授予的权限列表 (JSON: ["chat", "skill_call"])
	DispatchAt  time.Time `gorm:"column:dispatch_at" json:"dispatch_at"`
	ExpireAt    time.Time `gorm:"column:expire_at" json:"expire_at"` // 有效期 (可选)
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DigitalEmployeeDispatchGORM) TableName() string {
	return "digital_employee_dispatches"
}

// TableName 设置表名
func (B2BConnectionGORM) TableName() string {
	return "b2b_connections"
}

// DigitalEmployeeKpiGORM 数字员工考核日志
type DigitalEmployeeKpiGORM struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EmployeeID uint      `gorm:"index;column:employee_id" json:"employee_id"`
	MetricName string    `gorm:"size:100;column:metric_name" json:"metric_name"` // 考核项: response_speed, accuracy, satisfaction
	Score      float64   `gorm:"column:score" json:"score"`
	Detail     string    `gorm:"type:text;column:detail" json:"detail"` // 考核详情 (关联的消息 ID 或评价内容)
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 设置表名
func (DigitalEmployeeKpiGORM) TableName() string {
	return "digital_employee_kpis"
}

// AIAgentGORM 智能体定义 (LLM 核心配置)
type AIAgentGORM struct {
	ID           uint        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name         string      `gorm:"size:100;not null;column:name" json:"name"`
	Description  string      `gorm:"type:text;column:description" json:"description"`
	SystemPrompt string      `gorm:"type:text;column:system_prompt" json:"system_prompt"`
	ModelID      uint        `gorm:"index;column:model_id" json:"model_id"`
	Model        AIModelGORM `gorm:"foreignKey:ModelID" json:"model"`
	Temperature  float32     `gorm:"default:0.7;column:temperature" json:"temperature"`
	MaxTokens    int         `gorm:"default:2048;column:max_tokens" json:"max_tokens"`
	Tools        string      `gorm:"type:text;column:tools" json:"tools"` // JSON array of tool names or IDs
	IsVoice      bool        `gorm:"default:false;column:is_voice" json:"is_voice"`
	VoiceID      string      `gorm:"size:100;column:voice_id" json:"voice_id"`
	VoiceName    string      `gorm:"size:100;column:voice_name" json:"voice_name"`
	VoiceLang    string      `gorm:"size:50;column:voice_lang" json:"voice_lang"`
	VoiceRate    float32     `gorm:"default:1.0;column:voice_rate" json:"voice_rate"`                      // 语速 0.1 - 10
	OwnerID      uint        `gorm:"index;column:owner_id" json:"owner_id"`                                // 所属用户 ID
	Visibility   string      `gorm:"size:20;default:'public';column:visibility" json:"visibility"`         // public, private, link_only
	RevenueRate  float64     `gorm:"type:decimal(10,4);default:0;column:revenue_rate" json:"revenue_rate"` // 收益率 (每 1k tokens 收益多少算力)
	CallCount    int         `gorm:"default:0;column:call_count" json:"call_count"`                        // 使用次数 (热门排序依据)
	CreatedAt    time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (AIAgentGORM) TableName() string {
	return "ai_agents"
}

// AISessionGORM AI 对话会话
type AISessionGORM struct {
	ID         uint        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  string      `gorm:"size:100;uniqueIndex;not null;column:session_id" json:"session_id"` // 会话唯一 ID
	UserID     uint        `gorm:"index;column:user_id" json:"user_id"`                               // 所属用户
	AgentID    uint        `gorm:"index;column:agent_id" json:"agent_id"`                             // 关联智能体
	Agent      AIAgentGORM `gorm:"foreignKey:AgentID" json:"agent"`                                   // 关联智能体详情
	Topic      string      `gorm:"size:200;column:topic" json:"topic"`                                // 对话主题
	LastMsg    string      `gorm:"type:text;column:last_msg" json:"last_msg"`                         // 最后一条消息预览
	Platform   string      `gorm:"size:50;column:platform" json:"platform"`
	Status     string      `gorm:"size:20;default:'active';column:status" json:"status"`
	ContextRaw string      `gorm:"type:text;column:context_raw" json:"context_raw"` // 额外的上下文元数据 (JSON)
	CreatedAt  time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (AISessionGORM) TableName() string {
	return "ai_sessions"
}

// AIChatMessageGORM AI 对话历史消息
type AIChatMessageGORM struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID  string    `gorm:"size:100;index;column:session_id" json:"session_id"`
	UserID     uint      `gorm:"index;column:user_id" json:"user_id"` // 所属用户
	Role       string    `gorm:"size:20;column:role" json:"role"`     // system, user, assistant, tool
	Content    string    `gorm:"type:text;column:content" json:"content"`
	ToolCalls  string    `gorm:"type:text;column:tool_calls" json:"tool_calls"` // JSON 存储 ToolCall 详情
	UsageToken int       `gorm:"default:0;column:usage_token" json:"usage_token"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (AIChatMessageGORM) TableName() string {
	return "ai_chat_messages"
}

// BotSkillPermissionGORM 机器人技能授权表
type BotSkillPermissionGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BotID     string    `gorm:"index;size:64;column:bot_id" json:"bot_id"`
	SkillName string    `gorm:"index;size:100;column:skill_name" json:"skill_name"`
	IsAllowed bool      `gorm:"default:true;column:is_allowed" json:"is_allowed"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (BotSkillPermissionGORM) TableName() string {
	return "bot_skill_permissions"
}

// AIAgentTraceGORM AI Agent 执行追踪日志
type AIAgentTraceGORM struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID string    `gorm:"index;size:64;column:session_id" json:"session_id"`
	BotID     string    `gorm:"index;size:64;column:bot_id" json:"bot_id"`
	Step      int       `gorm:"column:step" json:"step"`
	Type      string    `gorm:"size:32;column:type" json:"type"` // reasoning, tool_call, tool_result, llm_response
	Content   string    `gorm:"type:text;column:content" json:"content"`
	Metadata  string    `gorm:"type:text;column:metadata" json:"metadata"` // JSON 格式的额外信息
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 设置表名
func (AIAgentTraceGORM) TableName() string {
	return "ai_agent_traces"
}
